package core

type MqttClient interface {
	Connect() Error
	Publish(ctx Context, topic string, payload []byte) Error
	Subscribe(ctx Context, topic string, handler MqttMessageHandler) Error
	Unsubscribe(ctx Context, topic string) Error
	Disconnect(ctx Context) Error
}

type MqttMessage struct {
	MessageID int64
	Topic     string
	Payload   any
}

type MqttMessageHandler func(ctx Context, client MqttClient, message MqttMessage)
