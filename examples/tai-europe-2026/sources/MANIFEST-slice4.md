# Corpus manifest — Slice 4 (cited-source faithfulness: the leverage premise)

Every source document the Slice-4 corpus holds, in the report’s own naming. The backticked id is
canonical: a claim’s `cites` column names the document a claim rests on with these ids, and a
faithfulness run marks a claim `unverifiable` (no model call) when a cited id is not listed here.

This is the **Slice-4** manifest. Slice 4 (PLAN.md §Slice 4) is the LEVERAGE PREMISE — every sourced
factual claim, in IO-1 (Part A pp28–30, Part B pp79–89) and the compute deep dive (pp189–196), about
what Europe controls in the AI chip supply chain and what leverage it buys — one `evidence`-route leaf
per row in `../claims-slice4.txt`. It grounds each leaf against the report’s OWN external citation — a
fetched web page held under `sources/cited/4x-*` — not another part of the report. `MANIFEST.md` stays
the Slice-1/2 manifest and `MANIFEST-slice3.md` the Slice-3 one; this file is loaded only for a Slice-4 run.

## Roles for Slice 4 (faithfulness config)

The report is the TARGET; the cited web pages are the SOURCES. As in Slice 3, this inverts Slice 1/2’s
held-report arrangement: the report is the thing under test and is NOT held, while each external
truth-maker it cites IS.

- **TARGET (assayed, not held):** the report’s sentences (`../claims-slice4.txt`, verbatim from
  `report.txt`). Not in this manifest — grounding a report sentence on the report’s own wording would
  be trivial.
- **SOURCE (held):** the 12 fetched cited pages below, held under `sources/cited/` as `4x-NN-<host>.txt`
  (stdlib HTMLParser strip of the raw `.html`). The check is cite-scoped faithfulness — each leaf is
  judged against ONLY its cited page’s passages (spec/TREE.md § Cite-scoped retrieval).

**No `single_source` line.** Claim and truth-maker are two DIFFERENT documents (report vs. cited page),
so the single-source judge rules are off: a leaf `absent` here means “the held cited page does not carry
this claim”, a genuine grounding miss against a second document — not “said once, not repeated”.
`LoadSingleSource` returns false for this file.

Structurally weakest verdict: `evidence` grounding proves only that the cited page was retrieved and
carries the claim, never that the report’s inferential use of it is sound (CLAUDE.md § axis boundary).
Read `faithful` here as “the cited page says it”, nothing larger — and read the two **interested-party**
sources (H.6) and the two **citation-orbit** sources (H.5) below more skeptically still.

## Scope gaps flagged to PW (do not fabricate to close them)

- The report’s strongest **“share of the layer” / monopoly** claim — “ASML (Netherlands) and suppliers
  such as Zeiss (Germany) or Trumpf (Germany) hold monopolies over relevant [layers of the AI stack]”,
  Figure 10 (report.txt:2818, objective O2.2, ~p60) — sits OUTSIDE the scoped pages. Slice 4 grounds the
  assets ENUMERATION (s01) and the EUV-monopoly step (s02) but not that Figure-10 layer-monopoly
  assertion, which is the leverage thesis’s load-bearing claim. Its own citation is a separate fetch.
- **ASM** (ASM International, ALD equipment), named in the task, does not appear anywhere in the report
  (grep `report.txt`). No leaf; not fabricated.

## Not held — named so their absence is `unverifiable`, never silently `absent`

The ENISA leaf (s10) cites a NOT-HELD id on purpose (see `../claims-slice4.txt` header). Both its
truth-makers are walls fetched but not usably in the corpus: `4x-11-euractiv.com` (Cloudflare
“Just a moment…” interstitial, HTTP 403) and the Bloomberg alternate `4x-14-bloomberg.com`
(“Are you a robot?”, HTTP 403). s10 cites the euractiv file, left out of the held set below, so it
renders `unverifiable` (no model call: “cited document(s) not in corpus”), never silently grounded.

## Held cited ids (registered for `LoadHeld`)

The 12 captures that fetched HTTP 200 and are usable as text (each a backticked `- ` bullet so
`manifest.LoadHeld` sees it). `fig`: `present` (the claim is in the extracted text), `figure-only`
(page renders the datum only in an interactive chart / data tool, so it is not in the static
extraction), `headline-only` (only the page title extracted; body client-rendered) — read
`figure-only` and `headline-only` leaves most skeptically. Full URL / SHA-256 provenance is the
§ Slice-4 cited sources table below.

