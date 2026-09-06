# spec/CLI.md — assay command-line interface

> **Scope note — assay-era spec, to be ported to elenchus2.**
> This describes the tool as it stands now: one flat `assay.go` at the repo root, built to the
> binary `assay`. It is written to survive the move to **elenchus2**, the planned binary that will
> carry the `cmd/crossexam` + `internal/` layout from `~/CLAUDE.md`. When that binary exists this
> spec is ported to it, `assay` is renamed `crossexam`, and the behaviour below travels unchanged.
> Until then **`assay` and `crossexam` name the same tool**, and file paths here (`internal/brief`,
> `internal/tree`) are the assay-era homes for code that elenchus2 will inherit.

## Synopsis

```
assay [flags] [INPUT_FILE]
```

`INPUT_FILE` is the prose or claims file to assay; with no file, input is read from stdin (or from
`-text`). Output goes to stdout; progress, the SUMMARY block, and warnings go to stderr.

## Modes

The mode is chosen by flags, one assay per run:

| Invocation                         | Mode                | What it settles                                          |
|------------------------------------|---------------------|---------------------------------------------------------|
| `assay INPUT`                      | Dialectic (default) | Logical substance — producer↔critic per claim           |
| `assay -source SRC INPUT`          | Faithfulness        | Does the summary represent the source transcript         |
| `assay -evidence INPUT`            | Evidence-grounding  | Each claim against external evidence via web search      |
| `assay -source SRC -evidence INPUT`| Grounding on intent | Grounds the intended proposition, not the literal words  |
| `assay -audit -source SRC INPUT`   | Audit               | All three modes in one cross-tab (always markdown)       |

`-audit` requires `-source`. `-source -evidence` grounds `what_source_actually_says` when the
faithfulness pass returns `partial`/`overstated`.

## Environment

| Variable            | Effect                                                              |
|---------------------|--------------------------------------------------------------------|
| `ANTHROPIC_API_KEY` | Required. The run aborts if unset.                                 |
| `ANTHROPIC_MODEL`   | Overrides the default model (`claude-sonnet-4-6`). `-model` wins.  |

## Flags

| Flag           | Default                     | Meaning                                                        |
|----------------|-----------------------------|---------------------------------------------------------------|
| `-model`       | `$ANTHROPIC_MODEL` or `claude-sonnet-4-6` | Model id for every API call.                    |
| `-source`      | `""`                        | Transcript file → faithfulness mode.                          |
| `-evidence`    | `false`                     | Evidence-grounding mode (enables web search).                 |
| `-audit`       | `false`                     | Run all three modes and emit a cross-tab (needs `-source`).   |
| `-text`        | `""`                        | Inline input instead of a file.                               |
| `-md`          | `false`                     | Emit markdown tables (applies to the full table / `audit.md`).|
| `-v`, `-verbose`| `false`                    | Verbose: show every API call.                                 |
| `-no-color`    | `false`                     | Disable ANSI colour.                                          |
| `-max-rounds`  | `2`                         | Producer–critic rounds per claim (substance only).           |
| `-max-claims`  | `0`                         | Cap claims graded, `0` = unlimited (bounds grounding cost).   |
| `-n`           | `1`                         | Repeat each claim N times; show modal verdict + agreement (faithfulness). |
| `-progress`    | `true`                      | Per-case stderr lines + 60s heartbeat.                        |
| `-quiet`       | `false`                     | Suppress per-case lines + heartbeat; keeps SUMMARY and chain. |
| `-usage-out`   | `""`                        | Append one JSON usage record per run to this file.            |
| `-chain-dir`   | `eval/<stamp>-<model>/`     | Directory for the Tier-2 JSONL verification chain.            |
| `-full`        | `false`                     | Print the full table to stdout instead of the brief report.   |
| `-tree`        | off                         | Print the tree report to stdout. `-tree=full` expands all.    |
| `-from`        | `""`                        | Render from a saved chain JSONL; no model calls. Needs INPUT.  |

`<stamp>` is `YYYYMMDD-HHMM` (local). The eval directory is therefore `eval/<stamp>-<model>/`, and
`audit.md` and the chain JSONL both land in it. (This corrects the `eval/<stamp>/` prose in
CLAUDE.md; `assay.go` already stamps the model in.)

## Output: stdout renderers and the eval directory

Three stdout renderers share one set of verdicts:

- **brief** — the default (below).
- **full** — `-full`, the current table renderer (colour, or markdown under `-md`).
- **tree** — `-tree`, a text tree keyed to the source document's headings (below).

Regardless of which renderer stdout gets, **every run always writes the full table to
`eval/<stamp>-<model>/audit.md`**, always writes the chain JSONL to the same directory, and always
writes an HTML tree to `eval/<stamp>-<model>/tree.html` (below). `audit.md` is the markdown rendering
under `-md`, otherwise the no-colour text rendering of the same table. The chain JSONL is unchanged
by any of this.

