package tree

// root.go renders the root summary block — the one screen above the tree, in both the text output and
// tree.html. It is a partition of the run's findings into named classes (every finding lands in exactly
// one, so the counts sum to N and nothing is dropped) plus the value line. spec/TREE.md § Root node
// rendering is the contract. No model call reaches here: every line comes from the verdict rows.

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"assay/internal/brief"
)

// RootMeta carries the header the block prints above the class counts.
type RootMeta struct {
	Title, Date string
	SourceDocs  int // M: documents in the manifest; 0 drops the "checked against M source documents" clause
	Runs        int // chains merged by -from; >1 adds the stability line (settled/wobble/contested)
}

// class is a finding's bucket in the block. The iota order is the print order.
type class int

const (
	clContested class = iota // the stability partition, taken first: sources that disagree across runs
	clHolds
	clContradicted
	clOverstated
	clUnsupported
	clOpinion
	clUnverifiable
	clSchemaFail // verdict unverifiable because the judge's reason failed the schema, not a missing document
	clGeneralised
	numClasses
)

const (
	rootIDCap   = 3  // ids shown per id-listing class before "and N more"
	rootWordCap = 34 // words of the stakes line shown on a detail line — fits both ≤12-word halves whole
	noCap       = 0  // writeIDClass cap for a class that shows every line — the contested group only
)

// classify places a row in exactly one class, in this precedence: opinion (never judged), then the
// stability partition (a contested leaf is filed by disagreement, not by verdict — spec/TREE.md § Root
// node rendering), then the schema gate (a malformed judge reason overrides the verdict), then the
// verdict itself. Opinion runs first so an opinion's verdict — even contradicted or absent — never
// lands it in a verdict-based class; contested runs before the verdict so it is excluded from every
// verdict group.
func classify(r brief.Row) class {
	if brief.IsOpinion(r) {
		return clOpinion
	}
	if r.Class == "contested" {
		return clContested
	}
	if r.SchemaFail {
		return clSchemaFail
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

	// The stability partition is taken first: the findings the runs could not agree on are the ones a
	// reader most needs, and they are excluded from every verdict class below. It shows every contested
	// line (noCap) — a reader must see the whole disagreement, not a sample of it — where the verdict
	// classes below stay capped at rootIDCap.
	writeIDClass(&b, "Sources don't settle these", buckets[clContested], contestedDetail, noCap)
	if n := len(buckets[clHolds]); n > 0 {
		fmt.Fprintf(&b, "Holds: %d findings say what their sources say.\n", n)
	}
	writeIDClass(&b, "Contradicted by their own sources", buckets[clContradicted], soWhatDetail, rootIDCap)
	writeIDClass(&b, "Overstated", buckets[clOverstated], soWhatDetail, rootIDCap)
	writeIDClass(&b, "Unsupported by any held source", buckets[clUnsupported], soWhatDetail, rootIDCap)
	if n := len(buckets[clOpinion]); n > 0 {
		fmt.Fprintf(&b, "Committee opinions, not checked: %d\n", n)
	}
	writeIDClass(&b, "Unverifiable, document not held", buckets[clUnverifiable], missingDocDetail, rootIDCap)
	writeIDClass(&b, "Unverifiable, judge output malformed", buckets[clSchemaFail], schemaFailDetail, rootIDCap)
	writeGeneralised(&b, buckets[clGeneralised])

	// When -from merged several runs, report how stable the verdicts were across them. The three
	// classes partition the *judged* findings; opinions are not judged, so they sit outside this line
	// (counted under Committee opinions above) and the three plus that count sum to N. The contested
	// figure here therefore equals the top group's count — both are the judged findings the runs could
	// not settle.
	if meta.Runs > 1 {
		settled, wobble, contested := stabilityCounts(rows)
		fmt.Fprintln(&b)
		fmt.Fprintf(&b, "Across %d runs: %d settled, %d wobble, %d contested.\n",
			meta.Runs, settled, wobble, contested)
	}

	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "Changes %d summary lines across %d chapters.\n",
		brief.OpenedSections(rows), brief.OpenedBranches(rows))
	return b.String()
}

