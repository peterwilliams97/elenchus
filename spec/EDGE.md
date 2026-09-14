# spec/EDGE.md — the edge-level adversarial pass

> **Scope note — spec only, no code yet.** This spec defines a format extension, one model call, one
> in-code admission check, and a rollup rule. Nothing here is implemented. It is the sister of
> `spec/ARGUMENT.md` and reuses that spec's node model, verdict vocabulary and rollup lattice — read
> ARGUMENT.md first. Where ARGUMENT.md derives a recommendation's judgement bottom-up from the
> **faithfulness** of its finding leaves, this pass judges the **inference itself**: given a finding
> that stands, does the recommendation it supports actually follow from it. That question no existing
> pass asks (`docs/todo/adversary-design-2026-09-14.md` §4, §1: 46 F→R edges, 0 carry a verdict).

## What the edge pass is, and the one thing it cannot do

The report tree (`spec/TREE.md`) settles *does the finding say what its source says*. The argument
tree (`spec/ARGUMENT.md`) propagates that up. Neither touches the step from a finding to the
recommendation it licenses — the enthymeme the report leaves implicit. A finding can be perfectly
faithful and the recommendation still not follow: `F-STANCE` faithfully reports an **association**
("a communicated AI stance amplifies performance"); `R-STANCE` recommends an **intervention**
("establish a policy"). The gap between them is where this pass works.

The pass may do exactly two things to an edge: **refute** it (name a concrete defeater → the edge is
`open`) or **decline to** (`unchallenged`). It may **never certify** that the recommendation follows.
This is the axis boundary in `CLAUDE.md` § The axis boundary, applied to inference: reasoning can
kill an edge with a counter-world but can never confirm one — positive grounding of "R follows from
F" needs the truth-maker (an intervention estimate, a typicality count), which is a retrieval, not a
deduction. So `unchallenged` means **"no admitted defeater this pass"**, never **"the inference is
sound."** The schema (below) has no field in which the model can assert soundness.

**Scope of the pass.** Only **load-bearing, stated** finding→recommendation edges, on findings that
still stand. Out of scope, because they are already in Needs-you and there is no standing inference
to attack:

- `?` edges — the report never asserted the linkage (`spec/ARGUMENT.md` § Edges); a human question,
  not a defeasible inference.
- childless recommendations — `R-TRANSFORM` (dora), `R6`/`R10` (vic-lceic): no finding to reason from.
- edges whose finding leaf-derives to `fails` (`contradicted`/`absent`/`unsupported`) — the premise
  has collapsed, so "a world where F holds and R fails" has no F to hold. The recommendation already
  `fails` on the leaf; running the edge would spend a call to no effect.

## 1. Format — the `scheme=` tag

An in-scope F→R edge gains one field: `scheme=<value>` at the head of the finding line's note column
(the third `|` field), alongside the existing `?`/`[opinion]` markers and provenance. It sits on the
**child (finding) line**, exactly as the `?` edge-marker does, because it is a property of the edge
up to the parent recommendation, not of the finding's own content. Values, the five consultant-report
schemes (`docs/todo/adversary-design-2026-09-14.md` § Operational form):

| value       | inference pattern                              | the report tier it usually types |
|-------------|------------------------------------------------|-------------------|
| `practical` | goal G, means M → do M                         | every dora capability rec; every master-plan rule |
| `survey`    | N% of respondents say X → X (position-to-know) | quocirca findings |
| `example`   | this case has property P → the class does (example→general) | master-plan's rules-from-anecdotes |
| `trend`     | indicator moved → the thing it signals moved (sign/trend) | dora's "amplifies" findings, ai-index growth |
| `classification` | the definition used here carries the conclusion | master-plan REC9 (the case-ordering) |

The `scheme` names what the **inference from finding to recommendation** is, which is not always what
the finding **is**: dora's findings are `trend`-shaped associations, but each is fed into a
means→ends recommendation, so the **edge** is `practical`. The tag is authored once per edge when the
tree is built, like the load-bearing/`?` decision it sits beside.

