package benchmarks

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/ssgreg/logf/v2"
)

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

var benchCtx = context.Background()

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

// --- File I/O (sync via UnbufferedEntryWriter) ---
// Note: UnbufferedEntryWriter is not goroutine-safe (encoder shares buffer).
// No parallel variant here — use ChannelWriter for concurrent workloads.

func BenchmarkLogfFileIO(b *testing.B) {
	f, err := os.CreateTemp("", "logf-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	w := logf.NewUnbufferedEntryWriter(logf.LevelDebug, logf.NewWriteAppender(f, logf.NewJSONEncoder.Default()))
	logger := logf.NewLogger(w).WithCaller(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info(benchCtx, "request handled")
	}
}
