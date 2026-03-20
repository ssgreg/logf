package benchmarks

import (
	"bufio"
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	logf "github.com/ssgreg/logf/v2"
	"github.com/ssgreg/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func tempFile(b *testing.B) *os.File {
	b.Helper()
	f, err := os.CreateTemp("", "bench-fp-*")
	if err != nil {
		b.Fatal(err)
	}
	return f
}

// --- Zap: no buffer (direct file write per message) ---

func BenchmarkFileParallel_Zap_NoBuffer(b *testing.B) {
	f := tempFile(b)
	defer os.Remove(f.Name())
	defer f.Close()

	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	enc := zapcore.NewJSONEncoder(cfg)
	core := zapcore.NewCore(enc, zapcore.AddSync(f), zapcore.DebugLevel)
	logger := zap.New(core)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("request handled", zapSixHeavy()...)
		}
	})
	b.StopTimer()
	logger.Sync()
}

// --- Zap: 256 KB buffer ---

func BenchmarkFileParallel_Zap_Buffer256KB(b *testing.B) {
	f := tempFile(b)
	defer os.Remove(f.Name())
	defer f.Close()

	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	enc := zapcore.NewJSONEncoder(cfg)
	buf := &zapcore.BufferedWriteSyncer{
		WS:            zapcore.AddSync(f),
		Size:          256 * 1024,
		FlushInterval: time.Second,
	}
	core := zapcore.NewCore(enc, buf, zapcore.DebugLevel)
	logger := zap.New(core)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("request handled", zapSixHeavy()...)
		}
	})
	b.StopTimer()
	buf.Stop()
}

// --- logf: Router → file (no buffer, direct write) ---

func BenchmarkFileParallel_Logf_NoBuffer(b *testing.B) {
	f := tempFile(b)
	defer os.Remove(f.Name())
	defer f.Close()

	enc := logf.JSON().EncodeTime(logf.RFC3339NanoTimeEncoder).EncodeDuration(logf.NanoDurationEncoder).Build()
	h, closeFn, err := logf.NewRouter().
		Route(enc, logf.Output(logf.LevelDebug, f)).
		Build()
	if err != nil {
		b.Fatal(err)
	}
	logger := logf.New(h).WithCaller(false)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.Info(ctx, "request handled", logfSixHeavy()...)
		}
	})
	b.StopTimer()
	closeFn()
}

// --- logf: Router → SlabWriter 128 KB × 2 → file ---

func BenchmarkFileParallel_Logf_Slab128Kx2(b *testing.B) {
	f := tempFile(b)
	defer os.Remove(f.Name())
	defer f.Close()

	sw := logf.NewSlabWriter(f).SlabSize(128*1024).SlabCount(2).
		FlushInterval(time.Second).Build()
	enc := logf.JSON().EncodeTime(logf.RFC3339NanoTimeEncoder).EncodeDuration(logf.NanoDurationEncoder).Build()
	h, closeFn, err := logf.NewRouter().
		Route(enc, logf.Output(logf.LevelDebug, sw)).
		Build()
	if err != nil {
		b.Fatal(err)
	}
	logger := logf.New(h).WithCaller(false)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.Info(ctx, "request handled", logfSixHeavy()...)
		}
	})
	b.StopTimer()
	closeFn()
	sw.Close()

	// SlabWriter is blocking — 0 drops by design.
	b.ReportMetric(0, "drops")
}

func BenchmarkFileParallel_Logf_Slab128Kx2_Drop(b *testing.B) {
	f := tempFile(b)
	defer os.Remove(f.Name())
	defer f.Close()

	sw := logf.NewSlabWriter(f).SlabSize(128*1024).SlabCount(2).
		FlushInterval(time.Second).DropOnFull().Build()
	enc := logf.JSON().EncodeTime(logf.RFC3339NanoTimeEncoder).EncodeDuration(logf.NanoDurationEncoder).Build()
	h, closeFn, err := logf.NewRouter().
		Route(enc, logf.Output(logf.LevelDebug, sw)).
		Build()
	if err != nil {
		b.Fatal(err)
	}
	logger := logf.New(h).WithCaller(false)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.Info(ctx, "request handled", logfSixHeavy()...)
		}
	})
	b.StopTimer()
	_ = closeFn()
	_ = sw.Close()

	b.ReportMetric(float64(sw.Stats().Dropped), "drops")
}

func BenchmarkFileParallel_Logf_Slab64Kx1(b *testing.B) {
	f := tempFile(b)
	defer os.Remove(f.Name())
	defer f.Close()

	// 1 slab = minimal config, blocking. Equivalent to a single 64KB buffer.
	sw := logf.NewSlabWriter(f).SlabSize(64*1024).SlabCount(1).
		FlushInterval(time.Second).Build()
	enc := logf.JSON().EncodeTime(logf.RFC3339NanoTimeEncoder).EncodeDuration(logf.NanoDurationEncoder).Build()
	h, closeFn, err := logf.NewRouter().
		Route(enc, logf.Output(logf.LevelDebug, sw)).
		Build()
	if err != nil {
		b.Fatal(err)
	}
	logger := logf.New(h).WithCaller(false)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.Info(ctx, "request handled", logfSixHeavy()...)
		}
	})
	b.StopTimer()
	_ = closeFn()
	_ = sw.Close()

	// 1 slab blocking — 0 drops by design.
	b.ReportMetric(0, "drops")
}

