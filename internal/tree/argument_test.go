package tree

// argument_test pins the argument tree's derivation rule (spec/ARGUMENT.md § Internal judgement) and
// the `?`-edge override the CLI adds on top of it: a contested leaf opens the root; a `?` edge opens
// the recommendation it hangs under; an all-holds tree holds to the root; opinion propagation; the
// rendered page shape; the recommendation count is every root child but `base` (any id prefix); and an
// `uncorroborated` leaf weakens rather than fails under a single-source corpus. Fixtures sit at the foot.

import (
	"os"
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

// TestDisputedFlag pins the disputed-leaf propagation (spec/ARGUMENT.md § Disputed leaves): a leaf a human
// read otherwise than the machine flags every node whose derived verdict is load-bearing on it — its
// finding and the recommendation above — while an adjudication the machine agreed with flags nothing. The
// derived verdicts do not move: R-BAD still fails on the machine's contradicted, R-OK still holds.
func TestDisputedFlag(t *testing.T) {
	arg := joinLines(
		"root  | Root.  | x",
		"    R-BAD  | Rec bad.  | x",
		"        F-BAD  | Finding bad.  | x",
		"            CBAD",
		"    R-OK  | Rec ok.  | x",
		"        F-OK  | Finding ok.  | x",
		"            COK",
	)
	root := mustBuild(t, arg, []brief.Row{
		{ID: "CBAD", Text: "bad claim", Faith: "contradicted"},
		{ID: "COK", Text: "ok claim", Faith: "faithful"},
	})
	machine := map[string]string{"CBAD": "contradicted", "COK": "faithful"}
	adjs := []adjudicate.Adjudication{
		{ID: "CBAD", Verdict: "faithful", Initials: "PW", Reason: "the source states it directly"},
		{ID: "COK", Verdict: "faithful", Initials: "PW", Reason: "agrees with the machine"},
	}
	disputed := disputedSet(adjudicate.NewOverlay(machine, adjs))

	rBad := childByID(root, "R-BAD")
	fBad := childByID(rBad, "F-BAD")
	rOk := childByID(root, "R-OK")

	// The disputed leaf is load-bearing for its finding and the recommendation above it.
	if got := disputedFor(fBad, disputed); len(got) != 1 || got[0] != "CBAD" {
		t.Errorf("F-BAD disputedFor = %v, want [CBAD]", got)
	}
	if got := disputedFor(rBad, disputed); len(got) != 1 || got[0] != "CBAD" {
		t.Errorf("R-BAD disputedFor = %v, want [CBAD]", got)
	}
	// An adjudication the machine agreed with is not in the disputed set, so it flags nothing.
	if got := disputedFor(rOk, disputed); len(got) != 0 {
		t.Errorf("R-OK disputedFor = %v, want none (adjudication agreed)", got)
	}
	// The derived verdicts are unmoved by the disagreement — the flag annotates, it does not re-judge.
	if j := rBad.Judgement(); j != jFails {
		t.Errorf("R-BAD judgement = %q, want fails", j)
	}
	if j := rOk.Judgement(); j != jHolds {
		t.Errorf("R-OK judgement = %q, want holds", j)
	}
	// The clause reaches the rendered page and the stdout root block, on the disputed branch only.
	page, block := ArgumentPage(root, nil, "Report", "what", false, adjudicate.NewOverlay(machine, adjs))
	if !strings.Contains(page, "(machine; human disagrees: CBAD)") {
		t.Errorf("page missing the disputed clause for CBAD")
	}
	if strings.Contains(page, "human disagrees: COK") {
		t.Errorf("agreed leaf COK must not produce a disputed clause")
	}
	if !strings.Contains(block, "(machine; human disagrees: CBAD)") {
		t.Errorf("root block missing the disputed clause for CBAD:\n%s", block)
	}
	if strings.Contains(block, "human disagrees: COK") {
		t.Errorf("root block: agreed leaf COK must not produce a disputed clause")
	}
}

// TestSubstanceGatesHold is the substance-rollup refuter (spec/SUBSTANCE-CORPUS.md): the substance axis
// gates the one `holds` case. A settled-faithful leaf holds when substance is absent (a
// faithfulness-only run, back-compat) or `substantive` (the control shape); a settled-faithful leaf whose
// substance came back `hollow` or `partial` weakens instead — the report copies its source, but the
// proposition does not survive the dialectic. Substance never rescues a leaf faithfulness already sank,
// so a `partial` FAITHFULNESS verdict stays weakened whatever substance says.
func TestSubstanceGatesHold(t *testing.T) {
	for _, tc := range []struct {
		name       string
		faith, sub string
		class      string
		want       string
	}{
		{"faithful, no substance run → holds (back-compat)", "faithful", "", "", jHolds},
		{"faithful + substantive → holds (control)", "faithful", "substantive", "", jHolds},
		{"faithful + hollow → weakened (defect leaf)", "faithful", "hollow", "", jWeakened},
		{"faithful + partial substance → weakened", "faithful", "partial", "", jWeakened},
		{"faithful + wobble → weakened whatever substance", "faithful", "substantive", "wobble", jWeakened},
		{"partial faithfulness stays weakened, substance can't rescue", "partial", "substantive", "", jWeakened},
	} {
		r := brief.Row{ID: "F1", Text: "f1", Faith: tc.faith, Substance: tc.sub, Class: tc.class}
		if got := leafJudgement(&r); got != tc.want {
			t.Errorf("%s: leafJudgement = %q, want %q", tc.name, got, tc.want)
		}
	}

	// An opinion (route=evaluative) leaf is never judged, so its substance verdict — however damning —
	// does not reclassify it: it stays opinion and contributes nothing. This is why the two evaluative
	// refuter leaves are tested empirically on their substance VERDICT, not on a tally change.
	op := brief.Row{ID: "F1", Text: "f1", Faith: "faithful", Substance: "hollow", Route: "evaluative"}
	if got := leafJudgement(&op); got != jOpinion {
		t.Errorf("evaluative leaf should stay opinion despite hollow substance, got %q", got)
	}
}

// TestSchemeTagsInCorpora is the spec/EDGE.md §1 refuter: the two in-scope corpora carry a scheme tag
// on exactly the in-scope finding→recommendation edges and nowhere else — dora's 8 capability findings
// (F-STANCE…F-VSM) and master-plan's 15 findings (F1…F15). It parses the shipped argument.txt files
// and counts nodes whose edge to the parent is typed; a dropped tag, a stray tag on a rec or base line,
// or a corpus edit that adds an edge reddens it. The count is the whole assertion — an empty parse
// would read 0 and fail, per the zero-output rule.
func TestSchemeTagsInCorpora(t *testing.T) {
	for _, tc := range []struct {
		path string
		want int
	}{
		{"../../examples/dora-2026/argument.txt", 8},
		{"../../examples/master-plan/argument.txt", 15},
	} {
		text, err := os.ReadFile(tc.path)
		if err != nil {
			t.Fatalf("read %s: %v", tc.path, err)
		}
		root, err := parseArg(string(text))
		if err != nil {
			t.Fatalf("parseArg %s: %v", tc.path, err)
		}
		if got := countTagged(root); got != tc.want {
			t.Errorf("%s: %d scheme-tagged edges, want %d", tc.path, got, tc.want)
		}
	}
}

// TestSchemeBadValueFailsParse: a scheme value outside the closed set (spec/EDGE.md §1) stops the parse
// rather than typing the edge to nothing, while a set member parses and lands on the node's `Scheme`.
// The pair is the refuter — the closed set must both admit its members and reject a typo, so a corpus
// with a mistyped scheme cannot render as if the edge were untagged.
func TestSchemeBadValueFailsParse(t *testing.T) {
	bad := joinLines(
		"root  | Root.  | x",
		"    R1  | Rec one.  | x",
		"        F1  | Finding one.  | scheme=analogy; §Somewhere",
	)
	if _, err := parseArg(bad); err == nil {
		t.Fatalf("an unknown scheme value must fail the parse")
	}
	good := joinLines(
		"root  | Root.  | x",
		"    R1  | Rec one.  | x",
		"        F1  | Finding one.  | scheme=practical; §Somewhere",
	)
	root, err := parseArg(good)
	if err != nil {
		t.Fatalf("a valid scheme must parse: %v", err)
	}
	if f1 := childByID(childByID(root, "R1"), "F1"); f1 == nil || f1.Scheme != "practical" {
		t.Fatalf("scheme not read onto the node: %+v", f1)
	}
}

// ── fixtures ────────────────────────────────────────────────────────────────

// TestEdgeOpenRollsUpToRecommendation: a recommendation whose only finding is faithful-and-settled
// (leaf-derived holds) but whose F→R edge is open (an admitted defeater) becomes open, and its reason
// names the edge defeater (spec/EDGE.md §4). This is the gap the edge pass exists to surface — the
// finding is true and the recommendation still doesn't follow — which no leaf-faithfulness computation
// can reach. The mutation, clearing the edge, must return the recommendation to holds, proving the edge
// verdict, not the fixture, opened it.
func TestEdgeOpenRollsUpToRecommendation(t *testing.T) {
	arg := joinLines(
		"root  | Root proposition.  | x",
		"    R1  | Recommendation one.  | x",
		"        F1  | Finding one.  | scheme=practical; x",
		"            CM1",
	)
	root := mustBuild(t, arg, []brief.Row{{ID: "CM1", Text: "c", Faith: "faithful"}})
	rec := childByID(root, "R1")
	finding := childByID(rec, "F1")
	if got := rec.Judgement(); got != jHolds {
		t.Fatalf("with no edge verdict the recommendation should hold, got %q", got)
	}
	finding.EdgeVerdict = jOpen
	finding.EdgeWorld = "a regulated org whose binding constraint is delivery stability"
	if got := rec.Judgement(); got != jOpen {
		t.Fatalf("an open edge on a holds finding should open the recommendation, got %q", got)
	}
	if r := recReason(rec); !strings.Contains(r, "edge defeater") {
		t.Fatalf("the recommendation's reason should name the edge defeater, got %q", r)
	}
	// Mutation: clear the edge → back to holds.
	finding.EdgeVerdict, finding.EdgeWorld = "", ""
	if got := rec.Judgement(); got != jHolds {
		t.Fatalf("clearing the edge should return the recommendation to holds, got %q", got)
	}
}

// TestEdgeMethodLineRendersOnce: a method-level defeater lifted to the root (RootMethods) renders once
// in the root block, above the recommendations, and the edges it was lifted off contribute nothing
// (their EdgeVerdict is unchallenged after the lift). With RootMethods empty the block is unchanged.
func TestEdgeMethodLineRendersOnce(t *testing.T) {
	arg := joinLines(
		"root  | Root proposition.  | x",
		"    R1  | Recommendation one.  | x",
		"        F1  | Finding one.  | scheme=practical; x",
		"            CM1",
	)
	root := mustBuild(t, arg, []brief.Row{{ID: "CM1", Text: "c", Faith: "faithful"}})
	childByID(root, "R1").Children[0].EdgeVerdict = "unchallenged" // lifted off this edge
	root.RootMethods = []string{"method: a shared common-cause world — defeats every edge of this evidence type."}
	block := argRootBlock(root, nil)
	if n := strings.Count(block, "method: a shared common-cause world"); n != 1 {
		t.Fatalf("method line should render exactly once, got %d", n)
	}
	if got := childByID(root, "R1").Judgement(); got != jHolds {
		t.Fatalf("an edge lifted to the root leaves the recommendation on its leaf verdict (holds), got %q", got)
	}
}

func joinLines(lines ...string) string { return strings.Join(lines, "\n") + "\n" }

// countTagged sums the nodes whose edge to the parent carries a scheme tag (spec/EDGE.md §1) — every
// node with a non-empty `Scheme`, over the whole tree.
func countTagged(n *ArgNode) int {
	c := 0
	if n.Scheme != "" {
		c++
	}
	for _, ch := range n.Children {
		c += countTagged(ch)
	}
	return c
}

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
