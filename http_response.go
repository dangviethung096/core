package core

import "net/http"

type HttpResponse interface {
	GetStatusCode() int
	GetBody() any
	GetReponseCode() int
	SetResponseContentType(ContentType)
	GetResponseContentType() ContentType
	GetMessage() string
	SetMessage(message string)
}

type ContentType string

const (
	JSON_CONTENT_TYPE            = "application/json"
	FORM_URLENCODED_CONTENT_TYPE = "application/x-www-form-urlencoded"
	TEXT_HTML_CONTENT_TYPE       = "text/html"
	TEXT_PLAIN_CONTENT_TYPE      = "text/plain"
)

type httpResponse struct {
	statusCode   int
	body         any
	responseCode int
	contentType  ContentType
	message      string
}

func (resp *httpResponse) GetStatusCode() int {
	return resp.statusCode
}

func (resp *httpResponse) GetBody() any {
	return resp.body
}

func (resp *httpResponse) GetReponseCode() int {
	return resp.responseCode
}

func (resp *httpResponse) SetResponseContentType(ct ContentType) {
	resp.contentType = ct
}

func (resp *httpResponse) GetResponseContentType() ContentType {
	return resp.contentType
}

func (resp *httpResponse) GetMessage() string {
	return resp.message
}

func (resp *httpResponse) SetMessage(message string) {
	resp.message = message
}

func NewDefaultHttpResponse(body any) HttpResponse {
	return &httpResponse{
		statusCode:   http.StatusOK,
		body:         body,
		responseCode: API_CODE_SUCCESS,
		contentType:  JSON_CONTENT_TYPE,
	}
}

func NewHttpResponse(responseCode int, body any) HttpResponse {
	return &httpResponse{
		responseCode: responseCode,
		body:         body,
		statusCode:   http.StatusOK,
		contentType:  JSON_CONTENT_TYPE,
	}
}

func NewHttpResponseWithMessage(code int, message string, body any) HttpResponse {
	return &httpResponse{
		responseCode: code,
		body:         body,
		message:      message,
		statusCode:   http.StatusOK,
		contentType:  JSON_CONTENT_TYPE,
	}
}

type responseBody struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data"`
}
