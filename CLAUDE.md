# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`assay` is a **report → argument-tree checker**: it takes a long report (the Victorian LCEIC report
is the worked corpus) decomposed into atomic claims tagged with their §-heading trail, retrieves the
source passages each claim is about (BM25 + local embeddings), runs a schema-enforced faithfulness
judge per claim against only those passages, and lays the verdicts out as two trees —

- the **report tree**: findings grouped by §-heading, one faithfulness verdict per leaf (`spec/TREE.md`);
- the **argument tree**: the same leaves arranged by inferential structure — claim → finding →
  recommendation → root thesis — with every internal node's judgement DERIVED bottom-up from its
  children, never authored (`spec/ARGUMENT.md`).

`-review` then builds a self-contained, PDF-linked two-pane site and `assay serve` serves it over
HTTP so the page's `#page=N` links land on the right page (`spec/SERVE.md`). The specs are the
contract; code follows them. Build with `./build.sh` or `go build`.

The tool's **older single-file modes** — dialectic (default), faithfulness (`-source`), evidence
(`-evidence`), audit (`-audit`), over a single prose file — still ship and are specified in
`spec/CLI.md`; `assay.py` (the retired Python original) is archived at git SHA `469ebe4`.

**Docs roster:** `spec/` (`CLI.md`, `TREE.md`, `ARGUMENT.md`, `SERVE.md`) holds the authoritative
specs that the code follows; `README.md` is the user-facing abstract; `BACKGROUND.md` is the design-rationale +
failure-envelope / destructive-self-criticism doc (where verdicts can't be trusted, plus the
destructive-test specs); `TESTING.md` is the testing program — the four-layer test taxonomy, what
each layer's results are allowed to mean, and the live destructive-test status board; `SESSION.md` is
the parking lot of deferred work; `rigour-map/decision_log.jsonl` is the decision/change log.

### Tracks

**Rigour-application map**: A parallel track (no Go code yet) that classifies
`(task, phase)` → `{advantage, disadvantage, irrelevant}` from a labeled decision log. Phase 0 is
data collection; `rigour-map/decision_log.jsonl` is the corpus. Open decision: seed the classifier
now vs. log unaided first to protect the disagreement baseline — not yet resolved.

## Running it

```sh
./build.sh                                             # test + build → ./assay
```

**Judge run** — the one pass that makes model calls: retrieved, N=3, against the manifest, writing a
Tier-2 chain plus `tree.html` and `audit.md` under `-chain-dir` (`spec/TREE.md`):

```sh
source ./setkey.sh && ./assay -backend anthropic -model claude-sonnet-4-6 -retrieve bm25 -n 3 \
  -manifest examples/vic-lceic/sources/MANIFEST.md \
  -source examples/vic-lceic/sources/hearings,examples/vic-lceic/sources/submissions,examples/vic-lceic/sources/qon \
  -chain-dir evidence/<date>-full/sonnet-retrieved \
  -tree examples/vic-lceic/claims-machine-full.txt
```

**Backfill passage ids** into a saved chain — no model call; adds each quote's `quote_passages` by
verbatim-matching it against that record's own passages, unresolved on zero or multiple matches:

```sh
./assay -backfill-passages <chain>.jsonl \
  -source ../sources/hearings,../sources/submissions,../sources/qon
```

