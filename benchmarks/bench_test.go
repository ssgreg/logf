package benchmarks

import (
	"context"
	"testing"

	"github.com/ssgreg/logf"
	"github.com/ssgreg/logf/logfc"
	"go.uber.org/zap"
)

var disableOthers = true

func BenchmarkDisabledPlainText(b *testing.B) {
	b.Run("logf", func(b *testing.B) {
		logger, _ := newLogger(logf.LevelError)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0))
		}
	})
	b.Run("logf.check", func(b *testing.B) {
		logger, _ := newLogger(logf.LevelError)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.AtLevel(logf.LevelInfo, func(log logf.LogFunc) {
				log(getMessage(0))
			})
		}
	})
	if disableOthers == true {
		return
	}

	b.Run("uber/zap", func(b *testing.B) {
		logger := newZapLogger(zap.ErrorLevel)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0))
		}
	})
	b.Run("uber/zap.check", func(b *testing.B) {
		logger := newZapLogger(zap.ErrorLevel)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if m := logger.Check(zap.InfoLevel, getMessage(0)); m != nil {
				m.Write()
			}
		}
	})
	b.Run("rs/zerolog", func(b *testing.B) {
		logger := newDisabledZerolog()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info().Msg(getMessage(0))
		}
	})
	b.Run("rs/zerolog.check", func(b *testing.B) {
		logger := newDisabledZerolog()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if e := logger.Info(); e.Enabled() {
				logger.Info().Msg(getMessage(0))
			}
		}
	})
	b.Run("sirupsen/logrus", func(b *testing.B) {
		logger := newDisabledLogrus()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0))
		}
	})
}

func BenchmarkDisabledPlainTextWithAccumulatedFields(b *testing.B) {
	b.Run("logf", func(b *testing.B) {
		logger := logf.NewDisabledLogger().With(fakeFields()...)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0))
		}
	})
	b.Run("logf.check", func(b *testing.B) {
		logger := logf.NewDisabledLogger().With(fakeFields()...)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.AtLevel(logf.LevelInfo, func(log logf.LogFunc) {
				log(getMessage(0))
			})
		}
	})
	if disableOthers == true {
		return
	}

	b.Run("uber/zap", func(b *testing.B) {
		logger := newZapLogger(zap.ErrorLevel).With(fakeZapFields()...)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0))
		}
	})
	b.Run("uber/zap.check", func(b *testing.B) {
		logger := newZapLogger(zap.ErrorLevel).With(fakeZapFields()...)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if m := logger.Check(zap.InfoLevel, getMessage(0)); m != nil {
				m.Write()
			}
		}
	})
	b.Run("rs/zerolog", func(b *testing.B) {
		logger := fakeZerologContext(newDisabledZerolog().With()).Logger()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info().Msg(getMessage(0))
		}
	})
	b.Run("rs/zerolog.check", func(b *testing.B) {
		logger := fakeZerologContext(newDisabledZerolog().With()).Logger()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if e := logger.Info(); e.Enabled() {
				logger.Info().Msg(getMessage(0))
			}
		}
	})
	b.Run("sirupsen/logrus", func(b *testing.B) {
		logger := newDisabledLogrus().WithFields(fakeLogrusFields())
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0))
		}
	})
}

func BenchmarkDisabledTextWithFields(b *testing.B) {
	b.Run("logf", func(b *testing.B) {
		logger := logf.NewDisabledLogger()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0), fakeFields()...)
		}
	})
	b.Run("logf.check", func(b *testing.B) {
		logger := logf.NewDisabledLogger()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.AtLevel(logf.LevelInfo, func(log logf.LogFunc) {
				logger.Info(getMessage(0), fakeFields()...)
			})
		}
	})
	if disableOthers == true {
		return
	}

	b.Run("uber/zap", func(b *testing.B) {
		logger := newZapLogger(zap.ErrorLevel)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0), fakeZapFields()...)
		}
	})
	b.Run("uber/zap.check", func(b *testing.B) {
		logger := newZapLogger(zap.ErrorLevel)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if m := logger.Check(zap.InfoLevel, getMessage(0)); m != nil {
				m.Write(fakeZapFields()...)
			}
		}
	})
	b.Run("rs/zerolog", func(b *testing.B) {
		logger := newDisabledZerolog()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			fakeZerologFields(logger.Info()).Msg(getMessage(0))
		}
	})
	b.Run("rs/zerolog.check", func(b *testing.B) {
		logger := newDisabledZerolog()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if e := logger.Info(); e.Enabled() {
				fakeZerologFields(logger.Info()).Msg(getMessage(0))
			}
		}
	})
	b.Run("sirupsen/logrus", func(b *testing.B) {
		logger := newDisabledLogrus()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.WithFields(fakeLogrusFields()).Info(getMessage(0))
		}
	})
}

