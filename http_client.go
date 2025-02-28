package core

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

type bodyType string

const (
	BodyType_NONE       bodyType = "none"
	BodyType_JSON       bodyType = "json"
	BodyType_XML        bodyType = "xml"
	BodyType_URLEncoded bodyType = "x-www-form-urlencoded"
	BodyType_FORM_DATA  bodyType = "multipart/form-data"
)

type HttpClientCallback func(ctx Context, response HttpClientResponse)

type httpClientBuilder struct {
	ctx Context
	*http.Client
	transport     *http.Transport
	defaultProxy  func(*http.Request) (*url.URL, error)
	url           string
	method        string
	body          any
	headers       map[string][]string
	callback      HttpClientCallback
	bodyType      bodyType
	queries       map[string][]string
	formData      map[string][]string
	retry         bool
	errorResponse any
}

type HttpClientBuilder interface {
	Build() HttpClientBuilder
	SetUrl(url string) HttpClientBuilder
	SetMethod(method string) HttpClientBuilder
	SetBody(body any) HttpClientBuilder
	AddQuery(key string, value string) HttpClientBuilder
	AddHeader(key string, value string) HttpClientBuilder
	SetHeaders(headers map[string][]string) HttpClientBuilder
	AddFormData(key string, value string) HttpClientBuilder
	SetFormData(formData map[string][]string) HttpClientBuilder
	SetMultiPartFormData(contentType string) HttpClientBuilder
	SetCallback(callback HttpClientCallback) HttpClientBuilder
	SetContext(ctx Context) HttpClientBuilder
	SetRetry() HttpClientBuilder
	SetErrorBody(errorResponse any) HttpClientBuilder
	IgnoreTLSCertificate() HttpClientBuilder
	SetProxy(stringProxyUrl string) HttpClientBuilder
	GetContext() Context
	GetUrl() string
	GetMethod() string
	SetBodyType(bodyType bodyType) HttpClientBuilder

	/*
	 * Make a http request and call to server
	 * @param response any Response to set
	 * @return HttpClientResponse, Error
	 */
	Request(response any) (HttpClientResponse, Error)

	/*
	 * Make a http request and call to internal server
	 * @param response any Response to set
	 * @return HttpClientResponse, Error
	 */
	RequestInternal(response any) (HttpClientResponse, Error)

	GetFile(path string) (string, Error)
}

func NewClient() HttpClientBuilder {
	clientBuilder := &httpClientBuilder{
		Client: &http.Client{
			Timeout: HTTP_CLIENT_TIMEOUT,
		},
		bodyType:      BodyType_JSON,
		ctx:           coreContext,
		body:          nil,
		headers:       make(map[string][]string),
		queries:       make(map[string][]string),
		retry:         false,
		errorResponse: nil,
		formData:      nil,
		transport:     &http.Transport{},
	}

	stringProxyUrl := Config.Proxy.GetConfigUrl()
	if proxyUrl, err := url.Parse(stringProxyUrl); stringProxyUrl != BLANK && err == nil {
		clientBuilder.defaultProxy = http.ProxyURL(proxyUrl)
	} else {
		clientBuilder.defaultProxy = http.ProxyURL(nil)
	}

	clientBuilder.transport.Proxy = clientBuilder.defaultProxy
	clientBuilder.Client.Transport = clientBuilder.transport

	return clientBuilder
}

/*
* Create new http client builder
* @return HttpClientBuilder
 */
func (builder *httpClientBuilder) Build() HttpClientBuilder {
	return builder
}

/*
* Set url for http client
* @param url string Url to set
* @return HttpClientBuilder
 */
func (builder *httpClientBuilder) SetUrl(url string) HttpClientBuilder {
	builder.url = url
	return builder
}

/*
* Set method for http client
* @param method string Method to set
* @return HttpClientBuilder
 */
func (builder *httpClientBuilder) SetMethod(method string) HttpClientBuilder {
	builder.method = method
	return builder
}

/*
* Set body for http client
* @param body any Body to set
* @return HttpClientBuilder
 */
func (builder *httpClientBuilder) SetBody(body any) HttpClientBuilder {
	builder.body = body
	return builder
}

/*
* Set header for http client
* @return HttpClientBuilder
 */
func (builder *httpClientBuilder) AddHeader(key string, value string) HttpClientBuilder {
	if builder.headers == nil {
		builder.headers = make(map[string][]string)
	}

	if val, ok := builder.headers[key]; ok {
		builder.headers[key] = append(val, value)
	} else {
		builder.headers[key] = []string{value}
	}

	return builder
}

