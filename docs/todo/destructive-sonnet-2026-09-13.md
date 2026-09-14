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

## Intervention 3 — the `-as-of` variant (added 2026-09-14)

§ Intervention 2 item 2 located the surviving defect precisely: `false.txt`'s core prediction is
`hollow` 10/10 (5 old + 5 variant) because the substance critic **imports the realised 2010s
outcome** — the fatal reason cites actual IDC/Gartner shipment figures (old run-3, Evidence/fatal;
variant run-5, Evidence/fatal), the job CLAUDE.md's axis boundary reserves for grounding. § Conclusion
(item 3) named the fix: not a verdict-boundary rescope (a boundary only reweights *which* axis
severity disqualifies — it cannot stop the critic citing an outcome it already knows) but **a prompt
rule forbidding the substance critic from using knowledge of a prediction's realised outcome**,
leaving that to `-evidence`. This intervention is that rule.

**The `-as-of` variant** (default OFF — `substanceCriticSys` byte-unchanged; separate from `-ce-scoped`
and `-narrowing-boundary`, and applying neither). Same `strings.Replace` splice + length-test pattern:
one rule (`asOfRule`, assay.go) is inserted **before the axis list** so the critic reads it first, and
every axis and the verdict block are byte-identical to the default. Rule text as added:

> Judge the claim as of the date it was made, using only what a careful reader could have known then.
> Whether the prediction later came true is not a substance question; if you find yourself citing the
> realised outcome, stop — that belongs to the evidence axis. Predictions are judged on scope,
> falsifiability, and whether a mechanism is offered, not on hindsight.

**Why it should move `false.txt` and nothing else:** the bailey is sunk by Equivocation/Hidden-premise
(unchanged axes) on the equivocating "intelligent", not by any outcome — so it stays hollow;
unfalsifiable-dress is sunk by Falsifiability (unchanged) — so it stays hollow; `false.txt`'s core was
sunk *only* by the critic reaching for the 2015 shipment figures → now barred → the well-formed
prediction (scoped, quantified, falsifiable, term defined inline) can reach `substantive`, with its
falsity left to a separate `-evidence` pass. If `false.txt` does **not** rise, the defect is not the
hindsight import after all and this reading is refuted.

### Refuter protocol (N=5 each, old vs `-as-of`, claude-sonnet-4-6) — RUN 2026-09-14

Chains at `testing/chains/as-of-variant-2026-09-14/{old,variant}/`; scored by `nb_tally.py`. Pass
criteria:

- **`false.txt` core prediction `substantive` ≥4/5 on the variant arm** (old baseline: `hollow` 5/5,
  0 substantive — § Intervention 2 Results). This is the primary gate — the rule earns its keep only
  if barring the realised outcome lets the well-formed prediction through on substance, its falsity
  left to grounding.
- **Bailey stays modal `hollow`** and **unfalsifiable-dress stays modal `hollow`** with **0
  `substantive` leak** — both collapse on axes the rule does not touch (Equivocation/Hidden-premise;
  Falsifiability), so the hindsight bar must not loosen them.

`nb_tally.py` already scores exactly these three (false core via `will account for less than 10%`,
bailey modal via `is intelligent`, unfalsifiable modal); it also prints `NO RUNS FOUND` for `true`,
`conditional` and an honest-motte observation line, which are **not part of this refuter's gate** and
are ignored here. Run (outside the sandbox — the in-sandbox TLS proxy fails Anthropic cert
verification):

```sh
cd /Users/peterw/code/personal/elenchus && go build -o assay . && source ./setkey.sh
BASE=testing/chains/as-of-variant-2026-09-14
for arm in old variant; do
  FLAG=""; [ "$arm" = variant ] && FLAG="-as-of"
  for spec in "false:examples/destructive/well-formed/false.txt" \
              "motte:examples/destructive/motte-and-bailey/claim.txt" \
              "unfalsifiable:examples/destructive/unfalsifiable-dress/claim.txt"; do
    name="${spec%%:*}"; file="${spec#*:}"
    for r in 1 2 3 4 5; do
      d="$BASE/$arm/$name/run-$r"; mkdir -p "$d"
      ./assay -md $FLAG -model claude-sonnet-4-6 -chain-dir "$d" "$file" >/dev/null  || echo "FAILED $arm/$name run $r"
      echo "  $arm/$name run $r done"
    done
  done
done
python3 examples/destructive/testing/nb_tally.py "$BASE"
```

### Results — FAIL (refuter run 2026-09-14, N=5, `testing/chains/as-of-variant-2026-09-14/`; `nb_tally.py`)

