# OVERVIEW.md

A map of the whole repository: crossexam, its frozen specification, the decision log, the calibration
data, the example fixtures, the documents, and the build gate. crossexam is one part of the repo, not
the whole of it.

The binary is **not built yet.** All five `.go` files are package-comment scaffolds plus a no-op
`main`; `spec/` is the contract the built binary must satisfy, not a description of working software.

## crossexam — `cmd/crossexam` + `internal/*`

crossexam reads prose (a file argument, `-text STRING`, or stdin), splits it into atomic claims with
a `decompose` step, and runs one or more of four modes over those claims. It writes per-claim
verdicts to stdout (plain text, or markdown tables with `-md`), progress and a `SUMMARY` block to
stderr, and a per-claim JSONL verification chain to `eval/<stamp>-<model>/`. The default model is
`claude-sonnet-4-6` (override with `-model` or `ANTHROPIC_MODEL`); `ANTHROPIC_API_KEY` is required.

The four modes and their verdicts:

- **substance** (default) — is each claim well-formed and falsifiable? A producer call writes the
  claim's strongest defensible version; a critic call then attacks that version across seven named
  axes (evidence, hidden premise, falsifiability, equivocation, base rate and magnitude,
  counterexample, causality vs. correlation). Verdicts: `substantive` / `partial` / `hollow`.
- **faithfulness** (`-source FILE`) — did the source actually say it? Compares each summary claim
  against the source text. Verdicts: `faithful` / `partial` / `overstated` / `absent` / `contradicted`.
- **grounding** (`-evidence`) — is the claim true, checked against external evidence via web search?
  Verdicts: `supported` / `mixed` / `refuted` / `unverifiable`. `supported` currently means a cited
  URL was retrieved, not that a span on the page backs the sentence.
- **audit** (`-audit -source FILE`) — runs all three modes and cross-tabulates the results per claim.

The Go packages:

- `internal/client` — the Anthropic API boundary: HTTP transport, retry, response parsing, usage
  accounting, truncation handling. The only place the network is stubbed for tests.
- `internal/claims` — the claim, fragment, and verdict types; verdict-enum validation; JSON
  extraction; the `decompose`/split input handling. Pure: no network, no global state.
- `internal/modes` — orchestrates the four modes, driving the producer/critic loops and the audit
  cross-tabulation. The mode prompts live here as constants, next to the code that parses their
  output; they are to be ported verbatim from `spec/PROMPTS.md`.
- `internal/render` — produces the terminal (ANSI) and markdown output from result values.
- `cmd/crossexam` — parses flags and wires the packages together. Currently a no-op `main` stub so
  `go build ./...` links.

## `spec/` — frozen v1 contract (read-only)

Copied byte-for-byte from v1 (`elenchus/spec/`) on 2026-06-13 and not edited since, except the one
authorized binary-name change recorded in `CLAUDE.md`. It was extracted from v1's single Go file
(`assay.go`) and is the input contract the rebuild reproduces. Six files:

- `PROMPTS.md` — every system prompt, reproduced verbatim from the v1 source, with the user-prompt
  formats taken from the call sites.
- `CLI.md` — the flags, environment variables, input forms, and output formats, with v1 line numbers.
- `BEHAVIOR.md` — what the program does (orchestration rules), each traceable to a source line range
  with a GIVEN/WHEN/THEN example.
- `FINDINGS.md` — the operating envelope, calibration results, known bugs, and permanent limits, with
  provenance markers separating measured results from design expectations.
- `FIXTURES.md` — the manifest of every probe and worked example under `examples/`: what each tests,
  its provenance, its run protocol, and its file list.
- `LESSONS.md` — structural failures in the v1 code (e.g. the `cfg` god object) that the rebuild
  should not repeat, each with v1 line numbers as evidence.

## `rigour-map/decision_log.jsonl`

An append-only log, one JSON object per build decision, written **before** the work it describes.
Each record carries the task and phase, a hypothesis, the alternatives considered and why they were
rejected, the gut-preferred action vs. the action actually taken (`gut` / `acted`), whether the
action is reversible, and free-text notes. It records what each decision did and whether reasoning
the decision through changed the action from the first instinct. Currently six entries (d018–d023);
the `d0xx` numbering continues v1's sequence, so earlier numbers are not in this file.

## `testing/`

- `calibration_log.jsonl` — the calibration ledger, 21 append-only lines, one JSON object per probe
  run (fields include `date`, `model`, `fixture`, `mode`, `runs`, `verdict_counts`, `axis_mentions`);
  additional record types share the file. This is v1 baseline data the rebuild must reproduce.
- `SCHEMA.md` — documents that ledger's record types and explains the per-fragment false-pass metric
  (a defect-carrying fragment rated `substantive` with no axis finding) versus the run-level counts.

## `examples/`

Input fixtures with expected results, used as the rebuild's acceptance baseline:

- `dan_shipper/` — a worked example: `dan_shipper.txt` (a podcast transcript) and `dan_summary.txt`
  (12 predictions drawn from it).
- `destructive/` — eight probes, each an engineered defect crossexam should catch: `axis-gaps`,
  `bare-vs-contextualized`, `causal-narrative`, `hidden-premise`, `laundering`, `motte-and-bailey`,
  `reference-class`, `unfalsifiable-dress`. Each has its input, a `DEFECT.md`/`EXPECTED.md`, and dated
  `results/`; `run.sh` runs them N times per probe. The `laundering` source is gitignored and
  recreated from its `PROVENANCE.md`.
- `reflexive/` — crossexam run on its own claims (`claims.txt`, `headline.txt`) with dated `results/`.
- `url-length/` — a single worked example built around the claim "the maximum URL length is 2048".

## Documents

- `README.md` — what crossexam does, the four modes, usage commands, and current status.
- `THEORY.md` — the rationale for keeping faithfulness, substance, and grounding in separate columns,
  and why grounding reports a distribution over runs.
- `CRITIQUE.md` — a documentation-level critique of the repo's claims, read through analytic
  philosophy; a review of the docs, not of code.
- `PLAN.md` — the build sequence (one package per session), the acceptance test (reproduce v1's
  calibration distributions), and a design queue of open questions.
- `FAQ.md` — first-time-reader questions answered, with provenance to repo files.
- `FAQ.answers.md` — the internal draft those answers were written from, with provenance tags retained.
- `LIMITS.md` — the published failure probes and the measured limits and known biases they map.
- `CLAUDE.md` — repo-specific working rules: spec is frozen, the structural and carried disciplines.
- `SESSION.md` — dated session handoffs, newest first.
- `OVERVIEW.md` — this file.

## Build / test gate — `build.sh`

`build.sh` runs four steps in order and stops on the first failure: `golangci-lint run`, `staticcheck
./...`, `go test ./...`, `go build ./...`. `.golangci.yml` configures the linters; `go.mod` sets the
module path and pins the Go version. The binary is not built, so the gate exercises the scaffolds, not
working software.
