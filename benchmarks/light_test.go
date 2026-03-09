package benchmarks

import (
	"context"
	"testing"

	"github.com/ssgreg/logf/v2"
	"go.uber.org/zap"
)

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
