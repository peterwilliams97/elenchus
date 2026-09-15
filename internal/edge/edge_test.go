package edge

// edge_test pins the code-side decisions of the edge pass (spec/EDGE.md §3–§4): the admission check
// rejects on each of its steps with a case that fails exactly that step and passes it otherwise; the
// template rule lifts an anchor shared across two edges to the root once and reduces both edges to
// unchallenged; and the per-scheme schema carries the fields and CQ enums the prompt promises. The
// model call is not exercised here — this package is the pure half. Fixtures sit at the foot.

import (
	"strings"
	"testing"
)

// TestAdmitRejectsEachStep drives Admit past a good defeater and then one mutation per step, asserting
// each mutation fails at, and only at, its own step (spec/EDGE.md §3). A rule the check stopped
// enforcing would let its mutation through — the point of a rejecting case per rule.
func TestAdmitRejectsEachStep(t *testing.T) {
	fq := findingAndQuotes()
	if ok, step := Admit(goodDefeater(), fq); !ok {
		t.Fatalf("the good defeater should be admitted, rejected at step %q", step)
	}
	cases := []struct {
		name string
		mut  func(*Defeater)
		step string
	}{
		{"empty world", func(d *Defeater) { d.World = "   " }, "world"},
		{"empty settles", func(d *Defeater) { d.Settles = "" }, "settles"},
		{"kind off the closed set", func(d *Defeater) { d.Kind = "vibes" }, "kind"},
		{"competing_goal is no longer a kind", func(d *Defeater) { d.Kind = "competing_goal" }, "kind"},
		{"anchor not in report", func(d *Defeater) { d.Anchor = "a fintech startup" }, "anchor"},
		{"empty anchor", func(d *Defeater) { d.Anchor = "" }, "anchor"},
	}
	for _, tc := range cases {
		d := goodDefeater()
		tc.mut(&d)
		ok, step := Admit(d, fq)
		if ok {
			t.Errorf("%s: expected rejection, was admitted", tc.name)
			continue
		}
		if step != tc.step {
			t.Errorf("%s: rejected at step %q, want %q", tc.name, step, tc.step)
		}
	}
}

// TestAdmitAnchorConfinedToThisEdge pins step 3 as amended for run 5 (spec/EDGE.md §3): the anchor is
// the cost clause, admissible only in THIS edge's own finding or a verified quote. It is rejected when
// drawn only from the recommendation — the run-3 defect (anchor in R while the world named something R
// never states) — and when drawn from ANOTHER finding's text — the run-4 defect (F-ACCESS anchored on
// F-DATA, a string present in the report but outside the premise under attack). Both reject at step
// "anchor", because Admit is given only this edge's finding and quotes; a cost clause in a verified
// quote of this edge admits.
func TestAdmitAnchorConfinedToThisEdge(t *testing.T) {
	fq := findingAndQuotes()

	inQuote := goodDefeater()
	inQuote.Anchor = "organizational performance" // present in this edge's verified quote
	if ok, step := Admit(inQuote, fq); !ok {
		t.Errorf("a cost-clause anchor drawn from this edge's verified quote should admit, rejected at %q", step)
	}

	inRecOnly := goodDefeater()
	inRecOnly.Anchor = "strategic prerequisite" // present only in the recommendation
	if !strings.Contains(norm(recText()), norm(inRecOnly.Anchor)) {
		t.Fatalf("fixture drift: %q is no longer in the recommendation", inRecOnly.Anchor)
	}
	if ok, step := Admit(inRecOnly, fq); ok || step != "anchor" {
		t.Errorf("an anchor drawn only from the recommendation should be rejected at anchor, got ok=%v step=%q", ok, step)
	}

	inOtherFinding := goodDefeater()
	inOtherFinding.Anchor = "broad data access" // present only in another finding's text
	if !strings.Contains(norm(otherFindingText()), norm(inOtherFinding.Anchor)) {
		t.Fatalf("fixture drift: %q is no longer in the other finding's text", inOtherFinding.Anchor)
	}
	if ok, step := Admit(inOtherFinding, fq); ok || step != "anchor" {
		t.Errorf("an anchor drawn from another finding's text should be rejected at anchor, got ok=%v step=%q", ok, step)
	}
}

// TestAdmitAnchorNormalised: a curly apostrophe or a line wrap between the anchor and its appearance in
// this edge's finding must not fail step 3 — the anchor check runs the same normalisation as the
// evidence-quote grounding, so present-despite-formatting still admits.
func TestAdmitAnchorNormalised(t *testing.T) {
	d := goodDefeater()
	d.Anchor = "binding constraint is delivery stability"
	fq := []string{"an org whose binding constraint is\ndelivery stability, per §2"}
	if ok, step := Admit(d, fq); !ok {
		t.Fatalf("wrapped anchor should still be found in this edge's finding, rejected at %q", step)
	}
}

// TestResolveTemplateRule: two edges whose admitted defeaters share an anchor (case-insensitively) are
// method-level — one MethodNote at the root, both edges reduced to unchallenged (spec/EDGE.md §3 rule
// 4). This is dora's correlation-as-cause: a world that defeats every associational edge belongs at the
// root once, not on each edge, where per-edge it would be the free attack the admission check exists to
// catch.
func TestResolveTemplateRule(t *testing.T) {
	shared := Defeater{World: "AI-mature orgs adopt a stance because already high-performing (common cause)",
		Kind: "condition", Anchor: "common cause", Settles: "whether the number is observational or an intervention estimate"}
	shared2 := shared
	shared2.Anchor = "Common Cause" // same anchor, different case — the template rule folds it
	edges := []EdgeResult{
		{FindingID: "F-STANCE", RecID: "R-STANCE", Verdict: Open, Defeater: shared},
		{FindingID: "F-DATA", RecID: "R-DATA", Verdict: Open, Defeater: shared2},
	}
	verdict, world, methods := Resolve(edges)
	if len(methods) != 1 {
		t.Fatalf("shared anchor should lift ONE method note, got %d", len(methods))
	}
	if got := methods[0].Edges; len(got) != 2 {
		t.Fatalf("the method note should name both edges, got %v", got)
	}
	for _, id := range []string{"F-STANCE", "F-DATA"} {
		if verdict[id] != Unchallenged {
			t.Errorf("%s should be unchallenged after the lift, got %q", id, verdict[id])
		}
		if world[id] != "" {
			t.Errorf("%s should carry no per-edge world after the lift, got %q", id, world[id])
		}
	}
}

