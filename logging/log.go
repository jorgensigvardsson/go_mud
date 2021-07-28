package logging

import "fmt"

type Logger interface {
	Printf(text string, args ...interface{})
}

type consoleLogger struct{}
type nullLogger struct{}
type composableLogger struct {
	loggers []Logger
}

func NewConsoleLogger() Logger {
	return &consoleLogger{}
}

func NewNullLogger() Logger {
	return &nullLogger{}
}

func ComposeLogs(loggers ...Logger) Logger {
	return &composableLogger{loggers: loggers}
}

func (log *consoleLogger) Printf(text string, args ...interface{}) {
	fmt.Printf(text, args...)
}

func (log *nullLogger) Printf(text string, args ...interface{}) {
	// A null logger does nothing
}

func (log *composableLogger) Printf(text string, args ...interface{}) {
	for _, v := range log.loggers {
		v.Printf(text, args...)
	}
}
