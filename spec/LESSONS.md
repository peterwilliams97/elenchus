# LESSONS.md — Structural Failures to Not Repeat

Each failure is stated plainly with the line numbers as evidence. These are architectural problems
in the current codebase that the new repo should not inherit.

---

## 1. The `cfg` god object

`assay.go:46–68` — the `cfg` struct has 15 fields mixing three completely different concerns:

```go
type cfg struct {
    model        string          // setting
    apiKey       string          // setting
    maxRounds    int             // setting
    maxClaims    int             // setting
    verbose      bool            // setting
    noColor      bool            // setting
    asMarkdown   bool            // setting
    showProgress bool            // setting
    quiet        bool            // setting
    usageOut     string          // setting
    chainFile    string          // ← live run-state: "set by runners before case loop"
    usage        *usageCounters  // ← live run-state: accumulated during execution
    tally        *runTally       // ← live run-state: accumulated during execution
    call func(...)               // ← test hook: overrides production dispatch
    httpClient *http.Client      // ← test hook: overrides production HTTP client
}
```

Settings (`model`, `maxRounds`, etc.) are fixed for the lifetime of a run and belong on a config
object. Live run-state (`chainFile`, `usage`, `tally`) accumulates during execution and is mutated
by every caller — this is shared mutable state, and the comment "set by runners before case loop"
(assay.go:57) is the exact temporal-coupling smell: a field that has no valid value until some
other code runs first and sets it. Test hooks (`call`, `httpClient`) exist to override production
dispatch without a network and belong in an injectable dependency, not in the same struct as
`model` and `quiet`.

Because every function is a method on `cfg` (`assayClaim`, `faithClaim`, `evidenceClaim`,
`runSubstance`, `runFaithfulness`, `runEvidence`, `runAudit`, `callJSON`, `callJSONSourced`,
`callClaude`, `progressStart`, `progressDone`, all terminal renderers — roughly 20 methods), there
is no boundary between the configuration layer and the execution layer. A function that needs only
`model` and `apiKey` still carries a reference to `tally` and `call`. A test that wants to stub
the API must build the entire cfg.

The fix in the new repo: separate `Config` (immutable, value type, set at startup) from `Run`
(mutable state for one execution — tally, usage, chainFile) from `testOverrides` (call, httpClient
— injectable at construction, not part of the domain type). Functions should take what they
actually need, not `*cfg`.

---

## 2. The single-file rule removed package-boundary pressure

`assay.go` is the entire implementation: prompts, API client, result types, JSON shims, rendering,
progress, chain writing, usage accounting, cost estimation — 1640 lines in one file.

The immediate consequence is that every symbol is in the same package and accessible everywhere
with no cost. There is no mechanism that prevents `printSummary` from calling `mdCell`, or
`runAudit` from reading `chainFile` directly, or `callClaude` from reaching into `usage`. In a
single-package single-file codebase, the compiler enforces nothing about layering; only
discipline does, and discipline erodes.

The specific failure this produced: the test seam (`cfg.call`, `cfg.httpClient`) was retrofitted
onto the god object rather than designed as a boundary. The seam is functional but it works
against the type system rather than with it. A genuine boundary — say, a `Caller` interface passed
to functions that need it — would have forced the question "who actually needs API access" at
the time the function was written.

The fix in the new repo: at minimum, separate `api/` (the HTTP client and retry logic),
`pipeline/` (the orchestration: decompose, assayClaim, faithClaim, evidenceClaim), and `render/`
(terminal + markdown output). Package boundaries are free compiler-enforced documentation.

---

## 3. Gates measured behavior but never structure

The CI gate (`go test ./...` via `build.sh`) tests four things:
- Pure functions: `splitSummary`, `extractJSON`, `unmarshalLoose`, `mdCell`, `tally`
  (`assay_test.go:1–~100`)
- Prompt presence: two regex assertions that specified strings appear in `substanceCriticSys` and
  `faithCriticSys` (`TestFaithCriticSysLiteralization`, `TestSubstanceCriticSysConditionDiscipline`)
- Stub-driven behavior: three scenarios where `cfg.call` is replaced with a canned response and the
  real orchestration runs (`TestConditionLaunderingDowngrade`, `TestConditionLaunderingLoopStop`,
  `TestRunEvidencePropositionSubstitution`)

What the gate never measures:

- That `cfg` fields are accessed only from the phase where they are valid (temporal coupling is
  invisible to `go test`)
- That the verdict-enum contract is enforced (W10: any string passes `unmarshalLoose` into
  `Verdict` with no check; `assay.go:937`)
- That `maxTokens` truncation routes to `error` rather than a partial verdict (W11)
- That the loop terminates — `assayClaim`'s loop condition (`assay.go:443–448`) has four guards
  but no test stubs a critic that always returns `needs_another_round=true` to confirm the flag
  is honored
- That audit cross-tab alignment is correct (no test for column-to-claim correspondence)

The result: every behavioral regression that does not appear in one of the three stub scenarios is
invisible to CI. The Layer-2 planned tests (TESTING.md §"Layer 2") are the repair, but they were
never written because each feature session added behavior first and tests never caught up.

The fix in the new repo: write the Layer-2 tests at the time the feature is written, not after.
The seam (`cfg.call`) is already there; using it for the planned tests (max-rounds, verdict-enum,
truncation, audit alignment) costs nothing structurally. The mutation test pass (TESTING.md §"Teeth
check") is the structural check: a test that survives its own logic being inverted is testing
nothing.

---

## 4. Feature-only sessions with no consolidation

The decision log (d008 through d017) shows a clean pattern: each session ships a feature or
calibration, and no session spends time on consolidation. The debt accumulated:

- `cfg` grew from ~10 fields to 15 without a pause to ask whether the struct should be split
- W10 (no verdict-enum validation) was identified as "expected red today, cheapest real fix"
  (SESSION.md) in the same pass that added `survives_only_by_conditioning` — and was deferred
- The `best_case` field in `faithDefenderSys` (assay.go:614) appears in the prompt but not in
  `defenderJSON` (assay.go:1026–1029); it generates silently and is discarded — this kind of
  drift goes unnoticed in a single-file codebase with no interface boundary
- `README.md:164–166` became stale when `crossCheckEvidence` was added (d009/d010) and the
  staleness was flagged in three separate documents (BACKGROUND.md, SESSION.md, d012) without
  being fixed because it was always "a separate maintenance edit, not bundled here"

The pattern is: features are shipped, debt is named, debt is deferred, debt compounds. The
deferred items accumulate in SESSION.md and TODO.md rather than being addressed in the session
that created them, because the session boundary makes each item "someone else's debt." In a
single-person project with no review, the deferral policy needs to be stricter: a feature that
introduces a known gap (W10, the stale README, the dead `best_case` field) should not ship until
the gap is either fixed or explicitly accepted with a test that will fail until it is.

The fix in the new repo: the consolidation pass is part of the session, not deferred to the next
one. Concretely: before a session closes, check (1) did any W-entry become fixable? (2) did any
struct grow a field that doesn't belong to its original concern? (3) is any comment or doc now
stale? The cost of a 15-minute consolidation at the end of a session is far below the cost of
backfilling it across multiple future sessions.
