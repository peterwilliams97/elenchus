# Step-3 2×2 comparison — faithfulness judge across model × retrieval

Cells present: sonnet-full, sonnet-retrieved, qwen-retrieved, qwen-oracle. Same 8 claims (`claims-faith-refuter.txt`), N=3, one `config.json` frozen per cell. Numbers below trace to each cell's chain / usage / run.stderr; nothing is synthetic.

## Agreement table (verdict · N=3 spread)

| claim | sonnet-full | sonnet-retrieved | qwen-retrieved | qwen-oracle |
|---|---|---|---|---|
| F8 | partial 3/3 | partial 3/3 | faithful 3/3 | faithful 3/3 |
| F12 | faithful 2/3 | faithful 3/3 | faithful 3/3 | faithful 3/3 |
| F17 | faithful 3/3 | faithful 3/3 | faithful 3/3 | faithful 3/3 |
| F29 | partial 3/3 | partial 3/3 | partial 2/3 | faithful 2/3 |
| F31 | partial 3/3 | partial 3/3 | faithful 3/3 | faithful 3/3 |
| M1 | contradicted 3/3 | contradicted 3/3 | contradicted 3/3 | contradicted 3/3 |
| M2 | contradicted 3/3 | contradicted 3/3 | contradicted 3/3 | contradicted 3/3 |
| M3 | absent 3/3 | absent 3/3 | contradicted 2/3 | absent 2/3 |

## Per-cell metrics

| metric | sonnet-full | sonnet-retrieved | qwen-retrieved | qwen-oracle |
|---|---|---|---|---|
| mutant catch rate (of 3) | 3/3 | 3/3 | 3/3 | 3/3 |
| secs / claim | 42.8s | 36.8s | 116.6s | 64.9s |
| $ total (24 calls) | $1.8517 | $0.5255 | $0 | $0 |
| $ / claim | $0.2315 | $0.0657 | $0 | $0 |
| quote-verification rejects | 15 | 19 | 34 | 16 |
| cache hit | 96% | 89% | 0% | 0% |
| schema retries | 0 | 0 | 0 | 0 |
