package benchmarks

import (
	"context"
	"os"
	"testing"
	"time"

	logf "github.com/ssgreg/logf/v2"
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
			logger.Info("request handled", zapTwoScalars()...)
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
			logger.Info("request handled", zapTwoScalars()...)
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

	enc := logf.NewJSONEncoder(logf.JSONEncoderConfig{
		EncodeTime:     logf.RFC3339NanoTimeEncoder,
		EncodeDuration: logf.NanoDurationEncoder,
	})
	h, closeFn, err := logf.NewRouter().
		Route(enc, logf.Output(logf.LevelDebug, f)).
		Build()
	if err != nil {
		b.Fatal(err)
	}
	logger := logf.NewLogger(h).WithCaller(false)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.Info(ctx, "request handled", logfTwoScalars()...)
		}
	})
	b.StopTimer()
	closeFn()
}

// --- logf: Router → BufferedWriter 256 KB → file ---

func BenchmarkFileParallel_Logf_Buffer256KB(b *testing.B) {
	f := tempFile(b)
	defer os.Remove(f.Name())
	defer f.Close()

	bw, closeBW := logf.NewBufferedWriter(f, logf.WithBufSize(256*1024))
	enc := logf.NewJSONEncoder(logf.JSONEncoderConfig{
		EncodeTime:     logf.RFC3339NanoTimeEncoder,
		EncodeDuration: logf.NanoDurationEncoder,
	})
	h, closeFn, err := logf.NewRouter().
		Route(enc, logf.Output(logf.LevelDebug, bw)).
		Build()
	if err != nil {
		b.Fatal(err)
	}
	logger := logf.NewLogger(h).WithCaller(false)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.Info(ctx, "request handled", logfTwoScalars()...)
		}
	})
	b.StopTimer()
	closeFn()
	closeBW()
}

// --- logf: Router → SlabWriter 64 KB × 4 → file ---

func BenchmarkFileParallel_Logf_Slab64Kx4(b *testing.B) {
	f := tempFile(b)
	defer os.Remove(f.Name())
	defer f.Close()

	sw := logf.NewSlabWriter(f, 64*1024, 4,
		logf.WithFlushInterval(time.Second),
	)
	enc := logf.NewJSONEncoder(logf.JSONEncoderConfig{
		EncodeTime:     logf.RFC3339NanoTimeEncoder,
		EncodeDuration: logf.NanoDurationEncoder,
	})
	h, closeFn, err := logf.NewRouter().
		Route(enc, logf.Output(logf.LevelDebug, sw)).
		Build()
	if err != nil {
		b.Fatal(err)
	}
	logger := logf.NewLogger(h).WithCaller(false)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			logger.Info(ctx, "request handled", logfTwoScalars()...)
		}
	})
	b.StopTimer()
	closeFn()
	sw.Close()

	// SlabWriter is blocking — 0 drops by design.
	b.ReportMetric(0, "drops")
}
