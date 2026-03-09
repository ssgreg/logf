package benchmarks

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/ssgreg/logf/v2"
)

func logfFields() []logf.Field {
	return []logf.Field{
		logf.Int("int", 42),
		logf.String("string", "hello"),
		logf.String("path", "/api/v1/users"),
		logf.Int64("latency_us", 1234),
		logf.Bool("ok", true),
	}
}

func fakeFields() []logf.Field {
	return []logf.Field{
		logf.Int("int", tenInts[0]),
		logf.ConstInts("ints", tenInts),
		logf.String("string", tenStrings[0]),
		logf.Strings("strings", tenStrings),
		logf.Time("tm", tenTimes[0]),
		// logf.Duration("dur", time.Second),
		// logf.Durations("durs", []time.Duration{time.Second, time.Millisecond}),
		// // logf.Any("times", tenTimes),
		logf.Object("user1", oneUser),
		// // logf.Any("user2", oneUser),
		// // logf.Any("users", tenUsers),
		logf.NamedError("error", errExample),
	}
}

func newLogger(l logf.Level) (*logf.Logger, logf.ChannelWriterCloseFunc) {
	encoder := logf.NewJSONEncoder.Default()
	w, close := logf.NewChannelWriter(l, logf.ChannelWriterConfig{
		Appender: logf.NewWriteAppender(io.Discard, encoder),
	})

	return logf.NewLogger(w), close
}

func newSyncLogger(l logf.Level) *logf.Logger {
	encoder := logf.NewJSONEncoder.Default()
	w := logf.NewSyncWriter(l, logf.NewWriteAppender(io.Discard, encoder))
	return logf.NewLogger(w)
}

var benchCtx = context.Background()

// --- Disabled path ---

func BenchmarkLogfDisabledLog(b *testing.B) {
	logger := logf.NewDisabledLogger()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info(ctx, "request handled")
	}
}

func BenchmarkLogfDisabledLogWithFields(b *testing.B) {
	logger := logf.NewDisabledLogger()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info(ctx, "request handled", logfFields()...)
	}
}

// --- File I/O (async via ChannelWriter) ---

func BenchmarkLogfBufferedFileIOWithFields(b *testing.B) {
	f, err := os.CreateTemp("", "logf-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	w, close := logf.NewChannelWriter(logf.LevelDebug, logf.ChannelWriterConfig{
		Appender: logf.NewWriteAppender(f, logf.NewJSONEncoder.Default()),
	})
	defer close()
	logger := logf.NewLogger(w).WithCaller(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info(benchCtx, "request handled", fakeFields()...)
	}
}

func BenchmarkLogfBufferedParallelFileIOWithFields(b *testing.B) {
	f, err := os.CreateTemp("", "logf-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	w, close := logf.NewChannelWriter(logf.LevelDebug, logf.ChannelWriterConfig{
		Appender: logf.NewWriteAppender(f, logf.NewJSONEncoder.Default()),
	})
	defer close()
	logger := logf.NewLogger(w).WithCaller(false)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info(benchCtx, "request handled", fakeFields()...)
		}
	})
}

func BenchmarkLogfBufferedFileIO(b *testing.B) {
	f, err := os.CreateTemp("", "logf-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	w, close := logf.NewChannelWriter(logf.LevelDebug, logf.ChannelWriterConfig{
		Appender: logf.NewWriteAppender(f, logf.NewJSONEncoder.Default()),
	})
	defer close()
	logger := logf.NewLogger(w).WithCaller(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info(benchCtx, "request handled")
	}
}

func BenchmarkLogfBufferedParallelFileIO(b *testing.B) {
	f, err := os.CreateTemp("", "logf-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	w, close := logf.NewChannelWriter(logf.LevelDebug, logf.ChannelWriterConfig{
		Appender: logf.NewWriteAppender(f, logf.NewJSONEncoder.Default()),
	})
	defer close()
	logger := logf.NewLogger(w).WithCaller(false)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info(benchCtx, "request handled")
		}
	})
}

// --- File I/O (sync via SyncWriter) ---

func BenchmarkLogfFileIOWithFields(b *testing.B) {
	f, err := os.CreateTemp("", "logf-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	w := logf.NewSyncWriter(logf.LevelDebug, logf.NewWriteAppender(f, logf.NewJSONEncoder.Default()))
	logger := logf.NewLogger(w).WithCaller(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info(benchCtx, "request handled", fakeFields()...)
	}
}

func BenchmarkLogfParallelFileIOWithFields(b *testing.B) {
	f, err := os.CreateTemp("", "logf-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	w := logf.NewSyncWriter(logf.LevelDebug, logf.NewWriteAppender(f, logf.NewJSONEncoder.Default()))
	logger := logf.NewLogger(w).WithCaller(false)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info(benchCtx, "request handled", fakeFields()...)
		}
	})
}

func BenchmarkLogfFileIO(b *testing.B) {
	f, err := os.CreateTemp("", "logf-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	w := logf.NewSyncWriter(logf.LevelDebug, logf.NewWriteAppender(f, logf.NewJSONEncoder.Default()))
	logger := logf.NewLogger(w).WithCaller(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info(benchCtx, "request handled")
	}
}

func BenchmarkLogfParallelFileIO(b *testing.B) {
	f, err := os.CreateTemp("", "logf-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	w := logf.NewSyncWriter(logf.LevelDebug, logf.NewWriteAppender(f, logf.NewJSONEncoder.Default()))
	logger := logf.NewLogger(w).WithCaller(false)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info(benchCtx, "request handled")
		}
	})
}
