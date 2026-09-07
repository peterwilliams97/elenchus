#!/usr/bin/env python3
"""Prompt-1 table: the 5 real claims (F8 F12 F17 F29 F31) judged by Sonnet/retrieved/N=3 on the
hearings-only corpus (before) vs the hearings+submissions corpus (after), with the source type of
each verified quote. Pure reader over committed chains — no model, no fabrication."""
import json, os

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.abspath(os.path.join(HERE, "..", "..", "..", ".."))
IDS = ["F8", "F12", "F17", "F29", "F31"]
BEFORE = os.path.join(ROOT, "examples/vic-lceic/evidence/2026-09-07-compare/sonnet-retrieved/claims-faith-refuter.faithfulness.jsonl")
AFTER = os.path.join(HERE, "sonnet-retrieved/claims-faith-5.faithfulness.jsonl")


def load(path):
    out = {}
    if os.path.exists(path):
        for ln in open(path):
            r = json.loads(ln)
            out[r["idx"]] = r
    return out


def vs(r):
    return f"{r['verdict']} {r.get('spread','')}".strip() if r else "—"


def srcmix(r):
    if not r:
        return "—"
    srcs = r.get("detail", {}).get("quote_sources", [])
    if not srcs:
        return "(no verified quote)"
    n = {}
    for s in srcs:
        n[s] = n.get(s, 0) + 1
    return ", ".join(f"{k}×{v}" for k, v in sorted(n.items()))


def main():
    before, after = load(BEFORE), load(AFTER)
    usage = {}
    up = os.path.join(HERE, "sonnet-retrieved/usage.jsonl")
    if os.path.exists(up):
        usage = json.loads(open(up).read().strip().splitlines()[-1])

    out = ["# Prompt 1 — submissions rerun (5 real claims)", "",
           "Sonnet · retrieved · N=3 · current judge. **Before** = the committed 2026-09-07 2×2 "
           "`sonnet-retrieved` cell (hearings only). **After** = the same judge over the "
           "hearings **+ 42 written submissions** corpus. Source type is the origin of each verified "
           "quote. Chains only — nothing synthetic.",
           "",
           f"Spend: {usage.get('est_usd','—')} · {usage.get('calls','—')} calls · "
           f"{usage.get('wall_seconds',0):.0f}s.",
           "",
           "| claim | verdict before | verdict after | source of verified quotes (after) |",
           "|---|---|---|---|"]
    changed = 0
    for i, cid in enumerate(IDS):
        b, a = before.get(i), after.get(i)
        if b and a and b["verdict"] != a["verdict"]:
            changed += 1
        out.append(f"| {cid} | {vs(b)} | {vs(a)} | {srcmix(a)} |")
    out.append("")
    out.append(f"_{changed} of {len(IDS)} verdicts changed when submissions were added._")
    with open(os.path.join(HERE, "COMPARE.md"), "w") as f:
        f.write("\n".join(out))
    print("\n".join(out))


if __name__ == "__main__":
    main()
