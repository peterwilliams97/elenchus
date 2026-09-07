package tree

// root.go renders the root summary block — the one screen above the tree, in both the text output and
// tree.html. It is a partition of the run's findings into named classes (every finding lands in exactly
// one, so the counts sum to N and nothing is dropped) plus the value line. spec/TREE.md § Root node
// rendering is the contract. No model call reaches here: every line comes from the verdict rows.

import (
	"fmt"
	"sort"
	"strings"

	"assay/internal/brief"
)

// RootMeta carries the header the block prints above the class counts.
type RootMeta struct {
	Title, Date string
	SourceDocs  int // M: documents in the manifest; 0 drops the "checked against M source documents" clause
}

// class is a finding's bucket in the block. The iota order is the print order.
type class int

const (
	clHolds class = iota
	clContradicted
	clOverstated
	clUnsupported
	clOpinion
	clUnverifiable
	clGeneralised
	numClasses
)

const (
	rootIDCap   = 3  // ids shown per id-listing class before "and N more"
	rootWordCap = 34 // words of the stakes line shown on a detail line — fits both ≤12-word halves whole
)

// classify places a row in exactly one class. Opinion resolution runs first, so an opinion's
// faithfulness verdict — even contradicted or absent — never lands it in a verdict-based class.
func classify(r brief.Row) class {
	if brief.IsOpinion(r) {
		return clOpinion
	}
	switch verdictOf(r) {
	case "faithful":
		return clHolds
	case "contradicted":
		return clContradicted
	case "overstated":
		return clOverstated
	case "unverifiable":
		return clUnverifiable
	case "partial":
		return clGeneralised
	default:
		// absent, unsupported, and any error/unknown verdict: not backed by a held source. Kept in one
		// visible class rather than dropped, so the counts still sum to N.
		return clUnsupported
	}
}

// RootBlock renders the block for `rows`. It ends with a trailing newline; the caller prints the tree
// after a blank line.
func RootBlock(rows []brief.Row, meta RootMeta) string {
	buckets := make([][]brief.Row, numClasses)
	for _, r := range rows {
		c := classify(r)
		buckets[c] = append(buckets[c], r)
	}

	var b strings.Builder
	fmt.Fprintln(&b, titleLine(meta, len(rows)))
	fmt.Fprintln(&b)

	if n := len(buckets[clHolds]); n > 0 {
		fmt.Fprintf(&b, "Holds: %d findings say what their sources say.\n", n)
	}
	writeIDClass(&b, "Contradicted by their own sources", buckets[clContradicted], soWhatDetail)
	writeIDClass(&b, "Overstated", buckets[clOverstated], soWhatDetail)
	writeIDClass(&b, "Unsupported by any held source", buckets[clUnsupported], soWhatDetail)
	if n := len(buckets[clOpinion]); n > 0 {
		fmt.Fprintf(&b, "Committee opinions, not checked: %d\n", n)
	}
	writeIDClass(&b, "Unverifiable, document not held", buckets[clUnverifiable], missingDocDetail)
	writeGeneralised(&b, buckets[clGeneralised])

	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "Changes %d summary lines across %d chapters.\n",
		brief.OpenedSections(rows), brief.OpenedBranches(rows))
	return b.String()
}

// titleLine is line 1. The "checked against M source documents" clause is dropped when M is unknown;
// the title and/or date prefix is dropped when the claims file declared none.
func titleLine(meta RootMeta, n int) string {
	var head string
	switch {
	case meta.Title != "" && meta.Date != "":
		head = meta.Title + ", " + meta.Date + " — "
	case meta.Title != "":
		head = meta.Title + " — "
	}
	if meta.SourceDocs > 0 {
		return fmt.Sprintf("%s%d findings checked against %d source documents", head, n, meta.SourceDocs)
	}
	return fmt.Sprintf("%s%d findings", head, n)
}

// writeIDClass writes a class header and, indented under it, up to rootIDCap detail lines (one per row,
// formatted by `detail`), then "and N more" when the class holds more. Nothing is written when empty.
func writeIDClass(b *strings.Builder, label string, rows []brief.Row, detail func(brief.Row) string) {
	if len(rows) == 0 {
		return
	}
	fmt.Fprintf(b, "%s: %d\n", label, len(rows))
	shown := rows
	if len(shown) > rootIDCap {
		shown = shown[:rootIDCap]
	}
	for _, r := range shown {
		fmt.Fprintf(b, "  %s\n", detail(r))
	}
	if len(rows) > rootIDCap {
		fmt.Fprintf(b, "  and %d more\n", len(rows)-rootIDCap)
	}
}

