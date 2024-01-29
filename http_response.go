package core

import "net/http"

type HttpResponse interface {
	GetStatusCode() int
	GetBody() any
	GetReponseCode() int
}

type httpResponse struct {
	statusCode   int
	body         any
	responseCode int
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

func NewDefaultHttpResponse(body any) HttpResponse {
	return &httpResponse{
		statusCode:   http.StatusOK,
		body:         body,
		responseCode: http.StatusOK,
	}
}

func NewHttpResponse(responseCode int, body any) HttpResponse {
	return &httpResponse{
		responseCode: responseCode,
		body:         body,
		statusCode:   http.StatusOK,
	}
}

type responseBody struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data"`
}
