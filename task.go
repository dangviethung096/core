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
	ctx.LogInfo("Receive StartTaskRequest: %+v", *request)

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
	if nextTime.Before(now) {
		ctx.LogError("Task is expired: %#v", *request)
		return ERROR_TASK_IS_EXPIRED
	}

	// Calculate new time
	bucket := GetBucket(nextTime) // Generate bucket id

	// Get id from id genrator
	var taskId int64

	// Init transaction
	tx, err := DBSession().BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		ctx.LogError("Begin transaction fail: %v, err = %s", *request, err.Error())
		return ERROR_ADD_TASK_SYSTEM_FAIL
	}
	defer tx.Rollback()

	// check task in database
	var id string
	var startTime string
	var source string
	var loopCount, interval int64
	var queueName string
	row := DBSession().QueryRowContext(ctx, "SELECT id, start_time, loop_count, interval, source, queue_name FROM scheduler_tasks WHERE task_name = $1", request.TaskName)
	err = row.Scan(&id, &startTime, &loopCount, &interval, &source, &queueName)
	if err == nil {
		if startTime == request.Time.Format(time.RFC3339) && loopCount == int64(request.Loop) && interval == request.Interval && source == Config.Server.Name && queueName == request.QueueName {
			ctx.LogInfo("Task %#v already exist in db with id = %s", *request, id)
			return ERROR_TASK_ALREADY_EXISTED
		}

		ctx.LogInfo("Replace task: %s, startTime: %s, loopCount: %d, interval: %d", id, startTime, loopCount, interval)
		if _, err := tx.ExecContext(ctx, "DELETE FROM scheduler_tasks WHERE id = $1;", id); err != nil {
			ctx.LogError("delete task fail: %s, err = %s", id, err.Error())
			return ERROR_REMOVE_OLD_TASK_FAIL
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM scheduler_todo WHERE task_id = $1;", id); err != nil {
			ctx.LogError("delete todo task fail: %s, err = %s", id, err.Error())
			return ERROR_REMOVE_OLD_TASK_FAIL
		}
	}

	ctx.LogInfo("Insert new task: %d, startTime: %s, loopCount: %d, interval: %d", taskId, request.Time.String(), request.Loop, request.Interval)
	row = tx.QueryRowContext(ctx,
		"INSERT INTO scheduler_tasks(task_name, queue_name, data, done, loop_index, loop_count, next, interval, start_time, source, next_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id;",
		request.TaskName, request.QueueName, request.Data, false, loopIndex, request.Loop, nextTime.Unix(), request.Interval, request.Time.Format(time.RFC3339), Config.Server.Name, nextTime.Format(time.RFC3339))
	if err := row.Scan(&taskId); err != nil {
		ctx.LogError("Insert task fail: %v, err = %s", *request, err.Error())
		return ERROR_ADD_TASK_SYSTEM_FAIL
	}

	if _, err := tx.ExecContext(ctx, "INSERT INTO scheduler_todo(task_id, bucket, next_time, source) VALUES ($1, $2, $3, $4);", taskId, bucket, nextTime.Format(time.RFC3339), Config.Server.Name); err != nil {
		ctx.LogError("Insert task fail: %v, err = %s", *request, err.Error())
		return ERROR_ADD_TASK_SYSTEM_FAIL
	}

	if err := tx.Commit(); err != nil && err != sql.ErrTxDone {
		ctx.LogError("Commit fail: %v, err = %s", *request, err.Error())
		return ERROR_ADD_TASK_SYSTEM_FAIL
	}

	ctx.LogInfo("Insert new task: %d success", taskId)
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
	tx, err := DBSession().BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		ctx.LogError("Begin transaction fail: %v, error = %s", *request, err.Error())
		return ERROR_STOP_TASK_FAIL
	}

	// Delete todo in database
	if _, err := tx.ExecContext(ctx, "DELETE FROM todo WHERE id = $1", request.Id); err != nil {
		ctx.LogError("Delete task from todo fail: %d, %s", request.Id, err.Error())
		return ERROR_STOP_TASK_FAIL
	}

	// Delete task in database
	if _, err := tx.ExecContext(ctx, "DELETE FROM task WHERE id = $1", request.Id); err != nil {
		ctx.LogError("Delete task from tasks fail: %d, %s", request.Id, err.Error())
		return ERROR_STOP_TASK_FAIL
	}
	return nil
}

func TriggerTask(ctx Context, taskName string, taskData []byte) Error {
	return pushTaskToQueue(ctx, taskName, taskData)
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
