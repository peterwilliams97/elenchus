# Retrieval refuter — Sonnet N=3 evidence vs fused bm25+embed retrieval

**Finding: fused BM25 + nomic-embed retrieval (token budget 8000) covers 12 of the 18 passages the
full-corpus Sonnet judge quoted, up from 7 under BM25 top-8. The residual 6 split into 3 unreachable
by any ranker and 3 reachable only past a ~35K-token budget — the irreducible scattered-evidence
limit.** Reported, not fixed further: per the step-1 spec, a miss "is the report, not a fix."

## What this checks

The oracle is the committed Sonnet N=3 faithfulness run over the full 15-transcript corpus
(`eval/n3/`, `model claude-sonnet-4-6 · calls 48`). For each of the 5 real claims
(`claims-faith-5.txt`: F8, F12, F17, F29, F31 = chain idx 0–4), it took every verbatim source span the
defender quoted and asked: does the passage that span comes from appear in the retriever's returned
set for that claim (query = claim text + §-heading labels, `-max-tokens 8000`)?

Retrieval is the same one the judge sees: BM25 and a nomic-embed-text cosine ranker fused by
reciprocal-rank fusion combined by *best* rank (not summed — the evidence is complementary, strong in
one ranker and weak in the other; summing penalises a true single-ranker hit and dropped a passage
BM25 alone had covered). Passages are admitted in fused order up to the token budget.

Reproduce: `go test . -run TestRetrievalRefuterFused -v` (in `refuter_b_test.go`). Deterministic, no
model call — it reads the committed embedding cache under `sources/hearings/.embcache`. Seed that
cache once with `ASSAY_EMBED=1` (needs a local ollama + `nomic-embed-text`).

## Result

- **18 quoted spans checked · 12 in the retrieved set · 6 missed · 0 unmatched.** (Was 7 in top-8
  under BM25 `-k 8`.)
- **F17** (evidence in one hearing, Yarra City Council) and **F29** (federal-funding share) are now
  fully covered. F17 is the positive control: both its spans rank 1 and 5.
- The 2 spans the old exact-substring locator could not pin now locate via a tail/middle probe window.
  One of them is a **non-verbatim** Sonnet quote — the defender wrote "The **arts** sector had shut
  down" where the source (`6_bendigo…#t15`) says "The **creative** sector had shut down". It is located
  and still just outside the budget (fused rank 17 of 16 admitted).
- The 6 misses are two kinds:
  - **Reachable, past the budget** — `6_bendigo…#t15` (rank 17), `1_st-martins…#t49` (76),
    `2_arena…#t127` (78). Admitting rank 78 needs ~35K tokens for one claim, which trades away the
    point of retrieval.
  - **Unreachable by any ranker** — `2_regional-arts…#t47` "Absolutely inadequate." (fused rank 1303:
    a bare answer whose meaning lives entirely in the question it answers), `2_community-music…#t37`
    "I do think music teachers…" (604: generic phrasing), `2_community-music…#t10` "underscore Jo
    Porter's remarks" (184: a cross-reference). BM25 score and cosine are both near the floor.

## Why it still misses (mechanism, not a bug)

The full-corpus judge quotes evidence spread across many hearings for one claim — F31 (regional
funding) draws on 4 transcripts. Adding a semantic ranker reaches evidence phrased differently from
the claim (`1_st-martins…#t54` sits at BM25 rank 46 but cosine rank 6, and is now covered). But it
cannot manufacture signal that is not in the passage: a terse answer, a generic sentence, and a
cross-reference to another witness carry almost no lexical *or* semantic overlap with the claim in
isolation, so both rankers place them in the deep tail. Retrieval is lossy exactly where the evidence
is distributed or context-dependent — which is a minority of these spans, but an irreducible one.

## Table (verbatim test output)

```
── Refuter: Sonnet N=3 quotes vs fused bm25+embed retrieval (token cap 8000) ──
F8  "COVID-19 took away critical training opportunities from"  (retrieved 16 passages)
    MISS  rank 17  (outside budget)  2025-03-13/6_bendigo-theatre-company-and-sertori-consulting#t15  "The arts sector had shut down, and as a "
    MISS  rank 604 (outside budget)  2025-03-12/2_community-music-victoria#t37  "I do think music teachers or anyone who "
    ok    rank 9    2025-03-13/1_st-martins-and-theatre-works#t54  "I can add to that the education bodies t"
F12  "COVID-19 led to worsening mental health in the industri"  (retrieved 19 passages)
    ok    rank 15   2025-03-13/1_st-martins-and-theatre-works#t51  "We are dealing with a huge increase in a"
    MISS  rank 76  (outside budget)  2025-03-13/1_st-martins-and-theatre-works#t49  "The mental health impact afterwards on a"
    MISS  rank 78  (outside budget)  2025-03-13/2_arena-rawcus-lamama#t127  "I remember one young person that I had w"
    ok    rank 1    2025-02-27/3_multicultural-arts-victoria#t11  "our report Beyond Tokenism calls for sup"
F17  "Ticket prices are a key barrier for people wanting to e"  (retrieved 23 passages)
    ok    rank 1    2025-02-27/1_yarra-city-council#t14  "What we do know is people will go to pro"
    ok    rank 5    2025-02-27/1_yarra-city-council#t17  "2.1 million fewer people have attended f"
F29  "Victoria receives its fair share of funding from Creati"  (retrieved 23 passages)
    ok    rank 1    2025-03-12/1_creative-victoria-and-vicscreen#t20  "Creative Australia's 2023–24 annual re"
    ok    rank 3    2025-03-13/3_a-new-approach#t28  "Victoria is the only state or territory "
    ok    rank 17   2025-03-12/3_theatre-network-australia#t9  "on average in the past decade, when look"
F31  "Regional Victoria does not receive its fair share of fu"  (retrieved 22 passages)
    ok    rank 6    2025-02-27/2_regional-arts-victoria#t44  "regional Victorians receive $1.06 per ca"
    MISS  rank 1303 (outside budget)  2025-02-27/2_regional-arts-victoria#t47  "Absolutely inadequate."
    MISS  rank 184 (outside budget)  2025-03-12/2_community-music-victoria#t10  "I would like to particularly underscore "
    ok    rank 14   2025-03-13/4_public-galleries-association-of-victoria#t11  "only one of which is a regional gallery."
    ok    rank 14   2025-03-13/4_public-galleries-association-of-victoria#t11  "In the most recent round Victoria's publ"
    ok    rank 8    2025-03-13/3_a-new-approach#t30  "when we cut the data from Creative Austr"
── quotes checked: 18 · in retrieved set: 12 · missed: 6 · unmatched: 0 ──
```
