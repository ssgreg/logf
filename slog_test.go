package logf

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"
)

// entrySink collects entries and encodes them to JSON via logf encoder.
type entrySink struct {
	mu      sync.Mutex
	entries []Entry
	enc     Encoder
}

func newSink() *entrySink {
	return &entrySink{
		enc: NewJSONEncoder(JSONEncoderConfig{
			DisableFieldTime:   true,
			DisableFieldCaller: true,
		}),
	}
}

func (s *entrySink) Handle(_ context.Context, e Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = append(s.entries, e)
	return nil
}

func (s *entrySink) Enabled(_ context.Context, _ Level) bool {
	return true
}

func (s *entrySink) last() Entry {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.entries[len(s.entries)-1]
}


func (s *entrySink) json(e Entry) string {
	buf, _ := s.enc.Encode(e)
	result := strings.TrimRight(buf.String(), "\n")
	buf.Free()
	return result
}

func (s *entrySink) lastJSON() string {
	return s.json(s.last())
}

type slogTestStringer struct{ s string }

func (ts slogTestStringer) String() string { return ts.s }

func TestSlogHandlerJSON(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*slog.Logger)
		want string
	}{
		{
			"basic",
			func(l *slog.Logger) { l.Info("hello", "key", "value", "count", 42) },
			`{"level":"info","msg":"hello","key":"value","count":42}`,
		},
		{
			"levels/info",
			func(l *slog.Logger) { l.Info("info") },
			`{"level":"info","msg":"info"}`,
		},
		{
			"levels/warn",
			func(l *slog.Logger) { l.Warn("warn") },
			`{"level":"warn","msg":"warn"}`,
		},
		{
			"levels/error",
			func(l *slog.Logger) { l.Error("error") },
			`{"level":"error","msg":"error"}`,
		},
		{
			"levels/debug-enabled",
			func(l *slog.Logger) { l.Debug("debug") },
			`{"level":"debug","msg":"debug"}`,
		},
		{
			"with-attrs",
			func(l *slog.Logger) { l.With("component", "auth").Info("login", "user", "alice") },
			`{"level":"info","msg":"login","component":"auth","user":"alice"}`,
		},
		{
			"group",
			func(l *slog.Logger) { l.WithGroup("http").Info("req", "method", "GET", "path", "/api") },
			`{"level":"info","msg":"req","http":{"method":"GET","path":"/api"}}`,
		},
		{
			"group-multiple",
			func(l *slog.Logger) { l.WithGroup("http").WithGroup("request").Info("got", "method", "GET") },
			`{"level":"info","msg":"got","http":{"request":{"method":"GET"}}}`,
		},
		{
			"group+attrs",
			func(l *slog.Logger) {
				l.WithGroup("http").With("host", "localhost").Info("req", "method", "GET")
			},
			`{"level":"info","msg":"req","http":{"host":"localhost","method":"GET"}}`,
		},
		{
			"group+attrs-before-group",
			func(l *slog.Logger) {
				l.With("app", "myapp").WithGroup("http").Info("req", "method", "GET")
			},
			`{"level":"info","msg":"req","app":"myapp","http":{"method":"GET"}}`,
		},
		{
			"group-deep",
			func(l *slog.Logger) {
				l.WithGroup("http").With("host", "localhost").WithGroup("request").Info("got", "path", "/api")
			},
			`{"level":"info","msg":"got","http":{"host":"localhost","request":{"path":"/api"}}}`,
		},
		{
			"types",
			func(l *slog.Logger) { l.Info("t", "bool", true, "int", 42, "float", 3.14, "str", "hello") },
			`{"level":"info","msg":"t","bool":true,"int":42,"float":3.14,"str":"hello"}`,
		},
		{
			"error",
			func(l *slog.Logger) { l.Error("oops", "err", errors.New("fail")) },
			`{"level":"error","msg":"oops","err":"fail"}`,
		},
		{
			"stringer",
			func(l *slog.Logger) { l.Info("s", "val", slogTestStringer{"hello"}) },
			`{"level":"info","msg":"s","val":"hello"}`,
		},
		{
			"group-attr",
			func(l *slog.Logger) {
				l.Info("g", slog.Group("user", slog.String("name", "alice"), slog.Int("age", 30)))
			},
			`{"level":"info","msg":"g","user":{"name":"alice","age":30}}`,
		},
		{
			"empty-group",
			func(l *slog.Logger) { l.Info("e", slog.Group("empty")) },
			`{"level":"info","msg":"e"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sink := newSink()
			logger := slog.New(NewSlogHandler(sink))
			tt.fn(logger)
			got := sink.lastJSON()
			if got != tt.want {
				t.Errorf("\n got: %s\nwant: %s", got, tt.want)
			}
		})
	}
}

func TestSlogHandlerDebugFiltered(t *testing.T) {
	w := newLeveledTestHandler(LevelInfo)
	slog.New(NewSlogHandler(w)).Debug("debug-should-not-appear")

	// Debug is below LevelInfo, so Enabled returns false and
	// slog.Logger won't call Handle at all — nothing written.
	if len(w.Entries) != 0 {
		t.Error("debug entry should have been filtered")
	}
}

func TestSlogHandlerWithGroupEmpty(t *testing.T) {
	sink := newSink()
	h := NewSlogHandler(sink)
	if h.WithGroup("") != h {
		t.Error("WithGroup(\"\") should return same handler")
	}
}

func TestSlogHandlerWithAttrsEmpty(t *testing.T) {
	sink := newSink()
	h := NewSlogHandler(sink)
	if h.WithAttrs(nil) != h {
		t.Error("WithAttrs(nil) should return same handler")
	}
}

func TestSlogHandlerCaller(t *testing.T) {
	sink := newSink()
	slog.New(NewSlogHandler(sink)).Info("with-caller")

	e := sink.last()
	if e.CallerPC == 0 {
		t.Error("CallerPC should be set from slog.Record.PC")
	}
}

func TestSlogHandlerContext(t *testing.T) {
	sink := newSink()
	logger := slog.New(NewSlogHandler(NewContextWriter(sink)))

	ctx := With(context.Background(), String("request_id", "abc-123"))
	logger.InfoContext(ctx, "with-bag")

	got := sink.lastJSON()
	want := `{"level":"info","msg":"with-bag","request_id":"abc-123"}`
	if got != want {
		t.Errorf("\n got: %s\nwant: %s", got, want)
	}
}

func TestSlogHandlerEnabled(t *testing.T) {
	w := newLeveledTestHandler(LevelWarn)
	h := NewSlogHandler(w)
	ctx := context.Background()

	tests := []struct {
		level slog.Level
		want  bool
	}{
		{slog.LevelDebug, false},
		{slog.LevelInfo, false},
		{slog.LevelWarn, true},
		{slog.LevelError, true},
	}
	for _, tt := range tests {
		if got := h.Enabled(ctx, tt.level); got != tt.want {
			t.Errorf("Enabled(%v): got %v, want %v", tt.level, got, tt.want)
		}
	}
}

func TestSlogHandlerTime(t *testing.T) {
	sink := newSink()
	before := time.Now()
	slog.New(NewSlogHandler(sink)).Info("timed")
	after := time.Now()

	e := sink.last()
	if e.Time.Before(before) || e.Time.After(after) {
		t.Errorf("time %v not in [%v, %v]", e.Time, before, after)
	}
}

func TestSlogHandlerImmutability(t *testing.T) {
	sink := newSink()
	h := NewSlogHandler(sink)

	slog.New(h.WithAttrs([]slog.Attr{slog.String("a", "1")})).Info("h1")
	slog.New(h.WithAttrs([]slog.Attr{slog.String("b", "2")})).Info("h2")

	want := []string{
		`{"level":"info","msg":"h1","a":"1"}`,
		`{"level":"info","msg":"h2","b":"2"}`,
	}
	for i, w := range want {
		got := sink.json(sink.entries[i])
		if got != w {
			t.Errorf("entry[%d]:\n got: %s\nwant: %s", i, got, w)
		}
	}
}
