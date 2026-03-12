package benchmarks

import (
	"context"
	"log/slog"
	"testing"
)

func newSlogLogfDiscard() *slog.Logger {
	return newLogfSync().Slog()
}

func newSlogLogfDiscardInfo() *slog.Logger {
	return newLogfSyncInfo().Slog()
}

func newSlogLogfDiscardWithCaller() *slog.Logger {
	return newLogfSyncWithCaller().Slog()
}

// B0: DisabledLevel
func BenchmarkSlogLogf_DisabledLevel(b *testing.B) {
	logger := newSlogLogfDiscardInfo()
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.DebugContext(ctx, "request handled")
	}
}

// B1: NoFields
func BenchmarkSlogLogf_NoFields(b *testing.B) {
	logger := newSlogLogfDiscard()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.InfoContext(ctx, "request handled")
		}
	})
}

// B2: TwoScalars
func BenchmarkSlogLogf_TwoScalars(b *testing.B) {
	logger := newSlogLogfDiscard()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.LogAttrs(ctx, slog.LevelInfo, "request handled", slogTwoScalars()...)
		}
	})
}

// B3: TwoScalarsInGroup
func BenchmarkSlogLogf_TwoScalarsInGroup(b *testing.B) {
	logger := newSlogLogfDiscard()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.LogAttrs(ctx, slog.LevelInfo, "request handled",
				slog.Group("request", slog.String("method", "GET"), slog.Int("status", 200)),
			)
		}
	})
}

// B4: SixScalars
func BenchmarkSlogLogf_SixScalars(b *testing.B) {
	logger := newSlogLogfDiscard()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.LogAttrs(ctx, slog.LevelInfo, "request handled", slogSixScalars()...)
		}
	})
}

// B5: SixHeavy — uses slog.Any("user", heavyUser) to show bridge
// recognizing logf.ObjectEncoder (no json/reflect).
func BenchmarkSlogLogf_SixHeavy(b *testing.B) {
	logger := newSlogLogfDiscard()
	heavy := func() []slog.Attr {
		return []slog.Attr{
			slog.String("body", string(heavyBytes)),
			slog.Time("timestamp", heavyTime),
			slog.Any("ids", heavyInts64),
			slog.Any("tags", heavyStrings),
			slog.Duration("latency", heavyDuration),
			slog.Any("user", heavyUser),
		}
	}
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.LogAttrs(ctx, slog.LevelInfo, "request handled", heavy()...)
		}
	})
}

// B6: ErrorField
func BenchmarkSlogLogf_ErrorField(b *testing.B) {
	logger := newSlogLogfDiscard()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.LogAttrs(ctx, slog.LevelInfo, "request handled", slog.Any("error", errExample))
		}
	})
}

// B7: WithPerCall+NoFields
func BenchmarkSlogLogf_WithPerCall_NoFields(b *testing.B) {
	logger := newSlogLogfDiscard()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.With(slogTwoScalarsArgs()...).InfoContext(ctx, "request handled")
		}
	})
}

// B8: WithPerCall+TwoScalars
func BenchmarkSlogLogf_WithPerCall_TwoScalars(b *testing.B) {
	logger := newSlogLogfDiscard()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.With(slogTwoScalarsArgs()...).LogAttrs(ctx, slog.LevelInfo, "request handled", slogTwoScalars()...)
		}
	})
}

// B9: WithCached+NoFields
func BenchmarkSlogLogf_WithCached_NoFields(b *testing.B) {
	logger := newSlogLogfDiscard().With(slogTwoScalarsArgs()...)
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.InfoContext(ctx, "request handled")
		}
	})
}

// B10: WithCached+TwoScalars
func BenchmarkSlogLogf_WithCached_TwoScalars(b *testing.B) {
	logger := newSlogLogfDiscard().With(slogTwoScalarsArgs()...)
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.LogAttrs(ctx, slog.LevelInfo, "request handled", slogTwoScalars()...)
		}
	})
}

// B11: WithBoth+TwoScalars
func BenchmarkSlogLogf_WithBoth_TwoScalars(b *testing.B) {
	logger := newSlogLogfDiscard().With(slogTwoScalarsArgs()...)
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.With(slogTwoScalarsArgs()...).LogAttrs(ctx, slog.LevelInfo, "request handled", slogTwoScalars()...)
		}
	})
}

// B12: WithGroupCached+TwoScalars
func BenchmarkSlogLogf_WithGroupCached_TwoScalars(b *testing.B) {
	logger := newSlogLogfDiscard().WithGroup("request").With(slogTwoScalarsArgs()...)
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.LogAttrs(ctx, slog.LevelInfo, "request handled", slogTwoScalars()...)
		}
	})
}

// B13: Caller+TwoScalars
func BenchmarkSlogLogf_Caller_TwoScalars(b *testing.B) {
	logger := newSlogLogfDiscardWithCaller()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.LogAttrs(ctx, slog.LevelInfo, "request handled", slogTwoScalars()...)
		}
	})
}

// --- A: With micro-benchmarks (no log call) ---

// A1: With
func BenchmarkSlogLogf_With(b *testing.B) {
	logger := newSlogLogfDiscard()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = logger.With(slogTwoScalarsArgs()...)
	}
}

// A2: WithOnTop
func BenchmarkSlogLogf_WithOnTop(b *testing.B) {
	logger := newSlogLogfDiscard().With(slogTwoScalarsArgs()...)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = logger.With(slogTwoScalarsArgs()...)
	}
}

// A3: WithGroup
func BenchmarkSlogLogf_WithGroup(b *testing.B) {
	logger := newSlogLogfDiscard()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = logger.WithGroup("request")
	}
}
