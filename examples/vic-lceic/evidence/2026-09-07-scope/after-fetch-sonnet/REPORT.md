# Rerun of F10/F33/F54b after fetching the cited documents

Sonnet · retrieved · N=3 · hearings + submissions + **QoN responses** corpus, with
`-manifest sources/MANIFEST.md`. Spend: **$0.11** (6 calls — F10 skipped as unverifiable, no model
call; F33 and F54b judged ×3). This closes the loop that began with the three "absent" verdicts of
2026-09-07-scope: each was reclassified `unverifiable` by the manifest check, the two fetchable
documents were fetched, and the claims were re-judged against them.

| claim | 1st run | manifest check | after fetch | why |
|---|---|---|---|---|
| **F10** | absent | unverifiable (`submission:19/attachment-1`) | **faithful 3/3** *(after a 2nd fetch)* | The attachment is not on the parliament pages, but the same document is published on TNA's own site (tna.org.au). Fetched it, and F10 verifies against it — "Between 2017–18 and 2021–22 attendance at performing arts events by children dropped …" (quote source: submission attachment). See `../f10-after-attachment/`. |
| **F33** | absent | unverifiable (`qon:abc/2025-03-21`) | **absent 3/3** | The ABC QoN response is now in the corpus and the judge read it, but it carries headcount/FTE tables, not the "no state or territory reflects its population" assertion — so `absent` is now the *right* verdict: cited source present, claim not found in it. |
| **F54b** | absent | unverifiable (`qon:sbs/2025-04-10`) | **faithful 3/3** | The SBS QoN response states it verbatim — "SBS's network reaches 3.42 million people every month in Victoria … 49% of the Victorian population, according to the OzTAM VOZ database." Verified quote source: **qon**. |

## What this shows

The manifest + cites + `unverifiable` machinery does exactly its job: it separated "we cannot check
because we lack the document" from "we checked and it is not there." Fetching the two available
documents resolved two of the three — one to `faithful` (the QoN confirms the claim), one to a
*genuine* `absent` (the document is present and simply does not make the claim) — and left the one
whose document is unavailable as `unverifiable`, with a standing so-what of "fetch
submission:19/attachment-1".
