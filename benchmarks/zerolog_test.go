package benchmarks

import (
	"bufio"
	"io"
	"os"
	"testing"

	"github.com/rs/zerolog"
)

func newZerolog() zerolog.Logger {
	return zerolog.New(io.Discard).With().Timestamp().Logger()
}

func newDisabledZerolog() zerolog.Logger {
	return newZerolog().Level(zerolog.Disabled)
}

func zerologFields(e *zerolog.Event) *zerolog.Event {
	return e.
		Int("int", 42).
		Str("string", "hello").
		Str("path", "/api/v1/users").
		Int64("latency_us", 1234).
		Bool("ok", true)
}

func fakeZerologFields(e *zerolog.Event) *zerolog.Event {
	return e.
		Int("int", tenInts[0]).
		Interface("ints", tenInts).
		Str("string", tenStrings[0]).
		Interface("strings", tenStrings).
		Time("fm", tenTimes[0]).
		// Interface("times", tenTimes).
		Interface("user1", oneUser).
		// Interface("user2", oneUser).
		// Interface("users", tenUsers).
		Err(errExample)
}

func fakeZerologContext(c zerolog.Context) zerolog.Context {
	return c.
		Int("int", tenInts[0]).
		Interface("ints", tenInts).
		Str("string", tenStrings[0]).
		Interface("strings", tenStrings).
		Time("tm", tenTimes[0]).
		// Interface("times", tenTimes).
		Interface("user1", oneUser).
		// Interface("user2", oneUser).
		// Interface("users", tenUsers).
		Err(errExample)
}

// --- Disabled path ---

func BenchmarkZerologDisabledLog(b *testing.B) {
	logger := newDisabledZerolog()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().Msg("request handled")
	}
}

func BenchmarkZerologDisabledLogWithFields(b *testing.B) {
	logger := newDisabledZerolog()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		zerologFields(logger.Info()).Msg("request handled")
	}
}

// --- Enabled path ---

func BenchmarkZerologPlainText(b *testing.B) {
	logger := newZerolog()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().Msg("request handled")
	}
}

func BenchmarkZerologTextWithFields(b *testing.B) {
	logger := newZerolog()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		zerologFields(logger.Info()).Msg("request handled")
	}
}

// --- Logger.With ---

func BenchmarkZerologLoggerWith(b *testing.B) {
	logger := newZerolog()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = logger.With().
			Int("int", 42).
			Str("string", "hello").
			Str("path", "/api/v1/users").
			Int64("latency_us", 1234).
			Bool("ok", true).
			Logger()
	}
}

func BenchmarkZerologLoggerWithOnTop(b *testing.B) {
	logger := newZerolog().With().
		Int("int", 42).
		Str("string", "hello").
		Str("path", "/api/v1/users").
		Int64("latency_us", 1234).
		Bool("ok", true).
		Logger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = logger.With().
			Int("int", 42).
			Str("string", "hello").
			Str("path", "/api/v1/users").
			Int64("latency_us", 1234).
			Bool("ok", true).
			Logger()
	}
}

// --- Accumulated fields (pre-attached via With) ---

func BenchmarkZerologTextWithAccumulatedFields(b *testing.B) {
	logger := newZerolog().With().
		Int("int", 42).
		Str("string", "hello").
		Str("path", "/api/v1/users").
		Int64("latency_us", 1234).
		Bool("ok", true).
		Logger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().Msg("request handled")
	}
}

// --- Parallel ---

func BenchmarkZerologParallelPlainText(b *testing.B) {
	logger := newZerolog()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info().Msg("request handled")
		}
	})
}

// --- Buffered File I/O ---

func BenchmarkZerologBufferedFileIOWithFields(b *testing.B) {
	f, err := os.CreateTemp("", "zerolog-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	lbw := &lockedBufWriter{bw: bufio.NewWriterSize(f, 4096)}
	defer lbw.Flush()
	logger := zerolog.New(lbw).With().Timestamp().
		Int("int", 42).
		Str("string", "hello").
		Str("path", "/api/v1/users").
		Int64("latency_us", 1234).
		Bool("ok", true).
		Logger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().Msg("request handled")
	}
}

func BenchmarkZerologBufferedParallelFileIOWithFields(b *testing.B) {
	f, err := os.CreateTemp("", "zerolog-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	lbw := &lockedBufWriter{bw: bufio.NewWriterSize(f, 4096)}
	defer lbw.Flush()
	logger := zerolog.New(lbw).With().Timestamp().
		Int("int", 42).
		Str("string", "hello").
		Str("path", "/api/v1/users").
		Int64("latency_us", 1234).
		Bool("ok", true).
		Logger()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info().Msg("request handled")
		}
	})
}

func BenchmarkZerologBufferedFileIO(b *testing.B) {
	f, err := os.CreateTemp("", "zerolog-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	lbw := &lockedBufWriter{bw: bufio.NewWriterSize(f, 4096)}
	defer lbw.Flush()
	logger := zerolog.New(lbw).With().Timestamp().Logger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().Msg("request handled")
	}
}

func BenchmarkZerologBufferedParallelFileIO(b *testing.B) {
	f, err := os.CreateTemp("", "zerolog-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	lbw := &lockedBufWriter{bw: bufio.NewWriterSize(f, 4096)}
	defer lbw.Flush()
	logger := zerolog.New(lbw).With().Timestamp().Logger()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info().Msg("request handled")
		}
	})
}

// --- File I/O ---

func BenchmarkZerologFileIO(b *testing.B) {
	f, err := os.CreateTemp("", "zerolog-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	logger := zerolog.New(f).With().Timestamp().Logger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().Msg("request handled")
	}
}

func BenchmarkZerologParallelFileIO(b *testing.B) {
	f, err := os.CreateTemp("", "zerolog-bench-*.log")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	logger := zerolog.New(f).With().Timestamp().Logger()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info().Msg("request handled")
		}
	})
}
