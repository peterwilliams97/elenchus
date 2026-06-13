# EXPECTED — motte-and-bailey

Expected results are a **distribution over runs**, never a single golden verdict. assay's judgment
layer is non-deterministic (see `../../../TESTING.md`, "The central constraint"). Read the modal
verdict and the spread, not one run.

## Envelope (pre-registered, before calibration)

- **Equivocation (or an adjacent term-shift catch — Counterexample / Hidden-premise) should fire in
  the large majority of runs.** The claim should not survive as `substantive`.
- Likely modal verdict: **`hollow`**, or **`partial`** surviving only as the narrow, honest motte
  ("the engine adapts its output from feedback") with the strong "genuinely intelligent" reading
  stripped. A `partial` whose surviving claim is the motte is a *success*, not a leak — the critic
  refused the bailey.
- **False pass (the thing this probe hunts):** any run rating the claim **`substantive`** with **no**
  finding touching the equivocation / term-shift / motte-retreat. Record every such run — false
  passes are the costly direction.

## What a clean run means (and does not mean)

A clean run says only: *the envelope held on this fixture, this model, this time.* It does **not**
say "the critic understands motte-and-bailey." Note the training-data caveat in `../README.md`:
this probe is public and may eventually be recognised rather than reasoned about. Date every result.

## Observed — 2026-06-10, `claude-haiku-4-5-20251001`, N=10

Distribution (claim-verdicts summed across runs; `decompose` split each run into ~6–7 atomic claims):
**40 hollow · 19 partial · 7 substantive · 2 error.** Equivocation named in every run's critique.
Full data in `results/2026-06-10-claude-haiku-4-5-20251001.md`.

- **No clean false pass:** no run rated the strong "genuinely intelligent" claim `substantive`
  *without* an equivocation/term-shift finding. The target axis fired; the bailey claims collapsed to
  `hollow`.
- **A real structural finding (W4 atomization asymmetry):** the substance `decompose` step **splits
  the motte from the bailey into separate atomic claims**, so the critic evaluates each in isolation
  rather than catching the *retreat between them* as one move. The isolated motte ("systems that learn
  from feedback are intelligent") is the source of the 7 `substantive` verdicts — defensible *on its
  own*, which is exactly the motte's job. The equivocation is still named on the holistic bailey
  claim, but the move's structure is partly dissolved by atomization before the critic sees it. This
  is a property of the decompose-then-grade pipeline, recorded as an envelope note — not a fixture bug.

## Observed — 2026-06-11, `claude-haiku-4-5-20251001`, N=10 (fragment-attributed)

Fresh calibration with per-fragment chains persisted (the 2026-06-10 chains were discarded by an
instrument error — see `../../../testing/SCHEMA.md` and `decision_log` d016 — so this is a *new
sample*, not a rescore of 2026-06-10). Distribution (64 fragment-verdicts, `claims_per_run` 5–7):
**41 hollow · 13 partial · 9 substantive · 1 error.** `N_valid` = 63 (the error is excluded from
envelope rates).

**Fragment attribution of the 9 `substantive` verdicts** — the breakdown the 2026-06-10 aggregate
could not give: **8 land on the defensible MOTTE in isolation** ("the engine adapts / learns from
interactions"), which is the critic being *correct*, exactly as the 2026-06-10 W4 note anticipated (a
surviving motte is a success, not a leak). **1 clean false pass:** in run 10 `decompose` did *not*
split the move — it kept the equivocation conditional ("if one accepts a system learns and adapts,
one has granted it is intelligent") as one fragment, and the critic rated it `substantive` with an
**empty critique** (no Equivocation / Counterexample / Hidden-premise finding). That empty critique
is the false-pass signature: not a wrong axis, but *no* axis. **False-pass rate: 1/63**, versus the
naive reading of all 9 `substantive` as 9 leaks — the unit-of-analysis correction in action.

## Calibration results

Populated by `../run.sh motte-and-bailey`. Latest summary lives in `results/`; ledger in
`../../../testing/calibration_log.jsonl`. Status: **`[x]`** — calibrated 2026-06-10/haiku, then
fragment-attributed 2026-06-11/haiku (1 clean false pass / 63; 8/9 `substantive` correct-on-motte).
"`[x]`" here means *measured and logged*, never "the critic is correct."
