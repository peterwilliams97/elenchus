package tree

// argument.go renders the report as an ARGUMENT tree (spec/ARGUMENT.md): a root proposition refined
// through recommendations and findings down to atomic-claim leaves, with every internal node's
// judgement DERIVED bottom-up from its children — never authored. It is the sister of the section-path
// Render in tree.go, and reuses that file's leaf rendering (leafVerdict, truncate, renderLeafHTML)
// unchanged, so a leaf looks identical in both trees.
//
// The judgement of an internal node is the WORST contribution of any child, over the ordering
// holds < weakened < open < fails (spec/ARGUMENT.md § Internal judgement). holds and fails come only
// from load-bearing children; open and weakened fire on any child. A `?` edge — a linkage the report
// does not establish — always contributes open: the reader must open a step the report never asserted.

import (
	"fmt"
	"html"
	"strings"

	"assay/internal/brief"
)

// ArgNode is one vertex of the argument tree. An atomic-claim leaf carries a `row` (its pooled
// verdict, matched by ID from the merged chain); every other node is internal and its judgement is
// derived from its children. `Query` marks the edge to this node's PARENT as a `?` edge — the report
// co-locates the two but does not state the parent rests on this child.
type ArgNode struct {
	ID       string
	Content  string // the proposition, authored, in the report's words; empty on a bare claim-leaf line
	Query    bool   // the edge from the parent to this node is a `?` edge (report does not establish it)
	Children []*ArgNode
	row      *brief.Row // non-nil only on an atomic-claim leaf, keyed by ID from the merged rows
}

// Judgement values, worst last. `opinion` is outside the ordering: an opinion is not evidence and
// contributes nothing to a parent (spec/ARGUMENT.md § Opinions contribute nothing).
const (
	jHolds    = "holds"
	jWeakened = "weakened"
	jOpen     = "open"
	jFails    = "fails"
	jOpinion  = "opinion"
)

// BuildArgument parses an argument.txt (spec/ARGUMENT.md line format) and hangs each merged row on
// its atomic-claim leaf by ID. It returns an error naming any claim-leaf id — a childless `F…` node —
// with no matching row, because a leaf the chain does not cover would silently render as an unsettled
// gap rather than the missing input it is.
func BuildArgument(argText string, rows []brief.Row) (*ArgNode, error) {
	root := parseArg(argText)
	if root == nil {
		return nil, fmt.Errorf("argument file has no root line")
	}
	byID := make(map[string]*brief.Row, len(rows))
	for i := range rows {
		byID[rows[i].ID] = &rows[i]
	}
	attachRows(root, byID)
	if missing := unmatchedLeaves(root); len(missing) > 0 {
		return nil, fmt.Errorf("argument leaves not in the chain: %s", strings.Join(missing, ", "))
	}
	return root, nil
}

// parseArg reads the argument tree by indentation: a node's parent is the nearest preceding line
// indented less than it. A line is `<id>[ ?]  | <content>  | <note>` (internal or attached leaf) or a
// bare `<id>` (an atomic-claim leaf). Blank lines and `#` comments are skipped; the note field is
// discarded here — it is provenance for a human, not part of the tree.
func parseArg(text string) *ArgNode {
	type frame struct {
		indent int
		node   *ArgNode
	}
	var root *ArgNode
	var stack []frame
	for _, raw := range strings.Split(text, "\n") {
		line := strings.TrimRight(raw, " \t")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		head := trimmed
		content := ""
		if i := strings.IndexByte(trimmed, '|'); i >= 0 {
			head = strings.TrimSpace(trimmed[:i])
			rest := trimmed[i+1:]
			if j := strings.IndexByte(rest, '|'); j >= 0 {
				rest = rest[:j]
			}
			content = strings.TrimSpace(rest)
		}
		fields := strings.Fields(head)
		if len(fields) == 0 {
			continue
		}
		n := &ArgNode{ID: fields[0], Content: content}
		for _, f := range fields[1:] {
			if f == "?" {
				n.Query = true
			}
		}
		for len(stack) > 0 && stack[len(stack)-1].indent >= indent {
			stack = stack[:len(stack)-1]
		}
		if len(stack) == 0 {
			root = n
		} else {
			p := stack[len(stack)-1].node
			p.Children = append(p.Children, n)
		}
		stack = append(stack, frame{indent, n})
	}
	return root
}

