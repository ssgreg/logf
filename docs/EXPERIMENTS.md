# SlabWriter Performance Experiments

Platform: M1 Pro, 10 goroutines, parallel file I/O, 200B JSON log entries.
Baseline: original SlabWriter with `atomic.Int64` counters = **425 ns/op**.

## What shipped

| Change | ns/op | Delta | Mechanism |
|--------|-------|-------|-----------|
| Baseline (atomic counters) | 425 | — | |
| Remove `atomic.Int64.Add` for `written` inside mutex | 348 | **-18%** | Atomic inside mutex causes cache-line bouncing across cores; plain `int++` under mutex has zero atomic overhead |
| + Remove write loop, direct `copy` | 348 | **-2%** | After early swap, message guaranteed to fit — loop with bounds check replaced by single `copy` |
| + Conditional `msgCount++` (only in dropOnFull) | ~347 | ~1 ns | Avoid unnecessary increment on hot path |
| + Early swap (doesn't fit in remainder) | no regression | | Prevents torn writes; one extra comparison on hot path |
| + Oversized message support (> slabSize) | no regression | | Dedicated `make+copy` only for rare large messages |
| + `safeReport` with `recover()` | no impact | | Prevents errW panic from crashing ioLoop |

**Final: 348 ns/op (was 425). -18% from baseline.**

## What didn't ship

### Lock-free CAS (atomic reserve + writers counter)

**Idea:** Replace mutex with CAS on packed state `[writers:16|pos:48]`.
Each Write does `CompareAndSwap` to reserve space, copies data, then
`atomic.Add(-writersInc)` to signal completion. Mutex only for swap.

**Result: 385 ns (+11% vs shipped 348 ns).**

Why: CAS + atomic decrement = 2 atomic ops per Write. Under 10-goroutine
contention, these atomic ops cause the same cache-line bouncing that
killed the `written` counter. The mutex version does 1 Lock + 1 Unlock
(which are also atomic ops, but the runtime's adaptive spinning is more
efficient than our CAS retry loop for ~25 ns critical sections).

**Lesson:** Lock-free is not always faster. For very short critical
sections (<50 ns), Go's mutex with adaptive spinning beats hand-rolled
CAS under moderate contention.

### Copy outside mutex (reserve under lock, copy outside)

**Idea:** Reserve space and capture slab pointer under mutex, copy data
outside mutex. Use `atomic.Int32` writers counter so swap knows when
all copies are done.

**Result: 428 ns (+23% vs shipped 348 ns).**

Why: Two `writers.Add` calls (increment under lock + decrement outside)
are two atomic ops — same problem as lock-free CAS. The 15 ns saved by
moving copy outside is overwhelmed by 2×5 ns atomic overhead amplified
by cache bouncing.

**Lesson:** Moving work outside a mutex only helps if the
synchronization cost of the "completion signal" is cheaper than the
work itself. For 200-byte memcpy (~15 ns), it's not.

### Spinlock (CAS-based, no goroutine parking)

**Idea:** Replace `sync.Mutex` with a custom spinlock that spins
(with `runtime.procyield` PAUSE hint) instead of parking goroutines.
Avoids the ~1µs park/wake overhead visible in CPU profiles.

**Result: 340 ns (-2.3% vs shipped 348 ns).**

Why it wasn't shipped: Requires `//go:linkname runtime.procyield` —
an unstable internal API that can break on Go upgrades. The 8 ns gain
doesn't justify the maintenance risk.

**Lesson:** Spinlock does help for <50 ns critical sections, but Go's
`//go:linkname` is too fragile for production code.

### Sharded SlabWriter (per-P or per-goroutine shards)

**Idea:** N independent slab shards (one per GOMAXPROCS), each with
its own mutex. Goroutines pick a shard via stack-address hash.
Eliminates cross-goroutine mutex contention almost entirely.

**Result: 175 ns (-50% vs shipped 348 ns).**

Why it wasn't shipped: Breaks strict message ordering between
goroutines (different shards → different slabs → reordered at
destination). Also: +50% memory, stack-address hash is imperfect
(occasional collisions cause contention spikes), more complex
Flush/Close (must iterate all shards).

Sharding results by shard count:

| Shards | ns/op | vs baseline |
| ------ | ----- | ----------- |
| 1 (shipped) | 348 | — |
| 2 | 293 | -16% |
| 4 | 194 | -44% |
| auto (GOMAXPROCS=10) | 175 | -50% |

**Lesson:** Sharding is the only technique that fundamentally reduces
contention. All other approaches (lock-free, spinlock, copy-outside)
are limited by the single-mutex bottleneck. The trade-off is ordering.

## Key insight: atomics inside mutex

The single most impactful discovery: `atomic.Int64.Add` inside a mutex
causes a +22% performance regression (425→348 ns) on ARM64. The atomic
instruction's memory barrier forces the CPU to drain its store buffer,
invalidating cache lines on all cores. The next `mutex.Unlock` (which
is also an atomic) then suffers a cache miss.

**Rule: never use atomic operations on data protected by a mutex.**
Use plain `int++` instead. Read the value under mutex in `Stats()`.

This effect is invisible in single-goroutine benchmarks (~5 ns overhead)
and only manifests under parallel contention with cache pressure from
other work (encoder pool operations, buffer allocation). The
`BenchmarkParallelConcurrentSlabCachePressure` benchmark was created
specifically to catch this class of regression.

## CPU profile breakdown (fileparallel, 10 goroutines)

| Component | % CPU |
|-----------|-------|
| Mutex contention (usleep + pthread_cond) | 65% |
| Mutex lock/unlock | 20% |
| Encoder (JSON) | 5% |
| File I/O (ioLoop) | 3% |
| Everything else | 7% |

The mutex is 85% of CPU. The encoder is already fast (5%). Further
optimization requires reducing contention (sharding) or reducing
critical section time (already minimized to ~25 ns memcpy).
