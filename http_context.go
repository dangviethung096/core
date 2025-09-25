package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
)

type HttpContext interface {
	Context
	Next()
	GetRequestHeader(key string) string
	GetQueryParam(key string) string
	GetArrayQueryParam(key string) []string
	GetFormData(key string) string
	SetResponseHeader(key string, value []string)
	AddResponseHeader(key string, value string)
	AddResponseHeaders(key string, values []string)
	GetResponseHeader(key string) []string
	RedirectURL(url string)
	GetCancelFunc() func()
	GetUrlParam(key string) string
	GetContextID() string
	GetTimeout() time.Duration
	GetCookie(key string) (*http.Cookie, Error)
	ResetCookie(name string)
	SetCookie(key string, value string, maxAge int)
	Query(key string) string
	GetHeader(key string) string
	SetTempData(key string, value any)
	GetTempData(key string) any
	EndResponse(statusCode int, header *http.Header, body []byte)
}

/*
* Context type: which carries deadlines, cancellation signals,
* and other request-scoped values across API boundaries and between processes.
 */
type httpContext struct {
	*gin.Context
	URL            *url.URL
	Method         string
	requestBody    []byte
	isRequestEnd   bool
	request        *http.Request
	rw             http.ResponseWriter
	isResponseEnd  bool
	urlParams      map[string]string
	responseHeader map[string][]string
	requestID      string
	timeout        time.Duration
	tempData       map[string]any
	cancelFunc     context.CancelFunc
}

/*
* GetContext: Get context from pool
* @return: Context
 */
func GetHttpContext(c *gin.Context) *httpContext {
	ctx := httpContextPool.Get().(*httpContext)
	ctx.Context = c
	ctx.timeout = contextTimeout
	ctx.isResponseEnd = false
	// Clear the maps instead of creating new ones
	for k := range ctx.responseHeader {
		delete(ctx.responseHeader, k)
	}
	for k := range ctx.urlParams {
		delete(ctx.urlParams, k)
	}
	for k := range ctx.tempData {
		delete(ctx.tempData, k)
	}
	ctx.requestID = ID.GenerateID()
	ctx.cancelFunc = func() {
		ctx.LogInfo("Cancel context")
	}
	return ctx
}

func NewHttpContext() *httpContext {
	ctx := httpContextPool.Get().(*httpContext)
	ctx.Context = nil
	ctx.timeout = contextTimeout
	ctx.isResponseEnd = false
	// Clear the maps instead of creating new ones
	for k := range ctx.responseHeader {
		delete(ctx.responseHeader, k)
	}
	for k := range ctx.urlParams {
		delete(ctx.urlParams, k)
	}
	for k := range ctx.tempData {
		delete(ctx.tempData, k)
	}
	ctx.requestID = ID.GenerateID()
	ctx.cancelFunc = func() {
		ctx.LogInfo("Cancel context")
	}
	return ctx
}

/*
* PutContext: Put context to pool
* @params: Context
* @return: void
 */
func PutHttpContext(ctx HttpContext) {
	if ctx == nil {
		return
	}

	ctxValue := ctx.(*httpContext)
	// Release memory of context: urlParams, responseHeader, tempData
	ctxValue.urlParams = nil
	ctxValue.responseHeader = nil
	ctxValue.tempData = nil
	ctxValue.cancelFunc = nil
	ctxValue.Context = nil
	// Put context to pool
	httpContextPool.Put(ctxValue)
}

/*
* Next: Set isRequestEnd to false
* This funciton must to be called when you want to call next middleware
* @return: void
 */
func (ctx *httpContext) Next() {
	ctx.isRequestEnd = false
}

/*
* GetRequestHeader: Get request header by key
* @params: key string
* @return: string
 */
func (ctx *httpContext) GetRequestHeader(key string) string {
	return ctx.request.Header.Get(key)
}

/*
* GetQueryParam: Get query param by key
* @params: key string
* @return: string
 */
func (ctx *httpContext) GetQueryParam(key string) string {
	return ctx.request.URL.Query().Get(key)
}

/*
* ListQueryParam: Get list query param by key
* @params: key string
* @return: []string
 */
func (ctx *httpContext) GetArrayQueryParam(key string) []string {
	return ctx.request.URL.Query()[key]
}

/*
* GetFormData: get data in body when context/type is application/x-www-form-urlencoded
* in header request
* @return string
* if key exist in form data return value of key, otherwise return empty string
 */
