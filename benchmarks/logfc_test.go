package benchmarks

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/ssgreg/logf/v2"
	"github.com/ssgreg/logf/v2/logfc"
)

// --- Disabled path ---

func BenchmarkLogfcDisabledLog(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logfc.Info(ctx, "request handled")
	}
}

func BenchmarkLogfcDisabledLogWithFields(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logfc.Info(ctx, "request handled", logfFields()...)
	}
}

// --- logf vs logfc: direct comparison ---

func BenchmarkLogfInfoInLoop(b *testing.B) {
	w := logf.NewSyncWriter(logf.LevelDebug, logf.NewWriteAppender(io.Discard, logf.NewJSONEncoder.Default()))
	logger := logf.NewLogger(w).WithCaller(false)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info(ctx, "request handled")
	}
}

func BenchmarkLogfcInfoInLoop(b *testing.B) {
	w := logf.NewSyncWriter(logf.LevelDebug, logf.NewWriteAppender(io.Discard, logf.NewJSONEncoder.Default()))
	ctx := logfc.New(context.Background(), logf.NewLogger(w).WithCaller(false))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logfc.Info(ctx, "request handled")
	}
}

func BenchmarkLogfWithInLoop(b *testing.B) {
	w := logf.NewSyncWriter(logf.LevelDebug, logf.NewWriteAppender(io.Discard, logf.NewJSONEncoder.Default()))
	logger := logf.NewLogger(w).WithCaller(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = logger.With(logfFields()...)
	}
}

func BenchmarkLogfcWithInLoop(b *testing.B) {
	w := logf.NewSyncWriter(logf.LevelDebug, logf.NewWriteAppender(io.Discard, logf.NewJSONEncoder.Default()))
	ctx := logfc.New(context.Background(), logf.NewLogger(w).WithCaller(false))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = logfc.With(ctx, logfFields()...)
	}
}

// --- Context creation (per-request overhead) ---

func BenchmarkLogfcContextWith(b *testing.B) {
	w := logf.NewSyncWriter(logf.LevelDebug, logf.NewWriteAppender(io.Discard, logf.NewJSONEncoder.Default()))
	ctx := logfc.New(context.Background(), logf.NewLogger(w).WithCaller(false))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = logfc.With(ctx, logfFields()...)
	}
}

func BenchmarkLogfcContextWithGroup(b *testing.B) {
	w := logf.NewSyncWriter(logf.LevelDebug, logf.NewWriteAppender(io.Discard, logf.NewJSONEncoder.Default()))
	ctx := logfc.New(context.Background(), logf.NewLogger(w).WithCaller(false))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = logfc.WithGroup(ctx, "request")
	}
}

func BenchmarkLogfcContextWithAndLog(b *testing.B) {
	w := logf.NewSyncWriter(logf.LevelDebug, logf.NewWriteAppender(io.Discard, logf.NewJSONEncoder.Default()))
	baseCtx := logfc.New(context.Background(), logf.NewLogger(w).WithCaller(false))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := logfc.With(baseCtx, logfFields()...)
		logfc.Info(ctx, "request handled")
	}
}

// --- Sync plain text ---

func BenchmarkLogfcSyncPlainText(b *testing.B) {
	w := logf.NewSyncWriter(logf.LevelDebug, logf.NewWriteAppender(io.Discard, logf.NewJSONEncoder.Default()))
	ctx := logfc.New(context.Background(), logf.NewLogger(w).WithCaller(false))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logfc.Info(ctx, getMessage(0))
	}
}

func BenchmarkLogfcSyncTextWithFields(b *testing.B) {
	w := logf.NewSyncWriter(logf.LevelDebug, logf.NewWriteAppender(io.Discard, logf.NewJSONEncoder.Default()))
	ctx := logfc.New(context.Background(), logf.NewLogger(w).WithCaller(false))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logfc.Info(ctx, getMessage(0), logfFields()...)
	}
}

func BenchmarkLogfcSyncTextWithAccumulatedFields(b *testing.B) {
	w := logf.NewSyncWriter(logf.LevelDebug, logf.NewWriteAppender(io.Discard, logf.NewJSONEncoder.Default()))
	ctx := logfc.New(context.Background(), logf.NewLogger(w).WithCaller(false))
	ctx = logfc.With(ctx, logfFields()...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logfc.Info(ctx, getMessage(0))
	}
}

func BenchmarkLogfcSyncTextWithGroupAndFields(b *testing.B) {
	w := logf.NewSyncWriter(logf.LevelDebug, logf.NewWriteAppender(io.Discard, logf.NewJSONEncoder.Default()))
	ctx := logfc.New(context.Background(), logf.NewLogger(w).WithCaller(false))
	ctx = logfc.WithGroup(ctx, "http")
	ctx = logfc.With(ctx, logfFields()...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logfc.Info(ctx, getMessage(0), logf.Int("status", 200))
	}
}

// --- File I/O (async via ChannelWriter) ---

func BenchmarkLogfcBufferedFileIO(b *testing.B) {
	f, err := os.CreateTemp("", "logfc-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	w, close := logf.NewChannelWriter(logf.LevelDebug, logf.ChannelWriterConfig{
		Appender: logf.NewWriteAppender(f, logf.NewJSONEncoder.Default()),
	})
	defer close()
	ctx := logfc.New(context.Background(), logf.NewLogger(w).WithCaller(false))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logfc.Info(ctx, "request handled")
	}
}

func BenchmarkLogfcBufferedParallelFileIO(b *testing.B) {
	f, err := os.CreateTemp("", "logfc-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	w, close := logf.NewChannelWriter(logf.LevelDebug, logf.ChannelWriterConfig{
		Appender: logf.NewWriteAppender(f, logf.NewJSONEncoder.Default()),
	})
	defer close()
	ctx := logfc.New(context.Background(), logf.NewLogger(w).WithCaller(false))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logfc.Info(ctx, "request handled")
		}
	})
}

// --- File I/O (sync via SyncWriter) ---

func BenchmarkLogfcFileIO(b *testing.B) {
	f, err := os.CreateTemp("", "logfc-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	w := logf.NewSyncWriter(logf.LevelDebug, logf.NewWriteAppender(f, logf.NewJSONEncoder.Default()))
	ctx := logfc.New(context.Background(), logf.NewLogger(w).WithCaller(false))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logfc.Info(ctx, "request handled")
	}
}

func BenchmarkLogfcParallelFileIO(b *testing.B) {
	f, err := os.CreateTemp("", "logfc-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	w := logf.NewSyncWriter(logf.LevelDebug, logf.NewWriteAppender(f, logf.NewJSONEncoder.Default()))
	ctx := logfc.New(context.Background(), logf.NewLogger(w).WithCaller(false))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logfc.Info(ctx, "request handled")
		}
	})
}
