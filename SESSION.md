# SESSION.md

A reverse-chronological log of session handoffs: what each working session changed and why, plus any
debt or carry-forward left for the next session. Newest first, one `##` section per session.

## 2026-06-14 — consolidation: license the repo (d036)

This entry records a consolidation session that added licensing only — no feature work, no
behaviour change, no other edits. Pre-registered as **d036**.

**The split.** Source code and build config under Apache-2.0; reader-facing documentation prose
under CC BY 4.0; the third-party inputs under `examples/` covered by neither. Copyright owner
Peter Williams, 2026.

**Files added at the repo root.** `LICENSE` is the verbatim Apache-2.0 text (fetched from
apache.org, 202 lines — byte-for-byte identical to the source, not reproduced from memory).
`LICENSE-docs` is the verbatim CC BY 4.0 legal code (fetched from creativecommons.org, 396 lines)
prefixed with a one-line statement of what it covers. `NOTICE` states the copyright and the split
in three plain sentences.

**Boundary (user-confirmed).** CC BY 4.0: README.md, THEORY.md, CRITIQUE.md, OVERVIEW.md, LIMITS.md,
FAQ.md, FAQ.answers.md. Apache-2.0: all `.go`, build config (.gitignore, .golangci.yml, build.sh,
go.mod), `spec/` (the extracted contract — LICENSE only states coverage; nothing is written inside
the frozen tree), and operational docs (CLAUDE.md, SESSION.md, PLAN.md, REVIEWER.md,
testing/SCHEMA.md, the decision and calibration logs). The three files the brief did not name
(REVIEWER.md, testing/SCHEMA.md, testing/calibration_log.jsonl) were confirmed as Apache.

**SPDX.** `// SPDX-License-Identifier: Apache-2.0` is now the first line of all 16 `.go` files
(blank line below it, so each package doc comment stays attached to its `package`). No copyright
block — one machine-readable line per file.

**README.** A new "Licensing" section near the end points at LICENSE, LICENSE-docs, and NOTICE.
The `# elenchus` title is unchanged.

**Gate.** `./build.sh` green: golangci-lint 0 issues, staticcheck clean, `go test ./...` ok across
5 packages, `go build ./...` ok — via `GOROOT=/usr/local/Cellar/go/1.26.4/libexec`. spec/ read-only
and untouched.

## 2026-06-14 — consolidation: one typed verdict vocabulary (d034)

Consolidation session (no feature work), pre-registered as **d034**. Extracted the verdict
vocabulary into a single typed definition and pointed the disagreement render at it.

**Two files changed.** `internal/claims/claims.go` gains `type Verdict string` + the full closed
vocabulary as named consts (verbatim from spec/BEHAVIOR.md "Verdict enums and routing", grouped by
mode; shared values partial/error defined once) — replacing the old substance-only 4-const block.
The rest of claims (Axis, SubstanceDetail, Substance, substanceVerdicts, ValidSubstanceVerdict) is
untouched. `internal/render/disagreements.go` now references the consts: `tierOf`'s switch types
over `claims.Verdict`, and the six clause maps are `map[claims.Verdict]string`. No bare verdict
literal remains in render code.

**The consts are deliberately untyped.** Typing them (`Substantive Verdict = …`) would break three
out-of-scope call sites that compare/assign against `string` — `claims`' own `substanceVerdicts`
map, `modes` (`verdict = claims.Hollow` into a string field), and `cmd` (`r.Verdict == claims.Error`).
An untyped string const is assignable to both `string` and `Verdict`, so it serves the existing
result path and the new typed maps/switch with zero edits outside the two files. End-to-end typing
(retyping `Substance.Verdict` to `Verdict` and threading it through modes/cmd) is deferred — a
larger change outside consolidation scope.

