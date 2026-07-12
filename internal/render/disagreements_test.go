// SPDX-License-Identifier: Apache-2.0

package render

import (
	"strings"
	"testing"
)

// The -disagreements rule (REVIEWER.md Q1 / decision d029) is deterministic: each
// audit column's verdict maps to a tier {PASS, FAIL, WEAK, NULL} via the spec's
// vcolor colour routing, and a claim surfaces iff at least one column is PASS and
// at least one is FAIL. Its entire behaviour is the 4^3 = 64 tier-triples plus the
// verdict->tier mapping over each mode's vocabulary. These tests pin both, with the
// triples constructed directly — no model is run, no audit pipeline is built.

// --- verdict -> tier mapping, over the full vocabulary of each mode -------------

func TestTierOf(t *testing.T) {
	cases := []struct {
		verdict string
		want    tier
	}{
		// faithfulness (5): faithful, partial, overstated, absent, contradicted
		{"faithful", tierPass},
		{"partial", tierWeak},
		{"overstated", tierWeak},
		{"absent", tierFail},
		{"contradicted", tierFail},
		// substance (3): substantive, partial, hollow
		{"substantive", tierPass},
		{"hollow", tierFail},
		// grounding (4): supported, mixed, refuted, unverifiable
		{"supported", tierPass},
		{"mixed", tierWeak},
		{"refuted", tierFail},
		{"unverifiable", tierNull},
		// non-verdicts that audit can still put in a column
		{"error", tierNull},
		{"skipped (over cap)", tierNull},
		// anything unrecognised falls to NULL (vcolor's "all others" grey bucket)
		{"banana", tierNull},
		{"", tierNull},
	}
	for _, c := range cases {
		if got := tierOf(c.verdict); got != c.want {
			t.Errorf("tierOf(%q) = %v, want %v", c.verdict, got, c.want)
		}
	}
}

// --- the surface rule, over all 64 tier-triples --------------------------------

func TestSurfaceAll64(t *testing.T) {
	allTiers := []tier{tierPass, tierFail, tierWeak, tierNull}

	// Independent oracle: a claim surfaces iff its tiers include both a PASS and a
	// FAIL. Written separately from surfaces() so a mutation to either is caught.
	oracle := func(ts [3]tier) bool {
		var hasPass, hasFail bool
		for _, x := range ts {
			switch x {
			case tierPass:
				hasPass = true
			case tierFail:
				hasFail = true
			}
		}
		return hasPass && hasFail
	}

	seen, surfaced := 0, 0
	for _, a := range allTiers {
		for _, b := range allTiers {
			for _, c := range allTiers {
				seen++
				want := oracle([3]tier{a, b, c})
				if want {
					surfaced++
				}
				if got := surfaces(a, b, c); got != want {
					t.Errorf("surfaces(%v,%v,%v) = %v, want %v", a, b, c, got, want)
				}
			}
		}
	}
	if seen != 64 {
		t.Fatalf("enumerated %d tier-triples, want 64", seen)
	}
	// Independent arithmetic check on the oracle itself: of 64 triples, those with
	// >=1 PASS and >=1 FAIL number 64 - |missing PASS or missing FAIL|
	// = 64 - (27 + 27 - 8) = 18.
	if surfaced != 18 {
		t.Fatalf("oracle surfaced %d of 64, want 18", surfaced)
	}
}

// Explicit truth-table anchors: a handful of triples with hand-written wants, so a
// shared bug in surfaces() and the oracle above cannot pass unnoticed.
func TestSurfaceAnchors(t *testing.T) {
	cases := []struct {
		a, b, c tier
		want    bool
	}{
		{tierPass, tierPass, tierPass, false}, // all pass: agree
		{tierFail, tierFail, tierFail, false}, // all fail: agree
		{tierWeak, tierWeak, tierWeak, false}, // all weak: unsettled
		{tierNull, tierNull, tierNull, false}, // all null: nothing said
		{tierPass, tierFail, tierPass, true},  // faithful/hollow/supported shape
		{tierPass, tierFail, tierWeak, true},  // pass+fail, weak third
		{tierPass, tierNull, tierFail, true},  // pass+fail, null third
		{tierFail, tierWeak, tierPass, true},  // order-independent
		{tierPass, tierWeak, tierNull, false}, // pass but no fail: not contested
		{tierFail, tierWeak, tierNull, false}, // fail but no pass: clear reject
	}
	for _, c := range cases {
		if got := surfaces(c.a, c.b, c.c); got != c.want {
			t.Errorf("surfaces(%v,%v,%v) = %v, want %v", c.a, c.b, c.c, got, c.want)
		}
	}
}

