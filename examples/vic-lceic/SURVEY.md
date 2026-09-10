# SURVEY — read-only state of elenchus / vic-lceic (2026-09-08)

Read-only. No code changed, no runs made, no key touched. Evidence for every claim below was read in
this session (file paths + tool output). Line-count and verdict tallies are reproducible with the
commands quoted.

## 1. Git

- Branch: `refresh/routes-prep-2026-09-07`
- HEAD: `aad89f9c5d7f740cece2d8712c962ea1e92407b7`
- Uncommitted: `M examples/vic-lceic/evidence/2026-09-07-compare/adjudication.md`, `M examples/vic-lceic/routes.md`
- Untracked: `LICENSE`, `diff.txt`, `docs/faithfulness_prompt.txt`, `docs/ollama.md`, `eval-value.stderr`,
  `eval-value.tree`, `examples/vic-lceic/current/root-block-tree.txt`
- Last 10 commits:
  ```
  aad89f9 vic-lceic: model-guess route run (Sonnet, 29/29, $ actuals in jsonl)
  820973c CLAUDE.md: key rule now permits in-session `source ./setkey.sh && <command>`
  e784b65 vic-lceic: add model-guess route pass for the regex-undecided Findings
  86b8275 vic-lceic: correct F44/F45 section to §4.4.2 (body is authority)
  c89806a vic-lceic: add routes-prep.md — regex-only route pre-pass over the 55 Findings
  973ddfe vic-lceic/eval-f46b: re-rerun under the narrowed per-field rule (plain_retries 1, $0.07)
  f999eb2 judge: narrow the no-copy rule per field; source_says may not contrast or negate
  dfa078e vic-lceic/eval-f46b: re-rerun under the no-copy rule (plain_retries 1, $0.07)
  670f7bb CLAUDE.md: pin the API-key handling rule — key never in-session, human runs via setkey.sh
  29dacf5 judge: reject report_says/source_says that copy a 3+-word run from the claim or a quote
  ```

## 2. Files, and what CLAUDE.md doesn't mention

CLAUDE.md still describes a single-file `assay.go` ("Everything lives in assay.go"). That is **stale**.
The repo now carries a full `internal/` tree the CLAUDE.md Architecture section never names:

```
internal/backend/backend.go          backend interface + shared req/resp
internal/backend/anthropic/          Messages API wrapper (+ _test.go)
internal/backend/ollama/             /api/chat wrapper (+ _test.go)
internal/backend/fake/               test double (no _test.go)
internal/brief/                      Qualify (needs-you tiers) + OpenedBranches (+ _test.go)
internal/embed/                      nomic-embed ranker (no _test.go)
internal/manifest/                   MANIFEST scan / canonical ids (+ _test.go)
internal/retrieve/                   BM25+embed RRF passage retrieval (+ rank.go, _test.go)
internal/tree/                       Render + RootBlock (+ root_test.go, tree_test.go)
```

Also new since the CLAUDE.md text: `spec/` (CLI.md, TREE.md), `docs/VALUE.md`, `judge_test.go`,
`refuter_b_test.go`, `oracle_gen_test.go`, the whole `examples/vic-lceic/` corpus workflow, and
`-backend`/`-retrieve`/`-manifest`/`-tree`/`-from`/`-oracle` flags. **CLAUDE.md's "What this is" and
"Architecture" sections predate the elenchus2/crossexam refactor and should be rewritten.** `spec/CLI.md`
and `spec/TREE.md` are the current authority (both carry "assay-era spec, to be ported to elenchus2"
scope notes).

## 3. Spec → code (BUILT / PARTIAL / MISSING)

### spec/TREE.md

| requirement | status | symbol | refuter |
|---|---|---|---|
| group claims by `key=Label` path into a tree | BUILT | `internal/tree/tree.go` `Render`, `ParsePath` | `TestParsePath`, `TestRenderDefaultExpansion` |
| width ≤7 (regroup), depth ≤4 (fail), label ≤12w, score target | BUILT | `tree.go` `regroupWide`, `deepestLeaf`, `maxWidth/maxDepth/scoreWarn` | `TestWidthRegroup`, `TestDepthFail`, `TestRenderComplexityHeader` |
| default expansion = only Needs-you branches; `-tree=full` all | BUILT | `tree.go` | `TestRenderDefaultExpansion`, `TestSmallTreeExpandsAll`, `TestRenderExpandAll` |
| Needs-you tier rule `Qualify` (tiers 0–5) | BUILT | `internal/brief/brief.go` `Qualify` | `TestQualifyTiers`, `TestSelectedOrder` |
| `OpenedBranches` = value line "changes N summary lines" | BUILT | `brief.go` `OpenedBranches` | `TestOpenedBranches` |
| root summary block, class partition, opinion resolved first | BUILT | `internal/tree/root.go` `RootBlock` | `TestRootBlock` + `TestRootBlockRouteGates` (route-gates-partition mutation) |
| `route` gates opinion class; read from chain, fall back to claims | BUILT | `root.go` | `TestRootBlockRouteGates` |
| leaf verdict + N=3 spread + verified quotes + so_what | BUILT | `tree.go` `RenderHTML`, `SpreadInLeaf`, `SoWhatOnNeedsLeaf` | `TestSpreadInLeaf`, `TestSoWhatOnNeedsLeaf`, `TestRenderHTML` |
| 4 code-side verdict rules (quote-by-ref, need-quote, cites→unverifiable, floor) | BUILT | see CLI.md rows below | see below |
| `-from` re-render, no model call; resumable run | BUILT | `assay.go` `runFromChain`, `readChain` | `TestReadChainOrdersAndValidates`, `TestReadChainSparseToleratesGaps`, `TestFromReplayBackendAgnostic` |
| `current/` PROVENANCE (one line per claim) | PARTIAL | `current/PROVENANCE.md` is prose+rollup, **not** one line per claim as TREE.md §current/ specifies | — |