func (ctx *httpContext) GetFormData(key string) string {
	return ctx.request.PostForm.Get(key)
}

func (ctx *httpContext) SetResponseHeader(key string, value []string) {
	if value != nil {
		ctx.responseHeader[key] = value
	} else {
		ctx.responseHeader[key] = []string{}
	}
}

func (ctx *httpContext) AddResponseHeader(key string, value string) {
	if val, ok := ctx.responseHeader[key]; ok && val != nil {
		val = append(val, value)
		ctx.responseHeader[key] = val
	} else {
		ctx.responseHeader[key] = []string{value}
	}
}

func (ctx *httpContext) AddResponseHeaders(key string, values []string) {
	if val, ok := ctx.responseHeader[key]; ok && val != nil {
		val = append(val, values...)
		ctx.responseHeader[key] = val
	} else {
		if values != nil {
			ctx.responseHeader[key] = values
		} else {
			ctx.responseHeader[key] = []string{}
		}
	}
}

/*
* GetResponseHeader
 */
func (ctx *httpContext) GetResponseHeader(key string) []string {
	if val, ok := ctx.responseHeader[key]; ok && val != nil {
		return val
	}
	return []string{}
}

/*
* Redirect url
 */
func (ctx *httpContext) RedirectURL(url string) {
	ctx.isResponseEnd = true
	http.Redirect(ctx.rw, ctx.request, url, http.StatusSeeOther)
}

/*
* GetCancelFunc: Get the cancel function
* @params: void
* @return: func()
 */
func (ctx *httpContext) GetCancelFunc() func() {
	return ctx.cancelFunc
}

/*
* writeError: write error http response to user
 */
func (ctx *httpContext) writeError(httpErr HttpError) {
	ctx.rw.Header().Set("Content-Type", "application/json")
	ctx.rw.Header().Set("Request-Id", ctx.requestID)
	for key, values := range ctx.responseHeader {
		headerValue := BLANK
		for i, value := range values {
			if i == 0 {
				headerValue = value
			} else {
				headerValue += "," + value
			}
		}

		ctx.rw.Header().Set(key, headerValue)
	}

	resBody := responseBody{
		Code:    httpErr.GetCode(),
		Message: httpErr.GetMessage() + " (RequestID: " + ctx.requestID + ")",
		Data:    httpErr.GetErrorData(),
	}

	body, err := json.Marshal(resBody)
	if err != nil {
		ctx.LogError("Marshal error json. RequestId: %s, Error: %s", ctx.requestID, err.Error())
		ctx.endResponse(http.StatusInternalServerError, `{"code":500,"message":"Internal server error(Marshal error response data)","errorData":null,"data":null}`)
		return
	}

	ctx.endResponse(int(httpErr.GetStatusCode()), string(body))
}

/*
* writeSuccess: write success http response to user
 */
func (ctx *httpContext) writeSuccess(httpRes HttpResponse) {
	ctx.rw.Header().Set("Request-Id", ctx.requestID)

	for key, values := range ctx.responseHeader {
		headerValue := BLANK
		for i, value := range values {
			if i == 0 {
				headerValue = value
			} else {
				headerValue += "," + value
			}
		}

		ctx.rw.Header().Set(key, headerValue)
	}

	var body []byte

	if httpRes.GetResponseContentType() == TEXT_PLAIN_CONTENT_TYPE {
		// Set text plain response
		ctx.rw.Header().Set("Content-Type", TEXT_PLAIN_CONTENT_TYPE)
		body = []byte(httpRes.GetBody().(string))
	} else {
		ctx.rw.Header().Set("Content-Type", JSON_CONTENT_TYPE)

		resBody := responseBody{
			Code:    httpRes.GetReponseCode(),
			Message: httpRes.GetMessage(),
			Data:    httpRes.GetBody(),
		}

		var errOrigin error
		body, errOrigin = json.Marshal(resBody)
		if errOrigin != nil {
			ctx.LogError("Marshal json. RequestId: %s, Error: %v", ctx.requestID, errOrigin)
			ctx.endResponse(http.StatusInternalServerError, `{"code":500,"message":"Internal server error(Marshal response data)","errorData":null,"data":null}`)
			return
		}

	}

	ctx.endResponse(int(httpRes.GetStatusCode()), string(body))
}

