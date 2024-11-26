package core

import (
	"database/sql"
	"fmt"
	"math"
	"time"
)

const TASK_TEMPLATE_KEY = "TASK:%d"
const TASK_PREFIX_QUEUE_NAME = "task_queue_"

type TaskStatus string

const (
	TaskStatus_Doing TaskStatus = "DOING"
	TaskStatus_Done  TaskStatus = "DONE"
)

type worker struct {
	id string
}

func NewWorker() *worker {
	return &worker{
		id: ID.GenerateID(),
	}
}

func (w *worker) Start(delay time.Duration, interval time.Duration) {
	go func() {
		time.Sleep(delay)
		ticker := time.NewTicker(interval)
		LogInfo("Start task!")
		for {
			select {
			case <-done:
				ticker.Stop()
				return
			case <-ticker.C:
				w.execute()
			}
		}
	}()
}

/*
* todo struct: contain taskid and bucket value that
* describe about bucket time that task must be done
 */
type todo struct {
	taskId int64
	bucket int64
}

/*
* execute: find all task from todo table and do it
 */
func (w *worker) execute() {
	bucket := GetBucket(time.Now())

	// Get all task from database: table: todo
	result, err := DBSession().QueryContext(coreContext, "SELECT task_id, bucket FROM scheduler_todo WHERE bucket <= $1 and source = $2", bucket, Config.Server.Name)
	if err != nil {
		LogError("Execute tasks fail: %v", err)
		return
	}

	todos := []todo{}
	var taskId int64
	var tBucket int64

	for result.Next() {
		err := result.Scan(&taskId, &tBucket)
		if err != nil {
			LogError("Get task fail: %v", err)
			return
		}

		todos = append(todos, todo{
			taskId: taskId,
			bucket: tBucket,
		})
	}

	if len(todos) == 0 {
		return
	}

	for _, todo := range todos {
		// Block this task by redis or lwt in database: use distributed log
		taskKey := fmt.Sprintf(TASK_TEMPLATE_KEY, todo.taskId)
		locker := NewPgLock(mainDbSession, taskKey)
		err := locker.Lock()
		if err != nil {
			LogInfo("Key %s existed: %v", taskKey, err)
			continue
		}

		defer locker.Unlock()
		// Process data
		LogDebug("Execute task: %d", todo.taskId)
		w.process(todo.bucket, todo.taskId)
	}
}

func (w *worker) process(bucket int64, id int64) {
	var t task
	// Get task detail from database in table: tasks
	row := DBSession().QueryRowContext(coreContext, "SELECT id, queue_name, data, done, loop_index, loop_count, next, interval, start_time FROM scheduler_tasks WHERE id = $1", id)
	err := row.Scan(&t.ID, &t.QueueName, &t.Data, &t.Done, &t.LoopIndex, &t.LoopCount, &t.Next, &t.Interval, &t.StartTime)
	if err != nil {
		LogError("Get task fail: %v", err)
		return
	}
	LogDebug("Execute task: id = %d, name = %s", t.ID, t.QueueName)

	if t.Done {
		// Delete task in table: todo
		if _, err := DBSession().ExecContext(coreContext, "DELETE FROM scheduler_todo WHERE task_id = $1", id); err != nil {
			LogError("Cannot delete todo task: %d", id)
		}
		return
	}

	// Start run this task: use rabbitmqt
	now := time.Now()
	err = pushTaskToQueue(coreContext, t.QueueName, t.Data)
	if err != nil {
		LogError("Cannot run task: %v, err = %s", t, err.Error())
		_, err := DBSession().ExecContext(coreContext, "INSERT INTO scheduler_done(bucket, task_id, operation_time, status) VALUES ($1, $2, $3, $4)", bucket, t.ID, now.Format(time.RFC3339), TASK_FAIL)
		if err != nil {
			LogError("Cannot insert task to done table: %v", err)
		}
	} else {
		_, err := DBSession().ExecContext(coreContext, "INSERT INTO scheduler_done(bucket, task_id, operation_time, status) VALUES ($1, $2, $3, $4)", bucket, t.ID, now.Format(time.RFC3339), TASK_DONE)
		if err != nil {
			LogError("Cannot insert task to done table: %v", err)
		}
	}

	startTime, err := time.Parse(time.RFC3339, t.StartTime)
	if err != nil {
		LogError("Cannot parse start time: %v", err)
		return
	}

	loopIndex, nextTime := calculateNextTime(startTime, t.Interval)
	t.LoopIndex = uint64(loopIndex)
	t.Next = nextTime.Unix()

	if t.LoopIndex < t.LoopCount {
		next := time.Unix(t.Next, 0)
		newBucket := GetBucket(next)
		// Update new task in table: todo, task (time of next task)
		tx, err := DBSession().BeginTx(coreContext, &sql.TxOptions{})
		if err != nil {
			LogError("Start transaction fail: %v", err)
		}
		defer tx.Rollback()
		// Delete old todo
		if _, err := tx.ExecContext(coreContext, "DELETE FROM scheduler_todo WHERE task_id = $1 AND bucket = $2", id, bucket); err != nil {
			LogError("Fail to delete task in todo: id = %d, bucket %d", id, bucket)
		}

		// Insert new record in todo task
		if _, err := tx.ExecContext(coreContext, "INSERT INTO scheduler_todo(task_id, bucket, next_time, source) VALUES ($1, $2, $3, $4);", id, newBucket, next.Format(time.RFC3339), Config.Server.Name); err != nil {
			LogError("Update todo task fail: id = %d, bucket = %d, err = %s", id, newBucket, err.Error())
		}

		// Update in task
		if _, err := tx.ExecContext(coreContext, "UPDATE scheduler_tasks SET next = $1, loop_index = $2, next_time = $3 WHERE id = $4", t.Next, t.LoopIndex, nextTime.Format(time.RFC3339), t.ID); err != nil {
			LogError("Update task fail: %v", t)
		}

		if err := tx.Commit(); err != nil {
			if err != sql.ErrTxDone {
				LogError("Commit transaction fail: task = %v, %s", t, err.Error())
			}
		}

	} else {
		tx, err := DBSession().BeginTx(coreContext, &sql.TxOptions{})
		if err != nil {
			LogError("Start transaction fail: %v", err)
		}
		defer tx.Rollback()

		// Delete old todo
		if _, err := tx.ExecContext(coreContext, "DELETE FROM scheduler_todo WHERE task_id = $1 AND bucket = $2", id, bucket); err != nil {
			LogError("Fail to delete task in todo: id = %d, bucket %d", id, bucket)
		}

		// Update task in task table: is it done? => done
		if _, err := tx.ExecContext(coreContext, "UPDATE scheduler_tasks SET done = true, loop_index = $1 WHERE id = $2", t.LoopIndex, id); err != nil {
			LogError("Update task in table fail: task = %v, err = %s", t, err.Error())
		}

		if err := tx.Commit(); err != nil {
			if err != sql.ErrTxDone {
				LogError("Commit transaction fail: task = %v, %s", t, err.Error())
			}
		}
	}
}

func calculateNextTime(start time.Time, interval int64) (uint64, time.Time) {
	if interval <= 0 {
		return math.MaxUint64, time.Unix(0, 0)
	}
	now := time.Now().Unix()
	startTime := start.Unix()
	var loopIndex uint64
	var nextTime time.Time
	if startTime < now {
		loopIndex = uint64((now-startTime)/interval) + 1
		nextTime = time.Unix(int64(loopIndex)*interval+startTime, 0)
	} else {
		loopIndex = 0
		nextTime = start
	}

	return uint64(loopIndex), nextTime
}
