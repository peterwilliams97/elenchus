# Provenance — LCEIC cultural & creative industries faithfulness sources

All sources fetched **2026-09-06** from parliament.vic.gov.au for a faithfulness assay of the LCEIC
final report. Raw hearing transcripts live under `sources/hearings/` (gitignored: analysis, not
redistribution); this file records every URL so the set is reconstructable. Derived artifacts
(`report-summary.md`, `submissions-index.md`, `claims-machine.txt`) are tracked.

## Report under analysis (the claim side)

- *Inquiry into the cultural and creative industries in Victoria* — Final report, LCEIC, June 2025.
- https://www.parliament.vic.gov.au/4a5044/contentassets/303bc6816a0f43068227fdfc9ae79b59/lceic-60-06-cultural-and-creative-industries-vic.pdf
- Local read copy: `~/Downloads/lceic-60-06-cultural-and-creative-industries-vic.pdf` (160 pp, 4 MB).

## Public hearings (the source side) — 3 days, 15 session transcript PDFs

The Committee held **3 hearing days**; the site publishes **one PDF per witness session** (15
total), not one file per day. Witness rosters per day are in the report's Appendix A.2 (printed
p92–93). Base host: `https://www.parliament.vic.gov.au`.

### 27 February 2025 — Davui Room, East Melbourne (`sources/hearings/2025-02-27/`)
- `1_yarra-city-council.pdf` — /4a4389/contentassets/95e4a206af154745aa4086feaf02048d/final_1.-yarra-city-council.pdf
- `2_regional-arts-victoria.pdf` — /4a438a/contentassets/aaab3c107f92459ea255393bcd8a6a53/final_2.-regional-arts-victoria.pdf
- `3_multicultural-arts-victoria.pdf` — /4a438b/contentassets/d5a3a79ecd36458d8f385e776c40c025/final_3.-multicultural-arts-victoria.pdf
- `4_abc.pdf` — /4a438c/contentassets/bc8a0be32cce41e5b5ee35bbaf2609c4/final_4.-australian-broadcasting-corporation.pdf

### 12 March 2025 — Davui Room, East Melbourne (`sources/hearings/2025-03-12/`)
- `1_creative-victoria-and-vicscreen.pdf` — /4a438f/contentassets/df11c2bdaefd4333956c9207204f107b/final_1.-creative-victoria-and-vicscreen.pdf
- `2_community-music-victoria.pdf` — /4a4390/contentassets/607df5f21b8449599e9de27faddb84ce/final_2.-community-music-victoria.pdf
- `3_theatre-network-australia.pdf` — /4a4391/contentassets/4c00f9c4fb0b43f0b0e740813be0aa21/final_3.-theatre-network-australia.pdf
- `4_association-of-artist-managers-and-music-victoria.pdf` — /4a4391/contentassets/1e54e69ffba6450db42f0bed496b72d3/final_4.-association-of-artists-managers.pdf
- `5_sbs.pdf` — /4a4392/contentassets/ba3a1f0f25124fe1b2e201f239130e93/final_5.-sbs.pdf

### 13 March 2025 — Davui Room, East Melbourne (`sources/hearings/2025-03-13/`)
- `1_st-martins-and-theatre-works.pdf` — /4a4393/contentassets/fcc6cbd2f97c4941b5cbbfb73430f54a/final_01.-st-martins-youth-arts-centre-and-theatre-works.pdf
- `2_arena-rawcus-lamama.pdf` — /4a4393/contentassets/f3001afd8e544542bc7f494a29e8ffe9/final_02.-arena-theatre-company-rawcus-theatre-company-and-la-mama-theatre.pdf
- `3_a-new-approach.pdf` — /4a4394/contentassets/aa3fa5145def404ea0421676ce9a88f1/final_03.-a-new-approach.pdf
- `4_public-galleries-association-of-victoria.pdf` — /4a4395/contentassets/9b5004afd292463c8e9d30a059393f28/final_04.-public-galleries-association-of-victoria.pdf
- `5_australian-museums-and-galleries-association-victoria.pdf` — /4a4395/contentassets/8157253c75bd493cb7ae02a30d44adf5/final_05.-australian-museums-and-galleries-association-victoria.pdf
- `6_bendigo-theatre-company-and-sertori-consulting.pdf` — /4a4396/contentassets/94a1cb63b0c44b38a38e9785dae8fb93/final_06.-bendigo-theatre-company-and-sertori-consulting.pdf

All 15 downloaded successfully (187–217 KB each, valid PDF v1.6).

## Submissions index — 42 entries

Source of the index: the report's **Appendix A.1** (printed p91), transcribed verbatim into
`submissions-index.md`. The live submissions page is JS-rendered and returned no machine-readable
list on fetch; individual PDFs are per-submission `contentassets` URLs (three confirmed, in the
index). No submission URLs are guessed. Non-public entries (name withheld ×2, confidential ×3, and
one `TEST PATTERN` placeholder row) are marked in the index.

