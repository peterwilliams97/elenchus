# spec/TREE.md — the report tree

> **Scope note — assay-era spec, to be ported to elenchus2.** Same disposition as `spec/CLI.md`:
> paths (`internal/tree`, `internal/brief`) are the assay-era homes; the behaviour travels unchanged
> when `assay` becomes `crossexam`.

## What the tree is

The **tree** is the report's findings laid out as a tree and hung with a faithfulness verdict per
finding. It is the top-level artifact of the LCEIC assay: one node per §-heading, one leaf per claim,
each leaf carrying the judge's verdict, its reason, and the verified evidence quotes. A reader scans
the tree to see, finding by finding, whether the report faithfully represents what the hearings,
submissions, and responses to questions on notice actually say.

Each claim in a claims file carries a `key=Label/key=Label` **path** (its §-heading trail, e.g.
`2=Chapter 2 — Overview…/2.2.1=Loss of audiences and income`). `internal/tree` groups the claims by
that path into a tree; a claim with an empty path lands under a single `unplaced` branch.

## Producing the tree

The complete tree is one uniform run over the whole report:

    assay -backend anthropic -model claude-sonnet-4-6 -retrieve bm25 -n 3 \
          -manifest examples/vic-lceic/sources/MANIFEST.md \
          -source <hearings>,<submissions>,<qon> \
          -chain-dir evidence/<date>-full/sonnet-retrieved \
          -tree  examples/vic-lceic/claims-machine-full.txt

All 69 findings, Sonnet, retrieved, N=3, on the final corpus, under the judge and rules below. The run
writes a **Tier-2 chain** (one JSON record per finding: verdict, N=3 spread, the full `samples` list —
every one of the N sample verdicts in the order they were drawn, so the dissent behind a `2/3` is
recoverable — retrieved passage ids, verified quotes and their source type, reason) plus `tree.html`
and `audit.md` in the chain directory.

Rendering is separable from judging. `assay -from <chain.jsonl> <claims.txt>` re-renders the brief,
the tree, and `tree.html` from a saved chain with **no model calls** — so a tree can be rebuilt,
restyled, or merged for free. A run is also **resumable**: rerun the same command and every finding
already in the chain is reused (grouping completes findings out of index order, so the partial chain
may have gaps; the reader tolerates them). `-fresh` re-judges everything.

### Merging runs — `-from a.jsonl,b.jsonl,…`

`-from` takes a comma-separated list of chain paths. **One** path replays a single run and behaves
exactly as before — byte-for-byte unchanged. **Two or more** merge the runs leaf by leaf, so the
reader sees not one judge pass but the agreement across several full runs of the same corpus.

- **Verdict.** Each run contributes its sample verdicts — the `samples` list when it sampled `-n>1`,
  else its single `verdict` as one sample. The merged leaf's verdict is the **majority (modal)** over
  the whole pool of every sample from every chain; an even split breaks toward the worse verdict
  (`faithVerdictOrder`, worst-first), the same tie-rule a single `-n>1` run uses.
- **Agreement.** The `k/N` fraction is recomputed over the pool (`N` = total samples across all
  chains) and the minority verdicts are named as the dissent — identically to a single `-n>1` run, so
  `partial 4/6 ≠ faithful` reads the same whether the six draws came from one run or three.
- **Verdict, no majority (a tie).** When the pool's highest sample count is shared by two or more
  verdicts, there is no majority the merge can name. `modalVerdict` would still return one (worst-first),
  but that pick is an artifact of the tie-break, not agreement — so a tie is **not** rendered as a
  verdict. The leaf carries a **split descriptor** (the tied verdicts, worst-first) and its class is
  **contested**, whether or not the tie crosses the support divide. A 2/2 same-side split
  (`faithful`/`partial`) and a three-way one-each split are both ties.
- **Class.** Each merged leaf is one of three **stability classes**, from the spread of verdicts in
  its pool:
  - **settled** — every sample is the same verdict.
  - **wobble** — the verdict moves but stays on one side of the support divide (supported =
    `faithful`/`partial`; not-supported = `overstated`/`absent`/`contradicted`/`unsupported`;
    `unverifiable` is its own side) **and a majority still holds**. Instability that does not change
    *whether the source backs the claim* and still resolves to one verdict.
  - **contested** — the verdicts **cross** the divide (at least one run says the source backs the
    claim and at least one says it does not, or could not check), **or the pool has no majority at
    all** (a tie). Either way the judge does not settle the very thing the tree exists to settle.

