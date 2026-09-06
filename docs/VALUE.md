# docs/VALUE.md — what a run is worth

## The measure

The value of a run is the difference between what a reader would do after reading the report's
**executive summary alone** and what they would do after reading **our tree**. Readers act on the
top of the tree. A leaf a reader never opens changed no decision, so it carries no value — only
cost.

- **Value** = summary lines that change a reader's action, *correctly*. Each is a top-level branch
  the tree opens because something below it needs attention.
- **Cost** = model spend, plus the minutes a reader spends on lines that did not need them.

Nothing on the reachable-but-unopened part of the tree is value. It is the price of finding the few
lines that are.

## The error that matters

The failure modes are not symmetric.

- **Worst: a misleading finding that reaches the summary as clean.** The reader acts on a false
  impression and never looks closer, because nothing told them to. This is the one error the tool
  exists to prevent, and every design choice pays to avoid it.
- **Small: a clean finding flagged for attention.** The reader opens a branch, spends a minute, and
  moves on. A handful of these per run is the correct price of never committing the first error.

So when in doubt, flag. A false flag costs a minute; a missed distortion costs the decision.

## The worked example — F29

`eval/cache-after`, claim F29: *"Victoria receives its fair share of funding from Creative
Australia."* Verdict **partial**. The reason: it is true only of the per-capita Creative Australia
slice — 26% of population against 26% of that programme's money — and that programme is about **3% of
total government arts funding**. The witness said exactly this. A reader who believed the summary
would conclude Victoria is funded fairly overall. It is not; the claim is true of one-thirtieth of
the money.

That finding must reach the top of the tree. **Today it collapses** — `partial` is not in the
Needs-you set, so its branch stays shut and the distortion reaches the reader as clean. That is the
worst error, and F29 is a live instance of it.

## What the code must do

1. **Needs-you gains a tier:** `partial` where the judge marks a **gap** — a scope, denominator,
   time-range, or attribution mismatch between what the summary implies and what the source
   supports. The gap is a structured field the judge returns (`{none, scope, denominator, timerange,
   attribution, other}`), not prose the tool greps. F29's gap is `denominator`; it now opens.

2. **The value line.** Every run prints, under the header, `changes N summary lines` — the count of
   top-level branches the tree opens. That number *is* the run's value; the cost sits in the USAGE
   line beside it, and a reader can weigh one against the other.

3. **Every needs-you leaf states the stakes.** A one-line *so what* from the judge (≤ 20 words):
   what a reader who believed the report would get wrong. It rides in the chain and shows on the
   leaf, so a flag is never just a colour — it says what is at risk.

Benign narrowing (`gap = none`) stays collapsed: flagging it would spend the reader's minutes for no
change in action, which is cost without value.