### spec/CLI.md

| requirement | status | symbol | refuter |
|---|---|---|---|
| backend interface, anthropic + ollama + fake | BUILT | `internal/backend/*` | `TestDispatchThroughBackend`, `TestChainStampsBackend`, backend pkg tests |
| ollama: no web search, think:false default, length→error | BUILT | `internal/backend/ollama` | `TestSupportsWebSearchFalse`, `TestThinkDefaultFalse`, `TestLengthTruncationErrors` |
| anthropic: retries, max_tokens truncation→error, cache blocks, strict tool | BUILT | `internal/backend/anthropic` | `TestComplete429ThenSuccess`, `TestCompleteMaxTokensTruncation`, `TestCompleteCacheControl`, `TestSchemaForcesStrictTool` |
| retrieval: BM25+embed RRF best-rank, Q-stitch, roles, hard filter, floor | BUILT | `internal/retrieve` | `TestRetrieveDeterministicAndCapped`, `TestRetrieveWitnessHardFilter`, `TestRetrieveFloorAbsent`, `TestRolesFromRealCorpus` |
| retrieval refuter vs Sonnet full-corpus spans (15/18) | BUILT | `internal/retrieve` | `TestRetrievalRefuterFusedCoversSonnetQuotes` |
| single schema-enforced faithfulness judge, retry-once | BUILT | `assay.go` faith judge path | `TestJudgeSchemaValid`, `TestSchemaRetryCountedOnce` |
| scope discipline (`faithful` needs subject/scope/direction) | BUILT | judge prompt | `TestFaithJudgeSysScopeDiscipline` |
| quote-by-reference verbatim check (rule 1) | BUILT | `quoteInPassage` | `TestQuoteInPassageVerbatimVsParaphrase`, `TestFaithJudgeRejectsNonVerbatimQuotes` |
| verdict needs a verified quote (rule 2) | BUILT | `groundVerdict` | `TestGroundVerdictDowngrade` |
| cites→unverifiable, no model call (rule 3) | BUILT | `internal/manifest` + judge path | `TestMissingCites`, `TestScanDerivesCanonicalIDs`, `TestLoadQoN`, `TestLoadSubmissions` |
| retrieval floor→absent (rule 4) | BUILT | `retrieve` | `TestRetrieveFloorAbsent` |
| no-copy / restatement judge rules (HEAD-era) | BUILT | judge prompt | `TestFaithJudgePlainRestatementRetry`, `TestFaithJudgeVerbatimCopyRetry`, `TestReportSaysBad`, `TestSourceSaysBad`, `TestContrastWord`, `TestVerbatimRun` |
| evidence grounding cross-check (URL provenance) | BUILT | `crossCheckEvidence`, `normalizeURL` | `TestCrossCheckEvidenceClaimedAbsent/Present/ZeroRetrieved/ZeroSources/Exemptions` |
| brief report layout + selection | BUILT | `brief.go` | `TestBriefReport`, `TestBriefOverflow`, `TestRenderStakes` |
| `-retrieve=oracle` + `-oracle` flag | BUILT (code) / **MISSING from spec** | `assay.go:151-152`, `oracle_gen_test.go` | `TestEmitOracle`, `TestByIDsOracle` — but CLI.md flag table lists only `bm25\|none` (0 hits for "oracle") |

