# MUTATION_BENCH.md — Pre-registration: crossexam mutation benchmark

Status: RATIFIED 2026-07-15 (Comrade), conditional on the A2.2 pre-flight gates. Thresholds in §6
are no longer proposals: they are struck and replaced in full by Amendment 1 §A1.3.
Repo home: elenchus (crossexam). This document must be committed before any harness code exists (pre-registration precedes implementation).

## 0. Invariants

- I1. Ground truth is the injection manifest, never a model verdict. Each injected defect is its own
  oracle.
- I2. The critic emits findings; the benchmark verdict is computed deterministically from
  findings vs. manifest. No LLM self-grading anywhere in the scoring path.
- I3. Injection is deterministic:
  same (base fixture, defect class, seed, tool version) → byte-identical mutant. Re-runnable.
- I4. Fixtures committed to the elenchus repo are synthetic or fully scrubbed.
  No Solo-private pitch content, customer identity, or commercial numbers in tracked fixtures.
  Solo docs may be benchmarked locally as unversioned inputs only.
- I5. spec/ remains frozen and read-only. The benchmark lives outside it.
- I6. Scoring results are appended to an append-only ledger (bench/RESULTS.md or JSONL);
  prior rows are never edited, only superseded by reference.
- I7. Any run against a pre-decompose-fix build must record that build provenance in the result row.
  SCP-class results from such builds are attributed to harness limitation, not grader capability
  (per calibration-run N=5 finding).

## 1. Purpose

Convert "does crossexam catch defects?" from anecdote into a measured, per-class recall number,
using documents seeded with known defects. Secondary: measure false-positive rate on clean fixtures,
so precision claims about the critic carry a number (Determinism Ledger discipline: model-based
claims carry measurement).

## 2. Defect taxonomy (v1 — seven classes)

Each class defines: what is injected, how injection is mechanized, and what counts as a catch.

| ID  | Class                    | Injection | Catch condition |
|-----|--------------------------|-----------|-----------------|
| FAB | Fabricated grounding     | A claim's citation is rewritten to reference evidence absent from the source set, or a supporting number is altered so it no longer appears in any source | Finding flags the claim ID with a grounding/fabrication class |
| OVR | Over-claim               | Quantifier or scope widened beyond evidence: "measured on corpus X" → "measured", "19/20" → "all", hedges deleted | Finding flags the claim ID with an over-claim/faithfulness class |
| SCP | Scope-shift equivocation | Subject widens mid-argument between premise and conclusion (port of examples/destructive/scope-shift pattern into a generator) | Finding flags the claim pair with a scope/equivocation class |
| ORF | Orphan reference         | A cross-reference is retargeted to a section, artifact, or exhibit that does not exist in the document set | Finding flags the reference with an orphan/missing-referent class |
| CON | Internal contradiction   | A claim is inserted or edited to contradict an earlier committed claim or documented dead end in the same document set | Finding flags either claim ID with a contradiction class, naming the counterpart |
| MFC | Missing falsification criteria | A claim marked testable has its kill condition / criterion deleted | Finding flags the claim ID with a missing-criteria class |
| TIER | Determinism-tier violation | A model-based claim is rewritten in deterministic-tier language (measured/probabilistic result presented as re-runnable proof), or vice versa | Finding flags the claim ID with a tier/vocabulary class |

Notes:
- SCP is the sentinel class for the decompose fix. If SCP recall is 0 on a post-fix build, that is
  a grader finding; on a pre-fix build it is a known harness artifact (I7).
- ORF is the class the ARA paper found LLM auditors weakest on (~22% detection). If v1 confirms low
  ORF recall here, the pre-registered response is to move orphan detection into a deterministic
  structural pre-pass, not to prompt-tune the critic.
- TIER is elenchus-generic in mechanism but exists because of the Solo Determinism Ledger. It
  generalizes: any two-tier claim discipline (proved vs. measured) can instantiate it.

## 3. Fixtures

