# current/ — provenance

Merged from **two uniform runs of the same corpus** (Sonnet, retrieved, N=3, full corpus + manifest):

- `evidence/2026-09-09-full/sonnet-retrieved/` → `current.faithfulness.jsonl`
- `evidence/2026-09-10-full/sonnet-retrieved/` → `current-2026-09-10.faithfulness.jsonl`

`assay -from current.faithfulness.jsonl,current-2026-09-10.faithfulness.jsonl claims.txt -tree` merges
the two leaf-by-leaf (majority verdict over the pooled samples) and renders `tree.html`, `audit.md`,
and `root-block-tree.txt` with no model calls.

The argument tree (`spec/ARGUMENT.md`) renders with, flags **before** the positional claims file
(Go's flag parser stops at the first non-flag argument):

    assay -from current.faithfulness.jsonl,current-2026-09-10.faithfulness.jsonl \
      -argument ../argument.txt -manifest ../sources/MANIFEST.md -refs ../claims-machine.txt \
      claims.txt

`-manifest` and `-refs` add per-leaf provenance with no corpus and no model call: every verified quote
ends in `— <doc>, <witness>, <locator>` (the source document's canonical id, its witness/author from
the manifest, and the page/transcript line), resolved from the quote's passage id via
`manifest.CanonicalDoc` + the manifest labels; a quote whose passage id has no manifest entry renders
`(passage <id>, unresolved)` and is never dropped. Each claim also carries its report section
(`report: §…`, from `claims-machine.txt`'s `ref=`).

The passage id per quote lives in each chain record's `quote_passages`, added by
`assay -backfill-passages <chain> -source ../sources/hearings,../sources/submissions,../sources/qon` —
a one-off, no-model pass that verbatim-matches each quote against that record's own retrieved passages
and writes the id only when exactly one passage matches, leaving it unresolved on zero or more than one
(never guessing). A live judge run now persists the same field directly (the judge's cited passage id).
Of the 119 quotes the argument tree renders, **113 resolve**; the 6 unresolved each appear verbatim in
two of the record's passages — a hearing turn also quoted in the ABC/SBS response to questions on
notice, or adjacent context-stitched turns — so the backfill declines to pick.

Stability across the two runs, over the **judged** findings (opinions excluded — they are never
judged): **48 settled, 3 wobble, 9 contested** (60 judged + 9 Committee opinions = 69). The 9 contested
findings — where the two runs cross the support divide or leave no majority — form the root block's top
group *Sources don't settle these* and are surfaced first in Needs-you:

F3, F5, F19, F23, F34d, F39c, F49a, F49b, F52 — of which F3, F23, F34d and F49b are ties (no majority),
each rendered `split a/b` rather than a verdict.

Three findings fail the schema gate — the judge's reason came back as a raw tag rather than prose — so
their verdicts are forced to `unverifiable` and they form the *Unverifiable, judge output malformed*
class: **F4, F34a, F34c**.
