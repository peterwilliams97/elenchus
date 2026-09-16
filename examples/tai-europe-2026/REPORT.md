# REPORT — Slice 3 segmenter seam (plan-segmentation.md item 7)

Implemented the Segmenter seam minimally so Slice-3's cite-scoped `route=evidence` grounding can run.
No model call, no commit.

## What shipped

- **Seam** (`internal/retrieve/segment.go`): `Segmenter` interface; `HansardSegmenter` wraps
  `splitData` (the split-out body of `splitFile`, unchanged); `PlainSegmenter` splits on blank lines
  into paragraph passages, merging each fragment under 200 runes into the next paragraph. Selection is
  by the `/cited/` path substring in `passagesForFile`; passage id base `cited/<stem>`, `Source=cited`.
- **`splitFile`** refactored to read the file then call `splitData(path, data)` — the Hansard segmenter
  hands bytes without a second `os.ReadFile`. `Format` gains a `cited web page` branch.
- **`citedExternalBases`** (assay.go): `cited/<stem>.txt` → base `cited/<stem>` already resolved via
  the generic `.txt` rule (shared with `leaderboards/`); only the doc comment changed.
- **MANIFEST** (`sources/MANIFEST.md` § Held cited ids): the 16 usable captures registered as backticked
  `- ` bullets so `LoadHeld` sees them. `08b`/`09` (wall) and the two `uncited/` loci stay unregistered.
- **Spec**: `spec/TREE.md` § The segmenter seam.

## Refuters (all green under `go test ./...`)

| refuter | where | result |
|---|---|---|
| (a) golden — existing corpora byte-identical | `retrieve.TestSegmenterGoldenUnchanged` | 8 corpus roots hashed field-for-field vs pre-seam capture; all match. |
| (b) 16 held captures ≥3 passages, base+source | `retrieve.TestPlainSegmenterCited` | 16/16 ≥3 (min 4: `06-wemustactnow`); base `cited/<stem>`, `Source=cited`. |
| (c) Slice-3 dry run, no model | `TestSlice3CiteScopedClassification` | 20 leaves → 16 eligible + 4 unverifiable + 0 floored. |

The 4 unverifiable: `c08b-bloomberg`, `c09-theinfo` (walled, unregistered), `c05-jagged`,
`c19-privateinv` (cite an `uncited/…` id no held document answers to).

## Deviation from the prompt's refuter (b) — flag to PW

Refuter (b) asked that "the MANIFEST-recorded figure sentence lands in exactly one passage" for **each**
of the 16. It is asserted for **2** (`07-pacingthefrontier`, `06-wemustactnow` — both `fig=present`
with a verbatim anchor quote), not all 16, because the MANIFEST itself records most captures as
`figure-only` (datum only in an interactive chart) or `absent-in-extraction` — there is no verbatim
figure sentence in the static text to land. That is the corpus's property (MANIFEST § Slice-3 cited
sources), not the segmenter's. Asserting a "figure lands once" check on a capture whose figure is not in
the text would be a fabricated pass. The ≥3-passage check does cover all 16.

## Not done / open for PW

- `single_source: true` in the MANIFEST drives the judge prompt toward the single-source rules
  ("other parts of the same document"). Slice-3 grounds against **external** cited pages, not the
  report, so that flag's rules do not fit the cite-scoped route. Item 7 (the seam) does not touch the
  judge prompt; decide whether the Slice-3 run needs single_source off before trusting verdicts.

## N=3 Sonnet judge command for Slice 3

```sh
source ./setkey.sh && ./assay -backend anthropic -model claude-sonnet-4-6 -retrieve bm25 -n 3 \
  -manifest examples/tai-europe-2026/sources/MANIFEST.md \
  -source examples/tai-europe-2026/sources/cited \
  -chain-dir examples/tai-europe-2026/evidence/2026-09-16-slice3/sonnet \
  examples/tai-europe-2026/claims-slice3.txt
```

Expect 16 judged leaves (cite-scoped against their `cited/<stem>` capture) and 4 `unverifiable` with no
model call. Flags precede the positional claims file (Go's parser stops at the first non-flag).