**Header line.** The brief and tree renderers print one provenance line before their content —
`model <id> · calls <n> · $<cost> (rates <date>) · wall <d>` — so a pasted report carries the model,
the API-call count, the estimated cost, and the wall time it took. Under `-from` every figure is
zero (no model call was made), which is the point.

**Value line.** Directly under the header, `changes N summary lines` — the count of distinct
top-level branches the tree opens (`brief.OpenedBranches`), which `docs/VALUE.md` defines as the
run's value: how many executive-summary lines a reader would change after reading the tree. The USAGE
line beside it carries the cost, so value and cost can be weighed against each other.

`-full` and `-tree` are mutually exclusive; passing both is an error (`-full and -tree select
different stdout renderers; choose one`).

## Prompt caching

The source transcript is identical across every claim in a faithfulness or audit run, so it is sent
in its own content block marked `cache_control: ephemeral`, ahead of the per-claim text, and the
system prompt is cached the same way. The first call of each `(system + source)` pair writes the
cache; the rest read it. The USAGE line reports `cache_read` and `cache_create` token counts, and
`estimateCost` bills reads and writes at their own rates. On the LCEIC 8-claim faithfulness fixture
this cut a run from `$7.68` (`cache_read=0`) to `$1.93`. Substance, decompose, and evidence-grounding
have no reusable corpus, so only their system prompt is cached.

## Replaying a saved chain (`-from`)

`-from CHAIN.jsonl INPUT` reconstructs the verdict rows from a saved Tier-2 chain and renders them
through the same stdout renderers a live run uses (brief / `-tree` / `-full`), **making no model
call** — the header line therefore shows `calls 0 · $0.0000`. It rewrites `audit.md` and `tree.html`
in the chain's own directory.

The INPUT claims file is required: it supplies the ids and heading paths the tree needs (the chain
records carry only claim text and idx), and each record's claim text is checked against it. A
mismatch — a chain paired with the wrong claims file, a differing count, a gap in the idx sequence,
or mixed modes — aborts with a message naming the offending claim. The mode (faithfulness, substance,
grounding, audit) is read from the chain itself.

## Brief report (default)

Brief is the default stdout output for every mode. `-full` prints the current full table instead.
Brief is computed by plain Go over the verdict structs — **no model call touches the brief.**

### Layout

```
<claim count> claims — <verdict counts for the mode(s) run>

Needs you (N):
<id> | <verdict> | <claim text ≤ 60 chars> | <reason ≤ 60 chars>
...                                                              (up to 10 rows)
and K more in <path>                                            (only if N > 10)

Full table: eval/<stamp>-<model>/audit.md
```

- **Line 1** — the claim count and the per-verdict counts for whichever mode(s) ran (e.g.
  `12 claims — faithful 7, overstated 3, contradicted 2` for faithfulness; audit lists all three
  axes).
- **`Needs you (N)`** — `N` is the total number of qualifying claims (the selection rule below),
  not the number shown. Rows are the first 10 after sorting.
- **Row** — four pipe-delimited fields: claim id, verdict, claim text truncated to 60 chars, and a
  reason truncated to 60 chars. The reason is the mode's existing rationale field on the verdict
  struct — faithfulness `faith.Evidence` (the critic finding), grounding `evidence.Finding`,
  substance `substance.Reason` — never a new model call.
- **`and K more in <path>`** — printed only when `N > 10`; `K = N − 10`, `<path>` is the `audit.md`
  path. Omitted when `N ≤ 10`.
- **`Full table:`** — always the last line, naming the `audit.md` path.

If `N = 0` the block is `Needs you (0):` with no rows, then the `Full table:` line. (A clean run is
a real result, not zero-output: the claim count on line 1 is the non-zero denominator.)

### Selection rule

A claim qualifies if it matches any tier. Tiers are the priority order:

- **a.** faithfulness verdict is `contradicted` or `absent`.
- **b.** audit rows where faithfulness is `faithful` **and** grounding is `refuted` (the laundering
  signature: a claim faithfully transcribed from the source yet false about the world).
- **c.** faithfulness verdict is `overstated` **and** the claim text contains a number (`[0-9]`).
- **e.** faithfulness verdict is `partial` **and** the judge's `gap` field is not `none` — a scope,
  denominator, time-range, or attribution mismatch between what the summary implies and what the
  source supports. This is the value tier of `docs/VALUE.md`: a `partial` whose narrowing would
  change a reader's action (the F29 denominator gap) must reach the summary rather than collapse.
  The gap is a structured judge field, so the rule reads a verdict, not prose.
- **d.** grounding verdict is `refuted`.

