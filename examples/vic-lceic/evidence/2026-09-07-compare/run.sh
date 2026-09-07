#!/usr/bin/env bash
# Step-3 2x2 eval driver. Same 8 claims (claims-faith-refuter.txt), N=3, one config.json per cell.
# Cells: {sonnet,qwen} x {retrieved, ...} plus sonnet/full and qwen/oracle. Run one cell per arg, or
# "qwen" / "sonnet" / "all". Qwen cells cost $0 (local ollama); Sonnet cells bill the Anthropic API.
set -euo pipefail
cd "$(dirname "$0")/../../../.."   # repo root
BIN=./assay
CLAIMS=examples/vic-lceic/claims-faith-refuter.txt
CORPUS=examples/vic-lceic/sources/hearings
ORACLE=examples/vic-lceic/oracle-18-spans.json
OUT=examples/vic-lceic/evidence/2026-09-07-compare
SHA=b961595
N=3

cell() { # name backend model retrieve extra...
  local name=$1 backend=$2 model=$3 retrieve=$4; shift 4
  local dir="$OUT/$name"; mkdir -p "$dir"
  cat > "$dir/config.json" <<CFG
{"cell":"$name","backend":"$backend","model":"$model","retrieve":"$retrieve","n":$N,
 "max_tokens":10000,"floor":0,"embed":true,"claims":"$CLAIMS","corpus":"$CORPUS",
 "git_sha":"$SHA","date":"2026-09-07","extra":"$*"}
CFG
  echo ">>> cell $name ($backend/$model, retrieve=$retrieve, N=$N)"
  $BIN -backend "$backend" -model "$model" -retrieve "$retrieve" -n $N \
    -source "$CORPUS" -chain-dir "$dir" -usage-out "$dir/usage.jsonl" \
    -tree -quiet "$@" "$CLAIMS" > "$dir/tree.txt" 2> "$dir/run.stderr" || { echo "cell $name FAILED"; tail -5 "$dir/run.stderr"; return 1; }
  grep -A6 SUMMARY "$dir/run.stderr" || true
}

QWEN=qwen3.6:27b-q4_K_M
case "${1:-all}" in
  qwen-retrieved) cell qwen-retrieved ollama "$QWEN" bm25 ;;
  qwen-oracle)    cell qwen-oracle    ollama "$QWEN" oracle -oracle "$ORACLE" ;;
  sonnet-full)    cell sonnet-full    anthropic claude-sonnet-4-6 none ;;
  sonnet-retrieved) cell sonnet-retrieved anthropic claude-sonnet-4-6 bm25 ;;
  qwen)   "$0" qwen-retrieved; "$0" qwen-oracle ;;
  sonnet) "$0" sonnet-full; "$0" sonnet-retrieved ;;
  all)    "$0" qwen; "$0" sonnet ;;
esac
