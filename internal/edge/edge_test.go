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
	if ok, step := Admit(goodDefeater()); !ok {
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
		{"anchor not in world", func(d *Defeater) { d.Anchor = "a fintech startup" }, "anchor"},
		{"empty anchor", func(d *Defeater) { d.Anchor = "" }, "anchor"},
	}
	for _, tc := range cases {
		d := goodDefeater()
		tc.mut(&d)
		ok, step := Admit(d)
		if ok {
			t.Errorf("%s: expected rejection, was admitted", tc.name)
			continue
		}
		if step != tc.step {
			t.Errorf("%s: rejected at step %q, want %q", tc.name, step, tc.step)
		}
	}
}

// TestAdmitAnchorNormalised: a curly apostrophe or a line wrap between the anchor and its appearance in
// the world must not fail step 3 — the anchor check runs the same normalisation as the evidence-quote
// grounding, so present-despite-formatting still admits.
func TestAdmitAnchorNormalised(t *testing.T) {
	d := goodDefeater()
	d.World = "an org whose binding constraint is\ndelivery stability"
	d.Anchor = "binding constraint is delivery stability"
	if ok, step := Admit(d); !ok {
		t.Fatalf("wrapped anchor should still be found in the world, rejected at %q", step)
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
// exactly that scheme's five CQ slugs and no other's, carries the four defeater kinds, and forbids a
// "sound" field the axis boundary rules out.
func TestSchemaPinsFieldsAndCQs(t *testing.T) {
	raw, ok := Schema("practical")
	if !ok {
		t.Fatal("practical is a known scheme; Schema returned not-ok")
	}
	s := string(raw)
	for _, want := range []string{`"warrant"`, `"none_admitted"`, `"defeater"`, `"questions_considered"`,
		`"population"`, `"condition"`, `"definition"`, `"competing_goal"`,
		`"goal_held"`, `"alt_means"`, `"side_effects"`, `"feasible"`, `"goal_conflict"`} {
		if !strings.Contains(s, want) {
			t.Errorf("practical schema missing %s", want)
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

// goodDefeater is a fully-admissible defeater: dora's R-PLAT side-effect world, anchor quoted inside it.
func goodDefeater() Defeater {
	return Defeater{
		World:            "a regulated org whose binding constraint is delivery stability, where the platform's instability cost outweighs its performance gain",
		Kind:             "competing_goal",
		Anchor:           "binding constraint is delivery stability",
		Settles:          "F-PLAT's instability effect size against that org's own weighting",
		CriticalQuestion: "side_effects",
	}
}
