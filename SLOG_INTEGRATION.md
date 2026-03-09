# slog Integration Roadmap

logf and slog are not competing APIs — they are complementary.
slog provides stdlib portability; logf provides async I/O, context-scoped
fields, and buffered writing. The goal is full interchangeability:
use either API anywhere, share the same pipeline, see the same fields.

Best of both worlds, not a choice between them.

## Current State

- `logf.NewSlogHandler(w, opts)` — drop-in `slog.Handler` backed by
  logf's channel writer, buffered appender, and JSON encoder.
- `logf.With(ctx, fields...)` — attaches fields to `context.Context`
  via Bag. The slog handler picks them up automatically through
  `ContextWriter`.
- `logfc` package — helper to store/retrieve `logf.Logger` in context.

## 1. Production Constructor

**Status: postponed. Priority: high.**

Current setup requires 5 lines and knowledge of the internal wiring order
(encoder → appender → channel writer → context writer). Even the author
forgets the sequence. This must become a single call:

```go
rt, close := logf.NewRuntime(logf.Config{
    Output: file,
    Level:  logf.LevelInfo,
})
defer close()

logger := rt.Logger()           // *logf.Logger
slogger := rt.SlogLogger()      // *slog.Logger (same pipeline)
```

Key requirement: the user should never need to know what `WriteAppender`,
`ChannelWriter`, or `ContextWriter` are to get a working production logger.
One function, one config, done.

## 2. SlogFromContext

**Status: postponed. Priority: high. Requires design work.**

Bridge between logf context and slog world. One context, two API entry
points, single async pipeline.

Problem: your code uses `logfc.Get(ctx)`, a colleague's library uses
`slog.InfoContext(ctx, ...)`. Without a bridge, these are two parallel
worlds — different outputs, different fields, different buffering.

```go
// Middleware — one line:
ctx = logf.NewContext(ctx, logger)

// Your code — logf API:
logfc.Info(ctx, "processing", logf.String("order_id", id))

// Colleague's library — slog API, same pipeline:
slogger := logf.SlogFromContext(ctx)
slogger.InfoContext(ctx, "validating", "amount", amount)
// ↑ goes through logf async channel, sees request_id from Bag
```

Open questions:

