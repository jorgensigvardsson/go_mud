package logging

import "fmt"

type Logger interface {
	WriteLine(text string)
	WriteLinef(text string, args ...interface{})
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

func (log *consoleLogger) WriteLine(text string) {
	fmt.Println(text)
}

func (log *consoleLogger) WriteLinef(text string, args ...interface{}) {
	log.WriteLine(fmt.Sprintf(text, args...))
}

func (log *consoleLogger) Printf(text string, args ...interface{}) {
	fmt.Printf(text, args...)
}

func (log *nullLogger) WriteLine(text string) {
	// A null logger does nothing
}

func (log *nullLogger) WriteLinef(text string, args ...interface{}) {
	// A null logger does nothing
}

func (log *nullLogger) Printf(text string, args ...interface{}) {
	// A null logger does nothing
}

func (log *composableLogger) WriteLine(text string) {
	for _, v := range log.loggers {
		v.WriteLine(text)
	}
}

func (log *composableLogger) WriteLinef(text string, args ...interface{}) {
	formattedText := fmt.Sprintf(text, args...)
	log.WriteLine(formattedText)
}

func (log *composableLogger) Printf(text string, args ...interface{}) {
	for _, v := range log.loggers {
		v.Printf(text, args...)
	}
}
