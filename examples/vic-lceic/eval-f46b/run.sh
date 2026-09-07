#!/usr/bin/env bash
# Rerun F46b alone over the full corpus, to inspect the report_says/source_says leaf the new judge
# writes (replacing the free-form so_what). Needs ANTHROPIC_API_KEY; ~$0.10 for one Sonnet judge
# call over the cached corpus prefix. The leaf is printed at the end and written to the chain JSONL.
set -euo pipefail
cd "$(dirname "$0")/../../.."          # repo root
[ -x ./assay ] || go build -o assay .

SRC=examples/vic-lceic/sources
DIR=examples/vic-lceic/eval-f46b

# hearings-all.txt is an aggregate of the per-hearing files; set it aside so the corpus is exactly
# hearings/ + submissions/ + qon/, matching the full run's F46b passage IDs. It is gitignored, so the
# move is non-destructive and the trap restores it on exit.
ALL="$SRC/hearings-all.txt"; STASH=""
if [ -f "$ALL" ]; then STASH="${TMPDIR:-/tmp}/hearings-all.$$.txt"; mv "$ALL" "$STASH"; fi
restore() { [ -n "$STASH" ] && mv "$STASH" "$ALL" 2>/dev/null || true; }
trap restore EXIT

./assay -source "$SRC" -manifest "$SRC/MANIFEST.md" \
  -retrieve bm25 -embed -floor 0 -max-tokens 10000 -n 1 \
  -chain-dir "$DIR" -usage-out "$DIR/usage.jsonl" -quiet \
  "$DIR/claims-f46b.txt" 2> "$DIR/run.stderr"

echo "== F46b leaf =="
python3 -m json.tool "$DIR/claims-f46b.faithfulness.jsonl"
echo "== SUMMARY =="
grep -A6 SUMMARY "$DIR/run.stderr" || true
