// logfc example: logger-in-context pattern.
//
// Store the logger in context once, then log from anywhere without
// passing *logf.Logger through function arguments.
package main

import (
	"context"
	"os"

	"github.com/ssgreg/logf/v2"
	"github.com/ssgreg/logf/v2/logfc"
)

func main() {
	logger := logf.NewLogger().
		Level(logf.LevelInfo).
		Output(os.Stdout).
		Context().
		Build()

	// Store logger in context.
	ctx := logfc.New(context.Background(), logger)

	// Log from anywhere — just pass ctx:
	logfc.Info(ctx, "application started")
	// → {"level":"info",..,"msg":"application started"}

	// Add fields for all downstream logs:
	ctx = logfc.With(ctx, logf.String("service", "orders"))
	handleRequest(ctx)
}

func handleRequest(ctx context.Context) {
	// "service" field is included automatically:
	logfc.Info(ctx, "handling request", logf.String("path", "/api/orders"))
	// → {"level":"info",..,"msg":"handling request","service":"orders","path":"/api/orders"}

	// Add request-scoped fields:
	ctx = logfc.With(ctx, logf.String("request_id", "req-123"))

	processOrder(ctx)
}

func processOrder(ctx context.Context) {
	// All accumulated fields (service, request_id) included:
	logfc.Info(ctx, "order processed", logf.Int("items", 5))
	// → {"level":"info",..,"msg":"order processed","service":"orders","request_id":"req-123","items":5}

	// Need slog for a third-party library? Derive from context:
	slogger := logfc.Get(ctx).Slog()
	slogger.InfoContext(ctx, "slog from logfc", "key", "value")
}
