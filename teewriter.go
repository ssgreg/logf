package logf

import (
	"context"
	"errors"
)

// NewTeeWriter returns an Handler that fans out Handle calls to all
// provided writers. Enabled returns true if any writer is enabled.
// Errors from individual writers are joined via errors.Join.
func NewTeeWriter(writers ...Handler) Handler {
	switch len(writers) {
	case 0:
		return nopHandler{}
	case 1:
		return writers[0]
	default:
		return &teeWriter{writers: append([]Handler{}, writers...)}
	}
}

type teeWriter struct {
	writers []Handler
}

func (t *teeWriter) Handle(ctx context.Context, e Entry) error {
	var errs []error
	for _, w := range t.writers {
		if !w.Enabled(ctx, e.Level) {
			continue
		}
		if err := w.Handle(ctx, e); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (t *teeWriter) Enabled(ctx context.Context, lvl Level) bool {
	for _, w := range t.writers {
		if w.Enabled(ctx, lvl) {
			return true
		}
	}
	return false
}
