package errors

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// LogLevel represents the severity of log messages
type LogLevel int

const (
	// Log levels
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

// Logger provides structured logging capabilities
type Logger struct {
	level     LogLevel
	requestID string
	fields    map[string]interface{}
}

// DefaultLogger is used when not provided by context
var DefaultLogger = &Logger{
	level:  LevelInfo,
	fields: make(map[string]interface{}),
}

// FromContext creates a new logger with context from a Fiber request
func FromContext(c *fiber.Ctx) *Logger {
	requestID, _ := c.Locals("requestID").(string)
	if requestID == "" {
		requestID = "unknown"
	}

	return &Logger{
		level:     LevelInfo,
		requestID: requestID,
		fields:    make(map[string]interface{}),
	}
}

// WithField adds a field to the log output
func (l *Logger) WithField(key string, value interface{}) *Logger {
	newLogger := &Logger{
		level:     l.level,
		requestID: l.requestID,
		fields:    make(map[string]interface{}),
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add new field
	newLogger.fields[key] = value
	return newLogger
}

// WithFields adds multiple fields to the log output
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	newLogger := &Logger{
		level:     l.level,
		requestID: l.requestID,
		fields:    make(map[string]interface{}),
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add new fields
	for k, v := range fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

// format creates a formatted log message
func (l *Logger) format(level, msg string) string {
	// Format fields
	var fieldStr string
	if len(l.fields) > 0 {
		parts := make([]string, 0, len(l.fields))
		for k, v := range l.fields {
			parts = append(parts, fmt.Sprintf("%s=%v", k, v))
		}
		fieldStr = " " + strings.Join(parts, " ")
	}

	// Format with timestamp, request ID, level, message, and fields
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	return fmt.Sprintf("[%s] [%s] [%s] %s%s", timestamp, l.requestID, level, msg, fieldStr)
}

// Debug logs debug level messages
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level <= LevelDebug {
		msg := fmt.Sprintf(format, args...)
		log.Println(l.format("DEBUG", msg))
	}
}

// Info logs info level messages
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= LevelInfo {
		msg := fmt.Sprintf(format, args...)
		log.Println(l.format("INFO", msg))
	}
}

// Warn logs warning level messages
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level <= LevelWarn {
		msg := fmt.Sprintf(format, args...)
		log.Println(l.format("WARN", msg))
	}
}

// Error logs error level messages
func (l *Logger) Error(format string, args ...interface{}) {
	if l.level <= LevelError {
		msg := fmt.Sprintf(format, args...)
		log.Println(l.format("ERROR", msg))
	}
}

// Fatal logs error and exits the application
func (l *Logger) Fatal(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Println(l.format("FATAL", msg))
	os.Exit(1)
}

// GetRequestLogger is a helper to get a logger from a Fiber context
func GetRequestLogger(c *fiber.Ctx) *Logger {
	return FromContext(c)
}
