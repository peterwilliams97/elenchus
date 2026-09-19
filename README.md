# elenchus

Reports are written to persuade.

[`assay`](assay.go) 
* reads a report against the documents it is based on
* asks, claim by claim, whether the sources bear each one out. 
* Propagates those checks up through the  findings and recommendations to the report's thesis. 
* Produces a summary for a human reader with a manageable number of claims that need to be checked. 
Each claim is shown  with the passage from the report that supports it. 

![review.html — the DORA report PDF on the left, assay's reading of the report on the right](docs/screenshots/dora-screenshot.png)

The worked example above is the 2025 DORA *State of AI-assisted Software Development* report (v.2025.2),
checked against its own figures and text: <https://peterwilliams97.github.io/elenchus-dora-2026/>.
It is two panes. 

* The **left** shows the source, in this case the report itself, open at p62 
(*Quality internal platforms*, Figure 44). 
* The **right** is assay's top-down reading: the report's root thesis,
each recommendation under it with a derived verdict, each recommendation opening to the findings it
rests on, and each finding opening to the atomic claims that support it.

The screenshot shows one recommendation being checked.

The report recommends investing in an internal platform (**R-PLAT**). That recommendation rests on
one finding (**F-PLAT**), and the finding rests on one claim (**CM10**): that with a good internal
platform, AI helps organizational performance more but also adds friction.

The AI judge marked CM10 **contradicted**, three runs out of three. It decided the report blames the
extra friction on small batches, not on platforms.

The judge was wrong. Page 62, open in the left pane, says: "Conversely, we found that AI's neutral
effect on respondents' reported experiences of friction is made harmful. That is, respondents
experience more friction in organizations with quality internal platforms."

A human reviewer therefore marked CM10 **faithful**. That verdict appears beside the machine's
**contradicted** badge. The verdicts above it are still computed by the code, so R-PLAT and F-PLAT
read **fails (machine; human disagrees: CM10)**.

On the right, the **report: §Quality internal platforms p62** link opens that page on the left. The
**reason** line is the judge's explanation, and the quote under it is the report text the judge
read. Clicking any `§` link or quote on the right opens its page on the left.


The top of the right pane has the thesis, the tally over the nine recommendations, and the
recommendations themselves (`examples/dora-2026/site/index.html`, contents from
`examples/dora-2026/argument.txt`):

```
AI's primary role in software development is that of an amplifier — it magnifies the strengths of
high-performing organizations and the dysfunctions of struggling ones; the greatest returns come not
from the tools themselves but from the underlying organizational system.

Of 9 recommendations, 3 hold (R-DATA, R-BATCH, R-USER), 3 are weakened (R-STANCE, R-ACCESS, R-VSM
(machine; human disagrees: VS4)), 2 fail (R-VC: no held source supports F-VC, R-PLAT: no held source
supports F-PLAT (machine; human disagrees: CM10)), 1 is opinion (R-TRANSFORM ?).
2 findings support no recommendation.

weakened  R-STANCE       Clarify and socialize your AI policies — establish a clear, communicated policy on permitted tools and usage.
holds     R-DATA         Treat your data as a strategic asset — invest in the quality, accessibility, and unification of internal data sources.
weakened  R-ACCESS       Connect AI to your internal context — give AI tools secure access to internal documentation, codebases, and data.
fails     R-VC           Embrace and fortify your safety nets — make teams proficient in rollback and revert features.
holds     R-BATCH        Reduce the size of work items — enforce the discipline of working in small batches.
holds     R-USER         Center users' needs in product strategy — keep the user as the product's North Star.
fails     R-PLAT         Invest in your internal platform — treat it as a product and the strategic prerequisite for unlocking AI's organizational value. (machine; human disagrees: CM10)
weakened  R-VSM          Use value stream management to turn AI investment into a competitive advantage. (machine; human disagrees: VS4)
opinion   R-TRANSFORM ?  Treat AI adoption as an organizational transformation, not a tools purchase — redesign workflows, roles, governance, and culture.
```

## What we did on the DORA report

1. **One source.** The corpus is the report itself (v.2025.2, 142 pages). The survey data behind it
   is not public, so the check is internal: does the report's own text and figures support what the
   report says elsewhere? Nothing here tests whether the survey numbers are true.
2. **Claims.** The report was broken into 136 atomic claims, each tied to a section and page, and
   arranged under the report's thesis, nine recommendations and their findings.
3. **Machine judging.** Claude Sonnet 4.6 judged each claim three times against passages retrieved
   from the report. Each claim gets a verdict and an agreement count such as 3/3.
4. **Derived verdicts.** Code, not a person, rolls the claim verdicts up the tree. Result: 3
   recommendations hold, 3 are weakened, 2 fail, 1 is opinion.
5. **Human check.** A human has read 11 of the 136 claims against the page and agreed with the
   machine on 4. Of the 7 disagreements, 4 were neighbouring labels (partial vs overstated, partial
   vs faithful), 1 was a problem the machine missed (FT2), and 2 were **contradicted** verdicts that
   were simply wrong (SD1, CM10). CM10 was wrong 3/3, and it is the reason R-PLAT shows as failing.
6. **Disagreement is shown, not hidden.** A human verdict sits beside the machine's and never
   replaces it. Any node that rests on a disputed claim says so:
   `R-PLAT fails (machine; human disagrees: CM10)`.

What this means for a reader: the other 125 verdicts are machine drafts that nobody has checked, and
a unanimous 3/3 is not proof of correctness. The tool's job is to put each claim next to the page it
rests on so that checking takes seconds. The reading is still yours.

## Worked examples

- <https://peterwilliams97.github.io/elenchus-dora-2026/> — the 2025 DORA *State of AI-assisted
  Software Development* report (v.2025.2), its headline findings checked against the report's own
  data, with the human adjudications shown beside the machine verdicts.
- <https://peterwilliams97.github.io/elenchus-vic-lceic/> — the Victorian LCEIC inquiry into the
  cultural and creative industries, checked claim-by-claim against the hearing transcripts and
  submissions. Run the same corpus twice and about 20% of leaf verdicts move (of 60 judged findings:
  48 settle, 3 wobble, 9 contested); a leaf that moves is marked contested and sent to the reader.

## How to read the right pane

Every node — the root, each recommendation, each finding — carries a **derived** verdict. The five
node labels, from `spec/ARGUMENT.md`:

- **holds** — every load-bearing child holds. The node stands on the ground its content claims.
- **weakened** — no child fails or opens, but some child is only partial, or wobbles between runs.
  The node still stands, but on narrower ground than its content claims.
- **open** — no load-bearing child fails, but the report never states what the node rests on (a `?`
  edge, or no finding attached), or a finding it rests on is contested (a leaf label, defined below).
  Not settled either way — the reader has to open it.
- **fails** — a load-bearing child fails: a supporting claim is contradicted, unsupported, or absent
  from the sources. One collapsed load-bearing premise collapses the node.
- **opinion** — the node is a pure value judgement. That is not evidence, so faithfulness cannot
  settle it and the tree does not pretend to.

**open** is a fact about the report, not about assay: it marks a place where the report never says
what its recommendation rests on — the material, not the tool that read it.

Under every finding sit its atomic claims, each with a **leaf faithfulness verdict** — whether the
source bears the claim out, never whether it is true (`spec/ARGUMENT.md`, `spec/TREE.md`):

- **faithful** — the source states the claim's subject, scope, and direction.
- **partial** — the source is adjacent, narrower, or broader: it bears on the claim without stating
  it (a scope, denominator, timerange, or attribution gap).
- **overstated** — the claim inflates the source: a hedge dropped, a number pushed, a certainty the
  source does not carry.
- **absent** — the cited document is held and was searched, but says nothing on the claim.
- **contradicted** — the source states the opposite of the claim.

(Two more the vocabulary carries: **unsupported** — no held source states it — and **unverifiable** —
the cited document is not in the corpus.)

A leaf also carries a **stability class** from re-running the judge. **contested** marks a leaf where
two identical runs land on different verdicts because the sources don't settle the claim either way —
a fact about the report, not about assay. A contested leaf **opens** the node above it: that is what
the node label `open` records when a finding is contested rather than left unstated.

Three more things ride on a leaf:

- **3/3** — the agreement. Each claim is judged N times (here N=3); `k/N` is how many of those
  samples landed on the verdict shown. `3/3` is unanimous; a split such as `2/3` is a claim the judge
  itself was unsure of.
- **so what** — one line, in plain words, naming the gap between what the report says and what the
  source says (`The report says X; the source only says Y`). It is the leaf's headline for a reader
  deciding whether to look closer.
- **reason** — the judge's own one-line justification for the verdict, citing the passages it read.

## Where the judgements come from

Three inputs, kept separate on purpose:

- The **report** supplies the propositions and the structure — each node's content, in the report's
  own words, and which findings a recommendation rests on.
- The **sources** supply the quotes each claim is checked against, carrying the document and page
  each came from.
- The **code** derives every verdict. No judgement is written by hand; each is computed from the
  leaves up (`spec/ARGUMENT.md`). Writing a verdict into the tree is the one edit the format forbids —
  it would let the builder's opinion pass as a result.

What counts as a source differs by run. For **LCEIC** the sources are external — the hearing
transcripts, submissions, and questions-on-notice, each quote carrying its document, witness, and
page — so a claim is checked against a document other than the one making it.

On both runs, machine verdicts a human has read are shown beside the machine badge with the
reviewer's initials, and the argument page reports how many adjudicated leaves the judge agreed on
(`spec/SERVE.md` § Adjudications).

## Run it on your own report

You supply four things:

- **claims** — the report's atomic claims, one file (`claims.txt`), plus `claims-machine.txt` mapping
  each claim to its report section.
- **argument** — the tree structure: root → recommendations → findings → claims, content only, no
  verdicts (`argument.txt`).
- **manifest** — `sources/MANIFEST.md`, mapping each source document's id to its file, witness, and
  locator.
- **sources** — the source documents themselves (`sources/hearings`, `sources/submissions`,
  `sources/qon`, `report.pdf`).

A faithfulness run (the model judge pass) produces the chain (`*.faithfulness.jsonl`) — this is
the model pass, roughly **$8 per run**. Everything below makes no model calls. The exact commands
(from [`examples/vic-lceic/current/PROVENANCE.md`](examples/vic-lceic/current/PROVENANCE.md), run
from that directory):

```sh
# Merge two grounding runs leaf-by-leaf (majority verdict) and render the report tree.
assay -from current.faithfulness.jsonl,current-2026-09-10.faithfulness.jsonl claims.txt -tree

# Render the argument tree, add per-leaf quote provenance, and write review.html.
assay -from current.faithfulness.jsonl,current-2026-09-10.faithfulness.jsonl \
  -argument ../argument.txt -manifest ../sources/MANIFEST.md -refs ../claims-machine.txt \
  -review \
  claims.txt

# One-off: attach each quote's passage id so provenance resolves (no model call).
assay -backfill-passages <chain> -source ../sources/hearings,../sources/submissions,../sources/qon
```

`-review` needs `-argument` and `-manifest`; it writes `review.html` alongside `index.html`, with
the report `§` and quote links rewritten to open the source PDFs in the left pane. `assay serve
<site-dir>` then serves the built site over HTTP so the `#page=N` links land on the right page.

## For readers

The narrative walkthrough of the LCEIC run is a Google Doc:
<https://docs.google.com/document/d/1GekNT4G8ESlQ__HrKafCTeaEOi7v2AVHBR3pqevM7HQ>.

## The name

Named for the Socratic *elenchus*: refuting a claim not by asserting its opposite, but by drawing out
what the claimant is committed to and showing where those commitments collide. The
[link](https://plato.stanford.edu/entries/plato-ethics-shorter/) is where the chain of rigour starts;
this repo is an attempt to continue it on modern work artifacts. It keeps three questions apart that
careless reasoning collapses into one: did the source actually say this? Is the claim well-formed?
Is it borne out by evidence? A claim can pass the first two and fail the third.

## Other modes

`assay` began as a single-document dialectical filter, and those modes still ship. Each runs one file
rather than a report tree; build with `./build.sh`, `ANTHROPIC_API_KEY` required.

- **Substance** (default) — `./assay memo.txt`. Decomposes the prose into atomic claims and runs a
  producer–critic dialectic on each: a Producer steelmans the claim blind to the attack, a Critic
  attacks on seven fixed axes (evidence, hidden premise, falsifiability, equivocation, base
  rate / magnitude, counterexample, causality vs correlation). Verdict:
  `substantive` · `partial` · `hollow`.
- **Faithfulness** — `./assay -source transcript.txt summary.txt`. Checks whether each claim in the
  summary is borne out verbatim by the source, never whether it is true. Verdict:
  `faithful` · `partial` · `overstated` · `absent` · `contradicted`.
- **Grounding** — `./assay -evidence claims.txt`. Web-searches for each claim's truth-maker. Verdict:
  `supported` · `mixed` · `refuted` · `unverifiable`. The only mode that touches truth; predictions
  correctly come back `unverifiable`.
- **Audit** — `./assay -audit -source transcript.txt -md summary.txt`. All three at once,
  cross-tabulated; the signal lives where the columns disagree.

Full flag list, design rationale, and the failure envelope are in `BACKGROUND.md` and `TESTING.md`.
