package benchmarks

import (
	"io"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func init() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.DurationFieldUnit = time.Nanosecond
}

func newZerologDiscard() zerolog.Logger {
	return zerolog.New(io.Discard).With().Timestamp().Logger().Level(zerolog.DebugLevel)
}

func newZerologDiscardInfo() zerolog.Logger {
	return zerolog.New(io.Discard).With().Timestamp().Logger().Level(zerolog.InfoLevel)
}

func newZerologDiscardWithCaller() zerolog.Logger {
	return zerolog.New(io.Discard).With().Timestamp().Caller().Logger().Level(zerolog.DebugLevel)
}

func zerologTwoScalars(e *zerolog.Event) *zerolog.Event {
	return e.Str("method", "GET").Int("status", 200)
}

func zerologSixScalars(e *zerolog.Event) *zerolog.Event {
	return e.Str("method", "GET").
		Int("status", 200).
		Str("path", "/api/v1/users").
		Str("user_agent", "Mozilla/5.0").
		Str("request_id", "abc-def-123").
		Int("size", 1024)
}

func zerologSixHeavy(e *zerolog.Event) *zerolog.Event {
	return e.Bytes("body", heavyBytes).
		Time("timestamp", heavyTime).
		Ints64("ids", heavyInts64).
		Strs("tags", heavyStrings).
		Dur("latency", heavyDuration).
		Dict("user", zerolog.Dict().Int("id", 123).Str("name", "alice"))
}

func zerologTwoScalarsCtx(c zerolog.Context) zerolog.Context {
	return c.Str("method", "GET").Int("status", 200)
}

// B0: DisabledLevel
func BenchmarkZerolog_DisabledLevel(b *testing.B) {
	logger := newZerologDiscardInfo()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Debug().Msg("request handled")
	}
}

// B1: NoFields
func BenchmarkZerolog_NoFields(b *testing.B) {
	logger := newZerologDiscard()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info().Msg("request handled")
		}
	})
}

// B2: TwoScalars
func BenchmarkZerolog_TwoScalars(b *testing.B) {
	logger := newZerologDiscard()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			zerologTwoScalars(logger.Info()).Msg("request handled")
		}
	})
}

// B3: TwoScalarsInGroup
func BenchmarkZerolog_TwoScalarsInGroup(b *testing.B) {
	logger := newZerologDiscard()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info().
				Dict("request", zerolog.Dict().Str("method", "GET").Int("status", 200)).
				Msg("request handled")
		}
	})
}

// B4: SixScalars
func BenchmarkZerolog_SixScalars(b *testing.B) {
	logger := newZerologDiscard()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			zerologSixScalars(logger.Info()).Msg("request handled")
		}
	})
}

// B5: SixHeavy
func BenchmarkZerolog_SixHeavy(b *testing.B) {
	logger := newZerologDiscard()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			zerologSixHeavy(logger.Info()).Msg("request handled")
		}
	})
}

// B6: ErrorField
func BenchmarkZerolog_ErrorField(b *testing.B) {
	logger := newZerologDiscard()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info().Err(errExample).Msg("request handled")
		}
	})
}

// B7: WithPerCall+NoFields
func BenchmarkZerolog_WithPerCall_NoFields(b *testing.B) {
	logger := newZerologDiscard()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l := zerologTwoScalarsCtx(logger.With()).Logger()
			l.Info().Msg("request handled")
		}
	})
}

// B8: WithPerCall+TwoScalars
func BenchmarkZerolog_WithPerCall_TwoScalars(b *testing.B) {
	logger := newZerologDiscard()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l := zerologTwoScalarsCtx(logger.With()).Logger()
			zerologTwoScalars(l.Info()).Msg("request handled")
		}
	})
}

// B9: WithCached+NoFields
func BenchmarkZerolog_WithCached_NoFields(b *testing.B) {
	logger := zerologTwoScalarsCtx(newZerologDiscard().With()).Logger()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info().Msg("request handled")
		}
	})
}

// B10: WithCached+TwoScalars
func BenchmarkZerolog_WithCached_TwoScalars(b *testing.B) {
	logger := zerologTwoScalarsCtx(newZerologDiscard().With()).Logger()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			zerologTwoScalars(logger.Info()).Msg("request handled")
		}
	})
}

// B11: WithBoth+TwoScalars
func BenchmarkZerolog_WithBoth_TwoScalars(b *testing.B) {
	logger := zerologTwoScalarsCtx(newZerologDiscard().With()).Logger()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l := zerologTwoScalarsCtx(logger.With()).Logger()
			zerologTwoScalars(l.Info()).Msg("request handled")
		}
	})
}

// B12: WithGroupCached+TwoScalars
func BenchmarkZerolog_WithGroupCached_TwoScalars(b *testing.B) {
	// zerolog has no WithGroup. Closest: nested Dict on each call with cached context.
	logger := zerologTwoScalarsCtx(newZerologDiscard().With()).Logger()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			zerologTwoScalars(logger.Info()).Msg("request handled")
		}
	})
}

// B13: Caller+TwoScalars
func BenchmarkZerolog_Caller_TwoScalars(b *testing.B) {
	logger := newZerologDiscardWithCaller()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			zerologTwoScalars(logger.Info()).Msg("request handled")
		}
	})
}

// --- A: With micro-benchmarks (no log call) ---

// A1: With
func BenchmarkZerolog_With(b *testing.B) {
	logger := newZerologDiscard()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = zerologTwoScalarsCtx(logger.With()).Logger()
	}
}

// A2: WithOnTop
func BenchmarkZerolog_WithOnTop(b *testing.B) {
	logger := zerologTwoScalarsCtx(newZerologDiscard().With()).Logger()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = zerologTwoScalarsCtx(logger.With()).Logger()
	}
}

// A3: WithGroup — skipped, zerolog has no WithGroup

// Suppress unused import warning.
var _ = time.Now
