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

**`practical`'s fixed CQs are `alt_means` and `side_effects` only.** The three that named the
recommendation's *goal* — `goal_held`, `goal_conflict`, `feasible` — are removed. In a report
addressed to a stated audience the goal is stipulated by the genre, so each of the three is
answerable for *any* practical recommendation by inventing an addressee with other goals or
constraints, and a world that changes the addressee is outside the report's domain (§ 3, step 5).
Run 2 (below) confirmed this concretely: all five of its opened edges opened on a `goal_conflict`
defeater naming a competing goal the report never addresses. The two that survive stay
report-internal by construction — `alt_means` names a cheaper or faster route to the same goal, and
`side_effects` names a cost the recommendation's own text already carries (F-PLAT's instability
clause, F-BATCH's cost clause).

**Output** — schema-enforced, exactly one of two shapes. There is no third "certified" shape:

    warrant            : string    // reconstructed W: "for R to follow from F you need W"
    defeater           : object | null
      world            : string    // a state of the world, plausible in the report's domain, in
                                    //   which the verified finding holds and the recommendation fails
      kind             : enum       // population | condition | definition
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
2. `kind` ∈ {`population`, `condition`, `definition`}.
3. `anchor` non-empty, and `anchor` is a **verbatim substring (via `norm`) of the finding text, the
   recommendation text, or the verified quotes** — **not** of the `world`. The same
   present-in-the-text discipline `quoteInPassage`/`groundVerdict` already run for evidence quotes
   (`assay.go`), but pointed at the report, not at the model's own prose. **Rationale:** a defeater
   the report itself names is checkable against the report; one the model imports from outside is a
   claim about the world, and confirming *that* is a retrieval, not a deduction — it belongs in
   grounding, past the axis boundary this pass may not cross (§ What the edge pass is). Anchoring in
   `world` (the pre-run rule) let the model manufacture its own referent and quote itself, which is
   why run 1 admitted a world the report never mentions (Refuter run 1, below). **Consequence:** v1
   admits only **report-internal** defeaters — a world built from a population, condition or
   definition the report's own text already names. Correlation-as-cause is admissible only on
   edges whose finding is worded associationally (dora's "amplifies"); where the report states a
   mechanism, the defeater must find its referent in that wording or fail admission.
4. **Template rule — cross-edge, applied after 1–3 across the whole report.** If an admitted
   defeater's `anchor` (case-insensitive) recurs in admitted defeaters on more than one edge of the
   same report, the defeater is **method-level**, not edge-level: a world that defeats every edge is a
   property of the report's evidence type, not of any one inference, and per-edge it is the free
   attack § 3 exists to catch. It is recorded once at the root (reason line `method: <world>`) and
   **removed from every edge**, which then count as `none-admitted` for that sample. dora's
   correlation-as-cause is the expected case — one `population`/`condition` world defeats each
   associational edge, so it belongs at the root once, not on all 8.
5. **Domain rule.** The `world` must hold **for the report's addressed audience as stated**. A
   defeater that varies the addressee — swaps in an organisation with other goals or a binding
   constraint the report never puts on its reader — is not a defeater of the inference: the report's
   genre stipulates its audience and the goal that audience holds, so a competing goal outside that
   stipulation is outside the report's domain. This is why `practical`'s three goal-naming CQs
   (`goal_held`, `goal_conflict`, `feasible`) are removed (§ 2): each is answerable only by inventing
   an addressee, and the world it produces fails this rule. Run 2's five opened edges (below) all
   failed it — a `goal_conflict` world (regulated org, rollback forbidden) that swaps the addressee —
   yet each passed steps 1–3 because its `anchor` was quoted from the finding, so the rule catches
   what anchoring cannot.

The concreteness the note demands ("a population, a condition, a definition") is
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

**Run 2 — re-registered after the step-3 fix (14 Sept).** Same corpus, same model of record, same
N=5, and **the same pass/fail lines above still bind** — run 2 is judged against them unchanged. The
one prediction the amended anchor rule (§ 3, step 3: anchor in the report, not in `world`) changes:
**admitted-`open` rate drops sharply** from run 1's 4/7, because the imported compliance/returns goal
that opened those edges now fails admission (it is nowhere in dora's text), and no other report-named
defeater is expected on the causal edges. **`R-VC` is again predicted `unchallenged`** — this time
for the recorded reason that run 1's opening world was report-external and step 3 now rejects it.
Registered before the run, per `CLAUDE.md` § Pre-register the attempt.

**Run 3 — re-registered after the CQ reduction + domain rule (14 Sept).** Same corpus, model of
record and N=5, against `practical`'s reduced CQ set (`alt_means`, `side_effects` only; § 2) and the
§ 3 step-5 domain rule. **The same pass/fail lines above still bind.** Run 2's opened edges all opened
on `goal_conflict`, which is now neither an offered CQ nor an admissible world, so:

- **admitted-`open` ≤ 2/7.** With the three addressee-varying CQs gone, no report-external competing
  goal can be offered, and the domain rule rejects any that reach the world anyway.
- **if any edge opens, it is `R-BATCH` on `side_effects`,** with an `anchor` in `F-BATCH`'s own cost
  clause — the one report-internal cost a practical edge names.
- **`R-VC` is `unchallenged`** — its run-2 opening world (swapped addressee) is exactly what step 5
  rejects.
- **offered drops well below run 2's 24** — two CQs where five were on offer.

All other FAIL lines (empty/partial run, false-attack bias, contested-rate ceiling) apply unchanged.
Registered before the run, per `CLAUDE.md` § Pre-register the attempt.

**Calibration note, not a gate.** The offered-vs-admitted gap (§ 3) is read as a signal, never a
pass/fail line: the check is form-only, so both `offered`==`admitted` (rejected nothing) and a
nonzero rejection count are satisfiable by construction and settle nothing about whether the pass
discriminates. A human reads the gap alongside the per-scheme rates in (a); it is where the check's
grip is judged, not where the run is failed.

Layer 3 discipline holds (`TESTING.md`): this is **calibration read by a human**, never a CI gate. A
green refuter means "the edge attacker discriminated on this set, this run," never "the recommendation
follows."

## Refuter run 1 — haiku N=5 dora, 14 Sept: FAIL (recorded negative)

First execution of the § 5 refuter on dora, `claude-haiku-4-5-20251001`, N=5. Chain:
`testing/chains/edge-20260914-2050-*`. **Verdict: FAIL** — kept here as a recorded negative, per
`CLAUDE.md` § Pre-register the attempt (log dead ends, not only survivors).

**Results.** 7 in-scope edges (R-PLAT dropped — its finding leaf-derives to `fails`, § Scope). Edge
verdicts: **open 4/7**. Admission: **25 defeaters offered, 21 admitted, 4 rejected** (all four
rejections at step 3, no anchor). The **template rule (§ 3, rule 4) did not fire** — no `anchor`
recurred verbatim across edges, so no world lifted to the root. Positive control **`R-VC` open 3/5**
(predicted `unchallenged`, § 5b). Stability: **contested 6/7** (verdict flips across the 5 runs — far
above the ~⅓ ceiling of § 5c).

**Diagnosis — one defeater wearing 21 hats.** All four opened edges (and R-VC's three opening samples)
open on the **same** world: a `competing_goal` — a compliance mandate or a shareholder-returns
obligation — that **the report never mentions**, imported as the goal G' that displaces the
recommendation's means M. It is `goal_conflict` applied to every practical edge, exactly the
method-level attack § 3 rule 4 exists to lift to the root once. Two things let it through per-edge
instead:

1. **The anchor rule pointed at the model's own prose.** Step 3 (pre-run) admitted an `anchor` that
   was a substring of `world` — but the model *wrote* `world`, so it manufactured its own referent
   ("compliance mandate") and quoted itself. The check verified a pointer into the model's sentence,
   not into the report, so a world the report never raises passed as concrete.
2. **The template rule could not catch it.** Each sample phrased the same goal differently
   ("regulatory compliance", "shareholder returns", "audit obligations"), so no `anchor` string
   recurred verbatim and rule 4 never fired. The single method-level world scattered into per-edge
   admissions instead of collecting at the root.

R-VC opening is the § 5b hard-fail condition: its admitted defeater is this same imported goal, which
does **not** survive human inspection as concrete (the report names no such constraint), so **admission
leaked** — the check passed a world it should have rejected. That is the fail, not merely the flip
count.

The fix is the amended step 3 above: anchor in the **report** (finding / recommendation / verified
quote), not in `world`. A defeater the report itself names stays admissible; the imported
compliance/returns goal now fails admission because "compliance mandate" appears nowhere in dora's
text. Re-registered as run 2 in § 5.

## Refuter run 2 — haiku N=5 dora, 14 Sept 21:13: FAIL (recorded negative)

Second execution of the § 5 refuter on dora, `claude-haiku-4-5-20251001`, N=5, against the amended
step 3 (anchor in the report, not in `world`). Chain: `testing/chains/edge-20260914-2113-haiku`.
**Verdict: FAIL** — kept as a recorded negative per `CLAUDE.md` § Pre-register the attempt. The
earlier `testing/chains/edge-20260914-2106-*` chain is **void**: it ran a stale binary and is a
byte-repeat of run 1, not a run of the amended check — do not read it as run 2.

**Results.** 7/7 in-scope edges reached. Edge verdicts: **open 5/7**. Admission: **24 defeaters
offered, 23 admitted, 1 rejected**. The **template rule (§ 3, rule 4) did not fire**. Positive
control **`R-VC` open 4/5** (predicted `unchallenged`, § 5b); on inspection its world **swaps the
addressee** — a regulated org for which rollback is forbidden — which is not concrete in the report's
domain. Stability: **contested 4/7** (verdict flips across the 5 runs — above the ~⅓ ceiling of
§ 5c).

**Diagnosis — the anchor rule held and was irrelevant.** The step-3 fix did its job: every admitted
defeater's `anchor` was quoted from the finding or recommendation, so no report-external anchor
leaked, and the imported compliance/returns world of run 1 is gone. The five opened edges are all
`critical_question=goal_conflict` — the `anchor` is report-internal, but the `world` still names an
**external competing goal** the report never addresses. The anchor rule constrains where the quote
comes from; it does nothing about **which CQ** the model is allowed to ask, and the free move is the
CQ, not the anchor. `goal_held`, `goal_conflict` and `feasible` are answerable for **any** practical
recommendation by inventing an addressee with other goals or constraints; in a report addressed to a
stated audience the goal is stipulated by the genre, and a world that changes the addressee is
outside the report's domain. `R-VC`'s opening world is exactly this — a swapped addressee — which is
the § 5b hard-fail condition (its admitted defeater does not survive human inspection as concrete).

The fix is § 2's CQ reduction (`practical` → `alt_means`, `side_effects` only) and § 3's step 5
(domain rule), both above. Re-registered as run 3 in § 5.

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
