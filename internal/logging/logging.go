package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// Logger represents a RedTriage logger
type Logger struct {
	logger zerolog.Logger
	level  zerolog.Level
	format string
	output io.Writer
}

// LogLevel represents the logging level
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// LogFormat represents the logging format
type LogFormat string

const (
	LogFormatText LogFormat = "text"
	LogFormatJSON LogFormat = "json"
)

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	return NewLoggerWithConfig(LogLevelInfo, LogFormatText, os.Stdout)
}

// NewLoggerWithConfig creates a new logger with specific configuration
func NewLoggerWithConfig(level LogLevel, format LogFormat, output io.Writer) *Logger {
	// Set global log level
	zerolog.SetGlobalLevel(parseLogLevel(level))
	
	// Create logger
	var logger zerolog.Logger
	
	switch format {
	case LogFormatJSON:
		logger = zerolog.New(output).With().Timestamp().Logger()
	default:
		// Text format with color
		output = zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: time.RFC3339,
			FormatLevel: func(i interface{}) string {
				if i == nil {
					return "????"
				}
				if ll, ok := i.(string); ok {
					switch ll {
					case "debug":
						return "\x1b[36mDBG\x1b[0m"
					case "info":
						return "\x1b[32mINF\x1b[0m"
					case "warn":
						return "\x1b[33mWRN\x1b[0m"
					case "error":
						return "\x1b[31mERR\x1b[0m"
					case "fatal":
						return "\x1b[31mFTL\x1b[0m"
					case "panic":
						return "\x1b[31mPNC\x1b[0m"
					default:
						return strings.ToUpper(ll)
					}
				}
				return strings.ToUpper(fmt.Sprintf("%v", i))
			},
			FormatMessage: func(i interface{}) string {
				if i == nil {
					return ""
				}
				return fmt.Sprintf("%s", i)
			},
			FormatFieldName: func(i interface{}) string {
				return fmt.Sprintf("\x1b[36m%s\x1b[0m=", i)
			},
			FormatFieldValue: func(i interface{}) string {
				return fmt.Sprintf("\x1b[32m%v\x1b[0m", i)
			},
		}
		logger = zerolog.New(output).With().Timestamp().Logger()
	}
	
	return &Logger{
		logger: logger,
		level:  parseLogLevel(level),
		format: string(format),
		output: output,
	}
}

// NewFileLogger creates a logger that writes to a file
func NewFileLogger(level LogLevel, format LogFormat, logPath string) (*Logger, error) {
	// Ensure log directory exists
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}
	
	// Open log file
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	
	// Create multi-writer for both file and console
	multiWriter := io.MultiWriter(os.Stdout, file)
	
	return NewLoggerWithConfig(level, format, multiWriter), nil
}

