# reference-material.md — future-work source material (external)

`../elenchus_material/` (sibling of the repo, **machine-local path — not committed here, not on any
remote**) holds the source material the next build phase draws from. The plan is to build as much as
possible *from the PDFs*; Peter is doing the prep of deciding which parts of each to analyse, so
**the PDFs are unread by design** until that selection lands. Read-only from this repo — never write
there.

- **`suriving_ai_coding,md`** (authored, v8 2026-09-06) — READ. "Surviving human-to-AI coding:
  master plan," the conceptual parent of assay's stance. Its vocabulary is already half-encoded in
  `CLAUDE.md`: **checker/statement/refuter** (= the never-weaken-tests + seen-to-fail rules,
  `refuter_b_test.go`), **"a test must say when it didn't look"** (= the zero-output rule and
  `absent` vs `unverifiable`, `internal/manifest`), **producer/partition/differential testing**
  (= the `-source` corpus and the segmentation seam). Its central claim — "a check is only as good as
  the inputs you feed it," ordered easy→hard→hardest (documents / WPP / security) — is the same
  wall as roadmap item 5: a consistency tester is buildable now, a correctness tester is blocked
  on a real input population that must be obtained, never synthesised.
- **`Anthropic-Detecting-and-countering-091026.pdf`** — UNREAD (pending Peter's selection).
- **`A Model for Organizational Interaction.pdf`** — UNREAD (pending Peter's selection).
- **`ai_index_report_2026.pdf`** — UNREAD (pending Peter's selection).
