package logf

import (
	"context"
	"time"
)

// NewLogger returns a new Logger with the given EntryWriter.
// Level filtering is controlled by the writer's Enabled method.
func NewLogger(w EntryWriter) *Logger {
	return &Logger{
		w:         w,
		addCaller: true,
	}
}

// NewDisabledLogger return a new Logger that logs nothing as fast as
// possible.
func NewDisabledLogger() *Logger {
	return NewLogger(nopWriter{})
}

var defaultDisabledLogger = NewDisabledLogger()

// DisabledLogger returns a default instance of a Logger that logs nothing
// as fast as possible.
func DisabledLogger() *Logger {
	return defaultDisabledLogger
}

// Logger is the fast, asynchronous, structured logger.
//
// The Logger wraps EntryWriter to check logging level and provide a bit of
// syntactic sugar.
type Logger struct {
	w EntryWriter

	bag        *Bag
	name       string
	addCaller  bool
	callerSkip int
}

// LogFunc allows to log a message with a bound level.
type LogFunc func(context.Context, string, ...Field)

// Enabled reports whether logging at the given level is enabled.
func (l *Logger) Enabled(ctx context.Context, lvl Level) bool {
	return l.w.Enabled(ctx, lvl)
}

// AtLevel calls the given fn if logging a message at the specified level
// is enabled, passing a LogFunc with the bound level.
func (l *Logger) AtLevel(ctx context.Context, lvl Level, fn func(LogFunc)) {
	if !l.w.Enabled(ctx, lvl) {
		return
	}

	fn(func(ctx context.Context, text string, fs ...Field) {
		l.write(ctx, 1, lvl, text, fs)
	})
}

// WithName returns a new Logger adding the given name to the calling one.
// Name separator is a period.
//
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

// WithCaller returns a new Logger with caller reporting enabled or disabled.
func (l *Logger) WithCaller(enabled bool) *Logger {
	cc := l.clone()
	cc.addCaller = enabled

	return cc
}

// WithCallerSkip returns a new Logger with increased number of skipped
// frames. It's usable to build a custom wrapper for the Logger.
func (l *Logger) WithCallerSkip(skip int) *Logger {
	cc := l.clone()
	cc.callerSkip = skip

	return cc
}

// With returns a new Logger with the given additional fields.
func (l *Logger) With(fs ...Field) *Logger {
	cc := l.clone()
	cc.bag = l.bag.With(fs...)

	return cc
}

// WithGroup returns a new Logger that nests all subsequent fields
// (from With and per-call) under the given group name.
// Produces nested JSON objects: WithGroup("http") + Int("status", 200)
// → {"http":{"status":200}}.
func (l *Logger) WithGroup(name string) *Logger {
	if name == "" {
		return l
	}

	cc := l.clone()
	cc.bag = l.bag.WithGroup(name)

	return cc
}

// Debug logs a debug message with the given text, optional fields and
// fields passed to the Logger using With function.
func (l *Logger) Debug(ctx context.Context, text string, fs ...Field) {
	if !l.w.Enabled(ctx, LevelDebug) {
		return
	}

	l.write(ctx, 1, LevelDebug, text, fs)
}

// Info logs an info message with the given text, optional fields and
// fields passed to the Logger using With function.
func (l *Logger) Info(ctx context.Context, text string, fs ...Field) {
	if !l.w.Enabled(ctx, LevelInfo) {
		return
	}

	l.write(ctx, 1, LevelInfo, text, fs)
}

// Warn logs a warning message with the given text, optional fields and
// fields passed to the Logger using With function.
func (l *Logger) Warn(ctx context.Context, text string, fs ...Field) {
	if !l.w.Enabled(ctx, LevelWarn) {
		return
	}

	l.write(ctx, 1, LevelWarn, text, fs)
}

// Error logs an error message with the given text, optional fields and
// fields passed to the Logger using With function.
func (l *Logger) Error(ctx context.Context, text string, fs ...Field) {
	if !l.w.Enabled(ctx, LevelError) {
		return
	}

	l.write(ctx, 1, LevelError, text, fs)
}

// Log logs a message at the given level.
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

	_ = l.w.WriteEntry(ctx, e)
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

// LogDepth logs using the given logger at the specified level, adding depth
// extra frames to the caller skip count. It is intended for wrapper packages
// like logfc to avoid Logger allocation on each call.
func LogDepth(l *Logger, ctx context.Context, depth int, lvl Level, text string, fs ...Field) {
	if !l.w.Enabled(ctx, lvl) {
		return
	}

	l.write(ctx, depth+1, lvl, text, fs)
}

// NewContext returns a new Context with the given Logger inside it.
func NewContext(parent context.Context, logger *Logger) context.Context {
	return context.WithValue(parent, contextKeyLogger{}, logger)
}

// FromContext returns the Logger associated with this context or
// DisabledLogger() if no value is associated.
func FromContext(ctx context.Context) *Logger {
	value := ctx.Value(contextKeyLogger{})
	if value == nil {
		return DisabledLogger()
	}

	return value.(*Logger)
}

type contextKeyLogger struct{}