// writeGeneralised writes the partial class: the count, then up to 2 highlight lines for partials whose
// narrowing changes what a reader would do — a non-none gap plus a number or place name in the claim —
// ranked by spread (3/3 before 2/3), then id.
func writeGeneralised(b *strings.Builder, rows []brief.Row) {
	if len(rows) == 0 {
		return
	}
	fmt.Fprintf(b, "Generalised from narrower evidence: %d\n", len(rows))
	var hi []brief.Row
	for _, r := range rows {
		if r.Gap != "" && r.Gap != "none" && hasNumberOrPlace(r.Text) {
			hi = append(hi, r)
		}
	}
	sort.SliceStable(hi, func(i, j int) bool {
		si, sj := spreadRatio(hi[i].Spread), spreadRatio(hi[j].Spread)
		if si != sj {
			return si > sj // 3/3 before 2/3
		}
		return natLess(hi[i].ID, hi[j].ID)
	})
	if len(hi) > 2 {
		hi = hi[:2]
	}
	for _, r := range hi {
		fmt.Fprintf(b, "  %s\n", soWhatDetail(r))
	}
}

// soWhatDetail is "id: so-what (≤20 words)", or the id alone when the leaf carries no so-what.
func soWhatDetail(r brief.Row) string {
	if s := truncWords(r.SoWhat, rootWordCap); s != "" {
		return r.ID + ": " + s
	}
	return r.ID
}

// missingDocDetail is "id: <document>" for an unverifiable leaf, the document being what its so-what
// tells the reader to fetch ("fetch qon:abc/…" → "qon:abc/…").
func missingDocDetail(r brief.Row) string {
	doc := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(r.SoWhat), "fetch "))
	if doc == "" {
		return r.ID
	}
	return r.ID + ": " + doc
}

// spreadRatio turns a "k/N" spread into k/N as a float for ranking; a spread with no ratio ranks last.
func spreadRatio(s string) float64 {
	k, n, ok := strings.Cut(s, "/")
	if !ok {
		return 0
	}
	kf, nf := atof(k), atof(n)
	if nf == 0 {
		return 0
	}
	return kf / nf
}

func atof(s string) float64 {
	var n float64
	for _, r := range strings.TrimSpace(s) {
		if r < '0' || '9' < r {
			return 0
		}
		n = n*10 + float64(r-'0')
	}
	return n
}

// hasNumberOrPlace reports a digit anywhere in the claim, or a proper-noun place-name signal: a
// mixed-case capitalised word (≥2 letters) that is not the claim's first word. A sentence-initial
// capital and an all-caps acronym (ABC, SBS, NSW) do not count — the first is grammatical, the second
// is an organisation, not a place.
func hasNumberOrPlace(text string) bool {
	for _, r := range text {
		if '0' <= r && r <= '9' {
			return true
		}
	}
	words := strings.Fields(text)
	for i, w := range words {
		if i == 0 {
			continue
		}
		if isPlaceWord(w) {
			return true
		}
	}
	return false
}

// isPlaceWord reports a word that reads as a proper place noun: starts uppercase, holds a lowercase
// letter (so it is not an all-caps acronym), and has at least two letters. Trailing punctuation is
// ignored so "York," and "Sweden." still match.
func isPlaceWord(w string) bool {
	w = strings.TrimFunc(w, func(r rune) bool {
		return !('A' <= r && r <= 'Z' || 'a' <= r && r <= 'z')
	})
	if len(w) < 2 || !('A' <= w[0] && w[0] <= 'Z') {
		return false
	}
	for _, r := range w[1:] {
		if 'a' <= r && r <= 'z' {
			return true
		}
	}
	return false
}

// truncWords returns the first n whitespace-separated words of s, appending "…" when it truncates.
func truncWords(s string, n int) string {
	words := strings.Fields(strings.ReplaceAll(strings.TrimSpace(s), "\n", " "))
	if len(words) <= n {
		return strings.Join(words, " ")
	}
	return strings.Join(words[:n], " ") + "…"
}
