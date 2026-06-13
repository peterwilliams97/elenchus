#!/usr/bin/env bash
#
# run.sh — Layer 3 calibration runner for the destructive axis probes.
#
# This is NOT a test and NOT a CI gate. crossexam's judgment layer is non-deterministic
# (see ../../TESTING.md, "The central constraint"); a flaky pass/fail check on a verdict
# trains us to ignore failures. So this runs each probe N times, tallies the verdict
# DISTRIBUTION, and logs it for a human to read — never green/red. Do not wire it into
# build.sh or `go test`.
#
# Usage:
#   ./run.sh [PROBE|all] [N] [MODEL]
#
#   PROBE   a probe directory name (e.g. motte-and-bailey) or "all"   [default: all]
#   N       number of runs per probe                                  [default: 10]
#   MODEL   model id                                                  [default: claude-haiku-4-5-20251001]
#
# Examples:
#   ./run.sh                              # all probes, 10 runs each, haiku
#   ./run.sh laundering 10                # just the laundering audit, 10 runs
#   ./run.sh motte-and-bailey 5 claude-sonnet-4-6
#
# Per run it appends one JSON line to ../../testing/calibration_log.jsonl and writes a
# dated markdown summary into each probe's results/ directory.
#
# Portable to macOS's stock bash 3.2 (no associative arrays, no namerefs): counts are
# accumulated in temp files and summarized with `sort | uniq -c`.
#
# Requires: the crossexam binary at the repo root (run ./build.sh first) and ANTHROPIC_API_KEY.

set -euo pipefail

DESTDIR="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$DESTDIR/../.." && pwd)"
BIN="$ROOT/crossexam"
LOG="$ROOT/testing/calibration_log.jsonl"
# Persisted Tier-2 chains. The false-pass metric is defined per-DEFECT-fragment, so the
# per-fragment verdicts MUST survive the run — earlier versions wrote -chain-dir to $TMP and
# deleted it on EXIT, discarding exactly the data the headline metric needs (instrument error,
# decision_log d016). Chains now persist here, one dir per run, gitignored (raw crossexam output;
# the committed artifact is the attribution summary in each probe's results/).
CHAINROOT="$ROOT/testing/chains"

PROBE="${1:-all}"
N="${2:-10}"
MODEL="${3:-claude-haiku-4-5-20251001}"

# Substance-mode probes (run on claim.txt). The laundering probe is audit-mode, handled separately.
SUBSTANCE_PROBES="motte-and-bailey reference-class hidden-premise unfalsifiable-dress causal-narrative axis-gaps"

# Axis keywords we scan the critic's "Why" column for (approximate — substring match, lower-cased).
AXIS_KEYS="equivocat|hidden premise|falsifiab|base rate|magnitude|counterexample|causal|correlation|evidence"

[ -x "$BIN" ] || { echo "error: crossexam binary not found at $BIN — run ./build.sh first" >&2; exit 1; }
[ -n "${ANTHROPIC_API_KEY:-}" ] || { echo "error: set ANTHROPIC_API_KEY first" >&2; exit 1; }
mkdir -p "$ROOT/testing"

DATE="$(date +%Y-%m-%d)"
TMP="$(mktemp -d "${TMPDIR:-/tmp}/crossexam-calib.XXXXXX")"
trap 'rm -rf "$TMP"' EXIT

# summarize FILE  -> prints to stdout: first a markdown table body, then a "===" line, then a
# JSON object body. Empty file yields empty table and empty JSON. Parsed by the callers below.
summarize() {
  local file="$1" tbl="" jsn="" first=1 count val
  if [ -s "$file" ]; then
    while read -r count val; do
      [ -n "$val" ] || continue
      tbl="${tbl}| ${val} | ${count} |"$'\n'
      [ $first -eq 0 ] && jsn="${jsn},"
      jsn="${jsn}\"${val}\":${count}"
      first=0
    done < <(sort "$file" | uniq -c | sed 's/^ *//')
  fi
  printf '%s\n===\n%s\n' "$tbl" "$jsn"
}

