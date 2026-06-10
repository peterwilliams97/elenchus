# DEFECT — axis-incompleteness probes (target: **defects the seven axes do NOT name**)

> Openly-constructed adversarial probes. The construction *is* the ground truth. See `../README.md`.
>
> **Outcome is unknown by design.** Unlike the other five probes, this one has no "should fire" axis.
> It maps the *edge* of the seven-axis instrument. A miss here is a **mapped envelope limit
> (BACKGROUND.md W6), recorded — not a bug.** A catch (via an adjacent axis) is an interesting datum
> about the axes' reach.

The seven fixed axes in `substanceCriticSys` are: Evidence, Hidden premise, Falsifiability,
Equivocation, Base rate / magnitude, Counterexample, Causality vs correlation. None of them *names*
the three classical defects below. The question this probe answers empirically: does the critic catch
them anyway through an adjacent axis, or do they pass clean?

## The three claims (one defect each)

### 1. Category error
> "A monolith is a more honest architecture than microservices."

*Honest* is a property of agents and statements; it cannot be truly predicated of an architectural
style. The sentence is grammatically fine and rhetorically familiar (engineers say things like this),
but it commits a **category mistake** — applying a moral-epistemic predicate to a structural object.
No axis names category errors. Likely to pass clean.

### 2. Composition (part → whole)
> "Every engineer on this team is individually excellent, so this is an excellent team."

The **composition fallacy**: inferring a property of the whole from the same property of each part.
Individually excellent engineers can compose a poor team (coordination, communication, conflicting
styles). No axis names composition/division. **Counterexample** is the most likely adjacent catch —
"excellent individuals, dysfunctional team" is a clean counterexample — so this one may not pass.

### 3. Survivorship framing
> "The founders who changed the world all ignored the skeptics and bet everything on one idea, so the
> lesson is to ignore the skeptics and bet everything."

**Survivorship bias**: the inference counts only the winners and is blind to the identically-behaving
founders who ignored the skeptics, bet everything, and *failed* — they are not in the sample. No axis
names survivorship. **Base rate / magnitude** is the most likely adjacent catch (the missing
denominator of failures is a base-rate omission), so this one may not pass either.

## Why these three, and what clean-passing would mean

These are well-known reasoning defects that the seven-axis list simply does not enumerate. If the
critic flags them, it is reaching beyond its named axes (informative — the axes under-describe the
critic's actual behavior). If it passes any `substantive`/`partial` with no relevant finding, that is
the **honest failure** recorded in BACKGROUND.md W6: the named defect sits *outside the envelope*
until an axis is added. Either result advances the envelope map; neither is a bug to be fixed by
patching the fixture green.

## Axes each is otherwise clean on

All three avoid equivocation, staked statistics, and unfalsifiable loops — they are short, plain, and
carry exactly one un-named structural defect, so any finding the critic *does* produce is attributable
to that defect's nearest axis (if any).
