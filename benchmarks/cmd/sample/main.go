package main

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/rs/zerolog"
	logf "github.com/ssgreg/logf/v2"
	"github.com/ssgreg/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	heavyBytes    = make([]byte, 16) // shortened for readability
	heavyTime     = time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	heavyInts64   = []int64{1, 2, 3, 4, 5, 6, 7, 8}
	heavyStrings  = []string{"api", "auth", "prod", "v2"}
	heavyDuration = 42 * time.Millisecond
	errExample    = errors.New("fail")
)

type benchUser struct {
	ID   int
	Name string
}

func (u *benchUser) EncodeLogfObject(enc logf.FieldEncoder) error {
	enc.EncodeFieldInt64("id", int64(u.ID))
	enc.EncodeFieldString("name", u.Name)
	return nil
}

func (u *benchUser) MarshalZerologObject(e *zerolog.Event) {
	e.Int("id", u.ID).Str("name", u.Name)
}

// --- logf helpers ---

func logfEnc() logf.Encoder {
	return logf.NewJSONEncoder(logf.JSONEncoderConfig{
		EncodeTime:     logf.RFC3339NanoTimeEncoder,
		EncodeDuration: logf.NanoDurationEncoder,
	})
}

func logfPrint(label string, enc logf.Encoder, fields []logf.Field) {
	e := logf.Entry{Level: logf.LevelInfo, Text: "request handled", Time: time.Now(), Fields: fields}
	buf, _ := enc.Encode(e)
	fmt.Printf("  logf:    %s\n", buf.String())
	buf.Free()
}

// --- zap helpers ---

func zapLogger() (*zap.Logger, *bytes.Buffer) {
	var b bytes.Buffer
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	cfg.EncodeDuration = zapcore.NanosDurationEncoder
	enc := zapcore.NewJSONEncoder(cfg)
	core := zapcore.NewCore(enc, zapcore.AddSync(&b), zap.DebugLevel)
	return zap.New(core), &b
}

func zapPrint(label string, fields ...zap.Field) {
	l, b := zapLogger()
	l.Info("request handled", fields...)
	fmt.Printf("  zap:     %s\n", b.String())
}

// --- zerolog helpers ---

func zerologLogger() (zerolog.Logger, *bytes.Buffer) {
	var b bytes.Buffer
	return zerolog.New(&b).With().Timestamp().Logger(), &b
}

// --- slog helpers ---

func slogLogger() (*slog.Logger, *bytes.Buffer) {
	var b bytes.Buffer
	return slog.New(slog.NewJSONHandler(&b, &slog.HandlerOptions{Level: slog.LevelDebug})), &b
}

// --- logrus helpers ---

func logrusLogger() (*logrus.Logger, *bytes.Buffer) {
	var b bytes.Buffer
	l := logrus.New()
	l.Out = &b
	l.Formatter = &logrus.JSONFormatter{TimestampFormat: time.RFC3339Nano}
	l.Level = logrus.DebugLevel
	return l, &b
}

