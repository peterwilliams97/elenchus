# Reflexive pass — crossexam run on its own claims

The tool's own thesis, turned on the tool. assay exists to keep three questions apart — *faithful?
well-formed? true?* — and to refuse to claim more than each mode can settle. If that discipline is
real, it must survive being applied to **assay's own load-bearing sentences**. This directory is that
test (TESTING.md §3d, work packages B1–B2, 2026-06-11, `claude-haiku-4-5-20251001`).

Two passes:

- **B1 — grounding canary** (`headline.txt`): the headline is a claim about *effects on people*
  ("seeing the three columns disagree causes a trained user to recognise their skill as
  transferable"). By the tool's own axis boundary that cannot be settled from the armchair. Required
  result: `unverifiable` every run.
- **B2 — substance** (`claims.txt`): eight verbatim load-bearing claims from `README.md` and
  `CLAUDE.md`, run through the dialectic. Does the tool rate its own claims the way it asks you to?

## B1 — canary result

`./assay -evidence headline.txt`, N=5: **`unverifiable` 5/5.** The grounding column did **not** confirm
the tool's own value proposition from the armchair — no self-sealing failure. A `supported` here would
be the earliest sign that grounding had started pronouncing without a truth-maker; it did not happen.
(Now wired into `run.sh` as the always-on final step of every full calibration pass.)

This is the *most important* reflexive result and it came out right: the boundary holds even when the
claim is the tool's own headline.

## B2 — substance result

`./assay claims.txt`, N=3. `decompose` expanded 8 claims into ~19–21 atomic fragments per run.
Fragment-verdict distribution (59 total): **12 substantive · 18 partial · 27 hollow · 2 error**
(`results/2026-06-11-claude-haiku-4-5-20251001.md`).

### Pre-registered expectations vs. observed

| Expectation (written up front) | Observed | Reading |
|---|---|---|
| Headline rates **partial-at-best** (we already carry it as a hypothesis) | **partial** (2/2 appearances) | **Confirmed.** The tool rates its own headline exactly as honestly as we frame it — a hypothesis, not a finding. |
| Axis-boundary claims rate **substantive** | **Mixed — many `hollow`/`partial`** | **Surprise — envelope data.** Not because the claims are unsound, but because of a *referent collision* (below). |

### The finding: decontextualisation collides the referent

This is the headline result, and it is the **same lesson as the bare-vs-contextualized probe**
(`../destructive/bare-vs-contextualized/`), demonstrated reflexively. Stripped to a bare atomic line,
the tool's specialist vocabulary loses its meaning:

- In **4 fragments** the critic read "**assay**" as a *chemical* assay — "when an assay method is
  properly calibrated and validated for a specific substance…" — not the tool. "assay is a mirror, not
  a crutch" was graded against metallurgy.
- "Two of assay's three **columns** cannot reach the thing the third column is for" → **`hollow` 3/3**:
  decontextualised, "columns" and "the third column" have no referent, so the critic correctly finds
  the *bare sentence* opaque — while the underlying claim is true and well-formed in context.
- "Predictions come back unverifiable under **`-evidence`**" → `hollow`: an opaque flag with no frame.

The critic is not *wrong* that these bare lines are under-specified. The point is sharper: **assay's
own envelope — substance on decontextualised specialist claims is unreliable — bites assay's own
documentation.** Every claim the tool grades arrives decontextualised by construction, so this is not
a quirk of our claims; it is the structural false-attack surface the bare-vs-contextualized probe
isolates, caught here in the mirror.

### Layer-4 spot-audit (trained reader marks agree / disagree)

The tool's own thesis is that the human is the verifier of last resort. So:

| Claim (verbatim) | modal verdict | agree? | why |
|---|---|---|---|
| "A claim can be faithfully reported, internally well-reasoned, and still false." | `substantive` | **agree** | The load-bearing separation claim; sound and well-formed. The tool rates its own foundational claim substantive. |
| "Seeing the three columns disagree causes a trained user to recognise their skill as transferable." | `partial` | **agree** | It *is* a hypothesis about effects on people; partial-at-best is the honest verdict, matching how we carry it. |
| "Two of assay's three columns cannot reach the thing the third column is for." | `hollow` | **disagree** | True and well-formed in context; rated `hollow` only because decontextualisation stripped the referent. → envelope limit, not a README defect. |
| "Reasoning can refute a grounding claim … but can never confirm one." | split (`substantive`/`partial` on *refute*; `hollow` on *never confirm*) | **partly agree** | The refute half grades well; the bare universal "never confirm one" reads as over-strong without the axis-boundary frame. |
| "assay is a mirror, not a crutch." | `hollow`/`partial` | **disagree with the basis** | Graded against a *chemical-assay* reading (referent collision). The metaphor is fine in context. |

**Disagreement triage** (per TESTING.md Layer 4): every disagree above maps to **one** recorded
envelope limit — *decontextualised specialist / self-referential claims collide their referent and
skew `hollow`/`partial`* — the same family as bare-vs-contextualized. Recorded, not patched. Candidate
new 3b fixture noted in `../../SESSION.md`: a **referent-ambiguity probe** (a domain-polysemous term
graded with vs without a one-line domain anchor).

## What we did NOT do

We did **not** edit a single README/CLAUDE.md claim in response to these verdicts (measurement before
intervention). None of the `hollow` verdicts exposes a genuinely hollow *claim* — they expose the
tool's decontextualisation envelope acting on sound claims. The one expected soft verdict (headline →
`partial`) is the result we *want*: the tool refuses to over-credit its own headline. Had any sentence
rated genuinely hollow on its merits, it would be logged in `../../SESSION.md` as a candidate edit with
the verdict attached — none did.
