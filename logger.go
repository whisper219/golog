package golog

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"time"
)

//log level, from max to min,
//LOG_LEVEL_STDOUT output log to stdout
const (
	LOG_LEVEL_FATAL = iota
	LOG_LEVEL_ERROR
	LOG_LEVEL_WARN
	LOG_LEVEL_INFO
	LOG_LEVEL_DEBUG
	LOG_LEVEL_TRACE
	LOG_LEVEL_ALL
	LOG_LEVEL_STDOUT
)

const (
	LOG_SHIFT_BY_SIZE = iota
	LOG_SHIFT_BY_DAY
	LOG_SHIFT_BY_HOUR
	LOG_SHIFT_BY_MINUTE
)

const LOG_DEFAULT_TICKER_INTERVAL = 1 * time.Second

type LogBufferInfo struct {
	content string
	time    time.Time
}

type Log struct {
	log_path            string
	log_prefix          string
	log_level           int
	log_num             int
	log_size            int64
	log_shift_type      int
	log_buffer          map[int]LogBufferInfo
	log_buffer_ttl      time.Duration
	check_buffer_ticker *time.Ticker
	closed              chan bool
}

// create Log obj
// log_path: the directory of the log file, etc: "../log/"
// log_prefix: the name of the log file,
// log_level: defined in this package, etc: LOG_LEVEL_INFO
// log_num: max saved log file, only enabled in LOG_SHIFT_BY_SIZE, if log file num exceed log_num, the earliest log file will be deleted, todo: enabled in all shift type
// log_size: max log file size (bytes), only enabled in LOG_SHIFT_BY_SIZE
// log_shift_type: defined in this package, etc: LOG_SHIFT_BY_SIZE
// log_buffer_ttl: only for "buffer-flush" log, if log content hasn't be flushed and time reaches ttl, it will force log content to flush to the file.
func NewRawLog(log_path string, log_prefix string, log_level int, log_num int, log_size int64,
	log_shift_type int, log_buffer_ttl time.Duration) *Log {
	var l Log
	l.log_path = log_path
	l.log_prefix = log_prefix
	l.log_level = log_level
	l.log_num = log_num
	l.log_size = log_size
	l.log_shift_type = log_shift_type
	l.log_buffer = make(map[int]LogBufferInfo)
	l.log_buffer_ttl = log_buffer_ttl
	l.closed = make(chan bool)

	//todo: check arguments

	if log_buffer_ttl > 0 {
		go l.checkLogExpire()
	}
	return &l
}

// modify log config
func (l *Log) ModConf(log_path string, log_prefix string, log_level int, log_num int, log_size int64,
	log_shift_type int, log_buffer_ttl time.Duration) {
	l.log_path = log_path
	l.log_prefix = log_prefix
	l.log_level = log_level
	l.log_num = log_num
	l.log_size = log_size
	l.log_shift_type = log_shift_type
	l.log_buffer_ttl = log_buffer_ttl

	//todo: check ttl > 0, restart checkLogExpire()
}

func (l *Log) Close() {
	if l.check_buffer_ticker != nil {
		l.check_buffer_ticker.Stop()
		l.closed <- true
	}
}

func (l *Log) RawLog(content string) (err error) {
	if l.log_level != LOG_LEVEL_STDOUT {
		var file *os.File
		if file, err = l.getLogFile(); err != nil {
			return err
		}
		defer file.Close()

		if _, err = fmt.Fprintln(file, content); err != nil {
			return err
		}
	} else {
		fmt.Println(content)
	}
	return nil
}

func (l *Log) LogTrace(buffer_idx int, format string, v ...interface{}) {
	if l.log_level >= LOG_LEVEL_TRACE {
		l.logFormat(buffer_idx, LOG_LEVEL_TRACE, format, v...)
	}
}

func (l *Log) LogDebug(buffer_idx int, format string, v ...interface{}) {
	if l.log_level >= LOG_LEVEL_DEBUG {
		l.logFormat(buffer_idx, LOG_LEVEL_DEBUG, format, v...)
	}
}

func (l *Log) LogInfo(buffer_idx int, format string, v ...interface{}) {
	if l.log_level >= LOG_LEVEL_INFO {
		l.logFormat(buffer_idx, LOG_LEVEL_INFO, format, v...)
	}
}

func (l *Log) LogWarn(buffer_idx int, format string, v ...interface{}) {
	if l.log_level >= LOG_LEVEL_WARN {
		l.logFormat(buffer_idx, LOG_LEVEL_WARN, format, v...)
	}
}

func (l *Log) LogError(buffer_idx int, format string, v ...interface{}) {
	if l.log_level >= LOG_LEVEL_ERROR {
		l.logFormat(buffer_idx, LOG_LEVEL_ERROR, format, v...)
	}
}