**Render from a saved chain** — no model calls. Flags go **before** the positional claims file
(Go's parser stops at the first non-flag). `-from` takes a comma-list of chains and merges them
leaf-by-leaf; `-argument` + `-manifest` + `-refs` render the argument tree with per-leaf provenance;
`-review` also builds the self-contained PDF-linked site under the example dir (`spec/SERVE.md`):

```sh
# run from examples/vic-lceic/current/
./assay -from current.faithfulness.jsonl,current-2026-09-10.faithfulness.jsonl \
  -argument ../argument.txt -manifest ../sources/MANIFEST.md -refs ../claims-machine.txt \
  -review \
  claims.txt
```

**Serve the built site** over HTTP — loopback only, opens the browser unless `-no-open` (`spec/SERVE.md`):

```sh
./assay serve examples/vic-lceic/site        # → http://127.0.0.1:8080/review.html
```

**`ANTHROPIC_API_KEY` lives in `./setkey.sh`.** Runs that need it use `source ./setkey.sh &&
<command>` in one Bash call. Never cat, echo, grep or read setkey.sh; never print
`$ANTHROPIC_API_KEY`; never write the key value into a command, file, or message. If a run fails for
auth, report that and stop.

`ANTHROPIC_MODEL` env var overrides the default model (`claude-sonnet-4-6`).

**Flags beyond those shown:** `-retrieve bm25|none|oracle` (per-claim retrieval; `none` sends the
whole corpus, `oracle` reads `-oracle`); `-n N` (repeat each claim N times → modal verdict + `k/N`
agreement, the stability class); `-fresh` (re-judge every claim rather than resume an existing
chain); `-tree=full` / `-full` (expand every node / print the full table); `-make-manifest <root>`
(scan a sources root, write its `MANIFEST.md`, exit); `-floor` (top-passage cosine below which a
claim is `absent` with no model call); `-embed` / `-embed-model` (the semantic ranker fused with
BM25); `-quiet`, `-progress`, `-usage-out FILE`, `-v`/`-verbose`, `-no-color`. The single-file modes
add `-max-rounds N`, `-max-claims N`, `-evidence`, `-audit`, `-md` (`spec/CLI.md`).

## Architecture

The report → argument-tree path lives in `internal/`; `assay.go` is the CLI and the model/judge
path. `assay.go` is large (~3,800 lines) and holds the parts that touch the network or `main`.
**New code goes in `internal/`, not `assay.go`** — a package there per concern, each naming its spec
in its doc comment.

### `internal/` — one package per concern

- **`backend`** — the single `Complete` seam every LLM provider implements (subpackages `anthropic`,
  `ollama`, and the test `fake`), so a mode runner calls one method and never names a provider. Holds
  the request/response shapes, `MaxTokens`, and the web-search / JSON-schema fields. Contract: `spec/CLI.md` §Backends.
- **`brief`** — renders the default stdout "Needs you" report from the verdict rows, deterministically
  and with no model call. `Qualify` is shared with `tree`, so the brief and the tree agree on which
  claims a human must look at.
- **`embed`** — the local Ollama embedding client (`nomic-embed-text`, L2-normalised vectors) that
  supplies the semantic half of faithfulness retrieval, kept separate from the judge backend.
- **`manifest`** — the held-document set derived from the `sources/` filesystem and keyed by canonical
  id, so a run can tell `absent` (checked, not there) from `unverifiable` (the truth-maker is missing).
- **`retrieve`** — turns the corpus into role-tagged speaker-turn passages and ranks them against a
  claim with BM25, so the judge sees only the passages a claim is about, not the whole corpus.
- **`tree`** — renders the verdict rows three ways: the **report tree** by §-heading path (`tree.go`,
  `spec/TREE.md`), the **argument tree** with judgements derived bottom-up (`argument.go`,
  `spec/ARGUMENT.md`), and the **root summary block** above both (`root.go`).

### `assay.go` — the CLI and the judge path

`main` parses flags and dispatches; the `serve` subcommand serves a built `-review` site (`spec/SERVE.md`).
The judge path is here because it makes model calls: per claim it retrieves passages, then `faithJudge`
makes one schema-enforced call (verdict + evidence quotes) via `callSchema` over the chosen `backend`;
`quoteInPassage` / `groundVerdict` verify every cited quote against its passage with **no model**, and
`faithJudgeRepeat` samples `-n` times for a stability class. Runs write a Tier-2 JSONL chain that
`-from` re-renders into trees with no further model calls. Decomposition, the older single-file mode
runners, and usage/tally accounting also live here. Its size is the reason a new concern starts as an
`internal/` package instead.

## Hard rule: never fabricate inputs, and propagate provenance to conclusions
This is a claim-validation tool. Its credibility is its substrate. Fabricated inputs don't just
weaken a result; they invert the tool's entire purpose.

- NEVER synthesize a fixture, test input, sample document, dataset, or "example" of a real artifact.
   If a real one is required and cannot be fetched or obtained, STOP and report
  "could not obtain real <X>" for that item. Do not substitute a fabricated stand-in.
- Fetching real public documents into a gitignored folder is correct — that is analysis, not
  redistribution. "Do not COMMIT copyrighted docs" never means "do not USE real docs."
- A populated results table is NOT success. Any eval, calibration, or verdict is only as valid as
  its inputs. Before reporting a finding, signal, or recommendation, confirm every input is real and
  state its provenance (source + a verifiable snippet).
- Propagate provenance: if ANY input is synthetic, unverified, or fictional, mark every downstream
  conclusion INVALID / inconclusive. Never present it as a finding.
- "Looks like the real thing" ≠ "is the real thing." Optimize for the substrate, not the artifact.

## Testing

Run `go test ./...` after every code change — including fixture edits, prompt tweaks, and
test-file changes. `./build.sh` is the gate check (runs `go test ./...` then builds), but
`go test ./...` alone is faster during iteration. Never report a change as done without a
passing run.

`./build.sh` runs `go test ./...` then builds the binary.

Test coverage in `assay_test.go`:
- **Pure functions:** `splitSummary` (newlines, run-together, bullets, blank lines, single-line),
  `extractJSON` (fences, embedded braces, preamble), `unmarshalLoose` (trailing commas, new
  condition-laundering fields), `mdCell`, `tally`.
- **Prompt presence:** `TestFaithCriticSysLiteralization`, `TestSubstanceCriticSysConditionDiscipline`
  — assert the calibration-fix instructions are in the prompts.
- **Integration via stub:** the three key behaviour tests use `cfg.call` (the injection seam) to
  return canned JSON without network calls, then drive the real method:
  - `TestConditionLaunderingDowngrade` — `c.assayClaim` downgrades partial→hollow when
    `survives_only_by_conditioning=true`.
  - `TestConditionLaunderingLoopStop` — loop exits after one critic call (not two) when
    `survives_only_by_conditioning=true && needs_another_round=true`.
  - `TestRunEvidencePropositionSubstitution` — `c.runEvidence` grounds `what_source_actually_says`
    (not the literal claim) when faithfulness returns overstated.

For end-to-end validation use `examples/dan_shipper/` with `-model claude-haiku-4-5-20251001` for
speed, then re-run on the default model for the verdict to trust.

These `go test` tests are **confirmatory** (nominal input → intended behaviour fires) and cover
Layers 1–2 (plumbing + orchestration) of the four-layer program in `TESTING.md`. The complementary
**destructive-testing program** (Layer 3 — adversarial inputs that push each mode past its limit to
map the failure envelope) is specified in BACKGROUND.md §2.2 and tracked layer-by-layer in
`TESTING.md`. It is now **partly built**: the §3b adversarial axis probes ship as seven public worked
examples under `examples/destructive/` (motte-and-bailey, reference-class, hidden-premise,
unfalsifiable-dress, causal-narrative, axis-gaps, laundering, each with `claim.txt` / `DEFECT.md` /
`EXPECTED.md` / `results/` + a top-level `run.sh` and reader README), first-calibrated 2026-06-10 on
`claude-haiku-4-5-20251001` at N=10 (logged to `testing/calibration_log.jsonl`). Layer 3 is
*calibration, never a CI gate* — its results are read by a human and mean "the envelope held on this
set, this time," never "the critic is correct." The constructive gold sets (§3a), cross-model probes
(§3c), and several Layer 1–2 lockdowns remain unbuilt — see `SESSION.md` for the prioritized queue.
Atomization asymmetry to keep in mind when testing: substance uses the LLM `decompose`, while
faithfulness/evidence/audit use the regex `splitSummary` — different failure surfaces (BACKGROUND.md
W4), so test the one your change actually touches.

## The axis boundary (durable design note)

The three columns are not equally reachable from any one competence:
- Close reading reaches FAITHFULNESS.
- Dialectic reaches SUBSTANCE.
- Reasoning can REFUTE a grounding claim (an internal contradiction kills it with no lookup) but can
  NEVER CONFIRM one. Positive grounding always needs the truth-maker — a retrieval, not a deduction,
  however rigorous.

The characteristic failure is laundering confidence: scoring real wins on faithfulness and
substance, then pronouncing on grounding with borrowed authority the first two columns never
licensed. Fluency on the reachable columns manufactures unearned confidence on the unreachable
one. The discipline is not "apply more rigour" — it is knowing the boundary of what your rigour
can settle and going to the truth-maker past that line. Two of assay's three columns cannot reach
the thing the third column is for. See examples/url-length for a worked demonstration.

`crossCheckEvidence` operationalises this boundary but does not abolish it: it proves a cited URL was
actually retrieved (provenance), never that the page supports the sentence (content). So `supported`
stays the structurally weakest verdict — read it most skeptically, alongside the faithfulness-pass
intent reconstructions (`intendedProposition`) that silently decide *which* proposition gets grounded.

**Operating envelope** (full version: BACKGROUND.md §2.3). Trust most: faithfulness WITH a source
present (the truth-maker sits in the context window) and refutation of self-contradictory claims.
Trust least: positive grounding confirmation; specialist-knowledge / novel-domain claims (Producer
and Critic share one `c.model`, so a blind spot in one role survives in the other); and the tool's
own headline (unfalsifiable from the armchair). BACKGROUND.md §2.1 pins the eleven weaknesses behind
this envelope to specific code loci — consult the relevant row before changing a mode.

## Working disciplines

- Every output is assayable, and decision-carrying outputs get run through `./assay` before they're
  trusted. Internal/planning work uses faithfulness (`-source`) and substance modes; grounding
  (`-evidence`) is expected to return `unverifiable` on intentions and predictions — that's the
  correct result, not a failure. Grounding only earns its keep on factual claims about the world.
- Grounded-summary discipline: every summary names the fuller source it compresses and is
  self-screened against the faithfulness critic's seven distortion modes before it's emitted — watch
  overstatement (hedges → certainties) and literalization (provocation → literal commitment)
  hardest. Summaries that carry decisions get the full `./assay -source notes.txt summary.txt` pass.
- Pre-register the attempt, not the success: write the `decision_log` entry when a change STARTS
  (hypothesis + alternatives not taken); fill in `outcome` later, including `abandoned`/`reverted`.
  Never log only survivors — dead ends are the denominator the classifier needs most.
- Goal-link or tag-as-detour: every change names the standing rigour-map objective (see Tracks
  above) it serves, or is tagged `maintenance`/`detour`. Detours and maintenance never get promoted
  to "the project focus."
- Separate done from advanced: every session report states (i) what shipped, (ii) whether the
  rigour-map goal moved and by how much, (iii) what toward the goal is still NOT done. "Green"
  never stands alone.
- Roads not taken: keep a parking lot (`SESSION.md`) of deferred and abandoned items so reversals
  and dead ends survive the session boundary.
- Adjudication initials: cc never files a line in an `adjudications.txt` under a human's initials. A
  leaf verdict cc drafts carries `cc`; a human promotes it by editing the line — correcting the verdict
  or reason as they read the leaf — and swapping `cc` for their own initials (spec/SERVE.md
  § Adjudications). A draft left under `PW` renders on the page as PW's own call, so borrowing the
  initials asserts a reading the human never made. Write `cc` and stop; promotion is the human's edit.


## Reporting to the human
The first line of every report is a locator: `example: <dir>` for corpus work (name the example the
run is about, e.g. `example: examples/ai-index-2026-coding`), or `repo` for code-only work that
touches no single example. It comes before the five lines below.

End-of-task reports are at most 5 lines:
  1. Done / not done, and the one number that matters.
  2. Anything I changed that you didn't ask for.
  3. Anything I couldn't do.
  4. Where the details are (file path).
  5. What you need to decide next, if anything.
Everything else goes in a file under the task directory
(REPORT.md), not in the chat. Tables, anomaly lists and
provenance belong in the file.