# TESTING.md

The testing program for assay/elenchus. Companion to BACKGROUND.md (design rationale and
failure-envelope analysis); this file tracks what is tested, how, and what each layer's results are
allowed to mean.

## The central constraint

[assay](assay.go)'s core output is a judgment, and a judgment has no deterministic ground truth. The
tool therefore cannot be tested as one thing. It splits into three layers, each tested with the
instrument that layer actually admits. A result from one layer must never masquerade as a result
from another. In particular:
**a clean Layer 3 run is never reported as "the critic is correct," only as "the envelope held on this set, this time."**

| Layer               | What it is                      | Instrument                        | CI gate? |
|---------------------|---------------------------------|-----------------------------------|----------|
| 1. Plumbing         | Pure functions, no LLM          | Exact-output unit tests + fuzzing | Yes — hard gate |
| 2. Orchestration    | The machine around the judgment | Stub via `cfg.call` seam + mutation testing | Yes — hard gate |
| 3. Critic precision | The judgment itself             | Constructive gold sets, adversarial fixtures, N-run distributions | **No** — calibration runs, read by a human |

Layers 1–2 are deterministic and live in `go test ./...` (gated by `./build.sh`). Layer 3 is
measurement: distributions logged over time like a stress-strain curve, never a green check.
A flaky pass/fail gate on a non-deterministic judgment trains us to ignore failures.

Provenance rule applies to test inputs with full force (see CLAUDE.md): no fabricated fixtures
standing in for real artifacts. Where a real document is required and cannot be obtained, the entry
below is marked BLOCKED, not faked.

Status legend: `[x]` done · `[ ]` planned · `[~]` partial · `[!]` known-failing by design ·
`[B]` blocked on a real input

---

## Layer 1 — Deterministic plumbing

Pure functions. Exact assertions. Highest-leverage cheap layer: every downstream verdict lands on
whatever unit decomposition produces, and an atomization bug is invisible at the verdict layer.

- [x] `splitSummary` — newlines, run-together, bullets, blank lines, single-line
- [x] `extractJSON` — fences, embedded braces, preamble
- [x] `unmarshalLoose` — trailing commas, condition-laundering fields
- [x] `mdCell`, `tally`
- [ ] **Fuzz `splitSummary`** (`go test -fuzz`) — merged/run-on claims, nested hedges, mixed list
      markers, unicode bullets, pathological whitespace. Property: no input text is silently dropped
      or duplicated across the split.
- [ ] **Decomposition unit-integrity fixtures** — inputs where the correct atomic-claim boundary is
      known by construction (e.g. two claims joined by "and", a hedge that scopes over one clause
      only). Assert claim count and that each hedge stays attached to its clause.
- [ ] **Round-trip property** — concatenated split output preserves all substantive tokens of the
     input (order-preserving, modulo list markers).

## Layer 2 — Orchestration via the `cfg.call` seam

Deterministic tests of the machine around the judgment: given the critic returned *this* JSON, does
the code behave correctly? No network, no model reasoning under test.

Existing (confirmatory):
- [x] `TestConditionLaunderingDowngrade` — partial→hollow when `survives_only_by_conditioning`
- [x] `TestConditionLaunderingLoopStop` — loop exits after one critic call, not two
- [x] `TestRunEvidencePropositionSubstitution` — grounds `what_source_actually_says`, not the
      literal claim, when faithfulness returns overstated
- [x] Prompt-presence tests (`TestFaithCriticSysLiteralization`,
      `TestSubstanceCriticSysConditionDiscipline`)

Planned (lock down every branch):
- [ ] **Verdict-enum routing** — each verdict value maps to correct downstream handling:
      `hollow` excluded from residue; `partial` carries `SurvivingClaim`; `error` is never
      silently counted as a win in `tally` or the cross-tab.
- [ ] **`callJSON` retry discipline** — retries exactly once on malformed JSON, then degrades to
      `error` rather than hanging or retrying forever. Stub returns garbage twice; assert two calls
      and an `error` verdict.
- [ ] **`-max-rounds` honored** — stub a critic that always returns `needs_another_round=true`;
      assert the loop stops at the flag value.
