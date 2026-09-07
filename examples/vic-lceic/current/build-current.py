#!/usr/bin/env python3
"""Build examples/vic-lceic/current/ from the Sonnet evidence cells: for each claim, keep the record
from the newest run (by chain-file mtime), write a merged chain + a matching claims file + a
provenance line per claim (which run it came from). No model — pure merge over committed chains. The
merged chain is then rendered to tree.html with `assay -from`."""
import json, os, glob

HERE = os.path.dirname(os.path.abspath(__file__))
EVID = os.path.join(HERE, "..", "evidence")
CLAIMS_FILES = [os.path.join(HERE, "..", f) for f in ("claims-faith-refuter.txt", "claims-faith-scope.txt")]
MACHINE = os.path.join(HERE, "..", "claims-machine.txt")

# Root-block header carried at the top of claims.txt; assay reads "# title:" / "# date:" for line 1.
TITLE = "Inquiry into the cultural and creative industries in Victoria"
DATE = "June 2025"


def norm(s):
    return " ".join(s.split())


def load_routes():
    """claim-id → route from claims-machine.txt (the human file that declares route=)."""
    import re
    routes = {}
    for ln in open(MACHINE):
        if ln.startswith("#") or "|" not in ln:
            continue
        m = re.search(r"route=(\S+)", ln)
        if m:
            routes[ln.split("|")[0].strip()] = m.group(1)
    return routes


def load_claim_meta():
    """text → (id, path) from the claims files (first 3 tab fields)."""
    meta = {}
    for f in CLAIMS_FILES:
        for ln in open(f):
            parts = ln.rstrip("\n").split("\t")
            if len(parts) >= 3:
                meta[norm(parts[2])] = (parts[0], parts[1])
    return meta


def main():
    meta = load_claim_meta()
    routes = load_routes()
    # Sonnet chains only, tagged by source dir + mtime.
    chains = []
    for f in glob.glob(os.path.join(EVID, "*", "*", "*.faithfulness.jsonl")):
        cfg = os.path.join(os.path.dirname(f), "config.json")
        if not os.path.exists(cfg) or json.load(open(cfg)).get("backend") != "anthropic":
            continue
        rel = os.path.relpath(os.path.dirname(f), EVID)
        chains.append((os.path.getmtime(f), rel, f))
    chains.sort(reverse=True)  # newest first

    latest = {}   # claim-id → (record, source_dir)
    for mtime, rel, f in chains:
        for ln in open(f):
            r = json.loads(ln)
            key = norm(r["claim"])
            if key not in meta:
                continue
            cid, path = meta[key]
            if cid in latest:      # a newer run already claimed it
                continue
            latest[cid] = (r, path, rel)

    order = sorted(latest, key=lambda cid: (latest[cid][1], cid))  # by §path then id
    chain_out, claims_out, prov = [], [], []
    for i, cid in enumerate(order):
        r, path, rel = latest[cid]
        r = dict(r)
        r["idx"], r["total"], r["backend"] = i, len(order), "anthropic"
        r["route"] = routes.get(cid, "")  # propagate route into the chain so the root block can partition
        chain_out.append(json.dumps(r))
        claims_out.append(f"{cid}\t{path}\t{r['claim']}")
        prov.append(f"| {cid} | {r['verdict']} {r.get('spread','')} | {rel} |")

    open(os.path.join(HERE, "current.faithfulness.jsonl"), "w").write("\n".join(chain_out) + "\n")
    header = f"# title: {TITLE}\n# date: {DATE}\n"
    open(os.path.join(HERE, "claims.txt"), "w").write(header + "\n".join(claims_out) + "\n")
    with open(os.path.join(HERE, "PROVENANCE.md"), "w") as f:
        f.write("# current/ — provenance (which run each claim's verdict came from)\n\n")
        f.write("Merged from the Sonnet evidence cells, newest run per claim (by chain mtime). This is a\n")
        f.write("heterogeneous stopgap — claims come from runs with different corpora/judge versions;\n")
        f.write("`evidence/2026-09-08-full/` is the uniform replacement. `assay -from` renders `tree.html`.\n\n")
        f.write("| claim | verdict | source run |\n|---|---|---|\n" + "\n".join(prov) + "\n")
    print(f"merged {len(order)} claims from {len(chains)} Sonnet chains")


if __name__ == "__main__":
    main()
