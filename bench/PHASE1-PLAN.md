# PHASE1-PLAN.md — research findings for the mutation benchmark

Status: RESEARCH OUTPUT. No implementation code written. Answers the four questions in
MUTATION_BENCH.md §9, then proposes a task breakdown. Phase 2 stays blocked until §6 thresholds
are amended (see "Threshold consequences") and MUTATION_BENCH.md is committed.

Every claim below cites the file and line it came from. Where I could not confirm something, I say so.

---

## 0. Headline

**Five of the seven defect classes cannot be scored against this build, and no amount of harness
work changes that.**
Substance mode is the only mode wired (`cmd/crossexam/main.go:73-75` makes
`-source`/`-evidence`/`-audit`/`-md` a fatal error). Substance mode grades
**one claim string at a time, with no source document and no sight of any other claim** —
`AssayClaim(c, claim, maxRounds)`
takes a single claim and nothing else (`internal/modes/substance.go:56`), and the producer and
critic prompts receive only that claim's text (`substance.go:64-76`).

So:

| Class | What it needs to be detectable                       | Available here? |
|-------|------------------------------------------------------|-----------------|
| FAB   | the source set, to know a citation is absent from it | No — needs faithfulness mode (`-source`), a fatal stub |
| OVR   | the evidence, to know the quantifier went beyond it  | No — same |
| SCP (as specced: premise→conclusion)                         | two claims seen together | No — see §1 |
| ORF   | the document set, to know a referent is missing      | No — no wired mode sees a document |
| CON   | an earlier claim, to contradict against              | No — the critic never sees another claim |
| **MFC** | the claim alone                                    | **Yes** — this is the Falsifiability axis |
| TIER  | the deterministic/model-based tier convention        | No — the prompts encode no such convention |

A benchmark that runs FAB against this build measures nothing about the grader. It measures that
substance mode was never given the source. That is a foregone conclusion, and it costs API calls to
re-derive.

The spec was written against README.md's usage block. CLAUDE.md already flags that trap: the README
"is the eventual contract, not working software."

---

## 1. Question 1 — decompose-fix status

**There is no decompose fix. There is no plan for one. The phrase names something that does not exist.**

Evidence:

- `grep -rn -i "decompose.fix|decompose fix"` across the repo returns two hits, both in
  `examples/destructive/scope-shift/EXPECTED.md` (lines 58, 68), and both describe the limitation as
  **"not fixed"**. Nothing anywhere describes a fix.
- PLAN.md §1 (the build sequence) contains no decompose change.
- PLAN.md §3 (the design queue, items (a)–(e)) contains no decompose change. The closest item, (e),
  is a *new mode* for source-coverage traversal, not a decompose change.
- SESSION.md carry-forward Item A (lines 30-38) is a **fixture**, not a code fix: a scope-shift
  written as one indivisible sentence, built specifically to *route around* decompose so the critic
  can be tested in isolation. It changes no code. SESSION.md calls the existing probe its
  "harness-level twin" and says the pairing is "deliberate, not redundant."

MUTATION_BENCH.md §I7 and §T3 both turn on a "pre-fix" / "post-fix" build distinction. That
distinction has no referent today and none scheduled, so **T3 can never fire** — there will never be
a "post-decompose-fix build" to measure.

**The deeper problem, which I'd flag as the main design finding:** SCP as defined in §2 is a
*cross-claim* defect (the manifest sets `counterpart_claim_id` to the premise claim). Decompose's
entire job is to split prose into atomic, independently-evaluable claims — that is the first line of
its prompt (`internal/modes/prompts.go:11-14`). A "fix" that preserves cross-claim moves is in
tension with atomism itself, which is CRITIQUE.md's known atomism gap, not a bug with a patch.
SCP recall ≈ 0 on this architecture is a **prediction of the design**, not a measurement waiting to
be taken — and the 2026-07-12 run already took it (0/5, SESSION.md:15-21).

**Recommendation.** Split SCP into the two classes the fixture pair already implies:

- **SCP-X** (cross-claim, premise→conclusion): tests decompose, predicted recall 0, **out of v1**.
  It is the sentinel for a future decompose design item, and it should be *entered in PLAN.md §3*
  rather than measured repeatedly.
- **SCP-1** (single indivisible sentence, per SESSION.md Item A): tests the *critic*, is measurable
  today, and is the question 2026-07-12 left open. **In v1.**

Then amend T3 to bind to SCP-1, where it is reachable.

### Build provenance stamping

**Nothing stamps a version today.** `build.sh` has no `-ldflags` (I read the whole file); there is
no version variable in any Go file (`grep -rn "ldflags|Version"` finds only unrelated hits:
`anthropic-version` HTTP headers in `internal/client/client.go:25,167`).

Minimal addition, in the benchmark harness rather than in `crossexam`: record at scoring time
`git rev-parse HEAD` plus a **dirty-tree flag** from `git status --porcelain`. The dirty flag is not
optional — a result row from a modified tree is not reproducible, and §I3/§I6 are worthless without
it. Alternative (cleaner, slightly more invasive): stamp a `version` variable via
`-ldflags -X` in `build.sh` and have the harness read it back. Either works; the harness-side
version needs no change to the shipped binary, so I lean that way.

**Also unpinned:** the default model is `claude-sonnet-4-6` (`main.go:60`), while PLAN.md §2's
calibration baseline is `claude-haiku-4-5-20251001` and SESSION.md Item B's re-run condition
specifies haiku for comparability. A benchmark run that does not pass `-model` explicitly silently
gets sonnet. The harness must pin the model, never inherit the default.

---

## 2. Question 2 — where injector/scorer fit

**Recommendation: a sibling binary (`cmd/mutbench`), not `crossexam` subcommands.**

The subcommand form in §7 collides with the existing CLI contract. `crossexam` has no subcommand
dispatch: `flag.Parse()` runs, then `flag.Arg(0)` is read as the **input file path**
(`main.go:69,81-85`). So `crossexam mutbench gen` would try to read a file named `mutbench`, and
adding dispatch would change the input-resolution contract that spec/CLI.md freezes (`-text` > file
arg > piped stdin > built-in default). spec/ is frozen; that change is a design-queue item, not a
port-time edit.

Three further reasons the sibling binary is the better fit:

1. **The chain file is already the seam.** `-chain-dir` exists (`main.go:68`) and `writeChain`
   persists one JSONL record per claim (`cmd/crossexam/chain.go:48-85`). The scorer reads that file.
   It needs no crossexam internals.
2. **`gen` and `score` are pure file→file.** No network, no API key, no clock, no randomness beyond
   the manifest seed. They need no test fake, so CLAUDE.md's "the fake lives in `internal/client`
   only" rule is untouched.
3. **The Goodhart guard in §8 becomes structural rather than a promise.** Keeping the benchmark out
   of the shipped binary means the graded pipeline cannot reach the benchmark.

Layout: `cmd/mutbench` (flags + wiring) and `internal/mutbench` (injector, scorer, manifest types).
`bench/` stays data-and-docs only, per §7. spec/ is untouched (§I5 holds).

---

## 3. Question 3 — does the finding output support the §5 match rule?

**No, on both halves. §5 as written is not satisfiable today.** This is the largest contract gap.

### (a) There is no stable claim ID, and one cannot be threaded through decompose

- `claims.Substance{Claim, Verdict, Reason, Detail}` — `Claim` is the claim **text**, not an ID
  (`internal/claims/claims.go:66-72`).
- `chainRecord{Idx, Total, Mode, Claim, Verdict, ...}` — the only anchor is `Idx`, a **position in
  the decompose output** (`chain.go:18-26,66-69`).
- `Decompose` returns `[]string` (`substance.go:43-49`). It is a **model call**. It is neither
  deterministic nor identity-preserving: the 2026-07-12 run split one passage into **5–7 fragments,
  varying run to run on the same input** (`EXPECTED.md`, "Observed" section).

So `Idx` is not stable across runs of the same input, and a fixture claim ID cannot survive
decompose. There is no deterministic key from a finding back to `anchor_claim_id`. §5's "No fuzzy
text matching. No LLM in the loop" cannot be met while decompose is in the path.

