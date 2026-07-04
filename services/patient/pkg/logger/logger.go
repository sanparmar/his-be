package logger

import (
	"context"
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger wraps zerolog for structured logging with PHI masking.
type Logger struct {
	logger zerolog.Logger
}

// New creates a new logger with specified level.
func New(level string, output io.Writer) *Logger {
	zeroLevel, _ := zerolog.ParseLevel(level)
	zlg := zerolog.New(output).
		Level(zeroLevel).
		With().
		Timestamp().
		Logger()

	return &Logger{logger: zlg}
}

// WithContext returns a new logger with context.
func (l *Logger) WithContext(ctx context.Context) *Logger {
	return &Logger{logger: l.logger.With().Ctx(ctx).Logger()}
}

// Info logs an info message.
func (l *Logger) Info(msg string, fields ...Field) {
	event := l.logger.Info()
	applyFields(event, fields)
	event.Msg(msg)
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string, fields ...Field) {
	event := l.logger.Warn()
	applyFields(event, fields)
	event.Msg(msg)
}

// Error logs an error message.
func (l *Logger) Error(msg string, err error, fields ...Field) {
	event := l.logger.Error().Err(err)
	applyFields(event, fields)
	event.Msg(msg)
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string, fields ...Field) {
	event := l.logger.Debug()
	applyFields(event, fields)
	event.Msg(msg)
}

// Field represents a structured log field.
type Field struct {
	Key   string
	Value interface{}
}

// String creates a string field.
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int creates an int field.
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// ErrorField creates an error field.
func ErrorField(err error) Field {
	return Field{Key: "error", Value: err}
}

// MaskPHI masks sensitive PHI fields (phone, email, DOB) for logging.
func MaskPHI(field Field) Field {
	switch field.Key {
	case "phone", "email", "dob":
		return Field{Key: field.Key, Value: "***MASKED***"}
	}
	return field
}

// applyFields applies fields to a zerolog event.
func applyFields(event *zerolog.Event, fields []Field) {
	for _, f := range fields {
		f = MaskPHI(f)
		event.Interface(f.Key, f.Value)
	}
}

// Global logger instance
var glog *Logger

// Init initializes the global logger.
func Init(level string) {
	glog = New(level, os.Stdout) // Use os.Stdout for dev/debugging
}

// Info logs an info message using the global logger.
func Info(msg string, fields ...Field) {
	if glog != nil {
		glog.Info(msg, fields...)
	} else {
		log.Info().Msg(msg)
	}
}

// Error logs an error message using the global logger.
func Error(msg string, err error, fields ...Field) {
	if glog != nil {
		glog.Error(msg, err, fields...)
	} else {
		log.Error().Err(err).Msg(msg)
	}
}