func BenchmarkAccumulateFields(b *testing.B) {
	b.Run("logf", func(b *testing.B) {
		logger, close := newLogger(logf.LevelDebug)
		defer close()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l := logger.With(fakeFields()...)
			_ = l
		}
	})
	b.Run("logfc", func(b *testing.B) {
		logger, close := newLogger(logf.LevelDebug)
		defer close()

		ctx := logf.NewContext(context.Background(), logger)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l := logfc.MustWith(ctx, fakeFields()...)
			_ = l
		}
	})
	if disableOthers == true {
		return
	}

	b.Run("uber/zap", func(b *testing.B) {
		logger := newZapLogger(zap.ErrorLevel)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l := logger.With(fakeZapFields()...)
			_ = l
		}
	})
	b.Run("rs/zerolog", func(b *testing.B) {
		logger := newZerolog()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l := fakeZerologContext(logger.With()).Logger()
			_ = l
		}
	})
	b.Run("sirupsen/logrus", func(b *testing.B) {
		logger := newLogrus()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l := logger.WithFields(fakeLogrusFields())
			_ = l
		}
	})
}

func BenchmarkAddName(b *testing.B) {
	b.Run("logf", func(b *testing.B) {
		logger, close := newLogger(logf.LevelDebug)
		defer close()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l := logger.WithName("test")
			_ = l
		}
	})
	if disableOthers == true {
		return
	}

	b.Run("uber/zap", func(b *testing.B) {
		logger := newZapLogger(zap.ErrorLevel)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l := logger.Named("test")
			_ = l
		}
	})
}

func BenchmarkPlainTextWithCaller(b *testing.B) {
	b.Run("logf", func(b *testing.B) {
		logger, close := newLogger(logf.LevelDebug)
		defer close()
		logger = logger.WithCaller()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0))
		}
	})
	if disableOthers == true {
		return
	}

	b.Run("uber/zap", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel).WithOptions(zap.AddCaller())
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0))
		}
	})
}

func BenchmarkAccumulateFieldsWithAccumulatedFields(b *testing.B) {
	b.Run("logf", func(b *testing.B) {
		logger, close := newLogger(logf.LevelDebug)
		defer close()
		logger = logger.With(fakeFields()...)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l := logger.With(fakeFields()...)
			_ = l
		}
	})
	b.Run("logfc", func(b *testing.B) {
		logger, close := newLogger(logf.LevelDebug)
		defer close()

		logger = logger.With(fakeFields()...)
		ctx := logf.NewContext(context.Background(), logger)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = logfc.MustWith(ctx, fakeFields()...)
		}
	})
	if disableOthers == true {
		return
	}

	b.Run("uber/zap", func(b *testing.B) {
		logger := newZapLogger(zap.ErrorLevel).With(fakeZapFields()...)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l := logger.With(fakeZapFields()...)
			_ = l
		}
	})
	b.Run("rs/zerolog", func(b *testing.B) {
		logger := fakeZerologContext(newZerolog().With()).Logger()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l := fakeZerologContext(logger.With()).Logger()
			_ = l
		}
	})
	b.Run("sirupsen/logrus", func(b *testing.B) {
		logger := newLogrus().WithFields(fakeLogrusFields())
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l := logger.WithFields(fakeLogrusFields())
			_ = l
		}
	})
}

func BenchmarkPlainText(b *testing.B) {
	b.Run("logf", func(b *testing.B) {
		logger, close := newLogger(logf.LevelDebug)
		defer close()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0))
		}
	})
	if disableOthers == true {
		return
	}

	b.Run("uber/zap", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0))
		}
	})
	b.Run("rs/zerolog", func(b *testing.B) {
		logger := newZerolog()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info().Msg(getMessage(0))
		}
	})
	b.Run("sirupsen/logrus", func(b *testing.B) {
		logger := newLogrus()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0))
		}
	})
}

func BenchmarkPlainTextWithAccumulatedFields(b *testing.B) {
	b.Run("logf", func(b *testing.B) {
		logger, close := newLogger(logf.LevelDebug)
		defer close()
		logger = logger.With(fakeFields()...)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0))
		}
	})
	b.Run("logfc", func(b *testing.B) {
		logger, close := newLogger(logf.LevelDebug)
		defer close()

		logger = logger.With(fakeFields()...)
		ctx := logf.NewContext(context.Background(), logger)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logfc.MustInfo(ctx, getMessage(0))
		}
	})
	if disableOthers == true {
		return
	}

	b.Run("uber/zap", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel).With(fakeZapFields()...)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0))
		}
	})
	b.Run("rs/zerolog", func(b *testing.B) {
		logger := fakeZerologContext(newZerolog().With()).Logger()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info().Msg(getMessage(0))
		}
	})
	b.Run("sirupsen/logrus", func(b *testing.B) {
		logger := newLogrus().WithFields(fakeLogrusFields())
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0))
		}
	})
}

func BenchmarkTextWithFields(b *testing.B) {
	b.Run("logf", func(b *testing.B) {
		logger, close := newLogger(logf.LevelDebug)
		defer close()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0), fakeFields()...)
		}
	})
	if disableOthers == true {
		return
	}

	b.Run("uber/zap", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0), fakeZapFields()...)
		}
	})
	b.Run("rs/zerolog", func(b *testing.B) {
		logger := newZerolog()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			fakeZerologFields(logger.Info()).Msg(getMessage(0))
		}
	})
	b.Run("sirupsen/logrus", func(b *testing.B) {
		logger := newLogrus()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.WithFields(fakeLogrusFields()).Info(getMessage(0))
		}
	})
}
