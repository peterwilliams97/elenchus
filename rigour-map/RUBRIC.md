# Rigour-Application Rubric

Classifies the dialectical-rigour methodology per
**(business_area × mode)** → `{advantage | disadvantage | irrelevant}`. The methodology is three
filters with different domain-fit, so the unit of classification is the (area, mode) cell, and the
per-area result is the 3-vector. A flat per-area label averages away the real signal — don't use one.

The label is *earned from an assay run*, never asserted. Inputs are the seven verified fixtures in
`fixtures/raw/` (provenance in their headers). A populated label table is not a result until every
cell cites the run output it came from.

## What each mode's verdict distribution looks like, per label

Read the label off three things: (1) the verdict spread, (2) your spot-check overrule rate against
the mode's oracle, (3) whether the column disagrees informatively with the others in the audit
cross-tab. Cost/yield breaks ties.

**ADVANTAGE** — the mode earns its cost:
- verdicts are *spread* (not collapsed to one value), AND
- on spot-check they track the mode's oracle (grounding→world; faithfulness→source; substance→a
  careful reader), i.e. low human-overrule rate, AND
- at least one verdict overturns a "looks-fine" reading — catches a false / overstated / hollow
  claim that would otherwise pass, AND
- in the cross-tab it contributes a column that disagrees *informatively* with the others
  (e.g. `faithful + substantive + refuted`).

**IRRELEVANT** — runs clean, adds nothing:
- verdicts collapse to a single value (all `substantive`; or all `unverifiable` because the content
  is prediction/intention/stipulation), OR
- the column is entailed by another (redundant), OR
- a stronger external process already owns this validation (audit, courts, a native 5-Whys) and the
  mode merely restates it.

**DISADVANTAGE** — actively misleads or burns cost for ~zero yield:
- high human-overrule rate AND the errors point the wrong way — the critic over-attacks rhetoric,
  literalizes puffery, or mis-frames a stipulation/definition as an empirical claim, producing false
  alarms; OR a high `unverifiable` rate gets read as endorsement (grounding on predictions = false
  comfort), OR
- cost is high (faithfulness re-sends the source per claim; grounding is one search per claim) and
  yield ≈ 0.

## Oracles (what you spot-check each verdict against)

- **grounding** — the world. Verify a sample against primary sources; read URLs from the API
  citation blocks, not the printed list (the model retypes those). `unverifiable` on a
  prediction/intention is correct, not a miss.
- **faithfulness** — the source text. Needs `(source, summary)` pairs; see `fixtures/summaries/`.
- **substance** — a careful human reader. No external oracle; report inter-rater agreement, not
  accuracy. This column is the softest; say so.

## Reading procedure

1. Run `rigour-map/run_eval.sh` (Haiku first for cost — accounting's 726 segments dominate).
2. For each (area, mode) cell: record verdict distribution, spot-check N claims against the oracle,
   note overrules.
3. Build the audit cross-tab for the areas that have summary probes; mark the informative
   disagreements.
4. Re-run the default model only on cells where the Haiku verdict was borderline or you overruled it.
5. Label each cell by the definitions above, citing the run file + the claims that decided it.
6. Diff against the SEALED PRIORS below. The cells where the run overturns the prior are the
   evaluation's real output — and each overrule becomes a canned-JSON stub test in the `cfg.call`
   seam (same shape as `TestConditionLaunderingDowngrade`).

---

## SEALED PRIORS — pre-registered hypotheses

> Do NOT consult while labeling. These are predictions read off the fixture *structure* before any
> run, sealed to protect the disagreement baseline. Open after labeling and diff. Authored
> 2026-06-01, pre-run.

Headline: grounding-fit tracks "present-tense world-facts"; substance-fit tracks "causal/economic
reasoning"; faithfulness is the broadest win, tracking "representation diverging from a source."

| area                       | faithfulness                                | substance | grounding | load-bearing mode |
|----------------------------|---------------------------------------------|-----------|-----------|-------------------|
| legal (DMCA notice)        | ADVANTAGE                                   | DISADVANTAGE (performative "good-faith belief" isn't falsifiable → mis-frame) | IRRELEVANT (truth-oracle is courts, not the world) | faithfulness |
| accounting (shareholder letter) | ADVANTAGE (mgmt narrative vs filing)   | IRRELEVANT (claims trivially well-formed) | ADVANTAGE (grounds the *narrative* spin the audit doesn't police) | grounding |
| sales (YC pitch)           | ADVANTAGE (overstatement is the whole game) | DISADVANTAGE (over-attacks vision/rhetoric) | DISADVANTAGE (mostly `unverifiable` predictions → false comfort) | faithfulness |
| marketing (press release)  | ADVANTAGE (quotes + literalization)         | ADVANTAGE (superlatives like "only/best" are falsifiable) | ADVANTAGE (greenwashing specs are verifiable) | all three (richest disagreement) |
| pm (EIP-1559)              | IRRELEVANT (no source as fetched)           | ADVANTAGE (causal/economic claims; stresses causality axis) | MIXED (was prediction in 2019; post-hoc checkable now) | substance |
| engineering (postmortem)   | IRRELEVANT (primary doc, no source)         | ADVANTAGE *at the margin over the native 5-Whys* | IRRELEVANT (internal past event, no public truth-maker) | substance |
| contract (agreement)       | ADVANTAGE (plain-English summary vs obligation) | DISADVANTAGE/IRRELEVANT (stipulations aren't empirical claims) | IRRELEVANT (clauses aren't world-facts) | faithfulness |

Secondary predictions to watch:
- The "DISADVANTAGE" cells are the falsification targets — if the substance critic *doesn't*
  over-attack the sales pitch, the prior is wrong and that's the most informative result.
- marketing should be the single best demonstrator (all three columns active and disagreeing). If it
  isn't, the methodology's value claim is weaker than assumed.
- engineering vs contract probes the "stronger process already owns it" sense of IRRELEVANT.
