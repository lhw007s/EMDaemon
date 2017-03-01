package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	INFO = iota + 1
	DEBUG
	ERR
	CRI
	MAX_LOG_TYPE
)

type ILogger interface {
	Init()
	createLoggerInstance(logType int)
	WriteLog(logType int, msg string)
	findLogInstance(logType int) *log.Logger
	doDailyChangeLoggerFile(logger *Logger)
	MakeLoggerFile(logType int) *Logger
}

type Logger struct {
	Type               int
	instance           *log.Logger
	fd                 *os.File
	nextCreateFileTime int64
	mutex              *sync.Mutex
}

type LoggerManager struct {
	//logInstPool sync.Pool
	debug *Logger
	info  *Logger
	cri   *Logger
	error *Logger
}

func NewLogManager() *LoggerManager {
	return &LoggerManager{
		&Logger{0, new(log.Logger), new(os.File), 0, new(sync.Mutex)},
		&Logger{0, new(log.Logger), new(os.File), 0, new(sync.Mutex)},
		&Logger{0, new(log.Logger), new(os.File), 0, new(sync.Mutex)},
		&Logger{0, new(log.Logger), new(os.File), 0, new(sync.Mutex)},
	}
}

func (manager *LoggerManager) Init() {
	manager.MakeLoggerFile(INFO)
	manager.MakeLoggerFile(DEBUG)
	manager.MakeLoggerFile(ERR)
	manager.MakeLoggerFile(CRI)
}

func (manager *LoggerManager) MakeLoggerFile(logType int) *Logger {

	var logger *Logger
	switch logType {
	case INFO:
		logger = manager.info
	case DEBUG:
		logger = manager.debug
	case ERR:
		logger = manager.error
	case CRI:
		logger = manager.cri
	}

	logger.Type = logType

	manager.createLogInstance(logger)
	manager.createLogFile(logger, logType)

	currentTime := time.Now()
	NextDateTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day()+1, currentTime.Hour(), 0, 0, 0, time.UTC)
	logger.nextCreateFileTime = NextDateTime.Unix()

	return logger
}

func (manager *LoggerManager) createLogInstance(logger *Logger) {
	var buf bytes.Buffer
	logger.instance = log.New(&buf, "", log.Ldate|log.Ltime)
	//fd := createLogFile(logType)
}

func (manager *LoggerManager) createLogFile(logger *Logger, logType int) {

	//env 파일에서내용을 참조한다.
	var logPath string = GEnvManager.Log["log"]["path"]
	var fileName string
	switch logType {
	case INFO:
		fileName = GEnvManager.Log["log"]["info_filename"]
	case DEBUG:
		fileName = GEnvManager.Log["log"]["debug_filename"]
	case ERR:
		fileName = GEnvManager.Log["log"]["err_filename"]
	case CRI:
		fileName = GEnvManager.Log["log"]["cri_filename"]
	}

	//Log%Y%M%D_debug.log
	currentTime := time.Now()
	strDate, err := NewTimeManager().TimeToStr("YMD", &currentTime)
	if err != nil {
		panic("로그파일 생성 오류")
	}

	fileName = strings.Replace(fileName, "%Y%M%D", strDate, -1)

	var fullSrc string = logPath + "/" + fileName
	fd, err := os.OpenFile(fullSrc, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		panic("로그파일 생성 오류")
	}

	logger.instance.SetOutput(fd)
}

func (manager *LoggerManager) WriteLog(logType int, format string, msg ...interface{}) {

	logger := manager.findLogInstance(logType)

	logger.mutex.Lock()
	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	_, line := f.FileLine(pc[0])

	if format == "" {
		format = " %s "
	}

	logMsg := fmt.Sprintf(": %s [line:%d] ", f.Name(), line)
	logMsg = logMsg + fmt.Sprintf(format, msg...)
	logger.instance.Print(logMsg)
	logger.mutex.Unlock()
}

func (manager *LoggerManager) Debug(format string, msg ...interface{}) {
	manager.WriteLog(DEBUG, format, msg...)
}

func (manager *LoggerManager) Cri(format string, msg ...interface{}) {
	manager.WriteLog(CRI, format, msg...)
}

func (manager *LoggerManager) Error(format string, msg ...interface{}) {
	manager.WriteLog(ERR, format, msg...)
}

func (manager *LoggerManager) Info(format string, msg ...interface{}) {
	manager.WriteLog(INFO, format, msg...)
}

func (manager *LoggerManager) doDailyChangeLoggerFile(logger *Logger) {
	if logger.nextCreateFileTime <= time.Now().Unix() {
		manager.MakeLoggerFile(logger.Type)
	}
}

func (manager *LoggerManager) findLogInstance(logType int) *Logger {
	var logger *Logger
	switch logType {
	case INFO:
		logger = manager.info
	case DEBUG:
		logger = manager.debug
	case ERR:
		logger = manager.error
	case CRI:
		logger = manager.cri
	}

	manager.doDailyChangeLoggerFile(logger)

	return logger
}
