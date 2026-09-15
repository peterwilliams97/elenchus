#!/usr/bin/env python3
# Tally the edge-level adversarial pass (spec/EDGE.md §2–§5) from a saved chain — no model call, no
# commit. Usage: edge_tally.py <DIR>  where DIR is a testing/chains/edge-*/ directory holding one
# <base>.edge.jsonl per corpus (runEdgePass writes `<argument-base>.edge.jsonl`, so dora/master-plan
# land in separate files). Reads the chain, never re-judges. Layer-3 calibration a human reads, never
# a CI gate (TESTING.md): a green §5 means "the edge attacker discriminated on this run", never "the
# recommendation follows".
import json, glob, os, re, sys, collections

# Per corpus (spec/EDGE.md §5): the full roster of scheme-tagged F→R finding ids. runEdgePass judges an
# edge only when its finding has NOT leaf-derived to `fails` (§ Scope), so the chain holds exactly the
# in-scope edges and the excluded ones are the roster findings absent from the chain — each dropped
# because its finding leaf-derived to `fails`, the only reason a scheme-tagged edge leaves the pass. The
# tally reads the reached set from the chain and reports reached / (tagged − excluded), naming every
# excluded edge and why, so a legitimate scope-drop stays distinct from a truncated run (which the chain's
# `total` field catches: records < total ⇒ the pass declared more in-scope edges than reached the file).
# dora's F-PLAT drops this way (spec/EDGE.md Refuter run 1); every master-plan `example` finding drops,
# so only 4 of its 15 tagged edges are on held findings (spec/EDGE.md § master-plan run). A corpus that
# matches neither roster leaves partiality UNCHECKED.
EXPECT = {
    "dora": {"tagged": ["F-STANCE", "F-DATA", "F-ACCESS", "F-VC", "F-PLAT", "F-BATCH", "F-USER", "F-VSM"]},
    "master-plan": {"tagged": ["F1", "F2", "F3", "F4", "F5", "F6", "F7", "F8", "F9", "F10", "F11", "F12",
                               "F13", "F14", "F15"]},
}

# norm mirrors edge.norm (internal/edge/edge.go): lower-case, fold curly punctuation and dashes to
# ASCII, collapse whitespace. The template rule keys anchors through it, so the anchor listing (§3)
# that exists to expose the near-duplicates the exact-match key MISSED must normalise identically —
# a different fold would invent matches the code never made, or hide ones it did.
_PUNCT = {"’": "'", "‘": "'", "“": '"', "”": '"', "—": "-", "–": "-"}
_WS = re.compile(r"\s+")


def norm(s):
    s = "".join(_PUNCT.get(ch, ch) for ch in s.strip().lower())
    return _WS.sub(" ", s)


def trunc(s, n=80):
    s = " ".join(s.split())
    return s if len(s) <= n else s[: n - 1] + "…"


# detect_corpus reads the corpus from the chain itself, never from the *.edge.jsonl filename (which is
# the argument base, e.g. "argument", and names no corpus). It prefers the `corpus` field runEdgePass
# stamps on every edge record (the example dir, e.g. "dora-2026"), and falls back to the chain dir path
# the tally was pointed at. Returns a key of EXPECT, or None when neither names a known corpus.
def detect_corpus(recs, chain_dir):
    for r in recs:
        c = (r.get("detail") or {}).get("corpus", "") or ""
        for k in EXPECT:
            if k in c.lower():
                return k
    low = (chain_dir or "").lower()
    for k in EXPECT:
        if k in low:
            return k
    return None


def load(path):
    recs = []
    for line in open(path):
        line = line.strip()
        if not line:
            continue
        r = json.loads(line)
        if r.get("mode") == "edge":  # a mixed dir may hold other chains; take only edge records
            recs.append(r)
    return recs


