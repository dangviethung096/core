package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator"
)

type optionalParams struct {
	haveUrlParam bool
	urlPattern   string
	urlParamKeys []string
}

type Route struct {
	URL     Url
	Method  string
	handler func(writer http.ResponseWriter, request *http.Request, optional optionalParams)
}

type Url struct {
	Path   string
	Params []string
}

type ApiMiddleware func(ctx *HttpContext) HttpError

type Handler[T any] func(ctx *HttpContext, request T) (HttpResponse, HttpError)

var urlRegex = regexp.MustCompile(`.*[{].*[}].*`)

/*
* Register api: register api to routeMap
* @param url: url of api
* @param handler: handler of api
* @param middleware: middleware of api
* @return void
 */
func RegisterAPI[T any](url string, method string, handler Handler[T], middlewares ...ApiMiddleware) {
	var isRegexPath = false
	var urlParams []string
	if urlRegex.MatchString(url) {
		url, urlParams = convertRegexUrl(url)
		isRegexPath = true
	}
	coreContext.LogInfo("Register api: %s %s", method, url)

	// Check if T is a struct
	tType := reflect.TypeOf((*T)(nil)).Elem()
	if tType.Kind() != reflect.Struct {
		LogFatal("Handler request parameter must be a struct, got: %s", tType.Kind())
	}
	// Create a new handler
	h := func(writer http.ResponseWriter, request *http.Request, optional optionalParams) {
		// Create a new context
		ctx := getHttpContext()
		defer putHttpContext(ctx)
		buildContext(ctx, writer, request)

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

		if optional.haveUrlParam {
			// convert param
			ctx.convertUrlParams(optional.urlPattern, request.URL.Path, optional.urlParamKeys)
		}

		// Unmarshal json request body to model T
		req := initRequest[T]()
		requestContentType := strings.ToLower(ctx.GetRequestHeader(CONTENT_TYPE_KEY))
		if len(ctx.requestBody) != 0 {
			if strings.Contains(requestContentType, JSON_CONTENT_TYPE) {
				if err := json.Unmarshal(ctx.requestBody, &req); err != nil {
					coreContext.LogInfo("Unmarshal request body fail. RequestId: %s, Error: %s", ctx.requestID, err.Error())
					ctx.writeError(NewDefaultHttpError(400, "Bad request (Marshal requeset body)"))
					return
				}
			} else if strings.Contains(requestContentType, FORMDATA_CONTENT_TYPE) {
				buffer := bytes.NewBuffer(ctx.requestBody)
				ctx.request.Body = io.NopCloser(buffer)
				ctx.request.ParseForm()
			}
		}

		// Validate go struct with tag
		errValidate := validate.StructCtx(ctx, req)
		if errValidate != nil {
			errMessage := "Request invalid: "
			for _, err := range errValidate.(validator.ValidationErrors) {
				errMessage = fmt.Sprintf("%s {Field: %s, Tag: %s, Value: %s}", errMessage, err.Field(), err.Tag(), err.Value())
			}
			ctx.writeError(NewHttpError(http.StatusBadRequest, ERROR_BAD_BODY_REQUEST, errMessage, nil))
			return
		}

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
		}
	}

	if !isRegexPath {
		routeSlice, ok := routeMap[url]
		if ok {
			routeSlice = append(routeSlice, Route{
				Method: method,
				URL: Url{
					Path:   url,
					Params: nil,
				},
				handler: h,
			})
			routeMap[url] = routeSlice
		} else {
			routeMap[url] = []Route{
				{
					Method: method,
					URL: Url{
						Path:   url,
						Params: nil,
					},
					handler: h,
				},
			}
		}
	} else {
		routeSlice, ok := routeRegexMap[url]
		if ok {
			routeSlice = append(routeSlice, Route{
				Method: method,
				URL: Url{
					Path:   url,
					Params: urlParams,
				},

				handler: h,
			})
			routeRegexMap[url] = routeSlice
		} else {
			routeRegexMap[url] = []Route{
				{
					Method: method,
					URL: Url{
						Path:   url,
						Params: urlParams,
					},
					handler: h,
				},
			}
		}
	}
}

func initRequest[T any]() T {
	var request T
	ref := reflect.New(reflect.TypeOf(request)).Elem()
	return ref.Interface().(T)
}

func buildContext(ctx *HttpContext, writer http.ResponseWriter, request *http.Request) HttpError {
	// Assign response writer and request
	ctx.rw = writer
	ctx.request = request

	ctx.requestID = ID.GenerateID()
	// Get url
	ctx.URL = request.URL.Path
	ctx.Method = request.Method

	// Get request body
	buffer := bytes.NewBuffer(ctx.requestBody)
	buffer.Reset()

	if _, err := io.Copy(buffer, request.Body); err != nil {
		LogError("Read request body fail. RequestId: %s, Error: %s", ctx.requestID, err.Error())
		return HTTP_ERROR_READ_BODY_REQUEST_FAIL
	}

	if err := request.Body.Close(); err != nil {
		LogError("Close request body fail. RequestId: %s, Error: %s", ctx.requestID, err.Error())
		return HTTP_ERROR_CLOSE_BODY_REQUEST_FAIL
	}

	ctx.requestBody = buffer.Bytes()
	return nil
}

func convertRegexUrl(url string) (string, []string) {
	params := make([]string, 0)
	param := BLANK
	// Convert url
	newUrl := "^"
	start := false
	for _, letter := range url {
		if letter == '{' && !start {
			newUrl += REGEX_URL_PATH_ELEMENT
			start = !start
			continue
		} else if letter == '}' && start {
			start = !start
			params = append(params, param)
			param = BLANK
			continue
		}

		if !start {
			newUrl += string(letter)
		} else {
			param += string(letter)
		}
	}
	return newUrl + "$", params
}
