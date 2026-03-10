package benchmarks

import (
	"context"
	"log/slog"
	"testing"

	"github.com/ssgreg/logf/v2"
	"go.uber.org/zap"
)

func BenchmarkLightTextWithAccumulatedFields(b *testing.B) {
	b.Run("logf", func(b *testing.B) {
		logger, close := newLogger(logf.LevelDebug)
		defer close()
		logger = logger.WithCaller(false).With(logfFields()...)
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(ctx, getMessage(0))
		}
	})
	b.Run("logf.sync", func(b *testing.B) {
		logger := newSyncLogger(logf.LevelDebug).WithCaller(false).With(logfFields()...)
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(ctx, getMessage(0))
		}
	})
	b.Run("uber/zap", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel).With(zapFields()...)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0))
		}
	})
	b.Run("log/slog", func(b *testing.B) {
		logger := newSlogLogger().With(slogFields()...)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0))
		}
	})
	b.Run("rs/zerolog", func(b *testing.B) {
		logger := newZerolog().With().
			Int("int", 42).
			Str("string", "hello").
			Str("path", "/api/v1/users").
			Int64("latency_us", 1234).
			Bool("ok", true).
			Logger()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info().Msg(getMessage(0))
		}
	})
	b.Run("sirupsen/logrus", func(b *testing.B) {
		logger := newLogrus().WithFields(logrusFields())
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0))
		}
	})
}

func BenchmarkLightTextWithFields(b *testing.B) {
	b.Run("logf", func(b *testing.B) {
		logger, close := newLogger(logf.LevelDebug)
		defer close()
		logger = logger.WithCaller(false)
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(ctx, getMessage(0), logfFields()...)
		}
	})
	b.Run("logf.sync", func(b *testing.B) {
		logger := newSyncLogger(logf.LevelDebug).WithCaller(false)
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(ctx, getMessage(0), logfFields()...)
		}
	})
	b.Run("uber/zap", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0), zapFields()...)
		}
	})
	b.Run("log/slog", func(b *testing.B) {
		logger := newSlogLogger()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0), slogFields()...)
		}
	})
	b.Run("rs/zerolog", func(b *testing.B) {
		logger := newZerolog()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			zerologFields(logger.Info()).Msg(getMessage(0))
		}
	})
	b.Run("sirupsen/logrus", func(b *testing.B) {
		logger := newLogrus()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.WithFields(logrusFields()).Info(getMessage(0))
		}
	})
}

func BenchmarkGroupField(b *testing.B) {
	b.Run("logf", func(b *testing.B) {
		logger, close := newLogger(logf.LevelDebug)
		defer close()
		logger = logger.WithCaller(false)
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(ctx, getMessage(0),
				logf.Group("request",
					logf.String("method", "GET"),
					logf.String("path", "/api/v1/users"),
					logf.Int("status", 200),
				),
				logf.String("user", "alice"),
				logf.Int64("latency_us", 1234),
			)
		}
	})
	b.Run("logf.sync", func(b *testing.B) {
		logger := newSyncLogger(logf.LevelDebug).WithCaller(false)
		ctx := context.Background()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(ctx, getMessage(0),
				logf.Group("request",
					logf.String("method", "GET"),
					logf.String("path", "/api/v1/users"),
					logf.Int("status", 200),
				),
				logf.String("user", "alice"),
				logf.Int64("latency_us", 1234),
			)
		}
	})
	b.Run("uber/zap", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0),
				zap.Dict("request",
					zap.String("method", "GET"),
					zap.String("path", "/api/v1/users"),
					zap.Int("status", 200),
				),
				zap.String("user", "alice"),
				zap.Int64("latency_us", 1234),
			)
		}
	})
	b.Run("log/slog", func(b *testing.B) {
		logger := newSlogLogger()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0),
				slog.Group("request",
					slog.String("method", "GET"),
					slog.String("path", "/api/v1/users"),
					slog.Int("status", 200),
				),
				slog.String("user", "alice"),
				slog.Int64("latency_us", 1234),
			)
		}
	})
}
