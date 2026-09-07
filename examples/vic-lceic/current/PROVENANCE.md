# current/ — provenance

`current/` points at the first complete tree of the report: **all 69 findings from
`../evidence/2026-09-08-full/sonnet-retrieved/`**, one uniform run — Sonnet, retrieved, N=3, on the
final corpus (hearings + submissions + QoN responses), under the manifest and the four verdict rules
(`spec/TREE.md`). `assay -from current.faithfulness.jsonl claims.txt` renders `tree.html`.

Rollup (69): 23 faithful · 26 partial · 13 absent · 2 overstated · 2 contradicted · 1 unsupported ·
2 unverifiable (F38a/b — cite the ABC QoN of 27 Feb 2025, which is not held). Spend $5.42.

Superseded the earlier heterogeneous per-claim merge (`build-current.py`), which stands as the $0
stopgap builder for use between full runs.