Extended syntax on two real dora edges (`examples/dora-2026/argument.txt`), both `practical` because
dora's rec tier is uniformly means→ends — the tag added, content unchanged:

    R-PLAT  | Invest in your internal platform — treat it as a product and the strategic prerequisite for unlocking AI's organizational value.  | p64; remedies F-PLAT
        F-PLAT  | A quality internal platform amplifies AI's positive influence on organizational performance (negligible when platform quality is low, strong when high), at the cost of a small but credible increase in delivery instability.  | scheme=practical; §Quality internal platforms p62; §The strategic imperative p71

    R-STANCE  | Clarify and socialize your AI policies — establish a clear, communicated policy on permitted tools and usage.  | p63; remedies F-STANCE
        F-STANCE  | A clear and communicated AI stance amplifies AI's positive impact on individual effectiveness and organizational performance, and turns its neutral effect on friction beneficial.  | scheme=practical; §Clear and communicated AI stance p52

## 2. The edge call — input and schema-enforced output

One model call per in-scope edge (repeated N times, § 5). It runs on the same `backend` seam as
`faithJudge` (`internal/backend`), with JSON-schema enforcement via `callSchema`.

**Input** (assembled in code, not authored):

- the **finding** text, and its **verified quote** — the `faithful`/`partial` evidence quote the
  report-tree pass already grounded and `quoteInPassage`-verified for this leaf. The edge reasons
  from what the source was shown to say, never from the finding's phrasing alone.
- the **recommendation** text.
- the edge's **scheme**, and that scheme's **fixed critical questions** — the short list from the
  table above, supplied verbatim in the prompt. The model asks the scheme's questions, not questions
  of its own.

**Output** — schema-enforced, exactly one of two shapes. There is no third "certified" shape:

    warrant            : string    // reconstructed W: "for R to follow from F you need W"
    defeater           : object | null
      world            : string    // a state of the world, plausible in the report's domain, in
                                    //   which the verified finding holds and the recommendation fails
      kind             : enum       // population | condition | definition | competing_goal
      anchor           : string     // the concrete referent the world names, VERBATIM within `world`
      settles          : string     // the source or observation that would decide it
      critical_question: enum       // which of the scheme's fixed CQs this defeater answers
    none_admitted      : bool       // true ⇔ defeater == null
    questions_considered: [enum]    // the scheme CQs examined; required when none_admitted

`defeater == null` with a populated `questions_considered` is the model saying "I asked the scheme's
questions and none opened a concrete world." That is a report of failure to refute — the axis
boundary forbids reading it as confirmation.

## 3. Admission check — in code, not in the prompt

The prompt asks the model to *produce* a defeater; whether it *counts* is decided by code, so a
model that emits "could be equivocating" for every edge (the free-fatal failure of the substance
critic, `docs/todo/destructive-sonnet-2026-09-13.md`) buys nothing. A returned `defeater` is
**admitted** only if all hold; the check is mechanical, no NLP:

1. `world` non-empty after trim, and `settles` non-empty after trim.
2. `kind` ∈ {`population`, `condition`, `definition`, `competing_goal`}.
3. `anchor` non-empty, and `anchor` is a **verbatim substring of `world`** — the same
   present-in-the-text discipline `quoteInPassage`/`groundVerdict` already run for evidence quotes
   (`assay.go`). The model must point at the concrete thing its world turns on *inside* the world it
   wrote; the code verifies the pointer lands.
4. **Template rule — cross-edge, applied after 1–3 across the whole report.** If an admitted
   defeater's `anchor` (case-insensitive) recurs in admitted defeaters on more than one edge of the
   same report, the defeater is **method-level**, not edge-level: a world that defeats every edge is a
   property of the report's evidence type, not of any one inference, and per-edge it is the free
   attack § 3 exists to catch. It is recorded once at the root (reason line `method: <world>`) and
   **removed from every edge**, which then count as `none-admitted` for that sample. dora's
   correlation-as-cause is the expected case — one `population`/`condition` world defeats each
   associational edge, so it belongs at the root once, not on all 8.

The concreteness the note demands ("a population, a condition, a definition, a competing goal") is
carried by `kind` + `anchor`: the model must name the referent's kind from a closed set and quote the
referent, and the code confirms the quote sits in the world. No sentence-level judgement is made — a
world that gestures without naming ("it might not generalize") fails at step 3 because it has no
anchor to quote.

