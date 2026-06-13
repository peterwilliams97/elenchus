# CLAUDE.md

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
  a 7th field requires a split proposal, approved before coding.
- Options are immutable after `main()`. Run-state lives in the function that runs.
- The API test fake lives in `internal/client` only. No other package stubs the network.
- Every session is feature OR consolidation, never both. Debt found mid-feature is logged
  for its own consolidation session, not side-fixed.
- Reviews are of code (the diff). A review of a report or doc must say so.
- Prompts in `internal/modes` are verbatim from `spec/PROMPTS.md`; any deliberate change
  to a prompt is a design-queue item (PLAN.md §3), never a port-time edit.

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

## Carried disciplines

- **fetch-or-STOP** — required external inputs are fetched; if unobtainable, STOP and mark blocked, never substitute.
- **provenance propagation** — synthetic input taints downstream; conclusions drawn from it are marked invalid.
- **red-then-green** — the failing test exists before the code that passes it.
- **pre-register decisions** — decisions are recorded before coding, not rationalised after.
