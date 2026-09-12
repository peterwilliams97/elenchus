package tree

// argument_test pins the argument tree's derivation rule (spec/ARGUMENT.md § Internal judgement) and
// the `?`-edge override the CLI adds on top of it: a contested leaf opens the root; a `?` edge opens
// the recommendation it hangs under; an all-holds tree holds to the root; opinion propagation; the
// rendered page shape; the recommendation count is every root child but `base` (any id prefix); and an
// `uncorroborated` leaf weakens rather than fails under a single-source corpus. Fixtures sit at the foot.

import (
	"strings"
	"testing"

	"assay/internal/adjudicate"
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

// TestArgumentPage pins the card page's content: the report title is the <title> and <h1>, the `what`
// sentence and the thesis card carry the proposition and tally, the recommendation card shows its
// derived badge, id, proposition and deciding child, the `?` companion carries the `?` marker, and the
// claim leaf card carries its verdict badge, claim text and reason. The stdout root block is unchanged.
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
	page, rootBlock := ArgumentPage(root,
		map[string]Leaf{"F30": {Reason: "the source says so"}},
		"The Cultural Industries Inquiry", "This page checks the report against its sources.", false, nil)
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
		"<title>The Cultural Industries Inquiry</title>", // the report title, not the shared "assay tree"
		"<h1>The Cultural Industries Inquiry</h1>",
		"This page checks the report against its sources.", // the `what` sentence
		`<section class="thesis">`,
		"industries matter and the Government should act", // the thesis, in its card
		`<details class="card">`, `<details class="card leaf">`,
		`<span class="badge b-open">open</span>`,      // R4's derived judgement
		`<span class="badge b-holds">faithful</span>`, // a settled-faithful leaf's badge
		"Advocate release of the breakdowns.",
		`<span class="dc">(F29 ?)</span>`, // the deciding child, small text
		`<span class="q"> ?</span>`,       // the `?` marker on the companion's id
		"breakdowns not released", "fair share received",
		"reason: the source says so",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("page missing %q:\n%s", want, page)
		}
	}
}

// TestArgumentCardPage is the spec/SERVE.md refuter: the page is a thesis card first, then the
// recommendation cards in the argument file's order, with each claim card nested inside a finding card
// inside a recommendation card. It builds two recommendations and a base child and asserts the source
// order survives to the page and the three-deep nesting holds — the order and depth a reader relies on
// to read the case top-down.
func TestArgumentCardPage(t *testing.T) {
	root := mustBuild(t, joinLines(
		"root  | Root thesis.  | x",
		"    R1  | First recommendation.  | x",
		"        F1  | Finding one.  | x",
		"            F1",
		"    R2  | Second recommendation.  | x",
		"        F2  | Finding two.  | x",
		"            F2",
		"    base ?  | Descriptive base.  | x",
		"        F3  | Base finding.  | x",
		"            F3",
	), []brief.Row{
		{ID: "F1", Text: "claim one", Faith: "faithful"},
		{ID: "F2", Text: "claim two", Faith: "absent"},
		{ID: "F3", Text: "claim three", Faith: "faithful"},
	})
	page, _ := ArgumentPage(root, nil, "Report", "what", false, nil)

	thesis := strings.Index(page, `<section class="thesis">`)
	r1 := strings.Index(page, "First recommendation.")
	r2 := strings.Index(page, "Second recommendation.")
	base := strings.Index(page, "Descriptive base.")
	if !(thesis >= 0 && thesis < r1 && r1 < r2 && r2 < base) {
		t.Fatalf("cards out of order: thesis=%d r1=%d r2=%d base=%d", thesis, r1, r2, base)
	}
	// R1 (holds) and R2 (fails) carry their derived badges; a base child on a `?` edge takes the marker.
	for _, want := range []string{
		`<span class="badge b-holds">holds</span>`,
		`<span class="badge b-fails">fails</span>`,
		`<span class="q"> ?</span>`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("page missing %q", want)
		}
	}
	// Nesting: within R1's card, a finding card, within it the claim leaf card — recommendation →
	// finding → claim, matching the argument file's depth. The claim text sits after two card opens.
	seg := page[r1:r2]
	firstCard := strings.Index(seg, `<details class="card">`) // the finding card, nested under R1
	leaf := strings.Index(seg, `<details class="card leaf">`) // the claim card, nested under the finding
	if !(firstCard >= 0 && leaf > firstCard) {
		t.Fatalf("R1 does not nest finding→claim: findingCard=%d leaf=%d\n%s", firstCard, leaf, seg)
	}
}

