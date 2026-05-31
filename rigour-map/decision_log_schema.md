# Decision Log Schema

Each row in `decision_log.jsonl` is one labeled decision about whether applying intellectual
rigour helped, hurt, or was moot in a specific **(task, phase)** context. The corpus trains the
rigour-application map: a classifier `(task, phase) → {advantage, disadvantage, irrelevant}`.

Two-stage by design. **Capture** is filled in ~20 seconds *before you act* — a gut call recorded
after the outcome is hindsight, not a prediction the map can be scored against. **Resolve** is
filled later, when the outcome is observable, and assigns the ground-truth label.

## Fields

### Capture (before acting)

| Field      | Type    | Description |
|------------|---------|-------------|
| `id`       | string  | Unique row id, e.g. `"d001"` |
| `date`     | ISO date| `"2026-05-31"` |
| `task`     | string  | The work task, one line, e.g. `"close-enterprise-deal"`, `"draft-adr"`, `"review-contract"` |
| `phase`    | string  | `"generate"` / `"validate"` / `"decide"` / `"execute"` — the work phase, **not** an assay pipeline stage. This is the discriminating variable: rigour flips from asset to liability between `generate` and `validate`. |
| `context`  | string  | Role/domain tag, e.g. `"papercut/ipp"`, `"solo/sales"`, `"monist/fundraise"` |
| `mode`     | string  | If the rigour applied was `assay`: `"substance"` / `"faithfulness"` / `"evidence"` / `"audit"`. `"none"` if rigour was applied without the tool. |
| `gut`      | string  | `"rigour"` / `"skip"` — your pre-committed instinct, recorded before you act |
| `acted`    | string  | `"rigour"` / `"skip"` — what you actually did (may differ from `gut`) |

### Resolve (when the outcome is observable)

| Field        | Type     | Description |
|--------------|----------|-------------|
| `reversible` | bool     | Was the decision cheap to reverse? (a feature, not part of the label) |
| `label`      | string   | `"advantage"` / `"disadvantage"` / `"irrelevant"` — ground truth, assigned from outcome per the definitions below |
| `rationale`  | string   | One or two sentences grounding the label: what happened, why rigour helped/hurt/was moot |
| `notes`      | string?  | Optional caveats, edge cases, follow-ups |

### Phase 1+ (added once the map exists)

| Field      | Type   | Description |
|------------|--------|-------------|
| `map_call` | string | What the map predicted: `"advantage"` / `"disadvantage"` / `"irrelevant"` |

## Label definitions

- **advantage** — applying rigour at this (task, phase) caught a real problem or improved the
  output in a way that would not have happened otherwise.
- **disadvantage** — applying rigour degraded the output, added noise, or consumed effort with no
  offsetting gain (the deal cooled, the idea was strangled, momentum was lost).
- **irrelevant** — rigour ran but the output would have been the same without it; neither helped
  nor hurt.

Assign `label` from the **observed outcome**, not from how it felt at the time.

## Why `gut` is load-bearing

The map has value only where it disagrees with instinct. Recording `gut` before acting is what
makes the two goal metrics computable: the **disagreement win-rate** (where `gut` and `map_call`
differ, how often is the map right against `label`) and the **false-negative gate** (`acted="skip"`
on a task whose `label` turned out `"advantage"` — the silent wrong answer). Drop `gut` and both
metrics become uncomputable; the corpus can describe but not score.

## Append discipline

- Rows are append-only. Never edit or delete a committed row.
- One decision per row. Two separable observations from one session → two rows.
- `gut` and `acted` are committed at capture and never revised once the outcome is known.
- Delete the two example rows before using the corpus for training.
