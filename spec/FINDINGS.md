# FINDINGS.md — Operating Envelope, Calibration Results, Known Bugs, Permanent Limits

Condensed from TESTING.md, SESSION.md, BACKGROUND.md, and testing/calibration_log.jsonl.
Provenance markers distinguish measured results (with date, model, N) from design expectations
(pre-calibration reasoning).

---

## Operating envelope

Source: TESTING.md §"Operating envelope", BACKGROUND.md §2.3. Rows marked *measured* cite a
specific calibration entry in testing/calibration_log.jsonl.

| Region                                                   | Expected reliability | Basis |
|--------------------------------------------------------------|------------------|-----|
| Faithfulness with source present                             | Highest          | Truth-maker is in the context window (STV asymmetry: being shown the reference is what moves verifier quality). Both Defender and Critic can quote the source directly. |
| Refutation of internally contradictory claims                | High             | Reachable by reasoning alone; an internal contradiction is a deduction, not a retrieval. |
| `unverifiable` on predictions / intentions / armchair claims | High — by design | *Measured:* resolved-prediction case (Ballmer iPhone forecast) returned `refuted` 10/10, not `unverifiable` (the boundary held against laundering); reflexive grounding canary returned `unverifiable` 5/5 (haiku, 2026-06-11). |
| Relational defects (equivocation, motte-and-bailey)          | Partly dissolved by decomposition before grading | *Measured:* motte calibration 2026-06-11 — `decompose` typically splits motte from bailey into separate fragments; the retreat between them is partly graded fragment-by-fragment. The move leaks when decompose keeps the conditional whole (1 case in 63 valid verdicts). |
| Substance on sweeping forward predictions                    | Skews `hollow` (false-attack direction) | *Measured:* laundering + bare-vs-contextualized (2026-06-11). Bare Ballmer forecast `hollow` 10/10; central prediction `hollow` 10/10 even with full context. Fatal axis is **Counterexample**, not Evidence. Context does not relieve the verdict. |
| Substance on novel / specialist domains                      | Lowest — treat with most suspicion | W1: Producer and Critic share the same model; a blind spot in one role survives in the other. No reference external to parametric knowledge. |
| Positive grounding confirmation (`supported`)                | Bounded by retrieval quality; never armchair | `crossCheckEvidence` proves a URL was retrieved, not that the page supports the sentence (W2/W3). The structurally weakest verdict. |
| Tool's own headline claim | Untestable by the tool itself    | Permanent limit (W9). *Measured:* reflexive canary `unverifiable` 5/5 (haiku, 2026-06-11) — the grounding boundary held; this is the only reachable check. |

---

## Calibration results per probe

First calibration: 2026-06-10, `claude-haiku-4-5-20251001`, N=10 each.
Fragment-attributed re-calibration: 2026-06-11, same model, N=10 each.
Source: testing/calibration_log.jsonl lines 1–22.

**Unit-of-analysis correction (TESTING.md, 2026-06-11).** The 2026-06-10 tallies are per-atomic-
claim. Each probe carries ONE engineered defect; `decompose` expands it into several fragments, most
of which are clean scaffolding. A `substantive` verdict on a clean fragment is the critic being
correct, not a false pass. **The false-pass metric is defined per defect-carrying fragment: a defect
fragment rated `substantive` with no target-axis (or legitimate-adjacent) finding — in practice an
empty critique.** The 2026-06-10 fragment chains are **unrecoverable** (instrument error d016;
chains written to $TMP and deleted on EXIT). Those lines are retained as first distribution
snapshots, annotated `fragment_attribution: unrecoverable`; all false-pass counts come from the
2026-06-11 fragment-attributed runs.

### motte-and-bailey (Equivocation axis)

Claim: fluent argument that retreats from a strong sense of "intelligent" to a trivial one.

| Date | N | Verdict distribution (fragment totals) | False-pass |
|---|---|---|---|
| 2026-06-10 | 10 | 40 hollow · 19 partial · 7 substantive · 2 error | unrecoverable (instrument error d016) |
| 2026-06-11 (fragment-attributed) | 10 | N_valid=63: 41 hollow · 13 partial · 9 substantive · 1 error | **1/63** — run 10: equivocation conditional kept whole by decompose; rated substantive with **empty critique**. 8/9 substantive land on the defensible motte alone (critic correct). |

Reading: equivocation fires in 10/10 runs. The false-pass signature is an empty critique — no axis
fired. `decompose` usually dissolves the motte/bailey split before grading; when it does not, the
move can leak.

