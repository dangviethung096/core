package core

import (
	"encoding/json"
	"fmt"
)

type TaskInfo struct {
	Data []byte
}

type TaskMessage struct {
	Data          []byte `json:"data"`
	TaskID        uint64 `json:"task_id"`
	TaskName      string `json:"task_name"`
	TaskQueueName string `json:"task_queue_name"`
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

		var taskMessage TaskMessage
		err := json.Unmarshal(data, &taskMessage)
		if err != nil {
			newCtx.LogError("Error when unmarshal task message: %v", err)
			return
		}

		// Handle task
		newCtx.LogInfo("Handle task from nats message: topic: %s, task_id: %d, task_name: %s, task_queue_name: %s", topic, taskMessage.TaskID, taskMessage.TaskName, taskMessage.TaskQueueName)
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

func pushTaskToQueue(ctx Context, taskMessage TaskMessage) Error {
	topicName := fmt.Sprintf("%s%s", TASK_PREFIX_QUEUE_NAME, taskMessage.TaskQueueName)
	data, err := json.Marshal(taskMessage)
	if err != nil {
		ctx.LogError("Error when marshal task message: %v", err)
		return ERROR_TASK_REQUEST_INVALID
	}
	return MessageQueue().Publish(ctx, topicName, data)
}
