// SPDX-License-Identifier: Apache-2.0

package mutbench

import (
	"strings"
	"testing"
)

func scpClaim() Claim {
	return Claim{
		ID:       "S01",
		Class:    "SCP-1",
		Template: "The 40-millisecond p99 latency measured on {SCOPE} puts our service's p99 latency at 40 milliseconds.",
		Slots: map[string]Slot{
			"SCOPE": {Clean: "all 14 production replicas", Mutant: "the Frankfurt replica"},
		},
		Scopes: Scopes{Narrow: "the Frankfurt replica", Broad: "our service"},
	}
}

// Render substitutes the named slot value. The mutant must carry the narrow
// scope and the clean twin the wide one; a template left unsubstituted (a stray
// {SLOT}) is the failure that would silently ship a broken mutant.
func TestRenderSubstitutesSlot(t *testing.T) {
	c := scpClaim()

	mut, err := Render(c, Mutant)
	if err != nil {
		t.Fatalf("Render(mutant): %v", err)
	}
	want := "The 40-millisecond p99 latency measured on the Frankfurt replica puts our service's p99 latency at 40 milliseconds."
	if mut != want {
		t.Errorf("mutant text\n got: %q\nwant: %q", mut, want)
	}

	clean, err := Render(c, Clean)
	if err != nil {
		t.Fatalf("Render(clean): %v", err)
	}
	if !strings.Contains(clean, "all 14 production replicas") {
		t.Errorf("clean twin lost its wide scope: %q", clean)
	}
	if strings.Contains(mut, "{") || strings.Contains(clean, "{") {
		t.Errorf("unsubstituted slot survived: mutant=%q clean=%q", mut, clean)
	}
}

// A slot named in the template but absent from Slots must error, not render a
// literal "{SCOPE}" into a mutant that then gets graded as if it were prose.
func TestRenderMissingSlotIsError(t *testing.T) {
	c := scpClaim()
	c.Slots = map[string]Slot{}
	if _, err := Render(c, Mutant); err == nil {
		t.Fatal("want error for template slot with no value, got nil")
	}
}

// I3: same (fixture, class, seed, version) -> byte-identical mutant.
func TestGenerateIsDeterministic(t *testing.T) {
	claims := []Claim{scpClaim()}
	a, err := Generate(claims, "F-BASE-1", 17, "v0.1.0-test")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	b, err := Generate(claims, "F-BASE-1", 17, "v0.1.0-test")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(a) != 1 {
		t.Fatalf("want 1 mutant, got %d", len(a))
	}
	if a[0].MutantID != b[0].MutantID || a[0].MutantSHA256 != b[0].MutantSHA256 {
		t.Errorf("not deterministic:\n a=%+v\n b=%+v", a[0], b[0])
	}
	if a[0].MutantSHA256 == a[0].BaseSHA256 {
		t.Error("mutant and base hash identical: injection did nothing")
	}
}

// A2.1: E1-template mutants are tagged confound:arith and must never pool into
// the SCP-1 class number.
func TestGenerateCarriesConfoundTag(t *testing.T) {
	c := scpClaim()
	c.ID = "S11"
	c.Tags = []string{"confound:arith"}
	rows, err := Generate([]Claim{c}, "F-BASE-1", 1, "v")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(rows[0].Tags) != 1 || rows[0].Tags[0] != "confound:arith" {
		t.Errorf("tag lost in generation: %v", rows[0].Tags)
	}
	if Scorable(rows[0]) {
		t.Error("confound:arith row must be quarantined from the class number")
	}
}

// A3.2: axis-fired is per input — any axis emitted a non-empty critique on any
// claim derived from that input. Empty critique on every claim means not fired.
func TestAxisFiredPerInput(t *testing.T) {
	fired := []ChainRec{
		{Claim: "a", Detail: ChainDetail{CritiqueByAxis: nil}},
		{Claim: "b", Detail: ChainDetail{CritiqueByAxis: []Axis{{Axis: "Falsifiability", Severity: "fatal"}}}},
	}
	if !AxisFired(fired) {
		t.Error("want fired=true when any claim has a non-empty critique")
	}

	quiet := []ChainRec{
		{Claim: "a", Detail: ChainDetail{CritiqueByAxis: nil}},
		{Claim: "b", Detail: ChainDetail{CritiqueByAxis: []Axis{}}},
	}
	if AxisFired(quiet) {
		t.Error("want fired=false when no claim has a critique")
	}
	if AxisFired(nil) {
		t.Error("want fired=false for no chain records")
	}
}