**Gate (behavior identical).** The existing `internal/render/disagreements_test.go` (64 tier-triples
+ verdict→tier map + rendered text) passes **unchanged** — zero edits to it — and the `&&`→`||` and
mis-tier mutations still fail. Full build/vet/test/lint green (5 pkgs, 0 lint issues) via
`GOROOT=/usr/local/Cellar/go/1.26.4/libexec`. spec/ read-only and clean.

**CLAUDE.md.** Added a Structural-rules bullet (enum-like vocabularies get one typed definition;
others import the consts, none re-spells literals) and a newest-first Replace-Degraded entry
(the bare verdict literals duplicated across tierOf + six maps, a renamed verdict falling through
to NULL silently).

## 2026-06-14 — substance check: ./crossexam runs one mode end to end (d033)

Feature session. The user re-scoped (after a Step-0 check found client/claims/modes were empty
stubs with nothing to wire) to build substance only across four packages in one session, one feature
end to end, deviating from one-package-per-session — pre-registered as **d033** before any code.
Bottom-up, each layer red-then-green against the `internal/client` fake (no network in tests).

**What shipped (substance mode only).**
- `internal/client` — bespoke raw-HTTP Anthropic client (not the SDK; matches the frozen spec and
  adds no dependency): `CallJSON` (parse + one retry with augmented system), HTTP retry on
  429/503/529 with full-jitter backoff + Retry-After, `max_tokens=1500`, truncation→error, usage
  accounting. The injectable seam is `Doer`; `Stub`/`DoerFunc` are the one network fake (exported so
  upstream tests reuse it — no other package stubs the network).
- `internal/claims` — substance verdict constants + `Substance`/`SubstanceDetail`/`Axis` types +
  `ValidSubstanceVerdict` boundary check (pure).
- `internal/modes` — the three substance prompts **verbatim** from spec/PROMPTS.md (verified
  byte-for-byte by script) + `Decompose` + `AssayClaim` (the producer-critic loop exactly per
  BEHAVIOR.md Mode 1: maxRounds, the four-part continue condition, the survives-only-by-conditioning
  hollow downgrade). Prompts live here per the package contract.
- `internal/render` — `TermSubstance` + `Summary` + `UsageLine`, string-returning to match the
  existing `Disagreements` idiom (keeps errcheck clean; caller does the IO).
- `cmd/crossexam` — replaced the no-op `main()`: flag parsing, `resolveInput` (-text > file > stdin
  > defaultInput), API-key check, run-state, dispatch, best-effort chain JSONL, and a fatal-with-
  message guard on -source/-evidence/-audit/-md (never a silent no-op).

**Proof, two ways.** (1) Loop/wiring tests through the fake at the modes and cmd
levels — a substance verdict comes out, no network. (2) Real `ANTHROPIC_API_KEY` smoke run on a
two-line file (`claude-sonnet-4-6`): both claims rated `partial`, 5 API calls, full SUMMARY/USAGE to
stderr, and a chain JSONL with the envelope + substanceDetail (steelman, per-axis critique). Actual
stdout/stderr pasted into the session report.

