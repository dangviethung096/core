package core

import (
	"encoding/json"
	"fmt"
)

type EventHandler[T any] func(ctx Context, data T)

/*
* Register event to event bus
* @param: event string
* @param: handler EventHandler[T]
 */
func RegisterEvent[T any](event string, handler EventHandler[T]) {
	eventTopic := fmt.Sprintf("event.%s", event)

	err := MessageQueue().Subscribe(coreContext, eventTopic, func(topic string, data []byte) {
		ctx := GetContextWithoutTimeout()
		defer PutContext(ctx)

		request := initRequest[T]()
		if data != nil {
			if err := json.Unmarshal(data, &request); err != nil {
				ctx.LogError("Unmarshal data fail. topic: %s, Error: %v", topic, err)
				return
			}
		}

		handler(ctx, request)
	})
	if err != nil {
		coreContext.LogFatal("Fail to subscribe topic: %s, Error: %v", event, err)
	}
}

/*
Publish event to event bus
@param: ctx Context
@param: event string
@param: data T
@return: Error
*/
func PublishEvent[T any](ctx Context, event string, data T) Error {
	eventTopic := fmt.Sprintf("event.%s", event)
	byteData, err := json.Marshal(data)
	if err != nil {
		ctx.LogError("Cannot sent publish event. Marshal data fail. topic: %s, Error: %v", eventTopic, err)
		return ERROR_CANNOT_PUBLISH_MESSAGE
	}

	return MessageQueue().Publish(ctx, eventTopic, byteData)
}
