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
- **Call things by their name; don't invent names and don't write "the tool".** Things
  that already have a name (`crossexam`, the packages, the modes, the files) are
  referred to by that name, not by a coined label and not by a vague placeholder. If
  something genuinely has no name yet, describe what it does in concrete terms rather
  than minting a term and using it as if it were established.

## Carried disciplines

- **fetch-or-STOP** — required external inputs are fetched; if unobtainable, STOP and mark blocked, never substitute.
- **provenance propagation** — synthetic input taints downstream; conclusions drawn from it are marked invalid.
- **red-then-green** — the failing test exists before the code that passes it.
- **pre-register decisions** — decisions are recorded before coding, not rationalised after.

## Replace Degraded Claude Code

A running, itemised list of concrete failures in this repo's sessions — so degradation is recorded,
not waved away. Read it before working; do not repeat what is here. Newest first.

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
