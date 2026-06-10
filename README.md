# [elenchus](https://plato.stanford.edu/entries/plato-ethics-shorter/)

Named for the Socratic *elenchus*: refuting a claim not by asserting its opposite, but by drawing
out what the claimant is committed to and showing where those commitments collide. The link is where
the chain of rigour starts; this repo is an attempt to continue it on modern work artifacts.

A dialectical filter for claims. It takes prose — a strategy memo, a pundit's "predictions," a
summary of a talk — breaks it into atomic claims, and tests each one. It doesn't tell you what to
think; it makes the structure of a claim visible so you can decide.

## Goals

The tool exists to do three things, in priority order:

1. **Keep three different questions apart.** Of any claim you can ask: did the source actually say
   this? (*sense / attribution*) Is it well-formed and falsifiable — the kind of thing that *could*
   be true? (*structure*) Is it actually true, against external evidence? (*reference /
   truth-makers*) A claim can be faithfully reported, internally well-reasoned, and still false.
   Most careless reasoning collapses these three into a single undifferentiated "is this good?";
   most careful reasoning is keeping them apart. The tool answers each question in its own mode and
   lets the answers disagree.

2. **Make already-held rigour visible to its owner.** assay is a mirror, not a crutch. It is built
   for people trained in classical rigour — the trivium, the scientific method, an arts or science
   degree — who have landed in technical workplaces and cannot see where that training applies. The
   training is invisible to the person who holds it. assay instantiates the rigour they already
   have as three concrete columns — *faithfulness*, *substance*, *grounding* — on a real work
   artifact, so the transferable skill becomes visible to its owner. The recognition is the
   product; the verdict on any one claim is secondary.

3. **Never claim more than each mode's competence licenses.** Close reading can settle attribution.
   Dialectic can settle structure. Neither can *confirm* truth — positive grounding always needs a
   truth-maker: a retrieval, not a deduction, however rigorous. The characteristic failure of
   fluent reasoning is laundering confidence across that line — winning on the reachable questions
   and pronouncing on the unreachable one with borrowed authority. The tool is built to refuse that
   move, including about itself.

## How the design serves the goals

Each goal is load-bearing somewhere specific in the design.

**Three modes, three verdict vocabularies, no aggregate score** *(goal 1)*. Substance returns
`substantive`/`partial`/`hollow`; faithfulness returns
`faithful`/`partial`/`overstated`/`absent`/`contradicted`; grounding returns
`supported`/`mixed`/`refuted`/`unverifiable`. There is deliberately no combined rating: the modes
cannot be averaged, only cross-tabulated (`-audit`), so the signal lives in where the columns
disagree — the disagreement *is* the lesson.

