package log

import (
	"context"
	"fmt"
	"maps"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Tracer struct {
	ctx         context.Context
	startTime   time.Time
	methodName  string
	serviceName string
	uri         string
	fields      map[string]any
}

func NewTracer(ctx context.Context, methodName, serviceName string) *Tracer {
	return &Tracer{
		ctx:         ctx,
		startTime:   time.Now(),
		methodName:  methodName,
		serviceName: serviceName,
		fields:      make(map[string]any),
	}
}

func (t *Tracer) WithField(key string, value any) *Tracer {
	t.fields[key] = value
	return t
}

func (t *Tracer) WithFields(fields map[string]any) *Tracer {
	maps.Copy(t.fields, fields)
	return t
}

func (t *Tracer) TraceRequest(err error) {
	t.trace(Request, err)
}

func (t *Tracer) TraceResponse(err error) {
	t.trace(Response, err)
}

func (t *Tracer) Debug(code Code, event Event, msg string, args ...any) {
	t.logWithLevel(DEBUG, code, event, nil, msg, args...)
}

func (t *Tracer) Info(code Code, event Event, msg string, args ...any) {
	t.logWithLevel(INFO, code, event, nil, msg, args...)
}

func (t *Tracer) Warning(code Code, event Event, msg string, args ...any) {
	t.logWithLevel(WARNING, code, event, nil, msg, args...)
}

func (t *Tracer) Error(code Code, event Event, err error, msg string, args ...any) {
	t.logWithLevel(ERROR, code, event, err, msg, args...)
}

func (t *Tracer) Alert(code Code, event Event, err error, msg string, args ...any) {
	t.logWithLevel(ALERT, code, event, err, msg, args...)
}

func (t *Tracer) Panic(code Code, event Event, err error, msg string, args ...any) {
	t.logWithLevel(PANIC, code, event, err, msg, args...)
}

func (t *Tracer) logWithLevel(severity Severity, code Code, event Event, err error, msg string, args ...any) {
	duration := time.Since(t.startTime).Milliseconds()

	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}

	f := field{
		severity:    severity,
		code:        code,
		event:       event,
		serviceName: t.serviceName,
		methodName:  t.methodName,
		uri:         t.uri,
		duration:    duration,
		err:         err,
		msg:         msg,
	}

	logFields := f.toMap()
	maps.Copy(logFields, t.fields)

	if logger == nil {
		entry := WithFields(logFields)
		switch severity {
		case DEBUG:
			entry.Debug(msg, args...)
		case INFO:
			entry.Info(msg, args...)
		case WARNING:
			entry.Warning(msg, args...)
		case ERROR:
			entry.Error(msg, args...)
		case ALERT, PANIC:
			entry.Alert(msg, args...)
		}
		return
	}

	entry := logger.WithFields(logFields)
	switch severity {
	case DEBUG:
		entry.Debug(msg, args...)
	case INFO:
		entry.Info(msg, args...)
	case WARNING:
		entry.Warning(msg, args...)
	case ERROR:
		entry.Error(msg, args...)
	case ALERT, PANIC:
		entry.Alert(msg, args...)
	}
}

func (t *Tracer) trace(event Event, err error) {
	if err == nil {
		t.Info(OK, event, "%s OK response", t.methodName)
		return
	}

	code := determineErrorCode(err)
	errMsg := err.Error()

	// check for common error patterns
	errLower := strings.ToLower(errMsg)

	// client errors that are OK to log as INFO
	if strings.Contains(errLower, "not found") || strings.Contains(errLower, "notfound") {
		t.Info(OK, event, errMsg)
		return
	}

	// client errors
	if code == ClientError ||
		strings.Contains(errLower, "bad request") ||
		strings.Contains(errLower, "invalid argument") ||
		strings.Contains(errLower, "permission denied") ||
		strings.Contains(errLower, "unauthenticated") {
		t.Error(ClientError, event, err, errMsg)
		return
	}

	// server errors
	t.Error(ServerError, event, err, errMsg)
}

// determineErrorCode maps gRPC status codes to log Code enum.
func determineErrorCode(err error) Code {
	if err == nil {
		return OK
	}

	st, ok := status.FromError(err)
	if !ok {
		return ServerError
	}

	switch st.Code() {
	case codes.OK:
		return OK
	case codes.InvalidArgument, codes.NotFound, codes.AlreadyExists,
		codes.PermissionDenied, codes.FailedPrecondition, codes.Aborted,
		codes.OutOfRange, codes.Unauthenticated:
		return ClientError
	default:
		return ServerError
	}
}
