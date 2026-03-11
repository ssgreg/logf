# Benchmarks — Methodology

## 1. Categories

### A. logf core (no writer)

Standalone measurements that don't involve writers:

- **With**: measure `logger.With(fields...)` — Bag creation cost
- **WithGroup**: measure `logger.WithGroup(name)` — group Bag cost
- **logfc context**: measure overhead of extracting logger from context

### B. logf scenarios (Discard + Async)

Full log call including encoding. Writer: SyncWriter → `io.Discard` (encodes fully,
no I/O noise). Async: only 2-3 tests for measuring ChannelWriter overhead,
not the full matrix.

| #  | Name                    | Description                                          |
|----|-------------------------|------------------------------------------------------|
| 0  | DisabledLevel           | `logger.Debug()` when level=Info — fast path cost    |
| 1  | NoFields                | `logger.Info("msg")` — baseline                      |
| 2  | TwoScalars              | `Info("msg", String, Int)` — typical case            |
| 3  | TwoScalarsInGroup       | `Info("msg", Group("g", String, Int))` — inline nested object (logf.Group / zap.Dict / slog.Group / zerolog.Dict) |
| 4  | SixScalars              | `Info("msg", 6× scalar)` — heavy typical             |
| 5  | SixHeavy                | `Info("msg", 6× heavy)` — slices, Time, Bytes, etc. |
| 6  | ErrorField              | `Info("msg", Error(err))` — common pattern           |
| 7  | WithPerCall+NoFields    | With inside loop + `Info("msg")`                     |
| 8  | WithPerCall+TwoScalars  | With inside loop + `Info("msg", 2× scalar)`          |
| 9  | WithCached+NoFields     | With outside loop + `Info("msg")`                    |
| 10 | WithCached+TwoScalars   | With outside loop + `Info("msg", 2× scalar)`         |
| 11 | WithBoth+TwoScalars     | With outside + With inside + `Info("msg", 2×)`       |
| 12 | WithGroupCached+TwoScalars | WithGroup outside + `Info("msg", 2× scalar)`      |
| 13 | Caller+TwoScalars       | AddCaller + `Info("msg", 2× scalar)`                 |

### C. Field type regression (encode only, Sync Text)

One benchmark per unique field type — catches optimization regressions during refactoring.
Each test: single field of given type, full encode cycle.

| Field type                | Constructor example           |
|---------------------------|-------------------------------|
| Bool                      | `Bool("k", true)`             |
| Int64                     | `Int64("k", 42)`              |
| Float64                   | `Float64("k", 3.14)`          |
| String                    | `String("k", "value")`        |
| Duration                  | `Duration("k", 5*time.Second)` |
| Time                      | `Time("k", time.Now())`       |
| Error                     | `NamedError("k", err)`        |
| Bytes                     | `Bytes("k", []byte{...})`     |
| Ints64 (slice)            | `Ints64("k", []int64{...})`   |
| Strings (slice)           | `Strings("k", []string{...})` |
| Bools (slice)             | `Bools("k", []bool{...})`     |
| Floats64 (slice)          | `Floats64("k", []float64{...})` |
| Durations (slice)         | `Durations("k", []Duration{...})` |
| Object                    | `Object("k", &obj)`           |
| Array                     | `Array("k", &arr)`            |
| Group                     | `Group("k", String, Int)`     |
| Stringer                  | `Stringer("k", &obj)`         |
| Formatter                 | `Formatter("k", &obj, verb)`  |
| Any (value type)          | `Any("k", 42)`                |
| Any (Snapshotter)         | `Any("k", &snapshotObj)`      |
| Any (Stringer fallback)   | `Any("k", &stringerObj)`      |
| Any (reference type)      | `Any("k", map[...]...)`       |

### D. Competitors (logf scenario numbers, sync only)

Same scenarios as logf section B, using each competitor's idiomatic API.
Same logical field values for fair comparison.

**slog→logf**: benchmarks create a logf Logger first (with the same config as
native logf tests — caller OFF, level Debug), then derive `slog.Logger` via
`Logger.Slog()`. This ensures the slog→logf bridge inherits the same settings
and measures only the bridge overhead, not accidental caller cost.

