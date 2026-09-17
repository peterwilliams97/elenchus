# Corpus manifest — Slice 3 (cited-source faithfulness)

Every source document the corpus holds, in the report's own naming. The backticked id is
canonical: a claim's `cites` column names the documents a claim rests on with these ids, and a
faithfulness run marks a claim `unverifiable` (no model call) when a cited id is not listed here.

This is the **Slice-3** manifest. Slice 3 (PLAN.md §Slice 3) is the empirical diagnostic section
"Why Europe needs to prepare for transformative AI" (printed pp20–26), one `evidence`-route leaf per
row in `../claims-slice3.txt`. It grounds each leaf against the report's OWN external citation — a
fetched web page held under `sources/cited/` — not another part of the report. `MANIFEST.md` stays
the Slice-1/2 manifest; this file is loaded only for a Slice-3 run.

## Roles for Slice 3 (faithfulness config)

The report is the TARGET; the cited web pages are the SOURCES. This inverts Slice 1/2's held-report
arrangement: there the report (Part B / footnote 2) was the held source; here the report is the thing
under test and is NOT held, while each external truth-maker it cites IS.

- **TARGET (assayed, not held):** the report's sentences (`../claims-slice3.txt`, reconstructed
  verbatim from `report.txt`). Not in this manifest — the check is whether the cited page supports the
  number the report attributes to it, so grounding a report sentence on the report's own wording would
  be trivial. `report-partB.txt` and `report-fn2.txt` are deliberately NOT held here.
- **SOURCE (held):** the 16 fetched cited pages below, one per load-bearing citation, held under
  `sources/cited/` as `<stem>.txt` (stdlib HTMLParser strip of the raw `.html`). The check is
  cite-scoped faithfulness — each leaf is judged against ONLY its cited page's passages
  (spec/TREE.md § Cite-scoped retrieval), the segmenter splitting each capture into ≥3 paragraph
  passages under the `cited/<stem>` id base.

**No `single_source` line.** Claim and truth-maker are two DIFFERENT documents (report vs. cited
page), so the single-source judge rules are off: a leaf `absent` here means "the held cited page does
not carry this number", a genuine grounding miss against a second document — not "said once, not
repeated". `LoadSingleSource` returns false for this file.

Structurally weakest verdict: `evidence` grounding proves only that the cited page was retrieved and
carries the figure, never that the report's inferential use of it is sound (CLAUDE.md § axis
boundary). Read `faithful` here as "the cited page says the number", nothing larger.

## Not held — named so their absence is `unverifiable`, never silently `absent`

Four Slice-3 leaves cite a NOT-HELD id on purpose (see `../claims-slice3.txt` header): `08b`
(Bloomberg, HTTP 403 bot-block) and `09` (The Information, login wall) were fetched but are walls, so
not usably in the corpus; `05` (`uncited/p23-jagged`) and `19` (`uncited/p24-private-investment`)
carry no URI annotation on the PDF, so no held document answers to them. Each names its (un)available
truth-maker and is left out of the held set below, so the leaf renders `unverifiable` (no model call:
"cited document(s) not in corpus"), never silently grounded.

## Held cited ids (registered for `LoadHeld`)

The 16 captures that fetched HTTP 200 and are usable as text. Each is a backticked `- ` bullet so
`manifest.LoadHeld` sees it: a Slice-3 leaf citing one of these is judged cite-scoped against the
capture's `cited/<stem>` passages; one citing an id NOT in this list is `unverifiable`. The `fig`
tag is whether the cited figure is present as extractable text (`present`), rendered only in an
interactive chart / data tool (`figure-only`), or held-but-not-located (`absent`) — read `figure-only`
and `absent` leaves most skeptically. Full URL / SHA-256 provenance for every id is the
§ Slice-3 cited sources table in `MANIFEST.md`; this file registers the ids, that file records the bytes.

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

## Held cited DATA ids — Epoch chart CSVs (registered for `LoadHeld`)