func (ctx *httpContext) writeDefaultSuccess() {
	ctx.rw.Header().Set("Request-Id", ctx.requestID)

	for key, values := range ctx.responseHeader {
		headerValue := BLANK
		for i, value := range values {
			if i == 0 {
				headerValue = value
			} else {
				headerValue += "," + value
			}
		}

		ctx.rw.Header().Set(key, headerValue)
	}

	ctx.rw.Header().Set("Content-Type", JSON_CONTENT_TYPE)

	resBody := responseBody{
		Code:    DEFAULT_INTEGER,
		Message: "Success",
		Data:    nil,
	}

	body, errOrigin := json.Marshal(resBody)
	if errOrigin != nil {
		ctx.LogError("Marshal json. RequestId: %s, Error: %v", ctx.requestID, errOrigin)
		ctx.endResponse(http.StatusInternalServerError, `{"code":500,"message":"Internal server error(Marshal response data)","errorData":null,"data":null}`)
		return
	}

	ctx.endResponse(http.StatusOK, string(body))
}

/*
* endResponse: call write header if it is not called before and write body to writer
 */
func (ctx *httpContext) endResponse(statusCode int, body string) {
	if !ctx.isResponseEnd {
		ctx.isResponseEnd = true
		// end response
		ctx.rw.WriteHeader(statusCode)
		ctx.Writer.Write([]byte(body))
		ctx.Writer.(http.Flusher).Flush()
	}
}

/*
* endResponse: call write header if it is not called before and write body to writer
 */
func (ctx *httpContext) GetUrlParam(key string) string {
	return ctx.urlParams[key]
}

/*
* GetContextID: Get the context id
* @params: void
* @return: string
 */
func (ctx *httpContext) GetContextID() string {
	return ctx.requestID
}

/*
* GetTimeout: Get the timeout
* @params: void
 */
func (ctx *httpContext) GetTimeout() time.Duration {
	return ctx.timeout
}

/*
* GetCookie: Get cookie by key
* @params: key string
* @return: *http.Cookie, Error
* if key exist in cookie return cookie, otherwise return error
* Error: ERROR_FROM_LIBRARY
 */
func (ctx *httpContext) GetCookie(key string) (*http.Cookie, Error) {
	cookie, err := ctx.request.Cookie(key)
	if err != nil {
		return cookie, NewError(ERROR_FROM_LIBRARY, err.Error())
	}
	return cookie, nil
}

/*
* ResetCookie: Reset cookie by name, value and maxAge
* @params: name string, value string, maxAge int
* @return: void
 */
func (ctx *httpContext) ResetCookie(name string) {
	http.SetCookie(ctx.rw, &http.Cookie{
		Name:   name,
		Value:  BLANK,
		MaxAge: -1,
		Path:   "/",
	})
}

/*
* SetCookie: Set cookie by key, value and maxAge
* @params: key string, value string, maxAge int
* @return: void
 */
func (ctx *httpContext) SetCookie(key string, value string, maxAge int) {
	http.SetCookie(ctx.rw, &http.Cookie{
		Name:     key,
		Value:    value,
		MaxAge:   maxAge,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
	})
}

/*
* SetTempData: Set temp data
* @params: key string, value any
* @return: void
 */
func (ctx *httpContext) SetTempData(key string, value any) {
	if ctx.tempData == nil {
		ctx.tempData = make(map[string]any)
	}

	ctx.tempData[key] = value
}

/*
* GetTempData: Get temp data
* @params: key string
* @return: any
 */
func (ctx *httpContext) GetTempData(key string) any {
	return ctx.tempData[key]
}

/*
* EndResponse: End response
* @params: statusCode int, header *http.Header, body []byte
* @return: void
 */
func (ctx *httpContext) EndResponse(statusCode int, header *http.Header, body []byte) {
	if !ctx.isResponseEnd {
		ctx.isResponseEnd = true

		if header != nil {
			for key, values := range *header {
				for _, value := range values {
					ctx.rw.Header().Set(key, value)
				}
			}
		}

		ctx.rw.WriteHeader(statusCode)
		if body != nil {
			fmt.Fprint(ctx.rw, string(body))
		}
		ctx.rw.(http.Flusher).Flush()
	}
}

func (ctx *httpContext) Query(key string) string {
	return ctx.Context.Query(key)
}

func (ctx *httpContext) GetHeader(key string) string {
	return ctx.Context.GetHeader(key)
}
