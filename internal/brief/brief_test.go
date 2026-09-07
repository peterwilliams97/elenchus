package brief

// brief_test pins the deterministic selection and rendering in brief.go: the tier rules in Qualify,
// the total order in Selected, the id ordering in idLess, and the shape of the Brief report
// (counts line, Needs-you rows, overflow line, trailing Full-table line). No model call reaches
// this package, so every case here is exact. Fixtures sit at the foot of the file.

import (
	"strings"
	"testing"
)

func TestQualifyTiers(t *testing.T) {
	for _, tc := range qualifyCases {
		gotTier, gotOK := Qualify(tc.faith, tc.substance, tc.grounding, tc.gap, tc.text)
		if gotOK != tc.ok || (tc.ok && gotTier != tc.tier) {
			t.Errorf("%s: Qualify(%q,%q,%q,%q,%q) = (%d,%v), want (%d,%v)",
				tc.name, tc.faith, tc.substance, tc.grounding, tc.gap, tc.text, gotTier, gotOK, tc.tier, tc.ok)
		}
	}
}

func TestSelectedOrder(t *testing.T) {
	rows := []Row{
		{ID: "F10", Text: "b", Grounding: "refuted"},                                    // tier 4
		{ID: "F2", Text: "a", Faith: "absent"},                                          // tier 0
		{ID: "F30", Text: "has 5 things", Faith: "overstated"},                          // tier 2 (number)
		{ID: "F3", Text: "faithful yet false", Faith: "faithful", Grounding: "refuted"}, // tier 1
		{ID: "F4", Text: "clean", Faith: "faithful"},                                    // no tier — dropped
		{ID: "F1", Text: "z", Faith: "contradicted"},                                    // tier 0
	}
	sels := Selected(rows)
	got := make([]string, len(sels))
	for i, s := range sels {
		got[i] = s.Row.ID
	}
	want := []string{"F1", "F2", "F3", "F30", "F10"} // tier asc, then id natural; F4 dropped
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("Selected order = %v, want %v", got, want)
	}
}

func TestBriefReport(t *testing.T) {
	rows := []Row{
		{ID: "F1", Text: "supported claim", Faith: "faithful"},
		{ID: "F2", Text: "not in the source at all", Faith: "absent", FaithReason: "no supporting quote found"},
	}
	out, err := Brief(rows, "1 faithful · 1 absent", "eval/x/audit.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"2 claims — 1 faithful · 1 absent",
		"Needs you (1):",
		"F2 | absent | not in the source at all | no supporting quote found",
		"Full table: eval/x/audit.md",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Brief output missing %q\n---\n%s", want, out)
		}
	}
}

func TestBriefOverflow(t *testing.T) {
	var rows []Row
	for _, id := range []string{"F1", "F2", "F3", "F4", "F5", "F6", "F7", "F8", "F9", "F10", "F11", "F12"} {
		rows = append(rows, Row{ID: id, Text: id, Faith: "absent"})
	}
	out, _ := Brief(rows, "12 absent", "audit.md")
	if !strings.Contains(out, "Needs you (12):") {
		t.Errorf("want total count 12 in header, got:\n%s", out)
	}
	if !strings.Contains(out, "and 2 more in audit.md") {
		t.Errorf("want overflow line for 12 selected, got:\n%s", out)
	}
	if strings.Count(out, "| absent |") != 10 {
		t.Errorf("want exactly 10 rows shown, got %d\n%s", strings.Count(out, "| absent |"), out)
	}
}

func TestIdLess(t *testing.T) {
	for _, tc := range idLessCases {
		got := idLess(tc.a, tc.b)
		if (got < 0) != tc.less || (got == 0) != tc.equal {
			t.Errorf("idLess(%q,%q) = %d, want less=%v equal=%v", tc.a, tc.b, got, tc.less, tc.equal)
		}
	}
}

// ── fixtures ────────────────────────────────────────────────────────────────

var qualifyCases = []struct {
	name                                   string
	faith, substance, grounding, gap, text string
	tier                                   int
	ok                                     bool
}{
	{"contradicted → a", "contradicted", "", "", "none", "x", 0, true},
	{"absent → a", "absent", "", "", "none", "x", 0, true},
	{"unsupported → a", "unsupported", "", "", "none", "x", 0, true}, // a positive verdict with no verified quote

	{"laundering → b", "faithful", "", "refuted", "none", "x", 1, true},
	{"overstated+number → c", "overstated", "", "", "none", "up 5%", 2, true},
	{"overstated no number → miss", "overstated", "", "", "none", "many", 0, false},
	// Tier e (docs/VALUE.md): a partial reaches the summary iff the judge marked a real gap. Each
	// gap class is a named case; adding a class means adding a line here, not editing the harness.
	{"partial denominator → e (F29)", "partial", "", "", "denominator", "fair share", 3, true},
	{"partial scope → e", "partial", "", "", "scope", "x", 3, true},
	{"partial timerange → e", "partial", "", "", "timerange", "x", 3, true},
	{"partial attribution → e", "partial", "", "", "attribution", "x", 3, true},
	{"partial other → e", "partial", "", "", "other", "x", 3, true},
	{"partial gap none → miss", "partial", "", "", "none", "x", 0, false},
	{"partial gap empty → miss", "partial", "", "", "", "x", 0, false},
	{"grounding refuted → d", "", "", "refuted", "none", "x", 4, true},
	{"clean faithful → miss", "faithful", "", "", "none", "x", 0, false},
	{"substance only → miss", "", "hollow", "", "none", "x", 0, false},
}

var idLessCases = []struct {
	a, b        string
	less, equal bool
}{
	{"F2", "F10", true, false},  // numeric, not lexical
	{"F2a", "F2b", true, false}, // suffix breaks the tie
	{"F2", "F2", false, true},   // equal
	{"F10", "F2", false, false}, // greater
}

// TestOpenedBranches pins the value line: the count of distinct top-level branches holding a
// Needs-you claim. Two chapters each hold one flagged claim; clean and unplaced rows do not count.
func TestOpenedBranches(t *testing.T) {
	rows := []Row{
		{ID: "F1", Path: "2=Chapter 2/2.1=x", Faith: "absent"},                      // needs, ch2
		{ID: "F2", Path: "2=Chapter 2/2.2=y", Faith: "faithful"},                    // clean
		{ID: "F3", Path: "3=Chapter 3/3.1=z", Faith: "partial", Gap: "denominator"}, // needs, ch3
		{ID: "F4", Faith: "faithful"},                                               // clean, unplaced
	}
	if n := OpenedBranches(rows); n != 2 {
		t.Errorf("OpenedBranches = %d, want 2 (ch2, ch3)", n)
	}
}
