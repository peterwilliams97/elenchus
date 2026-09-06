package main

// refuter_b_test is the Step-2 retrieval refuter: for each of the 5 real LCEIC claims, every source
// passage the Sonnet N=3 faithfulness run (eval/n3) quoted as support must appear in this retriever's
// top-8. A quote whose passage is missing means retrieval dropped evidence the judge relied on — a
// real gap, reported here rather than hidden. The claim set is claims-faith-5.txt; the quotes are read
// from the committed eval/n3 chain; the corpus is the real hearings tree. Nothing is synthetic.

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"assay/internal/retrieve"
)

func TestRetrievalRefuterTop8CoversSonnetQuotes(t *testing.T) {
	const (
		claimsFile = "examples/vic-lceic/claims-faith-5.txt"
		chainFile  = "eval/n3/claims-faith-refuter.faithfulness.jsonl"
		corpusDir  = "examples/vic-lceic/sources/hearings"
		k          = 8
	)
	ix, err := retrieve.Load(corpusDir)
	if err != nil {
		t.Fatalf("load corpus: %v", err)
	}
	claims := readClaimLines(t, claimsFile)   // 5 claims, in file order
	quotesByIdx := readChainQuotes(t, chainFile)

	misses, matched, quotesChecked := 0, 0, 0
	f17Total, f17Covered := 0, 0
	fmt.Println("\n── Refuter (b): Sonnet N=3 quotes vs retrieval top-8 ──")
	for i, cl := range claims {
		top := ix.Search(cl.text+" "+hintFromPath(cl.path), k, retrieveTokenCap)
		topSet := map[string]int{}
		for r, p := range top {
			topSet[p.ID] = r + 1 // 1-based rank
		}
		full := ix.Search(cl.text+" "+hintFromPath(cl.path), 0, 0) // full ranking for miss diagnosis
		fullRank := map[string]int{}
		for r, p := range full {
			fullRank[p.ID] = r + 1
		}
		fmt.Printf("%s  %q\n", cl.id, firstN(cl.text, 60))
		for _, q := range quotesByIdx[i] {
			quotesChecked++
			pid := passageContaining(ix, q)
			covered := pid != "" && topSet[pid] > 0
			switch {
			case pid == "":
				fmt.Printf("    NO-PASSAGE-MATCH  %q\n", firstN(q, 60))
			case topSet[pid] > 0:
				matched++
				fmt.Printf("    ok  rank %d  %s  %q\n", topSet[pid], pid, firstN(q, 45))
			default:
				misses++
				where := "not retrieved (no query-term overlap)"
				if r := fullRank[pid]; r > 0 {
					where = fmt.Sprintf("rank %d (below top-%d)", r, k)
				}
				fmt.Printf("    MISS  %s  %s  %q\n", where, pid, firstN(q, 45))
			}
			if cl.id == "F17" {
				f17Total++
				if covered {
					f17Covered++
				}
			}
		}
	}
	fmt.Printf("── quotes checked: %d · in top-%d: %d · missed (rank>%d or score 0): %d ──\n",
		quotesChecked, k, matched, k, misses)

	// This is a MEASUREMENT, not a gate: the spec says a miss "is the report, not a fix", so we do not
	// fail on scattered-evidence misses — they are the finding, printed above and recorded in
	// examples/vic-lceic/RETRIEVAL_REFUTER.md. We fail only if the measurement did not actually run
	// (zero-output rule) or if the positive control — F17, whose evidence sits in one hearing — is not
	// fully covered, which would mean retrieval is broken rather than merely lossy on scattered claims.
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

// passageContaining returns the id of the passage whose text contains the quote (normalised
// whitespace, first 60 chars as the key to tolerate stored-quote truncation), or "" if none.
func passageContaining(ix *retrieve.Index, quote string) string {
	key := norm(quote)
	if len(key) > 60 {
		key = key[:60]
	}
	for _, p := range ix.Passages {
		if strings.Contains(norm(p.Text), key) {
			return p.ID
		}
	}
	return ""
}