- F-BASE-n: 3–5 synthetic claims documents in the pitch/argument genre, each 15–40 claims, written
  clean and audited by hand once at creation. Committed under bench/fixtures/.
- Each fixture ships with a claim index: stable claim IDs anchoring every claim, reference, and
  criterion. Injection targets claim IDs, not raw byte offsets, so the matcher is deterministic and
  robust to whitespace.
- One fixture must contain a documented dead end and a testable claim with an explicit kill
  condition, so CON and MFC have targets.
- Clean fixtures are also run through the critic unmutated (§5, false-positive measurement).

## 4. Manifest schema

One JSONL row per mutant, written by the injector at generation time:

```json
{
  "mutant_id": "F-BASE-1.FAB.s17.a3c9",
  "base_fixture": "F-BASE-1",
  "defect_class": "FAB",
  "anchor_claim_id": "C12",
  "counterpart_claim_id": null,
  "seed": 17,
  "injector_version": "mutbench v0.1.0 <commit>",
  "base_sha256": "…",
  "mutant_sha256": "…",
  "note": "citation retargeted from EX-3 to nonexistent EX-9"
}
```

- counterpart_claim_id is set only for CON (the contradicted claim) and SCP (the premise claim).
- One defect per mutant in v1. Multi-defect mutants are a v2 question (interaction effects are real
  but confound per-class recall).

## 5. Scoring (deterministic)

- Match rule: a finding matches a manifest row iff (a) it anchors on anchor_claim_id (or counterpart
  for CON/SCP), and (b) its finding class maps to the defect class under a committed class-mapping
  table (bench/CLASSMAP.md, versioned). No fuzzy text matching. No LLM in the loop.
- Per-class recall: caught / injected, per defect class, per model, per crossexam build.
- False-positive rate: findings emitted on clean fixtures / total clean-fixture runs, reported per class.
- N per cell: minimum 10 mutants per class per run (≥70 mutants total per model×build cell). Below
  that, the row is marked underpowered and cannot trigger §6 responses.
- Every result row records: model, model version, crossexam commit, injector version, decompose-fix
  status, date.

## 6. Pre-registered thresholds (PROPOSED — ratify or amend before first scored run)

Committed before data, per kill-criteria doctrine. These trigger logged responses, not silent tuning.

- T1. FAB recall ≥ 0.90. FAB is the easy class (ARA reports 100% fabrication detection);
  below 0.90 indicates harness or class-mapping defect before grader defect.
  Response: audit matcher and prompts, re-run.
- T2. ORF recall < 0.50 → pre-registered response is deterministic pre-pass for orphan detection
  (structural check, zero LLM), and ORF is thereafter excluded from grader-recall claims.
- T3. SCP recall on a post-decompose-fix build < 0.60 → grader capability gap logged in BEHAVIOR.md;
  scope-shift claims about crossexam capability are suspended until fixed.
- T4. False-positive rate on clean fixtures > 0.15 per document → findings-quality gap logged; no
  capability claims cite recall numbers without the paired FP rate.
- T5. No public claim about crossexam detection capability may cite a number from an underpowered
  cell (§5) or a pre-fix SCP cell.

Ratification authority: Comrade. (Elenchus repo; Matthew's acceptance-authority role is Solo-scoped
and does not extend here.)

## 7. Harness shape (directory level only — package layout to be confirmed in CC research phase)

```
bench/
  fixtures/        # committed clean fixtures + claim indexes
  CLASSMAP.md      # finding-class ↔ defect-class mapping, versioned
  MUTATION_BENCH.md  # this document
  RESULTS.jsonl    # append-only result ledger
mutants/           # generated, gitignored (reproducible from manifest+seed)
```

Injector and scorer are subcommands
(working names: `crossexam mutbench gen`, `crossexam mutbench score`) or a sibling binary — CC
research phase decides against the existing five-package layout and BEHAVIOR.md contracts. Hard
build gate applies (golangci-lint + go test + go build).

