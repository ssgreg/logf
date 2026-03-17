package logf

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeeWriterEmpty(t *testing.T) {
	w := NewTeeWriter()
	// Zero writers — behaves like nopHandler.
	assert.False(t, w.Enabled(context.Background(), LevelDebug))
	assert.NoError(t, w.Handle(context.Background(), Entry{Text: "x"}))
}

func TestTeeWriterSingle(t *testing.T) {
	sink := &testHandler{}
	w := NewTeeWriter(sink)
	// Single writer — returned as-is, no wrapper.
	assert.Equal(t, sink, w)
}

func TestTeeWriterFanOut(t *testing.T) {
	s1 := &testHandler{}
	s2 := &testHandler{}
	w := NewTeeWriter(s1, s2)

	ctx := context.Background()
	require.NoError(t, w.Handle(ctx, Entry{Text: "hello", Level: LevelInfo}))

	require.Len(t, s1.Entries, 1)
	require.Len(t, s2.Entries, 1)
	assert.Equal(t, "hello", s1.Entries[0].Text)
	assert.Equal(t, "hello", s2.Entries[0].Text)
}

func TestTeeWriterHandleFiltersDisabled(t *testing.T) {
	errOnly := newLeveledTestHandler(LevelError)
	all := newLeveledTestHandler(LevelDebug)

	tee := NewTeeWriter(errOnly, all)
	ctx := context.Background()

	// Info entry — only "all" writer should receive it.
	require.NoError(t, tee.Handle(ctx, Entry{Text: "info", Level: LevelInfo}))
	assert.Empty(t, errOnly.Entries)
	assert.Len(t, all.Entries, 1)

	// Error entry — both writers should receive it.
	require.NoError(t, tee.Handle(ctx, Entry{Text: "err", Level: LevelError}))
	assert.Len(t, errOnly.Entries, 1)
	assert.Len(t, all.Entries, 2)
}

func TestTeeWriterEnabledAny(t *testing.T) {
	w1 := newLeveledTestHandler(LevelError) // only errors
	w2 := newLeveledTestHandler(LevelDebug) // everything

	tee := NewTeeWriter(w1, w2)
	ctx := context.Background()

	// Debug is enabled because w2 accepts it.
	assert.True(t, tee.Enabled(ctx, LevelDebug))
	// Error is enabled because both accept it.
	assert.True(t, tee.Enabled(ctx, LevelError))
}

func TestTeeWriterEnabledNone(t *testing.T) {
	w1 := newLeveledTestHandler(LevelError)
	w2 := newLeveledTestHandler(LevelError)

	tee := NewTeeWriter(w1, w2)
	// Info not enabled by either writer.
	assert.False(t, tee.Enabled(context.Background(), LevelInfo))
}

type errWriter struct {
	err error
}

func (w *errWriter) Handle(context.Context, Entry) error { return w.err }
func (w *errWriter) Enabled(context.Context, Level) bool     { return true }

func TestTeeWriterErrorJoin(t *testing.T) {
	e1 := errors.New("disk full")
	e2 := errors.New("network down")

	w := NewTeeWriter(&errWriter{e1}, &errWriter{e2})
	err := w.Handle(context.Background(), Entry{Text: "x"})

	require.Error(t, err)
	assert.ErrorIs(t, err, e1)
	assert.ErrorIs(t, err, e2)
}

func TestTeeWriterPartialError(t *testing.T) {
	sink := &testHandler{}
	fail := &errWriter{errors.New("fail")}

	w := NewTeeWriter(fail, sink)
	err := w.Handle(context.Background(), Entry{Text: "x"})

	// First writer fails, second still receives the entry.
	require.Error(t, err)
	require.Len(t, sink.Entries, 1)
	assert.Equal(t, "x", sink.Entries[0].Text)
}

func TestTeeWriterDefensiveCopy(t *testing.T) {
	s1 := &testHandler{}
	s2 := &testHandler{}
	writers := []Handler{s1, s2}

	w := NewTeeWriter(writers...)

	// Mutate original slice — tee should not be affected.
	writers[0] = nil
	require.NoError(t, w.Handle(context.Background(), Entry{Text: "safe"}))
	require.Len(t, s1.Entries, 1)
}
