package core

import (
	"context"
	"time"
)

/*
* Context type: which carries deadlines, cancellation signals,
* and other request-scoped values across API boundaries and between processes.
 */
type Context struct {
	context.Context
	cancelFunc context.CancelFunc
	requestID  string
	Timeout    time.Duration
}
