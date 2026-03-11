#!/usr/bin/env bash
#
# Run competitor-overlapping benchmarks and generate a markdown comparison table.
#
# Usage:
#   ./benchmarks/report.sh              # count=1 (quick)
#   ./benchmarks/report.sh 5            # count=5 (stable)
#
# Output: benchmarks/results/report-YYYYMMDD-HHMMSS.md
#
set -euo pipefail

export COUNT="${1:-1}"
export TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
DIR="$(cd "$(dirname "$0")" && pwd)"
RESULTS_DIR="$DIR/results"
mkdir -p "$RESULTS_DIR"

OUTFILE="$RESULTS_DIR/report-${TIMESTAMP}.md"
RAW="$RESULTS_DIR/raw-${TIMESTAMP}.txt"

# Scenarios that exist across competitors (B matrix).
# Pattern matches all B-scenario benchmarks for all APIs.
PATTERN='Benchmark(Logf|Slog|SlogLogf|Zap|Zerolog|Logrus)_(DisabledLevel|NoFields|TwoScalars$|TwoScalarsInGroup|SixScalars|SixHeavy|ErrorField|WithPerCall_NoFields|WithPerCall_TwoScalars|WithCached_NoFields|WithCached_TwoScalars|WithBoth_TwoScalars|WithGroupCached_TwoScalars|Caller_TwoScalars)'

echo "Running benchmarks (count=$COUNT)..."
(cd "$DIR" && go test -run='^$' -bench="$PATTERN" -benchmem -count="$COUNT" -timeout=15m .) | tee "$RAW"

echo ""
echo "Generating report → $OUTFILE"

# Parse raw output and build markdown table.
awk '
BEGIN {
    # Scenario display order
    split("DisabledLevel,NoFields,TwoScalars,TwoScalarsInGroup,SixScalars,SixHeavy,ErrorField,WithPerCall_NoFields,WithPerCall_TwoScalars,WithCached_NoFields,WithCached_TwoScalars,WithBoth_TwoScalars,WithGroupCached_TwoScalars,Caller_TwoScalars", order, ",")
    for (i in order) order_idx[order[i]] = i

    # API display order
    split("Logf,Slog,SlogLogf,Zap,Zerolog,Logrus", apis, ",")
    api_label["Logf"]    = "logf"
    api_label["Slog"]    = "slog"
    api_label["SlogLogf"]= "slog→logf"
    api_label["Zap"]     = "zap"
    api_label["Zerolog"]  = "zerolog"
    api_label["Logrus"]  = "logrus"
}

/^Benchmark/ {
    name = $1
    sub(/-[0-9]+$/, "", name)    # strip -10 suffix

    # Extract API and scenario from name: Benchmark{API}_{Scenario}
    sub(/^Benchmark/, "", name)
    # Match API prefix
    api = ""
    if      (name ~ /^SlogLogf_/) { api = "SlogLogf"; sub(/^SlogLogf_/, "", name) }
    else if (name ~ /^Logf_/)     { api = "Logf";     sub(/^Logf_/, "", name) }
    else if (name ~ /^Slog_/)     { api = "Slog";     sub(/^Slog_/, "", name) }
    else if (name ~ /^Zap_/)      { api = "Zap";      sub(/^Zap_/, "", name) }
    else if (name ~ /^Zerolog_/)  { api = "Zerolog";   sub(/^Zerolog_/, "", name) }
    else if (name ~ /^Logrus_/)   { api = "Logrus";    sub(/^Logrus_/, "", name) }
    else next

    scenario = name

    # Find ns/op and allocs/op
    ns = ""; allocs = ""
    for (i = 2; i <= NF; i++) {
        if ($(i+1) == "ns/op")     ns = $i + 0
        if ($(i+1) == "allocs/op") allocs = $i + 0
    }
    if (ns == "") next

    # Keep last result (if count>1, last iteration)
    key = api SUBSEP scenario
    data_ns[key] = ns
    data_allocs[key] = allocs
    scenarios[scenario] = 1
}

function fmt_ns(v) {
    if (v < 10) return sprintf("%.1f", v)
    return sprintf("%d", int(v + 0.5))
}

function fmt_scenario(s) {
    gsub(/_/, " ", s)
    return s
}

END {
    # Header
    printf "# Benchmark Report — %s\n\n", ENVIRON["TIMESTAMP"]
    printf "count=%s\n\n", ENVIRON["COUNT"]
    printf "| Scenario |"
    for (a = 1; a <= length(apis); a++) printf " %s |", api_label[apis[a]]
    printf "\n"
    printf "|---|"
    for (a = 1; a <= length(apis); a++) printf "---:|"
    printf "\n"

    # Rows in order
    for (s = 1; s <= length(order); s++) {
        sc = order[s]
        if (!(sc in scenarios)) continue
        printf "| %s |", fmt_scenario(sc)
        for (a = 1; a <= length(apis); a++) {
            key = apis[a] SUBSEP sc
            if (key in data_ns) {
                printf " %s ns (%d) |", fmt_ns(data_ns[key]), data_allocs[key]
            } else {
                printf " — |"
            }
        }
        printf "\n"
    }

    printf "\n— = not supported by this logger\n"
    printf "\nCell format: `ns/op (allocs/op)`\n"
}
' "$RAW" > "$OUTFILE"

echo ""
cat "$OUTFILE"
echo ""
echo "Saved: $OUTFILE"
echo "Raw:   $RAW"