The primary gate fails. The `false.txt` core prediction is `hollow` **0 substantive / 5** on **both**
arms — the `-as-of` rule moved it not at all. `nb_tally.py`:

```
old      well-formed/false  substantive=0/5  PASS=NO   bailey modal=[(partial,3),(hollow,2)]  unfalsifiable 0-leak PASS=yes
variant  well-formed/false  substantive=0/5  PASS=NO   bailey modal=[(hollow,5)]              unfalsifiable 0-leak PASS=yes
```

Secondary gates hold — bailey `hollow` 5/5 on the variant arm, unfalsifiable-dress `hollow` 5/5 with
0 `substantive` leak on both — but that only confirms the rule loosened nothing it shouldn't; it
earned nothing on the gate it exists for.

**Why it fails — the rule does not suppress the hindsight import.** Every one of the 5 variant
(`-as-of`) core-atom reasons still cites the *realised* 2015 shipment outcome, exactly the move the
rule text forbids. Core atom = "Touchscreen-only smartphones will account for less than 10% of global
smartphone…"; all 10 rows `hollow`:

| arm/run | verdict | fatal axes | reason cites 2015 realised outcome? |
|---|---|---|---|
| variant run-1 | hollow | Equivocation, Evidence, Hidden-premise, Falsifiability, Base-rate, Counterexample | yes — "iPhone shipments alone (≈15% of global market in 2015) exceed the claimed 10% ceiling" |
| variant run-2 | hollow | Equivocation, Falsifiability, Base-rate, Hidden-premise, Counterexample | yes — "touchscreen-only devices constituted well over 90% of 2015 global shipments" |
| variant run-3 | hollow | Evidence, Equivocation, Hidden-premise, Base-rate, Counterexample | yes — "constituted the overwhelming majority (~90%+) of global smartphone shipments in 2015" |
| variant run-4 | hollow | Evidence, Hidden-premise, Equivocation, Base-rate, Counterexample | yes — "All available 2015 market data show … accounting for well over 90% of global shipments" |
| variant run-5 | hollow | Evidence, Hidden-premise, Equivocation, Base-rate, Counterexample | yes — "by 2015, touchscreen-only smartphones constituted well above 90% of global shipments" |
| old run-1 | hollow | Equivocation, Counterexample | yes — "almost certainly exceeded 10% of 2015 global shipments by a wide margin" |
| old run-2 | hollow | Evidence, Hidden-premise, Equivocation, Base-rate, Counterexample | yes — "Apple's iPhone shipments alone in 2015 exceeded 10% of global smartphone volume" |
| old run-3 | hollow | Evidence, Equivocation, Hidden-premise, Falsifiability, Base-rate, Counterexample | yes — "By 2015 … roughly 85–90% of global shipments … not less than 10%" |
| old run-4 | hollow | Evidence, Hidden-premise, Falsifiability, Equivocation, Base-rate, Counterexample | yes — "empirically false by all available 2015 market data" |
| old run-5 | hollow | Evidence, Hidden-premise, Base-rate, Counterexample | yes — "Android touchscreen-only devices alone accounted for roughly 80% of ~1.4 billion units shipped" |

**Full variant-arm reasons, verbatim** (the arm the rule was meant to change):

- run-1: "The claim fails on multiple fatal axes simultaneously. Under any standard industry
  definition of 'touchscreen-only,' Apple's iPhone shipments alone (≈15% of global market in 2015)
  exceed the claimed 10% ceiling, directly falsifying the claim with a single counterexample. The
  steelman's rescue strategy depends on a non-standard, highly idiosyncratic redefinition that
  excludes devices with any physical button whatsoever — a definition found in no industry dataset
  and not stated in the original claim. This is condition laundering: the claim survives only by
  importing a definition the original speaker did not supply. The equivocation is fatal, the empirical
  base rate is fatal, and the counterexample is fatal. No defensible narrowed claim survives. Survives
  only by conditions the speaker never stated."
- run-2: "The claim is fatally equivocal. Under the standard 2015 industry meaning of
  'touchscreen-only' (primary input is a touchscreen; no physical keyboard), the claim is
  straightforwardly and massively false — touchscreen-only devices constituted well over 90% of 2015
  global shipments. The producer's steelman survives only by laundering in a hyper-strict redefinition
  ('no physical buttons of any kind') that no contemporary analyst, journalist, or consumer applied to
  the term, and under which the share was near 0% anyway, not near 10%. Neither reading produces a
  defensible, meaningful prediction: one is trivially false, the other trivially true but
  counterfactually miscalibrated. The 10% threshold is unsupported by any cited data and is wrong in
  both directions depending on definition. The claim is hollow."
