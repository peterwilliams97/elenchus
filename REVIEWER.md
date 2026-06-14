# REVIEWER.md

A design proposal for making `crossexam` useful to a conference reviewer without turning it into
another model that emits a wall of text. Two questions, each investigated against the frozen `spec/`
and pre-registered in `rigour-map/decision_log.jsonl` (d029, d030). Research only — this document
proposes; it ships no code.

Each answer is tagged:

- **projection (cheap)** — a new way to render data `crossexam` already computes; no new model
  calls, no new fields on shared structs.
- **new behavior (deferred)** — would change what gets computed; written up as a design-queue item,
  not implemented.

---

## Q1 — show only the claims where the three filters disagree, as questions

**Tag: projection (cheap).**

**Status (2026-06-14, d031):** the rule and its renderer are built and tested in
`internal/render/disagreements.go` — `tierOf`, `surfaces`, `disagree`, and `Disagreements([]AuditRow)`,
pinned by a truth-table test over all 64 tier-triples plus the verdict→tier mapping
(`internal/render/disagreements_test.go`). The `-disagreements` flag wiring (c) is **not** built: it
needs the audit runner and flag parsing, which do not exist yet (`cmd/crossexam` is a no-op `main`).
See SESSION.md and decision d031.

### What `-audit` already computes per claim

`runAudit` (spec/BEHAVIOR.md, Mode 4) runs all three filters on every claim and ends with three
verdicts per claim:

- **faithfulness** — `faithful` · `partial` · `overstated` · `absent` · `contradicted` · `error`
- **substance** — `substantive` · `partial` · `hollow` · `error`
- **grounding** — `supported` · `mixed` · `refuted` · `unverifiable` · `skipped (over cap)` · `error`

These three strings already exist together for each claim: `progressDone` is handed
`"faith=<v> sub=<v> ev=<v>"` (spec/BEHAVIOR.md line 204), the audit JSONL nests all three detail
objects under `faith`/`substance`/`evidence` (spec/CLI.md, `auditDetail`), and `mdAudit` prints them
as the three columns of its table (spec/CLI.md lines 109–115). The proposal reads exactly those three
verdicts plus the claim string — nothing else.

### (a) The exact rule for "the columns disagree"

The three columns speak three different vocabularies, so they cannot be compared as equal/unequal
strings. The only cross-mode mapping the spec provides is the terminal colour routing, `vcolor`
(spec/BEHAVIOR.md lines 285–292). Borrow it as a severity tier:

| Tier     | Verdicts |
|----------|----------|
| **PASS** | `faithful`, `substantive`, `supported` |
| **WEAK** | `partial`, `overstated`, `mixed` |
| **FAIL** | `hollow`, `absent`, `refuted`, `contradicted` |
| **NULL** | `unverifiable`, `error`, `skipped (over cap)` |

**Surface a claim if, and only if, at least one of its three columns is PASS and at least one is
FAIL.** Everything else is dropped:

- all three PASS, or all three FAIL → the filters agree; no question to ask.
- no FAIL anywhere (only PASS/WEAK/NULL) → nothing failed; merely unsettled, not contested.
- no PASS anywhere (only FAIL/WEAK/NULL) → nothing passed; a clear reject, not contested.

WEAK and NULL never trigger a surface on their own. This is deliberate: grounding returns
`unverifiable` on most predictions, and counting that as "disagreement" would surface nearly every
row and reproduce the wall of text. A claim surfaces only when one lens genuinely says *yes* while
another genuinely says *no* — the case a reviewer actually has to adjudicate.

`error` and `skipped (over cap)` are NULL: a column that errored or was skipped neither passes nor
fails, so it can never be one half of a conflict. A claim with an errored column still surfaces only
if its other two columns supply a PASS and a FAIL between them.

**Worked check against `examples/dan_shipper/` (the 12-claim audit in that README).** Mapping each
row to tiers, the rule surfaces exactly **#3, #5, #12** — every "faithfully reported, evidence holds,
but hollow under scrutiny" row. It drops #9 (`overstated`/`hollow`/`refuted` — no PASS, a clean
reject) and #6/#8 (`absent` at faithfulness — no PASS, summarizer's noise), which is correct: those
need no adjudication. Three of twelve surface. That is the recognition output, not a dump.

