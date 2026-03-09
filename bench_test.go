package logf

import (
	"context"
	"io"
	"os"
	"sync"
	"testing"
)

func benchLogger(lvl Level, addCaller bool) (*Logger, ChannelWriterCloseFunc) {
	enc := NewJSONEncoder.Default()
	w, close := NewChannelWriter(lvl, ChannelWriterConfig{
		Appender: NewWriteAppender(io.Discard, enc),
	})
	l := NewLogger(w).WithCaller(addCaller)
	return l, close
}

func benchFields() []Field {
	return []Field{
		Int("int", 42),
		String("string", "hello"),
		String("path", "/api/v1/users"),
		Int64("latency_us", 1234),
		Bool("ok", true),
	}
}

var benchCtx = context.Background()

// --- Disabled path ---

func BenchmarkDisabledLog(b *testing.B) {
	logger, close := benchLogger(LevelError, false)
	defer close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info(benchCtx, "request handled")
	}
}

func BenchmarkDisabledLogWithFields(b *testing.B) {
	logger, close := benchLogger(LevelError, false)
	defer close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info(benchCtx, "request handled", benchFields()...)
	}
}

func BenchmarkDisabledAtLevel(b *testing.B) {
	logger, close := benchLogger(LevelError, false)
	defer close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.AtLevel(benchCtx, LevelInfo, func(log LogFunc) {
			log(benchCtx, "request handled")
		})
	}
}

// --- Enabled path ---

func BenchmarkPlainText(b *testing.B) {
	logger, close := benchLogger(LevelDebug, false)
	defer close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info(benchCtx, "request handled")
	}
}

func BenchmarkTextWithFields(b *testing.B) {
	logger, close := benchLogger(LevelDebug, false)
	defer close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info(benchCtx, "request handled", benchFields()...)
	}
}

func BenchmarkTextWithAccumulatedFields(b *testing.B) {
	logger, close := benchLogger(LevelDebug, false)
	defer close()
	logger = logger.With(benchFields()...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info(benchCtx, "request handled")
	}
}

// --- Caller ---

func BenchmarkPlainTextWithCaller(b *testing.B) {
	logger, close := benchLogger(LevelDebug, true)
	defer close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info(benchCtx, "request handled")
	}
}

func BenchmarkTextWithFieldsAndCaller(b *testing.B) {
	logger, close := benchLogger(LevelDebug, true)
	defer close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info(benchCtx, "request handled", benchFields()...)
	}
}

// --- Context bag ---

func BenchmarkContextBag(b *testing.B) {
	enc := NewJSONEncoder.Default()
	w, close := NewChannelWriter(LevelDebug, ChannelWriterConfig{
		Appender: NewWriteAppender(io.Discard, enc),
	})
	defer close()
	cw := NewContextWriter(w)
	logger := NewLogger(cw)

	ctx := With(benchCtx, String("request_id", "abc-123"), String("user_id", "u42"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info(ctx, "request handled")
	}
}

// --- FieldSource ---

type traceKeyBench struct{}

func BenchmarkFieldSource(b *testing.B) {
	enc := NewJSONEncoder.Default()
	w, close := NewChannelWriter(LevelDebug, ChannelWriterConfig{
		Appender: NewWriteAppender(io.Discard, enc),
	})
	defer close()

	src := FieldSource(func(ctx context.Context) []Field {
		if v := ctx.Value(traceKeyBench{}); v != nil {
			return []Field{String("trace_id", v.(string))}
		}
		return nil
	})
	cw := NewContextWriter(w, src)
	logger := NewLogger(cw)

	ctx := context.WithValue(benchCtx, traceKeyBench{}, "trace-abc")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info(ctx, "request handled")
	}
}

// --- Bag.With ---

func BenchmarkBagWith(b *testing.B) {
	ctx := With(benchCtx, String("k1", "v1"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = With(ctx, String("k2", "v2"))
	}
}

// --- Logger.With ---

func BenchmarkLoggerWith(b *testing.B) {
	logger, close := benchLogger(LevelDebug, false)
	defer close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = logger.With(benchFields()...)
	}
}

// --- LogDepth ---

func BenchmarkLogDepth(b *testing.B) {
	logger, close := benchLogger(LevelDebug, true)
	defer close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LogDepth(logger, benchCtx, 1, LevelInfo, "request handled")
	}
}