**Checks with no dedicated refuter:** `internal/backend/fake` and `internal/embed` have no `_test.go`
(embed's ranker is exercised only indirectly through `internal/retrieve`). `current/PROVENANCE.md`'s
one-line-per-claim format (TREE.md §current/) has no generator or test — it is hand-written prose.

## 4. Tests

`go test ./...` → **all pass** (cached). 7 packages `ok`; 3 report `[no test files]`
(`internal/backend`, `internal/backend/fake`, `internal/embed`). 116 `Test*`/`Fuzz*` functions across
11 test files. No `go vet` complaints.

## 5. Open items — routes, adjudication, judge-prompt reconciliation

### routes
- `claims-machine.txt`: **all 69** findings declare `route=` (0 lack it). Propagated into
  `claims-machine-full.txt` as the 5th tab column.
- `routes.md` (human review sheet): 69 finding rows, **human "witness said / committee concluded /
  fact about world" column filled on 0 of them** — entirely blank, awaiting a person (the header says
  "do not guess it").
- `routes-model.jsonl` / `routes-model.md`: the **29** regex-undecided findings, machine-guessed
  (Sonnet, commit `aad89f9`). This is the "29/29" of the last commit, not the full 69.

### adjudication.md (`evidence/2026-09-07-compare/`)
- 3 divergence rows (F8, F31, M3). **"who's right" column blank on all 3.** Footer: "1 resolved by
  the grounding rule [M3 → both absent], 2 remain for a human" (F8 partial-vs-faithful, F31
  partial-vs-faithful).

### Judge-prompt reconciliation — the disagreement to flag
- **The "first complete tree" (`evidence/2026-09-08-full/`, mirrored to `current/`) was judged under
  the PRE-CHANGE judge prompt.** Its `config.json` records `git_sha 20eebe5` (2026-09-07 16:32).
  Four judge-prompt commits land *after* that SHA (`37a8495`, `e1a449e`, `29dacf5`, `f999eb2` — the
  so_what→`report_says`/`source_says` rework and the per-field no-copy rule). `20eebe5` is a verified
  ancestor of all four (`git merge-base --is-ancestor`).
- Evidence in the chains: the full-run and `current/` records carry detail keys `so_what` + `source_says`
  but **no `report_says`** (old schema). The one leaf re-run under HEAD's judge, `eval-f46b/`, carries
  `report_says` **and** `source_says` (new schema).
- **So all 69 tree verdicts are pre-change; only F46b has a current-prompt result, and it disagrees:**
  - `current/` + `2026-09-08-full`: **F46b = contradicted** (a Tier-0 red leaf that reaches the summary).
  - `eval-f46b/` under the HEAD judge: **F46b = partial**.
  - That flip is not merged into `current/`. The tree on disk therefore shows a `contradicted` the
    current judge no longer returns.
- Verdict rollup **is internally consistent** between STATUS.md, `current/PROVENANCE.md`, the
  `2026-09-08-full` chain, and `current/current.faithfulness.jsonl`: all four say 69 = 23 faithful ·
  26 partial · 13 absent · 2 overstated · 2 contradicted · 1 unsupported · 2 unverifiable, $5.42.
- Minor schema drift: `current/current.faithfulness.jsonl` records carry `route` + `passages` keys;
  the `2026-09-08-full` chain records do **not** (re-serialized by a newer build; verdicts identical).

### STATUS.md vs disk — flagged disagreements
1. STATUS.md dates the baseline "2026-09-07" and calls `2026-09-08-full` the "first complete tree,"
   but that run's judge predates the current (HEAD) judge prompt by four commits. STATUS does not say
   the tree is stale w.r.t. the judge.
2. STATUS.md's date header (2026-09-07) vs the run directory / config date (2026-09-08) are
   inconsistent labels for the same run.
3. `current/PROVENANCE.md` claims "points at … `2026-09-08-full/sonnet-retrieved/`" but the two chain
   files differ in schema (route/passages present only in `current/`).

## 6. What a human must decide before the next run

1. **Re-run the full 69 under the HEAD judge, or accept the pre-change tree?** F46b already flips
   contradicted→partial; other leaves may move under the report_says/no-copy rules. The current tree
   is not a HEAD-judge artifact.
2. **Fill `routes.md`** (0/69 human-verified) — or bless `routes-model.jsonl`'s 29 machine guesses —
   before any route-gated partition (opinion class, Needs-you) is trusted. Where a filled route
   disagrees with `claims-machine.txt`, fix the claims file first.
3. **Adjudicate the 2 open divergences** (F8, F31 in `adjudication.md`) — "who's right" is blank.
4. **`-oracle` / `-retrieve=oracle`**: document in `spec/CLI.md` (implemented, untested against spec)
   or remove.
5. **Rewrite CLAUDE.md's "What this is" / "Architecture"** — they describe the retired single-file
   layout, not the `internal/` tree that now exists.
6. **`current/PROVENANCE.md`**: bring to TREE.md's one-line-per-claim format, or amend TREE.md to
   match the prose form actually in use.
7. Housekeeping: several untracked scratch files at repo root (`diff.txt`, `eval-value.stderr`,
   `eval-value.tree`) and untracked `LICENSE`, `docs/ollama.md`, `docs/faithfulness_prompt.txt` —
   decide track vs ignore.
