# Corpus manifest — documents held under sources/

Every source document the corpus holds, in the report's own naming. The backticked id is
canonical: a claim's `src`/`cites` column names the documents a claim rests on with these ids, and
a faithfulness run marks a claim `unverifiable` (no model call) when a cited id is not listed here.

## Roles for Slice 1 (faithfulness config)

This slice holds exactly one SOURCE document — **Part B** of the report — and assays a TARGET that
is not held as a source — **Part A** of the same report.

- **SOURCE (held):** `report-partB.txt`, the report's Detailed Recommendations (printed pp77–204).
  It is the truth-maker every Slice-1 leaf rests on; each claim's `src` column names it.
- **TARGET (assayed, not a held source):** Part A, the report's front-half objective statements
  (printed pp16–76). The 11 objective claims in `../claims-objectives.txt` are drawn from here. Part
  A is deliberately NOT in `sources/` — if it were, retrieval could ground a Part A claim on its own
  Part A wording and the faithfulness check would be trivially satisfied.

The check is: does the Part B implementation SUPPORT the Part A objective claim. Claim and truth-maker
are two parts of ONE report (`single_source: true`), so the single-source judge rules apply — they
tell an overstatement from a benign restatement when the source passages are other parts of the same
document. It does NOT ground the report's external citations: the 795 inline link targets are the
truth-makers for `evidence`-route claims, none is fetched here, and Slice 1 has no `evidence` leaves.

## Source (1)

- `report-partB.txt` — Part B, *Detailed Recommendations* (printed pp77–204), extracted from
  `report.pdf` below. 128 pages, one `\f` per page preserved; extracted 2026-09-15 with
  `python3` splitting `report.txt` on `\f` and keeping printed pages 77–204 (PDF pages 77–204,
  offset 0). 7,446 lines, 488,236 bytes. First page footer "77 … A TRANSFORMATIVE AI STRATEGY FOR
  EUROPE"; last "204 …". The Part B "WHY IT MATTERS" block for each objective (pp79–169) is the
  passage a Slice-1 leaf is judged against.

## Origin PDF — provenance and deep-link artifact, not the Slice-1 grounding source

- `report.pdf` — *A Transformative AI Strategy for Europe*, KIRA Center (Berlin), **Version 1.0,
  September 2026**, 206 pages, CC BY 4.0. Convenor Monika Schnitzer (LMU Munich); Editorial Lead
  Daniel Privitera (KIRA Center). One file; Part A (pp16–76), Part B (pp77–204), two deep dives
  (compute pp189–196, robotics pp197–204) and a costed European frontier-AI-project hypothetical
  (pp174–187) are bundled inside it, not separate PDFs. No numbered reference list: all 795 citations
  exist only as inline link anchors. `report-partB.txt` is the pp77–204 slice of this file; Part A
  page refs (`pNN` in `claims-objectives.txt`) and any future `#page=N` deep-link resolve here.
  - **SHA-256:** `1640cb5d2ab4006b9dfcd9d48a630be192767eac84aaf6024ac2b78cda3df563`
  - **Version string (verbatim, p205):** "Version 1.0, September 2026. The strategy may be updated to
    correct errors; the current version is available at https://transformative-ai.eu." The legal
    notice makes this an errata-tracked document, so THIS snapshot is the artifact under test; a later
    version is a diff target, not a replacement.
  - **Fetched:** 2026-09-15 from `https://transformative-ai.eu/download/a_transformative_ai_strategy_for_europe.pdf`
    (landing `https://transformative-ai.eu`). PDF v1.5, 3,952,879 bytes, A4, InDesign 21.5;
    embedded CreationDate 2026-09-14. Extracted to `report.txt` with `pdftotext -layout` (11,283 lines;
    figure captions garble where text overlaps chart art — the Figure 1/2 caption region is the known
    soft spot).
  - **Provenance confirmed on extraction:** the PDF carries exactly **795 URI link annotations across
    222 distinct hosts**, matching the catalogue's H.1 count — the corpus entry's headline parse
    figure reproduces from the held file.

## Not held — named so their absence is `unverifiable`, never silently `absent`

The 795 inline citation targets are the external truth-makers for every `evidence`-route claim. None
is fetched in this pass, and Slice 1 carries no `evidence` leaves. A grounding run against them cannot
be performed here; any conclusion depending on one is inconclusive, not clean. The load-bearing ones
for the empirical section are tabled in `PLAN.md` §Slice 3 (top hosts: epoch.ai, anthropic.com,
metr.org, europe2031.ai). Also not held and flagged in the catalogue as post-dating general knowledge
(H.8), so document-internal assertions until checked: the July 2026 OpenAI-agents / Hugging Face
incident; the ~5× rise in critical-software vulnerabilities after Claude Mythos Preview; the July 2026
~200-economist statement; the 1,000+ frontier-lab-employee statement.

## Report-link resolution

`report_page_offset` is the printed→PDF page offset. It looks **~0** here (printed page footers align
with PDF page indices in the empirical section: printed p24 content sits on PDF page 24), UNLIKE
mit-2026's offset of 2 — VERIFY before trusting deep-links, since a wrong offset silently misplaces
every `#page=N` link. `single_source: true`: claim (Part A) and its truth-maker (Part B) are two
parts of one report, so a leaf `absent` is not a grounding miss against a second held document —
there is none — but a Part A claim that Part B does not repeat or support (O3.1 by design).

report_page_offset: 0   # VERIFIED 2026-09-15 (Slice 1). pdftotext -layout emits one \f per PDF page;
                        # chunk index i (0-based) carries footer number i+1 across the whole file, so
                        # printed page N sits on PDF page N. Known-page cross-check: printed p24 (chunk
                        # 23) contains "Nusajaya" (the Malaysia compute passage that PLAN §Slice 2 places
                        # on printed p24). Offset is 0, not mit-2026's 2.

single_source: true

# Part A objective pages for the `report:` deep-links (spec/SERVE.md). The §-label in each leaf's
# claims-machine.txt ref (§IO-1 …) resolves to the objective's Part A statement page here, and
# -review rewrites it into report.pdf#page=N. Pages are the verified Part A TOC pages (PLAN.md).
sections:
  IO-1: 28
  IO-2: 31
  IO-3: 36
  IO-4: 40
  IO-5: 45
  O1.1: 50
  O1.2: 53
  O2.1: 57
  O2.2: 60
  O3.1: 64
  O3.2: 68
