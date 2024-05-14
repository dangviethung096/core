package core

import (
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type TaskInfo struct {
	Data []byte
}

type TaskHandler func(ctx Context, task TaskInfo)

/*
* Handle task: handle a task message is published from message queue
 */
func HandleTask(taskQueueName string, handler TaskHandler) Error {
	queueConfig := QueueConfig{
		ExchangeName: BLANK,
		QueueName:    fmt.Sprintf("%s%s", TASK_PREFIX_QUEUE_NAME, taskQueueName),
		RouteKey:     BLANK,
		Kind:         MESSAGE_QUEUE_KIND_DIRECT,
		AutoAck:      true,
		Durable:      false,
		AutoDelete:   false,
		Exclusive:    false,
		NoWait:       false,
		Args:         nil,
	}

	session, err := MessageQueue().CreateSimpleSession(queueConfig)

	if err != nil {
		LogError("Create message queue session fail: %s", err.Error())
		return err
	}

	// Retry to connect
	go func(sess *simpleMessageQueueSession, taskHandler TaskHandler) {
		for err := range sess.channel.NotifyClose(make(chan *amqp.Error)) {
			// Retry connect
			sess.channel.Close()
			// Add retry connect
			LogError("Channel disconnected: retry to connect: %s", err.Error())
			for !sess.recreateSession() {
				time.Sleep(time.Second * time.Duration(Config.RabbitMQ.RetryTime))
				sess.connection.retryConnect()
			}

			handleTask(sess, taskHandler)
		}
	}(session, handler)

	return handleTask(session, handler)
}

func handleTask(session *simpleMessageQueueSession, handler TaskHandler) Error {
	messages, errConsume := session.channel.Consume(session.config.QueueName, BLANK, true, session.config.Exclusive, false, session.config.NoWait, nil)
	if errConsume != nil {
		LogError("Error when handle task: %s", errConsume.Error())
		return ERROR_CANNOT_CONSUME_MESSAGES_FROM_RABBITMQ
	}

	go func(messages <-chan amqp.Delivery) {
		for message := range messages {
			ctx := GetContextWithTimeout(Config.GetTaskTimeout())
			ctx.LogInfo("Start handle task: %s", session.config.QueueName)
			handler(ctx, TaskInfo{
				Data: message.Body,
			})
			ctx.LogInfo("End handle task: %s", session.config.QueueName)
			PutContext(ctx)
		}
	}(messages)

	return nil
}
