package uci

import (
	"fmt"
	"io"
	"os"
)

type Logger interface {
	Printf(msg string, args ...any)
}

func NewNullLogger() Logger {
	return nullLogger{}
}

func NewSimpleLogger(w io.Writer) Logger {
	return simpleLogger{w: w}
}

func NewStdoutLogger() Logger {
	return simpleLogger{w: os.Stdout}
}

func NewStderrLogger() Logger {
	return simpleLogger{w: os.Stderr}
}

type nullLogger struct{}

func (nullLogger) Printf(string, ...any) {}

type simpleLogger struct {
	w io.Writer
}

func (l simpleLogger) Printf(msg string, args ...any) {
	_, _ = fmt.Fprintf(l.w, msg+"\n", args...)
}