# ── substance-mode probe ──────────────────────────────────────────────────────
run_substance() {
  local probe="$1" dir="$DESTDIR/$1"
  [ -f "$dir/claim.txt" ] || { echo "skip $probe: no claim.txt" >&2; return; }

  local vfile="$TMP/$probe.verdicts" afile="$TMP/$probe.axes"
  : > "$vfile"; : > "$afile"
  local r out tally low key
  r=1
  while [ "$r" -le "$N" ]; do
    local cdir="$CHAINROOT/$DATE-$MODEL/$probe/run-$r"; mkdir -p "$cdir"
    out="$("$BIN" -md -model "$MODEL" -chain-dir "$cdir" "$dir/claim.txt" 2>"$TMP/err" || true)"
    # Verdict distribution: parse the bold tally line, e.g. **2 hollow · 1 partial**
    tally="$(printf '%s\n' "$out" | grep -E '^\*\*[0-9]' | head -1 | sed 's/\*\*//g' || true)"
    if [ -n "$tally" ]; then
      # Split on the middle-dot separator; each token is "N verdict".
      printf '%s\n' "$tally" | tr '·' '\n' | while read -r tok; do
        tok="$(echo "$tok" | xargs)"
        [ -n "$tok" ] || continue
        local num="${tok%% *}" verd="${tok#* }"
        case "$num" in (''|*[!0-9]*) continue;; esac
        local i=0
        while [ "$i" -lt "$num" ]; do echo "$verd"; i=$((i+1)); done
      done >> "$vfile"
    else
      echo "parse-miss" >> "$vfile"
    fi
    # Axis mentions (approximate): one tick per run per keyword mentioned anywhere.
    low="$(printf '%s' "$out" | tr '[:upper:]' '[:lower:]')"
    printf '%s\n' "$AXIS_KEYS" | tr '|' '\n' | while read -r key; do
      case "$low" in (*"$key"*) echo "$key";; esac
    done >> "$afile"
    echo "  $probe run $r/$N done" >&2
    r=$((r+1))
  done

  emit_report "$probe" substance "$vfile" "$afile"
}

# ── audit-mode probe (laundering) ─────────────────────────────────────────────
run_audit() {
  local probe="laundering" dir="$DESTDIR/laundering"
  local src="$dir/sources/ballmer_usatoday_2007.txt"
  if [ ! -f "$src" ]; then
    echo "error: $src missing (gitignored). Recreate it from laundering/PROVENANCE.md before running." >&2
    return
  fi

  local ffile="$TMP/laundering.faith" sfile="$TMP/laundering.subst" gfile="$TMP/laundering.ground"
  : > "$ffile"; : > "$sfile"; : > "$gfile"
  local r out
  r=1
  while [ "$r" -le "$N" ]; do
    local cdir="$CHAINROOT/$DATE-$MODEL/laundering/run-$r"; mkdir -p "$cdir"
    out="$("$BIN" -md -audit -model "$MODEL" -chain-dir "$cdir" -source "$src" "$dir/summary.txt" 2>"$TMP/err" || true)"
    # Cross-tab data rows: | N | claim | faithful | substantive | grounded |
    # Keep only rows whose first cell is an integer (excludes the pattern-reading table).
    printf '%s\n' "$out" | while IFS='|' read -r _ num _claim faith subst ground _rest; do
      num="$(echo "$num" | xargs)"
      case "$num" in (''|*[!0-9]*) continue;; esac
      faith="$(echo "$faith"  | sed 's/\*\*//g' | xargs)"
      subst="$(echo "$subst"  | sed 's/\*\*//g' | xargs)"
      ground="$(echo "$ground"| sed 's/\*\*//g' | xargs)"
      [ -n "$faith"  ] && echo "$faith"  >> "$ffile"
      [ -n "$subst"  ] && echo "$subst"  >> "$sfile"
      [ -n "$ground" ] && echo "$ground" >> "$gfile"
    done
    echo "  laundering run $r/$N done" >&2
    r=$((r+1))
  done

  emit_audit_report "$ffile" "$sfile" "$gfile"
}

