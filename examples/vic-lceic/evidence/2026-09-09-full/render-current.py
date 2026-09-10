#!/usr/bin/env python3
"""Render examples/vic-lceic/current/ from THIS run's single uniform chain (all 69 findings, one
judge version). Supersedes ../../current/build-current.py, which merged heterogeneous partial cells;
here the whole tree comes from one run, so current/ is a straight re-order + re-key of this chain.

Pairs the chain (written in grouped processing order) back to claims-machine-full.txt's F-order,
carries route= into each record so the root block can partition, and writes the current/ inputs that
`assay -from` reads. The `assay -from` call itself (which emits tree.html / root-block-tree.txt /
audit.md) is the shell step after this script — see the run's REPORT/adjacent command."""
import json, os

HERE = os.path.dirname(os.path.abspath(__file__))
CUR = os.path.join(HERE, "..", "..", "current")
CHAIN = os.path.join(HERE, "claims-machine-full.faithfulness.jsonl")
FULL = os.path.join(HERE, "claims-machine-full.txt")

TITLE = "Inquiry into the cultural and creative industries in Victoria"
DATE = "June 2025"
RUN = "2026-09-09-full"


def norm(s):
    return " ".join(s.split())


def main():
    # F-order rows from the claims input: (id, path, text, cites, route).
    rows = []
    for ln in open(FULL):
        parts = ln.rstrip("\n").split("\t")
        if len(parts) >= 5 and parts[0].startswith("F"):
            rows.append(parts[:5])

    # chain records keyed by normalised claim text.
    by_text = {}
    for ln in open(CHAIN):
        r = json.loads(ln)
        by_text[norm(r["claim"])] = r
    if len(by_text) != len(rows):
        raise SystemExit(f"chain has {len(by_text)} claims, input has {len(rows)}")

    chain_out, claims_out = [], []
    for i, (cid, path, text, cites, route) in enumerate(rows):
        r = dict(by_text[norm(text)])
        r["idx"], r["total"], r["backend"], r["route"] = i, len(rows), "anthropic", route
        chain_out.append(json.dumps(r))
        claims_out.append(f"{cid}\t{path}\t{text}\t{cites}")

    open(os.path.join(CUR, "current.faithfulness.jsonl"), "w").write("\n".join(chain_out) + "\n")
    header = f"# title: {TITLE}\n# date: {DATE}\n"
    open(os.path.join(CUR, "claims.txt"), "w").write(header + "\n".join(claims_out) + "\n")
    with open(os.path.join(CUR, "PROVENANCE.md"), "w") as f:
        f.write("# current/ — provenance\n\n")
        f.write(f"Every finding's verdict comes from the one uniform run `evidence/{RUN}/` "
                "(Sonnet, retrieved, N=3,\nfull corpus + manifest). `assay -from` renders `tree.html` "
                "and `root-block-tree.txt` from\n`current.faithfulness.jsonl` paired with `claims.txt`.\n")
    print(f"rendered current/ from {len(rows)} findings ({RUN})")


if __name__ == "__main__":
    main()
