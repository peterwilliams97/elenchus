# 2026-09-10 sonnet-retrieved re-run — run-to-run variance vs 2026-09-09

Re-run of the 2026-09-09 `sonnet-retrieved` cell under identical config (corpus
`hearings+submissions+qon`, judge `claude-sonnet-4-6`, `-retrieve bm25 -n 3 -max-tokens 10000
-floor 0 -embed`, manifest `sources/MANIFEST.md`, input `claims-machine-full.txt`, 69 findings).
Only the harness differs: the working-tree `assay.go` records every `-n` sample's verdict in the
chain; `faithJudge` itself is unchanged, so the judge is the same. Purpose: measure how stable the
judge's majority verdict is across two independent runs of the same config.

## Rollup

| run | verified | faithful | partial | overstated | absent | contradicted | unsupported | unverifiable | est_usd | wall |
|---|---|---|---|---|---|---|---|---|---|---|
| 09-09 | 69/69 | 28 | 26 | 0 | 11 | 1 | 1 | 2 | $8.39 | 1h01m |
| 09-10 | 69/69 | 30 | 22 | 1 | 11 | 1 | 2 | 2 | $8.38 | 1h03m |

09-10 est_usd = $8.1087 (main) + $0.2673 (idx-5 re-judge) = **$8.376**.

## Per-leaf majority-verdict differences: 16 of 69

Of the 16, **7 were a 2/3 split in either run** (marked ✓); the other 9 flipped their majority
despite being unanimous 3/3 in *both* runs — the judge's stable answer itself moved.

| idx | 09-09 | 09-10 | 2/3 in either | claim (truncated) |
|---|---|---|---|---|
| 3  | partial (3/3)  | absent (2/3)      | ✓ | industries provide significant economic stimulus… |
| 5  | partial (3/3)  | absent (3/3)      |   | engagement brings communities together… (†) |
| 15 | partial (2/3)  | faithful (3/3)    | ✓ | insecure/low wages impact retention… |
| 19 | partial (3/3)  | absent (3/3)      |   | COVID-19 exposed/intensified systemic issues… |
| 23 | absent (3/3)   | partial (2/3)     | ✓ | infrequent grant programs leave more applicants… |
| 25 | partial (3/3)  | faithful (3/3)    |   | balance between infrastructure investment… |
| 37 | partial (3/3)  | faithful (2/3)    | ✓ | strategic goal to reflect NSW's population… |
| 38 | partial (3/3)  | overstated (3/3)  |   | disappointing ABC/SBS expanded their… |
| 43 | faithful (3/3) | partial (3/3)     |   | ABC content-production shaped by state screen… |
| 45 | absent (3/3)   | partial (3/3)     |   | rising production costs + budget reduction… |
| 54 | absent (3/3)   | partial (3/3)     |   | ABC's vital role in supporting the state… |
| 57 | faithful (3/3) | partial (2/3)     | ✓ | ABC delivers 16 unique regional local-radio programs… |
| 58 | partial (3/3)  | unsupported (2/3) | ✓ | producing over 112 hours of local content… |
| 61 | partial (3/3)  | absent (3/3)      |   | cost of producing content in regional Victoria… |
| 62 | absent (3/3)   | partial (3/3)     |   | producing content in regional areas delivers… |
| 68 | partial (2/3)  | faithful (3/3)    | ✓ | SBS deeply engages multilingual/multicultural… |

(†) idx 5 hit one transient `api.anthropic.com` timeout in the main pass; re-judged in a resume
pass to `absent (3/3)`. Its diff is a genuine re-judge, not the error. Pre-re-judge chain kept at
`claims-machine-full.faithfulness.jsonl.bak-with-err`.

## Reading

Majority verdict moved on 16/69 = 23% of leaves between two runs of the *same* judge and config.
Most of the churn is `partial ↔ absent/faithful` — adjacent verdicts on the support spectrum — but
the direction is not one-sided (partial→faithful and partial→absent both occur), and 9 flips
happened with 3/3 unanimity on both sides, so the instability is not confined to leaves that
already looked split. Provenance: all inputs real (LCEIC hearing/submission/QON corpus,
`sources/MANIFEST.md`); no fabricated inputs.