func BenchmarkFileParallel_Logf_Slab64Kx2_Drop(b *testing.B) {
	f := tempFile(b)
	defer os.Remove(f.Name())
	defer f.Close()

	// 2 slabs: 1 current + 1 spare. Minimal working drop config.
	sw := logf.NewSlabWriter(f).SlabSize(64*1024).SlabCount(2).
		FlushInterval(time.Second).DropOnFull().Build()
	enc := logf.JSON().EncodeTime(logf.RFC3339NanoTimeEncoder).EncodeDuration(logf.NanoDurationEncoder).Build()
	h, closeFn, err := logf.NewRouter().
		Route(enc, logf.Output(logf.LevelDebug, sw)).
		Build()
	if err != nil {
		b.Fatal(err)
	}
	logger := logf.New(h).WithCaller(false)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.Info(ctx, "request handled", logfSixHeavy()...)
		}
	})
	b.StopTimer()
	_ = closeFn()
	_ = sw.Close()

	b.ReportMetric(float64(sw.Stats().Dropped), "drops")
}

// --- slog: unbuffered ---

func BenchmarkFileParallel_Slog_NoBuffer(b *testing.B) {
	f := tempFile(b)
	defer os.Remove(f.Name())
	defer f.Close()

	logger := slog.New(slog.NewJSONHandler(f, &slog.HandlerOptions{Level: slog.LevelDebug}))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("request handled", slogSixHeavyArgs()...)
		}
	})
}

// --- slog: buffered (lockedBufWriter 256KB) ---

func BenchmarkFileParallel_Slog_Buffer256KB(b *testing.B) {
	f := tempFile(b)
	defer os.Remove(f.Name())
	defer f.Close()

	lbw := &lockedBufWriter{bw: bufio.NewWriterSize(f, 256*1024)}
	logger := slog.New(slog.NewJSONHandler(lbw, &slog.HandlerOptions{Level: slog.LevelDebug}))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("request handled", slogSixHeavyArgs()...)
		}
	})
	b.StopTimer()
	lbw.Flush()
}

// --- slog+logf: unbuffered ---

func BenchmarkFileParallel_SlogLogf_NoBuffer(b *testing.B) {
	f := tempFile(b)
	defer os.Remove(f.Name())
	defer f.Close()

	enc := logf.JSON().EncodeTime(logf.RFC3339NanoTimeEncoder).EncodeDuration(logf.NanoDurationEncoder).Build()
	h, closeFn, err := logf.NewRouter().
		Route(enc, logf.Output(logf.LevelDebug, f)).
		Build()
	if err != nil {
		b.Fatal(err)
	}
	logger := logf.New(h).WithCaller(false).Slog()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("request handled", slogSixHeavyArgs()...)
		}
	})
	b.StopTimer()
	closeFn()
}

// --- slog+logf: slab ---

func BenchmarkFileParallel_SlogLogf_Slab(b *testing.B) {
	f := tempFile(b)
	defer os.Remove(f.Name())
	defer f.Close()

	sw := logf.NewSlabWriter(f).SlabSize(128*1024).SlabCount(2).
		FlushInterval(time.Second).Build()
	enc := logf.JSON().EncodeTime(logf.RFC3339NanoTimeEncoder).EncodeDuration(logf.NanoDurationEncoder).Build()
	h, closeFn, err := logf.NewRouter().
		Route(enc, logf.Output(logf.LevelDebug, sw)).
		Build()
	if err != nil {
		b.Fatal(err)
	}
	logger := logf.New(h).WithCaller(false).Slog()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("request handled", slogSixHeavyArgs()...)
		}
	})
	b.StopTimer()
	closeFn()
	sw.Close()
}

// --- zerolog: unbuffered ---

func BenchmarkFileParallel_Zerolog_NoBuffer(b *testing.B) {
	f := tempFile(b)
	defer os.Remove(f.Name())
	defer f.Close()

	logger := zerolog.New(f).With().Timestamp().Logger()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			zerologSixHeavy(logger.Info()).Msg("request handled")
		}
	})
}

// --- zerolog: buffered (lockedBufWriter 256KB) ---

func BenchmarkFileParallel_Zerolog_Buffer256KB(b *testing.B) {
	f := tempFile(b)
	defer os.Remove(f.Name())
	defer f.Close()

	lbw := &lockedBufWriter{bw: bufio.NewWriterSize(f, 256*1024)}
	logger := zerolog.New(lbw).With().Timestamp().Logger()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			zerologSixHeavy(logger.Info()).Msg("request handled")
		}
	})
	b.StopTimer()
	lbw.Flush()
}

// --- logrus: unbuffered ---

func BenchmarkFileParallel_Logrus_NoBuffer(b *testing.B) {
	f := tempFile(b)
	defer os.Remove(f.Name())
	defer f.Close()

	logger := &logrus.Logger{
		Out:       f,
		Formatter: new(logrus.JSONFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.DebugLevel,
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.WithFields(logrusSixHeavy()).Info("request handled")
		}
	})
}

// --- logrus: buffered (lockedBufWriter 256KB) ---

func BenchmarkFileParallel_Logrus_Buffer256KB(b *testing.B) {
	f := tempFile(b)
	defer os.Remove(f.Name())
	defer f.Close()

	lbw := &lockedBufWriter{bw: bufio.NewWriterSize(f, 256*1024)}
	logger := &logrus.Logger{
		Out:       lbw,
		Formatter: new(logrus.JSONFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.DebugLevel,
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.WithFields(logrusSixHeavy()).Info("request handled")
		}
	})
	b.StopTimer()
	lbw.Flush()
}
