package log

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSeverity_String(t *testing.T) {
	tests := []struct {
		name string
		s    Severity
		want string
	}{
		{name: "debug", s: DEBUG, want: "DEBUG"},
		{name: "info", s: INFO, want: "INFO"},
		{name: "warning", s: WARNING, want: "WARNING"},
		{name: "error", s: ERROR, want: "ERROR"},
		{name: "alert", s: ALERT, want: "ALERT"},
		{name: "panic", s: PANIC, want: "PANIC"},
		{name: "unknown in-range", s: UNKNOWN_SEVERITY, want: ""},
		{name: "zero", s: 0, want: ""},
		{name: "out of range", s: Severity(999), want: ""},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := tc.s.String()
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestCode_String(t *testing.T) {
	tests := []struct {
		name string
		c    Code
		want string
	}{
		{name: "ok", c: OK, want: "OK"},
		{name: "client error", c: ClientError, want: "CLIENT_ERROR"},
		{name: "server error", c: ServerError, want: "SERVER_ERROR"},
		{name: "unknown in-range", c: UnknownCode, want: ""},
		{name: "zero", c: 0, want: ""},
		{name: "out of range", c: Code(999), want: ""},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := tc.c.String()
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestMapToLogrusLevel(t *testing.T) {
	tests := []struct {
		name  string
		level Level
		want  logrus.Level
	}{
		{name: "panic", level: LevelPanic, want: logrus.PanicLevel},
		{name: "fatal", level: LevelFatal, want: logrus.FatalLevel},
		{name: "error", level: LevelError, want: logrus.ErrorLevel},
		{name: "warn", level: LevelWarn, want: logrus.WarnLevel},
		{name: "info", level: LevelInfo, want: logrus.InfoLevel},
		{name: "debug", level: LevelDebug, want: logrus.DebugLevel},
		{name: "trace", level: LevelTrace, want: logrus.TraceLevel},
		{name: "default unknown", level: Level("FOO"), want: logrus.InfoLevel},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := mapToLogrusLevel(tc.level)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestField_ToMap(t *testing.T) {
	tests := []struct {
		name string
		f    field
		want map[string]any
	}{
		{
			name: "basic fields",
			f: field{
				severity:    INFO,
				code:        OK,
				serviceName: "test-service",
				methodName:  "test-method",
				msg:         "test message",
			},
			want: map[string]any{
				"severity":    "INFO",
				"code":        "OK",
				"service":     "test-service",
				"method_name": "test-method",
				"message":     "test message",
			},
		},
		{
			name: "with uri",
			f: field{
				severity:    ERROR,
				code:        ClientError,
				serviceName: "test-service",
				methodName:  "test-method",
				uri:         "/test",
				msg:         "error message",
			},
			want: map[string]any{
				"severity":    "ERROR",
				"code":        "CLIENT_ERROR",
				"service":     "test-service",
				"method_name": "test-method",
				"uri":         "/test",
				"message":     "error message",
			},
		},
		{
			name: "with error",
			f: field{
				severity:    ERROR,
				code:        ServerError,
				serviceName: "test-service",
				methodName:  "test-method",
				err:         assert.AnError,
				msg:         "error message",
			},
			want: map[string]any{
				"severity":    "ERROR",
				"code":        "SERVER_ERROR",
				"service":     "test-service",
				"method_name": "test-method",
				"error":       assert.AnError.Error(),
				"message":     "error message",
			},
		},
		{
			name: "with durations",
			f: field{
				severity:    DEBUG,
				code:        OK,
				serviceName: "test-service",
				methodName:  "test-method",
				duration:    100,
				msg:         "debug message",
			},
			want: map[string]any{
				"severity":    "DEBUG",
				"code":        "OK",
				"service":     "test-service",
				"method_name": "test-method",
				"durations":   int64(100),
				"message":     "debug message",
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := tc.f.toMap()
			assert.Equal(t, tc.want, got)
		})
	}
}