// --- the rule composed over real verdict strings -------------------------------

func TestDisagreeOverVerdicts(t *testing.T) {
	cases := []struct {
		faith, substance, grounding string
		want                        bool
	}{
		{"faithful", "hollow", "supported", true},       // speaker's noise: said it, holds up factually, but vacuous
		{"faithful", "substantive", "refuted", true},    // anti-signal: checkable and wrong
		{"absent", "hollow", "unverifiable", false},     // no pass anywhere: summarizer's noise
		{"overstated", "hollow", "refuted", false},      // no pass anywhere: clean reject
		{"faithful", "partial", "mixed", false},         // no fail anywhere: unsettled
		{"partial", "substantive", "supported", false},  // no fail anywhere
		{"contradicted", "substantive", "mixed", true},  // source says opposite, yet survives scrutiny
		{"faithful", "substantive", "supported", false}, // all pass: agree, no question
	}
	for _, c := range cases {
		got := disagree(AuditRow{Claim: "x", Faith: c.faith, Substance: c.substance, Grounding: c.grounding})
		if got != c.want {
			t.Errorf("disagree(%s/%s/%s) = %v, want %v", c.faith, c.substance, c.grounding, got, c.want)
		}
	}
}

// --- the rendered output -------------------------------------------------------

func TestDisagreementsRender(t *testing.T) {
	rows := []AuditRow{
		{Claim: "The future of work will happen inside Codex or Claude Code.", Faith: "partial", Substance: "partial", Grounding: "mixed"}, // dropped (no pass+fail)
		{Claim: "SaaS is not dead.", Faith: "faithful", Substance: "hollow", Grounding: "supported"},                                       // surfaced #2
		{Claim: "PMs will thrive in the AI era.", Faith: "faithful", Substance: "hollow", Grounding: "mixed"},                              // surfaced #3
	}
	out := Disagreements(rows)

	// The dropped row never appears.
	if strings.Contains(out, "future of work") {
		t.Errorf("non-disagreeing claim was rendered:\n%s", out)
	}
	// Original claim numbering is preserved (position in the full slice, 1-based).
	if !strings.Contains(out, "Claim 2 [faithful · hollow · supported]") {
		t.Errorf("missing/incorrect header for surfaced claim 2:\n%s", out)
	}
	if !strings.Contains(out, "Claim 3 [faithful · hollow · mixed]") {
		t.Errorf("missing/incorrect header for surfaced claim 3:\n%s", out)
	}
	// The conflict sentence: PASS clauses joined, then "but", then FAIL clauses.
	// Claim 2 is faithful + supported (two PASS) but hollow (FAIL).
	wantSentence := "The source really says it and the evidence backs it, but it is hollow under scrutiny."
	if !strings.Contains(out, wantSentence) {
		t.Errorf("missing conflict sentence for claim 2.\nwant substring: %q\ngot:\n%s", wantSentence, out)
	}
	// Claim 3 is faithful (one PASS) but hollow (FAIL); mixed grounding is WEAK and
	// contributes no clause.
	if !strings.Contains(out, "The source really says it, but it is hollow under scrutiny.") {
		t.Errorf("claim 3 sentence should omit the WEAK grounding clause:\n%s", out)
	}
	// The reviewer prompt is fixed and present once per surfaced claim.
	if n := strings.Count(out, "Keep it, qualify it, or cut it?"); n != 2 {
		t.Errorf("expected the reviewer prompt twice (one per surfaced claim), got %d:\n%s", n, out)
	}
	// The claim text is quoted.
	if !strings.Contains(out, `"SaaS is not dead."`) {
		t.Errorf("claim text not quoted:\n%s", out)
	}
}

// An audit run with no disagreements renders to the empty string (nothing to ask).
func TestDisagreementsEmpty(t *testing.T) {
	rows := []AuditRow{
		{Claim: "a", Faith: "faithful", Substance: "substantive", Grounding: "supported"},
		{Claim: "b", Faith: "absent", Substance: "hollow", Grounding: "refuted"},
	}
	if out := Disagreements(rows); out != "" {
		t.Errorf("expected empty output when no claim disagrees, got:\n%s", out)
	}
}
