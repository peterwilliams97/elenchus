# Fixture Collection Report

## Current state — 2026-06-01

Three real public-domain documents are present in `fixtures/raw/` (gitignored):

| File | Source | Size | Provenance snippet |
|------|--------|------|--------------------|
| `legal.txt` | Project Gutenberg — *The Autobiography of Benjamin Franklin* (PG #148) | 395 KB | "The Project Gutenberg eBook of The Autobiography of Benjamin Franklin" |
| `engineering.html` | SQLite architecture page — sqlite.org/arch.html | 22 KB | `<!DOCTYPE html>` — sqlite.org served directly |
| `pm.md` | 18F Technology Budgeting Handbook — github.com/18F/technology-budgeting | 2.2 KB | "# Technology Budgeting Handbook" |

`TestFixtureIngestion` in `assay_test.go` reads from `fixtures/raw/` and passes each file through
`splitSummary`. Missing files are `t.Skip`'d — not substituted.

## Still missing

| Slot | Source attempted | Why failed |
|------|-----------------|------------|
| `accounting.txt` | SEC EDGAR 10-K MD&A (Apple, CIK 320193) | SEC returned HTML error page — not raw text |
| `contract.xml` | GSA FAR-Data repo, FAR-XML/52.xml | GitHub 404 — path does not exist |

These slots remain empty. Per CLAUDE.md no-fabrication rule, they are skipped in the test rather
than substituted.

## History

The original 2026-05-31 session halted with zero fixtures (all seven curl attempts failed). The
three files above were fetched in a follow-up session on 2026-06-01. The original failure record
is preserved in git history at `a1684e0`.
