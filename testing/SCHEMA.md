# calibration_log.jsonl — record schema

The Layer-3 calibration ledger. One JSON object per line, append-only. Lines are **never
rewritten** — corrections are added as new lines that reference the originals (see *annotation*
below). Multiple record *types* share the file, distinguished by their fields.

The companion per-fragment evidence lives under `testing/chains/<date>-<model>/<probe>/run-N/`
(persisted Tier-2 chains, gitignored — raw crossexam output). The ledger summarises; the chains are the
fragment-level record the summaries are derived from.

## Why the schema grew (2026-06-11)

The original schema logged only **run-level** verdict counts. But the headline metric — the
**false-pass rate** — is defined **per defect-carrying fragment**: each one-defect probe is split by
`decompose` into several atomic fragments, most of them *clean scaffolding* (e.g. the honest motte,
or a real figure), and a `substantive` verdict on a clean fragment is the critic being **right**, not
a leak. A false pass is specifically a **defect-carrying fragment rated `substantive` with no
target-axis (or legitimate-adjacent) finding**. That distinction is invisible at the run-level
aggregate, so the aggregate `substantive` count systematically *overstates* false passes.

The 2026-06-10 runs could not be re-scored at the fragment level because their chains were written to
a `$TMP` dir that `run.sh` deleted on exit (instrument error, `decision_log` d016 — now fixed by
persisting chains). So the fragment-attributed metrics start from a **fresh 2026-06-11 calibration**,
not a rescore of the 2026-06-10 sample (which no longer exists). See the *annotation* record type.

## Record types

### 1. `distribution` (the original format; emitted by `run.sh`)

Substance/grounding probe:
```json
{"date","model","fixture","mode","runs","verdict_counts":{...},"axis_mentions":{...}}
```
Audit probe (laundering): `faithful_counts` / `substantive_counts` / `grounded_counts` instead of
`verdict_counts`. `verdict_counts` are summed **across all fragments of all runs** — useful as a
distribution snapshot, but NOT a false-pass count.

### 2. `fragment_attribution` (new 2026-06-11; the false-pass metric)

```json
{"date","model","fixture","mode",
 "analysis":"fragment_attribution","fragment_attributed":true,
 "runs",
 "claims_per_run":[ ... ],          // fragments decompose produced, per run
 "verdicts_total","errors","n_valid", // n_valid = verdicts_total - errors
 "substantive_total",                // substantive (+ any non-enum near-miss like "substantial")
 "substantive_on_clean_scaffolding", // critic correctly affirmed an honest sub-claim = NOT a leak
 "substantive_on_defect_axis_fired", // defect fragment passed substantive but the axis still fired = named, not clean leak
 "false_pass",                       // defect fragment, substantive, NO target/adjacent finding (often empty critique)
 "target_axis","note"}
```
**Denominator rule:** `error` verdicts are instrument failures, **excluded** from `n_valid` and from
every envelope rate; they are reported in their own `errors` field. **False-pass rate** =
`false_pass / n_valid`.

### 3. `annotation` (supersede a prior line without rewriting it)

```json
{"date","annotation_of":{"date","fixture","model","mode"},
 "fragment_attribution":"unrecoverable","superseded_for":"false_pass_metrics",
 "superseded_by":{"date","analysis","fixture"},"reason"}
```
Used for the 2026-06-10 lines: retained untouched as the first distribution snapshot, but marked
unattributable for false-pass purposes (chains destroyed). The original line is the record of the
instrument error; the annotation points to the 2026-06-11 fragment-attributed line that supersedes it
for metrics.

### 4. `reflexive-canary` (TESTING.md 3d; always-on)

```json
{"date","model","fixture":"reflexive-canary","mode":"grounding","runs",
 "verdict_counts":{...},"canary":"headline","required":"unverifiable every run","held":true|false}
```
`./crossexam -evidence` on the repo's own headline. `unverifiable` every run = the boundary held. Any
`supported` = a self-sealing failure inside the instrument (grounding confirming the tool's own value
proposition from the armchair) — investigate before trusting any grounding verdict.

## Reading rule

False-pass and false-attack metrics are read from `fragment_attribution` (and the chains), never from
a `distribution` line's raw `substantive` count. A `distribution` line is a spread snapshot; it does
**not** license a false-pass claim on its own.