The `figure-only` cited pages above render their numbers only as interactive Epoch charts / data
tools, so the static HTML strip carries no figure (why `16b`, `14` were pre-registered `absent`). PW
fetched the underlying CSV exports on 17 Sept 2026, held under `sources/cited/data/` as raw `.csv`.
Each is a backticked `- ` bullet so `manifest.LoadHeld` registers it; the id is `cited/data/<file>`,
extension kept. SHA-256 recorded inline here per this addendum — this deviates from the "bytes live
in `MANIFEST.md`" note above, at PW's instruction that the data-addendum provenance sit beside the
ids it registers.

**NOT YET RETRIEVABLE (a fourth Slice-3 gap, CSV-specific).** `retrieve.Load` walks `.txt` only
(`internal/retrieve/retrieve.go:131`), so a `.csv` file is never ingested; `citedExternalBases`
(`assay.go:1007`) maps only `paper:`/`.txt` ids, so a `.csv` cite yields no base; and the plain
segmenter splits on blank lines (`internal/retrieve/segment.go:64`), which a CSV has none of, so it
would emit ONE passage for the whole file, not one per row — and CSV records span multiple physical
lines (quoted `Note`/`Quote` fields wrap), so a line-based split is wrong too. A CSV rule
(`encoding/csv`; header + one data row per passage; id base `cited/data/<file>`) plus a `.csv` case
in `Load`'s walker and in `citedExternalBases` must land before a Slice-3 run grounds any leaf on
these. That is Go code on its own branch, not folded into this fixture edit — see PLAN.md §Slice 3
"Data addendum". Until it lands, the three re-pointed leaves render `absent` on their secondary page
cite, not `faithful`.

Backs a Slice-3 leaf:

- `cited/data/ai_supercomputers.csv` — Epoch AI-supercomputers dataset (16b, p24; Status/Country/Power Capacity MW/H100 equivalents) — sha256:b6c20761d425fc9f3a9da9e795a72f33bbd8abbd39a39c5c1b85b9d4c3cd4b9d
- `cited/data/14-ai_companies_revenue_reports.csv` — Epoch AI-companies revenue reports (14 + 08a, p23–24; annualised revenue by company/date) — sha256:52e5727c0f8a3111ca69064b6d3579ec474d361ec759487caa110eb34bc391f2

Held, no Slice-3 leaf cites them yet (siblings of the two exports above):

- `cited/data/14-ai_companies.csv` — Epoch AI-companies index — sha256:9d69565e1091972d1345e7afe3dba780497933fe9500642917dea2c1ad10c552
- `cited/data/14-ai_companies_compute_spend.csv` — Epoch AI-companies compute spend — sha256:afd48524a17d81c5bea17daf3e63acf97587c34171d59ab402bced2ffb462f82
- `cited/data/14-ai_companies_funding_rounds.csv` — Epoch AI-companies funding rounds — sha256:fd912cf2bb692316c3462e793095c35bccc12d59d160b2d717e4171591d6add2
- `cited/data/14-ai_companies_staff_reports.csv` — Epoch AI-companies staff reports — sha256:c000b7b427720ef3d14557188e069bafbb71ad31cdbeb9282703f86a9c89f1a7
- `cited/data/14-ai_companies_usage_reports.csv` — Epoch AI-companies usage reports — sha256:ebc5da027521115f1b0042d1db848abd14b5b4cda62d7263bda99ef4a7525474
- `cited/data/ai_models.csv` — Epoch notable AI models — sha256:4752692efd4d6b66f212d02e82428f56f3d538882af76167c6a9c8750d6ec05f

## Report-link resolution

`report_page_offset` is the printed→PDF page offset, for `-review` deep-links from a Slice-3 leaf's
`pNN` page field back into `report.pdf#page=N`. Verified 0 for this report (Slice 1): `pdftotext
-layout` emits one `\f` per PDF page, so printed page N sits on PDF page N. Slice-3 leaves carry a
printed `pNN` directly (no named `sections:` table — that block is Slice-1's Part A objective pages
and lives in `MANIFEST.md`).

report_page_offset: 0   # VERIFIED 2026-09-15 (Slice 1); printed page N == PDF page N.
