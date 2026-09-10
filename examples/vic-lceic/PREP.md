# PREP — pre-flight for the HEAD-judge re-run (2026-09-08)

Read-only survey (item 3 writes only a command string here; nothing run, committed, stashed, or
switched). Evidence read this session; helper semantics quoted from `assay.go` and `judge_test.go`.

## 1. Branch position and the two uncommitted diffs

The baseline ref is the tag **`v1-assay`** (`df22801`); there is no bare `assay` ref (`git rev-parse
assay` → *Needed a single revision*). Against it:

- `git log --oneline v1-assay..HEAD` → **43 commits ahead**, HEAD `aad89f9`. Range spans the whole
  faithfulness/retrieval/tree build (from `4d47837` prompt-caching through the judge-prompt reworks
  `37a8495`→`f999eb2` to the route sheets `4563b39`/`aad89f9`).
- `git log --oneline HEAD..v1-assay` → **0 commits** (HEAD strictly ahead; `v1-assay` is an ancestor).

Both working-tree diffs are **whitespace-only markdown table re-alignment**, no content change:

- `examples/vic-lceic/routes.md` — the `| id |` column widened to `| id  |` and the header rule
  padded (`|---|` → `|-----|`); every row's finding/route/blank cells are byte-identical. Human
  review column still empty on all 69.
- `examples/vic-lceic/evidence/2026-09-07-compare/adjudication.md` — same: the 3 divergence rows
  (F8, F31, M3) re-padded for column alignment, plus loss of the trailing newline. No verdict, quote,
  or "who's right" cell changed.

(Full unified diffs are in this session's tool log; both are cosmetic and safe to commit or discard.)

## 2. Route precedence on `-from` re-render — chain wins; proposed override

**Current behaviour: the chain wins.** `assay.go:2202` `routeOr(chainRoute, claimsRoute)` returns
`chainRoute` whenever it is non-empty, else the claims-file route:

```go
func routeOr(chainRoute, claimsRoute string) string {
    if strings.TrimSpace(chainRoute) != "" { return chainRoute }
    return claimsRoute
}
```

Every `-from` render call site passes the **chain** record's route first: `runFromChain`
(`assay.go:2625, 2642, 2657, 2684`) builds each row with `Route: routeOr(r.Route, routes[i])`, where
`r.Route` is the chain record and `routes[i]` is parsed from the claims file (`:2594`). So once a
chain record carries a `route`, **editing `claims-machine.txt` and re-rendering with `-from` does not
change the leaf's route** — the fallback only fires for pre-`route` chains (the very case the doc
comment describes). This is why `current/`'s route edits would need a live re-judge, not a free
`-from`, to take effect.

### Smallest change: a `-route-from-claims` flag, `-from` only

`-from` already **requires and validates** the claims file (it aborts on any claim-text mismatch,
`:2601`), so the claims file is a trusted, editable source of truth on that path. Add one bool flag
that inverts the precedence for the render path only — no change to live-run behaviour, no schema
change:

- Declare `flag.BoolVar(&c.routeFromClaims, "route-from-claims", false, "on -from, let an edited
  claims-file route override the chain's")` next to the other flags (`assay.go:~176`).
- At the four `runFromChain` call sites, choose the argument order by the flag:
  ```go
  route := routeOr(r.Route, routes[i])
  if c.routeFromClaims { route = routeOr(routes[i], r.Route) }
  ```
  (or add `routePref(chain, claims string, preferClaims bool)` and call it in all four places).

This keeps `routeOr`'s existing default (chain wins, back-compat for old chains) and makes the
override explicit and opt-in, so no existing re-render changes silently.

**Refuter** — `TestFromRouteFromClaimsOverrides` (mirrors `TestRootBlockRouteGates`, driven through
the precedence): build a one-record chain whose route is `evaluative` and a claims-file row for the
same claim id whose route is `source`, with the record's verdict `absent`. Render twice.
- Default (`routeFromClaims=false`): leaf route resolves `evaluative` → root block puts it in
  *Committee opinions, not checked* (opinion resolved before verdict).
- With `-route-from-claims`: leaf route resolves `source` → the `absent` verdict now lands the leaf
  in *Unsupported by any held source*, moving that count by one.
The verdict/route pair is chosen so the two precedences produce **different root-block partitions**,
so the test fails if precedence is not actually honoured (not merely if a string field differs).

