# master-plan judge runs — 2026-09-13 (COMPLETE)

## Result

Two Sonnet N=3 bm25 runs over the 124-leaf master-plan corpus, retrieving over `report.md` + the four
held companions, merged leaf-by-leaf.

- **Rollup (124 leaves, merged modal):** 75 faithful · 39 unsupported · 7 unverifiable · 3 partial.
- **Stability (across the 2 runs):** 110 settled · 2 wobble · 12 contested.
- **Argument tree:** of 12 recommendations — **1 holds** (REC10), **3 open** (REC2/F3, REC8/F9,
  REC12/F15; open because their deciding finding is contested), **8 fail** (REC1/F1, REC3/F4, REC4/F5,
  REC5/F6, REC6/F7, REC7/F8, REC9/F10, REC11/F14 — each "no held source supports" the finding). 5
  findings support no recommendation.
- **Cost:** $19.80 total — original runs A $6.63 + B $6.44, resume A $3.38 + B $3.35. (Two resume
  attempts hit `credit balance too low` and errored fast; not billed.)

### Root thesis (verbatim, authored in `argument.txt` — not a judge output)

> Stop reviewing the code; review the test instead, and let the test review the code — because the
> model that writes the code will be replaced repeatedly and the only thing fixed across model
> versions is the test, so trust has to live in the test. The goal is to survive the transition to
> AI-generated code.

## Provenance

| Item | Value |
|---|---|
| chains | `evidence/2026-09-13-a/…faithfulness.jsonl` (124 rec), `evidence/2026-09-13-b/…` (124 rec) |
| model | `claude-sonnet-4-6`, backend anthropic, `-retrieve bm25 -n 3` |
| corpus | `report.md` + `refuter-runs-2026-08-15.md`, `flt-kloc-review.md`, `software-correctness.md`, `vuln-finder-hive.md` |
| backfill | passage ids resolved A 93/93, B 91/91, 0 unresolved |
| site | `site/index.html` (argument tree) + `site/review.html`; serve: `../../assay serve site` |

## Integrity caveats (do not affect verdict validity, but are divergences from the documented run)

1. **Derived `.txt` corpus.** The tracked sources are flat `.md`; `internal/retrieve` yields 0
   passages for them (only `report.txt`/`report-*.txt` and files under `submissions/`·`qon/` get
   paragraph-split, all else goes to the Hansard speaker-turn parser). Runs used a byte-identical
   `.txt` copy in the session scratchpad — `report.txt` + companions under `submissions/`. Verdicts
   are against the real doc text; manifest/`cites` matching keys on `.md` ids independently. **Follow-up
   for a maintainer: teach `retrieve` to accept `.md`/generic prose, else this example stays
   unrunnable as documented.**
2. **Sandbox off** for the API (in-sandbox `api.anthropic.com` fails TLS, `x509: OSStatus -26276`).
3. **7 leaves = "unverifiable (judge output malformed)"** — the grounding check (`groundVerdict`)
   rejected the judge's output at render and reclassified them, so their stored raw verdict is not
   trusted. Read alongside the 12 contested as the low-confidence tail.
4. `report.md` is markdown → no report PDF, so `site/review.html`'s left pane is blank (`pdfs
   copied=0 missing=1`), exactly as `sources/MANIFEST.md` states. `site/index.html` is self-contained.
