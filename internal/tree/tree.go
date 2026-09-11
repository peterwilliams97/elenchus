package tree

// tree renders claim verdicts as a text tree whose shape comes from the source document's own
// headings, carried in each row's Path as "/"-separated `key=label` segments — so this package
// holds no knowledge of any particular corpus, only the generic shape. spec/CLI.md §"Tree report"
// is the contract. Expansion reuses brief.Qualify/Selected so the tree and the brief agree on which
// claims a human must look at; a node stays collapsed to one line unless a claim below it needs
// attention (or expandAll is set).

import (
	"fmt"
	"html"
	"sort"
	"strconv"
	"strings"

	"assay/internal/brief"
)

// Complexity limits from spec/CLI.md. width and depth are hard (depth failure aborts the render);
// score is a target reported for a human to judge, never a failure.
const (
	maxWidth  = 7  // children per node
	maxDepth  = 4  // root to leaf
	maxLabel  = 12 // words in an internal-node label
	scoreWarn = 20 // worst-path (depth × mean fan-out) above which the tree is flagged tangled
	// smallTree is the leaf count at or below which the whole tree is expanded regardless of the
	// Needs-you gate: a tree this small is easier to read whole than to reason about which branches
	// the gate opened, and it still fits on one screen.
	smallTree = 7
)

// node is one tree vertex. Leaves carry a row; internal nodes carry children keyed for stable
// ordering. `count` is the direct-child count shown on an internal node's line.
type node struct {
	key, label string
	row        *brief.Row // non-nil only on a leaf
	children   []*node
}

func (n *node) leaf() bool { return n.row != nil }

// Render returns the whole tree report: a complexity header line, then the tree. A depth violation
// (after any width regrouping) returns a one-line failure naming the offending path instead of a
// tree — the caller prints it as-is. `expandAll` forces every internal node open (`-tree=full`);
// otherwise a node opens only when a claim below it is in the brief's Needs-you set.
func Render(rows []brief.Row, expandAll bool, auditPath string) string {
	root := build(rows)
	regroupWide(root)

	if d, path := deepestLeaf(root, nil); d > maxDepth {
		return fmt.Sprintf("COMPLEXITY FAIL: depth %d exceeds %d on path %s\n",
			d, maxDepth, strings.Join(path, " › "))
	}

	// A tree of at most smallTree leaves is shown whole; the Needs-you gate only governs which
	// branches open once the tree is too large to read in full.
	if len(rows) <= smallTree {
		expandAll = true
	}

	needs := needsYou(rows)
	var b strings.Builder
	depth, width := deepest(root), widest(root)
	score := worstScore(root)
	status := "ok"
	if score > scoreWarn {
		status = fmt.Sprintf("tangled (>%d)", scoreWarn)
	}
	fmt.Fprintf(&b, "tree — depth %d, max width %d, score %.1f (target ≤%d) %s\n\n",
		depth, width, score, scoreWarn, status)

	// The root is always opened; without that nothing below it could ever show.
	for _, c := range root.children {
		render(&b, c, 1, needs, expandAll)
	}
	if auditPath != "" {
		fmt.Fprintf(&b, "\nFull table: %s\n", auditPath)
	}
	return b.String()
}

// Leaf carries the per-claim detail the HTML tree reveals when a leaf is opened: the verdict reason
// and the verbatim source spans, both taken from the verification chain and keyed by row ID.
type Leaf struct {
	Reason string
	Quotes []Quote
}

// Quote is one verified source span plus its already-resolved provenance suffix. `Prov` is built by
// the caller (which owns the manifest, so this package stays corpus-agnostic): either
// "— <doc>, <witness>, <locator>" for a resolved passage or "(passage <id>, unresolved)" when the
// passage id has no manifest entry. An empty Prov renders the bare quote, unchanged.
type Quote struct {
	Text string
	Prov string
}

const htmlHead = `<!doctype html>
<html lang="en"><head><meta charset="utf-8"><title>assay tree</title>
<style>body{font-family:monospace;margin:1.5rem}details{margin-left:2ch}summary{cursor:pointer}` +
	`.d{margin:.3em 0 .6em 2ch}.op{color:#888}.root{margin:0 0 1.5rem}</style>
</head><body>
`
const htmlTail = "</body></html>\n"