A returned defeater that **fails admission** is discarded and the edge counts as `none-admitted`
(→ `unchallenged`). The rejection is **logged** — edge id, failed step, and the raw returned
`world`/`anchor` — so a run states how many defeaters were *offered* and how many *admitted*. The
gap between those two numbers is the calibration signal (§ 5): a large gap means the model attacks
freely and the check is holding the line.

## 4. Rollup — how `open`/`unchallenged` combine with the leaf verdicts

The edge verdict is the modal admitted-defeater result over the N samples (§ 5):

- **`open`** — an admitted defeater stands. The inference is contested: F holds, yet a concrete world
  makes R fail.
- **`unchallenged`** — no admitted defeater (none returned, or all rejected).

These do **not** replace a recommendation's leaf-derived verdict from `spec/ARGUMENT.md` § Internal
judgement; they combine with it. A recommendation now has two independent inputs — the conjunction
over its finding leaves' **faithfulness** (unchanged), and the **edge** verdict on the F→R inference —
and its final judgement is the **worse** of the two under ARGUMENT.md's existing precedence:

    fails  >  open  >  weakened  >  holds

The edge maps in as: **`open` edge → contributes `open`**; **`unchallenged` edge → contributes
nothing** (holds-level, never raises the node). So an admitted defeater moves the recommendation to
at least `open` **regardless of leaf faithfulness** — a recommendation whose every leaf is
`faithful`/`settled` (leaf-derived `holds`) but whose edge is `open` becomes `open`. That case — "the
finding is true and the recommendation still doesn't follow" — is the entire gap this pass exists to
surface, and nothing in the leaf-faithfulness computation could ever reach it.

**On the reused name `open`.** ARGUMENT.md's recommendation lattice already has `open`, meaning "not
settled, the reader must open it," reached by a contested leaf or a `?` edge. An admitted defeater is
a **third reason** for the same class, not a new class — exactly as ARGUMENT.md's root tally already
splits `open` into distinct reasons (`?`-edge/childless vs contested finding). The recommendation's
reason line names which fired: `open (edge defeater: <world>)` vs `open (finding contested)` vs
`open (edge to root not stated)`. A recommendation can be `open` for more than one reason at once;
all are listed.

