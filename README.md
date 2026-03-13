# logf

[![GoDoc](https://godoc.org/github.com/ssgreg/logf?status.svg)](https://godoc.org/github.com/ssgreg/logf)
[![Build Status](https://github.com/ssgreg/logf/actions/workflows/go.yml/badge.svg)](https://github.com/ssgreg/logf/actions/workflows/go.yml)
[![Go Report Status](https://goreportcard.com/badge/github.com/ssgreg/logf)](https://goreportcard.com/report/github.com/ssgreg/logf)
[![Coverage Status](https://codecov.io/gh/ssgreg/logf/branch/master/graph/badge.svg)](https://codecov.io/gh/ssgreg/logf)

Async structured logging for Go. Fast, resilient, slog-compatible.

## Why logf

slog from the standard library is sufficient for most applications. logf targets
scenarios where slog falls short:

- **High-throughput logging (>100K logs/sec)** — logf's buffered file I/O is 2×
  faster than slog (285 ns/op vs 518 ns/op) due to built-in write batching.
- **Unstable I/O / spike protection** — under simulated slow disk (2% of writes
  sleep 1ms), slog p99 = 2.5ms while logf p99 = 71µs. slog holds a mutex over
  the entire `Handle()` call (encode + write), causing cascading delays under
  contention. logf's async channel decouples callers from I/O entirely.
- **Request-scoped fields via context** — `logf.With(ctx, fields...)` attaches
  fields to `context.Context` that are automatically included in every log entry.
  Neither slog nor zap support this natively; they require passing a derived
  logger through the call stack.
- **Decoupled encoding and I/O** — logf is the only Go logger where encoding
  and writing are fully separated. This enables async writes, multiple encoders
  per destination, and per-destination isolation in a single pipeline. The caller
  goroutine encodes in parallel and never blocks on `write(fd)`.

## Design Philosophy

**The logger must never be the reason your service slows down.**

Writing to a local file is fast. If the disk cannot keep up, you have far bigger
problems than logging. But production systems rarely log to a single file — they
fan out to remote destinations: Logstash, Loki, Elasticsearch, S3 archivers.
Any of these can lag, stall, or go down entirely.

In a traditional synchronous logger, a slow remote destination blocks the
`write()` call, which blocks the goroutine, which blocks your HTTP handler.
One sluggish log collector degrades every request in the service. The logger
becomes the bottleneck.

logf solves this with **channel-based isolation**. Each log entry is encoded in
the caller's goroutine (parallel across all CPUs) and sent to a channel
(~20 ns). A dedicated consumer goroutine handles the actual I/O. The caller
never waits for disk or network — it moves on to serve the next request.

When multiple destinations are involved, each gets its own channel and consumer
goroutine. A stalled Logstash does not slow down local file writes. Error logs
routed to an alerting pipeline do not compete with high-volume info logs.
Destinations are physically isolated — not just logically separated.

This is not a throughput optimization. Throughput is bounded by the slowest
destination regardless of architecture. This is a **latency isolation**
strategy: the price your application pays per log call stays constant (~100 ns)
no matter what happens downstream.

## Architecture

**Buffered I/O** eliminates the dominant cost in file logging: the `write(fd)` syscall
(~2us on SSD). All loggers converge to similar encoding speed (~200-500ns) once I/O
is buffered. The difference is whether buffering is built-in or requires manual setup.

**Async channel architecture** (logf) encodes each entry in the caller goroutine
(parallel across all CPUs) and sends the resulting bytes to a channel (~20 ns).
A dedicated consumer goroutine handles only the `write(fd)` call. This separates
CPU-bound work (encoding) from I/O-bound work (writing), keeping caller latency
consistent regardless of I/O conditions.

**Sync buffered architecture** (zap `BufferedWriteSyncer`) buffers writes in memory
but encoding still happens in the caller goroutine under a mutex. When the buffer
is full, the caller that triggers the flush pays for the entire `write(fd)` while
holding the mutex, blocking other goroutines.

**Pre-encoded accumulated fields** (zap) encode `.With()` fields once at creation
time and copy raw bytes on each log call. This is faster than re-encoding or
cache lookup on every call, but less flexible for dynamic field sources.

## slog Compatibility

logf provides `logf.NewSlogHandler` — a drop-in `slog.Handler` implementation
backed by logf's channel writer and encoder. This gives slog users the benefits
of async logging, buffered I/O, and context-scoped fields without changing
their application code.

```go
// Standard slog — sync, unbuffered
logger := slog.New(slog.NewJSONHandler(file, nil))

// slog + logf backend — async, buffered, same API
w, close := logf.NewChannelWriter(logf.ChannelWriterConfig{
    Appender: logf.NewWriteAppender(file, logf.NewJSONEncoder(logf.JSONEncoderConfig{})),
})
defer close()
logger := slog.New(logf.NewSlogHandler(w, nil))
```

The rest of the application code stays the same. logf is not another logging API —
it is an infrastructure layer behind the standard one. Think of it as choosing
nginx vs another HTTP server: the HTTP interface is the same, but the backend
determines throughput, latency, and resilience.

**Context-scoped fields** are a unique logf feature available through the slog handler.
Middleware can attach fields to `context.Context` via `logf.With(ctx, fields...)`,
and the handler automatically includes them in every log entry — without passing
a derived logger through the call stack. Neither zap nor standard slog support this.

## TODO

- `Stack(key)` / `StackSkip(key, skip)` field constructor — capture current goroutine stack trace
