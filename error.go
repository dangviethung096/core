package core

import (
	"fmt"
)

type Error interface {
	GetCode() int
	GetMessage() string
	Error() string
	Equal(err Error) bool
}

func NewError(code int, message string) Error {
	return &coreError{
		code:    code,
		message: message,
	}
}

type coreError struct {
	code    int
	message string
}

func (err coreError) Error() string {
	return fmt.Sprintf("Code: %d, message: %s", err.code, err.message)
}

func (err coreError) GetCode() int {
	return err.code
}

func (err coreError) GetMessage() string {
	return err.message
}

func (err coreError) Equal(anotherError Error) bool {
	return err.code == anotherError.GetCode() && err.message == anotherError.GetMessage()
}

var (
	ERROR_SERVER_ERROR                          Error = NewError(1, "Internal Server Error")
	ERROR_BAD_REQUEST                           Error = NewError(2, "Bad request")
	ERROR_MODEL_IS_NOT_STRUCT                   Error = NewError(3, "Model is not a struct")
	ERROR_PARAM_IS_NOT_A_POINTER_OF_STRUCT      Error = NewError(4, "Param is not pontier of struct")
	ERROR_MODEL_HAVE_NO_FIELD                   Error = NewError(5, "Model have no field")
	ERROR_NOT_FOUND_PRIMARY_KEY                 Error = NewError(6, "Not found primary key")
	ERROR_NOT_FOUND_IN_DB                       Error = NewError(7, "Not found in database")
	ERROR_CANNOT_CREATE_RABBITMQ_CHANNEL        Error = NewError(8, "Cannot create RabbitMQ channel")
	ERROR_CANNOT_DECLARE_EXCHANGE               Error = NewError(9, "Cannot declare exchange")
	ERROR_CANNOT_DECLARE_QUEUE                  Error = NewError(10, "Cannot declare queue")
	ERROR_CANNOT_BIND_QUEUE                     Error = NewError(11, "Cannot bind queue")
	ERROR_CANNOT_CONNECT_RABBITMQ               Error = NewError(12, "Cannot connect to RabbitMQ")
	ERROR_CANNOT_PUBLISH_MESSAGE_TO_RABBITMQ    Error = NewError(13, "Cannot publish message to RabbitMQ")
	ERROR_CANNOT_CONSUME_MESSAGES_FROM_RABBITMQ Error = NewError(14, "Cannot consume messages from RabbitMQ")
	ERROR_CANNOT_CREATE_HTTP_REQUEST            Error = NewError(15, "Cannot create http request")
	ERROR_SEND_HTTP_REQUEST_FAIL                Error = NewError(16, "Send http request fail")
	ERROR_CANNOT_UNMARSHAL_HTTP_RESPONSE        Error = NewError(17, "Cannot unmarshal http response")
	ERROR_NIL_PARAM                             Error = NewError(18, "Nil param")
	ERROR_INSERT_TO_DB_FAIL                     Error = NewError(19, "Insert to database fail")
	ERROR_ADD_TASK_SYSTEM_FAIL                  Error = NewError(20, "Add task to system fail")
	ERROR_TASK_TIME_LESS_THAN_NOW               Error = NewError(21, "Scheduler time is less than now")
	ERROR_TASK_INTERVAL_INVALID                 Error = NewError(22, "Task interval is invalid")
	ERROR_TASK_REQUEST_INVALID                  Error = NewError(23, "Task request is invalid")
	ERROR_STOP_TASK_FAIL                        Error = NewError(24, "Stop task fail")
	ERROR_CREATE_MESSAGE_QUEUE_SESSION_FAIL     Error = NewError(25, "Create message queue session fail")
	ERROR_TASK_ALREADY_EXISTED                  Error = NewError(26, "Task has already existed")
	ERROR_REMOVE_OLD_TASK_FAIL                  Error = NewError(27, "Remove old task failed!")
	ERROR_TASK_IS_EXPIRED                       Error = NewError(28, "Task is expired!")
	ERROR_DB_ERROR                              Error = NewError(29, "Database error")
	ERROR_CANNOT_CREATE_FILE                    Error = NewError(30, "Cannot create file")
	ERROR_CANNOT_COPY_FILE                      Error = NewError(31, "Cannot copy file")
	ERROR_NO_SUPPORT_FOR_MANY_PRIMARY_KEYS      Error = NewError(33, "No support for many primary keys")
	ERROR_DB_UNIQUE_VIOLATION                   Error = NewError(34, "Unique violation")
	ERROR_DB_FOREIGN_KEY_VIOLATION              Error = NewError(35, "Foreign key violation")
	ERROR_CANNOT_SUBSCRIBE_QUEUE                Error = NewError(36, "Cannot subscribe queue")
	ERROR_CANNOT_PUBLISH_MESSAGE                Error = NewError(37, "Cannot publish message")
	ERROR_WHERE_QUERY_IS_EMPTY                  Error = NewError(38, "Where query is empty")
)