**Recommendation: invoke crossexam once per claim, bypassing decompose's splitting**, so the
anchor is 1:1 by construction. Two things make this cheap rather than a compromise:

- Substance mode **already grades each claim in isolation** — `AssayClaim` takes one claim and no
  context. Per-claim invocation is what the pipeline does anyway, minus the split. It distorts
  nothing.
- It collapses cost. Only the anchor claim needs grading per mutant (~2 API calls per round,
  `maxRounds=2`), instead of all 15–40 claims.

The cost is real and must be recorded in every result row: **v1 measures the critic, not the
decompose→critic pipeline.** I'd add a `stage_under_test: "critic"` field to the manifest and the
result row so no reader can take a v1 number as a pipeline number.

This needs one pre-flight check before anything is built — see Task 1. Decompose still runs on the
single claim (it is unconditional, `main.go:110`), so I need to measure how often it returns that
claim unchanged. If that rate is not ~1.0, this anchoring approach fails and Phase 2 needs redesign.
Better to spend ~20 API calls learning that now.

### (b) There is no class field

The critic emits `Axis{Axis, Finding, Severity}` where `Axis` is **free text straight from the
model** (`claims.go:50-54`). Nothing validates it. The seven axis names appear only inside the
prompt prose (`prompts.go:23-31`).

Contrast: verdicts *are* a closed typed vocabulary with boundary validation
(`type Verdict string` + consts + `ValidSubstanceVerdict`, `claims.go:23-83`). Axes are the same
kind of thing — a closed set the code needs to match on — and CLAUDE.md's structural rule already
covers them. Note this is independently justified: the repo's own degradation log has a 2026-06-14
entry for exactly this failure with verdict literals. **Minimal contract change:**
`type AxisName string` + seven named consts + a validator, in `internal/claims`.

One caution: the model returns free text, so strict validation means a malformed axis becomes an
`error` verdict, changing run behaviour for a benchmark's benefit. I'd rather **normalize and record
whether the axis was recognized** than hard-fail, and raise hard validation as its own consolidation
item. Flagging the choice rather than making it.

### (c) The catch conditions are not decidable from the chain — and this is where the benchmark
### risks the exact failure crossexam exists to catch

§2's catch conditions require the finding to **name** the defect ("naming the counterpart" for CON,
"flags the claim pair" for SCP). The chain carries a free-text `finding` sentence. Deciding "names
the counterpart" from free text is fuzzy matching or an LLM read — both banned by §5/§I2.

This is not hypothetical. It is the documented lesson of 2026-07-12 (SESSION.md:23-26):

> `run.sh`'s `equivocat` substring scan matched 5/5 … but never the subject shift; the per-run human
> read showed 0/5 target catch.

The automated scan certified a catch that did not happen. An axis fires on nearly every claim — PLAN
§3(b) records that an *empty* critique is so rare it is the signature of a false pass. So
"an axis fired" is the norm, and scoring recall on axis-firing measures almost nothing.

If the benchmark reports axis-firing as recall, it spends a win on the question it can reach
(did an axis fire) as authority on the one it cannot (did the critic name the defect). That is
confidence laundering — the target failure named in the first line of CLAUDE.md.

**Recommendation: two metrics, computed separately, never merged, never summed into one recall
number.**

- **axis-fired** — deterministic, from `(axis, severity)` in the chain. Cheap, automatable, and
  explicitly *not* recall. It is a necessary condition for a catch, not a catch.
- **target-catch** — the human read of the persisted critique, at small N. This is the real number.
  EXPECTED.md already calls the human read "load-bearing, not ceremony."

Naming: do **not** call these tier-1/tier-2. "Tier-2 chain" already means the JSONL chain
(spec/CLI.md), and TIER is a defect class name in §2. Three unrelated meanings of "tier" is exactly
the name-collision CLAUDE.md bans. Use "axis-fired" and "target-catch".

This means §5's "N per cell: minimum 10 mutants" is affordable for axis-fired and expensive for
target-catch. State the N for each separately.

