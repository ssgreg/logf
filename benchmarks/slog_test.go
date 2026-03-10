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

func fakeSlogFields() []any {
	return []any{
		slog.Int("int", tenInts[0]),
		slog.Any("ints", tenInts),
		slog.String("string", tenStrings[0]),
		slog.Any("strings", tenStrings),
		slog.Time("tm", tenTimes[0]),
		slog.Any("user1", oneUser),
		slog.Any("error", errExample),
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

func BenchmarkSlogTextWithGroupAndFields(b *testing.B) {
	logger := newSlogLogger().WithGroup("http").With(slogFields()...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled", "status", 200)
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

func BenchmarkSlogLoggerWithOnTop(b *testing.B) {
	logger := newSlogLogger().With(slogFields()...)

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
	w, close := logf.NewChannelWriter(logf.LevelDebug, logf.ChannelWriterConfig{
		Appender: logf.NewWriteAppender(io.Discard, logf.NewJSONEncoder.Default()),
	})
	defer close()
	logger := slog.New(logf.NewSlogHandler(w))

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

	w, close := logf.NewChannelWriter(logf.LevelDebug, logf.ChannelWriterConfig{
		Appender: logf.NewWriteAppender(f, logf.NewJSONEncoder.Default()),
	})
	defer close()
	logger := slog.New(logf.NewSlogHandler(w))

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

	w, close := logf.NewChannelWriter(logf.LevelDebug, logf.ChannelWriterConfig{
		Appender: logf.NewWriteAppender(f, logf.NewJSONEncoder.Default()),
	})
	defer close()
	logger := slog.New(logf.NewSlogHandler(w))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("request handled")
		}
	})
}

// --- Logger.Slog() bridge ---

func BenchmarkLogfSlogDiscard(b *testing.B) {
	w, close := logf.NewChannelWriter(logf.LevelDebug, logf.ChannelWriterConfig{
		Appender: logf.NewWriteAppender(io.Discard, logf.NewJSONEncoder.Default()),
	})
	defer close()
	logger := logf.NewLogger(w).WithCaller(false).Slog()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled")
	}
}

func BenchmarkLogfSlogTextWithFields(b *testing.B) {
	w, close := logf.NewChannelWriter(logf.LevelDebug, logf.ChannelWriterConfig{
		Appender: logf.NewWriteAppender(io.Discard, logf.NewJSONEncoder.Default()),
	})
	defer close()
	logger := logf.NewLogger(w).WithCaller(false).Slog()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled", slogFields()...)
	}
}

func BenchmarkLogfSlogTextWithAccumulatedFields(b *testing.B) {
	w, close := logf.NewChannelWriter(logf.LevelDebug, logf.ChannelWriterConfig{
		Appender: logf.NewWriteAppender(io.Discard, logf.NewJSONEncoder.Default()),
	})
	defer close()
	logger := logf.NewLogger(w).WithCaller(false).Slog().With(slogFields()...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled")
	}
}

func BenchmarkLogfSlogTextWithGroupAndFields(b *testing.B) {
	w, close := logf.NewChannelWriter(logf.LevelDebug, logf.ChannelWriterConfig{
		Appender: logf.NewWriteAppender(io.Discard, logf.NewJSONEncoder.Default()),
	})
	defer close()
	logger := logf.NewLogger(w).WithCaller(false).Slog().
		WithGroup("http").With(slogFields()...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled", "status", 200)
	}
}

func BenchmarkLogfSlogSyncPlainText(b *testing.B) {
	w := logf.NewSyncWriter(logf.LevelDebug, logf.NewWriteAppender(io.Discard, logf.NewJSONEncoder.Default()))
	logger := logf.NewLogger(w).WithCaller(false).Slog()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled")
	}
}

func BenchmarkLogfSlogSyncTextWithFields(b *testing.B) {
	w := logf.NewSyncWriter(logf.LevelDebug, logf.NewWriteAppender(io.Discard, logf.NewJSONEncoder.Default()))
	logger := logf.NewLogger(w).WithCaller(false).Slog()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled", slogFields()...)
	}
}

func BenchmarkLogfSlogSyncTextWithAccumulatedFields(b *testing.B) {
	w := logf.NewSyncWriter(logf.LevelDebug, logf.NewWriteAppender(io.Discard, logf.NewJSONEncoder.Default()))
	logger := logf.NewLogger(w).WithCaller(false).Slog().With(slogFields()...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled")
	}
}

func BenchmarkLogfSlogSyncTextWithGroupAndFields(b *testing.B) {
	w := logf.NewSyncWriter(logf.LevelDebug, logf.NewWriteAppender(io.Discard, logf.NewJSONEncoder.Default()))
	logger := logf.NewLogger(w).WithCaller(false).Slog().
		WithGroup("http").With(slogFields()...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled", "status", 200)
	}
}

func BenchmarkLogfSlogBufferedFileIO(b *testing.B) {
	f, err := os.CreateTemp("", "logfslog-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	w, close := logf.NewChannelWriter(logf.LevelDebug, logf.ChannelWriterConfig{
		Appender: logf.NewWriteAppender(f, logf.NewJSONEncoder.Default()),
	})
	defer close()
	logger := logf.NewLogger(w).WithCaller(false).Slog()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled")
	}
}

func BenchmarkLogfSlogBufferedParallelFileIO(b *testing.B) {
	f, err := os.CreateTemp("", "logfslog-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	w, close := logf.NewChannelWriter(logf.LevelDebug, logf.ChannelWriterConfig{
		Appender: logf.NewWriteAppender(f, logf.NewJSONEncoder.Default()),
	})
	defer close()
	logger := logf.NewLogger(w).WithCaller(false).Slog()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("request handled")
		}
	})
}

func BenchmarkLogfSlogFileIO(b *testing.B) {
	f, err := os.CreateTemp("", "logfslog-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	w := logf.NewSyncWriter(logf.LevelDebug, logf.NewWriteAppender(f, logf.NewJSONEncoder.Default()))
	logger := logf.NewLogger(w).WithCaller(false).Slog()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled")
	}
}

func BenchmarkLogfSlogParallelFileIO(b *testing.B) {
	f, err := os.CreateTemp("", "logfslog-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	w := logf.NewSyncWriter(logf.LevelDebug, logf.NewWriteAppender(f, logf.NewJSONEncoder.Default()))
	logger := logf.NewLogger(w).WithCaller(false).Slog()

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
