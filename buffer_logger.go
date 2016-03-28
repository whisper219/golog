package golog

import (
	"fmt"
	"sync"
	"time"
)

const LOG_DEFAULT_TICKER_INTERVAL = 1 * time.Second

type LogBufferInfo struct {
	content string
	time    time.Time
}

type BufferLog struct {
	logger              *Log
	log_buffer          map[int]*LogBufferInfo
	log_buffer_ttl      time.Duration
	check_buffer_ticker *time.Ticker
	closed              chan bool
	mutex               *sync.Mutex
}

// create BufferLog obj
// log_path: the directory of the log file, etc: "../log/"
// log_prefix: the name of the log file,
// log_level: defined in this package, etc: LOG_LEVEL_INFO
// log_num: max saved log file, only enabled in LOG_SHIFT_BY_SIZE, if log file num exceed log_num, the earliest log file will be deleted, todo: enabled in all shift type
// log_size: max log file size (bytes), only enabled in LOG_SHIFT_BY_SIZE
// log_shift_type: defined in this package, etc: LOG_SHIFT_BY_SIZE
// log_buffer_ttl: only for "buffer-flush" log, if log content hasn't be flushed and time reaches ttl, it will force log content to flush to the file.
func NewBufferLog(log_path string, log_prefix string, log_level int, log_num int, log_size int64,
	log_shift_type int, log_buffer_ttl time.Duration) *BufferLog {
	l := new(BufferLog)
	l.logger = NewLog(log_path, log_prefix, log_level, log_num, log_size, log_shift_type)
	l.log_buffer = make(map[int]*LogBufferInfo)
	l.log_buffer_ttl = log_buffer_ttl
	l.closed = make(chan bool)
	l.mutex = new(sync.Mutex)

	//todo: check arguments

	if log_buffer_ttl > 0 {
		go l.checkLogExpire()
	}
	return l
}

// modify log config
func (l *BufferLog) ModConf(log_path string, log_prefix string, log_level int, log_num int, log_size int64,
	log_shift_type int, log_buffer_ttl time.Duration) {
	l.logger.ModConf(log_path, log_prefix, log_level, log_num, log_size, log_shift_type)
	l.log_buffer_ttl = log_buffer_ttl

	//todo: check ttl > 0, restart checkLogExpire()
}

func (l *BufferLog) Close() {
	if l.check_buffer_ticker != nil {
		l.check_buffer_ticker.Stop()
		l.closed <- true
	}
}

func (l *BufferLog) LogTrace(buffer_idx int, format string, v ...interface{}) {
	if l.logger.log_level >= LOG_LEVEL_TRACE {
		l.logFormat(buffer_idx, LOG_LEVEL_TRACE, format, v...)
	}
}

func (l *BufferLog) LogDebug(buffer_idx int, format string, v ...interface{}) {
	if l.logger.log_level >= LOG_LEVEL_DEBUG {
		l.logFormat(buffer_idx, LOG_LEVEL_DEBUG, format, v...)
	}
}

func (l *BufferLog) LogInfo(buffer_idx int, format string, v ...interface{}) {
	if l.logger.log_level >= LOG_LEVEL_INFO {
		l.logFormat(buffer_idx, LOG_LEVEL_INFO, format, v...)
	}
}

func (l *BufferLog) LogWarn(buffer_idx int, format string, v ...interface{}) {
	if l.logger.log_level >= LOG_LEVEL_WARN {
		l.logFormat(buffer_idx, LOG_LEVEL_WARN, format, v...)
	}
}

func (l *BufferLog) LogError(buffer_idx int, format string, v ...interface{}) {
	if l.logger.log_level >= LOG_LEVEL_ERROR {
		l.logFormat(buffer_idx, LOG_LEVEL_ERROR, format, v...)
	}
}

func (l *BufferLog) LogFatal(buffer_idx int, format string, v ...interface{}) {
	if l.logger.log_level >= LOG_LEVEL_FATAL {
		l.logFormat(buffer_idx, LOG_LEVEL_FATAL, format, v...)
	}
}

//easy way to log content to file
func (l *BufferLog) Log(buffer_idx int, format string, v ...interface{}) {
	l.logFormat(buffer_idx, LOG_LEVEL_ALL, format, v...)
}

func (l *BufferLog) FlushLogBuffer(buffer_idx int) {
	if buffer_info, ok := l.log_buffer[buffer_idx]; ok {
		l.logger.RawLog(buffer_info.content + fmt.Sprintf("\n[Log end], [Key=%d]\n", buffer_idx))
		l.mutex.Lock()
		delete(l.log_buffer, buffer_idx)
		l.mutex.Unlock()
	}
}

func (l *BufferLog) logFormat(buffer_idx int, log_level int, format string, v ...interface{}) {
	log_content := l.logger.formatLogContent(log_level, format, v...)

	if buffer_idx != 0 {
		if buffer_info, ok := l.log_buffer[buffer_idx]; ok {
			l.mutex.Lock()
			buffer_info.content = buffer_info.content + "\n" + log_content
			//l.log_buffer[buffer_idx] = buffer_info
			l.mutex.Unlock()
		} else {
			l.mutex.Lock()
			l.log_buffer[buffer_idx] = &LogBufferInfo{fmt.Sprintf("[Log Begin], [Key=%d]\n", buffer_idx) + log_content, time.Now()}
			l.mutex.Unlock()
		}
	} else {
		l.logger.RawLog(log_content)
	}
}

func (l *BufferLog) checkLogExpire() {
	l.check_buffer_ticker = time.NewTicker(LOG_DEFAULT_TICKER_INTERVAL)
	for {
		select {
		case <-l.check_buffer_ticker.C:
			for idx, value := range l.log_buffer {
				if duration := time.Since(value.time); duration >= l.log_buffer_ttl {
					l.FlushLogBuffer(idx)
				}
			}
		case <-l.closed:
			for idx, _ := range l.log_buffer {
				l.FlushLogBuffer(idx)
			}
			l.logger.Close()
			return
		}
	}
}
