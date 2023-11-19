package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type Config struct {
	Path       string `yaml:"path"`
	Name       string `yaml:"name"`
	Ext        string `yaml:"ext"`
	TimeFormat string `yaml:"timeFormat"`
}

type logLevel int

const (
	DEBUG logLevel = iota
	INFO
	WARNNING
	ERROR
	FATAL
)

const (
	flags              = log.LstdFlags
	defaultCallerDepth = 2
	bufferSize         = 1e5
)

type logEntry struct {
	msg   string
	level logLevel
}

var (
	levelFlags = []string{"DEBUG", "INFO", "WARNNING", "ERROR", "FATAL"}
)

type Logger struct {
	logFile   *os.File
	logger    *log.Logger
	entryChan chan *logEntry
	entryPool *sync.Pool
}

var DefaultLogger = NewStdoutLogger()

func NewStdoutLogger() *Logger {
	logger := &Logger{
		logFile:   nil,
		logger:    log.New(os.Stdout, "", flags),
		entryChan: make(chan *logEntry, bufferSize),
		entryPool: &sync.Pool{
			New: func() interface{} {
				return &logEntry{}
			},
		},
	}
	go func() {
		for e := range logger.entryChan {
			_ = logger.logger.Output(0, e.msg)
			logger.entryPool.Put(e)
		}
	}()

	return logger
}

func NewFileLogger(config *Config) (*Logger, error) {
	fileName := fmt.Sprintf("%s-%s.%s", config.Name, time.Now().Format(config.TimeFormat), config.Ext)

	logFile, err := mustOpen(fileName, config.Path)
	if err != nil {
		return nil, fmt.Errorf("logging.Join: %v", err)
	}

	mw := io.MultiWriter(os.Stdout, logFile)
	logger := &Logger{
		logFile:   logFile,
		logger:    log.New(mw, "", flags),
		entryChan: make(chan *logEntry, bufferSize),
		entryPool: &sync.Pool{
			New: func() interface{} {
				return &logEntry{}
			},
		},
	}

	go func() {
		for e := range logger.entryChan {
			logFilename := fmt.Sprintf("%s-%s.%s", config.Name, time.Now().Format(config.TimeFormat), config.Ext)
			if path.Join(config.Path, logFilename) != logger.logFile.Name() {
				logFile, err := mustOpen(logFilename, config.Path)
				if err != nil {
					panic("open log " + logFilename + " failed" + err.Error())
				}

				logger.logFile = logFile

				logger.logger = log.New(io.MultiWriter(os.Stdout, logFile), "", flags)
			}

			_ = logger.logger.Output(0, e.msg)
			logger.entryPool.Put(e)
		}
	}()

	return logger, nil
}

func Setup(config *Config) {
	logger, err := NewFileLogger(config)
	if err != nil {
		panic(err)
	}

	DefaultLogger = logger
}

func (logger *Logger) Output(level logLevel, callerDepth int, msg string) {
	var formattedMsg string

	_, file, line, ok := runtime.Caller(callerDepth)
	if ok {
		formattedMsg = fmt.Sprintf("[%s][%s:%d] %s", levelFlags[level], filepath.Base(file), line, msg)
	} else {
		formattedMsg = fmt.Sprintf("[%s] %s", levelFlags[level], msg)
	}

	entry, ok := logger.entryPool.Get().(*logEntry)
	if !ok {
		entry = &logEntry{}
	}

	entry.msg = formattedMsg
	entry.level = level
	logger.entryChan <- entry
}

func Debug(v ...interface{}) {
	msg := fmt.Sprintln(v...)
	DefaultLogger.Output(DEBUG, defaultCallerDepth, msg)
}

func Debugf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	DefaultLogger.Output(DEBUG, defaultCallerDepth, msg)
}

func Info(v ...interface{}) {
	msg := fmt.Sprintln(v...)
	DefaultLogger.Output(INFO, defaultCallerDepth, msg)
}

func Infof(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	DefaultLogger.Output(INFO, defaultCallerDepth, msg)
}

func Warn(v ...interface{}) {
	msg := fmt.Sprintln(v...)
	DefaultLogger.Output(WARNNING, defaultCallerDepth, msg)
}

func Warnf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	DefaultLogger.Output(WARNNING, defaultCallerDepth, msg)
}

func Error(v ...interface{}) {
	msg := fmt.Sprintln(v...)
	DefaultLogger.Output(ERROR, defaultCallerDepth, msg)
}

func Errorf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	DefaultLogger.Output(ERROR, defaultCallerDepth, msg)
}

func Fatal(v ...interface{}) {
	msg := fmt.Sprintln(v...)
	DefaultLogger.Output(FATAL, defaultCallerDepth, msg)
}
