package jsonlog

import (
	"encoding/json"
	"io"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

const (
	InfoLevel Level = iota
	ErrorLevel
	FatalLevel
	NilLevel
)

type Level int8

type Logger struct {
	out   io.Writer
	level Level
	mu    sync.Mutex
}

func (lv Level) String() string {
	switch lv {
	case InfoLevel:
		return "INFO"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	case NilLevel:
		return "NIL"
	default:
		return ""
	}
}

func New(out io.Writer, level Level) *Logger {
	return &Logger{
		out:   out,
		level: level,
		mu:    sync.Mutex{},
	}
}

func (l *Logger) Info(message string, properties map[string]string) {
	l.print(InfoLevel, message, properties)
}

func (l *Logger) Error(err error, properties map[string]string) {
	l.print(ErrorLevel, err.Error(), properties)
}

func (l *Logger) FatalErr(err error, properties map[string]string) {
	l.print(FatalLevel, err.Error(), properties)
	os.Exit(1)
}

func (l *Logger) print(level Level, message string, properties map[string]string) (int, error) {
	if l.level > level {
		return 0, nil
	}

	aux := struct {
		Level      string            `json:"level"`
		Time       string            `json:"time"`
		Message    string            `json:"message"`
		Properties map[string]string `json:"properties,omitempty"`
		Trace      string            `json:"trace,omitempty"`
	}{
		Level:      level.String(),
		Time:       time.Now().UTC().Format(time.RFC3339),
		Message:    message,
		Properties: properties,
	}

	if level >= ErrorLevel {
		aux.Trace = string(debug.Stack())
	}

	var line []byte
	line, err := json.Marshal(aux)
	if err != nil {
		line = []byte(ErrorLevel.String() + ": unable to marshal log message: " + err.Error())
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	return l.out.Write(append(line, '\n'))
}

func (l *Logger) Write(b []byte) (int, error) {
	return l.print(ErrorLevel, string(b), nil)
}
