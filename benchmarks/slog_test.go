package benchmarks

import (
	"bufio"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/ssgreg/logf/v2"
)

func newSlogLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func newDisabledSlogLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
}

func slogFields() []any {
	return []any{
		"int", 42,
		"string", "hello",
		"path", "/api/v1/users",
		"latency_us", int64(1234),
		"ok", true,
	}
}

// --- Disabled path ---

func BenchmarkSlogDisabledLog(b *testing.B) {
	logger := newDisabledSlogLogger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled")
	}
}

func BenchmarkSlogDisabledLogWithFields(b *testing.B) {
	logger := newDisabledSlogLogger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled", slogFields()...)
	}
}

func BenchmarkSlogDisabledCheck(b *testing.B) {
	logger := newDisabledSlogLogger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if logger.Enabled(nil, slog.LevelInfo) {
			logger.Info("request handled")
		}
	}
}

// --- Enabled path ---

func BenchmarkSlogPlainText(b *testing.B) {
	logger := newSlogLogger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled")
	}
}

func BenchmarkSlogTextWithFields(b *testing.B) {
	logger := newSlogLogger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled", slogFields()...)
	}
}

func BenchmarkSlogTextWithAccumulatedFields(b *testing.B) {
	logger := newSlogLogger().With(slogFields()...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled")
	}
}

// --- Caller ---

func BenchmarkSlogPlainTextWithCaller(b *testing.B) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled")
	}
}

func BenchmarkSlogTextWithFieldsAndCaller(b *testing.B) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled", slogFields()...)
	}
}

// --- Logger.With ---

func BenchmarkSlogLoggerWith(b *testing.B) {
	logger := newSlogLogger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = logger.With(slogFields()...)
	}
}

// --- Parallel ---

func BenchmarkSlogParallelPlainText(b *testing.B) {
	logger := newSlogLogger()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("request handled")
		}
	})
}

// --- slog via logf handler ---

func BenchmarkSlogViaLogfDiscard(b *testing.B) {
	w, close := logf.NewChannelWriter(logf.ChannelWriterConfig{
		Appender: logf.NewWriteAppender(io.Discard, logf.NewJSONEncoder.Default()),
	})
	defer close()
	logger := slog.New(logf.NewSlogHandler(w, nil))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled")
	}
}

func BenchmarkSlogViaLogfFileIO(b *testing.B) {
	f, err := os.CreateTemp("", "sloglogf-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	w, close := logf.NewChannelWriter(logf.ChannelWriterConfig{
		Appender: logf.NewWriteAppender(f, logf.NewJSONEncoder.Default()),
	})
	defer close()
	logger := slog.New(logf.NewSlogHandler(w, nil))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled")
	}
}

func BenchmarkSlogViaLogfParallelFileIO(b *testing.B) {
	f, err := os.CreateTemp("", "sloglogf-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	w, close := logf.NewChannelWriter(logf.ChannelWriterConfig{
		Appender: logf.NewWriteAppender(f, logf.NewJSONEncoder.Default()),
	})
	defer close()
	logger := slog.New(logf.NewSlogHandler(w, nil))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("request handled")
		}
	})
}

// --- Buffered File I/O ---

func BenchmarkSlogBufferedFileIOWithFields(b *testing.B) {
	f, err := os.CreateTemp("", "slog-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	bw := bufio.NewWriterSize(f, 4096)
	defer bw.Flush()
	logger := slog.New(slog.NewJSONHandler(bw, &slog.HandlerOptions{Level: slog.LevelDebug})).With(slogFields()...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled")
	}
}

func BenchmarkSlogBufferedParallelFileIOWithFields(b *testing.B) {
	f, err := os.CreateTemp("", "slog-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	bw := bufio.NewWriterSize(f, 4096)
	defer bw.Flush()
	logger := slog.New(slog.NewJSONHandler(bw, &slog.HandlerOptions{Level: slog.LevelDebug})).With(slogFields()...)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("request handled")
		}
	})
}

func BenchmarkSlogBufferedFileIO(b *testing.B) {
	f, err := os.CreateTemp("", "slog-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	bw := bufio.NewWriterSize(f, 4096)
	defer bw.Flush()
	logger := slog.New(slog.NewJSONHandler(bw, &slog.HandlerOptions{Level: slog.LevelDebug}))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled")
	}
}

func BenchmarkSlogBufferedParallelFileIO(b *testing.B) {
	f, err := os.CreateTemp("", "slog-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	bw := bufio.NewWriterSize(f, 4096)
	defer bw.Flush()
	logger := slog.New(slog.NewJSONHandler(bw, &slog.HandlerOptions{Level: slog.LevelDebug}))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("request handled")
		}
	})
}

// --- File I/O ---

func BenchmarkSlogFileIO(b *testing.B) {
	f, err := os.CreateTemp("", "slog-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	logger := slog.New(slog.NewJSONHandler(f, &slog.HandlerOptions{Level: slog.LevelDebug}))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled")
	}
}

func BenchmarkSlogParallelFileIO(b *testing.B) {
	f, err := os.CreateTemp("", "slog-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	logger := slog.New(slog.NewJSONHandler(f, &slog.HandlerOptions{Level: slog.LevelDebug}))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("request handled")
		}
	})
}
