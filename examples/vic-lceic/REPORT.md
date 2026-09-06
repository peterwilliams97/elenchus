# REPORT — LCEIC faithfulness tree run (2026-09-06)

## Result

5-claim faithfulness pass ran clean: **5/5 verified, 0 errored — faithful 4, partial 1**. Tree
renders in 6 lines, complexity OK (depth 3, width 3, score 6.0, target ≤20). Model
`claude-sonnet-4-6`, 10 calls, 1.58M input tokens, wall 3m14s, **$4.82**.

- Tree (stdout): `examples/vic-lceic/tree-2026-09-06.txt`
- Full table: `examples/vic-lceic/eval-5/audit.md`
- Verification chain: `examples/vic-lceic/eval-5/claims-faith-5.faithfulness.jsonl`

Per-claim (from `audit.md`):

| id  | §     | claim | verdict |
|-----|-------|-------|---------|
| F8  | 2.2.1 | COVID-19 took away critical training opportunities | **partial** |
| F12 | 2.2.2 | COVID-19 worsened mental health, esp. children/young people | faithful |
| F17 | 2.2.5 | Ticket prices are a key barrier to engagement | faithful |
| F29 | 3.3.5 | Victoria receives its fair share from Creative Australia | faithful |
| F31 | 3.3.6 | Regional Victoria does NOT receive its fair share | faithful |

## Refuter run (2026-09-06)

**PASS — no mutant came back faithful.** `claims-faith-refuter.txt` = the 5 originals + 3 mutants;
8/8 verified, 0 errored; `claude-sonnet-4-6`, 16 calls, wall 4m16s, **$7.68**. Tree (16 lines,
depth 3, score 7.0, OK): `tree-refuter-2026-09-06.txt`; table: `eval-refuter/audit.md`.

| mutant | construction | expected | got | verdict |
|--------|--------------|----------|-----|---------|
| M1 | F31 negated ("Regional Victoria **receives** its fair share") | contradicted | **contradicted** | ✓ |
| M2 | invented ("ticket prices have fallen since 2019") | absent | **contradicted** | ✓ (stronger: source says prices *stagnant*, not fallen) |
| M3 | F12 mis-attributed to Claire Febey | partial/contradicted | **absent** | ✓ on the gate (not faithful); off the expectation |

Gate = "any mutant faithful → fail." None faithful → pass. M3 landed on **absent** rather than the
expected partial/contradicted: the critic treated the false attribution as integral — Claire Febey
(Creative Victoria, economic testimony) did not say this, so no support exists for the claim *as
stated*. Defensible, but it means a wrong-witness mutant is caught as "no support" rather than
"distorted support"; noted, not changed.

**Verdict drift between runs (nondeterminism).** F12 and F29 were `faithful` in the 5-claim run and
`partial` in the refuter run (same model, same source, same claim text). The critic is not
deterministic; a single verdict is a sample, not a fixed value. This does not touch the refuter gate
(mutants stayed well clear of faithful), but any single faithful/partial verdict here should be read
as one draw. F8/F17/F31 verdicts held across both runs.

**F17 note #3 resolved.** F17's report citation is a dataset + submission, but the transcripts *do*
independently support it — the `-v` defender quotes name Vicky Guglielmo (Yarra CC): "the ticket
price and the promise of what it can deliver … is a really key barrier," plus Ariel Blum's "2.1
million fewer people … 47 per cent citing the ticket prices" and Joshua Lowe (TNA). The critic did
not over-credit. Quotes captured in `f17-verbose.txt`.

**Renderer rule added.** `-tree` now expands the whole tree when leaves ≤ 7 (`smallTree`,
`internal/tree/tree.go`), regardless of the Needs-you gate — so the earlier 5-claim tree would now
show its leaves. Tested (`TestSmallTreeExpandsAll`; `TestRenderDefaultExpansion` reworked to 8
leaves to keep exercising the gate). `./build.sh` green.

## What was built (Option 1 + shell preprocessing, per the plan)

Renderer path, spec'd in `spec/CLI.md`, with tests:

- `internal/tree/tree.go` — text tree keyed to the source's headings; complexity constraints (width
  ≤7 → regroup into ≤7 runs; depth ≤4 → fail; label ≤12 words; score target ≤20). `tree_test.go`:
  6 tests (default expansion, expand-all, width regroup, depth fail, parsePath, natLess).
- `internal/brief/brief_test.go` — 5 tests for the previously-untested brief package (Qualify
  tiers, Selected order, Brief report, overflow, idLess).
- `assay.go` — imported both packages; added `-full` and `-tree`/`-tree=full` (mutually exclusive);
  `present()` always writes the full table to `eval/<stamp>-<model>/audit.md` and routes stdout to
  brief (new default) / full / tree; `parseClaimLine` reads an `<id>\t<path>\t<text>` prefix; wired
  into `runFaithfulness` and `runSubstance`. `./build.sh` green (go test + staticcheck + vet +
  build).

Shell preprocessing (not features — deliberately kept out of the binary):

- `pdftotext -layout` on all 15 hearing PDFs → `sources/hearings/*.txt`, concatenated in date order
  into `sources/hearings-all.txt` (9991 lines, 696K). Gitignored, per `PROVENANCE.md`.
- `to-faith-claims.pl report-summary.md claims-machine.txt > claims-faith.txt` (69 claims) — turns
  the pipe/`route=` format into assay's line format with an `<id>\t<path>\t<text>` prefix; path
  labels (`Chapter N — …`, `§X.Y.Z <title>`) lifted from `report-summary.md` (data, not code).
  `claims-faith-5.txt` = the five run here (F8, F12, F17, F29, F31).

Still NOT built (per plan): PDF extraction inside the binary, directory `-source`, and `route=`
dispatch. The pieces below the "prove it first" line remain for the scale-up.

## Notes (observations, not fixes — verdicts and tree untouched by hand)

1. **The tree collapsed both chapters** to one line each (`[3]`, `[2]`) with no leaves. This is
   correct: default `-tree` opens a branch only when a claim below it is in the brief's Needs-you
   set (contradicted/absent, laundering, overstated+number, grounding-refuted). All 5 verdicts are
   faithful/partial — none qualify — so a clean run collapses. `audit.md` and `-tree=full` show the
   leaves. A clean run is a real result, not zero-output: the claim count and header carry it.
2. **Recall held despite the needle-in-haystack risk.** Each faithfulness call put the whole 696K
   corpus (~150K tokens) in context; support for a claim can sit in one witness's testimony buried
   in 9991 lines. The model still found support (4 faithful), so recall held *here* — but this risk
   bites hardest on *negative* verdicts (a missed quote reads as "absent"), and none of these five
   came back absent, so the risk is unexercised, not disproven.
3. **F17 (ticket prices) verdict = faithful, but its report citation is a dataset + a written
   submission** (Audience Atlas 2024; Musica Viva Submission 23), NOT a hearing transcript. The
   source here was transcripts only, so either the transcripts independently support the claim or
   the critic over-credited. Flagged for a look; not changed.
4. **F29 and F31 both faithful though marked "in tension"** in `claims-machine.txt` (Victoria gets
   its fair share / regional Victoria does not). A transcript can faithfully carry both a claim and
   its regional counterpoint, so this is plausible, but worth an eye.
5. **Heartbeat overstates errors mid-run.** The 60s ticker printed `errored=4` at case 1/5;
   `assay.go:1637` computes `errored := (total - verified) - skipped`, counting unprocessed cases as
   errored. Cosmetic — the final SUMMARY and the chain both show 0 errored — but misleading on long
   runs. Minor bug, left for a decision.
6. **audit.md is the markdown full table** (no-color-text variant from `spec/CLI.md` note 1
   deferred), and it drops the id/path columns the tree carries — `mdFaith` is the existing
   renderer, reused as-is.

## Starting state (why this was more than "finish Step 2")

At HEAD `ccdd62f` the requested command could not run: `internal/brief` existed but was unimported
and untested; no `-tree`/`-full`, no tree package, no always-on `audit.md`; `-source` read one file
via `os.ReadFile` (assay.go); no PDF extraction; and `claims-machine.txt`'s pipe/`route=` format
would have been mangled by `splitSummary`. Option 1 built the spec'd renderer and moved ingestion to
shell, which is what ran above.
