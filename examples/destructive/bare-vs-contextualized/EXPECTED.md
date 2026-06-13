# EXPECTED — bare vs contextualized

Distribution over runs, never a golden verdict. The unit of observation is the **difference between
the two inputs**, not either verdict alone.

## Pre-registered hypothesis (written before running)

The substance critic penalises a **bare, decontextualised** claim on the **Evidence** axis and pushes
it to `hollow`; supplying the claim's real in-text argument relieves that penalty.

**Decision rule (pre-registered):**

- **Penalty CONFIRMED** if the verdict distribution shifts **materially toward `substantive` /
  `partial`** under `contextual.txt` relative to `bare.txt` — i.e. `bare` is mostly `hollow`
  (replicating the laundering 10/10) while `contextual` produces a clear share of `substantive` /
  `partial`. The shift is the signal; the magnitude is the effect size.
- **Penalty NOT confirmed (or not the cause)** if both inputs land on the same modal verdict — either
  both `hollow` (the critic is harsh on the *prediction* regardless of context; the bare-line penalty
  is not the mechanism) or both `substantive`/`partial` (the bare line was not in fact being
  penalised for decontextualisation). Either null is itself envelope data: it relocates the cause.

Both readings are real outcomes. A null does not waste the probe — it rules the Evidence-axis-penalty
mechanism in or out.

## Why this matters if confirmed

Every substance input is a decontextualised atomic line by construction. A confirmed penalty means the
tool is **systematically** harsh on its own standard input shape — a false-attack bias that inflates
`hollow`. That is a critic-calibration finding, fixable in `substanceCriticSys` (separate package).

## What a clean reading means

Only: *on this fixture, this model, this time, context did / did not move the verdict.* This claim is
famous (the model may "know" the iPhone succeeded), so read the **substance** verdict — form, not
truth — not any leakage about the outcome. Date every result; mind the training-data caveat in
`../README.md`.

## Observed — 2026-06-11, `claude-haiku-4-5-20251001`, N=10 each

**`bare.txt`** (1 fragment/run): `hollow` **10/10** — replicates the laundering finding exactly.
**`contextual.txt`** (6 fragments/run): 28 `hollow` · 21 `partial` · 11 `substantive`. Full data in
`results/2026-06-11-claude-haiku-4-5-20251001.md`.

**Verdict on the pre-registered hypothesis: NOT confirmed** — and the same unit-of-analysis discipline
that this session applied to the 3b leaks (A1) is what overturns the naive reading here.

- At the **aggregate**, the distribution *does* shift toward `substantive`/`partial` under context
  (0% → 53%) — which the literal pre-registered rule would read as "penalty confirmed."
- But at the **claim level**, the **central prediction rates `hollow` 10/10 in *both* inputs**. The
  entire shift is carried by the *added* sub-claims (price, market size), which are independently more
  concrete. Context does **not** relieve the verdict on the claim under test — it just supplies extra
  gradeable material around it.
- The **fatal axis is Counterexample** (9/10 bare, 8/10 contextual), **not Evidence** (Evidence only
  *weakens*). A sweeping forward prediction always admits a constructible counterexample, and that is
  what kills it — regardless of in-text support.

So the false-attack on the bare forecast is **real but mis-diagnosed by the hypothesis**: its driver
is Counterexample-against-a-sweeping-prediction, not an Evidence/decontextualisation penalty. This is
**measurement-before-intervention paying off** — the proposed `substanceCriticSys` "evaluate form, not
in-text support" fix is **withdrawn**, because the claim *has* in-text support in `contextual.txt` and
still rates `hollow`. The real question (how substance should treat falsifiable forward predictions it
can always counterexample) is reframed in `../../../SESSION.md`. Status: **`[x]`** — measured and
logged; "`[x]`" means *measured*, not "passed".
