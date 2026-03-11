# Lazy Fields — Design Notes

## Three Models of "Lazy" in Logging

### 1. zap WithLazy — deferred encoding (sync.Once)

```go
// zap/zapcore/lazy_with.go
type lazyWithCore struct {
    Core
    sync.Once
    fields []Field // already captured values
}

func (d *lazyWithCore) initOnce() {
    d.Once.Do(func() {
        d.Core = d.Core.With(d.fields) // encoding happens here
    })
}
```

- **Value capture**: eager (at Field creation)
- **Encoding**: lazy (sync.Once, on first log or chained With)
- **Fresh per log**: no — values frozen at WithLazy() call
- **Why needed**: zap's `core.With()` encodes eagerly. WithLazy skips encoding
  for loggers that might never be used (error paths, rare branches)

### 2. slog LogValuer — lazy value capture (per log)

```go
type LogValuer interface {
    LogValue() slog.Value
}

func (v Value) Resolve() Value {
    for i := 0; i < 100; i++ {
        if v.Kind() != KindLogValuer { return v }
        v = v.LogValuer().LogValue()
    }
    return AnyValue(errors.New("too many LogValue calls"))
}
```

- **Value capture**: lazy (LogValue() called at Resolve time)
- **Encoding**: at Handle time, after Resolve
- **Fresh per log**: yes — LogValue() called each time
- **Chains**: LogValue can return another LogValuer (resolved recursively, max 100)
- **Transparent**: any type implementing `LogValue() Value` becomes lazy automatically
- **Use cases**:
  - Dynamic metrics (`pool.Active()` — fresh each log)
  - Masking sensitive data (`Token.LogValue()` → masked string)
  - Expensive computation skipped if level disabled

### 3. logf — lazy encoding by architecture

```
logf With() → Field structs in Bag (no encoding)
                 ↓
First log    → Bag cache encodes once
                 ↓
Next logs    → Bag cache hit (byte copy)
```

- **Value capture**: eager (at Field creation)
- **Encoding**: lazy (deferred to first log, then cached via Bag slot cache)
- **Fresh per log**: no for With(); yes for Snapshotter
- **No WithLazy needed**: logf's With() already does what zap WithLazy does

## logf Snapshotter — lazy value capture (existing)

```go
type Snapshotter interface {
    TakeSnapshot() interface{}
}
```

- Called at log time in caller's goroutine (before async send)
- Returns `interface{}` — loses type info, encoder uses reflection
- Original purpose: race-safety with ChannelWriter
- Side effect: lazy value capture (fresh per log)
- Works via `Any("key", &obj)` — no special wrapper needed

### Snapshotter vs slog LogValuer

| Aspect              | slog LogValuer              | logf Snapshotter           |
|---------------------|-----------------------------|----------------------------|
| API                 | `Any("k", &obj)` (both)    | `Any("k", &obj)` (both)   |
| Return type         | typed `slog.Value`          | `interface{}` (untyped)    |
| Chains              | yes (recursive)             | no                         |
| Ergonomics          | just add `LogValue()` method | just add `TakeSnapshot()`  |
| Purpose             | user-facing lazy API        | internal race-safety       |
| Documented as lazy  | yes                         | no                         |

## slog→logf Bridge: LogValuer Handling

- **Per-call fields**: `attrToField()` calls `Value.Resolve()` → LogValue() called each log ✅
- **WithAttrs (accumulated)**: `convertAttrs()` → `attrToField()` → `Resolve()` called once at
  WithAttrs time → value frozen in Bag. Valid per slog spec (handler may resolve eagerly).

### Future: preserve LogValuer laziness in WithAttrs

Currently WithAttrs resolves LogValuer eagerly (converts to native Field).
With FieldValuer in place, bridge could store LogValuer natively:

```go
func (h *slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
    fields := make([]Field, 0, len(attrs))
    for _, a := range attrs {
        if lv, ok := a.Value.Any().(slog.LogValuer); ok {
            // Wrap in FieldValuer adapter — stays lazy in Bag
            fields = append(fields, Any(a.Key, slogLogValuerAdapter{lv}))
        } else {
            fields = append(fields, attrToField(a))
        }
    }
    return &slogHandler{
        w:    h.w,
        bag:  h.bag.With(fields...),
        ...
    }
}

// Adapter: slog.LogValuer → logf.FieldValuer
type slogLogValuerAdapter struct{ lv slog.LogValuer }

func (a slogLogValuerAdapter) LogValue() Field {
    v := slog.Value(a.lv.LogValue()).Resolve()
    return slogValueToField("", v) // convert resolved Value → Field
}
```

This way LogValuer fields in With() would be re-evaluated at each log,
matching stdlib slog behavior. Requires FieldValuer to be implemented first.

Note: Bag cache will cache the first evaluation. To get truly fresh values
per log, FieldValuer fields would need to bypass or invalidate cache — this
is a design decision (cache consistency vs freshness). Options:

1. **Cache normally** — LogValue() called once, cached. Consistent with
   logf's encode-once philosophy. Different from stdlib slog but valid per spec.
2. **Skip cache for FieldValuer nodes** — fresh each log. Matches stdlib
   slog behavior but loses Bag cache performance advantage.
3. **Hybrid** — cache by default, opt-out via `WithDynamic()` for fresh values.

Recommendation: option 1 (cache normally). logf's value proposition is
performance. Users wanting per-log freshness use per-call fields.

## Key Insight

logf's lazy encoding is a **core architectural advantage**, not a bolt-on optimization.
zap had to add WithLazy in v1.26 to partially catch up because their `core.With()`
does eager encoding by design.

## Proposal: Position Lazy as logf's Key Advantage

### Problem

Lazy is logf's strongest differentiator but it's hidden:
- `With()` is lazy — but nowhere stated as an advantage
- `Snapshotter` — name says "snapshot for race-safety", not "lazy value"
- `TakeSnapshot() interface{}` — untyped, reflection, not ergonomic
- No unified narrative "logf = lazy by design"

### Level 1 — Positioning (documentation only, no code changes)