// RenderHTML renders the same tree as Render into a self-contained HTML page built from nested
// <details>/<summary> elements — no JavaScript. Every node is open by default; a leaf's body reveals
// the full claim text, the verdict reason, and the quoted source lines from `details` (keyed by row
// ID). `rootBlock` (from RootBlock) sits above the tree, matching the text output. Width regrouping
// matches Render; the depth ceiling is not enforced here (the text renderer is the gate). Styling is
// monospace only, per spec/CLI.md.
func RenderHTML(rows []brief.Row, details map[string]Leaf, header, rootBlock string) string {
	root := build(rows)
	regroupWide(root)
	var b strings.Builder
	b.WriteString(htmlHead)
	if header != "" {
		fmt.Fprintf(&b, "<pre>%s</pre>\n", html.EscapeString(header))
	}
	if rootBlock != "" {
		fmt.Fprintf(&b, "<pre class=\"root\">%s</pre>\n", html.EscapeString(rootBlock))
	}
	for _, c := range root.children {
		renderHTML(&b, c, details)
	}
	b.WriteString(htmlTail)
	return b.String()
}

// renderHTML writes one node as an open <details> block. An internal node's summary ends in its
// child count; a leaf's summary ends in its verdict, and its body carries the revealed detail.
func renderHTML(b *strings.Builder, n *node, details map[string]Leaf) {
	if n.leaf() {
		renderLeafHTML(b, n.row, details)
		return
	}
	fmt.Fprintf(b, "<details open><summary>%s  %s  [%d]</summary>\n",
		html.EscapeString(n.key), html.EscapeString(label12(n.label)), directCount(n))
	for _, c := range n.children {
		renderHTML(b, c, details)
	}
	b.WriteString("</details>\n")
}

// renderLeafHTML writes one atomic-claim leaf as an open <details> block: a summary ending in the
// verdict (grey for an opinion) and a body revealing the claim text, the judge's reason, and the
// quoted source lines from `details`. Both the section-path tree (renderHTML) and the argument tree
// (renderArgHTML in argument.go) call it, so a leaf renders identically in both — spec/ARGUMENT.md's
// "leaves unchanged" contract lives here.
func renderLeafHTML(b *strings.Builder, row *brief.Row, details map[string]Leaf) {
	r := *row
	verd := html.EscapeString(leafVerdict(r))
	if brief.IsOpinion(r) {
		verd = `<span class="op">` + verd + `</span>`
	}
	fmt.Fprintf(b, "<details open><summary>%s  %s  %s</summary>\n",
		html.EscapeString(r.ID), html.EscapeString(truncate(r.Text, 60)), verd)
	b.WriteString(`<div class="d">`)
	if r.SoWhat != "" {
		fmt.Fprintf(b, "so what: %s\n", html.EscapeString(r.SoWhat))
	}
	fmt.Fprintf(b, "claim: %s\n", html.EscapeString(r.Text))
	if r.Section != "" {
		fmt.Fprintf(b, "report: %s\n", html.EscapeString(r.Section)) // the report section this claim rests on
	}
	d := details[r.ID]
	if d.Reason != "" {
		fmt.Fprintf(b, "reason: %s\n", html.EscapeString(d.Reason))
	}
	for _, q := range d.Quotes {
		line := q.Text
		if q.Prov != "" {
			line += " " + q.Prov // "— <doc>, <witness>, <loc>" or "(passage <id>, unresolved)"
		}
		fmt.Fprintf(b, "&gt; %s\n", html.EscapeString(line))
	}
	b.WriteString("</div></details>\n")
}

// build groups rows under a synthetic root by their Path segments, in path order, with the row
// hanging as a leaf under its deepest segment. A row with no Path lands under a single "unplaced"
// node, per spec/CLI.md.
func build(rows []brief.Row) *node {
	root := &node{key: "·"}
	for i := range rows {
		r := rows[i]
		segs := parsePath(r.Path)
		cur := root
		for _, s := range segs {
			cur = child(cur, s.key, s.label)
		}
		leafKey := r.ID
		if leafKey == "" {
			leafKey = fmt.Sprintf("c%d", i+1)
		}
		cur.children = append(cur.children, &node{key: leafKey, row: &rows[i]})
	}
	sortTree(root)
	return root
}

type seg struct{ key, label string }