func (builder *httpClientBuilder) AddFormData(key string, value string) HttpClientBuilder {
	builder.bodyType = BodyType_URLEncoded
	if builder.formData == nil {
		builder.formData = make(map[string][]string)
	}

	if val, ok := builder.formData[key]; ok {
		builder.formData[key] = append(val, value)
	} else {
		builder.formData[key] = []string{value}
	}

	return builder
}

func (builder *httpClientBuilder) SetContext(ctx Context) HttpClientBuilder {
	builder.ctx = ctx
	return builder
}

/*
* Set headers for http client
* It may be will replace old headers that already existed if same key is set
* @param headers map[string]string Headers to set
* @return HttpClientBuilder
 */
func (builder *httpClientBuilder) SetHeaders(headers map[string][]string) HttpClientBuilder {
	// If header is nil, assign map of params to header and return
	if builder.headers == nil {
		builder.headers = headers
		return builder
	}

	for key, value := range headers {
		builder.headers[key] = value
	}
	return builder
}

func (builder *httpClientBuilder) SetFormData(formData map[string][]string) HttpClientBuilder {
	builder.bodyType = BodyType_URLEncoded
	// If header is nil, assign map of params to header and return
	if builder.formData == nil {
		builder.formData = formData
		return builder
	}

	for key, value := range formData {
		builder.formData[key] = value
	}
	return builder
}

func (builder *httpClientBuilder) SetMultiPartFormData(contentType string) HttpClientBuilder {
	builder.bodyType = BodyType_FORM_DATA
	builder.AddHeader(CONTENT_TYPE_KEY, contentType)
	return builder
}

/*
* Set error body: set error reponse, client will save error data struct
* if status code > 399, and errorResponse is a point of struct type
* CAUTION: must pass a pointer of struct in errorResponse param
* @param errorResponse any
* @return HttpClientBuilder
 */
func (builder *httpClientBuilder) SetErrorBody(errorResponse any) HttpClientBuilder {
	builder.errorResponse = errorResponse
	return builder
}

/*
* Make a http request and call to server
* @param response any Response to set
* @return HttpClientResponse, Error
 */
func (builder *httpClientBuilder) Request(response any) (HttpClientResponse, Error) {
	// Reset builder after request
	defer builder.resetBuilder()
	// Run request timeY
	start := time.Now()
	defer func() {
		end := time.Now()
		diff := end.UnixNano() - start.UnixNano()

		builder.ctx.LogInfo("Request time: %fs", time.Duration(diff).Seconds())
	}()

	// Check if response is a pointer of struct
	if err := paramIsPointerOfStruct(response); err != nil {
		builder.ctx.LogError("Response param is not a pointer of struct, error: %v", err)
		return nil, err
	}

	var body *bytes.Buffer
	if builder.body != nil && builder.bodyType == BodyType_JSON {
		// Handle json body
		bodyBytes, err := json.Marshal(builder.body)
		if err != nil {
			builder.ctx.LogError("Cannot marshal body: body = %v, err = %v", builder.body, err)
		}
		builder.ctx.LogInfo("HttpRequest: url = %s, body: %s", builder.url, string(bodyBytes))
		body = bytes.NewBuffer(bodyBytes)
	} else if builder.bodyType == BodyType_URLEncoded {
		// Handle form data
		if builder.formData != nil {
			data := url.Values{}
			for key, values := range builder.formData {
				for _, value := range values {
					if data.Get(key) == BLANK {
						data.Set(key, value)
					} else {
						data.Add(key, value)
					}
				}
			}
			body = bytes.NewBuffer([]byte(data.Encode()))
		}
	} else {
		body = builder.body.(*bytes.Buffer)
	}

	// Init a request
	req, err := http.NewRequest(builder.method, builder.url, body)
	if err != nil {
		builder.ctx.LogError("Cannot create new http request: url = %s, method = %s, err = %v", builder.url, builder.method, err)
		return nil, ERROR_CANNOT_CREATE_HTTP_REQUEST
	}

	// Set headers
	for key, values := range builder.headers {
		for _, value := range values {
			if req.Header.Get(key) == BLANK {
				req.Header.Set(key, value)
			} else {
				req.Header.Add(key, value)
			}
		}
	}
	builder.ctx.LogInfo("HttpRequest: url = %s, headers: %#v", builder.url, builder.headers)

	//Set Form Data
	if builder.formData != nil {
		req.Header.Add(CONTENT_TYPE_KEY, FORMDATA_CONTENT_TYPE)
	}

	// Set query
	if len(builder.queries) > 0 {
		queries := req.URL.Query()
		for key, values := range builder.queries {
			for _, value := range values {
				queries.Add(key, value)
			}

		}
		req.URL.RawQuery = queries.Encode()
	}

	res, errRequest := builder.request(req, response)
	if errRequest != nil && builder.retry {
		// Retry request
		for i := 0; i < Config.HttpClient.RetryTimes; i++ {
			time.Sleep(time.Millisecond * time.Duration(Config.HttpClient.WaitTimes))
			res, errRequest = builder.request(req, response)
			if errRequest == nil {
				break
			}
		}
	}

	return res, errRequest
}

