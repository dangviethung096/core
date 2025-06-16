package core

import (
	"context"
	"fmt"
	"log"
	"net/http/httptest"
	"runtime"
	"strconv"
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
* @params: format string, args ...any
* @return: void
 */
func (ctx *rootContext) LogInfo(format string, args ...any) {
	logStr, logInfo := ctx.formatWithCallStack(format, 2, args...)
	logStr = "[INFO] " + logStr
	log.Println(logStr)
	if Config.Log.UseElasticsearch {
		esClient.IndexDocument(coreContext, Config.Log.ElasticIndex, BLANK, logInfo)
	}
}

func (ctx *rootContext) LogInfoWithCallStack(format string, callStack int, args ...any) {
	logStr, logInfo := ctx.formatWithCallStack(format, callStack, args...)
	logStr = "[INFO] " + logStr
	log.Println(logStr)
	if Config.Log.UseElasticsearch {
		esClient.IndexDocument(coreContext, Config.Log.ElasticIndex, BLANK, logInfo)
	}
}

/*
* Debug: Log Debug with context information
* @params: format string, args ...any
* @return: void
 */
func (ctx *rootContext) LogDebug(format string, args ...any) {
	logStr, logInfo := ctx.formatWithCallStack(format, 2, args...)
	logStr = "[DEBUG] " + logStr
	log.Println(logStr)
	if Config.Log.UseElasticsearch {
		esClient.IndexDocument(coreContext, Config.Log.ElasticIndex, BLANK, logInfo)
	}
}

func (ctx *rootContext) LogDebugWithCallStack(format string, callStack int, args ...any) {
	logStr, logInfo := ctx.formatWithCallStack(format, callStack, args...)
	logStr = "[DEBUG] " + logStr
	log.Println(logStr)
	if Config.Log.UseElasticsearch {
		esClient.IndexDocument(coreContext, Config.Log.ElasticIndex, BLANK, logInfo)
	}
}

/*
* Error: Log Error with context information
* @params: format string, args ...any
* @return: void
 */
func (ctx *rootContext) LogError(format string, args ...any) {
	logStr, logInfo := ctx.formatWithCallStack(format, 2, args...)
	logStr = "[ERROR] " + logStr
	log.Println(logStr)
	if Config.Log.UseElasticsearch {
		esClient.IndexDocument(coreContext, Config.Log.ElasticIndex, BLANK, logInfo)
	}
}

func (ctx *rootContext) LogErrorWithCallStack(format string, callStack int, args ...any) {
	logStr, logInfo := ctx.formatWithCallStack(format, callStack, args...)
	logStr = "[ERROR] " + logStr
	log.Println(logStr)
	if Config.Log.UseElasticsearch {
		esClient.IndexDocument(coreContext, Config.Log.ElasticIndex, BLANK, logInfo)
	}
}

/*
* Warning: Log Warning with context information
* @params: format string, args ...any
* @return: void
 */
func (ctx *rootContext) LogWarning(format string, args ...any) {
	logStr, logInfo := ctx.formatWithCallStack(format, 2, args...)
	logStr = "[WARNING] " + logStr
	log.Println(logStr)
	if Config.Log.UseElasticsearch {
		esClient.IndexDocument(coreContext, Config.Log.ElasticIndex, BLANK, logInfo)
	}
}

func (ctx *rootContext) LogWarningWithCallStack(format string, callStack int, args ...any) {
	logStr, logInfo := ctx.formatWithCallStack(format, callStack, args...)
	logStr = "[WARNING] " + logStr
	log.Println(logStr)
	if Config.Log.UseElasticsearch {
		esClient.IndexDocument(coreContext, Config.Log.ElasticIndex, BLANK, logInfo)
	}
}

/*
* Panic: Log Panic with context information
* @params: format string, args ...any
* @return: void
 */
func (ctx *rootContext) LogPanic(format string, args ...any) {
	logStr, logInfo := ctx.formatWithCallStack(format, 2, args...)
	logStr = "[Panic] " + logStr
	log.Panicln(logStr)
	if Config.Log.UseElasticsearch {
		esClient.IndexDocument(coreContext, Config.Log.ElasticIndex, BLANK, logInfo)
	}
}

func (ctx *rootContext) LogPanicWithCallStack(format string, callStack int, args ...any) {
	logStr, logInfo := ctx.formatWithCallStack(format, callStack, args...)
	logStr = "[Panic] " + logStr
	log.Panicln(logStr)
	if Config.Log.UseElasticsearch {
		esClient.IndexDocument(coreContext, Config.Log.ElasticIndex, BLANK, logInfo)
	}
}

/*
* Fatal: Log Fatal with context information
* @params: format string, args ...any
* @return: void
 */
func (ctx *rootContext) LogFatal(format string, args ...any) {
	logStr, logInfo := ctx.formatWithCallStack(format, 2, args...)
	logStr = "[FATAL] " + logStr
	log.Fatalln(logStr)
	if Config.Log.UseElasticsearch {
		esClient.IndexDocument(coreContext, Config.Log.ElasticIndex, BLANK, logInfo)
	}
}

func (ctx *rootContext) LogFatalWithCallStack(format string, callStack int, args ...any) {
	logStr, logInfo := ctx.formatWithCallStack(format, callStack, args...)
	logStr = "[FATAL] " + logStr
	log.Fatalln(logStr)
	if Config.Log.UseElasticsearch {
		esClient.IndexDocument(coreContext, Config.Log.ElasticIndex, BLANK, logInfo)
	}
}

/*
* format: Format the ctx: add to message log the file name
* and line number of the code that calls the ctx interface
* @params: format string, args ...any
* @return: string
 */

func (ctx *rootContext) formatWithCallStack(format string, callStack int, args ...any) (string, logMessage) {
	// Format the ctx
	logStr := fmt.Sprintf(format, args...)
	logInfo := logMessage{
		RequestID: BLANK,
		Message:   logStr,
	}

	// Get the file name and line number of the code that calls the ctx interface
	pc, file, line, ok := runtime.Caller(callStack)
	functionName := BLANK
	if ok {
		path := strings.Split(file, "/")
		if len(path) > 3 {
			file = strings.Join(path[len(path)-3:], "/")
		}
		// Get function name
		functionPath := strings.Split(runtime.FuncForPC(pc).Name(), "/")
		if len(functionPath) > 0 {
			functionName = functionPath[len(functionPath)-1]
		}

		logStr = fmt.Sprintf("%s:%d:%s, RequestID: %s , Message: %s", file, line, functionName, BLANK, logStr)
	}

	logInfo.Caller = functionName
	logInfo.File = file + ":" + strconv.Itoa(line)
	logInfo.Timestamp = time.Now().Format(time.RFC3339)

	// Return the formatted string
	return logStr, logInfo
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
	ctx.requestID = ID.GenerateID()
	// Init new request
	ctx.rw = httptest.NewRecorder()
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

func GetContextWithoutTimeout() Context {
	ctx := contextPool.Get().(*rootContext)
	ctx.Context, ctx.cancelFunc = context.WithCancel(coreContext)
	ctx.contextID = ID.GenerateID()
	return ctx
}

func GetContextWithDefaultTimeout() Context {
	ctx := contextPool.Get().(*rootContext)
	ctx.Context, ctx.cancelFunc = context.WithTimeout(coreContext, contextTimeout)
	ctx.timeout = contextTimeout
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
