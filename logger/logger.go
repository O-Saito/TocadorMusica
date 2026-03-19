package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	WithProfile(name string) Logger
}

type logger struct {
	output  io.Writer
	level   Level
	profile string
	mu      sync.Mutex
}

type Option func(*logger)

func ParseLevel(s string) Level {
	switch s {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn", "warning":
		return WARN
	case "error":
		return ERROR
	default:
		return DEBUG
	}
}

func parseLevel(s string) Level {
	return ParseLevel(s)
}

type filteredWriter struct {
	w            io.Writer
	consoleLevel Level
	mu           sync.Mutex
}

func (f *filteredWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	level := ERROR
	for _, lvl := range []Level{DEBUG, INFO, WARN, ERROR} {
		if bytes.Contains(p, []byte("["+lvl.String()+"]")) {
			level = lvl
			break
		}
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if level >= f.consoleLevel || level == ERROR {
		return f.w.Write(p)
	}
	return len(p), nil
}

func NewWithFile(profile string, fileLevel, consoleLevel Level, startTime string) (Logger, io.Closer, error) {
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logFile, err := os.Create(filepath.Join(logDir, fmt.Sprintf("%s_%s.log", profile, startTime)))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create log file: %w", err)
	}

	filteredConsole := &filteredWriter{
		w:            os.Stdout,
		consoleLevel: consoleLevel,
	}

	writer := &multiWriter{
		writers: []io.Writer{logFile, filteredConsole},
	}

	l := &logger{
		output:  writer,
		level:   fileLevel,
		profile: profile,
	}

	return l, logFile, nil
}

type multiWriter struct {
	writers []io.Writer
}

func (mw *multiWriter) Write(p []byte) (n int, err error) {
	for _, w := range mw.writers {
		if _, err := w.Write(p); err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

func WithOutput(w io.Writer) Option {
	return func(l *logger) {
		l.output = w
	}
}

func WithLevel(level string) Option {
	return func(l *logger) {
		l.level = parseLevel(level)
	}
}

func WithProfile(profile string) Option {
	return func(l *logger) {
		l.profile = profile
	}
}

func New(opts ...Option) Logger {
	l := &logger{
		output:  os.Stdout,
		level:   DEBUG,
		profile: "",
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

func (l *logger) log(level Level, msg string, args ...interface{}) {
	if level < l.level {
		return
	}

	_, file, line, _ := runtime.Caller(2)

	message := msg
	if len(args) > 0 {
		message = formatMessage(msg, args...)
	}

	profile := l.profile
	if profile == "" {
		profile = "-"
	}

	timestamp := time.Now().Format("2006-01-02T15:04:05")
	goroutines := runtime.NumGoroutine()

	format := "[%s] [goroutines=%d] [%s] [%s] [%s:%d] %s\n"
	logLine := fmt.Sprintf(format, timestamp, goroutines, level, profile, filepath.Base(file), line, message)

	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprint(l.output, logLine)
}

func formatMessage(msg string, args ...interface{}) string {
	if len(args) == 0 {
		return msg
	}

	result := msg
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			key, ok := args[i].(string)
			if ok {
				result += fmt.Sprintf(" %s=%v", key, args[i+1])
			} else {
				result += fmt.Sprintf(" %v", args[i+1])
			}
		}
	}
	return result
}

func (l *logger) Debug(msg string, args ...interface{}) {
	l.log(DEBUG, msg, args...)
}

func (l *logger) Info(msg string, args ...interface{}) {
	l.log(INFO, msg, args...)
}

func (l *logger) Warn(msg string, args ...interface{}) {
	l.log(WARN, msg, args...)
}

func (l *logger) Error(msg string, args ...interface{}) {
	l.log(ERROR, msg, args...)
}

func (l *logger) WithProfile(name string) Logger {
	return &logger{
		output:  l.output,
		level:   l.level,
		profile: name,
		mu:      l.mu,
	}
}