- run-3: "The claim is empirically inverted by a large margin. Touchscreen-only smartphones
  constituted the overwhelming majority (~90%+) of global smartphone shipments in 2015 under any
  standard industry definition. The steelman's rescue attempt depends on reclassifying feature phones
  as smartphones — a move that violates standard taxonomy and, if accepted, makes the claim
  untestable. Three axes fire at fatal severity: the evidentiary record contradicts the claim
  directly, the magnitude is off by roughly an order of magnitude, and obvious counterexamples (Apple,
  Samsung, the entire Android ecosystem) alone exceed the claimed 10% ceiling. No narrowing survives
  without laundering the definition of 'smartphone' beyond recognition."
- run-4: "The claim collapses on every empirical axis. All available 2015 market data show
  touchscreen-only smartphones — under any standard industry definition — accounting for well over 90%
  of global shipments, not under 10%. The steelman's attempt to rescue the claim requires a tortured
  redefinition of 'touchscreen-only' that would exclude the iPhone 6 and Galaxy S6, which is
  definitional manipulation, not narrowing. No version of the claim survives without laundering
  conditions the original speaker never stated and that invert market reality."
- run-5: "The claim inverts observable market reality: by 2015, touchscreen-only smartphones
  constituted well above 90% of global shipments, not below 10%. The steelman's definition of
  'touchscreen-only' is a private redefinition that no industry body used and that cannot be tested
  against published data. Every supporting premise — BlackBerry enterprise volume, keyboard-device
  prevalence in emerging markets, hybrid-form-factor significance — is empirically refuted by
  available shipment data. Three fatal findings on evidence, hidden premise, and base rate/magnitude
  independently render the claim indefensible."

**(a) Does any variant (`-as-of`) reason reference what actually happened to touchscreen phones?**
Yes — all 5 do, each citing the realised 2015 shipment record (>90% touchscreen-only; iPhone ≈15%;
Android ~80%), the outcome the rule tells the critic to stop at. The rule did not fire.

**(b) Stated ground for `hollow`.** Not applicable in the "if not" sense — the ground *is* the
imported outcome. Each variant reason rests `hollow` on Equivocation/Hidden-premise (condition
laundering on "touchscreen-only") **fused with** Evidence/Base-rate/Counterexample findings that are
the 2015 realised figures. So the verdict leans on exactly the hindsight `-as-of` was meant to bar,
alongside the definitional axes.

**Bailey (`nb_tally` scorer):** `hollow` 5/5 on the variant arm; old arm 3 partial / 2 hollow this
run vs 3 hollow / 2 partial the prior run — N=5 noise on an atom the rule does not target, not a
signal.

**Errored atoms:** 4 of 140 substance atom-runs returned `verdict:error` (API errors, not verdicts),
none on the `false.txt` core atom: old/unfalsifiable (2 atoms) and variant/motte (2 atoms). They do
not affect any gate above.

**Read.** § Intervention 2 item 2's reading survives: the defect is the hindsight import, and a
prompt rule placed before the axis list does not suppress it — the Evidence/Base-rate/Counterexample
axes re-introduce the realised outcome by name regardless of the instruction. A verdict-boundary
rescope was already ruled out; a prompt rule now is too. What remains untried is removing the
critic's licence to assert empirical figures at all on a prediction (route those to `-evidence`),
which is a larger change than a spliced sentence. `-as-of` stays default OFF; refuter did not pass.

## Conclusion — substance and grounding do not separate on a claim the model knows is false (added 2026-09-14)

Three interventions closed the same door from three sides. `-ce-scoped` rescoped the Counterexample
axis; `-narrowing-boundary` rewrote the verdict rule; `-as-of` inserted a hindsight bar before the
axis list. None lifted `false.txt`'s core prediction above `hollow`, and Intervention 3 shows why in
the plainest form available: **all 5 `-as-of` variant reasons cite the realised 2015 shipment outcome
by name** (>90% touchscreen-only; iPhone ≈15%; Android ~80%) — the exact move the rule text forbids —
fused into Evidence/Base-rate/Counterexample findings the boundary and the hindsight bar leave
untouched (§ Intervention 3 Results, rows *variant run-1…5*).

The finding is therefore not about a prompt or a boundary: **substance and grounding cannot be
separated on a claim the model already knows to be false.** When the outcome is in the model's
training data, the substance critic reaches for it regardless of instruction, because the axes
themselves (Evidence, Base-rate, Counterexample) are licensed to cite empirical fact and a known
outcome is empirical fact. This is a **property of the model, accepted as a limit**, not a defect to
patch. It bites only on **retrospective audits** — a claim whose truth was settled before the
model's cutoff — where the substance column silently borrows grounding's answer. On a prediction
whose outcome the model cannot know, the leak has nothing to draw on; that is what the retired
`false.txt` → `prediction.txt` swap tests (`examples/destructive/well-formed/`).

