package core

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

func generateNatsContextFromRequest(request *Request) *natsContext {
	ctx := natsContextPool.Get().(*natsContext)
	ctx.request = request

	for key, value := range request.Data {
		ctx.data[key] = value
	}

	ctx.response = &Response{
		Headers:      make(map[string]*Header),
		ResponseBody: nil,
		StatusCode:   DEFAULT_INTEGER,
		RedirectLink: BLANK,
	}
	ctx.Context, ctx.cancelFunc = context.WithTimeout(coreContext, ctx.timeout)

	return ctx
}

func putNatsContext(ctx *natsContext) {
	for k := range ctx.data {
		delete(ctx.data, k)
	}
	ctx.request = nil
	ctx.response = nil
	ctx.msg = nil
	ctx.cancelFunc()
	natsContextPool.Put(ctx)
}

type natsContext struct {
	context.Context
	request  *Request
	response *Response
	msg      *nats.Msg

	timeout    time.Duration
	cancelFunc func()

	data         map[string]any
	isEndRequest bool
}

func (ctx *natsContext) GetContextID() string {
	return ctx.request.ContextId
}

func (ctx *natsContext) GetTimeout() time.Duration {
	return ctx.timeout
}

func (ctx *natsContext) GetCancelFunc() func() {
	return ctx.cancelFunc
}

func (ctx *natsContext) Next() {
	ctx.isEndRequest = false
}

func (ctx *natsContext) GetRequestHeader(key string) string {
	val, ok := ctx.request.Headers[key]
	if !ok || len(val.Value) == 0 {
		return BLANK
	}
	return val.Value[0]
}

func (ctx *natsContext) GetQueryParam(key string) string {
	val, ok := ctx.request.QueryParams[key]
	if !ok || len(val.Value) == 0 {
		return BLANK
	}
	return val.Value[0]
}

func (ctx *natsContext) GetArrayQueryParam(key string) []string {
	val, ok := ctx.request.QueryParams[key]
	if !ok || len(val.Value) == 0 {
		return []string{}
	}
	return val.Value
}

func (ctx *natsContext) GetFormData(key string) string {
	val, ok := ctx.request.FormData[key]
	if !ok || len(val.Value) == 0 {
		return BLANK
	}
	return val.Value[0]
}

func (ctx *natsContext) SetResponseHeader(key string, value []string) {
	ctx.response.Headers[key] = &Header{
		Key:   key,
		Value: value,
	}
}

func (ctx *natsContext) AddResponseHeader(key string, value string) {
	val, ok := ctx.response.Headers[key]
	if !ok {
		ctx.response.Headers[key] = &Header{
			Key:   key,
			Value: []string{value},
		}
	} else {
		ctx.response.Headers[key] = val
		val.Value = append(val.Value, value)
	}
}

func (ctx *natsContext) AddResponseHeaders(key string, values []string) {
	val, ok := ctx.response.Headers[key]
	if !ok {
		ctx.response.Headers[key] = &Header{
			Key:   key,
			Value: values,
		}
	} else {
		val.Value = append(val.Value, values...)
		ctx.response.Headers[key] = val
	}
}

func (ctx *natsContext) GetResponseHeader(key string) []string {
	if val, ok := ctx.response.Headers[key]; ok && val != nil {
		return val.Value
	}
	return []string{}
}
func (ctx *natsContext) RedirectURL(url string) {
	ctx.response.RedirectLink = url
	ctx.isEndRequest = true
}

func (ctx *natsContext) GetUrlParam(key string) string {
	val, ok := ctx.request.UrlParams[key]
	if !ok {
		return BLANK
	}
	return val
}

func (ctx *natsContext) GetCookie(key string) (*http.Cookie, Error) {
	val, ok := ctx.request.Cookies[key]
	if !ok {
		return nil, NewError(ERROR_CODE_FROM_COOKIE, "Cookie not found")
	}

	return &http.Cookie{
		Name:        val.Name,
		Value:       val.Value,
		Quoted:      val.Quoted,
		Path:        val.Path,
		Domain:      val.Domain,
		Expires:     time.Unix(val.Expires, 0),
		RawExpires:  val.RawExpires,
		MaxAge:      int(val.MaxAge),
		Secure:      val.Secure,
		HttpOnly:    val.HttpOnly,
		SameSite:    http.SameSite(val.SameSite),
		Partitioned: val.Partitioned,
		Raw:         val.Raw,
		Unparsed:    val.Unparsed,
	}, nil
}

func (ctx *natsContext) ResetCookie(name string) {
	cookie := &http.Cookie{
		Name:   name,
		Value:  BLANK,
		MaxAge: -1,
		Path:   "/",
	}
	val, ok := ctx.response.Headers["Set-Cookie"]
	if val == nil || !ok {
		val = &Header{
			Key:   "Set-Cookie",
			Value: []string{cookie.String()},
		}
	} else {
		val.Value = append(val.Value, cookie.String())
	}

	ctx.response.Headers["Set-Cookie"] = val
}

