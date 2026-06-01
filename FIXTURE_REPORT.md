# Fixture Collection Report

## Current state — 2026-06-01

**Status: 7/7 obtained.** All slots populated with real, verified documents.

| function    | file               | source                                                                                          | verify-snippet (present in file) |
|-------------|--------------------|-------------------------------------------------------------------------------------------------|----------------------------------|
| legal       | `legal.md`         | github/dmca — RIAA DMCA notice re youtube-dl (2020-10-23)                                      | "perjury, we submit that the RIAA is authorized to act" |
| accounting  | `accounting.md`    | SEC EDGAR 8-K Ex-99.1 — Coinbase Q3 2025 shareholder letter                                    | "Total revenue in Q3 was $1.9 billion" |
| sales       | `sales.md`         | YC launches — Risely AI (O4v)                                                                   | "1 in 3 students drop out because of poor student service" |
| marketing   | `marketing.md`     | Fairphone Gen. 6 press release PDF (fairphone.com, 2025-07)                                    | "over 50% fair or recycled materials" (line-wrapped in PDF extraction) |
| pm          | `pm.md`            | ethereum/EIPs — EIP-1559 fee-market proposal                                                   | "There is a base fee per gas in protocol" |
| engineering | `engineering.md`   | GitLab blog — database outage postmortem 2017-01-31                                             | "around 300 GB of data had already been removed" |
| contract    | `contract.md`      | github/site-policy — GitHub Marketplace Developer Agreement (effective 2025-05-27)              | "govern your participation in GitHub" |

All files are in `fixtures/raw/` (gitignored — analysis use only, not redistribution).
`TestFixtureIngestion` in `assay_test.go` runs each through `splitSummary`; missing files are
`t.Skip`'d, never substituted.

## Superseded stopgap files (deleted)

The following genre-mismatched files from the 2026-06-01 Phase 1 session have been removed:

| deleted file        | was                                          |
|---------------------|----------------------------------------------|
| `legal.txt`         | Project Gutenberg — Franklin autobiography   |
| `engineering.html`  | SQLite architecture page (HTML)              |
| `pm.md`             | 18F Technology Budgeting Handbook            |

## History

- **2026-05-31** — original Phase 2 session halted with 0/7 fixtures (all curl attempts failed).
  Failure record at git SHA `a1684e0`.
- **2026-06-01** — Phase 1 stopgap: 3 genre-mismatched public-domain docs fetched into
  `testdata/fixtures/` then moved to `fixtures/raw/`.
- **2026-06-01** — Phase 2 complete: all 7 real-genre docs fetched, snippets verified, stopgaps
  deleted. This report reflects final state.