**Decisions settled in d033 (see the log):** raw HTTP over the SDK (frozen spec + no-new-deps +
HTTP-level fake); model default `claude-sonnet-4-6` per frozen CLI.md (not the claude-api skill's
opus default); `decompose` in modes not claims (purity + prompts-location); 7-key JSON schemas split
via embedded structs to honor the ≤6-fields rule; `est_usd` reported `n/a` (the v1 price table is
out of this repo's boundary — not invented); the ANSI `termSubstance` layout is a minimal readable
rendering (the spec pins only the markdown schema, out of scope).

**Gate:** gofmt / vet / `go build ./...` / `go test ./...` / golangci-lint all green (via
`GOROOT=/usr/local/Cellar/go/1.26.4/libexec` — the env-GOROOT bug from d031, still unfixed by the
user). spec/ read-only and clean.

**Out of scope, for follow-on sessions:** faithfulness, grounding, -audit, the -disagreements
wiring, -md, the est_usd price table, the ANSI/color output detail, and the §2 calibration run.

**Carry-forward (tooling hazard, recurring):** the markdown auto-formatter reflowed `spec/CLI.md`
(table alignment) after it was merely read — same on-read mutation that hit `spec/BEHAVIOR.md` last
session. Restored with `git checkout -- spec/CLI.md`; spec/ clean. The frozen-spec invariant remains
at risk from a format-on-save tool that does not exclude `spec/`; the guard suggested last session
(editorconfig/formatter ignore for `spec/`, or a pre-commit check that `spec/` matches HEAD bar the
d022 lines) is still not in place. Always `git status --porcelain spec/` before committing.

## 2026-06-14 — build the -disagreements render rule, red-then-green (d031, d032)

Feature session implementing d029 Q1 (the `-disagreements` rule). Red-then-green; `spec/` read-only
and clean.

**What shipped.** `internal/render/disagreements.go`: `tierOf` (verdict→tier via the spec's vcolor
routing — PASS/WEAK/FAIL/NULL), `surfaces` (a claim disagrees iff ≥1 column PASS and ≥1 FAIL),
`disagree`, and `Disagreements([]AuditRow) string` (renders only surfaced claims, each as one
plain-English question + its three verdicts, numbered by original position). RED first:
`internal/render/disagreements_test.go` written and run to a compile failure before any
implementation, then GREEN. The test pins the whole behaviour — all 64 tier-triples (independent
oracle + the arithmetic fact that exactly 18 surface + hand-written truth-table anchors), the
verdict→tier map over each mode's full vocabulary, the rule over real verdict strings, and the
rendered text. Verdict-triples are constructed directly; no model run, no dan_shipper as input.
Mutation-checked (`||` for `&&`, and a mis-tiered `faithful`, each fail the suite). Gate green: gofmt,
go vet, `go build ./...`, `go test ./...`, golangci-lint (0 issues). No new model calls; no new fields
on shared structs (`AuditRow` is a render input).

**Reported blocked, not faked (the STOP half of the brief).** The `-disagreements` flag and its wiring
into `cmd/crossexam` are NOT built: there is no flag parsing, no audit runner, and no audit-result
type to render from — `cmd/crossexam` is still `func main() {}`. Wiring needs PLAN §1 steps 1–4 (the
whole binary), which the brief forbids inventing this session. `render.AuditRow` is the adapter
boundary the future audit→render wiring will fill.

**Grounding deferred (d032, blocked).** The truth table grounds *what the rule computes*; it does not
ground *whether the rule surfaces the claims a reviewer cares about*. That needs audit output over a
corpus (N runs per input, distributions recorded), is blocked on the base build, and must never be
validated against a single fixture — dan_shipper stays an illustration in REVIEWER.md, never the
validation set. Pre-registered with its protocol.

**Docs.** REVIEWER.md Q1 gained a Status line; its two example blocks were updated to the actual
rendered output (the clause table uses "the source really says it", not the earlier looser
"Faithfully reported") — the feature's own doc, not an unrelated side-fix.

**Carry-forward (environment, reported not self-fixed).** A bare `go build`/`go test` fails with
`compile: version "go1.26.3" does not match go tool version "go1.26.4"`: an exported
`GOROOT=/usr/local/go` (a stale go1.26.3 tree) shadows the Homebrew `go1.26.4` binary
(`/usr/local/bin/go` → Cellar). Worked around this session per-command with
`GOROOT=/usr/local/Cellar/go/1.26.4/libexec`. The user's shell profile was not modified (repo
boundary). Permanent fix is the user's: unset/repoint `GOROOT`, or align go.mod. Note go.mod pins
`go 1.26.3`.

## 2026-06-14 — reviewer-facing design: -disagreements projection + review-to-rebuttal (d029, d030)

Research/design session, docs only. Two questions investigated against the frozen `spec/` and
pre-registered; no code, no spec edits (`spec/` clean, 6 files).

