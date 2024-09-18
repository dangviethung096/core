package core

import (
	"fmt"
)

type TaskInfo struct {
	Data []byte
}

type TaskHandler func(ctx Context, task TaskInfo)

/*
* Handle task: handle a task message is published from message queue
 */
func HandleTask(ctx Context, taskQueueName string, handler TaskHandler) Error {
	topicName := fmt.Sprintf("%s%s", TASK_PREFIX_QUEUE_NAME, taskQueueName)
	group := "TaskGroup"

	natsHandler := func(topic string, data []byte) {
		newCtx := GetContextWithoutTimeout()
		defer PutContext(newCtx)
		// Handle task
		newCtx.LogInfo("Handle task: %s", taskQueueName)
		handler(newCtx, TaskInfo{
			Data: data,
		})
	}

	err := MessageQueue().SubscribeGroup(ctx, topicName, group, natsHandler)
	if err != nil {
		ctx.LogError("Error when handle task: %s", err.Error())
		return err
	}

	return nil
}

func pushTaskToQueue(ctx Context, taskQueueName string, taskData []byte) Error {
	topicName := fmt.Sprintf("%s%s", TASK_PREFIX_QUEUE_NAME, taskQueueName)
	return MessageQueue().Publish(ctx, topicName, taskData)
}
