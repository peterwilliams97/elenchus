#!/usr/bin/env python3
"""Step-3 2x2 comparison generator. Reads the four cell directories (each with a config.json, a
faithfulness chain JSONL, a usage.jsonl, and a run.stderr) and writes COMPARE.md (agreement table,
mutant catch rate, secs/$ per claim, quote-verification failures) plus adjudication.md (one row per
Sonnet/retrieved vs Qwen/retrieved divergence, both judges' quotes side by side, an empty verdict
column left for a human). Pure reader — no model, no fabrication; every number traces to a cell file.
"""
import json, os, re, sys

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.abspath(os.path.join(HERE, "..", "..", "..", ".."))
CELLS = ["sonnet-full", "sonnet-retrieved", "qwen-retrieved", "qwen-oracle"]
CLAIMS_FILE = os.path.join(ROOT, "examples/vic-lceic/claims-faith-refuter.txt")
CHAIN = "claims-faith-refuter.faithfulness.jsonl"
MUTANTS = {"M1", "M2", "M3"}          # F31, F17, F12 mutants; a "catch" = a clear rejection
CATCH = {"contradicted", "absent", "overstated"}


def claim_ids():
    ids = []
    for ln in open(CLAIMS_FILE):
        ln = ln.strip()
        if ln:
            ids.append(ln.split("\t")[0])
    return ids


def read_cell(name):
    d = os.path.join(HERE, name)
    chain = {}
    cp = os.path.join(d, CHAIN)
    if os.path.exists(cp):
        for ln in open(cp):
            r = json.loads(ln)
            chain[r["idx"]] = r
    usage = {}
    up = os.path.join(d, "usage.jsonl")
    if os.path.exists(up):
        for ln in open(up):
            usage = json.loads(ln)   # last record wins
    stderr = ""
    sp = os.path.join(d, "run.stderr")
    if os.path.exists(sp):
        stderr = open(sp, errors="replace").read()
    cfg = {}
    cp2 = os.path.join(d, "config.json")
    if os.path.exists(cp2):
        cfg = json.load(open(cp2))
    return {"name": name, "chain": chain, "usage": usage, "stderr": stderr, "cfg": cfg,
            "present": bool(chain)}


def grep_int(text, pat):
    m = re.search(pat, text)
    return int(m.group(1)) if m else None


def post_rule(verdict, nquotes):
    """The groundVerdict rule (assay.go) recomputed from a committed record: every verdict except
    absent needs >=1 verified quote. With none, contradicted->absent, faithful/partial/overstated->
    unsupported; absent and the non-verdicts (error/unsupported) are unchanged."""
    if nquotes >= 1:
        return verdict
    if verdict == "contradicted":
        return "absent"
    if verdict in ("faithful", "partial", "overstated"):
        return "unsupported"
    return verdict


def verdict_cell(cell, idx):
    r = cell["chain"].get(idx)
    if not r:
        return "—"
    v = r["verdict"]
    sp = r.get("spread")
    base = f"{v} {sp}" if sp else v
    post = post_rule(v, len(r.get("detail", {}).get("quotes", [])))
    return f"{base} → **{post}**" if post != v else base


