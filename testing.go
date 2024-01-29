package core

type TestApiInfo struct {
	URL      string
	Method   string
	Headers  map[string]string
	Queries  map[string]string
	Body     any
	FormData map[string]string
}

type TestApiResponseData struct {
	Code      int    `json:"code"`
	ErrorData any    `json:"errorData,omitempty"`
	Message   string `json:"message,omitempty"`
	Data      any    `json:"data"`
}

func TestAPI(apiInfo TestApiInfo) (HttpClientResponse, Error) {
	client := NewClient().
		SetUrl(apiInfo.URL).
		SetMethod(apiInfo.Method)

	if apiInfo.Headers != nil {
		for key, value := range apiInfo.Headers {
			client.AddHeader(key, value)
		}
	}

	if apiInfo.FormData == nil {
		client.SetBody(apiInfo.Body)
	} else {
		client.AddHeader(CONTENT_TYPE_KEY, FORMDATA_CONTENT_TYPE)
		for key, value := range apiInfo.FormData {
			client.AddFormData(key, value)
		}
	}

	if apiInfo.Queries != nil {
		for key, value := range apiInfo.Queries {
			client.AddQuery(key, value)
		}
	}

	responseData := TestApiResponseData{}
	return client.Request(&responseData)
}