**Q1 — recognition output (d029, projection/cheap).** Proposed `-disagreements`: a render-only
flag over the data `-audit` already computes (the three per-claim verdicts handed to `progressDone`
and drawn by `mdAudit`). Surfaces only claims where, under the existing `vcolor` tiering
(PASS=faithful/substantive/supported, FAIL=hollow/absent/refuted/contradicted), at least one column
is PASS and at least one is FAIL — i.e. genuine conflicts a reviewer must adjudicate; unanimous and
merely-unsettled rows drop. Each surfaced claim renders as one plain-English question (deterministic
clause-per-column template, no model call) plus its three verdicts. Verified against
`examples/dan_shipper/`: surfaces #3/#5/#12, drops #6/#8/#9. Confirmed: no new model calls, no new
fields on shared structs — one new render function + flag wiring.

**Q2 — review-to-rebuttal (d030, new behavior/deferred).** Settled from spec alone: faithfulness
decomposes the *downstream* text, not `-source`, so `crossexam -source review.txt rebuttal.txt` keys
`absent` to *rebuttal* claims, not review points. "Which reviewer points went unanswered" needs the
reverse (source-decomposition + coverage) traversal — new behavior. Written up as **PLAN.md §3(e)**;
not implemented. No example pair drafted (that branch's condition — `absent` keyed to source points
— is false). The `-source rebuttal.txt review.txt` swap was considered and rejected as a non-fit.

Deliverables: `REVIEWER.md` (the proposal), `rigour-map/decision_log.jsonl` d029/d030, PLAN.md §3(e)
+ header note, this entry. Spec-silent points flagged in REVIEWER.md (cross-mode agreement is not a
spec concept; the tiering rule and error→NULL handling are pre-registered design choices, not spec).
Note: d028 was the README-title failure logged in CLAUDE.md (commit `dc3e324`) with no jsonl entry,
so this session continues the log at d029.

**Carry-forward (tooling hazard):** during this session a markdown auto-formatter reflowed
`spec/BEHAVIOR.md` (table alignment + line wraps, and it introduced a double-space in
`evidence,  what_source_actually_says`) after the file was merely read — no human or Claude edit. It
was restored with `git checkout -- spec/BEHAVIOR.md`; `spec/` ended clean. The frozen-spec invariant
is at risk from a format-on-save tool that does not exclude `spec/`; the next session should consider
a guard (e.g. an editorconfig/formatter ignore for `spec/`, or a pre-commit check that `spec/` matches
HEAD except the d022 lines). Always `git status --porcelain spec/` before committing.

## 2026-06-14 — fix project name (elenchus, not elenchus2); start failure log (d026)

PLAN.md:1 named the project by the clone directory (`elenchus2`); the git remote is
`peterwilliams97/elenchus`, so the project name is `elenchus`. Fixed PLAN.md:1, and added a
CLAUDE.md `## Replace Degraded Claude Code` section — a newest-first itemised failure log to read
before working — starting with this failure. `SESSION.md:59` left as-is (a factual go.mod
module-path statement, not a project-name claim). Docs only; `spec/` clean. Registered as **d026**.

## 2026-06-14 — name crossexam in OVERVIEW; "don't invent names" rule (d025)

Replaced the generic "the tool" with the actual name `crossexam` throughout OVERVIEW.md (intro,
section heading, mode paragraph, examples lines, README doc line), and added a CLAUDE.md Writing
rule: call things by their existing name, don't invent names or write "the tool". Docs only;
`spec/` clean. Registered as **d025**.

## 2026-06-14 — move §3c from atomism to externalism (d024)

Deleted the §3c sentence from CRITIQUE.md's atomism gap (it is reference-fixing, already covered
under externalism) and repointed PLAN.md §3c's cross-link from `#the-atomism-gap` to
`#the-externalism-gap`. Docs only; `spec/` clean. Registered as **d024**.

## 2026-06-13 — add OVERVIEW.md, a plain whole-repo map (d023)

