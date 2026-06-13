# CLAUDE.md

## spec/ is frozen input

`spec/` was copied once, byte-for-byte, from v1 (`elenchus/spec/`) on 2026-06-13.
It is never edited, extended, or deleted.

Any session that finds `spec/` missing or modified **STOPS and reports** — it never
substitutes the v1 repo or memory as a source.

## Structural rules

- No new fields on shared structs. New state is passed as parameters. Any struct wanting
  a 7th field requires a split proposal, approved before coding.
- Options are immutable after `main()`. Run-state lives in the function that runs.
- The API test fake lives in `internal/client` only. No other package stubs the network.
- Every session is feature OR consolidation, never both. Debt found mid-feature is logged
  for its own consolidation session, not side-fixed.
- Reviews are of code (the diff). A review of a report or doc must say so.
- Prompts in `internal/modes` are verbatim from `spec/PROMPTS.md`; any deliberate change
  to a prompt is a design-queue item (PLAN.md §3), never a port-time edit.

## Carried disciplines

- **fetch-or-STOP** — required external inputs are fetched; if unobtainable, STOP and mark blocked, never substitute.
- **provenance propagation** — synthetic input taints downstream; conclusions drawn from it are marked invalid.
- **red-then-green** — the failing test exists before the code that passes it.
- **pre-register decisions** — decisions are recorded before coding, not rationalised after.
