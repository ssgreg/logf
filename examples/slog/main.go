// slog integration example: logf as backend for slog, third-party
// libraries, and mixed logf/slog usage in the same application.
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/ssgreg/logf/v2"
)

func main() {
	// Setup logf with context support.
	logger := logf.NewLogger().
		Level(logf.LevelInfo).
		Output(os.Stdout).
		Context().
		Build()

	ctx := context.Background()

	// --- 1. Derive slog logger from logf ---
	// Same pipeline, same fields, same encoder.
	slogger := logger.Slog()
	slogger.Info("hello from slog", "key", "value")

	// --- 2. Accumulated fields carry over ---
	dbLogger := logger.WithName("db").With(logf.String("component", "postgres"))
	dbSlog := dbLogger.Slog()
	dbSlog.Info("connected", "host", "localhost:5432")
	// → {"logger":"db","msg":"connected","component":"postgres","host":"localhost:5432"}

	// --- 3. Context fields work through slog ---
	ctx = logf.With(ctx, logf.String("request_id", "abc-123"))

	// logf sees the context fields:
	logger.Info(ctx, "logf log", logf.Int("status", 200))
	// → {"msg":"logf log","request_id":"abc-123","status":200}

	// slog also sees them (via ContextHandler):
	slogger.InfoContext(ctx, "slog log", "status", 200)
	// → {"msg":"slog log","request_id":"abc-123","status":200}

	// --- 4. Set as default slog backend ---
	// All slog calls across the application go through logf's pipeline.
	slog.SetDefault(slogger)
	slog.Info("using default slog", "via", "logf")

	// --- 5. Pass to third-party libraries ---
	// Any library accepting *slog.Logger gets the same pipeline.
	thirdPartyLibrary(dbSlog, ctx)
}

// thirdPartyLibrary simulates a library that accepts *slog.Logger.
func thirdPartyLibrary(logger *slog.Logger, ctx context.Context) {
	// The library doesn't know about logf, but context fields
	// (request_id) appear in its logs automatically.
	logger.InfoContext(ctx, "query executed", "rows", 42)
	// → {"logger":"db","msg":"query executed","component":"postgres","request_id":"abc-123","rows":42}
}