## 8. Explicit non-goals (v1)

- No multi-defect mutants.
- No cross-document defects (defects spanning fixture boundaries).
- No prompt optimization loops driven by benchmark score (Goodhart guard: the benchmark measures; it
  does not train).
- No use of Solo pitch content in committed fixtures (I4).

## 9. Two-phase CC prompt (Phase 1 — research, then STOP)

> Read BEHAVIOR.md, the five-package layout, and bench/MUTATION_BENCH.md (this spec). Confirm: (1) current decompose-fix status and how build provenance can be stamped into a result row; (2) where injector/scorer subcommands fit the existing orchestration contracts without touching spec/; (3) whether the current finding output format carries stable claim-ID anchors and a class field sufficient for the §5 match rule — if not, enumerate the minimal contract change; (4) proposed claim-index format for fixtures. Produce a written plan covering these four items plus a task breakdown. Do not write implementation code. STOP after the plan.

Phase 2 (implement) is issued only after the plan is ratified and §6 thresholds are ratified or amended.

---

# MUTATION_BENCH.md — Amendment 1 (append; do not edit §§0–9)

Status: RATIFIED 2026-07-15 (Comrade), conditional on the A2.2 pre-flight gates.
Supersedes by reference: §2 (classes), §5 (scoring), §6 (all thresholds), §7 (harness shape).
Triggered by bench/PHASE1-PLAN.md findings against the as-built binary.

## A1.0 Findings of record

- Substance mode is the only wired mode; it grades a single claim with no source set, no document,
  no cross-claim visibility. FAB, OVR, ORF, CON, SCP-X, TIER are unmeasurable on this build.
  No harness work changes this.
- "Decompose fix" has no referent in the repo. SCP-X recall ≈ 0 is an architectural prediction
  (atomization vs. cross-claim structure), not a pending measurement. Recorded as a standing
  constraint in CRITIQUE.md; re-opened only by a decompose redesign decision.
- Chain findings carry neither stable claim IDs (Idx is positional over nondeterministic decompose
  output) nor typed classes (Axis.Axis is free text). The §5 match rule is undecidable from the chain.
- Axis-level critique fires on nearly every claim; finding-emission FP is ~1.0 by construction. The
  2026-07-12 result (equivocat scan 5/5 certified, human read 0/5) is the canonical instance of the
  laundering risk.

## A1.1 v1 scope (amends §2)

Measurable classes: MFC, SCP-1 (single-sentence equivocation).
Deferred, blocked on faithfulness/document modes: FAB, OVR, ORF, CON, TIER.
Blocked on architecture decision: SCP-X.

SCP-1 admission gate: the Phase 1 plan must contain three authored SCP-1 exemplars, each with a
one-line disjointness argument distinguishing it from OVR. If three clean exemplars cannot be
produced, SCP-1 is struck and v1 ships MFC only.

## A1.2 Metrics (amends §5)

Two metrics, never merged, never averaged:
- axis-fired — deterministic; axis emitted non-empty critique on the mutated claim.
  Explicitly NOT recall. Reported only as a harness diagnostic.
- target-catch — human read; the real recall number. Read protocol: reader sees critique text and
  the manifest row, records catch y/n before seeing the verdict. Protocol violations void the row.

False-positive metric is over-catch, defined on the verdict: clean claim receives a failing/flagged
verdict. Finding-emission FP is not reported.

N ≥ 10 mutants per class per cell stands. Human-read burden at v1 scope: ≤ 20 reads per model×build cell.

## A1.3 Thresholds (replaces §6 in full)

- A-T1. MFC target-catch ≥ 0.80. MFC is the Falsifiability axis's core job on exactly the input
  shape substance mode is built for; below floor is a grader capability gap logged in BEHAVIOR.md.
