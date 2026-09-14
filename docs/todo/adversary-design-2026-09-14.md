# Adversary design for elenchus over free-text reports (2026-09-14)

Source: chat with Claude, 14 Sept 2026. Seed, not decision.

## Compact representation of the attack class

Every attack on a claim-plus-support is one of three (Pollock 1987, defeasible reasoning):

1. Premise false — the finding isn't in the source or isn't what the source says. 
   Grounding: retrieval, not argument.
2. Undercutting defeater — premise true, doesn't reach the conclusion. All the informal fallacies 
   live here: equivocation, hidden premise, base rate, correlation-as-cause, gamed reference class, 
   survivorship, composition.
3. Rebutting defeater — independent reason the conclusion is false: counterexample, contradiction 
   elsewhere in the report, known outcome (which is grounding's job — the axis leak seen on Ballmer).

The seven substance axes are (2) and (3) unfolded. Aristotle's thirteen, Toulmin's six slots and 
fallacy lists are the same three at different resolutions.

## Operational form: Walton schemes + critical questions

A scheme is an inference pattern with a fixed short list of critical questions. The adversary names 
the scheme, then asks its questions.
Consultant reports use about five:

- Practical reasoning (goal G, means M → recommend M): is G actually held? alternative means? 
  side effects? is M feasible? does G conflict with other goals?
- Position to know / survey (N% say X → X): are respondents in a position to know? 
  is the sample the population being recommended to? was the question asked the claim being made?
- Example → general: is the case typical? how many counter-cases were looked for?
- Sign / trend: does the indicator track the thing? does the trend extrapolate?
- Classification: does the definition used here match the one that carries the conclusion?

## Two design rules for the tree

1. Attack edges, not leaves. Leaves are grounding. The adversary works at finding → recommendation 
   edges. First move on loose prose is reconstruction: "for R to follow from F you need W" — the 
   enthymeme made explicit — then name the scheme and ask its questions. Hidden-premise probe once 
   per edge, not once per claim.

2. Charge for attacks. Equivocation and Hidden premise fire on every bare
   general claim because they are free (boundary diagnosis, 14 Sept). Rule:
   an attack is admitted only if it is a named defeater — a concrete state
   of the world, plausible in the report's domain, under which the premise
   holds and the conclusion fails — and it states what source or
   observation would settle it. "Could be equivocating" is not admitted;
   "'security incident' means any helpdesk ticket on p.14 and a breach on
   p.31" is. This is the name-not-count rule as the entry condition, and it
   is what gives the `substantive` class room to exist.

## Related

- `docs/todo/audit-adversarial-2026-09-13.md` — where the adversarial (substance) leg is dormant on
  the report corpora, and the three real `faithful` leaves whose verified quote is present but whose
  claim doesn't follow (the finding → recommendation gap this note attacks at the edge).
- `docs/todo/destructive-sonnet-2026-09-13.md` — the calibration and the three failed interventions
  (`-ce-scoped`, `-narrowing-boundary`, `-as-of`). Its "free fatal on any general claim"
  (Equivocation + Hidden-premise spread across ≥2 axes) is the boundary diagnosis design rule 2
  answers with a named-defeater entry condition, and its Ballmer Counterexample-vs-grounding leak is
  the rebutting-defeater case in the attack-class list above.
- `spec/ARGUMENT.md` — the finding → recommendation → root structure whose edges design rule 1
  attacks; scheme reconstruction ("for R to follow from F you need W") is the `?`-edge / enthymeme
  made explicit.
- `examples/destructive/README.md` — the worked adversarial probes these schemes generalise
  (motte-and-bailey, reference-class, hidden-premise, unfalsifiable-dress, causal-narrative,
  axis-gaps, laundering).

## Applicability (2026-09-14)

Read-only over the five `argument.txt`. F→R = finding→recommendation edge (what this note attacks);
`?` = report never established the edge; childless-? recs name no finding at all.

### 1. F→R edges and verdicts today

| corpus | solid F→R | `?` F→R | childless-? recs | verdict on any edge today |
|---|---|---|---|---|
| vic-lceic | 9 | 5 | 2 (R6, R10) | 0 |
| quocirca | 9 | 0 | 3 (S1, S4, S5) | 0 |
| ai-index | 0 | 0 | 0 (no recs) | 0 |
| dora | 8 | 0 | 1 (R-TRANSFORM) | 0 |
| master-plan | 15 | 0 | 0 | 0 |

46 F→R edges, **0 carry a verdict** — as expected. Even after a judge run the derivation is a
*leaf-faithfulness conjunction/tally propagated up* (spec/ARGUMENT.md § Internal judgement); nothing
asks "does R follow from F." The tool computes hold/fail over whether the *finding* is faithful,
never over the *edge* — that inference is the gap.

### 2. Recommendations by Walton scheme

- **Practical reasoning** dominates the three rec-shaped reports — every dora capability (8), every
  vic-lceic advocate/index/reduce (9), every quocirca "do X" (9). Ex: `R-DATA` ← `F-DATA`.
- **Sign/trend** is the *finding* tier under those recs: dora's "X amplifies" findings are
  associational; ai-index `F-TB` (20%→77.3%). Ex: quocirca `B2` ← `KF5b` (maturity → data-loss).
- **Position-to-know / survey**: quocirca is a survey, so its findings are survey-typed even where
  the rec is practical. Ex: `B1` (elevate to enterprise risk) ← `KF2`.
- **Example→general** (and one **classification**, `REC9`): the whole master-plan risk — 12 universal
  rules generalised from single internal anecdotes. Ex: `REC3` (no shared code) ← `F4` (one `toInt32`).

### 3. Three edges with an available named defeater (admitted form)

- **dora `R-PLAT` ← `F-PLAT`** (side-effect CQ; the finding names its own defeater). World: an org
  whose binding constraint is delivery stability (regulated/safety-critical). F holds — platform
  amplifies performance *and* raises instability, per F-PLAT's own hedge — yet R (invest, it's the
  prerequisite) fails there. Settles it: F-PLAT's instability effect size vs. the org's own weighting.
