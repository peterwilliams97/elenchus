# SESSION — parking lot

Deferred and abandoned items, so reversals and dead ends survive the session boundary. Not a TODO
backlog (`TODO.md`) and not a change log (`rigour-map/decision_log.jsonl`) — this is the holding pen
for "decided to defer," with enough context to resume cold.

## Next tracked step (deferred from d011)

**Implement the BACKGROUND.md §2.2 destructive-test program.** The doc carries the specifications; the
tests are not written. Per "separate done from advanced," `BACKGROUND.md` (d011) is *done*; the test
suite is *advanced, not done*. Implementation is its own decision_log entry when it starts.

Order suggested by cost and signal (cheap/red-today first):

- **W10 — verdict-enum validation.** Stub `cfg.call` → `"verdict":"yes"`; assert reject/normalize.
  Expected red today. Likely pairs with a small guard near `unmarshalLoose` (assay.go:937) /
  per-mode verdict sets. Cheapest real fix.
- **W3 — citation-block provenance.** Test asserting displayed `-evidence` URLs come from API citation
  blocks, not model `sources`. **Expected red today by design** — this failing test *is* the recorded
  limit. Do not green it without actually reading citation blocks in `callClaude` (assay.go:792–808).
- **W11 — truncation/malformed → error masks signal.** Stub `max_tokens` + persistent bad JSON via the
  `httpClient` RoundTripper seam; assert the `error` carries a cause and coverage is surfaced.
- **W4 — splitSummary/decompose fuzz.** Run-on "and"/";" blobs, nested hedges, single-line multi-claim.
  Characterize the regex envelope (assay.go:945–946); no fabricated fixtures — derive adversarial
  inputs structurally.
- **W6 — axis-gap probes.** One claim each for category error, motte-and-bailey, reference-class
  gaming; run substance; record what the seven axes (assay.go:571) miss.
- **W1 / W2 / W5 — require the real model** (shared-misconception set; coherent-but-unanchored set;
  subtle intended-proposition source). Run on `claude-haiku-4-5-20251001` first for iteration, then a
  stronger model for the verdict to trust. Inputs must be real and provenanced.
- **W7 — calibration drift.** BLOCKED: needs a real, labeled gold set that does not exist in-repo. Do
  not fabricate labels. If none can be obtained, the honest output is "could not obtain real gold set."
- **W8 — survivorship audit** of `decision_log.jsonl` (ratio of dead-end to shipped rows).
- **W9 — no runnable test.** Permanent limit; the only check is review discipline (never state the
  headline as settled).

## Maintenance flagged, not yet done

- **README:164–166 is stale.** It still says `-evidence` URLs are model-retyped and unchecked;
  `crossCheckEvidence` (d009/d010, assay.go:512) now cross-checks them against retrieved URLs. Restate
  the gap at its true, narrower size (W3). Separate maintenance edit — not bundled with d011.

## Roads not taken

- (none recorded this session)