**Needs-you placement.** An edge-`open` recommendation **heads the Needs-you set** — printed above
the leaf-faithfulness tiers of `spec/TREE.md` (ahead of tier 0's contested leaves), because a
defeated inference on a standing finding is a different and higher-order defect than an unstable or
unfaithful leaf: the sources check out and the argument still has a hole. A method-level defeater
lifted to the root (§ 3, rule 4) renders **once**, above all the edge-`open` lines, as
`method: <world> — defeats every edge of this evidence type; settled against the report's evidence
type, not any one inference`. The **line is the defeater** — its `world` as the claim and `settles`
as what would decide it — e.g.

    R-PLAT — open: platform amplifies performance but raises delivery instability, so in a
      regulated org whose binding constraint is stability, "invest, it's the prerequisite" fails.
      Settles it: F-PLAT's instability effect size against that org's own weighting.

## 5. Refuter — pre-registered before any code is judged good

**Corpus.** dora (8 in-scope edges) and master-plan (15 in-scope edges — `REC1`…`REC12`, with
`REC9`'s three cases `F10`/`F11`/`F12` as three edges; `base` is `?`, out of scope). **N=5** per edge.
Model: the calibration model of record, `claude-haiku-4-5-20251001`, then re-run on the default
Sonnet for the verdict to trust (`CLAUDE.md` § Testing). Log to `testing/calibration_log.jsonl`
beside the existing destructive runs. **All lines below are registered before the run**, per
`CLAUDE.md` § Pre-register the attempt.

**(a) Expected admitted-defeater rate, per scheme** — the point of registering these is that the pass
should *discriminate*, opening the associational/anecdotal edges and sparing the near-analytic ones:

| scheme | edges in corpus | predicted admitted-`open` rate | why |
|---|---|---|---|
| `practical` (dora, all 8) | 8 | 3–5 of 8 | side-effect CQ (BATCH, PLAT name their own cost) + correlation-as-cause on the associational ones; not universal |
| `example` (master-plan REC3, REC7, REC9×3, REC11) | 6 | 4–6 of 6 | rules-from-N=1: "is the case typical" is almost always concretely nameable — the master-plan risk |
| `practical` (master-plan REC1, REC2, REC4, REC5, REC6, REC8, REC10, REC12) | 8 | 2–4 of 8 | mechanism-stated rules resist a concrete counter-world better than dora's associations |
| `classification` | 0 (REC9's ordering is an R→root edge, outside § 1's F→R scope) | — | not judged this pass; `survey`/`trend` are absent from this refuter — their corpora (quocirca, ai-index) are out of scope |

**(b) Positive control.** **`R-VC ← F-VC`** (dora, `practical`) is predicted **`unchallenged`**. The
rollback/revert safety-net inference is the most directly causal of the dora set — strong version
control recovers from AI's bad changes by mechanism, not by association — and I predict no concrete
world survives the admission check where strong version control holds yet building it fails to help
AI adoption. If **every** dora edge (or every master-plan edge) returns `open`, the pass has the
substance critic's false-attack bias and **fails** — an attacker that never spares an edge is not
discriminating, it is firing on structure. A single positive-control miss (R-VC opens, but its
admitted defeater survives a human read as a genuinely concrete world) is a **calibration note**, not
an automatic fail; R-VC opening on a defeater that a human judges non-concrete means the admission
check (§ 3) leaked and **fails**.

**(c) Stability.** An edge is **contested** if its edge verdict **flips** (`open` ↔ `unchallenged`)
across the 5 runs; its class is the modal verdict, flagged contested. Registered: contested-edge rate
below ~⅓ per corpus. A pass whose edges mostly flip is reporting sampling noise, not a defeater.

**Pass/fail, stated before the run:**

- **FAIL** if all 8 dora edges are `open`, or all 15 master-plan edges are `open` (false-attack bias).
- **FAIL** if `R-VC` is `open` and its admitted defeater does not survive human inspection as concrete.
- **FAIL** if the run is empty or partial — 0 edges judged, or fewer than the 8+15 in scope reached
  (`CLAUDE.md` § zero-output, § partial run).
- **PASS** only if: admitted-`open` rate < 100% in **both** corpora; the template rule (§ 3, rule 4)
  fired at least once on dora (its correlation-as-cause world is the expected case); `R-VC` is
  `unchallenged` after templates are removed (or its `open` defeater survives human inspection); and
  contested-edge rate is below ~⅓ per corpus.

**Calibration note, not a gate.** The offered-vs-admitted gap (§ 3) is read as a signal, never a
pass/fail line: the check is form-only, so both `offered`==`admitted` (rejected nothing) and a
nonzero rejection count are satisfiable by construction and settle nothing about whether the pass
discriminates. A human reads the gap alongside the per-scheme rates in (a); it is where the check's
grip is judged, not where the run is failed.

Layer 3 discipline holds (`TESTING.md`): this is **calibration read by a human**, never a CI gate. A
green refuter means "the edge attacker discriminated on this set, this run," never "the recommendation
follows."

## 6. Cost — N=5 on both corpora

23 in-scope edges (8 + 15) × N=5 = **115 edge calls** per model pass.

Baseline unit from the Sonnet N=3 figure of ~$10 per 100 claims → $10 / 100 / 3 ≈ **$0.033 / call**.

- Naive: 115 × $0.033 ≈ **$3.83**.
- Adjusted: an edge call's prompt carries finding + verified quote + recommendation + the scheme's CQ
  list, and its output is the structured-defeater object — ~1.3–1.5× a `faithJudge` call's tokens →
  **~$5.0–5.7** total for the Sonnet pass.
- Split: dora 8×5 = 40 calls ≈ **$1.3–2.0**; master-plan 15×5 = 75 calls ≈ **$2.5–3.8**.

The haiku calibration pass first (`claude-haiku-4-5-20251001`) is roughly an order of magnitude
cheaper and is where the pass/fail lines above are first read; the Sonnet figures are the numbers to
trust.
