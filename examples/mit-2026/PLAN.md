# PLAN — MIT Ad Hoc Committee on AI Use (Aug 2026)

Corpus catalogue entry: `docs/todo/Consultant & institutional reports.md` § G. Layout copied from
`examples/dora-2026/` (single-source internal-consistency config). No fetch / code / commit yet —
PW reads this first.

Report: *Report of MIT's Ad Hoc Committee on AI Use in Teaching, Learning, and Research Training*,
MIT, 13 Aug 2026, co-chairs Klopfer & Madden, 38 pp + appendices. Catalogue calls it the
highest-value artifact in the corpus for root-claim-vs-evidence testing.

## 1. Sources

- **Report PDF (HELD, target AND source):** `https://aiandeducation.mit.edu/report/` →
  `AI-Committee-Final-Report-Aug-13.pdf` (mirror `https://aihub.mit.edu`). URL status **[V]** —
  confirm the exact PDF path before download. → `sources/report.pdf` + extracted `report.txt`.
- The report is **one file**; its appendices are companions bundled inside it, not separate PDFs.
- **Cited but NOT held** (external truth-makers → `unverifiable`, named in MANIFEST "Not held"):
  Brookings, PsyArXiv, Pew, WSJ (the Sadun piece, §3.1.3), Healthy Minds Network, and MIT internal
  URLs (§3.3.7 Parley pricing). Also **not held:** the raw microdata behind the three surveys
  (AI Usage & Attitudes n=1,632; Quality of Life n≈8,200; The Tech Fall-2025 n=1,002) and The Tech's
  own published writeup — their figures are checkable only for internal consistency, never grounded.
- **Browser / by-hand (PW fetches):** WSJ (paywall/JS); any MIT internal URL behind SSO; Pew/
  Brookings/PsyArXiv are normally open — attempt in the batch, hand off only on failure.
- **gitignore:** add `examples/mit-2026/sources/*` + `!examples/mit-2026/sources/MANIFEST.md`
  (mirrors `.gitignore:73-74`). Fetched docs stay untracked; MANIFEST is the one tracked source file.

## 2. Roles — appendices are the source for the body (internal-contradiction config)

- **Claims-under-test:** §3 recommendation clusters (3 clusters, ~25 numbered sub-recs), the §1.1
  Landscape premises they rest on, and the §2 principles. Root thesis ≈ "MIT should lead boldly on
  AI in education and research training."
- **Source corpus:** Appendix C.1/C.2/C.3 survey tables (the evidence base), Appendix A (AI-use
  provenance statement), Appendix B (four-option syllabus menu).
- **Explicit:** the report's appendices ARE the source corpus for its own body. report.pdf is both
  assay TARGET and only held SOURCE, so every automated check is INTERNAL CONSISTENCY (does the body
  agree with its own appendix figures) — identical to dora-2026, not external grounding.

## 3. Decomposition

- **claims-machine.txt:** one atomic claim per line, 5 pipe fields as in dora
  (`id | route= | ref=§<heading> p<page> | claim | src=`). Routes: `survey` (a figure from one of the
  three surveys; truth-maker microdata not held → grounding `unverifiable`), `evidence` (external
  cited figure), `evaluative` (`[opinion]`, principles/posture), `method` (survey-design claims).
  Page-reference every leaf. `claims-machine-full.txt` carries the unpruned set.
- **argument.txt:** root → REC → F → atomic-claim leaves. Root = the "be bold / MIT should lead"
  thesis; 3 rec clusters as intermediate nodes → ~25 sub-recs (R-*) → one finding (F-*) each → leaves.
  Cross-cutting recs with no single-finding basis get `?` and attach under a `base` node (dora
  R-TRANSFORM pattern).
- **scheme= per F→REC edge** (on the finding line, dora-style): PW's prior is `survey` for most.
  Caveat to resolve at build: per `spec/EDGE.md` §1 the tag types the *inference*, not the finding —
  dora's survey-shaped findings feed means→ends recs and so are tagged `practical`. MIT's recs are
  mostly remedial means→ends ("mandate policies to fix confusion") → likely **`practical`**; a rec
  that just generalizes a sample statistic to all students → **`survey`**; anecdote-driven recs
  (e.g. §3.3.7 tool pricing) → possibly **`example`**. `-edge` is a registered practical run
  with pre-registered predictions (§6.3), not record-only.

## 4. Pre-registered expectation — §3.1.8 vs C.2/C.3

- **Leaf:** the atomic claim under R-3.1.8 stating "students report the AI guidance from instructors
  is often confusing and unclear" (echoing §1.1 Landscape). `route=survey`, ref §3.1.8.
- **Quotes that should surface** (from the held source appendices): C.2 — 72% of undergrads
  agreed MIT gave adequate AI guidance (highest of any group); C.3 — 75% believed faculty
  expectations on AI were clear.
- **Expected verdict: `contradicted`** — two appendix surveys inside the same document contradict
  the recommendation's stated premise; neither is acknowledged in §3.1.8.
- **Counts as a MISS:** verdict other than `contradicted` (`faithful`/`partial`) with C.2/C.3 in the
  retrieved set — a judge miss. Distinguish from a **retrieval/extraction miss:** if BM25 fails to
  surface C.2/C.3, or the appendix tables don't survive PDF→txt (they sit in Codex-generated graphs,
  possibly images — see §5), the leaf grounds `unverifiable` and the check never fires. That
  extraction risk is the main threat to this pre-registration; verify C.2/C.3 figures are present as
  text in `report.txt` before trusting a non-`contradicted` result.

## 5. G.3.1 numeric error in the Codex appendix

