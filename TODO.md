# assay — capability roadmap

Items are capabilities, not bug fixes. Template:

```
## N. <Title>
**What.** one or two sentences.
**Why.** why it matters to the project.
**Done when.** a concrete, checkable definition of done.
**Open.** unresolved questions or first steps.
```

---

## 1. Validated summarisation

**What.** Take a summary plus its source and return a *validated* summary — every claim checked on
all three axes (faithful? / substantive? / grounded?) — so the user gets either a trustworthy
summary or a precise map of where it fails. `-audit` is the current surface; mature it into the
primary "validate this summary" workflow with diffable markdown output.

**Why.** This is what the whole tool is for. Faithfulness, substance, and grounding are means; a
validated summary is the end.

**Done when.** A user points assay at a summary + source and gets the summary back annotated by
validation, with the failure cases (`overstated` / `absent` / `refuted` / `hollow`) surfaced, in
diffable markdown.

**Open.**
- Decomposition for `-audit` is settled — keep the regex/authored-claim split, do NOT
  LLM-decompose, because faithfulness must evaluate the summary's claims as authored and
  re-decomposition breaks row alignment. Record this rationale so it isn't re-litigated.
- Concurrency (a bounded worker pool, ~3–5, mind rate limits and `-v` output ordering) is the
  main latency win and lives under this item.

**Response (2026-09-12).** Right item to call the headline — the trees and the judge already exist;
what's missing is the *diffable-markdown-over-source* surface, not new judgement machinery. `-audit`
+ `runEvidence` (`assay.go:1003`) already produce the verdicts; the work is a renderer that
interleaves them with the summary text, not a new mode. Keep the no-LLM-decompose rule as stated —
it's the same reason `-from`'s `loadClaimsFile` strips `#` lines while live runs feed `splitSummary`
(see `SESSION.md` "Deferred — root-block session", the title-header gotcha): row alignment is the
invariant. Concurrency is real but has a documented ordering hazard — `-v`/`-progress` output and the
`k/N` sampling in `faithJudgeRepeat` both assume sequential emission; a worker pool needs a
result-ordering buffer, not just a semaphore. Sequence it after the renderer so latency work doesn't
block the capability.

---

## 2. Responding to critiques

**What.** A capability to ingest a critique — of code, an argument, a claim set — and assess each
point for validity *against the artifact* rather than accepting it on authority, grading where the
critique's confidence tracks its correctness. Essentially the dialectic and grounding turned on a
critique: per point, `valid` / `overstated` / `wrong`, with reasoning tied to the thing critiqued.

**Why.** Authoritative-sounding critiques mix real and bogus points; the project's whole stance is
empirical-over-authoritative. The recent architectural critique was the manual prototype — two of
four points landed and the loudest one was wrong.

**Done when.** assay can take a critique plus the thing it critiques and emit a per-point verdict
with grounded reasoning, explicitly flagging any confidence/accuracy mismatch.

**Open.**
- Mode design: does this reuse the grounding pass against the artifact, or is it a new track?
- How is "the artifact" supplied — code? a transcript? a claim set?

**Response (2026-09-12).** Reuses the existing pass, doesn't need a new track: a critique point is a
claim, and "assess it against the artifact" is faithfulness with the artifact as the source corpus —
`faithJudge` already grades a claim against retrieved passages. The confidence/accuracy-mismatch
grading you want is the axis boundary in `CLAUDE.md`: a critique's point can be REFUTED by internal
contradiction with no lookup, but CONFIRMED only against the artifact (the truth-maker). So the
verdict enum maps cleanly — `refuted`/`overstated`/`supported` already carry that distinction, and
`supported` stays the one to read most skeptically. The real blocker is the "artifact supply"
question, and it's the *same* blocker as item 7: `retrieve.go` only segments committee hearings, so
critiquing *code* needs a code/transcript segmenter first. Build item 7's segmentation seam before
this, then this mode is mostly prompt + wiring.

---

## 3. Standard code testing.

1. Fuzz
2. Peturb prompts
3. Test mulitple LLMs. Codex?
4. Run LLMs multiple times to check variance
5. Write a result tester
6. Test on other reports
7. generalise internal/retrieve/retrieve.go to other reports

**Response (2026-09-12), per item.**

1. **Fuzz** — scoped and pre-registered: `SESSION.md` W4 targets `splitSummary`/`decompose` on
   run-on `and`/`;` blobs and single-line multi-claim inputs (`assay.go` regex envelope). Do that
   one first; it's the parser most exposed to malformed input. "No fabricated fixtures — derive
   adversarial inputs structurally" (the repo's hard rule) applies.
2. **Perturb prompts** — this is the `-*Sys`-prompt regression problem. It has a live signal source:
   the `examples/destructive/` calibration set (`testing/calibration_log.jsonl`). Any prompt edit
   re-runs that set and diffs verdicts. Note the existing tests only assert *prompt presence*
   (`TestFaithCriticSysLiteralization` etc.), not behaviour under perturbation — that gap is the work.
3. **Test multiple LLMs. Codex?** — the backend seam supports it now (`backend.Backend`, 4 methods);
   adding OpenAI is ~1 day (see `docs/repo-map.md` "Adding an OpenAI backend"). "Codex" as a model is
   retired — this means a gpt-4.1/gpt-5 backend. Cross-model verdict *drift* is already a tracked
   line (`SESSION.md` 3c/W7): append to `calibration_log.jsonl`, never average across models.
4. **Run LLMs multiple times to check variance** — **already built.** `-n N` →
   `faithJudgeRepeat` (`assay.go:1498`) returns modal verdict + `k/N` agreement, sampling at
   `judgeSampleTemp` for N>1. Gap: it's faithfulness-only; substance (`assayClaim`) does not
   repeat-sample. Extend it there if variance on substance matters, else mark this item done.
5. **Write a result tester** — collides with the hardest constraint in the repo: a result tester
   needs a labeled gold set, and W7 is BLOCKED precisely because no real one exists in-repo and
   labels must not be fabricated (`CLAUDE.md` hard rule; `SESSION.md` W7). Check whether
   `oracle_gen_test.go` is already the seed of this before building parallel machinery — verify its
   scope first. The honest scope: a *consistency* tester (re-run stability, schema-validity, quote-
   grounding via `quoteInPassage`) is buildable now; a *correctness* tester is blocked on the gold set.
6. **Test on other reports** — `examples/quocirca-2026/sources/report.pdf` is the untracked new
   corpus for exactly this. Blocked by item 7: a plain PDF report has no speaker turns, so
   `retrieve.Load` can't segment it. Order: 7 then 6.
7. **Generalise `retrieve.go`** — the split is already favourable. The BM25 index (`build`,
   `bm25Scores`, `Search`, `tokenize`) is corpus-agnostic; only the *segmentation* is hearing-
   specific (`passagesForFile`, `reportPassages`, `qonPassages`, `submissionPassages`, roster/role
   `classify`). Generalising = extract a `Segmenter` seam (corpus → `[]Passage`) with the hearing
   parser as one implementation and a plain-prose/paragraph parser as another; `Index`, `Search`,
   and the judge path stay unchanged. Role tagging degrades cleanly — a non-hearing passage has no
   role, and the judge already handles that. This unblocks items 2 and 6.

**Cross-cut.** Items 2, 6, 7 all bottleneck on one thing: `retrieve.go` only knows committee
hearings. Build the segmentation seam (7) first and three items open at once. Items 1, 4 (done), and
the consistency half of 5 need nothing new. Items 3 and the correctness half of 5 are gated on
external inputs (an OpenAI key; a real labeled gold set) that must be obtained, never synthesised.
