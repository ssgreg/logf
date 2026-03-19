# Buffering Strategies

The Router has a single output mode — `Output` — that writes encoded data
directly from the caller goroutine. The caller controls buffering by choosing
which `io.Writer` to pass.

Throughout this section, "I/O spike" refers to a transient latency increase on
the destination — a common event for network targets (TCP retransmit, TLS
renegotiation, load-balancer failover, GC pause on the remote collector). A
single 50 ms spike at 10K msg/sec means 500 messages arrive while I/O is
stalled. The configuration determines whether those messages block callers, get
buffered, or are dropped.

## 1. Direct write — no buffer

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

## 2. SlabWriter — async batched I/O

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

## Capacity planning

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

## Why SlabWriter beats per-message channels

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

**Summary.** The channel added ~200 ns of per-message overhead (alloc + copy +
send) with no benefit over a mutex + memcpy into a slab. SlabWriter delivers
the same I/O isolation with less latency, zero allocations, better spike
tolerance, and simpler code (no consumer goroutine in the Router).
