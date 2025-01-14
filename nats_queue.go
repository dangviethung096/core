package core

import (
	"log"

	"github.com/nats-io/nats.go"
)

type natsClient struct {
	nc *nats.Conn
}

type NatsSubscriberHandler func(topic string, data []byte)

func connectToNatsQueue(queueUrl string) natsClient {
	nc, err := nats.Connect(queueUrl)
	if err != nil {
		log.Fatalf("Cannnot connect to Nats Server: %s, error = %v", queueUrl, err)
	}

	log.Printf("Connected to NATS server successfully: %s", queueUrl)

	return natsClient{nc: nc}
}

func (client natsClient) Subscribe(ctx Context, topic string, handler NatsSubscriberHandler) Error {
	natsHandler := func(msg *nats.Msg) {
		data := msg.Data
		handler(topic, data)
	}

	_, err := client.nc.Subscribe(topic, natsHandler)
	if err != nil {
		ctx.LogInfo("Fail to subscribe topic: %s, err = %v", topic, err)
		return ERROR_CANNOT_SUBSCRIBE_QUEUE
	}

	return nil
}

func (client natsClient) SubscribeGroup(ctx Context, topic string, group string, handler NatsSubscriberHandler) Error {
	natsHandler := func(msg *nats.Msg) {
		data := msg.Data
		handler(topic, data)
	}

	_, err := client.nc.QueueSubscribe(topic, group, natsHandler)
	if err != nil {
		ctx.LogInfo("Fail to subscribe topic: %s, err = %v", topic, err)
		return ERROR_CANNOT_SUBSCRIBE_QUEUE
	}

	return nil
}

func (client natsClient) Publish(ctx Context, topic string, data []byte) Error {
	err := client.nc.Publish(topic, data)
	if err != nil {
		ctx.LogInfo("Fail to publish message to topic: %s, err = %v", topic, err)
		return ERROR_CANNOT_PUBLISH_MESSAGE
	}

	return nil
}
