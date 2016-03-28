package main

import (
	"github.com/whisper219/golog"
	"time"
)

func testLogger() {
	//test log shift by size
	log := golog.NewLog("./", "test", golog.LOG_LEVEL_DEBUG, 5, 10000, golog.LOG_SHIFT_BY_SIZE)
	log.LogDebug("test log debug %d", 1)
	log.LogTrace("test log trace %d", 2)
	log.LogError("test log error %d", 3)

	for i := 0; i < 100; i++ {
		log.LogDebug("test log shift %d", i)
	}
	log.Close()

	//test log shift by time
	log = golog.NewLog("./", "test_shiftbymin", golog.LOG_LEVEL_DEBUG, 5, 10000, golog.LOG_SHIFT_BY_MINUTE)
	for i := 0; i < 3; i++ {
		log.LogDebug("test log shift %d", i)
		time.Sleep(time.Minute)
	}
	log.Close()
}

func testBufferLogger() {
	log := golog.NewBufferLog("./", "test", golog.LOG_LEVEL_DEBUG, 5, 10000, golog.LOG_SHIFT_BY_SIZE, time.Second*1)
	log.LogDebug(1, "test log debug %d", 1)
	log.LogTrace(2, "test log trace %d", 2)

	//test log shift
	for i := 0; i < 100; i++ {
		log.LogDebug(2, "test log shift %d", i)
	}
	log.FlushLogBuffer(1)
	time.Sleep(time.Second * 2)
	//log.FlushLogBuffer(2)
	log.Close()
}

func main() {
	testLogger()
	testBufferLogger()
}
