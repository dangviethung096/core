package core

import (
	"testing"
	"time"
)

func TestStartSchedule_ReturnSuccess(t *testing.T) {
	ctx := coreContext

	task := StartTaskRequest{
		QueueName: "test-queue",
		Time:      time.Now().Add(time.Second * 2),
		Loop:      1,
		Interval:  1,
		Data:      []byte("test-data"),
	}

	StartTask(ctx, &task)

	HandleTask(ctx, "test-queue", func(ctx Context, task TaskInfo) {
		t.Logf("Handle task: data = %s", string(task.Data))
	})

	time.Sleep(time.Second * 5)
}
