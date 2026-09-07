# Submissions index — LCEIC cultural & creative industries inquiry

**42 submissions**, transcribed verbatim from the report's **Appendix A.1** (printed p91), which is
the Committee's own authoritative list. The live submissions page
(`/inquiry-into-the-cultural-and-creative-industries-in-victoria/submissions`) is JS-rendered and
did not return a machine-readable list on fetch; individual submission PDFs are hosted on
parliament.vic.gov.au under per-submission `contentassets` hashes. A full URL-per-submission harvest
was **not** performed (would need the SPA rendered) and no URLs are guessed here.

## Fetch status — 2026-09-07 (verified this session, no URLs guessed)

**The three PDF URLs previously listed below are all dead (HTTP 404).** Re-checked with a browser
user-agent: submissions 9, 10, and 15 each return a 404 HTML error page, not a PDF. The site itself
is up and `contentassets` still serves PDFs — the report PDF
(`.../49eafb/contentassets/…/lceic-60-06-cultural-and-creative-industries-vic.pdf`) and the short
landing page `parliament.vic.gov.au/culturalcreativeindustries` both return 200 — so it is the
per-submission `…/submission-documents/NN.-name.pdf` paths specifically that were reorganised.

The current submissions listing
(`/get-involved/inquiries/inquiry-into-the-cultural-and-creative-industries-in-victoria/submissions/`)
is an **EPiServer/Optimizely SPA** (`/Static/assets/index-CIDfNjQj.js`, `find.js`), paginated
10-per-page ("Showing 1 to 10 of 42 records"): its server HTML carries **zero** submission PDF links.

**Resolved 2026-09-07 — fetch completed.** The SPA was rendered headless with Playwright (throwaway
Chromium, no profile, parliament.vic.gov.au only) and the `.pdf` hrefs read from the live DOM across
all 5 pages — no URLs guessed. **42 PDF links** (39 distinct submissions + attachments) downloaded to
`sources/submissions/` (gitignored), each verified as a real PDF and pdftotext-extracted. Every live
URL and the method are recorded in `PROVENANCE.md` (§ Written submissions). The `contentassets` hashes
below in the per-submission table are the **old, dead** ones; the live URLs are in `PROVENANCE.md`.
The table below is kept for the submitter names (from Appendix A.1); treat its URL column as stale.

Non-public entries are marked: submitters who withheld their name, confidential submissions, and one
placeholder row (`21 TEST PATTERN`) that appears verbatim in the source index.

| # | Submitter (verbatim) | Status | PDF URL |
|---|---|---|---|
| 1 | Coalition Against Duck Shooting | public | — |
| 2 | Name withheld | name withheld | — |
| 3 | Craig Coulson | public | — |
| 4 | Robert Heron | public | — |
| 5 | Name withheld | name withheld | — |
| 6 | Bo Kitty | public | — |
| 7 | Sophie Travers | public | — |
| 8 | Australian Publishers Association | public | — |
| 9 | A New Approach (ANA) | public (redacted) | https://www.parliament.vic.gov.au/495dc8/contentassets/2da66ae5d5b14de4baa761b648ff3554/submission-documents/09.-ana-a-new-approach-redacted.pdf |
| 10 | Interactive Games & Entertainment Association | public (redacted) | https://www.parliament.vic.gov.au/4907d2/contentassets/b5384136b831464ea487cc32e2e96181/submission-documents/10.-interactive-games-and-entertainment-association-redacted.pdf |
| 11 | Desmond Beer | public | — |
| 12 | Melbourne Fringe | public | — |
| 13 | Kate Larson | public | — |
| 14 | Parliamentary Budget Office | public | — |
| 15 | Regional Arts Victoria | public | https://www.parliament.vic.gov.au/495dc0/contentassets/f1c315520c7d423d8fd04b5860c5c7cd/15.-regional-arts-victoria.pdf |
| 16 | City Of Stonnington | public | — |
| 17 | Sir Zelman Cowen School of Music & Performance, Monash University | public | — |
| 18 | City of Yarra | public | — |
| 19 | Theatre Network Australia | public | — |
| 20 | Victorian Major Arts Festivals Alliance | public | — |
| 21 | TEST PATTERN | placeholder row in source index | — |
| 22 | AWG and AWGACS | public | — |
| 23 | MVA | public | — |
| 24 | Confidential | confidential | — |
| 25 | Arena Theatre Co | public | — |
| 26 | Victorian Independent and Youth Theatre Organisations | public | — |
| 27 | Ausdance Vic | public | — |
| 28 | NETS Victoria | public | — |
| 29 | Confidential | confidential | — |
| 30 | Confidential | confidential | — |
| 31 | Community Music Victoria | public | — |
| 32 | La Mama Theatre | public | — |
| 33 | PGAV | public | — |
| 34 | University of Melbourne | public | — |
| 35 | Australian Museums and Galleries Association Victoria | public | — |
| 36 | Martin Jackson | public | — |
| 37 | Sense & Centsability Pty Ltd | public | — |
| 38 | Nirmidha Sankar Kumara Suriyar | public | — |
| 39 | Live Performance Australia | public | — |
| 40 | Association of Artist Managers | public | — |
| 41 | Australian Broadcasting Corporation | public | — |
| 42 | SBS | public | — |

## Live index

- Inquiry home: https://www.parliament.vic.gov.au/culturalcreativeindustries
- Submissions page: https://www.parliament.vic.gov.au/get-involved/inquiries/inquiry-into-the-cultural-and-creative-industries-in-victoria/submissions
- Confirmed PDF URL patterns (two forms seen): `…/<code>/contentassets/<hash>/submission-documents/NN.-<slug>-redacted.pdf` and `…/<code>/contentassets/<hash>/NN.-<slug>.pdf`.