Sort key: tier (a < b < c < e < d), then claim id, then — to break any remaining tie
deterministically — the claim text. A claim matching more than one tier is counted once, under its
highest (earliest) tier. Tiers that need a signal the run didn't produce (e.g. tier b outside audit
mode, or tier e with no `gap` field) match nothing.

## Tree report

`-tree` renders the claims as a text tree whose shape comes from the **source document's own
headings**, not from the tool. For the LCEIC example the levels are chapter > section > finding >
claim.

### Path column

The claims file gains an optional 6th path column. `3.1/F17` = chapter 3, section 1, finding 17 —
the finding groups the claims decomposed from it. A claim with no path goes under a root node named
`unplaced`. (Assay-era note: `examples/vic-lceic/claims-machine.txt` is pipe-delimited with a
`ref=§<section>` field today; the path column is added from `report-summary.md`'s heading anchors in
the code phase, as data, not code.)

### Node line

One node per line:

```
<indent><id> <label ≤ 12 words> <[child count] | verdict>
```

- **Internal node** — trailing field is `[child count]` (e.g. `[7]`).
- **Leaf** (a claim) — trailing field is the verdict, or the tuple of verdicts for the mode(s) run.
  Under `-n N` (faithfulness) the verdict is the **modal** verdict over N runs, followed by the
  agreement fraction — `partial 2/3` means partial won 2 of 3 runs. A tie breaks to the worst
  verdict (`contradicted` > `absent` > `overstated` > `partial` > `faithful`), so a split surfaces
  the reading a human is likelier to need to check. The fraction rides in the chain (`spread`) and
  appears in the same place in `tree.html`.
- **So-what line** — a Needs-you leaf is followed by one indented line, `↳ so what: <…>`, the
  judge's ≤ 20-word statement of what a reader who believed the report would get wrong (`docs/VALUE.md`).
  It rides in the chain (`so_what`) and shows in the leaf body of `tree.html`. A leaf not in the
  Needs-you set gets no such line.
- **Label** — for an internal node, the document's heading text truncated to 12 words. Where no
  heading exists, and only there, a model is asked with the prompt *"one line, ≤ 12 words, using
  only nouns that appear in the children"*; a model-written label is flagged in the JSONL chain.

### Expansion

Default: an internal node is expanded only if some leaf below it has a verdict in the brief report's
`Needs you` set (the selection rule above); otherwise it is collapsed to its one node line with the
child count. `-tree=full` expands everything.

### Complexity constraints

Enforced by code, and printed at the top of the tree output:

| Constraint | Rule                                                                 |
|------------|----------------------------------------------------------------------|
| width      | max children per node ≤ 7                                             |
| depth      | root to leaf ≤ 4                                                      |
| label      | ≤ 12 words                                                            |
| score      | worst path (depth × mean fan-out on that path), target ≤ 20          |

- **Width violation** — insert a grouping level: split the children into runs of ≤ 7 in id order,
  each run labelled by its id span, e.g. `F1–F7`, `F8–F14`. This adds a level, so it is applied
  before the depth check.
- **Depth violation** (after any width grouping) — **fail**, naming the offending root-to-leaf path.

`score` is reported as a target, not a hard failure: it flags a tree that is legal on width and
depth but still too tangled to read, for a human to judge.

### HTML tree (`tree.html`)

Every run (and every `-from`) also writes the same tree to `eval/<stamp>-<model>/tree.html`, a
self-contained page built from nested `<details>`/`<summary>` elements — **no JavaScript**. Every
node is open by default; collapsing is the browser's native `<details>` toggle. A leaf's summary
carries its verdict; its body reveals the full claim text, the verdict reason, and the verbatim
source spans the defender cited, all taken from the chain. Styling is monospace only. The depth
ceiling is not re-enforced here — the text renderer is the gate — but width regrouping matches.

## Notes for review (decisions taken in this draft)

These are choices the two source prompts left open; flagged so they can be confirmed or changed
before code:

1. **`audit.md` when not `-md`** — holds the no-colour text rendering of the full table (a readable
   `.md`), and the markdown rendering under `-md`.
2. **`-full` + `-tree` together** — rejected as an error rather than silently picking one.
3. **`-tree` flag shape** — bare `-tree` (default expansion) and `-tree=full` (expand all) require a
   non-bool flag (a `flag.Value`, not `BoolVar`); off is the zero value.
4. **brief `reason` source** — the existing per-mode rationale field, truncated; no model call.
   Field names above are assay-era and confirmed against `assay.go` at code time.
5. **`-md` and the tree** — `-md` governs the full table / `audit.md` only; the tree is text-only.
6. **eval dir** — `eval/<stamp>-<model>/`, matching `assay.go`, not CLAUDE.md's `eval/<stamp>/`.