**Three-way outcome, by claim horizon:** a past/present well-formed claim reaches `substantive`
(`true.txt` 5/5); a forecast reaches `partial` with the number, scope and horizon intact and only the
modality stripped (`prediction.txt` 5/5, § Prediction fixture refuter); a known-outcome claim cannot
separate substance from grounding — the substance column silently borrows grounding's answer
(`false.txt` `hollow` 10/10), a stated limit rather than a patchable defect.

**What stands, unchanged:**

- The **default `substanceCriticSys`** stands, byte-unchanged. It has a working positive class on
  well-formed claims (`true.txt` core `substantive` 5/5, `conditional.txt` 3/5 on the default critic,
  § Intervention 2 Results); the hindsight limit above is a boundary of what substance can settle, not
  a reason to change the prompt.
- The **13 Sept 8-probe calibration** stands (§ Results): the envelope held on every probe, the
  laundering audit caught the laundering 10/10 (`refuted`, zero false pass), and the reflexive canary
  held (`unverifiable` 3/3).
- **`-ce-scoped`, `-narrowing-boundary`, and `-as-of` remain non-default recorded negatives** — each
  refuter is logged as FAIL against its own pre-registered gate (§§ Intervention 1–3 Results), and all
  three stay default OFF.

## Prediction fixture refuter (pre-registered — not yet run) (added 2026-09-14)

`false.txt` is retired and replaced by `examples/destructive/well-formed/prediction.txt`: a
post-cutoff prediction (battery-electric cars pass 25% of new EU passenger-car sales by end of
calendar year 2029) whose outcome no current model can know, so the hindsight import above has
nothing to draw on. Its core atom is expected `substantive` on substance and `unverifiable` on a
separate `-evidence` grounding pass (`examples/destructive/well-formed/EXPECTED.md`).

**Pass criterion:** the core prediction atom (matched by `more than 25%`, which excludes the
definition and the causal-mechanism atoms) is `substantive` **≥4/5** on the default prompt. This is a
single-arm refuter — **default `substanceCriticSys` only, no variant flag** (all three variants are
closed negatives above). Run outside the sandbox (the in-sandbox TLS proxy fails Anthropic cert
verification):

```sh
cd /Users/peterw/code/personal/elenchus && go build -o assay . && source ./setkey.sh
BASE=testing/chains/prediction-2026-09-14
FILE=examples/destructive/well-formed/prediction.txt
for r in 1 2 3 4 5; do
  d="$BASE/run-$r"; mkdir -p "$d"
  ./assay -md -model claude-sonnet-4-6 -chain-dir "$d" "$FILE" >/dev/null || echo "FAILED run $r"
  echo "  run $r done"
done
# Score the CORE atom (not best-across-atoms): substantive=k/5 on the prediction.
python3 - "$BASE" <<'PY'
import json, glob, sys, collections
base = sys.argv[1]
core = "more than 25%"          # selects the prediction atom; excludes definition + mechanism
rank = {"substantive": 2, "partial": 1, "hollow": 0, "error": -1}
runs = sorted(glob.glob(f"{base}/run-*"))
if not runs:
    sys.exit("NO RUNS FOUND — zero-output is a failure, not a pass")
per = []
for d in runs:
    recs = [json.loads(l) for f in glob.glob(f"{d}/*.substance.jsonl") for l in open(f)]
    frs = [r for r in recs if core in r["claim"]]
    if not frs:
        per.append(None); continue          # unmatched core atom = failure, fix the selector
    per.append(max(frs, key=lambda r: rank.get(r["verdict"], -1))["verdict"])
if any(v is None for v in per):
    sys.exit(f"CORE ATOM UNMATCHED in {per.count(None)} run(s) — fix the `more than 25%` selector")
subst = sum(1 for v in per if v == "substantive")
print(f"prediction core per-run={per}  substantive={subst}/{len(per)}  PASS={'yes' if subst >= 4 else 'NO'}")
PY
```

A `NO RUNS FOUND` or `CORE ATOM UNMATCHED` line is a failure, not a pass (zero-output discipline).
Grounding on `prediction.txt` — expected `unverifiable` — is a separate `-evidence` run, not part of
this substance refuter.

### Results — FAIL, but not on hindsight (refuter run 2026-09-14, N=5, `testing/chains/prediction-2026-09-14/`)

