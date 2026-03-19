# Architecture

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

## Pipeline

```text
Logger ──→ ContextHandler ──→ Router ──→ Encoder₁ ──→ SlabWriter₁ (file)
  │                              └──→ Encoder₂ ──→ SlabWriter₂ (HTTP)
  └─ .Slog() → slogHandler ─┘
```

- **Handler** = `Enabled(ctx, Level) + Handle(ctx, Entry)`. Thin, no Flush/Sync.
- **Writer** = `io.Writer + Flush + Sync`. Owns I/O lifecycle.
- **Router** = synchronous fan-out Handler. One Encode per encoder group,
  direct Write to each output.
- **SlabWriter** = pre-allocated slab pool + background I/O goroutine.
  `WithDropOnFull()` for non-blocking mode.

## Bag — cached field chain

`Bag` is an immutable linked list of fields. Each `With()` creates a new
node pointing to the parent — O(1), no copies. Bags are safe to share
across goroutines.

```text
Bag{fields:[status=200]} → Bag{group:"http"} → Bag{fields:[env=prod]} → nil
```

Each Bag node has a per-encoder slot cache (`AllocEncoderSlot`). The
encoder writes the fields once and caches the raw bytes. Subsequent
entries with the same Bag version skip encoding entirely. This makes
accumulated `.With()` fields essentially free on the hot path.

Groups (`WithGroup`) are stored as separate Bag nodes with no fields.
The encoder opens a JSON object for each group node and closes them
at the end. Empty trailing groups are suppressed automatically.

## ContextHandler

`ContextHandler` is a Handler middleware that extracts the Bag from
`context.Context` and attaches it to the Entry before passing it
downstream. Optional `FieldSource` functions extract additional fields
from context (e.g., OTel trace IDs, request metadata).

```text
Logger.Info(ctx, "msg")
  → ContextHandler.Handle(ctx, entry)
    → entry.Bag = BagFromContext(ctx)
    → entry.Fields = append(sourceFields(ctx), entry.Fields...)
    → next.Handle(ctx, entry)
```

Cost when unused (no Bag in context, no sources): one `ctx.Value()`
returning nil + one `len == 0` check. ~2-3 ns.

## slog integration

`Logger.Slog()` returns a `*slog.Logger` that shares the same Handler,
Bag, and name. It is a view of the same pipeline, not a separate logger.
Logging via slog and logf in the same request produces identical context
fields, the same encoder, the same destination.

The slog handler (`NewSlogHandler`) passes `testing/slogtest` — full
contract compliance including empty groups, inline groups, zero time,
and LogValuer resolution.

## I/O strategy

**Buffered I/O** eliminates the dominant cost in file logging: the `write(fd)`
syscall (~2 µs on SSD). All loggers converge to similar encoding speed
(~200–500 ns) once I/O is buffered. The difference is whether buffering is
built-in or requires manual setup.

logf's `SlabWriter` copies encoded bytes into pre-allocated slab buffers
under a mutex. A background I/O goroutine writes filled slabs to the
destination. The slab pool absorbs I/O spikes without blocking callers
or allocating memory. Zero per-message allocations.

See [BUFFERING.md](BUFFERING.md) for capacity planning and benchmarks.
