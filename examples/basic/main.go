// Basic example: NewLogger builder, logging levels, fields, groups.
package main

import (
	"context"
	"errors"
	"time"

	"github.com/ssgreg/logf/v2"
)

func main() {
	// Minimal setup: JSON to stderr, Debug level, caller enabled.
	// For colored text output, use: logf.NewLogger().EncoderFrom(logf.Text()).Build()
	logger := logf.NewLogger().Build()
	ctx := context.Background()

	// --- Levels ---
	// First output shown in full; subsequent outputs omit "ts" and "caller" for brevity.

	logger.Debug(ctx, "starting up")
	// → {"level":"debug","ts":"2025-01-15T12:00:00Z","caller":"basic/main.go:21","msg":"starting up"}

	logger.Info(ctx, "request handled", logf.String("method", "GET"), logf.Int("status", 200))
	// → {"level":"info",..,"msg":"request handled","method":"GET","status":200}

	logger.Warn(ctx, "slow query", logf.Duration("elapsed", 2*time.Second))
	// → {"level":"warn","msg":"slow query","elapsed":"2s"}

	logger.Error(ctx, "connection failed", logf.Error(errors.New("dial tcp: timeout")))
	// → {"level":"error","msg":"connection failed","error":"dial tcp: timeout"}

	// --- Accumulated fields (With) ---
	// Fields added with With are included in every subsequent log entry.
	reqLogger := logger.With(logf.String("request_id", "abc-123"), logf.String("user_id", "u42"))
	reqLogger.Info(ctx, "processing order")
	// → {"level":"info","msg":"processing order","request_id":"abc-123","user_id":"u42"}

	reqLogger.Info(ctx, "order complete", logf.Int("items", 3))
	// → {"level":"info","msg":"order complete","request_id":"abc-123","user_id":"u42","items":3}

	// --- Groups ---
	// Inline group: nested JSON object in a single log call.
	logger.Info(ctx, "http request", logf.Group("http",
		logf.String("method", "POST"),
		logf.String("path", "/api/orders"),
		logf.Int("status", 201),
	))
	// → {"level":"info","msg":"http request","http":{"method":"POST","path":"/api/orders","status":201}}

	// --- WithGroup ---
	// All subsequent fields are nested under the group key.
	httpLogger := logger.WithGroup("http")
	httpLogger.Info(ctx, "request", logf.String("method", "GET"), logf.Int("status", 200))
	// → {"level":"info","msg":"request","http":{"method":"GET","status":200}}

	// --- Named logger ---
	dbLogger := logger.WithName("db")
	dbLogger.Info(ctx, "connected", logf.String("host", "localhost:5432"))
	// → {"level":"info","logger":"db","msg":"connected","host":"localhost:5432"}

	// --- Conditional logging ---
	logger.AtLevel(ctx, logf.LevelDebug, func(log logf.LogFunc) {
		// This closure runs only if Debug is enabled.
		// Useful for expensive field computation.
		log(ctx, "debug details", logf.String("expensive", computeExpensiveValue()))
	})
	// → {"level":"debug","msg":"debug details","expensive":"computed-at-debug-only"}
}

func computeExpensiveValue() string {
	return "computed-at-debug-only"
}