## Gaps / not fetched (honest denominator)

- Individual submission PDFs (the actual faithfulness source content for submission-cited claims):
  **not** downloaded — only the index was requested. 3 of 42 URLs are confirmed; the rest need the
  SPA rendered or a per-submission lookup.
- The site search summary said "2 days of public hearings"; the report's Appendix A.2 says **3 days**
  (27 Feb, 12 Mar, 13 Mar). The report is treated as authoritative.

## Written submissions (the source side) — 42 PDFs, fetched 2026-09-07

Harvested from the inquiry submissions listing
(`/get-involved/inquiries/inquiry-into-the-cultural-and-creative-industries-in-victoria/submissions/`),
a JS-rendered EPiServer/Optimizely SPA paginated 10-per-page ("Showing 1 to 10 of 42 records").
The static HTML exposes no PDF links, so the page was rendered headless with Playwright (throwaway
Chromium context, no profile, parliament.vic.gov.au only) and the `.pdf` hrefs read from the live
DOM across all 5 pages — no URLs guessed. **The `contentassets` hashes differ from those in the
2026-09-06 `submissions-index.md`, which now 404**; these are the live URLs as of 2026-09-07.
42 PDF links (39 distinct submission numbers; some submissions carry an attachment PDF, e.g. 01.1,
09.1). Downloaded to `sources/submissions/` (gitignored: analysis, not redistribution), each
verified as a real PDF (`%PDF` header, HTTP 200) and pdftotext-extracted to a sibling `.txt`.