**A Producer blind to the critique axes** *(goal 1)*. In substance mode, the Producer steelmans the
claim without knowing the seven axes the Critic will attack on. It cannot teach to the test, so a
`substantive` verdict means the claim survived an attack it was not shaped to anticipate. The
producer–critic loop is the structure of a medieval
[obligational disputation](https://plato.stanford.edu/entries/obligationes/): a respondent
committed to a thesis, an opponent probing for the contradiction.

**Seven fixed, named critique axes** *(goal 2)*. Evidence, hidden premise, falsifiability,
equivocation, base rate / magnitude, counterexample, causality vs correlation. These are the moves
a classical training drilled — naming them on a claim from your actual job is how the invisible
skill becomes recognisable as the one you already have.

**A Defender who must quote** *(goal 1)*. In faithfulness mode, the Defender's support must be
verbatim from the source, and the Critic judges attribution only — including catching
*literalization*, where a rhetorical or ironic remark is reported as a sincere assertion. Whether
the claim is *true* never enters this mode; that is the separation, enforced.

**Grounding the intended proposition, not the literal words** *(goals 1 and 3)*. With
`-source` + `-evidence`, the faithfulness pass first reconstructs what the speaker actually
asserted, and *that* is what gets grounded — so grounding never scores a cheap win against a
straw-man of the phrasing.

**A provenance cross-check on grounding** *(goal 3)*. `crossCheckEvidence` compares the URLs the
model *cites* against the URLs the search API *actually retrieved*, and downgrades
`supported`/`mixed`/`refuted` to `unverifiable` on any mismatch. This polices the boundary
mechanically rather than by exhortation — though it proves only that a cited page was fetched,
never that the page backs the sentence, which is why `supported` remains the verdict to read most
skeptically (see [Honest limits](#honest-limits)).

**The tool's own headline is carried as a hypothesis** *(goal 3)*. The claim that seeing the three
columns disagree causes a trained user to recognise their skill as transferable is a grounding
claim about effects on people. By the tool's own axis boundary it cannot be settled from the
armchair — only by watching trained users use the tool. It is a hypothesis awaiting evidence, not a
finding, and the falsifiability press it invites is the shape of the honest limit, not an objection
to answer.

**The failure envelope is published, not hidden** *(goal 3)*.
[`examples/destructive/`](examples/destructive/) ships adversarial probes engineered to break each
mode, with results committed; [`TESTING.md`](TESTING.md) states what those results are and are not
allowed to mean. A claim-testing tool that hid where its own verdicts fail would be its own
counterexample.

## Who this is for

People trained in classical rigour — arts and science graduates — who have landed in modern
technical workplaces and can't see where their training applies. A philosophy graduate in a standup
does not think "this is dialectic"; they think they have a humanities degree while everyone else
has the useful skills. But separating what was said from what is well-formed from what is true —
the move the trivium drilled — is exactly the move a room full of fluent engineers will skip,
because fluency feels like knowledge and agreement feels like corroboration.

You watch the three columns disagree on something from your actual job, and you recognise: I
already know how to do this, and it is worth doing here.

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
| `-model ID`       | model id (default `claude-sonnet-4-6`, or `ANTHROPIC_MODEL`) |
| `-max-rounds N`   | producer–critic rounds per claim (substance only, default 2) |
| `-max-claims N`   | cap the number of claims graded (0 = unlimited; bounds an expensive run) |
| `-progress`       | per-case progress lines + 60s heartbeat on stderr (default on) |
| `-quiet`          | suppress per-case lines and heartbeat; keep the SUMMARY block and Tier-2 chain |
| `-usage-out FILE` | append one JSON usage record (tokens, web searches) per run to FILE |
| `-chain-dir DIR`  | write the Tier-2 JSONL verification chain here (default `eval/<stamp>/`) |
| `-v`              | show every Claude call: prompt and response |
| `-no-color`       | disable ANSI colour |

## Output

By default you see the dialectical *result* of each step — the steelman, the critique by axis, the
verdict — plus a final **SUMMARY** block on stderr (mode, model, verified/total, per-verdict counts,
and token/web-search usage). This is "what a careful reader would conclude." Progress lines and a 60s
heartbeat print to stderr so long grounding runs don't look hung; `-quiet` silences them while
keeping the SUMMARY. A per-case **Tier-2 JSONL** verification chain is written under `-chain-dir`
(default `eval/<stamp>/`) for offline audit, and `-usage-out` appends a one-line usage record per
run. `-v` instead shows every Claude call as it happens: system prompt, user prompt, and the full
response — "how the machine got there." Pipe verbose to a file (`> log.txt`) when you want the full
trace. Use `-md` to emit clean markdown tables suitable for piping into a document or review.

## Honest limits

- **The Critic is itself an LLM.** Its precision varies; it can wave through a hollow claim or
over-attack a sound one. The skill the tool is really training is catching where the machine critic
is wrong — don't treat its verdict as final.
- **Predictions come back `unverifiable` under `-evidence`, by design.** You can't ground a claim
about the future in present evidence. That's the correct result, not a failure — it's the line
between "falsifiable in principle" (which substance mode checks) and "settled by current evidence."
- **Grounding verifies URL *provenance*, not *content support*.** The model retypes its source URLs
from its search results, but `crossCheckEvidence` now cross-checks them against the URLs the API
actually fetched and downgrades `supported`/`mixed`/`refuted` → `unverifiable` when a cited URL was
never retrieved. That proves the page was *fetched*; it does **not** prove the page *backs the
sentence*. So `supported` stays the structurally weakest verdict — read it most skeptically. (The
still-narrower gap: displayed URLs come from the model's JSON, not yet from the API's citation
blocks; see BACKGROUND.md W3.)
- **Cost scales with claims.** Substance is roughly `1 + 2 × claims` calls; faithfulness sends the
full source twice per claim; grounding runs one search-enabled call per claim. Tune on
`-model claude-haiku-4-5-20251001`, then re-run on a stronger model for the verdict you'll trust.

## Worked examples

- [`examples/dan_shipper/`](examples/dan_shipper/) — Dan Shipper's 12 predictions about the future of
  work, run through all three filters with a cross-tabulated results table and pattern-reading guide.
- [`examples/url-length/`](examples/url-length/) — a worked demonstration of the axis boundary: a
  claim that wins on faithfulness and substance yet cannot be confirmed on grounding without going to
  the truth-maker.
- [`examples/destructive/`](examples/destructive/) — seven adversarial axis probes (motte-and-bailey,
  reference-class gaming, hidden premise, unfalsifiable-in-empirical-dress, causality-from-correlation,
  axis-incompleteness, and the laundering trap). Each doubles as a reader-facing demonstration and a
  destructive test that maps where a mode breaks; see [`TESTING.md`](TESTING.md) for the calibration
  results and what they are allowed to mean.
