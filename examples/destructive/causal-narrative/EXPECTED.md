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
the honest correlational `partial`. **3 `substantive` leaked** across ~45 claim-verdicts: runs where
the selection-effect/causal leap was not surfaced. Modest leak; record and re-check on other models.

## Calibration results

Populated by `../run.sh causal-narrative`. Summary in `results/`; ledger in
`../../../testing/calibration_log.jsonl`. Status: **`[x]`** — calibrated 2026-06-10/haiku, envelope
characterized (small `substantive` leak noted). "`[x]`" means *measured and logged*, not "passed".
