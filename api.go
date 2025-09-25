package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
)

type Url struct {
	Path   string
	Params []string
}

type ApiMiddleware func(ctx HttpContext) HttpError

type Handler[T any] func(ctx HttpContext, request T) (HttpResponse, HttpError)

/*
* Register api: register api to routeMap
* @param url: url of api
* @param handler: handler of api
* @param middleware: middleware of api
* @return void
 */
func RegisterAPI[T any](url string, method string, handler Handler[T], middlewares ...ApiMiddleware) {
	LogInfo("Register api: %s %s", method, url)

	// Check if T is a struct
	tType := reflect.TypeOf((*T)(nil)).Elem()
	if tType.Kind() != reflect.Struct {
		LogFatal("Handler request parameter must be a struct, got: %s", tType.Kind())
	}
	// Create a new handler
	ginHandler := func(c *gin.Context) {
		// Create a new context
		ctx := GetHttpContext(c)

		defer PutHttpContext(ctx)
		buildContext(ctx)

		ctx.LogInfo("Request: Url = %s, method = %s, header = %#v", ctx.URL, ctx.Method, ctx.request.Header)

		// Append to common middleware
		middlewareList := []ApiMiddleware{}
		middlewareList = append(middlewareList, commonApiMiddlewares...)
		middlewareList = append(middlewareList, middlewares...)

		// Call middleware of function
		for _, middleware := range middlewareList {
			ctx.isRequestEnd = true
			if err := middleware(ctx); ctx.isRequestEnd {
				if err != nil {
					ctx.writeError(err)
				}
				return
			}
		}

		// Init request
		req := initRequest[T]()

		ctx.LogInfo("Parse uri tag to req")
		// Parse uri tag to req
		if err := ctx.ShouldBindUri(&req); err != nil {
			ctx.LogError("Bind uri error: %s", err.Error())
			ctx.writeError(NewHttpError(http.StatusBadRequest, ERROR_BAD_BODY_REQUEST, err.Error(), nil))
			return
		}

		ctx.LogInfo("Parse json tag to req")
		requestContentType := strings.ToLower(ctx.GetRequestHeader(CONTENT_TYPE_KEY))
		if len(ctx.requestBody) != 0 {
			if strings.Contains(requestContentType, JSON_CONTENT_TYPE) {
				if err := json.Unmarshal(ctx.requestBody, &req); err != nil {
					ctx.LogError("Bind json error: %s", err.Error())
					ctx.writeError(NewHttpError(http.StatusBadRequest, ERROR_BAD_BODY_REQUEST, err.Error(), nil))
					return
				}
			} else if strings.Contains(requestContentType, FORM_URLENCODED_CONTENT_TYPE) {
				buffer := bytes.NewBuffer(ctx.requestBody)
				ctx.request.Body = io.NopCloser(buffer)
				ctx.request.ParseForm()
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
		requestBody := strings.ReplaceAll(string(ctx.requestBody), "\r", "")
		requestBody = strings.ReplaceAll(requestBody, "\n", "")

		ctx.LogInfo("Request: Url = %s, method = %s, header = %#v, body = %s", ctx.URL, ctx.Method, ctx.request.Header, requestBody)
		res, err := handler(ctx, req)
		if err != nil {
			ctx.LogError("Response error: Url = %s, body = %s", ctx.URL, err.Error())
			ctx.writeError(err)
			return
		}

		if res != nil {
			ctx.LogInfo("Response: Url = %s, body = %+v", ctx.URL, res.GetBody())
			ctx.writeSuccess(res)
			return
		}

		ctx.LogInfo("Response: Url = %s, body = nil", ctx.URL)
		ctx.writeDefaultSuccess()
	}

	switch method {
	case http.MethodGet:
		router.GET(url, ginHandler)
	case http.MethodPost:
		router.POST(url, ginHandler)
	case http.MethodPut:
		router.PUT(url, ginHandler)
	case http.MethodDelete:
		router.DELETE(url, ginHandler)
	}
}

func initRequest[T any]() T {
	var request T
	ref := reflect.New(reflect.TypeOf(request)).Elem()
	return ref.Interface().(T)
}

func buildContext(ctx *httpContext) HttpError {
	// Assign response writer and request
	ctx.rw = ctx.Writer
	ctx.request = ctx.Request

	// Get url
	ctx.URL = ctx.Request.URL
	ctx.Method = ctx.Request.Method

	bodyData, err := ctx.GetRawData()
	if err != nil {
		LogError("Read request body fail. RequestId: %s, Error: %s", ctx.requestID, err.Error())
		return HTTP_ERROR_READ_BODY_REQUEST_FAIL
	}

	ctx.requestBody = bodyData
	return nil
}
