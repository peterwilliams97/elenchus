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
