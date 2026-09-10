# REPORT — the rigour-map problem statement (2026-09-08)

Read-only investigation. No edits outside this file, no commits, no runs. Every count below carries
the command that produced it.

## Base rate first

This is one track that is under-defined, not a repo whose disciplines are broken. The append
disciplines mostly hold: 31/31 real rows carry an `outcome`, dead ends are logged at a plausible
rate (4/31 ≈ 13%), and the 2026-09-07 session was tightly human-directed, not adrift. The failure is
narrow and specific: the rigour-map *objective* is stated two incompatible ways, has no
done-condition, and the recent work was logged against an objective name that no document defines.

## Step 1 — what exists

`for f in ...; do [ -e "$f" ] && echo EXISTS ...; done`

| File | State |
|---|---|
| `CLAUDE.md`, `README.md`, `BACKGROUND.md`, `TESTING.md`, `SESSION.md` | EXISTS |
| `rigour-map/decision_log.jsonl` | EXISTS (33 lines) |
| `rigour-map/decision_log_schema.md` | EXISTS |
| `rigour-map/run_eval.sh` | EXISTS |
| `REPORT.md` | did not exist before this file |
| **`rigour-map/RUBRIC.md`** | EXISTS — **not named in the prompt's list**; it defines a *second, different* objective (see Step 2) |
| **`fixtures/` (and `fixtures/raw/`, `fixtures/summaries/`)** | **MISSING** — `find . -type d -name fixtures` returns nothing; RUBRIC.md and run_eval.sh both require it |

## Step 2 — the objective, as the repo states it (two incompatible versions)

There is no single objective. Two documents define a rigour-map classifier with the *same* output
taxonomy `{advantage, disadvantage, irrelevant}` but a *different* unit and a *different* corpus:

**Version A — the decision-log classifier.**
- `CLAUDE.md:32-33`: "**Rigour-application map**: A parallel track (no Go code yet) that classifies
  `(task, phase)` → `{advantage, disadvantage, irrelevant}` from a labeled decision log."
- `rigour-map/decision_log_schema.md:4-5`: "The corpus trains the rigour-application map: a
  classifier `(task, phase) → {advantage, disadvantage, irrelevant}`."
- Corpus: `rigour-map/decision_log.jsonl`. Phase 0 = data collection (`CLAUDE.md:34`).

**Version B — the fixture rubric.**
- `rigour-map/RUBRIC.md:3-5`: "Classifies the dialectical-rigour methodology per
  **(business_area × mode)** → `{advantage | disadvantage | irrelevant}` … the unit of
  classification is the (area, mode) cell."
- Corpus: "the seven verified fixtures in `fixtures/raw/`" (`RUBRIC.md:7`) — **which do not exist**.

The two disagree on the classification unit — `(task, phase)` vs `(business_area × mode)` — and draw
on different corpora. Any statement of "the rigour-map objective" has to pick one; the repo has not.