# summarise flattens one corpus's records into per-edge dicts. `final` is the post-template verdict
# the top-level record carries; `modal_raw` is recomputed from `samples` so a method-lifted edge
# (samples mostly open, final unchallenged) shows both, rather than hiding the lift. Per edge the
# samples split into open (admitted), unchallenged (none-admitted OR a rejected offer) and error;
# none_admitted is therefore (successful − offered), never counted off a declined sample's placeholder
# fields — those never reach `offered` in runEdgePass, so `rejected` steps are offers by construction.
def summarise(recs):
    edges = []
    for r in recs:
        d = r.get("detail", {}) or {}
        s = r.get("samples", []) or []
        opens, unch, errs = s.count("open"), s.count("unchallenged"), s.count("error")
        offered, admitted = d.get("offered", 0), d.get("admitted", 0)
        edges.append({
            "total": r.get("total", 0),           # aggs len runEdgePass stamped: records < total ⇒ truncated
            "fid": d.get("finding_id", ""),
            "rid": d.get("rec_id", ""),
            "scheme": d.get("scheme", ""),
            "final": r.get("verdict", ""),
            "modal_raw": "open" if opens > unch else "unchallenged",
            "opens": opens, "unch": unch, "errs": errs, "n": len(s),
            "flips": min(opens, unch),               # samples on the minority side; >0 ⇒ contested
            "offered": offered, "admitted": admitted,
            "none_adm": (len(s) - errs) - offered,   # successful samples the model declined to attack
            "rejected": d.get("rejected", []) or [],
            "world": d.get("world", ""),
            "anchor": d.get("anchor", ""),
            "cq": d.get("critical_question", "—"),   # edgeDetail drops Defeater.CriticalQuestion; "—" until added
            "method": bool(d.get("method_level", False)),
        })
    return edges


