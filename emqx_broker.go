package core

import (
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type emqxClient struct {
	mqtt.Client
}

func NewEmqxClient(emqxConfig EmqxConfig) MqttClient {
	log.Printf("Connect to emqx broker: %v", emqxConfig.Broker)
	broker := emqxConfig.Broker
	prefixClientID := emqxConfig.PrefixClient
	now := time.Now().UnixNano()
	clientID := fmt.Sprintf("%s_%d", prefixClientID, now)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)

	client := mqtt.NewClient(opts)
	emqxClient := &emqxClient{Client: client}
	err := emqxClient.Connect()
	if err != nil {
		log.Fatalf("Connect to emqx broker fail: %v\n", err)
	}

	return emqxClient
}

func (c *emqxClient) Connect() Error {
	token := c.Client.Connect()
	token.Wait()
	if token.Error() != nil {
		return NewError(ERROR_CODE_FROM_MQTT, token.Error().Error())
	}
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

func (c *emqxClient) Subscribe(ctx Context, topic string, handler MqttMessageHandler) Error {
	// Subscribe to the specified topic with QoS 1
	token := c.Client.Subscribe(topic, MQTT_QOS_EXACTLY_ONCE, func(client mqtt.Client, message mqtt.Message) {
		newContext := GetContextWithoutTimeout()
		defer PutContext(newContext)

		handler(newContext, c, MqttMessage{
			MessageID: int64(message.MessageID()),
			Topic:     message.Topic(),
			Payload:   message.Payload(),
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
