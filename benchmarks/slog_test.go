package benchmarks

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"
)

func newSlogDiscard() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func newSlogDiscardInfo() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelInfo}))
}

func newSlogDiscardWithCaller() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true}))
}

func slogTwoScalars() []slog.Attr {
	return []slog.Attr{
		slog.String("method", "GET"),
		slog.Int("status", 200),
	}
}

func slogSixScalars() []slog.Attr {
	return []slog.Attr{
		slog.String("method", "GET"),
		slog.Int("status", 200),
		slog.String("path", "/api/v1/users"),
		slog.String("user_agent", "Mozilla/5.0"),
		slog.String("request_id", "abc-def-123"),
		slog.Int("size", 1024),
	}
}

func slogSixHeavy() []slog.Attr {
	return []slog.Attr{
		slog.String("body", string(heavyBytes)),
		slog.Time("timestamp", heavyTime),
		slog.Any("ids", heavyInts64),
		slog.Any("tags", heavyStrings),
		slog.Duration("latency", heavyDuration),
		slog.Group("user", slog.Int("id", 123), slog.String("name", "alice")),
	}
}

func slogTwoScalarsArgs() []any {
	return []any{"method", "GET", "status", 200}
}

// B0: DisabledLevel
func BenchmarkSlog_DisabledLevel(b *testing.B) {
	logger := newSlogDiscardInfo()
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.DebugContext(ctx, "request handled")
	}
}

// B1: NoFields
func BenchmarkSlog_NoFields(b *testing.B) {
	logger := newSlogDiscard()
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.InfoContext(ctx, "request handled")
	}
}

// B2: TwoScalars
func BenchmarkSlog_TwoScalars(b *testing.B) {
	logger := newSlogDiscard()
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.LogAttrs(ctx, slog.LevelInfo, "request handled", slogTwoScalars()...)
	}
}

// B3: TwoScalarsInGroup
func BenchmarkSlog_TwoScalarsInGroup(b *testing.B) {
	logger := newSlogDiscard()
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.LogAttrs(ctx, slog.LevelInfo, "request handled",
			slog.Group("request", slog.String("method", "GET"), slog.Int("status", 200)),
		)
	}
}

// B4: SixScalars
func BenchmarkSlog_SixScalars(b *testing.B) {
	logger := newSlogDiscard()
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.LogAttrs(ctx, slog.LevelInfo, "request handled", slogSixScalars()...)
	}
}

// B5: SixHeavy
func BenchmarkSlog_SixHeavy(b *testing.B) {
	logger := newSlogDiscard()
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.LogAttrs(ctx, slog.LevelInfo, "request handled", slogSixHeavy()...)
	}
}

// B6: ErrorField
func BenchmarkSlog_ErrorField(b *testing.B) {
	logger := newSlogDiscard()
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.LogAttrs(ctx, slog.LevelInfo, "request handled", slog.Any("error", errExample))
	}
}

// B7: WithPerCall+NoFields
func BenchmarkSlog_WithPerCall_NoFields(b *testing.B) {
	logger := newSlogDiscard()
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.With(slogTwoScalarsArgs()...).InfoContext(ctx, "request handled")
	}
}

// B8: WithPerCall+TwoScalars
func BenchmarkSlog_WithPerCall_TwoScalars(b *testing.B) {
	logger := newSlogDiscard()
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.With(slogTwoScalarsArgs()...).LogAttrs(ctx, slog.LevelInfo, "request handled", slogTwoScalars()...)
	}
}

// B9: WithCached+NoFields
func BenchmarkSlog_WithCached_NoFields(b *testing.B) {
	logger := newSlogDiscard().With(slogTwoScalarsArgs()...)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.InfoContext(ctx, "request handled")
	}
}

// B10: WithCached+TwoScalars
func BenchmarkSlog_WithCached_TwoScalars(b *testing.B) {
	logger := newSlogDiscard().With(slogTwoScalarsArgs()...)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.LogAttrs(ctx, slog.LevelInfo, "request handled", slogTwoScalars()...)
	}
}

// B11: WithBoth+TwoScalars
func BenchmarkSlog_WithBoth_TwoScalars(b *testing.B) {
	logger := newSlogDiscard().With(slogTwoScalarsArgs()...)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.With(slogTwoScalarsArgs()...).LogAttrs(ctx, slog.LevelInfo, "request handled", slogTwoScalars()...)
	}
}

// B12: WithGroupCached+TwoScalars
func BenchmarkSlog_WithGroupCached_TwoScalars(b *testing.B) {
	logger := newSlogDiscard().WithGroup("request").With(slogTwoScalarsArgs()...)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.LogAttrs(ctx, slog.LevelInfo, "request handled", slogTwoScalars()...)
	}
}

// B13: Caller+TwoScalars
func BenchmarkSlog_Caller_TwoScalars(b *testing.B) {
	logger := newSlogDiscardWithCaller()
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.LogAttrs(ctx, slog.LevelInfo, "request handled", slogTwoScalars()...)
	}
}

// --- Parallel variants ---

func BenchmarkSlog_Parallel_NoFields(b *testing.B) {
	logger := newSlogDiscard()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.InfoContext(ctx, "request handled")
		}
	})
}

func BenchmarkSlog_Parallel_TwoScalars(b *testing.B) {
	logger := newSlogDiscard()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.LogAttrs(ctx, slog.LevelInfo, "request handled", slogTwoScalars()...)
		}
	})
}

func BenchmarkSlog_Parallel_WithCached_TwoScalars(b *testing.B) {
	logger := newSlogDiscard().With(slogTwoScalarsArgs()...)
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.LogAttrs(ctx, slog.LevelInfo, "request handled", slogTwoScalars()...)
		}
	})
}

// Suppress unused import warning.
var _ = time.Now