def report(base, corpus, edges):
    N = len(edges)
    print(f"\n===== {base}  ({corpus or 'corpus unknown'}, {N} edges) =====")

    # 1. Per edge: finding → rec, modal + final verdict, the sample split, flips, and the modal
    #    admitted world with its anchor and critical_question (the last is "—" until edgeDetail carries it).
    print("-- 1. per edge --")
    for e in edges:
        lift = "  *method-lifted" if e["method"] else ""
        con = "  CONTESTED" if e["flips"] > 0 else ""
        print(f"{e['fid']:>8} → {e['rid']:<8} [{e['scheme']}]  modal={e['modal_raw']} final={e['final']}{lift}"
              f"  {e['opens']}o/{e['unch']}u/{e['errs']}e  flips={e['flips']}/{e['n']}{con}")
        if e["world"]:
            print(f"           world: {trunc(e['world'])}")
            print(f"           anchor: {e['anchor']!r}   cq: {e['cq']}")

    # 2. Totals. Rejections tallied by the admission step that failed (spec/EDGE.md §3 steps 1–3).
    tot_s = sum(e["n"] for e in edges)
    tot_na = sum(e["none_adm"] for e in edges)
    tot_off = sum(e["offered"] for e in edges)
    tot_adm = sum(e["admitted"] for e in edges)
    rej = collections.Counter(step for e in edges for step in e["rejected"])
    print("-- 2. totals --")
    print(f"samples={tot_s}  none_admitted={tot_na}  offered={tot_off}  admitted={tot_adm}  "
          f"rejected={sum(rej.values())}")
    print(f"rejected by step: {dict(rej.most_common()) or '{}'}")
    # Admitted defeaters by critical_question — which of the scheme's CQs is doing the opening
    # (spec/EDGE.md §2). The chain carries one modal admitted defeater per open edge, so this counts
    # open edges by their defeater's CQ; a lopsided tally names the CQ the pass leans on.
    by_cq = collections.Counter(e["cq"] for e in edges if e["world"])
    print(f"admitted by critical_question: {dict(by_cq.most_common()) or '{}'}")

    # 3. Template rule (spec/EDGE.md §3 rule 4): which edges were lifted to the root as method-level,
    #    then EVERY distinct admitted anchor across edges by exact-match (norm) key with its edge count,
    #    fired or not — so a human can read the anchors and spot near-duplicates the exact key missed.
    print("-- 3. template rule --")
    fired = [e for e in edges if e["method"]]
    print(f"fired: {'yes' if fired else 'no'}")
    if fired:
        by_anchor = collections.OrderedDict()
        for e in fired:
            g = by_anchor.setdefault(norm(e["anchor"]), {"world": e["world"], "fids": []})
            g["fids"].append(e["fid"])
        for _, g in by_anchor.items():
            print(f"  method: {trunc(g['world'])}")
            print(f"          edges: {g['fids']}")
    groups = collections.OrderedDict()
    for e in edges:
        if e["world"] and e["anchor"]:  # admitted (world populated pre-template), incl. lifted edges
            g = groups.setdefault(norm(e["anchor"]), {"raws": [], "fids": []})
            if e["anchor"] not in g["raws"]:
                g["raws"].append(e["anchor"])
            g["fids"].append(e["fid"])
    print("distinct admitted anchors (count = edges carrying it; count>1 = template-eligible):")
    if not groups:
        print("  (none admitted)")
    for _, g in sorted(groups.items(), key=lambda kv: -len(kv[1]["fids"])):
        raws = " | ".join(repr(r) for r in g["raws"])
        print(f"  {len(g['fids'])}×  {raws}   edges={g['fids']}")

    # 4. Positive control R-VC (dora, practical): predicted unchallenged (spec/EDGE.md §5(b)).
    print("-- 4. positive control --")
    rvc = next((e for e in edges if e["rid"] == "R-VC" or e["fid"] == "F-VC"), None)
    if rvc is None:
        print("R-VC: not in this corpus")
    else:
        w = f"   world: {trunc(rvc['world'])}" if rvc["final"] == "open" else ""
        print(f"R-VC: final={rvc['final']}{w}")

    # 5. Against spec/EDGE.md §5's pre-registered lines. Each prints PASS/FAIL; overall PASS only if
    #    all hold. The template-rule line is flagged "read the anchors" when it did not fire, pointing
    #    the human at §3's listing to judge whether a near-duplicate anchor should have merged.
    print("-- 5. spec/EDGE.md §5 --")
    spec = EXPECT.get(corpus)
    nopen = sum(1 for e in edges if e["final"] == "open")
    contested = sum(1 for e in edges if e["flips"] > 0)
    checks = []

    # Partial-run check against the IN-SCOPE denominator (tagged − excluded). Excluded = the roster
    # findings absent from the chain, each dropped because its finding leaf-derived to `fails`
    # (spec/EDGE.md § Scope) — the only reason a scheme-tagged edge leaves the pass. A genuinely
    # truncated run is caught separately by the chain's `total`: runEdgePass stamps every record with the
    # aggregated in-scope count, so records (N) < total means edges were judged but not written. Naming
    # the excluded edges keeps a legitimate scope-drop distinct from that truncation.
    if spec is None:
        print(f"  empty/partial: reached {N}/? — corpus unknown, partiality UNCHECKED")
        checks.append(N > 0)  # zero-output is still a fail even when the denominator is unknown
    else:
        tagged = spec["tagged"]
        reached = {e["fid"] for e in edges}
        excluded = [f for f in tagged if f not in reached]
        in_scope = len(tagged) - len(excluded)
        declared = max((e["total"] for e in edges), default=0)  # aggs len the pass stamped
        stray = reached - set(tagged)  # a finding_id not on the roster — chain/roster mismatch
        ok = N > 0 and N == declared and not stray and N == in_scope
        excl_desc = ", ".join(f"{f} (finding leaf-derives to fails, § Scope)" for f in excluded) or "none"
        print(f"  empty/partial: reached {N}/{in_scope} in-scope  "
              f"({len(tagged)} tagged − {len(excluded)} excluded: {excl_desc})  → {'PASS' if ok else 'FAIL'}")
        if N != declared:
            print(f"    TRUNCATED: chain declares {declared} in-scope edges but only {N} records reached the file")
        if stray:
            print(f"    OFF-ROSTER: chain carries finding ids not in the {corpus} roster: {sorted(stray)}")
        checks.append(ok)

    ok = not (N > 0 and nopen == N)
    print(f"  false-attack (all open): open={nopen}/{N}  → {'PASS' if ok else 'FAIL'}")
    checks.append(ok)

    # Template rule (spec/EDGE.md §3 rule 4) is now a BACKSTOP, not a §5 pass condition: with practical's
    # CQ set reduced to side_effects only (§2) the model no longer offers the correlation-as-cause
    # world per edge, so the rule is not expected to fire on dora and dora's method-level point rides the
    # fixed root note instead. Print it informationally — it fails no run.
    fired_yes = bool(fired)
    flag = "" if fired_yes else "  ← read the anchors (§3): a near-dup the exact-match key missed?"
    print(f"  template rule fired: {'yes' if fired_yes else 'no'}  (informational — backstop, not a §5 gate){flag}")

    # Fixed root method note (spec/EDGE.md §3 rule 4, §5): the pass emits it from code only when EVERY
    # scheme-tagged edge — not just the in-scope ones — is `practical` (allPractical(schemeEdges), assay.go).
    # The note ("findings are associational; recommendations are interventions") lives in root.RootMethods,
    # not the chain, so it is inferred here from the scheme mix. The chain carries only the in-scope
    # edges, so the tally sees the excluded edges' schemes only through `excluded` (spec/EDGE.md § Scope):
    # any excluded edge could be non-`practical` (a master-plan `example` leaf-failed out of the pass keeps
    # the report mixed-scheme), so an all-`practical` inference is trustworthy ONLY when nothing was
    # excluded. For dora (1 excluded, F-PLAT, itself practical) the note IS a §5 PASS line, so dora's
    # excluded set is affirmed all-practical below rather than left indeterminate.
    reached_practical = N > 0 and all(e["scheme"] == "practical" for e in edges)
    excluded_ct = len(spec["tagged"]) - N if spec else 0
    if corpus == "dora":
        # dora's one excluded edge (F-PLAT) is practical, so the reached-edge inference holds for the full roster.
        all_practical = reached_practical
        print(f"  fixed root method note (all edges practical): {'emitted' if all_practical else 'NOT emitted'}"
              f"  → {'PASS' if all_practical else 'FAIL'}")
        checks.append(all_practical)
    elif not reached_practical:
        print("  fixed root method note: not emitted (a reached edge is not practical)  (informational off dora)")
    elif excluded_ct == 0:
        print("  fixed root method note: emitted (all scheme-tagged edges practical)  (informational off dora)")
    else:
        # Reached edges are all practical, but excluded ones (leaf-failed) are absent from the chain: the
        # gate reads the full roster, so the note's status cannot be settled from the chain — master-plan's
        # 6 `example` edges are excluded, so the pass emits NO note (spec/EDGE.md § master-plan run).
        print(f"  fixed root method note: indeterminate from chain — {excluded_ct} scheme-tagged edge(s) "
              f"excluded (leaf-failed), whose scheme the chain does not carry; the pass emits the note only "
              f"if ALL are practical  (informational off dora)")

    if corpus == "dora":
        if rvc is None:
            print("  R-VC: MISSING from dora corpus  → FAIL")
            checks.append(False)
        elif rvc["final"] == "unchallenged":
            print("  R-VC unchallenged  → PASS")
            checks.append(True)
        else:
            print("  R-VC open  → human inspection needed (not auto-PASS): is its world genuinely concrete?")
            checks.append(False)

    rate = contested / N if N else 1.0
    ok = rate < 1 / 3
    print(f"  contested rate: {contested}/{N} = {rate:.0%}  (< 33%)  → {'PASS' if ok else 'FAIL'}")
    checks.append(ok)

    print(f"  OVERALL: {'PASS' if all(checks) else 'FAIL'}")


def main():
    if len(sys.argv) != 2:
        sys.exit("usage: edge_tally.py <testing/chains/edge-*/ dir>")
    d = sys.argv[1]
    files = sorted(glob.glob(os.path.join(d, "*.edge.jsonl")))
    if not files:
        sys.exit(f"NO *.edge.jsonl FOUND in {d} — zero-output = FAILURE, not a pass")
    empty = True
    for f in files:
        recs = load(f)
        edges = summarise(recs)
        if not edges:
            print(f"\n===== {os.path.basename(f)} — 0 edge records — zero-output = FAILURE =====")
            continue
        empty = False
        corpus = detect_corpus(recs, d)
        report(os.path.basename(f)[: -len(".edge.jsonl")], corpus, edges)
    if empty:
        sys.exit("all *.edge.jsonl held 0 edge records — zero-output = FAILURE")


if __name__ == "__main__":
    main()
