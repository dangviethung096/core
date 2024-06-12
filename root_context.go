package core

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"
)

type rootContext struct {
	context.Context
	contextID  string
	timeout    time.Duration
	cancelFunc func()
}

/*
* GetContextID: Get the context ID
* @params: void
* @return: string
 */
func (root *rootContext) GetContextID() string {
	return root.contextID
}

/*
* GetTimeout: Get the timeout
* @params: void
* @return: int64
 */
func (root *rootContext) GetTimeout() time.Duration {
	return root.timeout
}

/*
* GetCancelFunc: Get the cancel function
* @params: void
* @return: func()
 */
func (root *rootContext) GetCancelFunc() func() {
	return root.cancelFunc
}

/*
* Info: Log Info with context information
* @params: format string, args ...interface{}
* @return: void
 */
func (ctx *rootContext) LogInfo(format string, args ...interface{}) {
	logStr := "[INFO] " + ctx.format(format, args...)
	log.Println(logStr)
}

/*
* Debug: Log Debug with context information
* @params: format string, args ...interface{}
* @return: void
 */
func (ctx *rootContext) LogDebug(format string, args ...interface{}) {
	logStr := "[DEBUG] " + ctx.format(format, args...)
	log.Println(logStr)
}

/*
* Error: Log Error with context information
* @params: format string, args ...interface{}
* @return: void
 */
func (ctx *rootContext) LogError(format string, args ...interface{}) {
	logStr := "[ERROR] " + ctx.format(format, args...)
	log.Println(logStr)
}

/*
* Warning: Log Warning with context information
* @params: format string, args ...interface{}
* @return: void
 */
func (ctx *rootContext) LogWarning(format string, args ...interface{}) {
	logStr := "[WARNING] " + ctx.format(format, args...)
	log.Fatalln(logStr)
}

/*
* Panic: Log Panic with context information
* @params: format string, args ...interface{}
* @return: void
 */
func (ctx *rootContext) LogPanic(format string, args ...interface{}) {
	logStr := "[Panic] " + ctx.format(format, args...)
	log.Panicln(logStr)
}

/*
* Fatal: Log Fatal with context information
* @params: format string, args ...interface{}
* @return: void
 */
func (ctx *rootContext) LogFatal(format string, args ...interface{}) {
	logStr := "[FATAL] " + ctx.format(format, args...)
	log.Fatalln(logStr)
}

/*
* format: Format the ctx: add to message log the file name
* and line number of the code that calls the ctx interface
* @params: format string, args ...interface{}
* @return: string
 */
func (ctx *rootContext) format(format string, args ...interface{}) string {
	// Format the ctx
	logStr := fmt.Sprintf(format, args...)

	// Get the file name and line number of the code that calls the ctx interface
	pc, file, line, ok := runtime.Caller(3)
	if ok {
		path := strings.Split(file, "/")
		if len(path) > 3 {
			file = strings.Join(path[len(path)-3:], "/")
		}
		// Get function name
		functionPath := strings.Split(runtime.FuncForPC(pc).Name(), "/")
		functionName := BLANK
		if len(functionPath) > 0 {
			functionName = functionPath[len(functionPath)-1]
		}

		logStr = fmt.Sprintf("%s:%d:%s, RequestID: %s , Message: %s", file, line, functionName, ctx.contextID, logStr)
	}

	// Return the formatted string
	return logStr
}

/*
* GetContextForTest: Get context for test
* Caution: This function is only used for test
* @return: Context
 */
func GetContextForTest() Context {
	ctx := contextPool.Get().(*rootContext)
	ctx.Context, ctx.cancelFunc = context.WithTimeout(coreContext, contextTimeout)
	ctx.contextID = ID.GenerateID()
	return ctx
}

func GetHttpContextForTest() *HttpContext {
	ctx := httpContextPool.Get().(*HttpContext)
	ctx.Context, ctx.cancelFunc = context.WithTimeout(coreContext, contextTimeout)
	ctx.requestID = ID.GenerateID()
	return ctx
}

/*
* Get child of core context with timeout as a parameter
 */
func GetContextWithTimeout(timeout time.Duration) Context {
	ctx := contextPool.Get().(*rootContext)
	ctx.Context, ctx.cancelFunc = context.WithTimeout(coreContext, timeout)
	ctx.timeout = timeout
	ctx.contextID = ID.GenerateID()
	return ctx
}

/*
* Get core context
* Return core context to http context pool
* @return: Context
 */
func GetContext() Context {
	return coreContext
}

/*
* Get context from http context pool
* Return context to http context pool
 */
func PutContext(ctx Context) {
	ctx.GetCancelFunc()()
	contextPool.Put(ctx)
}
