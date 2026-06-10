# BACKGROUND — design rationale & destructive self-criticism

This is a design-rationale and adversarial self-criticism document for elenchus / `assay.go`. It is
not a user guide (that is [README.md](README.md)) and not a change log (that is
[rigour-map/decision_log.jsonl](rigour-map/decision_log.jsonl)).
Its job is to state, as precisely as the code allows, **where this tool's verdicts cannot be trusted**
and to specify the destructive tests that would prove or bound each limit.

The tool's whole substrate is claim-validation; a document about its weaknesses is held to the same
standard as its inputs. Every claim here about external work carries its provenance.

## 0. Provenance ledger

| Source                                                                                | Status                          | How it is used here |
|---------------------------------------------------------------------------------------|---------------------------------|---------------------|
| STV — "Self-Trained Verification," Wu & Raghunathan (CMU), arXiv 2605.30290, May 2026 | Real, fetched, citable          | Cited normally |
| "Designing loops with Fable 5" (post)                                                 | **User-supplied; could NOT be independently fetched — X blocked automated access** | Every Fable-5-vs-Opus-4.7 figure is tagged `[UNVERIFIED, source claim]`. Single-author, unreplicated. Never presented as a grounded finding. |
| `examples/dan_shipper/`, `rigour-map/decision_log.jsonl`                              |  Real, in-repo                  | The only real corpora. No gold *labels* exist yet; none are fabricated here. |
| A calibration gold set                                                                | **Does not exist in this repo** | The W7 test specifies one as *required*. Until a real, provenanced set is obtained, that test is blocked — not faked. |

Per the hard rule in `CLAUDE.md`: if any input is synthetic or unverified, every conclusion drawn
from it is marked inconclusive. The Fable post's numbers are accordingly load-bearing for nothing in
this document; they are cited only as *claims the post makes*, never as measured facts.

---

## 1. elenchus and loop-based self-correction

elenchus and Fable-class self-correction loops share one mechanism and invert its purpose. Both sever
the verifier's context from the generator's, because a model critiques its own output poorly in the
same context:
Fable spawns a separate grader sub-agent;
elenchus makes the Producer blind to the critique axes (`producerSys`, assay.go:566) and runs the
Critic as a separate call (`substanceCriticSys`, assay.go:571). Fable's rubric of checkable criteria
is the analogue of elenchus's seven fixed substance axes, and "run until the rubric is satisfied" is
the producer–critic loop run to fixpoint (`assayClaim`, assay.go:428).

