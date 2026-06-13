# FIXTURES.md — Examples/ Manifest

Every probe and worked example in the repo, with what it tests, provenance, run protocol, and
the exact file list that moves to a new repo. All directories under `examples/` are copied as-is.

---

## examples/dan_shipper/

**What it tests:** nominal operation — all three modes via `-audit` on a real transcript.
The primary integration example. Used in CLAUDE.md as the recommended end-to-end validation
fixture during development: "use `examples/dan_shipper/` with `-model claude-haiku-4-5-20251001`
for speed, then re-run on the default model for the verdict to trust."

**Provenance:** Real. Dan Shipper (CEO of Every) made 12 predictions about the future of work on
Lenny's Podcast. `dan_shipper.txt` is the full transcript; `dan_summary.txt` is a 12-item
summary of those predictions. Both were fetched/transcribed and are included in the repo (not
gitignored — transcript is the user's own summary + a public appearance, not a re-published
copyrighted document).

**Files that move:**
```
examples/dan_shipper/dan_shipper.txt    — full transcript
examples/dan_shipper/dan_summary.txt   — 12-claim summary
examples/dan_shipper/README.md         — run instructions + pre-run results table + pattern guide
```

---

## examples/url-length/

**What it tests:** the axis-boundary lesson. Shows substance and faithfulness helping decisively
where grounding does nothing — and where over-confident reasoning after the armchair wins on the
first two columns creates a trap on the third. The example refuses to pre-fill its grounding
column; verdicts must be run.

**Provenance:** constructed claim (`claim.txt`), but the disputed factual question (where "2048"
came from) is a real public dispute with real sources (Microsoft IEInternals, W3C uri@w3.org
2010, RFC 9110). `README.md` documents the substance + grounding results from a real run (2026-
06-01, `claude-sonnet-4-6`).

**Files that move:**
```
examples/url-length/claim.txt   — the two competing claims
examples/url-length/README.md   — axis-boundary lesson + real run results (substance + grounding)
```

---

## examples/reflexive/

**What it tests:**
- **B1 (canary):** grounding on `headline.txt` must return `unverifiable` every run. Any
  `supported` is a self-sealing failure — the grounding column confirming the tool's own value
  proposition from the armchair. Wired as the always-on final step of `run.sh all` and
  `run.sh canary`.
- **B2 (substance):** `claims.txt` — eight verbatim load-bearing claims from README.md and
  CLAUDE.md run through the dialectic. Measures whether the tool rates its own claims honestly.
  Found: referent collision under decontextualisation (the critic read "assay" as a chemical
  assay in 4 fragments). Headline rated `partial` as pre-registered.

**Provenance:** constructed inputs (verbatim excerpts from the repo's own documentation — not
fabricated, because the construction IS the ground truth). Results in `results/` are real assay
output from 2026-06-11, `claude-haiku-4-5-20251001`.

**Files that move:**
```
examples/reflexive/claims.txt                                   — 8 verbatim claims from README/CLAUDE.md
examples/reflexive/headline.txt                                 — the canary headline sentence
examples/reflexive/README.md                                    — B1 + B2 results + Layer-4 spot-audit
examples/reflexive/results/2026-06-11-claude-haiku-4-5-20251001.md
```

---

## examples/destructive/

The eight adversarial axis probes, plus run.sh and the top-level README.

**What it tests:** the failure envelope. Each probe targets one weakness at a time; results are
distributions, not verdicts; calibration is logged. See also spec/FINDINGS.md for calibration
results.

**Run protocol** (from `run.sh`):
- `./run.sh [PROBE|all] [N] [MODEL]` — default: all, N=10, `claude-haiku-4-5-20251001`
- Substance probes run `./crossexam -md` on `claim.txt` (or `bare.txt` / `contextual.txt` for bare-vs-contextualized)
- Laundering runs `./crossexam -md -audit -source sources/ballmer_usatoday_2007.txt summary.txt`
- Each run writes per-run chains to `../../testing/chains/<date>-<model>/<probe>/run-N/` (gitignored)
- Writes a dated markdown summary to each `results/` directory
- Appends one JSON line per probe to `../../testing/calibration_log.jsonl`
- `./run.sh all` ends with the reflexive grounding canary automatically (`run.sh canary`)
- **Not wired into build.sh or `go test`** — calibration, never a CI gate

