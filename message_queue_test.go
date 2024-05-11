package core

import (
	"fmt"
	"log"
	"testing"
	"time"
)

const exchangeName = "exchange_name"
const queueName = "queue_name"
const routeKey = "route_key"
const kind = "direct"
const durable = false
const autoDelete = false
const exclusive = false
const noWaits = false

func InitMessageQueueTest() *messageQueue {
	Config.RabbitMQ.AMQPServerURL = "amqp://guest:guest@localhost:5672"
	rabbitMQConnection := connectRabbitMQ()
	return rabbitMQConnection
}

func publishMessage() {
	rabbitMQConnection := InitMessageQueueTest()
	defer rabbitMQConnection.connection.Close()

	session, err := rabbitMQConnection.CreateSession(QueueConfig{
		ExchangeName: exchangeName,
		QueueName:    queueName,
		RouteKey:     routeKey,
		Kind:         kind,
		Durable:      durable,
		AutoDelete:   autoDelete,
		Exclusive:    exclusive,
		NoWait:       noWaits,
		Args:         nil,
	})
	if err != nil {
		LoggerInstance.Error("error when create session: %v\n", err)
		return
	}

	defer session.CloseSession()

	for i := 0; i < 10; i++ {
		msg := fmt.Sprintf("Hello %d", i)
		err = session.Publish([]byte(msg))
		if err != nil {
			LoggerInstance.Error("error in publish message: %s, %s\n", msg, err)
		}
	}
}

func releaseMessageQueueTest(connection *messageQueue) {
	ch, err := connection.connection.Channel()
	if err != nil {
		LoggerInstance.Error("Cannot create a channel: %s", err.Error())
		return
	}
	defer ch.Close()

	if err = ch.ExchangeUnbind(queueName, routeKey, exchangeName, noWaits, nil); err != nil {
		LoggerInstance.Error("Exchange unbind fail: %s", err.Error())
		return
	}

	if err = ch.ExchangeDelete(exchangeName, false, false); err != nil {
		LoggerInstance.Error("Delete exchange fail: %s", err.Error())
		return
	}

	if data, err := ch.QueueDelete(queueName, false, false, false); err != nil {
		LoggerInstance.Error("Delete exchange fail: %s", err.Error())
		return
	} else {
		fmt.Printf("Queue Delete: %d\n", data)
	}

	connection.connection.Close()
}

func TestConnectMesageQueue_ReturnSuccess(t *testing.T) {
	rabbitMQConnection := InitMessageQueueTest()
	defer releaseMessageQueueTest(rabbitMQConnection)
	if rabbitMQConnection == nil {
		t.Errorf("connect to message queu fail")
	}
}

func TestCreateMessageQueueSession_ReturnSuccess(t *testing.T) {
	rabbitMQConnection := InitMessageQueueTest()
	if rabbitMQConnection == nil {
		t.Errorf("connect to message queu fail")
	}
	defer releaseMessageQueueTest(rabbitMQConnection)

	session, err := rabbitMQConnection.CreateSession(QueueConfig{
		ExchangeName: exchangeName,
		QueueName:    queueName,
		RouteKey:     routeKey,
		Kind:         kind,
		Durable:      durable,
		AutoDelete:   autoDelete,
		Exclusive:    exclusive,
		NoWait:       noWaits,
		Args:         nil,
	})

	if err != nil {
		t.Errorf("error when create session: %v", err)
		return
	}

	session.CloseSession()
}

func TestPublishMessageInMessageQueue_ReturnSuccess(t *testing.T) {
	rabbitMQConnection := InitMessageQueueTest()
	defer releaseMessageQueueTest(rabbitMQConnection)
	if rabbitMQConnection == nil {
		t.Errorf("connect to message queu fail")
	}

	session, err := rabbitMQConnection.CreateSession(QueueConfig{
		ExchangeName: exchangeName,
		QueueName:    queueName,
		RouteKey:     routeKey,
		Kind:         kind,
		Durable:      durable,
		AutoDelete:   autoDelete,
		Exclusive:    exclusive,
		NoWait:       noWaits,
		Args:         nil,
	})

	if err != nil {
		t.Errorf("error when create session: %v\n", err)
		return
	}

	defer session.CloseSession()

	err = session.Publish([]byte("Hello from there!"))
	if err != nil {
		t.Errorf("error in publish message: %v\n", err)
		return
	}
}

