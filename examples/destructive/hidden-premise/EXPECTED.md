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
honest signup figure. **3 `substantive` leaked** across ~35 claim-verdicts: runs where the
representativeness premise was not surfaced. Modest leak; record and re-check on other models.

## Calibration results

Populated by `../run.sh hidden-premise`. Summary in `results/`; ledger in
`../../../testing/calibration_log.jsonl`. Status: **`[x]`** — calibrated 2026-06-10/haiku, envelope
characterized (small `substantive` leak noted). "`[x]`" means *measured and logged*, not "passed".
