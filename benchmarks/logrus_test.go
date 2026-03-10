package benchmarks

import (
	"bufio"
	"io"
	"os"
	"testing"

	"github.com/ssgreg/logrus"
)

func newDisabledLogrus() *logrus.Logger {
	logger := newLogrus()
	logger.Level = logrus.ErrorLevel
	return logger
}

func newLogrus() *logrus.Logger {
	return &logrus.Logger{
		Out:       io.Discard,
		Formatter: new(logrus.JSONFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.DebugLevel,
	}
}

func logrusFields() logrus.Fields {
	return logrus.Fields{
		"int":        42,
		"string":     "hello",
		"path":       "/api/v1/users",
		"latency_us": int64(1234),
		"ok":         true,
	}
}

func fakeLogrusFields() logrus.Fields {
	return logrus.Fields{
		"int":     tenInts[0],
		"ints":    tenInts,
		"string":  tenStrings[0],
		"strings": tenStrings,
		"tm":      tenTimes[0],
		// "times":   tenTimes,
		"user1": oneUser,
		// "user2":   oneUser,
		// "users":   tenUsers,
		"error": errExample,
	}
}

// --- Disabled path ---

func BenchmarkLogrusDisabledLog(b *testing.B) {
	logger := newDisabledLogrus()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled")
	}
}

// --- Enabled path ---

func BenchmarkLogrusPlainText(b *testing.B) {
	logger := newLogrus()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled")
	}
}

func BenchmarkLogrusTextWithFields(b *testing.B) {
	logger := newLogrus()
	fields := logrus.Fields{
		"int":        42,
		"string":     "hello",
		"path":       "/api/v1/users",
		"latency_us": int64(1234),
		"ok":         true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.WithFields(fields).Info("request handled")
	}
}

// --- Logger.With ---

func BenchmarkLogrusLoggerWith(b *testing.B) {
	logger := newLogrus()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = logger.WithFields(logrusFields())
	}
}

func BenchmarkLogrusLoggerWithOnTop(b *testing.B) {
	logger := newLogrus().WithFields(logrusFields())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = logger.WithFields(logrusFields())
	}
}

// --- Accumulated fields (pre-attached via WithFields) ---

func BenchmarkLogrusTextWithAccumulatedFields(b *testing.B) {
	logger := newLogrus().WithFields(logrusFields())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled")
	}
}

// --- Parallel ---

func BenchmarkLogrusParallelPlainText(b *testing.B) {
	logger := newLogrus()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("request handled")
		}
	})
}

// --- Buffered File I/O ---

func BenchmarkLogrusBufferedFileIOWithFields(b *testing.B) {
	f, err := os.CreateTemp("", "logrus-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	bw := bufio.NewWriterSize(f, 4096)
	defer bw.Flush()
	logger := logrus.New()
	logger.Out = bw
	logger.Formatter = new(logrus.JSONFormatter)
	logger.Level = logrus.DebugLevel
	entry := logger.WithFields(logrus.Fields{
		"int":        42,
		"string":     "hello",
		"path":       "/api/v1/users",
		"latency_us": int64(1234),
		"ok":         true,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entry.Info("request handled")
	}
}

func BenchmarkLogrusBufferedParallelFileIOWithFields(b *testing.B) {
	f, err := os.CreateTemp("", "logrus-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	bw := bufio.NewWriterSize(f, 4096)
	defer bw.Flush()
	logger := logrus.New()
	logger.Out = bw
	logger.Formatter = new(logrus.JSONFormatter)
	logger.Level = logrus.DebugLevel
	entry := logger.WithFields(logrus.Fields{
		"int":        42,
		"string":     "hello",
		"path":       "/api/v1/users",
		"latency_us": int64(1234),
		"ok":         true,
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			entry.Info("request handled")
		}
	})
}

func BenchmarkLogrusBufferedFileIO(b *testing.B) {
	f, err := os.CreateTemp("", "logrus-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	bw := bufio.NewWriterSize(f, 4096)
	defer bw.Flush()
	logger := &logrus.Logger{
		Out:       bw,
		Formatter: new(logrus.JSONFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.DebugLevel,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled")
	}
}

func BenchmarkLogrusBufferedParallelFileIO(b *testing.B) {
	f, err := os.CreateTemp("", "logrus-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	bw := bufio.NewWriterSize(f, 4096)
	defer bw.Flush()
	logger := &logrus.Logger{
		Out:       bw,
		Formatter: new(logrus.JSONFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.DebugLevel,
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("request handled")
		}
	})
}

// --- File I/O ---

func BenchmarkLogrusFileIO(b *testing.B) {
	f, err := os.CreateTemp("", "logrus-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	logger := &logrus.Logger{
		Out:       f,
		Formatter: new(logrus.JSONFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.DebugLevel,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled")
	}
}

func BenchmarkLogrusParallelFileIO(b *testing.B) {
	f, err := os.CreateTemp("", "logrus-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
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
			logger.Info("request handled")
		}
	})
}
