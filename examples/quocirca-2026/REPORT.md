# quocirca-2026 — assay run (2026-09-12)

## 2026-09-12 (follow-up) — three fixes, one commit; sites rebuilt, uncommitted

Rendered from the same two chains (no new model calls). Code committed; both sites left uncommitted
(licensed report text).

1. **Recommendation = any root child but `base`** (`internal/tree/argument.go`). Dropped the `R<n>`
   spelling test; the count is now structural. Quocirca's root paragraph reads **"Of 12
   recommendations"** (S1–S7, B1–B5) instead of "Of 0". Resolves open-item #1 below. LCEIC still reads
   "Of 11 recommendations" (R1–R11).
2. **`single_source: true` manifest flag** (`internal/manifest`, `assay.go`, `internal/tree`). On a
   corpus that holds the report as its own only source, a leaf `absent` is not a grounding miss (there
   is no other document to miss) but a claim stated once and not repeated. It renders `uncorroborated`,
   the verdict key explains it ("said once in the report, not repeated elsewhere"), and the argument
   tree derives it as **weakened, not failed**. Effect on the tally: recommendations resting on
   single-page findings now weaken rather than collapse (S2/S3/S6/S7, B1/B2/B3 weakened; only B5 fails,
   on KF8 contradicted; B4 open on a contested KF5a).
3. This report gained the 8 overstated/contradicted leaves below.

### The 8 overstated / contradicted leaves (merged modal verdict, `evidence/2026-09-12-{a,b}`)

Six overstated, two contradicted (audit.md). Page numbers are report.pdf pages (`report_page_offset:
0`). Note on argument-tree badges: E3 and K37 are `route=evaluative`, so the tree greys them
`opinion` and E1 is a 3/3 split (`overstated`/`partial`) — their faithfulness verdicts, below, still
stand; the tree simply does not let an opinion or an unsettled leaf drive a recommendation.

**E1 — overstated (split 3/3 overstated/partial).**
- Claim, §Executive summary p2: "Two-thirds (67%) of organisations in the UK, France, Germany, and the US have experienced print-related data losses in the past year."
- Conflicts with, §Key findings p4: "Two-thirds (67%) of organisations report at least one print-related breach in the last year." — the source gives no four-country scope.

**E2 — overstated.**
- Claim, §Executive summary p2: "80% of organisations remain heavily reliant on printing to support their business operations."
- Conflicts with, §Key findings p4: "While print remains embedded in business operations (80% describe their organisation as reliant on print to support their activities)" — source says "reliant", not "heavily reliant".

**E3 — overstated (`route=evaluative`; tree badge opinion).**
- Claim, §Executive summary p2: "The only sustainable response is to increase resilience through a layered, continuously managed print security strategy that protects devices, users, documents, and workflows as part of the wider enterprise security environment."
- Conflicts with, §Key findings p5: "Buyers should move from reactive controls to continuous governance, while print vendors should help close the divide through assessment-led services, identity controls, fleet visibility, AI-enabled protection, and future-ready security roadmaps." — the source recommends continuous layered security but never calls it the "only sustainable response".

**E11 — overstated.**
- Claim, §Print remains a growing security vulnerability p2: "On average, each print-related breach costs £1 million."
- Conflicts with, §Key findings p4: "The average cost of a print-related data breach now exceeds £1 million" — source says "exceeds"; the claim drops the qualifier, turning a floor into a point figure.

**E21 — contradicted.**
- Claim, §Widespread AI integration presents both risk and reassurance p3: "54% are very or extremely concerned about AI-driven attacks."
- Conflicts with, §Key findings p5: "More (54%) are concerned that AI may be leveraged to create more sophisticated print-related threats, up from 40% in 2025." — the 54% is print-related threats, not AI-driven attacks generally.

**E23 — contradicted.**
- Claim, §Widespread AI integration presents both risk and reassurance p3: "85% say it is very or somewhat important that suppliers develop AI-driven security capabilities."
- Conflicts with, §Key findings p5: "55% now consider it very important that providers use AI and machine learning to identify potential security threats and cyberattacks, compared with 41% in 2025 and 34% in 2024" — the source figure is 55%; 85% appears nowhere.

**K22 — overstated.**
- Claim, §Key findings p4: "The average cost of a print-related data breach now exceeds £1 million."
- Conflicts with, §Recommendations (Buyer) p10: "With the average print-related breach now exceeding £1 million" — source says "breach"; the claim narrows to "data breach".

**K37 — overstated (`route=evaluative`; tree badge opinion).**
- Claim, §Key findings p5: "While print-specific quantum threats may not be immediate, suppliers have an opportunity to provide post-quantum roadmaps, device-refresh guidance, and education."
- Conflicts with, §Key findings p5 (backfill left the passage id unresolved — matched 0 or >1): "quantum resilience is moving onto the planning agenda for more mature buyers. Suppliers should publish practical roadmaps that explain how AI will be used for anomaly detection, automated remediation, and content protection and how post-quantum cryptography will be introduced across devices, firmware, certificates, and secure communications." — the source never says the threat is "not immediate".

