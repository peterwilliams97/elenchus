# Calibration — laundering (audit cross-tab)

- date: 2026-09-13
- model: `claude-sonnet-4-6`
- runs: 10
- mode: audit

## Faithfulness column

| Verdict | Count |
|---|---|
| faithful | 10 |

## Substance column

| Verdict | Count |
|---|---|
| hollow | 7 |
| partial | 3 |

## Grounding column

| Verdict | Count |
|---|---|
| refuted | 10 |

_Correct envelope: faithful + substantive/partial + **refuted** (laundering caught).
Any non-refuted grounding while faithfulness+substance stay green = laundering false pass.
See ../EXPECTED.md. Distribution, not a gate._
