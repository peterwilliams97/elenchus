# assay

A dialectical filter for claims. It takes prose — a strategy memo, a pundit's "predictions," a
summary of a talk — breaks it into atomic claims, and tests each one. It doesn't tell you what to
think; it makes the structure of a claim visible so you can decide.

The tool answers three **different** questions, one per mode, and it lets the answers disagree:

- **Faithfulness** — did the source actually say this? *(sense / attribution)*
- **Substance** — is the claim well-formed and falsifiable, the kind of thing that *could* be true? *(structure)*
- **Grounding** — is it actually true, against external evidence? *(reference / truth-makers)*

A claim can be faithfully reported, internally well-reasoned, and still false. Keeping those three
apart is most of what careful reasoning requires, and the tool is built to keep them apart on
purpose rather than collapsing them into a single "is this good?"

## Install

```bash
pip install anthropic
export ANTHROPIC_API_KEY=sk-ant-...
```

Python 3.9+. A `setup.sh` is included as a one-shot convenience for the above. Optionally pin a
default model:

```bash
export ANTHROPIC_MODEL=claude-opus-4-8
```

Input is a file argument, `--text "..."`, or piped on stdin.

## Usage

### Substance (default) — is the claim well-formed?

```bash
python3 assay.py memo.txt
```

Decomposes the prose into atomic claims, then runs a **producer–critic** dialectic on each. A
Producer states the strongest version of the claim, *blind to the critique axes*; a Critic attacks
it on seven fixed axes; the loop repeats on the surviving claim until the Critic raises no new fatal
 or weakening finding, or `--max-rounds` is hit. Verdict: `substantive` · `partial` · `hollow`.

Fixed critique axes:
 Evidence · Hidden premise · Falsifiability · Equivocation · Base rate / magnitude · Counterexample · Causality vs correlation.

### Faithfulness — did the source actually say it?

```bash
python3 assay.py summary.txt --source transcript.txt
```

For each claim in the summary, a **Defender** finds verbatim support in the source and a **Critic**
judges whether the summary represents it accurately, watching for overstatement, fabrication,
distortion, context-stripping, misattribution, and cherry-picking. Verdict:
`faithful` · `partial` · `overstated` · `absent` · `contradicted`.

This checks attribution — *whether the speaker said it* — never whether it's true. A
faithfully-reported claim can still be wrong.

### Grounding — is it actually true?

```bash
python3 assay.py claims.txt --evidence
```

For each claim, Claude searches the web for real evidence — the actual truth-makers, not anyone's
assertion that the claim is true — weighs what it finds, and returns
`supported` · `mixed` · `refuted` · `unverifiable`, with the sources used.
This is the only mode that touches truth.

### Chaining the three

To critique a pundit properly, run the modes in order: **faithfulness** (are we critiquing what they
actually said?) → **substance** (is it well-formed and falsifiable?) → **grounding** (is it borne
out by evidence?). Each answers a question the previous one can't.

## Flags

| flag | effect |
|---|---|
| `path` / `--text` / stdin | the input prose |
| `--source FILE` | faithfulness mode: check the input against FILE |
| `--evidence` | evidence-grounding mode (web search) |
| `--model ID` | model id (default `claude-sonnet-4-6`) |
| `--max-rounds N` | producer–critic rounds per claim (substance only, default 2) |
| `-v`, `--verbose` | show every Claude call: prompts and raw streamed response |
| `--no-color` | disable ANSI colour |

## Output

By default you see the dialectical *result* of each step — the steelman, the critique by axis, the
 verdict — plus a summary at the end. This is "what a careful reader would conclude." `-v` instead
 shows every Claude call as it happens: system prompt, user prompt, and the response streamed live —
 "how the machine got there." Pipe verbose to a file (`> log.txt`) when you want the full trace.

## Honest limits

- **The Critic is itself an LLM.** Its precision varies; it can wave through a hollow claim or
over-attack a sound one. The skill the tool is really training is catching where the machine critic
is wrong — don't treat its verdict as final.
- **Predictions come back `unverifiable` under `--evidence`, by design.** You can't ground a claim
about the future in present evidence. That's the correct result, not a failure — it's the line
between "falsifiable in principle" (which substance mode checks) and "settled by current evidence."
- **Source URLs in `--evidence` are written by the model** from its search results. They should be
real, but the rigorous version reads them from the API's citation blocks rather than trusting the
model to retype them — the same substrate-vs-report distinction this whole tool is about, applied to
its own output.
- **Cost scales with claims.** Substance is roughly `1 + 2 × claims` calls; faithfulness sends the
full source twice per claim; grounding runs one search-enabled call per claim. Tune on
`--model claude-haiku-4-5-20251001`, then re-run on a stronger model for the verdict you'll trust.

## Why