Wrote `OVERVIEW.md` describing every part of the repo (tool, frozen `spec/`, decision log, `testing/`,
`examples/`, docs, build gate) by what it reads/does/emits, existing terms only, binary-not-built and
spec-as-contract stated; docs only, `spec/` clean. Registered as **d023** (the brief said "d021" but
d021/d022 were already used this session). Debt noted, not side-fixed: `examples/reflexive/README.md:3`
still has a stray "assay" in body prose; `go build` currently fails on a `go.mod` 1.26.3 vs installed
1.26.4 toolchain mismatch (environmental).

## 2026-06-13 — change binary name in frozen spec; drop reader-facing lineage (d022)

Two follow-ups to the d021 rename, both at the user's explicit direction.

**Dropped reader-facing v1/v2 rebuild lineage.** README (`## Status`, Examples) and FAQ (the whole
"## History / why a v2 rebuild" Q&A, the intro's acceptance-test aside, the LLM answer's "v1's
post-mortem") no longer narrate the repo's own history — readers don't care. Kept the honest
"binary not built yet" status. Also removed the self-referential rename notes added earlier in d021
(CRITIQUE.md, reflexive README, `doc.go`). Internal docs that legitimately reference v1 (CLAUDE.md
frozen-spec rule, PLAN.md acceptance test, LIMITS.md `Evidence:` provenance citations) were left.

**Authorized one edit to the frozen spec (d022).** The d021 rename deliberately left `spec/` frozen,
producing a spec/code divergence (spec invoked `./assay`, the build ships `crossexam`). The user
chose to resolve it at the source (AskUserQuestion: "Edit spec + amend rule"). Changed **only the 3
binary-name surfaces** in `spec/FIXTURES.md`: the two `./assay` run commands (L86-87) and the
`/assay` built-binary tree line (L308) → `crossexam`. **Unchanged:** v1 source identifiers
(`assay.go`, `assayClaim`, `assay_test.go` — Go symbols, not the binary) and the chemical-assay
narrative. The `## spec/ is frozen input` rule in CLAUDE.md was amended to record this as the single
authorized deviation, so future sessions don't STOP on finding `spec/` modified; any *other* spec
change still triggers STOP-and-report. `spec/` is no longer byte-for-byte v1 in those 3 lines; the
acceptance test is unaffected (binary name is not a behavioural element). Logged as **d022**.

## 2026-06-13 — rename tool `assay` → `crossexam` + writing-discipline notes (d021)

Renamed the tool/binary from **`assay`** to **`crossexam`** at the user's direction (`assay`
collided with "chemical assay" — the exact misread the tool's own critic made on its own docs; see
the reflexive pass and the atomism gap). Name chosen by the user from {claimgrade, crossexam,
claimsort}: it names the producer-critic *method* (adversarial cross-examination; v1's "elenchus" is
the obscure Greek for it) and, unlike `claimcheck`, does not read as a fact-checker. Pre-registered
as **d021** before any edit.

**spec/ stays frozen and was NOT touched** — `spec/CLI.md`, `spec/FIXTURES.md`, `spec/PROMPTS.md`
still invoke `./assay`. The resulting **binary-name divergence is accepted and recorded** (the user's
chosen option), not hidden. Code side was trivial: all 5 `.go` files are package-comment scaffolds +
a no-op `main`; module name is `elenchus2`, unaffected. `git mv cmd/assay cmd/crossexam`; two doc.go
comments, `.gitignore` binary paths, and `README.md` build/usage updated.

**Rename vs. preserve (provenance discipline).** Renamed: tool identity in prose, runnable `./assay`
commands, the `cmd/assay` structural ref, and orphaned verb uses ("assayed" → "examined"). **Preserved
byte-for-byte:** dated run records (`examples/*/results/*`, `testing/calibration_log.jsonl`), v1
`assay.go:NNN` source citations, the whole **sealed reflexive experiment** (`examples/reflexive/` —
its `claims.txt` is graded input and the finding *depends* on the word "assay"; added one orienting
header note only), and every chemical-assay *cautionary* reference (CRITIQUE/FAQ/LIMITS/PLAN). The
critique's atomism-gap section gained a note: the rename is itself the "one-line anchor" move it warns
about — it fixes one instance, not the architecture. `FAQ.answers.md` (a superseded internal draft)
was left as-is.