/*
* Set callback for http client: if callback is set, http request will be asynchronous
* Http request will be called in go routine
* @param callback httpClientCallback Callback to set
* @return HttpClientBuilder
 */
func (builder *httpClientBuilder) SetCallback(callback HttpClientCallback) HttpClientBuilder {
	builder.callback = callback
	return builder
}

func (builder *httpClientBuilder) AddQuery(key string, value string) HttpClientBuilder {
	builder.queries[key] = append(builder.queries[key], value)
	return builder
}

func (builder *httpClientBuilder) SetRetry() HttpClientBuilder {
	builder.retry = true
	return builder
}

/*
* Set flag: ignore insecure server.
* Doesn't need verify certificate
 */
func (builder *httpClientBuilder) IgnoreTLSCertificate() HttpClientBuilder {
	builder.transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	return builder
}

/*
* SetProxy: set temporary proxy for http request
* Proxy will be reset to default proxy after http request is called
 */
func (builder *httpClientBuilder) SetProxy(stringProxyUrl string) HttpClientBuilder {
	if proxyUrl, err := url.Parse(stringProxyUrl); stringProxyUrl != BLANK && err == nil {
		builder.transport.Proxy = http.ProxyURL(proxyUrl)
	} else if err != nil {
		builder.ctx.LogError("Set proxy: %s", err.Error())
	}
	return builder
}

func (builder *httpClientBuilder) SetBodyType(bodyType bodyType) HttpClientBuilder {
	builder.bodyType = bodyType
	return builder
}

func (builder *httpClientBuilder) GetUrl() string {
	return builder.url
}

func (builder *httpClientBuilder) GetMethod() string {
	return builder.method
}

func (builder *httpClientBuilder) GetContext() Context {
	return builder.ctx
}

/*
* Reset builder after call
 */
func (builder *httpClientBuilder) resetBuilder() {
	builder.body = nil
	builder.ctx = coreContext
	builder.bodyType = BodyType_JSON
	builder.headers = make(map[string][]string)
	builder.url = BLANK
	builder.method = BLANK
	builder.queries = make(map[string][]string)
	builder.retry = false
	builder.errorResponse = nil
	builder.formData = nil
	builder.transport.TLSClientConfig = nil
	builder.transport.Proxy = builder.defaultProxy
}

func (builder *httpClientBuilder) request(req *http.Request, response any) (HttpClientResponse, Error) {
	builder.ctx.LogInfo("HttpRequest: url = %s%s, method = %s", req.Host, req.RequestURI, builder.method)
	// Send http request
	resp, err := builder.Do(req)
	if err != nil {
		builder.ctx.LogError("Cannot send http request: url = %s, method = %s, err = %s", builder.url, builder.method, err.Error())
		return nil, ERROR_SEND_HTTP_REQUEST_FAIL
	}

	defer resp.Body.Close()

	resVal := &httpClientResponse{
		responseBody: response,
		headers:      resp.Header,
		statusCode:   resp.StatusCode,
	}

	// Read body from buffer
	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		builder.ctx.LogError("Cannot read response body: url = %s, method = %s, err = %v", builder.url, builder.method, err)
		return resVal, ERROR_CANNOT_UNMARSHAL_HTTP_RESPONSE
	}
	resVal.rawResponse = resBody
	builder.ctx.LogInfo("HttpRequest: url = %s, response header: %+v", builder.url, resp.Header)
	builder.ctx.LogInfo("HttpRequest: url = %s response body: %s", builder.url, string(resBody))

	if resp.StatusCode > 399 && builder.errorResponse != nil && paramIsPointerOfStruct(builder.errorResponse) == nil {
		json.Unmarshal(resBody, builder.errorResponse)
	}

	// Read response
	err = json.Unmarshal(resBody, response)
	if err != nil {
		builder.ctx.LogError("Cannot unmarshal response: url = %s, method = %s, err = %v", builder.url, builder.method, err)
		return resVal, ERROR_CANNOT_UNMARSHAL_HTTP_RESPONSE
	}

	resVal.responseBody = response
	// Convert the body
	return resVal, nil
}

