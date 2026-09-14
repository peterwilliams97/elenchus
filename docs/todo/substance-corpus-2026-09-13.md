# Substance axis on the corpus pipeline — implementation + first run (2026-09-13)

Spec: `spec/SUBSTANCE-CORPUS.md`. Implements `-axis substance` for the `-source` corpus run: it runs
`assayClaim` (the dialectic/substance critic the `examples/destructive/` probes calibrate) per leaf,
writes `<chain-dir>/<fixture>.substance.jsonl` beside the faithfulness chain, and gates the argument
tree's one `holds` case on the substance verdict.

## What shipped

- **Flag `-axis faithfulness|substance` (comma-list)**, `assay.go`. Inert outside `-source`; the
  single-file modes select their axis directly.
- **Second chain in the corpus run** (`runSubstanceAxis`, `assay.go`): source-independent, one
  `assayClaim` per leaf, resumes from an existing substance chain like the faithfulness pass, fills
  `brief.Row.Substance`/`SubstanceReason` (fields that already existed). `appendChain` refactored to
  `appendChainTo(path, rec)` so both chains share one writer.
- **Render overlay** (`substanceSibling`, `attachSubstanceOverlay`, `assay.go`): `-from` loads a
  sibling `.substance.jsonl` when present, matched by idx + claim text; absent, the render is unchanged.
- **Rollup rule** (`leafJudgement`, `internal/tree/argument.go`): a settled-faithful leaf now `holds`
  only if substance is `substantive` **or absent** (back-compat); faithful + `hollow`/`partial` →
  `weakened`. An `opinion` leaf (`route=evaluative` / recommendation) is unchanged — never judged, so
  its substance verdict is shown but does not reclassify it.

## Refuters

**Deterministic (no model), `TestSubstanceGatesHold` in `internal/tree/argument_test.go` — GREEN.**
Pins the rollup: faithful+`""`→holds, faithful+substantive→holds, faithful+hollow→weakened,
faithful+partial→weakened, wobble→weakened, partial-faithfulness stays weakened (substance can't
rescue), evaluative+hollow stays opinion. `go test ./...` → all packages pass; `go build ./...` clean.

**Empirical (`assayClaim`, `claude-sonnet-4-6`, N=1, source-blind).** Run through the new
`-axis substance` path itself (`-retrieve none`, the faithfulness leg irrelevant; the substance chain
is what we read). Ran outside the sandbox — the in-sandbox TLS proxy fails Anthropic cert verification.

| leaf | expected | verdict | rounds | why (critic reason, abridged) |
|---|---|---|---|---|
| ai-index idx 49 — "2.7% in 2025, nearly double the 1.4%…" | non-substantive | **partial** ✓ | 1 | directional acceleration is falsifiable, but "2.7% in 2025" asserted as fact was a partial-year projection; strip the false precision → narrower defensible claim |
| ai-index idx 50 — Brynjolfsson "J-curve" framing | non-substantive | **hollow** ✓ | 1 | the content is an attribution to a 2026 publication that can't be verified to exist; Evidence + Falsifiability both fatal |
| dora idx 0 — AI "amplifier… magnifies strengths/dysfunctions" | non-substantive | **partial** ✓ | 2 | causal logic coherent + falsifiable, but no magnitude/base-rate survives and a key term risks circularity; narrows to a conditional claim |
| **control** — ai-index idx 37 — "Copilot → 26% more pull requests" | **substantive** | **partial** ✗ | 1 | critic flags an equivocation: 26% PR figure conflates a published 55% lab-task result with an undocumented observational metric; only a narrower lab-study claim survives |

**Outcome: defect-detection held 3/3 — every defect leaf came back non-substantive. The positive
control did NOT hold: nothing came back `substantive` at N=1.**

## Reading the control failure honestly

The refuter's discrimination test is only half-confirmed. Two readings of the control's downgrade,
both worth recording:

1. **A real catch.** The "26% more pull requests" line does conflate two distinct findings in the
   literature; a critic surfacing that is doing its job. If so, my control was not clean and the axis
   behaved correctly.
2. **A false attack — the more likely reading.** `assayClaim` is **source-blind**: it judged this from
   the model's own parametric knowledge of the Copilot studies, with no document in context. That is
   exactly the specialist-knowledge / false-attack risk the axis boundary (`CLAUDE.md`) and the
   `examples/destructive/bare-vs-contextualized/` probe warn about — Producer and Critic share one
   `c.model`, so an over-eager downgrade in one role is not checked by the other. At N=1, source-blind,
   sonnet awarded **no** `substantive` verdict to any of the four, which is itself a calibration signal:
   this critic, on bare decontextualized lines, is conservative and prone to inventing a defeater.

Either way it is a **mapped limit, not a silent pass**: the tool did not wave the defect leaves
through. But a fair positive-control test needs one of — a cleaner bare falsifiable claim, the
*contextualized* claim (its in-report argument attached, per the bare-vs-contextualized probe), or
N>1 to see whether `substantive` ever appears. That belongs with the dormant destructive-probe
revival (`docs/todo/audit-adversarial-2026-09-13.md` §c), which this run reinforces: the substance
critic's calibration on the current default model is untested, and this is the first data point.

## Notes / follow-ups

- The empirical run exercised the corpus `-axis substance` wiring end-to-end (both chains written).
  A full-corpus `-axis faithfulness,substance -tree` pass on ai-index/dora was **not** run — out of
  scope; only the four leaves were requested.
- Branch: edits were made on the pre-existing `goals/dora-2026-example-2026-09-13` branch (working
  tree already dirty from prior work; no commit made, per the request). A dedicated topic branch was
  not cut because switching a dirty tree mid-session is riskier than the deviation; flag for Peter.
- `critique_by_axis` in the chain uses capitalised keys (`Axis`/`Finding`/`Severity` — `critiqueItem`
  has no json tags). Harmless for the render, which reads the struct, but a reader `grep`-ing lowercase
  keys will miss them. Minor; not changed.
