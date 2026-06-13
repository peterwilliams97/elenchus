# EXPECTED — causality from correlation

Distribution over runs, never a golden verdict.

## Envelope (pre-registered, before calibration)

- **Causality-vs-correlation (or adjacent Hidden-premise) should fire in the large majority of
  runs**, naming the selection effect / unestablished causal leap.
- Likely modal verdict: **`partial`** — surviving only as the honest correlational claim ("teams that
  adopted the tool shipped 38% faster, association not causation; confounds unaddressed"), with the
  causal mechanism claim stripped. That stripped-causation `partial` is the intended success.
- **False pass:** any run rating the claim **`substantive`** while accepting the causal mechanism, with
  **no** finding about correlation-vs-causation or the selection effect. Record every such run.

## What a clean run means

The envelope held on this fixture, this model, this time. Date the result; mind the training-data
caveat in `../README.md`.

## Observed — 2026-06-10, `claude-haiku-4-5-20251001`, N=10

Distribution (~4–5 atomic claims/run): **23 hollow · 19 partial · 3 substantive.** Full data in
`results/2026-06-10-claude-haiku-4-5-20251001.md`. Causality findings present every run; magnitude
9/10.

Caught or narrowed in the large majority — the causal mechanism was stripped, surviving at best as
the honest correlational `partial`. The "**3 `substantive` leaked**" reading here was
per-atomic-claim and **could not say which fragment leaked or whether any finding fired** — superseded
by the 2026-06-11 fragment-attributed run below.

## Observed — 2026-06-11, `claude-haiku-4-5-20251001`, N=10 (fragment-attributed)

Fresh calibration with per-fragment chains persisted (2026-06-10 chains discarded by an instrument
error — `../../../testing/SCHEMA.md`, `decision_log` d016 — a *new sample*, not a rescore).
Distribution (44 fragment-verdicts, `claims_per_run` 4–5): **18 hollow · 21 partial · 4 substantive ·
1 error.** `N_valid` = 43 (error excluded).

**Fragment attribution of the 4 `substantive` verdicts: 0 clean false passes.** Every one had the
**named adjacent axis — Hidden premise — fire** (a hidden-premise catch on "adoption is the cause,
not a marker" is the same defect from another angle, which DEFECT.md counts as the probe *working*).
Three are decomposed sub-mechanism fragments ("removing context-switching friction keeps engineers in
flow"); the one defect-carrying mechanism fragment ("the causal mechanism linking adoption to faster
shipping is the removal of friction") still had Hidden-premise fire. **No defect fragment passed
`substantive` with a silent critique. False-pass rate: 0/43** — the earlier "3 leaked" was the
unit-of-analysis artifact, since each `substantive` was either clean sub-mechanism scaffolding or
caught by the adjacent axis.

## Calibration results

Populated by `../run.sh causal-narrative`. Summary in `results/`; ledger in
`../../../testing/calibration_log.jsonl`. Status: **`[x]`** — calibrated 2026-06-10/haiku, then
fragment-attributed 2026-06-11/haiku (**0 clean false passes / 43**; all `substantive` either clean
scaffolding or adjacent-caught). "`[x]`" means *measured and logged*, not "passed".
