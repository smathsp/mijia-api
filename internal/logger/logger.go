package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

// ANSI color codes.
const (
	colorDebug   = "\033[36m" // cyan
	colorInfo    = "\033[32m" // green
	colorWarning = "\033[33m" // yellow
	colorError   = "\033[31m" // red
	colorFatal   = "\033[1;31m" // bold red
	colorReset   = "\033[0m"
)

// Level represents a log level.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarning
	LevelError
	LevelFatal
)

var levelNames = [...]string{"DEBUG", "INFO", "WARNING", "ERROR", "FATAL"}
var levelColors = [...]string{colorDebug, colorInfo, colorWarning, colorError, colorFatal}

// Logger wraps the standard logger with level support and optional color.
type Logger struct {
	inner  *log.Logger
	level  Level
	useColor bool
	mu     sync.Mutex
}

var defaultLogger *Logger

func init() {
	useColor := isTerminal(os.Stderr)
	defaultLogger = &Logger{
		inner:   log.New(os.Stderr, "", 0),
		level:   LevelInfo,
		useColor: useColor,
	}
}

// SetLevel sets the log level for the default logger.
func SetLevel(l Level) {
	defaultLogger.mu.Lock()
	defer defaultLogger.mu.Unlock()
	defaultLogger.level = l
}

// SetLevelByName sets the log level by name string (DEBUG, INFO, WARNING, ERROR, FATAL).
func SetLevelByName(name string) {
	switch strings.ToUpper(name) {
	case "DEBUG":
		SetLevel(LevelDebug)
	case "INFO":
		SetLevel(LevelInfo)
	case "WARNING":
		SetLevel(LevelWarning)
	case "ERROR":
		SetLevel(LevelError)
	case "FATAL", "CRITICAL":
		SetLevel(LevelFatal)
	default:
		SetLevel(LevelInfo)
	}
}

// SetOutput sets the output destination for the default logger.
func SetOutput(w io.Writer) {
	defaultLogger.mu.Lock()
	defer defaultLogger.mu.Unlock()
	defaultLogger.inner.SetOutput(w)
}

func (l *Logger) output(level Level, msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if level < l.level {
		return
	}
	ts := formatTime()
	name := levelNames[level]
	if l.useColor {
		fmt.Fprintf(l.inner.Writer(), "%s%s - mijiaAPI - %s: %s%s\n",
			levelColors[level], ts, name, msg, colorReset)
	} else {
		fmt.Fprintf(l.inner.Writer(), "%s - mijiaAPI - %s: %s\n", ts, name, msg)
	}
}

// Debug logs a debug message.
func Debug(format string, args ...interface{}) {
	defaultLogger.output(LevelDebug, fmt.Sprintf(format, args...))
}

// Info logs an info message.
func Info(format string, args ...interface{}) {
	defaultLogger.output(LevelInfo, fmt.Sprintf(format, args...))
}

// Warning logs a warning message.
func Warning(format string, args ...interface{}) {
	defaultLogger.output(LevelWarning, fmt.Sprintf(format, args...))
}

// Error logs an error message.
func Error(format string, args ...interface{}) {
	defaultLogger.output(LevelError, fmt.Sprintf(format, args...))
}

// Fatal logs a fatal message and exits.
func Fatal(format string, args ...interface{}) {
	defaultLogger.output(LevelFatal, fmt.Sprintf(format, args...))
	os.Exit(1)
}
