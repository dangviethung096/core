package core

import (
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type emqxClient struct {
	mqtt.Client
}

func NewEmqxClient(emqxConfig EmqxConfig) MqttClient {
	broker := emqxConfig.Broker
	prefixClientID := emqxConfig.PrefixClient
	now := time.Now().UnixNano()
	clientID := fmt.Sprintf("%s_%d", prefixClientID, now)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)

	client := mqtt.NewClient(opts)

	return &emqxClient{Client: client}
}

func (c *emqxClient) Connect() Error {
	return nil
}

func (c *emqxClient) Publish(ctx Context, topic string, payload []byte) Error {
	// Publish the message to the specified topic with QoS 0 and no retained message
	token := c.Client.Publish(topic, MQTT_QOS_EXACTLY_ONCE, false, payload)

	// Wait for the token to confirm the message was sent
	token.Wait()

	// Check if there was an error during publishing
	if token.Error() != nil {
		// Return the error if publishing failed
		return NewError(ERROR_CODE_FROM_MQTT, token.Error().Error())
	}

	// Return nil if publishing was successful
	return nil
}

func (c *emqxClient) Subscribe(ctx Context, topic string, handler func(ctx Context, client MqttClient, message MqttMessage[any])) Error {
	// Subscribe to the specified topic with QoS 1
	token := c.Client.Subscribe(topic, MQTT_QOS_EXACTLY_ONCE, func(client mqtt.Client, message mqtt.Message) {
		payload := message.Payload()

		handler(ctx, c, MqttMessage[any]{
			MessageID: int64(message.MessageID()),
			Topic:     message.Topic(),
		})
	})

	// Wait for the subscription to complete
	token.Wait()

	// Check if there was an error during subscription
	if token.Error() != nil {
		// Return the error if subscription failed
		return NewError(ERROR_CODE_FROM_MQTT, token.Error().Error())
	}

	// Return nil if subscription was successful
	return nil
}

func (c *emqxClient) Unsubscribe(ctx Context, topic string) Error {
	// Unsubscribe from the specified topic
	token := c.Client.Unsubscribe(topic)

	// Wait for the unsubscription to complete
	token.Wait()

	// Check if there was an error during unsubscription
	if token.Error() != nil {
		// Return the error if unsubscription failed
		return NewError(ERROR_CODE_FROM_MQTT, token.Error().Error())
	}

	// Return nil if unsubscription was successful
	return nil
}

func (c *emqxClient) Disconnect(ctx Context) Error {
	c.Client.Disconnect(WAIT_MQTT_DISCONNECT_TIMEOUT)
	return nil
}
