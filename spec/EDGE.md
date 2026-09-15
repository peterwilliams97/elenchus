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

**`practical`'s fixed CQ is `side_effects` only.** The logic: for practical reasoning "M goes with
G → do M" is undercut, from inside the report, only by the report naming a cost of M. Goal,
feasibility and alternatives are the reader's; correlation→intervention is the method, not any one
edge's defect, and lives at the root once (§ 3, rule 4). A gap between the measured variable and the
recommended action is real, but it is a Needs-you question, not a model defeater — the model cannot
be held to it (run 4, below). The one CQ, with its anchor:

- **`side_effects`** — the finding or its verified quotes name a cost of M. Anchor = the cost clause,
  verbatim in the finding or quotes (F-PLAT's instability clause, F-BATCH's cost clause).

The three CQs that named the recommendation's *goal* — `goal_held`, `goal_conflict`, `feasible` — are
removed, and so is `alt_means`. The goal-naming three are answerable for *any* practical
recommendation by inventing an addressee with other goals or constraints, and a world that changes the
addressee is outside the report's domain (§ 3, step 5); run 2 (below) confirmed this — all five of its
opened edges opened on a `goal_conflict` defeater naming a competing goal the report never addresses.
`alt_means` is removed for a different reason (run 3, below): a cheaper route M' does not defeat "do
M", only "do M *rather than* M'", which the recommendation does not claim. `means_mismatch` is removed
after run 4 (below): its four admitted worlds were all "M done nominally but not realised" or "M
infeasible here" — implementation and feasibility, which are the reader's, not an inference the model
can settle — and one anchored on a *different* finding's text, the leak § 3 step 3 now closes.

**`example`'s fixed CQs are `named_exception` and `scope_dropped`.** The logic mirrors
`practical`'s: "case C shows property P → rule: always P" is undercut, from inside the report, only
by the report itself naming where P failed or bounding where C applies. Reasoning alone cannot
settle whether one case generalises — that needs a typicality count, a retrieval past the axis
boundary (§ What the edge pass is) — so the pass attacks only with a limit the report already
states. The two CQs, each with its anchor:

- **`named_exception`** — the finding or its verified quotes themselves name a case, condition or
  caveat where the pattern did **not** hold, and the recommendation states the rule as if it always
  does. Anchor = that exception clause, verbatim in **this edge's own** finding or its quotes.
- **`scope_dropped`** — the finding or its verified quotes state the scope of the case ("in this
  parser", "on Windows builds", "for our Go code") and the recommendation states the rule without
  it. Anchor = the scope phrase, verbatim in **this edge's own** finding or its quotes.

The two CQs that asked after the case's *typicality* — `typical_case` ("is the cited case typical
of the class the rule generalises to") and `counter_cases` ("how many counter-cases were looked for
before generalising") — are removed, for the reason `practical`'s goal-naming CQs went (runs 2–5,
§ 5): both are answerable for **any** anecdote by asserting a wider world the report never claims —
a class the case is or is not typical of, a search for counter-cases the report never ran — so the
report cannot be held to them. What survives is the report-internal move: the report flags its own
exception, or bounds its own scope, and the recommendation generalises past it. Both are checkable
against the report by the § 3 step-3 anchor rule `practical` earned, and nothing else opens an
`example` edge.

