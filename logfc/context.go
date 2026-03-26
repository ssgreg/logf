package logfc

import (
	"context"

	"github.com/ssgreg/logf/v2"
)

// New returns a new context.Context with the given logger associated with it.
func New(parent context.Context, logger *logf.Logger) context.Context {
	return logf.NewContext(parent, logger)
}

// Get returns the logf.Logger associated with ctx.
// Returns DisabledLogger() if no logger is associated.
func Get(ctx context.Context) *logf.Logger {
	return logf.FromContext(ctx)
}

// With returns a new context with the logger derived from ctx
// with the given additional fields.
func With(ctx context.Context, fields ...logf.Field) context.Context {
	return New(ctx, Get(ctx).With(fields...))
}

// WithName returns a new context with the logger derived from ctx
// with the given name appended.
func WithName(ctx context.Context, name string) context.Context {
	return New(ctx, Get(ctx).WithName(name))
}

// WithGroup returns a new context with the logger derived from ctx
// that nests all subsequent fields under the given group name.
func WithGroup(ctx context.Context, name string) context.Context {
	return New(ctx, Get(ctx).WithGroup(name))
}

// WithCaller returns a new context with the logger derived from ctx
// with caller reporting enabled.
func WithCaller(ctx context.Context) context.Context {
	return New(ctx, Get(ctx).WithCaller(true))
}

// WithCallerSkip returns a new context with the logger derived from ctx
// with increased number of skipped caller frames.
func WithCallerSkip(ctx context.Context, n int) context.Context {
	return New(ctx, Get(ctx).WithCallerSkip(n))
}

// AtLevel calls the given fn if logging at the specified level is enabled.
func AtLevel(ctx context.Context, level logf.Level, fn func(logf.LogFunc)) {
	Get(ctx).AtLevel(ctx, level, fn)
}

// Debug logs a debug message using the logger from ctx.
func Debug(ctx context.Context, text string, fs ...logf.Field) {
	logf.LogDepth(Get(ctx), ctx, 1, logf.LevelDebug, text, fs...)
}

// Info logs an info message using the logger from ctx.
func Info(ctx context.Context, text string, fs ...logf.Field) {
	logf.LogDepth(Get(ctx), ctx, 1, logf.LevelInfo, text, fs...)
}

// Warn logs a warning message using the logger from ctx.
func Warn(ctx context.Context, text string, fs ...logf.Field) {
	logf.LogDepth(Get(ctx), ctx, 1, logf.LevelWarn, text, fs...)
}

// Error logs an error message using the logger from ctx.
func Error(ctx context.Context, text string, fs ...logf.Field) {
	logf.LogDepth(Get(ctx), ctx, 1, logf.LevelError, text, fs...)
}

// Log logs a message at an arbitrary level using the logger from ctx.
func Log(ctx context.Context, level logf.Level, text string, fs ...logf.Field) {
	logf.LogDepth(Get(ctx), ctx, 1, level, text, fs...)
}
