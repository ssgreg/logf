# slog Integration — Decisions & TODOs

## TODO: Roadmap

### 1. Production Constructor (high)

Single call to set up the full pipeline. User should never need to know
what ChannelWriter or ContextWriter are.

```go
rt, close := logf.NewRuntime(logf.Config{
    Output: file,
    Level:  logf.LevelInfo,
})
defer close()
logger := rt.Logger()           // *logf.Logger
slogger := rt.SlogLogger()      // *slog.Logger (same pipeline)
```

### 2. SlogFromContext (high)

Bridge logf context to slog. One context, two APIs, single pipeline.

```go
ctx = logf.NewContext(ctx, logger)
slogger := logf.SlogFromContext(ctx)  // shares pipeline + Bag fields
```

Open: Logger.w is private — expose Writer(), store in context, or derive from Runtime.

### 3. WithAttrs (low)

`slog.Attr` → `logf.Field` conversion, stored in Bag. Cold path only.

### 4. OTel FieldSource (medium)

Sub-package `logfotel` — extracts trace_id/span_id from OTel context.

### 5. HTTP Field Middleware (medium)

Field injection (method, path, request_id), not logging middleware.

### 6. ErrAttr (trivial)

`logf.ErrAttr(err)` → `slog.Attr` with logf conventions.

### 7. ReplaceAttr support for slog handler

Empty-attr check in `attrToField` exists for this. Needs options struct.

### 8. Revisit level checker design

Current: static level baked into EntryWriter. Doesn't cover dynamic levels
(MutableLevel wrapper?), per-request override via context, or LevelEnabler
as separate optional interface.

### 9. Revisit ErrorEncoder

`DefaultErrorEncoder` uses global mutable state (cached closure).
`errorEncoderGetter`/`Default()` is over-engineered. `encodeError` copies
config on every call. May be unnecessary in v2 Handler/Encoder architecture.

## Accepted: Level on Writer, no WithLevel

**Done.** Level lives on Writer via `Enabled(context.Context, Level) bool`.
Logger has no level field, caches enabler at construction. teeWriter
returns OR of sub-writers. Removed from v1: Logger.level, WithLevel,
LevelChecker, LevelCheckerGetter.

## Accepted: WriteEntry returns error

**Done.** `WriteEntry(context.Context, Entry) error`. Logger ignores error.
Middleware can react: teeWriter uses `errors.Join`, channelWriter signals
backpressure.

## Design: Bag v2 — linked list + slot cache

Not yet implemented.

```go
type Bag struct {
    fields []Field
    parent *Bag       // linked list, O(1) With
    cache  [][]byte   // per-encoder slot, len = slotCount
}
```

- **With** creates a new node pointing to parent. No copy, no atomic.
- **Slot cache**: each encoder gets a slot index assigned by Runtime at
  creation. Encoder reads/writes only its slot. No sync on hot path.
- **Cache lifetime = Bag lifetime**. No LRU, no eviction.
- **Seed pattern**: `NewContext` seeds root Bag with slotCount from Logger.
  `logf.With` inherits slotCount from parent. No seed → nil cache → graceful
  degradation (encode every time).

Removes from v1: `Bag.version`, global atomic counter, field merge in With,
encoder-local LRU cache.

## Design: Groups — WithGroup + Group field

### `logf.Group(key, fields...)` — inline field

`FieldTypeGroup`. Encoder opens `{`, encodes sub-fields, closes `}`.
Works in With — stored in Bag, cached.

### `Logger.WithGroup(name)` — permanent namespace (slog compat)

Irreversible. Stored as `Bag.group string`. Encoder splits group nodes
(write `"name":{`, no caching) and field nodes (use slot cache).
Closing braces via `countGroups()` walk, O(depth), typically 1-3.

### WithGroup vs WithName

- **WithName** — logger identity, flat dotted path. On `Logger.name`.
- **WithGroup** — field namespace, nested JSON. On `Bag.group`.

## Design: Logger.Slog()

`logf.Logger` → `*slog.Logger` bridge. Transfers writer, Bag, name.
slog has no WithName — `Logger.Slog()` is the only way name reaches slog.

## Design: One Handler Interface

v2 merges EntryWriter and Appender into a single Handler interface.

```go
type Handler interface {
    Enabled(context.Context, Level) bool
    Handle(context.Context, Entry) error
    Flush() error
    Sync() error
}

type Encoder interface {
    Encode(*Buffer, Entry) error
}
```

**Why one interface.** v1 split (EntryWriter + Appender) forced channelWriter
to accept only Appender. Adding a second async destination required
nonBlockingAppender in logfx — a workaround. With one interface,
channelWriter takes Handler downstream. Full composition, no workarounds.

**Flush ≠ Sync.** Both on Handler because channelWriter's downstream is
Handler, and optional interfaces (Flusher/Syncer) break through middleware
wrapping. Flush = write buffered data to OS (cheap, on idle). Sync = fsync
(expensive, on shutdown). zap conflates both; logf separates for performance.

**ForwardHandler** — embed helper for middleware. Only override Handle:

```go
type ForwardHandler struct{ Next Handler }
// Enabled, Flush, Sync delegate to Next automatically.
```

## Design: TeeWriter — synchronous fan-out

```go
func NewTeeWriter(writers ...EntryWriter) EntryWriter
```

Delegates Handle, Flush, Sync to all targets in caller's goroutine.
Enabled returns true if any target is enabled. Errors via `errors.Join`.

Same goroutine, same thread — simple, no channels.

## Design: RouteWriter — multi-destination async

Each destination gets its own channelWriter with independent goroutine,
channel, buffer, and overflow strategy:

```text
Logger → contextWriter → routeWriter → channelWriter₁ → writeHandler (file)
                                      → channelWriter₂ → dumpServerHandler (HTTP)
```

- **Independent goroutines** — slow HTTP destination doesn't block fast file.
- **Independent overflow** — file blocks (nothing lost), HTTP drops (never blocks).
- **Independent lifecycle** — each channelWriter owns its context, Flush, Sync.

RouteWriter itself is synchronous fan-out (like TeeWriter) but its children
are channelWriters — so each destination is async independently.

### channelWriter: overflow + observer

From production experience (nonBlockingAppender in logfx/runvm-agent):

```go
type ChannelWriterConfig struct {
    Capacity    int
    OnOverflow  OverflowStrategy       // Block (default) | Drop
    Observer    ChannelWriterObserver   // nil = no-op (zero cost)
}
```

- **Block** — caller waits, nothing lost.
- **Drop** — caller never blocked, entry lost, observer notified.

Observer enables latency/error/drop metrics without coupling to a
specific metrics library. Eliminates need for nonBlockingAppender.

## Idea: Debug Ring Buffer

Post-mortem debugging: keep last N debug entries in memory (ring buffer).
On trigger (panic, error, signal) — dump to file. Caller pays ~50ns,
no encoding until dump. Depends on LevelEnabler per-destination filtering.
