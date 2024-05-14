package core

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type simpleMessageQueueSession struct {
	channel    *amqp.Channel
	connection *messageQueue
	config     QueueConfig
}

func (mq *messageQueue) CreateSimpleSession(config QueueConfig) (*simpleMessageQueueSession, Error) {
	channel, err := mq.connection.Channel()
	if err != nil {
		LogError("Could not open channel with RabbitMQ: %s", err.Error())
		return nil, ERROR_CANNOT_CREATE_RABBITMQ_CHANNEL
	}

	_, originalErr := channel.QueueDeclare(
		config.QueueName,
		config.Durable,
		config.AutoDelete,
		config.Exclusive,
		config.NoWait,
		config.Args,
	)

	if originalErr != nil {
		LogError("Error when declare queue: %v", originalErr)
		return nil, ERROR_CANNOT_DECLARE_QUEUE
	}

	session := &simpleMessageQueueSession{
		channel:    channel,
		connection: mq,
		config:     config,
	}

	return session, nil
}

func (mqs *simpleMessageQueueSession) CloseSession() {
	if mqs.channel != nil {
		mqs.channel.Close()
	}
}

func (mqs *simpleMessageQueueSession) recreateSession() bool {
	channel, err := mqs.connection.connection.Channel()
	if err != nil {
		LogError("Could not open channel with RabbitMQ: %s", err.Error())
		return false
	}

	_, originalErr := channel.QueueDeclare(
		mqs.config.QueueName,
		mqs.config.Durable,
		mqs.config.AutoDelete,
		mqs.config.Exclusive,
		mqs.config.NoWait,
		mqs.config.Args,
	)

	if originalErr != nil {
		LogError("Error when declare queue: %s", originalErr.Error())
		return false
	}

	mqs.channel = channel
	return true
}

func (mqs *simpleMessageQueueSession) publish(body []byte) Error {
	err := mqs.channel.PublishWithContext(
		coreContext,
		mqs.config.ExchangeName,
		mqs.config.QueueName,
		false,
		false,
		amqp.Publishing{
			ContentType: CONTENT_TYPE_TEXT,
			Body:        body,
		},
	)

	if err != nil {
		LogError("Publish message: error %s", err.Error())
		return ERROR_SERVER_ERROR
	}

	return nil
}
