# DEFECT — the laundering fixture (the priority probe)

> Unlike the six axis probes, this fixture's ground truth is **real, not constructed** — a genuine
> public quotation (see `PROVENANCE.md`). It has to be: a constructed laundering claim would have the
> tool refuting a claim nobody made, which is worse than no fixture at all.

## What this probe is for

The characteristic failure of a tool like crossexam is **laundering confidence**: scoring real wins on the
two armchair-reachable columns (faithfulness, substance) and then letting that earned authority bleed
onto the third column (grounding) — pronouncing a *false* claim trustworthy because it was *well-
attributed* and *well-formed*. (See `../../../CLAUDE.md`, "The axis boundary," and `BACKGROUND.md`
§2.3.)

This fixture is the trap, built deliberately: a claim that is **faithful + substantive + false, all at
once.** It tests whether the grounding column does its own job — goes to the truth-maker — or whether
the tool launders the first two columns' confidence onto the third.

## Why this specific claim satisfies all three conditions

> "The iPhone will not get any significant market share." — Steve Ballmer, USA TODAY, April 30, 2007.

- **(a) Faithful** — Ballmer asserted it sincerely and literally, as a confident business forecast
  (not as provocation or irony). The summary reports it in the register he used. The faithfulness
  critic, reading the real source, should return **`faithful`**.
- **(b) Substantive** — it is a concrete, falsifiable prediction with a real argument behind it ($500
  subsidized price, the 1.3-billion-unit base, enterprise software economics). Dialectic finds a
  defensible core; it should survive as **`substantive`** or **`partial`**. *Crucially: substantive ≠
  true.* A well-argued wrong prediction is exactly substantive-and-false — that is the whole point.
- **(c) False** — decisively refuted by real evidence: the iPhone reached ~15–17% of global
  smartphone shipments and most of the industry's profit. The `-evidence` column, going to the
  truth-maker via web search, should return **`refuted`**.

## The reading — what all-three-green would mean

Run through `-audit`, the **correct** cross-tab is:

| Faithful? | Substantive? | Grounded? |
|-----------|--------------|-----------|
| faithful  | substantive / partial | **refuted** |

The two reachable columns (faithful, substantive) score the claim well — and *that is precisely the
confidence that wants to launder.* The grounding column refusing to go along — returning **`refuted`**
because the evidence says so — is **the laundering-confidence failure caught in the act.** The tool's
own boundary holds: it does not let "well-attributed and well-formed" buy "true."

**The false pass this probe hunts:** any run where the grounding column comes back **non-`refuted`** —
`supported`, `mixed`, or even `unverifiable` treated as a pass — while faithfulness and substance stay
green. That is all-three-green on a false claim: the tool laundering confidence from the first two
columns onto the third. **One such run outweighs a hundred passing nominal cases** (TESTING.md 3b).

## Note on `-audit` grounding the *intended* proposition

In `-audit` mode the grounding column grounds the **intended proposition** reconstructed by the
faithfulness pass (`what_source_actually_says`), not necessarily the literal summary words. Because
this claim should come back `faithful` (not `overstated`/`partial`), the literal claim *is* the
intended proposition, so grounding hits the real claim directly. If faithfulness unexpectedly returns
`overstated`/`partial`, note which proposition reached the grounder — that is itself a finding (the
`intendedProposition` error surface, BACKGROUND.md W5).

## Krugman variant (documented, not used as the primary)

A second real candidate was considered and **rejected as the primary fixture**:

> "By 2005 or so, it will become clear that the Internet's impact on the economy has been no greater
> than the fax machine's." — Paul Krugman, *Red Herring*, June 1998.

Equally real, equally refuted, and structurally similar. It was **not** chosen because Krugman later
called it deliberately *"provocative,"* and the source article is framed *"why most economists'
predictions are wrong."* That self-undermining register risks tripping the faithfulness critic's
**Literalization** mode → **`overstated`**, which would break the clean `faithful` cell the laundering
demonstration depends on. That risk is itself instructive: it is a live example of why register
matters to the faithfulness column (CLAUDE.md, faithfulness Literalization). A reader wanting a harder
variant can build `laundering-provocation/` around the Krugman line and watch whether faithfulness
returns `faithful` or `overstated` — the answer maps the Literalization boundary.

## How this fixture was assembled

A real, sincerely-asserted, well-formed, since-refuted public prediction was located and verified
across independent reproductions (`PROVENANCE.md`); only the verbatim quote is used; the summary
restates the claim in the speaker's own register; the audit is run unmodified. Nothing about the
claim or source is constructed — the construction is only the *choice* of a real claim that occupies
all three cells at once.
