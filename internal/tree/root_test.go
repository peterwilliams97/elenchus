package tree

// root_test pins the root summary block line for line: the class partition, the ordering, the id
// caps, the Generalised spread ranking, and the value line. TestRootBlockRouteGates is its
// counterpart — it proves `route` gates the partition rather than decorating it, by mislabelling one
// evaluative record as source and watching a finding move between classes. spec/TREE.md § Root node
// rendering is the contract.

import (
	"strings"
	"testing"

	"assay/internal/brief"
)

func TestRootBlock(t *testing.T) {
	got := RootBlock(rootBlockRows(), RootMeta{Title: "Test Report", Date: "2025", SourceDocs: 60})
	if got != rootBlockGolden {
		t.Fatalf("RootBlock mismatch\n--- got ---\n%s\n--- want ---\n%s", got, rootBlockGolden)
	}
	if n := strings.Count(strings.TrimSpace(got), "\n") + 1; n > 21 {
		t.Errorf("block is %d lines; the class body must stay within 15 (≈21 with title, blanks, value line)", n)
	}
}

// TestRootBlockRouteGates flips one evaluative record (verdict absent) to route=source. The finding
// must leave "Committee opinions" and enter "Unsupported by any held source", changing that count from
// 3 to 4 — so the fourth id also collapses under the 3-id cap into "and 1 more".
func TestRootBlockRouteGates(t *testing.T) {
	rows := rootBlockRows()
	var flipped bool
	for i := range rows {
		if rows[i].ID == "F8" { // the evaluative-absent record
			rows[i].Route = "source"
			flipped = true
		}
	}
	if !flipped {
		t.Fatal("fixture drift: no F8 evaluative record to mislabel")
	}
	got := RootBlock(rows, RootMeta{Title: "Test Report", Date: "2025", SourceDocs: 60})
	if !strings.Contains(got, "Unsupported by any held source: 4") {
		t.Errorf("mislabelled route did not raise the Unsupported count to 4:\n%s", got)
	}
	if strings.Contains(got, "Committee opinions, not checked: 2") {
		t.Errorf("mislabelled route did not lower the Committee count from 2:\n%s", got)
	}
	if !strings.Contains(got, "and 1 more") {
		t.Errorf("the fourth Unsupported id should collapse to 'and 1 more':\n%s", got)
	}
}

// TestRootBlockContestedFirst is the stability-partition refuter. Two contested leaves — one that
// crossed the divide (modal faithful) and one that tied with no majority — must appear together in the
// top "Sources don't settle these" group and NOT in any verdict class: the crossing leaf's modal is
// faithful, so were it filed by verdict it would inflate Holds. The tie leaf must read "split a/b", not
// a verdict. The mutation clears the crossing leaf's Class, and it must then leave the top group and
// land in Holds — proving the partition is taken by stability first, ahead of the verdict.
func TestRootBlockContestedFirst(t *testing.T) {
	rows := []brief.Row{
		{ID: "F1", Path: "2=Ch2/2.1=§2.1", Text: "A settled faithful finding.", Faith: "faithful"},
		{ID: "F2", Path: "2=Ch2/2.1=§2.1", Text: "Runs cross the divide here.", Faith: "faithful",
			Class: "contested", Spread: "3/5", Dissent: "contradicted, absent"},
		{ID: "F3", Path: "2=Ch2/2.2=§2.2", Text: "Runs tie with no majority.", Faith: "partial",
			Class: "contested", Split: "faithful/partial"},
	}
	got := RootBlock(rows, RootMeta{Runs: 2})
	if !strings.Contains(got, "Sources don't settle these: 2") {
		t.Fatalf("contested leaves not filed under the top group:\n%s", got)
	}
	if !strings.Contains(got, "F2: faithful 3/5 ≠ contradicted, absent") {
		t.Errorf("crossing contested leaf should state its divide-crossing verdicts:\n%s", got)
	}
	if !strings.Contains(got, "F3: split faithful/partial") {
		t.Errorf("tie leaf should read 'split a/b', not a verdict:\n%s", got)
	}
	if !strings.Contains(got, "Holds: 1 findings say what their sources say.") {
		t.Errorf("only the settled F1 should hold; a contested leaf must be excluded from Holds:\n%s", got)
	}

	// Mutation: F2 is no longer contested. It must leave the top group and inflate Holds to 2.
	rows[1].Class, rows[1].Split = "", ""
	rows[2].Class, rows[2].Split = "", "" // keep the top group possible only via a real contested leaf
	got = RootBlock(rows, RootMeta{Runs: 2})
	if strings.Contains(got, "Sources don't settle these") {
		t.Errorf("no contested leaf remains; the top group must vanish:\n%s", got)
	}
	if !strings.Contains(got, "Holds: 2 findings say what their sources say.") {
		t.Errorf("clearing F2's contested class must return it to Holds (2):\n%s", got)
	}
}

