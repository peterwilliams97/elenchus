package tree

// tree_test pins the tree renderer's shape and its complexity rules: default expansion opens only
// the branches with a Needs-you claim below them, -tree=full opens everything, an over-wide node is
// regrouped into ≤7-child runs, and an over-deep path fails the render. natLess and parsePath are
// checked directly. Fixtures sit at the foot of the file.

import (
	"strings"
	"testing"

	"assay/internal/brief"
)

func TestRenderDefaultExpansion(t *testing.T) {
	// The Needs-you gate only governs trees larger than smallTree, so this fixture carries 8 leaves:
	// a clean 7-claim section that must collapse and a 1-claim section with an absent claim that
	// must open.
	var rows []brief.Row
	for _, id := range []string{"F1", "F2", "F3", "F4", "F5", "F6", "F7"} {
		rows = append(rows, brief.Row{ID: id, Path: "2=Chapter 2/2.2.1=Loss of audiences", Text: id, Faith: "faithful"})
	}
	rows = append(rows, brief.Row{ID: "F17", Path: "2=Chapter 2/2.2.5=Cost of living", Text: "ticket prices bar entry", Faith: "absent"})

	out := Render(rows, false, "")
	if !strings.Contains(out, "F17") {
		t.Errorf("Needs-you branch 2.2.5 should open and show F17:\n%s", out)
	}
	if strings.Contains(out, "F1  ") {
		t.Errorf("clean branch 2.2.1 should stay collapsed, but F1 leaf showed:\n%s", out)
	}
	if !strings.Contains(out, "2.2.1  Loss of audiences  [7]") {
		t.Errorf("collapsed 2.2.1 node line missing:\n%s", out)
	}
}

func TestSmallTreeExpandsAll(t *testing.T) {
	// Two faithful claims, neither in the Needs-you set: at or below smallTree the whole tree shows
	// anyway, so both leaves must appear without -tree=full.
	rows := []brief.Row{
		{ID: "F8", Path: "2=Chapter 2/2.2.1=Loss of audiences", Text: "training lost", Faith: "faithful"},
		{ID: "F12", Path: "2=Chapter 2/2.2.2=Mental health", Text: "worse mental health", Faith: "faithful"},
	}
	out := Render(rows, false, "")
	for _, id := range []string{"F8", "F12"} {
		if !strings.Contains(out, id) {
			t.Errorf("small tree (≤%d leaves) should show every leaf; %s missing:\n%s", smallTree, id, out)
		}
	}
}

func TestRenderExpandAll(t *testing.T) {
	out := Render(twoSectionRows(), true, "")
	for _, id := range []string{"F8", "F17"} {
		if !strings.Contains(out, id) {
			t.Errorf("expandAll should show every leaf; %s missing:\n%s", id, out)
		}
	}
}

func TestRenderComplexityHeader(t *testing.T) {
	out := Render(twoSectionRows(), true, "")
	if !strings.HasPrefix(out, "tree — depth 3, max width ") {
		t.Errorf("header should report depth 3:\n%s", out)
	}
	if !strings.Contains(out, "ok") {
		t.Errorf("small tree should be within the score target:\n%s", out)
	}
}

func TestWidthRegroup(t *testing.T) {
	var rows []brief.Row
	for _, id := range []string{"F1", "F2", "F3", "F4", "F5", "F6", "F7", "F8", "F9"} {
		rows = append(rows, brief.Row{ID: id, Path: "S=Section", Text: id, Faith: "faithful"})
	}
	out := Render(rows, true, "")
	if !strings.Contains(out, "F1–F7") || !strings.Contains(out, "F8–F9") {
		t.Errorf("9 children should regroup into F1–F7 and F8–F9 runs:\n%s", out)
	}
	if strings.Contains(out, "[8]") || strings.Contains(out, "[9]") {
		t.Errorf("no node may keep more than %d children after regroup:\n%s", maxWidth, out)
	}
}

func TestDepthFail(t *testing.T) {
	rows := []brief.Row{{ID: "F1", Path: "a=A/b=B/c=C/d=D", Text: "deep", Faith: "faithful"}}
	out := Render(rows, true, "")
	if !strings.HasPrefix(out, "COMPLEXITY FAIL:") {
		t.Errorf("a 4-segment path (leaf depth 5) must fail the depth check:\n%s", out)
	}
}

func TestParsePath(t *testing.T) {
	segs := parsePath("2=Chapter 2/2.2.5=Broader cost")
	if len(segs) != 2 || segs[0].key != "2" || segs[0].label != "Chapter 2" ||
		segs[1].key != "2.2.5" || segs[1].label != "Broader cost" {
		t.Errorf("parsePath keyed segments wrong: %+v", segs)
	}
	if p := parsePath(""); len(p) != 1 || p[0].key != "unplaced" {
		t.Errorf(`empty path should be one "unplaced" segment, got %+v`, p)
	}
}

