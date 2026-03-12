# Benchmarks — Methodology

## 1. Categories

### A. With micro-benchmarks (no log call)

Measure the cost of creating derived loggers. No `Info`/`Write` call — pure
logger construction overhead. Sequential (`b.N` loop), since `.With()` typically
happens once per request in middleware, not in a hot parallel path.

All loggers that support the operation get the same A-test.

| #  | Name       | What it measures                                      | Why it matters                                    |
|----|------------|-------------------------------------------------------|---------------------------------------------------|
| A1 | With       | `logger.With(twoScalars)` — derive with 2 fields      | Per-request middleware cost. logf: Bag alloc O(1). zap: encoder clone + pre-encode. |
| A2 | WithOnTop  | `.With()` on an already-derived logger                 | Stacking cost (nested middleware). Shows if second With is cheaper/same. |
| A3 | WithGroup  | `logger.WithGroup("name")` — derive with namespace    | Group/namespace creation cost. zerolog/logrus skip (no native WithGroup). |

Also in this category (logf-only, no competitors):

- **logfc GetFromContext** — `logf.FromContext(ctx)` — context.Value lookup cost
- **logfc PutToContext** — `logf.NewContext(ctx, logger)` — context.WithValue cost

### B. Log-call scenarios (Discard, parallel)

Full log call: field construction → encoding → write. Writer: `SyncWriter` →
`io.Discard` (full encode, no I/O noise).

**Execution mode:**
- **B0 (DisabledLevel)** — sequential (`b.N` loop). Measures level-check fast path
  only; parallelism adds no insight since the call returns immediately.
- **B1–B13** — parallel (`b.RunParallel`). Reflects production reality where
  multiple goroutines log concurrently. Exposes lock contention, pool scalability,
  and encoder thread-safety overhead.

| #  | Name                       | What it measures                                      | Why it matters                                    |
|----|----------------------------|-------------------------------------------------------|---------------------------------------------------|
| B0 | DisabledLevel              | `logger.Debug()` when level=Info                       | Fast-path cost. Should be ≤5 ns (atomic load).    |
| B1 | NoFields                   | `logger.Info("msg")` — no fields                       | Baseline: encoder acquire, timestamp, level, msg, write, release. |
| B2 | TwoScalars                 | `Info("msg", String, Int)`                             | Most common production case — 1-3 fields per log line. |
| B3 | TwoScalarsInGroup          | `Info("msg", Group("g", String, Int))`                 | Inline nested JSON object cost (logf.Group / zap.Dict / slog.Group / zerolog.Dict). |
| B4 | SixScalars                 | `Info("msg", 6× scalar)`                               | Heavier typical case — measures per-field encoding scaling. |
| B5 | SixHeavy                   | `Info("msg", Bytes+Time+Ints64+Strings+Duration+Object)` | Expensive types: base64, RFC3339, slice iteration, object marshaling. |
| B6 | ErrorField                 | `Info("msg", Error(err))`                              | Common pattern — error string encoding cost.      |
| B7 | WithPerCall+NoFields       | `logger.With(2s).Info("msg")` per iteration            | Per-request With + log. Simulates middleware adding request fields then logging. |
| B8 | WithPerCall+TwoScalars     | `logger.With(2s).Info("msg", 2s)` per iteration        | Same + per-call fields. Worst case: With + encode both sets. |
| B9 | WithCached+NoFields        | Pre-built `.With()` logger, then `Info("msg")`          | Amortized With — measures cached/pre-encoded context field benefit. |
| B10| WithCached+TwoScalars      | Pre-built `.With()` logger + `Info("msg", 2s)`          | Cached context + per-call fields. Key production pattern. |
| B11| WithBoth+TwoScalars        | Cached `.With()` + per-call `.With()` + `Info("msg", 2s)` | Full stack: service-level + request-level + call-level fields. |
| B12| WithGroupCached+TwoScalars | `.WithGroup().With()` cached + `Info("msg", 2s)`        | Namespaced context fields. Tests group encoding overhead. |
| B13| Caller+TwoScalars          | `AddCaller` + `Info("msg", 2s)`                         | runtime.Callers cost. Typically 100-200 ns extra.  |