**Output** — schema-enforced, exactly one of two shapes. There is no third "certified" shape:

    warrant            : string    // reconstructed W: "for R to follow from F you need W"
    defeater           : object | null
      world            : string     // a state of the world, plausible in the report's domain, in
                                    // which the verified finding holds and the recommendation fails
      kind             : enum       // population | condition | definition
      anchor           : string     // the concrete referent (the cost clause), VERBATIM in this
                                    //   edge's own finding or verified quotes (§ 3, step 3)
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
3. **Anchor check.** `anchor` non-empty and a **verbatim substring (via `norm`) of this edge's own
   finding text or its verified quotes** — **not** any other finding's text, **not** the recommendation
   text, and **not** the `world`. With `practical` reduced to `side_effects` (§ 2), the anchor is the
   cost clause the finding names, verbatim in that finding or its quotes. Confining it to *this* edge's
   finding closes the run-4 leak (below): `F-ACCESS`'s admitted anchor was quoted from `F-DATA`'s text
   — a string present in the report but not in the premise under attack.

   This is the present-in-the-text discipline `quoteInPassage`/`groundVerdict` already run for evidence
   quotes (`assay.go`), pointed at the report, not at the model's own prose. **Rationale:** a defeater
   the report itself names is checkable against the report; one the model imports from outside is a
   claim about the world, and confirming *that* is a retrieval, not a deduction — past the axis
   boundary this pass may not cross (§ What the edge pass is). Anchoring in `world` (the run-1 rule)
   let the model manufacture its own referent and quote itself (Refuter run 1, below); anchoring in
   the recommendation as well (the run-2 rule) let the anchor sit in R while the world named a goal R
   never states (run 3, below). Confining `anchor` to *this edge's* finding or its quotes — the premise
   under attack — forces the referent onto the thing being defeated; run 4 (below) added the *this-edge*
   restriction after `F-ACCESS` anchored on `F-DATA`'s text. **Consequence:** the pass admits only
   **report-internal**
   defeaters — a world built from a population, condition or definition the report's own text names.
4. **Template rule — cross-edge, applied after 1–3 across the whole report.** If an admitted
   defeater's `anchor` (case-insensitive) recurs in admitted defeaters on more than one edge of the
   same report, the defeater is **method-level**, not edge-level: a world that defeats every edge is a
   property of the report's evidence type, not of any one inference, and per-edge it is the free
   attack § 3 exists to catch. It is recorded once at the root (reason line `method: <world>`) and
   **removed from every edge**, which then count as `none-admitted` for that sample. The rule stays as
   a **backstop** — a defeater recurring verbatim across edges is method-level by construction — but
   with `practical`'s CQ set reduced to `side_effects` (§ 2) the model no longer offers the
   correlation-as-cause world per-edge, so the rule is no longer expected to fire on dora. dora's
   method-level point — findings are associational, recommendations are interventions — is instead
   emitted at the root as a fixed note by the pass (§ 5), not produced by the model. **There is no
   fixed root method note for `example` trees.** master-plan mixes `practical` and `example` edges
   and no single world defeats them all — a `named_exception` clause bounds the one finding that
   states it, not the report's evidence type — so the master-plan pass emits no root note; the
   template rule stays as its backstop for any `anchor` that does recur across edges.
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

**Corpus.** dora (8 in-scope edges) and master-plan (14 in-scope edges of 15 — `REC1`…`REC12`, with
`REC9`'s three cases `F10`/`F11`/`F12` as three edges; `REC9`'s case-ordering is `classification`, an
R→root edge outside § 1's F→R scope (§ 5(a)); `base` is `?`, out of scope). **N=5** per edge.
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
- **PASS** only if: admitted-`open` rate < 100% in **both** corpora; the fixed root method note for
  dora — "findings are associational; recommendations are interventions" — is emitted by the pass, not
  produced by the model; `R-VC` is `unchallenged` (or its `open` defeater survives human inspection);
  and contested-edge rate is below ~⅓ per corpus.

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

**Run 4 — re-registered after the CQ reduction to `side_effects`/`means_mismatch` + the fixed root
method note (15 Sept).** Same corpus, model of record and N=5, against `practical`'s two CQs (§ 2),
the § 3 per-CQ step-3 anchor rule, and the fixed root method note replacing the template-fired PASS
condition. **The same pass/fail lines above still bind.** Run 3 admitted 18 defeaters on
`alt_means`/`side_effects` with none rejected, but `alt_means` defeated "do M *rather than* M'" — a
comparative no recommendation makes — so those admissions were spurious (run 3 diagnosis, below). With
`alt_means` gone:

- **admitted-`open` 0–1/7.** No goal- or alternatives-based CQ remains; the only report-internal
  defeater a `practical` edge can raise is a named cost.
- **if any edge opens, it is `R-BATCH` on `side_effects`,** with an `anchor` in `F-BATCH`'s own cost
  clause — the one dora finding whose verified quote names a cost of its own means.
- **`R-VC` is `unchallenged`** — `F-VC` names no cost, and no goal or `alt_means` CQ remains to open it.
- **offered < 10** — two CQs, one of which (`means_mismatch`) fires only where finding and
  recommendation name different variables, which dora's edges do not.

All other FAIL lines (empty/partial run, false-attack bias, contested-rate ceiling) apply unchanged.
Registered before the run, per `CLAUDE.md` § Pre-register the attempt.

**Run 5 — re-registered after the CQ reduction to `side_effects` only + the this-edge anchor
confinement (15 Sept).** Same corpus, model of record and N=5, against `practical`'s single CQ (§ 2)
and § 3 step 3's confinement of `anchor` to the edge's own finding or quotes. **The same pass/fail
lines above still bind.** Run 4 admitted four `means_mismatch` worlds that were implementation or
feasibility, not inference defects, and one anchored on a different finding's text (run 4 diagnosis,
below). With `means_mismatch` gone and the anchor confined:

- **admitted-`open` 0–1/7.** `side_effects` is the only CQ, and it fires only where a finding's own
  verified quotes name a cost of its means.
- **if any edge opens, it is `R-BATCH` on `side_effects`,** and only if `F-BATCH`'s verified quotes
  name that cost.
- **`R-VC` is `unchallenged`** — `F-VC` names no cost, and no other CQ remains to open it.
- **offered < 8** — one CQ, with nothing to offer on a finding whose quotes name no cost.

All other FAIL lines (empty/partial run, false-attack bias, contested-rate ceiling) apply unchanged.
Registered before the run, per `CLAUDE.md` § Pre-register the attempt.

**Reading rule for run 5 and after.** With the one remaining CQ fully report-anchored — its `anchor` a
cost clause verbatim in the edge's own finding — an admitted defeater is a **finding to be read, not a
leak**: the report names the cost, the world is checkable against the report, and a human reads it as a
genuine hole in that recommendation. If haiku still opens edges on invented costs — a `side_effects`
world whose cost is nowhere in the finding's quotes — the admission check has held and the model has
not, so the pass moves to Sonnet unchanged and haiku is recorded as unable to hold the anchor
discipline (a model-fitness finding, not a spec change).

**Master-plan run — registered after dora's run-5 PASS, before any master-plan edge is judged
(15 Sept).** Corpus `examples/master-plan/argument.txt`, N=5 on `claude-haiku-4-5-20251001` first,
then the default Sonnet for the verdict to trust (`CLAUDE.md` § Testing). 15 F→R edges: **6
`example`** (`REC3←F4`, `REC7←F8`, `REC9←F10`, `REC9←F11`, `REC9←F12`, `REC11←F14`) and **8
`practical`** (`REC1`, `REC2`, `REC4`, `REC5`, `REC6`, `REC8`, `REC10`, `REC12`) in scope, **1
`classification`** (`REC9`'s case-ordering, R→root) **out** — the F→R pass runs `practical` and
`example` only (§ 1, § 5(a)). The dora pass/fail lines above bind, re-read for this corpus.
Registered before the run, per `CLAUDE.md` § Pre-register the attempt.

> **Retraction of the "in scope" premise (15 Sept, after the run — § Master-plan run below).** "6
> `example` + 8 `practical` in scope" assumed all 15 findings stand; they do not. master-plan has **15
> F→R edges but only 4 on held findings** — `F3`, `F9`, `F13`, `F15`, all `practical`. The other 11
> (all 6 `example`, plus practical `F1`/`F2`/`F5`/`F6`/`F7`) leaf-derive to `fails` and drop by
> § Scope, because master-plan's findings are the plan's own prescriptions with no held source
> (`route=method` leaves). The `example` predictions below — open ≤ 3/15, the `REC3 ← F4` positive
> control, the `named_exception`/`scope_dropped` anchors — are therefore **unreachable on this
> corpus**; only the `practical` prediction was testable, and it held on the 4 edges that stand.

Predictions:

- **open ≤ 3/15.** With `example` reduced to `named_exception`/`scope_dropped` (§ 2) and `practical`
  to `side_effects`, an edge opens only on a limit its own finding states, and most master-plan
  findings state their rule without one.
- **Any open `example` edge anchors on a scope phrase or exception clause its own finding contains** —
  `F8`'s "no number has been computed", `F10`'s "with stated caveats", `F11`'s "found nothing
  itself", `F12`'s and `F14`'s stated conditions are the report-internal limits available; nothing
  else opens an `example` edge.
- **Any open `practical` edge opens only on a named cost**, exactly as on dora (`R-BATCH`, run 5) —
  the single `side_effects` anchor.
- **Positive control: `REC3 ← F4` is predicted `unchallenged`.** Its finding — "Shared code hides
  bugs by construction — a shared toInt32 passed 10,000 checks with the bug present and again with
  the fix reverted" — states the rule with **no scope phrase and no exception clause**, so neither
  `example` CQ has an anchor to quote and the § 3 step-3 admission check leaves the edge un-opened.
  It is the one `example` finding stated without a bound.

**Pass/fail for the master-plan run** (the dora lines bind unchanged; these are the corpus-specific
adds):

- **FAIL** if **every** `example` edge opens (6/6 — false-attack bias on the reachable scheme), or if
  any admitted `anchor` is **not** verbatim in its own edge's finding or verified quotes (the run-4
  cross-finding leak, § 3 step 3).