# split_summary "<summarize output>"  -> sets SUM_TBL and SUM_JSON
split_summary() {
  SUM_TBL="$(printf '%s' "$1" | sed -n '1,/^===$/p' | sed '$d')"
  SUM_JSON="$(printf '%s' "$1" | sed -n '/^===$/,$p' | sed '1d')"
}

# ── reporting ─────────────────────────────────────────────────────────────────
emit_report() {
  local probe="$1" mode="$2" vfile="$3" afile="$4"
  local resfile="$DESTDIR/$probe/results/${DATE}-${MODEL}.md"

  split_summary "$(summarize "$vfile")"; local vtbl="$SUM_TBL" vjson="$SUM_JSON"
  split_summary "$(summarize "$afile")"; local atbl="$SUM_TBL" ajson="$SUM_JSON"

  {
    echo "# Calibration — $probe"
    echo
    echo "- date: $DATE"
    echo "- model: \`$MODEL\`"
    echo "- runs: $N"
    echo "- mode: $mode"
    echo
    echo "## Verdict distribution (claims, summed across runs)"
    echo
    echo "| Verdict | Count |"
    echo "|---|---|"
    [ -n "$vtbl" ] && echo "$vtbl"
    echo
    echo "## Axis mentions (approximate — substring match in the Why column, runs out of $N)"
    echo
    echo "| Axis keyword | Runs |"
    echo "|---|---|"
    [ -n "$atbl" ] && echo "$atbl"
    echo
    echo "_Read against ../EXPECTED.md. This is a distribution, not a verdict; the envelope held"
    echo "or it did not, on this model, this date. Not a pass/fail gate._"
  } > "$resfile"
  echo "wrote $resfile" >&2

  printf '{"date":"%s","model":"%s","fixture":"%s","mode":"%s","runs":%d,"verdict_counts":{%s},"axis_mentions":{%s}}\n' \
    "$DATE" "$MODEL" "$probe" "$mode" "$N" "$vjson" "$ajson" >> "$LOG"

  echo "--- $probe ($mode, $N runs, $MODEL) ---"
  echo "verdicts:"; [ -n "$vtbl" ] && echo "$vtbl" | sed 's/^/  /'
}

emit_audit_report() {
  local ffile="$1" sfile="$2" gfile="$3"
  local probe="laundering" resfile="$DESTDIR/laundering/results/${DATE}-${MODEL}.md"

  split_summary "$(summarize "$ffile")"; local ftbl="$SUM_TBL" fjson="$SUM_JSON"
  split_summary "$(summarize "$sfile")"; local stbl="$SUM_TBL" sjson="$SUM_JSON"
  split_summary "$(summarize "$gfile")"; local gtbl="$SUM_TBL" gjson="$SUM_JSON"

  {
    echo "# Calibration — laundering (audit cross-tab)"
    echo
    echo "- date: $DATE"
    echo "- model: \`$MODEL\`"
    echo "- runs: $N"
    echo "- mode: audit"
    echo
    echo "## Faithfulness column"; echo; echo "| Verdict | Count |"; echo "|---|---|"; [ -n "$ftbl" ] && echo "$ftbl"; echo
    echo "## Substance column";    echo; echo "| Verdict | Count |"; echo "|---|---|"; [ -n "$stbl" ] && echo "$stbl"; echo
    echo "## Grounding column";    echo; echo "| Verdict | Count |"; echo "|---|---|"; [ -n "$gtbl" ] && echo "$gtbl"; echo
    echo "_Correct envelope: faithful + substantive/partial + **refuted** (laundering caught)."
    echo "Any non-refuted grounding while faithfulness+substance stay green = laundering false pass."
    echo "See ../EXPECTED.md. Distribution, not a gate._"
  } > "$resfile"
  echo "wrote $resfile" >&2

  printf '{"date":"%s","model":"%s","fixture":"%s","mode":"audit","runs":%d,"faithful_counts":{%s},"substantive_counts":{%s},"grounded_counts":{%s}}\n' \
    "$DATE" "$MODEL" "$probe" "$N" "$fjson" "$sjson" "$gjson" >> "$LOG"

  echo "--- laundering (audit, $N runs, $MODEL) ---"
  echo "faithful:";    [ -n "$ftbl" ] && echo "$ftbl" | sed 's/^/  /'
  echo "substantive:"; [ -n "$stbl" ] && echo "$stbl" | sed 's/^/  /'
  echo "grounded:";    [ -n "$gtbl" ] && echo "$gtbl" | sed 's/^/  /'
}

