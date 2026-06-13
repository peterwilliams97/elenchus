# FAQ.answers.md — DRAFT answers for review (NOT the final FAQ)

This is the **answer pass** (step 2 of the FAQ workflow). It is a reviewable draft with
provenance, not standard-FAQ prose. Every answer ends with one or more provenance tags:

- `[GROUNDED: <file §section / phrase>]` — supported by the repo as written.
- `[CONTRACT: binary not built]` — describes promised behaviour the binary must satisfy, not
  observed runtime behaviour. The binary is not built (README §Status); no answer here reports a
  real run.
- `[LIMIT: <anchor>]` — a documented limit (CRITIQUE.md / LIMITS.md / README).
- `[OPEN: design analysis]` — not settled in the repo; my reasoning, separated from what the repo
  does.

No invented numbers, benchmarks, or runtime output. The only quantitative results cited are ones
already in the repo (the Ballmer N=10 row in README; the v1 distributions in PLAN.md §2).

---

## Goals

**What are the goals of this repo?**
The repo builds `assay`, a tool that splits prose into atomic claims and asks three deliberately
separate questions of each — faithfulness (did the source say it), substance (is it well-formed and
falsifiable), grounding (is it true) — and refuses to merge the verdicts. The stated target failure
is *confidence laundering*: real wins on the questions reasoning can reach (attribution, structure)
spent as authority on the one it cannot (truth). v2's concrete near-term goal is the acceptance
test — rebuild v1 in Go and reproduce v1's calibration distributions within sampling noise.
[GROUNDED: README.md §What it does / §Why; PLAN.md §2]

**Who is this for? Is it useless to me if I'm not already "trained to separate attribution from
structure from truth"?**
README names the audience as "people trained to separate attribution from structure from truth — and
who want to see that skill operating on material from their own work." The claimed mechanism is
recognition, not instruction: the tool does not install the skill, it makes a held skill visible (the
mirror hypothesis). So for an untrained user the *central* payoff is, by the repo's own account,
weaker — but the per-column verdicts (a refuted claim, an absent attribution) are still legible
without the meta-skill. Note this audience claim is itself carried as an unproven hypothesis, not a
finding. [GROUNDED: README.md §Who it's for; THEORY.md §The mirror hypothesis] [OPEN: design
analysis — usefulness to untrained users is not measured in the repo]

