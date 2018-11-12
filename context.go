package logf

import "context"

// NewContext returns a new Context with the given Logger inside it.
func NewContext(parent context.Context, logger *Logger) context.Context {
	return context.WithValue(parent, contextKeyLogger{}, logger)
}

// FromContext returns the Logger associated with this context or nil if
// no value is associated. Successive calls to FromContext returns the
// same result.
func FromContext(ctx context.Context) *Logger {
	value := ctx.Value(contextKeyLogger{})
	if value == nil {
		return nil
	}

	return value.(*Logger)
}

type contextKeyLogger struct{}
