# spec/ARGUMENT.md — the argument tree

> **Scope note — assay-era spec, no code yet.** This spec defines a data structure and a derivation
> rule; nothing here is implemented. It is the sister of `spec/TREE.md`, and reuses that spec's leaf
> verdict and stability classes wholesale — read TREE.md first. Where TREE.md groups a report's
> findings by their **§-heading** trail, this spec arranges the same findings by their **inferential**
> trail: what the report is trying to establish, and what it rests that on.

## What the argument tree is

A document is a **root proposition refined, step by step, down to evidence.** The argument tree makes
that refinement explicit: one node per step, edges pointing from a claim to the sub-claims that
support it, bottoming out at atomic claims whose truth an external check can settle.

It answers a different question from the report tree. The report tree (`spec/TREE.md`) asks, finding
by finding, *does the report faithfully represent its sources?* The argument tree asks *does the
report's case hold together?* — given each finding's verdict, does the recommendation it supports
still stand, and does the whole report's thesis. The report tree is the input; the argument tree is
the structure that propagates those leaf verdicts up to the conclusions that depend on them.

For the LCEIC report the refinement has four levels (`examples/vic-lceic/argument.txt`):

    root proposition
      └ recommendation        (R1–R11: what the report asks government to do)
          └ finding            (F1–F55: what the report concludes)
              └ atomic claim   (the ids in claims-machine.txt: what a source can settle)

Not every level is present on every branch: a finding that no recommendation rests on hangs directly
below the root, and a single-claim finding has one claim leaf whose id equals the finding's.

## A node is content plus a derived judgement

Every node carries two things and only two:

- **content** — the proposition, in the report's own words, kept short. This is **authored**: whoever
  builds the tree writes it, compressing the report's sentence to a line. It is the one thing in the
  tree a human puts there.
- **judgement** — where the evidence leaves this proposition standing. This is **derived, never
  authored.** No one writes a judgement into the tree; it is computed from the leaves up, and writing
  one by hand is the one edit the format forbids — it would let the builder's opinion masquerade as a
  result. (The same discipline TREE.md enforces on a leaf verdict, carried up the tree.)

`examples/vic-lceic/argument.txt` therefore carries content and structure but **no judgements at
all** — they do not exist until a chain has been pooled onto the leaves and propagated. A judgement
column in that file would be a fabricated result.

## Leaf judgement — the pooled verdict

A leaf is an atomic claim. Its judgement is exactly the **pooled merged-leaf verdict** from
`spec/TREE.md`: the modal faithfulness verdict over every sample in the pool, plus the leaf's
**stability class** (`settled` / `wobble` / `contested`, and a tie's `split`). Nothing new is defined
here — the argument tree reads the leaf straight off the merged chain. An opinion leaf
(`route=evaluative`, or a recommendation treated as a claim) renders `opinion` and is never judged,
per TREE.md's route gate.

The verdict vocabulary a leaf can present to its parent:

`faithful` · `partial` · `overstated` · `absent` · `contradicted` · `unsupported` · `unverifiable`,
each on a `settled` / `wobble` / `contested` / `split` leaf, or `opinion`.

## Internal judgement — the derivation rule

An internal node's judgement is a function of its children's judgements — where a child that is
itself internal contributes its own derived judgement, so the computation runs bottom-up. This
**conjunction rule** governs the two internal levels below the root — every **recommendation** and
**finding** node. The **root** is the one exception: it does not conjoin its children into a single
worst-case verdict but **tallies** them (§ The root). A recommendation or finding node is one of four:

- **holds** — every load-bearing child holds (a leaf `faithful`/`settled`, or an internal `holds`).
- **weakened** — no child fails or opens, but some child is `partial` or `wobble` (or internal
  `weakened`). The node still stands, but on narrower ground than its content claims.
- **open** — no load-bearing child fails, but some child is `contested` or `split` (or internal
  `open`). The runs cross the divide on a premise, so the node is not settled either way — the reader
  must open it.
- **fails** — a **load-bearing** child fails: `contradicted`, `unsupported`, or `absent` (or internal
  `fails`). One collapsed load-bearing premise collapses the node.

**Precedence.** The four can fire at once; the node takes the worst that any child triggers:

    fails  >  open  >  weakened  >  holds

So a node with one contested child and one contradicted load-bearing child is `fails`, not `open`.

**What "load-bearing" gates.** `holds` and `fails` are computed over the load-bearing children only —
this is the whole point of the load-bearing edge (below). `open` and `weakened` fire on **any** child:
a supporting child that is contested still opens the node, and a supporting child that is partial
still weakens it, because instability anywhere is something the reader needs to see; only *collapse*
is confined to the load-bearing set. One gap the four rules leave, closed here: a **non-load-bearing**
child that is `contradicted`/`unsupported`/`absent` does not sink the node (it was never essential)
but is not silently dropped either — it contributes **weakened**. A collapsed side-premise is
information, not nothing.

**Opinions contribute nothing.** An `opinion` child is not evidence, so it is excluded from the
hold/fail computation exactly as TREE.md excludes opinions from its partition. A node all of whose
children are opinions is itself `opinion` — unjudged. This is load-bearing for reading the LCEIC
tree: several findings are pure Committee value judgements (`route=evaluative`), and a recommendation
that rests only on those is `opinion`-based — faithfulness cannot settle it, and the tree must say so
rather than report a hold it never earned.

## Edges — load-bearing, and the `?` marker

An edge from a parent to a child carries two facts:

- **load-bearing or supporting.** A load-bearing edge means the parent's content genuinely depends on
  the child: remove it and the parent no longer stands. A supporting edge adds weight — context,
  corroboration, rationale — but the parent survives its loss. Only load-bearing children can make a
  parent `hold` or `fail`; supporting children can only `weaken` or `open` it (above).
- **stated or `?`.** A `?` marks an edge the **report does not establish**: the report co-locates the
  child and parent, or the child plausibly bears on the parent, but the report never says the parent
  rests on it. A `?` edge is a question for a human, not a guess by the builder — it is the format's
  way of refusing to invent a linkage the source did not assert. `?` edges are counted, not hidden,
  and a run over the tree reports how many remain.

## The root

The root is the **one proposition the report exists to establish** — the thing every recommendation
and finding is ultimately in service of. A report rarely states it as a single sentence; where it
does not, the root's content is a synthesis in the report's own words, and it is **flagged as
synthesised** with the nearest verbatim anchor cited, so a reader can tell an authored compression
from a quotation.

Unlike a recommendation or finding, the root's judgement is **not a conjunction** of its children — a
single failed recommendation does not collapse the whole report to `fails`. The root is a **tally** of
its **recommendations** (`R…`): the **count in each class** (`holds` / `weakened` / `open` / `fails` /
`opinion`), rendered as one prose sentence, worst class last, so a reader sees the shape of the whole
case at once, not the worst single branch —

    Of 11 recommendations, 2 hold (R1, R2), 1 is weakened (R3), 6 are open — 5 because the report
    doesn't say what they rest on, 1 because their findings are contested — 1 fails (R7: no held
    source supports F33), 1 is opinion (R9).

