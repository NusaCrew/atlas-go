package log

import (
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

var once sync.Once

var logger *Logger

type Logger struct {
	logrusEntry *logrus.Entry
	serviceName string
}

// Initialize initializes the package logger.
// It accepts an optional logrus.Formatter as the third parameter. If none is provided,
// it defaults to JSONFormatter. The variadic parameter keeps the function backward
// compatible with existing two-argument calls.
func Initialize(logLevel Level, serviceName string, formatterArgs ...logrus.Formatter) {
	once.Do(func() {
		log := logrus.New()

		log.SetLevel(mapToLogrusLevel(logLevel))

		// Choose formatter: use provided formatter if any, otherwise default to JSON
		var formatter logrus.Formatter
		if len(formatterArgs) > 0 && formatterArgs[0] != nil {
			formatter = formatterArgs[0]
		} else {
			formatter = &logrus.JSONFormatter{}
		}
		log.SetFormatter(formatter)
		log.SetOutput(os.Stdout)

		logger = &Logger{
			logrusEntry: logrus.NewEntry(log),
			serviceName: serviceName,
		}
	})
}

func newLogger(logrusEntry *logrus.Entry) *Logger {
	serviceName := ""
	if logger != nil {
		serviceName = logger.serviceName
	}
	return &Logger{
		logrusEntry: logrusEntry,
		serviceName: serviceName,
	}
}

func (l *Logger) WithField(key string, value any) *Logger {
	return newLogger(l.logrusEntry.WithField(key, value))
}

func (l *Logger) WithFields(args map[string]any) *Logger {
	return newLogger(l.logrusEntry.WithFields(args))
}

func (l *Logger) WithError(err error) *Logger {
	return newLogger(l.logrusEntry.WithError(err))
}

func (l *Logger) logF(severity Severity, msg string, args ...any) {
	if l == nil || l.logrusEntry == nil {
		switch severity {
		case DEBUG:
			logrus.Debugf(msg, args...)
		case INFO:
			logrus.Infof(msg, args...)
		case WARNING:
			logrus.Warnf(msg, args...)
		case ERROR:
			logrus.Errorf(msg, args...)
		default:
			logrus.Fatalf(msg, args...)
		}
		return
	}
	l.logrusEntry.WithFields(field{severity: severity, serviceName: l.serviceName, msg: msg, args: args}.toMap()).Log(mapSeverityToLogrusLevel(severity))
}

func (l *Logger) Debug(msg string, args ...any) {
	l.logF(DEBUG, msg, args...)
}

func (l *Logger) Info(msg string, args ...any) {
	l.logF(INFO, msg, args...)
}

func (l *Logger) Warning(msg string, args ...any) {
	l.logF(WARNING, msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.logF(ERROR, msg, args...)
}

func (l *Logger) Alert(msg string, args ...any) {
	l.logF(ALERT, msg, args...)
}

func (l *Logger) Panic(msg string, args ...any) {
	l.logF(PANIC, msg, args...)
}

// ------------ General Log Provider  ------------

func Debug(msg string, args ...any) {
	logger.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	logger.Info(msg, args...)
}

func Warning(msg string, args ...any) {
	logger.Warning(msg, args...)
}

func Error(msg string, args ...any) {
	logger.Error(msg, args...)
}

func Alert(msg string, args ...any) {
	logger.Alert(msg, args...)
}

func Panic(msg string, args ...any) {
	logger.Panic(msg, args...)
}

func WithField(key string, value any) *Logger {
	if logger == nil {
		return newLogger(logrus.NewEntry(logrus.StandardLogger()).WithField(key, value))
	}
	return logger.WithField(key, value)
}

func WithFields(args map[string]any) *Logger {
	if logger == nil {
		return newLogger(logrus.NewEntry(logrus.StandardLogger()).WithFields(args))
	}
	return logger.WithFields(args)
}

func WithError(err error) *Logger {
	if logger == nil {
		return newLogger(logrus.NewEntry(logrus.StandardLogger()).WithError(err))
	}
	return logger.WithError(err)
}
