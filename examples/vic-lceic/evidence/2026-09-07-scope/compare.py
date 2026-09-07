#!/usr/bin/env python3
"""Prompt-2 table: the judge prompt's scope change (faithful requires subject+scope+direction),
before vs after, per backend, on the hearings+submissions corpus. Scoring set excludes F8/F31.
Mutants (M1/M2/M3) must be contradicted/absent. The 5 new claims (F2a/F4/F10/F33/F54b) have no gold
— their verdicts + quotes become blank adjudication rows for a human. Pure reader over the chains."""
import json, os, re

HERE = os.path.dirname(os.path.abspath(__file__))
CHAIN = "claims-faith-scope.faithfulness.jsonl"
IDS = ["F12", "F17", "F29", "M1", "M2", "M3", "F2a", "F4", "F10", "F33", "F54b"]
MUTANTS = {"M1", "M2", "M3"}
NEW = {"F2a", "F4", "F10", "F33", "F54b"}   # no gold yet
CATCH = {"contradicted", "absent"}          # the mutant gold, per the spec
CELLS = ["before-sonnet", "after-sonnet", "before-qwen", "after-qwen"]


def load(cell):
    d = os.path.join(HERE, cell)
    chain, stderr = {}, ""
    cp = os.path.join(d, CHAIN)
    if os.path.exists(cp):
        for ln in open(cp):
            r = json.loads(ln)
            chain[r["idx"]] = r
    sp = os.path.join(d, "run.stderr")
    if os.path.exists(sp):
        stderr = open(sp, errors="replace").read()
    return {"chain": chain, "stderr": stderr, "present": bool(chain)}


def vs(cell, i):
    r = cell["chain"].get(i)
    return f"{r['verdict']} {r.get('spread','')}".strip() if r else "—"


def grepi(text, pat):
    m = re.search(pat, text)
    return m.group(1) if m else "—"


def main():
    cells = {c: load(c) for c in CELLS}
    present = [c for c in CELLS if cells[c]["present"]]

    out = ["# Prompt 2 — judge-prompt scope change (before vs after, per backend)", "",
           "faithful now requires the source to state the claim's **subject, scope and direction**; "
           "adjacent/broader statements are partial at best (gap=scope), with two worked negatives in "
           "the prompt. Hearings + submissions corpus, retrieved, N=3. Scoring set excludes F8/F31. "
           "Mutants (M1/M2/M3) gold = contradicted/absent; the 5 new claims (F2a/F4/F10/F33/F54b) have "
           "no gold — see adjudication.md.",
           "",
           "| claim | sonnet before | sonnet after | qwen before | qwen after |",
           "|---|---|---|---|---|"]
    for i, cid in enumerate(IDS):
        tag = " *(new)*" if cid in NEW else (" *(mutant)*" if cid in MUTANTS else "")
        row = [cid + tag, vs(cells["before-sonnet"], i), vs(cells["after-sonnet"], i),
               vs(cells["before-qwen"], i), vs(cells["after-qwen"], i)]
        out.append("| " + " | ".join(row) + " |")
    out.append("")

    # Per-cell metrics.
    out += ["## Per-cell metrics", "",
            "| metric | " + " | ".join(present) + " |",
            "|---|" + "|".join(["---"] * len(present)) + "|"]

    def mutant_catch(cell):
        c = t = 0
        for i, cid in enumerate(IDS):
            if cid in MUTANTS:
                t += 1
                r = cell["chain"].get(i)
                if r and r["verdict"] in CATCH:
                    c += 1
        return f"{c}/{t}"

    out.append("| mutants caught (of 3) | " + " | ".join(mutant_catch(cells[c]) for c in present) + " |")
    out.append("| no_quote_downgrades | " + " | ".join(grepi(cells[c]["stderr"], r"no_quote_downgrades (\d+)") for c in present) + " |")
    out.append("| quote_rejects | " + " | ".join(grepi(cells[c]["stderr"], r"quote_rejects (\d+)") for c in present) + " |")
    out.append("")
    with open(os.path.join(HERE, "COMPARE.md"), "w") as f:
        f.write("\n".join(out))

    # Adjudication: the 5 new claims (no gold) — after-run verdicts + quotes, who's-right blank.
    adj = ["# Adjudication — 5 new scope-risk claims (no gold yet)", "",
           "Verdicts and verified quotes from the **after** (scope-changed) judge on the combined "
           "corpus. These claims have no gold; the **who's right** column is blank for a human.", ""]
    have = [c for c in ("after-sonnet", "after-qwen") if cells[c]["present"]]
    if not have:
        adj.append("_(after runs not present yet)_")
    else:
        adj += ["| claim | backend | verdict | verified quotes | who's right |",
                "|---|---|---|---|---|"]
        for i, cid in enumerate(IDS):
            if cid not in NEW:
                continue
            for c in have:
                r = cells[c]["chain"].get(i)
                if not r:
                    continue
                qs = " / ".join(x[:90] for x in r.get("detail", {}).get("quotes", [])) or "(none)"
                adj.append(f"| {cid} | {c.replace('after-','')} | {r['verdict']} {r.get('spread','')} | {qs} |  |")
    with open(os.path.join(HERE, "adjudication.md"), "w") as f:
        f.write("\n".join(adj))
    print(f"wrote COMPARE.md + adjudication.md ({len(present)}/4 cells)")


if __name__ == "__main__":
    main()
