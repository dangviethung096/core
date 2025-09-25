package core

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
)

type TestApiInfo[T any] struct {
	URL      string
	Method   string
	Headers  map[string]string
	Queries  map[string]string
	Body     any
	FormData map[string]string
	Handler  Handler[T]
}

func TestAPI[T any](apiInfo TestApiInfo[T]) (HttpResponse, HttpError) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// Create a new context
	ctx := GetHttpContext(c)
	defer PutHttpContext(ctx)

	// Get url
	ctx.URL, _ = url.Parse(apiInfo.URL)

	ctx.Method = apiInfo.Method

	ctx.urlParams = apiInfo.Queries
	var req = apiInfo.Body.(T)

	// Validate go struct with tag
	errValidate := validate.StructCtx(ctx, req)
	if errValidate != nil {
		errMessage := "Request invalid: "
		for _, err := range errValidate.(validator.ValidationErrors) {
			errMessage = fmt.Sprintf("%s {Field: %s, Tag: %s, Value: %s}", errMessage, err.Field(), err.Tag(), err.Value())
		}

		return nil, NewHttpError(http.StatusBadRequest, ERROR_BAD_BODY_REQUEST, errMessage, nil)
	}

	// Call handler
	return apiInfo.Handler(ctx, req)
}

func TestAPIWithContext[T any](ctx *httpContext, apiInfo TestApiInfo[T]) (HttpResponse, HttpError) {
	// Get url
	ctx.URL, _ = url.Parse(apiInfo.URL)

	ctx.Method = apiInfo.Method

	ctx.urlParams = apiInfo.Queries
	var req = apiInfo.Body.(T)

	// Validate go struct with tag
	errValidate := validate.StructCtx(ctx, req)
	if errValidate != nil {
		errMessage := "Request invalid: "
		for _, err := range errValidate.(validator.ValidationErrors) {
			errMessage = fmt.Sprintf("%s {Field: %s, Tag: %s, Value: %s}", errMessage, err.Field(), err.Tag(), err.Value())
		}

		return nil, NewHttpError(http.StatusBadRequest, ERROR_BAD_BODY_REQUEST, errMessage, nil)
	}

	// Call handler
	return apiInfo.Handler(ctx, req)
}
