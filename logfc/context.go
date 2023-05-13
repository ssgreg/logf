package logfc

import (
	"context"

	"github.com/ssgreg/logf"
)

// New returns a new context.Context with the given logger associated with it.
//
// Note: the given logger will hide any other logger which
// was associated with this context before.
func New(parent context.Context, logger *logf.Logger) context.Context {
	return logf.NewContext(parent, logger)
}

// Get returns the logf.Logger associated with ctx
// or log.DisabledLogger() if no value is associated.
// Successive calls to Get return the same result.
func Get(ctx context.Context) *logf.Logger {
	logger := logf.FromContext(ctx)
	if logger == nil {
		logger = logf.DisabledLogger()
	}

	return logger
}

// MustGet returns the logf.Logger associated with ctx
// or panics if no value is associated.
// Successive calls to MustGet return the same result.
func MustGet(ctx context.Context) *logf.Logger {
	logger := logf.FromContext(ctx)
	if logger == nil {
		panic("logfc: provided context has no logf.Logger associated")
	}

	return logger
}

// With returns a new context.Context with new logf.Logger inside
// with provided fields appended to current logf.Logger associated with ctx
// If there is no logf.Logger is associated with ctx, logf.DisabledLogger()
// is used as a base logger.
func With(ctx context.Context, fields ...logf.Field) context.Context {
	return New(ctx, Get(ctx).With(fields...))
}

// MustWith returns a new context.Context with provided fields appended
// to fields of logf.Logger associated with ctx
// or panics if no logf.Logger is associated with ctx.
func MustWith(ctx context.Context, fields ...logf.Field) context.Context {
	return New(ctx, MustGet(ctx).With(fields...))
}

// WithName returns a new context.Context with a new logf.Logger
// adding the given name to the original one.
// If there is no logf.Logger is associated with ctx, logf.DisabledLogger()
// is used as a base logger.
// Name separator is a period.
func WithName(ctx context.Context, name string) context.Context {
	return New(ctx, Get(ctx).WithName(name))
}

// MustWithName returns a new context.Context with a new logf.Logger
// adding the given name to the original one
// or panics if no logf.Logger is associated with ctx.
// Name separator is a period.
func MustWithName(ctx context.Context, name string) context.Context {
	return New(ctx, MustGet(ctx).WithName(name))
}

// WithLevel returns a new context.Context with a new logf.Logger
// with the given additional level checker.
// If there is no logf.Logger is associated with ctx, logf.DisabledLogger()
// is used as a base logger.
func WithLevel(ctx context.Context, level logf.LevelCheckerGetter) context.Context {
	return New(ctx, Get(ctx).WithLevel(level))
}

// MustWithLevel returns a new context.Context with a new logf.Logger
// with the given additional level checker.
// MustWithLevel panics if no logf.Logger is associated with ctx.
// Name separator is a period.
func MustWithLevel(ctx context.Context, level logf.LevelCheckerGetter) context.Context {
	return New(ctx, MustGet(ctx).WithLevel(level))
}

// WithCaller returns a new context.Context with a new logf.Logger
// that adds a special annotation parameters
// to each logging message, such as the filename and line number of a caller.
// If there is no logf.Logger is associated with ctx, logf.DisabledLogger()
// is used as a base logger.
func WithCaller(ctx context.Context) context.Context {
	return New(ctx, Get(ctx).WithCaller())
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
// to each logging message, such as the filename and line number of a caller.
// If there is no logf.Logger is associated with ctx, logf.DisabledLogger()
// is used as a base logger.
func WithCallerSkip(ctx context.Context, n int) context.Context {
	return New(ctx, Get(ctx).WithCallerSkip(n))
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
	Get(ctx).AtLevel(level, fn)
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
	Get(ctx).WithCallerSkip(1).Debug(text, fs...)
}

// MustDebug logs a debug message with the given text, optional fields and
// fields passed to the logf.Logger using With function
// or panics if no logf.Logger is associated with ctx.
func MustDebug(ctx context.Context, text string, fs ...logf.Field) {
	MustGet(ctx).WithCallerSkip(1).Debug(text, fs...)
}

// Info logs an info message with the given text, optional fields and
// fields passed to the logf.Logger using With function
// or logs nothing if no logf.Logger is associated with ctx.
func Info(ctx context.Context, text string, fs ...logf.Field) {
	Get(ctx).WithCallerSkip(1).Info(text, fs...)
}

// MustInfo logs an info message with the given text, optional fields and
// fields passed to the logf.Logger using With function
// or panics if no logf.Logger is associated with ctx.
func MustInfo(ctx context.Context, text string, fs ...logf.Field) {
	MustGet(ctx).WithCallerSkip(1).Info(text, fs...)
}

// Warn logs a warning message with the given text, optional fields and
// fields passed to the logf.Logger using With function
// or logs nothing if no logf.Logger is associated with ctx.
func Warn(ctx context.Context, text string, fs ...logf.Field) {
	Get(ctx).WithCallerSkip(1).Warn(text, fs...)
}

// MustWarn logs a warning message with the given text, optional fields and
// fields passed to the logf.Logger using With function
// or panics if no logf.Logger is associated with ctx.
func MustWarn(ctx context.Context, text string, fs ...logf.Field) {
	MustGet(ctx).WithCallerSkip(1).Warn(text, fs...)
}

// Error logs an error message with the given text, optional fields and
// fields passed to the logf.Logger using With function
// or logs nothing if no logf.Logger is associated with ctx.
func Error(ctx context.Context, text string, fs ...logf.Field) {
	Get(ctx).WithCallerSkip(1).Error(text, fs...)
}

// MustError logs an error message with the given text, optional fields and
// fields passed to the logf.Logger using With function
// or panics if no logf.Logger is associated with ctx.
func MustError(ctx context.Context, text string, fs ...logf.Field) {
	MustGet(ctx).WithCallerSkip(1).Error(text, fs...)
}
