package core

import "net/http"

type HttpClientResponse interface {
	GetResponseBody() any
	GetHeaders() http.Header
	GetStatusCode() int
	GetRawResponseBody() []byte
}

type httpClientResponse struct {
	responseBody any
	headers      http.Header
	statusCode   int
	rawResponse  []byte
}

func (r *httpClientResponse) GetResponseBody() any {
	return r.responseBody
}

func (r *httpClientResponse) GetHeaders() http.Header {
	return r.headers
}

func (r *httpClientResponse) GetStatusCode() int {
	return r.statusCode
}

func (r *httpClientResponse) GetRawResponseBody() []byte {
	return r.rawResponse
}
