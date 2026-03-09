# Benchmarks

Comparative benchmarks for Go logging libraries: **logf**, **zap**, **zerolog**, **slog**, **logrus**.

## Test Environment

- Apple M1 Pro (10 cores)
- macOS, Go 1.21+
- Output format: JSON
- I/O targets: `io.Discard` and real temporary files

## Benchmark Categories

### 1. Micro-benchmarks (`zapbench_test.go`)

Standard `testing.B` benchmarks measuring `ns/op`, `B/op`, `allocs/op`.

**Scenarios:**

- **Disabled path** — level check overhead when logging is off
- **Plain text** — minimal log entry (level + timestamp + message)
- **With fields** — 5 fields passed per call (`int`, `string`, `string`, `int64`, `bool`)
- **Accumulated fields** — 5 fields via `.With()`, then plain log call
- **Parallel** — `b.RunParallel` with `GOMAXPROCS` goroutines

**I/O variants:**

- `io.Discard` — measures pure encoding cost
- File I/O (unbuffered) — each logger writes to a temp file, one `write(fd)` per entry
- File I/O (buffered) — with write buffering enabled where available

**Buffering configurations:**

- logf: built-in `writeAppender` buffer (4KB, flushes when full)
- zap: `zapcore.BufferedWriteSyncer` (256KB buffer, 30s flush timer)
- zerolog/slog/logrus: `bufio.Writer` (4KB) wrapper

### 2. Latency Distribution (`latency_test.go` — `TestLatencyDistribution`)

Measures per-call latency of each logger writing to a real file.
Reports percentiles: **p50, p90, p99, p999, max**.

50,000 entries per logger, single goroutine. Each call is individually timed
via `time.Now()` / `time.Since()`. Results are sorted and percentiles computed.

This test reveals tail latency differences between async and sync logging
architectures. Async loggers decouple the caller from I/O, so their p50
stays low regardless of disk speed, while sync loggers show spikes when
`write(fd)` is slow.

### 3. Slow I/O Simulation (`latency_test.go` — `TestSlowIOLatency`)

A `slowWriter` wrapper randomly sleeps on 2% of `Write()` calls (1ms delay),
simulating real-world disk latency spikes (NFS, container throttling, SSD GC).

Run in parallel (`GOMAXPROCS` goroutines), 50,000 total samples.
Reports the same percentiles as Test 2.

This models production conditions where disk I/O is not consistently fast.
Loggers with async or buffered writes absorb spikes without blocking callers.

### 4. Goroutine Scalability (`latency_test.go` — `TestGoroutineScalability`)

Measures throughput (logs/sec) as goroutine count increases: 1, 2, 4, 8, 16, 32, 64.
Each configuration writes 200,000 entries to a real temp file.

Shows how different synchronization strategies (channel, mutex, lock-free)
affect throughput under contention. Stable degradation is preferred over
high peak throughput with unpredictable drops.

## Running

```bash
# Micro-benchmarks (disable bench_test.go if it doesn't compile with v2 API)
go test -run='^$' -bench='BenchmarkZap|BenchmarkSlog|BenchmarkZerolog|BenchmarkLogrus' \
  -benchmem -count=1 -benchtime=3s ./benchmarks/

# logf-specific micro-benchmarks
go test -run='^$' -bench='Benchmark' -benchmem -count=1 -benchtime=3s .

# Latency distribution
go test -run='TestLatencyDistribution' -v -count=1 ./benchmarks/

# Slow I/O simulation
go test -run='TestSlowIOLatency' -v -count=1 ./benchmarks/

# Goroutine scalability
go test -run='TestGoroutineScalability' -v -count=1 ./benchmarks/
```

## Key Metrics

| Metric | What it shows |
| --- | --- |
| **ns/op** | Average latency per log call |
| **B/op** | Heap bytes allocated per call |
| **allocs/op** | Number of heap allocations per call |
| **p50/p99/p999** | Latency percentiles (tail latency) |
| **logs/sec** | Throughput under concurrent load |

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

## Architecture Notes

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
    Appender: logf.NewWriteAppender(file, logf.NewJSONEncoder.Default()),
})
defer close()
logger := slog.New(logf.NewSlogHandler(w, nil))
```

The rest of the application code stays the same. logf is not another logging API —
it is an infrastructure layer behind the standard one. Think of it as choosing
nginx vs another HTTP server: the HTTP interface is the same, but the backend
determines throughput, latency, and resilience.

Benchmarks include `slog + logf handler` vs `slog + standard handler` comparisons
to show the performance difference of the backend while keeping the same API.

**Context-scoped fields** are a unique logf feature available through the slog handler.
Middleware can attach fields to `context.Context` via `logf.With(ctx, fields...)`,
and the handler automatically includes them in every log entry — without passing
a derived logger through the call stack. Neither zap nor standard slog support this.