// parsePath splits a Path into ordered `key=label` segments. An empty Path yields the single
// "unplaced" segment; a segment with no "=" uses its text as both key and label.
func parsePath(path string) []seg {
	path = strings.TrimSpace(path)
	if path == "" {
		return []seg{{"unplaced", "unplaced"}}
	}
	var out []seg
	for _, part := range strings.Split(path, "/") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if k, label, ok := strings.Cut(part, "="); ok {
			out = append(out, seg{strings.TrimSpace(k), strings.TrimSpace(label)})
		} else {
			out = append(out, seg{part, part})
		}
	}
	if len(out) == 0 {
		return []seg{{"unplaced", "unplaced"}}
	}
	return out
}

// child returns the existing internal child with `key` under `parent`, or creates it. A later row
// may carry a fuller label for the same key; the first non-empty one wins so the label is stable.
func child(parent *node, key, label string) *node {
	for _, c := range parent.children {
		if !c.leaf() && c.key == key {
			if c.label == "" {
				c.label = label
			}
			return c
		}
	}
	n := &node{key: key, label: label}
	parent.children = append(parent.children, n)
	return n
}

// regroupWide enforces the width limit: an internal node with more than maxWidth children is split
// into runs of ≤ maxWidth in sorted order, each run wrapped in a synthetic node labelled by its
// key span (e.g. "F1–F7"). This adds a level, so it runs before the depth check — a regroup that
// pushes a leaf past maxDepth is then caught there, per spec/CLI.md.
func regroupWide(n *node) {
	for _, c := range n.children {
		if !c.leaf() {
			regroupWide(c)
		}
	}
	if len(n.children) <= maxWidth {
		return
	}
	var groups []*node
	for i := 0; i < len(n.children); i += maxWidth {
		end := min(i+maxWidth, len(n.children))
		run := n.children[i:end]
		span := run[0].key + "–" + run[len(run)-1].key
		groups = append(groups, &node{key: span, label: span, children: run})
	}
	n.children = groups
}

// render writes one node and, when it should open, its subtree. A leaf is one line ending in its
// verdict; an internal node is one line ending in its child count, expanded only when expandAll is
// set or a Needs-you claim sits below it.
func render(b *strings.Builder, n *node, depth int, needs map[string]bool, expandAll bool) {
	indent := strings.Repeat("  ", depth)
	if n.leaf() {
		fmt.Fprintf(b, "%s%s  %s  %s\n", indent, n.key, truncate(n.row.Text, 60), leafVerdict(*n.row))
		// A Needs-you leaf states its stakes: what a reader who believed the report would get wrong.
		if needs[n.key] && n.row.SoWhat != "" {
			fmt.Fprintf(b, "%s    ↳ so what: %s\n", indent, n.row.SoWhat)
		}
		return
	}
	fmt.Fprintf(b, "%s%s  %s  [%d]\n", indent, n.key, label12(n.label), directCount(n))
	if expandAll || opensFor(n, needs) {
		for _, c := range n.children {
			render(b, c, depth+1, needs, expandAll)
		}
	}
}

// opensFor reports whether any leaf below `n` is a Needs-you claim, which is the sole reason a node
// opens by default.
func opensFor(n *node, needs map[string]bool) bool {
	if n.leaf() {
		return needs[n.key]
	}
	for _, c := range n.children {
		if opensFor(c, needs) {
			return true
		}
	}
	return false
}

// needsYou is the set of row ids the brief would list under "Needs you"; the tree opens exactly the
// branches leading to them so the two reports never disagree.
func needsYou(rows []brief.Row) map[string]bool {
	set := make(map[string]bool)
	for _, s := range brief.Selected(rows) {
		set[s.Row.ID] = true
	}
	return set
}

// directCount counts a node's direct children; a synthetic width-group counts its own run, so the
// number on the line always matches what expansion would print beneath it.
func directCount(n *node) int { return len(n.children) }

func verdictOf(r brief.Row) string {
	switch {
	case r.Faith != "":
		return r.Faith
	case r.Grounding != "":
		return r.Grounding
	default:
		return r.Substance
	}
}

// leafVerdict is what a leaf line ends in: the verdict, plus the "k/N" agreement when a repeat run
// (-n>1) recorded a spread — e.g. "partial 2/3". A merged tie carries no majority verdict, so it ends
// in "split a/b" (the tied verdicts) instead. Both the text and HTML renderers use it so the two agree
// on what a split verdict looks like.
func leafVerdict(r brief.Row) string {
	if brief.IsOpinion(r) {
		return "opinion" // a Committee value judgment or recommendation: not judged, rendered grey in HTML
	}
	if r.Split != "" {
		return "split " + r.Split // a merged tie has no majority verdict — name the tied verdicts, not one
	}
	v := verdictOf(r)
	if r.Spread != "" {
		v += " " + r.Spread
	}
	if r.Dissent != "" {
		v += " ≠ " + r.Dissent // name the minority verdict(s) behind a split, e.g. "partial 2/3 ≠ faithful"
	}
	return v
}

