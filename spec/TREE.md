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
writes a **Tier-2 chain** (one JSON record per finding: verdict, N=3 spread, retrieved passage ids,
verified quotes and their source type, reason) plus `tree.html` and `audit.md` in the chain directory.

Rendering is separable from judging. `assay -from <chain.jsonl> <claims.txt>` re-renders the brief,
the tree, and `tree.html` from a saved chain with **no model calls** — so a tree can be rebuilt,
restyled, or merged for free. A run is also **resumable**: rerun the same command and every finding
already in the chain is reused (grouping completes findings out of index order, so the partial chain
may have gaps; the reader tolerates them). `-fresh` re-judges everything.

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
| 0 | `contradicted`, `absent`, or `unsupported` | faithfulness finding |
| 1 | `faithful` **and** grounding `refuted` (the laundering signature) | grounding finding |
| 2 | `overstated` **and** the claim contains a number | faithfulness finding |
| 3 | `partial` **and** a non-`none` gap (scope/denominator/timerange/attribution/other) | faithfulness finding |
| 4 | grounding `refuted` | grounding finding |
| 5 | `unverifiable` (a cited document is not in the corpus) — its own tier, below the red ones | faithfulness finding |

`brief.OpenedBranches` counts the distinct top-level branches that hold a Needs-you finding — the
run's headline "changes N summary lines" (the value line from `docs/VALUE.md`). By default the tree
expands only the Needs-you branches; `-tree=full` expands every node.

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
