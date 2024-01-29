package core

import (
	"fmt"
	"net/http"
)

type HttpError interface {
	GetStatusCode() int
	GetCode() int
	GetMessage() string
	GetErrorData() any
	Error() string
}

func NewHttpError(statusCode int, code int, message string, errorData any) HttpError {
	return &httpError{
		statusCode: statusCode,
		code:       code,
		message:    message,
		errorData:  errorData,
	}
}

func NewDefaultHttpError(code int, messsage string) HttpError {
	return &httpError{
		statusCode: http.StatusOK,
		code:       code,
		message:    messsage,
		errorData:  nil,
	}
}

func NewHttpErrorFromError(err Error) HttpError {
	return &httpError{
		statusCode: http.StatusOK,
		code:       err.GetCode(),
		message:    err.GetMessage(),
		errorData:  nil,
	}
}

type httpError struct {
	statusCode int
	code       int
	message    string
	errorData  any
}

func (err httpError) Error() string {
	return fmt.Sprintf("StatusCode: %d, Code: %d, message: %s, error data: %v", err.statusCode, err.code, err.message, err.errorData)
}

func (err httpError) GetCode() int {
	return err.code
}

func (err httpError) GetMessage() string {
	return err.message
}

func (err httpError) GetStatusCode() int {
	return err.statusCode
}

func (err httpError) GetErrorData() any {
	return err.errorData
}

var (
	HTTP_ERROR_READ_BODY_REQUEST_FAIL  = NewHttpError(http.StatusInternalServerError, ERROR_CODE_READ_BODY_REQUEST_FAIL, "Read body request fail", nil)
	HTTP_ERROR_BAD_REQUEST             = NewHttpError(http.StatusBadRequest, ERROR_CODE_READ_BODY_REQUEST_FAIL, "Read body request fail", nil)
	HTTP_ERROR_CLOSE_BODY_REQUEST_FAIL = NewHttpError(http.StatusInternalServerError, ERROR_CODE_CLOSE_BODY_REQUEST_FAIL, "Close body request fail", nil)
)
