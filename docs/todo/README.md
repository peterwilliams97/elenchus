# docs/todo/ — plans, orientation, and design seeds

The working set of "what to build next and why," split out of what had grown into one sprawling
file. Orientation (how the code is laid out) lives here too, because a plan is only readable next to
the map of what it changes. None of this is a contract — the contracts are `spec/`. None of it has
been actioned unless a file says so.

## Standing concern — don't lose adversarial review

The one worry that outranks every capability below: **the adversarial-review leg is going cold, and
if it dies the tool quietly reverts to "trust the green checks" — the exact failure
`examples/destructive/README.md` was built to prevent.** The destructive probes are elenchus applied
to assay's own critic, and they are the only thing standing between the instrument and unexamined
trust. Current status (`design-notes.md`, full detail):

- **Not a gate, by design** — `run.sh` is not wired into `build.sh`/`go test`, so nothing forces it
  to run. That is a deliberate choice (no flaky judgment gate), not neglect — but it means the leg
  survives only if someone re-runs it on purpose.
- **Stale** — last calibration `2026-06-11`, `claude-haiku-4-5` **only**; never run on the current
  default `claude-sonnet-4-6` or opus. ~3 months cold.
- **Weakest *and* least-exercised** — the axis boundary already marks substance/elenchus the weakest
  validator; its calibration being the least-tested compounds the risk rather than excusing it.

Treat reviving and *keeping* this calibration current on every model change as a first-class,
recurring obligation, not roadmap step 4. It is cheap to run and expensive to have silently lost.

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

## Suggested build order

**Before anything else:** revive the `examples/destructive/` calibration on the current models
(`design-notes.md`, roadmap 3.2) and keep it current — the standing concern above. It is a
prerequisite for trusting *any* result the work below produces, and it gets cheaper to lose the
longer it waits.

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
