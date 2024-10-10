package ptah_agent

import (
	"context"
	"log/slog"
)

type contextKey uint8

const loggerKey contextKey = iota

func Logger(ctx context.Context) *slog.Logger {
	return ctx.Value(loggerKey).(*slog.Logger)
}

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func ContextWithLoggerValues(ctx context.Context, args ...any) (context.Context, *slog.Logger) {
	logger := Logger(ctx).With(args...)

	return WithLogger(ctx, logger), logger
}
