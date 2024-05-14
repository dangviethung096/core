package core

import (
	"context"
	"time"
)

/*
* Context type: which carries deadlines, cancellation signals,
* and other request-scoped values across API boundaries and between processes.
 */

type Context interface {
	// Reimplement from context.Context
	context.Context
	logger
	Value(key any) any
	GetContextID() string
	GetTimeout() time.Duration
	GetCancelFunc() func()
}