// ── complexity metrics ──────────────────────────────────────────────────────

// deepest returns the maximum edge count from root to any leaf.
func deepest(n *node) int {
	if n.leaf() || len(n.children) == 0 {
		return 0
	}
	m := 0
	for _, c := range n.children {
		if d := deepest(c) + 1; d > m {
			m = d
		}
	}
	return m
}

// deepestLeaf returns the depth and the key path of the deepest leaf, for the depth-failure message.
func deepestLeaf(n *node, path []string) (int, []string) {
	if n.leaf() {
		return 0, append(path, n.key)
	}
	best, bestPath := 0, path
	for _, c := range n.children {
		here := path
		if n.key != "·" {
			here = append(append([]string{}, path...), n.key)
		}
		if d, p := deepestLeaf(c, here); d+1 > best {
			best, bestPath = d+1, p
		}
	}
	return best, bestPath
}

// widest returns the largest direct-child count of any internal node.
func widest(n *node) int {
	if n.leaf() {
		return 0
	}
	m := len(n.children)
	for _, c := range n.children {
		if w := widest(c); w > m {
			m = w
		}
	}
	return m
}

// worstScore is the largest, over all leaves, of (edges on the path) × (mean fan-out of the
// internal nodes on that path) — the spec's tangle metric. It flags a tree that is legal on width
// and depth yet still too bushy to read.
func worstScore(root *node) float64 {
	var worst float64
	var walk func(n *node, edges int, fanSum, fanN int)
	walk = func(n *node, edges int, fanSum, fanN int) {
		if n.leaf() {
			if fanN == 0 {
				return
			}
			mean := float64(fanSum) / float64(fanN)
			if s := float64(edges) * mean; s > worst {
				worst = s
			}
			return
		}
		for _, c := range n.children {
			walk(c, edges+1, fanSum+len(n.children), fanN+1)
		}
	}
	walk(root, 0, 0, 0)
	return worst
}

// ── formatting helpers ──────────────────────────────────────────────────────

func label12(s string) string {
	words := strings.Fields(s)
	if len(words) <= maxLabel {
		return s
	}
	return strings.Join(words[:maxLabel], " ") + "…"
}

func truncate(s string, n int) string {
	s = strings.ReplaceAll(strings.TrimSpace(s), "\n", " ")
	if len([]rune(s)) <= n {
		return s
	}
	return string([]rune(s)[:n-1]) + "…"
}

// sortTree orders every node's children: internal nodes before leaves is not imposed — instead all
// children sort by key with natural (digit-aware) order, so "2.2.5" < "2.2.10" and "F2" < "F17".
func sortTree(n *node) {
	sort.SliceStable(n.children, func(i, j int) bool {
		return natLess(n.children[i].key, n.children[j].key)
	})
	for _, c := range n.children {
		if !c.leaf() {
			sortTree(c)
		}
	}
}

// natLess compares two keys chunk by chunk, digit runs numerically and the rest lexically, giving
// the ordering a reader expects for section numbers and finding ids alike. Total, so it is a valid
// sort predicate.
func natLess(a, b string) bool {
	ac, bc := chunks(a), chunks(b)
	for i := 0; i < len(ac) && i < len(bc); i++ {
		x, y := ac[i], bc[i]
		xn, xok := toInt(x)
		yn, yok := toInt(y)
		switch {
		case xok && yok:
			if xn != yn {
				return xn < yn
			}
		case x != y:
			return x < y
		}
	}
	return len(ac) < len(bc)
}

// chunks splits a key into maximal runs of digits and non-digits: "F2.10a" → ["F", "2", ".", "10", "a"].
func chunks(s string) []string {
	var out []string
	var cur strings.Builder
	digit := false
	for i, r := range s {
		d := '0' <= r && r <= '9'
		if i > 0 && d != digit {
			out = append(out, cur.String())
			cur.Reset()
		}
		cur.WriteRune(r)
		digit = d
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}

func toInt(s string) (int, bool) {
	n, err := strconv.Atoi(s)
	return n, err == nil
}
