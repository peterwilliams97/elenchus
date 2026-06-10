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

## Calibration results

Populated by `../run.sh motte-and-bailey`. Latest summary lives in `results/`; ledger in
`../../../testing/calibration_log.jsonl`. Status: **`[x]`** — calibrated 2026-06-10/haiku, envelope
characterized (equivocation fires; atomization-split finding recorded). "`[x]`" here means *measured
and logged*, never "the critic is correct."
