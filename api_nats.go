package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

func RegisterNatsAPI[T any](url string, method string, handler Handler[T], middlewares ...ApiMiddleware) {
	subscribeTopic := ConvertUrlToNatsTopic(method, url)
	LogInfo("Register api: %s %s = %s", method, url, subscribeTopic)

	// Check if T is a struct
	tType := reflect.TypeOf((*T)(nil)).Elem()
	if tType.Kind() != reflect.Struct {
		LogFatal("Handler request parameter must be a struct, got: %s", tType.Kind())
	}

	_, err := queueClient.nc.Subscribe(subscribeTopic, func(msg *nats.Msg) {
		// Create a new context
		request := Request{}
		proto.Unmarshal(msg.Data, &request)

		ctx := generateNatsContextFromRequest(&request)
		ctx.msg = msg
		defer putNatsContext(ctx)

		ctx.LogInfo("Request: Url = %s, method = %s, header = %#v", ctx.request.Url, ctx.request.Method, ctx.request.Headers)

		// Append to common middleware
		middlewareList := []ApiMiddleware{}
		middlewareList = append(middlewareList, commonApiMiddlewares...)
		middlewareList = append(middlewareList, middlewares...)

		// Call middleware of function
		for _, middleware := range middlewareList {
			ctx.isEndRequest = true
			if err := middleware(ctx); ctx.isEndRequest {
				if err != nil {
					ctx.writeError(err)
				}
				return
			}
		}

		// Init request
		req := initRequest[T]()

		ctx.LogInfo("Parse json tag to req")
		requestContentType := strings.ToLower(ctx.GetRequestHeader(CONTENT_TYPE_KEY))
		if len(ctx.request.Body) != 0 {
			if strings.Contains(requestContentType, JSON_CONTENT_TYPE) {
				if err := json.Unmarshal(ctx.request.Body, &req); err != nil {
					ctx.LogError("Bind json error: %s", err.Error())
					ctx.writeError(NewHttpError(http.StatusBadRequest, ERROR_BAD_BODY_REQUEST, err.Error(), nil))
					return
				}
			}
		}

		ctx.LogInfo("Validate go struct with tag")
		// Validate go struct with tag
		errValidate := validate.StructCtx(ctx, req)
		if errValidate != nil {
			errMessage := "Request invalid: "
			for _, err := range errValidate.(validator.ValidationErrors) {
				errMessage = fmt.Sprintf("%s {Field: %s, Tag: %s, Value: %s}", errMessage, err.Field(), err.Tag(), err.Value())
			}
			ctx.LogError("Validate go struct with tag error: %s", errMessage)
			ctx.writeError(NewHttpError(http.StatusBadRequest, ERROR_BAD_BODY_REQUEST, errMessage, nil))

			return
		}

		ctx.LogInfo("Call handler")
		// Call handler
		requestBody := strings.ReplaceAll(string(ctx.request.Body), "\r", "")
		requestBody = strings.ReplaceAll(requestBody, "\n", "")

		ctx.LogInfo("Request: Url = %s, method = %s, header = %#v, body = %s", ctx.request.Url, ctx.request.Method, ctx.request.Headers, requestBody)
		res, err := handler(ctx, req)
		if err != nil {
			ctx.LogError("Response error: Url = %s, body = %s", ctx.request.Url, err.Error())
			ctx.writeError(err)
			return
		}

		if res != nil {
			ctx.LogInfo("Response: Url = %s, body = %+v", ctx.request.Url, res.GetBody())
			ctx.writeSuccess(res)
			return
		}

		ctx.LogInfo("Response: Url = %s, body = nil", ctx.request.Url)
		ctx.writeDefaultSuccess()

	})
	if err != nil {
		LogFatal("Fail to subscribe topic: %s, err = %v", url, err)
	}

}

func ConvertUrlToNatsTopic(method string, url string) string {
	url = strings.Trim(url, "/")
	urlTopic := fmt.Sprintf("%s.%s", method, url)
	return strings.ReplaceAll(urlTopic, "/", ".")
}