- **FAIL** on an empty or partial run — fewer than the 14 in-scope edges reached (`CLAUDE.md`
  § zero-output, § partial run).
- **PASS** only if: open ≤ 3/15; `REC3 ← F4` is `unchallenged` (or its `open` defeater survives human
  inspection as a report-internal bound); every admitted anchor sits in its own edge's finding or
  quotes; and **contested-edge rate is below ⅓ on the Sonnet pass** (§ 5c).

`argument.txt` currently tags every F→R finding `practical` or `example`; the `classification`
re-tag of `REC9`'s ordering is tree work, outside this spec change.

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

## Refuter run 3 — haiku N=5 dora, 14 Sept 22:14: FAIL (recorded negative)

Third execution of the § 5 refuter on dora, `claude-haiku-4-5-20251001`, N=5, against run 3's
`practical` CQ set (`alt_means`, `side_effects`) and the § 3 step-5 domain rule. Chain:
`testing/chains/edge-20260914-2214-haiku`. **Verdict: FAIL** — kept as a recorded negative per
`CLAUDE.md` § Pre-register the attempt.

**Results.** 7/7 in-scope edges reached. Edge verdicts: **open 4/7**. Admission: **18 defeaters
offered, 18 admitted, 0 rejected**; **none_admitted 17** (up from run 2's 11). The **template rule
(§ 3, rule 4) did not fire**. Positive control **`R-VC` open 4/5** on `alt_means` ("code review is the
operative lever"; predicted `unchallenged`, § 5b). Stability: **contested 3/7** (below the ~⅓ ceiling
of § 5c). Admitted by CQ: `alt_means` 3, `side_effects` 1.

**Diagnosis — the surviving CQs still attack the wrong thing.** Two defects, both in what a CQ is
allowed to claim:

1. **`alt_means` does not defeat "do M".** It defeats only "do M *rather than* M'" — a comparative the
   recommendation never makes. A cheaper or faster route to the same goal leaves "M goes with G, so do
   M" standing; naming M' is a reason to also weigh M', not a defeater of M. `R-VC` opened 4/5 on
   exactly this ("code review is the operative lever") — a rival means, not a defeated inference.
2. **The one `side_effects` defeater anchored on the means, not on the cost.** Its `anchor` was
   verbatim in the finding (passing step 3), but it named M rather than a cost of M, so the cost in the
   `world` was invented — unanchored in fact while the anchor pointed elsewhere.

The common root: **the anchor rule has been checking the target of the attack, not the attack.** Step
3 confirmed the referent sat in the report, but not that it was the *cost* (`side_effects`) or that the
recommendation made the comparative `alt_means` needs. The fix is § 2's reduction to `side_effects` +
`means_mismatch` and § 3's per-CQ step 3: `alt_means` removed outright, `side_effects` anchored on the
cost clause specifically, and `means_mismatch` given two anchors so the attack lands on the F→R gap and
nothing else. Re-registered as run 4 in § 5.

## Refuter run 4 — haiku N=5 dora, 15 Sept 09:20: FAIL (recorded negative)

Fourth execution of the § 5 refuter on dora, `claude-haiku-4-5-20251001`, N=5, against `practical`'s
`side_effects`/`means_mismatch` CQ set (run-4 registration, § 5), the per-CQ step-3 anchor rule and
the fixed root method note. Chain: `testing/chains/edge-20260915-0920-haiku`. **Verdict: FAIL** — kept
as a recorded negative per `CLAUDE.md` § Pre-register the attempt.

**Results.** 7/7 in-scope edges reached. Edge verdicts: **open 4/7**. Admission: **18 defeaters
offered, 18 admitted, 0 rejected**, with **16 samples `none_admitted`**; every admitted defeater is on
**`means_mismatch`** (`side_effects` admitted none), and the four opened edges all open on
`means_mismatch`. Positive control **`R-VC` unchallenged** — the first run in which it holds, the
§ 5b prediction met at last — and **`R-DATA` unchallenged**. Stability: **contested 4/7** (verdict
flips across the 5 runs — above the ~⅓ ceiling of § 5c).

**Diagnosis — `means_mismatch` opens on implementation, not inference; the anchor spans findings.**
Two defects, both fatal to `means_mismatch` as a CQ:

1. **The four opened worlds are not inference defects.** Each is either "M done nominally but not
   realised" or "M infeasible here" — the recommendation's means carried out in name only, or
   unbuildable in the reader's setting. Both are implementation and feasibility questions, which § 2
   already assigns to the reader; neither is a world in which the finding holds and the *inference*
   from it fails. `means_mismatch` licensed the model to attack the doing of M, not the following of R
   from F, so its admissions are Needs-you questions mis-filed as edge defeaters.
2. **The anchor spanned other findings.** `F-ACCESS`'s admitted `anchor` was a substring of `F-DATA`'s
   text, not of `F-ACCESS`'s own finding or quotes — a string present in the report but outside the
   premise under attack. The pre-run step-3 rule checked the anchor against "the finding text or the
   verified quotes" without pinning *which* finding, so a cross-finding quote passed as report-internal.

`R-VC` and `R-DATA` holding is the one positive sign: with the correlation-as-cause and goal worlds
gone, the causal edges no longer open. But open 4/7 on `means_mismatch` and contested 4/7 both breach
the § 5 lines, and the two defects above are the cause.

The fix is § 2's further CQ reduction (`practical` → `side_effects` only) and § 3 step 3's confinement
of `anchor` to *this* edge's own finding or quotes, both above. Re-registered as run 5 in § 5.

## Refuter run 5 — sonnet N=5 dora, 15 Sept 12:28: PASS

Fifth execution of the § 5 refuter on dora, run on the verdict model `claude-sonnet-4-6`, N=5,
against `practical`'s single `side_effects` CQ (§ 2) and § 3 step 3's confinement of `anchor` to the
edge's own finding or quotes. Chain: `testing/chains/edge-20260915-1228-sonnet`. **Verdict: PASS** —
the first run to clear every § 5 line.

**Results — per edge (§ 1):**

    F-STANCE → R-STANCE  [practical]  modal=unchallenged  final=unchallenged  0o/5u/0e  flips=0/5
    F-DATA   → R-DATA    [practical]  modal=unchallenged  final=unchallenged  0o/5u/0e  flips=0/5
    F-ACCESS → R-ACCESS  [practical]  modal=unchallenged  final=unchallenged  0o/5u/0e  flips=0/5
    F-VC     → R-VC      [practical]  modal=unchallenged  final=unchallenged  0o/5u/0e  flips=0/5
    F-BATCH  → R-BATCH   [practical]  modal=open          final=open          3o/2u/0e  flips=2/5  CONTESTED
       anchor: 'observed gains in individual effectiveness would be somewhat less'   cq: side_effects
    F-USER   → R-USER    [practical]  modal=unchallenged  final=unchallenged  0o/5u/0e  flips=0/5
    F-VSM    → R-VSM      [practical]  modal=unchallenged  final=unchallenged  0o/5u/0e  flips=0/5

**Totals (§ 2).** samples=35, none_admitted=32, offered=3, admitted=3, rejected=0; admitted by
critical_question: `side_effects` (F-BATCH only). Template rule (§ 3, rule 4) **did not fire** — one
distinct admitted anchor, carried by F-BATCH alone, so nothing lifts to the root. Positive control
**`R-VC` unchallenged** (§ 5b prediction met). Fixed root method note (all edges `practical`)
emitted (§ 5).

**Against the registration — PASS on every line.** `R-BATCH` opens on the same cost clause run 5
named — `side_effects`, anchor `'observed gains in individual effectiveness would be somewhat less'`,
verbatim in F-BATCH's own verified quotes. `R-ACCESS` and `R-VC` are **unchallenged**. open **1/7**
(below the false-attack ceiling), contested **1/7 = 14% (< ⅓**, § 5c), 7/7 in-scope edges reached
(no empty/partial run).

**Reading `R-BATCH` (run-5 reading rule).** The `practical` edge pass is calibrated on the model of
record as of this run — `claude-sonnet-4-6`. `R-BATCH`'s world is that rule's exemplar: the report
itself names the cost of small batches and resolves it by assertion, and the edge pass returns that
weighing to the reader rather than certifying it — an admitted defeater to be read as a genuine hole
in that one recommendation, not a leak in the check.

## Master-plan run — haiku N=5, 15 Sept 14:47: NO-RESULT for `example` (recorded)

First execution of the § 5 master-plan refuter, `claude-haiku-4-5-20251001`, N=5. Chain:
`testing/chains/edge-20260915-1447-haiku-mp`. **Verdict: NO-RESULT for `example`** — the run reached
none of the scheme master-plan was chosen to exercise, so it neither passes nor fails the `example`
lines; the 4 `practical` edges that stand cleared the dora § 5 lines re-read for this corpus. Kept as a
recorded negative per `CLAUDE.md` § Pre-register the attempt (log dead ends, not only survivors).

**Results.** **4 of 15** scheme-tagged F→R edges reached — `F3→REC2`, `F9→REC8`, `F13→REC10`,
`F15→REC12`, **all `practical`, all `unchallenged`, offered 0/20** (no defeater offered across the 20
samples). The other **11** dropped by § Scope, their findings leaf-deriving to `fails`: every `example`
edge (`REC3←F4`, `REC7←F8`, `REC9←F10`, `REC9←F11`, `REC9←F12`, `REC11←F14`) and five `practical`
ones (`F1`, `F2`, `F5`, `F6`, `F7`). master-plan's findings are the plan's own prescriptions with no
held source — `route=method` leaves — so they collapse to `fails` on the leaf and the edge has no F to
hold. `F3` is the one finding the tree annotates with a held truth-maker
(`refuter-runs-2026-08-15.md`); `F9`/`F13`/`F15` are the other three whose leaves do not collapse.
Re-read against the dora § 5 lines for the 4 that stand: open **0/4**, contested **0/4**, 4/4 in-scope
edges reached against the in-scope denominator of 4 — no empty/partial run.

**The applicability premise was wrong (retraction filed at the registration, § 5).** master-plan was
registered as the corpus that exercises `example`, on the premise that its 15 F→R edges all stand.
Only **4 are on held findings**, and every one of those 4 is `practical`, so the `example` scheme is
never reached. The registration's `example` predictions are unreachable here; its `practical`
prediction is met on the 4 edges that stand.

**The `example` CQs are spec'd and built but unexercised.** `named_exception` and `scope_dropped`
(§ 2) — the admission check, their per-CQ anchors, the code path — ship, but no corpus in the refuter
has a **held** `example` finding for them to attack. Calibrating them needs either a corpus with held
`example` findings, or a constructed probe under `examples/destructive/`: a `scope-dropped` case and a
`named-exception` case, each with a **held** finding stating the bound its recommendation generalises
past. **Deferred** — parked in `SESSION.md`.

**Fixed root method note — correctly withheld.** master-plan is mixed-scheme (§ 3 rule 4: no fixed
note for `example` trees), and it stays mixed-scheme even though every `example` edge leaf-failed out
of the pass. The note gate reads the full scheme roster, not the in-scope edges
(`allPractical(schemeEdges)`, `assay.go`), so the pass emits no root note on master-plan. Gating on the
in-scope edges alone would have wrongly emitted it once the `example` edges dropped —
`assay_test.go`'s `TestRunEdgePassFixedMethodNote` pins the mixed-tree-with-leaf-failed case.

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