- `cited/4x-01-chipexplorer.eto.tech.txt` — ETO Advanced-Chips Supply Chain Explorer (assets, p28) — figure-only
- `cited/4x-02-dwarkesh.com.txt` — Dwarkesh × Dylan Patel on EUV/TSMC bottlenecks (p192) — present
- `cited/4x-03-tomshardware.com.txt` — Samsung/SK hynix HBM shortage to 2027 (p192) — present
- `cited/4x-04-chosun.com.txt` — TSMC 2nm sold out (p192) — headline-only
- `cited/4x-05-epochai.substack.com.txt` — Epoch “Is a compute crunch coming?” (p190) — present
- `cited/4x-06-semianalysis.com.txt` — Semianalysis GPU-shortage / rental capacity (p190) — present
- `cited/4x-07-europe2031.ai.txt` — europe2031 compute forecast (5% EU, p192) — present (H.5 citation orbit)
- `cited/4x-08-epoch.ai.txt` — Epoch trends-in-AI-supercomputers (5% EU, p192) — present (H.5)
- `cited/4x-09-anthropic.com-glasswing.txt` — Anthropic Project Glasswing (p28) — present (INTERESTED PARTY, H.6)
- `cited/4x-10-anthropic.com-fable-mythos.txt` — Anthropic Fable/Mythos access notice (p28) — present (INTERESTED PARTY, H.6)
- `cited/4x-12-europarl.europa.eu.txt` — EP debate: US export controls of AI chips (p28) — present
- `cited/4x-13-antonleicht.me.txt` — “When should nations sell their data” (leverage framing, p28) — present

| leaf | anchor sentence (verbatim, `report.txt` page) | URL (PDF URI annotation) | file | fig |
|------|-----------------------------------------------|--------------------------|------|-----|
| s01 | “Member States host various assets … world-leading chipmaking equipment (e.g. ASML’s … EUV … machines), critical suppliers … (e.g. Zeiss … Trumpf …), and proprietary datasets …” (p28) | https://chipexplorer.eto.tech/ | 4x-01-chipexplorer.eto.tech.html | figure-only (data tool) |
| s02 | “… EUV … machines to become the biggest bottleneck – … only one company in the world, the Dutch ASML, is able to produce them.” (p192) | https://www.dwarkesh.com/p/dylan-patel | 4x-02-dwarkesh.com.html | present |
| s03 | “… high-bandwidth memory … is fully sold out for 2026 … could persist through 2027 and beyond.” (p192) | https://www.tomshardware.com/tech-industry/artificial-intelligence/samsung-and-sk-hynix-warn-ai-driven-memory-shortages-could-last-until-2027-and-beyond-… | 4x-03-tomshardware.com.html | present |
| s04 | “TSMC’s 2-nanometer chip production process … is reportedly sold out until 2028.” (p192) | https://www.chosun.com/english/industry-en/2026/03/29/JRVMGDAFLFCUVDACCSFX4CF37A/ | 4x-04-chosun.com.html | headline-only (“TSMC 2-Nanometer Sold Out”; body client-rendered) |
| s05 | “Epoch AI finds that global inference supply … is growing ~3.4x per year … demand … closer to 10x per year.” (p190) | https://epochai.substack.com/p/is-a-compute-crunch-coming | 4x-05-epochai.substack.com.html | present |
| s06 | “‘on-demand GPU rental capacity is sold out across all GPU types’ … ‘the last flight out’ …” (p190) | https://newsletter.semianalysis.com/p/the-great-gpu-shortage-rental-capacity | 4x-06-semianalysis.com.html | present |
| s07a | “… such as the EU, which only hosts around 5% of global AI compute …” (p192) | https://europe2031.ai/compute-forecast/ | 4x-07-europe2031.ai.html | present (H.5) |
| s07b | same sentence — Epoch AI-supercomputers trend (p192) | https://epoch.ai/publications/trends-in-ai-supercomputers | 4x-08-epoch.ai.html | present (H.5) |
| s08 | “… in April 2026, Anthropic shared its most capable model with only selected partners under Project Glasswing …” (p28) | https://www.anthropic.com/glasswing | 4x-09-anthropic.com-glasswing.html | present (H.6) |
| s09 | “… in June 2026, the US government issued an export control directive … forcing Anthropic to disable these models for all customers.” (p28) | https://www.anthropic.com/news/fable-mythos-access | 4x-10-anthropic.com-fable-mythos.html | present (H.6) |
| s10 | “… ENISA … only got invited to access the model after sustained negotiations …” (p28) | https://www.euractiv.com/news/anthropics-mythos-update-puts-eu-back-on-hold-over-ai-access/ | 4x-11-euractiv.com.html | **wall** — HTTP 403 Cloudflare (“Just a moment…”) |
| s10 (alt) | same sentence — Bloomberg on the ENISA/Mythos grant (p28) | https://www.bloomberg.com/news/articles/2026-06-01/anthropic-to-give-eu-s-cybersecurity-agency-access-to-mythos | 4x-14-bloomberg.com.html | **wall** — HTTP 403 (“Are you a robot?”) |
| s11 | “Europe’s companies also do not currently get preferential access to the chips built with European equipment …” (p28) | https://www.europarl.europa.eu/news/en/agenda/plenary-news/2025-02-10/6/us-export-controls-of-ai-chips-debate-with-the-commission | 4x-12-europarl.europa.eu.html | present (debate agenda; preferentiality clause is the report’s gloss) |
| s12 | “They do hold various indirect powers – such as export controls, investment screening or … the Anti-Coercion Instrument (ACI) – that can be used to turn valuable assets into concrete leverage …” (p28) | https://writing.antonleicht.me/p/when-should-nations-sell-their-data | 4x-13-antonleicht.me.html | present (leverage-framing essay; ACI/screening list is the report’s) |

