# elenchus

Three questions, kept separate.

## What it does

crossexam takes prose — a summary, a transcript, a list of claims — and splits it into atomic claims. ([critique](CRITIQUE.md#the-atomism-gap))
Of each claim it asks three separate questions:
did the source actually say it (**faithfulness**),
is it well-formed and falsifiable (**substance**),
and is it actually true (**grounding**).
It refuses to merge the answers — each question gets its own mode and its own verdict, and a win on
one is never spent on another. ([critique](CRITIQUE.md#the-analyticsubstance-vs-syntheticgrounding-gap))

## Why

The target failure is confidence laundering:
real wins on the questions reasoning can reach, spent as authority on the one it can't.
Reasoning can verify attribution and structure, and it can refute a claim by deduction —
but it cannot confirm one.
Positive grounding is a retrieval, not a deduction ([critique](CRITIQUE.md#the-given-gap)), however rigorous.
A system that runs faithfulness and substance checks and then pronounces on truth from the armchair
has crossed that line; the tool is built to refuse the move, including about itself. ([critique](CRITIQUE.md#the-reflexive-thesis-is-itself-ungrounded))
The full argument is in [THEORY.md](THEORY.md).

## Show it

One claim, three verdicts (`claude-haiku-4-5-20251001`, 2026-06-10, N=10):

| Claim | Faithful? | Substantive? | Grounded? |
|---|---|---|---|
| "There's no chance that the iPhone is going to get any significant market share." — Steve Ballmer, USA TODAY CEO Forum, 2007 | faithful 10/10 | hollow 10/10 | refuted 10/10 |

The summary reported Ballmer accurately, and what he said was false — two different facts, kept
in different columns. The `hollow` in the middle is not a verdict to trust: it is the tool's own
measured bias on forward predictions ([LIMITS.md](LIMITS.md#known-biases-and-bugs)), printed
rather than hidden. That honest row is the whole demonstration.

## What it can't do

The critic is a language model, not an oracle: a blind spot in the producer survives in the
critic, and specialist domains deserve the most suspicion. ([critique](CRITIQUE.md#the-externalism-gap)) Verdicts are distributions, not
facts ([critique](CRITIQUE.md#the-psychologism-gap)) — a single run is one draw, and the calibration protocol reports full distributions per
probe. `supported` is the structurally weakest verdict: it currently proves a URL was retrieved,
not that the page backs the sentence. We publish the probes that break the tool and the measured
envelope they map: [LIMITS.md](LIMITS.md).

## Status

The binary is not built yet. The build sequence and acceptance test are in [PLAN.md](PLAN.md);
until that passes, the commands below are the contract the binary must satisfy, not a
description of working software.

## Usage

```sh
export ANTHROPIC_API_KEY=sk-...   # required; fatal if unset
go build -o crossexam ./cmd/crossexam
```

The three modes, and the cross-tab that runs all of them:

```sh
./crossexam claims.txt                                  # substance — is each claim well-formed?
./crossexam -source transcript.txt summary.txt          # faithfulness — did the source say it?
./crossexam -evidence claims.txt                        # grounding — is it true, per external evidence?
./crossexam -audit -source transcript.txt summary.txt   # all three, cross-tabulated
```

Results go to stdout; progress and the SUMMARY block go to stderr. Verdicts by mode — substance:
`substantive / partial / hollow`; faithfulness: `faithful / partial / overstated / absent /
contradicted`; grounding: `supported / mixed / refuted / unverifiable`. `-md` emits markdown
tables (audit always does). Default model is `claude-sonnet-4-6`; override with `-model` or
`ANTHROPIC_MODEL`. Input is `-text STRING`, a file argument, or stdin. A per-claim verification
chain is written as JSONL under `eval/<stamp>-<model>/` by default (`-chain-dir` to relocate).
Full flag and output reference: [spec/CLI.md](spec/CLI.md).

## Examples

[examples/dan_shipper/](examples/dan_shipper/) shows the tool working end-to-end on a real
transcript. [examples/destructive/](examples/destructive/) shows where it bends, on purpose —
eight adversarial probes with dated verdict distributions.

## Who it's for

People trained to separate attribution from structure from truth — and who want to see that
skill operating on material from their own work. That seeing the three columns disagree causes
such a user to recognise the skill as transferable is a hypothesis awaiting field evidence, not
a finding ([THEORY.md](THEORY.md)).
