# Dan Shipper predictions — worked example

Dan Shipper (CEO of Every) made 12 predictions about the future of work on Lenny's Podcast.
`dan_summary.txt` contains those 12 claims; `dan_shipper.txt` is the full transcript.

This example runs all three assay filters in one command and cross-tabulates the results.

## How to run

From the repo root:

```bash
# Build first if you haven't already
./build.sh

S=examples/dan_shipper

# All three filters in one cross-tab (fast pass on haiku)
./assay -audit -source $S/dan_shipper.txt -md -model claude-haiku-4-5-20251001 $S/dan_summary.txt

# Or run modes individually
./assay -source $S/dan_shipper.txt $S/dan_summary.txt            # faithfulness only
./assay $S/dan_summary.txt                                       # substance only
./assay -source $S/dan_shipper.txt -evidence $S/dan_summary.txt  # grounding on intended propositions
```

Re-run on `claude-sonnet-4-6` (the default) for the verdict you'll trust; calibration is
model-sensitive.

## Results

Cross-tabulation produced by:

```
./assay -audit -source dan_shipper.txt -md -model claude-haiku-4-5-20251001 dan_summary.txt
```

The audit grounding column uses the **intended proposition** — what the speaker actually asserted
(`what_source_actually_says` when faithfulness flagged the claim as `partial` or `overstated`),
not the literal summary wording. A `refuted` in the grounding column therefore means the
*intended* claim is refuted — a real anti-signal — not a straw man.

| # | Claim | Faithful? | Substantive? | Grounded? |
|---|---|---|---|---|
| 1 | The future of work will happen inside Codex or Claude Code. | partial | partial | mixed |
| 2 | Every company will have one "super-agent" inside their Slack that every employee talks to regularly. | overstated | partial | mixed |
| 3 | SaaS is not dead—in fact, Dan is bullish on SaaS stocks. His contrarian take: "I would buy SaaS stocks right now." | faithful | **hollow** | supported |
| 4 | SaaS economics will shift: users will bring their own AI tokens into apps, which actually improves SaaS margins. | partial | partial | mixed |
| 5 | PMs will thrive in the AI era. | faithful | **hollow** | mixed |
| 6 | Full-stack designers will become superheroes. | **absent** | **hollow** | unverifiable |
| 7 | The AI job apocalypse is not happening. | faithful | partial | mixed |
| 8 | Forward deployed engineer is the new most essential role. | **absent** | partial | mixed |
| 9 | CLIs are over. | overstated | **hollow** | **refuted** |
| 10 | Automation is a lie. | partial | substantive | supported |
| 11 | We will read way more AI-generated writing and we will like it. | faithful | partial | mixed |
| 12 | We'll be building software for humans and agents to use together. | faithful | **hollow** | supported |

## Reading the results

| Pattern | Reading |
|---------|---------|
| faithful + partial/substantive + supported | **Signal** — a real, checkable claim that holds. |
| faithful + partial/substantive + mixed or unverifiable | **Genuine bet** — well-formed and honestly attributed, not yet settled. |
| faithful + partial/substantive + **refuted** | **Anti-signal** — checkable and wrong. Investigate before acting on it. |
| **overstated** or **absent** + any | **Summarizer's noise** — the distortion is in the summary. Go back to the source; the faithfulness column tells you where. |
| faithful + **hollow** | **Speaker's noise** — vacuous, but accurately reported. |

Note: because the audit grounds the *intended* proposition rather than the literal words, an
`overstated` or `absent` grounding verdict is rare — the substitution handles the register before
grounding. If grounding returns `refuted` on an `overstated` row, it means even the narrower
intended claim is refuted — a stronger indictment than a straw-man refutation.

**Signal — #10 (automation is a lie).** After the Literalization fix, faithfulness correctly
returns `partial` (Dan said it, but as a rhetorical provocation; the summary treated it as literal).
The grounding pass uses the intended proposition — "automation always needs a human in the loop" —
and returns `supported`. Three-column alignment: the claim is (roughly) accurately attributed,
substantively defensible, and evidentially supported. Strong signal, correct register.

**Strongest signal — #3 (SaaS not dead) and #7 (AI job apocalypse).** Both faithfully reported,
substantively partial or hollow but grounded supported/mixed — genuine falsifiable positions worth
tracking.

**Genuine bets — #11, #1, #4.** Faithfully or partially reported, substantively partial, grounded
mixed. Well-formed and honestly attributed; not yet settled by evidence.

**Summarizer's noise — #2, #6, #8.** Overstated or absent at faithfulness. Go back to the
transcript — you're not evaluating what Dan said.

**Worst claim — #9 (CLIs are over).** Overstated by the summary, substance hollow, grounding
refuted. Even the intended narrowed claim ("CLIs are no longer the primary work surface") is
refuted by evidence (CLIs are in a renaissance). Three-way bad: the summarizer inflated it, the
claim dissolved on scrutiny, and the evidence cuts against it.

**Speaker's noise — #3, #5, #12.** Faithfully reported but substance returned hollow — the
original claims are unfalsifiable or equivocate. Dan said them; they just don't hold up under
dialectic.