| #  | Scenario                   | slog | zap | zerolog | logrus | slog→logf |
|----|----------------------------|------|-----|---------|--------|-----------|
| 0  | DisabledLevel              |  ✓   |  ✓  |    ✓    |   ✓    |     ✓     |
| 1  | NoFields                   |  ✓   |  ✓  |    ✓    |   ✓    |     ✓     |
| 2  | TwoScalars                 |  ✓   |  ✓  |    ✓    |   ✓    |     ✓     |
| 3  | TwoScalarsInGroup          |  ✓   |  ✓  |    ✓    |   —    |     ✓     |
| 4  | SixScalars                 |  ✓   |  ✓  |    ✓    |   ✓    |     ✓     |
| 5  | SixHeavy                   |  ✓   |  ✓  |    ✓    |   ✓    |     ✓     |
| 6  | ErrorField                 |  ✓   |  ✓  |    ✓    |   ✓    |     ✓     |
| 7  | WithPerCall+NoFields       |  ✓   |  ✓  |    ✓    |   ✓    |     ✓     |
| 8  | WithPerCall+TwoScalars     |  ✓   |  ✓  |    ✓    |   ✓    |     ✓     |
| 9  | WithCached+NoFields        |  ✓   |  ✓  |    ✓    |   ✓    |     ✓     |
| 10 | WithCached+TwoScalars      |  ✓   |  ✓  |    ✓    |   ✓    |     ✓     |
| 11 | WithBoth+TwoScalars        |  ✓   |  ✓  |    ✓    |   ✓    |     ✓     |
| 12 | WithGroupCached+TwoScalars |  ✓   |  ✓  |    ✓    |   —    |     ✓     |
| 13 | Caller+TwoScalars          |  ✓   |  ✓  |    ✓    |   ✓    |     ✓     |

Notes:

- logrus: no native group support (3, 12 skipped)
- zerolog: groups via `Dict()` — closest equivalent
- zap: groups via `zap.Namespace()` — closest equivalent
- slog: native `slog.Group()` and `WithGroup()`
- All competitors use io.Discard (or closest equivalent, no I/O noise)
- slog→logf handler: `slog.New(logf.NewSlogHandler(w))` — full B matrix, key production use-case
- **If a competitor lacks a feature for a scenario — do NOT implement a local workaround. Ask what to do.**

---

## 2. Naming Convention

```
Benchmark{API}_{Writer}_{Scenario}

API:     Logf, Logfc, SlogLogf, Slog, Zap, Zerolog, Logrus
Writer:  Discard (default, omit in name), Async (explicit)
```

Examples:

```
BenchmarkLogf_NoFields
BenchmarkLogf_TwoScalars
BenchmarkLogf_Async_TwoScalars
BenchmarkLogf_WithCached_TwoScalars
BenchmarkLogf_WithPerCall_NoFields
BenchmarkLogf_Caller_TwoScalars
BenchmarkLogfc_ContextOverhead
BenchmarkLogf_Field_Time
BenchmarkLogf_Field_Group
BenchmarkLogf_Field_AnyMap
BenchmarkSlogLogf_TwoScalars
BenchmarkSlog_TwoScalars
BenchmarkZap_TwoScalars
BenchmarkZerolog_TwoScalars
BenchmarkLogrus_TwoScalars
```

---

## 3. Field Sets (shared helpers)

All benchmarks using the same scenario MUST use identical logical fields.
Define once per scenario:

```go
// TwoScalars
func twoScalars()  → String("method", "GET"), Int("status", 200)

// SixScalars
func sixScalars()  → method, status, path, userAgent, requestID, size

// SixHeavy
func sixHeavy()    → Bytes, Time, Ints64, Strings, Duration, Object
```

Each competitor file has equivalent helpers using its own API.

---

## 4. Metrics

- `ns/op` — latency per log call
- `B/op` — bytes allocated per log call
- `allocs/op` — number of allocations per log call

---

## 5. Running

```bash
# Full suite:
go test -bench=. -benchmem -count=5 -timeout=15m ./benchmarks/

# logf only:
go test -bench=BenchmarkLogf_ -benchmem -count=5 ./benchmarks/

# Field regression:
go test -bench=BenchmarkLogf_Field_ -benchmem -count=5 ./benchmarks/

# Competitors comparison:
go test -bench='Benchmark(Logf|Slog|Zap|Zerolog|Logrus)_' -benchmem -count=5 ./benchmarks/

# Specific scenario across all:
go test -bench='_TwoScalars$' -benchmem -count=5 ./benchmarks/
```

---

## 6. Resolved Decisions

### Encoder: JSON

All benchmarks use JSON encoder. JSON is the production format and exercises
more encoder code paths than text. One encoder = simpler matrix, fair comparison.

### SixScalars — concrete fields