A **contested** leaf enters Needs-you at **tier 0** and, within tier 0, ranks **ahead of** the
settled red verdicts (a unanimous `contradicted` included): an unstable finding is the one a reader
most needs to open, because no single run's verdict on it can be trusted. This is a class-based
tiebreak inside tier 0, not id order.

The root block gains one line beneath the class partition when more than one run is merged:

    Across <R> runs: <s> settled, <w> wobble, <c> contested.

The three counts partition the **judged** findings. Opinions (`route=evaluative` and recommendations)
are never judged, so a stability class on one is a category error: they are excluded from this line and
reported under *Committee opinions* instead, and the three counts plus that opinion count sum to `N`.
The `contested` figure here therefore equals the top group's count below — both are the judged findings
the runs could not settle. A single-chain `-from` prints no such line.

The support divide above is the **faithfulness** axis. A merge of substance or grounding chains pools
the verdict and fraction the same way, but each distinct verdict counts as its own side, so any
disagreement reads as contested. **Audit chains are not mergeable** (their verdict is a composite of
three axes) and a multi-chain `-from` over an audit chain is rejected.

**Refuters.** A 3-chain fixture with three leaves — one where all runs agree, one where the verdict
moves within a side, one where it crosses — yields exactly one `settled`, one `wobble`, one
`contested` (`assay_test.go` `TestMergeChainsThreeClasses`). `TestStabilityClass` additionally pins a
same-side 3-way split and a 2/2 same-side split as `contested` (no majority), and `TestSplitDescriptor`
pins the tied-verdict line (worst-first `a/b`, `""` when a majority exists). And in `internal/brief`, a
contested leaf whose modal verdict is `faithful` sorts **ahead of** a settled `contradicted` leaf,
proving the contested-first tiebreak — not id order — decides the head of tier 0.

### `current/`

`examples/vic-lceic/current/` is the pointer to the tree in force. It holds the merged chain, its
claims file, `tree.html`, and a **`PROVENANCE.md` with one line per claim** naming the run each
verdict came from. It is repointed at the newest complete run (`evidence/<date>-full/`); between full
runs it may be a heterogeneous merge of the latest per-claim verdict across cells, which the
provenance lines make explicit.

## Verdicts and the Needs-you selection

A finding's leaf carries one faithfulness verdict:

`faithful` | `partial` | `overstated` | `absent` | `contradicted` | `unsupported` | `unverifiable`

The tree does not shout every leaf. It surfaces a **Needs-you** set — the findings a reader must look
at — chosen by `internal/brief`'s tier rule (`Qualify`), highest priority first:

| tier | fires on | shown reason |
|---|---|---|
| 0 | a **contested** merged leaf (ranked first within the tier), then `contradicted`, `absent`, or `unsupported` | faithfulness finding |
| 1 | `faithful` **and** grounding `refuted` (the laundering signature) | grounding finding |
| 2 | `overstated` **and** the claim contains a number | faithfulness finding |
| 3 | `partial` **and** a non-`none` gap (scope/denominator/timerange/attribution/other) | faithfulness finding |
| 4 | grounding `refuted` | grounding finding |
| 5 | `unverifiable` (a cited document is not in the corpus) — its own tier, below the red ones | faithfulness finding |

`brief.OpenedBranches` counts the distinct top-level branches that hold a Needs-you finding — the
run's headline "changes N summary lines" (the value line from `docs/VALUE.md`). By default the tree
expands only the Needs-you branches; `-tree=full` expands every node.

## Root node rendering

Above the tree — in the text output and at the top of `tree.html` — sits a **root summary block**: the
one screen a reader sees before any branch. It is `tree.RootBlock`, assembled entirely from the
verdict rows and the chain (`spec/CLI.md` § Tier-2 chain); it makes **no model call**, so `-from`
prints it for free. Every count in it is a partition of the run's findings — each finding lands in
exactly one class — so the classes sum to the headline N and nothing is silently dropped.