- **dora `R-STANCE` ← `F-STANCE`** (correlation-as-cause; applies to all 8 dora edges). World:
  AI-mature orgs adopt a stance *because* already high-performing (common cause). F holds as an
  association; a policy in a struggling org moves nothing → R fails. Settles it: whether the report's
  number is observational or from an intervention — a causal estimate, not a regression.
- **master-plan `REC3` ← `F4`** (example→general, is the case typical / N=1). World: the `toInt32`
  anecdote holds, but shared-code checkers over a stable stdlib parser catch bugs it can't speak to →
  "must not share" is over-broad. Settles it: of shared-code checkers examined, how many hid vs caught.

### 4. What the tool cannot do today that this needs

- **Edge-level judge call.** Only model call is per-leaf `faithJudge` (assay.go); nothing takes
  (finding, recommendation, scheme) → "does R follow, and name a defeater." New pass, new prompt.
- **Scheme tag on the edge.** `argument.txt` edges carry only load-bearing/`?` + a note; no
  Walton-scheme field, so parser and format need one before the questions can be asked.
- **Admission check in code.** The named-defeater condition (concrete world + what settles it) is
  prose only; no struct holds a defeater, nothing rejects "could be equivocating."
- **A `substantive` verdict class.** hold/weakened/open/fails all derive from leaf *faithfulness*;
  no slot for "premise true, inference defeated" — design rule 2's class has no home yet.

**Net.** Highest leverage on **dora** (8 uniform correlation→intervention edges) and **master-plan**
(12 rules from anecdotes); moderate on vic-lceic/quocirca (many `?` edges the move can't attack — no
stated edge, only a question for a human); **near-zero on ai-index** (no recs, 0 F→R). And the
edge-attacker can only *refute/open* an edge, never certify R follows (the axis boundary).