The **reason a recommendation is not settled comes from its deciding child** (§ Internal judgement):
an **open** recommendation is either one the report never says what it rests on — a `?`-edge or
childless deciding child (`R4`'s `F29 ?`; `R6`/`R10`, which attach no finding) — or one whose finding
is **contested** — a stated child that opens (`R8`'s `F39`); a **failed** one names the load-bearing
child that collapsed it (`R7: no held source supports F33`). A `?` edge does not reclassify a root
child: each is counted by its own derived judgement, and carries a trailing `?` where the report does
not establish its link to the root.

The `base` node — the descriptive findings no recommendation rests on (§ How the LCEIC tree is
authored) — is **excluded from the tally**: it is not a recommendation, so counting it among them
would misstate the case. It is reported instead by one sentence beneath the tally, naming how many
findings it carries — `41 findings support no recommendation.`

*Refuter.* A root with one failed recommendation and two holding renders the tally (`2 hold (…), 1
fails (…)`), never the bare verdict `fails` — collapsing to `fails` would be the conjunction rule, which
the root does not run. The `base` node is absent from every count in the tally (reported by its own
sentence), and the open recommendations split into exactly two reasons — `?`-edge/childless versus
contested — that sum to the open total.

## How the LCEIC tree is authored — the attachment rule

The one authored decision, beyond writing each node's content, is which findings a recommendation
rests on. The report never prints "Recommendation N rests on Findings X, Y," so the linkage is read
from the report's own **structure**, under one rule:

- **solid (load-bearing) edge** — the recommendation sits in a numbered subsection (§) whose finding
  it directly acts on or remedies, subject for subject. `R1` (index grants to CPI) under §3.2.1's
  `F20` (funding not indexed) is the clean case.
- **`?` edge** — the report co-locates a recommendation with a finding in the same subsection, but
  does not state the recommendation rests on it (`R4`'s companion `F29`), **or** the recommendation
  has no finding to match at all — it sits at a section head, or spans a whole subsection with no
  single subject match (`R6`, `R10`). A `?` edge names candidate findings in a comment where the
  report suggests them, but commits to no child.
- **unattached finding** — a finding no recommendation acts on hangs directly under the root on a `?`
  edge. The report's whole Chapter 2 overview, and many Chapter 3–4 findings, are descriptive base:
  they establish that the sector matters, was damaged, and is under-funded, but the report routes
  them to no recommendation. They are kept in the tree (nothing dropped) and marked for what they
  are.

Every finding lands in exactly one place — under a recommendation (solid or `?`) or unattached — so
the 55 findings partition, and the count is checkable. Provenance for every solid and `?` edge is the
report's "Findings and recommendations" section (printed pp. xi–xvii), which prints findings and
recommendations in order with their subsection page anchors.
