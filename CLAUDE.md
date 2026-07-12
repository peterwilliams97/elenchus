# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`crossexam` splits prose into atomic claims and asks three separate questions of each — did the
source say it (**faithfulness**), is it well-formed and falsifiable (**substance**), is it true
(**grounding**) — and refuses to merge the answers. The target failure is *confidence laundering*:
spending a win on a question reasoning can reach (attribution, structure) as authority on the one it
can't (truth). Argument in THEORY.md; reader docs in README.md, OVERVIEW.md, CRITIQUE.md, LIMITS.md.

**Project name is `elenchus`** (git repo); `crossexam` is the binary; `elenchus2` is only the clone
dir and the Go module path (`github.com/peterwilliams97/elenchus2`). Don't name the project after
the module or the directory.

**Current build state:** substance mode only is wired. `main.go` makes `-source`/`-evidence`/
`-audit`/`-md` a **fatal error**, not a silent no-op — the README "Usage" block is the eventual
contract, not working software. PLAN.md §1 is the build sequence; §3 is the design queue.

## Build & test

```sh
./build.sh                              # the gate: golangci-lint → staticcheck → go test ./... → go build ./...; stops on first failure
go build -o crossexam ./cmd/crossexam   # build the binary
go test ./...                           # all tests (no network — internal/client fake is the only seam)
go test ./internal/modes -run TestAssay # a single package / single test
export ANTHROPIC_API_KEY=sk-...         # required to run the binary; fatal if unset
./crossexam claims.txt                  # substance mode: is each claim well-formed?
```

Lint gate (`.golangci.yml`): `govet`, `staticcheck`, `errcheck`, `gocyclo` (max complexity 15),
`funlen` (max 80 lines / 40 statements per function). Complexity and length caps are enforced —
split functions rather than suppressing.

## Architecture

Pipeline: raw prose → decompose into claims → grade each claim (producer–critic loop) → render +
persist a JSONL chain. Packages, in dependency order:

- **`internal/client`** — the *only* network seam (the Anthropic Messages API). HTTP transport,
  retry on 429/503/529 with full-jitter backoff (honors `Retry-After`), markdown-fence stripping,
  trailing-comma repair, one JSON parse-retry, usage/token accounting, and `stop_reason=max_tokens`
  → error. Callers use `CallJSON(system, user, &v)`. The API test fake (`Stub`/`DoerFunc` in
  `fake.go`) lives here and nowhere else; tests set `c.retryBase = 0` to zero the delay.
- **`internal/claims`** — pure types, no I/O. Owns the closed verdict vocabulary as named consts
  (`type Verdict string`) and validates verdicts at the boundary (`ValidSubstanceVerdict`) so a
  malformed verdict becomes `error`, never reaching render. Every package switching on a verdict
  imports these consts; no re-spelled literals.