def main():
    ids = claim_ids()
    cells = {c: read_cell(c) for c in CELLS}
    present = [c for c in CELLS if cells[c]["present"]]

    out = ["# Step-3 2×2 comparison — faithfulness judge across model × retrieval",
           "",
           f"Cells present: {', '.join(present) or '(none yet)'}. Same 8 claims "
           "(`claims-faith-refuter.txt`), N=3, one `config.json` frozen per cell. Numbers below trace "
           "to each cell's chain / usage / run.stderr; nothing is synthetic.",
           ""]

    # Agreement table: claim × cell. A `→ **post**` annotation is the verdict AFTER the groundVerdict
    # rule (every verdict but absent needs a verified quote), recomputed from the committed chains.
    flips = []
    for i, cid in enumerate(ids):
        for c in present:
            r = cells[c]["chain"].get(i)
            if r:
                p = post_rule(r["verdict"], len(r.get("detail", {}).get("quotes", [])))
                if p != r["verdict"]:
                    flips.append(f"{c} {cid}: {r['verdict']} → {p}")
    out += ["## Agreement table (verdict · N=3 spread)", "",
            "`→ **post**` marks a verdict changed by the code-side grounding rule (a verdict with no "
            "verified quote, recomputed from the committed chain — no rerun). "
            + (f"Post-rule flips: {', '.join(flips)}." if flips else "No post-rule flips."),
            "",
            "| claim | " + " | ".join(present) + " |",
            "|---|" + "|".join(["---"] * len(present)) + "|"]
    for i, cid in enumerate(ids):
        row = [cid] + [verdict_cell(cells[c], i) for c in present]
        out.append("| " + " | ".join(row) + " |")
    out.append("")

    # Per-cell rollups.
    out += ["## Per-cell metrics", "",
            "| metric | " + " | ".join(present) + " |",
            "|---|" + "|".join(["---"] * len(present)) + "|"]

    def rowfmt(label, fn):
        return "| " + label + " | " + " | ".join(fn(cells[c]) for c in present) + " |"

    def mutant_catch(cell):
        caught = tot = 0
        for i, cid in enumerate(ids):
            if cid in MUTANTS:
                tot += 1
                r = cell["chain"].get(i)
                if r and r["verdict"] in CATCH:
                    caught += 1
        return f"{caught}/{tot}"

    def per_claim_secs(cell):
        w = cell["usage"].get("wall_seconds")
        return f"{w/len(ids):.1f}s" if w else "—"

    def per_claim_usd(cell):
        e = cell["usage"].get("est_usd")
        if not e:
            return "$0"
        try:
            # est_usd is like "$1.8517 (rates 2026-06-01)" — take the number before the space.
            n = float(str(e).lstrip("$").split()[0])
            return f"${n/len(ids):.4f}"
        except (ValueError, IndexError):
            return str(e)

    def total_usd(cell):
        e = cell["usage"].get("est_usd")
        if not e:
            return "$0"
        try:
            return f"${float(str(e).lstrip('$').split()[0]):.4f}"
        except (ValueError, IndexError):
            return str(e)

    def quote_rejects(cell):
        n = grep_int(cell["stderr"], r"quote_rejects (\d+)")
        return str(n) if n is not None else "—"

    def cache_hit(cell):
        n = grep_int(cell["stderr"], r"cache_hit (\d+)%")
        return f"{n}%" if n is not None else "—"

    def schema_retries(cell):
        n = grep_int(cell["stderr"], r"schema_retries (\d+)")
        return str(n) if n is not None else "—"

    out.append(rowfmt("mutant catch rate (of 3)", mutant_catch))
    out.append(rowfmt("secs / claim", per_claim_secs))
    out.append(rowfmt("$ total (24 calls)", total_usd))
    out.append(rowfmt("$ / claim", per_claim_usd))
    out.append(rowfmt("quote-verification rejects", quote_rejects))
    out.append(rowfmt("cache hit", cache_hit))
    out.append(rowfmt("schema retries", schema_retries))
    out.append("")

    with open(os.path.join(HERE, "COMPARE.md"), "w") as f:
        f.write("\n".join(out))

    # Adjudication: Sonnet/retrieved vs Qwen/retrieved divergences.
    sr, qr = cells["sonnet-retrieved"], cells["qwen-retrieved"]
    adj = ["# Adjudication — Sonnet/retrieved vs Qwen/retrieved divergences", "",
           "One row per claim where the two retrieved-cell judges gave different verdicts. Both "
           "judges' verified quotes are shown side by side. The **who's right** column is left blank "
           "on purpose — it is for a human to fill.", ""]
    if not (sr["present"] and qr["present"]):
        adj.append("_(both retrieved cells not yet present)_")
    else:
        adj += ["| claim | Sonnet verdict | Sonnet quotes | Qwen verdict | Qwen quotes | who's right | post-rule |",
                "|---|---|---|---|---|---|---|"]
        ndiv = resolved = 0
        for i, cid in enumerate(ids):
            s, q = sr["chain"].get(i), qr["chain"].get(i)
            if not (s and q):
                continue
            if s["verdict"] != q["verdict"]:
                ndiv += 1
                sq = " / ".join(x[:90] for x in s.get("detail", {}).get("quotes", [])) or "(none)"
                qq = " / ".join(x[:90] for x in q.get("detail", {}).get("quotes", [])) or "(none)"
                sp = post_rule(s["verdict"], len(s.get("detail", {}).get("quotes", [])))
                qp = post_rule(q["verdict"], len(q.get("detail", {}).get("quotes", [])))
                note = f"resolved → both {sp}" if sp == qp else f"still differs ({sp} vs {qp})"
                if sp == qp:
                    resolved += 1
                adj.append(f"| {cid} | {s['verdict']} {s.get('spread','')} | {sq} | "
                           f"{q['verdict']} {q.get('spread','')} | {qq} |  | {note} |")
        adj.append("")
        adj.append(f"_{ndiv} divergence(s); {resolved} resolved by the grounding rule, "
                   f"{ndiv - resolved} remain for a human._")
    with open(os.path.join(HERE, "adjudication.md"), "w") as f:
        f.write("\n".join(adj))

    print(f"wrote COMPARE.md and adjudication.md ({len(present)}/4 cells present)")


if __name__ == "__main__":
    main()