// TestResolveDistinctAnchorsStay: two Open edges with DIFFERENT anchors are edge-level, not method — no
// lift, both stay open, each keeps its own defeater world. The mutation half of the template-rule test:
// it is the SHARED anchor, not merely two open edges, that triggers the lift.
func TestResolveDistinctAnchorsStay(t *testing.T) {
	edges := []EdgeResult{
		{FindingID: "F1", RecID: "R1", Verdict: Open, Defeater: Defeater{World: "a world naming widget latency", Anchor: "widget latency"}},
		{FindingID: "F2", RecID: "R2", Verdict: Open, Defeater: Defeater{World: "a world naming staffing gaps", Anchor: "staffing gaps"}},
	}
	verdict, world, methods := Resolve(edges)
	if len(methods) != 0 {
		t.Fatalf("distinct anchors should lift no method note, got %d", len(methods))
	}
	for _, id := range []string{"F1", "F2"} {
		if verdict[id] != Open {
			t.Errorf("%s should stay open, got %q", id, verdict[id])
		}
		if world[id] == "" {
			t.Errorf("%s should keep its per-edge world", id)
		}
	}
}

// TestSchemaPinsFieldsAndCQs pins the per-scheme schema (spec/EDGE.md §2), the edge-pass counterpart of
// the substance/faithfulness prompt-presence tests: the practical schema constrains critical_question to
// exactly its single CQ slug (side_effects) and no other's, carries the three defeater kinds
// (population, condition, definition), and forbids a "sound" field the axis boundary rules out.
func TestSchemaPinsFieldsAndCQs(t *testing.T) {
	raw, ok := Schema("practical")
	if !ok {
		t.Fatal("practical is a known scheme; Schema returned not-ok")
	}
	s := string(raw)
	for _, want := range []string{`"warrant"`, `"none_admitted"`, `"defeater"`, `"questions_considered"`,
		`"population"`, `"condition"`, `"definition"`, `"side_effects"`} {
		if !strings.Contains(s, want) {
			t.Errorf("practical schema missing %s", want)
		}
	}
	// The removed items must be gone: competing_goal is no longer a kind (spec/EDGE.md §3 step 2); the
	// goal-naming CQs are no longer offered (§2, §3 step 5); alt_means was removed in run 4 (§2); and
	// means_mismatch — with its anchor_rec field — was removed in run 5 (§2).
	for _, gone := range []string{`"competing_goal"`, `"goal_held"`, `"feasible"`, `"goal_conflict"`,
		`"alt_means"`, `"means_mismatch"`, `"anchor_rec"`} {
		if strings.Contains(s, gone) {
			t.Errorf("practical schema still carries removed value %s", gone)
		}
	}
	// A CQ slug from another scheme must not leak into this one's enum.
	if strings.Contains(s, "typical_case") {
		t.Error("practical schema leaked the example scheme's CQ slug")
	}
	if strings.Contains(strings.ToLower(s), "sound") {
		t.Error("schema must have no field asserting the inference is sound (axis boundary)")
	}
	if _, ok := Schema("nope"); ok {
		t.Error("an unknown scheme must return not-ok")
	}
}

// goodDefeater is a fully-admissible side_effects defeater: dora's R-PLAT cost world, its anchor the cost
// clause the finding names ("delivery instability", see findingAndQuotes). Admission checks this edge's
// finding and quotes, not the recommendation, not another finding, and not the world. Kind is `condition`
// (a state under which the instability cost bites); `competing_goal` is no longer a kind — a world that
// swaps the audience's goals is out of the report's domain (spec/EDGE.md §3 step 5).
func goodDefeater() Defeater {
	return Defeater{
		World:            "a state where the platform's instability cost, under load, outweighs its performance gain",
		Kind:             "condition",
		Anchor:           "delivery instability",
		Settles:          "F-PLAT's instability effect size against that org's own weighting",
		CriticalQuestion: "side_effects",
	}
}

// findingAndQuotes is the finding text and one verified quote `anchor` must land in (spec/EDGE.md §3
// step 3) — the R-PLAT edge's premise. It excludes the recommendation (recText) and every other finding
// (otherFindingText), the two texts `anchor` may NOT draw from under run 5's this-edge confinement.
func findingAndQuotes() []string {
	return []string{
		"A quality internal platform amplifies performance at the cost of a small increase in delivery instability.",
		"a quality internal platform amplifies AI's positive influence on organizational performance",
	}
}

// recText is the recommendation the admission check refuses `anchor` from — a phrase present only here,
// never in the R-PLAT edge's finding or quotes, is rejected at step "anchor".
func recText() string {
	return "Invest in your internal platform — treat it as the strategic prerequisite for unlocking AI's value."
}

// otherFindingText is a DIFFERENT edge's finding (dora's F-DATA), the source of the run-4 leak: an anchor
// verbatim in it but absent from the R-PLAT premise ("broad data access") must be rejected at "anchor".
func otherFindingText() string {
	return "Broad data access amplifies AI's positive influence on organizational performance."
}
