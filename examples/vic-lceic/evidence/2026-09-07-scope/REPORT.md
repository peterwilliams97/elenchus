# Prompt 2 — judge-prompt scope change

**Date:** 2026-09-07 · **Spend:** ~$1.59 (before-sonnet $0.79 + after-sonnet $0.80; qwen $0) ·
hearings + submissions corpus, retrieved, N=3, both backends. Scoring set (11 claims) excludes F8/F31
by design. Tables in `COMPARE.md`; the 5 new claims' rows in `adjudication.md`.

The change to `faithJudgeSys`: `faithful` requires the source to state the claim's **subject, scope
and direction**; an adjacent or broader statement is `partial` at best (`gap=scope`). Two worked
negatives were added, from Peter's readings — the lost-training-opportunities claim (adjacent: online
training happened but degraded) and regional Victoria vs. regional Australia (broader place).

## What the runs say

- **No verdict changed before→after, either backend.** Sonnet was already applying scope discipline
  (its before and after verdicts are identical on all 11 claims); qwen's only movements (F29
  partial↔faithful, F10 absent 3/3→2/3, M1 contradicted 3/3→2/3) are N=3 temperature spread, not a
  systematic tightening. The prompt change is **safe** — no regressions.
- **Mutants still caught 3/3 in every cell** (M1/M2/M3 contradicted or absent). The scope wording did
  not weaken the mutant floor.
- **Why no movement:** the two claims the scope rule was written to catch (F8, F31) were *excluded*
  from the scoring set by design, so the set tests non-regression on the other claims, not the rule's
  positive effect. Confirming the rule flips F8/F31 would need them back in a set — they are the
  worked negatives in the prompt, so re-scoring them would be marking the prompt against its own
  examples.
- **The 5 new scope-risk claims** (no gold; `adjudication.md`): F2a (320,000 Victorians) and F4
  (Victoria higher than all jurisdictions incl. NY/Sweden/UK) both rated **faithful** by both judges;
  F10, F33, F54b rated **absent** (no verified quote). **F4 is the one to look at** — the verified
  quote is "Victoria rated higher … than all other jurisdictions *where this research was conducted*",
  a scope hedge the claim drops by naming New York, Sweden and the UK; a human may call it
  `partial`/`gap=scope`. That judgment is left blank in `adjudication.md`.

## Also shipped

- **Resume** (`-fresh` to opt out): a faithfulness run reuses any claim already in the chain, so the
  OOM kill that stopped `after-qwen` at 3/11 was recoverable — the restart continued rather than
  re-paying. Grouping completes claims out of order, so the resume reader tolerates gaps in the
  partial chain. `no_quote_downgrades 1` in the qwen cells is the grounding rule firing on one
  ungrounded sample.
