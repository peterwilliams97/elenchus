# THEORY.md

## The axis boundary

Three questions can be asked of any claim in a text:
did the source actually say it (**faithfulness**),
is it well-formed and falsifiable (**substance**), and
is it actually true (**grounding**)?
The tool answers each in a separate mode and keeps the answers separate ([critique](CRITIQUE.md#the-analyticsubstance-vs-syntheticgrounding-gap)), because the modes reach
different things.

Close reading reaches faithfulness.
The faithfulness critic's task is sense-preservation, not truth ([critique](CRITIQUE.md#the-force--speech-act-gap)):
whether the summary reports the speaker accurately, never whether the speaker was right.
A faithfully reported claim can still be wrong; that question belongs to a different column.

Dialectic reaches substance.
The producer-critic loop tests whether a claim survives seven named challenges ([critique](CRITIQUE.md#the-rule-following-gap)) —
evidence,
hidden premise,
falsifiability,
equivocation,
base rate and magnitude,
counterexample,
causality versus correlation —
applied to the strongest defensible version of the claim.
A claim that survives is well-formed and falsifiable.
It is not thereby true. ([critique](CRITIQUE.md#the-analyticsubstance-vs-syntheticgrounding-gap))

Reasoning can refute a grounding claim without a lookup: an internal contradiction
kills a claim by deduction alone. But reasoning cannot confirm one. Positive grounding
is a retrieval — the actual truth-maker, a page that backs the sentence ([critique](CRITIQUE.md#the-given-gap)) — not a
deduction, however rigorous. That boundary is not a contingent limitation of the
design; it is the limit of what close reading and dialectic can do. ([critique](CRITIQUE.md#the-reflexive-thesis-is-itself-ungrounded))

The failure this guards against is laundering: confidence earned on the reachable
columns — faithfulness and substance — spent on the unreachable one. A system that
runs faithfulness and substance checks and then pronounces on truth from the armchair
has crossed the line the third mode exists to police. The tool is built to refuse that
move, including about itself.

---

## The producer and the limit of separation

In substance mode, two calls run in sequence.
The producer is given a claim and asked for its strongest defensible version and the conditions
required for it to hold.
The critic then attacks that version across seven named axes.
The producer's prompt withholds the axes entirely: it asks only for the best honest case, without
naming what the critic will test. A critic the generator can anticipate is worthless — it shapes
the output to survive rather than to survive scrutiny.

The calls also run with independent context. The critic receives the producer's
steelman as its input — it sees the framing — but it runs in a fresh context and did
not author what it grades. This removes self-defense bias: the critic is not evaluating
a position it has already committed to.

The limit of this design is as sharp as the design itself. An independent critic still
holds no truth-maker. Separating production from critique tests whether the strongest
version of a claim survives named structural challenges, not whether that structure
corresponds to anything in the world. This is the difference between this tool and
verifier systems graded against an external score or database. Separation buys honesty
about what the claim asserts. It cannot buy truth about whether the assertion holds. ([critique](CRITIQUE.md#the-analyticsubstance-vs-syntheticgrounding-gap))

---

## Verdicts as distributions

The judge is a language model, and language models are non-deterministic. A single run
on a single claim yields one sample from a distribution. Reporting that sample as a
stable verdict would be the tool making the move it exists to catch — asserting a claim
without the evidence needed to support it.

The calibration data makes this concrete. The motte-and-bailey probe ran ten times on
the same engineered defect. In nine runs, the defect-carrying material was caught —
rated hollow or partial. In run 10, decompose itself behaved differently — it kept the
equivocation conditional whole, a fragment the other runs never produced — and the
critic passed it with an empty critique. Variance enters at decomposition as well as at
critique; the false-pass rate (1 in 63 valid verdicts) is a property of the
distribution, not of the claim.

The calibration protocol is designed around this: N runs per probe, fragment-level
distributions reported in full, each false pass examined for its signature. A verdict
from a single run on a single claim should be read as one draw. Repeated runs shift the
question from what the model said to what the claim tends to do under examination. ([critique](CRITIQUE.md#the-psychologism-gap))

---

## The mirror hypothesis

The tool carries a claim about its own effect: that seeing the three columns disagree
causes a trained user to recognise their skill as transferable. The mechanism is
recognition, not instruction. The user already has the skill — a trained capacity to
separate attribution from structure from truth — and the tool produces a context in
which they can see it operating on something from their actual work. The columns do not
install the skill; they make it visible to the person who holds it.

The test condition is field evidence from trained users, not from people who built the
tool. The reflexive canary (unverifiable 5/5, haiku, 2026-06-11) confirms only that the
grounding module does not currently confirm this claim from the armchair — that is the
correct result for a hypothesis at this stage, not a positive finding. The claim is
carried as a hypothesis awaiting field evidence.