- A-T2. SCP-1 target-catch ≥ 0.60 (if SCP-1 survives A1.1's gate).
- A-T3. Over-catch on clean fixtures ≤ 0.30. Above ceiling: capability claims suspended; recall
  numbers may not be cited without the paired over-catch rate regardless.
- A-T4. axis-fired is never reported as, aggregated with, or converted into recall, in any document,
  internal or external.
- A-T5. No public claim cites a number from an underpowered cell, an unratified read protocol, or
  any SCP-X cell.
- A-T6 (design commitment, not threshold). When document mode exists, orphan-reference detection is
  implemented as a deterministic structural pre-pass; ORF is never assigned to the grader and never
  enters grader-recall claims.

Struck: T1, T2, T3, T4 of §6 (unrunnable, pre-empted, unreachable, degenerate respectively — see A1.0).

## A1.4 Harness shape (amends §7)

Sibling binary `mutbench` (no subcommand dispatch exists in crossexam; flag.Arg(0) is the input path).
Consumes/produces the chain JSONL via -chain-dir. Result rows stamp git rev-parse HEAD + dirty flag
for both binaries.

Fixture format: claims.jsonl sidecar carries claim IDs and authored injection slots; IDs never
appear inline in critic-visible text. Injection is slot substitution — deterministic, zero-network,
satisfies I3.

## A1.5 Run order

1. Task 1 pre-flight: decompose identity rate on single-claim input, ~20 calls. Gate: rate ≥ 0.90 or
   the anchoring scheme and A1.2 are re-opened before any other task runs.
2. Ratify this amendment (Comrade), conditional on Task 1 result.
3. Phase 2 CC prompt issued for v1 scope only.

## A1.6 Cross-repo correction of record

The Solo §7 DL-ratification constraint previously read "hand-run or use a build post-dating the
decompose fix." No fix exists or is planned; the constraint is now hand-run only. Log as
correction-by-reference in the Solo decision ledger.

---

# MUTATION_BENCH.md — Amendment 2 (append; do not edit §§0–9 or Amendment 1)

Status: RATIFIED 2026-07-15 (Comrade), conditional on the A2.2 pre-flight gates. Amends
A1.1 (exemplar gate outcome),
A1.2 (reporting),
A1.5 (pre-flight gate).
Triggered by the gate defect and exemplars in the Phase 1 follow-up of 2026-07-14.

## A2.0 Findings of record

- A1.5's "decompose identity rate" was underspecified. Cardinality-1 alone admits
  dissolution-by-rewording: decompose can normalize a mutant's narrow scope away without splitting
  the claim, passing the gate while erasing the defect. The surviving fragment is the
  bare-unsupported shape the 2026-07-12 run showed the critic reads as an evidence gap — same
  dissolution, different mechanism, invisible to a cardinality gate.
- MFC is immune to dissolution-by-rewording: its injection deletes the criterion, and decompose
  cannot restore absent input. Second independent basis for MFC as v1's robust class.

## A2.1 SCP-1 admission gate: satisfied (amends A1.1)

Disjointness test adopted as the class boundary, mechanical form: if deleting the narrow-scope
restrictor leaves the defect intact, the defect is OVR; if the defect exists only while both scopes
are present in the sentence, it is SCP-1. Three exemplars (checkout / latency / hiring-funnel
template, single matrix verb, SCOPE as the only moving slot) pass the test and are adopted as the v1
SCP-1 generator templates.

E1 (checkout) carries an embedded arithmetic sub-claim (1:50 < 2:00), retained deliberately for
continuity with SESSION Item A.
Firewall: mutants generated from the E1 template are tagged `confound:arith` and reported as a
separate diagnostic row. They are never pooled into SCP-1 recall. The SCP-1 class number is computed
from frozen-value templates (E2, E3, and any future template with no derivable sub-claim) only.

## A2.2 Pre-flight gate (replaces A1.5 step 1)

Two-part gate, same ~20-call budget per model:

- G1 anchoring: cardinality-1 rate ≥ 0.90 across all exemplar templates.
- G2 defect-survival (SCP-1 templates only): both scopes — narrow restrictor and broad subject —
  present in the returned claim, any wording, rate ≥ 0.90. G2 is a human-read check under the A1.2
  read protocol (recorded before sight of any downstream output).

G1 and G2 are per-model gates: decompose behavior is model-dependent, so the pre-flight runs on
every model v1 will score, with -model pinned explicitly and stamped into the pre-flight record.
A gate result from one model licenses scoring on that model only.

Failure handling: G1 failure re-opens the anchoring scheme and A1.2 before any other task runs.
G2 failure strikes SCP-1; v1 ships MFC only (A1.1's fallback, reached via A2.2).

MFC requires G1 only (A2.0: dissolution-immune).

## A2.3 Commit and epoch

MUTATION_BENCH.md §§0–9, Amendment 1, and this amendment are committed together on a branch per
standing rule; the §I6 append-only ledger epoch begins at that commit. Status lines flip to
RATIFIED in the ratification commit, not before. bench/PHASE1-PLAN.md is committed in the same
branch as the finding-of-record it is.

## A2.4 Cross-references

- A1.6 hand-off prompt (Solo decision ledger correction-by-reference) is dispatched to a Solo session;
it is outside this repo's scope and this ledger records only that it was handed off.

---

# MUTATION_BENCH.md — Ratification record, 2026-07-15 (append; do not edit §§0–9 or Amendments 1–2)

Ratifying authority: Comrade (§6). Recorded as an append per the append-only invariant; no prior
section is edited, and the three status lines carry the flip A2.3 requires.

## R.1 What is ratified

§§0–9, Amendment 1, and Amendment 2 are RATIFIED as of 2026-07-15, **conditional on the A2.2
pre-flight gates passing**. §6's thresholds were struck by A1.3 before ratification and are not
revived by it; the live thresholds are A-T1 … A-T6.

## R.2 v1 scoring model set

**`claude-sonnet-4-6` only.** Haiku is out of v1. A haiku cell may be added later only via a fresh
run under this protocol, never by back-fill or by comparison to the v1 sonnet cell.

Pre-flight set = scoring set (A2.2), so the pre-flight is **single-model**. This retires the
per-model multiplication flagged in bench/PHASE1-PLAN.md §9.2: the pre-flight is ~20 calls and ~20
reads, not ~40.

Consequence to carry forward: PLAN.md §2's calibration baseline and SESSION.md Item B's re-run
condition both specify `claude-haiku-4-5-20251001`. **No v1 number is comparable to those baselines**,
and none may be cited against them. The 2026-07-12 scope-shift datum (N=5, sonnet, hand-run) is the
only prior result on the v1 model, and it is `[~]` — measured, not settled.

## R.3 Transcription byte-check (§9.5, A2.3)

Performed 2026-07-15 against the author's hashes. **PASS**: commit `6211245`, split at the two
section rules, hashes to `1a8ac0b1…` / `13331f6f…` / `9ccf81b0…` at 112/55/35 lines — all three
match the authored originals. Delta at `6211245` = section rules only.

The documents were subsequently re-wrapped in the working tree, which introduced two word-splits
(`c`+`laims` in §1; `harnesswork` in PHASE1-PLAN.md §0). Both were caught by this check and repaired
before this commit. The ratified text is **content-identical** to the authored originals (0 differing
tokens outside markdown table padding) but is **no longer byte-identical** to them. Both hash sets
are recorded in bench/PHASE1-PLAN.md §9.5. Any future byte-check runs against the R.3 baseline, not
the authored one.

## R.4 Standing authorizations and limits

- Phase 2 implementation is authorized for v1 scope only, after the A2.2 gates resolve.
- Scoring runs are **not** authorized. They require the read protocol and separate authorization.
- Nothing is pushed without instruction.

## R.5 Correction of record — status-line condition (2026-07-15)

The three status lines at §§0–9, Amendment 1, and Amendment 2 read "RATIFIED 2026-07-15 (Comrade),
conditional on the A2.2 pre-flight gates." Amendment 3 §A3.3 struck both gates: G1's referent
(per-claim correspondence) no longer exists, and G2's hazard was empirically falsified for
single-sentence input. The condition was not met — G1 recorded 1/8 = 0.12, FAIL — and it was not
discharged; its referent was removed by a later ratified amendment. The ratifications therefore stand
unconditional as of 03d499e.

A reader encountering "RATIFIED … conditional on the A2.2 pre-flight gates" alongside the recorded G1
FAIL must not read the ratification as void or pending: the pre-flight falsified the anchoring
scheme, not the ratification. A3.6 carried all thresholds forward unmodified.

Correction is by reference only. The three status lines are not edited: each was true when written,
and §§0–9 and Amendments 1–2 carry their own instruction not to edit them.

## R.6 Pre-registration — over-catch predicate at cardinality > 1

(2026-07-15, before any read is evaluated). A clean input is over-caught iff any claim derived from
it receives a verdict mapped to failing/flagged. The verdict-type → failing/flagged mapping is
enumerated from the verdict type system in code, committed in this entry, and is not derived from run
data. Anchor-only counting is rejected: the measured unit is the pipeline (A3.1), and a flag on any
fragment of a clean document is a flag the user sees.

### The mapping

Enumerated from `substanceVerdicts` in `internal/claims/claims.go` — the closed set
`ValidSubstanceVerdict` gates on. Substance is the only wired mode, so no other verdict can reach a
v1 chain.

| Verdict | Const | Failing/flagged | Basis |
|---|---|---|---|
| `substantive` | `claims.Substantive` | **No** | The claim stands as written. This is the survival bar. |
| `partial` | `claims.Partial` | **Yes** | The claim did **not** survive as stated; the critic replaced it with a narrower `surviving_claim`. The user's claim comes back rewritten, which is a flag they see. |
| `hollow` | `claims.Hollow` | **Yes** | No defensible core. |
| `error` | `claims.Error` | **Excluded from the denominator** | Not a judgement about the claim — a transport or parse failure. |

### Basis, from the type system and prior doc precedent — not from run data

`partial` is the only live question, and two prior entries already treat `substantive` as the
survival bar and group `partial` with `hollow` as non-survival:

- `examples/destructive/scope-shift/EXPECTED.md`: "The success condition requires the honest narrow
  core to **survive**", and "A grade of `hollow`/`partial` is therefore attributable to the
  scope-shift and nothing else."
- SESSION.md, 2026-07-12: "the company-wide claim never survived `substantive` (5/5)."

`error`'s exclusion follows PLAN.md §2's standing rule that "`error` verdicts are excluded from
n_valid". It is a definitional loose end rather than a live one: the 2026-07-15 run recorded zero.

### Disclosure

CC enumerated this mapping and had prior sight of one clean-twin verdict (`M01.clean`, reported
while verifying the run completed). The mapping is derived from the type definitions and the two
precedents cited above, and no verdict distribution over the 42 chains was computed before this entry
was committed. The prior sight is recorded here so R.6's "not derived from run data" can be read
against what CC had actually seen, rather than asserted.

---

# MUTATION_BENCH.md — Amendment 3 (append; do not edit §§0–9, A1, A2)

Status: RATIFIED 2026-07-15 (Comrade). Amends A1.2 (metrics/read protocol), A2.2 (pre-flight gates, struck), A1.3/A-T5 (claim wording). Triggered by the pre-flight of 2026-07-15 (G1 = 1/8 = 0.12, FAIL; scheme re-opened per A2.2).

## A3.0 Findings of record

- Anchoring falsified: decompose re-predicates single-sentence input into 3–5 claims; no clause boundary is required. Claim-position anchoring is dead. The lone G1 pass was the MFC mutant (bare clause, nothing left to split).
- Dissolution prediction falsified for single-sentence input: decompose did not erase the scope transfer; it isolated the transfer as its own claim (E2 [2], E3 [2]). E3 named the equivocation unprompted, violating its "Do not evaluate" prompt contract. The 2026-07-12 attribution (decompose dissolves cross-claim structure) stands for cross-claim input only. Correction-by-reference to be logged in CRITIQUE.md.
- Design defect of record: G1 gated the deterministic match rule of §5, which A1.2 had already replaced with human target-catch. The gate protected an abandoned scheme; ~2/3 of pre-flight spend was attributable to this.
- Component-attribution hazard: decompose is a model component that sometimes emits meta-level findings inside a step the scoring path treated as preprocessing. Any per-component recall claim on this architecture is confounded.

## A3.1 Unit of measurement (amends A1.2, A-T5)

The measured unit is the pipeline: decompose → producer → critic → verdict, at a stamped (commit, model). No result row, ledger entry, or external claim attributes recall to "the critic" or any single component. Rows name the pipeline.

Component attribution is a recorded diagnostic only: the reader tags catch-source ∈ {decompose, critic, verdict} per catch. Diagnostic rows never aggregate into a component recall number.

## A3.2 Anchoring (replaces the PHASE1-PLAN §3(a) scheme)

Anchor is the input claim ID in the sidecar (claims.jsonl). Manifest rows map to full-chain outputs per input; the reader performs the mapping by reading. Decompose output cardinality is recorded per input as a diagnostic, ungated. axis-fired is redefined per input: any axis emitted non-empty critique on any claim derived from input i. Still deterministic; still never recall (A-T4 unchanged).

## A3.3 Gates struck (amends A2.2)

G1 and G2 are struck. G1's referent (per-claim correspondence) no longer exists. G2's hazard (defect erased pre-grader, measuring a foregone conclusion) was empirically falsified for single-sentence input by the pre-flight; residual dissolution is henceforth inside the measurement — a dissolved defect that no chain text identifies is a system miss, scored as such.

No pre-flight gate blocks Phase 2. The pre-flight record (25 calls, G1 = 0.12, isolation finding) is retained in the ledger as the falsification event for the prior scheme.

## A3.4 Read protocol (amends A1.2)

Per input: reader sees the full chain (decompose output, critic critique, verdict) plus the manifest row, in randomized input order, and records before any pooled results: catch y/n; catch-source; free note.

Catch standard: chain text must identify the specific defect. Neutral re-predication of the inferential move is NOT a catch; meta-level naming of the move IS. Calibration boundary pair, verbatim in the protocol:
- NOT a catch — E2 [2]: "The Frankfurt replica's p99 latency determines the service's p99 latency."
- Catch — E3 [2]: "The author equates the pass rate among onsite candidates with the hiring funnel's overall pass rate."

The read sheet of 2026-07-15 (bench/preflight/READ-SHEET-2026-07-15.md) is VOID — it poses the struck G2 question against decompose output only. Reads were not spent; the sealed key is discarded unopened.

## A3.5 Decompose prompt-contract tension (record, not action)

E3 [2] violates decompose's "Do not evaluate" instruction. The prompt is NOT to be tightened for v1: suppressing meta-level output would raise component purity by lowering pipeline recall — tuning the tool against its users to clean up a measurement. Recorded in CRITIQUE.md as a standing architectural fact; BEHAVIOR.md must eventually either license decompose findings or a redesign must remove them. Out of v1 scope.

## A3.6 Thresholds and non-tuning declaration

A-T1 (MFC ≥ 0.80), A-T2 (SCP-1 ≥ 0.60), A-T3 (over-catch ≤ 0.30) carry over unchanged as pipeline numbers. Declaration of record: these were ratified before the pre-flight, whose isolation finding is weakly favorable to SCP-1 recall; they are retained unmodified so no later reading can construe them as tuned to the signal.

## A3.7 Explicitly not built (v1)

- Decompose-only probe / entry point: serves a struck question. Not built.
- Any per-component recall reporting path.
Phase 2 may parallelize pipeline calls if the change is trivial; no other harness scope is added by this amendment.

