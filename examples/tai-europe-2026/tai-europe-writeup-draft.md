# **Does the European AI strategy's back half deliver its front half? — draft**

&nbsp;

*Draft for editing. Every number below is from `examples/tai-europe-2026/PLAN.md` and `adjudications.txt`; anything marked \[check\] I inferred and you should confirm before it ships.*

On 15 September 2026 the KIRA Center published *A Transformative AI Strategy for Europe* — 206 pages, eleven objectives, 795 inline citations, two Nobel laureates among the authors. The same day, Wolfgang Munchau wrote that reports like it "raise false hopes". That's a claim about the document. It can be checked, edge by edge, and neither he nor the authors did.

We checked one part of it with [assay](https://github.com/peterwilliams97/elenchus): does the implementation half (Part B, pp77–204) actually carry what the objectives half (Part A, pp16–76) promises? The document states each of its eleven objectives twice — once as a claim in Part A, once as an implementation in Part B — so this is a question the document itself sets up.

## The method, briefly

Each Part A objective statement was treated as a claim; Part B was held as the only source. An AI model (Claude Sonnet 4.6, three samples per claim) judged whether Part B supports the Part A statement, quoting the passages it relied on. Then I read all eleven pairs on the page and either upheld or overturned each machine verdict, with a one-line reason recorded per objective. The machine's verdicts, my overrides, and the reasons are all in the repository.

The cost was about $1.50 of model calls for the eleven \[check: excluding the comment-line bug run\] and about an hour of reading.

## What the document does

**Six of eleven objectives are not stated in Part B as they are stated in Part A.** They fall into three kinds:

1. **A rationale given in the front half that the back half never makes** — four objectives. Part A says Europe should reform broadly because it's hard to predict where AI value will land (O2.1); Part B argues for breadth on a different ground and never mentions unpredictability. Part A opens O2.2 with "experts disagree" about which layer of the value chain matters; Part B states the conclusion without the premise. Part A ties Europe's ability to absorb labour shocks to its share of AI value creation (O3.2); Part B doesn't. Part A says nothing else in the strategy is achievable without institutional capacity (IO-2); Part B doesn't say that either.  
2. **A factual specificity the back half doesn't carry** — one objective. Part A describes the loss of frontier access as "export controls on Anthropic's most capable models"; Part B says only that access was "suddenly withdrawn by foreign governments". The front half names a mechanism and a company the back half doesn't. Whether it was in fact an export control is a question for the cited source, not for this check.  
3. **No implementation at all** — one objective. O3.1, on enforcing the AI Act and GPAI Code, is stated in Part A as a mandate ("Europe must ensure that their enforcement is prioritised, targeted, proportionate…"). Part B's section for it is four lines: "there are no detailed recommendations for this objective."

**Five of eleven hold.** For those I overruled the machine, and the reasons matter for anyone using tools like this: the model had judged the objective against Part B's "why it matters" paragraph and not against the ACTION blocks that follow it. The commitments Part A makes — community benefit-sharing in the compute buildout (IO-3), crisis prevention and emergency response (IO-4), model-weight security tiers (IO-5), an anti-capture fund so at-risk companies don't need foreign capital (O1.2) — are all there, as legislation and timetables, in the actions. The fifth (IO-1) was a phrasing difference the model read as a claim.

## What the tool got wrong, and what we did about it

Two things, both now in the repository as recorded negatives.

The retrieval miss above: four false positives, all of one kind. For documents keyed by objective, the fix is to retrieve the whole section by ID rather than the best-matching paragraphs. Not yet built; noted.

The second was found by the positive control. We checked the document's headline compute comparison — Europe hosts three times the AI compute of one Malaysian site — against its own footnote (2.1 GW vs 0.662 GW). It holds; 2.1/0.662 \= 3.17. But the first run returned `contradicted` on a sentence whose own reasoning field ended "Verdict: faithful." The cause was the order of fields in the judge's output schema: it was asked for the verdict before the reasoning, so it committed to an enum before it had thought. We moved the verdict last and re-ran a 136-claim corpus we'd already adjudicated by hand: run-to-run noise on the old schema was 2%; the change moved 22% of verdicts, almost all milder; and on the five leaves with a human verdict, the judge went from agreeing with the human on none to agreeing on three. \[check: numbers from spec/TREE.md § Refuters\]

That is a better judge than we had on Monday, found by a control that was supposed to be boring.

## What this doesn't say

It doesn't say whether Europe can catch up, whether the objectives are the right ones, or whether the authors were complacent. The tool won't argue with a document about its goals; it only holds the document to what it says elsewhere. Munchau may be right. But "false hopes" is a claim about the relationship between a strategy's diagnosis and its prescriptions, and the honest version of that claim is narrower than his: one objective has no prescription, four have rationales that appear once and aren't carried through, and six are implemented as promised.

## Next

The diagnostic section — the twenty numbers on AI progress and Europe's position that the urgency argument rests on — each traced to its cited source. That's the part that would tell us whether the document's picture of "falling behind" is what its own sources say. It's the expensive part, and it isn't done.

---

*Reproduce: `examples/tai-europe-2026/` in the repository — PLAN.md (pre-registration and results), adjudications.txt (my eleven calls), evidence/ (the model chains).*

&nbsp;