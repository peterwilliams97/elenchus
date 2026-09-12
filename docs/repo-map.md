# repo-map.md

Orientation for reading `assay` — where each concern lives and the key symbols that carry it.
Line numbers are from HEAD at the time of writing and drift; grep the symbol name if one has
moved. Authoritative behaviour is the `spec/` files, not this map.

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
  elenchus on a target.
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

## Organization opportunities (top-level sprawl)

Assessed 2026-09-12. Reference counts are `git grep -l` over tracked files; treat as drift-prone.
Nothing here has been actioned — this is a to-decide list, not a change record.

### Quick wins — low risk
- **`tree.txt` — delete from tracking.** A tracked 480-line filesystem dump (last touched
  `4dfd751`) that drifts the instant any file moves and is not documentation.
- **Untracked scratch is already gitignored** (`demo-dan.log`, `diff.txt`, `eval-value.stderr`,
  `eval-value.tree`, `eval/`) — no repo change needed; `rm` locally for a clean listing.
- **`FIXTURE_REPORT.md`** — dated 2026-06-01, 1 reference. Stale one-off; fold into `testing/` or
  `docs/`, or delete.

### Medium — real churn, do selectively
- **Root markdown splits by reference weight.** Keep at root (living docs + convention):
  `README.md`, `CLAUDE.md`, `BACKGROUND.md` (11 refs), `TESTING.md` (13), `SESSION.md` (11) —
  moving each edits 11–13 files of cross-links. Relocate candidates into `docs/` (cheap):
  `TODO.md` (2), `REPORT.md` (4), `FIXTURE_REPORT.md` (1).
- **Three test-ish dirs read as one to a newcomer** — `eval/` (gitignored scratch run outputs),
  `testdata/` (Go-convention fixtures), `testing/` (tracked calibration logs). Distinct, but the
  names don't say so; a one-line header in each closes the gap cheaply.

### Structural — biggest opportunity, a decision not a tidy-up
- **The Go layout departs from the `~/CLAUDE.md` `cmd/` + `internal/` standard** (no `.go` at the
  repo root). Here the root carries `assay.go` (~3,900 lines) plus four test files
  (`assay_test.go`, `judge_test.go`, `oracle_gen_test.go`, `refuter_b_test.go`). Clean target:
  `cmd/assay/main.go` for flag parsing / `serve` dispatch, the judge path extracted into an
  `internal/` package, `assay.go` split by concern like the rest of `internal/`.
