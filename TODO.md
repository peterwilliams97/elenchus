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