### reference-class (Base rate / magnitude axis)

Claim: a real statistic compared against a gamed reference class to inflate apparent magnitude.

| Date | N | Verdict distribution | False-pass |
|---|---|---|---|
| 2026-06-10 | 10 | 14 hollow · 16 partial · **0 substantive** | — nothing to attribute |

Reading: **0/30 substantive**. The gamed class never passed clean. The DEFECT.md self-prediction
("expect elevated substantive/partial leakage here") was **falsified** by the observed 0 substantive.
No 2026-06-11 re-run needed.

### hidden-premise (Hidden premise axis)

Claim: a conclusion valid only under the unstated premise that cloud adoption is already complete.

| Date | N | Verdict distribution (fragment totals) | False-pass |
|---|---|---|---|
| 2026-06-10 | 10 | 17 hollow · 15 partial · 3 substantive · 1 error | unrecoverable (instrument error d016) |
| 2026-06-11 (fragment-attributed) | 10 | N_valid=36: 15 hollow · 13 partial · 7 substantive · 1 "substantial" · 1 error | **1/36** — run 8: "customer choice indicates where engineering effort should be allocated" rated substantive with **empty critique**. 6/8 substantive on the clean honest figure ("91% of new signups chose cloud") — correct. 1 defect fragment ("retire on-prem") substantive but Hidden-premise+Base-rate+Causality fired — defect named, not a clean leak. |

W10 noted here: run 3 returned `"substantial"` (non-enum near-miss); passed through and rendered
as-is (calibration_log line 14 note field).

### unfalsifiable-dress (Falsifiability axis)

