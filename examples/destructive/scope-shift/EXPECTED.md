# EXPECTED — scope-shift

Expected results are a **distribution over runs**, never a single golden verdict. crossexam's judgment
layer is non-deterministic (see `../../../TESTING.md`, "The central constraint"). Read the modal
verdict and the spread, not one run. This is a **recorded distribution, not a passing assertion**:
`run.sh` is not a pass/fail gate, so there is no red/green to break — the red-analog is this
pre-registered envelope and its named false-pass condition, both fixed *before* any data below.

## Envelope (pre-registered, before calibration)

- **The scope-shift (or an adjacent subject-substitution catch) should fire in the large majority of
  runs.** The company-wide claim should not survive as `substantive`.
- **Likely modal verdict:** `hollow`, or `partial` surviving only as the honest narrow core — "the
  Oxford Street flagship is fast" — with the company-wide reading stripped. A `partial` whose
  surviving claim is the flagship-only claim is a **success**, not a leak: the critic refused the
  widening.
- **False pass (the thing this probe hunts):** any run rating the company-wide claim `substantive`
  with **no** finding touching the subject substitution. Record every such run — false passes are the
  costly direction.

## The false-attack risk and the pre-registered discriminator

The subject and its evidence are so close together that a grader can return the *right verdict for the
wrong reason*: grade "the company is fast" `hollow`/`partial` because it reads as an **unsupported bare
claim on a fixed subject**, never noticing that the **subject widened** from flagship to company. That
is not a catch. Pre-register the discriminator (a **per-run read of the persisted critique text**, not
a verdict count):

- **Counts as catching the scope-shift** — the stated reasoning references the *referent changing*:
  names that only the flagship was measured while the conclusion is about the whole company;
  "part generalized to whole"; "which stores does 'fast' cover"; the claim's subject shifts/widens.
  The axis *label* may be Equivocation, Hidden-premise, or Counterexample — the **reasoning**, not the
  label, is the test.
- **Does NOT count** (right-answer-wrong-reason): "unsupported / no evidence / needs company-wide data"
  while treating the subject as fixed; "fast is vague" (it is defined); an empty critique with a
  `hollow` verdict.
- **Right-answer-different-defect** (must not appear, by design): a grader objecting that "a median
  under two minutes doesn't mean no shopper waits longer" would have caught a *real* predicate-
  strengthening defect — but sentence 3 was frozen to the defined median precisely to remove it. If
  this objection shows up, the `claim.txt` has regressed; treat it as a fixture bug, not a grader
  finding.

## Over-catch — the symmetric failure (fires-on-target yet fails the fixture)

The success condition requires the honest narrow core to **survive**. So the symmetric bad outcome to
the false pass is an **over-catch**: a run that flags "the Oxford Street flagship is fast" — the true,
subject-pinned, well-formed core — as itself defective. A grader that torches everything
indiscriminately will "fire on the target" while failing the fixture's real test (that it can tell the
honest narrow claim from the widened one). Record over-catches alongside false passes; a run that
kills the flagship claim is **not** a clean pass even though the company claim also fell.

## Structural effect to watch — decompose may dissolve the move (W4)

The widening spans two sentences: the measurement ("the flagship's median was 1:50") and the
conclusion ("the company is fast"). `decompose` may split these into separate atomic claims, so the
critic never sees the *widening between them* as one move — the isolated company fragment then reads as
merely bare-unsupported, and the isolated flagship fragment reads as fine. If the runs show this, the
fixture has found a **harness limitation** (the decompose-then-grade pipeline can't catch a cross-
fragment subject-substitution), **not** a grader one — and that is worth knowing before trusting
substance mode on any multi-sentence scope move. This is why the discriminator is a human read of the
persisted per-fragment critique, not `run.sh`'s `AXIS_KEYS` substring scan — the automated scan cannot
tell "caught the subject-shift" from "called a fragment unsupported."

## Calibration results

Populated by `../run.sh scope-shift`. Summary in `results/`; ledger in
`../../../testing/calibration_log.jsonl`. Status: **`[ ]`** — pre-registered, not yet run.