- [ ] **Audit cross-tab alignment** — three stubbed verdict sets in; assert each row's
      faithfulness/substance/grounding cells correspond to the same claim (no off-by-one,
      no column swap). This is the table the whole tool's signal lives in.
- [ ] **Tool-use block consumption** — stubbed evidence response containing tool-use blocks
      is consumed without corrupting the extracted verdict.
- [ ] **`maxTokens` truncation behavior** — a truncated (unparseable) critic response routes
      to `error`, never to a partial verdict.

Teeth check:
- [ ] **Mutation testing pass** — invert each guarded branch in `assayClaim`, `faithClaim`,
      `runEvidence` (e.g., flip the `SurvivesOnlyByConditions` check); confirm at least one
      test goes red per mutation. A seam test that survives its own logic being inverted is
      testing nothing. Record the mutation→killed-by mapping here when run.

## Layer 3 — Critic precision (destructive testing)

The frontier. Modeled on industrial destructive testing: the goal is not to confirm nominal inputs
pass but to characterize the failure envelope — engineer inputs that push each axis past its limit
and find where it breaks.

**Protocol for every Layer 3 item:** fixed input, N≥10 runs, across
`claude-haiku-4-5-20251001` / `claude-sonnet-4-6` / `claude-opus-4-8`. Report a distribution
(verdict frequencies), not a single verdict. Log to `testing/calibration_log.jsonl` with
date, model, fixture id, and verdict counts. Metrics tracked over time:

- **False-pass rate** — hollow claims rated `substantive`
- **False-attack rate** — sound claims rated `hollow`
- **Verdict stability** — modal-verdict frequency per fixture per model
- **Boundary compliance** — rate of correct `unverifiable` on armchair-unreachable claims

### 3a. Constructive gold sets (label known by construction, not by our say-so)

A gold set we write *and* label ourselves is circular — it measures whether the critic agrees with
our intuition, the exact proxy this tool exists to externalize. Only labels that are constructively
true qualify:

- [ ] **Self-contradictory claims** — reasoning can refute these with no lookup; constructible
      deterministically (claim ∧ ¬claim in dressed prose). Expected: `refuted` (grounding) /
      fatal finding (substance). Any pass is a hard failure of the critic.
- [ ] **Future-tense predictions** — must return `unverifiable` under `-evidence`. Directly
      asserts the axis boundary holds: falsifiable-in-principle is not settled-by-evidence.
- [ ] **Known-distortion faithfulness pairs** — take a REAL source document (fetched, not
      fabricated; provenance recorded) and apply one controlled edit per fixture: hedge
      promoted to certainty; fabricated detail inserted; context stripped; attribution
      shifted; rhetorical claim literalized; cherry-pick. The label is the edit we made —
      the source is the truth-maker. Assert the faithfulness critic names the specific
      distortion mode. *This is the strongest set in the program: it mirrors the STV result
      that a verifier is reliable precisely when it holds the reference — which is also why
      faithfulness is the most testable mode and substance the least.*
      - [B] Requires real source documents; record each fixture's provenance inline.
- [ ] **Arithmetic/internal-consistency claims** — quantitative claims whose falsity is checkable by
      computation alone. Expected: refutation without retrieval.

### 3b. Adversarial axis probes (one axis at a time)

Built as public worked examples under [`examples/destructive/`](examples/destructive/) — each probe
doubles as a reader-facing demonstration and a destructive test. Here `[x]` means **calibrated and
logged** on a stated model/date with the envelope characterized — never "the critic is correct" (see
the central constraint). First calibration: 2026-06-10, `claude-haiku-4-5-20251001`, N=10 each
(`testing/calibration_log.jsonl`).