Claim: an un-falsifiable claim dressed in empirical language ("the pattern holds across
domains … wherever one looks").

| Date | N | Verdict distribution | False-pass |
|---|---|---|---|
| 2026-06-10 | 10 | 30 hollow · 9 partial · **0 substantive** · 1 error | — nothing to attribute |

Reading: **0/39 substantive** (strongest catch). Falsifiability fires in 10/10 runs.
No re-run needed.

### causal-narrative (Causality vs correlation axis)

Claim: a mechanism story laid over a single correlation, with no causal evidence cited.

| Date | N | Verdict distribution (fragment totals) | False-pass |
|---|---|---|---|
| 2026-06-10 | 10 | 23 hollow · 19 partial · 3 substantive | unrecoverable (instrument error d016) |
| 2026-06-11 (fragment-attributed) | 10 | N_valid=43: 18 hollow · 21 partial · 4 substantive · 1 error | **0/43** — all 4 substantive had Hidden-premise fire (same defect from an adjacent angle = probe working). No defect fragment passed with an empty critique. |

Reading: the "3 leaked" reading from 2026-06-10 was the unit-of-analysis artifact.
False-pass 0/43 on the corrected metric.

### axis-gaps (category error / composition / survivorship — no target axis by design)

Claim set: three defect types the seven axes do not directly name.

| Date | N | Verdict distribution (fragment totals) | Mapped-limit leak |
|---|---|---|---|
| 2026-06-10 | 10 | 55 hollow · 10 partial · 4 substantive | unrecoverable (instrument error d016) |
| 2026-06-11 (fragment-attributed) | 10 | N_valid=70: 55 hollow · 14 partial · 1 substantive | **1/70** — survivorship fragment ("founders who changed the world all ignored the skeptics") rated substantive; survivorship-specific axes silent; Hidden-premise fired weakly. In the same run the sibling fragment was caught hollow via Base rate / Counterexample. |

Envelope finding: composition was caught via Counterexample (10/10 across runs); survivorship via
Base rate (8/10). **Category error is the genuine un-named gap** — those fragments stayed quiet.
`[~]` stays: envelope-mapping is logged and re-checked, never "passed."

### laundering (all three modes via -audit)

Real source: Steve Ballmer, USA TODAY CEO Forum, April 30, 2007. "There's no chance that the
iPhone is going to get any significant market share." Fetched and provenanced in
examples/destructive/laundering/PROVENANCE.md. Source file gitignored; recreatable from the
verbatim snippet in PROVENANCE.md.

| Date | N | Faithfulness | Substance | Grounding |
|---|---|---|---|---|
| 2026-06-10 | 10 | faithful 10/10 | **hollow 10/10** | refuted 10/10 |

Two findings pulling opposite ways:

**Affirmative non-occurrence — the laundering false pass did NOT occur.** Grounding returned
`refuted` 10/10: the resolved-prediction blindness the probe hunts (grounding confirming a
famous-but-false forecast from the armchair) never happened on haiku. Retained as a live
hypothesis for sonnet/opus runs. `[~]` stays.

**Measured false-attack (first logged datum for that metric).** Substance rated the bare,
falsifiable Ballmer prediction `hollow` 10/10. This diverges from the README spec for
`substantive` ("well-formed and falsifiable"). Investigated by the bare-vs-contextualized probe.

### bare-vs-contextualized (critic calibration probe — no engineered claim defect)

Same Ballmer claim, twice: bare one-liner vs. with the real in-text argument (from the laundering
source). Tests whether the false-attack is caused by an Evidence-axis decontextualisation penalty.

| Date | N | Bare verdict | Contextual central-claim verdict | Aggregate shift |
|---|---|---|---|---|
| 2026-06-11 | 10 each | hollow 10/10; fatal axis: Counterexample 9/10 | central claim: **hollow 10/10**; fatal axis: Counterexample 8/10 | aggregate shift toward substantive/partial under context = added concrete sub-claims (decomposition artifact), not the central prediction moving |

**Hypothesis NOT confirmed.** The central prediction rates `hollow` 10/10 with and without context.
The aggregate shift toward `substantive/partial` in the contextual run is entirely the added
concrete sub-claims (`decompose` surfacing them separately). The fatal axis is **Counterexample**
(a counterexample to a sweeping forward prediction is always constructible), not Evidence.
Context does not relieve the verdict. See §"Withdrawn hypothesis" below.

### reflexive canary

Source: examples/reflexive/headline.txt — "Seeing the three columns disagree causes a trained
user to recognise their skill as transferable."

| Date | N | Verdicts |
|---|---|---|
| 2026-06-11 | 5 | **unverifiable 5/5** — canary held |

Required result every run: `unverifiable`. Any `supported` = self-sealing failure (grounding
confirming the tool's own value proposition from the armchair). None occurred.

---

## Withdrawn hypothesis

**Evidence-axis decontextualisation penalty** (SESSION.md 2026-06-11, §A4).

The original hypothesis: substance rates bare, decontextualised claims `hollow` due to an
Evidence-axis penalty on in-text support. If confirmed, the fix was to add an instruction to
`substanceCriticSys` ("evaluate form, not in-text support").

The bare-vs-contextualized probe **falsified this at the claim level**. The central prediction rates
`hollow` 10/10 in both bare and contextual conditions. The fatal axis is **Counterexample**, not
Evidence (Evidence only weakens). A counterexample to a sweeping forward prediction is always
constructible; context does not remove that counterexample.

**Therefore the planned `substanceCriticSys` prompt fix is withdrawn.** It would not move the
verdict because the contextual claim *has* in-text support and still dies on Counterexample.

**Reframed open question (SESSION.md):** how should substance treat a falsifiable *forward
prediction* that Counterexample can always be constructed against? "iPhone won't get significant
market share" is well-formed and falsifiable, yet Counterexample kills it as if a single
conceivable counterexample refutes a probabilistic forecast. Candidate: distinguish a
counterexample that defeats a universal from one that merely contests a prediction. A substance-
critic design question, not a prompt tweak — spec before touching `substanceCriticSys`.

---

## Open design questions

All from SESSION.md 2026-06-11.

**1. Forward prediction / Counterexample.** The Counterexample axis fires fatally on probabilistic
forward predictions because it always finds a conceivable counterexample. Whether that is correct
behavior is an unresolved design question. A precise fix would require the critic to distinguish
"a counterexample that defeats a universal" from "one that merely contests a prediction." No
implementation path yet; requires spec.

**2. Referent ambiguity.** Running substance on the repo's own claims (examples/reflexive/) showed
the critic reading "assay" as a *chemical* assay in 4 fragments, and rating "two of assay's three
columns" `hollow` because "columns" has no referent out of context. This is the same envelope as
bare-vs-contextualized, caught reflexively: decontextualised specialist terms collide their
referents and skew toward `hollow`. Candidate new 3b fixture: a referent-ambiguity probe — a
domain-polysemous term graded with vs without a one-line domain anchor.

**3. Empty-critique false-pass signature.** Both confirmed false passes (motte run10, hidden-premise
run8) share one signature: the defect fragment rated `substantive` with an **empty critique** (no
axis fired at all). This is the empirical false-pass signature on haiku. It is not yet validated
across models or used to define a detection rule. A Layer-2 test could assert that any `substantive`
verdict with an empty critique (all severities `"clears"` or no critique items) triggers a warning.

---

## Known bugs (W-entries)

Source: BACKGROUND.md §2.1 weakness taxonomy, calibration_log.jsonl observations.

| # | Bug | Code locus | Status |
|---|---|---|---|
| W1 | Correlated producer/critic blind spots | `assayClaim` runs producerSys then substanceCriticSys through the same `c.model` (assay.go:64, 433, 440). A flaw the model cannot see as Producer it also cannot catch as Critic. | Open; no test possible without cross-model harness |
| W2 | Positive grounding confirmation unreachable by reasoning | `crossCheckEvidence` checks URL presence, not content support (assay.go:512). An internally coherent, externally unanchored claim can return `supported`. | Open; structural limit |
| W3 | Citation-span provenance gap (narrowed) | `callClaude` extracts retrieved URLs from `web_search_tool_result` and matches model-cited URLs host+path via `normalizeURL` (assay.go:792–808, 551). Proves a URL was retrieved, not that *this sentence* is backed by *that span*. True API citation blocks never read. | Open; smaller than README:164–166 states (that paragraph is stale per SESSION.md) |
| W4 | Decomposition / atomization errors | `splitSummary` is pure regex (assay.go:945–946); `decompose` is one LLM call (assay.go:418). A paragraph joining claims with "and"/";" and no list markers → one mega-claim graded as a unit. `decompose` may over-split or merge nested hedges. | Open; fuzz test planned but not built |
| W5 | Literalization / intent-reconstruction error | `intendedProposition` (assay.go:316) substitutes `what_source_actually_says` when faith verdict ∈ {partial, overstated}; reconstruction comes from `faithCriticSys`. If wrong, `evidenceClaim` grounds the wrong proposition. | Open |
| W6 | Axis incompleteness | Seven fixed axes in `substanceCriticSys` (assay.go:572). Category error, composition/division, survivorship framing not named. | *Partly measured:* composition caught via Counterexample (10/10), survivorship via Base rate (8/10). **Category error is the confirmed un-named gap.** |
| W7 | Calibration drift across models | Whole pipeline keys on `c.model`; no labeled gold set exists. Same fixture on haiku vs sonnet vs opus can yield different verdicts; magnitude unknown. | Blocked — requires a real labeled gold set |
| W8 | Survivorship bias in the decision corpus | `rigour-map/decision_log.jsonl` historically skipped dead ends (d001/d003/d004/d006 all backfilled). Any classifier built on the corpus inherits a denominator missing its own failures. | Partly repaired by backfill; ongoing discipline |
| W9 | Headline self-claim unfalsifiable from the armchair | `README.md` value proposition about effects on people. By the tool's own axis boundary it cannot be settled by reasoning. | Permanent limit; reflexive canary is the only reachable check |
| W10 | No verdict-enum validation | `unmarshalLoose` (assay.go:937) accepts any string into `Verdict`; no caller checks membership. *Observed in calibration:* hidden-premise run3 (2026-06-11) returned `"substantial"` (non-enum near-miss); rendered as-is. | Open; expected to fail today; cheap to test |
| W11 | Single-reparse / truncation collapses signal to `error` | `callJSONSourced` retries once (assay.go:683–693); `maxTokens=1500` truncation returns an error → `assayClaim` yields `Verdict:"error"`. Long critiques or stubborn malformed JSON collapse to `error` silently. | Open |

---

## Permanent limits (no test possible)

Source: BACKGROUND.md §2.3, TESTING.md §3d.

**The headline claim is untestable by the tool (W9).** "Seeing the three columns disagree causes
a trained user to recognise their skill as transferable" is a grounding claim about effects on
people. It cannot be settled by reasoning, by web search, or by assay running on itself. It is
carried as a hypothesis awaiting field evidence. The reflexive canary (`unverifiable` 5/5 on haiku,
2026-06-11) proves only that grounding is not currently self-sealing — not that the claim is true.

**Positive grounding confirmation (W2/W3) always needs the truth-maker.** No amount of reasoning
or prompt improvement can make `supported` reliable without a retrieved page whose *content*
backs the sentence. This is not a bug to fix; it is the axis boundary.

**Correlated blind spots (W1) require cross-model dissociation.** When Producer and Critic share
a model, errors invisible in either role are invisible to the loop. The only instrument is
divergence: a cross-model producer/critic run (not yet built) or a gold-labeled disagreement
audit.
