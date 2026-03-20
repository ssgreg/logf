package benchmarks

import (
	"context"
	"testing"

	logf "github.com/ssgreg/logf/v2"
)

// B0: DisabledLevel
func BenchmarkLogf_DisabledLevel(b *testing.B) {
	logger := newLogfSyncInfo()
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Debug(ctx, "request handled")
	}
}

// B1: NoFields
func BenchmarkLogf_NoFields(b *testing.B) {
	logger := newLogfSync()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.Info(ctx, "request handled")
		}
	})
}

// B2: TwoScalars
func BenchmarkLogf_TwoScalars(b *testing.B) {
	logger := newLogfSync()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.Info(ctx, "request handled", logfTwoScalars()...)
		}
	})
}

// B3: TwoScalarsInGroup
func BenchmarkLogf_TwoScalarsInGroup(b *testing.B) {
	logger := newLogfSync()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.Info(ctx, "request handled",
				logf.Group("request", logf.String("method", "GET"), logf.Int("status", 200)),
			)
		}
	})
}

// B4: SixScalars
func BenchmarkLogf_SixScalars(b *testing.B) {
	logger := newLogfSync()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.Info(ctx, "request handled", logfSixScalars()...)
		}
	})
}

// B5: SixHeavy
func BenchmarkLogf_SixHeavy(b *testing.B) {
	logger := newLogfSync()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.Info(ctx, "request handled", logfSixHeavy()...)
		}
	})
}

// B6: ErrorField
func BenchmarkLogf_ErrorField(b *testing.B) {
	logger := newLogfSync()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.Info(ctx, "request handled", logf.NamedError("error", errExample))
		}
	})
}

// B7: WithPerCall+NoFields
func BenchmarkLogf_WithPerCall_NoFields(b *testing.B) {
	logger := newLogfSync()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.With(logfTwoScalars()...).Info(ctx, "request handled")
		}
	})
}

// B8: WithPerCall+TwoScalars
func BenchmarkLogf_WithPerCall_TwoScalars(b *testing.B) {
	logger := newLogfSync()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.With(logfTwoScalars()...).Info(ctx, "request handled", logfTwoScalars()...)
		}
	})
}

// B9: WithCached+NoFields
func BenchmarkLogf_WithCached_NoFields(b *testing.B) {
	logger := newLogfSync().With(logfTwoScalars()...)
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.Info(ctx, "request handled")
		}
	})
}

// B10: WithCached+TwoScalars
func BenchmarkLogf_WithCached_TwoScalars(b *testing.B) {
	logger := newLogfSync().With(logfTwoScalars()...)
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.Info(ctx, "request handled", logfTwoScalars()...)
		}
	})
}

// B11: WithBoth+TwoScalars
func BenchmarkLogf_WithBoth_TwoScalars(b *testing.B) {
	logger := newLogfSync().With(logfTwoScalars()...)
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.With(logfTwoScalars()...).Info(ctx, "request handled", logfTwoScalars()...)
		}
	})
}

// B12: WithGroupCached+TwoScalars
func BenchmarkLogf_WithGroupCached_TwoScalars(b *testing.B) {
	logger := newLogfSync().WithGroup("request").With(logfTwoScalars()...)
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.Info(ctx, "request handled", logfTwoScalars()...)
		}
	})
}

// B13: Caller+TwoScalars
func BenchmarkLogf_Caller_TwoScalars(b *testing.B) {
	logger := newLogfSyncWithCaller()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.Info(ctx, "request handled", logfTwoScalars()...)
		}
	})
}

// --- A: With micro-benchmarks (no log call) ---

// A1: With
func BenchmarkLogf_With(b *testing.B) {
	logger := newLogfSync()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = logger.With(logfTwoScalars()...)
	}
}

// A2: WithOnTop
func BenchmarkLogf_WithOnTop(b *testing.B) {
	logger := newLogfSync().With(logfTwoScalars()...)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = logger.With(logfTwoScalars()...)
	}
}

// A3: WithGroup
func BenchmarkLogf_WithGroup(b *testing.B) {
	logger := newLogfSync()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = logger.WithGroup("request")
	}
}

// --- Router variants (Router + slab buffer, block mode, parallel) ---

func BenchmarkLogf_Router_NoFields(b *testing.B) {
	logger, close := newLogfRouter()
	defer close()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.Info(ctx, "request handled")
		}
	})
}

func BenchmarkLogf_RouterSlab_NoFields(b *testing.B) {
	logger, close := newLogfRouterSlab()
	defer close()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.Info(ctx, "request handled")
		}
	})
}