func (ctx *natsContext) SetCookie(key string, value string, maxAge int) {
	cookie := &http.Cookie{
		Name:     key,
		Value:    value,
		MaxAge:   maxAge,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
	}

	val, ok := ctx.response.Headers["Set-Cookie"]
	if val == nil || !ok {
		val = &Header{
			Key:   "Set-Cookie",
			Value: []string{cookie.String()},
		}
	} else {
		val.Value = append(val.Value, cookie.String())
	}

	ctx.response.Headers["Set-Cookie"] = val
}

func (ctx *natsContext) Query(key string) string {
	if val, ok := ctx.request.QueryParams[key]; val != nil && ok && len(val.Value) > 0 {
		return val.Value[0]
	}
	return BLANK
}

func (ctx *natsContext) GetHeader(key string) string {
	if val, ok := ctx.request.Headers[key]; val != nil && ok && len(val.Value) > 0 {
		return val.Value[0]
	}
	return BLANK
}

func (ctx *natsContext) SetTempData(key string, value any) {
	ctx.data[key] = value
}

func (ctx *natsContext) GetTempData(key string) any {
	val, ok := ctx.data[key]
	if !ok {
		return nil
	}
	return val
}

func (ctx *natsContext) EndResponse(statusCode int, header *http.Header, body []byte) {
	ctx.response.StatusCode = int32(statusCode)
	ctx.response.ResponseBody = body
	ctx.isEndRequest = true

	data, err := proto.Marshal(ctx.response)
	if err != nil {
		ctx.LogError("Marshal error proto. RequestId: %s, Error: %s", ctx.GetContextID(), err.Error())
		ctx.msg.Respond(ERROR_NATS_INTERNAL_SERVER)
		return
	}

	ctx.msg.Respond(data)
}

func (ctx *natsContext) endResponse(statusCode int, body []byte) {
	ctx.response.StatusCode = int32(statusCode)
	ctx.response.ResponseBody = body
	data, err := proto.Marshal(ctx.response)
	if err != nil {
		ctx.LogError("Marshal error proto. RequestId: %s, Error: %s", ctx.GetContextID(), err.Error())
		ctx.msg.Respond(ERROR_NATS_INTERNAL_SERVER)
		return
	}

	ctx.isEndRequest = true

	ctx.msg.Respond(data)
}

func (ctx *natsContext) writeError(httpErr HttpError) {
	resBody := responseBody{
		Code:    httpErr.GetCode(),
		Message: httpErr.GetMessage() + " (RequestID: " + ctx.GetContextID() + ")",
		Data:    httpErr.GetErrorData(),
	}

	body, err := json.Marshal(resBody)
	if err != nil {
		ctx.LogError("Marshal error json. RequestId: %s, Error: %s", ctx.GetContextID(), err.Error())
		ctx.endResponse(http.StatusInternalServerError, ERROR_NATS_INTERNAL_SERVER)
		return
	}

	ctx.endResponse(httpErr.GetStatusCode(), body)
}

func (ctx *natsContext) writeSuccess(httpRes HttpResponse) {
	ctx.response.StatusCode = int32(httpRes.GetStatusCode())

	var body []byte

	if httpRes.GetResponseContentType() == TEXT_PLAIN_CONTENT_TYPE {
		// Set text plain response
		ctx.response.Headers["Content-Type"] = &Header{
			Key:   "Content-Type",
			Value: []string{TEXT_PLAIN_CONTENT_TYPE},
		}
		body = []byte(httpRes.GetBody().(string))
	} else {
		ctx.response.Headers["Content-Type"] = &Header{
			Key:   "Content-Type",
			Value: []string{JSON_CONTENT_TYPE},
		}

		resBody := responseBody{
			Code:    httpRes.GetReponseCode(),
			Message: httpRes.GetMessage(),
			Data:    httpRes.GetBody(),
		}

		var errOrigin error
		body, errOrigin = json.Marshal(resBody)
		if errOrigin != nil {
			ctx.LogError("Marshal json. RequestId: %s, Error: %v", ctx.GetContextID(), errOrigin)
			ctx.endResponse(http.StatusInternalServerError, ERROR_NATS_INTERNAL_SERVER)
			return
		}
	}

	ctx.endResponse(httpRes.GetStatusCode(), body)
}

func (ctx *natsContext) writeDefaultSuccess() {
	ctx.isEndRequest = true

	resBody := responseBody{
		Code:    DEFAULT_INTEGER,
		Message: "Success",
		Data:    nil,
	}

	body, errOrigin := json.Marshal(resBody)
	if errOrigin != nil {
		ctx.LogError("Marshal json. RequestId: %s, Error: %v", ctx.GetContextID(), errOrigin)
		ctx.endResponse(http.StatusInternalServerError, ERROR_NATS_INTERNAL_SERVER)
		return
	}

	ctx.endResponse(http.StatusOK, body)
}

func (ctx *natsContext) GetRawRequest() []byte {
	return ctx.request.Body
}
