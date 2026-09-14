# Corpus manifest — documents held under sources/

Every source document the corpus holds, in the plan's own naming. The backticked id is canonical: a
claim's `cites` column names the documents a claim rests on with these ids (LoadHeld reads them), and
a faithfulness run marks a claim `unverifiable` (no model call) when a cited id is not listed here.
Each held doc's id equals its filename, so a run's `-source` can load it.

The TARGET of the assay is the plan itself, `report.md`. Five documents are held: the plan, the IPP
refuter evidence doc it links, and the three companion docs it references (the FLT/Prove2Me writeup,
the "Making Software Correct" doc behind the PDF-library method, and the security /
vulnerability-finder companion). The plan is a STRATEGY / DESIGN document, not a survey — most of its
content is normative prescription (`route=method`). The checks this corpus supports:

1. **Internal consistency** — does `report.md` state a claim at its `ref` §, and do restated figures
   agree (OV8/DP2 on the PDF defect record; WP8/ME6/OV12 on 0.5 declarations per review-hour;
   WP4/ST9/LP2 on the 10,000-check run). `report.md` is markdown: no pages, so no printed→PDF offset
   and no report deep-links (`spec/SERVE.md` does not apply); the check keys on the plan's § headings.
2. **Grounding against a held companion** — a claim whose truth-maker is one of the other four held
   docs can be grounded, not merely checked for consistency. Coverage below.

A claim whose truth-maker is NOT held (internal codec/WPP/redaction records, the PDF production defect
record, un-held primary papers) grounds to `unverifiable` by design, not omission — see "Not held".

## Held (5)

- `report.md` — *Surviving human-to-AI coding: master plan*, draft v8, 2026-09-06. The TARGET, and
  the source for every internal-consistency check.
- `refuter-runs-2026-08-15.md` — *Refuter runs for the branch's key fixes — 2026-08-15*, from the
  PaperCut `ipp` repo (`docs/reviews/`, branch `pd2639/stack`). Grounds **R-REF** ("two of six tests
  could not fail" — its summary table marks `217e650` alone and `4b7f57a` undetectable on revert, the
  2-of-6); also cited by **ST6, AP4**.
- `flt-kloc-review.md` — *Applying Fermat's Last Theorem formalisation to KLOC-scale PR review*
  (draft, 2026-09-05), the companion the plan names in its header. Grounds **WT4** (13M lines of
  Lean; kernel checks every step) and **ST4** (Prove2Me proof-sketch = NL description + Lean
  statement; blinded read-back). Rule 8b: this is PaperCut's writeup — its own primaries (Anthropic's
  research; arXiv:2608.28433) are marked full-fetch there but are NOT held here, so grounding WT4/ST4
  is corroboration against the companion, not the primary.
- `software-correctness.md` — *Making Software Correct*, the talk/doc behind the UniDoc PDF-library
  method. Grounds the document-products case and the sampling principles it originates: **OV8/DP2**
  (the "no serious bugs in the first 6 months" success criterion — stated here; the several-years
  outcome is NOT held, see below), **DP1** (producer = printer driver), **DP3** (author found/fixed
  bugs after shipping, "ran the automation iteratively over a week"), **DP5** ((a) bound by producers
  / (c) corpus by origin / (d) rasterize+count, grayscale compare / (e) run repeatedly), **SA4**
  ("a handful of patterns"; few generators share behaviours), **SA5** (producer clustering), **LP4**
  (run the checks while developing), **NS4** (the handful-of-patterns judgement, still uncounted),
  and corroborates **OV1/OV2** (humans can't reason about large code).
- `vuln-finder-hive.md` — *Generalising the vulnerability finder across PaperCut, starting with Hive*
  (v2, 6 Sept 2026), the security companion, itself a tracking doc against this plan. Grounds the
  security case: **SE1** (attacker economics; adaptive producer), **SE2** (line B blinded simulation +
  line A attack-surface enumeration), **SE3** (the blind; positives are the measured gap, negatives
  weak), **SE4/SE5** (a prior AI review is a baseline only if model versions were pinned), **SE6**
  (line A is the only line yielding a coverage denominator, and goes first), **SA9/SA10** (the
  attacker case; "the instance travels, the reason for it doesn't"). NOTE: this file's last line is a
  184 KB embedded base64 PNG (the template diagram) — strip it before feeding the doc to a run, or
  retrieval will choke on one 184 KB "passage".

## Not held — named so their absence is `unverifiable`, never silently `absent`

These are the truth-makers the plan's remaining results and citations rest on. None is in this
corpus; a grounding run against them cannot be performed, and any conclusion depending on one is
inconclusive, not clean.

### PaperCut internal data (the plan's own results)

- **PDF library production defect record** — the customer-report channel behind "no customer-reported
  PDF defect over several years" (the OUTCOME half of OV8/DP2; the 6-month criterion half IS held in
  `software-correctness.md`). Per the plan's §What this does not settle (NS3), the before-method
  defect count that would make this evidence rather than anecdote does not yet exist.
