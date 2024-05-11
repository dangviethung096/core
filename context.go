package core

import (
	"time"
)

/*
* Context type: which carries deadlines, cancellation signals,
* and other request-scoped values across API boundaries and between processes.
 */

type Context interface {
	// Reimplement from context.Context
	Deadline() (deadline time.Time, ok bool)
	Done() <-chan struct{}
	Err() error
	Value(key any) any
	GetContextID() string
	GetTimeout() time.Duration
	GetCancelFunc() func()
	LogInfo(format string, args ...interface{})
	LogDebug(format string, args ...interface{})
	LogError(format string, args ...interface{})
	LogWarning(format string, args ...interface{})
	LogFatal(format string, args ...interface{})
}
