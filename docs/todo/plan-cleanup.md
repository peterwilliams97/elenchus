# plan-cleanup.md — organization opportunities (top-level sprawl)

Assessed 2026-09-12. Reference counts are `git grep -l` over tracked files; treat as drift-prone.
Nothing here has been actioned — a to-decide list, not a change record. (The `assay`-vs-`elenchus`
rename is a related decision, kept in `repo-map.md` §Naming.)

## Quick wins — low risk
- **`tree.txt` — delete from tracking.** A tracked 480-line filesystem dump (last touched
  `4dfd751`) that drifts the instant any file moves and is not documentation.
- **Untracked scratch is already gitignored** (`demo-dan.log`, `diff.txt`, `eval-value.stderr`,
  `eval-value.tree`, `eval/`) — no repo change needed; `rm` locally for a clean listing.
- **`FIXTURE_REPORT.md`** — dated 2026-06-01, 1 reference. Stale one-off; fold into `testing/` or
  `docs/`, or delete.

## Medium — real churn, do selectively
- **Root markdown splits by reference weight.** Keep at root (living docs + convention):
  `README.md`, `CLAUDE.md`, `BACKGROUND.md` (11 refs), `TESTING.md` (13), `SESSION.md` (11) —
  moving each edits 11–13 files of cross-links. Relocate candidates into `docs/` (cheap):
  `REPORT.md` (4), `FIXTURE_REPORT.md` (1). (`TODO.md` already moved to `docs/todo/roadmap.md`.)
- **Three test-ish dirs read as one to a newcomer** — `eval/` (gitignored scratch run outputs),
  `testdata/` (Go-convention fixtures), `testing/` (tracked calibration logs). Distinct, but the
  names don't say so; a one-line header in each closes the gap cheaply.

## Structural — biggest opportunity, a decision not a tidy-up
- **The Go layout departs from the `~/CLAUDE.md` `cmd/` + `internal/` standard** (no `.go` at the
  repo root). Here the root carries `assay.go` (~3,900 lines) plus four test files
  (`assay_test.go`, `judge_test.go`, `oracle_gen_test.go`, `refuter_b_test.go`). Clean target:
  `cmd/assay/main.go` for flag parsing / `serve` dispatch, the judge path extracted into an
  `internal/` package, `assay.go` split by concern like the rest of `internal/`.
- **It contradicts the repo's own written convention.** `elenchus/CLAUDE.md` deliberately documents
  `assay.go` at root as the CLI, and `build.sh:8` hard-codes `go build -o assay .` — so a move
  breaks `build.sh`, the README invocations, and spec references in the same commit. This is a
  decision to make on purpose (Peter's call), not a cleanup-pass side effect.