// stabilityCounts tallies the merged-run stability classes carried on the judged rows. Opinions are
// skipped — they are never judged, so a stability class on one is a category error, and counting them
// here would make the contested figure disagree with the top group, which excludes them. A judged row
// with no Class (a single-chain render never sets one) counts as settled.
func stabilityCounts(rows []brief.Row) (settled, wobble, contested int) {
	for _, r := range rows {
		if brief.IsOpinion(r) {
			continue
		}
		switch r.Class {
		case "wobble":
			wobble++
		case "contested":
			contested++
		default:
			settled++
		}
	}
	return settled, wobble, contested
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

// writeIDClass writes a class header and, indented under it, up to `cap` detail lines (one per row,
// formatted by `detail`), then "and N more" when the class holds more. A `cap` of noCap shows every
// row with no "and N more" — the contested group takes that, the verdict classes pass rootIDCap.
// Nothing is written when empty.
func writeIDClass(b *strings.Builder, label string, rows []brief.Row, detail func(brief.Row) string, cap int) {
	if len(rows) == 0 {
		return
	}
	fmt.Fprintf(b, "%s: %d\n", label, len(rows))
	shown := rows
	if cap > 0 && len(shown) > cap {
		shown = shown[:cap]
	}
	for _, r := range shown {
		fmt.Fprintf(b, "  %s\n", detail(r))
	}
	if cap > 0 && len(rows) > cap {
		fmt.Fprintf(b, "  and %d more\n", len(rows)-cap)
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

// contestedDetail is the one plain line a contested leaf shows: the claim head then the disagreement
// clause as a second sentence — "The report says <claim>. Runs split a/b." when the pool tied and no
// verdict can be named, else "The report says <claim>. <modal> <spread> against <dissent>." (the
// verdicts that crossed the divide). The join strips any trailing periods off the head — one from a
// normal so-what, two when the so-what itself ended doubled — and re-adds exactly one before the split
// clause, so the two never run together and never double the period. It reads as prose so the top group
// carries the claim, not just verdict names — the refuter is `TestRootBlockContestedFirst`.
func contestedDetail(r brief.Row) string {
	head := reportSaysHead(r)
	tail := contestedSplit(r)
	if head == "" {
		return r.ID + ": " + tail
	}
	return r.ID + ": " + strings.TrimRight(head, ".") + ". " + tail
}

// reportSaysHead is the "The report says <claim>" clause the contested line opens with — the claim
// lower-cased at its first letter and stripped of its trailing period, so `contestedDetail` places
// exactly one period before the disagreement clause that follows. It reuses the leaf's
// assembled so-what (the judge's plain restatement) where the modal chain produced one, and falls back
// to the claim text when the modal verdict was faithful and left no so-what — either way the line
// carries the claim rather than only verdict names.
func reportSaysHead(r brief.Row) string {
	if s := strings.TrimSpace(r.SoWhat); strings.HasPrefix(s, "The report says ") {
		if i := strings.Index(s, ". "); i >= 0 {
			s = s[:i+1] // just the report-says sentence; the settled tail is not this leaf's to claim
		}
		return sentenceJoin(s)
	}
	if t := truncWords(r.Text, rootWordCap); t != "" {
		return sentenceJoin("The report says " + t + ".")
	}
	return ""
}

// sentenceJoin turns a closed "The report says X." head into a "The report says x" opener: it strips
// the trailing period and lower-cases the claim's sentence-initial capital (the character after the
// fixed "The report says " prefix); `contestedDetail` re-adds one period before the disagreement clause.
// An acronym opener is left alone — a claim beginning "COVID-19" or "ABC" has an upper-case second
// letter, so lower-casing only the first would produce "cOVID-19"/"aBC" and defeat the join; the same
// upper-case-run test isPlaceWord uses to tell an acronym from an ordinary word.
func sentenceJoin(s string) string {
	const prefix = "The report says "
	s = strings.TrimSuffix(s, ".")
	if !strings.HasPrefix(s, prefix) {
		return s
	}
	r := []rune(s[len(prefix):])
	if len(r) == 0 {
		return s
	}
	if len(r) >= 2 && unicode.IsUpper(r[1]) {
		return prefix + string(r) // acronym opener: keep the capital
	}
	r[0] = unicode.ToLower(r[0])
	return prefix + string(r)
}

// contestedSplit is the second sentence: "Runs split a/b." for a tie (no majority to name), else the
// crossing verdicts as "<modal> <spread> against <dissent>."
func contestedSplit(r brief.Row) string {
	if r.Split != "" {
		return "Runs split " + r.Split + "."
	}
	v := verdictOf(r)
	if r.Spread != "" {
		v += " " + r.Spread
	}
	if r.Dissent != "" {
		v += " against " + r.Dissent
	}
	return v + "."
}

// schemaFailDetail is "id: judge output malformed" — the flag for a leaf whose verdict was forced to
// unverifiable because the judge's reason was empty or carried a raw tag. The full malformed reason is
// preserved on the tree leaf itself (schemaLeafFlag); the block only names the failure.
func schemaFailDetail(r brief.Row) string {
	return r.ID + ": judge output malformed"
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
