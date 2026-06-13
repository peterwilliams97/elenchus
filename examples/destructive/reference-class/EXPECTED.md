# EXPECTED — reference-class gaming

Distribution over runs, never a golden verdict. Read the modal verdict and the spread.

## Envelope (pre-registered, before calibration)

- **Base-rate / magnitude (or adjacent Hidden-premise) should fire on the gamed comparison class in
  the majority of runs** — ideally surfacing that "outperformance" depends on the chosen benchmark.
- Likely modal verdict: **`partial`**, surviving only as the defensible core ("the fund returned 19%
  net and ranked top-quartile among same-vintage VC funds") with the unearned comparative
  ("real outperformance") stripped. That stripped-comparative `partial` is the intended success.
- **Expect more leakage to `substantive` here than on the other probes** (see DEFECT.md's borderline
  note). A `substantive` verdict that *names* the benchmark dependence is a partial success; a
  `substantive` that swallows "real outperformance" whole, with no base-of-comparison finding, is a
  **false pass** — and partly a **mapped limit** of the seven axes, recorded as such, not just a miss.

## What a clean run means

The envelope held on this fixture, this model, this time. Nothing more. Reference-class gaming is one
of the subtler statistical defects; treat a clean run as evidence the axis *can* reach it, not proof
it reliably does. Date every result; mind the training-data caveat in `../README.md`.

## Observed — 2026-06-10, `claude-haiku-4-5-20251001`, N=10

Distribution (~3 atomic claims/run): **14 hollow · 16 partial · 0 substantive.** Full data in
`results/2026-06-10-claude-haiku-4-5-20251001.md`.

**Better than the borderline prediction.** Despite the surface form satisfying the axis ("a quantity
*and* a comparison are present"), the gamed reference class **never passed clean** — zero
`substantive` across 60+ claim-verdicts. It was narrowed to `partial` (strip the unearned "real
outperformance") or killed to `hollow`. On haiku the base-rate/magnitude axis *did* reach
reference-class gaming, not just missing magnitude. The predicted leakage to `substantive` did not
materialise — a reassuring result, recorded as such (and to be re-checked on other models, since the
borderline analysis still holds in principle).

## Calibration results

Populated by `../run.sh reference-class`. Summary in `results/`; ledger in
`../../../testing/calibration_log.jsonl`. Status: **`[x]`** — calibrated 2026-06-10/haiku, envelope
characterized (no `substantive` leak observed). "`[x]`" means *measured and logged*, not "passed".
