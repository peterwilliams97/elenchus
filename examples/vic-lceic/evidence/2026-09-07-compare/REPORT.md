# Step-3 2×2 eval — faithfulness judge across model × retrieval

**Date:** 2026-09-07 · **Spend:** $2.38 of the $6 budget (Sonnet cells; Qwen cells $0) · **Judge:**
the single schema-enforced faithfulness judge (step 2), N=3 per claim. Same 8 claims
(`claims-faith-refuter.txt`: 5 real F-claims + 3 mutants M1/M2/M3). One `config.json` frozen per cell.
Every number traces to a cell's chain / usage / run.stderr — see `COMPARE.md` for the tables,
`adjudication.md` for the divergence rows. Nothing is synthetic.

The four cells: **sonnet-full** (whole corpus, the Sonnet gold, rerun on the new judge — not reused
from `eval/n3`), **sonnet-retrieved** (fused bm25+embed, `-max-tokens 10000`), **qwen-retrieved**
(same retrieval, `qwen3.6:27b-q4_K_M`), **qwen-oracle** (the exact gold passages per claim, via
`-retrieve=oracle`).

## What the cells say

- **Retrieval preserves Sonnet's judgment.** sonnet-full and sonnet-retrieved return the *same verdict
  on all 8 claims* (only F12's N=3 spread differs, 2/3 vs 3/3). Retrieval is 3.5× cheaper per claim
  ($0.066 vs $0.232) and slightly faster, at no verdict cost — the core result the retrieval work was
  for.
- **Qwen is systematically more lenient than Sonnet on the real claims.** It rates F8 and F31
  *faithful* where Sonnet rates them *partial* (it misses the training-loss distortion in F8 and the
  denominator gap in F31). This is the weaker-judge failure the axis-boundary note predicts: fluent,
  but it under-detects overreach.
- **Every cell catches every mutant (3/3).** M1 (contradicted), M2 (contradicted), M3 all rejected by
  all four cells. The mutant floor holds regardless of model or retrieval — the distortions the
  mutants encode are gross enough that even the lenient judge rejects them.
- **The oracle does not uniformly help Qwen.** Given the perfect passages, qwen-oracle moves M3 to
  *absent* (matching Sonnet, better) but moves F29 to *faithful* (more lenient, worse than
  qwen-retrieved's partial). Perfect retrieval is not sufficient for a weak judge — the gap is in the
  judging, not only the evidence.
- **The grounding check fires on every model.** Quote-verification rejects: qwen-retrieved 34 (it
  paraphrases most), sonnet 15–19, qwen-oracle 16. No model quotes perfectly; code catches the
  non-verbatim ones on all of them.
- **Cache + grouping work on the Anthropic path.** cache_hit 96% (full) / 89% (retrieved) — claim
  grouping over a shared prefix kept it hot. 0 schema retries in any cell: the schema held.

## Caveats

- Qwen wall time is 3× Sonnet's (117 s/claim retrieved) — free but slow.
- The Anthropic path uses a forced strict-tool call for structured output (not `output_config.format`,
  whose raw-HTTP shape was not live-verifiable); it returned real, schema-valid verdicts on all 24
  live calls, 0 retries.
- The first sonnet-full attempt failed on a sandbox TLS interception ($0, no billing); the Anthropic
  cells were rerun with the sandbox disabled.

## Adjudication

3 divergences between sonnet-retrieved and qwen-retrieved (F8, F31, M3), with both judges' verified
quotes side by side, in `adjudication.md`. The **who's right** column is left blank — that call is a
human's.