**SHA-256 (over held `.html`, = raw response bytes; fetched 2026-09-16 with `curl`, Chrome UA):**

```
9aba73eeb8968f344f1fdc559810b1846cffac1098fc8885242eb9e00fd3a3e5  4x-01-chipexplorer.eto.tech.html
2976b927bac4717fdc1e035c82fc1bebb79d721f7b20d4a072ae83cf818ac9c5  4x-02-dwarkesh.com.html
9fcce4f53eb5e1b5eaaec8c07bd70a1e3c1f8c2202593882a16dd4208990c588  4x-03-tomshardware.com.html
af507b62053224470ae76361a974425cec0bfced62197323f08de0e5a2ccb10d  4x-04-chosun.com.html
2dae87b7f4fb1442fc005283f485a55ed93e268c5ed9c4653df18ce468b9d265  4x-05-epochai.substack.com.html
993d44ed2e073c1751f109123c8785a9758469c1845de700f416d2774937e760  4x-06-semianalysis.com.html
7e9b623cae5b6a501127c1dab49f64ae4248d26ef43d7d760a21ffc792035750  4x-07-europe2031.ai.html
e23414cb42a69a83e52300ae037b2ab5f1a5cbfa6ea0186a60cddc1d95530bcf  4x-08-epoch.ai.html
7524f1bbb976a090d416274c62021f478e55fbbbe8c2b05bf49144e2d14e3b7d  4x-09-anthropic.com-glasswing.html
57baaedcf0b514bd96f655ac6f2ea1c5939669e9720e61d1426ee1bb8be27d62  4x-10-anthropic.com-fable-mythos.html
dcefaf0a56e69ebd14f8ca012c05cf10874ba337fe54e0863dba14148ca5c974  4x-11-euractiv.com.html
333b4da0cfb7540325eeec79f2563ef3c564cb9057fbe5eb6af048c30ee93d23  4x-12-europarl.europa.eu.html
e7a4aa7cd8ef76db88ad92a994de9aeb911972157f5c63829dacc53ad7ad9802  4x-13-antonleicht.me.html
acf715096d54a5bf0b74c8879b79dc2760a6232d9b4298f27f1c3362c75cca73  4x-14-bloomberg.com.html
```

## Report-link resolution

`report_page_offset` 0 (VERIFIED Slice 1): `pdftotext -layout` emits one `\f` per PDF page, so printed
page N sits on PDF page N. Every URL above was read off the PDF’s own URI annotation with pypdf (PDF
pages 28 and 190–192, by rect y-coordinate against the text line — never guessed).

report_page_offset: 0   # VERIFIED 2026-09-15 (Slice 1); printed page N == PDF page N.
