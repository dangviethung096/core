package core

import (
	"log"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

type natsClient struct {
	nc *nats.Conn
	// Add fields to track subscriptions
	subscriptions map[string]*subscriptionInfo
	mu            sync.RWMutex
	queueUrl      string
	reconnecting  bool
}

type subscriptionInfo struct {
	topic   string
	group   string
	handler NatsSubscriberHandler
}

type NatsSubscriberHandler func(topic string, data []byte)

func connectToNatsQueue(queueUrl string) {
	queueClient = &natsClient{
		subscriptions: make(map[string]*subscriptionInfo),
		queueUrl:      queueUrl,
	}

	queueClient.connectWithReconnection()
}

func (client *natsClient) connectWithReconnection() {
	// Set up connection options with reconnection
	opts := []nats.Option{
		nats.ReconnectWait(1 * time.Second),
		nats.MaxReconnects(-1), // Unlimited reconnections
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			LogError("NATS disconnected: %v", err)
			client.reconnecting = true
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			LogInfo("NATS reconnected to: %s", nc.ConnectedUrl())
			client.reconnecting = false
			// Resubscribe to all topics after reconnection
			client.resubscribeAll()
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			client.reconnecting = false
			LogError("NATS connection closed")
		}),
	}

	nc, err := nats.Connect(client.queueUrl, opts...)
	if err != nil {
		log.Fatalf("Cannot connect to NATS Server: %s, error = %v", client.queueUrl, err)
	}

	client.nc = nc
	log.Printf("Connected to NATS server successfully: %s", client.queueUrl)
}

func (client *natsClient) resubscribeAll() {
	client.mu.RLock()
	defer client.mu.RUnlock()

	LogInfo("Resubscribing to %d topics after reconnection", len(client.subscriptions))

	for _, subInfo := range client.subscriptions {
		if subInfo.group != BLANK {
			// Resubscribe to queue group
			_, err := client.nc.QueueSubscribe(subInfo.topic, subInfo.group, func(msg *nats.Msg) {
				subInfo.handler(subInfo.topic, msg.Data)
			})
			if err != nil {
				LogError("Failed to resubscribe to topic: %s, group: %s, err: %v", subInfo.topic, subInfo.group, err)
			}
			LogInfo("Resubscribed to topic: %s, group: %s", subInfo.topic, subInfo.group)
		} else {
			// Resubscribe to regular topic
			_, err := client.nc.Subscribe(subInfo.topic, func(msg *nats.Msg) {
				subInfo.handler(subInfo.topic, msg.Data)
			})
			if err != nil {
				LogError("Failed to resubscribe to topic: %s, err: %v", subInfo.topic, err)
			}
			LogInfo("Resubscribed to topic: %s", subInfo.topic)
		}
	}
}

func (client *natsClient) Subscribe(ctx Context, topic string, handler NatsSubscriberHandler) Error {
	client.mu.Lock()
	defer client.mu.Unlock()

	// Store subscription info for reconnection
	key := topic
	client.subscriptions[key] = &subscriptionInfo{
		topic:   topic,
		handler: handler,
	}

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

func (client *natsClient) SubscribeGroup(ctx Context, topic string, group string, handler NatsSubscriberHandler) Error {
	client.mu.Lock()
	defer client.mu.Unlock()

	// Store subscription info for reconnection
	key := topic + ":" + group
	client.subscriptions[key] = &subscriptionInfo{
		topic:   topic,
		group:   group,
		handler: handler,
	}

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

func (client *natsClient) Publish(ctx Context, topic string, data []byte) Error {
	// Check if we're reconnecting and wait a bit
	if client.reconnecting {
		time.Sleep(100 * time.Millisecond)
	}

	err := client.nc.Publish(topic, data)
	if err != nil {
		ctx.LogInfo("Fail to publish message to topic: %s, err = %v", topic, err)
		return ERROR_CANNOT_PUBLISH_MESSAGE
	}

	return nil
}

// Add method to manually reconnect if needed
func (client *natsClient) Reconnect() {
	client.reconnecting = true
	client.nc.Close()
	client.connectWithReconnection()
}

// Add method to close connection
func (client *natsClient) Close() {
	if client.nc != nil {
		client.nc.Close()
	}
}