// parseLogLevel converts LogLevel to zerolog.Level
func parseLogLevel(level LogLevel) zerolog.Level {
	switch level {
	case LogLevelDebug:
		return zerolog.DebugLevel
	case LogLevelInfo:
		return zerolog.InfoLevel
	case LogLevelWarn:
		return zerolog.WarnLevel
	case LogLevelError:
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...map[string]interface{}) {
	if len(fields) > 0 {
		l.logger.Debug().Fields(fields[0]).Msg(msg)
	} else {
		l.logger.Debug().Msg(msg)
	}
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...map[string]interface{}) {
	if len(fields) > 0 {
		l.logger.Info().Fields(fields[0]).Msg(msg)
	} else {
		l.logger.Info().Msg(msg)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...map[string]interface{}) {
	if len(fields) > 0 {
		l.logger.Warn().Fields(fields[0]).Msg(msg)
	} else {
		l.logger.Warn().Msg(msg)
	}
}

// Error logs an error message
func (l *Logger) Error(msg string, fields ...map[string]interface{}) {
	if len(fields) > 0 {
		l.logger.Error().Fields(fields[0]).Msg(msg)
	} else {
		l.logger.Error().Msg(msg)
	}
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, fields ...map[string]interface{}) {
	if len(fields) > 0 {
		l.logger.Fatal().Fields(fields[0]).Msg(msg)
	} else {
		l.logger.Fatal().Msg(msg)
	}
}

// WithField adds a field to the logger
func (l *Logger) WithField(key string, value interface{}) *Logger {
	newLogger := l.logger.With().Interface(key, value).Logger()
	return &Logger{
		logger: newLogger,
		level:  l.level,
		format: l.format,
		output: l.output,
	}
}

// WithFields adds multiple fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	newLogger := l.logger.With().Fields(fields).Logger()
	return &Logger{
		logger: newLogger,
		level:  l.level,
		format: l.format,
		output: l.output,
	}
}

// WithError adds an error field to the logger
func (l *Logger) WithError(err error) *Logger {
	newLogger := l.logger.With().Err(err).Logger()
	return &Logger{
		logger: newLogger,
		level:  l.level,
		format: l.format,
		output: l.output,
	}
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = parseLogLevel(level)
	zerolog.SetGlobalLevel(l.level)
}

// GetLevel returns the current logging level
func (l *Logger) GetLevel() LogLevel {
	switch l.level {
	case zerolog.DebugLevel:
		return LogLevelDebug
	case zerolog.InfoLevel:
		return LogLevelInfo
	case zerolog.WarnLevel:
		return LogLevelWarn
	case zerolog.ErrorLevel:
		return LogLevelError
	default:
		return LogLevelInfo
	}
}

// IsLevelEnabled checks if a specific level is enabled
func (l *Logger) IsLevelEnabled(level LogLevel) bool {
	return parseLogLevel(level) >= l.level
}

// LogCommand logs command execution
func (l *Logger) LogCommand(cmd string, args []string, duration time.Duration, err error) {
	fields := map[string]interface{}{
		"command":  cmd,
		"args":     args,
		"duration": duration.String(),
	}
	
	if err != nil {
		fields["error"] = err.Error()
		l.Error("Command execution failed", fields)
	} else {
		l.Info("Command execution completed", fields)
	}
}

// LogArtifact logs artifact collection
func (l *Logger) LogArtifact(artifactType string, path string, size int64, duration time.Duration, err error) {
	fields := map[string]interface{}{
		"type":     artifactType,
		"path":     path,
		"size":     size,
		"duration": duration.String(),
	}
	
	if err != nil {
		fields["error"] = err.Error()
		l.Error("Artifact collection failed", fields)
	} else {
		l.Info("Artifact collected", fields)
	}
}

// LogDetection logs detection rule execution
func (l *Logger) LogDetection(ruleName string, severity string, matches int, duration time.Duration, err error) {
	fields := map[string]interface{}{
		"rule":     ruleName,
		"severity": severity,
		"matches":  matches,
		"duration": duration.String(),
	}
	
	if err != nil {
		fields["error"] = err.Error()
		l.Error("Detection rule execution failed", fields)
	} else {
		l.Info("Detection rule executed", fields)
	}
}

// LogSystem logs system information
func (l *Logger) LogSystem() {
	fields := map[string]interface{}{
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
		"go_version": runtime.Version(),
		"num_cpu":    runtime.NumCPU(),
	}
	
	l.Info("System information", fields)
}

// LogPerformance logs performance metrics
func (l *Logger) LogPerformance(operation string, duration time.Duration, memoryUsage uint64) {
	fields := map[string]interface{}{
		"operation":   operation,
		"duration":    duration.String(),
		"memory_mb":   memoryUsage / 1024 / 1024,
		"timestamp":   time.Now().UTC(),
	}
	
	l.Info("Performance metric", fields)
}

// Close closes the logger and any associated resources
func (l *Logger) Close() error {
	// For file-based loggers, we might need to close the file
	if closer, ok := l.output.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// Global logger instance
var globalLogger *Logger

// InitGlobalLogger initializes the global logger
func InitGlobalLogger(level LogLevel, format LogFormat) {
	globalLogger = NewLoggerWithConfig(level, format, os.Stdout)
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() *Logger {
	if globalLogger == nil {
		globalLogger = NewLogger()
	}
	return globalLogger
}

// Convenience functions for global logger
func Debug(msg string, fields ...map[string]interface{}) {
	GetGlobalLogger().Debug(msg, fields...)
}

func Info(msg string, fields ...map[string]interface{}) {
	GetGlobalLogger().Info(msg, fields...)
}

func Warn(msg string, fields ...map[string]interface{}) {
	GetGlobalLogger().Warn(msg, fields...)
}

func Error(msg string, fields ...map[string]interface{}) {
	GetGlobalLogger().Error(msg, fields...)
}

func Fatal(msg string, fields ...map[string]interface{}) {
	GetGlobalLogger().Fatal(msg, fields...)
}
