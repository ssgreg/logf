package logfc

import (
	"context"

	"github.com/ssgreg/logf"
)

// New returns a new Context with the given Logger inside it.
// Avoid heavy using of NewContext and prefer using Context directly for better performance.
func New(parent context.Context, logger *logf.Logger) context.Context {
	return logf.NewContext(parent, logger)
}

// Get returns the Logger associated with this context or nil if
// no value is associated. Successive calls to FromContext returns the
// same result.
func Get(ctx context.Context) *logf.Logger {
	return logf.FromContext(ctx)
}

// With returns a new Context with provided fields appended to its fields.
func With(ctx context.Context, fields ...logf.Field) context.Context {
	return New(ctx, Get(ctx).With(fields...))
}

// Debug logs a debug message with the given text, optional fields and
// fields passed to the Logger using With function.
func Debug(ctx context.Context, text string, fs ...logf.Field) {
	Get(ctx).Debug(text, fs...)
}

// Info logs an info message with the given text, optional fields and
// fields passed to the Logger using With function.
func Info(ctx context.Context, text string, fs ...logf.Field) {
	Get(ctx).Info(text, fs...)
}

// Warn logs a warning message with the given text, optional fields and
// fields passed to the Logger using With function.
func Warn(ctx context.Context, text string, fs ...logf.Field) {
	Get(ctx).Warn(text, fs...)
}

// Error logs an error message with the given text, optional fields and
// fields passed to the Logger using With function.
func Error(ctx context.Context, text string, fs ...logf.Field) {
	Get(ctx).Error(text, fs...)
}