### (d) The false-positive rate in §5 is defined so that it is always ~1.0

§5 defines it as "findings emitted on clean fixtures / total clean-fixture runs." The critic emits
findings on essentially every claim (PLAN §3(b), above). So this metric is ~1.0 by construction and
carries no information.

**Recommendation:** define the false positive on the **verdict**, not on finding-emission — a clean,
well-formed claim graded `hollow` or `partial`. That is the measurable harm, and it is what
EXPECTED.md already calls an **over-catch**:

> a run that kills the flagship claim is **not** a clean pass even though the company claim also fell.

Note there is a known confound to pre-register: PLAN §3(c) records that decontextualized fragments
skew hollow (the referent-anchoring gap). Clean-fixture claims graded in isolation will inherit that
skew, so the v1 false-positive number measures the critic-in-isolation, consistent with (a).

---

## 4. Question 4 — proposed claim-index format

Given per-claim invocation, the v1 fixture **is** the claim index. §3's prose document buys nothing
measurable in v1 (nothing reads a document), so prose is deferred to v2 along with decompose and
faithfulness.

`bench/fixtures/F-BASE-1/claims.jsonl`, one row per claim:

```json
{
  "id": "C12",
  "text": "The Oxford Street flagship's median checkout time was 1 minute 50 seconds last week.",
  "clean": true,
  "slots": {
    "criterion": {"span": [72, 118]},
    "subject": {"narrow": "the Oxford Street flagship", "broad": "the company"}
  }
}
```

Design notes:

- **IDs live in the sidecar structure, never inline in `text`.** An inline `[C12]` marker is text the
  critic would see and could comment on — it would contaminate the measurement.
- **`slots` are authored injection sites.** This keeps injection **mechanical and template-driven**,
  so §I3 (byte-identical re-generation) holds with zero network and zero LLM. An LLM-written mutant
  would be more natural but would break I3 unless cached.
- **Validity caveat to pre-register:** mechanically-injected defects are cleaner than organic ones,
  so recall on them is an **upper bound** on real-document recall. Worth stating in the ledger rather
  than discovering later.
- `base_sha256` / `mutant_sha256` in §4 then hash `claims.jsonl`, not a prose document.

**SCP-1 gets its clean twin for free**, which is worth noticing: the clean fixture is the
subject-pinned claim ("the flagship is fast"), the mutant is the widened one ("the company is fast").
So the over-catch measurement (§3(d)) and the recall measurement are the same fixture pair. The
injection is a deterministic subject swap across the `subject` slot — exactly SESSION.md Item A's
shape.

**MFC is the flagship class for v1** and the one class where the deterministic scorer is trustworthy,
because the defect *is* an axis: injection deletes the `criterion` span, and the catch is the
Falsifiability axis firing fatal. Here "axis-fired" and "target-catch" nearly coincide — which is
precisely why they must stay separate metrics for the classes where they do not.

---

## 5. Threshold consequences (§6 must be amended before Phase 2)

> **SUPERSEDED by MUTATION_BENCH.md Amendment 1 §A1.3**, which strikes T1–T4 and replaces §6 in
> full. This section is kept unedited as the record of why. See §9 for the Amendment 1 responses.

| Threshold                              | Status                                                        | Proposed amendment |
|----------------------------------------|---------------------------------------------------------------|--------------------|
| T1 FAB ≥ 0.90                          | **Unrunnable** — FAB needs faithfulness mode, a fatal stub    | Defer to when `-source` is wired |
| T2 ORF < 0.50 → deterministic pre-pass | **Skip the measurement** — ORF recall is 0 by architecture (no wired mode sees a document). The pre-registered response is already the right design | Build the structural pre-pass directly; it needs no LLM and no crossexam mode. Never route ORF through the critic |
| T3 SCP post-fix < 0.60                 | **Unreachable** — no fix exists or is planned (§1)            | Rebind to SCP-1 (critic in isolation), where it is measurable now |
| T4 FP > 0.15/doc                       | **Runnable, but the metric is degenerate as defined** (§3(d)) | Redefine on verdict (over-catch), not finding-emission. Denominator per claim, not per document — §5 reports per class, §6 says per document |
| T5 no claims from underpowered/pre-fix cells | **Keep, and it now bites harder**                       | Every v1 cell is `stage_under_test: critic`. No v1 number may be cited as a pipeline number |

