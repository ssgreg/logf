package benchmarks

import (
	"io"
	"testing"

	"github.com/ssgreg/logrus"
)

func newLogrusDiscard() *logrus.Logger {
	l := logrus.New()
	l.Out = io.Discard
	l.Formatter = &logrus.JSONFormatter{}
	l.Level = logrus.DebugLevel
	return l
}

func newLogrusDiscardInfo() *logrus.Logger {
	l := logrus.New()
	l.Out = io.Discard
	l.Formatter = &logrus.JSONFormatter{}
	l.Level = logrus.InfoLevel
	return l
}

// logrus (ssgreg fork) has no ReportCaller — B13 Caller skipped below

func logrusTwoScalars() logrus.Fields {
	return logrus.Fields{
		"method": "GET",
		"status": 200,
	}
}

func logrusSixScalars() logrus.Fields {
	return logrus.Fields{
		"method":     "GET",
		"status":     200,
		"path":       "/api/v1/users",
		"user_agent": "Mozilla/5.0",
		"request_id": "abc-def-123",
		"size":       1024,
	}
}

func logrusSixHeavy() logrus.Fields {
	return logrus.Fields{
		"body":      heavyBytes,
		"timestamp": heavyTime,
		"ids":       heavyInts64,
		"tags":      heavyStrings,
		"latency":   heavyDuration,
		"user":      map[string]any{"id": 123, "name": "alice"},
	}
}

// B0: DisabledLevel
func BenchmarkLogrus_DisabledLevel(b *testing.B) {
	logger := newLogrusDiscardInfo()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Debug("request handled")
	}
}

// B1: NoFields
func BenchmarkLogrus_NoFields(b *testing.B) {
	logger := newLogrusDiscard()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled")
	}
}

// B2: TwoScalars
func BenchmarkLogrus_TwoScalars(b *testing.B) {
	logger := newLogrusDiscard()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.WithFields(logrusTwoScalars()).Info("request handled")
	}
}

// B3: TwoScalarsInGroup — skipped, logrus has no native group support

// B4: SixScalars
func BenchmarkLogrus_SixScalars(b *testing.B) {
	logger := newLogrusDiscard()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.WithFields(logrusSixScalars()).Info("request handled")
	}
}

// B5: SixHeavy
func BenchmarkLogrus_SixHeavy(b *testing.B) {
	logger := newLogrusDiscard()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.WithFields(logrusSixHeavy()).Info("request handled")
	}
}

// B6: ErrorField
func BenchmarkLogrus_ErrorField(b *testing.B) {
	logger := newLogrusDiscard()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.WithError(errExample).Info("request handled")
	}
}

// B7: WithPerCall+NoFields
func BenchmarkLogrus_WithPerCall_NoFields(b *testing.B) {
	logger := newLogrusDiscard()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.WithFields(logrusTwoScalars()).Info("request handled")
	}
}

// B8: WithPerCall+TwoScalars
func BenchmarkLogrus_WithPerCall_TwoScalars(b *testing.B) {
	logger := newLogrusDiscard()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.WithFields(logrusTwoScalars()).WithFields(logrusTwoScalars()).Info("request handled")
	}
}

// B9: WithCached+NoFields
func BenchmarkLogrus_WithCached_NoFields(b *testing.B) {
	logger := newLogrusDiscard().WithFields(logrusTwoScalars())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("request handled")
	}
}

// B10: WithCached+TwoScalars
func BenchmarkLogrus_WithCached_TwoScalars(b *testing.B) {
	logger := newLogrusDiscard().WithFields(logrusTwoScalars())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.WithFields(logrusTwoScalars()).Info("request handled")
	}
}

// B11: WithBoth+TwoScalars
func BenchmarkLogrus_WithBoth_TwoScalars(b *testing.B) {
	logger := newLogrusDiscard().WithFields(logrusTwoScalars())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.WithFields(logrusTwoScalars()).WithFields(logrusTwoScalars()).Info("request handled")
	}
}

// B12: WithGroupCached+TwoScalars — skipped, logrus has no native group support

// B13: Caller+TwoScalars — skipped, ssgreg/logrus fork has no ReportCaller
