package main

// refuter_b_test is the Step-1 retrieval refuter: for each of the 5 real LCEIC claims, every source
// passage the Sonnet N=3 faithfulness run (eval/n3) quoted as support must appear in the fused
// bm25+embed retrieved set (token-budgeted, the same set the judge sees). A quote whose passage is
// missing means retrieval dropped evidence the judge relied on — reported here, with the passage's
// full fused rank, rather than hidden. The claim set is claims-faith-5.txt; the quotes are read from
// the committed eval/n3 chain; the corpus is the real hearings tree. Nothing is synthetic.
//
// Embeddings: the corpus and the 5 query vectors are read from the committed .embcache so this runs
// offline under `go test`. Set ASSAY_EMBED=1 (with a local ollama + nomic-embed-text) to (re)compute
// and write that cache — the one online step that seeds what the committed test then reads.

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"assay/internal/embed"
	"assay/internal/retrieve"
)

func TestRetrievalRefuterFusedCoversSonnetQuotes(t *testing.T) {
	const (
		claimsFile = "examples/vic-lceic/claims-faith-5.txt"
		chainFile  = "eval/n3/claims-faith-refuter.faithfulness.jsonl"
		corpusDir  = "examples/vic-lceic/sources/hearings"
	)
	ix, err := retrieve.Load(corpusDir)
	if err != nil {
		t.Fatalf("load corpus: %v", err)
	}
	var emb retrieve.Embedder
	if os.Getenv("ASSAY_EMBED") == "1" {
		emb = embed.New("", "", nil) // online: compute + write the cache
	}
	if err := ix.AttachEmbeddings(emb, embedCacheDir(corpusDir)); err != nil {
		t.Fatalf("attach embeddings: %v", err)
	}
	if !ix.HasEmbeddings() {
		t.Fatalf("no embeddings available: seed the cache once with ASSAY_EMBED=1 (needs ollama + nomic-embed-text)")
	}

	claims := readClaimLines(t, claimsFile)
	quotesByIdx := readChainQuotes(t, chainFile)

	misses, matched, quotesChecked, unmatched := 0, 0, 0, 0
	f17Total, f17Covered := 0, 0
	fmt.Printf("\n── Refuter: Sonnet N=3 quotes vs fused bm25+embed retrieval (token cap %d) ──\n", retrieveTokenCap)
	for i, cl := range claims {
		q := claimQuery(cl.text, cl.path)
		got := ix.Retrieve(q, retrieveTokenCap, 0)
		full := ix.FusedRanking(q)
		fullRank := map[string]int{}
		for r, p := range full {
			fullRank[p.ID] = r + 1
		}
		fmt.Printf("%s  %q  (retrieved %d passages)\n", cl.id, firstN(cl.text, 55), len(got.Passages))
		for _, quote := range quotesByIdx[i] {
			quotesChecked++
			pid := passageContaining(ix, quote)
			covered := pid != "" && got.Ranks[pid] > 0
			switch {
			case pid == "":
				unmatched++
				fmt.Printf("    UNMATCHED  (quote not verbatim in any passage)  %q\n", firstN(quote, 55))
			case covered:
				matched++
				fmt.Printf("    ok    rank %-3d  %s  %q\n", got.Ranks[pid], pid, firstN(quote, 40))
			default:
				misses++
				fmt.Printf("    MISS  rank %-3d (outside budget)  %s  %q\n", fullRank[pid], pid, firstN(quote, 40))
			}
			if cl.id == "F17" {
				f17Total++
				if covered {
					f17Covered++
				}
			}
		}
	}
	fmt.Printf("── quotes checked: %d · in retrieved set: %d · missed: %d · unmatched: %d ──\n",
		quotesChecked, matched, misses, unmatched)

	// A MEASUREMENT, not a gate (per the step-1 spec: a miss "is the report, not a fix"). We fail only
	// if the measurement did not run (zero-output) or the positive control F17 — evidence in one
	// hearing — is not fully covered, which would mean retrieval is broken rather than merely lossy.
	if len(claims) != 5 || quotesChecked < 14 {
		t.Fatalf("refuter did not run over the full oracle: %d claims, %d quotes", len(claims), quotesChecked)
	}
	if f17Total == 0 || f17Covered != f17Total {
		t.Errorf("positive control F17 not fully covered (%d/%d) — retrieval broken, not merely lossy", f17Covered, f17Total)
	}
}

// ── helpers ──────────────────────────────────────────────────────────────────

type claimLine struct{ id, path, text string }

func readClaimLines(t *testing.T, path string) []claimLine {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read claims: %v", err)
	}
	var out []claimLine
	for _, ln := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if strings.TrimSpace(ln) == "" {
			continue
		}
		id, p, txt := parseClaimLine(ln)
		out = append(out, claimLine{id, p, txt})
	}
	return out
}

// readChainQuotes returns the defender quotes for chain idx 0..4 (the 5 real claims).
func readChainQuotes(t *testing.T, path string) map[int][]string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open chain: %v", err)
	}
	defer f.Close()
	out := map[int][]string{}
	dec := json.NewDecoder(f)
	for dec.More() {
		var rec struct {
			Idx    int `json:"idx"`
			Detail struct {
				Quotes []string `json:"quotes"`
			} `json:"detail"`
		}
		if err := dec.Decode(&rec); err != nil {
			t.Fatalf("decode chain: %v", err)
		}
		if rec.Idx < 5 {
			out[rec.Idx] = rec.Detail.Quotes
		}
	}
	return out
}

func firstN(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

var (
	wsRe     = regexp.MustCompile(`\s+`)
	footerRe = regexp.MustCompile(`(?i)(mon|tues|wednes|thurs|fri|satur|sun)day \d{1,2} [a-z]+ \d{4} legislative council economy and infrastructure committee \d{1,3}`)
	punct    = strings.NewReplacer("’", "'", "‘", "'", "“", `"`, "”", `"`, "—", "-", "–", "-")
)

// norm lowercases, unifies curly punctuation/dashes to ASCII, strips the page-footer boilerplate that
// Hansard injects mid-turn, and collapses whitespace — so a quote matches its passage despite a footer
// splitting it or a curly apostrophe differing. This is analysis hygiene, not a retrieval change.
func norm(s string) string {
	s = punct.Replace(strings.ToLower(strings.TrimSpace(s)))
	s = wsRe.ReplaceAllString(s, " ")
	s = footerRe.ReplaceAllString(s, " ")
	return wsRe.ReplaceAllString(s, " ")
}

// passageContaining returns the id of the passage a quote comes from. It first tries an exact
// (normalised) 60-char substring; when the defender lightly re-joined or altered a word — Sonnet
// wrote "arts sector" where the source says "creative sector" — the head window misses, so it also
// tries a tail and a middle window. A quote that matches none is genuinely non-verbatim and reported
// UNMATCHED rather than force-fitted to a passage.
func passageContaining(ix *retrieve.Index, quote string) string {
	for _, w := range windows(norm(quote), 60) {
		for _, p := range ix.Passages {
			if strings.Contains(norm(p.Text), w) {
				return p.ID
			}
		}
	}
	return ""
}

// windows returns up to three probe substrings of s — head, tail, middle — each width chars, so a
// single altered or re-joined span in the quote does not defeat the whole match. A short string is
// returned whole.
func windows(s string, width int) []string {
	if len(s) <= width {
		return []string{s}
	}
	mid := len(s)/2 - width/2
	return []string{s[:width], s[len(s)-width:], s[mid : mid+width]}
}
