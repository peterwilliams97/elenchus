# repo-map.md — orientation

Where each concern lives and the key symbols that carry it. Line numbers are from HEAD at the time
of writing and drift; grep the symbol name if one has moved. Authoritative behaviour is the `spec/`
files, not this map. Plans and decisions live in the sibling files under `docs/todo/` — see
`README.md` there.

## What this repo actually does

In one line: **it reads a long report against the documents that report is built on, and tells you
which of the report's claims the sources actually support.**

The problem it addresses: a big report (the worked case is the Victorian LCEIC report on the
cultural and creative industries) makes hundreds of claims, cites hundreds of source documents
(hearing transcripts, written submissions, questions-on-notice), and no human has time to check
every claim against every source. `assay` does that check mechanically.

The pipeline, end to end:

1. **Decompose** — break the report into atomic claims, each tagged with the `§`-heading it sits
   under.
2. **Retrieve** — for each claim, pull the specific source passages it is about (BM25 + local
   embeddings), so the judge sees only the relevant passages, not the whole 25K-token corpus
   (`internal/retrieve`).
3. **Judge** — one schema-enforced model call per claim decides faithfulness against *only* those
   passages: `faithful` / `overstated` / `absent` (checked, not there) / `refuted` / `unverifiable`
   (the source document isn't in the held set). Every quote the model cites is then verified against
   its passage with **no model** (`quoteInPassage`), and the run can repeat each claim `-n` times for
   a stability class (`assay.go` judge path).
4. **Lay it out as two trees** — the **report tree** groups verdicts by heading; the **argument
   tree** arranges the same leaves by inferential structure (claim → finding → recommendation → root
   thesis) with every internal node's verdict *derived bottom-up from its children, never authored*
   (`internal/tree`, `spec/TREE.md`, `spec/ARGUMENT.md`).
5. **Publish** — `-review` builds a self-contained two-pane website (source PDF on the left, assay's
   reading on the right, click a claim → the PDF jumps to that page) and `assay serve` serves it
   (`spec/SERVE.md`).

The output is a page a human reads: e.g. "of 11 recommendations, 2 hold, 1 is weakened, 6 are open,
1 fails, 1 is opinion." The point is not to replace the reader but to tell them **where to look** —
which claims are unsupported and why.

The design's load-bearing idea (`CLAUDE.md`, the "axis boundary"): a claim can be **refuted** by pure
reasoning, but it can only be **confirmed** by going to the source — grounding always needs the
truth-maker, never a deduction, however clever. So the tool's whole discipline is knowing which
verdicts it has actually earned. Older single-file modes — dialectic, faithfulness, evidence, audit
over one prose file — still ship (`spec/CLI.md`), but the report→argument-tree path above is the
centre of gravity.

## Naming — `assay` vs `elenchus` (assessed 2026-09-12)

A three-way mismatch, all pointing at `elenchus` as the stale label, not `assay`:

- repo/directory name: **elenchus**; `go.mod` module: **assay** (`go.mod:1`); `README.md` title
  `# elenchus`, but every sentence under it describes **assay** checking claims against sources.
- **`assay.go` is the accurate name.** "Assay" = testing a material's true composition against a
  standard (assaying ore for its metal content) — exactly what the tool now does: test each claim
  against its truth-maker and report its true content. It matches the module and the binary.
- **The centre of gravity has moved off elenchus.** Elenchus (Socratic refutation by
  cross-examination) maps to the dialectic/substance mode — which the project's own axis boundary
  identifies as the *weakest* leg (reasoning can refute but never confirm grounding). Vocabulary
  confirms the shift: `faithful` appears in ~88 files, `dialectic` in ~15, `adversar*` in ~12,
  `elenchus` in ~10 (one being the repo name). "Adversarial" now mostly describes the *testing
  program* that attacks the tool (`TESTING.md`, `examples/destructive/`), not the tool performing
  elenchus on a target. See `design-notes.md` for where the elenchus leg stands today (dormant).
- **The call is Peter's, not a tidy-up.** Either rename the repo to `assay`, or keep `elenchus` as
  the umbrella method-name with `assay` the instrument under it — in which case the cheap fix is one
  README opening line relating the two, so a first-time reader isn't given two names for one thing.

## Top-level shape

- `assay.go` (~3,900 lines) — the CLI and the **only** code that touches the network (the judge
  path). `main` parses flags and dispatches; the `serve` subcommand is handled before flag parsing.
- `internal/` — one package per concern; new code goes here, not in `assay.go`.
- `spec/` — the contracts the code follows: `CLI.md`, `TREE.md`, `ARGUMENT.md`, `SERVE.md`.

## `assay.go` — CLI + judge path

The judge path lives here because it makes model calls.

- `main` — `assay.go:156` intercepts `serve` (`os.Args[1] == "serve"`) before parsing; the `serve`
  FlagSet is built at `assay.go:437`, served by `serveHandler` (`assay.go:486`), a plain
  `http.FileServer`.
- `faithJudge` — `assay.go:1181` — one schema-enforced call per claim (verdict + evidence quotes).
- `callSchema` — `assay.go:2398` — the single JSON-schema call helper over the chosen `backend`.
- `quoteInPassage` — `assay.go:1495` — verifies a cited quote against its passage, **no model**.
- `groundVerdict` — `assay.go:1405` — grounds the verdict against verified quotes, no model.
- `faithJudgeRepeat` — `assay.go:1429` — samples `-n` times → modal verdict + `k/N` stability class.
- `assayClaim` — `assay.go:1104` — the substance (dialectic) path; `decompose` — `assay.go:1094` —
  the LLM claim atomizer used by substance mode.
- `runEvidence` — `assay.go:1003` — grounding mode; `crossCheckEvidence` — `assay.go:2102` — proves
  a cited URL was retrieved (provenance), never that the page supports the sentence (content).

A judge run writes a Tier-2 JSONL chain; `-from` re-renders that chain into trees with no further
model calls.

## `internal/` — one package per concern

- **`backend`** — the `Complete` seam every provider implements (subpackages `anthropic`, `ollama`,
  and the test `fake`). A mode runner calls one method, never names a provider. Spec: `CLI.md`
  §Backends.
- **`brief`** — deterministic stdout "Needs you" report from verdict rows, no model call. `Qualify`
  is shared with `tree` so brief and tree agree on which claims need a human.
- **`embed`** — local Ollama embedding client (`nomic-embed-text`, L2-normalised), the semantic half
  of faithfulness retrieval; kept separate from the judge backend.
- **`manifest`** — held-document set from the `sources/` filesystem, keyed by canonical id; lets a
  run tell `absent` (checked, not there) from `unverifiable` (truth-maker missing).
- **`retrieve`** — corpus → role-tagged speaker-turn passages, ranked against a claim with BM25, so
  the judge sees only the passages a claim is about.
- **`tree`** — renders verdict rows three ways:
  - `tree.go` — report tree by §-heading path (`spec/TREE.md`);
  - `argument.go` — argument tree, internal node judgements DERIVED bottom-up from children, never
    authored (`spec/ARGUMENT.md`);
  - `root.go` — the root summary block above both.
  Each `.go` has a `_test.go` beside it.

## The axis boundary (why to trust some verdicts less)

- Close reading reaches FAITHFULNESS; dialectic reaches SUBSTANCE; reasoning can REFUTE a grounding
  claim but never CONFIRM one — positive grounding needs the truth-maker (a retrieval, not a
  deduction).
- `crossCheckEvidence` operationalises the boundary but does not abolish it: `supported` stays the
  structurally weakest verdict. Full envelope in `BACKGROUND.md` §2.3.

## The bookkeeping files (not code)

Three files carry the project's process discipline, each a different kind of record — don't confuse
them.

- **`REPORT.md`** — the "details go in a file, not the chat" convention (`CLAUDE.md:250-254`).
  Chat reports cap at 5 lines; tables, anomaly lists, provenance and per-command counts go into a
  `REPORT.md` under the relevant task directory. Not one canonical file — a recurring filename for
  the durable half of any task's output, colocated with the work.
  - `./REPORT.md` (root) — a one-off read-only investigation of the rigour-map problem statement,
    dated 2026-09-08; stale.
  - `examples/vic-lceic/**/REPORT.md` — per-run records beside each judge run's evidence directory
    (`evidence/<date>-.../<model>/REPORT.md`).
- **`SESSION.md`** — the **parking lot**: deferred and abandoned work, so reversals and dead ends
  survive the session boundary. Explicitly *not* a TODO backlog and *not* the change log — it's the
  holding pen for "decided to defer," with enough context to resume cold. Holds per-session what-
  shipped/what-carried blocks, withdrawn hypotheses, and the destructive-test queue (W1–W11) with
  each item's state. Newest dated entry 2026-09-07; treat its line references as drift-prone.
- **`rigour-map/decision_log.jsonl`** — the append-only decision/change log (pre-register the
  attempt when a change starts; fill in `outcome` later, including `abandoned`/`reverted`).