**calibration_log.jsonl schema:** `{"date", "model", "fixture", "mode", "runs",
"verdict_counts": {verdict: count}, "axis_mentions": {keyword: count}}` for substance/evidence;
`{"faithful_counts", "substantive_counts", "grounded_counts"}` for audit (laundering); additional
fragment-attribution fields for 2026-06-11 re-calibration entries (`fragment_attributed: true`,
`n_valid`, `false_pass`, `target_axis`, `note`); annotation objects for superseded 2026-06-10
lines (`annotation_of`, `fragment_attribution: "unrecoverable"`, `superseded_for/by`).
Full schema: `testing/SCHEMA.md`.

---

### examples/destructive/motte-and-bailey/

**Targets:** Equivocation axis.
**Defect:** A key term ("intelligent") retreats from a strong sense (a system has inner experience)
to a trivial one (it adapts/learns from interactions) when challenged. The strong sense is the
bailey; the trivial sense is the motte.
**Provenance:** Openly constructed. The construction is the ground truth; no real source required
or used.
**Calibration:** 2026-06-10 snapshot + 2026-06-11 fragment-attributed (false-pass 1/63). See
spec/FINDINGS.md.

**Files that move:**
```
examples/destructive/motte-and-bailey/claim.txt
examples/destructive/motte-and-bailey/DEFECT.md
examples/destructive/motte-and-bailey/EXPECTED.md
examples/destructive/motte-and-bailey/results/2026-06-10-claude-haiku-4-5-20251001.md
examples/destructive/motte-and-bailey/results/2026-06-11-claude-haiku-4-5-20251001.md
```

---

### examples/destructive/reference-class/

**Targets:** Base rate / magnitude axis.
**Defect:** A real, accurate statistic compared against a gamed reference class chosen to make a
modest result look impressive ("top quartile of its category" where the category was cherry-picked
to exclude larger players).
**Provenance:** Openly constructed.
**Calibration:** 2026-06-10: 0/30 substantive. Design self-prediction ("elevated leakage") was
**falsified**. No re-run needed.

**Files that move:**
```
examples/destructive/reference-class/claim.txt
examples/destructive/reference-class/DEFECT.md
examples/destructive/reference-class/EXPECTED.md
examples/destructive/reference-class/results/2026-06-10-claude-haiku-4-5-20251001.md
```

---

### examples/destructive/hidden-premise/

**Targets:** Hidden premise axis.
**Defect:** A business conclusion ("retire on-prem products") valid only under the unstated
load-bearing premise that cloud adoption is already universal and complete. The honest figure
(91% of new signups chose cloud) is present and is clean scaffolding.
**Provenance:** Openly constructed.
**Calibration:** 2026-06-10 snapshot + 2026-06-11 fragment-attributed (false-pass 1/36).
W10 observed: one run returned `"substantial"` (non-enum near-miss, rendered as-is).

**Files that move:**
```
examples/destructive/hidden-premise/claim.txt
examples/destructive/hidden-premise/DEFECT.md
examples/destructive/hidden-premise/EXPECTED.md
examples/destructive/hidden-premise/results/2026-06-10-claude-haiku-4-5-20251001.md
examples/destructive/hidden-premise/results/2026-06-11-claude-haiku-4-5-20251001.md
```

---

### examples/destructive/unfalsifiable-dress/

**Targets:** Falsifiability axis.
**Defect:** An un-falsifiable claim ("the pattern holds across domains … wherever one looks")
dressed in empirical vocabulary. No observation could disconfirm it.
**Provenance:** Openly constructed.
**Calibration:** 2026-06-10: 0/39 substantive — strongest catch. No re-run needed.

**Files that move:**
```
examples/destructive/unfalsifiable-dress/claim.txt
examples/destructive/unfalsifiable-dress/DEFECT.md
examples/destructive/unfalsifiable-dress/EXPECTED.md
examples/destructive/unfalsifiable-dress/results/2026-06-10-claude-haiku-4-5-20251001.md
```

---

### examples/destructive/causal-narrative/

