# Audit — what assay checks today, and what the dormant adversarial leg misses (2026-09-13)

Read-only. No model calls, no commit. Evidence is committed chains under `examples/*/evidence/` and
code read this session at the cited `assay.go:NN` / `internal/…` lines.

## a) What the tool actually produces on the four corpora

**Finding up front:** every committed run on all five report corpora (`vic-lceic`, `quocirca-2026`,
`ai-index-2026-coding`, `dora-2026`, `master-plan`) is a **faithfulness** chain
(`*.faithfulness.jsonl`). The **substance** and **grounding** axes are implemented but are only
reachable through the single-file `-source`/`-evidence`/`-audit` modes — **neither runs on any of the
four/five corpora.** So of assay's three axes, exactly one is exercised on the reports.

Verdict tally across the committed corpus chains (`grep '"verdict"'`):

    416 faithful · 211 absent · 193 partial · 55 unsupported · 36 contradicted · 28 overstated
     11 unverifiable · 10 error

### Faithfulness axis — the report tree (the only axis run on the corpora)

| verdict class | code path that computes it | axis |
|---|---|---|
| `faithful`/`partial`/`overstated`/`absent`/`contradicted` | `faithJudge` `assay.go:1352` → `callSchema` (schema-enforced) | faithfulness |
| quote-presence downgrade → `unsupported`/`absent` | `groundVerdict` `assay.go:1585` (no verified quote ⇒ contradicted→absent, faithful/partial/overstated→unsupported) | faithfulness |
| verbatim quote check that feeds the downgrade | `quoteInPassage` `assay.go:1675` (substring, normalised whitespace/curly punct only — `normQuote` :1693) | faithfulness |
| `unverifiable` (cited doc not held) | manifest held-check; `internal/manifest/manifest.go` (`Held` :105), decided pre-model | faithfulness |
| `absent` (retrieval floor / cite-scoped miss) | `passagesForClaim` `assay.go:870`, `citedExternalBases` :907, `-floor` | faithfulness |
| `unverifiable` (schema gate — empty/raw-tag reason) | `brief.SchemaFailed` (`spec/TREE.md` §rules 5) | faithfulness |
| modal verdict + `k/N` spread + stability class (`settled`/`wobble`/`contested`/`split`) | `faithJudgeRepeat` `assay.go:1609`, `modalVerdict` :2306, `spreadFromSamples` :1638 | faithfulness |
| single-source direction rules (overstated/contradicted only vs. a narrower, more-detailed passage) | `faithJudgeSingleSourceRules`; `dropOwnParagraph` — active on `quocirca`, `dora`, `master-plan` (`single_source: true`); `ai-index` is now `single_source: false` | faithfulness |
| report-tree rollup (§-heading tree, root block, Needs-you tiers) | `internal/tree/tree.go`, `root.go`; selection `internal/brief` `Qualify` | faithfulness (rollup only) |

**Argument-tree rollup.** `internal/tree/argument.go` derives internal-node judgements
(`holds`/`weakened`/`open`/`fails`) bottom-up from leaf verdicts: `Judgement()` :159, `leafJudgement`
:196, `decidingChild` :257. It **reads faithfulness leaves and propagates them** — it computes no new
axis. (Note: `spec/ARGUMENT.md`'s scope banner still says "no code yet"; that is stale — the
derivation is implemented and tested in `argument_test.go`.)

### Substance axis (elenchus) — implemented, NOT run on the corpora

`assayClaim` `assay.go:1275`: producer/critic loop (`producerSys`, `substanceCriticSys` :2409) over
the seven fixed axes → verdict ∈ `{substantive, partial, hollow}`; condition-laundering downgrade to
`hollow` when `survives_only_by_conditioning` (:1303). Reachable only via the default dialectic mode
and `-audit`. No corpus chain is a substance chain.

### Grounding axis (evidence) — implemented, NOT run on the corpora

`evidenceClaim` `assay.go:2329`: web-search → verdict ∈ `{supported, mixed, refuted, unverifiable}`;
`crossCheckEvidence` :2350 downgrades to `unverifiable` unless a model-cited URL is in the actually
retrieved set (proves *retrieval*, never *content* — the axis boundary). Reachable only via
`-evidence`/`-audit`. No corpus chain is a grounding chain.

## b) What we miss because the adversarial-review leg is dormant

The destructive probes (`examples/destructive/`) calibrate the **substance** critic (six axes) plus
the axis-gap and laundering cases. Because the corpora are run **faithfulness-only**, none of that
critic ever fires on them. Faithfulness is *orthogonal* to substance: a `faithful` verdict means "the
report compresses its source without distortion," never "the claim is true or the inference holds."
So for every probe defect, **the current corpus pipeline can pass a claim carrying it** — not because
the critic misses it, but because the critic is never invoked.

| probe / defect | would the faithfulness-only corpus pipeline pass a claim with this defect? |
|---|---|
| motte-and-bailey (equivocation: strong sense retreats to trivial) | **Yes** — faithfulness checks fidelity to the source term, not which sense is load-bearing |
| reference-class / base-rate (real number vs. gamed baseline) | **Yes** — the number is quoted verbatim; the comparison class is never assessed |
| hidden-premise (conclusion valid only under unstated premise) | **Yes** — the stated words match the source; the missing premise is invisible to faithfulness |
| unfalsifiable-dress (no observation could disconfirm) | **Yes** — falsifiability is a substance axis, not run |
| causal-narrative (mechanism story over one correlation) | **Yes** — the correlational sentence is faithful; causal overreach is unchecked |
| axis-gaps (category error, composition, survivorship) | **Yes** — not named by any axis, faithfulness included |
| laundering (faithful + substantive + **false**) | **Yes** — the whole point: faithful+substantive can co-occur with world-falsity; only the grounding column (not run) would catch it |