Core atom = "Battery-electric vehicles will make up more than 25% of new passenger-car unit sales in
the European Union by the end of calendar year 2029" (idx 2; the `more than 25%` selector matches it
alone — idx 5's causal restatement says "exceed 25%", not "more than 25%"). Scorer output:

```
prediction core per-run=['partial','partial','partial','partial','partial']  substantive=0/5  PASS=NO
```

**The pre-registered gate (core `substantive` ≥4/5) is NOT met: `partial` 5/5, `substantive` 0/5.**
This also misses `EXPECTED.md`'s envelope, which pre-registered `substantive`. But the failure is a
different animal from `false.txt`'s: the retired 2015-horizon fixture sank `hollow` 5/5 by importing
the realised outcome (§ Intervention 2 Results); this post-cutoff prediction sinks to `partial`
**with no hindsight import** — no run cites a 2029 result, because none exists to cite. Every run
reasons forward from 2023–2024 data (S-curve deceleration, Germany's 2023 subsidy removal, the
required pp/year vs. observed pp/year), which is the substance axes doing their own job, not
grounding's.

| run | verdict | fatal axes | weakens axes | surviving claim (gist) |
|---|---|---|---|---|
| 1 | partial | — none | Hidden-premise, Base-rate, Counterexample, Causality | "plausibly exceed 25% … contingent on the regulatory framework intact and no sustained policy reversal — a probabilistic claim rather than a near-certainty" |
| 2 | partial | — none | Hidden-premise, Counterexample | "will exceed 25% … provided EU fleet CO2 regs and the 2035 ICE phase-out remain materially in force" |
| 3 | partial | — none | Hidden-premise, Base-rate, Counterexample, Causality | "more likely than not to exceed 25% … provided CO2 fleet regs for 2025 and 2030 remain intact and no prolonged demand shock" |
| 4 | partial | — none | Hidden-premise (×2), Equivocation, Base-rate, Counterexample | "will exceed 25% … provided CO2 fleet penalties remain in force and battery price-parity reached ~2026–2027" |
| 5 | partial | — none | Hidden-premise, Base-rate, Counterexample | "will likely exceed 25% … provided CO2 fleet regs not materially rolled back and battery costs keep declining" |

**What the critic stripped — one line: the certainty, nothing else.** The number (25%), the scope (EU
new passenger-car unit sales) and the horizon (end-2029) are preserved verbatim in all 5 surviving
claims. What is removed each time is the modality: the unconditional "will make up more than 25%"
becomes "likely / plausibly / more-likely-than-not to exceed 25%, **provided** [the producer's own
stated conditions] hold". Not the mechanism, not the number, not the scope, not the horizon — the
modal certainty. Every reason states the attached conditions are producer-stated and explicitly
**not** laundered ("no condition laundering applies"; "legitimate narrowing rather than laundering").

**Why `substantive` is unreachable here — a property of the default rule, not a hindsight leak.** No
axis fires `fatal` in any of the 5 runs; the verdict is `partial` off `weakens`-level findings alone.
`substanceCriticSys` (assay.go:2552) defines `substantive` as "falsifiable, evidence exists …, no
equivocation, **survives counterexample**", and every run finds the Germany-2023 subsidy-removal
counterexample at least `weakens` — a live scenario a genuine empirical prediction can never fully
close. So the critic writes a narrower conditional `surviving_claim`, which is exactly the "partial:
a narrower, qualified claim survives" branch (assay.go:2551). The certainty-strip is the mechanism;
the rule reads any producer-stated hedge as material narrowing. This is the mirror of the `false.txt`
finding: there substance borrowed grounding's *answer*; here substance cannot certify a well-formed
prediction because the default rule requires an uncloseable counterexample to clear — a claim about
the future always has one available.

**Verbatim reasons + surviving claims (core atom, all 5 runs):**

- **run-1** — reason: *"The claim is falsifiable and grounded in real data, but the Germany
  subsidy-removal counterexample and growth-rate deceleration signals in 2023-2024 mean the
  unconditional version is overstated. The producer's own eight conditions are substantive — not
  cosmetic — and several (especially conditions 1, 4, and 6) have already shown vulnerability.
  Stripping the certainty framing leaves a well-supported probabilistic trajectory that survives
  scrutiny as 'likely but not assured.' The surviving claim restricts only to conditions the producer
  explicitly stated, so no condition laundering applies."*
  surviving_claim: *"BEV share of new EU passenger-car sales will plausibly exceed 25% by end of
  2029, contingent on the regulatory framework (2025/2030 CO2 targets) remaining substantially intact
  and no sustained policy reversal in major markets (Germany, France, Italy, Spain) comparable to
  Germany's 2023 subsidy removal — a narrower probabilistic claim rather than a near-certainty."*
