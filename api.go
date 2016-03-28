package golog

type Logger interface {
	LogTrace(format string, v ...interface{})
	LogDebug(format string, v ...interface{})
	LogInfo(format string, v ...interface{})
	LogWarn(format string, v ...interface{})
	LogError(format string, v ...interface{})
	LogFatal(format string, v ...interface{})
	Log(format string, v ...interface{})
	Close()
}

type BufferLogger interface {
	LogTrace(buffer_idx int, format string, v ...interface{})
	LogDebug(buffer_idx int, format string, v ...interface{})
	LogInfo(buffer_idx int, format string, v ...interface{})
	LogWarn(buffer_idx int, format string, v ...interface{})
	LogError(buffer_idx int, format string, v ...interface{})
	LogFatal(buffer_idx int, format string, v ...interface{})
	Log(buffer_idx int, format string, v ...interface{})
	FlushLogBuffer(buffer_idx int)
	Close()
}

func GetStdLogger() Logger {
	return NewLog("", "", LOG_LEVEL_STDOUT, 0, 0, 0)
}
