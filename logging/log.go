package logging

import (
	"fmt"
	"strings"
	"time"
)

type Logger interface {
	Println(args ...interface{})
	Printlnf(text string, args ...interface{})
	Printf(text string, args ...interface{})
	Close()
}

type consoleLogger struct{}
type nullLogger struct{}
type composableLogger struct {
	loggers []Logger
}
type timestampDecoratorLogger struct {
	logger    Logger
	onNewLine bool
}

type logEntryType int

const (
	let_Println logEntryType = iota
	let_Printlnf
	let_Printf
)

type logEntry struct {
	t    logEntryType
	text string
	args []interface{}
}

type synchronizingLogger struct {
	logChannel chan logEntry
	logger     Logger
}

func NewSynchronizingLoggerDecorator(logger Logger, bufferSize int) Logger {
	synchedLogger := &synchronizingLogger{logger: logger, logChannel: make(chan logEntry)}
	go startLogthread(synchedLogger)
	return synchedLogger
}

func NewTimestampLoggerDecorator(logger Logger) Logger {
	return &timestampDecoratorLogger{logger: logger, onNewLine: true}
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

func (log *consoleLogger) Println(args ...interface{}) {
	fmt.Println(args...)
}

func (log *consoleLogger) Printlnf(text string, args ...interface{}) {
	log.Println(fmt.Sprintf(text, args...))
}

func (log *consoleLogger) Printf(text string, args ...interface{}) {
	fmt.Printf(text, args...)
}

func (log *consoleLogger) Close() {
	// Can't close console
}

func (log *nullLogger) Println(args ...interface{}) {
	// A null logger does nothing
}

func (log *nullLogger) Printlnf(text string, args ...interface{}) {
	// A null logger does nothing
}

func (log *nullLogger) Printf(text string, args ...interface{}) {
	// A null logger does nothing
}

func (log *nullLogger) Close() {
	// A null logger does nothing
}

func (log *composableLogger) Println(args ...interface{}) {
	for _, v := range log.loggers {
		v.Println(args)
	}
}

func (log *composableLogger) Printlnf(text string, args ...interface{}) {
	formattedText := fmt.Sprintf(text, args...)
	log.Println(formattedText)
}

func (log *composableLogger) Printf(text string, args ...interface{}) {
	for _, v := range log.loggers {
		v.Printf(text, args...)
	}
}

func (log *composableLogger) Close() {
	for _, v := range log.loggers {
		v.Close()
	}
}

func splitNewLines(text string) []string {
	separator := func(c rune) bool {
		return c == '\n'
	}

	return strings.FieldsFunc(text, separator)
}

func newTimestamp() string {
	time := time.Now()

	milliSecond := time.Nanosecond() / (1000 * 1000)
	return fmt.Sprintf(
		"%04d-%02d-%02d %02d:%02d:%02d.%03d",
		time.Year(),
		time.Month(),
		time.Day(),
		time.Hour(),
		time.Minute(),
		time.Second(),
		milliSecond,
	)
}

func (log *timestampDecoratorLogger) Println(args ...interface{}) {
	lines := splitNewLines(fmt.Sprintln(args...))
	timestamp := newTimestamp()

	for _, line := range lines {
		if log.onNewLine {
			log.logger.Printf("[%s] ", timestamp)
		}
		log.logger.Println(line)
		log.onNewLine = true
	}
}

func (log *timestampDecoratorLogger) Printlnf(text string, args ...interface{}) {
	lines := splitNewLines(fmt.Sprintln(args...))
	timestamp := newTimestamp()

	for _, line := range lines {
		if log.onNewLine {
			log.logger.Printf("[%s] ", timestamp)
		}
		log.logger.Println(line)
		log.onNewLine = true
	}
}

func (log *timestampDecoratorLogger) Printf(text string, args ...interface{}) {
	formatted := fmt.Sprintln(args...)
	lastNewLineIndex := strings.LastIndex(formatted, "\n")
	timestamp := newTimestamp()

	if lastNewLineIndex < 0 {
		// If there is no newline in the formatted string...
		if log.onNewLine {
			// Make sure we emit a timestamp if we're on a new line
			log.logger.Printf("[%s] ", timestamp)
		}
		// ... we will just Printf!
		log.logger.Printf(formatted)
		// And now we know we're not on a new line!
		log.onNewLine = false
	} else {
		lines := splitNewLines(formatted)
		for i, line := range lines {
			if log.onNewLine {
				log.logger.Printf("[%s] ", timestamp)
			}

			if i < len(lines) {
				// We're on an intermediate line, so let's jus Println it...
				log.logger.Println(line)

				// ... and not that we now are on a new line!
				log.onNewLine = true
			} else {
				// We're printing the last line. Let's see if the original text actually ended with a newline!
				// If so, we need to Println and note that we have a new line indeed
				if lastNewLineIndex == len(formatted)-1 {
					log.logger.Println(line)
					log.onNewLine = true
				} else {
					log.logger.Printf(line)
					log.onNewLine = false
				}
			}
		}
	}
}

func (log *timestampDecoratorLogger) Close() {
	log.logger.Close()
	log.onNewLine = true
}

func (log *synchronizingLogger) Println(args ...interface{}) {
	log.logChannel <- logEntry{
		t:    let_Println,
		args: args,
	}
}

func (log *synchronizingLogger) Printlnf(text string, args ...interface{}) {
	log.logChannel <- logEntry{
		t:    let_Printlnf,
		text: text,
		args: args,
	}
}

func (log *synchronizingLogger) Printf(text string, args ...interface{}) {
	log.logChannel <- logEntry{
		t:    let_Printf,
		text: text,
		args: args,
	}
}

func (log *synchronizingLogger) Close() {
	close(log.logChannel)
}

func startLogthread(logger *synchronizingLogger) {
	for e := range logger.logChannel {
		switch e.t {
		case let_Printf:
			logger.logger.Printf(e.text, e.args...)
		case let_Println:
			logger.logger.Println(e.args...)
		case let_Printlnf:
			logger.logger.Printlnf(e.text, e.args...)
		}
	}
}
