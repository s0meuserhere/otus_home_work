package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

const localEnv = "local"

type ctxKey struct{}

func WithContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, logger)
}

func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

type Logger struct {
	logger *slog.Logger
}

func New(level string, env string) *Logger {
	var l slog.Level

	switch strings.ToLower(level) {
	case "debug":
		l = slog.LevelDebug
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	case "info":
		l = slog.LevelInfo
	default:
		l = slog.LevelInfo
	}

	var slogger *slog.Logger
	switch env {
	case localEnv:
		slogger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level:     l,
			AddSource: true,
		}))
	default:
		slogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     l,
			AddSource: true,
		}))
	}

	slogger.With("env", env).With("level", level).Info("logger initialized")

	return &Logger{slogger}
}

func (l *Logger) Slog() *slog.Logger {
	return l.logger
}

func (l *Logger) Info(msg string, args ...interface{}) {
	l.logger.Info(msg, args...)
}

func (l *Logger) Error(msg string, args ...interface{}) {
	l.logger.Error(msg, args...)
}

func (l *Logger) Debug(msg string, args ...interface{}) {
	l.logger.Debug(msg, args...)
}

func (l *Logger) Warn(msg string, args ...interface{}) {
	l.logger.Warn(msg, args...)
}