### The `route` field

Each claim carries a `route` — what kind of thing can settle it — declared in `claims-machine.txt`
(`route=…`) and propagated into `claims-machine-full.txt` (a fifth tab column) and each chain record
(`"route"`). It aligns with the axis boundary in CLAUDE.md:

- `evidence` — a factual claim about the world; needs an external truth-maker.
- `source` — compresses witness/submission testimony; checked against the corpus (faithfulness).
- `evaluative` — a Committee value judgment ("disappointing", "should"); only checkable as *did
  stakeholders say this*, never groundable as world-fact.
- `data-gap` — a claim about a dataset's own limits (a breakdown "is not released").

An **opinion** — `route=evaluative`, or a recommendation (an `R`-prefixed id) — is not judged: its leaf
renders verdict `opinion` (grey in `tree.html`), it never enters Needs-you, and at render it costs
`$0`. `route` is read from the chain record; a record without one falls back to the claims file, so a
chain written before `route` existed still renders correctly once the claims file carries it.

### The block, line for line

1. **Title line.** `<report title>, <date> — <N> findings checked against <M> source documents` — the
   title and date from the claims file's `# title:` / `# date:` header, `M` from the manifest
   (`-manifest`, else a `# sources:` header). The "checked against M source documents" clause is
   dropped when `M` is unknown (0).
2. A blank line, then one line per **class** below, each omitted when its count is 0, in this order.
   Placement runs in a fixed precedence, so each finding lands in exactly one class: **opinion first**
   (`route=evaluative` or a recommendation — its verdict never places it once it is an opinion), then
   the **stability partition** (a `contested` leaf is filed by disagreement, not by verdict, and so is
   excluded from every verdict class), then the **schema gate** (a malformed judge reason forces the
   verdict to unverifiable), then the **verdict** itself.