**What is the "mirror hypothesis," and has it been shown to be true?**
The mirror hypothesis is the repo's claim about its own effect: that seeing the three columns
disagree causes an already-trained user to recognise their skill as transferable — recognition, not
instruction. It has **not** been shown true. The test condition is field evidence from trained users
(not builders), which the repo does not have. The reflexive canary (`unverifiable` 5/5, haiku,
2026-06-11) confirms only that the grounding module does not self-confirm the claim from the armchair
— the correct null result at this stage, not a positive finding. CRITIQUE.md additionally flags the
repo's load-bearing thesis as itself ungrounded. [GROUNDED: THEORY.md §The mirror hypothesis;
LIMITS.md §Permanent limits] [LIMIT: CRITIQUE.md#the-reflexive-thesis-is-itself-ungrounded]

---

## What it is

**In one sentence, what is assay and what do I get from it?**
assay is a CLI that decomposes prose into atomic claims and returns, per claim, up to three
independent verdicts — faithfulness, substance, grounding — kept in separate columns so a win on one
is never spent as authority on another; what you get is a per-claim table of verdicts plus a JSONL
verification chain, not a single "good/bad" score. [GROUNDED: README.md §What it does; spec/CLI.md
§Output formats] [CONTRACT: binary not built]

**Is this a fact-checker?**
Only one of the three modes (grounding, `-evidence`) attempts truth, and the repo is emphatic that
even that is bounded: positive grounding is a retrieval not a deduction, and the `supported` verdict
currently only proves a URL was retrieved, not that the page backs the sentence. The other two modes
are explicitly *not* about truth — faithfulness checks sense-preservation, substance checks
well-formedness. So it is a fact-checker only in the narrow, heavily-qualified grounding column, and
a weak one to lean on there. [GROUNDED: README.md §What it does; THEORY.md §The axis boundary]
[LIMIT: CRITIQUE.md#the-given-gap]

**What do I actually see when I run it?**
Per the CLI *contract* (I cannot show a real run): results go to stdout — an ANSI table by default,
markdown with `-md`, and audit is always markdown — while progress, a `SUMMARY` block, and a `USAGE`
line go to stderr. Each mode prints a per-claim table (e.g. substance is `| # | Claim | Verdict | Why
|`), and a per-claim Tier-2 JSONL verification chain is written under `eval/<stamp>-<model>/`. I
cannot show actual output because the binary is not built. [CONTRACT: binary not built — spec/CLI.md
§Output formats; README.md §Status]

**One-sentence difference between the three modes — and when do I pick which?**
Faithfulness (`-source`) asks *did the source actually say it* — pick it when you have a summary and
its source transcript;
substance (default) asks *is the claim well-formed and falsifiable* — pick it when you have claims
and no external referent to check against;
grounding (`-evidence`) asks *is it true per external evidence* — pick it when the claim is
checkable against the world and you accept retrieval-bounded confidence;
`-audit` runs all three cross-tabulated and requires `-source`.
[GROUNDED: README.md §Usage; spec/CLI.md §Flags]

**Why three separate modes and verdicts instead of one "is this claim good?" score?**
Because the three questions reach different things, and merging them is the exact failure the tool
targets: a single score lets a win on faithfulness or substance (reachable by reasoning) be spent as
authority on truth (not reachable by reasoning alone). Keeping the columns separate is the structural
refusal of confidence laundering. Note CRITIQUE.md argues the substance/grounding line itself may not
be cleanly drawable in principle. [GROUNDED: README.md §Why; THEORY.md §The axis boundary]
[LIMIT: CRITIQUE.md#the-analyticsubstance-vs-syntheticgrounding-gap]

---

## Problem statements

**All useful work needs a problem statement. Which of the 3 questions (faithfulness, substance,
grounding) requires the problem to be stated?**
This is the subtle one and the glib answer is wrong. On the surface, faithfulness obviously needs an
external referent (the source transcript) and grounding needs one (the world / evidence), while
**substance looks like the self-contained mode** — it grades a claim's internal form with no external
problem stated. The context-principle critique inverts this: substance is precisely the mode whose
referents must be supplied. Decomposing into atomic fragments strips the context that fixes meaning —
the critic read "assay" as a chemical assay and rated "two of assay's three columns" hollow because
"columns" had no referent out of context. So substance is the mode that most needs the problem/context
stated, yet its contract pretends it does not, and it fails *silently* when context is missing.
Faithfulness states its problem explicitly (the source); grounding states it (truth against
evidence); substance hides the requirement. [GROUNDED: THEORY.md §The axis boundary; PLAN.md §3c]
[LIMIT: CRITIQUE.md#the-atomism-gap]

---

## Method choice

**We chose Dialectic. What are other ways of achieving our goals? Compare them to dialectic.**
*What the repo actually does (grounded):* substance runs a producer→critic dialectic over seven fixed
axes; faithfulness runs a defender→critic pair; grounding runs a single web-search call. [GROUNDED:
THEORY.md §§The axis boundary / The producer; spec/PROMPTS.md]
*Alternatives I am analysing (OPEN):*
- **Checklist / rubric scoring** (one LLM scores against fixed criteria): cheaper and more
  reproducible, but has no adversary — a plausible-but-hollow claim is likelier to pass, and it loses
  the steelman-then-attack separation THEORY.md uses to remove self-defense bias.
- **Verifier graded against an external score/database:** gives a real truth signal *where a database
  exists*, but only there; this is exactly the design THEORY.md contrasts itself with — it buys truth,
  not honesty about what was said.
- **Formal / deductive checking** (proof, type-checking, constraint solving): conclusive where claims
  are formalizable, useless on natural-language prose.
- **Ensemble / self-consistency** (sample N, take the mode): the calibration protocol already does
  this at the verdict level; it reduces variance but adds no adversarial axis.
Dialectic's distinctive value here is the asymmetry THEORY.md leans on (it can refute by deduction but
not confirm) plus critic-from-producer independence; its distinctive cost is the rule-following gap —
the axes underdetermine their own application. [OPEN: design analysis] [LIMIT:
CRITIQUE.md#the-rule-following-gap]

**What are alternatives to the Producer-Critic loop?**
(a) Single-pass critic with no producer — no steelman, so it risks attacking a weak version (the
strawman the producer exists to prevent); (b) multi-critic panel or two-advocate debate judged by a
third — more coverage, more cost, more variance; (c) self-critique / reflexion, one model grading its
own output — THEORY.md explicitly rejects this for self-defense bias; (d) iterative refinement with no
adversary (rewrite-until-clean) — improves prose, does not test survival; (e) human-in-the-loop
critic. The repo's current loop is specifically producer-blind-to-the-axes plus an
independent-context critic, chosen so the generator cannot shape output to survive a critic it can
anticipate. [OPEN: design analysis] [GROUNDED for current design: THEORY.md §The producer and the
limit of separation]

**How is assay different from a fact-checker, an LLM-as-judge eval harness, or a RAG hallucination
detector?**
A fact-checker collapses everything into true/false; assay refuses to, treating truth as one of three
columns and the least reachable. An LLM-as-judge eval harness grades an output against a task/rubric
to emit a score; assay's substance mode is adversarial (producer vs critic) and reports a verdict
*distribution*, not a scalar — and explicitly disclaims that the distribution is a property of the
claim. A RAG hallucination detector checks whether generated text is supported by retrieved context;
that maps closest to assay's *faithfulness* mode (sense-preservation against a source) — but assay
separates that from truth entirely, whereas a hallucination detector typically conflates "grounded in
the docs" with "correct." [GROUNDED: README.md §What it does; THEORY.md §The axis boundary] [LIMIT:
CRITIQUE.md#the-psychologism-gap] [OPEN: design analysis for the three comparisons]

---

## Trust & cost

**If the judge is itself an LLM that can be wrong, why should I trust any verdict?**
You should not trust any *single* verdict — the repo says so: a verdict from one run is "one draw"
from a distribution, and the calibration protocol reports full distributions per probe rather than a
point verdict. The narrower claim the repo does make is that some columns are more trustworthy than
others: faithfulness-with-source-present and refutation-of-internally-contradictory-claims are highest
(truth-maker in the window / reachable by deduction), while `supported` grounding and
specialist-domain substance are weakest. Trust is bounded and column-specific, never blanket.
[GROUNDED: THEORY.md §Verdicts as distributions; LIMITS.md §Operating envelope] [LIMIT:
CRITIQUE.md#the-psychologism-gap]

**The docs say a single run is "one draw." Do I have to run everything ~10 times? What does that cost
in time and API spend?**
The N=10 (canary N=5) protocol is the *calibration / acceptance* procedure — how the repo measures
the tool's own bias envelope — not a mandate for every production use. The README's Ballmer row and
PLAN.md §2 distributions are N=10 calibration results. For a one-off read you can run once and read
the verdict as one draw; you repeat only when you need the distribution (a high-stakes verdict, or to
detect skew). I cannot give time or dollar figures: the binary is not built and no benchmark exists.
assay does emit a per-run `est_usd` and a `USAGE` line, but those are contract outputs, not measured
numbers — and `est_usd` is null when the model is not in the price table. [GROUNDED: THEORY.md
§Verdicts as distributions; PLAN.md §2] [CONTRACT: binary not built — spec/CLI.md §USAGE / §usage-out]

**Which verdicts are safe to rely on, and which aren't?**
From LIMITS.md's operating envelope — *most reliable:* faithfulness with the source present
(truth-maker in the window) and refutation of internally-contradictory claims (deduction, not
retrieval); *reliable by design:* `unverifiable` on unresolved predictions. *Least reliable:*
`supported` grounding (proves a URL was retrieved, not that the page backs the sentence — structurally
weakest verdict); substance on sweeping forward predictions (skews `hollow` via Counterexample); and
substance on novel/specialist domains (producer and critic share a model, so a blind spot survives
both). Treat any `substantive` verdict with an empty critique as suspect — that is the documented
false-pass signature. [GROUNDED: LIMITS.md §Operating envelope / §Known biases and bugs] [LIMIT:
CRITIQUE.md#the-given-gap, CRITIQUE.md#the-externalism-gap]

---

## Limits & scope

**What is assay bad at? When should I not use it?**
Bad at: positive truth confirmation (the Given gap — grounding-as-retrieval cannot deliver a
truth-maker, possibly even in principle); relational defects like motte-and-bailey (`decompose`
splits the move into separate fragments and grades the pieces, so the move *between* them goes
ungraded); sweeping forward predictions (skew `hollow`); specialist/novel domains (shared-model blind
spot); and rhetoric/irony (no machinery for illocutionary force — it can literalize a performative).
Do not use it as an oracle, as a single-run truth verdict, or on decontextualised specialist claims
where the referent is severed. [GROUNDED: LIMITS.md §Known biases and bugs; README.md §What it can't
do] [LIMIT: CRITIQUE.md#the-given-gap, CRITIQUE.md#the-force--speech-act-gap,
CRITIQUE.md#the-externalism-gap]

**CRITIQUE.md says the "grounded?" column may be unreachable even in principle. So why ship a
grounding mode at all?**
Three reasons the repo's own logic supports. (1) The *negative* direction is reachable — reasoning can
refute (internal contradiction; settled-false predictions like the Ballmer forecast, `refuted` 10/10),
even if it can never confirm. (2) Shipping the column with its weakness printed (`supported` = URL
retrieved only) *is* the anti-laundering stance — hiding the column would hide the limit, the move the
tool exists to refuse. (3) CRITIQUE.md's own "Options for the future" lists marking grounding
permanently provisional, or dropping the completeness claim, as live strategic options — i.e. the repo
has not resolved this and presents it as a fork, not a finished feature. So grounding ships partly as a
refutation engine and partly as an unresolved design question, never as a working truth oracle.
[LIMIT: CRITIQUE.md#the-given-gap; CRITIQUE.md §Options for the future] [GROUNDED: LIMITS.md §Operating
envelope — refutation High, `supported` weakest] [OPEN: design analysis of the "why ship it" rationale]

**Does it work on specialist or technical claims?**
This is where the repo says trust it *least*. Producer and critic share one model (W1), so a blind
spot in the producer survives the critic; LIMITS.md marks specialist/novel domains "Lowest — treat
with most suspicion." The externalism critique sharpens it: the meaning of specialist terms is fixed
by a division of linguistic labor the critic cannot defer to, and decontextualised fragments sever
reference (the "assay"-as-chemical-assay collision); a one-line domain anchor patches symptoms but
does not restore the causal-historical reference chain. So it runs, but specialist claims are its
worst-grounded case. [GROUNDED: LIMITS.md §Operating envelope (specialist row); README.md "specialist
domains deserve the most suspicion"] [LIMIT: CRITIQUE.md#the-externalism-gap]

---

## History

**Where did this come from? What's the relationship to v1 (elenchus), and why a v2 rebuild?**
assay is the v2 rebuild of v1 (elenchus). `spec/` was copied byte-for-byte from v1 on 2026-06-13 and
is frozen — it is the input contract, never edited. v1 was a single Go file (`assay.go`, cited
throughout spec/CLI.md and spec/PROMPTS.md by line number); v2 re-implements it as a multi-package Go
project (internal/client, claims, modes, render; cmd/assay). v2's acceptance test is to reproduce v1's
calibration distributions within sampling noise — so the rebuild is validated against v1's *measured*
behaviour, not against fresh goals. The documented motivation is re-architecture into stateless
packages (PLAN.md §1, LESSONS.md references); a deeper "why rebuild at all" is not separately argued in
the v2 docs I read. [GROUNDED: CLAUDE.md §spec/ is frozen input; PLAN.md §1–2; README.md §Status]
[OPEN: the motivation beyond re-architecture is not spelled out in the v2 docs]

**Is it finished — can I use it today?**
No. README §Status: "the binary is not built yet." The internal/ packages are scaffold stubs
(`doc.go` package comments only), and the commands in the README are "the contract the binary must
satisfy, not a description of working software." You cannot run it today. [GROUNDED: README.md
§Status; observed: internal/*/doc.go and cmd/assay/doc.go are package-comment scaffolds only]
[CONTRACT: binary not built]

---

## LLMs

**Can we make this repo LLM-agnostic, to try different LLMs and compare performance?**
Today it is not. The model is an Anthropic model id (default `claude-sonnet-4-6`, override via
`-model` / `ANTHROPIC_MODEL`), grounding uses Anthropic-specific web-search tooling, and CLAUDE.md
mandates "one wrapper per external dependency." Making it agnostic would mean abstracting the client
behind an interface (provider → text + tools + usage) — feasible, since the modes already consume
parsed JSON rather than raw API types — but web-search grounding and citation handling are
provider-specific and would need per-provider adapters or a shared search seam. The calibration
baselines are model-specific (haiku, dated), so cross-model comparison means re-running the protocol
per model. [GROUNDED for current state: spec/CLI.md §Flags / §Env vars; spec/PROMPTS.md §6; CLAUDE.md
§API discipline] [OPEN: design analysis of the agnostic refactor]

**How would we measure that performance?**
The repo already defines the instrument: the calibration protocol — N runs per probe, fragment-level
distributions reported in full, false-pass counts read per `fragment_attribution` records, `error`
verdicts excluded from `n_valid` — run against the destructive probe fixtures with known engineered
defects. "Performance" here is not accuracy-versus-gold-label but the *bias envelope*: does each probe
land in its expected distribution (equivocation fires 10/10; reference-class and unfalsifiable-dress
stay at 0 substantive; the canary holds) and what is the false-pass rate. To compare LLMs you would
hold prompts and fixtures fixed, swap the model, and compare distributions and false-pass counts.
There is no single scalar score, by design. [GROUNDED: PLAN.md §2; LIMITS.md] [OPEN: extending the
existing protocol to cross-model comparison]

**Can we remove LLMs from this repo's structure completely? If so, how does the residual differ from
the LLM-enhanced whole?**
The scaffolding that survives without LLMs: decomposition-into-fragments as a pipeline, the seven-axis
taxonomy as a checklist, the producer/critic/defender control flow, the verdict enums, the JSONL
chain, the cross-tab, the calibration harness, and cost/usage accounting. What you lose is every step
that requires reading natural language — steelmanning a claim, finding distortions against a source,
searching for evidence, assigning a verdict. The residual is a bookkeeping-and-protocol skeleton: it
can enforce that the three questions are asked and kept separate, record verdicts, and run the
calibration math, but it cannot *produce* a verdict. So the LLM is the judgment; the structure is the
discipline that constrains and audits the judgment — and CRITIQUE.md's argument that the structure's
value (separation, anti-laundering) is partly independent of the judge still needs a judge for there
to be anything to separate. [OPEN: design analysis] [GROUNDED for what is mechanical vs model-driven:
spec/PROMPTS.md; spec/CLI.md]

**Can I run it against my own local model?**
Not as written. The client targets the Anthropic API (`ANTHROPIC_API_KEY` required, fatal if unset;
Anthropic model ids; Anthropic web-search tool). A local model behind an Anthropic-compatible endpoint
*might* work if it honoured the same API and tool surface, but grounding mode's web search and
JSON-tool behaviour are the hard part and are not guaranteed by a local model. There is no documented
local-model path. [GROUNDED: spec/CLI.md §Environment variables; README.md §Usage] [CONTRACT: binary
not built] [OPEN: no local-model support is designed]

---

## Uses

Each example shows the decompose → three-questions flow and names where a documented limit bites.
These illustrate the *contract*; none is a real run (the binary is not built), and no verdict below is
asserted as a measured result — only as how the flow would route and which axis/limit applies.

**Code review — worked example.**
Input: a PR description, "This refactor makes the parser 10× faster and eliminates all allocations in
the hot path." Decompose → `["the refactor makes the parser 10× faster", "the refactor eliminates all
allocations in the hot path"]`. *Faithfulness* (with the diff/commit as `-source`): did the change
actually do this, or does the summary overstate the diff? *Substance:* "10× faster" trips Base
rate/magnitude (a real quantity needs a comparison — versus what workload?); "eliminates all
allocations" trips Falsifiability/Counterexample (one allocation refutes "all"). *Grounding:* only
checkable if an external benchmark is retrievable. Limit that bites: specialist-domain suspicion —
producer and critic share a model that may not know the codebase, so a wrong `substantive` on a
domain-specific perf claim survives both roles; also, faithfulness mode is built for transcript
summaries, not diffs. [CONTRACT: binary not built] [LIMIT: CRITIQUE.md#the-externalism-gap; LIMITS.md
§Operating envelope (specialist row)]

**History books — worked example.**
Input: "The Treaty of Versailles caused World War II by economically crippling Germany." Decompose →
`["the Treaty of Versailles economically crippled Germany", "that economic crippling caused World War
II"]`. *Substance:* the second fragment trips Causality-vs-correlation directly (cause asserted from
association) and Hidden premise (it assumes a counterfactual). *Faithfulness:* applies only if you
have the source the summary came from. *Grounding:* the first fragment is partly checkable against
economic data; the causal claim is the contested, hard-to-confirm kind. Limit that bites: the atomism
gap — splitting "caused" from "crippled" grades each piece but loses the causal *move* that is the
actual historical claim, the same way decomposition dissolves a motte-and-bailey. [CONTRACT: binary
not built] [LIMIT: CRITIQUE.md#the-atomism-gap; LIMITS.md "Decomposition dissolves relational
defects"]

**Literature review — worked example.**
Input: a related-work paragraph, "Prior work (Smith 2021) showed method X outperforms Y; we are the
first to do Z." Decompose → `["Smith 2021 showed X outperforms Y", "no prior work has done Z"]`.
*Faithfulness* (with Smith 2021 as `-source`): did Smith actually show that, or is it overstated /
context-stripped? — this is faithfulness's strongest case, the truth-maker is in the window.
*Substance:* "first to do Z" trips Falsifiability and Base rate (what would refute it — one prior
instance). *Grounding:* "first to do Z" is a universal negative — refutable by one counterexample,
never confirmable by retrieval. Limit that bites: the Given gap — grounding can refute "first" by
finding a prior work but can never confirm it, and `supported` would only prove a URL was retrieved.
[CONTRACT: binary not built] [GROUNDED: LIMITS.md §Operating envelope — faithfulness-with-source is the
highest-reliability row] [LIMIT: CRITIQUE.md#the-given-gap]

**Vibe-coding — worked example.**
Input: a coding assistant's claim, "I've added input validation and the function now handles all edge
cases." Decompose → `["input validation was added", "the function handles all edge cases"]`.
*Faithfulness* (with the actual diff as source): did it add validation, or only claim to? *Substance:*
"all edge cases" trips Counterexample/Falsifiability (one unhandled case refutes it; "all" is the
classic hollow-skewing universal). *Grounding:* only checkable by running tests, which assay does not
do — it reads the prose claim, not the behaviour. Limit that bites: assay assays the *claim about* the
code, never the code — a confident-but-false "handles all edge cases" with a plausible rationale can
pass while the code is broken — and universal claims skew toward `hollow`. [CONTRACT: binary not
built] [LIMIT: LIMITS.md §Known biases and bugs (forward/universal skew); CRITIQUE.md#the-externalism-gap]

**Social-media influencer post decoding — worked example.**
Input: "This one morning habit boosted my productivity 300% — the science is clear." Decompose →
`["a morning habit boosted the speaker's productivity 300%", "the science is clear on this"]`.
*Substance:* "300%" trips Base rate/magnitude (a quantity needs a baseline), "the science is clear"
trips Evidence and Equivocation (a buzzword shield). *Faithfulness:* if decoding what a cited study
"shows," run it with the study as source — likely overstated/cherry-pick. *Grounding:* the
productivity number is anecdotal/unverifiable; "the science" is retrievable and often refutable. Limit
that bites: the force/speech-act gap — influencer copy is performative and hyperbolic, and grading it
on literal truth-conditions misreads register; faithfulness mode has a Literalization guard, but
substance and grounding can still treat hype as a sincere constative. [CONTRACT: binary not built]
[LIMIT: CRITIQUE.md#the-force--speech-act-gap]

**Science (exploring spectra libraries to deduce inorganic chemistry insights) — worked example.**
Input: a prose conclusion, "The 580 cm⁻¹ band confirms octahedral Fe–O coordination in the sample."
Decompose → `["a band appears at 580 cm⁻¹", "that band indicates octahedral Fe–O coordination", "the
sample contains octahedral Fe–O"]`. Be honest about what is happening: **assay validates the prose
claims, not the spectra.** It cannot read a spectrum or a spectral library; it can only assess the
sentences a chemist writes about them. *Faithfulness* would check the prose against a cited source
text; *substance* would test the inference (Hidden premise: the assignment assumes a reference
assignment; Evidence; Causality); *grounding* would web-search the claim. The limits bite hardest
here: grounding is retrieval-bound, so at best it finds a paper *asserting* the assignment — another
claim, not the truth-maker (the Given gap) — while specialist-domain suspicion is maximal: band
assignment is expert content fixed by a division of linguistic labor that the shared producer/critic
model cannot reliably defer to (the externalism gap; README: "specialist domains deserve the most
suspicion"). Do not use assay to *derive* chemistry insight; at most use it to flag overclaiming in
chemistry prose, and trust even that least. [CONTRACT: binary not built] [LIMIT:
CRITIQUE.md#the-given-gap, CRITIQUE.md#the-externalism-gap; README.md "specialist domains deserve the
most suspicion"]