func TestNatLess(t *testing.T) {
	for _, tc := range natCases {
		if got := natLess(tc.a, tc.b); got != tc.want {
			t.Errorf("natLess(%q,%q) = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}

// ── fixtures ────────────────────────────────────────────────────────────────

// twoSectionRows returns three claims: two under §2.2 (one clean, one absent) and one under §3.3,
// so default expansion, collapse, and the depth-3 header are all exercised by one tree.
func twoSectionRows() []brief.Row {
	return []brief.Row{
		{ID: "F8", Path: "2=Chapter 2/2.2.1=Loss of audiences", Text: "training lost", Faith: "faithful"},
		{ID: "F17", Path: "2=Chapter 2/2.2.5=Cost of living", Text: "ticket prices bar entry", Faith: "absent"},
		{ID: "F29", Path: "3=Chapter 3/3.3.5=Fair share", Text: "fair federal funding", Faith: "partial"},
	}
}

var natCases = []struct {
	a, b string
	want bool
}{
	{"2.2.5", "2.2.10", true}, // numeric, not lexical
	{"F2", "F17", true},       // finding numbers order numerically
	{"F17", "F2", false},      // and the reverse
	{"2.2.5", "2.2.5", false}, // equal is not less
	{"F1–F7", "F8–F9", true},  // regroup span labels order by first key
}

// TestRenderHTML pins the HTML tree: nested <details open>, a leaf whose summary carries the verdict
// and whose body reveals claim text, reason, and quoted source lines from the chain detail.
func TestRenderHTML(t *testing.T) {
	rows := []brief.Row{
		{ID: "F1", Path: "2=Chapter 2/2.2=Access", Text: "ticket prices bar entry", Faith: "overstated"},
	}
	details := map[string]Leaf{
		"F1": {Reason: "the source hedged this", Quotes: []string{"we saw some price sensitivity"}},
	}
	out := RenderHTML(rows, details, "model X · calls 0 · $0.0000", "")
	for _, want := range []string{
		"<details open>", "overstated", "claim: ticket prices bar entry",
		"reason: the source hedged this", "&gt; we saw some price sensitivity",
		"font-family:monospace", "model X",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("RenderHTML output missing %q\n---\n%s", want, out)
		}
	}
}

// TestSpreadInLeaf pins the -n spread rendering: a leaf with a "k/N" spread ends in "verdict k/N" in
// both the text and HTML trees.
func TestSpreadInLeaf(t *testing.T) {
	rows := []brief.Row{{ID: "F1", Text: "x", Faith: "partial", Spread: "2/3"}}
	if txt := Render(rows, true, ""); !strings.Contains(txt, "partial 2/3") {
		t.Errorf("text tree missing spread:\n%s", txt)
	}
	if h := RenderHTML(rows, map[string]Leaf{}, "", ""); !strings.Contains(h, "partial 2/3") {
		t.Errorf("html tree missing spread:\n%s", h)
	}
}

// TestDissentInLeaf pins that a split leaf names the minority verdict after its fraction — the -from
// dissent rendering — while a leaf with a spread but no recorded dissent stays back-compat unchanged.
func TestDissentInLeaf(t *testing.T) {
	split := []brief.Row{{ID: "F1", Text: "x", Faith: "partial", Spread: "2/3", Dissent: "faithful"}}
	if txt := Render(split, true, ""); !strings.Contains(txt, "partial 2/3 ≠ faithful") {
		t.Errorf("text tree missing dissent:\n%s", txt)
	}
	if h := RenderHTML(split, map[string]Leaf{}, "", ""); !strings.Contains(h, "partial 2/3 ≠ faithful") {
		t.Errorf("html tree missing dissent:\n%s", h)
	}
	bare := []brief.Row{{ID: "F1", Text: "x", Faith: "partial", Spread: "2/3"}}
	if txt := Render(bare, true, ""); strings.Contains(txt, "≠") {
		t.Errorf("no-dissent leaf grew a dissent marker:\n%s", txt)
	}
}

// TestSplitLeaf pins that a merged tie (Split set) ends its tree leaf line in "split a/b", not a modal
// verdict with a dissent marker — a tie has no majority to name, and the counterpart is TestDissentInLeaf
// (a non-tie split, which keeps the "modal k/N ≠ dissent" form). Both text and HTML renderers agree.
func TestSplitLeaf(t *testing.T) {
	tie := []brief.Row{{ID: "F1", Text: "x", Faith: "partial", Spread: "2/4",
		Dissent: "faithful", Split: "faithful/partial"}}
	if txt := Render(tie, true, ""); !strings.Contains(txt, "split faithful/partial") || strings.Contains(txt, "≠") {
		t.Errorf("text tie leaf should read 'split a/b' and carry no dissent marker:\n%s", txt)
	}
	if h := RenderHTML(tie, map[string]Leaf{}, "", ""); !strings.Contains(h, "split faithful/partial") {
		t.Errorf("html tie leaf missing split descriptor:\n%s", h)
	}
}

// TestSoWhatOnNeedsLeaf pins that a Needs-you leaf carries its one-line stakes in the text tree and
// that the HTML body shows it, while a clean leaf gets no so-what line.
func TestSoWhatOnNeedsLeaf(t *testing.T) {
	rows := []brief.Row{
		{ID: "F29", Path: "3=Ch3/3.3=Funding", Text: "fair share", Faith: "partial", Gap: "denominator",
			SoWhat: "Reader thinks Victoria funded fairly overall; true of only 3% of the money."},
		{ID: "F1", Path: "3=Ch3/3.3=Funding", Text: "clean", Faith: "faithful"},
	}
	txt := Render(rows, false, "")
	if !strings.Contains(txt, "↳ so what:") || !strings.Contains(txt, "3% of the money") {
		t.Errorf("needs-you leaf missing so-what:\n%s", txt)
	}
	if strings.Count(txt, "↳ so what:") != 1 {
		t.Errorf("so-what should appear once (only on the needs-you leaf):\n%s", txt)
	}
	if h := RenderHTML(rows, map[string]Leaf{}, "", ""); !strings.Contains(h, "so what:") {
		t.Errorf("html missing so-what:\n%s", h)
	}
}
