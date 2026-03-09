package benchmarks

import (
	"io"
	"os"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func fakeZapFields() []zapcore.Field {
	return []zap.Field{
		zap.Int("int", tenInts[0]),
		zap.Ints("ints", tenInts),
		zap.String("string", tenStrings[0]),
		zap.Strings("strings", tenStrings),
		zap.Time("fm", tenTimes[0]),
		// zap.Times("times", tenTimes),
		zap.Object("user1", oneUser),
		// zap.Any("user2", oneUser),
		// zap.Any("users", tenUsers),
		zap.Error(errExample),
	}
}

func zapFields() []zap.Field {
	return []zap.Field{
		zap.Int("int", 42),
		zap.String("string", "hello"),
		zap.String("path", "/api/v1/users"),
		zap.Int64("latency_us", 1234),
		zap.Bool("ok", true),
	}
}

// A Syncer is a spy for the Sync portion of zapcore.WriteSyncer.
type Syncer struct {
	err    error
	called bool
}

// SetError sets the error that the Sync method will return.
func (s *Syncer) SetError(err error) {
	s.err = err
}

// Sync records that it was called, then returns the user-supplied error (if
// any).
func (s *Syncer) Sync() error {
	s.called = true
	return s.err
}

// Called reports whether the Sync method was called.
func (s *Syncer) Called() bool {
	return s.called
}

// A Discarder sends all writes to io.Discard.
type Discarder struct{ Syncer }

// Write implements io.Writer.
func (d *Discarder) Write(b []byte) (int, error) {
	return io.Discard.Write(b)
}

func newZapLogger(lvl zapcore.Level) *zap.Logger {
	ec := zap.NewProductionEncoderConfig()
	ec.EncodeDuration = zapcore.NanosDurationEncoder
	ec.EncodeTime = zapcore.ISO8601TimeEncoder
	enc := zapcore.NewJSONEncoder(ec)
	return zap.New(zapcore.NewCore(
		enc,
		&Discarder{},
		lvl,
	))
}

func newZapFileLogger(f *os.File) *zap.Logger {
	ec := zap.NewProductionEncoderConfig()
	ec.EncodeDuration = zapcore.NanosDurationEncoder
	ec.EncodeTime = zapcore.ISO8601TimeEncoder
	enc := zapcore.NewJSONEncoder(ec)
	return zap.New(zapcore.NewCore(enc, zapcore.AddSync(f), zap.DebugLevel))
}

func newZapBufferedFileLogger(f *os.File) (*zap.Logger, func()) {
	ec := zap.NewProductionEncoderConfig()
	ec.EncodeDuration = zapcore.NanosDurationEncoder
	ec.EncodeTime = zapcore.ISO8601TimeEncoder
	enc := zapcore.NewJSONEncoder(ec)
	bws := &zapcore.BufferedWriteSyncer{WS: zapcore.AddSync(f)}
	logger := zap.New(zapcore.NewCore(enc, bws, zap.DebugLevel))
	return logger, func() { bws.Stop() }
}

// --- Disabled path ---

func BenchmarkZapDisabledLog(b *testing.B) {
	logger := newZapLogger(zap.ErrorLevel)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled")
	}
}

func BenchmarkZapDisabledLogWithFields(b *testing.B) {
	logger := newZapLogger(zap.ErrorLevel)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled", zapFields()...)
	}
}

func BenchmarkZapDisabledCheck(b *testing.B) {
	logger := newZapLogger(zap.ErrorLevel)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if m := logger.Check(zap.InfoLevel, "request handled"); m != nil {
			m.Write()
		}
	}
}

// --- Enabled path ---

func BenchmarkZapPlainText(b *testing.B) {
	logger := newZapLogger(zap.DebugLevel)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled")
	}
}

func BenchmarkZapTextWithFields(b *testing.B) {
	logger := newZapLogger(zap.DebugLevel)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled", zapFields()...)
	}
}

func BenchmarkZapTextWithAccumulatedFields(b *testing.B) {
	logger := newZapLogger(zap.DebugLevel).With(zapFields()...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled")
	}
}

// --- Caller ---

func BenchmarkZapPlainTextWithCaller(b *testing.B) {
	logger := newZapLogger(zap.DebugLevel).WithOptions(zap.AddCaller())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled")
	}
}

func BenchmarkZapTextWithFieldsAndCaller(b *testing.B) {
	logger := newZapLogger(zap.DebugLevel).WithOptions(zap.AddCaller())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled", zapFields()...)
	}
}

// --- Logger.With ---

func BenchmarkZapLoggerWith(b *testing.B) {
	logger := newZapLogger(zap.DebugLevel)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = logger.With(zapFields()...)
	}
}

// --- Parallel ---

func BenchmarkZapParallelPlainText(b *testing.B) {
	logger := newZapLogger(zap.DebugLevel)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("request handled")
		}
	})
}

// --- File I/O ---

func BenchmarkZapBufferedFileIOWithFields(b *testing.B) {
	f, err := os.CreateTemp("", "zap-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	logger, stop := newZapBufferedFileLogger(f)
	defer stop()
	logger = logger.With(zapFields()...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled")
	}
}

func BenchmarkZapBufferedParallelFileIOWithFields(b *testing.B) {
	f, err := os.CreateTemp("", "zap-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	logger, stop := newZapBufferedFileLogger(f)
	defer stop()
	logger = logger.With(zapFields()...)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("request handled")
		}
	})
}

func BenchmarkZapBufferedFileIO(b *testing.B) {
	f, err := os.CreateTemp("", "zap-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	logger, stop := newZapBufferedFileLogger(f)
	defer stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled")
	}
}

func BenchmarkZapBufferedParallelFileIO(b *testing.B) {
	f, err := os.CreateTemp("", "zap-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	logger, stop := newZapBufferedFileLogger(f)
	defer stop()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("request handled")
		}
	})
}

func BenchmarkZapFileIO(b *testing.B) {
	f, err := os.CreateTemp("", "zap-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	logger := newZapFileLogger(f)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled")
	}
}

func BenchmarkZapParallelFileIO(b *testing.B) {
	f, err := os.CreateTemp("", "zap-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	logger := newZapFileLogger(f)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("request handled")
		}
	})
}
