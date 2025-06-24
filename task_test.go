package core

import (
	"testing"
	"time"
)

func TestStartSchedule_ReturnSuccess(t *testing.T) {
	ctx := coreContext

	task := StartTaskRequest{
		TaskName:  "test-queue",
		QueueName: "test-queue",
		Time:      time.Now().Add(time.Second * 2),
		Loop:      0,
		Interval:  0,
		Data:      []byte("test-data"),
	}

	StartTask(ctx, &task)

	HandleTask(ctx, "test-queue", func(ctx Context, task TaskInfo) {
		t.Logf("Handle task: data = %s", string(task.Data))
	})

	time.Sleep(time.Second * 5)
}

func TestStopTaskByTaskName(t *testing.T) {
	ctx := coreContext

	task := StartTaskRequest{
		TaskName:  "test-queue",
		QueueName: "test-queue",
		Time:      time.Now().Add(time.Second * 2),
		Loop:      0,
		Interval:  0,
		Data:      []byte("test-data"),
	}

	StartTask(ctx, &task)

	err := StopTaskByTaskName(ctx, "test-queue")
	if err != nil {
		t.Fatalf("Failed to stop task: %v", err)
	}

	time.Sleep(time.Second * 5)
}

func TestStartTask_ReplaceTask(t *testing.T) {
	ctx := coreContext

	task := StartTaskRequest{
		TaskName:  "test-queue",
		QueueName: "test-queue",
		Time:      time.Now().Add(time.Second * 60),
		Loop:      0,
		Interval:  0,
		Data:      []byte("test-data"),
	}

	StartTask(ctx, &task)

	task.Data = []byte("test-data-2")
	StartTask(ctx, &task)

	time.Sleep(time.Second * 5)
}
