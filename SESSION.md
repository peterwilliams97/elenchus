# SESSION — parking lot

Deferred and abandoned items, so reversals and dead ends survive the session boundary. Not a TODO
backlog (`TODO.md`) and not a change log (`rigour-map/decision_log.jsonl`) — this is the holding pen
for "decided to defer," with enough context to resume cold.

## 3b adversarial axis probes — BUILT (d013), calibration ongoing

The §2.2 / 3b axis probes are now public worked examples under `examples/destructive/` (seven probes
+ `run.sh` + reader README), pre-registered and first-calibrated 2026-06-10 on
`claude-haiku-4-5-20251001`, N=10 (`testing/calibration_log.jsonl`). This advances — but does not
close — the W6 line in the deferred list below. Deferred follow-ups from the calibration:

- **Laundering substantive-cell re-run (priority).** On haiku the laundering audit returned
  `faithful + hollow + refuted` (boundary held — no false pass), but substance rated Ballmer's bare
  forecast `hollow`, so the textbook `substantive + refuted` demonstration did not appear. **Re-run
  `examples/destructive/laundering/` on `claude-sonnet-4-6` and `claude-opus-4-8`** to see whether a
  stronger model credits the *argued* prediction as `substantive`/`partial` (then the trap springs
  fully). If even strong models call it `hollow`, consider a second real fixture whose substance-core
  is a present-tense factual assertion rather than a forecast. Until then the priority demonstration
  is *advanced, not done*.
- **Cross-model calibration (3c/W7 overlap).** The whole 3b set is calibrated on haiku only. Re-run on
  sonnet/opus to characterize verdict drift (the regression signal for any `*Sys` prompt edit). Append
  to the same `calibration_log.jsonl`; do not average across models.
- **Envelope finding — decompose splits motte-from-bailey (W4).** Substance `decompose` atomizes the
  motte and bailey into separate claims, so the *retreat between them* is partly dissolved before the
  critic grades. Worth a dedicated W4 fixture: a single claim whose defect lives in the *join* between
  two clauses, to measure how often decompose severs it.
- **Envelope finding — adjacent-axis reach is wider than the gap taxonomy assumed.** Composition was
  caught via Counterexample (10/10) and survivorship via Base-rate (8/10); only **category error**
  passed as a clean un-named gap. Re-check on other models before treating "axes miss composition /
  survivorship" as settled — on haiku they largely do not.

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
- **W6 — axis-gap probes.** ✅ BUILT + first-calibrated (d013) as `examples/destructive/` (incl.
  `axis-gaps/`, `motte-and-bailey/`, `reference-class/`). Haiku finding: category error is the genuine
  un-named gap; composition/survivorship are caught via adjacent axes. Remaining: cross-model re-run
  (see the 3b section above).
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
