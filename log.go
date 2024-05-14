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
	coreContext.(*rootContext).LogInfo(format, args...)
}

func LogDebug(format string, args ...interface{}) {
	coreContext.(*rootContext).LogError(format, args...)
}

func LogWarning(format string, args ...interface{}) {
	coreContext.(*rootContext).LogWarning(format, args...)
}

func LogError(format string, args ...interface{}) {
	coreContext.(*rootContext).LogError(format, args...)
}

func LogFatal(format string, args ...interface{}) {
	coreContext.(*rootContext).LogFatal(format, args...)
}

func LogPanic(format string, args ...interface{}) {
	coreContext.(*rootContext).LogPanic(format, args...)
}