// TestArgumentTitle: the "# title:" header is read as the report name and every other line — comments
// without a title key, and the tree lines — is ignored; a file with no such header yields "" so the
// caller falls back to the file name (never the thesis).
func TestArgumentTitle(t *testing.T) {
	withTitle := joinLines(
		"# argument.txt — the report as an argument tree",
		"# title: Inquiry into X — Committee, June 2025",
		"root  | The report argues X.  | SYNTHESISED",
	)
	if got := ArgumentTitle(withTitle); got != "Inquiry into X — Committee, June 2025" {
		t.Errorf("ArgumentTitle = %q, want the header value", got)
	}
	noTitle := joinLines(
		"# Line format (indentation = depth):",
		"root  | The report argues X.  | SYNTHESISED",
	)
	if got := ArgumentTitle(noTitle); got != "" {
		t.Errorf("ArgumentTitle = %q, want \"\" when no title header (caller falls back to file name)", got)
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
	_, rootBlock := ArgumentPage(root, nil, "", "", false, nil)
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

// TestArgRootTallyCountsNonRRecommendations: a recommendation is any root child but `base`, whatever
// its id prefix. Quocirca numbers its recommendations S1–S7 / B1–B5, none with an `R…` spelling; the
// tally must count all four here (two S, two B) and exclude the `base` node's finding, not fall to
// zero on the missing `R`. Mutating an S id to `base` would drop it from the count.
func TestArgRootTallyCountsNonRRecommendations(t *testing.T) {
	root := mustBuild(t, joinLines(
		"root  | Root.  | x",
		"    S1  | Supplier rec one.  | x",
		"        F1  | Finding one.  | x",
		"            F1",
		"    S2  | Supplier rec two.  | x",
		"        F2  | Finding two.  | x",
		"            F2",
		"    B1  | Buyer rec one.  | x",
		"        F3  | Finding three.  | x",
		"            F3",
		"    B2  | Buyer rec two.  | x",
		"        F4  | Finding four.  | x",
		"            F4",
		"    base  | Descriptive base.  | x",
		"        F5  | A finding no rec rests on.  | x",
		"            F5",
	), []brief.Row{
		{ID: "F1", Text: "f1", Faith: "faithful"},
		{ID: "F2", Text: "f2", Faith: "faithful"},
		{ID: "F3", Text: "f3", Faith: "faithful"},
		{ID: "F4", Text: "f4", Faith: "faithful"},
		{ID: "F5", Text: "f5", Faith: "faithful"},
	})
	_, rootBlock := ArgumentPage(root, nil, "", "", false, nil)
	lines := strings.Split(rootBlock, "\n")
	if want := "Of 4 recommendations"; !strings.Contains(lines[1], want) {
		t.Errorf("tally missing %q:\n%s", want, lines[1])
	}
	if want := "1 finding supports no recommendation."; !strings.Contains(lines[2], want) {
		t.Errorf("base sentence missing %q:\n%s", want, lines[2])
	}
}

// TestArgUncorroboratedWeakensNotFails is the single_source refuter: a leaf verdict `uncorroborated`
// (assay.go's remap of `absent` when the report is its own only source) weakens its recommendation, it
// does not fail it — the contrast is the same fixture with `absent`, which fails. The page shows the
// verdict word `uncorroborated` on the leaf and, with singleSource=true, the key line explaining it.
func TestArgUncorroboratedWeakensNotFails(t *testing.T) {
	arg := joinLines(
		"root  | Root.  | x",
		"    R1  | Rec one.  | x",
		"        F1  | Finding one.  | x",
		"            F1",
	)
	unc := mustBuild(t, arg, []brief.Row{{ID: "F1", Text: "f1", Faith: "uncorroborated"}})
	if got := childByID(unc, "R1").Judgement(); got != jWeakened {
		t.Fatalf("uncorroborated leaf should weaken its recommendation, got %q", got)
	}
	// The contrast: absent (a corpus with other held sources) fails the same recommendation.
	abs := mustBuild(t, arg, []brief.Row{{ID: "F1", Text: "f1", Faith: "absent"}})
	if got := childByID(abs, "R1").Judgement(); got != jFails {
		t.Fatalf("absent leaf should fail its recommendation, got %q", got)
	}

	page, _ := ArgumentPage(unc, nil, "Report", "what", true, nil)
	for _, want := range []string{
		`<span class="badge b-weakened">uncorroborated</span>`, // the leaf badge: weakened colour, verdict word
		"said once in the report, not repeated elsewhere",      // the key line
	} {
		if !strings.Contains(page, want) {
			t.Errorf("single_source page missing %q", want)
		}
	}
	// Without single_source the key must not carry the uncorroborated line.
	if plain, _ := ArgumentPage(unc, nil, "Report", "what", false, nil); strings.Contains(plain, "said once in the report") {
		t.Errorf("multi-source key should not carry the uncorroborated line")
	}
}

// TestArgumentPageAdjudications drives ArgumentPage with a human overlay and pins the two ways the page
// uses it (spec/SERVE.md § Adjudications): the index reports the agreement count and lists the
// disagreement with its reason, and an adjudicated leaf shows the human verdict beside the machine
// badge. Two leaves — F1 the human agrees with (faithful == faithful), F2 not (overstated vs partial).
func TestArgumentPageAdjudications(t *testing.T) {
	arg := joinLines(
		"root  | Root.  | x",
		"    R1  | Rec one.  | x",
		"        F1  | Finding one.  | x",
		"            F1",
		"        F2  | Finding two.  | x",
		"            F2",
	)
	root := mustBuild(t, arg, []brief.Row{
		{ID: "F1", Text: "claim one", Faith: "faithful"},
		{ID: "F2", Text: "claim two", Faith: "partial"},
	})
	adjs := []adjudicate.Adjudication{
		{ID: "F1", Verdict: "faithful", Initials: "PW", Reason: "source says exactly this"},
		{ID: "F2", Verdict: "overstated", Initials: "PW", Reason: "claim inflates the source"},
	}
	machine := map[string]string{"F1": "faithful", "F2": "partial"}
	page, _ := ArgumentPage(root, nil, "Report", "what", false, adjudicate.NewOverlay(machine, adjs))
	for _, want := range []string{
		"2 leaves adjudicated, judge agreed on 1.",
		"<b>F2</b> — human overstated, judge partial: claim inflates the source",
		`<span class="human">PW: overstated</span>`, // the disagreeing leaf's human chip
		`<span class="human">PW: faithful</span>`,   // the agreeing leaf's human chip
	} {
		if !strings.Contains(page, want) {
			t.Errorf("adjudication page missing %q", want)
		}
	}
	// No overlay → no adjudication summary at all.
	if plain, _ := ArgumentPage(root, nil, "Report", "what", false, nil); strings.Contains(plain, "adjudicated") {
		t.Errorf("page without an overlay must carry no adjudication summary")
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
