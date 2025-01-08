package core

type logger interface {
	LogInfo(format string, args ...interface{})
	LogDebug(format string, args ...interface{})
	LogError(format string, args ...interface{})
	LogWarning(format string, args ...interface{})
	LogFatal(format string, args ...interface{})
	LogPanic(format string, args ...interface{})
}

// Implement the logger interface
func LogInfo(format string, args ...interface{}) {
	coreContext.(*rootContext).LogInfoWithCallStack(format, 3, args...)
}

func LogDebug(format string, args ...interface{}) {
	coreContext.(*rootContext).LogDebugWithCallStack(format, 3, args...)
}

func LogWarning(format string, args ...interface{}) {
	coreContext.(*rootContext).LogWarningWithCallStack(format, 3, args...)
}

func LogError(format string, args ...interface{}) {
	coreContext.(*rootContext).LogErrorWithCallStack(format, 3, args...)
}

func LogFatal(format string, args ...interface{}) {
	coreContext.(*rootContext).LogFatalWithCallStack(format, 3, args...)
}

func LogPanic(format string, args ...interface{}) {
	coreContext.(*rootContext).LogPanicWithCallStack(format, 3, args...)
}