// TestRootBlockSchemaFailure is the schema-gate refuter at the block. A leaf whose verdict was forced
// to unverifiable by a malformed judge reason (SchemaFail) must be filed under "Unverifiable, judge
// output malformed", flagged by id, and kept out of the verdict classes — even though its stored Faith
// still reads faithful. The mutation clears SchemaFail and restores faithful, and the leaf must then
// return to Holds, proving the flag gates the partition rather than decorating it.
func TestRootBlockSchemaFailure(t *testing.T) {
	rows := []brief.Row{
		{ID: "F1", Path: "2=Ch2/2.1=§2.1", Text: "A clean faithful finding.", Faith: "faithful"},
		{ID: "F2", Path: "2=Ch2/2.1=§2.1", Text: "Judge reason came back as a tag.", Faith: "unverifiable",
			SchemaFail: true},
	}
	got := RootBlock(rows, RootMeta{})
	if !strings.Contains(got, "Unverifiable, judge output malformed: 1") {
		t.Fatalf("schema-failed leaf not filed under its own class:\n%s", got)
	}
	if !strings.Contains(got, "F2: judge output malformed") {
		t.Errorf("schema-failed leaf should be flagged by id:\n%s", got)
	}
	if !strings.Contains(got, "Holds: 1 findings say what their sources say.") {
		t.Errorf("schema-failed leaf must be excluded from Holds:\n%s", got)
	}
	if strings.Contains(got, "Unverifiable, document not held") {
		t.Errorf("a schema failure is not a missing-document unverifiable:\n%s", got)
	}

	// Mutation: the reason was fine after all — clear the flag and restore the verdict.
	rows[1].SchemaFail, rows[1].Faith = false, "faithful"
	got = RootBlock(rows, RootMeta{})
	if strings.Contains(got, "judge output malformed") {
		t.Errorf("cleared schema flag must drop the malformed class:\n%s", got)
	}
	if !strings.Contains(got, "Holds: 2 findings say what their sources say.") {
		t.Errorf("clearing the flag must return the leaf to Holds (2):\n%s", got)
	}
}

// rootBlockRows is the 12-record fixture: two per multi-member class where useful, one evaluative
// record whose verdict is absent (the route-gate target) and one whose verdict is overstated (proof
// that opinion resolution overrides the verdict), and two Generalised partials at different spreads.
func rootBlockRows() []brief.Row {
	const c2a, c2b = "2=Ch2/2.1=§2.1", "2=Ch2/2.2=§2.2"
	const c3a, c3b, c3c = "3=Ch3/3.1=§3.1", "3=Ch3/3.2=§3.2", "3=Ch3/3.3=§3.3"
	return []brief.Row{
		{ID: "F1", Path: c2a, Text: "Baseline finding one.", Faith: "faithful"},
		{ID: "F2", Path: c2a, Text: "Baseline finding two.", Faith: "faithful"},
		{ID: "F3", Path: c2a, Text: "A directly reversed finding.", Faith: "contradicted", Route: "source",
			SoWhat: "Sources say the opposite of the claim."},
		{ID: "F4", Path: c2a, Text: "A finding inflated to 40% support.", Faith: "overstated", Route: "source",
			SoWhat: "The 40% figure overstates what the source supports."},
		{ID: "F5", Path: c2b, Text: "A finding no source states.", Faith: "unsupported", Route: "source",
			SoWhat: "No held source states this at all."},
		{ID: "F6", Path: c3a, Text: "A finding absent from its hearing.", Faith: "absent", Route: "source",
			SoWhat: "Cited hearing does not contain the claim."},
		{ID: "F7", Path: c3a, Text: "A finding absent from the dataset.", Faith: "absent", Route: "evidence",
			SoWhat: "Dataset does not back this figure."},
		{ID: "F8", Path: c3a, Text: "A Committee value judgment.", Faith: "absent", Route: "evaluative",
			SoWhat: "Not judged; an opinion."},
		{ID: "F9", Path: c3b, Text: "A Committee opinion, disappointing.", Faith: "overstated", Route: "evaluative",
			SoWhat: "Not judged; an opinion."},
		{ID: "F10", Path: c3b, Text: "A finding citing an unheld document.", Faith: "unverifiable", Route: "evidence",
			SoWhat: "fetch qon:abc/2025"},
		{ID: "F11", Path: c3c, Text: "Over 256 live music venues closed statewide.", Faith: "partial",
			Route: "source", Gap: "scope", Spread: "3/3", SoWhat: "Reader would overestimate the 256 closures statewide."},
		{ID: "F12", Path: c3c, Text: "Participation in New York exceeded Victoria.", Faith: "partial",
			Route: "source", Gap: "scope", Spread: "2/3", SoWhat: "Claim generalises a New York finding to Victoria."},
	}
}

const rootBlockGolden = `Test Report, 2025 — 12 findings checked against 60 source documents

Holds: 2 findings say what their sources say.
Contradicted by their own sources: 1
  F3: Sources say the opposite of the claim.
Overstated: 1
  F4: The 40% figure overstates what the source supports.
Unsupported by any held source: 3
  F5: No held source states this at all.
  F6: Cited hearing does not contain the claim.
  F7: Dataset does not back this figure.
Committee opinions, not checked: 2
Unverifiable, document not held: 1
  F10: qon:abc/2025
Generalised from narrower evidence: 2
  F11: Reader would overestimate the 256 closures statewide.
  F12: Claim generalises a New York finding to Victoria.

Changes 5 summary lines across 2 chapters.
`