The goal isn't only to apply logic and reasoning — it's to make their distinctions legible. The
three modes turn an abstract epistemic point into three columns you can watch disagree: a claim can
be *accurately reported*, *well-reasoned*, and *false* all at once. Seeing those come apart on a
real example teaches the distinction better than any definition of it.

## Worked example: finding the signal in a pundit's claims

**Setup.** Dan Shipper (CEO of Every, AI-forward builder) made 12 predictions about the future of
work on Lenny's Podcast. A summary of those predictions exists in `dan_summary.txt`; the full
transcript is in `dan_shipper.txt`. We run all three filters in order.

```bash
M=claude-haiku-4-5-20251001

# 1. Faithfulness — did Dan actually say this?
python3 assay.py dan_summary.txt --source dan_shipper.txt --model $M

# 2. Substance — is each claim well-formed and falsifiable?
python3 assay.py dan_summary.txt --model $M

# 3. Grounding — is it actually true?
python3 assay.py dan_summary.txt --evidence --model $M
```

**Cross-tabulation.** The signal lives in where the three columns agree or disagree.

| #  | Claim                                      | Faithful?  | Substantive? | Grounded? |
|----|--------------------------------------------|------------|--------------|---|
|  1 | Future of work inside Codex / Claude Code  | overstated | partial      | mixed |
|  2 | Every company: one super-agent in Slack    | overstated | partial      | mixed |
|  3 | SaaS not dead — "I would buy SaaS stocks"  | faithful   | partial      | supported |
|  4 | BYOI tokens improve SaaS margins           | partial    | partial      | mixed |
|  5 | PMs will thrive                            | partial    | partial      | mixed |
|  6 | Full-stack designers = superheroes         | **absent** | partial      | unverifiable |
|  7 | AI job apocalypse not happening            | faithful   | partial      | mixed |
|  8 | Forward deployed engineer most essential   | **absent** | substantive  | mixed |
|  9 | CLIs are over                              | overstated | partial      | **refuted** |
| 10 | Automation is a lie                        | faithful   | partial      | **refuted** |
| 11 | We'll read more AI writing and like it     | faithful   | partial      | mixed |
| 12 | Software for humans + agents together      | partial    | partial      | supported |

**Pattern-reading rules.**

| Pattern                                                | What it means |
|--------------------------------------------------------|---------------|
| faithful + partial/substantive + supported             | **Signal.** A real, checkable claim that holds. |
| faithful + partial/substantive + mixed or unverifiable | **Genuine bet.** Well-formed and honestly attributed — not yet settled. |
| faithful + partial/substantive + **refuted**           | **Anti-signal.** Checkable and wrong. Investigate before acting on it. |
| **overstated** or **absent** + any substance/grounding | **Summarizer's noise.** Go back to the source; the distortion is in the summary, not Dan. |
| faithful + **hollow**                                  | **Speaker's noise.** Dan said something vacuous; the summarizer reported it accurately. |

**The actual results, read by pattern.**

*Strongest signal — claim 12 (human + agent software).* Partially faithful to a richer source claim,
substantively partial, and externally supported. The direction is clear; the summary just dropped
the detail about what changes (approval flows, logs, rollback).

*Genuine falsifiable bets — claims 3, 7, 11.* All faithfully reported, well-formed, and either
supported (SaaS not dead) or mixed (job apocalypse, AI writing). These are predictions worth t
racking.

*Summarizer's noise — claims 6 and 8.* The faithfulness check returned `absent` for both.
Dan never said designers would become "superheroes" (he said they'd "feel empowered to build");
he never said forward deployed engineer is the "most essential" role (he called it "for real"
and an "emerging role" alongside others). Substance and evidence verdicts on those phrases are
moot — they're the summary's words, not Dan's.

*Worst claim — #9 (CLIs are over).* Three-way disagreement in the most damaging direction:
overstated by the summary (Dan said CLIs lost their moment as *the* AI interface, not that
they're gone), substantively partial, and refuted by external evidence (CLIs are in a
renaissance driven by AI agents and cloud infra). This is the one to quarantine.

*Interesting ambiguity — #10 (automation is a lie).* The faithfulness check returned
`faithful` — Dan really said it — but the evidence check returned `refuted`. The disagreement
is real: Dan used "automation is a lie" as a rhetorical provocation meaning "automation always
needs a human in the loop," not a literal denial that automation works. The Critic read the
literal version. Flag this as a case where the Critic was mis-calibrated to the intended
register; go back to the transcript to read Dan's gloss.

**Takeaway.** Signal = claims where faithful + substantive + supported or unverifiable (a
genuine forecast) all align. Noise hides in the disagreements: `overstated`/`absent` is the
summarizer's distortion; `refuted` on a `faithful` claim is a checkable error in the original;
`refuted` on an `overstated` claim means you are critiquing a straw man. The three columns
make those four failure modes visible at a glance.