- Logger.w is private. SlogFromContext needs EntryWriter access.
  Options: expose Logger.Writer(), store EntryWriter in context,
  or derive from Runtime (#1).
- Allocation: creating *slog.Logger on every call vs caching in context.
  Per-request call in middleware is fine; per-log-line is not.
- Interaction with slog.SetDefault: if Runtime sets global slog default,
  SlogFromContext is only needed for per-request logger variants.

## 3. WithAttrs — Unified Context Fields

Allow adding slog-style attributes to context, visible to both logf and slog.

```go
// Middleware — uses slog types, no logf import needed
ctx = logf.WithAttrs(ctx, slog.String("request_id", rid))

// Works in both:
logger.Info(ctx, "done")               // logf — sees request_id
slog.InfoContext(ctx, "done")           // slog — sees request_id
```

`WithAttrs` converts `slog.Attr` → `logf.Field` once (cold path) and stores
in the context Bag. No conversion on the hot path.

## 4. OTel FieldSource

Built-in `FieldSource` that extracts `trace_id` and `span_id` from
OpenTelemetry context. One line to enable:

```go
w := logf.NewContextWriter(next, logf.OTelFieldSource())
```

Every log entry automatically includes trace correlation. Works for both
logf and slog (through the shared ContextWriter).

Separate build tag or sub-package (`logfotel`) to avoid hard dependency
on OpenTelemetry.

## 5. HTTP Field Middleware

Not a logging middleware — a field injection middleware. Enriches context
with request metadata (method, path, request_id, remote_addr). Any logger
downstream sees these fields automatically.

```go
mux.Use(logf.HTTPMiddleware(logf.HTTPMiddlewareConfig{
    RequestID: true,    // generates or reads X-Request-ID
    Method:    true,
    Path:      true,
}))
```

Handler code just calls `logger.Info(ctx, "done")` — request fields are
already there.

## 6. ErrAttr — Dual Error Helper

```go
// logf
logger.Error(ctx, "failed", logf.Err(err))

// slog — same semantics, returns slog.Attr
slog.ErrorContext(ctx, "failed", logf.ErrAttr(err))
```

`ErrAttr` returns `slog.Attr` following logf conventions: `"error"` key,
verbose error field if the error implements `VerboseError`.

## Priority

| # | Feature         | Value                      | Complexity |
| - | --------------- | -------------------------- | ---------- |
| 1 | Production ctor | lowers barrier to entry    | low        |
| 2 | SlogFromContext | dual API from one context  | low        |
| 3 | WithAttrs       | unified context fields     | low        |
| 4 | OTel source     | observability ready        | medium     |
| 5 | HTTP middleware  | zero-config request fields | medium     |
| 6 | ErrAttr         | API polish                 | trivial    |

## Accepted Decisions

### Level Architecture — Final Decision

**Level lives on Writer (via LevelEnabler), not on Logger. No WithLevel.**

**Current implementation:** `Enabled(context.Context, Level) bool` added directly
to `EntryWriter` interface. Writer constructors accept level explicitly:
`NewChannelWriter(level, cfg)`, `NewUnbufferedEntryWriter(level, appender)`.
Each writer stores level and checks via `Level.Enabled(lvl)` — simple static
comparison. `ContextWriter` delegates `Enabled` to the next writer.

**TODO:** Revisit level checker design. Current static level on each writer
doesn't cover: dynamic levels (MutableLevel as writer wrapper?), per-request
level override via context, LevelEnabler as separate optional interface
(current design bakes Enabled into EntryWriter — may want to decouple later).

**TODO:** Rename `unbufferedEntryWriter` — name is misleading. It's a
synchronous writer (encodes and writes in the caller's goroutine), not just
"unbuffered". Also lacks a mutex, so parallel writes are unsafe. Options:
rename to `SyncWriter`/`DirectWriter`, add mutex or document single-goroutine
contract.

Logger has no level field. Level check is delegated to Writer via optional
`LevelEnabler` interface with context:

```go
type LevelEnabler interface {
    Enabled(context.Context, Level) bool
}
```

Logger caches the enabler at construction to avoid per-call type assertion:

```go
type Logger struct {
    w       EntryWriter
    enabler LevelEnabler  // cached from w, nil if w doesn't implement
    // ... name, bag
}

func NewLogger(w EntryWriter) *Logger {
    l := &Logger{w: w}
    if e, ok := w.(LevelEnabler); ok {
        l.enabler = e
    }
    return l
}

func (l *Logger) enabled(ctx context.Context, lvl Level) bool {
    if l.enabler != nil {
        return l.enabler.Enabled(ctx, lvl)
    }
    return true
}
```

teeWriter returns OR of all sub-writers:

```go
func (t *teeWriter) Enabled(ctx context.Context, level Level) bool {
    for _, w := range t.writers {
        if enabler, ok := w.(LevelEnabler); ok && enabler.Enabled(ctx, level) {
            return true
        }
    }
    return false
}
```

#### How we arrived at this decision

**Step 1: Enabled on Writer for fan-out.** Without per-destination level
check, fan-out with ring buffer forces all destinations to receive and
encode entries at the lowest level — wasting ~200ns per discarded entry.
Solution: optional `LevelEnabler` interface on `EntryWriter`.

**Step 2: v1 WithLevel bug discovered.** v1 `WithLevel` uses AND
composition — can only narrow, never widen:

```go
// v1: AND — broken
logger := NewLogger(LevelError, w)
logger.WithLevel(LevelDebug)  // Debug && Error = Error. Debug still blocked.
```

First fix idea: change AND to REPLACE. Makes WithLevel bidirectional.

**Step 3: duplicate level problem.** If level is on Logger AND on Writer,
fan-out requires specifying level twice — on Logger (min of all writers)
and on each Writer. Logger level is derived information, error-prone.

**Step 4: level on Writer only — eliminates duplication.** But then
WithLevel override on Logger can't be "undone" (no way to say "go back
to Writer's level"). Explored `WithLevel(LevelFromWriter)` sentinel —
ugly.

**Step 5: WithLevel is not needed.** Research confirmed: no major Go
logger (zap, zerolog, logrus, slog) has per-request level as a commonly
used pattern. Real per-request override (A/B testing) is more complex
than "just set Debug" — it's per-destination ("debug for file but not
for kibana"). This is Writer territory, not Logger. Libraries won't
override level; business logic will, and it will write custom solutions.

**Step 6: context in Enabled.** For per-request decisions (A/B testing,
per-tenant debug), Writer needs request context. slog does the same:
`Handler.Enabled(context.Context, slog.Level) bool`. This makes Writer
the single authority for level decisions — global, per-destination,
and per-request.

**Step 7: no logger name in Enabled.** Considered passing logger name
to allow per-component filtering ("debug for http handlers"). Decided
against: YAGNI, name is available in context fields if needed, and it
bloats the interface for every middleware writer.

#### What we remove from v1

- `Logger.level` field
- `WithLevel()` method
- `LevelChecker` type
- `LevelCheckerGetter` interface
- `LevelCheckerGetterFunc` type
- REPLACE/AND debate

#### What we add

- `LevelEnabler` optional interface on `EntryWriter`
- `Enabled(context.Context, Level) bool` — with context for per-request
- `MutableLevel` now implements `LevelEnabler` (was `LevelCheckerGetter`)

### WriteEntry returns error

v1 `WriteEntry` returns nothing:

```go
// v1
type EntryWriter interface {
    WriteEntry(context.Context, Entry)
}
```

v2 adds error return for chain (middleware) awareness:

```go
// v2
type EntryWriter interface {
    WriteEntry(context.Context, Entry) error
}
```

Logger ignores the error (same as slog `_ = handler.Handle(ctx, r)`),
but middleware writers can react:

- `teeWriter`: if one destination fails, continues to the rest, returns
  `errors.Join`.
- `filterWriter`: propagates error from next writer.
- `channelWriter`: returns error if channel is full (backpressure signal).

Without error return, chain writers are blind to downstream failures.

### Bag v2: linked list + encoder slot cache

v1 Bag merges all parent fields into a flat slice on every `With`. v2
uses a linked list and per-encoder-slot cache on the Bag itself.

#### v1 problems

1. `Bag.With` copies all parent fields — O(N) per call.
2. Encoder-local LRU cache (`map[int32][]byte`, limit=100) — LRU can
   evict live entries, `container/list` allocates per `Set`.
3. Global `atomic.AddInt32` for version on every Bag creation.

#### v2 design

```go
type Bag struct {
    fields []Field
    parent *Bag       // linked list, O(1) With
    cache  [][]byte   // per-encoder slot, len = slotCount
}
```

- **Linked list.** `With` creates a new node pointing to parent.
  No field copy, no merge, no atomic. O(1).
- **Slot cache.** Each encoder gets a slot index (0, 1, ...) assigned
  at Runtime creation (single-threaded). Encoder reads/writes only its
  slot. No mutex, no atomic, no CAS — different goroutines touch
  different slots.
- **Cache lifetime = Bag lifetime.** When Bag is GC'd, cache is GC'd.
  No LRU, no eviction, no limit.
- **version field removed.** Not needed — cache is on Bag, not in
  encoder-local map.

#### Slot assignment

Runtime assigns slots at creation (single-threaded, before any
logging):

```go
func NewRuntime(opts ...Option) *Runtime {
    slot := 0
    for _, dest := range rt.destinations {
        dest.encoder.slot = slot
        slot++
    }
    rt.slotCount = slot
}
```

Typical: JSON file (slot 0) + text stderr (slot 1) = 2 slots.

#### Encoder with slot cache

```go
func (f *jsonEncoder) encodeBag(bag *Bag) {
    if bag == nil {
        return
    }
    if data := bag.cache[f.slot]; data != nil {
        f.buf.AppendBytes(data)  // cache hit, ~3ns memcopy
        return
    }

    start := f.buf.Len()
    f.encodeBag(bag.parent)          // parent first (may hit cache)
    for _, field := range bag.fields {
        field.Accept(f)              // encode only own fields
    }

    encoded := make([]byte, f.buf.Len()-start)
    copy(encoded, f.buf.Data[start:])
    bag.cache[f.slot] = encoded      // cache includes parent bytes
}
```

Recursive: parent cache is reused. First log from `http` logger:
encode `base` (cache miss) + encode `http` (cache miss). Second log:
one cache hit for `http` (includes base bytes). Parent shared between
`http` and `db` loggers — base cache hit for both after first encode.

#### Context Bag: seed + inherit

Problem: `logf.With(ctx, fields...)` does not know slotCount.

Solution: `NewContext` seeds an empty root Bag with slotCount from
Logger. All subsequent `logf.With` calls inherit slotCount from parent.

```go
func NewContext(ctx context.Context, l *Logger) context.Context {
    ctx = context.WithValue(ctx, loggerKey, l)
    return context.WithValue(ctx, bagKey, &Bag{
        cache: make([][]byte, l.slotCount),  // seed
    })
}

func With(ctx context.Context, fs ...Field) context.Context {
    parent := bagFromContext(ctx)
    var cache [][]byte
    if parent != nil {
        cache = make([][]byte, len(parent.cache))  // inherit
    }
    return context.WithValue(ctx, bagKey, &Bag{
        fields: fs,
        parent: parent,
        cache:  cache,
    })
}
```

Typical middleware flow:

```go
ctx = logf.NewContext(ctx, logger)                    // seed
ctx = logf.With(ctx, logf.String("request_id", rid)) // inherited
ctx = logf.With(ctx, logf.String("user_id", uid))    // inherited
logger.Info(ctx, "done")                              // cache hit
```

If `logf.With` is called without seed (no `NewContext`), parent is nil,
cache is nil — graceful degradation, encoder encodes every time.

#### No sync, no atomic on hot path

- Slot assignment: single-threaded at creation.
- Encoder writes only its own slot — no cross-goroutine access.
- Cache allocation: at `NewContext` (seed) or `With` (inherit) —
  caller goroutine, before any channel send.
- Channel send provides happens-before for encoder reads.

#### Bag: what we remove from v1

- `Bag.version int32` and global `nextVersion` atomic counter
- `Bag.With` field merge (O(N) copy)
- Encoder-local `Cache` (LRU with `container/list`)

#### Bag: what we add

- `Bag.parent *Bag` — linked list
- `Bag.cache [][]byte` — per-encoder slot
- Seed pattern via `NewContext`
- Encoder `slot int` assigned by Runtime

### Groups: WithGroup and Group field

slog.Handler has two group mechanisms. logf supports both.

#### `logf.Group(key, fields...)` — inline field (native API)

New `FieldTypeGroup`. Self-contained: only listed fields are nested.

```go
logger.Info(ctx, "done",
    logf.Group("request", logf.String("id", "abc"), logf.Int("status", 200)),
    logf.String("duration", "50ms"),  // outside group
)
// → {"msg":"done", "request":{"id":"abc", "status":200}, "duration":"50ms"}
```

Works in `With` too — Group is a regular Field, stored in Bag, cached:

```go
logger2 := logger.With(logf.Group("request", logf.String("id", "abc")))
```

Encoder: open `{`, encode sub-fields, close `}`. No Bag interaction.

#### `Logger.WithGroup(name)` — permanent namespace (slog compat)

All subsequent fields (from With and per-call) are nested under group.
Irreversible — no CloseGroup. Needed for slog.Handler.WithGroup.

```go
httpLogger := logger.WithGroup("http")
httpLogger.Info(ctx, "done", logf.Int("status", 200))
// → {"msg":"done", "http":{"status":200}}
```

Implemented via `Bag.group` field:

```go
type Bag struct {
    fields []Field
    parent *Bag
    group  string     // empty = no group
    cache  [][]byte
}
```

Encoder tracks open groups, closes them after all fields:

```go
func (f *jsonEncoder) encodeBag(bag *Bag) {
    if bag == nil { return }
    if bag.cache != nil {
        if data := bag.cache[f.slot]; data != nil {
            f.buf.AppendBytes(data.bytes)
            f.openGroups += data.groups
            return
        }
    }
    start := f.buf.Len()
    startGroups := f.openGroups
    f.encodeBag(bag.parent)
    if bag.group != "" {
        f.addKey(bag.group)
        f.buf.AppendByte('{')
        f.openGroups++
    }
    for _, field := range bag.fields { field.Accept(f) }
    // cache stores bytes + open group count
}
```

#### WithGroup vs WithName

They do NOT compete:

- **WithName** — logger identity. Flat dotted path: `fb.agent.rt`.
  Stays on `Logger.name` and `Entry.LoggerName`.
- **WithGroup** — field namespace. Nested JSON: `{"http":{"status":200}}`.
  Lives in Bag.group.

WithName = "where am I?" (metadata). WithGroup = "nest my fields" (structure).

#### slog adapter

```go
func (h *slogHandler) WithGroup(name string) slog.Handler {
    h2 := h.clone()
    h2.bag = &Bag{group: name, parent: h.bag,
        cache: make([][]byte, h.slotCount)}
    return h2
}

func (h *slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
    h2 := h.clone()
    h2.bag = h.bag.With(h.slotCount, convertAttrs(attrs)...)
    return h2
}
```

Bag linked list preserves WithGroup/WithAttrs ordering naturally.
Multiple WithAttrs after WithGroup all land inside the group — no
duplicate keys, no wrapping hacks.

### Logger.Slog() — bridge to slog

logf is the primary API. slog.Logger is derived from logf.Logger:

```go
func (l *Logger) Slog() *slog.Logger {
    return slog.New(&slogHandler{
        w:         l.w,
        bag:       l.bag,
        name:      l.name,       // transferred
        slotCount: l.slotCount,
    })
}
```

Name transfers automatically:

```go
httpLogger := logf.NewLogger(w).WithName("fb").WithName("agent")

httpLogger.Info(ctx, "request")          // {"name":"fb.agent", ...}
httpLogger.Slog().Info("request")        // {"name":"fb.agent", ...}

logf.Default().Slog().Info("request")    // {"msg":"request"}  — no name, ok
```

slog has no WithName. Logger.Slog() is the only way name reaches slog.
Without it, slog handler has no name — empty string, field omitted.
Explicit, deterministic, no context dependency.

### Zero-cost opt-in

Every feature is pay-only-if-used. Three levels of involvement:

```go
// Level 1: Logger.With only — no context Bag at all
logger := logf.NewLogger(w).WithName("http").With(logf.String("env", "prod"))
logger.Info(ctx, "request", logf.Int("status", 200))
logger.Slog().Info("request", "status", 200)
// Both: {"name":"http", "env":"prod", "status":200}

// Level 2: Logger.With + context fields
ctx = logf.NewContext(ctx, logger)
ctx = logf.With(ctx, logf.String("request_id", rid))
logger.Info(ctx, "done")
// {"name":"http", "env":"prod", "request_id":"abc", ...}

// Level 3: Pure slog, logf is just backend
slog.SetDefault(slog.New(logf.NewSlogHandler(w)))
slog.Info("request", "status", 200)
// Async, buffered — but no name, no context fields
```

What you skip costs nothing:

- **No context Bag** → no seed/inherit allocations, encoder skips context bag.
- **No WithGroup** → `Bag.group` empty, encoder skips.
- **No Slog()** → slog handler never created.
- **No WithName** → name empty, field omitted from output.

### Interface architecture: one Handler, one Encoder

v2 merges EntryWriter and Appender into a single Handler interface.
Appender is eliminated.

```go
// Single pipeline interface — middleware and terminals alike.
// Renamed from EntryWriter. Appender merged in.
type Handler interface {
    Enabled(context.Context, Level) bool
    Handle(context.Context, Entry) error
    Flush() error
    Sync() error
}

// Serialization — pure, no I/O, no lifecycle.
type Encoder interface {
    Encode(*Buffer, Entry) error
}
```

Middleware uses embed helper to avoid Flush/Sync/Enabled boilerplate:

```go
type ForwardHandler struct{ Next Handler }

func (f ForwardHandler) Enabled(ctx context.Context, lvl Level) bool {
    return f.Next.Enabled(ctx, lvl)
}
func (f ForwardHandler) Flush() error { return f.Next.Flush() }
func (f ForwardHandler) Sync() error  { return f.Next.Sync() }

// Middleware — only Handle:
type samplingHandler struct {
    ForwardHandler
    rate int
}
func (h *samplingHandler) Handle(ctx context.Context, e Entry) error {
    if !h.sample() { return nil }
    return h.Next.Handle(ctx, e)
}
```

#### Why one interface (not two)

The v1 split (EntryWriter + Appender) forced channelWriter to accept
only Appender. Adding a second async destination required duplicating
the async runtime at Appender level (nonBlockingAppender in logfx) —
a workaround for channelWriter not accepting Handler downstream.

With one interface, channelWriter takes Handler:

```text
Logger → contextHandler → channelWriter → teeHandler → writeHandler (file)
                                                      → channelWriter₂ → dumpServerHandler (HTTP)
```

Full composition. No nonBlockingAppender needed. Each destination gets
its own channelWriter with independent channel and worker.

#### Why Flush and Sync on Handler

User never calls Flush/Sync directly — only Close().
channelWriter manages lifecycle internally (drain → flush → sync).

But Flush/Sync must be on the public interface because channelWriter's
downstream is Handler, and channelWriter needs to flush/sync it.
Optional interfaces (Flusher/Syncer) break when middleware wraps
a terminal — middleware hides the terminal's Flusher from channelWriter.

Flush ≠ Sync:

- **Flush** — write buffered data to OS (cheap, frequent, on idle)
- **Sync** — fsync to disk (expensive, rare, on close/error)

zap conflates both into Sync() because it has no runtime — user calls
`defer logger.Sync()` manually. logf's channelWriter separates them
for performance: Flush on every channel drain, Sync only on shutdown.

#### channelWriter owns its context

channelWriter does NOT pass request context through the channel.
It creates its own context for downstream I/O lifecycle:

```go
func (l *channelWriter) init(cfg Config) {
    l.ctx, l.cancel = context.WithCancel(context.Background())
    l.ch = make(chan Entry, l.Capacity)
    go l.worker()
}

func (l *channelWriter) close() {
    l.cancel()     // signal downstream: cancel last I/O
    close(l.ch)    // drain remaining
    l.Wait()
}

func (l *channelWriter) handle(e Entry) {
    l.Next.Handle(l.ctx, e)  // channelWriter's context, not request's
}
```

Request context is irrelevant at I/O level — it was about "what to log",
not "where to write". Downstream handlers use channelWriter's context
for graceful shutdown of network/DB connections.

#### channelWriter: overflow strategy and observer

From production experience (nonBlockingAppender in logfx/runvm-agent):
channelWriter v2 supports configurable overflow and opt-in metrics.

```go
type ChannelWriterConfig struct {
    Capacity    int
    OnOverflow  OverflowStrategy       // Block (default) | Drop
    Observer    ChannelWriterObserver   // nil = no-op (zero cost)
}

type ChannelWriterObserver interface {
    OnHandleStarted(entryTimestamp time.Time)
    OnHandleCompleted(entryTimestamp time.Time, err error)
    OnDropped(Entry)
    OnFlushStarted()
    OnFlushCompleted(error)
    OnSyncStarted()
    OnSyncCompleted(error)
}
```

- **Block** (v1 default) — caller waits, nothing lost.
- **Drop** — caller never blocked, entry lost, observer notified.

This eliminates the need for nonBlockingAppender — channelWriter
handles both modes. Observer enables latency/error/drop metrics
without coupling to a specific metrics library.

#### teeHandler — fan-out for Handler

```go
func NewTeeHandler(handlers ...Handler) Handler
```

Delegates Handle, Flush, Sync to all targets. Enabled returns true
if any target is enabled. Replaces compositeAppender from logfx.

```text
Logger → contextHandler → channelWriter → teeHandler → writeHandler (file)
                                                      → channelWriter₂ → dumpServerHandler
```

Each destination can have its own channelWriter (independent async,
independent overflow strategy). Or share one channelWriter upstream
if all destinations are fast enough.

### Debug Ring Buffer — separate design

**Status: idea. Separate from slog integration.**

Post-mortem debugging: when a problem occurs, debug logs are already lost.
Dynamic level only helps for reproducible issues ("enable debug, wait,
catch"). A ring buffer solves the non-reproducible case:

- Always keep last N debug entries in memory (circular buffer).
- On trigger (panic, error, signal) — dump the buffer to file.
- Caller pays only for Entry creation (~50ns), no encoding until dump.
- After dump, buffer resets and continues.
- Depends on LevelEnabler: file channelWriter skips Debug (Enabled=false),
  only ring buffer accepts them (Enabled=true).

This would be a unique feature — neither slog, zap, nor zerolog offer it.

## Design Principles

- **Interchangeability.** Use `logf.Logger` or `slog.Logger` anywhere
  in the same application. Both write to the same async pipeline, both
  see the same context fields. Switching between them is zero-cost.
- **Complementarity.** slog brings stdlib portability and ecosystem.
  logf brings async I/O, context Bag, buffered writing. Together they
  cover more than either alone.
- **Context is the field bus.** Fields attached to context are visible
  to any logger (logf or slog) that uses the shared ContextWriter.
- **One-line migration.** Switching from `slog.NewJSONHandler` to logf
  backend should require changing one line. No code changes elsewhere.
- **No forced dependency.** OTel, HTTP middleware, and other integrations
  live in sub-packages. Core logf has zero external dependencies beyond
  stdlib.