Name things properly in README/docs:
- `With()` = **zero-cost field accumulation** (vs zap's eager encoding)
- Bag cache = **encode-once semantics**
- Async pipeline = **caller never encodes**

Comparison table for README:

```markdown
|                     | zap With | zap WithLazy | logf With |
|---------------------|----------|--------------|-----------|
| At With() call      | encode   | nothing      | nothing   |
| At first log        | —        | encode+cache | encode+cache |
| Subsequent logs     | —        | cached       | cached    |
```

### Level 2 — FieldValuer interface (new API)

Typed lazy value — like slog LogValuer but native logf:

```go
type FieldValuer interface {
    LogValue() Field
}
```

Example — dynamic metrics:

```go
type PoolStats struct{ pool *Pool }

func (p *PoolStats) LogValue() Field {
    return Group("pool",
        Int("active", p.pool.Active()),
        Int("idle", p.pool.Idle()),
    )
}

logger.Info("tick", Any("pool", &stats))  // LogValue() at each log
```

Example — sensitive data masking:

```go
type SecretToken string

func (t SecretToken) LogValue() Field {
    return String("", "***"+string(t[len(t)-4:]))
}
```

Advantages over Snapshotter:
- **Typed return** — `Field` instead of `interface{}`
- **No reflection** — encoder knows type immediately
- **Same method name as slog** — `LogValue()`, familiar to users
- **Snapshotter stays** for backward compat

Priority in `Any()`:

```go
func Any(key string, value interface{}) Field {
    switch v := value.(type) {
    case FieldValuer:    // new, first priority
        ...
    case Snapshotter:    // backward compat
        ...
    case ObjectEncoder:
        ...
    // ...
    }
}
```

### Level 3 — Functional lazy (ad-hoc, no types needed)

```go
logger.With(LazyField(func() Field {
    return String("trace", expensiveLookup())
}))
```

For one-off lazy fields without creating a named type.
Internally: store func in Field.Any, call at encoding time.

### Level 4 — Snapshotter evolution

Options:
1. **Keep as is** — Snapshotter for race-safety, FieldValuer for lazy values
2. **Deprecate** — migrate users to FieldValuer (breaking change, v3)
3. **Bridge** — if type implements both, FieldValuer wins (checked first in Any)

Recommendation: option 3 — additive, no breakage, gradual migration.

### Summary

| Before                          | After                                      |
|---------------------------------|--------------------------------------------|
| "logf defers encoding" (hidden) | **"logf is lazy by design"** (positioning)  |
| Snapshotter (internal tool)     | **FieldValuer** (user-facing API)           |
| `interface{}` → reflection      | **`Field` → zero reflection**               |
| Not documented                  | **README: "Zero-cost With, encode-once cache, lazy values"** |

## Async Safety — Why Async is the Right Default

### The problem (perceived)

Async logging encodes in a consumer goroutine, not in the caller.
If a reference type (map, slice, pointer) is mutated between the log call
and encoding, the log output may not match the state at call time.

### The reality: 95% of log calls are value types

Real production log calls look like this:

```go
logger.Info("request handled",
    String("method", r.Method),     // string → immutable
    String("path", r.URL.Path),     // string → immutable
    Int("status", status),          // int → copied into Field
    Duration("latency", elapsed),   // int64 → copied into Field
    String("trace_id", traceID),    // string → immutable
)
```

No reference types. Every value copied into Field struct. Async-safe by design.

### logf already has the Xxx/ConstXxx pattern for slices

```go
Ints(k, v)       // make + copy → async-safe (like sync behavior)
ConstInts(k, v)  // raw pointer → caller guarantees immutability
```

Safe by default (`Ints`), fast opt-in (`ConstInts`). Already implemented.

### Sync loggers don't save you from concurrent mutation either

```go
data := map[string]int{"a": 1}
go func() { data["b"] = 2 }()
slog.Info("state", "data", data)  // RACE CONDITION → panic even with sync!
```

Sync only helps with sequential mutation (mutate AFTER log call returns).
For concurrent mutation, both sync and async are broken without a lock or copy.

### The narrow edge case async creates

Sequential mutation between log call and encoding:

```go
data := map[string]int{"a": 1}
logger.Info("state", Any("data", data))  // channel send (~50ns)
data["b"] = 2                             // before encoding
// Log shows {"a":1,"b":2} — not what was true at Info() time
```

This ONLY affects:
- `Any()` with reference types (map, slice, *struct)
- Sequential (not concurrent) mutation after log call
- NOT typed fields (String, Int64, Bool, Duration, etc.)

### Cost of sync vs async

| Metric        | Sync logger       | logf async            |
|---------------|-------------------|-----------------------|
| Caller cost   | encode + write(fd) | channel send (~50ns) |
| p99 latency   | 1µs - 2.5ms      | 50-100ns              |
| Under slow I/O | caller blocked   | caller free           |
| Value types   | safe              | **safe**              |
| Ref mutation  | safe (sequential) | needs Snapshot/copy   |
| Concurrent mut | **broken** (race) | **broken** (race)    |

### Positioning: "Safe by default, never blocks"

- **Typed fields (95%)** — copied by value, async-safe, zero I/O wait
- **Slice fields** — `Ints()` copies (safe), `ConstInts()` skips copy (fast)
- **Reference types** — `Snapshotter` or explicit copy. Same discipline as
  passing data to another goroutine — Go developers already understand this
- **The trade-off**: one narrow edge case (sequential ref mutation) in exchange
  for 50x better p99 latency on every single log call

Async doesn't create new problems. It creates one narrow edge case for
reference types, in exchange for dramatically better latency for everything else.

### Open: should `Any()` auto-snapshot?

Options for reference types in `Any()`:

1. **Current** — store as-is, Snapshotter opt-in. Consistent with `ConstXxx`.
2. **Auto-detect maps/slices** — copy in `Any()`. Safe by default but adds
   allocation for the common case (most refs are not mutated).
3. **`Snapshot(k, v)` constructor** — explicit "copy now", `Any` stays fast.
   Consistent with `Ints`/`ConstInts` split.
4. **`Any` copies, `ConstAny`/`Ref` skips** — flip the default. Breaking change.

Recommendation: option 3 — additive, explicit, no breakage. `Snapshot()` for
maps/slices/pointers, `Any()` stays zero-copy.

## Variant C: Pipeline Handles Copy (not constructors)

Original plan was always variant C — constructors don't copy, pipeline does.

### Responsibilities split

```
Constructor          Pipeline (caller goroutine)    Encoding (consumer goroutine)
───────────          ──────────────────────────     ────────────────────────────
Any() → raw ref      Snapshotter? → TakeSnapshot()  FieldValuer? → LogValue()
Ints() → raw ref     copy slices for known types     encode typed Field
                     (only ChannelWriter)             (always)
```

- **Constructor**: zero-alloc, stores as-is. No copy, no snapshot.
- **ChannelWriter**: in caller goroutine, before channel send — snapshot/copy
  mutable data. Only async pipeline pays this cost.
- **SyncWriter**: no copy needed — encoding happens in caller goroutine immediately.
- **Encoding**: FieldValuer resolved here — lazy typed value.

### Current `Any()` constructor — needs change

```go
// CURRENT (wrong place for snapshot):
if s, ok := rv.(Snapshotter); ok {
    return Field{Key: k, Type: FieldTypeAny, Any: s.TakeSnapshot()}  // eager!
}

// PROPOSED (constructor stores raw, pipeline snapshots):
// Any() just stores the value, doesn't call TakeSnapshot()
return Field{Key: k, Type: FieldTypeAny, Any: v}
// ChannelWriter calls TakeSnapshot() before channel send
```

### Slice constructors — same change

```go
// CURRENT: Ints() copies
func Ints(k string, v []int) Field {
    cc := make([]int, len(v))
    copy(cc, v)
    return ConstInts(k, cc)
}

// PROPOSED: Ints() = ConstInts() (no copy)
// ChannelWriter copies slices before channel send
// Ints/ConstInts distinction removed or Ints becomes alias
```

## FieldValuer vs Snapshotter — Orthogonal Concerns

Two interfaces, two problems, two moments in time:

| Aspect          | FieldValuer (lazy)                  | Snapshotter (async safety)            |
|-----------------|-------------------------------------|---------------------------------------|
| Problem         | lazy value computation + typing     | mutable data with async pipeline      |
| When called     | at encoding time                    | before channel send (caller goroutine)|
| Who calls       | encoder                             | ChannelWriter (pipeline)              |
| Returns         | `Field` (typed)                     | `interface{}` (snapshot of state)     |
| Needed w/ Sync? | yes — lazy is always useful         | no — encoding in caller, no race      |

### Priority when both present on same object

With **ChannelWriter** (async):

1. Snapshotter → TakeSnapshot() → snapshot stored in Field.Any
2. Encoding: snapshot is a plain value, encoded normally (not FieldValuer)

With **SyncWriter** (sync):

1. No snapshot (not needed)
2. Encoding: FieldValuer → LogValue() → encode typed Field

### Practical guidance

- **Mutable object** → implement Snapshotter (async safety)
- **Lazy computation** → implement FieldValuer (typed lazy value)
- **Both on same type** → unlikely in practice; Snapshotter wins with async,
  FieldValuer wins with sync. No conflict — they solve different problems.

They are orthogonal: Snapshotter = "when to copy", FieldValuer = "when to compute".

## Open Questions

- FieldValuer: should `LogValue()` return `Field` or `[]Field`?
- FieldValuer in With(): call LogValue() at each log (fresh) or once (cached)?
  - With() + Bag cache = encode once → LogValue() called once → cached. Same as slog WithAttrs.
  - Per-call fields: LogValue() at each log → fresh. Same as slog Handle.
  - This is the natural behavior, no special handling needed.
- LazyField: worth adding or YAGNI? FieldValuer covers most cases.
- Naming: `FieldValuer` vs `LazyField` vs `DynamicField` vs `LogValuer`?
  - `LogValuer` matches slog but returns different type (Field vs slog.Value)
  - `FieldValuer` is explicit about what it returns
