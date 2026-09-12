# repo-map.md

Orientation for reading `assay` — where each concern lives and the key symbols that carry it.
Line numbers are from HEAD at the time of writing and drift; grep the symbol name if one has
moved. Authoritative behaviour is the `spec/` files, not this map.

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
