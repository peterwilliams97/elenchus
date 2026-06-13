# EXPECTED — hidden-premise chain

Distribution over runs, never a golden verdict.

## Envelope (pre-registered, before calibration)

- **Hidden-premise (or adjacent Counterexample) should fire in the large majority of runs**,
  surfacing the unstated "new-signup mix represents installed-base value" assumption.
- Likely modal verdict: **`partial`** — surviving only as the narrow, honest claim ("91% of *new
  signups* chose cloud"), with the decision ("therefore retire on-prem") stripped because it rests
  on the unstated premise. A `partial` that strips the conclusion is the intended success.
- **False pass:** any run rating the claim **`substantive`** with **no** finding naming the
  representativeness assumption or a counterexample to it. Record every such run.

## What a clean run means

The envelope held on this fixture, this model, this time. Date the result; mind the training-data
caveat in `../README.md`.

## Observed — 2026-06-10, `claude-haiku-4-5-20251001`, N=10

Distribution (~3–4 atomic claims/run): **17 hollow · 15 partial · 3 substantive · 1 error.** Full
data in `results/2026-06-10-claude-haiku-4-5-20251001.md`. Counterexample fired 6/10, evidence 8/10.

Caught or narrowed in the large majority — the decision ("retire on-prem") was stripped, leaving the
honest signup figure. The "**3 `substantive` leaked**" reading here was **per-atomic-claim and could
not say *which* fragment leaked** — superseded by the 2026-06-11 fragment-attributed run below.

## Observed — 2026-06-11, `claude-haiku-4-5-20251001`, N=10 (fragment-attributed)

Fresh calibration with per-fragment chains persisted (2026-06-10 chains discarded by an instrument
error — `../../../testing/SCHEMA.md`, `decision_log` d016 — so this is a *new sample*, not a rescore).
Distribution (37 fragment-verdicts, `claims_per_run` 3–4): **15 hollow · 13 partial · 7 substantive ·
1 `substantial` · 1 error.** `N_valid` = 36 (error excluded).

**Fragment attribution of the 8 `substantive`/`substantial` verdicts** (matching the motte pattern):
**6 land on the clean honest figure** ("91% of new signups chose cloud") — correct behaviour, the
narrow defensible core. **1 defect-carrying fragment** ("retire on-prem", run 6) rated `substantive`
but with Hidden-premise + Base-rate + Causality findings **firing** — the premise *was* named, so not
a clean leak. **1 clean false pass:** run 8's "customer choice indicates where engineering effort
should be allocated", rated `substantive` with an **empty critique**. **False-pass rate: 1/36**, not
3 — the earlier "3 leaked" conflated correct-on-the-clean-figure verdicts with the one real leak.

*Instrument note (W10):* run 3's verdict string came back `"substantial"` — a non-enum near-miss
rendered as-is, because no code path validates verdict strings against the enum (BACKGROUND.md W10).
It was on a clean fact; counted within the substantive total and flagged here.

## Calibration results

Populated by `../run.sh hidden-premise`. Summary in `results/`; ledger in
`../../../testing/calibration_log.jsonl`. Status: **`[x]`** — calibrated 2026-06-10/haiku, then
fragment-attributed 2026-06-11/haiku (1 clean false pass / 36; 6/8 `substantive` correct-on-figure;
1 W10 non-enum verdict). "`[x]`" means *measured and logged*, not "passed".
