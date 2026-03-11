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
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.InfoContext(ctx, "request handled")
	}
}

// B2: TwoScalars
func BenchmarkSlogLogf_TwoScalars(b *testing.B) {
	logger := newSlogLogfDiscard()
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.LogAttrs(ctx, slog.LevelInfo, "request handled", slogTwoScalars()...)
	}
}

// B3: TwoScalarsInGroup
func BenchmarkSlogLogf_TwoScalarsInGroup(b *testing.B) {
	logger := newSlogLogfDiscard()
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.LogAttrs(ctx, slog.LevelInfo, "request handled",
			slog.Group("request", slog.String("method", "GET"), slog.Int("status", 200)),
		)
	}
}

// B4: SixScalars
func BenchmarkSlogLogf_SixScalars(b *testing.B) {
	logger := newSlogLogfDiscard()
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.LogAttrs(ctx, slog.LevelInfo, "request handled", slogSixScalars()...)
	}
}

// B5: SixHeavy
func BenchmarkSlogLogf_SixHeavy(b *testing.B) {
	logger := newSlogLogfDiscard()
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.LogAttrs(ctx, slog.LevelInfo, "request handled", slogSixHeavy()...)
	}
}

// B6: ErrorField
func BenchmarkSlogLogf_ErrorField(b *testing.B) {
	logger := newSlogLogfDiscard()
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.LogAttrs(ctx, slog.LevelInfo, "request handled", slog.Any("error", errExample))
	}
}

// B7: WithPerCall+NoFields
func BenchmarkSlogLogf_WithPerCall_NoFields(b *testing.B) {
	logger := newSlogLogfDiscard()
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.With(slogTwoScalarsArgs()...).InfoContext(ctx, "request handled")
	}
}

// B8: WithPerCall+TwoScalars
func BenchmarkSlogLogf_WithPerCall_TwoScalars(b *testing.B) {
	logger := newSlogLogfDiscard()
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.With(slogTwoScalarsArgs()...).LogAttrs(ctx, slog.LevelInfo, "request handled", slogTwoScalars()...)
	}
}

// B9: WithCached+NoFields
func BenchmarkSlogLogf_WithCached_NoFields(b *testing.B) {
	logger := newSlogLogfDiscard().With(slogTwoScalarsArgs()...)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.InfoContext(ctx, "request handled")
	}
}

// B10: WithCached+TwoScalars
func BenchmarkSlogLogf_WithCached_TwoScalars(b *testing.B) {
	logger := newSlogLogfDiscard().With(slogTwoScalarsArgs()...)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.LogAttrs(ctx, slog.LevelInfo, "request handled", slogTwoScalars()...)
	}
}

// B11: WithBoth+TwoScalars
func BenchmarkSlogLogf_WithBoth_TwoScalars(b *testing.B) {
	logger := newSlogLogfDiscard().With(slogTwoScalarsArgs()...)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.With(slogTwoScalarsArgs()...).LogAttrs(ctx, slog.LevelInfo, "request handled", slogTwoScalars()...)
	}
}

// B12: WithGroupCached+TwoScalars
func BenchmarkSlogLogf_WithGroupCached_TwoScalars(b *testing.B) {
	logger := newSlogLogfDiscard().WithGroup("request").With(slogTwoScalarsArgs()...)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.LogAttrs(ctx, slog.LevelInfo, "request handled", slogTwoScalars()...)
	}
}

// B13: Caller+TwoScalars
func BenchmarkSlogLogf_Caller_TwoScalars(b *testing.B) {
	logger := newSlogLogfDiscardWithCaller()
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.LogAttrs(ctx, slog.LevelInfo, "request handled", slogTwoScalars()...)
	}
}
