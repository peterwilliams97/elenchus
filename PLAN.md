# PLAN.md — elenchus2 build plan

Sources: spec/FINDINGS.md, spec/FIXTURES.md, spec/LESSONS.md, testing/SCHEMA.md. spec/ is frozen.
FINDINGS.md erratum: cites calibration_log.jsonl lines 1–22; the file has 21 — the log itself is
authoritative.

## 1. Build sequence

One package per session, in dependency order. Red-then-green throughout: the failing test exists
before the code that passes it. Standing gate, every session: diff review, then lint + `go test
./...` + `go build ./...` all green before the session closes. Sessions are feature or
consolidation, never mixed; debt found mid-feature is logged and gets its own consolidation
session, not a side-fix (LESSONS.md §4 is what mixing produced; no W-entry-style debt ships).

1. **internal/client** — Anthropic API client: HTTP, retry, JSON extraction, usage accounting.
   Tests: stubbed HTTP (the injectable seam LESSONS.md §1–2 demands), truncation routes to error.
2. **internal/claims** — claim/fragment/verdict types, decompose + splitSummary equivalents.
   Tests: verdict-enum validation at the boundary (kills W10 at birth), split/parse pure functions.
3. **internal/modes** — substance / faithfulness / evidence / audit orchestration + prompts
   (co-located with the code that parses their output). Tests: stub-driven loop termination,
   max-rounds, proposition substitution, audit alignment — plus a **mutation check**: invert each
   guard and confirm a test fails (a test that survives its own logic inverted is testing nothing).
4. **internal/render + cmd/assay** — terminal + markdown output, flag parsing, wiring. Tests:
   golden-output render tests; `Config` (immutable) separated from run-state per LESSONS.md §1.
5. **run.sh adaptation + smoke run** — port examples/destructive/run.sh to the v2 binary
   (chains persisted, never $TMP — instrument error d016), then one smoke probe end-to-end.

## 2. Acceptance test: reproduce v1's calibration

Before anything in §3: a full calibration run — same prompts (spec/PROMPTS.md, verified verbatim
against the binary), same fixtures (examples/ imported byte-identical per spec/FIXTURES.md, 44
files; laundering source re-fetched from the URL in PROVENANCE.md and verified against its
snippet — STOP and mark the probe blocked if unobtainable), same model
(`claude-haiku-4-5-20251001`), N=10 (canary N=5) — must reproduce v1's distributions within
sampling noise. Baseline: testing/calibration_log.jsonl + the dated results/ files in examples/.
False-pass counts are read per fragment_attribution records (testing/SCHEMA.md), never from raw
substantive totals; `error` verdicts are excluded from n_valid.

v1 numbers to match (fragment-attributed 2026-06-11 where it exists, else 2026-06-10):

| Probe                       | v1 distribution (valid fragments)                  | False-pass |
|-----------------------------|----------------------------------------------------|------------|
| motte-and-bailey            | N_valid=63: 41 hollow · 13 partial · 9 substantive | 1/63 |
| reference-class (06-10)     | 14 hollow · 16 partial · 0 substantive             | 0/30 |
| hidden-premise              | N_valid=36: 15 hollow · 13 partial · 7 subst. · 1 "substantial" | 1/36 |
| unfalsifiable-dress (06-10) | 30 hollow · 9 partial · 0 substantive              | 0/39 |
| causal-narrative            | N_valid=43: 18 hollow · 21 partial · 4 substantive | 0/43 |
| axis-gaps                   | N_valid=70: 55 hollow · 14 partial · 1 substantive | 1/70 (mapped-limit leak) |
| laundering (audit)          | faithful 10/10 · hollow 10/10 · refuted 10/10      | — |
| bare-vs-contextualized      | central claim hollow 10/10 in both inputs (fatal axis Counterexample 9/10 bare, 8/10 ctx); contextual aggregate 28 hollow · 21 partial · 11 substantive | criterion: central claim |
| reflexive canary            | unverifiable 5/5 — required every run              | any `supported` = stop |

A material divergence is a porting bug until proven otherwise. Expected qualitative anchors:
equivocation fires 10/10; reference-class and unfalsifiable-dress stay at 0 substantive;
category error stays the silent gap; the canary holds. Nothing in §3 starts before this passes.

## 3. Design queue (priority order, from spec/FINDINGS.md open questions)

**(a) Forward-prediction handling (the Counterexample question).** The Counterexample axis fires
fatally on probabilistic forward predictions because a conceivable counterexample is always
constructible — the bare and contextualized Ballmer forecast both rated hollow 10/10, fatal axis
Counterexample (9/10, 8/10), and the Evidence-penalty hypothesis was falsified, so this is a
critic design question, not a prompt tweak. Test: a graded pair — a categorical "no chance"
claim vs. a hedged quantified forecast of the same event. The fix works if the categorical form
still dies on Counterexample while the hedged quantified form escapes hollow.

**(b) Never-empty-critique prompt change.** Both confirmed false passes (motte run 10,
hidden-premise run 8) share one signature: a defect fragment rated substantive with an empty
critique — no axis fired at all. Test: a Layer-2 check that any substantive verdict with an
empty critique raises a warning, plus re-calibration of both probes showing false-pass 0 without
the clean-scaffolding substantive verdicts (8/9 and 6/8 of which were the critic being correct)
regressing toward hollow.

**(c) Referent anchoring for bare fragments.** Running substance on the repo's own claims showed
the critic reading "assay" as a chemical assay in 4 fragments and rating "two of assay's three
columns" hollow because "columns" has no referent out of context — decontextualised specialist
terms collide their referents and skew hollow. Test: the candidate referent-ambiguity fixture
from FINDINGS.md — a domain-polysemous term graded with vs. without a one-line domain anchor;
the fix works if the anchored and unanchored distributions converge.

**(d) Citation-block provenance.** `crossCheckEvidence` matches retrieved URLs by host+path,
proving a URL was retrieved, not that the cited span backs the sentence (W3; W2 makes
`supported` the structurally weakest verdict). True API citation blocks are never read. Test: a
fixture whose claim cites a genuinely retrieved page that does not support the sentence — today
it can pass URL-presence matching; with citation blocks read, it must not yield `supported`.