func TestConsumeMessageInMessageQueue_ReturnSuccess(t *testing.T) {
	rabbitMQConnection := InitMessageQueueTest()
	defer releaseMessageQueueTest(rabbitMQConnection)

	if rabbitMQConnection == nil {
		t.Errorf("connect to message queu fail")
	}
	publishMessage()

	session, err := rabbitMQConnection.CreateSession(QueueConfig{
		ExchangeName: exchangeName,
		QueueName:    queueName,
		RouteKey:     routeKey,
		Kind:         kind,
		Durable:      durable,
		AutoDelete:   autoDelete,
		Exclusive:    exclusive,
		NoWait:       noWaits,
		Args:         nil,
	})

	if err != nil {
		t.Errorf("error when create session: %v\n", err)
		return
	}

	defer session.CloseSession()

	messages := make(chan string)

	err = session.Consume(func(ctx Context, message RabbitmqMessage) {
		messages <- string(message.Body)
	})

	go func() {
		time.Sleep(time.Second)
		close(messages)
	}()

	if err != nil {
		t.Errorf("error in subscribe message: %v\n", err)
		return
	}

	flag := false
	for message := range messages {
		if message != "" {
			log.Printf("Receive message from channel: %s", message)
			flag = true
		}
	}

	if !flag {
		t.Errorf("Fail to receive message!")
	}
}

func consume(t *testing.T, messages chan string, id int, done chan bool) {
	rabbitMQConnection := InitMessageQueueTest()
	defer releaseMessageQueueTest(rabbitMQConnection)

	session, err := rabbitMQConnection.CreateSession(QueueConfig{
		ExchangeName: exchangeName,
		QueueName:    queueName,
		RouteKey:     routeKey,
		Kind:         kind,
		Durable:      durable,
		AutoDelete:   autoDelete,
		Exclusive:    exclusive,
		NoWait:       noWaits,
		Args:         nil,
	})

	if err != nil {
		t.Errorf("error when create session: %v\n", err)
		return
	}

	defer session.CloseSession()

	err = session.Consume(func(ctx Context, message RabbitmqMessage) {
		messages <- fmt.Sprintf("Message from %d: %s", id, string(message.Body))
		time.Sleep(time.Millisecond * 100)
	})

	if err != nil {
		t.Errorf("error in subscribe message: %v\n", err)
		return
	}

	switch {
	case <-done:
		return
	default:
		fmt.Printf("Wait 0.1 seconds")
		time.Sleep(time.Millisecond * 100)
	}
}

func TestConsumeMessageFromTwoChannelInMessageQueue_ReturnSuccess(t *testing.T) {
	messages := make(chan string)
	done := make(chan bool)

	go func() {
		consume(t, messages, 1, done)
	}()

	go func() {
		consume(t, messages, 2, done)
	}()

	publishMessage()

	go func() {
		time.Sleep(time.Second * 3)
		close(messages)
	}()

	flag := false
	for message := range messages {
		if message != "" {
			log.Printf("%s", message)
			flag = true
		}
	}

	if !flag {
		t.Errorf("Fail to receive message!")
	}

	done <- true
}

func TestConsumeMessageRetryConnect_ReturnSuccess(t *testing.T) {
	rabbitMQConnection := InitMessageQueueTest()
	defer releaseMessageQueueTest(rabbitMQConnection)

	session, err := rabbitMQConnection.CreateSession(QueueConfig{
		ExchangeName: exchangeName,
		QueueName:    queueName,
		RouteKey:     routeKey,
		Kind:         kind,
		Durable:      durable,
		AutoDelete:   autoDelete,
		Exclusive:    exclusive,
		NoWait:       noWaits,
		Args:         nil,
	})
	if err != nil {
		LoggerInstance.Error("error when create session: %v\n", err)
		return
	}

	defer session.CloseSession()

	session.Consume(func(ctx Context, msg RabbitmqMessage) {
		fmt.Printf("Consumer data: %s\n", string(msg.Body))
	})

	go func() {
		time.Sleep(time.Second * 5)
		publishMessage()
	}()

	time.Sleep(10 * time.Second)
}
