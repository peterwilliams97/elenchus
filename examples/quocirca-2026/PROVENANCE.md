# Provenance — quocirca-2026 faithfulness assay

Records the report under analysis and, for the evidence chains, which pair is canonical and which is
superseded. The fuller run narrative is `REPORT.md`; this file is the short "what is authoritative"
record so a later reader does not rebuild the site from the wrong chains.

## Report under analysis (both TARGET and only held SOURCE)

- Quocirca, *Print Security Landscape 2026 — How AI, Identity, and Quantum are reshaping the threat
  landscape*, RICOH excerpt, July 2026 (15 pp.).
- `sources/report.pdf` — the read copy; `sources/report.txt` is its pdftotext (gitignored under
  `sources/*`: analysis, not redistribution).
- Single-source corpus (`sources/MANIFEST.md` declares `single_source: true`): the report is its own
  only source, so the automated check is internal-consistency only. Survey figures' truth-maker (the
  study microdata) is not held — those claims are `unverifiable` for grounding by design.

## Evidence chains

Every chain is a full 60-leaf faithfulness run over `claims-machine-full.txt`, N=3, Sonnet
`claude-sonnet-4-6`, bm25+embed retrieval.

### Canonical pair (current judge) — build the site from these

- `evidence/2026-09-12-a-para/claims-machine-full.faithfulness.jsonl`
- `evidence/2026-09-12-b-para/claims-machine-full.faithfulness.jsonl`

The current-judge chains: they carry the paragraph-granularity single-source self-exclusion
(`dropOwnParagraph`, `assay.go`) that drops only the claim's own paragraph, not its whole page, so a
same-§ qualifier survives to ground the claim (`spec/TREE.md` § Single-source judging has a
direction; refuter `TestDropOwnParagraph`). Produced as a resume of the earlier chains with the 8
flagged leaves re-judged; the change moved K22 → `faithful` and K37 → `uncorroborated` (REPORT.md
follow-up 3). Backfilled 2026-09-12 (`-backfill-passages`, no model call): a-para 29/30, b-para 33/34
quotes resolved to report pages; unresolved quotes are left blank, never guessed. `site/` is built
from this pair (adjudications: 8 leaves, judge agreed on 2).

### Pre-fix pair (retired) — kept for the record, not a site source

- `evidence/2026-09-12-a/claims-machine-full.faithfulness.jsonl`
- `evidence/2026-09-12-b/claims-machine-full.faithfulness.jsonl`

The original full judge run (352 calls each, $2.6950 + $2.6660, ~36 min each, 0 errors). These use
the earlier page-granularity self-exclusion (`dropOwnPage`), which hid same-page qualifiers and so
mis-flagged K22 and K37. **Superseded by the `-para` pair above.** Do not rebuild the site from these.