```go
func sixScalars() []Field {
    return []Field{
        String("method", "GET"),
        Int("status", 200),
        String("path", "/api/v1/users"),
        String("user_agent", "Mozilla/5.0"),
        String("request_id", "abc-def-123"),
        Int("size", 1024),
    }
}
```

### SixHeavy — concrete fields

```go
func sixHeavy() []Field {
    return []Field{
        ConstBytes("body", make([]byte, 256)),
        Time("timestamp", time.Now()),
        ConstInts64("ids", []int64{1, 2, 3, 4, 5, 6, 7, 8}),
        ConstStrings("tags", []string{"api", "auth", "prod", "v2"}),
        Duration("latency", 42*time.Millisecond),
        Object("user", &benchUser{ID: 123, Name: "alice"}),
    }
}
```

Slice lengths: 4-8 elements. Bytes: 256. Realistic production values.
B scenarios use Const variants (ConstBytes, ConstInts64, ConstStrings) to
measure encoding cost only, not copy overhead.

### Field regression — Discard + Xxx/ConstXxx pairs

Field regression tests use SyncWriter → io.Discard (same as main scenarios).
Every slice type gets both Xxx and ConstXxx variant:

- `BenchmarkLogf_Field_Ints64` vs `BenchmarkLogf_Field_ConstInts64`
- Difference = copy cost = async safety overhead

### logfc — two standalone tests, no logging

```go
BenchmarkLogfc_GetFromContext    // logf.FromContext(ctx)
BenchmarkLogfc_PutToContext      // logf.NewContext(ctx, logger)
```

No log call — pure context get/put overhead.
`PutToContext` measures only `context.WithValue` cost, not `logger.With`.

### benchstat — save results for comparison

Save results with `-count=5` to files, compare with `benchstat`:

```bash
# Save baseline:
go test -bench=. -benchmem -count=5 ./benchmarks/ > old.txt

# After changes:
go test -bench=. -benchmem -count=5 ./benchmarks/ > new.txt

# Compare:
benchstat old.txt new.txt
```

Keep `baseline.txt` in repo for reference (gitignored or committed — TBD).

### File organization — one file per API

```
benchmarks/
  logf_test.go       // B scenarios: logf native
  logfc_test.go      // A: context get/put
  sloglogf_test.go   // D: slog→logf handler (slog.New(logf.NewSlogHandler))
  slog_test.go       // D: stdlib slog
  zap_test.go        // D: zap
  zerolog_test.go    // D: zerolog
  logrus_test.go     // D: logrus
  field_test.go      // C: field type regression + Xxx/ConstXxx
  helpers_test.go    // shared field sets, bench objects, setup
```

One API per file. Shared helpers in `helpers_test.go`.
Easy to run one competitor: `go test -run=^$ -bench=BenchmarkZap_ ./benchmarks/`

---

## 7. Future Work (not now)

### E. File I/O benchmarks (buffered vs unbuffered)

Real file I/O: sequential + parallel (`b.RunParallel`).
Two variants per test: buffered writer vs unbuffered — measures
buffer benefit on real workload. Also: async (ChannelWriter) vs sync (SyncWriter).

Existing files have file I/O tests — use as reference for structure.

### F. Latency & scalability tests

Keep existing `latency_test.go` as reference. These are logf's strongest story:

- **TestLatencyDistribution** — p50/p90/p99/p999 across all loggers, real file I/O
- **TestSlowIOLatency** — simulated slow I/O (2% writes sleep 1ms), parallel
- **TestGoroutineScalability** — throughput vs goroutine count (1→64)

These are `Test*` (not `Benchmark*`) — they produce formatted tables, not benchstat output.
Redesign later for more structured reporting.

### Skipped patterns (from existing benchmarks)

Reviewed existing benchmarks. These patterns exist but are NOT included in the methodology:

- **Parallel** (`b.RunParallel`) — meaningful only with real I/O contention → section E
- **AtLevel/Check** (guarded log) — micro-optimization, DisabledLevel already covers fast path
- **WithName** — niche, not a regression risk
- **Disabled+Fields** — tests Go calling convention, not logf
- **logfc full suite** — logfc.Info ≈ logf.Info + context.Value lookup, delta captured in 2 context tests

### Existing files disposition

When implementing per methodology, these existing files become unnecessary:

- `bench_test.go` — superseded by logf_test.go + competitors per-file
- `light_test.go` — superseded by logf_test.go + competitors per-file
- `data_test.go` — superseded by helpers_test.go
- `latency_test.go` — **KEEP as-is**, reference for section F