**Unit-of-analysis correction (2026-06-11).** The 2026-06-10 tallies are **per-atomic-claim**, but
each probe carries **one** engineered defect and `decompose` splits it into several fragments — most
of them **clean scaffolding** (the honest motte, a real figure). A `substantive` verdict on a clean
fragment is the critic being **right**, not a leak. The false-pass metric is therefore defined
**per defect-carrying fragment**: a false pass is a defect fragment rated `substantive` with **no
target-axis (or legitimate-adjacent) finding** — in practice, an *empty* critique. The four probes
with nonzero `substantive` (motte, hidden-premise, causal, axis-gaps) were re-run 2026-06-11 with
per-fragment chains persisted and the `substantive` verdicts attributed; reference-class (0/30) and
unfalsifiable-dress (0/39) had nothing to attribute. **`N_valid` excludes `error` verdicts**
(instrument failures), reported in their own column. The 2026-06-11 lines carry the false-pass
counts below; the 2026-06-10 lines are retained as the first distribution snapshot and annotated
`fragment_attribution: unrecoverable` (their chains were destroyed by an instrument error — see
[`testing/SCHEMA.md`](testing/SCHEMA.md) and the §3d note). Fragment-attributed false-pass rate
across the four probes: **3 / 212 valid fragment-verdicts (~1.4%)**, versus the naive 22-`substantive`
(~10%) reading — ~77% of `substantive` verdicts are the critic correctly affirming clean scaffolding.

- [x] **Motte-and-bailey fixture** — equivocation axis.
      [`examples/destructive/motte-and-bailey/`](examples/destructive/motte-and-bailey/). 2026-06-10
      snapshot: 40 hollow · 19 partial · 7 substantive · 2 error; equivocation fires. **2026-06-11
      fragment-attributed (N_valid=63):** of 9 `substantive`, **8 land on the defensible motte in
      isolation** (correct — W4 splits motte from bailey before grading) and **1 is a clean false
      pass** (run 10 kept the equivocation conditional whole and passed it with an *empty* critique).
      **False-pass: 1/63.**
- [x] **Reference-class gaming** — base-rate/magnitude axis dressed in a flattering class.
      [`examples/destructive/reference-class/`](examples/destructive/reference-class/). Haiku N=10:
      14 hollow · 16 partial · **0 substantive** — the gamed class never passed clean. **The DEFECT.md
      design note's falsifiable self-prediction — "expect more `substantive`/`partial` leakage here
      than on the other probes" — was FALSIFIED by the observed 0/30** `substantive`. No re-run needed
      (nothing to attribute).
