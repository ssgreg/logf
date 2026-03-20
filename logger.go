package logf

import (
	"context"
	"log/slog"
	"time"
)

// New returns a Logger wired to the given Handler. Level filtering is
// controlled by the handler's Enabled method, so you can use any Handler
// implementation — SyncHandler, Router, ContextHandler, or your own.
// For a friendlier builder-style API, use NewLogger() instead.
func New(w Handler) *Logger {
	return &Logger{
		w:         w,
		addCaller: true,
	}
}

// DisabledLogger returns a Logger that silently discards everything as
// fast as possible. Handy as a safe default when no logger is configured,
// and it is what FromContext returns when there is no Logger in the context.
func DisabledLogger() *Logger {
	return defaultDisabledLogger
}

// Logger is the main entry point for structured logging. It wraps a Handler,
// checks levels before doing any work, and provides the familiar
// Debug/Info/Warn/Error methods plus context-aware field accumulation via
// With and WithGroup. Loggers are immutable — every With/WithName/WithGroup
// call returns a new Logger, so they are safe to share across goroutines.
type Logger struct {
	w Handler

	bag        *Bag
	name       string
	addCaller  bool
	callerSkip int
}

// Enabled reports whether logging at the given level would actually produce
// output. Use this to guard expensive argument preparation.
func (l *Logger) Enabled(ctx context.Context, lvl Level) bool {
	return l.w.Enabled(ctx, lvl)
}

// LogFunc is a logging function with a pre-bound level, used by AtLevel.
type LogFunc func(context.Context, string, ...Field)

// AtLevel calls fn only if the specified level is enabled, passing it a
// LogFunc pre-bound to that level. This is perfect for guarding expensive
// log preparation without a separate Enabled check.
func (l *Logger) AtLevel(ctx context.Context, lvl Level, fn func(LogFunc)) {
	if !l.w.Enabled(ctx, lvl) {
		return
	}

	fn(func(ctx context.Context, text string, fs ...Field) {
		l.write(ctx, 1, lvl, text, fs)
	})
}

// WithName returns a new Logger with the given name appended to the
// existing name, separated by a period. Names appear in log output
// as "parent.child" and are great for identifying subsystems.
// Loggers have no name by default.
func (l *Logger) WithName(n string) *Logger {
	if n == "" {
		return l
	}

	cc := l.clone()
	if cc.name == "" {
		cc.name = n
	} else {
		cc.name += "." + n
	}

	return cc
}

// WithCaller returns a new Logger with caller reporting toggled on or off.
// When enabled (the default), every log entry includes the source file and line.
func (l *Logger) WithCaller(enabled bool) *Logger {
	cc := l.clone()
	cc.addCaller = enabled

	return cc
}

// WithCallerSkip returns a new Logger that skips additional stack frames
// when capturing caller info. Use this when you wrap the Logger in your
// own helper function so the reported caller points to your caller, not
// your wrapper.
func (l *Logger) WithCallerSkip(skip int) *Logger {
	cc := l.clone()
	cc.callerSkip = skip

	return cc
}

// With returns a new Logger that includes the given fields in every
// subsequent log entry. Fields are accumulated, not replaced — so
// calling With multiple times builds up context over time.
func (l *Logger) With(fs ...Field) *Logger {
	cc := l.clone()
	cc.bag = l.bag.With(fs...)

	return cc
}

// Slog returns a *slog.Logger backed by the same Handler, fields, and
// name as this Logger. Use it when you need to hand a standard library
// slog.Logger to code that does not know about logf.
func (l *Logger) Slog() *slog.Logger {
	return slog.New(&slogHandler{
		w:         l.w,
		bag:       l.bag,
		name:      l.name,
		addCaller: l.addCaller,
	})
}

// WithGroup returns a new Logger that nests all subsequent fields — both
// from With and from per-call arguments — under the given group name.
// In JSON output this produces nested objects:
//
//	WithGroup("http") + Int("status", 200) → {"http":{"status":200}}
func (l *Logger) WithGroup(name string) *Logger {
	if name == "" {
		return l
	}

	cc := l.clone()
	cc.bag = l.bag.WithGroup(name)

	return cc
}

// Debug logs a message at LevelDebug. If debug logging is disabled, this
// is a no-op — no fields are evaluated, no allocations happen.
func (l *Logger) Debug(ctx context.Context, text string, fs ...Field) {
	if !l.w.Enabled(ctx, LevelDebug) {
		return
	}

	l.write(ctx, 1, LevelDebug, text, fs)
}

// Info logs a message at LevelInfo. This is the default "something
// happened" level for normal operational events.
func (l *Logger) Info(ctx context.Context, text string, fs ...Field) {
	if !l.w.Enabled(ctx, LevelInfo) {
		return
	}

	l.write(ctx, 1, LevelInfo, text, fs)
}

// Warn logs a message at LevelWarn. Use this for situations that are
// unexpected but not broken — things a human should probably look at.
func (l *Logger) Warn(ctx context.Context, text string, fs ...Field) {
	if !l.w.Enabled(ctx, LevelWarn) {
		return
	}

	l.write(ctx, 1, LevelWarn, text, fs)
}

// Error logs a message at LevelError. Something went wrong and you want
// everyone to know about it.
func (l *Logger) Error(ctx context.Context, text string, fs ...Field) {
	if !l.w.Enabled(ctx, LevelError) {
		return
	}

	l.write(ctx, 1, LevelError, text, fs)
}

// Log logs a message at an arbitrary level. Use this when the level is
// determined at runtime; for the common cases prefer Debug/Info/Warn/Error.
func (l *Logger) Log(ctx context.Context, lvl Level, text string, fs ...Field) {
	if !l.w.Enabled(ctx, lvl) {
		return
	}

	l.write(ctx, 1, lvl, text, fs)
}

func (l *Logger) write(ctx context.Context, extraSkip int, lv Level, text string, fs []Field) {
	e := Entry{
		LoggerBag:  l.bag,
		Fields:     fs,
		Level:      lv,
		Time:       time.Now(),
		LoggerName: l.name,
		Text:       text,
	}
	if l.addCaller {
		e.CallerPC = CallerPC(1 + l.callerSkip + extraSkip)
	}

	_ = l.w.Handle(ctx, e)
}

func (l *Logger) clone() *Logger {
	return &Logger{
		w:          l.w,
		bag:        l.bag,
		name:       l.name,
		addCaller:  l.addCaller,
		callerSkip: l.callerSkip,
	}
}

// LogDepth logs at the given level, adding depth extra frames to the caller
// skip count. It is a package-level function (not a method) so that wrapper
// packages like logfc can log through an existing Logger without allocating
// a new one on every call.
func LogDepth(l *Logger, ctx context.Context, depth int, lvl Level, text string, fs ...Field) {
	if !l.w.Enabled(ctx, lvl) {
		return
	}

	l.write(ctx, depth+1, lvl, text, fs)
}

// NewContext returns a new Context carrying the given Logger. Retrieve it
// later with FromContext. No more threading loggers through your entire
// call stack like some kind of dependency injection nightmare.
func NewContext(parent context.Context, logger *Logger) context.Context {
	return context.WithValue(parent, contextKeyLogger{}, logger)
}

// FromContext returns the Logger stored in the context by NewContext, or
// a DisabledLogger if no Logger was stored. It is always safe to call —
// you will never get nil.
func FromContext(ctx context.Context) *Logger {
	value := ctx.Value(contextKeyLogger{})
	if value == nil {
		return DisabledLogger()
	}

	return value.(*Logger)
}

type contextKeyLogger struct{}

var defaultDisabledLogger = New(nopHandler{})
