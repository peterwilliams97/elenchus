# Retrieval refuter — Sonnet N=3 evidence vs fused bm25+embed retrieval

**Finding: fused BM25 + nomic-embed retrieval, with each witness passage carrying its preceding
question as context, covers 15 of the 18 passages the full-corpus Sonnet judge quoted, up from 7
under BM25 top-8. The residual 3 — a generic sentence, a cross-reference, and a deep-tail turn —
carry almost no lexical or semantic overlap with the claim and are the irreducible
scattered-evidence limit.** Reported, not fixed further: per the spec, a miss "is the report, not a
fix."

## What this checks

The oracle is the committed Sonnet N=3 faithfulness run over the full 15-transcript corpus
(`eval/n3/`, `model claude-sonnet-4-6 · calls 48`). For each of the 5 real claims
(`claims-faith-5.txt`: F8, F12, F17, F29, F31 = chain idx 0–4), it took every verbatim source span the
defender quoted and asked: does the passage that span comes from appear in the retriever's returned
set for that claim (query = claim text + §-heading labels, `-max-tokens 10000`)?

Retrieval is the same one the judge sees:

- **Passage = witness turn + its immediately preceding questioner/chair turn**, tagged separately
  (`Q — …` / `A — …`) inside the passage. A bare answer like Jo Porter's "Absolutely inadequate." is
  meaningless — and unretrievable — without the question it answers; the question is where the topical
  terms live. `Text` stays the witness's own words alone, so a quote is still matched to the answer
  precisely, while ranking and the judge see the question via the combined text.
- **Two rankers, fused.** BM25 over the combined text, and a nomic-embed-text cosine ranker over its
  vector. Fused by reciprocal-rank fusion combined by *best* rank (not summed — the evidence is
  complementary, strong in one ranker and weak in the other; summing penalises a true single-ranker
  hit and regressed the F17 positive control). Passages are admitted in fused order up to the budget.

Reproduce: `go test . -run TestRetrievalRefuterFused -v` (in `refuter_b_test.go`). Deterministic, no
model call — it reads the committed embedding cache under `sources/hearings/.embcache`. Seed that
cache once with `ASSAY_EMBED=1` (needs a local ollama + `nomic-embed-text`).

## Result

- **18 quoted spans checked · 15 in the retrieved set · 3 missed · 0 unmatched.** (Was 7 in top-8
  under BM25 `-k 8`; was 12 under fusion before the question-context change.)
- **The question-context change rescued the flagship miss:** "Absolutely inadequate."
  (`2_regional-arts…#t47`) moved from fused rank **1303 → 13**, now covered — its preceding turn is
  Gaelle Broad asking about "regional Victoria's share of national arts and culture spending …
  inadequate", which supplies every term the bare answer lacks.
- **The larger passages need a larger budget.** Carrying the question roughly doubled passage size, so
  the old 8000-token budget held fewer passages and briefly dropped two borderline wins. 10000 is the
  knee of the curve; coverage plateaus there:

  | `-max-tokens` | covered | avg passages/claim |
  |---------------|---------|--------------------|
  | 8000          | 12/18   | 19.6               |
  | **10000**     | **15/18** | **24.6**         |
  | 12000         | 15/18   | 30.6               |
  | 16000         | 15/18   | 44.8               |
  | 20000         | 15/18   | 62.4               |

- **The 3 residual misses are unreachable at any budget:** `2_community-music…#t37` "I do think music
  teachers…" (fused rank 628 — generic phrasing), `2_community-music…#t10` "underscore Jo Porter's
  remarks" (223 — a cross-reference to another witness), `2_arena…#t127` "I remember one young
  person…" (85 — a deep-tail anecdote). BM25 score and cosine are both near the floor; no ranker can
  manufacture signal a passage does not carry.

## Known positive for step-2b quote verification: a Sonnet paraphrase

`6_bendigo…#t15` is the **first recorded case for the step-2b verbatim-quote check.** The Sonnet
defender quoted "The **arts** sector had shut down, and as a mother I had no viable career prospects."
The source says "The **creative** sector had shut down …" — Sonnet substituted "arts" for "creative".
It is a paraphrase presented as a verbatim quote. The refuter's tail/middle-window locator still pins
it to the passage (so it counts toward coverage), but step 2b's grounding check — evidence quotes must
be a verbatim substring of the cited passage — must **reject** this record. Code catches what a fluent
paraphrase hides; this is the canary that proves it does.

## Table (verbatim test output)

```
── Refuter: Sonnet N=3 quotes vs fused bm25+embed retrieval (token cap 10000) ──
F8  "COVID-19 took away critical training opportunities from"  (retrieved 17 passages)
    ok    rank 17   2025-03-13/6_bendigo-theatre-company-and-sertori-consulting#t15  "The arts sector had shut down, and as a "
    MISS  rank 628 (outside budget)  2025-03-12/2_community-music-victoria#t37  "I do think music teachers or anyone who "
    ok    rank 16   2025-03-13/1_st-martins-and-theatre-works#t54  "I can add to that the education bodies t"
F12  "COVID-19 led to worsening mental health in the industri"  (retrieved 28 passages)
    ok    rank 14   2025-03-13/1_st-martins-and-theatre-works#t51  "We are dealing with a huge increase in a"
    ok    rank 7    2025-03-13/1_st-martins-and-theatre-works#t49  "The mental health impact afterwards on a"
    MISS  rank 85  (outside budget)  2025-03-13/2_arena-rawcus-lamama#t127  "I remember one young person that I had w"
    ok    rank 1    2025-02-27/3_multicultural-arts-victoria#t11  "our report Beyond Tokenism calls for sup"
F17  "Ticket prices are a key barrier for people wanting to e"  (retrieved 24 passages)
    ok    rank 1    2025-02-27/1_yarra-city-council#t14  "What we do know is people will go to pro"
    ok    rank 2    2025-02-27/1_yarra-city-council#t17  "2.1 million fewer people have attended f"
F29  "Victoria receives its fair share of funding from Creati"  (retrieved 23 passages)
    ok    rank 1    2025-03-12/1_creative-victoria-and-vicscreen#t20  "Creative Australia's 2023–24 annual re"
    ok    rank 2    2025-03-13/3_a-new-approach#t28  "Victoria is the only state or territory "
    ok    rank 16   2025-03-12/3_theatre-network-australia#t9  "on average in the past decade, when look"
F31  "Regional Victoria does not receive its fair share of fu"  (retrieved 31 passages)
    ok    rank 9    2025-02-27/2_regional-arts-victoria#t44  "regional Victorians receive $1.06 per ca"
    ok    rank 13   2025-02-27/2_regional-arts-victoria#t47  "Absolutely inadequate."
    MISS  rank 223 (outside budget)  2025-03-12/2_community-music-victoria#t10  "I would like to particularly underscore "
    ok    rank 17   2025-03-13/4_public-galleries-association-of-victoria#t11  "only one of which is a regional gallery."
    ok    rank 17   2025-03-13/4_public-galleries-association-of-victoria#t11  "In the most recent round Victoria's publ"
    ok    rank 14   2025-03-13/3_a-new-approach#t30  "when we cut the data from Creative Austr"
── quotes checked: 18 · in retrieved set: 15 · missed: 3 · unmatched: 0 ──
```
