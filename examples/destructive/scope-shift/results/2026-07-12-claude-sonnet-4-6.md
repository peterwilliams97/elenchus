# Calibration — scope-shift

- date: 2026-07-12
- model: `claude-sonnet-4-6`
- runs: 5
- mode: substance
- **how run:** by hand (`./crossexam <dir>/claim.txt`, N=5), **not** via `../run.sh`. `run.sh`'s
  substance path hard-codes `-md`, which is a fatal stub on this build (substance mode only is wired,
  CLAUDE.md "Current build state"); it would swallow the error and log `parse-miss` for every run.
  Logged as a harness bug for a consolidation session, not side-fixed here.

## Verdict distribution (claims, summed across 5 runs)

`decompose` split the 3-sentence passage into **5–7 atomic claims per run** (5, 6, 5, 7, 6). Totals:

| Verdict | Count |
|---|---|
| partial | 17 |
| substantive | 5 |
| hollow | 7 |

All **5 `substantive`** are the arithmetic tautology fragment "1 minute 50 seconds is under 2 minutes"
(one per run) — a correct `substantive`, not a leak.

## Axis mentions (approximate — substring match in the Why column, runs out of 5)

| Axis keyword | Runs |
|---|---|
| equivocat | 5 |
| evidence | 5 |
| falsifiab | 5 |
| counterexample | 4 |
| magnitude | 2 |
| causal | 1 |
| base rate | 0 |
| correlation | 0 |
| hidden premise | 0 |

**Do not read this table as the target axis firing.** `equivocat` matched 5/5 — but on the *predicate*
("fast" = checkout speed vs. total store speed) or on "flagship" (designation vs. superlative), **never
on the subject widening**. This is the pre-registered reason the automated scan cannot stand in for the
discriminator (EXPECTED.md).

## Discriminator read (per-run, human — the point of the probe)

The question is not "did a verdict fall?" but "did any critique name the subject substitution —
measured at the flagship, asserted of the company?"

| Run | company-median claim | "company is fast" claim | flagship measurement (narrow core) | scope-shift **named**? |
|---|---|---|---|---|
| 01 | partial (unsupported/no data) | hollow (bare adjective, condition-laundering) | partial — survived | **no** |
| 02 | partial (collapses without data) | hollow (survives-only-by-conditioning) | partial — survived | **no** |
| 03 | partial (representativeness/no data) | hollow (bare, condition-laundering) | **hollow — over-catch** | **no** |
| 04 | partial (no data/format variance) | partial ("by the stated definition"; self-referential bar) | partial — survived | **no** |
| 05 | partial (sampling/no data) | hollow (bare, equivocates on 'fast') | partial — survived | **no** |

- **Scope-shift caught for the right reason: 0/5.** Every objection to a company claim was
  *unsupportedness* ("no data cited," "unanchored"), *bare-assertion*, or *condition-laundering* (the
  steelman inventing conditions) — all treating "the company" as a **fixed** subject with an evidence
  gap. None noticed the evidence was about a **different, narrower** subject (the flagship).

## Envelope check (against EXPECTED.md)

- **Headline held (5/5):** the company-wide claim did **not** survive `substantive` (partial or
  hollow every run).
- **Target catch failed (0/5):** the scope-shift was never named — for the **pre-registered W4
  reason**, not grader weakness (below).
- **False pass: 0/5.** No company claim was rated `substantive`.
- **Over-catch: 1/5.** Run 03 rated the honest narrow core (the flagship's 1:50 measurement) `hollow`
  — the symmetric failure named in EXPECTED.md. Not a clean pass even though the company claim also
  fell.

## The finding — a harness limitation, not a grader one

This is the outcome EXPECTED.md pre-registered as the most valuable to learn. `decompose` splits the
**flagship measurement** and the **company conclusion** into separate atomic claims every run, so no
single fragment holds both the narrow evidence and the broad claim. The critic never sees the widening
*as one move* — it grades each fragment in isolation, where the company claim reads as merely
unsupported and the flagship claim as fine. **The scope-shift, being a cross-sentence subject
substitution, is dissolved before the critic sees it.** Two extra symptoms of the same dissolution:

- `decompose` also **manufactured spurious fragments** — "The Oxford Street flagship is the company's
  flagship store" (runs 02, 05), "...is located on Oxford Street" (run 04) — propositions the passage
  never asserts as standalone claims. More cross-fragment noise from the same step.
- The `hollow` on "the company is fast" is driven by `survives_only_by_conditioning` (CLAUDE.md), i.e.
  the critic catching the *producer's* steelman inventing conditions — a different laundering than the
  scope-shift, reached without ever connecting the two subjects.

**Reading:** on this model, this build, this time — substance mode + the current `decompose` cannot
catch a multi-sentence scope-widening *as such*; it catches the downstream symptom (an unsupported
broad claim) instead. Before trusting substance mode on scope moves, know the catch is **incidental,
not diagnostic**. The clean follow-up is a **single-sentence** scope-shift (widening within one
sentence, so `decompose` can't separate the subjects) — that isolates grader behavior from harness
dissolution and answers whether the critic *itself* can name a subject substitution when it is forced
to see both subjects at once.

_Read against ../EXPECTED.md. Distribution, not a verdict. N=5, one model, by-hand — a first datum,
not a settled envelope. Re-run at N≥10 and on haiku (for baseline comparability) once run.sh's `-md`
dependency is fixed._