- `01.-coalition-against-duck-shooting.pdf` — /4a436a/contentassets/17f4ca0472e74420a5a0b5d4ef4662ef/submission-documents/01.-coalition-against-duck-shooting.pdf
- `27.-ausdance-vic.pdf` — /4a437a/contentassets/84e466f4d4fa4de281dfd2d9663ef1b1/submission-documents/27.-ausdance-vic.pdf
- `28.-nets-victoria.pdf` — /4a437d/contentassets/af68f09a22894113b9b7ec4ed967c569/submission-documents/28.-nets-victoria.pdf
- `32.-la-mama-theatre.pdf` — /4a437e/contentassets/4b6a3fadd32148e59cfffcee1ee101a5/submission-documents/32.-la-mama-theatre.pdf
- `31.-community-music-victoria.pdf` — /4a437e/contentassets/8beaac93a31743a799742834702a309a/submission-documents/31.-community-music-victoria.pdf
- `33.-public-galleries-association-of-victoria-redacted.pdf` — /4a437e/contentassets/a2ddb3b351d44e6fa94b79396d453560/submission-documents/33.-public-galleries-association-of-victoria-redacted.pdf
- `34.-university-of-melbourne_redacted.pdf` — /4a4381/contentassets/2f908c6dcf9b4bc68a7bdffacc84bded/submission-documents/34.-university-of-melbourne_redacted.pdf
- `33.1-public-galleries-association-of-victoria_redacted.pdf` — /4a4381/contentassets/a2ddb3b351d44e6fa94b79396d453560/attachment-documents/33.1-public-galleries-association-of-victoria_redacted.pdf
- `36.-martin-jackson.pdf` — /4a4382/contentassets/3a92e4fe08af420b83ce69db2d3d4b67/submission-documents/36.-martin-jackson.pdf
- `35.-amaga-victoria.pdf` — /4a4382/contentassets/48cc3dd9baf043e59daf5f0f80e1ae72/submission-documents/35.-amaga-victoria.pdf
- `38.-nirmidha-sankar_redacted.pdf` — /4a4383/contentassets/0be45c11c8934f6fa9e31aa1429c996c/submission-documents/38.-nirmidha-sankar_redacted.pdf
- `37.-sense-and-centsability_redacted.pdf` — /4a4383/contentassets/4258050f13d845079932006a7239a026/submission-documents/37.-sense-and-centsability_redacted.pdf
- `39.-live-performance-australia_redacted.pdf` — /4a4383/contentassets/cd8a90a17a54485483b0dd694c3d3fc1/submission-documents/39.-live-performance-australia_redacted.pdf
- `40.-association-of-artist-managers.pdf` — /4a4384/contentassets/41fdda6b1a5a4e4a8647dcb850064df3/submission-documents/40.-association-of-artist-managers.pdf
- `41.-abc.pdf` — /4a4385/contentassets/93cdd91de05f42dbb3ad049226537760/submission-documents/41.-abc.pdf
- `42.-sbs.pdf` — /4a4385/contentassets/d273ac3b866c428ba647cc1c4ebf15f0/submission-documents/42.-sbs.pdf
- `01.1-coalition-against-duck-shooting.pdf` — /4a4391/contentassets/17f4ca0472e74420a5a0b5d4ef4662ef/attachment-documents/01.1-coalition-against-duck-shooting.pdf
- `02.-name-withheld.pdf` — /4a4391/contentassets/f07417999a7447ca93b575894f37da58/submission-documents/02.-name-withheld.pdf
- `03.-craig-coulson.pdf` — /4a4392/contentassets/9873c21420c54e7c80907bb5a7ae7450/submission-documents/03.-craig-coulson.pdf
- `04.-robert-heron_redacted.pdf` — /4a4392/contentassets/a082bbf8024b495fa6b39a7cb9b0f777/submission-documents/04.-robert-heron_redacted.pdf
- `05.-name-withheld.pdf` — /4a4392/contentassets/ef983b0bc6aa4cf2ac6ce5e24a043e6a/submission-documents/05.-name-withheld.pdf
- `06.-bo-kitty_redacted.pdf` — /4a4394/contentassets/1efee370cbac4636bb4dfd89d73f2560/submission-documents/06.-bo-kitty_redacted.pdf
- `07.-sophie-travers.pdf` — /4a4394/contentassets/bba69401a8de4a3f9b534dc9273223af/submission-documents/07.-sophie-travers.pdf
- `09.-ana-a-new-approach-redacted.pdf` — /4a4395/contentassets/2068ba1ac6bd4b4bb302637c8dac7601/submission-documents/09.-ana-a-new-approach-redacted.pdf
- `08.-australian-publishers-association_redacted.pdf` — /4a4395/contentassets/7721f8ca0c4c415588e2e1857301cdaf/submission-documents/08.-australian-publishers-association_redacted.pdf
- `09.1-ana-a-new-approach-redacted.pdf` — /4a4396/contentassets/2068ba1ac6bd4b4bb302637c8dac7601/attachment-documents/09.1-ana-a-new-approach-redacted.pdf
- `11.-desmond-beer.pdf` — /4a4396/contentassets/8d2907b4327244aa9b59280c69bd6566/submission-documents/11.-desmond-beer.pdf
- `10.-interactive-games-and-entertainment-association-redacted.pdf` — /4a4396/contentassets/9237d3390f694390aa5106f90217ffb2/submission-documents/10.-interactive-games-and-entertainment-association-redacted.pdf
- `13.-kate-larsen-redacted.pdf` — /4a4397/contentassets/2383b1d956e9411cbcc726d3435a823e/submission-documents/13.-kate-larsen-redacted.pdf
- `12.-melbourne-fringe.pdf` — /4a4397/contentassets/c454be33b55f4116a56edaf20a5ffa95/submission-documents/12.-melbourne-fringe.pdf
- `15.-regional-arts-victoria.pdf` — /4a4399/contentassets/0ed66d51cee94e16a689fc4e33b95d52/submission-documents/15.-regional-arts-victoria.pdf
- `14.-parliamentary-budget-office_redacted.pdf` — /4a4399/contentassets/d1489a470437496b88ef1131319754bf/submission-documents/14.-parliamentary-budget-office_redacted.pdf
- `16.-stonnington-council_redacted.pdf` — /4a439a/contentassets/04f06b887601479bb9eef5f38b55b2a8/submission-documents/16.-stonnington-council_redacted.pdf
- `17.-sir-zelman-cowen-school-of-music-and-performance.pdf` — /4a439a/contentassets/fb437a0ae9314b629f8098f747612bfe/submission-documents/17.-sir-zelman-cowen-school-of-music-and-performance.pdf
- `19.-theatre-network-australia.pdf` — /4a439b/contentassets/0c6cb1f32f8c43fc971604a68d4c8ed0/submission-documents/19.-theatre-network-australia.pdf
- `18.-city-of-yarra-redacted.pdf` — /4a439b/contentassets/dca486cad6324875898035050d1a16bd/submission-documents/18.-city-of-yarra-redacted.pdf
- `20.-victorian-major-arts-festivals-alliance_redacted.pdf` — /4a439c/contentassets/0ab9db84b837404299c4fff6e40122e8/submission-documents/20.-victorian-major-arts-festivals-alliance_redacted.pdf
- `21.-test-pattern_redacted.pdf` — /4a439c/contentassets/db87743d53134a9b841f28b7cdacfd43/submission-documents/21.-test-pattern_redacted.pdf
- `23.-mva.pdf` — /4a439d/contentassets/1d35701c00f04ec3be78eb35d0ff7f5e/submission-documents/23.-mva.pdf
- `22.-awg-and-awgacs.pdf` — /4a439d/contentassets/46e1e8c8e2b4486aae8eaa3d60aaef71/submission-documents/22.-awg-and-awgacs.pdf
- `26.-victorian-independent-and-youth-theatre-organisations.pdf` — /4a439e/contentassets/9d088196b15a46c19655f2e51e47e879/submission-documents/26.-victorian-independent-and-youth-theatre-organisations.pdf
- `25.-arena-theatre-co_redacted.pdf` — /4a439e/contentassets/9d73d945f3ec4adc98dfb8ca07640e94/submission-documents/25.-arena-theatre-co_redacted.pdf