func (l *Log) LogFatal(buffer_idx int, format string, v ...interface{}) {
	if l.log_level >= LOG_LEVEL_FATAL {
		l.logFormat(buffer_idx, LOG_LEVEL_FATAL, format, v...)
	}
}

//easy way to log content to file
func (l *Log) Log(buffer_idx int, format string, v ...interface{}) {
	l.logFormat(buffer_idx, LOG_LEVEL_ALL, format, v...)
}

func (l *Log) FlushLogBuffer(buffer_idx int) {
	if buffer_info, ok := l.log_buffer[buffer_idx]; ok {
		l.RawLog(buffer_info.content + fmt.Sprintf("\nLog Buffer end, Key=%d\n", buffer_idx))
		delete(l.log_buffer, buffer_idx)
	}
}

func (l *Log) logFormat(buffer_idx int, log_level int, format string, v ...interface{}) {
	log_content := l.formatLogContent(log_level, format, v...)

	if buffer_idx != 0 {
		if buffer_info, ok := l.log_buffer[buffer_idx]; ok {
			buffer_info.content = buffer_info.content + "\n" + log_content
			l.log_buffer[buffer_idx] = buffer_info
		} else {
			l.log_buffer[buffer_idx] = LogBufferInfo{fmt.Sprintf("\nLog Buffer Start, Key=%d\n", buffer_idx) + log_content, time.Now()}
		}
	} else {
		l.RawLog(log_content)
	}
}

func (l *Log) formatLogContent(log_level int, format string, v ...interface{}) string {
	var log_content string
	log_content = time.Now().Format("[2006-01-02] 15:04:05.99999")
	switch log_level {
	case LOG_LEVEL_FATAL:
		log_content += " [FATAL] "
	case LOG_LEVEL_ERROR:
		log_content += " [ERROR] "
	case LOG_LEVEL_WARN:
		log_content += " [WARN] "
	case LOG_LEVEL_DEBUG:
		log_content += " [DEBUG] "
	case LOG_LEVEL_INFO:
		log_content += " [INFO] "
	case LOG_LEVEL_TRACE:
		log_content += " [TRACE] "
	case LOG_LEVEL_ALL:
		log_content += " [LOG] "
	default:
		log_content += " [LOG] "
	}
	log_content += fmt.Sprintf(format, v...)

	_, full_filename, line, _ := runtime.Caller(3)
	full_filename_arr := bytes.Split([]byte(full_filename), []byte(`/`))
	short_filename := string(full_filename_arr[len(full_filename_arr)-1])
	log_content += fmt.Sprintf(" %s:%d", short_filename, line)
	return log_content
}

func (l *Log) checkLogExpire() {
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
			return
		}
	}
}

func (l *Log) shiftLogFile() (err error) {
	cur_logfile := l.getLogFileName(0)
	var file *os.File
	var stat os.FileInfo
	if file, err = os.OpenFile(cur_logfile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666); err != nil {
		return err
	}
	defer file.Close()
	if stat, err = file.Stat(); err != nil {
		return err
	}

	if l.log_shift_type == LOG_SHIFT_BY_SIZE {
		if stat.Size() >= l.log_size {
			for i := l.log_num - 2; i >= 0; i-- {
				var filename string = l.getLogFileName(i)
				if _, err := os.Stat(filename); err == nil {
					os.Rename(filename, l.getLogFileName(i+1))
				}
			}
		}
	} else if l.log_shift_type == LOG_SHIFT_BY_DAY {
		cur_time_str := time.Now().Format("20060102")
		if cur_time_str != stat.ModTime().Format("20060102") {
			os.Rename(cur_logfile, cur_logfile+"."+cur_time_str)
		}
	} else if l.log_shift_type == LOG_SHIFT_BY_HOUR {
		cur_time_str := time.Now().Format("20060102-15")
		if cur_time_str != stat.ModTime().Format("20060102-15") {
			os.Rename(cur_logfile, cur_logfile+"."+cur_time_str)
		}
	} else if l.log_shift_type == LOG_SHIFT_BY_MINUTE {
		cur_time_str := time.Now().Format("20060102-1504")
		if cur_time_str != stat.ModTime().Format("20060102-1504") {
			os.Rename(cur_logfile, cur_logfile+"."+cur_time_str)
		}
	}
	return nil
}

func (l *Log) getLogFile() (file *os.File, err error) {
	l.shiftLogFile()
	cur_logfile := l.getLogFileName(0)
	if file, err = os.OpenFile(cur_logfile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666); err != nil {
		return nil, err
	}

	return file, nil
}

func (l *Log) getLogFileName(idx int) (name string) {
	if idx == 0 {
		return fmt.Sprintf("%s%s%s", l.log_path, l.log_prefix, ".log")
	} else {
		return fmt.Sprintf("%s%s%d%s", l.log_path, l.log_prefix, idx, ".log")
	}
}