They diverge on [telos](https://en.wikipedia.org/wiki/Telos).
Fable's loop hillclimbs — make the artifact better.
elenchus's loop filters — make the claim's structure visible, downgrade, surface the residue — and
is indifferent to whether the claim wins. Only `substantive` and `partial` reach the residue;
everything else is shown and set aside.

They diverge on domain, and this is the load-bearing point.
Fable's wins — a training score, a SQL answer checked against a database — have truth-makers in hand,
so its grader is trustworthy because its [rubric](https://en.wikipedia.org/wiki/Rubric_(academic))
is grounded. That places Fable inside thec*reachable* region of elenchus's axis boundary (see
`CLAUDE.md`, "The axis boundary").
elenchus exists for the region with no checkable rubric, where an independent-context verifier is no
more grounded than self-critique. An independent context window removes the self-defense bias but
does not supply a truth-maker. This is STV's reference-conditioning result: the separate context is
not what makes a verifier good; being shown the reference is.

Fable's reported memory progression (fail → investigate → verify → distill → consult) is elenchus's
axes in motion. Its reported failure mode — a model writing "possibly X? Verify." into memory
then verifying roughly 17% of the time `[UNVERIFIED, source claim]` — is the laundering-confidence /
self-sealing failure with a number on it. "Verification coverage," on this reading, is effectively the
fraction of stored claims that crossed from *substantive* to *grounded*.

**Sharpest seam:** severing the verifier's context beats self-critique inside grounded domains;
everywhere else you still need the truth-maker, and Fable's own memory experiment is the cleanest
demonstration of why "everywhere else" is different.

---

## 2. ELENCHUS WEAKNESSES — a destructive-testing program

### 2.0 Framing: destructive testing, not confirmation

The existing suite is confirmatory by design. `TestConditionLaunderingDowngrade`,
`TestConditionLaunderingLoopStop`, `TestRunEvidencePropositionSubstitution`,
`TestCrossCheckEvidence*` — each feeds a *nominal* input and asserts the intended behavior fires.
That is regression protection. It is not characterization.

This program is the opposite. The goal is not to show elenchus passes good inputs; it is to
**engineer adversarial inputs that push each axis past its limit and find where it breaks** — the
same posture as proof-loading a beam until it yields, so the rated load is known rather than hoped.
A green result on a destructive test means the failure mode did not occur.
An *honest failure* means the test exposed a real limit; the disciplined response is to record it in
the operating envelope (§2.3), not to soften the test until it passes.
Several tests below (W3, W10 especially) are **expected to fail today** — the failing test *is* the
documented limit.

One asymmetry runs through several rows: substance mode atomizes with the LLM call `decompose`
(assay.go:418); faithfulness, evidence, and audit atomize with the pure-regex `splitSummary`
(assay.go:949). They have different failure surfaces and must be tested separately.

A standing correction this document records (and flags for a separate maintenance edit, not made
here): `README.md:164–166` still says `-evidence` source URLs are merely model-retyped and unchecked.
That is now **stale**. `d009`/`d010` added `crossCheckEvidence` (assay.go:512), which cross-checks
model-reported URLs against URLs actually pulled from `web_search_tool_result` blocks
(assay.go:792–808), downgrading to `unverifiable` when they do not match. The real gap (W3) is
*narrower* than the README claims and should be restated at its true size.

### 2.1 Weakness taxonomy

Substance: LLM `decompose` (418). Faithfulness / evidence / audit: regex `splitSummary` (949).

| #   | Weakness                                | Code locus | How it manifests |
|-----|-----------------------------------------|--------|---|
|  W1 | Correlated producer/critic blind spots  | `assayClaim` runs `producerSys` then `substanceCriticSys` through the *same* `c.model` via the single `cfg.call` seam (assay.go:64, 433, 440) | A flaw the model cannot see as Producer it also cannot catch as Critic. Making the Producer blind to the axes removes *self-defense* bias but not *shared parametric* blind spots. Connects to the STV asymmetry: separate context ≠ grounded verifier. |
|  W2 | Confirmation unreachable by reasoning   | `evidenceClaim` (491) can emit `supported`; `crossCheckEvidence` (512) checks **URL provenance**, not whether the retrieved page's *content* supports the claim | An internally-coherent, externally-unanchored claim can be returned `supported`; a real-but-misread URL passes the cross-check (the URL was retrieved) while the support is fictional. Reasoning can refute, never confirm. |
|  W3 | Citation-span provenance gap (narrowed) | `callClaude` extracts `retrievedSource{Title,URL}` from `web_search_tool_result` (792–808); the model's `sources` are model-typed JSON matched host+path by `normalizeURL` (551) | The cross-check confirms *a* cited URL was among those retrieved — not that *this sentence's* proposition is backed by *that span*. True API citation blocks are never read. Real, but smaller than `README.md:164–166` states. |
|  W4 | Decomposition / atomization errors      | `splitSummary` (949) is pure regex (`leadingMarkerRe` 945, `runOnMarkerRe` 946); `decompose` (418) is one LLM call | A paragraph joining three claims with "and"/";" and no list markers → **one mega-claim** graded as a unit (faith/evidence/audit). `decompose` may instead over-split or merge nested hedges. Either way every downstream verdict lands on the wrong unit. |
|  W5 | Literalization / intent-reconstruction error | `intendedProposition` (316) substitutes `what_source_actually_says` when the faith verdict ∈ {partial, overstated}; the reconstruction comes from `faithCriticSys` (616) | If the faith critic reconstructs the *wrong* intended proposition, `evidenceClaim` grounds the wrong thing — confidently grounding a claim the speaker never made. |
|  W6 | Axis incompleteness                     | The seven fixed axes in `substanceCriticSys` (571) | Category error, motte-and-bailey, and reference-class gaming *beyond* base-rate are not axes; such defects can pass `substantive`/`partial` uncaught. |
|  W7 | Calibration drift across models         | The whole pipeline keys on `c.model`; no gold set in repo | The same fixture on haiku vs sonnet vs opus can yield different verdicts; the magnitude is unknown and uncharacterized. Measuring it requires a **real labeled gold set that does not yet exist**. |
|  W8 | Survivorship bias in the corpus         | `rigour-map/decision_log.jsonl` | Dead ends were historically unlogged (d006 had to be backfilled). Any classifier or eval built on the corpus inherits a denominator missing its own failures. |
|  W9 | Headline self-claim unfalsifiable from the armchair | `README.md:27–32` | The value prop — making held rigour visible causes a trained user to recognise it as transferable — is a grounding claim about effects on people. By the tool's own axis boundary it cannot be settled by reasoning. A permanent limit, not a TODO. |
| W10 | No verdict-enum validation             | `unmarshalLoose` (937) accepts any string into `Verdict`; no caller checks membership against the documented enums | A model emitting `"verdict":"yes"` passes straight through and renders as if it were a real verdict. Cheap to test, currently unguarded. |
| W11 | Single-reparse / truncation collapses signal to `error` | `callJSONSourced` retries once (683–693); `maxTokens=1500` (assay.go:35) returns a truncation error → `assayClaim` yields `Verdict:"error"` | Long critiques or stubborn malformed JSON collapse to `error`, which the tally counts as "not verified" — silently shrinking coverage rather than surfacing *why*. |

### 2.2 Destructive-test specifications

"PASS" = elenchus surfaces the limit correctly (the failure mode does not silently occur).
"Honest FAILURE" = the test exposes a real limit; record it in §2.3 rather than patching it green.
None of these is implemented in this pass — they are specifications. Implementation is the next
tracked step (`SESSION.md`).

| #   | Destructive test (spec) | PASS looks like | Honest FAILURE looks like |
|-----|---|---|---|
|  W1 | Curated claims carrying a known *shared-misconception* flaw (something the model itself believes wrongly). Run full substance through the real `c.model`. | Critic flags the flaw; verdict is not `substantive`. | Producer *and* Critic both miss it → `substantive`/`partial`. Record: blind-to-axes ≠ independent grounding; specialist-knowledge claims sit outside the envelope. |
|  W2 | Red-team set of internally-coherent, externally-unanchored claims (plausible, un-retrievable). Run `-evidence`. | All return `unverifiable` (no truth-maker present). | Any returns `supported`/`refuted` from reasoning. Record: positive grounding confirmation is unreachable; `crossCheckEvidence` guards URL-presence, not content-support. |
|  W3 | Assert the displayed `-evidence` source URLs originate from **API citation blocks**, not the model's `sources` JSON. | (not achievable on today's code) | **FAILS today by design** — the code reads `web_search_tool_result` URLs and model-typed `sources`, never citation blocks. This failing test is the recorded limit; the fix is future work. |
|  W4 | Fuzz `splitSummary` and `decompose`: run-on claims joined by "and"/";" with no markers; nested hedges; single-line multi-claim blobs. | Atomic units recovered — or an explicit, stated cap on what the regex can split. | A 3-claim blob → 1 unit, or a hedge split mid-claim. Record `splitSummary`'s regex envelope and the LLM-vs-regex asymmetry. |
|  W5 | Source whose intended proposition is subtle; capture the string that reaches `evidenceClaim`. | `intendedProposition` routes the correct reconstruction; grounding hits the real claim. | The wrong proposition is grounded. Record intent-reconstruction as an error surface upstream of grounding. |
|  W6 | One probe claim per missing defect: a category error, a motte-and-bailey, a reference-class-gamed statistic. Run substance. | At least flagged via an adjacent axis (e.g. Equivocation catching the motte-and-bailey). | Passes `substantive`/`partial` clean. Record the three named gaps as out-of-envelope until axes are added. |
|  W7 | Same **real, provenanced** fixture across haiku/sonnet/opus; measure verdict agreement against gold labels. | Drift within a stated tolerance band → the envelope is quantified. | Wide disagreement → characterize the drift envelope; until then, treat single-model verdicts as model-conditional. **Blocked on a real labeled gold set; if none can be obtained, report "could not obtain real gold set" — never fabricate labels.** |
|  W8 | Audit `decision_log.jsonl`: ratio of `abandoned`/`reverted`/`superseded` to `shipped`; flag implausible all-survivor stretches. | Dead-end rows present at a plausible rate. | Near-zero logged failures → the corpus is contaminated; any eval or classifier built on it is marked INVALID per the provenance rule. |
|  W9 | No runnable test exists. The check is that this document *states* W9 as a hypothesis awaiting evidence, never as a settled finding. | The doc carries it as unsettled. | The only failure mode is *claiming it settled* — a discipline lapse, caught in review, not in code. |
| W10 | Stub `cfg.call` to return `"verdict":"yes"`. | The pipeline rejects or normalizes the non-enum verdict. | Garbage renders as a verdict → add enum validation (future work). Expected to fail today. |
| W11 | Stub a `max_tokens` truncation and a persistently malformed JSON; observe verdict and tally. | The `error` carries a cause and the tally shows reduced coverage explicitly. | Coverage silently shrinks with no surfaced reason → record the `error`-masks-signal behavior. |

