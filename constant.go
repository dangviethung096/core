package core

const (
	BLANK                  = ""
	CONTENT_TYPE_TEXT      = "text/plain"
	HTTP_CLIENT_TIMEOUT    = 0
	DEFAULT_CONSUMER_TAG   = "default_consumer"
	CONTENT_TYPE_KEY       = "Content-Type"
	ACCEPT_KEY             = "Accept"
	REGEX_URL_PATH_ELEMENT = "[\\w-]+"
)

// Kind of message queue:
// "direct", "fanout", "topic" and "headers".
const (
	MESSAGE_QUEUE_KIND_DIRECT  = "direct"
	MESSAGE_QUEUE_KIND_FANOUT  = "fanout"
	MESSAGE_QUEUE_KIND_TOPIC   = "topic"
	MESSAGE_QUEUE_KIND_HEADERS = "headers"
)

// Error code
const (
	ERROR_CODE_READ_BODY_REQUEST_FAIL  = 100
	ERROR_CODE_CLOSE_BODY_REQUEST_FAIL = 101
	ERROR_BAD_BODY_REQUEST             = 102
	ERROR_FROM_LIBRARY                 = 103
	ERROR_CODE_FROM_DATABASE           = 104
)

// Scheduler
const (
	TASK_DONE = "Done"
	TASK_FAIL = "Fail"
)

const (
	DB_ERROR_NAME_UNIQUE_VIOLATION      = "unique_violation"
	DB_ERROR_NAME_FOREIGN_KEY_VIOLATION = "foreign_key_violation"
	DB_ERROR_NAME_NOT_NULL_VIOLATION    = "not_null_violation"
)
