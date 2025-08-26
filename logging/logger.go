package logging

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

// Logger wraps logrus.Logger to provide a structured logging interface
type Logger struct {
	*logrus.Logger
}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	logger := logrus.New()
	
	// Set default level
	logger.SetLevel(logrus.InfoLevel)
	
	// Set default formatter
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	
	// Set default output
	logger.SetOutput(os.Stdout)
	
	return &Logger{Logger: logger}
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level string) {
	switch level {
	case "debug":
		l.Logger.SetLevel(logrus.DebugLevel)
	case "info":
		l.Logger.SetLevel(logrus.InfoLevel)
	case "warn":
		l.Logger.SetLevel(logrus.WarnLevel)
	case "error":
		l.Logger.SetLevel(logrus.ErrorLevel)
	case "fatal":
		l.Logger.SetLevel(logrus.FatalLevel)
	case "panic":
		l.Logger.SetLevel(logrus.PanicLevel)
	default:
		l.Logger.SetLevel(logrus.InfoLevel)
	}
}

// SetFormat sets the logging format
func (l *Logger) SetFormat(format string) {
	switch format {
	case "json":
		l.Logger.SetFormatter(&logrus.JSONFormatter{})
	case "text":
		l.Logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	default:
		l.Logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}
}

// SetOutput sets the logging output
func (l *Logger) SetOutput(output io.Writer) {
	l.Logger.SetOutput(output)
}

// WithField adds a field to the logger
func (l *Logger) WithField(key, value string) *logrus.Entry {
	return l.Logger.WithField(key, value)
}

// WithFields adds multiple fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *logrus.Entry {
	return l.Logger.WithFields(logrus.Fields(fields))
}

// Debug logs a debug message
func (l *Logger) Debug(args ...interface{}) {
	l.Logger.Debug(args...)
}

// Info logs an info message
func (l *Logger) Info(args ...interface{}) {
	l.Logger.Info(args...)
}

// Warn logs a warning message
func (l *Logger) Warn(args ...interface{}) {
	l.Logger.Warn(args...)
}

// Error logs an error message
func (l *Logger) Error(args ...interface{}) {
	l.Logger.Error(args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(args ...interface{}) {
	l.Logger.Fatal(args...)
}

// Panic logs a panic message and panics
func (l *Logger) Panic(args ...interface{}) {
	l.Logger.Panic(args...)
}