// attachRows binds a merged row to every childless node whose ID names a claim, leaving structural
// childless nodes (a recommendation with no attached finding, e.g. R6/R10) rowless — they derive as
// open, not as a missing leaf.
func attachRows(n *ArgNode, byID map[string]*brief.Row) {
	if len(n.Children) == 0 {
		if r, ok := byID[n.ID]; ok {
			n.row = r
		}
		return
	}
	for _, c := range n.Children {
		attachRows(c, byID)
	}
}

// unmatchedLeaves collects the ids of childless `F…` nodes (atomic-claim leaves by naming convention)
// that got no row — the claims the chain failed to cover. A rowless `R…` node is not a leaf, so it is
// not reported.
func unmatchedLeaves(n *ArgNode) []string {
	var out []string
	var walk func(*ArgNode)
	walk = func(n *ArgNode) {
		if len(n.Children) == 0 {
			if n.row == nil && (strings.HasPrefix(n.ID, "F") || strings.HasPrefix(n.ID, "f")) {
				out = append(out, n.ID)
			}
			return
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(n)
	return out
}

// Judgement derives one node's verdict, bottom-up. A leaf reads its pooled row; an internal node takes
// the worst contribution of any child (fails > open > weakened > holds). A `?` edge contributes open
// whatever the child says — the linkage itself is the open question. holds and fails reach the node
// only through load-bearing (stated) children; an opinion child on a stated edge contributes nothing.
// This conjunction is the judgement of a recommendation or finding node. The ROOT's displayed judgement
// is not this worst-case but a tally of its children (rootTally, spec/ARGUMENT.md § The root).
func (n *ArgNode) Judgement() string {
	if n.row != nil {
		return leafJudgement(n.row)
	}
	worst := -1
	live := false // any child that is not an ignored opinion — a node with none is opinion (or open if childless)
	for _, c := range n.Children {
		contrib := ""
		switch {
		case c.Query:
			contrib = jOpen // unestablished linkage: the reader must open the step, whatever the child holds
		case c.Judgement() == jOpinion:
			continue // opinion on a stated edge: not evidence, contributes nothing
		default:
			contrib = c.Judgement()
		}
		live = true
		if s := sev(contrib); s > worst {
			worst = s
		}
	}
	if !live {
		if len(n.Children) == 0 {
			return jOpen // a recommendation the report attaches to no finding: nothing settles it
		}
		return jOpinion // every child an opinion: the node rests only on opinions
	}
	return fromSev(worst)
}

// leafJudgement maps one pooled leaf to its contribution. An opinion is never judged. A contested (or
// tied `split`) leaf crosses the support divide across runs, so it opens its parent. unverifiable is
// "can't check" — neither support nor its refusal — so it opens too. A collapsed verdict fails; a
// narrowed one (partial/overstated) or a same-side wobble weakens; only faithful-and-settled holds.
func leafJudgement(r *brief.Row) string {
	if brief.IsOpinion(*r) {
		return jOpinion
	}
	if r.Class == "contested" {
		return jOpen
	}
	switch verdictOf(*r) {
	case "faithful":
		if r.Class == "wobble" {
			return jWeakened
		}
		return jHolds
	case "partial", "overstated":
		return jWeakened
	case "unverifiable":
		return jOpen
	case "contradicted", "unsupported", "absent":
		return jFails
	default:
		return jWeakened // an error/unknown verdict is not a clean hold nor a clear collapse
	}
}

func sev(j string) int {
	switch j {
	case jHolds:
		return 0
	case jWeakened:
		return 1
	case jOpen:
		return 2
	case jFails:
		return 3
	}
	return -1
}

func fromSev(s int) string {
	switch s {
	case 0:
		return jHolds
	case 1:
		return jWeakened
	case 3:
		return jFails
	default:
		return jOpen
	}
}

// isRecommendation reports an `R…`-numbered node — the recommendations the root block summarises, one
// line each. The atomic-claim leaves and findings are `F…`; the descriptive base is `base`.
func isRecommendation(id string) bool {
	return len(id) >= 2 && (id[0] == 'R' || id[0] == 'r') && '0' <= id[1] && id[1] <= '9'
}

// decidingChild names the child whose contribution set the node's judgement — the first child at the
// worst level — with a trailing "?" when it hangs on a `?` edge. It returns "" for a node that holds
// (nothing decides against it) or a childless recommendation (no child to point at).
func decidingChild(n *ArgNode) string {
	target := n.Judgement()
	if target == jHolds {
		return ""
	}
	for _, c := range n.Children {
		contrib := c.Judgement()
		if c.Query {
			contrib = jOpen
		}
		if contrib == target {
			if c.Query {
				return c.ID + " ?"
			}
			return c.ID
		}
	}
	return ""
}

// ArgumentPage returns the full argument.html and, separately, the plain-text root block it embeds
// (so the caller can also print it to stdout). The page is the root block above, then the findings
// and claims as a nested-<details> tree — the recommendations, their findings, and the atomic-claim
// leaves, the leaves rendered exactly as the section-path tree renders them.
func ArgumentPage(root *ArgNode, details map[string]Leaf, header string) (page, rootBlock string) {
	rootBlock = argRootBlock(root)
	var b strings.Builder
	b.WriteString(htmlHead)
	if header != "" {
		fmt.Fprintf(&b, "<pre>%s</pre>\n", html.EscapeString(header))
	}
	fmt.Fprintf(&b, "<pre class=\"root\">%s</pre>\n", html.EscapeString(rootBlock))
	for _, c := range root.Children {
		renderArgHTML(&b, c, details)
	}
	b.WriteString(htmlTail)
	return b.String(), rootBlock
}

// rootTally renders the root's judgement as a TALLY, not a conjunction (spec/ARGUMENT.md § The root):
// a prose sentence counting the RECOMMENDATIONS in each class, worst class last (holds, weakened,
// open, fails, then opinion outside the ordering). The descriptive `base` node — the findings no
// recommendation rests on — is excluded from the counts; baseSentence reports it separately, so the
// tally is a statement about the recommendations alone. Each recommendation is counted by its own
// derived Judgement (a `?` root edge does NOT reclassify it, that override is the conjunction rule the
// root does not run) and, where it does not hold, its reason comes from its deciding child: an open
// recommendation splits into those the report does not say what they rest on (a `?`-edge or childless
// deciding child) and those whose findings are contested (a stated child that opens); a failed one
// names the load-bearing child that collapsed it.
func rootTally(root *ArgNode) string {
	var holds, weakened, opinion, fails []string
	var openUnstated, openContested int
	recs := 0
	for _, c := range root.Children {
		if !isRecommendation(c.ID) {
			continue
		}
		recs++
		id := c.ID
		if c.Query {
			id += " ?"
		}
		switch c.Judgement() {
		case jHolds:
			holds = append(holds, id)
		case jWeakened:
			weakened = append(weakened, id)
		case jOpinion:
			opinion = append(opinion, id)
		case jFails:
			dc := strings.TrimSuffix(decidingChild(c), " ?")
			if dc == "" {
				fails = append(fails, c.ID)
			} else {
				fails = append(fails, fmt.Sprintf("%s: no held source supports %s", c.ID, dc))
			}
		case jOpen:
			// A `?`-edge or childless deciding child means the report never says what the
			// recommendation rests on; a stated deciding child that opens means its finding is contested.
			if dc := decidingChild(c); dc == "" || strings.HasSuffix(dc, " ?") {
				openUnstated++
			} else {
				openContested++
			}
		}
	}

	type clause struct {
		text   string
		isOpen bool
	}
	var clauses []clause
	// holds is always shown — the reference count — even at zero; the rest only when non-empty.
	clauses = append(clauses, clause{text: countClause(len(holds), verbHold(len(holds)), holds)})
	if len(weakened) > 0 {
		clauses = append(clauses, clause{text: countClause(len(weakened), isAre(len(weakened))+" weakened", weakened)})
	}
	if open := openUnstated + openContested; open > 0 {
		var reasons []string
		if openUnstated > 0 {
			reasons = append(reasons, fmt.Sprintf("%d because the report doesn't say what they rest on", openUnstated))
		}
		if openContested > 0 {
			reasons = append(reasons, fmt.Sprintf("%d because their findings are contested", openContested))
		}
		text := fmt.Sprintf("%d %s open", open, isAre(open))
		if len(reasons) > 0 {
			text += " — " + strings.Join(reasons, ", ")
		}
		clauses = append(clauses, clause{text: text, isOpen: true})
	}
	if len(fails) > 0 {
		clauses = append(clauses, clause{text: countClause(len(fails), verbFail(len(fails)), fails)})
	}
	if len(opinion) > 0 {
		clauses = append(clauses, clause{text: countClause(len(opinion), isAre(len(opinion))+" opinion", opinion)})
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Of %d %s,", recs, plural(recs, "recommendation", "recommendations"))
	for i, cl := range clauses {
		switch {
		case i == 0:
			b.WriteString(" ")
		case clauses[i-1].isOpen:
			b.WriteString(" — ") // the open aside closes on an em-dash before the next class
		default:
			b.WriteString(", ")
		}
		b.WriteString(cl.text)
	}
	b.WriteString(".")
	return b.String()
}

// countClause renders one verdict class as "<n> <verb> (<ids>)" — the id list dropped when empty (a
// zero-count holds clause reads "0 hold"). `verb` already agrees with `n` (verbHold/verbFail/isAre).
func countClause(n int, verb string, ids []string) string {
	if len(ids) == 0 {
		return fmt.Sprintf("%d %s", n, verb)
	}
	return fmt.Sprintf("%d %s (%s)", n, verb, strings.Join(ids, ", "))
}

func isAre(n int) string { return plural(n, "is", "are") }

// verbHold and verbFail conjugate "hold"/"fail" to the count: "1 holds", "2 hold"; "1 fails", "2 fail".
func verbHold(n int) string { return plural(n, "holds", "hold") }
func verbFail(n int) string { return plural(n, "fails", "fail") }

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// baseSentence counts the findings the descriptive `base` node carries — every finding no
// recommendation rests on — as one sentence beneath the tally. It sums the direct children of every
// non-recommendation root child, so a report with no base node yields "" (nothing to say).
func baseSentence(root *ArgNode) string {
	n := 0
	for _, c := range root.Children {
		if !isRecommendation(c.ID) {
			n += len(c.Children)
		}
	}
	if n == 0 {
		return ""
	}
	verb := plural(n, "finding supports", "findings support")
	return fmt.Sprintf("%d %s no recommendation.", n, verb)
}

// argRootBlock is the one screen above the tree: the root proposition, the root's tally paragraph and
// the base sentence, then one line per recommendation (R1–R11) — its proposition and judgement, plus
// the child that decides it when it does not hold. It ends with a trailing newline.
func argRootBlock(root *ArgNode) string {
	var b strings.Builder
	fmt.Fprintln(&b, root.Content)
	fmt.Fprintln(&b, rootTally(root))
	if s := baseSentence(root); s != "" {
		fmt.Fprintln(&b, s)
	}
	b.WriteString("\n")
	for _, c := range root.Children {
		if !isRecommendation(c.ID) {
			continue
		}
		j := c.Judgement()
		line := fmt.Sprintf("%s  %s  — %s", c.ID, c.Content, j)
		if j != jHolds {
			if dc := decidingChild(c); dc != "" {
				line += fmt.Sprintf(" (%s)", dc)
			}
		}
		fmt.Fprintln(&b, line)
	}
	return b.String()
}

// renderArgHTML writes one argument-tree node as an open <details> block. A leaf renders through the
// shared renderLeafHTML (identical to the section tree); an internal node's summary carries its id, a
// "?" when it hangs on a `?` edge, its content, its derived judgement, and its child count.
func renderArgHTML(b *strings.Builder, n *ArgNode, details map[string]Leaf) {
	if n.row != nil {
		renderLeafHTML(b, n.row, details)
		return
	}
	q := ""
	if n.Query {
		q = " ?"
	}
	fmt.Fprintf(b, "<details open><summary>%s%s  %s  [%s]  [%d]</summary>\n",
		html.EscapeString(n.ID), q, html.EscapeString(label12(n.Content)),
		html.EscapeString(n.Judgement()), len(n.Children))
	for _, c := range n.Children {
		renderArgHTML(b, c, details)
	}
	b.WriteString("</details>\n")
}