Also: §I7 should be rewritten. It attributes SCP results from "pre-fix builds" to harness limitation,
which is right in substance but names a build distinction that does not exist. The real rule is:
**SCP-X measures decompose; SCP-1 measures the critic; neither substitutes for the other.**

---

## 6. Proposed v1 scope

**Two classes: MFC and SCP-1**, plus over-catch on their clean twins. That is what substance mode can
be honestly graded on today.

Deferred, with a reason each, not silently dropped:

- FAB, OVR → blocked on faithfulness mode (`-source`). Note `claims.go:30-34` already carries the
  `overstated` verdict const, so OVR has a home waiting.
- ORF → goes to the deterministic structural pre-pass (T2's own pre-registered response).
- CON, SCP-X → need a mode that sees more than one claim. No such mode exists; PLAN §3(e) is the
  closest precedent and is explicitly "a new mode, not a flag on the existing one."
- TIER → the critic has no concept of the tier vocabulary. Needs a prompt or mode that encodes the
  convention, which is a PLAN §3 design item (prompts are verbatim from spec/PROMPTS.md; a
  deliberate change is never a port-time edit).

This is a large cut from seven classes to two. I'd rather deliver two numbers that mean something
than seven where five re-measure a foregone conclusion.

---

## 7. Task breakdown

Ordered. Each task is its own session; feature and consolidation never mix (CLAUDE.md).

| # | Task | Type | Blocks on |
|---|------|------|-----------|
| 0 | Commit MUTATION_BENCH.md + this plan under `bench/`. Amend §6 per §5 above. Ratify. | doc | Comrade |
| 1 | **Pre-flight: decompose identity rate on single-claim input.** ~20 API calls, hand-run, N≥10 on the pinned model. Measures how often `Decompose` returns a single atomic claim unchanged. **If not ~1.0, §3(a)'s anchoring fails and Phase 2 is redesigned, not patched.** Pre-register the threshold before running. | probe | 0 |
| 2 | Author `F-BASE-1` fixture: MFC claims (criterion slot) + the SCP-1 pair (subject slot). Hand-audit once at creation (§3). | feature | 1 |
| 3 | `internal/claims`: `AxisName` typed vocabulary + validator, normalize-and-record (§3(b)). Independently justified by CLAUDE.md's structural rule. | consolidation | — |
| 4 | `internal/mutbench` injector + manifest writer. Mechanical, seeded, no network. Red-then-green. | feature | 2, 3 |
| 5 | `bench/CLASSMAP.md` — strict axis↔class map. **Pre-register before any scored run; strict (target axis only), never permissive.** A permissive map re-creates the vacuous certificate. | doc | 2 |
| 6 | `internal/mutbench` scorer: reads chain JSONL + manifest → axis-fired metric. Pure, deterministic, unit-testable with no fake. | feature | 4, 5 |
| 7 | `cmd/mutbench` wiring: `gen` / `score`. Pins model + records git SHA and dirty flag (§1). | feature | 6 |
| 8 | First scored run + human target-catch read. Append to `bench/RESULTS.jsonl` (append-only, §I6). | probe | 7, 0 |

Task 1 is deliberately first and deliberately cheap. It can invalidate tasks 2–8, and it costs about
twenty API calls to find out.

---

## 8. Where this plan might be wrong

- **The decompose identity rate is unmeasured.** §3(a)'s whole anchoring scheme rests on it being
  ~1.0. I have not run it. That is Task 1, and it is the plan's single largest assumption.
- **I did not read `spec/CLI.md` in full** (254 lines) — only the parts `chain.go` and `main.go`
  cite. If it specifies a claim-ID field I did not see, §3(a) softens. I checked the built code, and
  the built code has no such field.
- **"Recall on two classes" may be too thin to be worth the harness.** If Comrade's goal is a
  headline detection number across a taxonomy, v1 cannot produce it, and the honest answer may be to
  wire faithfulness mode first and benchmark second. That is a scope call, not a technical one.
- **The axis vocabulary may not be closed in practice.** I have not sampled real runs to see how far
  the model's `axis` strings drift from the seven prompt names. Task 3 should measure the drift
  before choosing normalize-vs-validate.

---

## 9. Amendment 1 responses

Answers A1.1's SCP-1 admission gate, and raises one defect in A1.5's Task 1 gate that must be
settled before the pre-flight runs.

### 9.1 SCP-1 admission gate — three authored exemplars

**Disjointness criterion, stated once and applied to each.** OVR **removes the scope restrictor or
moves the quantifier**: the narrow scope vanishes from the sentence and the predicate strengthens
("measured on corpus X" → "measured"; "19/20" → "all"; hedge deleted). SCP-1 **keeps both scopes
visible and freezes the predicate**: the statistic, its value, and its hedging are identical on both
sides of the transfer, and the only defect is the asserted transfer from the measured scope to the
wider subject. So the test is mechanical, not a judgement call: *if you can delete the restrictor and
still have the defect, it was OVR; if both scopes must stay in the sentence for the defect to exist
at all, it is SCP-1.*

All three share one template, which makes injection a pure slot substitution per A1.4:

> `The {VALUE} {STATISTIC} measured {PREP} {SCOPE} puts {BROAD_SUBJECT}'s {STATISTIC} {VALUE}.`

The `SCOPE` slot is the only thing that moves. Clean twin = a scope that already covers the broad
subject (transfer sound). Mutant = a scope that is a proper part of it (transfer unsound). The matrix
verb *puts* carries the sentence's only assertion, so there is no clause boundary for `decompose` to
cut — which is the whole point of the single-sentence shape (SESSION.md Item A).

**E1 — checkout time** (continuity with SESSION.md Item A's stated shape)

- Mutant: *The 1-minute-50 median checkout time measured at **our Oxford Street flagship** puts the company's median checkout time under two minutes.*
- Clean twin: `SCOPE` = "all 32 of our stores".
- Disjointness: the restrictor "measured at our Oxford Street flagship" stays in the sentence and the two-minute threshold is identical on both sides; OVR of the same base would delete that restrictor or push "median" to "every shopper's".
- **Known confound, flagged not hidden:** this one embeds an arithmetic step (1:50 < 2:00). On 2026-07-12 *all five* `substantive` verdicts were that arithmetic fragment (EXPECTED.md, "Observed"). E2 and E3 are built to avoid it, so the confound stays isolated to E1.
- **Firewall (A2.1):** E1-template mutants are tagged `confound:arith`, reported as a separate
  diagnostic row, and **never pooled into SCP-1 recall**. The class number comes from frozen-value
  templates (E2, E3) only.

**E2 — service latency** (confound-free)

- Mutant: *The 40-millisecond p99 latency measured on **the Frankfurt replica** puts our service's p99 latency at 40 milliseconds.*
- Clean twin: `SCOPE` = "all 14 production replicas".
- Disjointness: both the measured scope and the asserted subject remain, and the statistic and value are identical across the transfer (p99, 40 ms — no comparison, no arithmetic); OVR would drop "measured on the Frankfurt replica" or raise p99 to "never exceeds".

**E3 — hiring funnel** (confound-free)

- Mutant: *The 92% pass rate measured among **the 40 candidates who reached onsite** puts our hiring funnel's pass rate at 92%.*
- Clean twin: `SCOPE` = "all 1,200 candidates who applied".
- Disjointness: the cohort restrictor stays put and 92% never moves; OVR would delete "who reached onsite" or strengthen 92% to "nearly all".

E2 and E3 freeze the value **identically** on both sides (40 = 40, 92 = 92), so no arithmetic
sub-claim exists to split off or to be graded `substantive` on its own. That is a deliberate
improvement on the 2026-07-12 fixture, not a restyling.

**Gate verdict: three clean exemplars produced. SCP-1 is admitted, subject to 9.2's G2.**

### 9.2 Defect in A1.5's Task 1 gate — "identity rate" is under-specified and passes a silent failure

A1.5 gates on "decompose identity rate ≥ 0.90" without saying what identity means. Both readings are
wrong, in different directions:

- **Byte-identity is the wrong measure and would fail spuriously.** `decompose`'s prompt instructs it
  to "Strip rhetoric, hedges, and connective filler" (`internal/modes/prompts.go:11-14`). It rewords
  **by design**. Byte-identity could be near 0 while anchoring is perfectly fine.
- **Cardinality-1 is the right anchoring measure but passes a silent dissolution.** What anchoring
  needs is that `decompose` returns exactly one claim — then the mutant → graded-claim map is 1:1
  regardless of wording. But `decompose` can also **normalize the mutant's subject away without
  splitting it**: rewrite E2 to "Our service's p99 latency is 40 milliseconds", dropping the
  Frankfurt scope. Cardinality 1 ✓, gate passes ✓, defect gone ✗ — and every SCP-1 mutant silently
  becomes the bare-unsupported fragment that 2026-07-12 already showed the critic reads as an
  evidence gap. That is the same dissolution as the 2026-07-12 split, by a different mechanism, and
  **the gate as written cannot see it.**

**Two-part gate, both measured by the same ~20 calls. Ratified as A2.2:**

- **G1 — cardinality-1 rate ≥ 0.90.** `Decompose` returns exactly one claim. This is the anchoring
  requirement. Applies to all exemplars.
- **G2 — defect-survival rate ≥ 0.90, SCP-1 only.** Both scopes still present in the single returned
  claim. This is the validity requirement. **If G2 fails, SCP-1 is struck and v1 ships MFC only** —
  which is already A1.1's own fallback, reached by a different route.

**A2.2 tightened both beyond what I proposed, in two ways that change the work:**

1. **G2 is a human read** under the A1.2 protocol, not the substring proxy floated below. That is the
   stricter call and it costs reads, but it is consistent: a substring check would ask whether the
   authored tokens survived, when the question is whether *the defect* survived. A paraphrase can
   keep both tokens and still dissolve the transfer.
2. **G1 and G2 are per-model gates.** `decompose` behaviour is model-dependent, so the pre-flight
   runs on **every model v1 will score**, `-model` pinned and stamped. A gate result licenses
   scoring on that model alone. This multiplies the pre-flight by the number of models — at two
   models (haiku for baseline comparability, sonnet as the default) that is ~40 calls and ~40 reads,
   not ~20. Worth pricing before Task 1 rather than discovering at Task 8.

**MFC needs no G2 and that is a structural fact worth recording:** MFC's injection *deletes* the
criterion, and `decompose` cannot restore a criterion that is not in its input. MFC is immune to
dissolution-by-rewording. This is a second, independent reason MFC is v1's robust class.

One note on method, to pre-empt an objection: G2 checks the harness's **own authored input** survived
a rewrite. It is not the banned fuzzy matching — A1.2/§5 ban fuzzy-matching *findings* to the
manifest, which is a claim about the grader. G2 is a claim about the harness, checked against strings
we authored, and it does not touch the critique text. (A2.2 made G2 a human read anyway, so the
objection is moot; the note stands as the record of why a deterministic G2 was available and not
taken.)

### 9.3 Consequential edits to §7's task table

- **Task 1 now depends on 9.1's exemplars** (G2 must run against real SCP-1 mutants, not generic
  claims). §9.1 delivers them, so Task 1 is unblocked now; no reordering needed.
- **Task 1's gate is 9.2's G1 + G2** (A2.2), not "identity rate ≥ 0.90", and **runs once per model
  v1 will score**, not once.
- **Task 4's injector gains the `confound:arith` tag** (A2.1) and Task 6's scorer must keep
  E1-template rows out of the SCP-1 pool and in their own diagnostic row. Cheaper to build in at
  Task 4 than to retrofit at Task 8.
- **Task 5 (`CLASSMAP.md`) shrinks to two classes** and, under A1.2/A-T4, now feeds only the
  axis-fired diagnostic. It no longer sits on the recall path at all — target-catch is a human read.
  It stays pre-registered and strict regardless.
- **Task 8 gains the A1.2 read protocol**: critique text + manifest row, catch recorded y/n *before*
  the verdict is seen, protocol violations void the row. That ordering needs the scorer to emit a
  verdict-suppressed read sheet — a small addition to Task 6, noted here so it is not discovered
  during Task 8.

### 9.4 A1.6 is cross-repo and cannot be done from this session

A1.6 directs a correction into the **Solo decision ledger**. This session owns `elenchus` only
(CLAUDE.md repo boundaries: never edit another repo, produce a hand-off prompt instead). The
hand-off prompt is in the session response, not committed here.

### 9.5 Provenance of `bench/MUTATION_BENCH.md` — transcribed, not fetched

`MUTATION_BENCH.md` existed nowhere on this machine and in no git branch or history when A2.3
directed it to be committed (searched 2026-07-14: filesystem-wide `find`, `git log --all`). The
committed file is **transcribed from the three chat messages that authored it** — §§0–9, Amendment 1,
Amendment 2 — with the `---` rules between them added as the only editorial addition. The author is
the authoritative source, so this is not a fetch-or-STOP substitution; it is dictation.

It does mean **the epoch commit's byte-fidelity rests on a transcription no one has diffed against
an original.** A pre-registration document's entire value is that it says later what it said before,
so this is worth one byte-check by the author against their source before the ratification commit
lands. If no original exists outside the chat, this commit *is* the original, and that fact belongs
in the record rather than in an assumption.

**Byte-check performed 2026-07-15 against the author's hashes. Result: PASS.** Commit `6211245`'s
`bench/MUTATION_BENCH.md`, split at the two section rules and each part hashed:

| Authored original | sha256 | Lines |
|---|---|---|
| `MUTATION_BENCH.md` (§§0–9) | `1a8ac0b1bc599acc7abae1d649af90d4d2a09de63fb9c197b6a3eeb674084f4f` | 112 |
| `MUTATION_BENCH-AMENDMENT-1.md` | `13331f6fef180fd2bdad7e9a1bbe9df7b158e01cee4477b1950cdf65303a6956` | 55 |
| `MUTATION_BENCH-AMENDMENT-2.md` | `9ccf81b0017bfb036dc2fb702305f3252deb6dd491e1a942873eb4a0e793cf30` | 35 |

All three match. Line counts match. **Transcription verified against authored originals; delta at
`6211245` = section rules only.** So the transcription route introduced no content change, and the
question §9.5 was written to raise is answered: an original did exist, and this repo now agrees with
it.

**Re-baselining after the 2026-07-15 reflow.** The documents were then re-wrapped in the working
tree. The canned finding above no longer describes the ratified file, so it is not reused: the delta
is now **section rules + line re-wrapping + markdown table-column padding**. The reflow also
introduced two word-splits, caught by this check and repaired before ratification — `claims` → `c` +
`laims` in §1 Purpose, and `harness work` → `harnesswork` in this document's §0. Post-repair, a
whitespace-insensitive token comparison against the authored originals shows **0 differing tokens in
Amendments 1 and 2, and in §§0–9 only the table separator row's column padding** — no prose content
differs anywhere.

The ratified files therefore **do not hash to the authored originals**, and claiming otherwise would
be false. New baseline, over the ratified (reflowed, repaired) text:

| Ratified section | sha256 | Lines |
|---|---|---|
| §§0–9 | `11f44a2d2dd4af72f036766cb762feeb6c6902de8c4c9b80858bbf89a7f3cc28` | 147 |
| Amendment 1 | `a7f4fa7cb625c51f95cb6433ca360ba095a4351ff937ba6747ce2ec562457146` | 82 |
| Amendment 2 | `6dcf40d56761da85e48811353673209ceece9973fd2efb0964d069009e49d34f` | 61 |

Amendments 1 and 2 hash identically before and after repair, because the reflow left them
content-clean; only §§0–9 needed a fix. Both hash sets are recorded because the pre-registration
claim that matters is **content** identity with the authored original, which is verified, not byte
identity with it, which the reflow ended.
