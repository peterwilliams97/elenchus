# spec/SUBSTANCE-CORPUS.md — a substance chain alongside the corpus faithfulness chain

> **Scope note — assay-era spec.** Paths (`assay.go`, `internal/tree`) are the assay-era homes; the
> behaviour travels unchanged if `assay` is renamed.

## Why

The report tree (`spec/TREE.md`) runs the **faithfulness** axis only. Faithfulness is orthogonal to
substance: a `faithful` leaf means "the report compresses its source without distortion," never "the
proposition survives the dialectic." `docs/todo/audit-adversarial-2026-09-13.md` §b shows three
`faithful` corpus leaves that carry an uncaught substance defect. This spec adds the **substance**
axis (`assayClaim`, `assay.go`) to the corpus pipeline so a leaf can be faithful *and* be marked where
its content does not hold up.

## The flag

`-axis <list>` — comma-list of axes to run in the `-source` (corpus faithfulness) run:
`faithfulness` (always run; the default), `substance`. `-axis faithfulness,substance` runs both. The
flag is inert outside the `-source` path (the single-file modes already select their axis directly).

## What a substance run produces

When `substance` is in `-axis`, the corpus run writes a **second** chain beside the faithfulness one:

    <chain-dir>/<fixture>.faithfulness.jsonl   (unchanged)
    <chain-dir>/<fixture>.substance.jsonl      (new)

The substance chain is byte-compatible with the existing single-file substance chain
(`substanceChainRecord`, `mode:"substance"`): one record per leaf, keyed by `idx`, verdict ∈
`{substantive, partial, hollow, error}`, `detail` = `substanceDetail` (steelman, per-axis critique,
surviving claim, rounds, reason). It reuses `assayClaim` unchanged — the same producer/critic loop the
`examples/destructive/` probes calibrate — so a corpus leaf is judged by the same critic as a probe.

**Source-independent.** `assayClaim` judges the claim *text* alone (the dialectic axis needs no
source), so the substance pass does not retrieve and makes no use of `-manifest`/`-source` passages.
It runs per leaf regardless of `route` — including `route=evaluative` leaves — because the audit's
point is that evaluative-and-faithful leaves are exactly where substance defects hide.

**Resume + N.** The substance chain resumes like the faithfulness one (`readChainSparse` on the
substance file, reuse unless `-fresh`). N-repeat (`-n`) is a faithfulness-judge feature and does not
apply; substance runs once per leaf.

## Leaf schema (substance.jsonl record)

    {"idx":N,"total":M,"mode":"substance","backend":"…","claim":"…",
     "verdict":"substantive|partial|hollow|error","elapsed_s":…,
     "detail":{"steelman":"…","critique_by_axis":[{"axis":…,"finding":…,"severity":…}],
               "surviving_claim":"…","rounds":N,"reason":"…"}}

## Rollup — how a faithful-but-hollow leaf propagates (`internal/tree/argument.go`)

`brief.Row` already carries `Substance`/`SubstanceReason`; the render path (`-from`) loads the sibling
`.substance.jsonl` (when present) and fills them by `idx`. `leafJudgement` changes in exactly one
place — the `faithful` case:

- `holds` now requires **both** faithful/settled **and** the substance axis clearing the leaf.
  Substance clears when it is `substantive` **or absent** (no substance chain — back-compat: a
  faithfulness-only run is unchanged).
- **faithful + substance `hollow` or `partial` → `weakened`.** The report copies its source, but the
  proposition does not survive the dialectic, so the node it supports stands on narrower ground. The
  substance verdict is shown on the leaf (`SubstanceReason`) so the reader sees *why*.

Everything else in `leafJudgement` is unchanged: a `partial`/`overstated`/`contradicted`/… faithfulness
verdict already dominates (it is at most `weakened`, at worst `fails`), so substance only ever
*lowers* a leaf that faithfulness left at `holds`, never rescues one faithfulness sank.

**Opinions are the one exception, by the existing rule.** A `route=evaluative` leaf (or a
recommendation) is `opinion` in the tree and contributes nothing to its parent (`leafJudgement`
returns `jOpinion` before the `faithful` case is reached). Its substance verdict is still recorded and
shown on the leaf for the reader, but it does **not** reclassify the leaf to `weakened` — an opinion is
not evidence, so it cannot weaken a parent it never supported. The rollup change therefore bites on
non-opinion faithful leaves (`route=source`/`evidence`/`data-gap`); on evaluative leaves the substance
verdict is informational. This is stated so the two evaluative refuter leaves below are understood
correctly: their substance verdict is the empirical refuter's target, not a tally change.

## Refuters

- **Deterministic rollup refuter** (`internal/tree/argument_test.go`, no model): four hand-built rows
  pin `leafJudgement` — `faithful`+`settled` with substance `""`→`holds` (back-compat),
  `substantive`→`holds` (the control shape), `hollow`→`weakened`, `partial`→`weakened`.
- **Empirical substance refuter** (`assayClaim` on `claude-sonnet-4-6`, four leaves from the audit):
  the three defect leaves must come back **non-substantive** and a control leaf **substantive** —
  - ai-index idx 49: *"…productivity growth reached 2.7% in 2025, nearly double the 1.4% average of
    the previous decade."* (reference-class / base-rate)
  - ai-index idx 50: *"Brynjolfsson (2026) frames the 2025 US productivity rise as the early stages of
    a 'J-curve'…"* (causal narrative)
  - dora idx 0: *"AI's primary role in software development is that of an amplifier…"* (equivocation /
    hidden premise)
  - **control** — ai-index idx 37: *"Software developers using GitHub Copilot completed 26% more pull
    requests."* (a bare, falsifiable, cited empirical fact → should stay `substantive`)

  Results in `docs/todo/substance-corpus-2026-09-13.md`. A defect leaf that comes back
  `substantive` is a mapped limit of the source-blind substance axis (a contextual defect the bare
  claim text does not carry), recorded — not silently passed.
