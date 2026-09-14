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

	"assay/internal/adjudicate"
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
// narrowed one (partial/overstated), a same-side wobble, or an `uncorroborated` leaf weakens; only
// faithful-and-settled holds. `uncorroborated` is a single-source corpus's remap of `absent` (assay.go
// rewrites it there): the report says it once and no second document repeats it — a weakening, not the
// grounding failure `absent` is when other sources were held and checked.
//
// The substance axis (spec/SUBSTANCE-CORPUS.md), when run, gates the one `holds` case: a faithful leaf
// holds only if substance also clears it (`substantive`, or absent — a faithfulness-only run is
// unchanged). A faithful-but-`hollow`/`partial` claim is `weakened` — the report copies its source, but
// the proposition does not survive the dialectic. Substance only ever lowers a leaf faithfulness left at
// `holds`; every worse faithfulness verdict already dominates, so it is not re-consulted there.
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
		if r.Substance == "hollow" || r.Substance == "partial" {
			return jWeakened
		}
		return jHolds
	case "partial", "overstated", "uncorroborated":
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

// isBase reports the descriptive `base` node — the one root child that is not a recommendation. Every
// other direct child of the root is a recommendation, whatever its id prefix: LCEIC numbers them
// `R1`–`R11`, Quocirca `S1`–`S7` / `B1`–`B5`, so the count must not hinge on an `R…` spelling.
func isBase(id string) bool {
	return id == "base"
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

// ArgumentTitle returns the report name from an argument.txt "# title:" header line — the name the
// site pages carry as their <title> and index.html as its <h1>. It returns "" when the file has no
// such line; the caller then falls back to the file name, never to the thesis, which stays the root
// block's alone (`root.Content`, rendered once by argRootBlock).
func ArgumentTitle(argText string) string {
	for _, ln := range strings.Split(argText, "\n") {
		t := strings.TrimSpace(ln)
		if !strings.HasPrefix(t, "#") {
			continue
		}
		body := strings.TrimSpace(strings.TrimPrefix(t, "#"))
		if k, v, ok := strings.Cut(body, ":"); ok && strings.EqualFold(strings.TrimSpace(k), "title") {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// ArgumentPage renders the argument tree as index.html (spec/SERVE.md § index.html) and, separately,
// returns the plain-text root block (argRootBlock) for the caller to print to stdout. The page is,
// top to bottom: the report title, the `what`/sources sentences, the thesis in a card with the
// recommendation tally beneath it, then one card per root child — the recommendations in the argument
// file's order, then the descriptive base — each opening down through its findings, claims, and the
// quotes those claims rest on. Every badge is DERIVED, never authored: a leaf's pooled verdict, an
// internal node's conjunction. `title` is the report name (ArgumentTitle); `what` is the
// corpus-specific "what this page checks / which sources are held" sentence the caller builds from the
// manifest — this package stays corpus-agnostic, so it takes the sentence rather than the held set.
// `adj` overlays a human's own leaf verdicts (spec/SERVE.md § Adjudications): it renders the human call
// beside the machine badge on each adjudicated leaf and, beneath the `what` sentence, the agreement
// count and the disagreements. nil when the corpus carries no adjudications — the page is then unchanged.
func ArgumentPage(root *ArgNode, details map[string]Leaf, title, what string, singleSource bool, adj *adjudicate.Overlay) (page, rootBlock string) {
	rootBlock = argRootBlock(root)
	var b strings.Builder
	fmt.Fprintf(&b, argHead, html.EscapeString(title))
	fmt.Fprintf(&b, "<main class=\"page\">\n<h1>%s</h1>\n", html.EscapeString(title))
	if what != "" {
		fmt.Fprintf(&b, "<p class=\"what\">%s</p>\n", html.EscapeString(what))
	}
	b.WriteString(adjStatHTML(adj))
	b.WriteString(`<button class="toggle" type="button">Expand all</button>` + "\n")
	// The thesis card is static — the root is not a toggle — so a reader meets the whole case (the
	// proposition, the recommendation tally, the base sentence) before opening any recommendation.
	b.WriteString(`<section class="thesis">` + "\n")
	fmt.Fprintf(&b, "<p class=\"prop\">%s</p>\n", html.EscapeString(root.Content))
	fmt.Fprintf(&b, "<p class=\"tally\">%s</p>\n", html.EscapeString(rootTally(root)))
	if s := baseSentence(root); s != "" {
		fmt.Fprintf(&b, "<p class=\"base\">%s</p>\n", html.EscapeString(s))
	}
	b.WriteString("</section>\n")
	for _, c := range root.Children {
		renderNodeCard(&b, c, details, adj)
	}
	b.WriteString(keyHTML(singleSource))
	b.WriteString("</main>\n")
	b.WriteString(argScript)
	b.WriteString("</body></html>\n")
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
		if isBase(c.ID) {
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
// recommendation rests on — as one sentence beneath the tally. It sums the direct children of the
// `base` root child, so a report with no base node yields "" (nothing to say).
func baseSentence(root *ArgNode) string {
	n := 0
	for _, c := range root.Children {
		if isBase(c.ID) {
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
// the base sentence, then one line per recommendation — every root child but `base` — its proposition
// and judgement, plus the child that decides it when it does not hold. It ends with a trailing newline.
func argRootBlock(root *ArgNode) string {
	var b strings.Builder
	fmt.Fprintln(&b, root.Content)
	fmt.Fprintln(&b, rootTally(root))
	if s := baseSentence(root); s != "" {
		fmt.Fprintln(&b, s)
	}
	b.WriteString("\n")
	for _, c := range root.Children {
		if isBase(c.ID) {
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

// renderNodeCard writes one internal argument node — a recommendation, a finding, or the descriptive
// base — as a closed <details> card: a badge for its DERIVED judgement, its id (with a trailing `?`
// on a `?` edge), its proposition, and, when it does not hold, the child that decides it in small
// text. An atomic-claim leaf renders through renderLeafCard.
func renderNodeCard(b *strings.Builder, n *ArgNode, details map[string]Leaf, adj *adjudicate.Overlay) {
	if n.row != nil {
		renderLeafCard(b, n.row, details, adj)
		return
	}
	j := n.Judgement()
	b.WriteString(`<details class="card"><summary>`)
	badgeSpan(b, j, j)
	b.WriteString(idSpan(n))
	fmt.Fprintf(b, `<span class="prop">%s</span>`, html.EscapeString(n.Content))
	if j != jHolds {
		if dc := decidingChild(n); dc != "" {
			fmt.Fprintf(b, `<span class="dc">(%s)</span>`, html.EscapeString(dc))
		}
	}
	b.WriteString("</summary>\n")
	for _, c := range n.Children {
		renderNodeCard(b, c, details, adj)
	}
	b.WriteString("</details>\n")
}

// renderLeafCard writes one atomic-claim leaf as a closed <details> card. Its badge is the pooled
// verdict, coloured on the severity scale (a contested/split leaf takes the contested hue); the
// summary carries the verdict word, the k/N agreement, the id and the claim head; the body reveals
// the full claim, the report section, the judge's reason, and the quotes the claim rests on — each
// quote's provenance kept in the exact "report: §… / … — <doc>, <witness>, <loc>" text renderReview
// (assay.go) rewrites into a PDF link.
func renderLeafCard(b *strings.Builder, row *brief.Row, details map[string]Leaf, adj *adjudicate.Overlay) {
	r := *row
	class, word, note := leafBadge(r)
	b.WriteString(`<details class="card leaf"><summary>`)
	badgeSpan(b, class, word)
	if note != "" {
		fmt.Fprintf(b, `<span class="note">%s</span>`, html.EscapeString(note))
	}
	// A human adjudication of this leaf sits beside the machine badge — both calls, no agreement colour
	// (the count is in the index). "PW: faithful", the initials then the verdict.
	if adj != nil {
		if a, ok := adj.By[r.ID]; ok {
			fmt.Fprintf(b, `<span class="human">%s: %s</span>`,
				html.EscapeString(a.Initials), html.EscapeString(a.Verdict))
		}
	}
	fmt.Fprintf(b, `<span class="id">%s</span><span class="claim">%s</span></summary>`+"\n",
		html.EscapeString(r.ID), html.EscapeString(truncate(r.Text, 80)))
	b.WriteString(`<div class="body">` + "\n")
	fmt.Fprintf(b, "<p class=\"claimfull\">%s</p>\n", html.EscapeString(r.Text))
	if r.SoWhat != "" {
		fmt.Fprintf(b, "<p class=\"meta\">so what: %s</p>\n", html.EscapeString(r.SoWhat))
	}
	if r.Section != "" {
		// "report: §X pN" verbatim — renderReview rewrites it into a link to the report PDF page.
		fmt.Fprintf(b, "<p class=\"meta\">report: %s</p>\n", html.EscapeString(r.Section))
	}
	d := details[r.ID]
	if d.Reason != "" {
		fmt.Fprintf(b, "<p class=\"meta\">reason: %s</p>\n", html.EscapeString(d.Reason))
	}
	for _, q := range d.Quotes {
		line := q.Text
		if q.Prov != "" {
			line += " " + q.Prov // "— <doc>, <witness>, <loc>": renderReview rewrites the provenance into a PDF link
		}
		fmt.Fprintf(b, "<blockquote>%s</blockquote>\n", html.EscapeString(line))
	}
	b.WriteString("</div></details>\n")
}

// idSpan writes a node's id, with a trailing `?` marker when the edge from its parent is a `?` edge —
// a linkage the report does not establish (spec/ARGUMENT.md § Edges).
func idSpan(n *ArgNode) string {
	if n.Query {
		return fmt.Sprintf(`<span class="id">%s<span class="q"> ?</span></span>`, html.EscapeString(n.ID))
	}
	return fmt.Sprintf(`<span class="id">%s</span>`, html.EscapeString(n.ID))
}

// badgeSpan writes a coloured badge. The class names the colour (b-holds … b-contested); the word is
// what the reader sees, so the colour is redundant with the text and a colour-blind reader loses
// nothing (spec/SERVE.md § Badges).
func badgeSpan(b *strings.Builder, class, word string) {
	fmt.Fprintf(b, `<span class="badge b-%s">%s</span>`, html.EscapeString(class), html.EscapeString(word))
}

// leafBadge maps one pooled leaf to its badge: the colour class, the verdict word shown, and a small
// note (the k/N agreement and any minority verdict). A contested or tied `split` leaf takes the
// contested hue whatever its modal verdict; an opinion is never judged. Otherwise the colour is the
// leaf's contribution to its parent (leafJudgement) so a claim and the finding above it agree on
// colour, while the word stays the verdict itself.
func leafBadge(r brief.Row) (class, word, note string) {
	if brief.IsOpinion(r) {
		return "opinion", "opinion", ""
	}
	if r.Split != "" {
		return "contested", "split", r.Split
	}
	note = strings.TrimSpace(r.Spread)
	if r.Dissent != "" {
		note = strings.TrimSpace(note + " ≠ " + r.Dissent)
	}
	if r.Class == "contested" {
		return "contested", verdictOf(r), note
	}
	return leafJudgement(&r), verdictOf(r), note
}

// argHead is the argument page's head through the opening <body>, with one %s for the report title.
// The CSS carries no per-cent sign so the whole block passes through fmt.Fprintf untouched. Layout
// (width, gutters) sits on the `.page` wrapper, not `body`, so renderReview can drop the body content
// into its right pane without the page's margins fighting the two-pane frame.
const argHead = `<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<title>%s</title>
<link rel="icon" href="favicon.svg" type="image/svg+xml">
<style>
:root{--holds:#009e73;--weak:#e69f00;--open:#0072b2;--fails:#d55e00;--opinion:#767676;--contested:#cc79a7}
*{box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,Helvetica,Arial,sans-serif;color:#1a1a1a;background:#fafafa;margin:0;line-height:1.5}
.page{max-width:60rem;margin:0 auto;padding:2rem 1.25rem 4rem}
h1{font-size:1.3rem;font-weight:600;line-height:1.3;margin:0 0 .6rem}
.what{color:#555;margin:0 0 1.25rem;max-width:46rem}
.adjstat{border:1px solid #e3e3e3;border-radius:8px;background:#fff;padding:.7rem 1rem;margin:0 0 1.25rem;font-size:.92rem}
.adjstat p{margin:0 0 .4rem;font-weight:500}
.adjstat ul{margin:0;padding-left:1.2rem;color:#555}
.adjstat li{margin:.15rem 0}
.human{font-size:.72rem;font-weight:600;letter-spacing:.02em;padding:.12rem .45rem;border-radius:999px;border:1px solid #999;color:#444;background:#fff;white-space:nowrap}
.thesis{border:1px solid #ddd;border-radius:8px;background:#fff;padding:1rem 1.15rem;margin:0 0 1.5rem}
.thesis .prop{font-size:1.05rem;font-weight:500;margin:0 0 .6rem}
.thesis .tally{margin:0 0 .3rem}
.thesis .base{color:#555;margin:0}
.toggle{font:inherit;font-size:.82rem;border:1px solid #ccc;border-radius:6px;background:#fff;padding:.28rem .7rem;cursor:pointer;margin:0 0 1rem}
details.card{border:1px solid #e3e3e3;border-radius:8px;background:#fff;margin:.5rem 0}
details.card>summary{cursor:pointer;list-style:none;padding:.55rem .8rem;display:flex;gap:.5rem;align-items:baseline;flex-wrap:wrap}
details.card>summary::-webkit-details-marker{display:none}
details.card[open]>summary{border-bottom:1px solid #eee}
details.card details.card{margin:.5rem .8rem}
.card .id{font-weight:600;white-space:nowrap}
.card .q{color:#999}
.card .prop{flex:1 1 18rem}
.card .dc{color:#777;font-size:.85rem;white-space:nowrap}
.leaf .body{padding:.5rem .85rem .85rem;color:#333;font-size:.95rem}
.leaf .claimfull{margin:.2rem 0 .5rem}
.leaf .meta{color:#555;margin:.2rem 0}
.leaf blockquote{margin:.5rem 0;padding:.25rem 0 .25rem .8rem;border-left:3px solid #ddd;color:#333}
.claim{flex:1 1 18rem}
.badge{font-size:.7rem;font-weight:700;letter-spacing:.03em;text-transform:uppercase;padding:.14rem .5rem;border-radius:999px;white-space:nowrap;color:#fff}
.b-holds{background:var(--holds)}
.b-weakened{background:var(--weak);color:#000}
.b-open{background:var(--open)}
.b-fails{background:var(--fails)}
.b-opinion{background:var(--opinion)}
.b-contested{background:var(--contested);color:#000}
.note{color:#777;font-size:.8rem;white-space:nowrap}
.key{border-top:1px solid #ddd;margin:2rem 0 0;padding-top:1.1rem;color:#555;font-size:.9rem;line-height:2}
.key .badge{margin-right:.35rem}
.nav{margin:1.1rem 0 0;font-size:.9rem}
.nav a,.key a{color:inherit}
</style>
</head><body>
`

// adjStatHTML renders the adjudication summary beneath the `what` sentence (spec/SERVE.md §
// Adjudications): "N leaves adjudicated, judge agreed on M", then one line per disagreement naming the
// human and machine verdicts and the human's reason. Empty when the corpus carries no adjudications, so
// the page is unchanged without a file.
func adjStatHTML(adj *adjudicate.Overlay) string {
	if adj == nil || adj.Result.N == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(`<section class="adjstat">` + "\n")
	fmt.Fprintf(&b, "<p>%d %s adjudicated, judge agreed on %d.</p>\n",
		adj.Result.N, plural(adj.Result.N, "leaf", "leaves"), adj.Result.Agreed)
	if len(adj.Result.Disagreements) > 0 {
		b.WriteString("<ul>\n")
		for _, d := range adj.Result.Disagreements {
			fmt.Fprintf(&b, "<li><b>%s</b> — human %s, judge %s: %s</li>\n",
				html.EscapeString(d.ID), html.EscapeString(d.Human),
				html.EscapeString(d.Machine), html.EscapeString(d.Reason))
		}
		b.WriteString("</ul>\n")
	}
	b.WriteString("</section>\n")
	return b.String()
}

// keyHTML is the verdict key and the two-pane link, at the foot of the page. It is generic to the
// argument tree — the verdict semantics, not this corpus — so it lives here rather than being passed
// in. Each line leads with the badge it defines, so the key doubles as the colour legend. On a
// single-source corpus (manifest `single_source: true`) it adds the `uncorroborated` line, the remap
// of a leaf `absent` when the report is its own only source: said once, not repeated elsewhere. The
// badge takes the weakened colour because that is how such a leaf derives (leafJudgement).
func keyHTML(singleSource bool) string {
	uncorroborated := ""
	if singleSource {
		uncorroborated = `<span class="badge b-weakened">uncorroborated</span> said once in the report, not repeated elsewhere.<br>` + "\n"
	}
	return `<div class="key">
<b>R</b> = recommendation, <b>F</b> = finding. Each badge shows its class as a word, so the colour is redundant.<br>
<span class="badge b-holds">holds</span> every claim underneath was found in a source saying what the report says.<br>
<span class="badge b-weakened">weakened</span> found, but the source says less.<br>
` + uncorroborated + `<span class="badge b-open">open</span> the report doesn't say what this rests on, or the sources don't settle a claim.<br>
<span class="badge b-fails">fails</span> a claim it rests on was not found in any held source.<br>
<span class="badge b-opinion">opinion</span> the Committee's own view; not checked.<br>
<span class="badge b-contested">contested</span> a claim the runs could not settle.
</div>
<p class="nav"><a href="review.html">Open the two-pane reading — the report on the left, this tree on the right</a> · <a href="sources/report.pdf">the report as published</a></p>
`
}

// argScript is the page's only JavaScript: it opens every card when the URL carries ?open=all, and
// wires the Expand-all button to open or close them all. One statement block — the page is otherwise
// static (spec/SERVE.md § index.html).
const argScript = `<script>
{const p=new URLSearchParams(location.search).get("open")==="all",d=document.querySelectorAll("details.card"),b=document.querySelector(".toggle");if(p)d.forEach(x=>{x.open=true});if(b)b.onclick=()=>{const o=[...d].some(x=>!x.open);d.forEach(x=>{x.open=o});b.textContent=o?"Collapse all":"Expand all"}}
</script>
`