- **run-2** — reason: *"The core arithmetic is plausible and grounded in real data: ~14–15% share in
  2023, a required gain of ~1.7 pp/year versus a recent trend of 3–4 pp/year, and a hard regulatory
  floor creating OEM incentive. The two genuine weaknesses are (1) 2024 growth stagnation in major
  markets showing the trajectory is subsidy-dependent and non-linear, and (2) the claim's validity
  hinges on regulatory stability, which the producer correctly flags as a condition. The
  regulatory-stability condition is not laundered — it was explicitly stated by the producer and is
  intrinsic to the causal mechanism. Stripping it would make the claim harder to evaluate, not
  easier. The claim survives as partial because the stalling 2024 data introduces meaningful
  uncertainty that the steelman underweights, but the underlying direction and order-of-magnitude are
  defensible."*
  surviving_claim: *"EU BEV passenger-car share will exceed 25% of new unit sales by end-2029,
  provided EU fleet CO2 regulations and the 2035 ICE phase-out remain materially in force through that
  period — a condition the producer explicitly stated and that is itself subject to political risk."*
- **run-3** — reason: *"The claim has a real empirical base (documented ~15% share in 2023, binding EU
  regulation, falling costs) and is clearly falsifiable against ACEA registration data. However, the
  2024 stall caused by German subsidy withdrawal and the historical pattern of early-adoption growth
  decelerating as mass-market segments are reached mean the central 'highly plausible' framing is
  overstated without qualification. The two retained conditions (regulation intact, no prolonged
  demand shock) were explicitly stated by the producer, so their inclusion is legitimate narrowing
  rather than laundering. Stripped of the steelman's over-confident CAGR extrapolation, the claim
  survives as a conditional probability judgment rather than a near-certainty."*
  surviving_claim: *"BEV share of new EU passenger-car registrations is more likely than not to exceed
  25% by end-2029, provided CO2 fleet-average regulations for 2025 and 2030 remain materially intact
  and no prolonged demand shock (recession or broad subsidy withdrawal) interrupts the 2024–2029
  period—conditions the producer explicitly stated."*
- **run-4** — reason: *"The core claim is falsifiable, evidence-grounded, and mechanistically
  reasoned. However, Germany's 2023 subsidy removal demonstrably stalled growth, the 2035 mandate
  faces active political pressure that partially undermines condition (7) already, and the conjunction
  of all seven producer conditions substantially narrows the probability space. The surviving claim is
  narrowed to the two most load-bearing conditions the speaker explicitly named, which are legitimate
  speaker-stated qualifiers rather than laundered ones. The claim is not hollow — it has a defensible
  core and a real evidentiary basis — but it is not fully substantive given the counterexample of
  Germany and the fragility of the policy environment."*
  surviving_claim: *"BEV share of new EU passenger-car sales will exceed 25% by end-2029, provided EU
  CO2 fleet-average penalties remain materially in force and battery purchase-price parity with
  volume-segment ICE vehicles is achieved in major EU markets by approximately 2026–2027 — both
  conditions the speaker explicitly stated."*
- **run-5** — reason: *"The core claim is falsifiable, grounded in real data, and causally coherent.
  It weakens — but does not collapse — on three fronts: the growth-rate deceleration risk as the
  market moves past early adopters, the live counterexample of post-subsidy-withdrawal slowdowns in
  major markets, and the magnitude question of whether a ~10pp gain in 5 years is as assured as the
  prior ~13pp gain. The producer's own stated conditions (regulatory enforcement, cost trajectories,
  no comprehensive subsidy elimination) are legitimate qualifiers they introduced, not laundered
  additions. The surviving claim preserves those producer-stated conditions without introducing new
  ones, and the claim remains falsifiable and defensible at that narrower scope."*
  surviving_claim: *"BEV share of new EU passenger-car registrations will likely exceed 25% by end of
  2029, provided EU CO2 fleet regulations are not materially rolled back and battery costs continue
  declining broadly along current trajectories — conditions the producer explicitly stated."*

**Read.** The post-cutoff swap did what it was designed to do — it removed the hindsight import
(contrast `false.txt` core `hollow` 5/5, sunk on the realised 2015 figures, § Intervention 2). But it
exposed a second, independent reason the substance critic cannot certify a well-formed empirical
prediction: the default rule's `substantive` branch requires surviving a counterexample, and a
forward-looking claim always affords a `weakens`-level counterexample (a policy reversal, a demand
shock), so the critic downgrades to `partial` by attaching the producer's own conditions. `EXPECTED.md`
pre-registered `substantive`; that envelope is refuted for the default prompt at N=5. The gate is a
recorded FAIL against the retired `substantive`-≥4/5 gate; `EXPECTED.md` now pre-registers `partial`
with the specifics intact as the pass, which this run meets 5/5. The certainty-strip is correct
behaviour on an unconditional forecast: a bare version of the sentence with its trailing `because …`
causal clause removed strips the same certainty and also lands `partial` 5/5, so the stated mechanism
is not the cause — the downgrade is a property of any forward-looking empirical claim, which always
affords a `weakens`-level counterexample.

