# TODO

## Done (this session)

- [x] Production Constructor → `NewLogger().Build()` builder API
- [x] slog contract compliance → passes `testing/slogtest.TestHandler`
- [x] WriterSlot for lazy destination initialization
- [x] SlabWriter message integrity (no torn writes)
- [x] SlabWriter performance (atomic counter regression fix, -18%)

## Backlog

### OTel FieldSource (medium)

Sub-package `logfotel` with `FieldSource` extracting trace_id/span_id
from OTel context. ContextHandler already supports FieldSource — this is
just a concrete implementation (~20 lines).

Do when there's real demand for OTel + logf integration.

### ReplaceAttr support (low)

Needs options struct on `NewSlogHandler`. Useful for redacting sensitive
fields, renaming keys, dropping attrs.

Add when users request it. Infrastructure exists (empty-attr check in
attrToField).

### Debug Ring Buffer (idea)

Post-mortem debugging: keep last N debug entries in memory, dump on
trigger. Interesting but complex (lifecycle, memory bounds, trigger
mechanism).

Revisit if post-mortem debugging becomes a real need.

### Sampling Handler (idea)

Per-level or per-key rate limiting as a Handler middleware. Drop every
Nth message at Debug/Info while passing all Warn/Error.

Advanced: `SamplingKey("http.request")` — a special field that is not
logged but provides a grouping key for sampling. Different call sites
with the same key share one counter. Gives the user explicit control
over what gets sampled together, unlike per-callsite (zap) or global
counter (zerolog) approaches.

### Stack trace field (low)

`Stack(key)` / `StackSkip(key, skip)` field constructor — capture current
goroutine stack trace. Useful for error-level logging.

## Decided against

- **SlogFromContext** — considered `logf.SlogFromContext(ctx) *slog.Logger`
  to get a slog logger from context directly. Not needed because
  `logfc.Get(ctx).Slog()` already works — one extra method call.

- **HTTP Field Middleware** — considered a built-in middleware that injects
  method, path, request_id into context. Every project has its own
  middleware conventions and field naming. Better as an example in README
  than a public API.

- **ErrAttr** — considered `logf.ErrAttr(err) slog.Attr` for slog
  compatibility. Not worth a public function — it's a one-liner:
  `slog.Any("error", err)`.

- **ForwardHandler** — considered an embeddable base struct for Handler
  middleware that auto-delegates Enabled. Not needed — Handler has only
  two methods, delegating Enabled is a single line.
