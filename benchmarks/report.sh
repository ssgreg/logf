#!/usr/bin/env bash
#
# Run competitor-overlapping benchmarks and generate a markdown comparison table.
#
# Usage:
#   ./benchmarks/report.sh                         # all loggers, all scenarios, count=1
#   ./benchmarks/report.sh 5                       # all, count=5
#   ./benchmarks/report.sh 1 Logf                  # one logger
#   ./benchmarks/report.sh 1 Logf,Zap              # several loggers
#   ./benchmarks/report.sh 1 Logf NoFields         # one logger, one scenario
#   ./benchmarks/report.sh 1 Logf,Zap NoFields,SixHeavy  # mix
#   ./benchmarks/report.sh 1 '' NoFields           # all loggers, one scenario
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

# All known loggers and scenarios.
ALL_APIS="Logf|Slog|SlogLogf|Zap|Zerolog|Logrus"
ALL_SCENARIOS='DisabledLevel|NoFields|TwoScalars$|TwoScalarsInGroup|SixScalars|SixHeavy|ErrorField|WithPerCall_NoFields|WithPerCall_TwoScalars|WithCached_NoFields|WithCached_TwoScalars|WithBoth_TwoScalars|WithGroupCached_TwoScalars|Caller_TwoScalars|Parallel_NoFields|Parallel_TwoScalars|Parallel_WithCached_TwoScalars|Async_NoFields|Async_TwoScalars|Async_SixScalars|AsyncParallel_NoFields|AsyncParallel_TwoScalars|AsyncParallel_WithCached_TwoScalars'

# Optional filters: $2 = loggers (comma-sep), $3 = scenarios (comma-sep).
APIS="${2:-}"
SCENARIOS="${3:-}"

# Case-insensitive alias resolution (bash 3 compatible).
resolve_api() {
    case "$(echo "$1" | tr '[:upper:]' '[:lower:]')" in
        logf)     echo Logf ;;
        slog)     echo Slog ;;
        sloglogf|slogf) echo SlogLogf ;;
        zap)      echo Zap ;;
        zerolog)  echo Zerolog ;;
        logrus)   echo Logrus ;;
        *)        echo "$1" ;;
    esac
}

resolve_scenario() {
    case "$(echo "$1" | tr '[:upper:]' '[:lower:]')" in
        disabledlevel|disabled)            echo DisabledLevel ;;
        nofields|none)                      echo NoFields ;;
        twoscalars|2scalars)               echo TwoScalars ;;
        twoscalarsingroup|2scalarsingroup) echo TwoScalarsInGroup ;;
        sixscalars|6scalars)               echo SixScalars ;;
        sixheavy|6heavy)                   echo SixHeavy ;;
        errorfield|error)                   echo ErrorField ;;
        withpercall_nofields|percall)       echo WithPerCall_NoFields ;;
        withpercall_twoscalars)            echo WithPerCall_TwoScalars ;;
        withcached_nofields|cached)         echo WithCached_NoFields ;;
        withcached_twoscalars)             echo WithCached_TwoScalars ;;
        withboth_twoscalars|both)          echo WithBoth_TwoScalars ;;
        withgroupcached_twoscalars|groupcached) echo WithGroupCached_TwoScalars ;;
        caller_twoscalars|caller)          echo Caller_TwoScalars ;;
        parallel_nofields|pnofields)       echo Parallel_NoFields ;;
        parallel_twoscalars|p2scalars)     echo Parallel_TwoScalars ;;
        parallel_withcached_twoscalars|pcached) echo Parallel_WithCached_TwoScalars ;;
        async_nofields)                    echo Async_NoFields ;;
        async_twoscalars)                  echo Async_TwoScalars ;;
        async_sixscalars)                  echo Async_SixScalars ;;
        asyncparallel_nofields|apnofields) echo AsyncParallel_NoFields ;;
        asyncparallel_twoscalars|ap2scalars) echo AsyncParallel_TwoScalars ;;
        asyncparallel_withcached_twoscalars|apcached) echo AsyncParallel_WithCached_TwoScalars ;;
        *)                                 echo "$1" ;;
    esac
}

