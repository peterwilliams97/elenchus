# Prompt 2 — judge-prompt scope change (before vs after, per backend)

faithful now requires the source to state the claim's **subject, scope and direction**; adjacent/broader statements are partial at best (gap=scope), with two worked negatives in the prompt. Hearings + submissions corpus, retrieved, N=3. Scoring set excludes F8/F31. Mutants (M1/M2/M3) gold = contradicted/absent; the 5 new claims (F2a/F4/F10/F33/F54b) have no gold — see adjudication.md.

| claim | sonnet before | sonnet after | qwen before | qwen after |
|---|---|---|---|---|
| F12 | partial 3/3 | partial 3/3 | faithful 3/3 | faithful 3/3 |
| F17 | faithful 3/3 | faithful 3/3 | faithful 3/3 | faithful 3/3 |
| F29 | partial 3/3 | partial 3/3 | partial 2/3 | faithful 2/3 |
| M1 *(mutant)* | contradicted 3/3 | contradicted 3/3 | contradicted 3/3 | contradicted 2/3 |
| M2 *(mutant)* | contradicted 3/3 | contradicted 3/3 | contradicted 3/3 | contradicted 3/3 |
| M3 *(mutant)* | absent 3/3 | absent 3/3 | absent 3/3 | absent 3/3 |
| F2a *(new)* | faithful 3/3 | faithful 3/3 | faithful 3/3 | faithful 3/3 |
| F4 *(new)* | faithful 3/3 | faithful 3/3 | faithful 3/3 | faithful 3/3 |
| F10 *(new)* | absent 3/3 | absent 3/3 | absent 3/3 | absent 2/3 |
| F33 *(new)* | absent 3/3 | absent 3/3 | absent 3/3 | absent 3/3 |
| F54b *(new)* | absent 3/3 | absent 3/3 | absent 3/3 | absent 3/3 |

## Per-cell metrics

| metric | before-sonnet | after-sonnet | before-qwen | after-qwen |
|---|---|---|---|---|
| mutants caught (of 3) | 3/3 | 3/3 | 3/3 | 3/3 |
| no_quote_downgrades | 0 | 0 | 1 | 1 |
| quote_rejects | 4 | 6 | 19 | 14 |