### Three real `faithful` leaves where the verified quote is present but the claim doesn't follow

The tree's positive verdict here is `faithful` (the faithfulness analogue of `-evidence`'s
`supported`). In each case `quoteInPassage` confirmed a verbatim substring, `groundVerdict` kept the
verdict, and the leaf renders green — while the claim carries an unchecked substance defect.

1. **`ai-index-2026-coding/evidence/2026-09-13-b/…faithfulness.jsonl` idx 49** — `faithful` 3/3,
   `route=evidence`. Claim: *"In the United States, productivity growth reached 2.7% in 2025, nearly
   double the 1.4% average of the previous decade."* Verified quote is that sentence verbatim. But the
   leaf sits in the AI-productivity chapter, and "nearly double the previous decade" is exactly the
   **reference-class** defect — the baseline is chosen, and the implicit AI attribution is a
   **causal-narrative** move. Faithfulness confirms only that the report copied the figure.

2. **`ai-index-2026-coding/…/idx 50`** — `faithful` 3/3, `route=evaluative`. Claim: *"Brynjolfsson
   (2026) frames the 2025 US productivity rise as the early stages of a 'J-curve,' in which
   organizations…"* Verified quote present. This is a **causal-narrative** laid over a single-year
   correlation; faithfulness passes because the framing is faithfully reported, not because the
   J-curve mechanism is established.

3. **`dora-2026/evidence/2026-09-13-b/…faithfulness.jsonl` idx 0** — `faithful` 3/3,
   `route=evaluative`. Claim: *"AI's primary role in software development is that of an amplifier: it
   magnifies the strengths of high-performing organizations and the dysfunctions of struggling ones."*
   The **quote of record** is a *different, weaker* proposition — *"AI has the potential to reshape how
   software is built, but it does not change organizational systems on its own. What it does do…is
   reflect how those systems actually operate."* The verbatim-substring grounding passed on an adjacent
   statement; the strong "primary role / magnifies" claim (a **motte-and-bailey / hidden-premise**
   shape) is not carried by the quote the leaf rests on. (Companion: idx 12, same pattern.)

Each is a true positive *for faithfulness* and an uncaught defect *on the axis the destructive probes
exist to guard* — which is dormant on these corpora.

## c) Proposed next step (one prompt-sized task) — re-run the calibration on two models

**Goal:** revive `examples/destructive/run.sh` on `claude-sonnet-4-6` (current default) and
`qwen3.6:27b-q4_K_M` (local, cross-model check). Do not run it here.

**Sonnet — no code change needed.** `run.sh` already takes the model positionally:

    cd examples/destructive && source ../../setkey.sh && ./run.sh all 10 claude-sonnet-4-6

(recreate `laundering/sources/ballmer_usatoday_2007.txt` first per `laundering/PROVENANCE.md` — it is
gitignored). This appends per-probe lines to `testing/calibration_log.jsonl` and dated summaries under
each `results/`.

**Qwen — three `run.sh` edits required**, because run.sh assumes Anthropic:
- **Backend is never passed.** Every `"$ASSAY"` call (:89, :132, :244) omits `-backend`, so it
  defaults to `anthropic`; a `qwen…` model id would be sent to Anthropic and error. Add a `BACKEND`
  arg (or infer: `case "$MODEL" in qwen*|*ollama*) BACKEND=ollama;; *) BACKEND=anthropic;; esac`) and
  pass `-backend "$BACKEND"` on all three assay invocations.
- **The API-key gate aborts local runs.** Line 55 hard-requires `ANTHROPIC_API_KEY`. Guard it with
  `[ "$BACKEND" = anthropic ]` so an Ollama run proceeds without a key.
- **Prereq:** `ollama pull qwen3.6:27b-q4_K_M` and `ollama serve` running. The canary
  (`-evidence`, web search) has no Ollama analogue — expect it to error or skip on the qwen pass; note
  that in the results rather than treating it as a breach.

The `run.sh` parsers read assay's own markdown output, not raw model text, so the tally/cross-tab
parsing is backend-agnostic once `-backend` is wired.

**Estimated cost.**
- Sonnet full pass: 6 substance probes × 10 runs (each a decompose + producer/critic loop) + a 10-run
  laundering audit (faith+substance+evidence per fragment) + a 3-run web-search canary. Order
  ~0.5–1M tokens total → roughly **$5–10** at Sonnet-4.6 rates. (Wide band — the audit and web-search
  legs dominate and vary with fragment count.)
- Qwen local: **$0 API** (local compute only; q4_K_M at 27B fits the 48 GB unified-memory budget).

**Not a gate.** Keep `run.sh` out of `build.sh`/`go test` (`examples/destructive/README.md`,
`design-notes.md`). The output is a distribution read by a human, dated and model-stamped; a clean run
licenses only "the envelope held on this fixture, this model, this time."