**Spec text for `spec/TREE.md` § The `route` field** (append after the existing "read from the chain;
a record without one falls back to the claims file" sentence):

> On a live judge run the route is taken from the claims file and written into the chain, so the
> chain is authoritative for every later `-from` render — an edited claims-file route does **not**
> reach a saved chain's leaves for free. To re-route a saved run without re-judging, pass
> `-route-from-claims`: on the `-from` path only, the claims-file route (which `-from` already
> validates against the chain) overrides the chain's, falling back to the chain when the claims file
> leaves a row's route blank. Live runs are unaffected; the default (chain wins) preserves rendering
> of pre-`route` chains.

*(Do not implement — proposal only.)*

## 3. The HEAD-judge full re-run command and expected cost

Per `spec/TREE.md` § Producing the tree, from the repo root, under the HEAD judge (`f999eb2`, the
per-field no-copy rule). The key is supplied by the human via `setkey.sh`:

```sh
source ./setkey.sh && ./assay \
  -backend anthropic -model claude-sonnet-4-6 -retrieve bm25 -n 3 -embed \
  -manifest examples/vic-lceic/sources/MANIFEST.md \
  -source examples/vic-lceic/sources/hearings-all.txt \
  -chain-dir examples/vic-lceic/evidence/2026-09-09-full/sonnet-retrieved \
  -usage-out examples/vic-lceic/evidence/2026-09-09-full/sonnet-retrieved/usage.jsonl \
  -tree examples/vic-lceic/claims-machine-full.txt
```

Notes before running (a human decides these):
- **`-source`** above names the single stitched corpus file (`sources/hearings-all.txt`). The
  2026-09-08 run's `config.json` records `corpus: "hearings+submissions+qon"` — confirm whether the
  final corpus is that one file or a comma-list of hearings,submissions,qon dirs, and match it, or
  the manifest→cites→`unverifiable` accounting will differ. **This is the one input to verify before
  running** (Hard-rule provenance: the corpus is the substrate).
- `-fresh` is **not** passed, so if `evidence/2026-09-09-full/…` already holds a partial chain the
  run resumes it; on a clean dir it judges all 69. Add `-fresh` to force a full re-judge.
- `sources/` is gitignored; the run reads the committed `.embcache` offline for the embed ranker.

**Expected cost ≈ $5.4 (bracket $5.2–$5.8).** Basis — the last full run's own
`usage.jsonl` actuals (`evidence/2026-09-08-full/sonnet-retrieved/usage.jsonl`, same model, N=3,
same 69 claims, same retrieval):

```
calls 201 · in 10,222 · out 76,965 · cache_read 7,821,328 · cache_create 502,787 · web 0
wall 2,204s (~37 min) · est $5.4170 (rates 2026-06-01)
```

The HEAD judge adds only **plain retries** (the no-copy/contrast steer, one extra call per tripped
field). eval-f46b logged `plain_retries 1` at **$0.07** for a single leaf. Item 4 estimates ~25
leaves carry a source_says that trips on the first attempt; at ≤1 retry each that is ≲ 25 extra
judge calls (~+12% calls, but retries re-hit the hot cache prefix, so ≲ +$0.3–0.4). So **budget
$5.4, cap ~$5.8.** Cost is dominated by cache-read tokens, unchanged by the judge edit.

## 4. Leaves most likely to move under the report_says / no-copy rules — **25 of 69**

Method and its limit. The new rules bite on two *fresh* judge fields, `report_says` and `source_says`
(`assay.go:1120-1131`): `sourceSaysBad = plainBadWord ∥ verbatimRun(quotes,5) ∥ contrastWord`;
`reportSaysBad = plainBadWord ∥ verbatimRun(claim,3)`. The old chain has no `report_says`, and its
`so_what` was under **no** discipline (its contrast is now supplied by the code-assembled stakes
frame), so proxying `report_says` from `so_what` over-fires — a naïve pass flags 57/69 on legitimate
claim references and is discarded. The **defensible** proxy is the one field that existed *and*
regenerates similarly: **`source_says`**. A leaf is counted if its 2026-09-08 `source_says`
- carries a contrast/negation word `contrastWord` catches (`not/only/just/instead/but/rather than`), or
- lifts a **5-word** verbatim run from one of its own verified quotes (`verbatimRun(quotes,5)`),

either of which forces a plain retry on re-run (`runTokens`/`splitWords` semantics per `assay.go`).

**Count: 25 leaves** (source_says non-empty on 67; the 2 empty are faithful/absent leaves that carry
none). By trip:
- contrastWord (11): F7, F13, F14, F23, F25, F28, F30, F34c, F34d, F39a, F41
- 5-word quote-lift (20): F2b, F3, F5, F9, F10, F11, F13, F15, F17, F23, F30, F34b, F34c, F39a, F41,
  F45, F46a, F46c, F49a, F54b
- **Union (25):** F2b, F3, F5, F7, F9, F10, F11, F13, F14, F15, F17, F23, F25, F28, F30, F34b, F34c,
  F34d, F39a, F41, F45, F46a, F46c, F49a, F54b

By current verdict: 22 `partial`, F13 + F34d `overstated`, **F46a `contradicted`**.

Two caveats on reading this as "will move":
- A plain retry **regenerates the field text at temperature; it does not deterministically flip the
  verdict.** These 25 are where the retry *fires*, i.e. where verdict perturbation is most likely —
  an upper-lit set, not a prediction of change.
- Verdicts can also move **outside** this set from ordinary N=3 temperature spread on a fresh judge.
  The one confirmed mover, **F46b (contradicted→partial** under the HEAD judge, `eval-f46b/`), is
  **not** in the 25 — its source_says is empty (it was `absent` in the old run), so the field rules
  don't touch it; it moved on a full re-judge. So treat 25 as "where the new rules bite," and expect
  a few more movers from spread. Only the actual re-run (item 3) settles it.
