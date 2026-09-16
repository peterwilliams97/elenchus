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

## Roles for Slice 2 (faithfulness config — restructured)

Slice 2 (PLAN.md §Slice 2) is the headline positive-control claim chain: the compute comparison
stated three times in the report — executive summary (p6), IO-3 body lead-in (p24), footnote 2
(p24). It is now a two-leaf faithfulness check, mirroring Slice 1's SOURCE/TARGET split.

- **SOURCE (held):** `report-fn2.txt`, footnote 2 (printed p24), which states 2.1 GW for all of
  Europe and 0.662 GW for the Nusajaya (Malaysia) site, plus the p192 body sentence
  (`report.txt:10398-10400`) that scopes the 2 GW figure to "Europe as a whole – including non-EU
  countries such as the UK and Norway". It is the truth-maker both Slice-2 leaves rest on; the two
  leaves in `../claims-slice2.txt` carry `src=report-fn2.txt`.
- **TARGET (assayed, not a held source):** the two prose statements — the executive-summary "three
  times as much" (p6) and the IO-3 body "roughly one third" (p24). They are deliberately NOT in
  `sources/` (they live inside `report.txt`, which is retained only as extraction origin, below): if
  they were, retrieval could ground each on its own wording and the check would be trivially satisfied.

The check is: does each prose statement reconcile with footnote 2's figures? 2.1 / 0.662 = 3.17 ≈
"three times" (CMP-EXEC) and its reciprocal ≈ "one third" (CMP-BODY), so both are expected `faithful`
(pre-registered in `spec/TREE.md` § Refuters pre-registered for the reorder, refuter (a)). A
non-faithful on either leaf is a calibration miss, not a document finding. This mirrors Slice 1's role
split (SOURCE = Part B, TARGET = Part A held out): footnote 2 is the held source, the prose the target.

## Source (1)

- `report-partB.txt` — Part B, *Detailed Recommendations* (printed pp77–204), extracted from
  `report.pdf` below. 128 pages, one `\f` per page preserved; extracted 2026-09-15 with
  `python3` splitting `report.txt` on `\f` and keeping printed pages 77–204 (PDF pages 77–204,
  offset 0). 7,446 lines, 488,236 bytes. First page footer "77 … A TRANSFORMATIVE AI STRATEGY FOR
  EUROPE"; last "204 …". The Part B "WHY IT MATTERS" block for each objective (pp79–169) is the
  passage a Slice-1 leaf is judged against.

## Slice-2 source (1)

- `report-fn2.txt` — footnote 2 (printed p24), extracted verbatim from `report.txt:950-952`
  (line-wrapping joined), plus the p192 body sentence (`report.txt:10398-10400`, verbatim). The held
  SOURCE for Slice 2's restructured faithfulness check: the two prose compute-comparison statements
  (p6, p24) are judged for faithfulness against the 2.1 GW / 0.662 GW figures it states. The p192
  sentence supplies the "including the UK and Norway" scope that footnote 2 omits, without which
  CMP-BODY scored partial for a scope with no truth-maker in this source. It carries an orientation
  header above the passages; each passage is a verbatim paragraph, so a cited quote grounds as a substring.

## Slice-2 extraction origin — retained, not the Slice-2 source

- `report.txt` — the whole report (printed pp1–206), `pdftotext -layout` of `report.pdf` below.
  751,724 bytes; one `\f` per PDF page, offset 0 (printed page N sits on PDF page N). Retained as the
  extraction origin for the Slice-2 claim texts (CMP-EXEC `report.txt:382-383`, CMP-BODY
  `report.txt:942-943`) and footnote 2 (`report.txt:950-952`), NOT held as the Slice-2 source: the
  restructured check grounds against `report-fn2.txt` alone, and holding the whole report would let
  retrieval ground a prose statement on its own wording. Slice 1 does not use it either (Part A is
  deliberately not held — see the Slice-1 roles above).

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

## Slice-3 cited sources — fetch pass, held under `sources/cited/`

