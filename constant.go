package core

const (
	BLANK                  = ""
	CONTENT_TYPE_TEXT      = "text/plain"
	HTTP_CLIENT_TIMEOUT    = 0
	DEFAULT_CONSUMER_TAG   = "default_consumer"
	CONTENT_TYPE_KEY       = "Content-Type"
	ACCEPT_KEY             = "Accept"
	REGEX_URL_PATH_ELEMENT = "[\\w-]+"
	DEFAULT_INTEGER        = 0
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
	ERROR_CODE_FROM_MQTT               = 105
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

const MAX_UPLOAD_FILE_SIZE = 50 << 20

const MAX_WEBSOCKET_READ_BUFFER_SIZE = 1024
const MAX_WEBSOCKET_WRITE_BUFFER_SIZE = 1024

const (
	DB_TYPE_POSTGRES = "postgres"
	DB_TYPE_ORACLE   = "oracle"
)

const MAXIMIZE_QUERY_COUNT_IN_ORACLE_DATABASE = 100000

const WAIT_MQTT_DISCONNECT_TIMEOUT = 250
const (
	MQTT_QOS_AT_MOST_ONCE  = 0
	MQTT_QOS_AT_LEAST_ONCE = 1
	MQTT_QOS_EXACTLY_ONCE  = 2
)

const (
	MQTT_MESSAGE_PAYLOAD_TYPE_JSON   = "json"
	MQTT_MESSAGE_PAYLOAD_TYPE_BYTES  = "bytes"
	MQTT_MESSAGE_PAYLOAD_TYPE_STRING = "string"
)
