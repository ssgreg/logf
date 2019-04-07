package logfc

import (
	"context"

	"github.com/ssgreg/logf"
)

// New returns a new context.Context with the given logger associated with it.
func New(parent context.Context, logger *logf.Logger) context.Context {
	return logf.NewContext(parent, logger)
}

// Get returns the logf.Logger associated with ctx and true
// or nil and false if no value is associated.
// Successive calls to Get returns the same result.
func Get(ctx context.Context) (logger *logf.Logger, ok bool) {
	logger = logf.FromContext(ctx)

	return logger, logger != nil
}

// MustGet returns the logf.Logger associated with ctx
// or panics if no value is associated.
// Successive calls to MustGet returns the same result.
func MustGet(ctx context.Context) *logf.Logger {
	logger, ok := Get(ctx)
	if !ok {
		panic("logfc: provided context has no logf.Logger associated")
	}

	return logger
}

// GetOrDisable returns the logf.Logger associated with ctx
// or logf.DisabledLogger() if no value is associated.
// Successive calls to GetOrDisable returns the same result.
func GetOrDisable(ctx context.Context) *logf.Logger {
	logger, ok := Get(ctx)
	if !ok {
		logger = logf.DisabledLogger()
	}

	return logger
}

// With returns a new context.Context with provided fields appended
// to fields of logf.Logger associated with ctx
// or returns unmodified ctx if no logf.Logger is associated with ctx.
func With(ctx context.Context, fields ...logf.Field) context.Context {
	logger, ok := Get(ctx)
	if !ok {
		return ctx
	}

	return New(ctx, logger.With(fields...))
}

// MustWith returns a new context.Context with provided fields appended
// to fields of logf.Logger associated with ctx
// or panics if no logf.Logger is associated with ctx.
func MustWith(ctx context.Context, fields ...logf.Field) context.Context {
	return New(ctx, MustGet(ctx).With(fields...))
}

// WithName returns a new context.Context with a new logf.Logger
// adding the given name to the original one
// or returns unmodified ctx if no logf.Logger is associated with ctx.
// Name separator is a period.
func WithName(ctx context.Context, name string) context.Context {
	logger, ok := Get(ctx)
	if !ok {
		return ctx
	}

	return New(ctx, logger.WithName(name))
}

// MustWithName returns a new context.Context with a new logf.Logger
// adding the given name to the original one
// or panics if no logf.Logger is associated with ctx.
// Name separator is a period.
func MustWithName(ctx context.Context, name string) context.Context {
	return New(ctx, MustGet(ctx).WithName(name))
}

// WithCaller returns a new context.Context with a new logf.Logger
// that adds a special annotation parameters
// to each logging message, such as the filename and line number of a caller
// or returns unmodified ctx if no logf.Logger is associated with ctx.
func WithCaller(ctx context.Context) context.Context {
	logger, ok := Get(ctx)
	if !ok {
		return ctx
	}

	return New(ctx, logger.WithCaller())
}

// MustWithCaller returns a new context.Context with a logf.Logger
// that adds a special annotation parameters
// to each logging message, such as the filename and line number of a caller
// or panics if no logf.Logger is associated with ctx.
func MustWithCaller(ctx context.Context) context.Context {
	return New(ctx, MustGet(ctx).WithCaller())
}

// WithCallerSkip returns a new context.Context with a new logf.Logger
// that adds a special annotation parameters with additional n skipped frames
// to each logging message, such as the filename and line number of a caller
// or returns unmodified ctx if no logf.Logger is associated with ctx.
func WithCallerSkip(ctx context.Context, n int) context.Context {
	logger, ok := Get(ctx)
	if !ok {
		return ctx
	}

	return New(ctx, logger.WithCallerSkip(n))
}

// MustWithCallerSkip returns a new context.Context with a logf.Logger
// that adds a special annotation parameters with additional n skipped frames
// to each logging message, such as the filename and line number of a caller
// or panics if no logf.Logger is associated with ctx.
func MustWithCallerSkip(ctx context.Context, n int) context.Context {
	return New(ctx, MustGet(ctx).WithCallerSkip(n))
}

// AtLevel calls the given fn if logging a message at the specified level
// is enabled, passing a logf.LogFunc with the bound level
// or does nothing if there is no logf.Logger associated with ctx.
func AtLevel(ctx context.Context, level logf.Level, fn func(logf.LogFunc)) {
	logger, ok := Get(ctx)
	if !ok {
		return
	}

	logger.AtLevel(level, fn)
}

// MustAtLevel calls the given fn if logging a message at the specified level
// is enabled, passing a logf.LogFunc with the bound level
// or panics if there is no logf.Logger associated with ctx.
func MustAtLevel(ctx context.Context, level logf.Level, fn func(logf.LogFunc)) {
	MustGet(ctx).AtLevel(level, fn)
}

// Debug logs a debug message with the given text, optional fields and
// fields passed to the logf.Logger using With function
// or logs nothing if no logf.Logger is associated with ctx.
func Debug(ctx context.Context, text string, fs ...logf.Field) {
	GetOrDisable(ctx).Debug(text, fs...)
}

// MustDebug logs a debug message with the given text, optional fields and
// fields passed to the logf.Logger using With function
// or panics if no logf.Logger is associated with ctx.
func MustDebug(ctx context.Context, text string, fs ...logf.Field) {
	MustGet(ctx).Debug(text, fs...)
}

// Info logs an info message with the given text, optional fields and
// fields passed to the logf.Logger using With function
// or logs nothing if no logf.Logger is associated with ctx.
func Info(ctx context.Context, text string, fs ...logf.Field) {
	GetOrDisable(ctx).Info(text, fs...)
}

// MustInfo logs an info message with the given text, optional fields and
// fields passed to the logf.Logger using With function
// or panics if no logf.Logger is associated with ctx.
func MustInfo(ctx context.Context, text string, fs ...logf.Field) {
	MustGet(ctx).Info(text, fs...)
}

// Warn logs a warning message with the given text, optional fields and
// fields passed to the logf.Logger using With function
// or logs nothing if no logf.Logger is associated with ctx.
func Warn(ctx context.Context, text string, fs ...logf.Field) {
	GetOrDisable(ctx).Warn(text, fs...)
}

// MustWarn logs a warning message with the given text, optional fields and
// fields passed to the logf.Logger using With function
// or panics if no logf.Logger is associated with ctx.
func MustWarn(ctx context.Context, text string, fs ...logf.Field) {
	GetOrDisable(ctx).Warn(text, fs...)
}

// Error logs an error message with the given text, optional fields and
// fields passed to the logf.Logger using With function
// or logs nothing if no logf.Logger is associated with ctx.
func Error(ctx context.Context, text string, fs ...logf.Field) {
	GetOrDisable(ctx).Error(text, fs...)
}

// MustError logs an error message with the given text, optional fields and
// fields passed to the logf.Logger using With function
// or panics if no logf.Logger is associated with ctx.
func MustError(ctx context.Context, text string, fs ...logf.Field) {
	MustGet(ctx).Error(text, fs...)
}
