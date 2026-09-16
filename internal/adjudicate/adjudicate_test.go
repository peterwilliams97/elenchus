package adjudicate

// adjudicate_test.go pins the file parse and the agreement tally. TestAgreeThreeTwo is the stated
// refuter: three adjudications, two matching the machine's verdicts and one not, so N=3, Agreed=2 and
// the single disagreement keeps the human's reason; and a fourth adjudication whose id names no judged
// leaf is IGNORED — not counted in N, not a disagreement — pinning that the count covers only ids with a
// machine verdict, whether or not that leaf sits on the argument tree (spec/SERVE.md § Adjudications).
// TestLoad pins the |-delimited line, the skipped `#` and blank lines, the page parse, that a missing
// file is not an error, and that a short line is.

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAgreeThreeTwo(t *testing.T) {
	adjs := []Adjudication{
		{ID: "E1", Verdict: "faithful", Reason: "summary drops the country list"},
		{ID: "E21", Verdict: "contradicted", Reason: "print-specific relabelled as general"},
		{ID: "E3", Verdict: "overstated", Reason: "only stated once"},
	}
	machine := map[string]string{"E1": "faithful", "E21": "contradicted", "E3": "partial"}
	got := Agree(machine, adjs)
	if got.N != 3 || got.Agreed != 2 {
		t.Fatalf("Agree = N %d agreed %d, want 3/2", got.N, got.Agreed)
	}
	if len(got.Disagreements) != 1 || got.Disagreements[0].ID != "E3" {
		t.Fatalf("want one disagreement, on E3, got %+v", got.Disagreements)
	}
	if d := got.Disagreements[0]; d.Human != "overstated" || d.Machine != "partial" || d.Reason == "" {
		t.Fatalf("disagreement lost a field: %+v", d)
	}
	// An adjudication whose id has no machine verdict is gated out: not counted in N, not a disagreement.
	// (`machine` is every judged leaf's verdict.) GHOST is absent from machine, so N stays 3.
	gated := append(append([]Adjudication{}, adjs...),
		Adjudication{ID: "GHOST", Verdict: "faithful", Reason: "leaf split away since this was written"})
	if r := Agree(machine, gated); r.N != 3 || r.Agreed != 2 || len(r.Disagreements) != 1 || r.Disagreements[0].ID != "E3" {
		t.Fatalf("an adjudication of an unjudged id must be ignored: got %+v", r)
	}
	// A file whose ids name no judged leaf gates to nothing at all.
	if r := Agree(machine, []Adjudication{{ID: "GHOST", Verdict: "faithful"}}); r.N != 0 || len(r.Disagreements) != 0 {
		t.Fatalf("a file of only unjudged ids must gate to N=0 with no disagreements, got %+v", r)
	}
}

// TestAgreeSingleSourceSpelling is the stated refuter for the single-source fold: a claim the chain
// judged 'absent' agrees whether the human wrote the raw verdict ('absent') or the page's display
// spelling ('uncorroborated'), and a genuine mismatch ('faithful') still disagrees.
func TestAgreeSingleSourceSpelling(t *testing.T) {
	machine := map[string]string{"E1": "absent", "E2": "absent", "E3": "absent"}
	adjs := []Adjudication{
		{ID: "E1", Verdict: "absent"},
		{ID: "E2", Verdict: "uncorroborated"},
		{ID: "E3", Verdict: "faithful"},
	}
	got := Agree(machine, adjs)
	if got.N != 3 || got.Agreed != 2 {
		t.Fatalf("Agree = N %d agreed %d, want 3/2", got.N, got.Agreed)
	}
	if len(got.Disagreements) != 1 || got.Disagreements[0].ID != "E3" {
		t.Fatalf("want one disagreement, on E3, got %+v", got.Disagreements)
	}
}

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "adjudications.txt")
	body := "# a comment\n\n" +
		"E1 | faithful | Two-thirds (67%) of organisations | p2 | PW | drops the country list\n" +
		"E3 | overstated | the only sustainable response | 5 | PW | only stated once\n"
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	adjs, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(adjs) != 2 {
		t.Fatalf("want 2 adjudications, got %d", len(adjs))
	}
	if a := adjs[0]; a.ID != "E1" || a.Verdict != "faithful" || a.Page != 2 || a.Initials != "PW" || a.Reason != "drops the country list" {
		t.Fatalf("first line parsed wrong: %+v", a)
	}
	if adjs[1].Page != 5 { // a bare number parses too, not only "pN"
		t.Fatalf("second line page = %d, want 5", adjs[1].Page)
	}
	if none, err := Load(filepath.Join(dir, "nope.txt")); err != nil || none != nil {
		t.Fatalf("missing file: adjs=%v err=%v, want nil/nil", none, err)
	}
	bad := filepath.Join(dir, "bad.txt")
	if err := os.WriteFile(bad, []byte("E1 | faithful | q | p2 | PW\n"), 0o644); err != nil { // 5 fields
		t.Fatal(err)
	}
	if _, err := Load(bad); err == nil {
		t.Fatalf("a line without six fields must error")
	}
}
