#!/usr/bin/env sh
# rigour-map eval harness.
# Runs each fixture through the modes its domain supports and writes -md cross-tabs to eval/.
# Haiku first for cost (accounting's ~726 segments dominate the bill); re-run the default model
# only on the cells that turn out to matter.
#
# Usage:
#   rigour-map/run_eval.sh                                  # haiku
#   rigour-map/run_eval.sh claude-opus-4-8                  # stronger model, targeted re-run
set -eu

MODEL="${1:-claude-haiku-4-5-20251001}"
STAMP="$(date +%Y%m%d-%H%M)"
OUT="eval/${STAMP}-${MODEL}"
mkdir -p "$OUT"

USAGE_JSONL="eval/usage.jsonl"

# Count total steps: substance + grounding per fixture, plus audit per summary probe.
total_steps=0
for f in fixtures/raw/*.md; do
  [ -e "$f" ] || continue
  total_steps=$((total_steps + 2))
done
for s in fixtures/summaries/*_faithful.md fixtures/summaries/*_distorted.md; do
  [ -e "$s" ] || continue
  total_steps=$((total_steps + 1))
done

step=0
wall_start=$(date +%s)

# Single-document modes run on all seven (substance always; grounding where world-facts exist).
for f in fixtures/raw/*.md; do
  [ -e "$f" ] || continue
  b="$(basename "$f" .md)"

  step=$((step + 1))
  t0=$(date +%s)
  echo "[${step}/${total_steps}] $(date '+%H:%M:%S') substance  ${b}"
  ./assay -md -model "$MODEL" -usage-out "$USAGE_JSONL" "$f" > "$OUT/${b}.substance.md"
  echo "[${step}/${total_steps}] done  substance  ${b}  +$(($(date +%s) - t0))s"

  step=$((step + 1))
  t0=$(date +%s)
  echo "[${step}/${total_steps}] $(date '+%H:%M:%S') grounding  ${b}"
  ./assay -md -model "$MODEL" -usage-out "$USAGE_JSONL" -evidence "$f" > "$OUT/${b}.grounding.md"
  echo "[${step}/${total_steps}] done  grounding  ${b}  +$(($(date +%s) - t0))s"
done

# Faithfulness + audit only where a summary probe exists (the source is the real fixture;
# the summary is the probe). See fixtures/summaries/MANIFEST.md.
for s in fixtures/summaries/*_faithful.md fixtures/summaries/*_distorted.md; do
  [ -e "$s" ] || continue
  area="$(basename "$s" | sed 's/_.*//')"
  src="fixtures/raw/${area}.md"
  [ -e "$src" ] || { echo "!! no source for $s ($src) — skip"; continue; }
  name="$(basename "$s" .md)"

  step=$((step + 1))
  t0=$(date +%s)
  echo "[${step}/${total_steps}] $(date '+%H:%M:%S') audit      ${name}  (source: ${area})"
  ./assay -md -model "$MODEL" -usage-out "$USAGE_JSONL" -audit -source "$src" "$s" > "$OUT/${name}.audit.md"
  echo "[${step}/${total_steps}] done  audit      ${name}  +$(($(date +%s) - t0))s"
done

wall_end=$(date +%s)
wall_total=$((wall_end - wall_start))

echo ""
echo "wrote $OUT"
echo "total steps: ${step}/${total_steps}  wall: ${wall_total}s"
echo ""

# Sum eval/usage.jsonl into a grand total. Uses jq if available, else awk.
if [ -f "$USAGE_JSONL" ]; then
  echo "=== usage summary (from $USAGE_JSONL) ==="
  if command -v jq >/dev/null 2>&1; then
    priced_rows=$(jq -r 'select(.est_usd != null and (.est_usd | startswith("$"))) | .est_usd' \
      "$USAGE_JSONL" 2>/dev/null | wc -l | tr -d ' ')
    jq -rs '
      map(select(.est_usd != null and (.est_usd | startswith("$")))) as $priced |
      {
        calls:        (map(.calls) | add // 0),
        input_tokens: (map(.input_tokens) | add // 0),
        out_tokens:   (map(.output_tokens) | add // 0),
        web_searches: (map(.web_searches) | add // 0),
        est_usd:      ($priced | map(.est_usd | ltrimstr("$") | split(" ")[0] | tonumber) | add // 0),
        priced_rows:  ($priced | length),
        total_rows:   length
      } |
      "  calls=\(.calls) in=\(.input_tokens) out=\(.out_tokens) web=\(.web_searches) est_usd=$\(.est_usd | . * 10000 | round / 10000) (\(.priced_rows)/\(.total_rows) rows priced)"
    ' "$USAGE_JSONL"
    skipped=$(jq -r 'select(.est_usd == null or (.est_usd | startswith("$") | not)) | .model' \
      "$USAGE_JSONL" 2>/dev/null | wc -l | tr -d ' ')
    if [ "$skipped" -gt 0 ]; then
      echo "  (${skipped} rows skipped from total — est_usd null or n/a; token counts above include all rows)"
    fi
  else
    # awk fallback: sum calls, tokens, web_searches; skip null/n/a est_usd rows.
    awk '
      BEGIN { calls=0; in=0; out=0; web=0; usd=0; priced=0; skipped=0 }
      {
        match($0, /"calls":([0-9]+)/, a);        calls += a[1]+0
        match($0, /"input_tokens":([0-9]+)/, a); in    += a[1]+0
        match($0, /"output_tokens":([0-9]+)/, a); out  += a[1]+0
        match($0, /"web_searches":([0-9]+)/, a); web   += a[1]+0
        if (match($0, /"est_usd":"\$([0-9.]+)/, a)) { usd += a[1]+0; priced++ }
        else { skipped++ }
      }
      END {
        printf "  calls=%d in=%d out=%d web=%d est_usd=$%.4f (%d priced rows)\n",
               calls, in, out, web, usd, priced
        if (skipped > 0)
          printf "  (%d rows skipped — est_usd null or n/a; token counts above include all rows)\n", skipped
      }
    ' "$USAGE_JSONL"
  fi
fi

echo ""
echo "next: label each (area x mode) cell per rigour-map/RUBRIC.md, citing these files,"
echo "      then diff against the SEALED PRIORS."