| class line | membership | detail shown |
|---|---|---|
| `Sources don't settle these: <n>` | class `contested` (the runs crossed the divide or had no majority), non-opinion — **the top group** | **every** contested claim (no 3-id cap — see Line budget), one plain line each, sentence-joined from the `The report says <claim>` opener (the claim in plain words — the leaf's `so_what` report-says sentence, or the claim text when the modal was faithful and left none — with its trailing period(s) dropped and first letter lower-cased), then exactly one period before the disagreement clause as a second sentence: `Runs split a/b.` for a tie, else `<modal> <spread> against <dissent>.` (the crossing verdicts) |
| `Holds: <n> findings say what their sources say.` | verdict `faithful`, non-opinion, non-contested | none |
| `Contradicted by their own sources: <n>` | verdict `contradicted`, non-opinion, non-contested | one indented line per claim: id, then `so_what` (≤ 20 words) |
| `Overstated: <n>` | verdict `overstated`, non-opinion, non-contested | ids + `so_what` |
| `Unsupported by any held source: <n>` | verdict `unsupported`, or verdict `absent` on a non-opinion claim (route `source`/`evidence`/`data-gap`), non-contested | ids, `so_what` where present |
| `Committee opinions, not checked: <n>` | opinions (`route=evaluative` and recommendations) | none (no ids) |
| `Unverifiable, document not held: <n>` | verdict `unverifiable` from a cited document not in the corpus | ids + the missing document name (from `so_what`, "fetch X" → X) |
| `Unverifiable, judge output malformed: <n>` | the schema gate fired — verdict forced to `unverifiable` because the judge's reason was empty or carried a raw tag (see § The rules that gate a verdict) | one plain line per claim: `id: judge output malformed` |
| `Generalised from narrower evidence: <n>` | verdict `partial`, non-contested | up to **2** indented lines for partials with a non-`none` gap **and** a number or place name in the claim text, ranked by spread (`3/3` before `2/3`), id + `so_what` |

3. A blank line, then the value line: `Changes <s> summary lines across <c> chapters.` — `s` is the
   number of distinct §-heading sections (a claim's full heading path) holding a Needs-you leaf, `c`
   the number of distinct chapters (top-level branches, `brief.OpenedBranches`) holding one.

An `absent` or `unsupported` verdict on an *opinion* never reaches the Unsupported class — the opinion
resolution runs first — which is what the mislabel refuter below turns on. A "number or place name" is
a digit anywhere in the claim, or a mixed-case capitalised word (≥ 2 letters) that is not the claim's
first word (so `New York`, `Victoria` mid-sentence, `NSW`-adjacent proper nouns qualify; a
sentence-initial capital and an all-caps acronym like `ABC` do not).

**Line budget.** Every verdict class listing ids shows at most the first **3**, then `and <k> more`.
The contested top group is the one exception — it shows **every** line, because a reader must see the
whole disagreement rather than a sample of it, so the block body grows with the contested count and is
no longer bounded to 15 lines. `so_what` text is used verbatim from the chain, truncated to 20 words; a
leaf with none prints the id alone.

### Refuters

`internal/tree` `TestRootBlock` builds a hand-made chain of 12 records covering the verdict classes and
checks the block line for line against a golden string. `TestRootBlockRouteGates` flips one
`route=evaluative` record to `route=source`; because that record's verdict is `absent`, the flip must
move it out of *Committee opinions* and into *Unsupported by any held source*, changing that count — the
test that `route` actually gates the partition, not just decorates it.

`TestRootBlockContestedFirst` is the stability-partition refuter: two contested leaves — one that
crossed the divide with a modal `faithful`, one that tied — must appear together under *Sources don't
settle these*, each as the `The report says <claim>.` head then the disagreement as a second sentence
(the tie ending `Runs split a/b.`, the crossing one `<modal> <spread> against <dissent>.`) carrying the
claim text and not just verdict names, and be kept out of the verdict classes; were the crossing leaf
filed by verdict it would inflate *Holds*. The mutation clears its `contested` class and it must then
move into *Holds*, proving stability is partitioned first. `TestRootBlockContestedUncapped` is the
cap-lifting refuter: a nine-contested fixture must print nine detail lines with no `and <k> more`
collapse, where a verdict class of the same size shows three. `TestRootBlockSchemaFailure` is the
schema-gate refuter: a leaf marked `SchemaFail` must file under *Unverifiable, judge output malformed*
(flagged by id) and stay out of *Holds* even though its stored verdict still reads `faithful`; clearing
the flag returns it to *Holds*.

## The rules that gate a verdict

The verdict on a leaf is not the judge's word taken on trust. Four code-side rules stand between the
model and the tree (`spec/CLI.md` § Faithfulness judge and § Retrieval carry the mechanics):

1. **Quote by reference.** Every evidence item is `{passage_id, quote}`; code verifies the quote is a
   verbatim substring of the named passage and drops any that is not. No model call.
2. **A verdict needs a verified quote.** Every verdict except `absent` must keep at least one verified
   quote: with none, `contradicted`→`absent` and `faithful`/`partial`/`overstated`→`unsupported`. A
   fluent but ungrounded judgment cannot reach a verdict.
3. **A cited document must be held.** A finding whose `cites` name a document not in the manifest is
   `unverifiable` (reason names the missing document, so-what "fetch X"), decided before any model
   call. This reserves `absent` for cited-document-present, claim-not-found.
4. **Retrieval floor.** When nothing clears `-floor`, the finding is `absent` from code, retrieved=0.
5. **Schema gate.** A leaf whose judge reason is empty or carries a raw markup tag (e.g. a leaked
   `</antml_parameter>`, the fingerprint of a truncated or malformed structured response) is a schema
   failure: its verdict is forced to `unverifiable` and the leaf is flagged, because a verdict whose
   reason did not survive the schema rests on nothing checkable, whatever the verdict string claims.
   Opinions are exempt — they are never judged, so an empty reason on one is expected. In the root block
   these leaves form the *Unverifiable, judge output malformed* class, distinct from the missing-document
   `unverifiable`. Refuters: `brief.SchemaFailed` (empty and raw-tag reasons fail; a plain finding, a
   reason containing `<` used arithmetically, and ordinary punctuation pass) and the root-block
   `TestRootBlockSchemaFailure` above.

Faithfulness itself is scope-disciplined: `faithful` requires the source to state the claim's subject,
scope, and direction; an adjacent or broader statement is `partial` at best (`gap=scope`).

### Reasoning before the enum

`judgeSchema` (`assay.go`) orders its properties reasoning-first: `report_says`, `source_says` and
`reason` are emitted before `gap`, `evidence`, and the `verdict` enum. A strict tool emits fields in
schema-property order, so this is the order the model generates them in — the enum is chosen only
after the plain restatements and the reason have been written, not pattern-matched ahead of them. The
edge and substance schemas already reason first: edge's `warrant` precedes its `defeater`
(`internal/edge/prompt.go`), and substance's `critique` precedes its `verdict` (`assay.go`, the
`Return ONLY JSON` template). This brings the faithfulness judge into line with them.

The calibration case that motivated the reorder is slice-2 CMP-BODY (the IO-3 body statement "roughly
one third of the AI compute capacity of Europe"), run Sonnet N=3 on 2026-09-16
(`examples/tai-europe-2026/evidence/2026-09-16-slice2/sonnet/claims-slice2.faithfulness.jsonl`,
record 1): samples were `overstated, contradicted, contradicted` (modal `contradicted`, 2/3) while the
record's own `critic_finding` reads "the ratio roughly one third — the claim is arithmetically
consistent with the source figures. Verdict: faithful." Two of three samples returned an enum their
own reasoning contradicted — the failure a verdict-first field order invites, where the enum is fixed
before the arithmetic is worked. Refuter: `TestJudgeSchemaReasoningBeforeEnum` (`judge_test.go`) pins
`reason` before `gap`/`evidence`/`verdict` in the schema source.

### Refuters for the reorder — results

Pre-registered before the runs (discipline: pre-register the attempt, not the success); the
prediction and the observed result are kept side by side below. **Decision: the reorder stands.**

- **(a) Enum agrees with reasoning on CMP-BODY; arithmetic accepted.** Predicted `faithful` 3/3 on
  both restructured Slice-2 leaves. Result: the reordered schema no longer returns an enum its own
  reasoning contradicts — the arithmetic (2.1 / 0.662 = 3.17 ≈ "three times" / "one third") is
  accepted. CMP-BODY still scored partial, but the cause was a source-scoping error in our fixture,
  not a judge miss: footnote 2 states "2.1 GW for all of Europe" with no country scope, so CMP-BODY's
  "including the UK and Norway" had no truth-maker. Fixed by holding the p192 body sentence alongside
  the footnote in `sources/report-fn2.txt` (see MANIFEST and `claims-slice2.txt`); re-registered
  `faithful`, pending the slice-2 rerun.
- **(b) DORA N=3 rerun scored against noise and adjudicated leaves.** Result: the a-vs-b noise floor
  (two runs of the unchanged config) differs on **3/136** leaves. The reorder-vs-b comparison changed
  **28/126** leaves (10 leaves errored, below, so 126 comparable) — 22 milder, 6 harsher. Agreement
  with the 5 PW-adjudicated DORA leaves rose **0/5 → 3/5** (AD19 moved toward the human verdict). The
  reorder moves well above the noise floor and improves agreement with the settled leaves, so it stands.
- **10 leaves errored on billing** (DG4–DG15, ME4, ME5) and are excluded from the 126; they are to be
  rerun into the same chain dir (command below).

**Corrected scoring rule.** A judge change is scored against the a/b noise floor and against the
human-adjudicated leaves — never against the previous run's verdicts. The previous run is not ground
truth: a change that moves verdicts is a fix if it moves them toward adjudicated leaves and beyond
what two identical runs differ by, regardless of how many earlier verdicts it disturbs.

### Cite-scoped retrieval

Which passages a claim is judged against is decided by its `cites`, not by the whole corpus:

- A claim that **cites a held document other than the report** is judged against that document (or
  documents) **alone**; every report passage is excluded from the judge's context for that claim. A
  `route=benchmark` claim cites a leaderboard capture as its truth-maker, so grounding it against the
  report restating the same figure would launder the report's own word into a corroboration it never
  earned — the axis boundary in CLAUDE.md. Retrieval draws from the cited documents so the judge sees
  the leaderboard, not the report echoing it.
- A claim with **no external cite** (it cites only report excerpts, or nothing) is judged against the
  report itself, under the single-source rules below.

Mechanically, `citedExternalBases` (assay.go) resolves a claim's `cites` to the passage-id bases of its
non-report documents — a report cite is any id matching a manifest report excerpt's PDF filename; the two
external shapes this corpus uses map `paper:<stem>` → `papers/<stem>` and a `<dir>/<stem>.txt` leaderboard
id → `<dir>/<stem>`. When that set is non-empty, `passagesForClaim` calls `retrieve.RetrieveFrom`, which
ranks the whole corpus in fused order then keeps only the passages whose base is in the set. With no
passage in the cited documents — or none clearing `-floor` — `RetrieveFrom` sets `Below`, so the claim is
`absent` from code with no model call: the report self-restatement, now excluded, cannot rescue it. The
leaderboard captures reach the index through `retrieve.leaderboardPassages`, which splits a capture under a
`leaderboards/` directory into paragraph passages tagged `Source=leaderboard`, keeping each table's rows
whole so a claim's figure grounds as a verbatim quote.

Refuters: `retrieve.TestCiteScopedRetrieval` pins the exclusion against the real ai-index corpus — a claim
whose figure the report restates (plain `Retrieve` reaches it), cite-scoped to the leaderboard the claim
names, returns only that capture and no report passage, and a cite to a document holding no passage falls
to `Below`. `retrieve.TestLeaderboardPassages` pins the capture ingestion (base, source tag, verbatim
figure), and `TestCitedExternalBases` (assay_test.go) pins the report-vs-external split of a `cites` field.

### The segmenter seam

A **Segmenter** turns one held document into passages — the parser a corpus file gets, chosen by an
explicit registry (`internal/retrieve/segment.go`, `passagesForFile`) rather than a general prose
splitter for everything. Two are named:

- **`HansardSegmenter`** — the existing speaker-turn parser (`splitData`, the body of `splitFile`),
  **unchanged**: committee transcripts split into turns with role and question context. A corpus file
  with no more specific routing lands here, exactly as before the seam.
- **`PlainSegmenter`** — splits a document on blank lines into paragraph passages, **merging each
  fragment under ~200 runes into the next paragraph** (a wrapped heading or a nav/menu line off an
  HTML-to-text strip is held and attached forward, so the prose stays whole rather than scattering into
  one-line passages). It is used for **any held id under `cited/`** — a fetched web-page capture, the
  external truth-maker for a `route=evidence` claim, which carries no speaker turns. The passage id base
  is `cited/<stem>` and `Source=cited`.

Selection is by the held id's `cited/` prefix (the on-disk path substring `/cited/`), the same shape the
submission/qon/leaderboard/report parsers route by. **Cite-scoped retrieval** then restricts a
`route=evidence` claim's candidate passages to its cited id's passages: `citedExternalBases` (assay.go)
maps a cite `cited/<stem>.txt` to the base `cited/<stem>`, and `RetrieveFrom` keeps only passages on that
base — so the claim is grounded against the page it cites, never the report restating the same number
(the axis boundary in CLAUDE.md). A cite naming an id not in the manifest (a walled or uncited target) is
`unverifiable` with no model call, decided before retrieval.

Refuters: `retrieve.TestSegmenterGoldenUnchanged` is the golden — every existing corpus's passage set
(vic-lceic hearings/submissions/qon, dora, quocirca, ai-index report/leaderboards/papers) is byte-identical
to a hash captured on the pre-seam code, so the refactor perturbed nothing the plain segmenter did not add
(master-plan is excluded: its Markdown `sources/` runs `retrieve=none` and reaches no segmenter).
`retrieve.TestPlainSegmenterCited` pins the plain segmenter over the 16 held tai-europe captures — each
yields ≥3 passages on the `cited/<stem>` base with `Source=cited`, and two `fig=present` anchor sentences
each land in exactly one passage. `TestSlice3CiteScopedClassification` (assay_test.go) is the end-to-end
dry run: the 20 Slice-3 leaves classify, with no model call, into 16 eligible + 4 unverifiable + 0 floored.

### Single-source judging has a direction

When the manifest is `single_source: true` the report is its own only source, so the passages a claim
is judged against are OTHER parts of the same document — two statements by one author, not a claim
weighed against an outside witness. The comparison has a direction, and only one direction is a
distortion. The judge prompt carries these rules for single-source runs (`faithJudgeSingleSourceRules`
in `assay.go`):

- A claim is `overstated` or `contradicted` **only if** another passage states the same fact with
  **narrower scope or a different value AND is the more detailed statement** — the body pinning down
  what the claim inflated or altered.
- A **summary that drops detail is not a conflict**: a passage that rounds, summarises, or drops a
  qualifier the claim keeps is the vaguer of the two, and the claim carrying more detail is `faithful`.
- **Figures that nest are consistent**: `very` (55%) sits inside `very or somewhat` (85%), a component
  share inside the total that contains it. A larger combined figure does not contradict a smaller
  sub-figure of it — `faithful`.
- **Complementary percentages are consistent**: two shares that partition one population — `some
  confidence` (70%) and `little or no trust` (30%) — are two faces of one 100% split, not two facts in
  conflict. A claim's figure and a passage's figure summing to 100 is never `contradicted`.
- **Figures within a point are the same figure rounded**: two figures within one percentage point of
  each other, or two shares of one split that fall within a point of summing to 100% (78% and a
  complementary 21% sum to 99%), differ only by rounding — `faithful`, never `contradicted`.
- **A claim at one level of a stated taxonomy does not conflict with the taxonomy's parent**: when the
  report defines the hierarchy itself — throughput's three factors, instability's two, both under
  software delivery performance — a claim that places an item at the level the source places it does not
  contradict a passage naming only the parent or a sibling level. `faithful`.
- A passage that contains the claim **near-verbatim** — same words, same figure, same scope — is
  `faithful`, whatever wording differs elsewhere.
- **With no conflicting passage retrieved**, a claim the passages neither pin down nor repeat is
  `absent` (rendered `uncorroborated`), never `overstated` or `contradicted`.

The self-exclusion that makes this an internal-consistency check drops the claim's own **paragraph**,
not its whole page: `dropOwnParagraph` (assay.go) removes the single report passage on the claim's §
that shares the most words with it — the paragraph the claim was decomposed from — so the claim cannot
confirm itself, while a qualifier one paragraph over in the same § survives and can still ground it.
The page-wide exclusion this replaced dropped every same-§ passage, so a fact restated one paragraph
over read `absent`; cross-page restatement corroborates as before. In a multi-excerpt corpus the drop
is scoped to the claim's own excerpt (the one its claim id routes to, `manifest.ReportFor`), keyed on
the excerpt stem a report passage id carries (`report`, `report-productivity`), so a same-page
paragraph of the OTHER excerpt is never mistaken for the claim's source. Refuter: `TestDropOwnParagraph`.

Refuters: `TestSingleSourceDirection` in `assay_test.go` pins six shapes — three from the quocirca
corpus (E1, a specific claim against a vaguer restatement; E23, nested percentages; K37, the claim
present near-verbatim) and three from the DORA corpus the base prompt scored `contradicted` (AD19,
complementary percentages 70/30; EX20, a rounded complement 78/21; SD1, a throughput sub-factor read as
conflicting with the instability level) — each `faithful` under these rules, each carrying the rule
sentence it rests on.

## Reading a leaf back to its evidence

Every leaf is traceable. Its chain record names the retrieved passage ids the judge saw, the verified
quotes and whether each came from a hearing, a submission, or a QoN response, and the modal verdict
plus its N=3 spread — so `partial 2/3` reads as "look here, the judge was not unanimous," and a quote
tagged `qon` reads as "this rests on a response to questions on notice." The evidence origin and the
manifest together make the tree auditable: what the finding rests on, whether that document is held,
and whether the words are actually there.

Because the chain also stores the `samples` list, `-from` reconstructs the spread from those raw
verdicts rather than the pre-baked `spread` string, and can name the dissent: a `2/3` leaf renders
`partial 2/3 ≠ faithful`, the minority verdict spelled out. A chain written before `samples` existed
lacks the field and still renders — the spread falls back to the stored `spread` string, with no
dissent shown. A merged **tie** has no majority to name, so its leaf line ends in `split a/b` (the
tied verdicts, worst-first) rather than a modal verdict with a `≠` dissent — refuter `TestSplitLeaf`.
