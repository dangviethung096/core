package core

type MqttClient interface {
	Publish(ctx Context, topic string, payload []byte) Error
	Subscribe(ctx Context, topic string, payload any, handler MqttMessageHandler) Error
	Unsubscribe(ctx Context, topic string) Error
	Disconnect(ctx Context) Error
}

type MqttMessage struct {
	MessageID int64
	Topic     string
	Payload   any
}

type MqttMessageHandler func(client MqttClient, message MqttMessage)
