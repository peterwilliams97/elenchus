# SESSION.md

## 2026-06-13 — doc-level philosophy critique recorded (consolidation, docs only)

This session installed a critic against the repo's own documentation and recorded its findings
faithfully rather than answering them. **CRITIQUE.md** (new, at repo root) holds nine sections
reading THEORY.md / README.md / PLAN.md through analytic philosophy from Frege forward — atomism,
analytic/synthetic separation, rule-following, the Myth of the Given, psychologism, speech-act
force, externalism, and the charge that the repo's own load-bearing thesis ("reasoning can refute
but never confirm") is itself an ungrounded substance-column claim. Seventeen inline
`([critique](CRITIQUE.md#…))` links were wired from the most questionable claims back to it (8 in
THEORY.md, 6 in README.md, 3 in PLAN.md); all anchors verified to resolve. The decision was
pre-registered before any edit as **d018** in `rigour-map/decision_log.jsonl` (this is the first
entry in elenchus2's own log — schema carried from v1's `rigour-map/decision_log.jsonl`, which the
user supplied for schema only and which was not modified). No code, test, prompt, or spec/ change;
spec/ confirmed clean.

**The split that is the decision.** The critique sorts into two kinds. *In-principle, not patchable:*
the Given gap (grounding-as-retrieval can never deliver a non-inferential truth-maker), the
analytic/synthetic separation taken as a universal claim, and holism vs. per-fragment grounding —
no prompt or fixture closes these, and pretending otherwise re-introduces the very laundering the
tool exists to refuse. *Tractable by narrowing scope:* the rule-following, speech-act-force,
externalism, and referent gaps are failures of the OPEN-world ambition; they shrink or vanish once
the input domain is closed and claim types are restricted, because context-dependence and
externalism stop biting when context is supplied by construction. Notably, PLAN.md §3's existing
open problems (§3a rule-following, §3c the context principle, §3d/W2–W3 the Given) are these same
critiques already surfacing as engineering work.

**Two options now on the table.** (1) **Reduce scope** — apply the producer–critic method to a
narrow problem class (fixed domain, anchored referents, claim types where the axes genuinely apply:
internal-contradiction refutation, attribution-fidelity over a known source) where the gaps are
designed out rather than apologised for. (2) **Drop the demand for complete answers** — reposition
the tool as producing partial results and targeted improvements (the refutation and faithfulness
flags it CAN earn), and either stop claiming the grounding column or mark it permanently
provisional. The current docs promise complete, separable, three-column adjudication; the
defensible product is a sharp instrument for the columns reasoning can actually reach. No option is
chosen here — recording the fork was the deliverable.

### Carry-forward / flags for the next session
- **Unrelated working-tree change present:** `build.sh` carries a `staticcheck ./...` step that was
  NOT authored this session and is NOT part of this doc-only change. It was left untouched and
  unbundled (one-logical-change-per-commit). Commit or revert it on its own.
- **Anchor slug note:** the heading "The analytic/substance vs. synthetic/grounding gap" slugs to
  `#the-analyticsubstance-vs-syntheticgrounding-gap` (lowercase, punctuation dropped, spaces→hyphens
  — the `/` is removed, not hyphenated). The links use this heading-derived slug, which differs from
  the originating instruction's hyphenated table form; the heading was preserved verbatim, so the
  links were corrected to match it (the one permitted slug repair).
- **An Author's response was NOT added** to CRITIQUE.md. The critique body stands unanswered by
  design; if a response is ever wanted it must be fenced, last, and must not soften or shorten the
  body.