- [x] **Hidden-premise chain** — conclusion valid only under an unstated load-bearing premise.
      [`examples/destructive/hidden-premise/`](examples/destructive/hidden-premise/). 2026-06-10
      snapshot: 17 hollow · 15 partial · 3 substantive · 1 error. **2026-06-11 fragment-attributed
      (N_valid=36):** of 8 `substantive`/`substantial`, **6 land on the clean honest figure** ("91% of
      new signups chose cloud") and 1 defect fragment ("retire on-prem") passed but with the
      hidden-premise axis *firing* (named, not a leak); **1 clean false pass** (run 8, empty critique).
      **False-pass: 1/36, not 3.** (W10: one verdict came back `"substantial"`, a non-enum near-miss.)
- [x] **Unfalsifiable-by-design claim in empirical dress** — falsifiability axis.
      [`examples/destructive/unfalsifiable-dress/`](examples/destructive/unfalsifiable-dress/).
      Haiku N=10: 30 hollow · 9 partial · **0 substantive** — strongest catch, as predicted. No re-run
      needed (nothing to attribute).
- [x] **Causality-from-correlation narrative** — plausible mechanism story over correlational
      evidence only.
      [`examples/destructive/causal-narrative/`](examples/destructive/causal-narrative/). 2026-06-10
      snapshot: 23 hollow · 19 partial · 3 substantive. **2026-06-11 fragment-attributed
      (N_valid=43):** of 4 `substantive`, **every one had the adjacent Hidden-premise axis fire** (the
      same defect from another angle = the probe working) and 3 are decomposed sub-mechanism fragments.
      **False-pass: 0/43** — the "3 leaked" reading was the unit-of-analysis artifact.
- [~] **Axis-incompleteness probes** — defects the seven axes do NOT name (category error,
      composition/division, survivorship framing). Outcome unknown by design; the result maps the
      envelope either way. [`examples/destructive/axis-gaps/`](examples/destructive/axis-gaps/).
      2026-06-10 snapshot: 55 hollow · 10 partial · 4 substantive; composition caught via
      **Counterexample** (10/10), survivorship via **Base rate** (8/10) — adjacent reach is wider than
      the gap taxonomy assumed; **category error remains the genuine un-named gap**. **2026-06-11
      fragment-attributed (N_valid=70):** **1 `substantive`**, on the survivorship fragment with the
      survivorship-specific axes silent (Hidden-premise fired weakly); category-error fragments stayed
      quiet this run. **Mapped-limit leak: 1/70.** Stays `[~]`: envelope-mapping is logged and
      re-checked, never "passed".
- [~] **The laundering fixture (priority)** — a claim that is *faithful, substantive, and false*
      simultaneously, run through `-audit`. Correct result: faithful + substantive + `refuted`; all
      three columns green = the laundering-confidence failure caught in the act.
      [`examples/destructive/laundering/`](examples/destructive/laundering/) — real source (Ballmer,
      USA TODAY 2007; provenance recorded). Haiku N=10: **faithful 10 + hollow 10 + refuted 10**.
      Two findings, pulling opposite ways:
      - **Affirmative non-occurrence — the laundering false pass did NOT occur.** Grounding returned
        `refuted` 10/10: the resolved-prediction blindness the probe hunts (grounding confirming a
        famous-but-false forecast from the armchair) **never happened — the boundary held.** Retained
        as a **live hypothesis for the sonnet/opus runs** (a stronger model that "knows" the iPhone
        succeeded could still refute cheaply; the haiku result does not retire the risk).
      - **Measured false-attack (first datum for the metric).** Substance rated the bare, falsifiable
        prediction `hollow` **10/10**, diverging from the README spec for `substantive` ("well-formed
        and falsifiable — the kind of thing that *could* be true"). This is the **first logged datum
        for the false-attack rate**. Hypothesis: the **Evidence axis penalises bare, decontextualised
        claims** — and since every assayed claim arrives decontextualised by construction, that would
        be a **systematic** bias. Tested directly by the bare-vs-contextualized probe below.
      Stays `[~]`: the textbook `substantive + refuted` cell still wants a sonnet/opus re-run (a forecast
      may need a stronger model to be credited `substantive`); pending in `SESSION.md`.
- [x] **Bare vs contextualized** — a *critic-calibration* probe (no engineered claim defect), testing
      the false-attack hypothesis above.
      [`examples/destructive/bare-vs-contextualized/`](examples/destructive/bare-vs-contextualized/).
      Same Ballmer claim, with vs without its real in-text argument (provenance from the laundering
      source). **2026-06-11, haiku, N=10 each: hypothesis NOT confirmed.** `bare.txt` `hollow` 10/10;
      `contextual.txt` 28 hollow · 21 partial · 11 substantive — but the **central prediction rates
      `hollow` 10/10 in *both*** (the aggregate shift is the *added* sub-claims, a decomposition
      artifact the A1 discipline catches). The fatal axis is **Counterexample**, not Evidence (Evidence
      only weakens). The false-attack is real but its cause is Counterexample-against-a-sweeping-
      prediction, not an Evidence/decontextualisation penalty — so the proposed prompt fix is
      **withdrawn** (SESSION.md). Measurement before intervention, vindicated.

### 3c. Correlated blind spots (cannot be self-tested)

Producer and critic share a model; errors the model cannot see in either role are invisible to any
single-model run by definition. The only instrument is divergence:

- [ ] **Cross-model producer/critic** — producer on model A, critic on model B (requires a small
      harness change or env override; spec the seam first). Disagreements vs. the same-model run
      surface shared blind spots.
- [ ] **Cross-model gold-set judging** — judge 3a/3b fixtures with a model different from the one
      under test; log divergence rate per fixture.
- [ ] **Calibration drift envelope** — same fixture set across haiku/sonnet/opus over time;
      characterize verdict drift as models or prompts change. This is the regression signal
      for prompt edits: any change to a `*Sys` prompt triggers a re-run of the full 3a/3b set.

### 3d. Known-failing and permanent limits

Documented limits live here as tests or standing entries — honest failures, not hidden gaps:

- [!] **Citation-block provenance** — `-evidence` source URLs are model-retyped, not read
      from API citation blocks. The test asserting citation-block provenance FAILS today by
      design; the failing test IS the documented limit. (Excluded from the CI gate; tracked
      here until the implementation reads citation blocks.)
- [x] **Reflexive canary** — run assay on its own README/CLAUDE.md claims. The headline claim
      (visible rigour → user recognition) is a grounding claim about effects on people and MUST
      return `unverifiable`. If grounding ever rates the tool's own value proposition `supported`,
      that is a self-sealing failure inside the instrument — the earliest warning that grounding has
      started confirming from the armchair. Now **wired into `run.sh` as the always-on final step of a
      full pass** (`run.sh canary`, or automatic after `run.sh all`), and demonstrated in
      [`examples/reflexive/`](examples/reflexive/). **Run and logged 2026-06-11,
      `claude-haiku-4-5-20251001`, N=5: `unverifiable` 5/5 — canary held, no self-sealing failure.**
      `[x]` = run and logged on this model/date, **never "passed"**: a held canary says only that the
      boundary held on this run, not that grounding is sound.
- **Permanent limit (no test possible):** the headline claim itself is settled only by
      watching trained users use the tool — never from the armchair, and never by assay.
      Carried as a hypothesis awaiting field evidence. Recorded here so it is never quietly
      promoted to a finding.

## Layer 4 — Human spot-audit (the irreducible floor)

Not a cop-out; the tool's own thesis. The human is the verifier of last resort, and "catching where
the machine critic is wrong" is the skill assay trains. Sampling is built into the program rather
than treated as failure:

- [ ] **Sampled verdict audit** — each calibration run, sample ~10 verdicts (stratified
      across verdict types, biased toward `substantive` — false passes are the costly
      direction). Trained reader marks agree/disagree + reason. Log to
      `testing/audit_log.jsonl`.
- [ ] **Disagreement triage** — every human-disagree becomes either a new 3b fixture (if
      constructible) or a recorded envelope limit (if not).

## Operating envelope (maintained summary)

Updated as Layer 3/4 results accumulate. No longer "from design analysis, not yet measured": the
haiku 3b/3d calibration (2026-06-10/11) measures several rows. Rows tagged *measured* cite a
calibration run; the remainder are still design expectations.

| Region                                        | Expected reliability               | Basis |
|-----------------------------------------------|------------------------------------|-------|
| Faithfulness with source present              | Highest                            | Reference in hand (STV asymmetry) |
| Refutation of internally contradictory claims | High                               | Reachable by reasoning alone |
| `unverifiable` on predictions/intentions      | High (by design)                   | *Measured:* by design; resolved-prediction case held 10/10, haiku, 2026-06-10 (laundering) |
| Relational defects (equivocation, motte-and-bailey) | Partly dissolved by decomposition before grading | *Measured:* motte calibration — `decompose` splits motte from bailey, so the move is graded fragment-by-fragment (W4) |
| Substance on sweeping forward predictions     | Skews `hollow` (false-attack direction) | *Measured:* laundering + bare-vs-contextualized (2026-06-11) — bare forecast `hollow` 10/10, unmoved by context; killed by **Counterexample**, not an Evidence/decontextualisation penalty |
| Substance verdicts, novel/specialist domains  | Lowest — treat with most suspicion | No reference; correlated blind spots |
| Positive grounding confirmation               | Bounded by search quality; never armchair | Needs the truth-maker |
| Tool's own headline claim                     | Untestable by the tool              | Permanent limit; reflexive canary `unverifiable` 5/5 (haiku, 2026-06-11) shows the boundary holding |

## Disciplines for this file

- Pre-register each test before building it: decision_log entry with hypothesis and alternatives not
  taken; fill in outcome including `abandoned`.
- A populated results table is not success — every fixture's provenance is stated or the entry is
  marked BLOCKED.
- Layer 3 results are read, summarized, and logged; never gated, never averaged into a single score.
- Separate done from advanced: each testing session reports what shipped, whether the envelope map
  moved, and what is still NOT covered.
