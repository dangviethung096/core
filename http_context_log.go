package core

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"
)

/*
* Info: Log Info with context information
* @params: format string, args ...any
* @return: void
 */
func (ctx *HttpContext) LogInfo(format string, args ...any) {
	logStr, logInfo := ctx.format(format, args...)
	logInfo.Level = "INFO"
	logStr = "[INFO] " + logStr
	log.Println(logStr)
	if Config.Log.UseElasticsearch {
		esClient.IndexDocument(coreContext, Config.Log.ElasticIndex, ctx.requestID, logInfo)
	}
}

/*
* Debug: Log Debug with context information
* @params: format string, args ...any
* @return: void
 */
func (ctx *HttpContext) LogDebug(format string, args ...any) {
	logStr, logInfo := ctx.format(format, args...)
	logInfo.Level = "DEBUG"
	logStr = "[DEBUG] " + logStr
	log.Println(logStr)
	if Config.Log.UseElasticsearch {
		esClient.IndexDocument(coreContext, Config.Log.ElasticIndex, ctx.requestID, logInfo)
	}
}

/*
* Error: Log Error with context information
* @params: format string, args ...any
* @return: void
 */
func (ctx *HttpContext) LogError(format string, args ...any) {
	logStr, logInfo := ctx.format(format, args...)
	logInfo.Level = "ERROR"
	logStr = "[ERROR] " + logStr
	log.Println(logStr)
	if Config.Log.UseElasticsearch {
		esClient.IndexDocument(coreContext, Config.Log.ElasticIndex, ctx.requestID, logInfo)
	}
}

/*
* Warning: Log Warning with context information
* @params: format string, args ...any
* @return: void
 */
func (ctx *HttpContext) LogWarning(format string, args ...any) {
	logStr, logInfo := ctx.format(format, args...)
	logInfo.Level = "WARNING"
	logStr = "[WARNING] " + logStr
	log.Println(logStr)
	if Config.Log.UseElasticsearch {
		esClient.IndexDocument(coreContext, Config.Log.ElasticIndex, ctx.requestID, logInfo)
	}
}

/*
* Fatal: Log Fatal with context information
* @params: format string, args ...any
* @return: void
 */
func (ctx *HttpContext) LogFatal(format string, args ...any) {
	logStr, logInfo := ctx.format(format, args...)
	logInfo.Level = "FATAL"
	logStr = "[FATAL] " + logStr
	log.Fatalln(logStr)
	if Config.Log.UseElasticsearch {
		esClient.IndexDocument(coreContext, Config.Log.ElasticIndex, ctx.requestID, logInfo)
	}
}

func (ctx *HttpContext) LogPanic(format string, args ...any) {
	logStr, logInfo := ctx.format(format, args...)
	logInfo.Level = "PANIC"
	logStr = "[PANIC] " + logStr
	log.Panicln(logStr)
	if Config.Log.UseElasticsearch {
		esClient.IndexDocument(coreContext, Config.Log.ElasticIndex, ctx.requestID, logInfo)
	}
}

/*
* format: Format the ctx: add to message log the file name
* and line number of the code that calls the ctx interface
* @params: format string, args ...any
* @return: string
 */
func (ctx *HttpContext) format(format string, args ...any) (string, logMessage) {
	// Format the ctx
	logStr := fmt.Sprintf(format, args...)
	logInfo := logMessage{
		RequestID: ctx.requestID,
		Message:   logStr,
	}

	// Get the file name and line number of the code that calls the ctx interface
	pc, file, line, ok := runtime.Caller(2)
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

		logStr = fmt.Sprintf("%s:%d:%s RequestID: %s, Message: %s", file, line, functionName, ctx.requestID, logStr)
	}

	logInfo.Caller = functionName
	logInfo.File = file
	logInfo.Timestamp = time.Now().Format(time.RFC3339)

	// Return the formatted string
	return logStr, logInfo
}
