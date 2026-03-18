# logf

[![GoDoc](https://godoc.org/github.com/ssgreg/logf?status.svg)](https://godoc.org/github.com/ssgreg/logf)
[![Build Status](https://github.com/ssgreg/logf/actions/workflows/go.yml/badge.svg)](https://github.com/ssgreg/logf/actions/workflows/go.yml)
[![Go Report Status](https://goreportcard.com/badge/github.com/ssgreg/logf)](https://goreportcard.com/report/github.com/ssgreg/logf)
[![Coverage Status](https://codecov.io/gh/ssgreg/logf/branch/master/graph/badge.svg)](https://codecov.io/gh/ssgreg/logf)

Structured logging for Go — context-aware, slog-native, fast.

## Features

- **Context-aware fields** — `logf.With(ctx, fields...)` attaches fields to context, automatically included in every log entry
- **Native slog bridge** — `logger.Slog()` returns a `*slog.Logger` sharing the same pipeline, fields, and name. Passes `testing/slogtest`
- **Router** — multi-destination fan-out with per-output level filtering and encoder groups
- **Async buffered I/O** — SlabWriter with pre-allocated slab pool, zero per-message allocations
- **Builder API** — `logf.NewLogger().Level(logf.LevelInfo).Build()` for quick setup
- **Zero-alloc hot path** — 0 allocs/op across all benchmarks

## Quick Start

```go
// Minimal — JSON to stderr:
logger := logf.NewLogger().Build()

// Production — custom encoder, stdout:
logger := logf.NewLogger().
    Level(logf.LevelInfo).
    Output(os.Stdout).
    EncoderFrom(logf.JSON().TimeKey("time")).
    Context().
    Build()

// slog backend:
slogger := logger.Slog()
```

## Why logf

slog from the standard library is sufficient for most applications. logf targets
scenarios where slog falls short:

- **High-throughput logging (>100K logs/sec)** — encoding is parallel across
  goroutines (no global mutex), writes are memcpy into pre-allocated slabs
  (~17 ns), and a background goroutine flushes large blocks. Result: 2× faster
  than slog (285 ns/op vs 518 ns/op) under parallel file I/O.
- **Unstable I/O / spike protection** — under simulated slow disk (2% of writes
  sleep 1ms), slog p99 = 2.5ms while logf p99 = 71µs. slog holds a mutex over
  the entire `Handle()` call (encode + write), causing cascading delays under
  contention. logf's slab pool decouples callers from I/O entirely.
- **Request-scoped fields via context** — `logf.With(ctx, fields...)` attaches
  fields to `context.Context` that are automatically included in every log entry.
  Neither slog nor zap support this natively; they require passing a derived
  logger through the call stack.
- **Decoupled encoding and I/O** — the Router encodes once per encoder group
  and fans out to multiple writers. Each writer can be sync or async (SlabWriter)
  independently. A stalled network destination does not block file writes.

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

logf solves this with **slab-based I/O isolation**. Each log entry is encoded in
the caller's goroutine (parallel across all CPUs) and copied into a pre-allocated
slab via a mutex (~17 ns memcpy). A background I/O goroutine flushes filled slabs
to the destination. The caller never waits for disk or network — it moves on to
serve the next request.

When multiple destinations are involved, each gets its own SlabWriter and I/O
goroutine. A stalled Logstash does not slow down local file writes. Error logs
routed to an alerting pipeline do not compete with high-volume info logs.
Destinations are physically isolated — not just logically separated.

This is not a throughput optimization. Throughput is bounded by the slowest
destination regardless of architecture. This is a **latency isolation**
strategy: the price your application pays per log call stays constant (~100 ns)
no matter what happens downstream.

## Architecture

**Buffered I/O** eliminates the dominant cost in file logging: the `write(fd)`
syscall (~2 µs on SSD). All loggers converge to similar encoding speed
(~200–500 ns) once I/O is buffered. The difference is whether buffering is
built-in or requires manual setup.

**Slab-based async architecture** (logf `SlabWriter`) copies encoded bytes into
pre-allocated slab buffers under a mutex. A background I/O goroutine writes
filled slabs to the destination. The slab pool absorbs I/O spikes without
blocking callers or allocating memory. Zero per-message allocations, 2.7×
faster than channel-based alternatives (107 ns vs 289 ns parallel file I/O).

**Sync buffered architecture** (zap `BufferedWriteSyncer`) buffers writes in
memory but encoding still happens in the caller goroutine under a mutex. When
the buffer is full, the caller that triggers the flush pays for the entire
`write(fd)` while holding the mutex, blocking other goroutines.

**Pre-encoded accumulated fields** (zap) encode `.With()` fields once at creation
time and copy raw bytes on each log call. This is faster than re-encoding or
cache lookup on every call, but less flexible for dynamic field sources.

## Buffering Strategies

The Router has a single output mode — `Output` — that writes encoded data
directly from the caller goroutine. The caller controls buffering by choosing
which `io.Writer` to pass.

Throughout this section, "I/O spike" refers to a transient latency increase on
the destination — a common event for network targets (TCP retransmit, TLS
renegotiation, load-balancer failover, GC pause on the remote collector). A
single 50 ms spike at 10K msg/sec means 500 messages arrive while I/O is
stalled. The configuration determines whether those messages block callers, get
buffered, or are dropped.

### 1. Direct write — no buffer

```go
h, close, _ := logf.NewRouter().
    Route(enc, logf.Output(logf.LevelDebug, file)).
    Build()
```

The simplest path. The caller goroutine encodes the entry and writes it directly
to the destination. No goroutines, zero per-message allocations.

- **Caller latency:** Full I/O cost per message (~1–2 µs page cache).
- **Data safety:** Strongest — data is in the kernel buffer before Handle returns.
- **Batching:** None — each message is a separate `write(fd)` syscall.
- **I/O spikes:** The caller stalls for the entire spike duration.

Best for: local files, debug mode, tests.

### 2. SlabWriter — async batched I/O

```go
sw := logf.NewSlabWriter(conn, 64*1024, 8, logf.WithFlushInterval(100*time.Millisecond))
defer sw.Close()
h, close, _ := logf.NewRouter().
    Route(enc, logf.Output(logf.LevelDebug, sw)).
    Build()
```

The caller copies encoded bytes into a pre-allocated slab (~17 ns memcpy under
mutex). A background I/O goroutine writes filled slabs to the destination.
When no new data arrives for `flushInterval`, the partial slab is flushed
automatically.

- **Caller latency:** Mutex + memcpy (~17 ns). The caller never blocks on I/O.
  When all slabs are in flight, Write blocks until a slab is recycled
  (backpressure).
- **Data safety:** Up to `slabCount × slabSize` of data in flight. Graceful
  shutdown via `Close()` flushes all remaining data.
- **Batching:** Yes — each slab is a single large Write call (e.g., 64 KB =
  ~256 messages at 256 bytes each).
- **I/O spikes:** The I/O goroutine stalls, but callers keep filling free slabs.
  The slab pool acts as a time buffer:
  `burst_tolerance = slabCount × slabSize / (msg_rate × avg_msg_size)`.
  With 8 × 64 KB slabs and 256-byte messages at 10K msg/sec: absorbs
  **~200 ms** of spike without blocking any caller. Only when all slabs are in
  flight does backpressure reach callers.
- **Idle flush:** When the message flow stops, a timer fires after
  `flushInterval` and writes the partial slab. The timer never fires during
  active flow (slabs fill before it expires), so batching is not degraded.

Best for: network destinations (Kibana, Loki, remote syslog), high-throughput
file logging, any scenario where I/O isolation matters.

### Capacity planning

SlabWriter has two parameters — `slabSize` and `slabCount` — that control
throughput and spike tolerance independently. Both can be derived from the
workload.

**Inputs:**

- `R` — message rate (msgs/sec)
- `M` — average encoded message size (bytes)
- `L` — destination write latency (seconds per call)
- `S` — maximum I/O spike duration to absorb without backpressure (seconds)

**Step 1: slabSize — throughput.**

Each slab becomes one `Write(slab)` call. To sustain rate `R`, the slab must
hold enough messages to cover the write latency:

```
slabSize ≥ R × M × L
```

| Rate | Msg size | Write latency | Min slabSize |
|---|---|---|---|
| 10K msg/s | 256 B | 1 ms (file) | 2.5 KB |
| 10K msg/s | 256 B | 10 ms (network) | 25 KB |
| 100K msg/s | 200 B | 10 ms (network) | 200 KB |

Round up to a power of two for alignment. For most workloads, 16–64 KB is a
good default.

**Step 2: slabCount — spike tolerance.**

During an I/O spike, the I/O goroutine is blocked. Callers keep filling free
slabs. The pool must hold enough slabs to absorb the spike:

```
slabCount ≥ S × R × M / slabSize
```

| Rate | Msg size | slabSize | Spike target | Min slabCount |
|---|---|---|---|---|
| 10K msg/s | 256 B | 64 KB | 50 ms | 2 |
| 10K msg/s | 256 B | 64 KB | 200 ms | 1 + 8 = 8 |
| 50K msg/s | 256 B | 64 KB | 100 ms | 20 |
| 100K msg/s | 200 B | 256 KB | 100 ms | 8 |

Add 1 for the slab currently being filled by producers.

**Step 3: memory budget.**

Total memory = `slabCount × slabSize`. If the budget is tight, reduce
`slabSize` (trades throughput for spike tolerance at the same memory cost):

| Config | Memory | Spike at 10K×256 B | Spike at 50K×256 B |
|---|---|---|---|
| 8 × 64 KB | 512 KB | 200 ms | 40 ms |
| 16 × 16 KB | 256 KB | 100 ms | 20 ms |
| 16 × 64 KB | 1 MB | 400 ms | 80 ms |
| 4 × 16 KB | 64 KB | 25 ms | 5 ms |

**Quick reference — common scenarios:**

```go
// Local file, moderate rate — 512 KB, absorbs 200 ms at 10K msg/s
sw := logf.NewSlabWriter(file, 64*1024, 8)

// Network (Kibana), high rate — 1 MB, absorbs 80 ms at 50K msg/s
sw := logf.NewSlabWriter(conn, 64*1024, 16, logf.WithFlushInterval(100*time.Millisecond))

// Low-memory sidecar — 64 KB total, absorbs 25 ms at 10K msg/s
sw := logf.NewSlabWriter(conn, 16*1024, 4, logf.WithFlushInterval(50*time.Millisecond))

// High-throughput pipeline — 2 MB, absorbs 100 ms at 100K msg/s of 200 B msgs
sw := logf.NewSlabWriter(conn, 256*1024, 8, logf.WithFlushInterval(100*time.Millisecond))
```

### Why SlabWriter beats per-message channels

Earlier versions of logf used a Go channel to pass each encoded message from
the caller to a consumer goroutine. SlabWriter replaces that design entirely.
Here is why.

**Per-message cost.** A channel requires `make([]byte, N)` + `copy` + `chan send`
for every message — about 250 ns and 1 allocation (208 bytes). SlabWriter does
`mutex.Lock` + `memcpy into slab` + `mutex.Unlock` — about 17 ns and 0
allocations. The channel only fires once per full slab (~328 messages), not once
per message.

**Benchmark results** (parallel file I/O, 10 goroutines, M1 Pro):

| Strategy | ns/op | allocs/op | MB/s |
|---|---|---|---|
| SlabWriter 8×64 KB | 107 | 0 | 1,860 |
| SlabWriter 2×32 KB (64 KB total) | 132 | 0 | 1,470 |
| Channel(20) + BufferedWriter 64 KB | 289 | 1 (208 B) | 692 |
| Channel(1000) + BufferedWriter 64 KB | 265 | 1 (208 B) | 754 |
| Channel(20) + SlabWriter 8×64 KB | 300 | 1 (208 B) | 669 |

SlabWriter is **2.7× faster** than any channel variant at the same memory
budget. Increasing the channel buffer from 20 to 1000 barely helps (289 → 265
ns) because the bottleneck is per-message overhead, not channel contention.

**Why the channel is the bottleneck, not the buffer behind it.** Both
`Channel + BufferedWriter` and `Channel + SlabWriter` show nearly identical
numbers (~289 vs ~300 ns). The buffer type behind the channel does not matter —
the channel dominates. Removing the channel removes the bottleneck.

**Spike tolerance scales with slab count, not channel size.** A channel of 20
entries × 256 bytes = 5 KB of spike buffer. Increasing to 1000 entries = 250 KB.
A SlabWriter with 8 × 64 KB = 512 KB of spike buffer, but each slab holds ~256
messages, so the effective burst capacity is ~2,000 messages vs ~1,000 for the
large channel — at half the per-message cost.

**Idle flush works without a channel.** The old channel consumer detected idle
by observing an empty channel (`select` with `default` branch → Flush). Without
a channel, SlabWriter uses a timer in its I/O goroutine: when the `full` slab
channel is empty, arm a timer; if it fires before a new slab arrives, flush the
partial slab. During active flow, slabs arrive before the timer expires, so
batching is never degraded.

**Summary.** The channel added ~200 ns of per-message overhead (alloc + copy +
send) with no benefit over a mutex + memcpy into a slab. SlabWriter delivers
the same I/O isolation with less latency, zero allocations, better spike
tolerance, and simpler code (no consumer goroutine in the Router).

## slog Compatibility

logf provides `logf.NewSlogHandler` — a drop-in `slog.Handler` implementation
backed by logf's Router and encoder. This gives slog users the benefits
of async logging, buffered I/O, and context-scoped fields without changing
their application code.

```go
// Standard slog — sync, unbuffered
logger := slog.New(slog.NewJSONHandler(file, nil))

// slog + logf backend — async, buffered, same API
enc := logf.JSON().Build()
sw := logf.NewSlabWriter(file, 64*1024, 8)
defer sw.Close()
h, close, _ := logf.NewRouter().
    Route(enc, logf.Output(logf.LevelDebug, sw)).
    Build()
defer close()
logger := slog.New(logf.NewSlogHandler(h))
```

The rest of the application code stays the same. logf is not another logging API —
it is an infrastructure layer behind the standard one. Think of it as choosing
nginx vs another HTTP server: the HTTP interface is the same, but the backend
determines throughput, latency, and resilience.

**`Logger.Slog()` produces a fully integrated `*slog.Logger`** — not a separate
logger, but a view of the same pipeline. It inherits accumulated `.With()` fields,
`.WithName()` identity, and the full Handler chain (Router, ContextHandler, etc.).
Logging via slog and logf in the same request produces identical context fields,
the same encoder, the same destination. There is no state divergence.

**Context-scoped fields** work through the slog handler as well.
Middleware can attach fields to `context.Context` via `logf.With(ctx, fields...)`,
and the handler automatically includes them in every log entry — without passing
a derived logger through the call stack. Neither zap nor standard slog support this.

## TODO

- `Stack(key)` / `StackSkip(key, skip)` field constructor — capture current goroutine stack trace
- Sync path error reporting — Logger currently discards `Handle()` errors (`_ =`), same as slog/zerolog/logrus. Consider a Handler decorator or Logger option to redirect write errors to a separate `io.Writer` (like zap's `ErrorOutput`). SlabWriter already has `WithErrorWriter` for async errors.
