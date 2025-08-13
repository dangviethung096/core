package core

import (
	"encoding/json"
	"net/http"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

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

	h := func(c *gin.Context) {
		ctx := getWebsocketContext(c)
		defer putWebsocketContext(ctx)

		// Run middlewares
		for _, middleware := range middlewares {
			if err := middleware(ctx, ctx.Writer, ctx.Request); err != nil {
				handshakeContext := getHttpContext(ctx.Context)
				buildContext(handshakeContext)
				handshakeContext.requestID = ctx.GetContextID()
				handshakeContext.writeError(err)
				putHttpContext(handshakeContext)
				return
			}
		}

		conn, err := websocketUpgrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if err != nil {
			ctx.LogError("websocket upgrade failed: %v", err)
			return
		}
		defer conn.Close()

		// Create channels for communication
		done := make(chan struct{})
		errChan := make(chan error)

		// Start read pump in a separate goroutine
		go func() {
			defer close(done)
			for {
				messageType, message, err := conn.ReadMessage()
				if err != nil {
					if !websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
						ctx.LogError("read error: %v", err)
						errChan <- err
					}
					return
				}

				// Handle message in a separate goroutine
				go func() {
					req := initRequest[T]()
					if err := json.Unmarshal(message, &req); err != nil {
						ctx.LogError("unmarshal error: %v", err)
						errChan <- err
						return
					}
					ctx.messageType = messageType

					res, err := handler(ctx, req)
					if err != nil {
						ctx.LogError("handler error: %v", err)
						errChan <- err
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

					// Use connection write mutex
					conn.SetWriteDeadline(time.Now().Add(writeWait))
					if err := conn.WriteJSON(wsResponse); err != nil {
						ctx.LogError("write error: %v", err)
						errChan <- err
					}
				}()
			}
		}()

		// Handle connection closure
		select {
		case <-done:
			return
		case err := <-errChan:
			ctx.LogError("websocket error: %v", err)
			return
		}
	}

	router.GET(url, h)
}

// Constants for websocket handling
const (
	writeWait = 10 * time.Second
)
