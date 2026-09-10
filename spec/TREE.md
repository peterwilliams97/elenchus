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
- **Class.** Each merged leaf is one of three **stability classes**, from the spread of verdicts in
  its pool:
  - **settled** — every sample is the same verdict.
  - **wobble** — the verdict moves but stays on one side of the support divide (supported =
    `faithful`/`partial`; not-supported = `overstated`/`absent`/`contradicted`/`unsupported`;
    `unverifiable` is its own side). Instability that does not change *whether the source backs the
    claim*.
  - **contested** — the verdicts **cross** the divide: at least one run says the source backs the
    claim and at least one says it does not (or could not check). The judge disagrees with itself
    about the very thing the tree exists to settle.

A **contested** leaf enters Needs-you at **tier 0** and, within tier 0, ranks **ahead of** the
settled red verdicts (a unanimous `contradicted` included): an unstable finding is the one a reader
most needs to open, because no single run's verdict on it can be trusted. This is a class-based
tiebreak inside tier 0, not id order.

The root block gains one line beneath the class partition when more than one run is merged:

    Across <R> runs: <s> settled, <w> wobble, <c> contested.

The three counts partition the findings (they sum to `N`), so the reader sees at a glance how much of
the report the judges agreed on. A single-chain `-from` prints no such line.

The support divide above is the **faithfulness** axis. A merge of substance or grounding chains pools
the verdict and fraction the same way, but each distinct verdict counts as its own side, so any
disagreement reads as contested. **Audit chains are not mergeable** (their verdict is a composite of
three axes) and a multi-chain `-from` over an audit chain is rejected.

**Refuters.** A 3-chain fixture with three leaves — one where all runs agree, one where the verdict
moves within a side, one where it crosses — yields exactly one `settled`, one `wobble`, one
`contested` (`assay_test.go`). And in `internal/brief`, a contested leaf whose modal verdict is
`faithful` sorts **ahead of** a settled `contradicted` leaf, proving the contested-first tiebreak —
not id order — decides the head of tier 0.

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
   `route=evaluative` is resolved to the opinion class *first*, before any verdict-based class, so a
   claim's verdict never places it once it is an opinion.

| class line | membership | detail shown |
|---|---|---|
| `Holds: <n> findings say what their sources say.` | verdict `faithful`, non-opinion | none |
| `Contradicted by their own sources: <n>` | verdict `contradicted`, non-opinion | one indented line per claim: id, then `so_what` (≤ 20 words) |
| `Overstated: <n>` | verdict `overstated`, non-opinion | ids + `so_what` |
| `Unsupported by any held source: <n>` | verdict `unsupported`, or verdict `absent` on a non-opinion claim (route `source`/`evidence`/`data-gap`) | ids, `so_what` where present |
| `Committee opinions, not checked: <n>` | opinions (`route=evaluative` and recommendations) | none (no ids) |
| `Unverifiable, document not held: <n>` | verdict `unverifiable` | ids + the missing document name (from `so_what`, "fetch X" → X) |
| `Generalised from narrower evidence: <n>` | verdict `partial` | up to **2** indented lines for partials with a non-`none` gap **and** a number or place name in the claim text, ranked by spread (`3/3` before `2/3`), id + `so_what` |

3. A blank line, then the value line: `Changes <s> summary lines across <c> chapters.` — `s` is the
   number of distinct §-heading sections (a claim's full heading path) holding a Needs-you leaf, `c`
   the number of distinct chapters (top-level branches, `brief.OpenedBranches`) holding one.

An `absent` or `unsupported` verdict on an *opinion* never reaches the Unsupported class — the opinion
resolution runs first — which is what the mislabel refuter below turns on. A "number or place name" is
a digit anywhere in the claim, or a mixed-case capitalised word (≥ 2 letters) that is not the claim's
first word (so `New York`, `Victoria` mid-sentence, `NSW`-adjacent proper nouns qualify; a
sentence-initial capital and an all-caps acronym like `ABC` do not).

**Line budget.** Any class listing ids shows at most the first **3**, then `and <k> more`; the whole
block body (the class region) stays within 15 lines. `so_what` text is used verbatim from the chain,
truncated to 20 words; a leaf with none prints the id alone.

### Refuter

`internal/tree` `TestRootBlock` builds a hand-made chain of 12 records covering every class and checks
the block line for line against a golden string. A companion mutation flips one `route=evaluative`
record to `route=source`; because that record's verdict is `absent`, the flip must move it out of
*Committee opinions* and into *Unsupported by any held source*, changing that count — the test that
`route` actually gates the partition, not just decorates it.

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

Faithfulness itself is scope-disciplined: `faithful` requires the source to state the claim's subject,
scope, and direction; an adjacent or broader statement is `partial` at best (`gap=scope`).

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
dissent shown.
