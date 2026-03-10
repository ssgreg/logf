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

func (s *entrySink) WriteEntry(_ context.Context, e Entry) error {
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

func (s *entrySink) len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.entries)
}

func (s *entrySink) json(e Entry) string {
	buf := NewBuffer()
	_ = s.enc.Encode(buf, e)
	return strings.TrimRight(buf.String(), "\n")
}

func (s *entrySink) lastJSON() string {
	return s.json(s.last())
}

type slogTestStringer struct{ s string }

func (ts slogTestStringer) String() string { return ts.s }

func TestSlogHandlerJSON(t *testing.T) {
	tests := []struct {
		name string
		opts *SlogHandlerOptions
		fn   func(*slog.Logger)
		want string
	}{
		{
			"basic",
			nil,
			func(l *slog.Logger) { l.Info("hello", "key", "value", "count", 42) },
			`{"level":"info","msg":"hello","key":"value","count":42}`,
		},
		{
			"levels/info",
			nil,
			func(l *slog.Logger) { l.Info("info") },
			`{"level":"info","msg":"info"}`,
		},
		{
			"levels/warn",
			nil,
			func(l *slog.Logger) { l.Warn("warn") },
			`{"level":"warn","msg":"warn"}`,
		},
		{
			"levels/error",
			nil,
			func(l *slog.Logger) { l.Error("error") },
			`{"level":"error","msg":"error"}`,
		},
		{
			"levels/debug-enabled",
			&SlogHandlerOptions{Level: slog.LevelDebug},
			func(l *slog.Logger) { l.Debug("debug") },
			`{"level":"debug","msg":"debug"}`,
		},
		{
			"with-attrs",
			nil,
			func(l *slog.Logger) { l.With("component", "auth").Info("login", "user", "alice") },
			`{"level":"info","msg":"login","component":"auth","user":"alice"}`,
		},
		{
			"group/dot-prefix",
			nil,
			func(l *slog.Logger) { l.WithGroup("http").Info("req", "method", "GET", "path", "/api") },
			`{"level":"info","msg":"req","http.method":"GET","http.path":"/api"}`,
		},
		{
			"group/dot-prefix-multiple",
			nil,
			func(l *slog.Logger) { l.WithGroup("http").WithGroup("request").Info("got", "method", "GET") },
			`{"level":"info","msg":"got","http.request.method":"GET"}`,
		},
		{
			"group/nested",
			&SlogHandlerOptions{NestedGroups: true},
			func(l *slog.Logger) { l.WithGroup("http").Info("req", "method", "GET") },
			`{"level":"info","msg":"req","http":{"method":"GET"}}`,
		},
		{
			"group/nested-multiple",
			&SlogHandlerOptions{NestedGroups: true},
			func(l *slog.Logger) { l.WithGroup("http").WithGroup("request").Info("got", "method", "GET") },
			`{"level":"info","msg":"got","http":{"request":{"method":"GET"}}}`,
		},
		{
			"group/dot-prefix+attrs",
			nil,
			func(l *slog.Logger) {
				l.With("app", "myapp").WithGroup("http").With("host", "localhost").Info("req", "method", "GET")
			},
			`{"level":"info","msg":"req","app":"myapp","http.host":"localhost","http.method":"GET"}`,
		},
		{
			"group/nested+attrs",
			&SlogHandlerOptions{NestedGroups: true},
			func(l *slog.Logger) {
				l.WithGroup("http").With("host", "localhost").Info("req", "method", "GET")
			},
			`{"level":"info","msg":"req","http":{"host":"localhost","method":"GET"}}`,
		},
		{
			"types",
			nil,
			func(l *slog.Logger) { l.Info("t", "bool", true, "int", 42, "float", 3.14, "str", "hello") },
			`{"level":"info","msg":"t","bool":true,"int":42,"float":3.14,"str":"hello"}`,
		},
		{
			"error",
			nil,
			func(l *slog.Logger) { l.Error("oops", "err", errors.New("fail")) },
			`{"level":"error","msg":"oops","err":"fail"}`,
		},
		{
			"stringer",
			nil,
			func(l *slog.Logger) { l.Info("s", "val", slogTestStringer{"hello"}) },
			`{"level":"info","msg":"s","val":"hello"}`,
		},
		{
			"group-attr",
			nil,
			func(l *slog.Logger) {
				l.Info("g", slog.Group("user", slog.String("name", "alice"), slog.Int("age", 30)))
			},
			`{"level":"info","msg":"g","user":{"name":"alice","age":30}}`,
		},
		{
			"empty-group",
			nil,
			func(l *slog.Logger) { l.Info("e", slog.Group("empty")) },
			`{"level":"info","msg":"e"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sink := newSink()
			logger := slog.New(NewSlogHandler(sink, tt.opts))
			tt.fn(logger)
			got := sink.lastJSON()
			if got != tt.want {
				t.Errorf("\n got: %s\nwant: %s", got, tt.want)
			}
		})
	}
}

func TestSlogHandlerDebugFiltered(t *testing.T) {
	sink := newSink()
	slog.New(NewSlogHandler(sink, nil)).Debug("debug")
	if sink.len() != 0 {
		t.Fatal("debug should be filtered at default level")
	}
}

func TestSlogHandlerWithGroupEmpty(t *testing.T) {
	sink := newSink()
	h := NewSlogHandler(sink, nil)
	if h.WithGroup("") != h {
		t.Error("WithGroup(\"\") should return same handler")
	}
}

func TestSlogHandlerWithAttrsEmpty(t *testing.T) {
	sink := newSink()
	h := NewSlogHandler(sink, nil)
	if h.WithAttrs(nil) != h {
		t.Error("WithAttrs(nil) should return same handler")
	}
}

func TestSlogHandlerCaller(t *testing.T) {
	sink := newSink()
	slog.New(NewSlogHandler(sink, nil)).Info("with-caller")

	e := sink.last()
	if e.CallerPC == 0 {
		t.Error("CallerPC should be set from slog.Record.PC")
	}
}

func TestSlogHandlerContext(t *testing.T) {
	sink := newSink()
	logger := slog.New(NewSlogHandler(NewContextWriter(sink), nil))

	ctx := With(context.Background(), String("request_id", "abc-123"))
	logger.InfoContext(ctx, "with-bag")

	got := sink.lastJSON()
	want := `{"level":"info","msg":"with-bag","request_id":"abc-123"}`
	if got != want {
		t.Errorf("\n got: %s\nwant: %s", got, want)
	}
}

func TestSlogHandlerEnabled(t *testing.T) {
	h := NewSlogHandler(newSink(), &SlogHandlerOptions{Level: slog.LevelWarn})
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
	slog.New(NewSlogHandler(sink, nil)).Info("timed")
	after := time.Now()

	e := sink.last()
	if e.Time.Before(before) || e.Time.After(after) {
		t.Errorf("time %v not in [%v, %v]", e.Time, before, after)
	}
}

func TestSlogHandlerImmutability(t *testing.T) {
	sink := newSink()
	h := NewSlogHandler(sink, nil)

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