---

## Original run (2026-09-12)

Root replaced, two full Sonnet N=3 judge runs, backfilled, rendered to a PDF-linked site. All on
branch `feat/single-report-corpus-2026-09-12`. No commits (as asked).

## Result

- **Rollup** (modal verdict, 60 argument-leaf claims, merged over both runs):
  absent 43 · overstated 6 · faithful 5 · partial 4 · contradicted 2.
  Supported (faithful+partial) = 9 / 60. The "absent"-heavy shape is the intended internal-consistency
  signal: each claim is judged against the REST of the report (its own § excluded), so a fact stated
  only on its own page has no in-document corroboration → absent; only cross-page restatements
  (the E↔K pairs) and cross-referenced figures come back faithful/partial.
- **Cost**: run-a $2.6950 + run-b $2.6660 = **$5.36** (Sonnet claude-sonnet-4-6, 352 calls each,
  cache_hit 99%, 0 errors, ~36 min each, concurrent).
- **Stability** (merged 6 samples/leaf): **settled 57 · wobble 0 · contested 3.**
- **Root paragraph (verbatim, §Executive summary p2):**
  "Despite heightened awareness of cybersecurity threats, print infrastructure remains a significant
  vulnerability for many organisations; the only sustainable response is to increase resilience through
  a layered, continuously managed print security strategy that protects devices, users, documents, and
  workflows as part of the wider enterprise security environment."

## Site

`examples/quocirca-2026/site/` — index.html + review.html, report.pdf copied (1), 0 missing.
All 60 report §-refs deep-link via the manifest rules: p2×18, p3×3, p4×39. Backfill: run-a 31/33,
run-b 32/34 quotes resolved to report pages (2 unresolved each — a quote matching 0 or >1 passages,
left blank, never guessed).

**Refuter met:** the "§Executive summary" leaf deep-links to `sources/report.pdf?p=2#page=2`
(verified in the built `review.html`).

## What was built to make this runnable (real content, no fabrication)

- `claims-machine-full.txt` — tab-delimited judge input, the 60 argument-leaf claims reformatted from
  the pipe `claims-machine.txt` (id, `p<N>=<§ heading>` path, claim verbatim, cites=report.pdf, route).
- `sources/report.txt` — pdftotext of the real report.pdf (15 pages, gitignored under `sources/*`).

## Code (branch `feat/single-report-corpus-2026-09-12`)

1. **Single-report corpus** (`internal/retrieve`): `reportPassages` splits report.txt into page-tagged
   paragraph passages (`report#p<page>#<n>`, Source=report). `internal/manifest` CanonicalDoc maps
   those to `report.pdf p.N`.
2. **Self-exclusion** (`assay.go` `dropOwnPage`/`pageOfPath`): each claim is judged against the rest of
   the report, its own printed page dropped — applied on the shared union at judge time (a per-claim
   pre-filter is defeated by the group-union cache), gated to Source=report (hearing/submission
   grounding is untouched).
3. **Manifest-driven report links** (your 2nd instruction): `manifest.LoadReportLinks` reads
   `report_page_offset:` and a `sections:` name→page table; `renderReview` resolves a §-ref (numbered
   via inline page + offset, named via the table) instead of the old hardcoded `printedToPDF=18` +
   `§[\d.]+` regex. vic-lceic MANIFEST now declares `report_page_offset: 18`; quocirca declares 0 + a
   sections table.

Tests added (all green via `./build.sh`): TestReportPassages, TestDropOwnPage, TestLoadReportLinks,
TestRenderReviewReportLinks, extended TestCanonicalDoc.

## Open items (NOT done / your call)

- **Root tally reads "Of 0 recommendations, 0 hold."** `internal/tree/argument.go:isRecommendation`
  only recognises the LCEIC `R<n>` id; quocirca's recommendations are `S1–S7`/`B1–B5`, so the root
  SUMMARY block counts 0. The tree derivation itself is correct (S/B nodes render with derived
  judgements) — only the headline miscounts. Pre-existing, same hardcoded-LCEIC class as the link
  rules were; not fixed (outside the two tasks). Fix options: accept `S`/`B`/`R` prefixes, or make the
  recommendation-id set data-driven like the link rules now are.
- **Named-section page precision**: the `sections:` table maps a named § to its START page, so the 13
  "Key findings" claims that sit on printed p5 deep-link to p4 (the section start). Refuter unaffected.
- **Licensed material**: `evidence/*/` chains and `site/` hold verbatim report quotes; only `sources/*`
  is gitignored. Add ignores for these before any commit (nothing committed this session).
- vic-lceic's `report_page_offset: 18` is hand-added to a `-make-manifest`-generated file; re-add after
  any regenerate (noted inline in that MANIFEST).

## FILED
- Chains: `evidence/2026-09-12-a,-b/claims-machine-full.faithfulness.jsonl` (backfilled) — FILED (uncommitted, gitignore pending).
- Site: `examples/quocirca-2026/site/` — FILED (uncommitted, licensed).
- This report: `examples/quocirca-2026/REPORT.md` — FILED.