### (b) The output shape

One entry per surfaced claim: a plain-English question built deterministically from the tier of each
column, followed by the three raw verdicts in brackets. No tables, no distributions, no percentages.

The question is assembled from one fixed clause per column, chosen by that column's verdict — a
template lookup, not a model call:

| Column       | PASS                        | WEAK                               | FAIL |
|--------------|-----------------------------|------------------------------------|------|
| faithfulness | "the source really says it" | "the source says a weaker version" (`partial`) / "the summary strengthened the source" (`overstated`) | "the source does not say it" (`absent`) / "the source says the opposite" (`contradicted`) |
| substance    | "it survives scrutiny"      | "only a narrower version survives" | "it is hollow under scrutiny" |
| grounding    | "the evidence backs it"     | "the evidence is split"            | "the evidence contradicts it" |

The PASS clauses and the FAIL clauses are joined with "but", then closed with a fixed prompt to the
reviewer. For `crossexam`'s own default input claim #3 (faithful · hollow · supported):

```
Claim 3 [faithful · hollow · supported]
  The source really says it and the evidence backs it, but it is hollow under scrutiny.
  Keep it, qualify it, or cut it?
  "SaaS is not dead — I would buy SaaS stocks right now."
```

And a faithful · substantive · refuted claim (an anti-signal — checkable and wrong) would read:

```
Claim N [faithful · substantive · refuted]
  The source really says it and it survives scrutiny, but the evidence contradicts it.
  Keep it, qualify it, or cut it?
  "..."
```

The closing question is the same every time ("Keep it, qualify it, or cut it?"); only the pass/fail
clauses vary with the verdicts. The reviewer reads one short question per contested claim and
decides.

### (c) The trigger

A flag: **`-disagreements`** (bool, default false). It names exactly what it does — it shows the
claims where the three filters disagree. It does not collide with any existing flag (`-model`,
`-source`, `-evidence`, `-audit`, `-text`, `-md`, `-v`/`-verbose`, `-no-color`, `-max-rounds`,
`-max-claims`, `-progress`, `-quiet`, `-usage-out`, `-chain-dir` — spec/CLI.md), and it does not
read as a merge "conflict", a CLI "flag", or a generic "review".

`-disagreements` runs the same three-filter audit computation and requires `-source` for the same
reason `-audit` does (faithfulness has no source otherwise). It changes only the rendering: instead
of `mdAudit`'s full table it emits the disagreement-questions list above. Two reasonable shapes for
the wiring, both pure rendering:

1. `-disagreements` implies the audit run, so `crossexam -disagreements -source S.txt summary.txt`
   stands alone (mirrors how `-audit` implies `-md`).
2. `-disagreements` is a modifier on `-audit`, fatal without it.

Either is a one-flag, render-only change. Shape (1) is preferred — one flag for one intent, matching
the "name it for what it does" rule.

### Confirmation: this is a projection, not new behavior

- **No new model calls.** The three verdicts already exist after the audit run; the filter and the
  question template are arithmetic and string lookup over them.
- **No new fields on shared structs.** The renderer needs the claim string and the three verdicts —
  exactly what `mdAudit` already reads to draw its table.
- **What is new is one render function** in `internal/render` (a new output path beside
  `mdAudit`/`termSubstance`/etc.), plus flag wiring in `cmd/crossexam`. That is permitted: a new
  renderer is not a new field on a shared struct and not a new call path to the API.

It can be done without new calls or new state. **Projection (cheap).**

---

## Q2 — review-to-rebuttal, and which reviewer points went unanswered

**Tag: new behavior (deferred).** Written up below and as PLAN.md §3(e); not implemented.

### The directionality question, settled from the spec alone

Faithfulness decomposes the **downstream text**, not the source. From spec/BEHAVIOR.md Mode 2:
`splitSummary(input)` splits the positional/`-text`/stdin input into claims; the `-source` file is
loaded separately and is never decomposed. Each downstream claim is then graded against the source
by `faithClaim` (defender finds source quotes; critic assigns the verdict). The verdict `absent`
means *this downstream claim is not found in the source* (spec/BEHAVIOR.md, Example D).

