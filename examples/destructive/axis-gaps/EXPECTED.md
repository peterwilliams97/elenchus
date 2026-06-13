# EXPECTED — axis-incompleteness probes

Distribution over runs, never a golden verdict. **And no "pass/fail" here** — the outcome is unknown
by design and the result *maps the envelope either way*.

## Envelope (pre-registered, before calibration)

There is no target axis and therefore no "should fire." Pre-registered guesses (to be checked, not
graded):

- **Category error** (monolith / honest) — **likely passes clean** (`substantive` or `partial`). No
  axis names category mistakes, and no adjacent axis fits well. A clean pass here is the cleanest
  illustration of an axis gap.
- **Composition** (excellent engineers → excellent team) — **may be caught** via **Counterexample**.
  If caught, the catch is via an adjacent axis, not a "composition axis."
- **Survivorship** (founders who changed the world) — **may be caught** via **Base rate / magnitude**
  (the missing denominator of failures). If caught, likewise adjacent.

## How to read the result

- A defect **passes `substantive`/`partial` with no relevant finding** → recorded as an **out-of-
  envelope limit** (BACKGROUND.md W6). This is an *honest failure*, logged, not patched. It tells a
  reader exactly which reasoning defects the seven axes miss.
- A defect is **caught via an adjacent axis** → recorded as evidence the critic reaches slightly
  beyond its named axes. Note *which* axis fired.

Both outcomes are findings. The only way to "fail" this probe is to misreport a miss as a catch (or
vice versa), or to quietly edit the fixture until it goes green.

## What a clean run means

Nothing about "the critic is correct." It maps where the instrument's named axes stop. Date the
result; mind the training-data caveat in `../README.md`.

## Observed — 2026-06-10, `claude-haiku-4-5-20251001`, N=10

Distribution (3 input claims, `decompose` expanded to ~6–7 atomic/run): **55 hollow · 10 partial ·
4 substantive.** Full data in `results/2026-06-10-claude-haiku-4-5-20251001.md`.

**The un-named defects were caught MORE than the gap analysis predicted — via adjacent axes.**

- **Composition** (excellent engineers → excellent team): **Counterexample** fired 10/10 runs. The
  "excellent individuals, dysfunctional team" counterexample is squarely on an existing axis, so
  composition is well-covered by the adjacent reach — *not* the clean gap the pre-registration
  guessed.
- **Survivorship** (founders who changed the world): **Base rate / magnitude** fired 8/10. The
  missing-denominator-of-failures reads as a base-rate omission, so survivorship is also substantially
  reachable.
- **Category error** (monolith / honest) is the most likely source of the **4 `substantive` leaks** —
  no axis names category mistakes and none fits adjacently, so it is the cleanest *true* gap of the
  three.

**Envelope mapped (the whole point):** the seven axes' adjacent reach covers composition and
survivorship better than the gap taxonomy assumed; **category error remains the genuine un-named
gap.** This is a finding, not a pass or a fail. The 4 `substantive` verdicts are *mapped limits*
(BACKGROUND.md W6), recorded, not patched.

## Observed — 2026-06-11, `claude-haiku-4-5-20251001`, N=10 (fragment-attributed)

Fresh calibration with per-fragment chains persisted (2026-06-10 chains discarded by an instrument
error — `../../../testing/SCHEMA.md`, `decision_log` d016 — a *new sample*, not a rescore).
Distribution (70 fragment-verdicts, `claims_per_run` 7): **55 hollow · 14 partial · 1 substantive ·
0 error.** `N_valid` = 70.

**Fragment attribution of the lone `substantive` verdict:** it is on the **survivorship** fragment
(run 5, "the founders who changed the world all ignored the skeptics"). The survivorship-specific
catches (**Base rate / Counterexample**) were **silent** on it; only Hidden-premise fired weakly. In
the *same run*, a sibling fragment ("…bet everything on one idea") was caught `hollow` with an
explicit **"Survivorship bias / Base rate"** axis — so the defect was reachable, just not on this
fragment. Recorded as **1 mapped-limit leak** (the un-named-gap behaviour this probe exists to map),
not a clean target catch. **Category-error fragments did not surface `substantive` this run** — the
standing gap was quiet rather than leaking here. **False-pass-class rate: 1/70** (down from the
per-claim aggregate of 4, which spanned both un-named gaps).

## Calibration results

Populated by `../run.sh axis-gaps`. Summary in `results/`; ledger in
`../../../testing/calibration_log.jsonl`. Status: **`[~]`** — calibrated 2026-06-10/haiku, then
fragment-attributed 2026-06-11/haiku (1 mapped-limit leak / 70, on survivorship; category-error quiet
this run). Stays `[~]` by design: "done" for envelope-mapping means *the limit is logged and
re-checked across models*, never *the critic passed*. Category-error leakage is the standing mapped
limit.
