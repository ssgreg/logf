#!/usr/bin/env bash
#
# Run file I/O benchmarks and latency tests, save results.
#
# Usage:
#   ./benchmarks/report-io.sh          # count=3
#   ./benchmarks/report-io.sh 5        # count=5
#
# Output: benchmarks/results/io-report-YYYYMMDD-HHMMSS.md
#
set -euo pipefail

COUNT="${1:-3}"
TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
DIR="$(cd "$(dirname "$0")" && pwd)"
RESULTS_DIR="$DIR/results"
mkdir -p "$RESULTS_DIR"

OUTFILE="$RESULTS_DIR/io-report-${TIMESTAMP}.md"

echo "=== File Parallel Benchmarks (count=$COUNT) ==="
echo ""
FP_RAW="$RESULTS_DIR/io-fileparallel-${TIMESTAMP}.txt"
(cd "$DIR" && go test -run='^$' -bench='BenchmarkFileParallel' -benchmem -count="$COUNT" -timeout=10m .) | tee "$FP_RAW"

echo ""
echo "=== Latency Distribution ==="
echo ""
LAT_RAW="$RESULTS_DIR/io-latency-${TIMESTAMP}.txt"
(cd "$DIR" && go test -run='TestLatencyDistribution' -v -timeout=5m .) 2>&1 | tee "$LAT_RAW"

echo ""
echo "=== Slow I/O Latency ==="
echo ""
SLOW_RAW="$RESULTS_DIR/io-slowio-${TIMESTAMP}.txt"
(cd "$DIR" && go test -run='TestSlowIOLatency' -v -timeout=5m .) 2>&1 | tee "$SLOW_RAW"

echo ""
echo "=== Goroutine Scalability ==="
echo ""
SCALE_RAW="$RESULTS_DIR/io-scalability-${TIMESTAMP}.txt"
(cd "$DIR" && go test -run='TestGoroutineScalability' -v -timeout=5m .) 2>&1 | tee "$SCALE_RAW"

# --- Generate report ---

cat > "$OUTFILE" <<HEADER
# I/O Benchmark Report — $TIMESTAMP

## File Parallel (count=$COUNT)

Parallel writes to a real file, 6 heavy fields per entry (bytes, time, ints, strings, duration, object).

HEADER

# Parse FileParallel results into a table.
awk '
/^Benchmark/ {
    name = $1
    sub(/-[0-9]+$/, "", name)
    sub(/^BenchmarkFileParallel_/, "", name)

    ns = ""
    bytes = ""
    allocs = ""
    drops = ""
    for (i = 2; i <= NF; i++) {
        if ($(i+1) == "ns/op")     ns = $i + 0
        if ($(i+1) == "B/op")      bytes = $i + 0
        if ($(i+1) == "allocs/op") allocs = $i + 0
        if ($(i+1) == "drops")     drops = $i + 0
    }
    if (ns == "") next

    # Keep last result per name
    data_ns[name] = ns
    data_bytes[name] = bytes
    data_allocs[name] = allocs
    data_drops[name] = drops
    if (!(name in order_idx)) {
        order_idx[name] = ++order_count
        order[order_count] = name
    }
}

function fmt_ns(v) {
    if (v < 10) return sprintf("%.1f", v)
    return sprintf("%d", int(v + 0.5))
}

END {
    printf "| Config | ns/op | B/op | allocs | drops |\n"
    printf "|---|---:|---:|---:|---:|\n"
    for (i = 1; i <= order_count; i++) {
        n = order[i]
        d = (data_drops[n] != "" && data_drops[n]+0 > 0) ? sprintf("%d", data_drops[n]+0) : "—"
        printf "| %s | %s | %d | %d | %s |\n", n, fmt_ns(data_ns[n]), data_bytes[n], data_allocs[n], d
    }
}
' "$FP_RAW" >> "$OUTFILE"

# Append latency test output.
cat >> "$OUTFILE" <<'SECTION'

## Latency Distribution

Parallel (NumCPU goroutines), 50K samples, real file I/O.

```
SECTION
grep 'p50=' "$LAT_RAW" >> "$OUTFILE"
echo '```' >> "$OUTFILE"

cat >> "$OUTFILE" <<'SECTION'

## Slow I/O Latency

Parallel, 5% of writes sleep 5ms to simulate I/O stalls.

```
SECTION
grep 'p50=' "$SLOW_RAW" >> "$OUTFILE"
echo '```' >> "$OUTFILE"

cat >> "$OUTFILE" <<'SECTION'

## Goroutine Scalability

Throughput (logs/sec) by goroutine count, 200K total ops, real file I/O.

```
SECTION
grep -E '(Logger|logf|zap|zerolog|slog)' "$SCALE_RAW" | grep -v '=== RUN' >> "$OUTFILE"
echo '```' >> "$OUTFILE"

echo ""
echo "=== Report ==="
echo ""
cat "$OUTFILE"
echo ""
echo "Saved: $OUTFILE"
