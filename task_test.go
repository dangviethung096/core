package core

import (
	"testing"
	"time"
)

func TestStartSchedule_ReturnSuccess(t *testing.T) {
	queueConfig := QueueConfig{
		ExchangeName: BLANK,
		QueueName:    "TestTask",
		RouteKey:     BLANK,
		Kind:         "direct",
		Durable:      false,
		AutoDelete:   false,
		Exclusive:    false,
		NoWait:       false,
		Args:         nil,
	}
	session, err := MessageQueue().CreateSimpleSession(queueConfig)
	defer session.CloseSession()
	if err != nil {
		t.Errorf("Create message fail: %s", err.Error())
	}

	data := "Hello from scheduler"
	StartTask(coreContext, &StartTaskRequest{
		QueueName: queueConfig.QueueName,
		Data:      []byte(data),
		Time:      time.Now().Add(time.Second * 5),
		Interval:  0,
		Loop:      0,
	})

	done := make(chan bool)
	go func() {
		HandleTask(queueConfig.QueueName, func(ctx Context, taskInfo TaskInfo) {
			if string(taskInfo.Data) != data {
				t.Errorf("Not receive exactly message: expected: %s, actually: %s", data, string(taskInfo.Data))
			}
			done <- true
		})
	}()

	if !(<-done) {
		t.Error("Task doesn't done!")
	}
}
