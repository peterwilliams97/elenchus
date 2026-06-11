# DEFECT — bare vs contextualized (target: the **Critic's calibration**, not a claim defect)

> Openly-constructed probe. Unlike the six axis probes, **this fixture has no engineered claim
> defect.** It targets the substance Critic's *own calibration* — specifically a suspected
> **false-attack** bias. The two inputs carry the *same* central claim; only the surrounding context
> differs. Any verdict difference between them is a property of the critic, not of the claim.

## What this probe is for

The laundering calibration (2026-06-10, haiku) found that substance rated Steve Ballmer's 2007
prediction — *"The iPhone will not get any significant market share"* — **`hollow` 10/10**. That
claim is paradigmatically *falsifiable*: a concrete empirical prediction that turned out decisively
false. The README spec for `substantive` is "well-formed and falsifiable — the kind of thing that
*could* be true." A falsifiable, well-formed prediction rated `hollow` is a **false attack** (a sound
claim killed), not a true negative.

The **hypothesis**: the substance critic penalises a **bare, decontextualised** claim — one presented
as a single line with no supporting argument — on the **Evidence** axis ("no source / no support
given"), and that penalty pushes it to `hollow`. If so, this is a **systematic** bias, because *every*
claim assayed in substance mode arrives as a decontextualised line by construction (the input is
prose, broken into atomic claims, each graded alone). The tool would then be structurally harsh on
exactly the inputs it is built to take.

## The two inputs (identical claim, different context)

- **`bare.txt`** — the Ballmer claim line exactly as in `../laundering/`:
  *"The iPhone will not get any significant market share."*
- **`contextual.txt`** — the same claim, wrapped in 2–3 sentences of its **real** supporting argument,
  drawn faithfully from the fetched source (the $500 subsidised price; ~1.3 billion phones sold
  annually; Apple's projected 2–3% vs a licensed platform's 60–80%). **Provenance applies:** every
  supporting fact is Ballmer's own, from the USA TODAY 2007 quotation recorded in
  `../laundering/PROVENANCE.md`. No supporting fact is invented.

## What the inputs deliberately hold constant

- The **central claim** is word-for-word identical in both.
- The claim's truth value is identical (both are the same false prediction).
- The claim's falsifiability is identical.
- Only the **presence of an in-text argument** varies. So a verdict shift isolates the critic's
  sensitivity to in-text support — the Evidence-axis-penalty signal.

## Measurement only — no fix in this package

This probe **measures**; it does not change `substanceCriticSys`. If the penalty is confirmed, the fix
(clarifying that substance evaluates a claim's *form*, not whether the claim cites its own support
in-text) is a **separate red-then-green follow-up** with its own Layer-2 prompt-presence test. Keeping
measurement and intervention apart is the discipline (see `../../../SESSION.md`).
