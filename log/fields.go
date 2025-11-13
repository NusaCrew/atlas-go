package log

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

// --------------- ENUMERATIONS ---------------
type Severity int

const (
	DEBUG Severity = iota + 1
	INFO
	WARNING
	ERROR
	ALERT
	PANIC
	UNKNOWN_SEVERITY
)

func (s Severity) String() string {
	switch s {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	case ALERT:
		return "ALERT"
	case PANIC:
		return "PANIC"
	default:
		return ""
	}
}

type Code int

const (
	OK Code = iota + 1
	ClientError
	ServerError
	UnknownCode
)

func (c Code) String() string {
	switch c {
	case OK:
		return "OK"
	case ClientError:
		return "CLIENT_ERROR"
	case ServerError:
		return "SERVER_ERROR"
	default:
		return ""
	}
}

type Level string

const (
	LevelPanic Level = "PANIC"
	LevelFatal Level = "FATAL"
	LevelError Level = "ERROR"
	LevelWarn  Level = "WARN"
	LevelInfo  Level = "INFO"
	LevelDebug Level = "DEBUG"
	LevelTrace Level = "TRACE"
)

func mapToLogrusLevel(level Level) logrus.Level {
	switch level {
	case LevelPanic:
		return logrus.PanicLevel
	case LevelFatal:
		return logrus.FatalLevel
	case LevelError:
		return logrus.ErrorLevel
	case LevelWarn:
		return logrus.WarnLevel
	case LevelInfo:
		return logrus.InfoLevel
	case LevelDebug:
		return logrus.DebugLevel
	case LevelTrace:
		return logrus.TraceLevel
	default:
		return logrus.InfoLevel
	}
}

func mapSeverityToLogrusLevel(severity Severity) logrus.Level {
	switch severity {
	case DEBUG:
		return logrus.DebugLevel
	case INFO:
		return logrus.InfoLevel
	case WARNING:
		return logrus.WarnLevel
	case ERROR:
		return logrus.ErrorLevel
	case ALERT, PANIC:
		return logrus.FatalLevel
	default:
		return logrus.InfoLevel
	}
}

type Event int

const (
	Request Event = iota + 1
	Response
	UnknownEvent
)

func (e Event) String() string {
	if e <= 0 || e >= UnknownEvent {
		return ""
	}
	return [...]string{"request", "response"}[e-1]
}

// --------------- Log Fields ---------------

type field struct {
	severity    Severity
	code        Code
	event       Event
	methodName  string
	serviceName string
	uri         string
	err         error
	msg         string
	duration    int64
	args        []any
}

func (f field) GetMessage() string {
	if len(f.args) == 0 {
		return strings.TrimSpace(f.msg)
	}
	return strings.TrimSpace(fmt.Sprintf(f.msg, f.args...))
}

func (f field) toMap() map[string]any {
	resultMap := map[string]any{
		"severity": f.severity.String(),
		"code":     f.code.String(),
		"service":  f.serviceName,
		"message":  f.GetMessage(),
	}

	if f.methodName != "" {
		resultMap["method"] = f.methodName
	}
	if f.event > 0 && f.event < UnknownEvent {
		resultMap["event_type"] = f.event.String()
	}
	if f.uri != "" {
		resultMap["uri"] = f.uri
	}
	if f.err != nil {
		resultMap["error"] = f.err.Error()
	}
	if f.duration != 0 {
		resultMap["duration"] = f.duration
	}

	return resultMap
}
