#!/usr/bin/env python3
# Tally the -narrowing-boundary refuter. Usage: nb_tally.py <BASE>  where BASE has old/ and variant/
# subtrees:
#   BASE/<arm>/true/run-*/*.substance.jsonl
#   BASE/<arm>/false/run-*/*.substance.jsonl
#   BASE/<arm>/conditional/run-*/*.substance.jsonl
#   BASE/<arm>/motte/run-*/*.substance.jsonl          (honest motte + bailey, attributed by regex)
#   BASE/<arm>/unfalsifiable/run-*/*.substance.jsonl
# Pass criteria (variant arm): each well-formed fixture's CORE atom substantive >=4/5; bailey and
# unfalsifiable modal hollow. The honest motte is reported as an observation, not a gate.
import json, glob, re, sys, collections

BASE = sys.argv[1]
honest = re.compile(r"(system that learns|learns from feedback|adapts? its output).*intellig", re.I)
bailey = re.compile(r"genuinely intelligent|\bis intelligent\b", re.I)
rank = {"substantive": 2, "partial": 1, "hollow": 0, "error": -1}

# Each well-formed fixture decomposes into several atoms, but only ONE carries the proposition the
# control is about — the prediction in false.txt, the conditional in conditional.txt, the wage fact
# in true.txt. Scoring `best` across every atom lets a trivially-true definitional atom (e.g.
# "touchscreen-only smartphones are handsets with no physical keyboard", substantive) stand in for
# the fragment and mask a hollow core. So score the CORE atom named in ../well-formed/EXPECTED.md.
# `will account for` (not `accounting for`) excludes false.txt's separate causal atom; `below 2.1`
# selects only conditional.txt's antecedent-bearing atom; `7.25` selects true.txt's wage fact.
core = {
    "true":        re.compile(r"7\.25"),
    "false":       re.compile(r"will account for less than 10%", re.I),
    "conditional": re.compile(r"below 2\.1", re.I),
}


def run_verdicts(arm, probe):
    out = []
    for d in sorted(glob.glob(f"{BASE}/{arm}/{probe}/run-*")):
        recs = [json.loads(l) for f in glob.glob(f"{d}/*.substance.jsonl") for l in open(f)]
        out.append(recs)
    return out


def best(recs, pred):
    frs = [r for r in recs if pred(r["claim"])]
    if not frs:
        return None
    return max(frs, key=lambda r: rank.get(r["verdict"], -1))["verdict"]


def modal(recs, pred=lambda c: True):
    vs = [r["verdict"] for r in recs if pred(r["claim"])]
    return collections.Counter(vs).most_common(1)[0][0] if vs else None


for arm in ("old", "variant"):
    print(f"\n===== {arm} =====")
    # Well-formed positive controls: score the CORE atom of each fixture, not the best across atoms.
    for frag in ("true", "false", "conditional"):
        runs = run_verdicts(arm, frag)
        if not runs:
            print(f"well-formed/{frag:11s} NO RUNS FOUND")  # zero-output = failure, not a pass
            continue
        cv = [best(r, core[frag].search) for r in runs]
        if any(v is None for v in cv):  # a run whose core atom went unmatched is a failure, not a pass
            print(f"well-formed/{frag:11s} CORE ATOM UNMATCHED in {cv.count(None)} run(s) — fix the regex")
            continue
        subst = sum(1 for v in cv if v == "substantive")
        print(f"well-formed/{frag:11s} runs={len(runs)}  core-per-run={cv}  substantive={subst}/{len(runs)}"
              f"   PASS={'yes' if subst >= 4 else 'NO'}")
    # Honest motte: observation only (never a gate) — subst-or-partial per run.
    m = run_verdicts(arm, "motte")
    hv = [best(r, honest.search) for r in m]
    bvv = [best(r, bailey.search) for r in m]
    hp = sum(1 for v in hv if v in ("substantive", "partial"))
    print(f"honest motte  runs={len(m)}  per-run={hv}  subst-or-partial={hp}/{len(m)}  (observation, not a gate)")
    bmod = collections.Counter([v for v in bvv if v]).most_common()
    print(f"bailey        modal={bmod}  per-run={bvv}   PASS={'yes' if bmod and bmod[0][0]=='hollow' else 'NO'}")
    # Unfalsifiable-dress: per-run modal; passes when modal hollow and no substantive leak.
    u = run_verdicts(arm, "unfalsifiable")
    um = [modal(r) for r in u]
    uleak = sum(1 for r in u for x in r if x["verdict"] == "substantive")
    umod = collections.Counter([v for v in um if v]).most_common()
    print(f"unfalsifiable per-run-modal={um}  substantive-leak-fragments={uleak}"
          f"   PASS={'yes' if umod and umod[0][0]=='hollow' and uleak==0 else 'NO'}")