- The defect: C.2 text reads "46% of undergrads and 60% of undergrads and graduate students" — the
  label "undergrads" appears twice; the second should read grad students/postdocs (chart: undergrads
  31+15=46, grads 27+33=60; total 22+19=41 ≈ "about 40%"). It sits in the appendix whose statistical
  summaries were Codex-generated (Appendix A disclosure) — declared provenance boundary + a defect
  inside it.
- **How to represent it:** encode the mislabelled sentence as one `route=survey` leaf (ref §App C.2)
  and, if the chart numbers extract, a companion leaf carrying the chart-derived figures, as an
  internal-consistency pair (does the text label match its own chart).
- **Honest limit:** the current faithfulness axis compares claim-text to source-text; here the source
  IS the flawed sentence, and the corrective numbers live in a chart that is likely an image → not
  extractable. So the label/arithmetic error is **not reliably machine-detectable** — it is a
  human-adjudication item unless PW hand-transcribes the chart into a passage. Record it as a
  pre-registered known-limit case (ties G.3.1 to the G.4 Codex boundary), not an expected auto-catch.

## 6. Run sequence

1. **Faithfulness N=3, Sonnet** (`-backend anthropic -model claude-sonnet-4-6 -retrieve bm25 -n 3`)
   against `sources/MANIFEST.md`, writing the Tier-2 chain + tree/audit under `evidence/<date>-full/`.
   The only model-calling pass; the verdict to trust.
2. **Argument render** — `-from <chain> -argument argument.txt -manifest -refs claims-machine.txt
   -review`, no model calls; builds the two-pane PDF-linked site.
3. **`-edge`** — a **registered practical run** (not record-only): expect `survey`/`practical`
   edges, Layer-3 calibration, never a gate. Register the predictions before the run and score
   against them after.
   - **Predictions:** the twelve `scheme=practical` edges are tabled in **§6.3** with a blank
     prediction column — PW fills each `side_effects` call (`open`/`unchallenged`) before the run.
4. **`-evidence`** on the external citations — expect `unverifiable` for every not-held source;
   earns its keep only on anything PW fetches (Pew/Brookings/PsyArXiv).
- **Cost:** **~$12 at N=3**, reading the standing ~$10/100-claims figure as **per-pass** (it already
  assumes a multi-sample pass, not per-call), ~120 claims — dominated by pass 1; render/backfill are
  model-free. Confirm the claim count after decomposition before the spend. (PW runs the API pass.)

### 6.3 Practical edges — pre-registered predictions

The twelve `scheme=practical` F→R edges in `argument.txt`. `side_effects` (spec/EDGE.md §2) is the only
CQ: the pass opens the edge iff the finding or its verified quote names a **cost of the recommended
means**, and returns `unchallenged` otherwise. "Cost anchor" below is that clause read verbatim from
the finding, or `none`. "Grounds?" is whether the finding has a real survey/evidence leaf (not just
`[opinion]`) for the edge call to quote — the edges marked no cannot fire the automated run and are
tabled for structure only. **Prediction** is blank for PW to fill (`open` / `unchallenged`) before the
run; score after.

| Edge (R ← F) | Finding gist | Cost anchor (side_effects) | Grounds? | Prediction |
|---|---|---|---|---|
| R-3.1.2 ← F-3.1.2 | AI saps p-sets/exams of assessment value | none (cost is of the status quo, not oral exams/portfolios) | no | |
| R-3.1.3 ← F-3.1.3 | AI-era workers succeed by experimentation (Sadun) | none | yes (R313c) | |
| R-3.1.5 ← F-3.1.5 | UROP engages 93%/58%; educates, not labor | none | yes (R315b/c) | |
| R-3.1.6 ← F-3.1.6 | grade rationing counterproductive, raises AI reliance | none (cost is of the rejected means, rationing) | yes (R316c) | |
| **R-3.1.8 ← F-3.1.8** | **students report AI guidance confusing/unclear** | **none — POSITIVE CONTROL** | **yes (R318a, survey)** | |
| R-3.1.9 ← F-3.1.9 | detectors risk arms race; misclassify NNES/neurodivergent | none (cost is of the rejected means, detectors) | yes (R319d) | |
| R-3.2.2 ← F-3.2.2 | students report anxiety/depression/loneliness | none (problem statement, not a cost of the means) | yes (R32c) | |
| R-3.2.3 ← F-3.2.3 | students uncomfortable with AI TAs; double-standard | none | yes (R323b, survey) | |
| R-3.3.3 ← F-3.3.3 | support units (TLL) already at capacity | none | no | |
| R-3.3.8 ← F-3.3.8 | open models rival commercial **but** largest need very large GPU clusters | "the largest open models require very large GPU clusters" | no | |
| R-3.3.9 ← F-3.3.9 | IS&T can see every Parley request and response | "IS&T can in theory see every user request and every backend response" | yes (R339c/d) | |
| R-3.3.10 ← F-3.3.10 | training and inference incur energy costs | "both training and inference incur energy costs" (a cost of AI use, the rec's subject, not of publishing cost info) | no | |

**Positive control: `R-3.1.8 ← F-3.1.8`** — the one practical edge whose finding names no cost and no
caveat of the recommended means (post a clear per-subject policy), and whose leaf is survey-grounded
(R318a) so the edge call has a real quote to reason from. Predicted `unchallenged`: if the pass opens
it, the model is inventing a defeater the report does not license. (Distinct from R318a's own
pre-registered `contradicted` faithfulness verdict vs C2h/C3d, §4 — that is the leaf, this is the edge.)

## 7. Adjudication

`adjudications.txt` starts empty (header only). Model-drafted leaf verdicts carry `cc`; PW promotes a
draft by editing the line and swapping `cc` for PW after reading the leaf on the page. cc never files
under PW.
