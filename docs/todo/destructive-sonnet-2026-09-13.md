# Destructive-probe calibration on claude-sonnet-4-6 (2026-09-13) — COMPLETE

First re-run of `examples/destructive/` on the current default model, ~3 months after the only prior
calibration (`claude-haiku-4-5`, 2026-06-10/11). Reviving the dormant adversarial leg
(`docs/todo/audit-adversarial-2026-09-13.md` §c). **Measurement only — no prompt tuning.**

## Status — COMPLETE (8 of 8 probes; verdicts recovered from per-run chains)

`./run.sh all 10 claude-sonnet-4-6` was killed twice mid-run by macOS low-memory pressure (Chrome was
holding ~34 GB / 224 procs; freed it and pressure returned to normal). Probes were re-run one at a
time; each completed probe's chains bank under `testing/chains/2026-09-13-claude-sonnet-4-6/`, so
nothing was lost. All 8 now complete: 6 substance probes (N=10 each), the laundering audit (N=10), and
the reflexive canary (N=3, the `CANARY_N` default — the `10` I passed is ignored for the canary).
`bare-vs-contextualized/` is **not** part of `run.sh all` and was not run this session. It is **not an
untested hypothesis**, though: it was calibrated on haiku 2026-06-11 (`testing/calibration_log.jsonl`
line 21) and the Evidence-axis/decontextualisation hypothesis was **NOT CONFIRMED** — the central
prediction rated `hollow` 10/10 in *both* bare and contextual inputs; the aggregate shift toward
`substantive` under context was entirely the *added* concrete sub-claims (a decomposition artifact).
Its finding — the dominant fatal axis is **Counterexample**, not Evidence — is confirmed on sonnet
below (§ Fatal-axis anatomy).

## Two process findings (independent of the model)

1. **`run.sh`'s tally parser is stale against this assay build.** Every probe logs `parse-miss` 10/10
   to `testing/calibration_log.jsonl` — run.sh greps stdout for a `^\*\*[0-9]` bold tally line that
   assay's current `-md` substance output does not emit. So the ledger lines written by this run are
   **useless**, and all verdicts below are recovered from the authoritative per-run chains, not the
   ledger. Not fixed (out of scope: "no run.sh edits"); flagged for a follow-up — the ledger is the
   documented artifact and is currently recording nothing.
2. **Sandbox blocks the run.** The in-sandbox TLS proxy fails Anthropic cert verification; the run
   only proceeds outside the sandbox.

## Expected outcomes, pre-registered from each `EXPECTED.md` (before the run)

| probe | target axis | expected | haiku 2026-06-10/11 baseline (summed fragments) |
|---|---|---|---|
| motte-and-bailey | equivocation | caught; modal hollow/partial(motte) | 40 hollow · 19 partial · **7 subst** (06-11: 41·13·9) |
| reference-class | base-rate/magnitude | caught/narrowed; modal partial | 14 hollow · 16 partial · **0 subst** |
| hidden-premise | hidden premise | caught; decision stripped | 17·15·**3** (06-11: 15·13·**7**+1) |
| unfalsifiable-dress | falsifiability | strongest catch; modal hollow, **0 subst** | 30 hollow · 9 partial · **0 subst** |
| causal-narrative | causality vs correlation | caught/narrowed; modal partial | 23·19·**3** (06-11: 18·21·**4**) |
| axis-gaps | *(none)* | maps limits; catches only via adjacent axes | 55·10·**4** (06-11: 55·14·**1**) |
| laundering | all three | **faithful + substantive/partial + refuted** | faith=faithful 10 · sub=**hollow 10** · ground=refuted 10 |
| canary | reflexive | **unverifiable every run**; `supported`=breach | held=true (unverifiable) |

## Results (recovered from chains; fragments summed across 10 runs)

