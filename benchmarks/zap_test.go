package benchmarks

import (
	"io"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newZapDiscard(lvl zapcore.Level) *zap.Logger {
	enc := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(enc, zapcore.AddSync(io.Discard), lvl)
	return zap.New(core)
}

func zapTwoScalars() []zap.Field {
	return []zap.Field{
		zap.String("method", "GET"),
		zap.Int("status", 200),
	}
}

func zapSixScalars() []zap.Field {
	return []zap.Field{
		zap.String("method", "GET"),
		zap.Int("status", 200),
		zap.String("path", "/api/v1/users"),
		zap.String("user_agent", "Mozilla/5.0"),
		zap.String("request_id", "abc-def-123"),
		zap.Int("size", 1024),
	}
}

// zapUser implements zapcore.ObjectMarshaler for Object benchmarks.
type zapUser struct {
	ID   int
	Name string
}

func (u *zapUser) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt("id", u.ID)
	enc.AddString("name", u.Name)
	return nil
}

func zapSixHeavy() []zap.Field {
	return []zap.Field{
		zap.Binary("body", heavyBytes),
		zap.Time("timestamp", heavyTime),
		zap.Int64s("ids", heavyInts64),
		zap.Strings("tags", heavyStrings),
		zap.Duration("latency", heavyDuration),
		zap.Object("user", &zapUser{ID: 123, Name: "alice"}),
	}
}

// B0: DisabledLevel
func BenchmarkZap_DisabledLevel(b *testing.B) {
	logger := newZapDiscard(zapcore.InfoLevel)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Debug("request handled")
	}
}

// B1: NoFields
func BenchmarkZap_NoFields(b *testing.B) {
	logger := newZapDiscard(zapcore.DebugLevel)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled")
	}
}

// B2: TwoScalars
func BenchmarkZap_TwoScalars(b *testing.B) {
	logger := newZapDiscard(zapcore.DebugLevel)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled", zapTwoScalars()...)
	}
}

// B3: TwoScalarsInGroup
func BenchmarkZap_TwoScalarsInGroup(b *testing.B) {
	logger := newZapDiscard(zapcore.DebugLevel)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled",
			zap.Dict("request", zap.String("method", "GET"), zap.Int("status", 200)),
		)
	}
}

// B4: SixScalars
func BenchmarkZap_SixScalars(b *testing.B) {
	logger := newZapDiscard(zapcore.DebugLevel)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled", zapSixScalars()...)
	}
}

// B5: SixHeavy
func BenchmarkZap_SixHeavy(b *testing.B) {
	logger := newZapDiscard(zapcore.DebugLevel)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled", zapSixHeavy()...)
	}
}

// B6: ErrorField
func BenchmarkZap_ErrorField(b *testing.B) {
	logger := newZapDiscard(zapcore.DebugLevel)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled", zap.Error(errExample))
	}
}

// B7: WithPerCall+NoFields
func BenchmarkZap_WithPerCall_NoFields(b *testing.B) {
	logger := newZapDiscard(zapcore.DebugLevel)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.With(zapTwoScalars()...).Info("request handled")
	}
}

// B8: WithPerCall+TwoScalars
func BenchmarkZap_WithPerCall_TwoScalars(b *testing.B) {
	logger := newZapDiscard(zapcore.DebugLevel)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.With(zapTwoScalars()...).Info("request handled", zapTwoScalars()...)
	}
}

// B9: WithCached+NoFields
func BenchmarkZap_WithCached_NoFields(b *testing.B) {
	logger := newZapDiscard(zapcore.DebugLevel).With(zapTwoScalars()...)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled")
	}
}

// B10: WithCached+TwoScalars
func BenchmarkZap_WithCached_TwoScalars(b *testing.B) {
	logger := newZapDiscard(zapcore.DebugLevel).With(zapTwoScalars()...)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled", zapTwoScalars()...)
	}
}

// B11: WithBoth+TwoScalars
func BenchmarkZap_WithBoth_TwoScalars(b *testing.B) {
	logger := newZapDiscard(zapcore.DebugLevel).With(zapTwoScalars()...)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.With(zapTwoScalars()...).Info("request handled", zapTwoScalars()...)
	}
}

// B12: WithGroupCached+TwoScalars
func BenchmarkZap_WithGroupCached_TwoScalars(b *testing.B) {
	logger := newZapDiscard(zapcore.DebugLevel).With(zap.Namespace("request")).With(zapTwoScalars()...)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled", zapTwoScalars()...)
	}
}

// B13: Caller+TwoScalars
func BenchmarkZap_Caller_TwoScalars(b *testing.B) {
	logger := newZapDiscard(zapcore.DebugLevel).WithOptions(zap.AddCaller())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled", zapTwoScalars()...)
	}
}

// --- Parallel variants ---

func BenchmarkZap_Parallel_NoFields(b *testing.B) {
	logger := newZapDiscard(zapcore.DebugLevel)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("request handled")
		}
	})
}

func BenchmarkZap_Parallel_TwoScalars(b *testing.B) {
	logger := newZapDiscard(zapcore.DebugLevel)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("request handled", zapTwoScalars()...)
		}
	})
}

func BenchmarkZap_Parallel_WithCached_TwoScalars(b *testing.B) {
	logger := newZapDiscard(zapcore.DebugLevel).With(zapTwoScalars()...)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("request handled", zapTwoScalars()...)
		}
	})
}

// Suppress unused import warning.
var _ = time.Now
