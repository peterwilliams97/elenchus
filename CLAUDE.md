# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`assay.go` — a self-contained Go CLI that runs prose through a three-stage dialectical filter using
the Claude API:
**decompose** (extract atomic claims) →
**dialectic** (producer↔critic loop per claim) →
**filter** (surface the substantive residue).
Build with `./build.sh` or `go build`.

Three operating modes, plus Go-only extras:

- **Dialectic** (default) — assay prose for logical integrity
- **Faithfulness** (`-source`) — check whether a summary accurately represents a source transcript
- **Evidence-grounding** (`-evidence`) — check each claim against external evidence via web search
- **Audit** (`-audit -source`) — all three modes in one cross-tab (Go only)
- **Markdown output** (`-md`) — emit markdown tables instead of terminal colour (Go only)

`assay.py` (the retired Python implementation) is archived at git SHA `469ebe4`.

**Docs roster:** `README.md` is the user-facing abstract; `BACKGROUND.md` is the design-rationale +
failure-envelope / destructive-self-criticism doc (where verdicts can't be trusted, plus the
destructive-test specs); `SESSION.md` is the parking lot of deferred work; `rigour-map/decision_log.jsonl`
is the decision/change log.

### Tracks

**Rigour-application map**: A parallel track (no Go code yet) that classifies
`(task, phase)` → `{advantage, disadvantage, irrelevant}` from a labeled decision log. Phase 0 is
data collection; `rigour-map/decision_log.jsonl` is the corpus. Open decision: seed the classifier
now vs. log unaided first to protect the disagreement baseline — not yet resolved.

## Running it

```sh
export ANTHROPIC_API_KEY=sk-ant-...

./build.sh                                             # test + build → ./assay

./assay memo.txt                                       # substance (default)
./assay -source transcript.txt summary.txt             # faithfulness
./assay -evidence claims.txt                           # grounding
./assay -source transcript.txt -evidence summary.txt   # grounding on intended proposition
./assay -audit -source transcript.txt -md summary.txt  # all three, markdown cross-tab
./assay -v memo.txt                                    # verbose: show every API call
./assay -model claude-opus-4-8 -max-rounds 3 memo.txt
```

`ANTHROPIC_MODEL` env var overrides the default model (`claude-sonnet-4-6`).

## Architecture

Everything lives in `assay.go`. The call graph by mode:

**Dialectic (default) — `runSubstance`**
```
main() → runSubstance(input)
  decompose(input)       # → []string of atomic claims via callJSON()
  for claim:
    assayClaim(claim)
      callJSON(producerSys, ...)   # steelman, blind to critique axes
      callJSON(substanceCriticSys, ...)  # 7 fixed axes → verdict + surviving_claim
                                         # + added_conditions + survives_only_by_conditioning
      loop if NeedsAnother && !SurvivesOnlyByConditions && rounds < maxRounds
      → downgrade to hollow at loop exit if SurvivesOnlyByConditions
  termSubstance / mdSubstance
```

**Faithfulness — `runFaithfulness`**
```
main() → runFaithfulness(input, src)
  splitSummary(input)    # splits numbered/bulleted lists or lines
  for claim:
    faithClaim(claim, src)
      callJSON(faithDefenderSys, ...)  # find strongest verbatim support
      callJSON(faithCriticSys, ...)    # verdict: faithful|partial|overstated|absent|contradicted
                                       # Literalization mode: rhetorical claim → overstated
  termFaith / mdFaith
```

**Evidence-grounding — `runEvidence`**
```
main() → runEvidence(input, src)
  splitSummary(input)
  for claim:
    if src != "": faithClaim(claim, src)
      → use SourceSays when verdict is partial|overstated
        (grounds the intended proposition, not literal words)
    evidenceClaim(proposition)
      callJSONSourced(evidenceSys, ..., withTools=true)  # web_search_20250305; captures retrieved URLs
      crossCheckEvidence(...)  # downgrade supported|mixed|refuted → unverifiable when the model's cited
                               # URLs are absent from the actually-retrieved set (d009/d010)
  termEvidence / mdEvidence
```

**Audit — `runAudit`** (Go only)
```
main() → runAudit(input, src)
  splitSummary(input)    # shared decomposition across all three modes
  for claim: faithClaim + assayClaim + evidenceClaim
  mdAudit(claims, fs, ss, es)  # always markdown
```

**Key design constraint:** `producerSys` deliberately omits the critique axes. The Producer must
steelman without knowing how it will be attacked.

**API plumbing:** `callClaude` makes a non-streaming POST to the Anthropic API. `callJSON` wraps it
with one retry on JSON parse failure, dispatching through `cfg.call` (nil in production → falls back
to `callClaude`; set to a stub in tests). `extractJSON` / `unmarshalLoose` handle markdown fences
and trailing commas. Web search uses `withTools=true`, which adds `web_search_20250305` to the
request. `callClaude` parses `web_search_tool_result` blocks into `retrievedSource`s (the URLs
actually fetched); `callJSONSourced` threads these back so `crossCheckEvidence` can compare them
(host+path, via `normalizeURL`) against the model's self-reported `sources`. Other tool-use blocks
(`server_tool_use`, etc.) are consumed silently (or logged in `-v`).

**Verdicts:**
- Dialectic: `"substantive"` | `"partial"` | `"hollow"`     | `"error"`. Only `substantive` and `partial` appear in the final residue. `partial` claims carry `SurvivingClaim`.
- Faithfulness: `"faithful"` | `"partial"` | `"overstated"` | `"absent"` | `"contradicted"`
- Evidence: `"supported"`    | `"mixed"`   | `"refuted"`    | `"unverifiable"`
  - Grounding integrity: `crossCheckEvidence` downgrades `supported`/`mixed`/`refuted` →
    `unverifiable` when the model's cited URLs are absent from the retrieved set (or nothing was
    retrieved/cited). This checks URL *provenance* (was the page fetched), NOT *content support*
    (does the page back the sentence) — see the axis boundary note. No code path validates the
    verdict strings against these enums (BACKGROUND.md W10); a junk verdict renders as-is.

**`maxTokens = 1500`** caps every Claude response. Raise this constant if critiques truncate. A
`max_tokens` cutoff (or stubborn malformed JSON past the single reparse) collapses a case to
`"error"`, which the tally counts as unverified rather than surfacing why (BACKGROUND.md W11).

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

These tests are **confirmatory** (nominal input → intended behaviour fires). The complementary
**destructive-testing program** — adversarial inputs that push each mode past its limit to map the
failure envelope — is specified in BACKGROUND.md §2.2 and tracked as the next step in `SESSION.md`;
it is not yet implemented. Atomization asymmetry to keep in mind when testing: substance uses the LLM
`decompose`, while faithfulness/evidence/audit use the regex `splitSummary` — different failure
surfaces (BACKGROUND.md W4), so test the one your change actually touches.

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
