package logf

import "context"

// NewContext returns new context with the provided Logger inside it.
func NewContext(parent context.Context, logger *Logger) context.Context {
	return context.WithValue(parent, contextKeyLogger{}, logger)
}

// FromContext returns Logger from context previously added to it with NewContext
// or nil if there is no Logger in context.
func FromContext(ctx context.Context) *Logger {
	value := ctx.Value(contextKeyLogger{})
	if value == nil {
		return nil
	}

	return value.(*Logger)
}

type contextKeyLogger struct{}
