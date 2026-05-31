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
git clone ...
./build.sh          # runs go test ./... then builds ./assay
export ANTHROPIC_API_KEY=sk-ant-...
```

Optionally pin a default model:

```bash
export ANTHROPIC_MODEL=claude-opus-4-8
```

Input is a file argument, `-text "..."`, or piped on stdin.

## Usage

### Substance (default) — is the claim well-formed?

```bash
./assay memo.txt
```

Decomposes the prose into atomic claims, then runs a **producer–critic** dialectic on each. A
Producer states the strongest version of the claim, *blind to the critique axes*; a Critic attacks
it on seven fixed axes; the loop repeats on the surviving claim until the Critic raises no new fatal
or weakening finding, or `-max-rounds` is hit. Verdict: `substantive` · `partial` · `hollow`.

Fixed critique axes:
Evidence · Hidden premise · Falsifiability · Equivocation · Base rate / magnitude · Counterexample · Causality vs correlation.

### Faithfulness — did the source actually say it?

```bash
./assay -source transcript.txt summary.txt
```

For each claim in the summary, a **Defender** finds verbatim support in the source and a **Critic**
judges whether the summary represents it accurately, watching for overstatement, fabrication,
distortion, context-stripping, misattribution, cherry-picking, and literalization (rhetorical or
ironic claims reported as sincere literal assertions). Verdict:
`faithful` · `partial` · `overstated` · `absent` · `contradicted`.

This checks attribution — *whether the speaker said it* — never whether it's true. A
faithfully-reported claim can still be wrong.

### Grounding — is it actually true?

```bash
./assay -evidence claims.txt
```

For each claim, Claude searches the web for real evidence — the actual truth-makers, not anyone's
assertion that the claim is true — weighs what it finds, and returns
`supported` · `mixed` · `refuted` · `unverifiable`, with the sources used.
This is the only mode that touches truth.

When a source is supplied alongside `-evidence`, the tool first runs the faithfulness pass to
reconstruct the *intended* proposition (what the speaker actually asserted, not the literal words),
then grounds that — so rhetorical or overstated claims are not evaluated against a straw-man:

```bash
./assay -source transcript.txt -evidence summary.txt
```

### All three at once — audit mode

```bash
./assay -audit -source transcript.txt -md summary.txt
```

Runs faithfulness, substance, and grounding on the same claim set and emits a cross-tabulation
markdown table. The signal lives in where the three columns agree or disagree.

### Chaining the three

To critique a pundit properly, run the modes in order: **faithfulness** (are we critiquing what they
actually said?) → **substance** (is it well-formed and falsifiable?) → **grounding** (is it borne
out by evidence?). Each answers a question the previous one can't.

## Flags

| flag              | effect |
|-------------------|--------|
| `path` / `-text`  | the input prose |
| `-source FILE`    | faithfulness mode (or proposition reconstruction for `-evidence`) |
| `-evidence`       | evidence-grounding mode (web search) |
| `-audit`          | all three modes, cross-tabulated (requires `-source`) |
| `-md`             | emit markdown tables instead of terminal colour |
| `-model ID`       | model id (default `claude-sonnet-4-6`) |
| `-max-rounds N`   | producer–critic rounds per claim (substance only, default 2) |
| `-v`              | show every Claude call: prompt and response |
| `-no-color`       | disable ANSI colour |

## Output

By default you see the dialectical *result* of each step — the steelman, the critique by axis, the
verdict — plus a summary at the end. This is "what a careful reader would conclude." `-v` instead
shows every Claude call as it happens: system prompt, user prompt, and the full response —
"how the machine got there." Pipe verbose to a file (`> log.txt`) when you want the full trace.
Use `-md` to emit clean markdown tables suitable for piping into a document or review.

## Honest limits

- **The Critic is itself an LLM.** Its precision varies; it can wave through a hollow claim or
over-attack a sound one. The skill the tool is really training is catching where the machine critic
is wrong — don't treat its verdict as final.
- **Predictions come back `unverifiable` under `-evidence`, by design.** You can't ground a claim
about the future in present evidence. That's the correct result, not a failure — it's the line
between "falsifiable in principle" (which substance mode checks) and "settled by current evidence."
- **Source URLs in `-evidence` are written by the model** from its search results. They should be
real, but the rigorous version reads them from the API's citation blocks rather than trusting the
model to retype them — the same substrate-vs-report distinction this whole tool is about, applied to
its own output.
- **Cost scales with claims.** Substance is roughly `1 + 2 × claims` calls; faithfulness sends the
full source twice per claim; grounding runs one search-enabled call per claim. Tune on
`-model claude-haiku-4-5-20251001`, then re-run on a stronger model for the verdict you'll trust.

## Why

The goal isn't only to apply logic and reasoning — it's to make their distinctions legible. The
three modes turn an abstract epistemic point into three columns you can watch disagree: a claim can
be *accurately reported*, *well-reasoned*, and *false* all at once. Seeing those come apart on a
real example teaches the distinction better than any definition of it.

## Worked example

See [`examples/dan_shipper/`](examples/dan_shipper/) — Dan Shipper's 12 predictions about the
future of work, run through all three filters with a cross-tabulated results table and
pattern-reading guide.
