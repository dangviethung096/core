package core

import (
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/gorilla/websocket"
)

type websocketRoute struct {
	url     string
	handler func(w http.ResponseWriter, r *http.Request)
}

var websocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  MAX_WEBSOCKET_READ_BUFFER_SIZE,
	WriteBufferSize: MAX_WEBSOCKET_WRITE_BUFFER_SIZE,
}

type WebsocketResponse struct {
	MessageType int
	Code        int
	Message     string
	Data        any
}

type WebsocketMiddleware func(ctx WebsocketContext, w http.ResponseWriter, r *http.Request) HttpError

type WebsocketHandler[T any] func(ctx WebsocketContext, data T) (*WebsocketResponse, Error)

func RegisterWebsocket[T any](url string, handler WebsocketHandler[T], middlewares ...WebsocketMiddleware) {
	LogInfo("Register Websocket: %s", url)

	// Check if T is a struct
	tType := reflect.TypeOf((*T)(nil)).Elem()
	if tType.Kind() != reflect.Struct {
		LogFatal("Handler request parameter must be a struct, got: %s", tType.Kind())
	}

	h := func(w http.ResponseWriter, r *http.Request) {
		// Get context
		ctx := getWebsocketContext()
		defer putWebsocketContext(ctx)

		// Run middlewares
		for _, middleware := range middlewares {
			err := middleware(ctx, w, r)
			if err != nil {
				// Return error
				handshakeContext := getHttpContext()
				buildContext(handshakeContext, w, r)
				handshakeContext.requestID = ctx.GetContextID()
				handshakeContext.writeError(err)
				putHttpContext(handshakeContext)
				return
			}
		}

		connection, err := websocketUpgrader.Upgrade(w, r, nil)
		if err != nil {
			ctx.LogError("websocket upgrade failed: %v", err)
			return
		}

		for {
			// Read a message
			messageType, message, err := connection.ReadMessage()
			if err != nil {
				ctx.LogInfo("Reading message end: %v", err)
				connection.Close()
				return
			}
			ctx.LogInfo("Received message: %v", string(message))
			// Unmarshal the received message
			req := initRequest[T]()
			err = json.Unmarshal(message, &req)
			if err != nil {
				ctx.LogError("Error unmarshalling message: %v", err)
				connection.Close()
				return
			}
			ctx.messageType = messageType

			res, err := handler(ctx, req)
			if err != nil {
				ctx.LogError("Error handling message: %v", err)
				connection.Close()
				return
			}

			if res.MessageType == 0 {
				res.MessageType = ctx.messageType
			}

			wsResponse := responseBody{
				Code:    res.Code,
				Message: res.Message,
				Data:    res.Data,
			}

			resJson, err := json.Marshal(wsResponse)
			if err != nil {
				ctx.LogError("Error marshalling response: %v", err)
				connection.Close()
				return
			}

			// Echo the message back
			if err := connection.WriteMessage(res.MessageType, resJson); err != nil {
				ctx.LogError("Error writing message: %v", err)
				connection.Close()
				return
			}
		}
	}

	websocketRouteMap[url] = websocketRoute{url: url, handler: h}
}