So for the canonical invocation:

```
crossexam -source review.txt rebuttal.txt
```

`-source` = `review.txt`; the positional arg `rebuttal.txt` is what gets decomposed. Every verdict —
including `absent` — is keyed to a **rebuttal** claim, graded against the review. The run answers
"did the rebuttal faithfully represent the review, or overstate / fabricate against it?" It does
**not** enumerate the review's points, so it cannot report which review points the rebuttal failed
to address.

**`absent` is keyed only to downstream claims.** Therefore "which reviewer points went unanswered" is
new behavior, not a projection. Per the brief, it is a design-queue item, not implemented this
session.

### Why it is genuinely new (and why the obvious workaround is not a clean fit)

The unanswered-point question needs the opposite traversal: decompose the **source** (the review)
into points and report, for each source point, whether the downstream (the rebuttal) covers it.
`crossexam` has no such traversal — faithfulness only ever decomposes the downstream side.

One could try to fake it by swapping the arguments:

```
crossexam -source rebuttal.txt review.txt    # treat the rebuttal as the "source"
```

Now each *review* point is graded against the rebuttal, and `absent` would mean "this review point is
not asserted in the rebuttal" — i.e. unanswered. But this misuses what `-source` means. The
faithfulness prompt is written for a *summary representing an original*: its distortion taxonomy is
fabrication / overstatement / context-stripping / misattribution / cherry-pick / literalization
(spec/PROMPTS.md §5). Those modes describe how a summary distorts a source; they do not map onto "the
rebuttal answered / overstated / left-unanswered a reviewer's point." The `absent` verdict would
carry over, but the rest of the machine would be reporting the wrong thing. The swap is a misread of
the source/summary roles, not a feature.

### The design-queue item (stated plainly)

The open directionality question, for PLAN.md §3(e):

> Should faithfulness be able to decompose the **source** and report coverage of each source point by
> the downstream text — the reverse of its current downstream-only traversal? Today
> `crossexam -source review.txt rebuttal.txt` grades each rebuttal claim against the review and keys
> `absent` to rebuttal claims; it cannot say which review points the rebuttal left unanswered.
> Surfacing unanswered source points requires a new source-decomposition-and-coverage traversal, plus
> a verdict vocabulary for a source point (answered / partially-answered / overstated-in-response /
> unanswered) distinct from the summary-distortion taxonomy. This is a new mode, not a flag on the
> existing one. Test (for whoever builds it): a review with N enumerable points and a rebuttal that
> answers some, overstates one, and silently drops one — the mode must name the dropped point as
> unanswered and the inflated one as overstated.

No example pair is drafted, because the brief draws the example only for the branch where `absent`
is already keyed to source points — which the spec shows it is not.

---

## What the frozen spec does not settle

- **Cross-mode agreement is not a spec concept.** `spec/` defines `vcolor` for *display colour*
  only; it never defines when the three columns "agree" or "disagree". The PASS/WEAK/FAIL tiering in
  Q1(a) reuses the colour routing as the disagreement criterion. That reuse is a design decision made
  and pre-registered here (d029), not something the spec dictates. A reviewer who wanted a stricter
  or looser rule (e.g. count WEAK-vs-PASS as disagreement) would be making a different, equally
  spec-unsupported choice.
- **The handling of an `error`/`skipped` column in the projection is spec-silent.** The spec says
  `error` never counts as a win (spec/BEHAVIOR.md, Error verdict rule) but says nothing about a
  cross-column projection. Q1 treats `error` and `skipped (over cap)` as NULL — never half of a
  conflict. This is a chosen default, flagged here, not a spec rule.
- **The exact question wording is authored, not specified.** The clause table in Q1(b) is a writing
  choice. The spec fixes the verdicts, not the prose that renders them.

None of these block the proposal: Q1's rule is a stated, pre-registered design choice over data the
spec fully defines, and Q2 is decisively answered by the spec (downstream-keyed) and deferred. There
is no point at which a required input was missing and an answer had to be invented.
