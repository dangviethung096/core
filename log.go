package core

import (
	"fmt"
	"log"
	"runtime"
	"strings"
)

type Logger interface {
	Info(format string, args ...interface{})
	Debug(format string, args ...interface{})
	Warning(format string, args ...interface{})
	Error(format string, args ...interface{})
	Fatal(format string, args ...interface{})
	Panic(format string, args ...interface{})
}

type logger struct {
}

func initLogger() *logger {
	return &logger{}
}

// Implement the logger interface
func (logger logger) Info(format string, args ...interface{}) {
	logStr := "[INFO] " + logger.format(format, args...)
	log.Println(logStr)
}

func (logger logger) Debug(format string, args ...interface{}) {
	logStr := "[DEBUG] " + logger.format(format, args...)
	log.Println(logStr)
}

func (logger logger) Warning(format string, args ...interface{}) {
	logStr := "[WARNING] " + logger.format(format, args...)
	log.Println(logStr)
}

func (logger logger) Error(format string, args ...interface{}) {
	logStr := "[ERROR] " + logger.format(format, args...)
	log.Println(logStr)
}

func (logger logger) Fatal(format string, args ...interface{}) {
	logStr := "[FATAL] " + logger.format(format, args...)
	log.Fatalln(logStr)
}

func (logger logger) Panic(format string, args ...interface{}) {
	logStr := "[PANIC] " + logger.format(format, args...)
	log.Panicln(logStr)
}

// Implement the format function
// Use the format function in the logger interface
// Read the file of code that calls the logger interface
// Print the line of code that calls the logger interface
func (logger logger) format(format string, args ...interface{}) string {
	// Format the logger
	logStr := fmt.Sprintf(format, args...)

	// Get the file name and line number of the code that calls the logger interface
	_, file, line, ok := runtime.Caller(2)
	if ok {
		path := strings.Split(file, "/")
		if len(path) > 3 {
			file = strings.Join(path[len(path)-3:], "/")
		}
		logStr = fmt.Sprintf("%s:%d: %s", file, line, logStr)
	}

	// Return the formatted string
	return logStr
}
