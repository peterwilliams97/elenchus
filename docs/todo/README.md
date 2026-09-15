# docs/todo/ — plans, orientation, and design seeds

The working set of "what to build next and why," split out of what had grown into one sprawling
file. Orientation (how the code is laid out) lives here too, because a plan is only readable next to
the map of what it changes. None of this is a contract — the contracts are `spec/`. None of it has
been actioned unless a file says so.

## Standing concern — don't lose adversarial review

The worry that outranks every capability below: **if the adversarial-review leg goes cold the tool
reverts to "trust the green checks" — the failure `examples/destructive/README.md` exists to
prevent.** The destructive probes are elenchus applied to assay's own critic.

- **Current on `claude-sonnet-4-6` as of 2026-09-14** — 8-probe calibration + three interventions in
  [`destructive-sonnet-2026-09-13.md`](destructive-sonnet-2026-09-13.md); the envelope held on every probe.
- **Ledger parser now reads chains** — `run.sh` reads each fragment's `"verdict"` from the per-run
  `*.substance.jsonl` chain, not stdout (the earlier stdout grep parse-missed every run).
- **Substance and grounding don't separate on a known-outcome claim** — a stated limit, not a bug
  (§ Conclusion, same doc): the substance critic imports an outcome it already knows.
- **Not a gate, by design.** The obligation: re-run `examples/destructive/run.sh` from a terminal on
  every default-model change — cheap to run, expensive to have silently lost.

## Files

- **`repo-map.md`** — orientation. What the repo does, the pipeline, the naming assessment
  (`assay` vs `elenchus`), the top-level shape, key symbols in `assay.go` and `internal/`, the axis
  boundary, and the bookkeeping files. Read this first.
- **`roadmap.md`** — the capability roadmap (was the root `TODO.md`). The numbered capabilities
  (validated summarisation, responding to critiques, standard code testing) with per-item responses.
  The deep-dive plans below expand its heavier items.
- **`plan-segmentation.md`** — the retrieval segmentation seam. Plan of record for roadmap items 2,
  6, 7, which all bottleneck on `retrieve.go` knowing only committee hearings. **Do this first** —
  one behaviour-preserving refactor unblocks three items.
- **`plan-backends.md`** — multi-provider LLM backends (OpenAI + DeepSeek/Qwen + Gemini). Roadmap
  item 3.3. Its prerequisite is W10 (verdict-enum validation), because non-OpenAI providers give
  JSON mode, not strict schema.
- **`plan-cleanup.md`** — top-level sprawl and the `cmd/`/`internal/` layout question. Housekeeping,
  not capability; the structural item is a decision, not a tidy-up.
- **`reference-material.md`** — pointer to `../elenchus_material/` (machine-local, uncommitted): the
  master plan (read) and three PDFs (unread by design, pending Peter's selection).
- **`design-notes.md`** — unsettled design seeds: whether surviving refutation is the best
  validation, and where the elenchus/substance leg stands today (dormant). May graduate to
  `BACKGROUND.md`.
- **`adversary-design-2026-09-14.md`** — built and calibrated on dora (`practical`); example scheme
  next: attack finding→recommendation edges with named defeaters. Of 46 F→R edges across the five
  corpora, dora's 7 in-scope now carry an edge verdict.
- **`Consultant & institutional reports.md`** — candidate test corpus: open-access consultant and
  institutional reports to analyse, grouped by subject (global/AU/Scotland, housing, AI-in-programming,
  EU competitiveness, AI-in-education), each with a **Test notes** column flagging where the argument
  structure is likely to strain. Top pick is the MIT Ad Hoc Committee report (§G) — a central
  recommendation contradicted by the report's own appendix surveys. None loaded yet.

## Suggested build order

Calibration is current as of 14 Sept; the open prerequisite for the adversarial leg is a scheme tag
on the edge in `spec/ARGUMENT.md` before any edge-level judge code.

Then, capability work:

1. **`plan-segmentation.md`** item 7 — the `Segmenter` seam (golden-test refuter, near-free). Opens
   items 6 and 2.
2. **W10** (verdict-enum validation, `SESSION.md`) — cheap, and the gate for both the segmentation
   follow-ons and any non-Anthropic backend.
3. **`plan-backends.md`** — the `openaicompat` backend, for cross-model calibration. Note the loop
   back to the standing concern: each new model is a new thing the destructive probes must be re-run
   against, never averaged over.

Housekeeping (`plan-cleanup.md`) and the naming/rename decision (`repo-map.md`) are independent and
Peter's to schedule.