**Also this session (writing discipline, per user):** added a `## Writing` section to repo `CLAUDE.md`
— plain English (no pompous/obscure language), and name components for what they do / never in a way
that confuses the reader; rewrote CRITIQUE.md's "load-bearing first step" into plain prose.

**Build state (reported, not fixed):** `go build ./...` exits non-zero, but the cause is environmental
and pre-existing — a toolchain mismatch (`go.mod` pins `go 1.26.3`; installed tool is `go1.26.4`),
failing the *standard library* compile, independent of this rename. The `build.sh` staticcheck change
remains uncommitted and unbundled (out of scope, the user's call), as noted at d020.

## 2026-06-13 — FAQ formatted into FAQ.md (consolidation, docs only, step 4 of 4)

This session formatted the reviewed `FAQ.answers.md` draft into the reader-facing **FAQ.md** (repo
root), the final step of the FAQ workflow. The approach was **verify-then-format, not
trust-and-polish**: review had found a fabricated citation in the draft, so every surviving provenance
tag was re-checked against its actual source file as the answer was lifted, then the internal
`[GROUNDED]/[CONTRACT]/[OPEN]/[LIMIT]` tags were stripped from the reader-facing text while the honesty
they encoded was preserved in prose. The decision was pre-registered as **d020** (same meta handling as
d018/d019: doc/meta, review-of-docs-not-a-diff, `label=irrelevant`, outcome abstained). A link-checker
verified all 23 in-repo anchor links in FAQ.md resolve to real headings.

**Git/provenance state at session start (reported, not fixed).** The d018 doc work *is* committed
(`a6e25b2` pre-register, `e81c767` critique+cross-links, `939a5ac` reflow). Uncommitted at start were
the d019 deliverables (`FAQ.answers.md` untracked, the `SESSION.md` edit, the d019 log line) plus two
unrelated pre-existing changes not authored by these sessions — the `CRITIQUE.md` "PLAN.md §3"
back-link and the `build.sh` staticcheck step.

**Reviewer corrections applied.** (a) The LLM-agnostic answer's claim crediting `CLAUDE.md` with "one
wrapper per external dependency" was **dropped as unverifiable** — that phrase is in neither the repo
`CLAUDE.md` nor `spec/LESSONS.md` (it lives only in the global `~/CLAUDE.md`, not a repo file). The
feasibility point was re-grounded to `spec/LESSONS.md §2` (the `api/` package boundary) plus the modes
already consuming typed JSON shims (`substanceJSON`/`faithJSON`/`evidenceJSON`, in `spec/PROMPTS.md` &
`spec/BEHAVIOR.md`); the valid `spec/CLI.md` + `spec/PROMPTS.md §6` grounding was kept. (b) The
fact-checker/judge/RAG answer's "explicitly disclaims that the distribution is a property of the claim"
was **softened** to THEORY.md's narrower statement — the false-pass *rate* is a property of the
distribution, not the claim, while repeated runs still read as "what the claim tends to do" — and the
[psychologism-gap](CRITIQUE.md#the-psychologism-gap) link (that very equivocation) was kept. (c)
Confirmed `LIMITS.md` carries a `## Operating envelope` section, so those citations stand and did not
need repointing to `spec/FINDINGS.md`.

**Tag report (changed / dropped / unverifiable).** One tag was unverifiable and dropped: the
`CLAUDE.md §API discipline` citation (correction a), with the claim re-grounded as above. One claim was
softened (correction b). All remaining internal tags were stripped by design, each underlying claim
re-verified against its source this session; no unverifiable claim was carried into the published FAQ.

**Reader-facing form.** Section order unchanged (Goals / What it is / Problem statements / Method
choice / Trust & cost / Limits & scope / History / LLMs / Uses). The eight design-analysis answers
(dialectic-alternatives, producer-critic-alternatives, the fact-checker/judge/RAG comparison,
why-ship-grounding, and all four LLM answers) are framed as reasoned design analysis, not settled fact
— seven with an explicit "_Design analysis, not a settled repo decision:_" lead, and measure-performance
with an inline qualifier because its core instrument (the calibration protocol) is grounded in PLAN.md
§2 and only the cross-model extension is analysis. The binary-not-built / "cannot run today" /
"trust this least" honesty is retained, the reader-facing `([critique](CRITIQUE.md#…))` and
THEORY/PLAN/LIMITS links are kept, and Uses carries one contract disclaimer at the top plus each
example's "limit that bites." No answer was re-answered or expanded; no new questions were added.

**Constraints / carry-forward.** Docs only — no code/test/prompt/spec edits; `spec/` confirmed clean
(6 files). `FAQ.answers.md` (the draft) was left in place untouched. The unrelated `CRITIQUE.md`
back-link and `build.sh` staticcheck change were **left unbundled and unreverted** by design — their
disposition (commit or revert, each on its own) is the user's call, not this session's.

## 2026-06-13 — FAQ answer pass drafted (consolidation, docs only)

This session drafted **FAQ.answers.md** (new, repo root): grounded, provenance-tagged answers to
every first-time-viewer FAQ question, in order, under the supplied section headings (Goals / What it
is / Problem statements / Method choice / Trust & cost / Limits & scope / History / LLMs / Uses).
This is the **answer pass** (step 2 of the 4-step FAQ workflow) — a reviewable draft, not the final
FAQ. Each answer carries at least one provenance tag: `[GROUNDED]` (supported by the repo),
`[CONTRACT: binary not built]` (promised behaviour, not observed — the binary is not built, README
§Status, and internal/*/doc.go are scaffolds), `[LIMIT]` (a CRITIQUE.md / LIMITS.md / README limit),
or `[OPEN: design analysis]` (my reasoning, separated from what the repo does). No invented numbers:
the only quantitative results cited are the repo's own (Ballmer N=10 row; PLAN.md §2 distributions).
The decision was pre-registered before drafting as **d019** in `rigour-map/decision_log.jsonl` (same
handling as d018: doc/meta work, review of docs not a diff, `label=irrelevant`, outcome abstained,
kept out of the rigour gut-vs-acted signal). All 8 cited CRITIQUE.md anchors were verified to resolve
to real headings.

**GROUNDED vs OPEN split.** Grounded (anchored to repo text): the goals, the three-mode definitions
and when to pick each, the "is it a fact-checker" answer, what-you-see (CONTRACT), the
problem-statement subtlety (substance is the mode whose referents must be supplied — THEORY.md +
PLAN.md §3c + CRITIQUE.md#the-atomism-gap), the trust/reliability ranking (LIMITS.md operating
envelope), the N=10-is-calibration-not-mandatory answer, the limits/scope answers, the
v1→v2 history, and "not finished." OPEN design analysis (reasoned, marked, kept separate from "what
the repo does"): the dialectic-vs-alternatives comparison, Producer-Critic alternatives, the
fact-checker/LLM-judge/RAG comparison, all four LLM questions (agnostic refactor, performance
measurement, removing LLMs, local model), and the "why ship grounding" rationale. The six Uses
examples are CONTRACT-tagged worked micro-examples (decompose → three-questions flow + the documented
limit that bites each); the science/spectra one is explicit that crossexam validates **prose claims, not
spectra**, with the Given and externalism gaps biting hardest. Two questions were softened to honest
non-answers rather than guessed: usefulness to untrained users (no field data) and the deeper "why
rebuild v2" motivation (not argued in the v2 docs) — both flagged `[OPEN]`, none required a STOP.

**Step 4 (format into FAQ.md) is pending review.** FAQ.md was NOT touched. spec/ confirmed clean; no
code/test/prompt diffs. Carry-forward: two pre-existing, unrelated working-tree changes I did NOT
author and left unbundled — `build.sh` (a `staticcheck ./...` step, already flagged below) and
`CRITIQUE.md` (a one-spot inline link added to "PLAN.md §3"); commit or revert each on its own.

## 2026-06-13 — doc-level philosophy critique recorded (consolidation, docs only)

This session installed a critic against the repo's own documentation and recorded its findings
faithfully rather than answering them. **CRITIQUE.md** (new, at repo root) holds nine sections
reading THEORY.md / README.md / PLAN.md through analytic philosophy from Frege forward — atomism,
analytic/synthetic separation, rule-following, the Myth of the Given, psychologism, speech-act
force, externalism, and the charge that the repo's own load-bearing thesis ("reasoning can refute
but never confirm") is itself an ungrounded substance-column claim. Seventeen inline
`([critique](CRITIQUE.md#…))` links were wired from the most questionable claims back to it (8 in
THEORY.md, 6 in README.md, 3 in PLAN.md); all anchors verified to resolve. The decision was
pre-registered before any edit as **d018** in `rigour-map/decision_log.jsonl` (this is the first
entry in elenchus2's own log — schema carried from v1's `rigour-map/decision_log.jsonl`, which the
user supplied for schema only and which was not modified). No code, test, prompt, or spec/ change;
spec/ confirmed clean.

**The split that is the decision.** The critique sorts into two kinds. *In-principle, not patchable:*
the Given gap (grounding-as-retrieval can never deliver a non-inferential truth-maker), the
analytic/synthetic separation taken as a universal claim, and holism vs. per-fragment grounding —
no prompt or fixture closes these, and pretending otherwise re-introduces the very laundering the
tool exists to refuse. *Tractable by narrowing scope:* the rule-following, speech-act-force,
externalism, and referent gaps are failures of the OPEN-world ambition; they shrink or vanish once
the input domain is closed and claim types are restricted, because context-dependence and
externalism stop biting when context is supplied by construction. Notably, PLAN.md §3's existing
open problems (§3a rule-following, §3c the context principle, §3d/W2–W3 the Given) are these same
critiques already surfacing as engineering work.

**Two options now on the table.** (1) **Reduce scope** — apply the producer–critic method to a
narrow problem class (fixed domain, anchored referents, claim types where the axes genuinely apply:
internal-contradiction refutation, attribution-fidelity over a known source) where the gaps are
designed out rather than apologised for. (2) **Drop the demand for complete answers** — reposition
the tool as producing partial results and targeted improvements (the refutation and faithfulness
flags it CAN earn), and either stop claiming the grounding column or mark it permanently
provisional. The current docs promise complete, separable, three-column adjudication; the
defensible product is a sharp instrument for the columns reasoning can actually reach. No option is
chosen here — recording the fork was the deliverable.

### Carry-forward / flags for the next session
- **Unrelated working-tree change present:** `build.sh` carries a `staticcheck ./...` step that was
  NOT authored this session and is NOT part of this doc-only change. It was left untouched and
  unbundled (one-logical-change-per-commit). Commit or revert it on its own.
- **Anchor slug note:** the heading "The analytic/substance vs. synthetic/grounding gap" slugs to
  `#the-analyticsubstance-vs-syntheticgrounding-gap` (lowercase, punctuation dropped, spaces→hyphens
  — the `/` is removed, not hyphenated). The links use this heading-derived slug, which differs from
  the originating instruction's hyphenated table form; the heading was preserved verbatim, so the
  links were corrected to match it (the one permitted slug repair).
- **An Author's response was NOT added** to CRITIQUE.md. The critique body stands unanswered by
  design; if a response is ever wanted it must be fenced, last, and must not soften or shorten the
  body.
