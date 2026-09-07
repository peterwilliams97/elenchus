# Adjudication — 5 new scope-risk claims (no gold yet)

Verdicts and verified quotes from the **after** (scope-changed) judge on the combined corpus. These claims have no gold; the **who's right** column is blank for a human.

> **Superseded in part (2026-09-07, appended not rewritten).** The `absent` rows for **F10, F33, F54b**
> below were `absent` *because the corpus lacked the document the report cites* — later reclassified
> `unverifiable`, then the documents were fetched and the claims re-judged. See the Resolution section
> at the foot of this file and `../2026-09-07-scope/after-fetch-sonnet/`, `.../f10-after-attachment/`,
> and the full tree `../2026-09-08-full/`. **F2a and F4 are unchanged and still open** for a human —
> F4 in particular (the source hedges "all jurisdictions *where this research was conducted*" vs the
> claim naming New York, Sweden and the UK: a candidate `partial`/`gap=scope`).

| claim | backend | verdict | verified quotes | who's right |
|---|---|---|---|---|
| F2a | sonnet | faithful 3/3 | More than 320,000 Victorians work in the creative economy. That is almost 9 per cent of to |  |
| F2a | qwen | faithful 3/3 | More than 320,000 Victorians work in the creative economy. |  |
| F4 | sonnet | faithful 3/3 | Victoria rated higher for cultural participation than all other jurisdictions where this r |  |
| F4 | qwen | faithful 3/3 | Victoria rated higher for cultural participation than all other jurisdictions where this r |  |
| F10 | sonnet | absent 3/3 | (none) |  |
| F10 | qwen | absent 2/3 | (none) |  |
| F33 | sonnet | absent 3/3 | (none) |  |
| F33 | qwen | absent 3/3 | (none) |  |
| F54b | sonnet | absent 3/3 | (none) |  |
| F54b | qwen | absent 3/3 | (none) |  |

## Resolution (after fetching the cited documents)

The three `absent` verdicts above traced to documents the corpus did not hold; the manifest check
reclassified them `unverifiable` ("fetch X"), the fetchable documents were fetched, and the claims
were re-judged. Verdicts in the first complete tree (`../2026-09-08-full/`):

| claim | scope run (above) | after fetching the cited doc | what settled it |
|---|---|---|---|
| F10 | absent | **faithful** | TNA Submission 19 Attachment 1 (fetched from tna.org.au): "Between 2017–18 and 2021–22 attendance at performing arts events by children dropped …" |
| F33 | absent | **absent** (now *verified*) | ABC QoN 21 Mar 2025 is held and was read; it carries headcount/FTE tables, not the "no state reflects population" claim — a genuine absent, not a missing-document absent. |
| F54b | absent | **faithful** | SBS QoN 10 Apr 2025 (fetched): "reaches 3.42 million people every month in Victoria … 49% of the Victorian population … OzTAM VOZ". |

F2a and F4 remain as judged above (`faithful`), their **who's right** still for a human.