func main() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.DurationFieldUnit = time.Nanosecond

	// ---- B1: NoFields ----
	fmt.Println("=== B1: NoFields ===")
	enc := logfEnc()
	logfPrint("", enc, nil)
	zapPrint("")
	zl, zlb := zerologLogger()
	zl.Info().Msg("request handled")
	fmt.Printf("  zerolog: %s\n", zlb.String())
	sl, sb := slogLogger()
	sl.Info("request handled")
	fmt.Printf("  slog:    %s\n", sb.String())
	rl, rb := logrusLogger()
	rl.Info("request handled")
	fmt.Printf("  logrus:  %s\n", rb.String())

	// ---- B2: TwoScalars ----
	fmt.Println("=== B2: TwoScalars ===")
	logfPrint("", enc, []logf.Field{logf.String("method", "GET"), logf.Int("status", 200)})
	zapPrint("", zap.String("method", "GET"), zap.Int("status", 200))
	zl, zlb = zerologLogger()
	zl.Info().Str("method", "GET").Int("status", 200).Msg("request handled")
	fmt.Printf("  zerolog: %s\n", zlb.String())
	sl, sb = slogLogger()
	sl.LogAttrs(nil, slog.LevelInfo, "request handled", slog.String("method", "GET"), slog.Int("status", 200))
	fmt.Printf("  slog:    %s\n", sb.String())
	rl, rb = logrusLogger()
	rl.WithFields(logrus.Fields{"method": "GET", "status": 200}).Info("request handled")
	fmt.Printf("  logrus:  %s\n", rb.String())

	// ---- B4: SixScalars ----
	fmt.Println("=== B4: SixScalars ===")
	logfPrint("", enc, []logf.Field{
		logf.String("method", "GET"), logf.Int("status", 200),
		logf.String("path", "/api/v1/users"), logf.String("user_agent", "Mozilla/5.0"),
		logf.String("request_id", "abc-def-123"), logf.Int("size", 1024),
	})
	zapPrint("", zap.String("method", "GET"), zap.Int("status", 200),
		zap.String("path", "/api/v1/users"), zap.String("user_agent", "Mozilla/5.0"),
		zap.String("request_id", "abc-def-123"), zap.Int("size", 1024))
	zl, zlb = zerologLogger()
	zl.Info().Str("method", "GET").Int("status", 200).
		Str("path", "/api/v1/users").Str("user_agent", "Mozilla/5.0").
		Str("request_id", "abc-def-123").Int("size", 1024).Msg("request handled")
	fmt.Printf("  zerolog: %s\n", zlb.String())
	sl, sb = slogLogger()
	sl.LogAttrs(nil, slog.LevelInfo, "request handled",
		slog.String("method", "GET"), slog.Int("status", 200),
		slog.String("path", "/api/v1/users"), slog.String("user_agent", "Mozilla/5.0"),
		slog.String("request_id", "abc-def-123"), slog.Int("size", 1024))
	fmt.Printf("  slog:    %s\n", sb.String())
	rl, rb = logrusLogger()
	rl.WithFields(logrus.Fields{"method": "GET", "status": 200, "path": "/api/v1/users",
		"user_agent": "Mozilla/5.0", "request_id": "abc-def-123", "size": 1024}).Info("request handled")
	fmt.Printf("  logrus:  %s\n", rb.String())

	// ---- B5: SixHeavy ----
	fmt.Println("=== B5: SixHeavy ===")
	user := &benchUser{ID: 123, Name: "alice"}
	logfPrint("", enc, []logf.Field{
		logf.Bytes("body", heavyBytes), logf.Time("timestamp", heavyTime),
		logf.Ints64("ids", heavyInts64), logf.Strings("tags", heavyStrings),
		logf.Duration("latency", heavyDuration), logf.Object("user", user),
	})
	zapPrint("", zap.Binary("body", heavyBytes), zap.Time("timestamp", heavyTime),
		zap.Int64s("ids", heavyInts64), zap.Strings("tags", heavyStrings),
		zap.Duration("latency", heavyDuration),
		zap.Object("user", zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
			enc.AddInt("id", 123)
			enc.AddString("name", "alice")
			return nil
		})))
	zl, zlb = zerologLogger()
	zl.Info().Bytes("body", heavyBytes).Time("timestamp", heavyTime).
		Ints64("ids", heavyInts64).Strs("tags", heavyStrings).
		Dur("latency", heavyDuration).Object("user", user).Msg("request handled")
	fmt.Printf("  zerolog: %s\n", zlb.String())
	sl, sb = slogLogger()
	sl.LogAttrs(nil, slog.LevelInfo, "request handled",
		slog.String("body", string(heavyBytes)), slog.Time("timestamp", heavyTime),
		slog.Any("ids", heavyInts64), slog.Any("tags", heavyStrings),
		slog.Duration("latency", heavyDuration),
		slog.Group("user", slog.Int("id", 123), slog.String("name", "alice")))
	fmt.Printf("  slog:    %s\n", sb.String())
	rl, rb = logrusLogger()
	rl.WithFields(logrus.Fields{"body": heavyBytes, "timestamp": heavyTime,
		"ids": heavyInts64, "tags": heavyStrings, "latency": heavyDuration,
		"user": map[string]any{"id": 123, "name": "alice"}}).Info("request handled")
	fmt.Printf("  logrus:  %s\n", rb.String())

	// ---- B6: ErrorField ----
	fmt.Println("=== B6: ErrorField ===")
	logfPrint("", enc, []logf.Field{logf.NamedError("error", errExample)})
	zapPrint("", zap.Error(errExample))
	zl, zlb = zerologLogger()
	zl.Info().Err(errExample).Msg("request handled")
	fmt.Printf("  zerolog: %s\n", zlb.String())
	sl, sb = slogLogger()
	sl.LogAttrs(nil, slog.LevelInfo, "request handled", slog.Any("error", errExample))
	fmt.Printf("  slog:    %s\n", sb.String())
	rl, rb = logrusLogger()
	rl.WithError(errExample).Info("request handled")
	fmt.Printf("  logrus:  %s\n", rb.String())

	twoF := []logf.Field{logf.String("method", "GET"), logf.Int("status", 200)}

	// ---- B7: WithPerCall+NoFields ----
	fmt.Println("=== B7: WithPerCall+NoFields ===")
	logfPrint("", enc, twoF)
	zapPrint("", zap.String("method", "GET"), zap.Int("status", 200))
	zl, zlb = zerologLogger()
	zl.Info().Str("method", "GET").Int("status", 200).Msg("request handled")
	fmt.Printf("  zerolog: %s\n", zlb.String())
	sl, sb = slogLogger()
	sl.LogAttrs(nil, slog.LevelInfo, "request handled", slog.String("method", "GET"), slog.Int("status", 200))
	fmt.Printf("  slog:    %s\n", sb.String())
	rl, rb = logrusLogger()
	rl.WithFields(logrus.Fields{"method": "GET", "status": 200}).Info("request handled")
	fmt.Printf("  logrus:  %s\n", rb.String())

	// ---- B9: WithCached+NoFields ----
	fmt.Println("=== B9: WithCached+NoFields (With pre-cached, no per-call fields) ===")
	logfPrint("", enc, twoF) // logf: fields come from Bag, same output
	zl2, zlb2 := zapLogger()
	zl2.With(zap.String("method", "GET"), zap.Int("status", 200)).Info("request handled")
	fmt.Printf("  zap:     %s\n", zlb2.String())
	zlCtx, zlb := zerologLogger()
	zlCached := zlCtx.With().Str("method", "GET").Int("status", 200).Logger()
	zlCached.Info().Msg("request handled")
	fmt.Printf("  zerolog: %s\n", zlb.String())
	sl, sb = slogLogger()
	slCached := sl.With(slog.String("method", "GET"), slog.Int("status", 200))
	slCached.Info("request handled")
	fmt.Printf("  slog:    %s\n", sb.String())
	rl, rb = logrusLogger()
	rlCached := rl.WithFields(logrus.Fields{"method": "GET", "status": 200})
	rlCached.Info("request handled")
	fmt.Printf("  logrus:  %s\n", rb.String())

	// ---- B10: WithCached+TwoScalars ----
	fmt.Println("=== B10: WithCached+TwoScalars (cached With + 2 per-call fields) ===")
	logfPrint("", enc, append(twoF, logf.String("path", "/api"), logf.Int("size", 1024)))
	zl2, zlb2 = zapLogger()
	zl2.With(zap.String("method", "GET"), zap.Int("status", 200)).Info("request handled",
		zap.String("path", "/api"), zap.Int("size", 1024))
	fmt.Printf("  zap:     %s\n", zlb2.String())
	zlCtx, zlb = zerologLogger()
	zlCached = zlCtx.With().Str("method", "GET").Int("status", 200).Logger()
	zlCached.Info().Str("path", "/api").Int("size", 1024).Msg("request handled")
	fmt.Printf("  zerolog: %s\n", zlb.String())
	sl, sb = slogLogger()
	slCached = sl.With(slog.String("method", "GET"), slog.Int("status", 200))
	slCached.LogAttrs(nil, slog.LevelInfo, "request handled", slog.String("path", "/api"), slog.Int("size", 1024))
	fmt.Printf("  slog:    %s\n", sb.String())
	rl, rb = logrusLogger()
	rlCached = rl.WithFields(logrus.Fields{"method": "GET", "status": 200})
	rlCached.WithFields(logrus.Fields{"path": "/api", "size": 1024}).Info("request handled")
	fmt.Printf("  logrus:  %s\n", rb.String())

	// ---- B11: WithBoth+TwoScalars ----
	fmt.Println("=== B11: WithBoth+TwoScalars (cached + per-call With + 2 fields) ===")
	logfPrint("", enc, append(twoF,
		logf.String("trace_id", "abc-123"),
		logf.String("path", "/api"), logf.Int("size", 1024)))
	zl2, zlb2 = zapLogger()
	zl2.With(zap.String("method", "GET"), zap.Int("status", 200)).
		With(zap.String("trace_id", "abc-123")).
		Info("request handled", zap.String("path", "/api"), zap.Int("size", 1024))
	fmt.Printf("  zap:     %s\n", zlb2.String())
	zlCtx, zlb = zerologLogger()
	zlCached = zlCtx.With().Str("method", "GET").Int("status", 200).Logger()
	zlCached.Info().Str("trace_id", "abc-123").Str("path", "/api").Int("size", 1024).Msg("request handled")
	fmt.Printf("  zerolog: %s\n", zlb.String())
	sl, sb = slogLogger()
	slCached = sl.With(slog.String("method", "GET"), slog.Int("status", 200))
	slCached.LogAttrs(nil, slog.LevelInfo, "request handled",
		slog.String("trace_id", "abc-123"), slog.String("path", "/api"), slog.Int("size", 1024))
	fmt.Printf("  slog:    %s\n", sb.String())
	rl, rb = logrusLogger()
	rlCached = rl.WithFields(logrus.Fields{"method": "GET", "status": 200})
	rlCached.WithFields(logrus.Fields{"trace_id": "abc-123", "path": "/api", "size": 1024}).Info("request handled")
	fmt.Printf("  logrus:  %s\n", rb.String())

	// ---- B12: WithGroupCached+TwoScalars ----
	fmt.Println("=== B12: WithGroupCached+TwoScalars (WithGroup + cached With + 2 fields) ===")
	// logf: group wraps cached fields
	logfPrint("", enc, []logf.Field{logf.Group("request",
		logf.String("method", "GET"), logf.Int("status", 200),
		logf.String("path", "/api"), logf.Int("size", 1024))})
	zl2, zlb2 = zapLogger()
	zl2.With(zap.Namespace("request"), zap.String("method", "GET"), zap.Int("status", 200)).
		Info("request handled", zap.String("path", "/api"), zap.Int("size", 1024))
	fmt.Printf("  zap:     %s\n", zlb2.String())
	// zerolog: no WithGroup, use Dict
	zlCtx, zlb = zerologLogger()
	zlCached = zlCtx.With().Str("method", "GET").Int("status", 200).Logger()
	zlCached.Info().Str("path", "/api").Int("size", 1024).Msg("request handled")
	fmt.Printf("  zerolog: %s\n", zlb.String())
	sl, sb = slogLogger()
	slCached = sl.WithGroup("request").With(slog.String("method", "GET"), slog.Int("status", 200))
	slCached.LogAttrs(nil, slog.LevelInfo, "request handled", slog.String("path", "/api"), slog.Int("size", 1024))
	fmt.Printf("  slog:    %s\n", sb.String())
	// logrus: no WithGroup
	rl, rb = logrusLogger()
	rlCached = rl.WithFields(logrus.Fields{"method": "GET", "status": 200})
	rlCached.WithFields(logrus.Fields{"path": "/api", "size": 1024}).Info("request handled")
	fmt.Printf("  logrus:  %s\n", rb.String())
}
