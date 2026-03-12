# logf

[![GoDoc](https://godoc.org/github.com/ssgreg/logf?status.svg)](https://godoc.org/github.com/ssgreg/logf)
[![Build Status](https://github.com/ssgreg/logf/actions/workflows/go.yml/badge.svg)](https://github.com/ssgreg/logf/actions/workflows/go.yml)
[![Go Report Status](https://goreportcard.com/badge/github.com/ssgreg/logf)](https://goreportcard.com/report/github.com/ssgreg/logf)
[![Coverage Status](https://codecov.io/gh/ssgreg/logf/branch/master/graph/badge.svg)](https://codecov.io/gh/ssgreg/logf)

Faster-than-light, asynchronous, structured logger in Go with zero allocation count.

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
- **Async logging out of the box** — logf's channel writer is the only Go logger
  with built-in asynchronous logging. The caller goroutine never blocks on
  `write(fd)`, paying only ~125ns per log call regardless of I/O conditions.

## Architecture

**Buffered I/O** eliminates the dominant cost in file logging: the `write(fd)` syscall
(~2us on SSD). All loggers converge to similar encoding speed (~200-500ns) once I/O
is buffered. The difference is whether buffering is built-in or requires manual setup.

**Async channel architecture** (logf) decouples the caller goroutine from both
encoding and I/O. The caller only pays for creating an Entry and sending it to
a channel (~125ns). Encoding and writing happen in a dedicated worker goroutine.
This provides consistent low latency regardless of I/O conditions.

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
