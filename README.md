# › logf

[![Go Reference](https://pkg.go.dev/badge/github.com/ssgreg/logf/v2.svg)](https://pkg.go.dev/github.com/ssgreg/logf/v2)
[![Build Status](https://github.com/ssgreg/logf/actions/workflows/go.yml/badge.svg)](https://github.com/ssgreg/logf/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/ssgreg/logf/v2)](https://goreportcard.com/report/github.com/ssgreg/logf/v2)
[![codecov](https://codecov.io/gh/ssgreg/logf/branch/master/graph/badge.svg)](https://codecov.io/gh/ssgreg/logf)

Structured logging for Go — context-aware, slog-native, fast.

## So you want to log things

You already have `slog`. It works. It's in the standard library. Why would
you need anything else?

Well, most of the time you don't. But then one day your service starts
handling 50K requests per second and you notice something funny: your
p99 latency spikes every time the log collector hiccups. Or you realize
that passing a logger through seventeen function arguments just to get
a `request_id` in your database layer is... not great.

That's where logf comes in. Think of it as slog's cool older sibling who
went to systems programming school and came back with opinions about
memory allocation.

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
// → {"level":"info","ts":"2026-03-19T14:04:02Z","msg":"hello, world","caller":"main.go:10","from":"logf"}
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

## Logging (the fun part)

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

## Context-aware fields

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
// Their logs go through YOUR pipeline. One config to rule them all.
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
sw := logf.NewSlabWriter(file).SlabSize(64*1024).SlabCount(8).Build()
router, close, _ := logf.NewRouter().Route(logf.JSON().Build(), logf.Output(logf.LevelInfo, sw)).Build()
slog.SetDefault(slog.New(logf.NewSlogHandler(logf.NewContextHandler(router))))

// Step 4 (optional): switch hot paths to logf for typed fields
logger := logf.New(logf.NewContextHandler(router))
logger.Info(ctx, "fast path", logf.Int("status", 200))
```

## Router (the traffic cop)

One log entry, multiple destinations, each with its own rules:

```go
fileSlab := logf.NewSlabWriter(file).SlabSize(64*1024).SlabCount(8).Build()
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
fileSlab := logf.NewSlabWriter(file).
    SlabSize(64*1024).
    SlabCount(8).
    FlushInterval(100*time.Millisecond).
    Build()

router, close, _ := logf.NewRouter().
    Route(enc,
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
sw := logf.NewSlabWriter(file).
    SlabSize(64*1024).
    SlabCount(8).
    FlushInterval(100*time.Millisecond).
    Build()
defer sw.Close()
```

When the I/O goroutine can't keep up? The slab pool absorbs the spike.
8 slabs × 64 KB = 512 KB of burst tolerance. At 10K msg/sec with
256-byte messages, that's ~200 ms of I/O stall with zero caller impact.

**Drop mode** — for when losing a log is better than blocking a request:

```go
sw := logf.NewSlabWriter(conn).
    SlabSize(64*1024).
    SlabCount(8).
    DropOnFull().
    FlushInterval(100*time.Millisecond).
    ErrorWriter(os.Stderr).
    Build()
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

## Why not just use slog?

Honestly? For most apps, slog is fine. logf is for when:

- You're logging **a lot** (>100K entries/sec) and encoding is parallel
  across goroutines with pre-allocated slabs (~17 ns per write)
- Your I/O is **unreliable** (slab pool gives you p99 = 71µs vs
  slog's p99 = 2.5ms under simulated slow disk)
- You want **context fields** without the ceremony (slog passes context
  through but never reads it)
- You need **fan-out** to multiple destinations with independent
  encoding and I/O strategies

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for the gory details.

## Who uses logf

- [Acronis](https://www.acronis.com) — global cybersecurity and data protection platform

## Testing

```go
// Silent tests (discard everything):
logger := logf.DisabledLogger()

// Capture logs for assertions:
var buf bytes.Buffer
logger := logf.NewLogger().Output(&buf).Build()
logger.Info(ctx, "hello")
// buf.String() has your JSON

// Logs in test output (visible with -v or on failure):
type testWriter struct{ t testing.TB }
func (w testWriter) Write(p []byte) (int, error) {
    w.t.Helper()
    w.t.Log(strings.TrimRight(string(p), "\n"))
    return len(p), nil
}
logger := logf.NewLogger().Output(testWriter{t}).Build()
```

## Log rotation

logf doesn't rotate logs — that's what `lumberjack` and `logrotate` are for:

```go
import "gopkg.in/natefinch/lumberjack.v2"

rotator := &lumberjack.Logger{
    Filename:   "/var/log/myapp.log",
    MaxSize:    100, // MB
    MaxBackups: 3,
    MaxAge:     28,
}
sw := logf.NewSlabWriter(rotator).SlabSize(64*1024).SlabCount(8).Build()
```

## Performance

Parallel benchmarks on Apple M1 Pro, Go 1.24, `count=5`.
Full results and methodology in [benchmarks/](benchmarks/).

### Latency (ns/op, lower is better)

| Scenario | logf | slog | slog+logf | zap | zerolog | logrus |
|---|---:|---:|---:|---:|---:|---:|
| No fields | 43 | 221 | 53 | 50 | 26 | 500 |
| 2 scalars | 94 | 237 | 133 | 126 | 32 | 820 |
| 6 fields (bytes, time, object…) | 257 | 722 | 836 | 611 | 147 | 1937 |
| With() per call | 200 | 363 | 254 | 579 | 196 | 812 |
| Caller + 2 scalars | 232 | 471 | 246 | 339 | 232 | — |
| With() (no log call) | 60 | 343 | 227 | 456 | 68 | 226 |
| WithGroup() (no log call) | 21 | 98 | 67 | 430 | — | — |

### Allocations (B/op / allocs)

| Scenario | logf | slog | slog+logf | zap | zerolog | logrus |
|---|---:|---:|---:|---:|---:|---:|
| No fields | 0 | 0 | 0 | 0 | 0 | 836 / 16 |
| 2 scalars | 112 / 1 | 0 | 112 / 1 | 128 / 1 | 0 | 1413 / 23 |
| 6 fields | 355 / 1 | 1046 / 13 | 710 / 5 | 1188 / 7 | 0 | 3220 / 46 |
| With() per call | 177 / 2 | 352 / 8 | 371 / 6 | 1427 / 6 | 512 / 1 | 1413 / 23 |
| With() | 176 / 2 | 352 / 8 | 368 / 6 | 1425 / 6 | 0 | 416 / 3 |
| WithGroup() | 64 / 1 | 184 / 4 | 128 / 3 | 1361 / 6 | — | — |

**The highlights:**

- **Faster than zap** on most scenarios. `With()` is 7.6× faster
  (60 ns vs 456 ns), `WithGroup()` is 20× faster (21 ns vs 430 ns).
  These are the "new logger per request" operations — they happen a lot.
- **2–5× faster than slog** across the board. slog shows 0 allocs on
  small field counts (inline buffer), but pays for it in latency.
- **6 fields: 257 ns** — zap needs 611 ns, slog needs 722 ns.
  logf keeps 1 alloc where slog does 13.
- **slog+logf** — keep the standard `slog.Logger` API, get 2–4× faster
  than stock slog. Caller lookup is nearly free: 246 ns vs slog's 471 ns.
  The one weak spot is 6 fields (836 ns) — slog's `any`-based
  attrs force reflection that logf's typed fields avoid.
- **zerolog is faster** (zero-alloc by design), but you pay for it:
  no multi-destination routing, no context-aware fields, no slog
  compatibility, no async I/O, and a fluent API where a forgotten
  `.Msg()` silently drops the entry.
- **Router: 36 ns** for a log call routed to io.Discard — encoding,
  level check, and dispatch included.
- **SlabWriter: 0 allocs, async I/O.** Your goroutine does a memcpy and
  moves on. Background I/O handles the rest.

### Real file I/O (parallel, 6 fields)

The benchmarks above use `io.Discard`. Here's what happens with a real
file and a realistic payload (bytes, time, []int, []string, duration,
object) — where SlabWriter's async architecture actually matters:

| Config | ns/op | B/op | allocs |
|---|---:|---:|---:|
| logf + SlabWriter | 744 | 353 | 1 |
| zerolog + bufio 256KB | 1098 | 0 | 0 |
| zap + BufferedWriteSyncer | 1097 | 1187 | 7 |
| slog + bufio 256KB | 1986 | 1076 | 17 |

All loggers use 256KB of buffering. With real I/O and realistic fields,
logf is **32% faster than zerolog and zap** — the typed encoder advantage
grows as field count and complexity increase.

**Under I/O pressure** (5% of writes stall for 5ms — think slow network,
overloaded disk):

| Logger | p50 | p99 | p999 |
|---|---:|---:|---:|
| logf (SlabWriter) | 833 ns | 43 µs | 163 µs |
| zap (buffered) | 917 ns | 56 µs | 180 µs |
| zerolog (unbuffered) | 7.5 µs | 5.9 ms | 10.2 ms |
| slog (unbuffered) | 14 µs | 17.9 ms | 24 ms |

Sync loggers block on every stalled write. logf copies bytes into a slab
and moves on — the background goroutine deals with the slow destination.
Your HTTP handler never notices.

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