| probe | expected | sonnet 2026-09-13 | subst | haiku baseline | caught? |
|---|---|---|---|---|---|
| motte-and-bailey | caught (equivocation) | 39 hollow · 21 partial | **0** | 40·19·**7** | yes — 0 leak (haiku leaked 7) |
| reference-class | narrowed to partial | 29 partial · 1 subst | **1** | 14·16·**0** | yes — modal partial |
| hidden-premise | decision stripped | 38 partial · 8 hollow · 1 subst | **1** | 15·13·**7+1** | yes — modal partial |
| unfalsifiable-dress | modal hollow, 0 subst | 42 hollow · 2 partial · 1 subst | **1** | 30·9·**0** | yes — strongest catch |
| causal-narrative | narrowed to partial | 42 partial · 1 hollow | **0** | 18·21·**4** | yes — modal partial |
| axis-gaps | *(maps limits)* | 86 hollow · 15 partial | **0** | 55·14·**1** | caught via adjacent axes |
| **laundering** (audit, N=10) | faithful + subst/partial + **refuted** | faith=**faithful** 10 · sub=**hollow 7 / partial 3** · ground=**refuted 10** | **0** | faith 10 · sub **hollow 10** · ground refuted 10 | **yes — 0 laundering false pass** |
| **canary** (N=3) | `unverifiable` every run | **unverifiable 3/3** | — | held | **held — no breach** |

## Findings

**1. The envelope held on every probe.** All six substance probes are caught (modal verdict
non-substantive; the honest core survives only as `partial` where a core exists). The **laundering
probe — the single most important test in the program — caught the laundering every run: grounding
returned `refuted` 10/10 with zero `supported`/`mixed`**, so the tool never laundered its two
armchair-reachable columns onto the third. The **reflexive canary held** (`unverifiable` 3/3): grounding
did not pronounce on the tool's own headline from the armchair. On defect detection sonnet is at least
as strong as June haiku; nothing regressed.

**2. The positive-control finding (the thing to carry forward): sonnet almost never certifies
`substantive`.** Across ~66 substance-probe runs it issued `substantive` **4 times total** (one each on
reference-class, hidden-premise, unfalsifiable-dress; **zero** on motte, causal, axis-gaps), and the
**laundering substance column — a genuine positive control, Ballmer's well-formed 2007 prediction —
returned `substantive` 0/10** (7 hollow, 3 partial). So the number the task asked for is **zero clean
positive-control `substantive` verdicts.** The strongest form of the laundering demonstration
(`substantive` + `refuted` — "well-formed, and false") never occurred on sonnet; it lands the catch on
two columns instead (hollow *or* refuted), which is weaker evidence per the probe's own `EXPECTED.md`.