### C. Field type regression (logf only, encode only)

One benchmark per unique field type — catches optimization regressions during refactoring.
Each test: single field of given type, full encode cycle via SyncWriter → io.Discard.

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

### D. Competitors (all categories, all loggers)

All loggers get the same A + B matrix using each competitor's idiomatic API.
Same logical field values for fair comparison.

**slog→logf (slogf)**: benchmarks create a logf Logger first (with the same config as
native logf tests — caller OFF, level Debug), then derive `slog.Logger` via
`Logger.Slog()`. This ensures the slog→logf bridge inherits the same settings
and measures only the bridge overhead, not accidental caller cost.

#### B matrix coverage

| #  | Scenario                   | logf | slog | slogf | zap | zerolog | logrus |
|----|----------------------------|------|------|-------|-----|---------|--------|
| B0 | DisabledLevel              |  ✓   |  ✓   |   ✓   |  ✓  |    ✓    |   ✓    |
| B1 | NoFields                   |  ✓   |  ✓   |   ✓   |  ✓  |    ✓    |   ✓    |
| B2 | TwoScalars                 |  ✓   |  ✓   |   ✓   |  ✓  |    ✓    |   ✓    |
| B3 | TwoScalarsInGroup          |  ✓   |  ✓   |   ✓   |  ✓  |    ✓    |   —    |
| B4 | SixScalars                 |  ✓   |  ✓   |   ✓   |  ✓  |    ✓    |   ✓    |
| B5 | SixHeavy                   |  ✓   |  ✓   |   ✓   |  ✓  |    ✓    |   ✓    |
| B6 | ErrorField                 |  ✓   |  ✓   |   ✓   |  ✓  |    ✓    |   ✓    |
| B7 | WithPerCall+NoFields       |  ✓   |  ✓   |   ✓   |  ✓  |    ✓    |   ✓    |
| B8 | WithPerCall+TwoScalars     |  ✓   |  ✓   |   ✓   |  ✓  |    ✓    |   ✓    |
| B9 | WithCached+NoFields        |  ✓   |  ✓   |   ✓   |  ✓  |    ✓    |   ✓    |
| B10| WithCached+TwoScalars      |  ✓   |  ✓   |   ✓   |  ✓  |    ✓    |   ✓    |
| B11| WithBoth+TwoScalars        |  ✓   |  ✓   |   ✓   |  ✓  |    ✓    |   ✓    |
| B12| WithGroupCached+TwoScalars |  ✓   |  ✓   |   ✓   |  ✓  |    ✓    |   —    |
| B13| Caller+TwoScalars          |  ✓   |  ✓   |   ✓   |  ✓  |    ✓    |   —    |

#### A matrix coverage

| #  | Scenario   | logf | slog | slogf | zap | zerolog | logrus |
|----|------------|------|------|-------|-----|---------|--------|
| A1 | With       |  ✓   |  ✓   |   ✓   |  ✓  |    ✓    |   ✓    |
| A2 | WithOnTop  |  ✓   |  ✓   |   ✓   |  ✓  |    ✓    |   ✓    |
| A3 | WithGroup  |  ✓   |  ✓   |   ✓   |  ✓  |    —    |   —    |

Notes:

- logrus: no native group support (B3, B12 skipped), no ReportCaller (B13 skipped), no WithGroup (A3 skipped)
- zerolog: groups via `Dict()` — closest equivalent; no WithGroup (A3 skipped)
- zap: groups via `zap.Namespace()` — closest equivalent for B12/A3
- slog: native `slog.Group()` and `WithGroup()`
- All loggers use io.Discard (or closest equivalent, no I/O noise)
- **If a competitor lacks a feature for a scenario — do NOT implement a local workaround. Mark as skipped.**

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

### Skipped patterns

Reviewed existing benchmarks. These patterns exist but are NOT included in the methodology:

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