# ── reflexive grounding canary (TESTING.md 3d) ────────────────────────────────
# Always-on final step of a full calibration pass. Runs ./crossexam -evidence on the repo's own
# headline; REQUIRED result is `unverifiable` every run. Any `supported` is a self-sealing failure
# inside the instrument (grounding confirming the tool's own value proposition from the armchair) —
# the earliest warning that grounding has started pronouncing from the armchair. Cheap, never a gate.
run_canary() {
  local headline="$ROOT/examples/reflexive/headline.txt"
  [ -f "$headline" ] || { echo "skip canary: $headline missing" >&2; return; }
  local cn="${CANARY_N:-3}" r=1 vfile="$TMP/canary.verdicts"; : > "$vfile"
  echo "--- reflexive grounding canary (TESTING.md 3d), N=$cn ---" >&2
  while [ "$r" -le "$cn" ]; do
    local cdir="$CHAINROOT/$DATE-$MODEL/reflexive-canary/run-$r"; mkdir -p "$cdir"
    "$BIN" -md -quiet -evidence -model "$MODEL" -chain-dir "$cdir" "$headline" >/dev/null 2>"$TMP/err" || true
    local v
    v="$(grep -hoE '"verdict":"[a-z]+"' "$cdir"/*.grounding.jsonl 2>/dev/null | head -1 | sed 's/.*:"//;s/"//')"
    [ -n "$v" ] && echo "$v" >> "$vfile"
    echo "  canary run $r/$cn: ${v:-parse-miss}" >&2
    r=$((r+1))
  done
  local held=true
  if grep -q '^supported$' "$vfile"; then
    held=false
    echo "*** CANARY BREACH: grounding returned 'supported' on the tool's own headline — self-sealing" >&2
    echo "*** failure (TESTING.md 3d). Investigate before trusting any grounding verdict. ***" >&2
  else
    echo "canary held (required: unverifiable every run; not a gate): $(sort "$vfile" | uniq -c | sed 's/^ *//' | tr '\n' ' ')" >&2
  fi
  split_summary "$(summarize "$vfile")"; local cjson="$SUM_JSON"
  printf '{"date":"%s","model":"%s","fixture":"reflexive-canary","mode":"grounding","runs":%d,"verdict_counts":{%s},"canary":"headline","required":"unverifiable every run","held":%s}\n' \
    "$DATE" "$MODEL" "$cn" "$cjson" "$held" >> "$LOG"
}

# ── dispatch ──────────────────────────────────────────────────────────────────
echo "calibration: probe=$PROBE N=$N model=$MODEL" >&2
echo "log: $LOG" >&2

case "$PROBE" in
  all)
    for p in $SUBSTANCE_PROBES; do run_substance "$p"; done
    run_audit
    run_canary   # always-on final step of a full pass (TESTING.md 3d)
    ;;
  canary)
    run_canary
    ;;
  laundering)
    run_audit
    ;;
  *)
    found=0
    for p in $SUBSTANCE_PROBES; do [ "$p" = "$PROBE" ] && found=1; done
    [ "$found" -eq 1 ] || { echo "error: unknown probe '$PROBE' (use one of: $SUBSTANCE_PROBES laundering canary all)" >&2; exit 1; }
    run_substance "$PROBE"
    ;;
esac

echo "done. ledger: $LOG" >&2