- **IPP codec review doc, `2026-09-05-rapid-ipp-codec.md` @ a46a4b7** (branch
  `human-to-ai/rapid-ipp-codec`) — truth-maker for the entire WPP first result: the 10,000-check run,
  the year overflow 57760 → −7776 at the commit before `be9e85c`, the four bugs (uint8(200) → −56,
  the 32-bit truncation, the two validator bugs), the shared-`toInt32` blind spot, the timestamp fix
  shape, the eleven commits / 5.5 agent-hours / ~2 review-hours, and the 0.5-declarations-per-
  review-hour row (WP3–WP8, ST9, LP2, LP3, ME6, OV12). Linked by the plan; not held here.
- **Redaction tooling records** — the `Incomplete()` gate and PASS/FAIL/ERROR/REFUSED outcomes (AP1,
  ST8); the two-reader unanimity rule (TC1); `/Info` metadata in 11 of 24 files and the XPS
  clean-bytes/dirty-render case (AP2, TC3); the append-only Determinism Ledger; 24 of 24 corpus files
  reading "Aspose" and 89% of files in 49 clusters (AP3); the `/ActualText` defeat. Several are
  self-graded `[cite path]` in the plan — unverified even there.
- **WPP lab records** — the 65/65 transport finding (WP-65, TC5, WP2), `ipptool` runs, the lab
  printers. Not held.

### Papers and external artifacts cited

- **Kuhn, Wallace and Gallo**, "Software fault interactions and implications for software testing",
  *IEEE TSE* 30(6), 2004 — the basis for "most failures are triggered by two or three conditions"
  (SA3). Not held.
- **Primary FLT sources** — Anthropic, "Formalizing Fermat's Last Theorem" (4 Sep 2026); Chen et al.,
  "Prove2Me", arXiv:2608.28433v2. The plan's WT4/ST4 are held only via the companion
  `flt-kloc-review.md`; these primaries are not held (Rule 8b).
- **Independent PDF implementations** — poppler, mupdf, qpdf, pdftk, pikepdf, named as
  differential-testing oracles (DP1). Named, not held as code or eval.
- **WPP / Windows Protected Print Mode** — the closed Microsoft print stack the WPP case must match
  (OV9, WP1). Closed source by definition; not held.

### Companion documents referenced but not supplied

- **The "document-based products" Google Doc** the plan links by URL. `software-correctness.md`
  covers the same PDF-library method and is held; whether the two are byte-identical is not asserted,
  so method claims ground against the held doc and any URL-specific content is treated as not held.

## Notes for a run

- Every held doc's id = its filename; `report.md` is markdown, so there is no `report_page_offset` and
  no `sections:` map (both are PDF deep-link machinery, `spec/SERVE.md`); `ref` keys on § headings.
- The `single_source: false` config line below declares five held docs. A leaf citing only
  `report.md` and found `absent` there is the plan stating something once with no second held document
  to corroborate it: `uncorroborated`, derived as weakened, not failed.

single_source: false