func (builder *httpClientBuilder) GetFile(path string) (string, Error) {
	// Reset builder after request
	defer builder.resetBuilder()
	// Run request timeY
	start := time.Now()
	defer func() {
		end := time.Now()
		diff := end.UnixNano() - start.UnixNano()

		builder.ctx.LogInfo("Request time: %fs", time.Duration(diff).Seconds())
	}()

	var body []byte
	if builder.body != nil && builder.bodyType == BodyType_JSON {
		var err error
		body, err = json.Marshal(builder.body)
		if err != nil {
			builder.ctx.LogError("Cannot marshal body: body = %v, err = %v", builder.body, err)
		}
		builder.ctx.LogInfo("HttpRequest: url = %s, body: %s", builder.url, string(body))
	}

	if builder.formData != nil {
		data := url.Values{}
		for key, values := range builder.formData {
			for _, value := range values {
				if data.Get(key) == BLANK {
					data.Set(key, value)
				} else {
					data.Add(key, value)
				}
			}
		}
		body = []byte(data.Encode())
	}

	// Init a request
	req, err := http.NewRequest(builder.method, builder.url, bytes.NewBuffer(body))
	if err != nil {
		builder.ctx.LogError("Cannot create new http request: url = %s, method = %s, err = %v", builder.url, builder.method, err)
		return BLANK, ERROR_CANNOT_CREATE_HTTP_REQUEST
	}

	// Set headers
	for key, values := range builder.headers {
		for _, value := range values {
			if req.Header.Get(key) == BLANK {
				req.Header.Set(key, value)
			} else {
				req.Header.Add(key, value)
			}
		}
	}
	builder.ctx.LogInfo("HttpRequest: url = %s, headers: %#v", builder.url, builder.headers)

	//Set Form Data
	if builder.formData != nil {
		req.Header.Add(CONTENT_TYPE_KEY, FORMDATA_CONTENT_TYPE)
	}

	// Set query
	if len(builder.queries) > 0 {
		queries := req.URL.Query()
		for key, values := range builder.queries {
			for _, value := range values {
				queries.Add(key, value)
			}

		}
		req.URL.RawQuery = queries.Encode()
	}

	resp, err := builder.Do(req)
	if err != nil {
		builder.ctx.LogError("Cannot send http request: url = %s, method = %s, err = %s", builder.url, builder.method, err.Error())
		return BLANK, ERROR_SEND_HTTP_REQUEST_FAIL
	}

	defer resp.Body.Close()

	out, err := os.Create(path)
	if err != nil {
		builder.ctx.LogError("Cannot create file: url = %s, err = %s", path, err.Error())
		return BLANK, ERROR_CANNOT_CREATE_FILE
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		builder.ctx.LogError("Cannot copy file: url = %s, err = %s", path, err.Error())
		return BLANK, ERROR_CANNOT_COPY_FILE
	}

	return path, nil
}

func (builder *httpClientBuilder) RequestInternal(response any) (HttpClientResponse, Error) {
	type internalResponse struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    any    `json:"data"`
	}

	ctx := builder.ctx

	var internalResponseValue internalResponse

	res, err := builder.Request(&internalResponseValue)
	if err != nil {
		return nil, err
	}

	if internalResponseValue.Code != API_CODE_SUCCESS {
		ctx.LogError("Request internal: code = %d, message = %s", internalResponseValue.Code, internalResponseValue.Message)
		return nil, NewError(internalResponseValue.Code, internalResponseValue.Message)
	}

	responseValue := res.(*httpClientResponse)

	newRawResponse, originError := json.Marshal(internalResponseValue.Data)
	if originError != nil {
		ctx.LogError("Cannot marshal data: data = %v, err = %v", internalResponseValue.Data, originError)
		return nil, ERROR_SERVER_ERROR
	}

	originError = json.Unmarshal(newRawResponse, response)
	if originError != nil {
		ctx.LogError("Cannot unmarshal data: data = %v, err = %v", internalResponseValue.Data, originError)
		return nil, ERROR_INVALID_STRUCTURE_FOR_RESPONSE
	}

	responseValue.rawResponse = newRawResponse
	responseValue.responseBody = response

	return responseValue, nil
}
