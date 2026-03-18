# slog Integration — Decisions & Status

## Architecture (current)

```text
Logger ──→ ContextHandler ──→ Router ──→ Encoder₁ ──→ SlabWriter₁ (file)
  │                              └──→ Encoder₂ ──→ SlabWriter₂ (HTTP)
  └─ .Slog() → slogHandler ─┘
```

- **Handler** = `Enabled(ctx, Level) + Handle(ctx, Entry)`. Thin, no Flush/Sync.
- **Writer** = `io.Writer + Flush + Sync`. Owns I/O lifecycle.
- **Router** = synchronous fan-out Handler. One Encode per encoder group,
  direct Write to each output. No channels — SlabWriter adds async if needed.
- **SlabWriter** = pre-allocated slab pool + background I/O goroutine.
  `WithDropOnFull()` for non-blocking mode.

## Done

### Bag v2 — linked list + slot cache

```go
type Bag struct {
    fields []Field
    parent *Bag               // O(1) With, no copies
    group  string             // WithGroup namespace
    cache  atomic.Pointer[bagCache] // per-encoder slot (max 2)
}
```

`AllocEncoderSlot()` returns 1-based index. Slot 0 = no caching (graceful
degradation). Cache lifetime = Bag lifetime. No LRU, no eviction.

Removes from v1: `Bag.version`, global atomic counter, field merge in With,
encoder-local LRU cache.

### Groups — WithGroup + Group field

- `Group(key, fields...)` — `FieldTypeGroup`. Encoder opens `{`, encodes
  sub-fields, closes `}`. Stored in Bag, cached.
- `Bag.WithGroup(name)` — permanent namespace. Encoder splits group nodes
  (write `"name":{`, no caching) and field nodes (use slot cache).
  Empty trailing groups suppressed via `skipTrailingGroups()`.
- `Group("", fields...)` — inline group (slog compat). Encoder emits fields
  at current level, no wrapping object.
- `Object("", obj)` — inline object, same as `Inline(obj)`.

### Level on Handler, no WithLevel

Level lives on Handler via `Enabled(context.Context, Level) bool`.
Logger caches enabler at construction. Router returns OR of sub-handlers.
`MutableLevel` provides atomic runtime level changes.

### Handler interface (simplified from design)

```go
type Handler interface {
    Enabled(context.Context, Level) bool
    Handle(context.Context, Entry) error
}
```

Design doc proposed Flush/Sync on Handler. Current architecture puts them
on Writer interface instead — Handler is pure logic, Writer owns I/O.
This is simpler: middleware (ContextHandler) only forwards Handle/Enabled.
Router calls Flush/Sync on Writer directly at close.

### Router (replaces TeeWriter + RouteWriter)

`NewRouter().Route(enc, Output(level, w)...).Build()` returns Handler + close.
Synchronous fan-out: one Encode per encoder group, direct Write to each
output. For async I/O, wrap writer in SlabWriter.

Replaces: TeeWriter (deleted), RouteWriter design with per-destination
channelWriters. Simpler: no channels in router, async is opt-in via
SlabWriter.

### SlabWriter (replaces ChannelWriter)

Pre-allocated slab pool + background I/O goroutine.
`WithDropOnFull()` for non-blocking mode. `WithFlushInterval()` for idle
flush. `Stats()` / `Dropped()` for observability.

Replaces: ChannelWriter (deleted). Simpler overflow model — boolean flag
instead of enum OverflowStrategy. No Observer interface (use `Dropped()`
counter instead).

### WithAttrs conversion

`slogHandler.WithAttrs()` converts `[]slog.Attr` → `[]Field`, stores in Bag.

### Logger.Slog()

`logf.Logger` → `*slog.Logger` bridge. Transfers handler, Bag, name.
slog has no WithName — `Logger.Slog()` is the only way name reaches slog.

### slog contract compliance

Passes `testing/slogtest.TestHandler`. All contract rules handled:
zero time, zero PC, resolved attrs, zero attrs, inline groups,
empty groups, WithGroup(""), WithAttrs(nil), concurrency, context.

## TODO: Roadmap

### 1. Production Constructor (high) — NEEDED

Single call to set up the full pipeline:

```go
handler, close, err := logf.NewHandler(logf.Config{
    Output: file,
    Level:  logf.LevelInfo,
})
defer close()
logger := logf.New(handler)
slogger := logger.Slog()
```

Currently users must manually wire Router + ContextHandler + SlabWriter +
Encoder. A convenience constructor eliminates boilerplate for the 80% case.
Power users still use Router directly.

**Verdict: worth doing.** Most users want "log JSON to file" in 3 lines.

### 2. SlogFromContext (medium) — RECONSIDER

```go
slogger := logf.SlogFromContext(ctx)
```

Requires storing Handler in context (Logger.w is private). Alternative:
store Logger in context, call `.Slog()` on it. `logfc.Get(ctx).Slog()`
already works if Logger is in context.

**Verdict: low value.** `logfc.Get(ctx).Slog()` is one extra method call.
Only worth it if slog-only users are a primary audience.

### 3. OTel FieldSource (medium) — DEFERRED

Sub-package `logfotel` with `FieldSource` extracting trace_id/span_id
from OTel context. ContextHandler already supports FieldSource — this is
just a concrete implementation.

**Verdict: do when needed.** ~20 lines of code, trivial to add later.
Depends on whether users actually use OTel + logf together.

### 4. HTTP Field Middleware (low) — SKIP

Field injection (method, path, request_id). Every project has its own
middleware conventions. A generic one adds little value.

**Verdict: skip.** Better as an example in docs than a sub-package.

### 5. ErrAttr (trivial) — SKIP

`logf.ErrAttr(err)` → `slog.Attr`. One-liner: `slog.Any("error", err)`.

**Verdict: skip.** Not worth a public API for a one-liner.

### 6. ReplaceAttr support (low) — DEFERRED

Needs options struct on `NewSlogHandler`. Useful for redacting sensitive
fields, renaming keys, dropping attrs. zap has this.

**Verdict: add when users request it.** Infrastructure exists (empty-attr
check in attrToField), but no demand yet.

### 7. ErrorEncoder cleanup (low) — DEFERRED

`errorEncoderGetter`/`Default()` is over-engineered. `encodeError` copies
config on every call. Works fine, just ugly.

**Verdict: clean up opportunistically.** Not blocking anything.

### 8. ForwardHandler (low) — SKIP

Embed helper for middleware: auto-delegate Enabled to Next.
With Handler being just Handle+Enabled, embedding is trivial:

```go
type MyHandler struct{ next Handler }
func (h *MyHandler) Enabled(ctx context.Context, l Level) bool { return h.next.Enabled(ctx, l) }
```

**Verdict: skip.** Two-method interface doesn't need a helper.

### 9. Debug Ring Buffer (idea) — DEFERRED

Post-mortem debugging: keep last N debug entries in memory, dump on
trigger. Interesting but complex (lifecycle, memory bounds, trigger
mechanism). No current demand.

**Verdict: revisit if post-mortem debugging becomes a real need.**
