# FAQ

First-time questions about **crossexam**, answered. Where an answer reasons about a design choice the
repo has not settled, it is flagged in italics as design analysis rather than stated as fact. Several
answers link to the deeper documents — [THEORY.md](THEORY.md), [PLAN.md](PLAN.md),
[LIMITS.md](LIMITS.md), and the running [critique](CRITIQUE.md) of the repo's own claims.

> **Note:** the binary is not built yet (see [Status](README.md#status)). The commands and outputs
> described below are the contract the binary must satisfy, not a description of working software. No
> verdict shown anywhere in this FAQ is a measured result.

---

## Goals

### What are the goals of this repo?

The repo builds `crossexam`, a tool that splits prose into atomic claims and asks three deliberately
separate questions of each —
faithfulness (did the source say it),
substance (is it well-formed and falsifiable),
grounding (is it true)
 — and refuses to merge the verdicts.
The target failure it guards against is *confidence laundering*: real wins on the questions
reasoning can reach
(attribution, structure) spent as authority on the one it cannot (truth).

### Who is this for? Is it useless to me if I'm not already "trained to separate attribution from structure from truth"?

The named audience is people already trained to separate attribution from structure from truth, who
want to see that skill operating on material from their own work. The claimed mechanism is
recognition, not instruction: the tool does not install the skill, it makes a held skill visible (the
[mirror hypothesis](THEORY.md#the-mirror-hypothesis)). So for an untrained user the *central* payoff
is, by the repo's own account, weaker — but the per-column verdicts (a refuted claim, an absent
attribution) are still legible without the meta-skill. That said, the value to untrained users is not
something the repo measures, and the audience claim itself is carried as an unproven hypothesis.

### What is the "mirror hypothesis," and has it been shown to be true?

The mirror hypothesis is the repo's claim about its own effect: that seeing the three columns
disagree causes an already-trained user to recognise their skill as transferable — recognition, not
instruction. It has **not** been shown true. The test condition is field evidence from trained users
(not the people who built the tool), which the repo does not have. The reflexive canary
(`unverifiable` 5/5, haiku, 2026-06-11) confirms only that the grounding module does not self-confirm
the claim from the armchair — the correct null result at this stage, not a positive finding. The
critique additionally flags this load-bearing thesis as
[itself ungrounded](CRITIQUE.md#the-reflexive-thesis-is-itself-ungrounded).

---

## What it is

### In one sentence, what is crossexam and what do I get from it?

crossexam is a CLI that decomposes prose into atomic claims and returns, per claim, up to three
independent verdicts — faithfulness, substance, grounding — kept in separate columns so a win on one
is never spent as authority on another; what you get is a per-claim table of verdicts plus a JSONL
verification chain, not a single "good/bad" score.

### Is this a fact-checker?

Only one of the three modes (grounding, `-evidence`) attempts truth, and the repo is emphatic that
even that is bounded: positive grounding is a retrieval, not a deduction, and the `supported` verdict
currently only proves a URL was retrieved, not that the page backs the sentence
([critique](CRITIQUE.md#the-given-gap)). The other two modes are explicitly *not* about truth —
faithfulness checks sense-preservation, substance checks well-formedness. So it is a fact-checker only
in the narrow, heavily-qualified grounding column, and a weak one to lean on there.

### What do I actually see when I run it?

Results go to stdout — an ANSI table by default, markdown with `-md`, and audit is always markdown —
while progress, a `SUMMARY` block, and a `USAGE` line go to stderr. Each mode prints a per-claim table
(substance, for instance, is `| # | Claim | Verdict | Why |`), and a per-claim verification chain is
written as JSONL under `eval/<stamp>-<model>/`. The binary is not built, so this describes the output
contract, not a real run.

### One-sentence difference between the three modes — and when do I pick which?

Faithfulness (`-source`) asks *did the source actually say it* — pick it when you have a summary and
its source transcript; substance (default) asks *is the claim well-formed and falsifiable* — pick it
when you have claims and no external referent to check against; grounding (`-evidence`) asks *is it
true per external evidence* — pick it when the claim is checkable against the world and you accept
retrieval-bounded confidence; `-audit` runs all three cross-tabulated and requires `-source`.

### Why three separate modes and verdicts instead of one "is this claim good?" score?

Because the three questions reach different things, and merging them is the exact failure the tool
targets: a single score lets a win on faithfulness or substance (reachable by reasoning) be spent as
authority on truth (not reachable by reasoning alone). Keeping the columns separate is the structural
refusal of confidence laundering. The critique argues the substance/grounding line itself may not be
cleanly drawable in principle ([critique](CRITIQUE.md#the-analyticsubstance-vs-syntheticgrounding-gap)).

---

## Problem statements

### All useful work needs a problem statement. Which of the 3 questions (faithfulness, substance, grounding) requires the problem to be stated?

This is the subtle one, and the glib answer is wrong. On the surface, faithfulness obviously needs an
external referent (the source transcript) and grounding needs one (the world / evidence), while
**substance looks like the self-contained mode** — it grades a claim's internal form with no external
problem stated. The context-principle critique inverts this: substance is precisely the mode whose
referents must be supplied. Decomposing into atomic fragments strips the context that fixes meaning —
in the repo's own dogfooding the critic read "assay" as a chemical assay and rated "two of assay's
three columns" hollow because "columns" had no referent out of context. So substance is the mode that
most needs the problem/context stated, yet its contract pretends it does not, and it fails *silently*
when context is missing. Faithfulness states its problem explicitly (the source); grounding states it
(truth against evidence); substance hides the requirement ([THEORY.md](THEORY.md#the-axis-boundary),
[critique](CRITIQUE.md#the-atomism-gap)).

---

## Method choice

### We chose Dialectic. What are other ways of achieving our goals? Compare them to dialectic.

*What the repo actually does:* substance runs a producer→critic dialectic over seven fixed axes;
faithfulness runs a defender→critic pair; grounding runs a single web-search call.

*Design analysis, not a settled repo decision:* alternatives to a dialectic for the same goals —

- **Checklist / rubric scoring** (one LLM scores against fixed criteria): cheaper and more
  reproducible, but has no adversary — a plausible-but-hollow claim is likelier to pass, and it loses
  the steelman-then-attack separation [THEORY.md](THEORY.md#the-producer-and-the-limit-of-separation)
  uses to remove self-defense bias.
- **Verifier graded against an external score/database:** gives a real truth signal *where a database
  exists*, but only there; this is exactly the design THEORY.md contrasts itself with — it buys truth,
  not honesty about what was said.
- **Formal / deductive checking** (proof, type-checking, constraint solving): conclusive where claims
  are formalizable, useless on natural-language prose.
- **Ensemble / self-consistency** (sample N, take the mode): the calibration protocol already does
  this at the verdict level; it reduces variance but adds no adversarial axis.

Dialectic's distinctive value here is the asymmetry THEORY.md leans on (it can refute by deduction but
not confirm) plus critic-from-producer independence; its distinctive cost is the rule-following gap —
the axes underdetermine their own application ([critique](CRITIQUE.md#the-rule-following-gap)).

### What are alternatives to the Producer-Critic loop?

*Design analysis, not a settled repo decision:* (a) single-pass critic with no producer — no
steelman, so it risks attacking a weak version (the strawman the producer exists to prevent); (b)
multi-critic panel or two-advocate debate judged by a third — more coverage, more cost, more variance;
(c) self-critique / reflexion, one model grading its own output — THEORY.md explicitly rejects this
for self-defense bias; (d) iterative refinement with no adversary (rewrite-until-clean) — improves
prose, does not test survival; (e) human-in-the-loop critic. The repo's current loop is specifically
producer-blind-to-the-axes plus an independent-context critic, chosen so the generator cannot shape
output to survive a critic it can anticipate
([THEORY.md](THEORY.md#the-producer-and-the-limit-of-separation)).

### How is crossexam different from a fact-checker, an LLM-as-judge eval harness, or a RAG hallucination detector?

*Design analysis, not a settled repo decision:* a fact-checker collapses everything into true/false;
crossexam refuses to, treating truth as one of three columns and the least reachable. An LLM-as-judge eval
harness grades an output against a task/rubric to emit a score; crossexam's substance mode is adversarial
(producer vs critic) and reports a verdict *distribution* rather than a scalar. It treats the
false-pass *rate* as a property of that distribution, not of the claim — while still reading repeated
runs as "what the claim tends to do under examination," and the critique flags exactly that slide from
distribution to claim-property as the [psychologism gap](CRITIQUE.md#the-psychologism-gap). A RAG
hallucination detector checks whether generated text is supported by retrieved context; that maps
closest to crossexam's *faithfulness* mode (sense-preservation against a source) — but crossexam separates
that from truth entirely, whereas a hallucination detector typically conflates "grounded in the docs"
with "correct."

---

## Trust & cost

### If the judge is itself an LLM that can be wrong, why should I trust any verdict?

You should not trust any *single* verdict — the repo says so: a verdict from one run is "one draw"
from a distribution, and the calibration protocol reports full distributions per probe rather than a
point verdict ([THEORY.md](THEORY.md#verdicts-as-distributions)). The narrower claim the repo does
make is that some columns are more trustworthy than others: faithfulness-with-source-present and
refutation-of-internally-contradictory-claims are highest (truth-maker in the window / reachable by
deduction), while `supported` grounding and specialist-domain substance are weakest
([LIMITS.md](LIMITS.md)). Trust is bounded and column-specific, never blanket. (Note too that
faithfulness preserves *what was said*, which can still drop *what was implicated* —
[critique](CRITIQUE.md#the-force--speech-act-gap).)

### The docs say a single run is "one draw." Do I have to run everything ~10 times? What does that cost in time and API spend?

The N=10 (canary N=5) protocol is the *calibration / acceptance* procedure — how the repo measures the
tool's own bias envelope — not a mandate for every production use. The README's Ballmer row and PLAN.md
§2 distributions are N=10 calibration results. For a one-off read you can run once and read the verdict
as one draw; you repeat only when you need the distribution (a high-stakes verdict, or to detect skew).
There are no time or dollar figures to quote: the binary is not built and no benchmark exists. crossexam
does emit a per-run `est_usd` and a `USAGE` line, but those are contract outputs, not measured numbers
— and `est_usd` is null when the model is not in the price table.

### Which verdicts are safe to rely on, and which aren't?

From the operating envelope in [LIMITS.md](LIMITS.md) — *most reliable:* faithfulness with the source
present (truth-maker in the window) and refutation of internally-contradictory claims (deduction, not
retrieval); *reliable by design:* `unverifiable` on unresolved predictions. *Least reliable:*
`supported` grounding (proves a URL was retrieved, not that the page backs the sentence — structurally
weakest verdict; [critique](CRITIQUE.md#the-given-gap)); substance on sweeping forward predictions
(skews `hollow` via Counterexample); and substance on novel/specialist domains (producer and critic
share a model, so a blind spot survives both; [critique](CRITIQUE.md#the-externalism-gap)). Treat any
`substantive` verdict with an empty critique as suspect — that is the documented false-pass signature.

---

## Limits & scope

### What is crossexam bad at? When should I not use it?

Bad at: positive truth confirmation (grounding-as-retrieval cannot deliver a truth-maker, possibly
even in principle; [critique](CRITIQUE.md#the-given-gap)); relational defects like motte-and-bailey
(`decompose` splits the move into separate fragments and grades the pieces, so the move *between* them
goes ungraded); sweeping forward predictions (skew `hollow`); specialist/novel domains (shared-model
blind spot; [critique](CRITIQUE.md#the-externalism-gap)); and rhetoric/irony (no machinery for
illocutionary force — it can literalize a performative; [critique](CRITIQUE.md#the-force--speech-act-gap)).
Do not use it as an oracle, as a single-run truth verdict, or on decontextualised specialist claims
where the referent is severed ([LIMITS.md](LIMITS.md)).

### CRITIQUE.md says the "grounded?" column may be unreachable even in principle. So why ship a grounding mode at all?

*Design analysis, not a settled repo decision:* three reasons the repo's own logic supports. (1) The
*negative* direction is reachable — reasoning can refute (internal contradiction; settled-false
predictions like the Ballmer forecast, `refuted` 10/10), even if it can never confirm. (2) Shipping
the column with its weakness printed (`supported` = URL retrieved only) *is* the anti-laundering
stance — hiding the column would hide the limit, the move the tool exists to refuse. (3) The critique's
own ["Options for the future"](CRITIQUE.md#options-for-the-future) lists marking grounding permanently
provisional, or dropping the completeness claim, as live strategic options — i.e. the repo has not
resolved this and presents it as a fork, not a finished feature. So grounding ships partly as a
refutation engine and partly as an unresolved design question, never as a working truth oracle.

### Does it work on specialist or technical claims?

This is where the repo says trust it *least*. Producer and critic share one model, so a blind spot in
the producer survives the critic; [LIMITS.md](LIMITS.md) marks specialist/novel domains "Lowest —
treat with most suspicion." The [externalism critique](CRITIQUE.md#the-externalism-gap) sharpens it:
the meaning of specialist terms is fixed by a division of linguistic labor the critic cannot defer to,
and decontextualised fragments sever reference (the "assay"-as-chemical-assay collision); a one-line
domain anchor patches symptoms but does not restore the causal-historical reference chain. So it runs,
but specialist claims are its worst-grounded case.

---

## Status

### Is it finished — can I use it today?

No. The [README](README.md#status) says the binary is not built yet; the internal packages are
scaffold stubs (package-comment files only), and the commands in the README are "the contract the
binary must satisfy, not a description of working software." You cannot run it today.

---

## LLMs

### Can we make this repo LLM-agnostic, to try different LLMs and compare performance?

*Design analysis, not a settled repo decision:* today it is not. The model is an Anthropic model id
(default `claude-sonnet-4-6`, override via `-model` / `ANTHROPIC_MODEL`), and grounding uses
Anthropic-specific web-search tooling ([spec/CLI.md](spec/CLI.md),
[spec/PROMPTS.md §6](spec/PROMPTS.md)). Making it agnostic would mean abstracting the client behind an
interface (provider → text + tools + usage). That is feasible: the spec already calls for a
separate `api/` package boundary holding the HTTP client and retry logic
([spec/LESSONS.md §2](spec/LESSONS.md)), and the modes already consume typed JSON shims
(`substanceJSON` / `faithJSON` / `evidenceJSON`) rather than raw API types
([spec/PROMPTS.md](spec/PROMPTS.md), [spec/BEHAVIOR.md](spec/BEHAVIOR.md)), so a provider swap behind
that seam need not touch the orchestration. The hard part is provider-specific: web-search grounding
and citation handling would need per-provider adapters or a shared search seam. The calibration
baselines are model-specific (haiku, dated), so cross-model comparison means re-running the protocol
per model.

### How would we measure that performance?

The repo already defines the instrument: the calibration protocol — N runs per probe, fragment-level
distributions reported in full, false-pass counts read per `fragment_attribution` records, `error`
verdicts excluded from `n_valid` — run against the destructive probe fixtures with known engineered
defects ([PLAN.md §2](PLAN.md)). "Performance" here is not accuracy-versus-gold-label but the *bias
envelope*: does each probe land in its expected distribution (equivocation fires 10/10; reference-class
and unfalsifiable-dress stay at 0 substantive; the canary holds), and what is the false-pass rate. *As
design analysis,* to compare LLMs you would hold prompts and fixtures fixed, swap the model, and compare
distributions and false-pass counts. There is no single scalar score, by design.

### Can we remove LLMs from this repo's structure completely? If so, how does the residual differ from the LLM-enhanced whole?

*Design analysis, not a settled repo decision:* the scaffolding that survives without LLMs —
decomposition-into-fragments as a pipeline, the seven-axis taxonomy as a checklist, the
producer/critic/defender control flow, the verdict enums, the JSONL chain, the cross-tab, the
calibration harness, and cost/usage accounting. What you lose is every step that requires reading
natural language — steelmanning a claim, finding distortions against a source, searching for evidence,
assigning a verdict. The residual is a bookkeeping-and-protocol skeleton: it can enforce that the three
questions are asked and kept separate, record verdicts, and run the calibration math, but it cannot
*produce* a verdict. So the LLM is the judgment; the structure is the discipline that constrains and
audits the judgment — and the critique's argument that the structure's value (separation,
anti-laundering) is partly independent of the judge still needs a judge for there to be anything to
separate.

### Can I run it against my own local model?

*Design analysis, not a settled repo decision:* not as written. The client targets the Anthropic API
(`ANTHROPIC_API_KEY` required, fatal if unset; Anthropic model ids; Anthropic web-search tool;
[spec/CLI.md](spec/CLI.md), [README §Usage](README.md#usage)). A local model behind an
Anthropic-compatible endpoint *might* work if it honoured the same API and tool surface, but grounding
mode's web search and JSON-tool behaviour are the hard part and are not guaranteed by a local model.
There is no documented local-model path.

---

## Uses

Each example below shows the decompose → three-questions flow and names where a documented limit bites.
**These illustrate the contract, not a real run** — the binary is not built, and no verdict shown is a
measured result; each is only how the flow would route and which axis or limit applies.

### Code review — worked example

Input: a PR description, "This refactor makes the parser 10× faster and eliminates all allocations in
the hot path." Decompose → `["the refactor makes the parser 10× faster", "the refactor eliminates all
allocations in the hot path"]`. *Faithfulness* (with the diff/commit as `-source`): did the change
actually do this, or does the summary overstate the diff? *Substance:* "10× faster" trips Base
rate/magnitude (a real quantity needs a comparison — versus what workload?); "eliminates all
allocations" trips Falsifiability/Counterexample (one allocation refutes "all"). *Grounding:* only
checkable if an external benchmark is retrievable. **Limit that bites:** specialist-domain suspicion —
producer and critic share a model that may not know the codebase, so a wrong `substantive` on a
domain-specific perf claim survives both roles; also, faithfulness mode is built for transcript
summaries, not diffs ([critique](CRITIQUE.md#the-externalism-gap)).

### History books — worked example

Input: "The Treaty of Versailles caused World War II by economically crippling Germany." Decompose →
`["the Treaty of Versailles economically crippled Germany", "that economic crippling caused World War
II"]`. *Substance:* the second fragment trips Causality-vs-correlation directly (cause asserted from
association) and Hidden premise (it assumes a counterfactual). *Faithfulness:* applies only if you have
the source the summary came from. *Grounding:* the first fragment is partly checkable against economic
data; the causal claim is the contested, hard-to-confirm kind. **Limit that bites:** the
[atomism gap](CRITIQUE.md#the-atomism-gap) — splitting "caused" from "crippled" grades each piece but
loses the causal *move* that is the actual historical claim, the same way decomposition dissolves a
motte-and-bailey.

### Literature review — worked example

Input: a related-work paragraph, "Prior work (Smith 2021) showed method X outperforms Y; we are the
first to do Z." Decompose → `["Smith 2021 showed X outperforms Y", "no prior work has done Z"]`.
*Faithfulness* (with Smith 2021 as `-source`): did Smith actually show that, or is it overstated /
context-stripped? — this is faithfulness's strongest case, the truth-maker is in the window.
*Substance:* "first to do Z" trips Falsifiability and Base rate (what would refute it — one prior
instance). *Grounding:* "first to do Z" is a universal negative — refutable by one counterexample,
never confirmable by retrieval. **Limit that bites:** the [Given gap](CRITIQUE.md#the-given-gap) —
grounding can refute "first" by finding a prior work but can never confirm it, and `supported` would
only prove a URL was retrieved.

### Vibe-coding — worked example

Input: a coding assistant's claim, "I've added input validation and the function now handles all edge
cases." Decompose → `["input validation was added", "the function handles all edge cases"]`.
*Faithfulness* (with the actual diff as source): did it add validation, or only claim to? *Substance:*
"all edge cases" trips Counterexample/Falsifiability (one unhandled case refutes it; "all" is the
classic hollow-skewing universal). *Grounding:* only checkable by running tests, which crossexam does not
do — it reads the prose claim, not the behaviour. **Limit that bites:** crossexam examines the *claim about*
the code, never the code — a confident-but-false "handles all edge cases" with a plausible rationale
can pass while the code is broken — and universal claims skew toward `hollow`
([LIMITS.md](LIMITS.md), [critique](CRITIQUE.md#the-externalism-gap)).

### Social-media influencer post decoding — worked example

Input: "This one morning habit boosted my productivity 300% — the science is clear." Decompose →
`["a morning habit boosted the speaker's productivity 300%", "the science is clear on this"]`.
*Substance:* "300%" trips Base rate/magnitude (a quantity needs a baseline), "the science is clear"
trips Evidence and Equivocation (a buzzword shield). *Faithfulness:* if decoding what a cited study
"shows," run it with the study as source — likely overstated/cherry-pick. *Grounding:* the
productivity number is anecdotal/unverifiable; "the science" is retrievable and often refutable.
**Limit that bites:** the [force/speech-act gap](CRITIQUE.md#the-force--speech-act-gap) — influencer
copy is performative and hyperbolic, and grading it on literal truth-conditions misreads register;
faithfulness mode has a Literalization guard, but substance and grounding can still treat hype as a
sincere constative.

### Science (exploring spectra libraries to deduce inorganic chemistry insights) — worked example

Input: a prose conclusion, "The 580 cm⁻¹ band confirms octahedral Fe–O coordination in the sample."
Decompose → `["a band appears at 580 cm⁻¹", "that band indicates octahedral Fe–O coordination", "the
sample contains octahedral Fe–O"]`. Be honest about what is happening: **crossexam validates the prose
claims, not the spectra.** It cannot read a spectrum or a spectral library; it can only assess the
sentences a chemist writes about them. *Faithfulness* would check the prose against a cited source
text; *substance* would test the inference (Hidden premise: the assignment assumes a reference
assignment; Evidence; Causality); *grounding* would web-search the claim. **The limits bite hardest
here:** grounding is retrieval-bound, so at best it finds a paper *asserting* the assignment — another
claim, not the truth-maker (the [Given gap](CRITIQUE.md#the-given-gap)) — while specialist-domain
suspicion is maximal: band assignment is expert content fixed by a division of linguistic labor that
the shared producer/critic model cannot reliably defer to (the
[externalism gap](CRITIQUE.md#the-externalism-gap); the README warns that "specialist domains deserve
the most suspicion"). Do not use crossexam to *derive* chemistry insight; at most use it to flag
overclaiming in chemistry prose, and trust even that least.
