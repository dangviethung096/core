package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

/*
* Context type: which carries deadlines, cancellation signals,
* and other request-scoped values across API boundaries and between processes.
 */
type HttpContext struct {
	context.Context
	URL            string
	Method         string
	requestBody    []byte
	responseBody   responseBody
	isRequestEnd   bool
	request        *http.Request
	rw             http.ResponseWriter
	isResponseEnd  bool
	urlParams      map[string]string
	responseHeader map[string][]string
	cancelFunc     context.CancelFunc
	requestID      string
	timeout        time.Duration
}

/*
* GetContext: Get context from pool
* @return: Context
 */
func getHttpContext() *HttpContext {
	ctx := httpContextPool.Get().(*HttpContext)
	ctx.Context, ctx.cancelFunc = context.WithTimeout(coreContext, contextTimeout)
	ctx.timeout = contextTimeout
	ctx.isResponseEnd = false
	return ctx
}

/*
* PutContext: Put context to pool
* @params: Context
* @return: void
 */
func putHttpContext(ctx *HttpContext) {
	ctx.cancelFunc()
	ctx.urlParams = make(map[string]string)
	ctx.responseHeader = make(map[string][]string)
	httpContextPool.Put(ctx)
}

/*
* Next: Set isRequestEnd to false
* This funciton must to be called when you want to call next middleware
* @return: void
 */
func (ctx *HttpContext) Next() {
	ctx.isRequestEnd = false
}

/*
* GetRequestHeader: Get request header by key
* @params: key string
* @return: string
 */
func (ctx *HttpContext) GetRequestHeader(key string) string {
	return ctx.request.Header.Get(key)
}

/*
* GetQueryParam: Get query param by key
* @params: key string
* @return: string
 */
func (ctx *HttpContext) GetQueryParam(key string) string {
	return ctx.request.URL.Query().Get(key)
}

/*
* ListQueryParam: Get list query param by key
* @params: key string
* @return: []string
 */
func (ctx *HttpContext) GetArrayQueryParam(key string) []string {
	return ctx.request.URL.Query()[key]
}

/*
* GetFormData: get data in body when context/type is application/x-www-form-urlencoded
* in header request
* @return string
* if key exist in form data return value of key, otherwise return empty string
 */
func (ctx *HttpContext) GetFormData(key string) string {
	return ctx.request.PostForm.Get(key)
}

func (ctx *HttpContext) SetResponseHeader(key string, value []string) {
	if value != nil {
		ctx.responseHeader[key] = value
	} else {
		ctx.responseHeader[key] = []string{}
	}
}

func (ctx *HttpContext) AddResponseHeader(key string, value string) {
	if val, ok := ctx.responseHeader[key]; ok && val != nil {
		val = append(val, value)
		ctx.responseHeader[key] = val
	} else {
		ctx.responseHeader[key] = []string{value}
	}
}

func (ctx *HttpContext) AddResponseHeaders(key string, values []string) {
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
func (ctx *HttpContext) GetResponseHeader(key string) []string {
	if val, ok := ctx.responseHeader[key]; ok && val != nil {
		return val
	}
	return []string{}
}

/*
* Redirect url
 */
func (ctx *HttpContext) RedirectURL(url string) {
	ctx.isResponseEnd = true
	http.Redirect(ctx.rw, ctx.request, url, http.StatusSeeOther)
}

/*
* writeError: write error http response to user
 */
func (ctx *HttpContext) writeError(httpErr HttpError) {
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

	ctx.responseBody.Code = httpErr.GetCode()
	ctx.responseBody.Message = httpErr.GetMessage()
	ctx.responseBody.Data = httpErr.GetErrorData()

	body, err := json.Marshal(ctx.responseBody)
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
func (ctx *HttpContext) writeSuccess(httpRes HttpResponse) {
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

	ctx.responseBody.Code = httpRes.GetReponseCode()
	ctx.responseBody.Message = BLANK
	ctx.responseBody.Data = httpRes.GetBody()

	body, err := json.Marshal(ctx.responseBody)
	if err != nil {
		ctx.LogError("Marshal json. RequestId: %s, Error: %s", ctx.requestID, err.Error())
		ctx.endResponse(http.StatusInternalServerError, `{"code":500,"message":"Internal server error(Marshal response data)","errorData":null,"data":null}`)
		return
	}

	ctx.endResponse(int(httpRes.GetStatusCode()), string(body))
}

/*
* endResponse: call write header if it is not called before and write body to writer
 */
func (ctx *HttpContext) endResponse(statusCode int, body string) {
	if !ctx.isResponseEnd {
		ctx.isResponseEnd = true
		// end response
		ctx.rw.WriteHeader(statusCode)
		fmt.Fprint(ctx.rw, body)
	}
}

/*
* endResponse: call write header if it is not called before and write body to writer
 */
func (ctx *HttpContext) GetUrlParam(key string) string {
	return ctx.urlParams[key]
}

func (ctx *HttpContext) convertUrlParams(pattern string, url string, params []string) {
	pattern = pattern[1 : len(pattern)-1]
	patternArray := strings.Split(pattern, "/")
	urlArray := strings.Split(url, "/")
	if len(urlArray) != len(patternArray) {
		LogError("Cannot convert url params: %s, %s", pattern, url)
		return
	}

	count := 0
	for i, patternElement := range patternArray {
		if patternElement == REGEX_URL_PATH_ELEMENT {
			ctx.urlParams[params[count]] = urlArray[i]
			count++
		}
	}
}

/*
* GetContextID: Get the context id
* @params: void
* @return: string
 */
func (ctx *HttpContext) GetContextID() string {
	return ctx.requestID
}

/*
* GetCancelFunc: Get the cancel function
* @params: void
* @return: func()
 */
func (ctx *HttpContext) GetCancelFunc() func() {
	return ctx.cancelFunc
}

/*
* GetTimeout: Get the timeout
* @params: void
 */
func (ctx *HttpContext) GetTimeout() time.Duration {
	return ctx.timeout
}

func (ctx *HttpContext) GetCookie(key string) (*http.Cookie, Error) {
	cookie, err := ctx.request.Cookie(key)
	if err != nil {
		return cookie, NewError(ERROR_FROM_LIBRARY, err.Error())
	}
	return cookie, nil
}

func (ctx *HttpContext) ResetCookie(name string, value string, maxAge int) {
	http.SetCookie(ctx.rw, &http.Cookie{
		Name:   name,
		Value:  value,
		MaxAge: maxAge,
	})
}

/*
* Return a page
 */
