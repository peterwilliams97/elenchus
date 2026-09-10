# current/ — provenance

Merged from **two uniform runs of the same corpus** (Sonnet, retrieved, N=3, full corpus + manifest):

- `evidence/2026-09-09-full/sonnet-retrieved/` → `current.faithfulness.jsonl`
- `evidence/2026-09-10-full/sonnet-retrieved/` → `current-2026-09-10.faithfulness.jsonl`

`assay -from current.faithfulness.jsonl,current-2026-09-10.faithfulness.jsonl claims.txt -tree` merges
the two leaf-by-leaf (majority verdict over the pooled samples) and renders `tree.html`, `audit.md`,
and `root-block-tree.txt` with no model calls.

Stability across the two runs, over the **judged** findings (opinions excluded — they are never
judged): **48 settled, 3 wobble, 9 contested** (60 judged + 9 Committee opinions = 69). The 9 contested
findings — where the two runs cross the support divide or leave no majority — form the root block's top
group *Sources don't settle these* and are surfaced first in Needs-you:

F3, F5, F19, F23, F34d, F39c, F49a, F49b, F52 — of which F3, F23, F34d and F49b are ties (no majority),
each rendered `split a/b` rather than a verdict.

Three findings fail the schema gate — the judge's reason came back as a raw tag rather than prose — so
their verdicts are forced to `unverifiable` and they form the *Unverifiable, judge output malformed*
class: **F4, F34a, F34c**.
