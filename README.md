# logf

[![Go Reference](https://pkg.go.dev/badge/github.com/ssgreg/logf/v2.svg)](https://pkg.go.dev/github.com/ssgreg/logf/v2)
[![Build Status](https://github.com/ssgreg/logf/actions/workflows/go.yml/badge.svg)](https://github.com/ssgreg/logf/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/ssgreg/logf/v2)](https://goreportcard.com/report/github.com/ssgreg/logf/v2)
[![codecov](https://codecov.io/gh/ssgreg/logf/branch/master/graph/badge.svg)](https://codecov.io/gh/ssgreg/logf)

Structured logging for Go — context-aware, slog-native, fast.

## Features

- **Context-aware fields** — `logf.With(ctx, fields...)` attaches fields to context, automatically included in every log entry
- **Native slog bridge** — `logger.Slog()` returns a `*slog.Logger` sharing the same pipeline, fields, and name. Passes `testing/slogtest`
- **Router** — multi-destination fan-out with per-output level filtering and encoder groups
- **Async buffered I/O** — SlabWriter with pre-allocated slab pool, zero per-message allocations
- **WriterSlot** — placeholder writer for lazy destination initialization
- **JSON and Text encoders** — `logf.JSON()` for production, `logf.Text()` for development (colored, human-readable)
- **Builder API** — `logf.NewLogger().Level(logf.LevelInfo).Build()` for quick setup
- **Zero-alloc hot path** — 0 allocs/op across all benchmarks

## Installation

```bash
go get github.com/ssgreg/logf/v2
```

## Quick Start

```go
// Minimal — JSON to stderr, debug level, caller enabled:
logger := logf.NewLogger().Build()

// Development — colored text output:
logger := logf.NewLogger().EncoderFrom(logf.Text()).Build()
// Mar 19 14:04:02.167 [INF] request handled › method=GET status=200

// Production — custom JSON encoder, stdout, context fields:
logger := logf.NewLogger().
    Level(logf.LevelInfo).
    Output(os.Stdout).
    EncoderFrom(logf.JSON().TimeKey("time")).
    Context().
    Build()
```

## Logging

```go
ctx := context.Background()

// Basic levels:
logger.Debug(ctx, "starting up")
// → {"level":"debug","msg":"starting up"}

logger.Info(ctx, "request handled", logf.String("method", "GET"), logf.Int("status", 200))
// → {"level":"info","msg":"request handled","method":"GET","status":200}

logger.Warn(ctx, "slow query", logf.Duration("elapsed", 2*time.Second))
// → {"level":"warn","msg":"slow query","elapsed":"2s"}

logger.Error(ctx, "connection failed", logf.Error(err))
// → {"level":"error","msg":"connection failed","error":"dial tcp: timeout"}

// Accumulated fields (carried on every subsequent log):
reqLogger := logger.With(logf.String("request_id", "abc-123"))
reqLogger.Info(ctx, "processing")
// → {"level":"info","msg":"processing","request_id":"abc-123"}

// Groups (nested JSON objects):
logger.Info(ctx, "done", logf.Group("http",
    logf.String("method", "GET"),
    logf.Int("status", 200),
))
// → {"msg":"done","http":{"method":"GET","status":200}}

// WithGroup — all subsequent fields nested under a key:
httpLogger := logger.WithGroup("http")
httpLogger.Info(ctx, "req", logf.String("method", "GET"), logf.Int("status", 200))
// → {"msg":"req","http":{"method":"GET","status":200}}

// Named logger:
dbLogger := logger.WithName("db")
dbLogger.Info(ctx, "query")  // → {"logger":"db","msg":"query"}
```

## Context-aware Fields

Attach fields to `context.Context` — they appear in every log entry
automatically, without passing a derived logger through the call stack.

```go
// Middleware adds request fields to context:
func middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := logf.With(r.Context(),
            logf.String("request_id", r.Header.Get("X-Request-ID")),
            logf.String("method", r.Method),
            logf.String("path", r.URL.Path),
        )
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Deep in the call stack — fields are included automatically:
func handleOrder(ctx context.Context, orderID string) {
    logger.Info(ctx, "processing order", logf.String("order_id", orderID))
    // → {"msg":"processing order","request_id":"abc","method":"POST","path":"/orders","order_id":"123"}
}
```

Enable with `.Context()` in the builder, or wrap with `NewContextHandler`:

```go
// Builder:
logger := logf.NewLogger().Context().Build()

// Manual (for Router pipelines):
handler := logf.NewContextHandler(router)

// With external field sources (e.g., OTel trace IDs).
// The FieldSource function is called on every log entry — trace_id
// appears automatically in all logs without any manual field passing:
handler := logf.NewContextHandler(router, func(ctx context.Context) []logf.Field {
    span := trace.SpanFromContext(ctx)
    if !span.SpanContext().IsValid() {
        return nil
    }
    return []logf.Field{
        logf.String("trace_id", span.SpanContext().TraceID().String()),
    }
})
// Now every logger.Info(ctx, "...") includes trace_id if the context
// carries an active span — zero changes to application logging code.
```

## logfc — Logger in Context

The `logfc` package stores the logger in `context.Context` and provides
top-level logging functions. No need to pass `*logf.Logger` through
function arguments:

```go
import "github.com/ssgreg/logf/v2/logfc"

// Store logger in context (typically in main or middleware):
ctx = logfc.New(ctx, logger)

// Log from anywhere — just pass ctx:
logfc.Info(ctx, "order processed", logf.Int("items", 3))
// → {"level":"info",..,"msg":"order processed","items":3}

// Derive logger in context (adds fields for all downstream logs):
ctx = logfc.With(ctx, logf.String("order_id", "ord-789"))
logfc.Info(ctx, "payment complete")
// → {"level":"info",..,"msg":"payment complete","order_id":"ord-789"}

// Get the underlying logger when needed:
slogger := logfc.Get(ctx).Slog()
```

If no logger is in context, `logfc` uses `DisabledLogger` — all calls
are no-ops with zero overhead.

## slog Integration

`logger.Slog()` produces a fully integrated `*slog.Logger` — not a separate
logger, but a view of the same pipeline. It inherits accumulated `.With()`
fields, `.WithName()` identity, and the full Handler chain.

```go
// logf logger with context support:
logger := logf.NewLogger().Level(logf.LevelInfo).Context().Build()

// Derive slog logger — same pipeline, same fields:
slogger := logger.Slog()

// Pre-accumulated fields carry over:
dbLogger := logger.WithName("db").With(logf.String("component", "postgres"))
dbSlog := dbLogger.Slog()
dbSlog.Info("connected")
// → {"logger":"db","msg":"connected","component":"postgres"}
```

### Third-party libraries

Many libraries accept `*slog.Logger`. Pass `logger.Slog()` — they
get the same encoder, destination, and async I/O as your application code:

```go
// Give libraries a slog logger derived from your logf pipeline:
db := sqlx.NewClient(sqlx.WithLogger(logger.Slog()))
cache := redis.NewClient(redis.WithLogger(logger.Slog()))

// Their logs go through your Router, SlabWriter, and encoder.
// No separate log configuration per dependency.
```

### logf solves slog's context problem

`slog.InfoContext(ctx, ...)` accepts a context, but the built-in
`slog.JSONHandler` and `slog.TextHandler` ignore it — the context
is passed through but never used. logf's `ContextHandler` actually
reads fields from it. This means `slog.InfoContext(ctx, "msg")`
produces richer output through logf than through standard slog:

```go
// Standard slog — context is ignored:
slog.InfoContext(ctx, "order placed")
// → {"msg":"order placed"}

// slog through logf — context fields (see Context-aware Fields above) included:
slog.InfoContext(ctx, "order placed")
// → {"msg":"order placed","request_id":"abc-123","trace_id":"def-456"}
```

No special slog middleware, no slog-context packages. Just
`ContextHandler` in the pipeline.

### Progressive enhancement

Start with pure slog and add logf features incrementally —
each step is independent, no big-bang migration:

```go
// Step 1: faster slog backend
slog.SetDefault(slog.New(logf.NewSlogHandler(
    logf.NewSyncHandler(logf.LevelInfo, os.Stderr, logf.JSON().Build()),
)))

// Step 2: add context fields (existing slog calls gain request_id etc.)
slog.SetDefault(slog.New(logf.NewSlogHandler(
    logf.NewContextHandler(syncHandler),
)))

// Step 3: add async I/O for throughput
slog.SetDefault(slog.New(logf.NewSlogHandler(
    logf.NewContextHandler(router), // router → SlabWriter → file
)))

// Step 4 (optional): switch hot paths to logf for typed fields
logger := logf.New(handler)
logger.Info(ctx, "fast path", logf.Int("status", 200))
```

### Mixed codebase

logf and slog can coexist in the same application. Both produce
consistent output through the same pipeline:

```go
logger := logf.NewLogger().Level(logf.LevelInfo).Context().Build()

// logf in your code:
logger.Info(ctx, "handled", logf.Int("status", 200))

// slog in a dependency:
slogger := logger.Slog()
slogger.InfoContext(ctx, "query executed", "rows", 42)

// Both outputs include the same context fields (request_id, trace_id)
// and go through the same encoder, router, and destination.
```

## Router (multi-destination)

Route log entries to multiple destinations with independent encoders,
level filters, and I/O strategies.

```go
enc := logf.JSON().Build()

// Same encoder, different levels per output:
router, close, _ := logf.NewRouter().
    Route(enc,
        logf.Output(logf.LevelDebug, fileWriter),   // all levels to file
        logf.Output(logf.LevelError, alertWriter),   // errors only to alerting
    ).
    Build()
defer close()
```

**Multiple encoder groups** — one Encode call per group, shared across
outputs in that group. Different groups can use different formats:

```go
jsonEnc := logf.JSON().Build()
textEnc := logf.Text().Build()

router, close, _ := logf.NewRouter().
    Route(jsonEnc,
        logf.Output(logf.LevelDebug, fileSlab),     // JSON to file (async)
    ).
    Route(textEnc,
        logf.Output(logf.LevelInfo, os.Stderr),     // colored text to console (sync)
    ).
    Build()
```

**Mixed sync/async** — direct write for console, SlabWriter for file.
A stalled file does not block stderr output:

```go
fileSlab := logf.NewSlabWriter(file, 64*1024, 8,
    logf.WithFlushInterval(100*time.Millisecond),
)

router, close, _ := logf.NewRouter().
    Route(enc,
        logf.Output(logf.LevelDebug, fileSlab),     // async to file
        logf.Output(logf.LevelInfo, os.Stderr),     // sync to console
    ).
    Build()
defer close()
defer fileSlab.Close()

// Or transfer SlabWriter ownership to Router with OutputCloser:
router, close, _ := logf.NewRouter().
    Route(enc,
        logf.OutputCloser(logf.LevelDebug, fileSlab), // Router.close() closes fileSlab
        logf.Output(logf.LevelInfo, os.Stderr),
    ).
    Build()
defer close() // flushes and closes fileSlab automatically
```

## SlabWriter (async buffered I/O)

Decouples the caller from I/O with pre-allocated slab buffers:

```go
sw := logf.NewSlabWriter(file, 64*1024, 8,
    logf.WithFlushInterval(100*time.Millisecond),
)
defer sw.Close()

router, close, _ := logf.NewRouter().
    Route(enc, logf.Output(logf.LevelDebug, sw)).
    Build()
```

- Caller pays ~17 ns (mutex + memcpy), never blocks on I/O
- Slab pool absorbs I/O spikes without dropping messages
- Message integrity: each message is fully delivered or fully dropped, never torn

**Drop mode** — for destinations where dropping is better than blocking
(e.g., metrics pipeline, non-critical remote collector):

```go
sw := logf.NewSlabWriter(conn, 64*1024, 8,
    logf.WithDropOnFull(),
    logf.WithFlushInterval(100*time.Millisecond),
    logf.WithErrorWriter(os.Stderr), // log I/O errors to stderr
)
```

When the I/O goroutine can't keep up and all slabs are in flight,
`Write` discards the current slab instead of blocking. The caller
is never delayed — the slab is reused immediately.

**Monitoring** — `Stats()` provides a snapshot for metrics scrapers:

```go
stats := sw.Stats()
// stats.QueuedSlabs  — slabs waiting for I/O
// stats.FreeSlabs    — slabs available in pool
// stats.Dropped      — total messages dropped (dropOnFull mode)
// stats.Written      — total messages accepted
// stats.WriteErrors  — total I/O errors

// Example: expose as Prometheus metrics
droppedGauge.Set(float64(stats.Dropped))
queuedGauge.Set(float64(stats.QueuedSlabs))
```

See [docs/BUFFERING.md](docs/BUFFERING.md) for capacity planning and benchmarks.

## WriterSlot (lazy destination)

Connect a destination after logger creation — useful when the output
is not available at startup:

```go
slot := logf.NewWriterSlot(logf.WithSlotBuffer(4096))
logger := logf.NewLogger().Output(slot).Build()

// Logger works immediately — writes are buffered.
logger.Info(ctx, "starting up")

// Later, when destination is ready:
slot.Set(file)
// Buffered data is flushed, subsequent writes go directly to file.
```

## Why logf

slog from the standard library is sufficient for most applications. logf
targets scenarios where slog falls short:

- **High-throughput logging** — encoding is parallel across goroutines,
  writes are memcpy into pre-allocated slabs (~17 ns). ~2× faster than
  slog under parallel file I/O.
- **Unstable I/O** — slab pool decouples callers from I/O. Under
  simulated slow disk, logf p99 = 71µs vs slog p99 = 2.5ms.
- **Request-scoped fields via context** — slog passes context through
  but built-in handlers ignore it. logf actually reads fields from it.
- **Decoupled encoding and I/O** — Router encodes once and fans out to
  multiple writers independently.

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for design details.

## Who uses logf

- [Acronis](https://www.acronis.com) — global cybersecurity and data protection platform

## Testing

```go
// Discard all logs (silent tests):
logger := logf.DisabledLogger()

// Capture logs to a buffer for assertions:
var buf bytes.Buffer
logger := logf.NewLogger().Output(&buf).Build()
logger.Info(ctx, "hello")
// buf.String() contains JSON output

// Send logs to testing.T (visible with -v or on failure):
type testWriter struct{ t testing.TB }
func (w testWriter) Write(p []byte) (int, error) {
    w.t.Helper()
    w.t.Log(strings.TrimRight(string(p), "\n"))
    return len(p), nil
}
logger := logf.NewLogger().Output(testWriter{t}).Build()
```

## Log Rotation

logf does not handle log rotation — use `lumberjack` or OS-level `logrotate`:

```go
import "gopkg.in/natefinch/lumberjack.v2"

rotator := &lumberjack.Logger{
    Filename:   "/var/log/myapp.log",
    MaxSize:    100, // MB
    MaxBackups: 3,
    MaxAge:     28,
}
sw := logf.NewSlabWriter(rotator, 64*1024, 8)
```

## Caveats

- **One allocation per log call with fields.** `logger.Info(ctx, "msg", field1, field2)`
  allocates a `[]Field` slice on the heap (Go variadic argument semantics).
  This is the single `1 allocs/op` visible in benchmarks. Calls without
  fields (`logger.Info(ctx, "msg")`) are zero-alloc.

- **Oversized messages allocate (SlabWriter only).** When using SlabWriter,
  messages larger than `slabSize` require `make([]byte, len(p))` + copy.
  Typical log entries (100–500 bytes) with typical slab sizes (16–64 KB)
  never hit this path. Sync handlers are not affected.

## Documentation

- [Architecture & design philosophy](docs/ARCHITECTURE.md)
- [Buffering strategies & capacity planning](docs/BUFFERING.md)
- [Performance experiments](docs/EXPERIMENTS.md)
- [TODO & roadmap](docs/TODO.md)
