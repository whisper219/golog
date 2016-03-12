package golog

import (
	"time"
)

type Logger interface {
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

var loggermap map[string]*Log = make(map[string]*Log)

func AddInstance(logger_key string, log_path string, log_prefix string, log_level int, log_num int, log_size int64,
	log_shift_type int, log_buffer_ttl time.Duration) {
	if logger, ok := loggermap[logger_key]; ok {
		logger.ModConf(log_path, log_prefix, log_level, log_num, log_size, log_shift_type, log_buffer_ttl)
	} else {
		signallogger := NewRawLog(log_path, log_prefix, log_level, log_num, log_size, log_shift_type, log_buffer_ttl)
		loggermap[logger_key] = signallogger
	}
}

func GetInstance(logger_key string) Logger {
	if logger, ok := loggermap[logger_key]; ok {
		return logger
	}
	return GetStdLogger()
}

func GetStdLogger() Logger {
	return NewRawLog("", "", LOG_LEVEL_STDOUT, 0, 0, 0, 0)
}