### 2.3 Operating envelope

The point of the program above is to draw this line and keep it drawn.

**Reliable region — verdicts you can lean on:**

- **Faithfulness with the source present.** Close reading reaches faithfulness: the truth-maker (the
  source transcript) is in the context window, and both Defender and Critic can quote it
  (`faithClaim`, assay.go:468). This is the one mode whose truth-maker is *supplied*, and it is the
  tool's most trustworthy column.
- **Refutation of self-contradictory claims.** Reasoning can kill a claim that contradicts itself
  with no lookup — an internal contradiction is a deduction, not a retrieval. The substance axes and
  a `refuted` grounding verdict on an internally incoherent claim are inside reach.

**Unreliable region — verdicts that need a truth-maker the tool may not have:**

- **Positive grounding confirmation** (W2/W3). `supported` is the structurally weakest verdict:
  `crossCheckEvidence` proves a URL was retrieved, not that it supports the sentence. Confirmation
  always needs the truth-maker, never the reasoning.
- **Specialist-knowledge claims** (W1). Where the model is wrong in both roles, the dialectic cannot
  rescue it.
- **Novel domains** with thin parametric coverage — same mechanism as W1.
- **The tool's own headline** (W9) — unsettleable from the armchair, by construction.

The honest summary: two of assay's three columns (substance, faithfulness-without-source) cannot reach
the thing the third column (grounding) is for, and the third reaches it only by retrieval, never by
reasoning. The characteristic failure is *laundering confidence* — scoring real wins on the reachable
columns and then pronouncing on grounding with borrowed authority the first two never licensed.

