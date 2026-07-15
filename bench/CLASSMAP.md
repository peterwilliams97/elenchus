# CLASSMAP.md — defect class ↔ critic axis, v1

Version: 1 (2026-07-15). Pre-registered before any scored run, per MUTATION_BENCH.md §5.
Scope: v1 classes only (MFC, SCP-1). Deferred classes get their rows when their modes exist.

## What this table is for, and what it is not for

This map feeds **one thing**: the axis-fired diagnostic. Nothing else.

It does **not** decide catches. Under A1.2 and A3.4 a catch is decided by a human reading the chain
text against the manifest row, and no table can stand in for that. §5 originally gave this map a
second job — the deterministic match rule — and Amendment 1 took that job away when it replaced the
deterministic scoring path with target-catch.

So the map is a convenience for reporting which axis the target class *would* fire, not an oracle.

## The table

| Class | Target axis | Rationale |
|---|---|---|
| MFC | `Falsifiability` | The injection deletes the kill condition. The axis's own prompt text is "what observation would show it false? If none exists, it is vacuous" (`internal/modes/prompts.go`). The defect *is* the axis's subject. |
| SCP-1 | `Equivocation` | The injection widens the subject of predication while freezing the predicate. The axis's prompt text is "does a key term shift meaning or hide behind a buzzword?" |

## The map is strict, and that is the point

**Target axis only. Never permissive.** A permissive map — accepting `Evidence` or
`Hidden premise` as an SCP-1 catch — would score the 2026-07-12 run as a success. It was not one.
Every objection that run raised against a company-wide claim was unsupportedness, bare assertion, or
condition-laundering: the right answer for the wrong reason, treating "the company" as a fixed
subject with an evidence gap rather than seeing the subject move. The human read scored the target
catch 0/5 while a substring scan certified 5/5.

Widening this table is therefore not a tuning knob. It is the mechanism by which a benchmark
launders a near-miss into a hit.

## Known limit, recorded not smoothed

`Axis.Axis` is free text from the model (`internal/claims/claims.go`); nothing validates it against
the seven axis names the prompt lists. So a spelling drift silently fails to match this table. That
is logged as a consolidation item (PHASE1-PLAN.md §9.3 task 3, `AxisName` typed vocabulary), and
until it lands the axis-fired diagnostic under-counts rather than over-counts — the safe direction
for a number that must never be read as recall.