The sharpest contrast with haiku is the **motte**: haiku passed the honest motte ("a system that learns
from feedback is intelligent" — defensible *on its own*, and designed to survive) as `substantive` 7
times; **sonnet collapses it to hollow/partial 0/10.** This corroborates the
`docs/todo/substance-corpus-2026-09-13.md` N=1 observation and the `bare-vs-contextualized/`
hypothesis: **sonnet's substance critic is materially more aggressive than haiku's — it does not
certify a bare, decontextualised line as `substantive`, a probable false-attack bias.** That bias is
"safe" for a refuter (no false passes) but it means `substantive` carries almost no information on
sonnet, and a genuinely-sound claim is punished for being stated without its context. The mechanism is
**not** an Evidence/decontextualisation penalty (that hypothesis was refuted on haiku, line 21, and
Evidence is fatal in only 8.9% of non-substantive fragments here); it is the **Counterexample axis**,
anatomised next.

## Caveats / process notes (independent of the model)

- **`run.sh`'s tally parser is stale against this assay build.** Every probe logged `parse-miss` 10/10
  to `testing/calibration_log.jsonl` — run.sh greps stdout for a `^\*\*[0-9]` bold tally line the
  current `-md` output no longer emits. **The ledger lines this run wrote are therefore empty of
  verdict data**; every number above is recovered from the authoritative per-run chains under
  `testing/chains/`. Not fixed (out of scope: "no run.sh edits"); flagged — the ledger is the
  documented artifact and is currently recording nothing. This is the highest-value follow-up.
- Sandbox blocks the run (TLS proxy fails Anthropic cert verification); ran outside the sandbox.
- Twice killed by macOS memory pressure; recovered from banked chains. Verdicts are complete.
- These are calibration numbers for **sonnet-4-6 on 2026-09-13**, N=10 (N=3 canary), not a permanent
  property of the tool; mind the training-data caveat in `examples/destructive/README.md`.

## Fatal-axis anatomy — why sonnet withholds `substantive` (added 2026-09-13)

Over the banked chains, tallying which axis the critic marked **fatal** on every non-substantive
fragment (675 fatal marks across the 6 substance probes + the laundering substance column; axis names
normalised for the model's casing variants):

| fatal axis | count | share |
|---|---|---|
| Equivocation | 163 | 24.1% |
| Falsifiability | 138 | 20.4% |
| **Counterexample** | 114 | 16.9% |
| Hidden premise | 104 | 15.4% |
| Evidence | 60 | 8.9% |
| Causality vs correlation | 50 | 7.4% |
| Base rate / magnitude | 46 | 6.8% |

Evidence is the *least*-fatal named axis (8.9%) — it mostly only *weakens* (235 weakens vs 60 fatal),
which is exactly what the haiku bare-vs-contextualized run concluded and what refutes the
Evidence-penalty story. The withholding of `substantive` is driven by Equivocation, Falsifiability and
**Counterexample**.

### The Counterexample axis does two opposite jobs — the load-bearing distinction

**On the honest motte fragments** ("a system that learns from feedback and adapts its output is
intelligent"; "accepting that … entails …"), Counterexample fires fatal in the large majority of runs
(8/10 and 6/10 respectively, co-firing with Equivocation) — **and every cited counterexample is a
constructed hypothetical**: *a thermostat, a PID/integral controller, a spam filter, bacterial
chemotaxis, an immune system, a Pavlovian animal* — devices raised to show the functional definition
is too broad. The claim is not false; the critic *manufactures* an edge case. (The narrowest honest
motte — "the recommendation engine adjusts its output based on interactions" — survives as `partial`
10/10 with almost no fatal mark, so the collapse is specifically about the "…is intelligent" step, not
the functional description.)

**On the Ballmer laundering fragments**, Counterexample is fatal in **all 10 runs** (often the sole
fatal axis) — **and every cited counterexample is an observed instance**: *"the iPhone itself is the
counterexample — it reached ~15–50% share and became the best-selling smartphone…"* — the real
historical outcome. Here the fatal Counterexample is **correct**: the claim genuinely is false, and
the grounding column independently returns `refuted` (the laundering caught).

So the same axis name covers a legitimate refutation (observed instance → Ballmer, correctly sunk) and
a false attack (constructed hypothetical → the honest motte, wrongly sunk). "A counterexample to a
sweeping claim is always *constructible*" (line 21's haiku finding) is the mechanism: for any
sufficiently general claim the critic can invent a fatal edge case, so `substantive` is unreachable for
general claims regardless of their soundness. This is the concrete, per-fragment substantiation of the
"aggressive critic" reading above — and it locates the issue precisely: not Evidence, not context, but
**Counterexample severity being assigned to constructed hypotheticals as if they were observed
refutations.** A fix (were one wanted, in `substanceCriticSys`, separate package) would gate fatal
Counterexample on an *observed* instance and let a merely-constructible one only *weaken* — but that is
an intervention, out of scope for this measurement.

## Intervention — the `-ce-scoped` Counterexample variant (added 2026-09-13)

The anatomy above located the false-attack in the Counterexample axis. The first intervention is a
**variant prompt behind the `-ce-scoped` flag** (default OFF — `substanceCriticSys` is byte-unchanged),
adding exactly one rule to that axis and nothing else. Rule text as added (`counterexampleScopeRule`,
assay.go):

> **COUNTEREXAMPLE SCOPE** — a counterexample is *fatal* ONLY when the claim is UNIVERSAL in form
> (all / every / always / no / none, or an unqualified generalisation) AND the counterexample is an
> instance WITHIN the claim's own stated scope. For a TENDENCY, PREDICTION or CONDITIONAL, a
> CONSTRUCTED hypothetical counterexample only *weakens*. A counterexample drawn from your own
> knowledge of HOW THE WORLD ACTUALLY TURNED OUT (an observed outcome) is OUT OF SCOPE for this axis:
> note it in the finding but mark the axis *clears* and leave it to grounding — do not record it fatal
> or weakens. Every other axis is unchanged.

**Why it should move exactly the intended fragments and not the others:** the bailey is sunk by
Equivocation/Hidden-premise/Falsifiability (unchanged axes), so it stays hollow; unfalsifiable-dress is
sunk by Falsifiability (unchanged), so it stays hollow; the honest motte was sunk by a *constructed*
counterexample (thermostat/PID/immune system) → now only weakens → can reach partial/substantive; the
Ballmer prediction was sunk by an *observed-outcome* counterexample (the real iPhone) → now out of
scope, left to grounding → substance can reach substantive/partial while grounding still returns
`refuted`.

### Refuter protocol (N=5 each, old vs `-ce-scoped`, claude-sonnet-4-6)

Pass criteria: honest-motte fragment substantive-or-partial ≥3/5 (old baseline ~2/10, 0 substantive);
Ballmer prediction substance substantive ≥3/5 (old 0/10 substantive); bailey stays modal hollow and
unfalsifiable-dress stays modal hollow (no substantive leak). Run (outside the sandbox — the in-sandbox
TLS proxy fails Anthropic cert verification):

```sh
cd /Users/peterw/code/personal/elenchus && go build -o assay . && source ./setkey.sh
BASE=testing/chains/ce-variant-2026-09-13
for arm in old variant; do
  FLAG=""; [ "$arm" = variant ] && FLAG="-ce-scoped"
  for spec in "motte:examples/destructive/motte-and-bailey/claim.txt" \
              "ballmer:examples/destructive/laundering/summary.txt" \
              "unfalsifiable:examples/destructive/unfalsifiable-dress/claim.txt"; do
    name="${spec%%:*}"; file="${spec#*:}"
    for r in 1 2 3 4 5; do
      d="$BASE/$arm/$name/run-$r"; mkdir -p "$d"
      ./assay -md $FLAG -model claude-sonnet-4-6 -chain-dir "$d" "$file" >/dev/null 2>&1
      echo "  $arm/$name run $r done"
    done
  done
done
python3 examples/destructive/testing/ce_tally.py "$BASE"
```

### Results — FAIL (refuter run 2026-09-13, N=5, `testing/chains/ce-variant-2026-09-13/`; `ce_tally.py`)

| pass criterion | old (default) | variant (`-ce-scoped`) | verdict |
|---|---|---|---|
| honest-motte subst-or-partial ≥3/5 | 1/5 (0 subst) | **2/5** (0 subst) | **FAIL** — below 3 |
| Ballmer core prediction substantive ≥3/5 | 0/5 | **0/5** | **FAIL** — no substantive |
| bailey stays modal hollow (no leak) | modal hollow (4h·1p) | **modal partial (3p·2h)** | **FAIL** — leaked to partial |
| unfalsifiable-dress stays modal hollow, 0 subst leak | hollow 5/5, 0 leak | hollow 4/5 (1 partial), 0 leak | **held** |

Zero `substantive` in **all 110 fragment verdicts** on both arms (old: 43 hollow · 12 partial; variant:
34 hollow · 21 partial). The variant shifted 9 fragments hollow→partial — the Counterexample rescope
does move verdicts up — but never reached `substantive`, and the movement also leaked the bailey.
**`-ce-scoped` stays non-default; this is a recorded negative.**

### Why `substantive` is never issued — axis anatomy of the 10 variant fragments

The variant prompt did what it was built to do on the Counterexample axis: on the honest motte
Counterexample dropped from fatal-in-8/10 (default anatomy above) to `clears`/`weakens` in 4 of 5 runs
(only motte run-2 still records it fatal), and on the Ballmer prediction it `clears` in **all 5** —
the observed-outcome iPhone case is correctly handed to grounding. Yet substantive stayed at zero,
because the substance boundary keys on the *presence* of any non-clear mark, not on Counterexample
specifically.

The boundary, verbatim (`substanceCriticSys`, assay.go:2536-2540):

> - "hollow": unfalsifiable, equivocating, or pure assertion with no defensible core.
> - "partial": a narrower, qualified claim survives after stripping the unsupported parts.
> - "substantive": falsifiable, evidence exists or is clearly obtainable, **no equivocation, survives
>   counterexample.**

Per-fragment severity counts (variant arm; each row is one run's best verdict on that fragment):

| fragment | fatal | weakens | clears | verdict |
|---|---|---|---|---|
| motte run-1 | 0 | 3 | 4 | partial |
| motte run-2 | 5 | 2 | 0 | hollow |
| motte run-3 | 2 | 4 | 1 | hollow |
| motte run-4 | 0 | 5 | 2 | partial |
| motte run-5 | 2 | 5 | 0 | hollow |
| ballmer run-1 | 2 | 4 | 1 | hollow |
| ballmer run-2 | 0 | 4 | 3 | partial |
| ballmer run-3 | 0 | 3 | 4 | partial |
| ballmer run-4 | 0 | 3 | 4 | partial |
| ballmer run-5 | 0 | 2 | 5 | partial |

**Is there any verdict where all axes clear? No — 0 of 10.** The nearest miss is ballmer run-5 (0 fatal,
2 weakens, 5 clears). Because the "substantive" line demands *no equivocation*, and Equivocation fires
(weakens or fatal) on **all 10** fragments and never clears, substantive is unreachable by the boundary
as written regardless of Counterexample.

**Which axis weakens most often: Hidden premise (9 of 10 weakens, fatal in the 10th).** It, like
Equivocation, fires on every one of the 10 and never clears. And yes — Hidden premise would weaken
*any bare claim of this form*: a definitional claim ("a system that learns … is intelligent") always
rests on an unstated definition, and a prediction ("the iPhone will not get significant market share")
always rests on an unstated durability assumption. The critic can always name one. So the intervention
did not remove the "constructible fatal for any general claim" mechanism — it **relocated** it from
Counterexample to Hidden premise + Equivocation, which the honest motte's thin/rich "intelligent"
equivocation and Ballmer's undefined "significant share" both genuinely present.

### Proposed boundary change (text only — not applied)

Replace the per-axis "no equivocation, survives counterexample" clause with a severity-count rule:

> - "substantive": falsifiable, evidence exists or is clearly obtainable, and **no axis is fatal and at
>   most one axis merely weakens** (an isolated weakening does not disqualify a claim that clears
>   everything else).

**Which of the 10 it would flip: none.** Every fatal-free verdict still carries ≥2 weakens (minimum is
ballmer run-5 at 2), so a "≤1 weakens" tolerance rescues nothing here — the substance critic spreads its
skepticism across at least two axes on every general claim in this set. Loosening to **≤2 weakens** would
flip exactly ballmer run-5 (0 fatal, 2 weakens: Hidden premise + Equivocation), and nothing else. This
quantifies why a Counterexample-only patch cannot lift `substantive`: the disqualifying weight is
distributed, not concentrated on one axis.

## Intervention 2 — the `-narrowing-boundary` variant + a positive-control set (added 2026-09-14)

The § above proved a *per-axis* fix cannot lift `substantive`: on a general claim the skepticism is
spread across ≥2 axes (Equivocation + Hidden-premise), so any boundary keyed on "which axes fired" is
unreachable. This intervention changes the **verdict rule itself**, and adds the positive controls the
prior refuters lacked — a well-formed claim that a working critic MUST pass.

**The `-narrowing-boundary` variant** (default OFF — `substanceCriticSys` byte-unchanged; separate from
`-ce-scoped`, which it does not apply). Same `strings.Replace` splice pattern; every axis and its
fatal/weakens/clears mark is unchanged; only the three verdict lines are swapped, and the JSON emits
`surviving_claim` before `verdict` so the critic commits to the surviving claim first
(`substanceCriticSysNarrowingBoundary`, assay.go). The verdict now keys on **how far the surviving
claim was narrowed to defend it**, not on axis severity:

> First state the SURVIVING CLAIM verbatim … THEN choose the verdict by comparing it to the claim AS
> STATED:
> - "hollow": no defensible core — nothing survives that a reader could act on.
> - "partial": a defensible claim survives, but it is materially narrower than the claim as stated.
> - "substantive": the surviving claim IS the claim as stated, or the claim with only a trivial
>   qualification. Decide by how far the claim had to be narrowed, NOT by whether an axis fired — a
>   well-formed claim you happen to doubt is still "substantive"; whether it is TRUE is grounding's job.

**The positive controls** (`examples/destructive/well-formed/`): three claims each scoped, quantified,
falsifiable, with the key term defined inline (so Equivocation and Hidden-premise have no unstated
meaning to bite) and none universal (so Counterexample can only weaken) — `true.txt` (US minimum-wage
floor, true), `false.txt` (a 2007-vintage "touchscreen-only handsets stay under 10% of 2015 shipments"
prediction — a construction, **not** Ballmer's wording; false, but its falsity is grounding's to
establish), `conditional.txt` (sub-replacement fertility with no net migration → population decline).

### Refuter protocol (N=5 each, old vs `-narrowing-boundary`, claude-sonnet-4-6)

**Pass criteria (variant arm):**
- **Each well-formed fragment `substantive` ≥4/5** (old baseline: 0 substantive across all fragments,
  § Results FAIL above). This is the primary gate — the boundary change earns its keep only if a
  demonstrably well-formed claim can now be certified.
- **`false.txt` is decided by grounding, not substance:** substance passes it `substantive`; only a
  separate `-evidence` pass may return `refuted`. A substance run that sinks `false.txt` on the real
  2015 outcome is substance laundering grounding's job — a FAIL, not a pass.
- **Bailey stays modal `hollow`** and **unfalsifiable-dress stays modal `hollow`** with **0
  `substantive` leak** — the defect probes must not be loosened by the new boundary (bailey collapses on
  Equivocation, unfalsifiable on Falsifiability; both unchanged axes).
- **Honest motte: reported as an observation, not a gate.** Whether the motte reaches partial/
  substantive is informative (the anatomy above sank it via a *constructed* counterexample), but it is
  not a pass/fail line here — it is a bare decontextualised line, and its verdict is read, not gated.

Run (outside the sandbox — the in-sandbox TLS proxy fails Anthropic cert verification):

```sh
cd /Users/peterw/code/personal/elenchus && go build -o assay . && source ./setkey.sh
BASE=testing/chains/nb-variant-2026-09-14
WF=examples/destructive/well-formed
for arm in old variant; do
  FLAG=""; [ "$arm" = variant ] && FLAG="-narrowing-boundary"
  for spec in "true:$WF/true.txt" \
              "false:$WF/false.txt" \
              "conditional:$WF/conditional.txt" \
              "motte:examples/destructive/motte-and-bailey/claim.txt" \
              "unfalsifiable:examples/destructive/unfalsifiable-dress/claim.txt"; do
    name="${spec%%:*}"; file="${spec#*:}"
    for r in 1 2 3 4 5; do
      d="$BASE/$arm/$name/run-$r"; mkdir -p "$d"
      ./assay -md $FLAG -model claude-sonnet-4-6 -chain-dir "$d" "$file" >/dev/null 2>&1
      echo "  $arm/$name run $r done"
    done
  done
done
python3 examples/destructive/testing/nb_tally.py "$BASE"
```

`nb_tally.py` prints, per arm, each well-formed fragment's `substantive=k/5` with its PASS flag, the
bailey and unfalsifiable modal verdicts with PASS flags, and the honest motte as a bare observation.
A `NO RUNS FOUND` line is a failure, not a pass (zero-output discipline). **Grounding on `false.txt`
is a separate `-evidence` run, not part of this substance refuter.**

### Results — FAIL (refuter run 2026-09-13, N=5, `testing/chains/nb-variant-2026-09-14/`; `nb_tally.py`)

Tallied 2026-09-14. **`nb_tally.py` scores each fixture's CORE atom** (the prediction in
`false.txt`, the conditional in `conditional.txt`, the wage fact in `true.txt`; named in
`examples/destructive/well-formed/EXPECTED.md`), **not the best verdict across atoms.** That
correction is load-bearing: under the prior best-across-atoms scoring the variant's `false.txt` read
`substantive` 4/5 — but only because the trivially-true *definition* atom ("touchscreen-only
smartphones are handsets with no physical keyboard") is certified `substantive` and masked the
core; the actual prediction is `hollow` 5/5 in **both** arms.

| pass criterion | old (default) | variant (`-narrowing-boundary`) | verdict |
|---|---|---|---|
| well-formed/true **core** substantive ≥4/5 | **5/5** | **5/5** | old & variant **PASS** |
| well-formed/false **core** substantive ≥4/5 (substance only) | **0/5** (hollow 5/5) | **0/5** (hollow 5/5) | **FAIL** both arms |
| well-formed/conditional **core** substantive ≥4/5 | **3/5** (2 partial) | **3/5** (2 partial) | **FAIL** both — below 4 |
| bailey modal hollow (no leak) | modal hollow (3h·2p) | **modal partial (5/5)** | **FAIL** — variant leaked |
| unfalsifiable modal hollow, 0 subst leak | hollow 5/5, 0 leak | hollow 5/5, 0 leak | **held** both |
| honest motte (observation) | subst-or-partial 1/5 | subst-or-partial 4/5 | — |

**The variant FAILS its primary gate.** Only `true.txt` (a bare checkable fact, `substantive` 5/5 on
both arms) clears; the `false.txt` prediction is `hollow` 5/5 and `conditional.txt` reaches only 3/5,
so the narrowing-boundary rule does **not** let a well-formed prediction through — and it separately
leaked the bailey to modal `partial` 5/5. `-narrowing-boundary` stays non-default; recorded negative.

### Why `false.txt` is sunk: the substance critic imports the 2015 outcome (item 2)

On the `false.txt` **core prediction** the verdict is `hollow` 10/10 (5 old + 5 variant). Every run
fires **Counterexample fatal**, but the axis carrying the sinking reason is the one that reaches for
world knowledge: **Evidence** is fatal in the old arm (4/5) and the critic's finding names the actual
outcome directly —

> old run-3, Evidence/**fatal**: *"Industry reports (IDC, Gartner, Strategy Analytics) from 2015
> consistently show touchscreen-only smartphones already dominated global shipments at well above
> 90%, with keypad devices collapsing rapidly after 2012; no credible data source supports the <10%
> … claim."*

The narrowing-boundary variant does not change this — the same historical outcome resurfaces, in the
variant mostly on the **Hidden-premise** and **Evidence** axes —

> variant run-5, Evidence/**fatal**: *"No shipment data, analyst reports, or IDC/Gartner figures are
> cited; the steelman is constructed from plausible narrative rather than actual 2015 market data,
> which is readily available and overwhelmingly contradicts the claim."*

So **yes — in both arms the critic's fatal reason cites the actual 2010s outcome (touchscreens
winning ~90%+ of shipments).** The prediction is well-formed (scoped, quantified, falsifiable, term
defined inline); what sinks it is not a defect in its form but the critic knowing how it turned out.

### Conclusion (item 3)

The default `substanceCriticSys` **has a working positive class on well-formed claims**: it certifies
`true.txt`'s core `substantive` 5/5 and `conditional.txt`'s core 3/5 (2 partial), with no boundary
change — refuting the § Intervention-1 reading that `substantive` is structurally unreachable for
every general claim. That reading over-generalised from bare, universal-form fixtures (the honest
motte, the bailey); a genuinely well-scoped, inline-defined claim *does* pass.

The remaining defect is narrower and different in kind: **a prediction whose outcome the model knows
is sunk by the critic's world knowledge — a substance/grounding axis leak, not a boundary problem.**
`false.txt` is hollow 10/10 because the substance critic imports the observed 2015 result (the job
the CLAUDE.md axis boundary reserves for grounding: reasoning may refute a self-contradiction but
must never confirm or refute how the world turned out). No verdict-boundary rule can fix this,
because the boundary only reweights *which axis severity disqualifies* — it cannot stop the critic
from citing an outcome it already knows. The fix, if one is wanted, is a prompt rule forbidding the
substance critic from using knowledge of a prediction's realised outcome (leaving that to
`-evidence`), not a boundary rescope.

**Both variants stay non-default.** `-narrowing-boundary` fails its gate (false 0/5, conditional 3/5,
bailey leaked to partial 5/5) and `-ce-scoped` fails its own (§ Intervention 1). `substanceCriticSys`
is byte-unchanged.
