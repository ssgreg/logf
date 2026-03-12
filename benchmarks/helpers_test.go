package benchmarks

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	logf "github.com/ssgreg/logf/v2"
)

// errExample is a representative error value used in ErrorField benchmarks.
var errExample = errors.New("fail")

// --- benchUser implements logf.ObjectEncoder for Object field benchmarks ---

type benchUser struct {
	ID   int
	Name string
}

func (u *benchUser) EncodeLogfObject(enc logf.FieldEncoder) error {
	enc.EncodeFieldInt64("id", int64(u.ID))
	enc.EncodeFieldString("name", u.Name)
	return nil
}

// --- benchArray implements logf.ArrayEncoder for Array field benchmarks ---

type benchArray []int

func (a benchArray) EncodeLogfArray(enc logf.TypeEncoder) error {
	for _, v := range a {
		enc.EncodeTypeInt64(int64(v))
	}
	return nil
}

// --- benchSnapshotter implements logf.Snapshotter for Any(Snapshotter) benchmarks ---

type benchSnapshotter struct {
	Value string
}

func (s *benchSnapshotter) TakeSnapshot() interface{} {
	return &benchSnapshotter{Value: s.Value}
}

func (s *benchSnapshotter) String() string {
	return s.Value
}

// --- benchStringer implements fmt.Stringer for Stringer/Any(Stringer) benchmarks ---

type benchStringer struct {
	Value string
}

func (s *benchStringer) String() string {
	return s.Value
}

// Verify interfaces at compile time.
var (
	_ logf.ObjectEncoder = (*benchUser)(nil)
	_ logf.ArrayEncoder  = benchArray(nil)
	_ logf.Snapshotter   = (*benchSnapshotter)(nil)
	_ fmt.Stringer       = (*benchStringer)(nil)
)

// --- Shared field set constructors for logf ---

func logfTwoScalars() []logf.Field {
	return []logf.Field{
		logf.String("method", "GET"),
		logf.Int("status", 200),
	}
}

func logfSixScalars() []logf.Field {
	return []logf.Field{
		logf.String("method", "GET"),
		logf.Int("status", 200),
		logf.String("path", "/api/v1/users"),
		logf.String("user_agent", "Mozilla/5.0"),
		logf.String("request_id", "abc-def-123"),
		logf.Int("size", 1024),
	}
}

var (
	heavyBytes    = make([]byte, 256)
	heavyTime     = time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	heavyInts64   = []int64{1, 2, 3, 4, 5, 6, 7, 8}
	heavyStrings  = []string{"api", "auth", "prod", "v2"}
	heavyDuration = 42 * time.Millisecond
	heavyUser     = &benchUser{ID: 123, Name: "alice"}
)

func logfSixHeavy() []logf.Field {
	return []logf.Field{
		logf.ConstBytes("body", heavyBytes),
		logf.Time("timestamp", heavyTime),
		logf.ConstInts64("ids", heavyInts64),
		logf.ConstStrings("tags", heavyStrings),
		logf.Duration("latency", heavyDuration),
		logf.Object("user", heavyUser),
	}
}

// --- logf logger constructors ---

func newLogfSync() *logf.Logger {
	enc := logf.NewJSONEncoder(logf.JSONEncoderConfig{
		EncodeTime:     logf.RFC3339NanoTimeEncoder,
		EncodeDuration: logf.NanoDurationEncoder,
	})
	w := logf.NewWriter(logf.LevelDebug, io.Discard, enc)
	return logf.NewLogger(w).WithCaller(false)
}

func newLogfSyncInfo() *logf.Logger {
	w := logf.NewWriter(logf.LevelInfo, io.Discard, logf.NewJSONEncoder(logf.JSONEncoderConfig{}))
	return logf.NewLogger(w).WithCaller(false)
}

func newLogfSyncWithCaller() *logf.Logger {
	w := logf.NewWriter(logf.LevelDebug, io.Discard, logf.NewJSONEncoder(logf.JSONEncoderConfig{}))
	return logf.NewLogger(w).WithCaller(true)
}

func newLogfAsync() (*logf.Logger, logf.ChannelWriterCloseFunc) {
	w, close := logf.NewChannelWriter(logf.LevelDebug, logf.ChannelWriterConfig{
		Appender: logf.NewWriteAppender(io.Discard, logf.NewJSONEncoder(logf.JSONEncoderConfig{})),
	})
	return logf.NewLogger(w).WithCaller(false), close
}

func newLogfPooledAsync() (*logf.Logger, func()) {
	w, close := logf.NewAsyncWriter(logf.LevelDebug, io.Discard, logf.NewJSONEncoder(logf.JSONEncoderConfig{}))
	return logf.NewLogger(w).WithCaller(false), close
}

// --- lockedBufWriter — thread-safe buffered writer (used by zerolog in latency tests) ---

type lockedBufWriter struct {
	mu sync.Mutex
	bw *bufio.Writer
}

func (w *lockedBufWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.bw.Write(p)
}

func (w *lockedBufWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.bw.Flush()
}
