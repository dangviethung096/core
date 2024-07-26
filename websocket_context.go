package core

import (
	"context"
	"time"
)

type WebsocketContext interface {
	// Reimplement from context.Context
	context.Context
	logger
	Value(key any) any
	GetContextID() string
	GetTimeout() time.Duration
	GetCancelFunc() func()
	GetMessageType() int
}

type websocketContext struct {
	context.Context
	requestID   string
	timeout     time.Duration
	cancelFunc  context.CancelFunc
	messageType int
}

func (w *websocketContext) GetContextID() string {
	return w.requestID
}

func (w *websocketContext) GetTimeout() time.Duration {
	return w.timeout
}

func (w *websocketContext) GetCancelFunc() func() {
	return w.cancelFunc
}

func (w *websocketContext) Value(key any) any {
	return w.Context.Value(key)
}

/*
* GetContext: Get context from pool
* @return: Context
 */
func getWebsocketContext() *websocketContext {
	ctx := websocketContextPool.Get().(*websocketContext)
	ctx.Context, ctx.cancelFunc = context.WithTimeout(coreContext, contextTimeout)
	ctx.timeout = contextTimeout
	ctx.requestID = ID.GenerateID()
	return ctx
}

/*
* PutContext: Put context to pool
* @params: Context
* @return: void
 */
func putWebsocketContext(ctx *websocketContext) {
	ctx.cancelFunc()
	// Release memory of context

	// Put context to pool
	websocketContextPool.Put(ctx)
}

func (w *websocketContext) GetMessageType() int {
	return w.messageType
}