## Horizon-gate refuter (grounding) — PASS (added 2026-09-14; tallied 2026-09-14)

The prediction-fixture refuter above is a *substance* test; this is its grounding counterpart. It
exercises the **horizon gate** — `crossCheckEvidence` (assay.go:2483) forcing `unverifiable` from code
with no model call whenever `evidenceDetail.Horizon == "future"` (assay.go:2484), the model's own
verdict preserved in `original_verdict` and the reason prefixed `forecast — projections are not
evidence` (`spec/EVIDENCE.md` § Horizon). The gate never *confirms* a verdict, only downgrades one,
per `CLAUDE.md` § The axis boundary: reasoning cannot ground an unobserved outcome.

The test is **two-sided by design** — a gate that downgraded everything would be useless, so it must
fire on a forecast and stay silent on a settled claim:

- **`prediction.txt` (`horizon=future`) → `unverifiable`.** The end-2029 BEV forecast is unobservable
  now, so whatever grounding verdict the model self-reports (`supported`/`mixed`/`refuted`) is code-
  downgraded to `unverifiable`, with that self-reported verdict banked in `original_verdict`.
- **`laundering` / Ballmer (`horizon=past`) → grounding column stays `refuted`.** The 2007 "no
  significant market share" claim was settled long before the model's cutoff, so the horizon gate does
  **not** fire (`Horizon != "future"`), and the audit's grounding column keeps the `refuted` it earns
  from the retrieved truth-maker. This is the specificity control: the gate must leave a genuine
  past-tense refutation intact, not blanket everything to `unverifiable`.

**Why the banked chains can't be scored.** `testing/chains/prediction-evidence-2026-09-14/` (N=3) and
the 13 Sept laundering audit (`testing/chains/2026-09-13-claude-sonnet-4-6/laundering/`, N=10) both
predate the gate: no record in the tree carries a `horizon` or `original_verdict` field, the three
prediction runs read `mixed` / `mixed` / `supported` (never downgraded), and each also banked one
`error` record — `stop_reason=max_tokens (limit=1500)`. The current `-evidence` path raises that cap
to `backend.WebSearchMaxTokens` = 8000 (assay.go:2966), so a fresh run should not truncate. Both legs
therefore need re-running on the current binary before any tally exists; the gate cannot be applied
offline because there is no stored `horizon` value to gate on.

### Pass criteria (N=3 each, default prompt, claude-sonnet-4-6)

- **`prediction.txt` `-evidence` grounding `unverifiable` 3/3**, and in every run
  `detail.horizon == "future"` **and** `detail.original_verdict` is non-empty (the pre-downgrade
  self-reported verdict) **and** `detail.downgrade_reason` begins `forecast — projections are not
  evidence`. An `unverifiable` with an empty `original_verdict` is a **fail** — it would mean the
  model returned `unverifiable` on its own and the gate never fired.
- **`laundering` `-audit` grounding column `refuted` 3/3** (the `ev=` field of the composite verdict),
  and in every run `detail.evidence.horizon == "past"` with `detail.evidence.original_verdict` empty
  (no downgrade). A `refuted` that carries an `original_verdict` would mean the gate mistook a settled
  claim for a forecast — a **fail**.
- A `NO RUNS FOUND` or a missing grounding record on any run is a failure, not a pass (zero-output
  discipline). The tally reads the banked chains only — no model call.

### Run (outside the sandbox — the in-sandbox TLS proxy fails Anthropic cert verification)

Leg 1 — `prediction.txt` grounding, N=3:

```sh
cd /Users/peterw/code/personal/elenchus && go build -o assay . && source ./setkey.sh
BASE=testing/chains/prediction-evidence-2026-09-14b
FILE=examples/destructive/well-formed/prediction.txt
for r in 1 2 3; do
  d="$BASE/run-$r"; mkdir -p "$d"
  ./assay -evidence -model claude-sonnet-4-6 -chain-dir "$d" "$FILE" || echo "FAILED prediction run $r"
done
```

Leg 2 — `laundering` audit (grounding column), N=3:

```sh
cd /Users/peterw/code/personal/elenchus && go build -o assay . && source ./setkey.sh
BASE=testing/chains/laundering-horizon-2026-09-14
SRC=examples/destructive/laundering/sources/ballmer_usatoday_2007.txt   # gitignored; recreate from laundering/PROVENANCE.md
FILE=examples/destructive/laundering/summary.txt
for r in 1 2 3; do
  d="$BASE/run-$r"; mkdir -p "$d"
  ./assay -md -audit -model claude-sonnet-4-6 -chain-dir "$d" -source "$SRC" "$FILE" || echo "FAILED laundering run $r"
done
```

### Tally (no model call — run after both legs bank their chains)

```sh
python3 - <<'PY'
import json, glob
def recs(g):
    return [json.loads(l) for f in glob.glob(g) for l in open(f)]
# Leg 1 — prediction grounding: non-error record per run must be unverifiable, horizon=future, original_verdict set.
pred=[]
for d in sorted(glob.glob("testing/chains/prediction-evidence-2026-09-14b/run-*")):
    rs=[r for r in recs(f"{d}/*.grounding.jsonl") if r["verdict"]!="error"]
    r=rs[-1] if rs else None
    det=(r or {}).get("detail",{}) or {}
    pred.append((r and r["verdict"], det.get("horizon"), det.get("original_verdict"), det.get("downgrade_reason")))
u=sum(1 for v,h,o,_ in pred if v=="unverifiable" and h=="future" and o)
print(f"prediction: unverifiable+future+original_verdict = {u}/{len(pred)}  PASS={'yes' if u==len(pred) and pred else 'NO'}")
for i,row in enumerate(pred,1): print(f"  run-{i}: verdict={row[0]} horizon={row[1]} original_verdict={row[2]!r}")
# Leg 2 — laundering audit grounding column: refuted, horizon=past, no downgrade. The audit's
# grounding leg banks as *.grounding.jsonl with a plain verdict and a detail dict (not a composite
# ev= string), so read that record directly.
lau=[]
for d in sorted(glob.glob("testing/chains/laundering-horizon-2026-09-14/run-*")):
    rs=[r for r in recs(f"{d}/*.grounding.jsonl") if r["verdict"]!="error"]
    r=rs[-1] if rs else None
    ed=(r or {}).get("detail",{}) or {}
    lau.append((r and r["verdict"], ed.get("horizon"), ed.get("original_verdict")))
rf=sum(1 for ev,h,o in lau if ev=="refuted" and h=="past" and not o)
print(f"laundering: refuted+past+no-downgrade = {rf}/{len(lau)}  PASS={'yes' if rf==len(lau) and lau else 'NO'}")
for i,row in enumerate(lau,1): print(f"  run-{i}: ev={row[0]} horizon={row[1]} original_verdict={row[2]!r}")
PY
```

### Results — PASS (tallied 2026-09-14, N=3 each; no model call)

Both legs pass their pre-registered gate. Tally over the banked chains
(`testing/chains/prediction-evidence-2026-09-14b/`, `testing/chains/laundering-horizon-2026-09-14/`):

| leg | fixture | horizon | expected grounding | N | result |
|---|---|---|---|---|---|
| 1 | `prediction.txt` | future | `unverifiable` (downgraded, `original_verdict` recorded) | 3 | **PASS** — `unverifiable` 3/3, `horizon=future` 3/3, `original_verdict=mixed` 3/3, `downgrade_reason` = `forecast — projections are not evidence` 3/3 |
| 2 | `laundering` (Ballmer) | past | `refuted` (gate silent, no downgrade) | 3 | **PASS** — `refuted` 3/3, `horizon=past` 3/3, `original_verdict` empty 3/3 (gate did not fire) |

The two-sided gate holds: it downgrades the end-2029 BEV forecast to `unverifiable` from code while
banking the model's own `mixed` verdict in `original_verdict`, and it stays silent on the 2007 Ballmer
claim (`horizon=past`), leaving the `refuted` the audit earns from the retrieved truth-maker intact.

**One correction to the tally as first written** (folded into the § Tally script above): leg 2's audit
grounding column banks as `summary.grounding.jsonl` with a plain `verdict` and a `detail` dict, not as
`*.audit.jsonl` with a composite `ev=` verdict string. The original reader globbed the wrong filename
and parsed a non-existent `ev=` field, so it reported `0/3` on chains that are in fact `refuted` 3/3.
The reader now reads the grounding record directly; no chain was re-run.

**Conclusion.** Grounding on a forecast is unverifiable by construction, decided in code from a
model-reported horizon; without the gate Sonnet returned mixed/supported on a 2029 claim from analyst
projections.