// --- Sync (no channel, encoding in caller goroutine) ---

func benchSyncLogger(lvl Level, addCaller bool) *Logger {
	enc := NewJSONEncoder.Default()
	w := NewUnbufferedEntryWriter(lvl, NewWriteAppender(io.Discard, enc))
	return NewLogger(w).WithCaller(addCaller)
}

func BenchmarkSyncPlainText(b *testing.B) {
	logger := benchSyncLogger(LevelDebug, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info(benchCtx, "request handled")
	}
}

func BenchmarkSyncTextWithFields(b *testing.B) {
	logger := benchSyncLogger(LevelDebug, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info(benchCtx, "request handled", benchFields()...)
	}
}

func BenchmarkSyncPlainTextWithCaller(b *testing.B) {
	logger := benchSyncLogger(LevelDebug, true)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info(benchCtx, "request handled")
	}
}

// --- Parallel (shows lock contention effect) ---

func BenchmarkParallelPlainText(b *testing.B) {
	logger, close := benchLogger(LevelDebug, false)
	defer close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info(benchCtx, "request handled")
		}
	})
}

// --- Parallel with real file I/O ---

func BenchmarkParallelFileIO(b *testing.B) {
	f, err := os.CreateTemp("", "logf-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	enc := NewJSONEncoder.Default()
	w, close := NewChannelWriter(LevelDebug, ChannelWriterConfig{
		Appender: NewWriteAppender(f, enc),
	})
	defer close()
	logger := NewLogger(w).WithCaller(false)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info(benchCtx, "request handled")
		}
	})
}

func BenchmarkSyncFileIOWithFields(b *testing.B) {
	f, err := os.CreateTemp("", "logf-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	enc := NewJSONEncoder.Default()
	w := NewUnbufferedEntryWriter(LevelDebug, NewWriteAppender(f, enc))
	logger := NewLogger(w).WithCaller(false).With(benchFields()...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info(benchCtx, "request handled")
	}
}

func BenchmarkSyncParallelFileIOWithFields(b *testing.B) {
	f, err := os.CreateTemp("", "logf-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	enc := NewJSONEncoder.Default()
	w := &lockedEntryWriter{w: NewUnbufferedEntryWriter(LevelDebug, NewWriteAppender(f, enc))}
	logger := NewLogger(w).WithCaller(false).With(benchFields()...)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info(benchCtx, "request handled")
		}
	})
}

func BenchmarkSyncFileIO(b *testing.B) {
	f, err := os.CreateTemp("", "logf-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	enc := NewJSONEncoder.Default()
	w := NewUnbufferedEntryWriter(LevelDebug, NewWriteAppender(f, enc))
	logger := NewLogger(w).WithCaller(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info(benchCtx, "request handled")
	}
}

// lockedEntryWriter wraps an EntryWriter with a mutex for parallel safety.
type lockedEntryWriter struct {
	mu sync.Mutex
	w  EntryWriter
}

func (w *lockedEntryWriter) WriteEntry(ctx context.Context, e Entry) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.w.WriteEntry(ctx, e)
}

func (w *lockedEntryWriter) Enabled(ctx context.Context, lvl Level) bool {
	return w.w.Enabled(ctx, lvl)
}

func BenchmarkSyncParallelFileIO(b *testing.B) {
	f, err := os.CreateTemp("", "logf-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	enc := NewJSONEncoder.Default()
	w := &lockedEntryWriter{w: NewUnbufferedEntryWriter(LevelDebug, NewWriteAppender(f, enc))}
	logger := NewLogger(w).WithCaller(false)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info(benchCtx, "request handled")
		}
	})
}

// --- Realistic pipeline ---

func BenchmarkRealisticPipeline(b *testing.B) {
	enc := NewJSONEncoder.Default()
	w, close := NewChannelWriter(LevelDebug, ChannelWriterConfig{
		Appender: NewWriteAppender(io.Discard, enc),
	})
	defer close()
	cw := NewContextWriter(w)
	logger := NewLogger(cw).WithCaller(true)

	ctx := With(benchCtx, String("request_id", "abc-123"))
	ctx = NewContext(ctx, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := FromContext(ctx)
		l.Info(ctx, "request handled", String("status", "ok"), Int("code", 200))
	}
}
