package core

import (
	"database/sql"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
)

type StartTaskRequest struct {
	TaskName  string `validate:"required"`
	QueueName string `validate:"required"`
	Data      []byte
	Time      time.Time `validate:"required"`
	Interval  int64     // Seconds
	Loop      int64
}

func StartTask(ctx Context, request *StartTaskRequest) Error {
	ctx.LogInfo("Receive StartTaskRequest: %#v", *request)

	if err := validateStartTaskRequest(request); err != nil {
		ctx.LogError("Validate task request fail: %v, err = %s", *request, err.Error())
		return ERROR_ADD_TASK_SYSTEM_FAIL
	}

	if request.Loop == -1 {
		request.Loop = math.MaxInt64
	}

	if request.Interval < 0 {
		ctx.LogError("Interval is invalid: %#v", *request)
		return ERROR_TASK_REQUEST_INVALID
	}

	var loopIndex uint64
	var nextTime time.Time

	if request.Interval != 0 {
		loopIndex, nextTime = calculateNextTime(request.Time, request.Interval)
	} else {
		loopIndex = 0
		nextTime = request.Time
	}

	now := time.Now()
	if nextTime.Before(now.Add(-time.Minute * 30)) {
		ctx.LogError("Task is in the past less than 30 minutes: %#v", *request)
		return ERROR_TASK_IS_EXPIRED
	}

	// Calculate new time
	bucket := GetBucket(nextTime) // Generate bucket id

	// Get id from id genrator
	var taskId int64

	// Init transaction
	tx, err := DBSession().GetOriginConnection(ctx).BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		ctx.LogError("Begin transaction fail: %v, err = %s", *request, err.Error())
		return ERROR_ADD_TASK_SYSTEM_FAIL
	}
	defer tx.Rollback()

	// Use UPSERT to handle concurrent requests safely
	// This will either insert a new task or update existing one atomically
	ctx.LogInfo("Upsert task: task name = %s, queue name = %s, startTime = %s, loopCount = %d, interval = %d",
		request.TaskName, request.QueueName, request.Time.String(), request.Loop, request.Interval)

	row := tx.QueryRowContext(ctx,
		`INSERT INTO scheduler_tasks(task_name, queue_name, data, done, loop_index, loop_count, next, interval, start_time, source, next_time) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) 
		 ON CONFLICT (task_name) DO UPDATE SET
			queue_name = EXCLUDED.queue_name,
			data = EXCLUDED.data,
			done = EXCLUDED.done,
			loop_index = EXCLUDED.loop_index,
			loop_count = EXCLUDED.loop_count,
			next = EXCLUDED.next,
			interval = EXCLUDED.interval,
			start_time = EXCLUDED.start_time,
			source = EXCLUDED.source,
			next_time = EXCLUDED.next_time
		 RETURNING id;`,
		request.TaskName, request.QueueName, request.Data, false, loopIndex, request.Loop, nextTime.Unix(), request.Interval, request.Time.Format(time.RFC3339), Config.Server.Name, nextTime.Format(time.RFC3339))

	if err := row.Scan(&taskId); err != nil {
		ctx.LogError("Upsert task fail: %v, err = %s", *request, err.Error())
		return ERROR_ADD_TASK_SYSTEM_FAIL
	}

	// Use UPSERT for scheduler_todo to handle concurrent requests safely
	// This will either insert a new todo entry or update existing one atomically
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO scheduler_todo(task_id, bucket, next_time, source) 
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (task_id) DO UPDATE SET
			bucket = EXCLUDED.bucket,
			next_time = EXCLUDED.next_time,
			source = EXCLUDED.source;`,
		taskId, bucket, nextTime.Format(time.RFC3339), Config.Server.Name); err != nil {
		ctx.LogError("Upsert todo fail: %v, err = %s", *request, err.Error())
		return ERROR_ADD_TASK_SYSTEM_FAIL
	}

	if err := tx.Commit(); err != nil && err != sql.ErrTxDone {
		ctx.LogError("Commit fail: %v, err = %s", *request, err.Error())
		return ERROR_ADD_TASK_SYSTEM_FAIL
	}

	ctx.LogInfo("Insert new task success: id = %d, task name = %s, queue name = %s, startTime = %s, loopCount = %d, interval = %d", taskId, request.TaskName, request.QueueName, request.Time.String(), request.Loop, request.Interval)
	return nil
}

func validateStartTaskRequest(request *StartTaskRequest) Error {
	if err := validate.Struct(*request); err != nil {
		LogError("Task request is invalid: %s", err.Error())
		return ERROR_TASK_REQUEST_INVALID
	}

	return nil
}

type StopTaskRequest struct {
	Id uint64
}

func StopTask(ctx Context, request *StopTaskRequest) Error {
	ctx.LogInfo("Receive StopTaskRequest: %#v", *request)
	tx, err := DBSession().GetOriginConnection(ctx).BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		ctx.LogError("Begin transaction fail: %v, error = %s", *request, err.Error())
		return ERROR_STOP_TASK_FAIL
	}

	// Delete todo in database
	if _, err := tx.ExecContext(ctx, "DELETE FROM scheduler_todo WHERE task_id = $1", request.Id); err != nil {
		ctx.LogError("Delete task from todo fail: %d, %s", request.Id, err.Error())
		return ERROR_STOP_TASK_FAIL
	}

	// Delete task in database
	if _, err := tx.ExecContext(ctx, "DELETE FROM scheduler_tasks WHERE id = $1", request.Id); err != nil {
		ctx.LogError("Delete task from tasks fail: %d, %s", request.Id, err.Error())
		return ERROR_STOP_TASK_FAIL
	}

	// Run transaction
	if err := tx.Commit(); err != nil {
		ctx.LogError("Commit fail: %v, error = %s", *request, err.Error())
		return ERROR_STOP_TASK_FAIL
	}

	ctx.LogInfo("Stop task %d success", request.Id)
	return nil
}

func StopTaskByTaskName(ctx Context, taskName string) Error {
	ctx.LogInfo("Receive StopTaskByTaskNameRequest: %s", taskName)
	var id uint64
	row := DBSession().GetOriginConnection(ctx).QueryRowContext(ctx, "SELECT id FROM scheduler_tasks WHERE task_name = $1", taskName)
	if err := row.Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			ctx.LogInfo("Task %s not found", taskName)
			return ERROR_TASK_NOT_FOUND
		}
		ctx.LogError("Get task id fail: %v, error = %s", taskName, err.Error())
		return ERROR_STOP_TASK_FAIL
	}

	ctx.LogInfo("Stop task_id = %d, task name = %s", id, taskName)
	err := StopTask(ctx, &StopTaskRequest{
		Id: id,
	})

	return err
}

func TriggerTask(ctx Context, taskName string, taskData []byte) Error {
	return pushTaskToQueue(ctx, TaskMessage{
		Data:          taskData,
		TaskID:        uint64(0),
		TaskName:      taskName,
		TaskQueueName: taskName,
	})
}

func StartOneTimeTask(ctx Context, queueName string, startTime time.Time, taskData []byte) (string, Error) {
	randomId := uuid.New().String()
	taskName := fmt.Sprintf("onetime_%s_%s", queueName, randomId)

	now := time.Now()
	if startTime.Before(now) {
		ctx.LogError("StartOneTimeTask: startTime is before now: %s", startTime.Format(time.RFC3339))
		return BLANK, ERROR_TASK_REQUEST_INVALID
	}

	return taskName, StartTask(ctx, &StartTaskRequest{
		TaskName:  taskName,
		QueueName: queueName,
		Time:      startTime,
		Data:      taskData,
		Interval:  0,
		Loop:      0,
	})
}

func GetTaskByName(ctx Context, taskName string) (*task, Error) {
	taskData := task{}
	query := "SELECT id, queue_name, data, done, loop_index, loop_count, next, interval, source, start_time, task_name FROM scheduler_tasks WHERE task_name = $1"
	ctx.LogInfo("Query = %s, taskName = %s", query, taskName)
	row := DBSession().GetOriginConnection(ctx).QueryRowContext(ctx, query, taskName)
	if err := row.Scan(&taskData.ID, &taskData.QueueName, &taskData.Data, &taskData.Done, &taskData.LoopIndex, &taskData.LoopCount, &taskData.Next, &taskData.Interval, &taskData.Source, &taskData.StartTime, &taskData.TaskName); err != nil {
		if err == sql.ErrNoRows {
			ctx.LogInfo("Task %s not found", taskName)
			return nil, ERROR_TASK_NOT_FOUND
		}
		ctx.LogError("Get task fail: %v, error = %s", taskName, err.Error())
		return nil, ERROR_SERVER_ERROR
	}

	return &taskData, nil
}
