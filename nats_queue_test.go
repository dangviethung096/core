package core

import (
	"fmt"
	"testing"
	"time"
)

func TestNatsQueuePubSub(t *testing.T) {
	ctx := coreContext
	queueClient.Subscribe(ctx, "test-topic", func(topic string, data []byte) {
		t.Logf("Received message: %s", string(data))
	})

	ctx.LogInfo("Publishing message to topic: %s", "test-topic")
	queueClient.Publish(ctx, "test-topic", []byte("Hello, World!"))

	time.Sleep(1 * time.Second)
}

func TestNatsInManySubs(t *testing.T) {
	ctx := coreContext
	for i := 0; i < 5; i++ {
		queueClient.Subscribe(ctx, "test-topic", func(topic string, data []byte) {
			t.Logf("Received message in sub %d: %s", i, string(data))
		})
	}

	for i := 0; i < 3; i++ {
		queueClient.Publish(ctx, "test-topic", []byte(fmt.Sprintf("Hello, World! %d", i)))
	}

	time.Sleep(3 * time.Second)
}

func TestNatsQueueGroup(t *testing.T) {
	ctx := coreContext
	queueClient.SubscribeGroup(ctx, "test-topic", "test-group", func(topic string, data []byte) {
		t.Logf("Received message in sub 01: %s", string(data))
	})

	queueClient.SubscribeGroup(ctx, "test-topic", "test-group", func(topic string, data []byte) {
		t.Logf("Received message in sub 02: %s", string(data))

	})

	for i := 0; i < 10; i++ {
		queueClient.Publish(ctx, "test-topic", []byte(fmt.Sprintf("Hello, World! %d", i)))
	}

	time.Sleep(3 * time.Second)
}
