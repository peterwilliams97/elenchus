# spec/ — the contract the code follows

`assay` (renamed `crossexam` under the planned **elenchus2** layout — the two name one tool) is a
**report → argument-tree checker**: it takes a long report, decomposed into atomic claims tagged with
their §-heading trail, and asks three separate questions of it — does each claim *say what its source
says*, does each claim *hold up under the dialectic*, and does each *recommendation actually follow
from the finding it rests on*. The answers are laid out as trees a human reads, never as a single
score.

These files are the **authoritative specs**. Code follows them, not the other way round: where a spec
and the code disagree, the spec is the contract and the code is the bug (`CLAUDE.md` § What this is).
None of this is user-facing prose — `README.md` at the repo root is the abstract; `BACKGROUND.md` is
the design rationale and failure envelope; these are the build contract.

## Why the specs are shaped this way — the axis boundary

Everything here serves one discipline, stated in `CLAUDE.md` § The axis boundary. The tool has three
competences and they do not reach equally far:

- **Close reading** reaches **faithfulness** — is the claim a distortion-free compression of its
  source. The truth-maker (the source) sits in the context window, so this is the column to trust.
- **Dialectic** reaches **substance** — does the proposition survive a producer↔critic exchange.
- **Reasoning** can **refute** a grounding claim (an internal contradiction kills it with no lookup)
  but can **never confirm** one. Positive grounding — "this is true of the world", "this
  recommendation follows" — always needs the truth-maker, which is a *retrieval*, not a deduction,
  however rigorous the reasoning.

The characteristic failure the tool is built to prevent is **laundering confidence**: scoring real
wins on the two reachable columns, then pronouncing on the unreachable one with borrowed authority.
So every spec below is written to keep the columns apart — a judge that cannot cite a verified quote
gets no verdict, a defeater the report does not name is not admitted, an inference is never certified
as sound. The credibility of a claim-validation tool is its substrate; these specs are the substrate.

## The specs, and how they compose

Read them in dependency order — each builds on the one before.

1. **[`CLI.md`](CLI.md)** — the command-line surface and the model/judge path. The modes (`dialectic`
   default, `-source` faithfulness, `-evidence` grounding, `-audit`), the **backend** seam (one
   `Complete` wrapper per provider), **retrieval** (corpus split into speaker-turn passages, ranked
   per claim by fused BM25 + local embeddings), and the **faithfulness judge** (one schema-enforced
   call whose every cited quote is verified against its passage in code, no model). This is where the
   axis boundary is enforced mechanically: no web search on Ollama ⇒ grounding returns `unverifiable`
   rather than a fabricated citation; a verdict with no verified quote is downgraded, not trusted.

2. **[`TREE.md`](TREE.md)** — the **report tree**: the report's findings grouped by their §-heading
   trail, one faithfulness verdict per leaf. Defines the merged-leaf verdict and its **stability
   class** (`settled` / `wobble` / `contested`) over N samples, the **Needs-you** selection that
   surfaces only the findings a reader must open, the **root summary block**, the **route** gate
   (opinions are never judged), **cite-scoped retrieval** (a claim citing an external truth-maker is
   judged against *that*, never against the report restating the same figure), and the five code-side
   rules that stand between the model and the tree. This is the top-level artifact.

3. **[`ARGUMENT.md`](ARGUMENT.md)** — the **argument tree**: the same leaves rearranged by
   *inferential* structure — claim → finding → recommendation → root thesis — with every internal
   node's judgement **derived bottom-up** from its children (`holds` / `weakened` / `open` / `fails`),
   never authored. The report tree is the input; this propagates its leaf verdicts up to the
   conclusions that depend on them. The root is a **tally** of its recommendations, not a conjunction:
   one failed branch does not collapse the whole case.

4. **[`EDGE.md`](EDGE.md)** — the **edge-level adversarial pass**: the one question the first two
   trees never ask — given a finding that *stands*, does the recommendation it licenses actually
   follow? It attacks the finding→recommendation inference with a Walton **scheme**'s fixed
   **critical questions**, and admits a **defeater** only if code confirms its **anchor** is a
   referent the report itself names (finding, recommendation, or verified quote — not the model's own
   prose). It may refute an edge (`open`) or decline (`unchallenged`); it may **never certify**
   soundness — the schema has no field for it. This is the axis boundary applied to inference. Its
   refuter runs are logged in-file as recorded negatives, per `CLAUDE.md` § Pre-register the attempt.

