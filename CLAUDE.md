# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

- **Build gate (run before every commit):** `./build.sh` — runs, stopping on the first failure:
  `golangci-lint run`, `staticcheck ./...`, `go test ./...`, `go build ./...`. All four must pass.
- **One package's tests:** `go test ./internal/client/`
- **One test:** `go test ./internal/client/ -run TestNonRetryableStatusIsError -count=1`
  (`-count=1` bypasses the test cache; use it for the red/green/break-it runs).
- **Run the binary:** `go build -o crossexam ./cmd/crossexam` then `./crossexam claims.txt`.
  `ANTHROPIC_API_KEY` must be set or it exits 1. Only substance mode is wired; `-source`,
  `-evidence`, `-audit`, `-md` are rejected loudly (see main.go).
- **Break-it check** (required when answering a code review): revert the one line the test guards,
  confirm the test goes red, restore it. Never claim a test has teeth without this.

Toolchain gotcha: two Go installs on this machine can mismatch (`go1.26.3` compiler under a
`go1.26.4` driver via a stale `GOROOT`). If `go build` reports a version mismatch, run with the
matching binary (`/usr/local/go/bin/go`) or unset `GOROOT`. Not a repo bug.

## Architecture

`crossexam` (module `github.com/peterwilliams97/elenchus2`) reads prose, splits it into atomic
claims, and grades each. Four modes are specified — substance, faithfulness, grounding, audit — but
**only substance is built**; the rest are stubbed and rejected at the flag boundary. The design
refuses to merge verdicts across modes (the "confidence laundering" target — see README.md, THEORY.md).

Dependency order (a package only imports ones above it):

- **`internal/client`** — the Anthropic API boundary and the *only* network seam. HTTP, retry
  (429/503/529, 4 attempts, Retry-After, jittered backoff), one parse-retry, truncation handling,
  usage accounting. `CallJSON(system, user, &out)` in, parsed JSON out. Tests inject `client.Stub`
  (fake.go) via the `Doer` interface — **no other package stubs the network.**
- **`internal/claims`** — pure types, no network, no global state. The verdict vocabulary
  (`type Verdict string` + named consts) and boundary validation (`ValidSubstanceVerdict`) live here;
  other packages import the consts and never re-spell the literals.
- **`internal/modes`** — mode orchestration and prompts. `Decompose` splits input into claims;
  `AssayClaim` runs the producer→critic loop (BEHAVIOR.md Mode 1). Prompts (prompts.go) are
  **verbatim from `spec/PROMPTS.md`**, co-located with the code that parses their output.
- **`internal/render`** — terminal and markdown output. Stateless: explicit values in, string out.
- **`cmd/crossexam`** — flags, input resolution (`-text` > file arg > piped stdin > default), the run
  loop, and the best-effort Tier-2 JSONL chain (chain.go, written under `eval/<stamp>-<model>/`).

`spec/` is the frozen contract (see below); the code implements it and tests check against its
GIVEN/WHEN/THEN examples. PLAN.md holds the build sequence; SESSION.md is the newest-first handoff log.

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

## Carried disciplines

- **fetch-or-STOP** — required external inputs are fetched; if unobtainable, STOP and mark blocked,
  never substitute.
- **provenance propagation** — synthetic input taints downstream; conclusions drawn from it are
  marked invalid.
- **red-then-green** — the failing test exists before the code that passes it.
- **pre-register decisions** — decisions are recorded before coding, not rationalised after.

## Testing

- **A test must fail when the code is wrong.** Break, on purpose, the exact thing the test
  covers and confirm the test fails. If it still passes, it tests nothing.
- **Assert the result, not just that the code ran.** "An error came back" or "it parsed" is
  not enough — check the value, the count, the message.
- **Pick the input where right and wrong code behave differently.** Count the attempts; put a
  boundary case exactly on the edge.
- **Coverage is not proof.** A line running is not a result being checked.
- **Write the failing test first, watch it fail, then make it pass.**
- **One test per spec rule** — each BEHAVIOR.md GIVEN/WHEN/THEN.
- **No network, no clock, no randomness.** Use the one fake (`internal/client`), set the retry
  delay to zero.
- **A test's name and its body must match.**

## Answering a code review

- A code-review fix is answered with the test code and the run output, never a prose "done".
  Show four things: the test, the run where it fails (red), the run where it passes (green), and
  the run where it fails again when the fix is reverted (the break-it check). No "all green"
  without the pasted `go test` output.

## Replace Degraded Claude Code

A running, itemised list of concrete failures in this repo's sessions — so degradation is recorded,
not waved away. Read it before working; do not repeat what is here. Newest first.

- **2026-06-15 — over-corrected the last plain-English pass: a dash-definition glued onto every
  phrase.** Fixing coined shorthand, REVIEW.md swung the other way: every phrase dragged a dash and a
  definition behind it ("built end to end — one feature taken all the way through so it runs from
  start to finish"; "Substance, the mode that tests whether a claim holds up under scrutiny or is
  hollow, is the only mode connected up so it runs"). The glosses say one idea two or three ways and
  bury the sentence. That is clutter, not clarity — the opposite failure from coined shorthand, same
  root cause: not trusting the plain words to carry the meaning. Fix: say the thing once, in short
  sentences, one idea each. Gloss a genuinely hard idea once (Frege, holism, the Given); for ordinary
  wording, write the plain words and add no gloss. The test for next time: if you are about to add a
  dash-gloss to explain a phrase, the phrase is wrong — replace it with the plain words and delete the
  gloss.
- **2026-06-14 — shipped coined in-house shorthand in a doc a newcomer can't read.**
  REVIEW.md used "substance vertical slice", "the slice", "wired"/"wired mode", and "walking-skeleton
  main()" — terms that only parse if you already built the repo. A reader who has never seen the code
  cannot tell what "the slice" is or what "wired" means here. Fix: say the plain words — "vertical
  slice"/"the slice" → "one feature built end to end"; "wired"/"wired mode" → "connected up so it
  runs"; "walking-skeleton main()" → "the bare start-to-finish wiring with most features stubbed out".
  The test for next time: if a phrase only makes sense to someone who already knows this repo, it is
  banned — write the plain words instead. (This is the writing rule in action: a phrase that needs the
  repo as its glossary is the same failure as one that needs an academic glossary.)
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
