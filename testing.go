package core

import (
	"fmt"
	"net/http"

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
	// Create a new context
	ctx := getHttpContext()
	defer putHttpContext(ctx)

	ctx.requestID = ID.GenerateID()
	// Get url
	ctx.URL = apiInfo.URL
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

func TestAPIWithContext[T any](ctx *HttpContext, apiInfo TestApiInfo[T]) (HttpResponse, HttpError) {
	ctx.requestID = ID.GenerateID()
	// Get url
	ctx.URL = apiInfo.URL
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
