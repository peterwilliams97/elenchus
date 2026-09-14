# Provenance — master-plan corpus

Records what is under analysis and what is authoritative, so a later reader does not rebuild from the
wrong source. No judge run has been made against this corpus yet: this is the CLAIM + ARGUMENT +
MANIFEST scaffold only (no `evidence/` chains, no `site/`).

## Document under analysis (TARGET)

- *Surviving human-to-AI coding: master plan*, draft v8, 2026-09-06.
- `sources/report.md` — copied verbatim from `~/code/personal/elenchus_material/ai_master_plan.md`
  (the read copy). Gitignored under `sources/*` (this example's `.gitignore`): analysis input, not
  redistributed. `sources/MANIFEST.md` is the tracked record of what is held.

## Held sources (5)

- `sources/report.md` — the plan (TARGET and the internal-consistency source).
- `sources/refuter-runs-2026-08-15.md` — from the PaperCut `ipp` repo
  (`docs/reviews/2026-08-15-refuter-runs.md`). Grounds R-REF ("two of six tests could not fail" — its
  summary table marks `217e650` alone and `4b7f57a` undetectable on revert); also ST6, AP4.
- `sources/flt-kloc-review.md` — *Applying FLT formalisation to KLOC-scale PR review* (the companion
  named in the plan's header). Grounds WT4, ST4 — corroboration against our writeup, not the primary
  (Anthropic research / arXiv:2608.28433 not held; Rule 8b).
- `sources/software-correctness.md` — *Making Software Correct* (the doc behind the UniDoc PDF
  method). Grounds the document-products case + sampling origins: OV8/DP2 (6-month criterion), DP1,
  DP3, DP5, SA4, SA5, LP4, NS4; corroborates OV1/OV2.
- `sources/vuln-finder-hive.md` — *Generalising the vulnerability finder … Hive first* (v2). Grounds
  the security case: SE1–SE6, SA9, SA10. Its last line is a 184 KB embedded base64 PNG — strip before
  a run.

The four companions were supplied by the user into `sources/`. All held docs are gitignored
(`.gitignore`: `sources/*`, keep MANIFEST) and never committed — internal PaperCut material.

## Deliverables in this directory

- `claims-machine.txt` — annotated pipe-format claims (routes, `ref` §, per-claim src/truth-maker).
- `claims-machine-full.txt` — the tab-delimited machine form the tool parses (`-tree`); 124 leaves,
  generated from `claims-machine.txt` (fields: id, §-path, claim, cites, route).
- `argument.txt` — the plan as an argument tree: root = the plan's thesis, 12 recommendations
  (REC1–REC12) as children, findings (F1–F15) under them, 124 atomic-claim leaves. Exact partition:
  every claim id is a leaf exactly once; descriptive/open material hangs under `base` on a `?` edge.
- `sources/MANIFEST.md` — held (2) vs not-held (PaperCut internal data, cited papers, linked Google
  Docs), each not-held truth-maker named so a grounding run reports `unverifiable`, never `absent`.

## What a run would and would not settle

- Groundable against a held companion (24 leaves): R-REF/ST6/AP4 (refuter-runs); WT4/ST4 (FLT
  companion, corroboration only per Rule 8b); OV8/DP1/DP2/DP3/DP5/SA4/SA5/LP4/NS4 and OV1/OV2
  (software-correctness); SE1–SE6/SA9/SA10 (vuln-finder-hive).
- Internal-consistency only: the remaining `route=method` / `route=definition` / `route=data-gap`
  claims, and the restated-figure pairs (OV8/DP2, WP8/ME6/OV12, WP4/ST9/LP2).
- `unverifiable` by design: the codec/WPP/redaction `route=result` claims (internal docs not held)
  and the several-years PDF outcome; `route=citation` for Kuhn/Wallace/Gallo 2004, the FLT primaries,
  poppler/mupdf/etc., and WPP; and every `route=prediction` ([improves]/[decays] — truth-maker future).
