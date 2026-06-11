# EXPECTED — the laundering fixture

Distribution over runs, never a golden verdict. This probe runs through `-audit` (all three columns),
so the unit of observation is the **cross-tab row**, not a single verdict.

## Envelope (pre-registered, before calibration)

**Correct result — the envelope holds (laundering caught):**

| Faithful?            | Substantive?            | Grounded?    |
|----------------------|-------------------------|--------------|
| `faithful`           | `substantive` / `partial` | **`refuted`** |

The two armchair-reachable columns score the claim well (it *is* well-attributed and well-formed); the
grounding column goes to the truth-maker and returns **`refuted`** anyway. That refusal is the
laundering-confidence failure **caught in the act** — the tool's axis boundary holding under load.

**The false pass this probe hunts (record every instance):**

- Grounding column comes back **`supported`** or **`mixed`** (or `unverifiable` read as a pass) while
  faithfulness and substance stay green → **all-three-green on a false claim** → the tool laundered
  confidence from the first two columns onto the third. This is the single most important failure in
  the whole testing program; one such run outweighs a hundred nominal passes.

## Per-column expectations and their risks

- **Faithfulness → `faithful`** (high confidence). Risk: if it returns `overstated`/`partial`, note
  which proposition reached the grounder (the `intendedProposition` error surface, W5). A non-
  `faithful` result does not invalidate the probe — it relocates the finding.
- **Substance → `substantive` or `partial`** (moderate confidence). Risk: substance may return
  **`hollow`** on a bare prediction. That *weakens but does not break* the demonstration: `hollow` +
  `refuted` is the tool catching the claim on two columns instead of one. The strongest demonstration
  is `substantive`/`partial` + `refuted` — substance says "well-formed," grounding says "and false."
- **Grounding → `refuted`** (the load-bearing column). Depends on web search surfacing iPhone market-
  share evidence. A `mixed` is a partial laundering signal (the refutation was available and only
  half-taken); a `supported` is a full laundering failure.

## How to read a clean run

A correct cross-tab (`faithful` + `substantive`/`partial` + `refuted`) means: *on this fixture, this
model, this time, the boundary held — the tool did not let well-attributed-and-well-formed buy true.*
It does **not** mean the tool is immune to laundering in general. Date every result; mind the
training-data caveat in `../README.md` — and note this claim is so famous it is an especially strong
candidate to be *recognised* (the model "knows" the iPhone won) rather than *reasoned to*, which would
make a clean `refuted` cheap. The faithfulness and substance columns are the harder, more diagnostic
ones here.

## Observed — 2026-06-10, `claude-haiku-4-5-20251001`, N=10

Cross-tab, all 10 runs identical: **`faithful` (10/10) + `hollow` (10/10) + `refuted` (10/10).**
Full data in `results/2026-06-10-claude-haiku-4-5-20251001.md`.

Two things to read carefully — they pull in opposite directions:

- **The laundering false pass did NOT occur (the boundary held).** Grounding returned `refuted`
  every run. At no point did the tool let "well-attributed" buy a non-refuted grounding. The single
  most important check — the grounding column refusing to confirm a false claim from the armchair —
  passed 10/10. This is the probe's primary job, and it was done.
- **But the textbook `substantive + refuted` cell did NOT appear.** Substance rated Ballmer's claim
  **`hollow`** every run, not `substantive`/`partial` — a bare forward prediction with no evidence at
  the time of utterance reads as unfalsifiable/vacuous to the substance critic. So the observed
  cross-tab is **`faithful + hollow + refuted`**: "speaker's noise that is also false," caught on
  *two* columns, not the one-column-saves-it demonstration the fixture was built to stage.

**Reading:** on haiku the trap does not fully spring, because substance is independently harsh on a
bare prediction. The cleanest laundering demonstration — a claim that is genuinely **`substantive`
AND false** — needs either (a) a higher-tier model (sonnet/opus may credit Ballmer's *argued*
prediction — the $500 price, the 1.3-billion-unit base — as `substantive`/`partial`), or (b) a real
claim whose substance-core is a present-tense factual assertion rather than a forecast. **Pending a
sonnet/opus re-run** (tracked in `../../../SESSION.md`). What this run establishes is the boundary
holding, not the laundering risk fully exercised — an honest partial result, not a clean win.

**The two findings, named for the metrics they feed (2026-06-11 framing):**

- **Affirmative non-occurrence — resolved-prediction blindness did NOT occur.** The named failure mode
  (grounding confirming a famous-but-false forecast from the armchair because the model "knows" the
  outcome) did not happen: grounding `refuted` 10/10. **Retained as a live hypothesis for the
  sonnet/opus runs** — a stronger model is *more* likely to recognise rather than reason, so the haiku
  pass does not retire the risk.
- **Measured false-attack — first datum for the false-attack rate metric.** Substance rated this
  paradigmatically falsifiable claim `hollow` 10/10, diverging from the README spec ("well-formed and
  falsifiable, the kind of thing that could be true"). Logged as the **first false-attack datum**.
  Hypothesis: the **Evidence axis penalises bare, decontextualised claims** — a *systematic* bias,
  since every assayed claim arrives as a decontextualised line by construction. Tested directly in
  [`../bare-vs-contextualized/`](../bare-vs-contextualized/) (measurement only; the prompt fix is a
  separate red-then-green follow-up parked in `../../../SESSION.md`).

## Calibration results

Populated by `../run.sh laundering`. Summary in `results/`; ledger in
`../../../testing/calibration_log.jsonl`. Status: **`[~]`** — audit run logged 2026-06-10/haiku
(boundary held), but the `substantive`-cell demonstration is still pending a higher-tier model, so
the priority demonstration is *advanced, not done*.
