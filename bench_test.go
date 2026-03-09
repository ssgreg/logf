package logf

import (
	"context"
	"io"
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

// Internal API benchmarks that require package-level access (package logf).
// These test functionality not reachable from the external benchmarks/ package:
//   - ContextWriter, FieldSource, With(ctx), FromContext, NewContext
//   - LogDepth (unexported-friendly caller depth)
//   - Bag.With (context-based field accumulation)
//
// Plain logging, sync/async, file I/O, and cross-logger comparisons
// live in benchmarks/ (external package).

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
