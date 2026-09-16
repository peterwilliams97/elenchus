# design-notes.md — design seeds

Unsettled design rationale. Seeds for work, not conclusions. Anything here that hardens may graduate
to `BACKGROUND.md`, where the design rationale lives.

## Is surviving refutation the best validation? (2026-09-12)

Peter: "I don't know of a better validation of any claim than surviving refutation." The position is
mostly right and its exact failure is the line `assay` is built on.

- **Where surviving refutation IS the ceiling.** Claims with no reachable truth-maker — theories,
  forecasts, universals. Popper: science corroborates, never verifies. Nothing beats attacking the
  claim and watching it stand.
- **But refutation is asymmetric — a decisive *falsifier*, never a decisive *validator*.** One
  counterexample kills a universal (modus tollens); survival proves nothing (affirming the
  consequent). So the honest word is *corroboration* — "not yet shown false, and it stuck its neck
  out" — not *validation*. Upgrading "un-refuted" to "validated" is the "laundering confidence"
  failure in `CLAUDE.md`.
- **For a correspondence claim with an accessible source, retrieval strictly dominates.** "The
  report cites D on page 12" — open page 12; no amount of armchair refutation substitutes. This is
  the axis boundary verbatim: reasoning can refute a grounding claim but never confirm one. `assay`
  doesn't out-argue "the source says X" — it reads the source. The tool is a monument to
  refutation-survival being insufficient here.
- **Survival's strength has an invisible bound: the refuter's imagination.** Surviving refutation
  only means surviving the refutations someone thought to try — a lower bound set by the adversary,
  not a property of the claim. Hence the repo's own rules: a refuter must name (not count) its cases;
  Producer and Critic share one `c.model` so a blind spot survives in both; `survives_only_by_
  conditioning` catches a claim that survived by dodging into unfalsifiability rather than by being
  true.
- **The reframe.** Refutation and grounding aren't rival validators — refutation is the cheapest way
  to establish *falsity* (no lookup), retrieval the only way to establish *positive support*. "Best
  validation" is claim-type-dependent, and the axis boundary already draws that map. Keeper:
  **surviving refutation is the best test of a claim's coherence and its nerve; it is never, on its
  own, evidence the claim is true — for that, the retrieval half exists.**

## Single-source adjudication: which verdict does the human file against? (2026-09-16)

On a `single_source: true` corpus `presentArgument` (`assay.go`) renames the machine's `absent` to
`uncorroborated` before it reaches the tree, because with the report as its own only source an
`absent` is not a grounding miss but a claim the report states once and no second document repeats —
it derives as *weakened*, not *failed* (`spec/ARGUMENT.md`, spec/SERVE.md § Adjudications). The leaf
card therefore *displays* `uncorroborated`, while the raw chain verdict is `absent`.

The **count** half of this is settled: `adjudicate.Agree` scores the human against the raw verdict
(the `machine` map is captured before the remap) and `canonVerdict` folds `uncorroborated`→`absent`,
so a human who wrote either spelling agrees with a chain that recorded `absent` (committed 2026-09-16,
`59e6cfb`). That closes the earlier `tai-europe-2026` slice-1 discrepancy — raw 6/11 vs a rendered
4/11 — at 6/11 (`examples/tai-europe-2026/PLAN.md` § Adjudicated 16 Sept).

**What is still unsettled** is the display, not the count: a human reading the page sees the machine
badge say `uncorroborated` but, filing from the raw verdict, writes `absent` — the human chip and the
machine badge then show different words on a leaf they *agree* on. Two ways out, and it is a UX call,
not a counting one: (i) show the human the raw `absent` on the card so the spellings match what they
type, or (ii) guide the human to file the display spelling `uncorroborated`. The fold makes both
harmless to the count, so this is cosmetic-plus-clarity, low priority. Peter's call.

## The elenchus leg is the dormant one (2026-09-13)

`examples/destructive/README.md` is the repo's adversarial/Socratic heart: its seven substance
probes (equivocation via motte-and-bailey, base-rate, hidden-premise, falsifiability, causality,
plus the axis-gaps and calibration probes) are exactly the moves of a Socratic cross-examination.
It is elenchus — but *meta*: probes that test whether assay's critic catches the fallacies, not the
tool performing elenchus on a target report.

Two facts about its current status, both relevant to the naming question in `repo-map.md`:

- **It is deliberately not a gate.** `run.sh` is not wired into `build.sh` or `go test` — by design
  ("a flaky pass/fail gate on a non-deterministic judgment just trains everyone to ignore
  failures"). `build.sh` runs `go test ./...` then builds; there is no `test.sh`. So the automated
  suite covers plumbing/orchestration (`TESTING.md` Layers 1–2), never the substance judgment
  (Layer 3).
- **In practice it has gone cold.** The last calibration in `testing/calibration_log.jsonl` is
  dated **2026-06-11, on `claude-haiku-4-5` only** — ~3 months stale as of this note, and never run
  on the current default (`claude-sonnet-4-6`) or opus. `SESSION.md` already flags the cross-model
  re-run as deferred.

So the elenchus/substance leg is doubly demoted: the axis boundary marks it the *weakest* validator
(cannot confirm grounding), and its calibration is the *least-exercised* part of the instrument. If
the repo keeps the `elenchus` name, reviving this calibration on current models is the work that
would earn it back; otherwise it is the clearest evidence the centre of gravity has moved to
faithfulness. Cross-references roadmap item 3.2 (perturb prompts / re-run the calibration set).
