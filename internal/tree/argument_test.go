package tree

// argument_test pins the argument tree's derivation rule (spec/ARGUMENT.md § Internal judgement) and
// the `?`-edge override the CLI adds on top of it. The three refuters are: one contested leaf opens
// the root; a `?` edge opens the recommendation it hangs under; an all-holds tree holds to the root.
// Two further tests pin opinion propagation and the rendered page shape. Fixtures sit at the foot.

import (
	"strings"
	"testing"

	"assay/internal/brief"
)

// TestArgContestedOpensRoot: a single contested leaf, on load-bearing edges all the way up, opens the
// root. The mutation — the same leaf settled-and-faithful — must return the root to holds, proving the
// contested class, not the fixture shape, is what opens it.
func TestArgContestedOpensRoot(t *testing.T) {
	arg := joinLines(
		"root  | Root proposition.  | x",
		"    R1  | Recommendation one.  | x",
		"        F1  | Finding one.  | x",
		"            F1",
	)
	root := mustBuild(t, arg, []brief.Row{{ID: "F1", Text: "f1", Faith: "faithful", Class: "contested"}})
	if got := root.Judgement(); got != jOpen {
		t.Fatalf("contested leaf should open the root, got %q", got)
	}
	settled := mustBuild(t, arg, []brief.Row{{ID: "F1", Text: "f1", Faith: "faithful"}})
	if got := settled.Judgement(); got != jHolds {
		t.Fatalf("mutation to a settled faithful leaf should hold, got %q", got)
	}
}

// TestArgQueryEdgeOpensRecommendation: a `?` edge opens its parent whatever the child holds — R4 rests
// load-bearing on F30 (which holds) and co-locates F29 on a `?` edge, so R4 is open, and the deciding
// child names the `?` companion. Dropping the `?` (both edges stated) must return R4 to holds.
func TestArgQueryEdgeOpensRecommendation(t *testing.T) {
	rows := []brief.Row{
		{ID: "F30", Text: "f30", Faith: "faithful"},
		{ID: "F29", Text: "f29", Faith: "faithful"},
	}
	withQuery := mustBuild(t, joinLines(
		"root  | Root.  | x",
		"    R4  | Recommendation four.  | x",
		"        F30  | Load-bearing finding.  | x",
		"            F30",
		"        F29 ?  | Companion, rest-on unstated.  | x",
		"            F29",
	), rows)
	r4 := childByID(withQuery, "R4")
	if got := r4.Judgement(); got != jOpen {
		t.Fatalf("a `?` edge should open the recommendation, got %q", got)
	}
	if dc := decidingChild(r4); dc != "F29 ?" {
		t.Fatalf("deciding child of the open R4 should be the `?` companion, got %q", dc)
	}

	stated := mustBuild(t, joinLines(
		"root  | Root.  | x",
		"    R4  | Recommendation four.  | x",
		"        F30  | Load-bearing finding.  | x",
		"            F30",
		"        F29  | Companion, now stated.  | x",
		"            F29",
	), rows)
	if got := childByID(stated, "R4").Judgement(); got != jHolds {
		t.Fatalf("with both edges stated and both leaves faithful, R4 should hold, got %q", got)
	}
}

// TestArgAllHoldsRoot: every leaf faithful-and-settled on load-bearing edges holds all the way to the
// root — the baseline the other two refuters mutate away from.
func TestArgAllHoldsRoot(t *testing.T) {
	root := mustBuild(t, joinLines(
		"root  | Root.  | x",
		"    R1  | Rec one.  | x",
		"        F1  | Finding one.  | x",
		"            F1",
		"    R2  | Rec two.  | x",
		"        F2  | Finding two.  | x",
		"            F2",
	), []brief.Row{
		{ID: "F1", Text: "f1", Faith: "faithful"},
		{ID: "F2", Text: "f2", Faith: "faithful"},
	})
	if got := root.Judgement(); got != jHolds {
		t.Fatalf("all-holds tree should hold at the root, got %q", got)
	}
}