The empirical section "Why Europe needs to prepare for transformative AI" (printed pp20–26). PLAN.md
§Slice 3's ~20 load-bearing citations, each link target read off the PDF's own URI annotation
(pypdf, PDF pages 21–25 — never guessed), fetched 2026-09-16 with `curl` into
`sources/cited/<nn>-<host>.html` and extracted to `.txt` (stdlib HTMLParser strip). SHA-256 is over
the held `.html` (raw response bytes). **Held for provenance only** — no grounding run here; every
`evidence`-route Slice-3 leaf stays `unverifiable` until a later fetch/grounding pass. The `PLAN item`
column is the plan's numbering; two anchors share one target (14≡15 epoch data tool, 17≡18 fdc-hub).

`fig` column — is the *cited figure* present as extractable text in the held file:
`present` (the number/quote is in the text), `figure-only` (page renders the datum only in an
interactive chart / data tool, so the exact figure is not in the static extraction), `absent`
(page held, cited figure not located in the extraction — read most skeptically).

### Held cited ids (registered for `LoadHeld`)

The 16 captures fetched HTTP 200 and usable as text. Each is a backticked `- ` bullet so
`manifest.LoadHeld` sees it: a Slice-3 leaf citing one of these is judged (cite-scoped against the
capture's `cited/<stem>` passages, spec/TREE.md § Cite-scoped retrieval), and one citing an id NOT in
this list is `unverifiable` with no model call. Deliberately absent: `cited/08b-bloomberg.com.txt`
(HTTP 403 bot-block) and `cited/09-theinformation.com.txt` (login wall) — held on disk but a wall, so
not usably in the corpus; and the two uncited loci (`uncited/p23-jagged`, `uncited/p24-private-investment`),
which have no held document at all. All four render `unverifiable`, not silently `absent`.

- `cited/01-epoch.ai.txt` — Epoch Capabilities Index (Fig 1, p21) — figure-only
- `cited/02-metr.org.txt` — METR time-horizons (4-minute tasks, p21) — figure-only
- `cited/03-anthropic.com.txt` — Anthropic recursive-self-improvement essay (Fig 2, p22) — present
- `cited/04-epoch.ai.txt` — Epoch: AI-capabilities progress has sped up (p22) — present
- `cited/06-wemustactnow.ai.txt` — ~200-economist statement (p23) — present
- `cited/07-pacingthefrontier.com.txt` — 1,386 frontier-lab-employee statement (p22) — present
- `cited/08a-epoch.ai.txt` — Epoch: OpenAI revenue-growth projection (p23) — present
- `cited/10-epoch.ai.txt` — Epoch: hyperscaler capex trend (€650bn, p23) — figure-only
- `cited/11-internationalaisafetyreport.org.txt` — International AI Safety Report 2026 (p24) — present
- `cited/12-redwoodresearch.org.txt` — Redwood: Hugging Face incident (p24) — present
- `cited/13-forbes.com.txt` — Forbes Global 2000 (ten most valuable, p24) — present
- `cited/14-epoch.ai-aicompanies.txt` — Epoch companies revenue data tool (p24) — figure-only
- `cited/16a-europe2031.ai.txt` — europe2031 compute forecast (5% EU, p24) — present
- `cited/16b-epoch.ai-supercomputers.txt` — Epoch AI-supercomputers trend (p24) — figure-only
- `cited/17-epoch.ai-fdc-hub.txt` — Epoch frontier-data-centers hub (Malaysia, p24) — figure-only
- `cited/20-foxphilip.substack.com.txt` — frontier-model-access economics (100×, p25) — absent

| PLAN | anchor sentence (verbatim, `report.txt` page) | URL | file | fig |
|------|-----------------------------------------------|-----|------|-----|
| 1 | "Based on ECI scores, model capabilities have been improving nearly twice as fast since April 2024 than before." (Fig 1 caption, p21) | https://epoch.ai/eci | 01-epoch.ai.html | figure-only |
| 2 | "In 2024, they struggled with basic programming tasks that took programmers 4 minutes." (p21) | https://metr.org/time-horizons/ | 02-metr.org.html | figure-only |
| 3 | "they can now solve increasingly complex and open-ended tasks on behalf of AI researchers (see Figure 2)." + Fig 2 caption "Source: Anthropic." (p22) — INTERESTED PARTY (H.6) | https://www.anthropic.com/institute/recursive-self-improvement | 03-anthropic.com.html | present (essay; success-rate chart itself figure-only) |
| 4 | "by some key metrics, the pace at which AI models are improving has begun accelerating (Figure 1)." (p22) | https://epoch.ai/data-insights/ai-capabilities-progress-has-sped-up | 04-epoch.ai.html | present |
| 6 | "In July 2026, around 200 economists, including 16 Nobel Prize winners, signed a statement warning that 'radically more powerful AI' could arrive in the coming decade …" (p23) — UNVERIFIED (H.8) | https://www.wemustactnow.ai/ | 06-wemustactnow.ai.html | present (statement text; page lists signatories) |
| 7 | "Over a thousand frontier lab employees jointly stated that 'the world's leading AI companies believe they could be close to automating AI research'." (p22) — UNVERIFIED (H.8) | https://www.pacingthefrontier.com/ | 07-pacingthefrontier.com.html | present ("A statement from 1,386 employees …") |
| 8 | "Yet it only took OpenAI the first eight months of the year to roughly double its annualised run rate, from over \$20 billion to more than \$40 billion." — 2.3× projection clause (p23) | https://epoch.ai/gradient-updates/openai-is-projecting-unprecedented-revenue-growth | 08a-epoch.ai.html | present |
| 8 | same sentence — the \$40 billion run-rate clause (p23) | http://www.bloomberg.com/news/articles/2026-08-13/openai-s-revenue-run-rate-tops-40-billion-ahead-of-ipo | 08b-bloomberg.com.html | **hand-fetch** — HTTP 403 bot-block ("Are you a robot?") |
| 9 | "Similarly, Anthropic projected 4x growth in its most optimistic scenario for 2026, but had already grown by 5.2× by May 2026." (p23) | https://www.theinformation.com/briefings/anthropic-increases-2026-revenue-forecast-20-18-billion | 09-theinformation.com.html | **hand-fetch** — login wall ("Save 25% to unlock this story") |
| 10 | "AI capex in 2026 is projected to be around €650 billion and growing by more than 70% annually." (p23) | https://epoch.ai/data-insights/hyperscaler-capex-trend | 10-epoch.ai.html | figure-only |
| 11 | "According to the International AI Safety Report, plausible risks include AI-enabled biological or chemical attacks, the erosion of human autonomy, and loss of control scenarios …" (p24) | https://internationalaisafetyreport.org/publication/international-ai-safety-report-2026 | 11-internationalaisafetyreport.org.html | present |
| 12 | "AI agents developed by OpenAI coordinated through a hidden message board and eventually hacked the servers of Hugging Face, a multi-billion dollar AI company …" (p24) — UNVERIFIED (H.8) | https://www.redwoodresearch.org/research/hugging-face-incident | 12-redwoodresearch.org.html | present |
| 13 | "None of the world's ten most valuable AI companies are European." (p24) | https://www.forbes.com/lists/global2000/ | 13-forbes.com.html | present (list; "none European" is derived, not stated) |
| 14/15 | "The annualised revenue run rate of European model developers is likely less than 2% of the combined run rate of … OpenAI and Anthropic, whose combined revenue has grown roughly 25-fold in two years – from \$4 billion to over \$100 billion." (p24) | http://epoch.ai/data/ai-companies?view=graph&tab=revenue | 14-epoch.ai-aicompanies.html | figure-only (data tool) |
| 16 | "The EU only hosts around 5% of global AI compute, compared to 75% in the US and 15% in China." (p24) | https://europe2031.ai/compute-forecast/ | 16a-europe2031.ai.html | present (5% / 15% in text; 75% not located) |
| 16 | same sentence — Epoch supercomputers trend (H.5 citation orbit) (p24) | https://epoch.ai/publications/trends-in-ai-supercomputers | 16b-epoch.ai-supercomputers.html | figure-only |
| 17/18 | "a single data centre site in Malaysia will reach roughly one third of the AI compute capacity of Europe …" + footnote 2 (Nusajaya 0.662 GW, Stargate+New Carlisle 2.4 GW vs 2.1 GW) (p24) — SHARED WITH SLICE 2, H.5 | https://epoch.ai/latest/introducing-the-frontier-data-centers-hub | 17-epoch.ai-fdc-hub.html | figure-only (frontier-data-centers hub) |
| 20 | "The market already pays roughly 100 times more for access to US frontier models than to the European fast follower." (p25) | https://foxphilip.substack.com/p/the-economics-of-frontier-model-access | 20-foxphilip.substack.com.html | absent (100× figure not located in extraction) |

**No URI annotation on the PDF page — recorded, not fetched (do not guess a host):**

- **PLAN item 5** — "AI capability growth is also 'jagged': models reach superhuman performance on
  some tasks while unexpectedly lagging on others." (p23). The RL-on-verifiable-outcomes /
  'jagged' caveat paragraph (top of p23) carries no URI annotation; it is the report's own
  acknowledge-then-dismiss inference step (H.7), not a cited claim.
- **PLAN item 19** — "In 2025, all of Europe combined received just 5–6% of global private AI
  investment." (p24). No adjacent URI annotation on pp21–25; uncited in the PDF at this locus. Needs
  a hand-check (a later mention elsewhere in the report may carry the link) before it is called
  absent for the document.

**SHA-256 (over held `.html`, = raw response bytes; fetched 2026-09-16):**

```
ce9a105015744b2ed88e858a0a7f2028c04aca2f0debb635b5e6aac2e21b2756  01-epoch.ai.html
15f78bac66faa53f3a3a18b9de508bb87af7e95b51f35d535ca217028aa1264c  02-metr.org.html
7f432c8fbc6115beef14b37b544a49abc16bd7e065bed50ff59fd5cb8c9e910c  03-anthropic.com.html
d33fffdc94f545cd9adcb1b40e6b3009c446df5a723dd2abc72328f150cda45a  04-epoch.ai.html
a290775d202b81f9aba28b60fe2f05e5d850cb1266e99f4ec881a56fd1374595  06-wemustactnow.ai.html
9e2f0ff99ee5d8fa373863d0d33cde581327eb70da2cb26e59bcbc5d926b321e  07-pacingthefrontier.com.html
e13fbf5ed86e5dc1514181f524f5542bac22ebdd30af9d23d3ae7600effb540d  08a-epoch.ai.html
42907fdc8ad447b15b2f8d842933f5c64ed811bd2dde74115124cf9a0bf41807  08b-bloomberg.com.html
1c066dd96fd0423ef3c7bce3e513d0efd038a0089dae2187632c032df179016c  09-theinformation.com.html
25a8b85d158d296816e93649eb716e9a0544d21ca13d30d700ecb6f7fa025ed9  10-epoch.ai.html
b8c8d6857c934e0b9338b790443abce8cf940fdb64a1723d14d1289be8c28dc6  11-internationalaisafetyreport.org.html
ec1da01649d7ea38d625530eac9d37c8094ac5ab79cdcf7ca0897dcdc66710cc  12-redwoodresearch.org.html
900710df406182bc65f0693f372b573c7fa64ef26c4f8d3f25828b80527f6ea9  13-forbes.com.html
50a3d5cbc32890a737eb1ff72a5da2bb87a2eef23b345093bf2f1af2ca489e5e  14-epoch.ai-aicompanies.html
7e9b623cae5b6a501127c1dab49f64ae4248d26ef43d7d760a21ffc792035750  16a-europe2031.ai.html
e23414cb42a69a83e52300ae037b2ab5f1a5cbfa6ea0186a60cddc1d95530bcf  16b-epoch.ai-supercomputers.html
eaf1e917a7170107956d20c9ecdceae407b1867ca05cd830f877c64dc987fab3  17-epoch.ai-fdc-hub.html
4778ffca1b4b3cee3bf2a4725e09f52c6124be6df8fd8f5efd80735e2a925741  20-foxphilip.substack.com.html
```

Counts: **16 held** (HTTP 200, fetched + extracted), **2 hand-fetch** (08b Bloomberg 403 bot-block;
09 The Information login wall), **2 no-annotation** (items 5, 19 — uncited at this locus, not
fetched). Held with the cited figure present as text: items 3, 4, 6, 7, 8(epoch), 11, 12, 13, 16
(europe2031). Figure-only (datum in an interactive chart, not the static extraction): 1, 2, 10, 14,
16(epoch trends), 17. Absent-in-extraction: 20.

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
