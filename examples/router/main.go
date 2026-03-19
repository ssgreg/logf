// Router example: multi-destination logging with independent encoders,
// level filters, and sync/async I/O.
package main

import (
	"context"
	"os"
	"time"

	"github.com/ssgreg/logf/v2"
)

func main() {
	// Open a log file.
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Async buffered writer for the file destination.
	fileSlab := logf.NewSlabWriter(file, 64*1024, 8,
		logf.WithFlushInterval(100*time.Millisecond),
	)

	// JSON encoder for file (full detail).
	fileEnc := logf.JSON().
		EncodeTime(logf.RFC3339NanoTimeEncoder).
		Build()

	// JSON encoder for console (compact).
	consoleEnc := logf.JSON().Build()

	// Route: file gets everything (async), console gets Info+ (sync).
	router, closeFn, err := logf.NewRouter().
		Route(fileEnc,
			logf.OutputCloser(logf.LevelDebug, fileSlab), // async, all levels
		).
		Route(consoleEnc,
			logf.Output(logf.LevelInfo, os.Stderr), // sync, Info+
		).
		Build()
	if err != nil {
		panic(err)
	}
	defer closeFn() // flushes file, syncs, closes SlabWriter

	// Wrap with ContextHandler for request-scoped fields.
	handler := logf.NewContextHandler(router)
	logger := logf.New(handler)

	ctx := context.Background()

	// Debug goes to file only (console filter is Info+):
	logger.Debug(ctx, "detailed trace", logf.String("component", "auth"))

	// Info goes to both file and console:
	logger.Info(ctx, "server started", logf.Int("port", 8080))

	// Error goes to both:
	logger.Error(ctx, "database unreachable", logf.String("host", "db:5432"))
}
