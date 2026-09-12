package adjudicate

// adjudicate_test.go pins the file parse and the agreement tally. TestAgreeThreeTwo is the stated
// refuter: three adjudications, two matching the machine's verdicts and one not, so N=3, Agreed=2 and
// the single disagreement keeps the human's reason. TestLoad pins the |-delimited line, the skipped `#`
// and blank lines, the page parse, that a missing file is not an error, and that a short line is.

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
	// A machine verdict the tree does not carry is a disagreement, not a silent agreement.
	if r := Agree(map[string]string{}, adjs[:1]); r.Agreed != 0 || len(r.Disagreements) != 1 || r.Disagreements[0].Machine != "(no verdict)" {
		t.Fatalf("missing machine verdict must disagree against (no verdict), got %+v", r)
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