**Done-condition: none is stated with a target.** The nearest thing to a success test:
- `decision_log_schema.md:61-63` names two metrics — the **disagreement win-rate** ("where `gut`
  and `map_call` differ, how often is the map right against `label`") and the **false-negative
  gate** (`acted="skip"` on a task whose `label` turned out `"advantage"`). Both require `map_call`,
  a `### Phase 1+` field "added once the map exists" (`schema:41`). **0 rows carry `map_call`**
  (`grep -c map_call` → 0), so neither metric is computable today.
- `RUBRIC.md:6, 51-58` gives a procedure ending in "Diff against the SEALED PRIORS … the cells where
  the run overturns the prior are the evaluation's real output." That is a procedure, not a pass
  mark, and it needs the missing `fixtures/`.

Finding: **no passage defines a done-condition that could come back negative.** There is a taxonomy,
two metrics, and a sealed-prior diff, but no target count, accuracy floor, or completion test.

## Step 3 — the open decision and the corpus

**The open decision** (`CLAUDE.md:34-35`): "seed the classifier now vs. log unaided first to
protect the disagreement baseline — not yet resolved."
- *Seed now* commits to building the map from the 31 rows and starts scoring `map_call` against
  `label`; it forecloses a clean disagreement baseline, because once you have seen the map's call
  you can no longer record an independent `gut`.
- *Log unaided first* keeps `gut` honest (recorded before any map exists) so the disagreement
  win-rate stays meaningful; it forecloses any map output until enough unaided rows accumulate, and
  it has no stated stopping rule for "enough."

**Corpus state** (`python3` over `rigour-map/decision_log.jsonl`; commands in the session log):
- Lines: 33. **Example rows still present:** `ex001`, `ex002`, both `notes` = "DELETE — example row
  only". The schema's last line (`append discipline`) says "Delete the two example rows before using
  the corpus for training." Not done.
- **Real entries: 31** (`d001`–`d031`). Date span **2026-05-01 → 2026-09-07** (`min/max` of
  `date`/`started_at`; none missing).
- **`outcome` filled: 31/31.** None left open.
- **Dead ends: 4** — `abandoned`: `d001`; `reverted`: `d003`, `d004`; `superseded`: `d006`. A
  non-zero rate (≈13%), so the W8 survivorship alarm (`BACKGROUND.md:108,129`) does not fire.
- **`goal_link`:** `grounding-integrity` 17 · `maintenance` 11 · `phase-0-corpus-collection` 3 ·
  `detour` 0.
- **Schema conformance — four deviations:**
  - `label` must be `{advantage, disadvantage, irrelevant}` (`schema:37, 49-54`). `d029`, `d030`,
    `d031` set `label="maintenance"`; `d008` has **no `label`**.
  - `gut` must be `{rigour, skip}` (`schema` Capture table). 3 rows (`d029`–`d031`) set
    `gut="advantage"` — the field was repurposed to hold a predicted label.
  - Two example rows undeleted (above).
- **The load-bearing finding:** `grounding-integrity` is the goal_link on **17 of 31 rows** but is
  **defined nowhere as a rigour-map objective**. `grep -rniE 'grounding-integrity'` finds it only at
  `BACKGROUND.md:232` ("d009/d010 grounding-integrity line" — a description of a code path) and as a
  goal_link value. It names the faithfulness-judge hardening line, which is tool development, not the
  classifier. So the majority of recent rows are goal-linked to a non-rigour-map objective wearing a
  goal_link's clothes.

## Step 4 — what the 2026-09-07 session did: specified, but not at this target

Transcript: `~/.claude/projects/-Users-peterw-code-personal-elenchus/34d788ee-af64-43b6-ae45-e7a48e3edabd.jsonl`
(5.8 MB). Extracted user-role text blocks (excluding `tool_result`, command wrappers, and system
notices): **21 blocks, of which 3 are harness/skill injections** (a skill preamble at 01:39, a
Chrome-extension system notice and an "[Request interrupted]" marker at 04:27), leaving **18 genuine
human turns.**

Classification — (a) redirect toward a goal vs (b) responding to something the assistant raised:
**≈14 (a) : ≈4 (b)** (`continue` 04:12; "is anything sill running?" 06:58; a bare file path 07:44;
plus the yes/no on the Chrome extension). The session was **densely specified**, not underspecified.

Five (a) redirect turns, verbatim openings, in order (timestamps from the transcript):
- `00:44:57` "Three steps, stop after each, ≤5 lines. Retrieval must pass its refuter before
  anything is compared again. 1. Fix retrieval. Target: all 18 Sonnet-quoted spans…"
- `01:58:19` "Go. Step 3, with Sonnet/full rerun on the new judge rather than reused from
  eval/n3 (~$4). Budget $6 for the run…"
- `03:36:37` "Rule, in code not prompt: contradicted requires ≥1 verified quote; with none it is
  downgraded to absent and the downgrade is counted in USAGE…"
- `04:37:17` "Prompt 1 — submissions rerun. Add sources/submissions/*.txt to the corpus with
  source=submission on each passage…"
- `05:51:54` "1. Corpus manifest: sources/MANIFEST.md listing every document held… 2. Claims file
  gains a cites column from report-summary.md's footnotes…"

Every redirect names retrieval, the faithfulness judge, the submissions corpus, or the LCEIC report
tree. **None names the rigour-map, the decision log, the classifier, Phase 0, or `map_call`.**

Git log, same window (`git log --since=2026-09-07 --until=2026-09-08 --oneline` → 20 commits).
Applying CLAUDE.md's own goal-link rule, every one serves either (i) the LCEIC faithfulness demo
(`vic-lceic: …`, an *application* of the tool) or (ii) faithfulness-judge hardening (`judge: …`, the
"grounding-integrity" line), plus two CLAUDE.md pins. **Zero touch the rigour-map classifier.** By
the rule they are all `detour`/`maintenance` relative to the rigour-map objective.

**Answer to the step's question:** neither "underspecified" nor "specified and ignored." The work
was specified in detail and executed faithfully — against a *different, coherent* objective
(tool hardening + a real-report demo). The goal-link rule did not bite because it had somewhere to
land: the rows were tagged `goal_link="grounding-integrity"`, which *looks* like a rigour-map goal
and is not one. The rule needs a closed list of valid rigour-map objectives to check against; it has
an open string field, so a detour can be logged as a goal and no gate notices.

## Step 5 — the problem statement

**Objective (one sentence, once the human picks a version):** build a classifier that, given a
work context, predicts whether applying the dialectical-rigour method will be an advantage,
a disadvantage, or irrelevant — where "work context" is *either* `(task, phase)` from the decision
log (Version A) *or* `(business_area × mode)` from the fixtures (Version B); the repo has not chosen.

**Done-condition (stated so it can fail):** on a held-out set of contexts, the map's `map_call`
beats the recorded `gut` on the disagreement subset (where `gut ≠ map_call`) by a margin the human
fixes in advance — e.g. map right on ≥ two-thirds of disagreements — with zero silent false
negatives (`acted="skip"` where `label="advantage"`) on that set. This can return negative: today it
is not even computable, because `map_call` exists on 0 rows.

**Explicitly out of scope:** the LCEIC report demo (`examples/vic-lceic/`); the faithfulness-judge
"grounding-integrity" hardening line; any Go code for the map (CLAUDE.md: "no Go code yet"); the
destructive-testing program (`TESTING.md` Layer 3). These are the repo's real recent work and are
legitimate — but they are detours *with respect to this objective* and should be tagged as such.

**Next single decidable step:** resolve the open decision at `CLAUDE.md:34-35` — it gates the
done-condition and it is the human's call, so this report stops at presenting both branches:

| Branch | Commits the project to | Forecloses | Precondition it exposes |
|---|---|---|---|
| **Seed the classifier now** | derive `map_call` from the 31 rows, start scoring against `label` | a clean disagreement baseline — a `gut` recorded after seeing a map is hindsight | must first pick Version A or B; Version B additionally needs the missing `fixtures/` restored |
| **Log unaided first** | keep recording `gut`-before-`map` rows until "enough" | any map output until then | needs a stated stopping rule ("enough" is undefined) and a closed goal_link vocabulary so new rows stop laundering detours as goals |

Both branches share two unblock-able prerequisites the human can decide independently of the seed
question: (1) **choose Version A or Version B** as *the* objective, or state they are two tracks;
(2) **close the `goal_link` vocabulary** to a fixed set so `grounding-integrity` is either promoted
to a named objective or demoted to `detour`.

## FILED / NOT-FILED

- All counts in Steps 1, 3, 4: FILED — commands are in this session's tool log; re-runnable against
  `rigour-map/decision_log.jsonl` and the named transcript at HEAD `aad89f9`.
- Transcript quotes (Step 4): FILED — verbatim openings, truncated at capture; timestamps from the
  transcript records.
