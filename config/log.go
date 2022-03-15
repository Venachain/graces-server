package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gggwvg/logrotate"
	"github.com/sirupsen/logrus"
)

var AllLevels = []Level{
	PanicLevel,
	FatalLevel,
	ErrorLevel,
	WarnLevel,
	InfoLevel,
	DebugLevel,
	TraceLevel,
}

var logrusPackage = "github.com/sirupsen/logrus"

const (
	// PanicLevel level, highest level of severity. Logs and then calls panic with the
	// message passed to Debug, Info, ...
	PanicLevel Level = iota
	// FatalLevel level. Logs and then calls `logger.Exit(1)`. It will exit even if the
	// logging level is set to Panic.
	FatalLevel
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	ErrorLevel
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	InfoLevel
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel
	// TraceLevel level. Designates finer-grained informational events than the Debug.
	TraceLevel
)

func MakeLogConfig() error {
	cfg := Config.LogConf
	logPath := cfg.Path
	if logPath == "" {
		return errors.New("unable to parse the log path from config")
	}
	err := os.MkdirAll(logPath, os.ModePerm)
	if err != nil {
		return errors.New("mkdir failed")
	}

	now := time.Now()
	fileDate := now.Format("2006-01-02")
	filename := fmt.Sprintf("%s%s.log", logPath, fileDate)
	opts := []logrotate.Option{
		logrotate.File(filename),
	}

	if cfg.Size == "" {
		return errors.New("unable to parse the log size from config")
	}
	opts = append(opts, logrotate.RotateSize(cfg.Size))
	logger, err := logrotate.NewLogger(opts...)
	if err != nil {
		return errors.New("new logger is error")
	}
	level := cfg.Level
	logrus.SetLevel(logrus.Level(level))
	logrus.SetOutput(os.Stdout)
	logrus.SetOutput(logger)
	writers := []io.Writer{
		logger,
		os.Stdout}
	fileAndStdoutWriter := io.MultiWriter(writers...)
	logrus.SetOutput(fileAndStdoutWriter)
	logrus.SetFormatter(new(LogFormatter))
	logger.Close()
	return nil
}

type LogFormatter struct{}

// Format 格式化日志输出
func (s *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := time.Now().Local().Format("2006-01-02 15:04:05")
	var codeFile string
	var length int

	pcs := make([]uintptr, 25)
	depth := runtime.Callers(4, pcs)
	frames := runtime.CallersFrames(pcs[:depth])
	var caller *runtime.Frame
	for f, again := frames.Next(); again; f, again = frames.Next() {
		pkg := getPackageName(f.Function)
		// If the caller isn't part of this package, we're done
		if pkg != logrusPackage {
			caller = &f //nolint:scopelint
			break
		}
	}
	if caller != nil {
		codeFile = filepath.Base(caller.File)
		length = caller.Line
	}
	// 日志输出格式
	msg := fmt.Sprintf("%s [%s] [%s:%d] %s\n", timestamp, strings.ToUpper(entry.Level.String()), codeFile, length, entry.Message)
	return []byte(msg), nil
}

func getPackageName(f string) string {
	for {
		lastPeriod := strings.LastIndex(f, ".")
		lastSlash := strings.LastIndex(f, "/")
		if lastPeriod > lastSlash {
			f = f[:lastPeriod]
		} else {
			break
		}
	}
	return f
}

func Info(format string, a ...interface{}) {
	logrus.Infof(format, a...)
}

func Warn(format string, a ...interface{}) {
	logrus.Warnf(format, a...)
}

func Debug(format string, a ...interface{}) {
	logrus.Debugf(format, a...)
}

func Error(format string, a ...interface{}) {
	logrus.Errorf(format, a...)
}

func Fatal(format string, a ...interface{}) {
	logrus.Fatalf(format, a...)
}

func Panic(format string, a ...interface{}) {
	logrus.Panicf(format, a...)
}
