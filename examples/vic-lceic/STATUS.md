# STATUS — LCEIC faithfulness baseline (2026-09-07)

**First complete tree.** All 69 report findings are judged in one uniform run —
`evidence/2026-09-08-full/` (Sonnet, retrieved, N=3, final corpus, judge + rules, $5.42): 23
faithful · 26 partial · 13 absent · 2 overstated · 2 contradicted · 1 unsupported · 2 unverifiable.
`current/` points at it (`spec/TREE.md` specifies the tree).

**What's run.** The LCEIC final-report findings are assayed for faithfulness against a real corpus of
15 hearing transcripts + 42 written submissions + 2 responses to questions on notice + 1 submission
attachment (`sources/`, gitignored; provenance and the 60-document `sources/MANIFEST.md` are tracked).
Retrieval fuses BM25 with nomic-embed-text by best-rank RRF, stitches each witness turn to its
preceding question, and is measured by `RETRIEVAL_REFUTER.md` (**15 of 18** Sonnet-quoted spans covered).
A single schema-enforced judge decides each claim and cites evidence by `{passage_id, quote}`, which
code verifies verbatim; a verdict asserting source content with no surviving quote is downgraded in
code (→ `absent`/`unsupported`), a claim whose cited document is not in the manifest is `unverifiable`
with no model call, and `faithful` requires the source to match the claim's subject, scope and
direction. Evals live under `evidence/`: the 2×2 (Sonnet/Qwen × full/retrieved/oracle), the
submissions rerun, the judge-scope before/after, and the manifest→fetch→rerun loop. **What's
verified.** Retrieval preserves Sonnet's verdicts at ~3.5× lower cost; Qwen is systematically more
lenient but still catches every mutant (3/3) in every cell; the grounding, downgrade, and unverifiable
rules fire live; and the three previously-"absent" scope claims (F10, F33, F54b) were each traced to
their cited document, which was fetched, turning two to `faithful` (SBS QoN and TNA's attachment state
them verbatim) and one to a *genuine* `absent` (the ABC QoN is present but does not make the claim).
Total model spend across all of it ≈ $4.6. **What's open (for a human).** The `adjudication.md` rows
— F4's scope hedge ("all jurisdictions *where this research was conducted*" vs the claim naming New
York/Sweden/UK) and the five new scope-risk claims — carry a blank "who's right" column; a full
per-finding `cites` column across `claims-machine.txt` is not yet written (only the scope scoring set
is annotated); and the few Qwen before/after wobbles in the scope run are N=3 temperature spread, not
a prompt effect.
