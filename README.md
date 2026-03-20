# › logf

[![Go Reference](https://pkg.go.dev/badge/github.com/ssgreg/logf/v2.svg)](https://pkg.go.dev/github.com/ssgreg/logf/v2)
[![Build Status](https://github.com/ssgreg/logf/actions/workflows/go.yml/badge.svg)](https://github.com/ssgreg/logf/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/ssgreg/logf/v2)](https://goreportcard.com/report/github.com/ssgreg/logf/v2)
[![codecov](https://codecov.io/gh/ssgreg/logf/branch/master/graph/badge.svg)](https://codecov.io/gh/ssgreg/logf)

Structured logging for Go — context-aware, slog-native, fast.

Used in production at [Acronis](https://www.acronis.com).

## So you want to log things

You already have `slog`. It works. But when your service hits 50K req/sec
and p99 spikes every time the log collector hiccups — or when you're
tired of threading loggers through seventeen function arguments just to
get a `request_id` — that's where logf comes in.

## What's in the box

- **Context-aware fields** — attach fields to `context.Context`, they show up in every log entry magically. No more threading loggers through your entire call stack like some kind of dependency injection nightmare.
- **Native slog bridge** — `logger.Slog()` gives you a real `*slog.Logger` that shares everything. Fields, name, pipeline. It's not a wrapper, it's the same logger wearing a different hat.
- **Router** — send logs to multiple destinations. JSON to file, colored text to console, errors to alerting. Each destination gets its own encoder and level filter. A stalled Kibana doesn't block your stderr.
- **SlabWriter** — async buffered I/O that copies your log into a pre-allocated slab in ~17 ns and moves on. A background goroutine handles the actual writing. Your HTTP handler never waits for disk.
- **WriterSlot** — don't know where you're logging to yet? No problem. Start logging, connect the destination later. Early logs are buffered.
- **JSON and Text encoders** — `logf.JSON()` for machines, `logf.Text()` for humans. The text encoder has colors, italics, and a `›` separator that makes your terminal look like it went to design school.
- **Builder API** — one line to start, chain methods to customize. No config structs with 47 fields.
- **Zero-alloc hot path** — the only allocation is Go's variadic `[]Field` slice. Everything else is pooled, pre-allocated, or stack-allocated.

## Getting started

```bash
go get github.com/ssgreg/logf/v2
```

Two lines to logging:

```go
logger := logf.NewLogger().Build()
logger.Info(ctx, "hello, world", logf.String("from", "logf"))
// → {"level":"info","ts":"2026-03-19T14:04:02Z","caller":"main.go:10","msg":"hello, world","from":"logf"}
```

Want colors? Say no more:

```go
logger := logf.NewLogger().EncoderFrom(logf.Text()).Build()
// Mar 19 14:04:02.167 [INF] hello, world › from=logf → main.go:10
```

Going to production? Crank it up:

```go
logger := logf.NewLogger().
    Level(logf.LevelInfo).
    Output(os.Stdout).
    Build()
```

## Logging

```go
ctx := context.Background()

// The classics:
logger.Debug(ctx, "starting up")
// → {"level":"debug","msg":"starting up"}

logger.Info(ctx, "request handled", logf.String("method", "GET"), logf.Int("status", 200))
// → {"level":"info","msg":"request handled","method":"GET","status":200}

logger.Warn(ctx, "slow query", logf.Duration("elapsed", 2*time.Second))
// → {"level":"warn","msg":"slow query","elapsed":"2s"}

logger.Error(ctx, "connection failed", logf.Error(err))
// → {"level":"error","msg":"connection failed","error":"dial tcp: timeout"}
```

**Accumulated fields** — set once, included forever:

```go
reqLogger := logger.With(logf.String("request_id", "abc-123"))
reqLogger.Info(ctx, "processing")
// → {"level":"info","msg":"processing","request_id":"abc-123"}

reqLogger.Info(ctx, "done", logf.Int("items", 3))
// → {"level":"info","msg":"done","request_id":"abc-123","items":3}
```

**Groups** — nest fields under a key:

```go
logger.Info(ctx, "done", logf.Group("http",
    logf.String("method", "GET"),
    logf.Int("status", 200),
))
// → {"msg":"done","http":{"method":"GET","status":200}}

// Or permanently with WithGroup:
httpLogger := logger.WithGroup("http")
httpLogger.Info(ctx, "req", logf.String("method", "GET"), logf.Int("status", 200))
// → {"msg":"req","http":{"method":"GET","status":200}}
```

**Named loggers** — know who's talking:

```go
dbLogger := logger.WithName("db")
dbLogger.Info(ctx, "connected")
// → {"logger":"db","msg":"connected"}
```

## Context-aware fields (the killer feature)

Here's the thing about logging in real applications: you want `request_id`
in every single log entry. With most loggers, that means passing a derived
logger through every function. With logf, you put fields in the context
and forget about them:

```go
// In your middleware — add fields once:
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

// Somewhere deep in the call stack — fields are just there:
func handleOrder(ctx context.Context, orderID string) {
    logger.Info(ctx, "processing order", logf.String("order_id", orderID))
    // → {"msg":"processing order","request_id":"abc","method":"POST","path":"/orders","order_id":"123"}
}
```

Enable it with `.Context()` in the builder:

```go
logger := logf.NewLogger().Context().Build()
```

Want to automatically extract trace IDs from OpenTelemetry spans?
Write a `FieldSource` and pass it to `.Context()`:

```go
// Define once:
func otelTraceSource(ctx context.Context) []logf.Field {
    span := trace.SpanFromContext(ctx)
    if !span.SpanContext().IsValid() {
        return nil
    }
    return []logf.Field{
        logf.String("trace_id", span.SpanContext().TraceID().String()),
    }
}

// Plug it in:
logger := logf.NewLogger().Context(otelTraceSource).Build()
```

That's it. From now on, whenever a context carries an active OTel span,
`trace_id` shows up in every log entry. You didn't change a single
logging call in your application.

## logfc — when you don't want to pass the logger at all

The `logfc` package puts the logger in the context. Not a global
singleton — a real logger that picks up new fields as the request
travels deeper through your code. Each layer adds its own details,
and by the time you log something ten functions down, the entry
carries the full story of how it got there:

```go
import "github.com/ssgreg/logf/v2/logfc"

// In main or middleware:
ctx = logfc.New(ctx, logger)

// Anywhere else — no logger argument needed:
logfc.Info(ctx, "order processed", logf.Int("items", 3))

// Add fields for everything downstream:
ctx = logfc.With(ctx, logf.String("order_id", "ord-789"))
logfc.Info(ctx, "payment complete")
// → includes order_id automatically

// Need slog? Pull it out:
slogger := logfc.Get(ctx).Slog()
```

If no logger is in context, everything is a no-op. Zero overhead. No panics.

## slog integration (they're best friends)

`logger.Slog()` doesn't create a new logger. It returns a `*slog.Logger`
that IS your logf logger, just with slog's API. Same fields, same name,
same pipeline, same destination. Log with either one — the output is
identical.

```go
// These two produce the same output:
logger.Info(ctx, "hello", logf.Int("n", 42))
logger.Slog().InfoContext(ctx, "hello", "n", 42)
```

**Give it to your dependencies:**

```go
db := sqlx.NewClient(sqlx.WithLogger(logger.Slog()))
cache := redis.NewClient(redis.WithLogger(logger.Slog()))
// Their logs go through your pipeline. One config for everything.
```

**Here's the neat part** — slog has `InfoContext(ctx, ...)` but the
built-in handlers completely ignore the context. logf actually reads
fields from it:

```go
// Standard slog — context is decoration:
slog.InfoContext(ctx, "order placed")
// → {"msg":"order placed"}

// slog through logf — context fields included:
slog.InfoContext(ctx, "order placed")
// → {"msg":"order placed","request_id":"abc-123","trace_id":"def-456"}
```

**Progressive enhancement** — start with slog, add logf features one at a time:

```go
// Step 1: just a faster backend — JSON to stderr
sync := logf.NewSyncHandler(logf.LevelInfo, os.Stderr, logf.JSON().Build())
slog.SetDefault(slog.New(logf.NewSlogHandler(sync)))

// Step 2: add context fields — existing slog calls magically gain request_id
slog.SetDefault(slog.New(logf.NewSlogHandler(
    logf.NewContextHandler(sync),
)))

// Step 3: add async I/O — swap stderr for SlabWriter → file
sw := logf.NewSlabWriter(file, 64*1024, 8)
router, close, _ := logf.NewRouter().Route(logf.JSON().Build(), logf.Output(logf.LevelInfo, sw)).Build()
slog.SetDefault(slog.New(logf.NewSlogHandler(logf.NewContextHandler(router))))

// Step 4 (optional): switch hot paths to logf for typed fields
logger := logf.New(logf.NewContextHandler(router))
logger.Info(ctx, "fast path", logf.Int("status", 200))
```

## Router (the traffic cop)

One log entry, multiple destinations, each with its own rules:

```go
fileSlab := logf.NewSlabWriter(file, 64*1024, 8)
jsonEnc := logf.JSON().Build()
textEnc := logf.Text().Build()

router, close, _ := logf.NewRouter().
    Route(jsonEnc,
        logf.OutputCloser(logf.LevelDebug, fileSlab), // everything to file (async)
        logf.Output(logf.LevelError, alertWriter),    // errors to alerting
    ).
    Route(textEnc,
        logf.Output(logf.LevelInfo, os.Stderr),       // colored text to console (sync)
    ).
    Build()
defer close() // flushes and closes fileSlab
```

The Router encodes once per encoder group. Two outputs sharing the same
encoder? One encode call. Stalled network destination? The file output
doesn't care — each writer is independent.

**Mix sync and async** — because console output should be instant but
file writes can be batched:

```go
fileSlab := logf.NewSlabWriter(file, 64*1024, 8,
    logf.WithFlushInterval(100*time.Millisecond),
)

router, close, _ := logf.NewRouter().
    Route(jsonEnc,
        logf.OutputCloser(logf.LevelDebug, fileSlab), // async, Router closes it
        logf.Output(logf.LevelInfo, os.Stderr),       // sync, direct write
    ).
    Build()
defer close() // flushes and closes fileSlab automatically
```

## SlabWriter (the speed demon)

Here's how it works: your goroutine copies log bytes into a pre-allocated
slab buffer under a mutex (~17 ns memcpy). A background goroutine writes
filled slabs to the destination. Your goroutine never touches the disk.
Never blocks on the network. Just copies bytes and moves on.

```go
sw := logf.NewSlabWriter(file, 64*1024, 8,
    logf.WithFlushInterval(100*time.Millisecond),
)
defer sw.Close()
```

When the I/O goroutine can't keep up? The slab pool absorbs the spike.
8 slabs × 64 KB = 512 KB of burst tolerance. At 10K msg/sec with
256-byte messages, that's ~200 ms of I/O stall with zero caller impact.

**Drop mode** — for when losing a log is better than blocking a request:

```go
sw := logf.NewSlabWriter(conn, 64*1024, 8,
    logf.WithDropOnFull(),
    logf.WithFlushInterval(100*time.Millisecond),
    logf.WithErrorWriter(os.Stderr),
)
```

**Keep an eye on it:**

```go
stats := sw.Stats()
// stats.Dropped      — messages lost (dropOnFull mode)
// stats.Written      — messages accepted
// stats.QueuedSlabs  — slabs waiting for I/O
// stats.WriteErrors  — I/O failures
```

See [docs/BUFFERING.md](docs/BUFFERING.md) for capacity planning.

## WriterSlot (the patient one)

Sometimes you need a logger before you know where the logs are going.
Config isn't parsed yet. The database connection isn't up. The cloud
SDK hasn't initialized.

WriterSlot lets you start logging immediately and connect the real
destination later:

```go
slot := logf.NewWriterSlot(logf.WithSlotBuffer(4096))
logger := logf.NewLogger().Output(slot).Build()

logger.Info(ctx, "booting up")       // buffered
logger.Info(ctx, "config loaded")    // buffered

slot.Set(file)                       // buffer flushed, future writes go to file

logger.Info(ctx, "ready to serve")   // written directly
```

## Odds and ends

### Testing

```go
// Silent tests (discard everything):
logger := logf.DisabledLogger()

// Capture logs for assertions:
var buf bytes.Buffer
logger := logf.NewLogger().Output(&buf).Build()
logger.Info(ctx, "hello")
// buf.String() has your JSON

// Logs in test output (visible with -v or on failure):
logger := logf.NewLogger().Output(testWriter{t}).Build()
```

Where `testWriter` adapts `testing.T` to `io.Writer`:

```go
type testWriter struct{ t testing.TB }
func (w testWriter) Write(p []byte) (int, error) {
    w.t.Helper()
    w.t.Log(strings.TrimRight(string(p), "\n"))
    return len(p), nil
}
```

### Log rotation

logf doesn't rotate logs — that's what `lumberjack` and `logrotate` are for:

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

### The fine print

- **One allocation per log call with fields.** That's Go's variadic
  `[]Field` slice. Calls without fields are zero-alloc.
- **Oversized messages allocate (SlabWriter only).** Messages bigger than
  `slabSize` get a dedicated buffer. Normal log entries (100–500 bytes)
  with normal slabs (16–64 KB) never hit this.

## Learn more

- [Architecture & design philosophy](docs/ARCHITECTURE.md)
- [Buffering strategies & capacity planning](docs/BUFFERING.md)
- [Performance experiments](docs/EXPERIMENTS.md)
- [TODO & roadmap](docs/TODO.md)