// TestArgOpinionOnlyRecommendation: a recommendation resting only on an evaluative (opinion) finding is
// itself opinion — faithfulness cannot settle it — and contributes nothing to the root, which still
// holds on its other, evidential recommendation.
func TestArgOpinionOnlyRecommendation(t *testing.T) {
	root := mustBuild(t, joinLines(
		"root  | Root.  | x",
		"    R9  | Opinion-only rec.  | x",
		"        F43  | Committee value judgement.  | [opinion]",
		"            F43",
		"    R1  | Evidential rec.  | x",
		"        F1  | Finding.  | x",
		"            F1",
	), []brief.Row{
		{ID: "F43", Text: "f43", Faith: "absent", Route: "evaluative"},
		{ID: "F1", Text: "f1", Faith: "faithful"},
	})
	if got := childByID(root, "R9").Judgement(); got != jOpinion {
		t.Fatalf("a rec resting only on an opinion should be opinion, got %q", got)
	}
	if got := root.Judgement(); got != jHolds {
		t.Fatalf("root should hold on R1 with the opinion R9 contributing nothing, got %q", got)
	}
}

// TestArgumentPage pins the rendered page: the root block leads with the proposition and the root
// TALLY (not a single verdict), prints one line per recommendation with its judgement (and the
// deciding child when it is not holds), and the tree below carries the atomic-claim leaf rendered by
// the shared leaf renderer.
func TestArgumentPage(t *testing.T) {
	root := mustBuild(t, joinLines(
		"root  | Victoria's industries matter and the Government should act.  | SYNTHESISED",
		"    R4  | Advocate release of the breakdowns.  | x",
		"        F30  | Breakdowns not released.  | x",
		"            F30",
		"        F29 ?  | Companion.  | x",
		"            F29",
	), []brief.Row{
		{ID: "F30", Text: "breakdowns not released", Faith: "faithful"},
		{ID: "F29", Text: "fair share received", Faith: "faithful"},
	})
	page, rootBlock := ArgumentPage(root, map[string]Leaf{"F30": {Reason: "the source says so"}}, "model X · calls 0")
	for _, want := range []string{
		"Victoria's industries matter",
		"Of 1 recommendation, 0 hold, 1 is open — 1 because the report doesn't say what they rest on.",
		"R4  Advocate release of the breakdowns.  — open (F29 ?)",
	} {
		if !strings.Contains(rootBlock, want) {
			t.Errorf("root block missing %q:\n%s", want, rootBlock)
		}
	}
	for _, want := range []string{
		"<pre class=\"root\">", "<details open>", "breakdowns not released",
		"claim: fair share received", "reason: the source says so", "[open]", "F29 ?",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("page missing %q:\n%s", want, page)
		}
	}
}

// TestArgRootTallyNotConjunction: the root is a tally, not a conjunction (spec/ARGUMENT.md § The root).
// A root with one failed child and two holding renders the per-class counts and names the failed child
// — it does NOT collapse to the bare verdict "fails" the way a recommendation node would. Under the old
// worst-child root rule this block read "Judgement: fails"; the refuter is that it no longer can.
func TestArgRootTallyNotConjunction(t *testing.T) {
	root := mustBuild(t, joinLines(
		"root  | Root.  | x",
		"    R1  | Rec one.  | x",
		"        F1  | Finding one.  | x",
		"            F1",
		"    R2  | Rec two.  | x",
		"        F2  | Finding two.  | x",
		"            F2",
		"    R3  | Rec three.  | x",
		"        F3  | Finding three.  | x",
		"            F3",
	), []brief.Row{
		{ID: "F1", Text: "f1", Faith: "faithful"},
		{ID: "F2", Text: "f2", Faith: "faithful"},
		{ID: "F3", Text: "f3", Faith: "absent"},
	})
	_, rootBlock := ArgumentPage(root, nil, "")
	tally := strings.Split(rootBlock, "\n")[1] // line 0 is the content, line 1 is the tally paragraph
	for _, want := range []string{"Of 3 recommendations", "2 hold (R1, R2)", "1 fails (R3: no held source supports F3)"} {
		if !strings.Contains(tally, want) {
			t.Errorf("root tally missing %q:\n%s", want, tally)
		}
	}
	if strings.Contains(tally, "Judgement: fails") || strings.TrimSpace(tally) == "fails" {
		t.Errorf("root must render the tally, not the bare verdict fails: %q", tally)
	}
}

// ── fixtures ────────────────────────────────────────────────────────────────

func joinLines(lines ...string) string { return strings.Join(lines, "\n") + "\n" }

func mustBuild(t *testing.T, arg string, rows []brief.Row) *ArgNode {
	t.Helper()
	root, err := BuildArgument(arg, rows)
	if err != nil {
		t.Fatalf("BuildArgument: %v", err)
	}
	return root
}

func childByID(n *ArgNode, id string) *ArgNode {
	for _, c := range n.Children {
		if c.ID == id {
			return c
		}
	}
	return nil
}