- **`internal/modes`** — orchestration. `Decompose` splits input into claim strings; `AssayClaim`
  runs the producer→critic loop (`maxRounds` bound; `survives_only_by_conditioning` forces `hollow`).
  Prompts live in `prompts.go`, verbatim from `spec/PROMPTS.md`, co-located with the code that parses
  their output. JSON DTOs here are flat by design (they mirror the model's schema).
- **`internal/render`** — terminal + markdown output (`TermSubstance`, `Summary`, `UsageLine`) and
  the `-disagreements` cross-tab. Golden-output tests.
- **`cmd/crossexam`** — flag parsing (`flag.XxxVar` into named vars), input resolution (`-text` >
  file arg > piped stdin > built-in default), wiring, and the Tier-2 JSONL chain writer (`chain.go`,
  best-effort: chain failure logs to stderr, never aborts the run). Chains go to
  `eval/<stamp>-<model>/` by default.

Convention: **results to stdout; progress, SUMMARY, and USAGE to stderr.** Config is immutable after
`main()`; run-state lives in the function that runs (`runOpts`, not a growing struct).

## spec/ is frozen input

`spec/` was copied once, byte-for-byte, from v1 (`elenchus/spec/`) on 2026-06-13.
It is never edited, extended, or deleted.

Any session that finds `spec/` missing or modified **STOPS and reports** — it never
substitutes the v1 repo or memory as a source.

**One authorized deviation (d022, 2026-06-13):** the binary name was changed `assay`
→ `crossexam` in `spec/FIXTURES.md` (the two `./assay` run commands and the `/assay`
built-binary line only). v1 source identifiers (`assay.go`, `assayClaim`,
`assay_test.go`) and the chemical-assay narrative are unchanged. This is the **only**
permitted edit to `spec/`; finding any other modification still means STOP and report.

## Structural rules

- No new fields on shared structs. New state is passed as parameters. Any struct wanting
  a 7th field requires a split proposal, approved before coding. This applies to shared
  run-state structs, not per-call JSON DTOs (which mirror a fixed external schema — keep flat).
- Options are immutable after `main()`. Run-state lives in the function that runs.
- The API test fake lives in `internal/client` only. No other package stubs the network.
- Every session is feature OR consolidation, never both. Debt found mid-feature is logged
  for its own consolidation session, not side-fixed.
- Reviews are of code (the diff). A review of a report or doc must say so.
- Prompts in `internal/modes` are verbatim from `spec/PROMPTS.md`; any deliberate change
  to a prompt is a design-queue item (PLAN.md §3), never a port-time edit.
- A closed set of string values the code switches on or matches (an enum-like vocabulary) gets
  one typed definition (`type X string` + named consts) in the package that owns it. Other
  packages import those consts; none re-spells the literals. The compiler then catches a renamed
  or mistyped value; bare literals let it fall through a default silently.

## Writing

- **Plain English.** We are not writing for an audience of pompous academics hiding
  behind obscure language. If a phrase needs a glossary, rewrite it. Prefer the short
  word and the concrete one. "The first step everything depends on" beats "the
  load-bearing first step." The docs name hard ideas (Frege, holism, the Given) — name
  them plainly; the difficulty is in the idea, never in the wording.
- **Name things for what they do; never name a component in a way that confuses the
  reader.** A name must not collide with a more common meaning of the word. When it
  does, two readers get burned: the human, and the tool itself when it decomposes its
  own prose into fragments — in the repo's own dogfooding the critic read "assay" as a
  chemical assay (CRITIQUE.md, the atomism gap). If the best descriptive name is
  ambiguous out of context, gloss it on first use.
- **Call things by their name; don't invent names and don't write "the tool".** Things
  that already have a name (`crossexam`, the packages, the modes, the files) are
  referred to by that name, not by a coined label and not by a vague placeholder. If
  something genuinely has no name yet, describe what it does in concrete terms rather
  than minting a term and using it as if it were established.
- **Every number, metric, status, or tool term gets its plain meaning in the same
  sentence** — what it measures, and whether it's good or bad. Don't assume the reader
  knows Go tooling, the package names, or what a metric means. A raw figure or term a
  non-Go reader can't act on is not reporting; it's noise.
  - "coverage modes 92.5%" → "the tests run 92.5% of the `modes` package's code; the
    other 7.5% never runs under test"
  - "exit 0" → "passed"
  - "go vet clean" → "`go vet`, Go's built-in checker for suspicious code, found nothing"

## Testing

- A test must fail when the code is wrong. Check it: break, on purpose, the exact thing
  the test covers, and confirm the test fails. If it still passes, it tests nothing.
- Assert the result, not just that the code ran. "An error came back" or "it parsed" is
  not enough — check the actual value.
- Pick the input where right and wrong code behave differently. Count the attempts; put
  the boundary case exactly on the edge.
- Coverage is not proof. A line running is not a result being checked.
- Write the failing test first, watch it fail, then make it pass.
- One test per spec rule — each `spec/BEHAVIOR.md` GIVEN/WHEN/THEN.
- No network, no clock, no randomness. Use the one fake (`internal/client`), and set the
  retry delay to zero (`c.retryBase = 0`).
- A test's name and its body must match: a test called "non-retryable" must prove the
  status was not retried, not merely that an error came back.
- A code-review fix is answered with the test code and the pasted `go test` output, never
  a prose "done". Show the test, the run where it fails (red), the run where it passes
  (green), and the run where it fails again when the fix is reverted (the break-it check).
  No "all green" without the pasted output.

## Carried disciplines

- **fetch-or-STOP** — required external inputs are fetched; if unobtainable, STOP and mark blocked,
  never substitute.
- **provenance propagation** — synthetic input taints downstream; conclusions drawn from it are
  marked invalid.
- **red-then-green** — the failing test exists before the code that passes it.
- **pre-register decisions** — decisions are recorded before coding, not rationalised after.

## Replace Degraded Claude Code

A running, itemised list of concrete failures in this repo's sessions — so degradation is recorded,
not waved away. Read it before working; do not repeat what is here. Newest first.

- **2026-06-14 — reports shipped raw numbers and tool terms with no plain meaning.**
  Reports stated figures and Go-tooling terms — "coverage modes 92.5%", "exit 0", "vet
  clean" — with nothing saying what they measure or whether they're good or bad. A reader
  who isn't a Go developer can't tell. The test for next time: if a number or term would
  stop a reader who isn't a Go developer, explain it in the same sentence — what it
  measures, and good or bad. See Writing, "Every number, metric, status, or tool term".
- **2026-06-14 — bare verdict literals duplicated across `tierOf` and six clause maps.**
  `internal/render/disagreements.go` re-spelled the verdict strings (`"faithful"`, `"hollow"`, …)
  in `tierOf`'s switch and again in the six PASS/FAIL clause maps, with no compiler link to a
  single definition. A renamed verdict would fall through `tierOf`'s `default` to `NULL` with no
  error. Fix: one typed `claims.Verdict` (`type Verdict string` + named consts) owns the
  vocabulary; render imports the consts (maps are `map[claims.Verdict]string`), and nothing
  re-spells a literal.
- **2026-06-14 — used `flag.String`/`Bool`/`Int` (pointer-returning) instead of `flag.XxxVar`.**
  `cmd/crossexam/main.go` declared flags as `model := flag.String(...)` and then dereferenced `*model`
  everywhere. Bind flags with `flag.StringVar(&model, ...)` / `BoolVar` / `IntVar` into named vars
  instead — no `*` at every use site, and it matches how the spec describes the binding (spec/CLI.md:
  "defaults are read from the `flag.XxxVar` calls in `main`"). Use the `Var` form for all flag
  declarations.
- **2026-06-14 — used literary metaphor instead of saying what the sentence means.**
  "The analytic/synthetic distinction wearing a binary" in CRITIQUE.md does not say anything — it
  reaches for a clever image instead of a meaning. The fix: "cast as a binary verdict." Write the
  concrete thing. Do not use model training on self-indulgent writing — no metaphors, no
  personification, no borrowed cleverness. If you cannot say it plainly, you do not understand it.
- **2026-06-14 — named README.md after the tool, not the project.** README.md:1 was `# crossexam`
  (the binary name). The repo is `elenchus` (`git remote -v`: `git@github.com:peterwilliams97/elenchus.git`).
  A README title is the project name. The binary name belongs in the body.
- **2026-06-14 — left a doc file with no statement of what it is.** SESSION.md opened straight into
  dated entries with no line saying what the file is (a newest-first log of session handoffs). Every
  doc starts with a one-line statement of what it is and what it is for.
- **2026-06-14 — used the directory name as the project name.** Wrote "elenchus2 build plan" in
  PLAN.md. `elenchus2` is the directory this repo is cloned into; the project name is the git repo
  name, `elenchus` (remote `git@github.com:peterwilliams97/elenchus.git`). A git project's name is
  its repo name, never the clone directory. Check `git remote -v` before naming the project.
