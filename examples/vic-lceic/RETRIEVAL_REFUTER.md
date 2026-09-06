# Retrieval refuter (Step 2b) — BM25 top-8 vs the Sonnet N=3 evidence

**Finding: BM25 top-8 retrieval drops most of the passages the full-corpus judge actually quoted.**
This is reported, not fixed — per the Step-2 spec, "if any is missing, that's the report, not a fix."

## What this checks

The oracle is the committed Sonnet N=3 faithfulness run over the full 15-transcript corpus
(`eval/n3/`, `model claude-sonnet-4-6 · calls 48`). For each of the 5 real claims
(`claims-faith-5.txt`: F8, F12, F17, F29, F31 = chain idx 0–4), it took every verbatim source span the
defender quoted and asked: does the passage that span comes from appear in this retriever's top-8 for
that claim (query = claim text + §-heading labels, `-k 8`, ~8K-token cap)?

Reproduce: `go test . -run TestRetrievalRefuter -v` (in `refuter_b_test.go`). Deterministic, no model.

## Result

- **18 quoted spans checked · 7 in top-8 · 9 missed (ranked below 8, or score 0) · 2 unmatched.**
- Only **F17** — whose evidence is concentrated in one hearing (Yarra City Council) — is fully covered
  (ranks 1 and 3). It is the test's positive control.
- **F12** (3 of 4 missed), **F31** (4 of 5 missed), **F8** and **F29** (1 each missed) fail. The missed
  passages rank 9, 12, 45, 46, 139 — and one ("Absolutely inadequate.") never retrieves at all because
  it shares no query term.
- The 2 `NO-PASSAGE-MATCH` spans (both F8) do exist in the corpus
  (`6_bendigo-theatre-company-and-sertori-consulting`, `2_community-music-victoria`) but were lightly
  re-joined by the defender, so exact-substring can't pin their rank; they are not counted as top-8
  hits either way.

## Why it misses (mechanism, not a bug)

The full-corpus judge quotes evidence **spread across many hearings** for one claim — F31 (regional
funding) draws on 4 different transcripts. BM25 top-8 on a single claim surfaces the lexically
concentrated passages and cannot reach evidence that is scattered or phrased without the claim's terms
(a cross-reference like "underscore Jo Porter's remarks", rank 139; a bare "Absolutely inadequate.",
score 0). Retrieval is lossy exactly where the evidence is distributed — which is most of these claims.

## Table (verbatim test output)

```
── Refuter (b): Sonnet N=3 quotes vs retrieval top-8 ──
F8  "COVID-19 took away critical training opportunities from thos"
    NO-PASSAGE-MATCH  "The arts sector had shut down, and as a mother I had no viab"
    NO-PASSAGE-MATCH  "I do think music teachers or anyone who has chosen music tea"
    MISS  rank 46 (below top-8)  2025-03-13/1_st-martins-and-theatre-works#t54  "I can add to that the education bodies that t"
F12  "COVID-19 led to worsening mental health in the industries, p"
    MISS  rank 9 (below top-8)  2025-03-13/1_st-martins-and-theatre-works#t51  "We are dealing with a huge increase in anxiet"
    MISS  rank 45 (below top-8)  2025-03-13/1_st-martins-and-theatre-works#t49  "The mental health impact afterwards on artist"
    MISS  rank 46 (below top-8)  2025-03-13/2_arena-rawcus-lamama#t127  "I remember one young person that I had worked"
    ok  rank 1  2025-02-27/3_multicultural-arts-victoria#t11  "our report Beyond Tokenism calls for support "
F17  "Ticket prices are a key barrier for people wanting to engage"
    ok  rank 1  2025-02-27/1_yarra-city-council#t14  "What we do know is people will go to programm"
    ok  rank 3  2025-02-27/1_yarra-city-council#t17  "2.1 million fewer people have attended fewer "
F29  "Victoria receives its fair share of funding from Creative Au"
    ok  rank 3  2025-03-12/1_creative-victoria-and-vicscreen#t20  "Creative Australia's 2023–24 annual report "
    ok  rank 2  2025-03-13/3_a-new-approach#t28  "Victoria is the only state or territory that "
    MISS  rank 12 (below top-8)  2025-03-12/3_theatre-network-australia#t9  "on average in the past decade, when looking a"
F31  "Regional Victoria does not receive its fair share of funding"
    ok  rank 5  2025-02-27/2_regional-arts-victoria#t44  "regional Victorians receive $1.06 per capita "
    MISS  not retrieved (no query-term overlap)  2025-02-27/2_regional-arts-victoria#t47  "Absolutely inadequate."
    MISS  rank 139 (below top-8)  2025-03-12/2_community-music-victoria#t10  "I would like to particularly underscore Jo Po"
    MISS  rank 9 (below top-8)  2025-03-13/4_public-galleries-association-of-victoria#t11  "only one of which is a regional gallery. Give"
    MISS  rank 9 (below top-8)  2025-03-13/4_public-galleries-association-of-victoria#t11  "In the most recent round Victoria's public ga"
    ok  rank 6  2025-03-13/3_a-new-approach#t30  "when we cut the data from Creative Australia "
── quotes checked: 18 · in top-8: 7 · missed (rank>8 or score 0): 9 ──
```