**Targets:** Causality vs correlation axis.
**Defect:** A plausible mechanism story ("removing friction keeps engineers in flow → faster
shipping") laid over a single correlation with no causal evidence for the mechanism.
**Provenance:** Openly constructed.
**Calibration:** 2026-06-10 snapshot + 2026-06-11 fragment-attributed (false-pass 0/43). The
"3 leaked" reading from 2026-06-10 was the unit-of-analysis artifact.

**Files that move:**
```
examples/destructive/causal-narrative/claim.txt
examples/destructive/causal-narrative/DEFECT.md
examples/destructive/causal-narrative/EXPECTED.md
examples/destructive/causal-narrative/results/2026-06-10-claude-haiku-4-5-20251001.md
examples/destructive/causal-narrative/results/2026-06-11-claude-haiku-4-5-20251001.md
```

---

### examples/destructive/axis-gaps/

**Targets:** none by design — maps what the seven axes *do not* name.
**Defects tested:** category error (treating a classification problem as an optimization problem),
composition (the team succeeds → every member succeeds), survivorship ("the founders who changed
the world all ignored the skeptics").
**Provenance:** Openly constructed.
**Calibration:** 2026-06-10 snapshot + 2026-06-11 fragment-attributed (mapped-limit leak: 1/70).
Envelope finding: composition caught via Counterexample (10/10), survivorship via Base rate
(8/10). **Category error is the confirmed un-named gap.** Stays `[~]`: never "passes."

**Files that move:**
```
examples/destructive/axis-gaps/claim.txt
examples/destructive/axis-gaps/DEFECT.md
examples/destructive/axis-gaps/EXPECTED.md
examples/destructive/axis-gaps/results/2026-06-10-claude-haiku-4-5-20251001.md
examples/destructive/axis-gaps/results/2026-06-11-claude-haiku-4-5-20251001.md
```

---

### examples/destructive/laundering/

**Targets:** all three modes simultaneously (via `-audit`). The priority probe.
**Defect:** a claim that is simultaneously faithful (accurately attributed), substantive (well-
formed and falsifiable), and false (refuted by evidence). Correct result: `faithful + substantive/
partial + refuted`. This demonstrates the laundering-confidence failure caught in the act.
**Provenance:** REAL source required (a constructed laundering claim would fabricate the faithful
cell). Source: Steve Ballmer, USA TODAY CEO Forum, April 30, 2007. Verbatim: "There's no chance
that the iPhone is going to get any significant market share. No chance."
Retrieved and verified 2026-06-10. Provenance recorded in PROVENANCE.md. Full text in
`sources/ballmer_usatoday_2007.txt`.

**Gitignore rule:** `examples/destructive/laundering/sources/` is gitignored. The source file is
real third-party text; the policy is analysis, not redistribution (per CLAUDE.md). After cloning,
the file must be recreated from the verbatim snippet in PROVENANCE.md. PROVENANCE.md includes
the one-liner `cat > sources/ballmer_usatoday_2007.txt <<'SRC' ... SRC`.

**Calibration:** 2026-06-10: `faithful 10 + hollow 10 + refuted 10`. The laundering false pass
did NOT occur (refuted 10/10). But substance rated the bare forecast `hollow` 10/10 (false-
attack; the textbook `substantive + refuted` cell did not appear). Priority: re-run on
`claude-sonnet-4-6` and `claude-opus-4-8` — a stronger model may credit the argued prediction
as `substantive/partial`, springing the trap fully.

**Files that move:**
```
examples/destructive/laundering/summary.txt            — the claim being audited
examples/destructive/laundering/DEFECT.md
examples/destructive/laundering/EXPECTED.md
examples/destructive/laundering/PROVENANCE.md          — verbatim snippet + one-liner to recreate sources/
examples/destructive/laundering/results/2026-06-10-claude-haiku-4-5-20251001.md
```

**Does NOT move** (gitignored):
```
examples/destructive/laundering/sources/ballmer_usatoday_2007.txt
```

---

### examples/destructive/bare-vs-contextualized/

**Targets:** the critic's calibration (not a claim defect). Tests whether the false-attack rate
on forward predictions is caused by an Evidence-axis penalty on bare, decontextualised claims.
**Defect:** none — this is a *calibration probe*, not an adversarial defect probe.
**Inputs:** two versions of the Ballmer iPhone claim:
- `bare.txt`: the bare one-liner prediction only
- `contextual.txt`: the same prediction with the full in-text argument from the source

**Provenance:** `contextual.txt` argument drawn from the laundering source (real; same provenance
as `laundering/`). The bare claim is derived directly from that source.
**Calibration:** 2026-06-11, haiku, N=10 each. Hypothesis NOT confirmed. See spec/FINDINGS.md
§"Withdrawn hypothesis."

**Files that move:**
```
examples/destructive/bare-vs-contextualized/bare.txt
examples/destructive/bare-vs-contextualized/contextual.txt
examples/destructive/bare-vs-contextualized/DEFECT.md
examples/destructive/bare-vs-contextualized/EXPECTED.md
examples/destructive/bare-vs-contextualized/results/2026-06-11-claude-haiku-4-5-20251001.md
```

---

### examples/destructive/ top-level

```
examples/destructive/run.sh     — calibration runner: N runs per probe, distributes verdicts,
                                   logs to testing/calibration_log.jsonl, writes results/ summaries,
                                   ends with reflexive canary. NOT wired into CI.
examples/destructive/README.md  — how to read distributions; training-data caveat; probe table
```

---

## Gitignored — does not move

```
examples/destructive/laundering/sources/   — real third-party text; recreate from PROVENANCE.md
/testing/chains/                           — raw per-run Tier-2 chains; large, not committed
/eval/                                     — per-run eval output from main binary runs
/crossexam                                 — built binary
```

---

## Complete file list (everything that moves)

```
examples/dan_shipper/dan_shipper.txt
examples/dan_shipper/dan_summary.txt
examples/dan_shipper/README.md
examples/url-length/claim.txt
examples/url-length/README.md
examples/reflexive/claims.txt
examples/reflexive/headline.txt
examples/reflexive/README.md
examples/reflexive/results/2026-06-11-claude-haiku-4-5-20251001.md
examples/destructive/run.sh
examples/destructive/README.md
examples/destructive/motte-and-bailey/claim.txt
examples/destructive/motte-and-bailey/DEFECT.md
examples/destructive/motte-and-bailey/EXPECTED.md
examples/destructive/motte-and-bailey/results/2026-06-10-claude-haiku-4-5-20251001.md
examples/destructive/motte-and-bailey/results/2026-06-11-claude-haiku-4-5-20251001.md
examples/destructive/reference-class/claim.txt
examples/destructive/reference-class/DEFECT.md
examples/destructive/reference-class/EXPECTED.md
examples/destructive/reference-class/results/2026-06-10-claude-haiku-4-5-20251001.md
examples/destructive/hidden-premise/claim.txt
examples/destructive/hidden-premise/DEFECT.md
examples/destructive/hidden-premise/EXPECTED.md
examples/destructive/hidden-premise/results/2026-06-10-claude-haiku-4-5-20251001.md
examples/destructive/hidden-premise/results/2026-06-11-claude-haiku-4-5-20251001.md
examples/destructive/unfalsifiable-dress/claim.txt
examples/destructive/unfalsifiable-dress/DEFECT.md
examples/destructive/unfalsifiable-dress/EXPECTED.md
examples/destructive/unfalsifiable-dress/results/2026-06-10-claude-haiku-4-5-20251001.md
examples/destructive/causal-narrative/claim.txt
examples/destructive/causal-narrative/DEFECT.md
examples/destructive/causal-narrative/EXPECTED.md
examples/destructive/causal-narrative/results/2026-06-10-claude-haiku-4-5-20251001.md
examples/destructive/causal-narrative/results/2026-06-11-claude-haiku-4-5-20251001.md
examples/destructive/axis-gaps/claim.txt
examples/destructive/axis-gaps/DEFECT.md
examples/destructive/axis-gaps/EXPECTED.md
examples/destructive/axis-gaps/results/2026-06-10-claude-haiku-4-5-20251001.md
examples/destructive/axis-gaps/results/2026-06-11-claude-haiku-4-5-20251001.md
examples/destructive/laundering/summary.txt
examples/destructive/laundering/DEFECT.md
examples/destructive/laundering/EXPECTED.md
examples/destructive/laundering/PROVENANCE.md
examples/destructive/laundering/results/2026-06-10-claude-haiku-4-5-20251001.md
examples/destructive/bare-vs-contextualized/bare.txt
examples/destructive/bare-vs-contextualized/contextual.txt
examples/destructive/bare-vs-contextualized/DEFECT.md
examples/destructive/bare-vs-contextualized/EXPECTED.md
examples/destructive/bare-vs-contextualized/results/2026-06-11-claude-haiku-4-5-20251001.md
```

Total committed files: 44.