5. **[`SERVE.md`](SERVE.md)** — `assay -review` builds a self-contained, PDF-linked two-pane site from
   a saved chain (no model calls), and `assay serve` serves it over loopback HTTP so the page's
   `#page=N` links land on the right page. Also specifies the **adjudications** overlay — a human's own
   verdicts beside the machine's, where `cc` marks a draft and a human promotes it by swapping in their
   initials (`CLAUDE.md` § Adjudication initials).

6. **[`SUBSTANCE-CORPUS.md`](SUBSTANCE-CORPUS.md)** — adds the **substance** axis to the corpus
   pipeline (`-axis substance`): a second chain beside the faithfulness one, so a leaf can be
   `faithful` *and* be marked where its content does not survive the dialectic. A faithful-but-hollow
   leaf lowers its parent from `holds` to `weakened` — substance only ever lowers a leaf faithfulness
   left standing, never rescues one it sank.

`CLI.md`, `TREE.md`, `SUBSTANCE-CORPUS.md` describe shipped behaviour; `ARGUMENT.md` and `EDGE.md`
carry a *spec-only, no-code-yet* scope note at the top and are the design ahead of the code.

## Glossary

Terms recur across the specs; each is defined once, here, and used verbatim after (`CLAUDE.md` § Call
things by their known names).

- **claim / atomic claim** — the smallest proposition a source can settle, decomposed from the report
  and tagged with its §-heading **path** and a **route**. The leaves of both trees.
- **finding** — what the report *concludes* (`F…`); groups the claims decomposed from it.
- **recommendation** — what the report asks government/industry to *do* (`R…`); rests on findings.
- **root proposition / thesis** — the one thing the whole report exists to establish.
- **route** — what kind of thing can settle a claim: `source` (checked against the corpus),
  `evidence` (a fact about the world; needs external retrieval), `evaluative` (a value judgement —
  an **opinion**, never judged), `data-gap` (a claim about a dataset's own limits).
- **faithfulness** — the axis asking *does the claim say what its source says*. Verdicts: `faithful`,
  `partial`, `overstated`, `absent`, `contradicted`, `unsupported`, `unverifiable`.
- **substance** — the axis asking *does the proposition survive the dialectic*. Verdicts:
  `substantive`, `partial`, `hollow`.
- **grounding / evidence** — the axis checking a claim against the world via web retrieval; the
  column reasoning can refute but not confirm (see the axis boundary above).
- **passage** — one speaker-turn of a transcript (or a paragraph of another source), role-tagged
  (`witness` / `questioner` / `chair`), the unit retrieval ranks and the judge quotes.
- **manifest** — the held-document set (`sources/MANIFEST.md`), each keyed by a canonical id, so a run
  tells `absent` (checked, not there) from `unverifiable` (the truth-maker is not held).
- **cites / cite-scoped retrieval** — a claim's declared truth-maker documents; when a claim cites an
  external document, it is judged against that alone and every report passage is excluded, so the
  report cannot launder its own restatement into corroboration.
- **Tier-2 chain** — the JSONL record per claim (verdict, N-sample spread, retrieved passage ids,
  verified quotes, reason) a judge run writes; `-from` re-renders trees from it with no model calls.
- **stability class** — `settled` (every sample agrees) / `wobble` (moves but stays one side of the
  support divide, majority holds) / `contested` (the runs cross the divide, or no majority) — the
  merged-leaf class over N samples.
- **Needs-you** — the surfaced subset of leaves a reader must look at, chosen by a fixed tier rule;
  the tree collapses everything else.
- **scheme / critical questions (CQ)** — a Walton inference pattern (`practical`, `survey`, `example`,
  `trend`, `classification`) and its fixed short list of questions the edge attacker asks; one such
  question is a **critical question**, abbreviated **CQ** in `EDGE.md` and the schema's
  `critical_question` field.
- **defeater / anchor** — a concrete counter-world in which the finding still holds yet the
  recommendation fails; its **anchor** is the referent the world turns on, which admission requires to
  appear verbatim in the report's own text.
- **edge verdict** — `open` (an admitted defeater stands) or `unchallenged` (none admitted); combines
  with a recommendation's leaf-derived judgement, never replaces it.
- **derived judgement** — an internal node's `holds` / `weakened` / `open` / `fails`, computed
  bottom-up from its children; authoring one by hand is the edit the argument-tree format forbids.
- **adjudication** — a human's own verdict on a leaf, shown beside the machine's; `cc` marks a draft,
  a human's initials mark a promoted one.
- **backend** — the single `Complete` seam every LLM provider implements (`anthropic`, `ollama`,
  test `fake`), so a mode never names a provider.
- **elenchus2 / `crossexam`** — the planned `cmd/` + `internal/` binary these assay-era specs port to
  unchanged; until it exists, `assay` and `crossexam` name the same tool.