- **It contradicts the repo's own written convention.** `elenchus/CLAUDE.md` deliberately documents
  `assay.go` at root as the CLI, and `build.sh:8` hard-codes `go build -o assay .` — so a move
  breaks `build.sh`, the README invocations, and spec references in the same commit. This is a
  decision to make on purpose (Peter's call), not a cleanup-pass side effect.

## Adding an OpenAI backend (feasibility, 2026-09-12)

The backend seam is built for this — a new provider is a subpackage, no mode/judge/tree change.

- **Seam.** Every provider implements the 4-method `backend.Backend` interface (`backend.go:59`):
  `Name`, `Model`, `SupportsWebSearch`, `Complete(Request) (Response, error)`. Mode runners call
  `Complete` and never name a provider (`spec/CLI.md` §Backends is the contract).
- **Scope.** New `internal/backend/openai/{openai.go,openai_test.go}` (~250 lines, modeled on the
  263-line `anthropic.go`), one `case "openai":` in the dispatch at `assay.go:319-328`, and a
  `spec/CLI.md` §Backends update in the same commit. No new dependency — `anthropic.go` hand-rolls
  raw HTTP/JSON against the API URL, so OpenAI should too (`one wrapper per dependency`). Test via
  the injected `*http.Client` RoundTripper seam (`New(model, apiKey, httpc)`), no live calls. ~1 day.
- **Field parity.** `Schema`/`SchemaName` → `response_format:{type:"json_schema",strict:true}` —
  *easier* than Anthropic's forced-tool-call because the response is JSON text, no `tool_use`
  unwrap. `Temperature` → `temperature`. `Cached` folds into the prompt (OpenAI auto-caches
  ≥1024-token prefixes; no explicit field). Usage: `prompt_tokens`→`InputTokens`,
  `completion_tokens`→`OutputTokens`, `cached_tokens`→`CacheRead`, `CacheCreate`=0.
- **Three gotchas.**
  1. Reasoning models (o-series, gpt-5 reasoning) use `max_completion_tokens`, reserve output
     budget for hidden reasoning, and reject `temperature`; a 1500-token `MaxTokens` cap truncates
     them. Target a standard chat model (gpt-4.1 / gpt-5 non-reasoning) first.
  2. Web search is the one non-trivial piece: `crossCheckEvidence` compares model-reported
     citations against actually-fetched URLs (Anthropic exposes these via `web_search_tool_result`
     blocks; OpenAI's annotations carry different provenance semantics). Ship the first cut with
     `SupportsWebSearch() bool { return false }` — the schema-enforced judge path is the main one,
     and deferring grounding matches the project's treatment of it as the weakest column.
  3. Key/model plumbing: needs an `OPENAI_API_KEY` analog to `ANTHROPIC_API_KEY`/`setkey.sh`, and a
     model-env name alongside `ANTHROPIC_MODEL`.
- **Naming.** "Codex" as a model is retired; this is an OpenAI *chat-model* backend
  (gpt-4.1 / gpt-5 / o-series). "Codex" now names OpenAI's CLI/agent, unrelated to this seam.

## Reference — future-work material (external, not in this repo)

`../elenchus_material/` (sibling of the repo, **machine-local path — not committed here, not on any
remote**) holds the source material the next build phase draws from. The plan is to build as much as
possible *from the PDFs*; Peter is doing the prep of deciding which parts of each to analyse, so
**the PDFs are unread by design** until that selection lands. Read-only from this repo — never write
there.

- **`suriving_ai_coding,md`** (authored, v8 2026-09-06) — READ. "Surviving human-to-AI coding:
  master plan," the conceptual parent of assay's stance. Its vocabulary is already half-encoded in
  `CLAUDE.md`: **checker/statement/refuter** (= the never-weaken-tests + seen-to-fail rules,
  `refuter_b_test.go`), **"a test must say when it didn't look"** (= the zero-output rule and
  `absent` vs `unverifiable`, `internal/manifest`), **producer/partition/differential testing**
  (= the `-source` corpus and item-7 segmentation). Its central claim — "a check is only as good as
  the inputs you feed it," ordered easy→hard→hardest (documents / WPP / security) — is the same
  wall as `TODO.md` item 5: a consistency tester is buildable now, a correctness tester is blocked
  on a real input population that must be obtained, never synthesised.
- **`Anthropic-Detecting-and-countering-091026.pdf`** — UNREAD (pending Peter's selection).
- **`A Model for Organizational Interaction.pdf`** — UNREAD (pending Peter's selection).
- **`ai_index_report_2026.pdf`** — UNREAD (pending Peter's selection).

## Plan of record — the retrieval segmentation seam (TODO items 2/6/7)

The cross-cut that `TODO.md` items 2, 6, 7 share is narrower than "`retrieve.go` only knows
hearings." Everything downstream of a `retrieve.Passage` is already corpus-agnostic (`Index`,
`build`, `bm25Scores`, `Search`, `tokenize`, `Format`, `ByIDs`). Three of the four parsers are
already generic prose splitters — `submissionPassages`, `qonPassages`, and `reportPassages`
(`retrieve.go:202`) tag `Source` and leave `Role`/`Context` empty; only `splitFile`
(`retrieve.go:302`) is Hansard-specific (speaker turns, `parseRoster`, `classify`, Q/A `Context`).
Role already degrades cleanly: `Format` (`retrieve.go:533`) branches on `Source`, so role-less
passages render fine and the judge never sees a missing role.

The one coupled point is `passagesForFile` (`retrieve.go:181`): it picks a parser by hardcoded path
substrings (`/submissions/`, `/qon/`, `report.txt`, else Hansard). That implicit switch is the only
thing you edit to add a corpus type.

### The seam (item 7 — do first)

Replace the substring switch with an explicit ordered registry:

```go
type Segmenter interface {
    Match(path string) bool                           // does this segmenter own the file?
    Split(path string, data []byte) ([]Passage, error)
}
```

`passagesForFile` reads the file once, walks a registered `[]Segmenter` (first `Match` wins), Hansard
last as default. The four existing parsers become four `Segmenter` values; registration order =
today's switch order. `Index`/`Search`/`Format` are untouched. This isolates the Hansard
role/context logic inside one segmenter rather than generalising it.

It is a **behaviour-altering-nothing change**, so (per `../elenchus_material/` master plan) the
refuter is near-free: golden-test the vic-lceic corpus — capture the current `[]Passage`, refactor,
assert byte-identical output. The old code is the test.

### After the seam

- **Item 6 (other reports)** — closer to done than first stated: a plain report already flows through
  `reportPassages` via the `report.txt` suffix. The real blocker is the page-number assumption —
  `reportPassages` maps `printed page = form-feed index + 1`, valid only "for a report whose PDF has
  no front-matter offset" (`retrieve.go:199`). A report with a cover/TOC (likely
  `examples/quocirca-2026/`) makes the `#page=N` links in the `-review` site (`spec/SERVE.md`) off by
  the offset. Fix: make the page offset a segmenter parameter (or read it from the manifest). Plus
  ordinary prep — a `MANIFEST.md` and decomposed claims for the new report (real inputs, never
  synthesised).
- **Item 2 (critique of code)** — the seam gives it a home, not the work: a code corpus needs a
  segmenter whose passages are Go declarations (`go/ast`, one passage per top-level decl). Genuinely
  new — a different producer (code, not hearings), so a different partition with its own segmenter.

### Sequence

7 (seam + golden refuter) → 6 (report-offset parameter + quocirca manifest/claims) → 2 (Go-decl
segmenter). One refactor unblocks the routing for all three; 6 and 2 then differ only in which
segmenter they add. If the seam is written up as a contract, it belongs in a new `spec/RETRIEVE.md`
(there is none today — retrieval is specified inside `spec/CLI.md` §Retrieval).

## Design starting point — is surviving refutation the best validation? (2026-09-12)

A seed for design work, not settled. Peter: "I don't know of a better validation of any claim than
surviving refutation." The position is mostly right and its exact failure is the line `assay` is
built on. (This is design rationale — it may graduate to `BACKGROUND.md`.)

- **Where surviving refutation IS the ceiling.** Claims with no reachable truth-maker — theories,
  forecasts, universals. Popper: science corroborates, never verifies. Nothing beats attacking the
  claim and watching it stand.
- **But refutation is asymmetric — a decisive *falsifier*, never a decisive *validator*.** One
  counterexample kills a universal (modus tollens); survival proves nothing (affirming the
  consequent). So the honest word is *corroboration* — "not yet shown false, and it stuck its neck
  out" — not *validation*. Upgrading "un-refuted" to "validated" is the "laundering confidence"
  failure in `CLAUDE.md`.
- **For a correspondence claim with an accessible source, retrieval strictly dominates.** "The
  report cites D on page 12" — open page 12; no amount of armchair refutation substitutes. This is
  the axis boundary verbatim: reasoning can refute a grounding claim but never confirm one. `assay`
  doesn't out-argue "the source says X" — it reads the source. The tool is a monument to
  refutation-survival being insufficient here.
- **Survival's strength has an invisible bound: the refuter's imagination.** Surviving refutation
  only means surviving the refutations someone thought to try — a lower bound set by the adversary,
  not a property of the claim. Hence the repo's own rules: a refuter must name (not count) its cases;
  Producer and Critic share one `c.model` so a blind spot survives in both; `survives_only_by_
  conditioning` catches a claim that survived by dodging into unfalsifiability rather than by being
  true.
- **The reframe.** Refutation and grounding aren't rival validators — refutation is the cheapest way
  to establish *falsity* (no lookup), retrieval the only way to establish *positive support*. "Best
  validation" is claim-type-dependent, and the axis boundary already draws that map. Keeper:
  **surviving refutation is the best test of a claim's coherence and its nerve; it is never, on its
  own, evidence the claim is true — for that, the retrieval half exists.**
