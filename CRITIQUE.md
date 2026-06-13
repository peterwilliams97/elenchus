# CRITIQUE.md

A doc-level critique of this repo's claims, read through analytic philosophy from
Frege forward. This is a review of documentation, not of code. It is not a bug
list: several items here are in-principle limits, not defects awaiting a patch.
The tell throughout is that PLAN.md §3's own open problems are these critiques
resurfacing as engineering work — §3a is the rule-following paradox, §3c is the
context principle, §3d and the W2/W3 limits are the Myth of the Given.

## The atomism gap
The load-bearing first step is `decompose` into atomic claims, then grading
fragments. Frege's context principle — never ask for the meaning of a word in
isolation, only in the context of a proposition — says atomic content is
context-relative. Quine's confirmation holism says claims meet evidence only as a
corporate body, never individually, so "is THIS atomic claim grounded?" is
malformed at the root. The Tractatus was the high-water mark of logical atomism,
and Wittgenstein himself retracted it. §3c (the critic reading "columns" as
referent-less, "assay" as a chemical assay) is the context principle biting; a
one-line anchor patches the symptom while the architecture keeps committing the
error elsewhere, silently.

## The analytic/substance vs. synthetic/grounding gap
The substance/grounding split — well-formed-and-falsifiable on one side, true on
the other — is the analytic/synthetic distinction wearing a binary. Quine's "Two
Dogmas" denies the distinction can be drawn in principle: falsifiability and
well-formedness already carry synthetic background commitments. The line
"survives the challenges → well-formed and falsifiable; not thereby true" is
exactly the line Quine says isn't there. Davidson sharpens it: the producer's
"strongest defensible version" IS the principle of charity, and charity is
constitutive of interpretation — you cannot fix what was said (faithfulness)
without already taking a stand on what is true (grounding). So the steelman bleeds
grounding into substance, violating the repo's own column-separation rule. The
one Quinean move the tool does make (distributional verdicts) is in tension with
the atomism and the analytic/synthetic split it also needs.

## The rule-following gap
The seven axes are rules, and Wittgenstein (PI §201; Kripke's reading) shows no
rule contains the rules for its own application. §3a is this paradox in the wild:
the Counterexample axis fires fatally on probabilistic forecasts because "a
conceivable counterexample is always constructible" — the rule underdetermines
its own application and gives no guidance. No prompt redesign closes this, because
application requires a shared practice the rule cannot encode. (Popper hit the
same wall: probabilistic statements are not strictly falsifiable.)

## The Given gap
This is where the repo is proudest and weakest at once. "Positive grounding is a
retrieval — the actual truth-maker, a page that backs the sentence" is Sellars'
Myth of the Given: a non-inferential foundation that grounds without itself
standing in the space of reasons. The page is another claim needing its own
justification (Russell's regress; Tarski's right-hand side of the T-schema is the
world, not a document). Reading the citation block better (§3d) yields a better
representation, never a Given. The honest conclusion the repo is groping toward
but has not stated: the third column may be unreachable IN PRINCIPLE — not by
armchair, and not by retrieval either.

## The psychologism gap
"Verdicts as distributions" relocates a logical/normative property (is this claim
well-formed?) into the statistical dispositions of a stochastic judge, then
reports judge-reliability as if it were a property of the claim. Frege's whole
campaign was against this. "What the claim tends to do under examination" does not
escape it: the distribution is over THIS model's behavior, and the bridge to a
claim-property is itself an armchair confirmation — the very move grounding mode
forbids.

## The force / speech-act gap
There is no machinery for illocutionary force or implicature. Faithfulness treats
utterances as constatives with truth-values to preserve, but Frege's judgment
stroke distinguishes content from the act of asserting it; Austin distinguishes
locution, illocution, perlocution; Grice distinguishes the said from the meant.
Ballmer's "no chance" may be bravado dismissing a competitor, not a literal
probabilistic assertion — grading it "refuted" on truth-conditions misreads a
performative as a constative, and a faithful summary can preserve what was said
while dropping what was implicated.

## The externalism gap
Putnam: meaning ain't in the head; the division of linguistic labor fixes
specialist content by experts the critic cannot defer to. Kripke: reference is
fixed by a causal-historical chain, not by description, so decontextualised
fragments sever it and a descriptive anchor may not restore it. The README's own
warning that "specialist domains deserve the most suspicion" is this gap surfacing.

## The reflexive thesis is itself ungrounded
THEORY.md's foundational claim — reasoning can refute but never confirm, positive
grounding is retrieval, "not a contingent limitation... the limit of what close
reading and dialectic can do" — is an a priori claim about the limits of
reasoning. By the repo's own taxonomy it lives in the SUBSTANCE column and has
never been grounded, yet it is asserted with the authority of settled truth. That
is the repo's own definition of laundering applied to its load-bearing sentence.
The reflexive canary tests the mirror hypothesis; it never turns on the separation
thesis itself. A self-collapse eval that bites should point there.

## Options for the future
These gaps split into two kinds, and the split is the decision.

In-principle, not patchable: the Given gap (grounding-as-retrieval), the
analytic/synthetic separation as a universal claim, holism vs. per-fragment
grounding. No prompt or fixture closes these; pretending otherwise reintroduces
the laundering the tool exists to refuse.

Tractable by narrowing scope: the rule-following, force, externalism, and
referent gaps are failures of the OPEN-world ambition. They shrink or vanish when
the input domain is closed and the claim types are restricted. A narrow tool —
fixed domain, anchored referents, claim types where the axes genuinely apply (e.g.
internal-contradiction refutation, attribution-fidelity over a known source) — can
be rigorous precisely because the context principle and externalism stop biting
when context is supplied by construction.

Two strategic readings follow. (1) Reduce scope: apply the producer–critic method
to a narrow problem class where the gaps are designed out rather than apologised
for. (2) Drop the demand for complete answers: reposition the tool as producing
partial results and targeted improvements — refutation and faithfulness flags it
CAN earn — and stop claiming the grounding column at all, or mark it permanently
provisional. The current docs promise complete, separable, three-column
adjudication; the defensible product is a sharp instrument for the columns
reasoning can actually reach.