resolve_list() {
    local input="$1"
    local resolver="$2"
    local result=""
    local IFS=','
    for p in $input; do
        local resolved
        resolved="$($resolver "$p")"
        [ -n "$result" ] && result="$result|"
        result="$result$resolved"
    done
    echo "$result"
}

if [ -n "$APIS" ]; then
    APIS_RE="$(resolve_list "$APIS" resolve_api)"
else
    APIS_RE="$ALL_APIS"
fi

if [ -n "$SCENARIOS" ]; then
    SCENARIOS_RE="$(resolve_list "$SCENARIOS" resolve_scenario)"
else
    SCENARIOS_RE="$ALL_SCENARIOS"
fi

PATTERN="Benchmark(${APIS_RE})_(${SCENARIOS_RE})"

echo "Running benchmarks (count=$COUNT, pattern=$PATTERN)..."
(cd "$DIR" && go test -run='^$' -bench="$PATTERN" -benchmem -count="$COUNT" -timeout=15m .) | tee "$RAW"

echo ""
echo "Generating report → $OUTFILE"

# Parse raw output and build markdown table.
awk '
BEGIN {
    # ── Scenario display order and labels ──
    # key = benchmark suffix, value = table label
    # To rename a scenario in the table, change the value here.
    split("DisabledLevel,NoFields,TwoScalars,TwoScalarsInGroup,SixScalars,SixHeavy,ErrorField,WithPerCall_NoFields,WithPerCall_TwoScalars,WithCached_NoFields,WithCached_TwoScalars,WithBoth_TwoScalars,WithGroupCached_TwoScalars,Caller_TwoScalars,Parallel_NoFields,Parallel_TwoScalars,Parallel_WithCached_TwoScalars,Async_NoFields,Async_TwoScalars,Async_SixScalars,AsyncParallel_NoFields,AsyncParallel_TwoScalars,AsyncParallel_WithCached_TwoScalars", order, ",")
    for (i in order) order_idx[order[i]] = i

    sc_label["DisabledLevel"]              = "Disabled level (DisabledLevel)"
    sc_label["NoFields"]                   = "No fields (NoFields)"
    sc_label["TwoScalars"]                 = "2 scalars (TwoScalars)"
    sc_label["TwoScalarsInGroup"]          = "2 scalars in group (TwoScalarsInGroup)"
    sc_label["SixScalars"]                 = "6 scalars (SixScalars)"
    sc_label["SixHeavy"]                   = "6 heavy (SixHeavy)"
    sc_label["ErrorField"]                 = "Error field (ErrorField)"
    sc_label["WithPerCall_NoFields"]       = "With/call (WithPerCall_NoFields)"
    sc_label["WithPerCall_TwoScalars"]     = "With/call + 2s (WithPerCall_TwoScalars)"
    sc_label["WithCached_NoFields"]        = "With/cached (WithCached_NoFields)"
    sc_label["WithCached_TwoScalars"]      = "With/cached + 2s (WithCached_TwoScalars)"
    sc_label["WithBoth_TwoScalars"]        = "With/both + 2s (WithBoth_TwoScalars)"
    sc_label["WithGroupCached_TwoScalars"] = "WithGroup + 2s (WithGroupCached_TwoScalars)"
    sc_label["Caller_TwoScalars"]          = "Caller + 2s (Caller_TwoScalars)"

    sc_label["Parallel_NoFields"]                      = "‖ No fields (Parallel_NoFields)"
    sc_label["Parallel_TwoScalars"]                    = "‖ 2 scalars (Parallel_TwoScalars)"
    sc_label["Parallel_WithCached_TwoScalars"]         = "‖ With/cached + 2s (Parallel_WithCached_TwoScalars)"
    sc_label["Async_NoFields"]                         = "⤳ No fields (Async_NoFields)"
    sc_label["Async_TwoScalars"]                       = "⤳ 2 scalars (Async_TwoScalars)"
    sc_label["Async_SixScalars"]                       = "⤳ 6 scalars (Async_SixScalars)"
    sc_label["AsyncParallel_NoFields"]                 = "⤳‖ No fields (AsyncParallel_NoFields)"
    sc_label["AsyncParallel_TwoScalars"]               = "⤳‖ 2 scalars (AsyncParallel_TwoScalars)"
    sc_label["AsyncParallel_WithCached_TwoScalars"]    = "⤳‖ With/cached + 2s (AsyncParallel_WithCached_TwoScalars)"

    # ── Logger display order and labels ──
    # key = benchmark prefix, value = table column header
    split("Logf,Slog,SlogLogf,Zap,Zerolog,Logrus", apis, ",")
    api_label["Logf"]     = "logf"
    api_label["Slog"]     = "slog"
    api_label["SlogLogf"] = "slogf"
    api_label["Zap"]      = "zap"
    api_label["Zerolog"]  = "zerolog"
    api_label["Logrus"]   = "logrus"
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

    # Find ns/op, B/op and allocs/op
    ns = ""; bytes = ""; allocs = ""
    for (i = 2; i <= NF; i++) {
        if ($(i+1) == "ns/op")     ns = $i + 0
        if ($(i+1) == "B/op")      bytes = $i + 0
        if ($(i+1) == "allocs/op") allocs = $i + 0
    }
    if (ns == "") next

    # Keep last result (if count>1, last iteration)
    key = api SUBSEP scenario
    data_ns[key] = ns
    data_bytes[key] = bytes
    data_allocs[key] = allocs
    scenarios[scenario] = 1
    api_seen[api] = 1
}