**The real skill the tool trains is catching where the machine critic is wrong.** Every weakness above
is a place where a fluent, confident verdict can be unearned. Using elenchus well means reading its
verdicts as a competent-but-fallible colleague's, hardest on exactly the `supported` calls and the
intent reconstructions where it has the least ground under it.

---

## 3. LIMITS OF FABLE-CLASS TOOLS / loop-based self-correction

These limits are the mirror image of §2: where elenchus's weakness is that it operates *without* a
truth-maker, the loop-based tools' weakness is that they quietly assume one.

### 3.1 Verifier bottleneck

The loop is only as good as its grader. Severing the verifier's context removes self-defense bias but
not *ungrounded-rubric* risk: when the rubric is a proxy for the goal rather than the goal itself, the
loop inflates the proxy — score inflation / reward hacking. STV's Figure 8 `[per STV, arXiv
2605.30290]` is the reference point: an independent context is necessary but not sufficient; being
shown the reference is what moves verifier quality.

### 3.2 Truth-maker dependence

These loops work where a real truth-maker exists — a training score, a database answer, a passing test
suite — and degrade *exactly* where elenchus is needed: claims about people, the future, contested
domains, anything with no checkable rubric. The capability that makes Fable strong on graded tasks
does not transfer to ungraded ones; it merely *feels* like it does.

### 3.3 Memory failure modes

Low verification coverage means unverified-but-plausible claims accumulate and compound across
sessions. The reported "possibly X? Verify." pattern, verified ~17% of the time
`[UNVERIFIED, source claim]`, is the concrete form: a note that *looks* like diligence but never
crossed from substantive to grounded. Worse, distilling such a claim into a "general rule" propagates
the error with added authority — the self-sealing failure, scaled.

### 3.4 Goodhart on the rubric

A well-formed rubric can be the wrong rubric. The loop optimizes the rubric it was given, not the goal
the rubric was meant to stand for. Tightening a loop against a misspecified criterion makes the output
more confidently wrong, not more correct.

### 3.5 "Design loops, don't steer" assumes capability implies calibration

The bitter-lesson-leaning advice to design loops and let capability do the rest assumes a capable model
is also a calibrated one. In grounded domains that holds — the truth-maker calibrates it. In ungrounded
domains a loop can converge confidently wrong, because nothing in the loop is pulling toward truth, only
toward rubric-satisfaction.

### 3.6 Benchmark-claim provenance

The "Designing loops with Fable 5" post's Fable-5-vs-Opus-4.7 numbers are single-author and
unreplicated, and could not be independently fetched (§0). Applying elenchus's own provenance rule to
the post: those figures are `unverifiable` against current evidence. They may be true; they are not
*shown* true here, and nothing in this document rests on them. Treating a single-source benchmark claim
as settled would be the exact laundering-confidence move the tool exists to catch — committed against
the tool's own source material.

---

## 4. References

- Wu & Raghunathan, "Self-Trained Verification," CMU, arXiv 2605.30290, May 2026. **Real, fetched.**
  Cited for: reference-conditioning result (independent context ≠ grounded verifier; being shown the
  reference is what helps) and the reward-hacking / score-inflation result (Figure 8).
- "Designing loops with Fable 5" (post). **User-supplied; could not be independently fetched (X blocked
  automated access).** All quantitative claims (Fable-5-vs-Opus-4.7 figures, ~17% verification
  coverage) are tagged `[UNVERIFIED, source claim]` throughout and are presented only as claims the
  post makes, never as findings.
- In-repo: `CLAUDE.md` ("The axis boundary"), `README.md` ("Honest limits"),
  `rigour-map/decision_log.jsonl` (d009/d010 grounding-integrity line).


## Footnotes

self-sealing = an unverified claim that erases its own "unverified" flag and then gets promoted to
a rule, so the error spreads with more authority than it ever earned, and the system can no longer
tell it was never checked.
