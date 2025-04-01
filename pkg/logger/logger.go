package logger

import (
	"context"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.Logger
	once   sync.Once
)

// InitLogger initializes a singleton zap logger
func InitLogger(environment string) *zap.Logger {
	once.Do(func() {
		var err error
		var config zap.Config

		// Determine logging configuration based on environment
		switch environment {
		case "production":
			config = zap.NewProductionConfig()
			config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		case "development":
			config = zap.NewDevelopmentConfig()
			config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
			config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		default:
			config = zap.NewDevelopmentConfig()
		}

		// Customize log output
		config.OutputPaths = []string{"stdout"}
		config.ErrorOutputPaths = []string{"stderr"}

		// Build logger
		logger, err = config.Build(zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
		if err != nil {
			panic(err)
		}
	})

	return logger
}

// GetLogger returns the initialized logger
func GetLogger() *zap.Logger {
	if logger == nil {
		return InitLogger("development")
	}
	return logger
}

// GetSugaredLogger returns a sugared logger for easier logging
func GetSugaredLogger() *zap.SugaredLogger {
	return GetLogger().Sugar()
}

// LoggerMiddleware creates a middleware for Echo framework logging
func LoggerMiddleware(logger *zap.Logger) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Start timer
			start := time.Now()

			// Process request
			err := next(c)

			// Log request details
			fields := []zap.Field{
				zap.String("method", c.Request().Method),
				zap.String("path", c.Path()),
				zap.Int("status", c.Response().Status),
				zap.Duration("latency", time.Since(start)),
				zap.String("remote_ip", c.RealIP()),
			}

			// Determine log level based on status code
			switch {
			case c.Response().Status >= 500:
				logger.Error("Server error", fields...)
			case c.Response().Status >= 400:
				logger.Warn("Client error", fields...)
			default:
				logger.Info("Request processed", fields...)
			}

			return err
		}
	}
}

// ContextLogger adds contextual logging capabilities
type ContextLogger struct {
	logger *zap.Logger
}

// NewContextLogger creates a new contextual logger
func NewContextLogger(ctx context.Context) *ContextLogger {
	return &ContextLogger{
		logger: GetLogger(),
	}
}

// With adds fields to the logger
func (cl *ContextLogger) With(fields ...zap.Field) *ContextLogger {
	return &ContextLogger{
		logger: cl.logger.With(fields...),
	}
}

// Info logs an info message
func (cl *ContextLogger) Info(msg string, fields ...zap.Field) {
	cl.logger.Info(msg, fields...)
}

// Error logs an error message
func (cl *ContextLogger) Error(msg string, fields ...zap.Field) {
	cl.logger.Error(msg, fields...)
}

// Warn logs a warning message
func (cl *ContextLogger) Warn(msg string, fields ...zap.Field) {
	cl.logger.Warn(msg, fields...)
}

// Debug logs a debug message
func (cl *ContextLogger) Debug(msg string, fields ...zap.Field) {
	cl.logger.Debug(msg, fields...)
}