function fmt_ns(v) {
    if (v < 10) return sprintf("%.1f", v)
    return sprintf("%d", int(v + 0.5))
}

function fmt_scenario(s) {
    if (s in sc_label) return sc_label[s]
    gsub(/_/, " ", s)
    return s
}

END {
    printf "# Benchmark Report — %s\n\n", ENVIRON["TIMESTAMP"]
    printf "count=%s\n\n", ENVIRON["COUNT"]

    # ── Table 1: ns/op ──
    printf "## Latency (ns/op)\n\n"
    printf "| Scenario |"
    for (a = 1; a <= length(apis); a++) printf " %s |", api_label[apis[a]]
    printf "\n|---|"
    for (a = 1; a <= length(apis); a++) printf "---:|"
    printf "\n"
    for (s = 1; s <= length(order); s++) {
        sc = order[s]
        if (!(sc in scenarios)) continue
        printf "| %s |", fmt_scenario(sc)
        for (a = 1; a <= length(apis); a++) {
            key = apis[a] SUBSEP sc
            if (key in data_ns) printf " %s |", fmt_ns(data_ns[key])
            else if (apis[a] in api_seen) printf " — |"
            else                printf " · |"
        }
        printf "\n"
    }

    # ── Table 2: allocations (B/op / allocs/op) ──
    printf "\n## Allocations (B/op / allocs)\n\n"
    printf "| Scenario |"
    for (a = 1; a <= length(apis); a++) printf " %s |", api_label[apis[a]]
    printf "\n|---|"
    for (a = 1; a <= length(apis); a++) printf "---:|"
    printf "\n"
    for (s = 1; s <= length(order); s++) {
        sc = order[s]
        if (!(sc in scenarios)) continue
        printf "| %s |", fmt_scenario(sc)
        for (a = 1; a <= length(apis); a++) {
            key = apis[a] SUBSEP sc
            if (key in data_ns) {
                b = data_bytes[key] + 0
                al = data_allocs[key] + 0
                if (b == 0 && al == 0) printf " 0 |"
                else                   printf " %d / %d |", b, al
            } else if (apis[a] in api_seen) {
                printf " — |"
            } else {
                printf " · |"
            }
        }
        printf "\n"
    }

    printf "\n— = not supported\n"
    printf "· = not tested\n"
}
' "$RAW" > "$OUTFILE"

echo ""
cat "$OUTFILE"
echo ""
echo "Saved: $OUTFILE"
echo "Raw:   $RAW"
