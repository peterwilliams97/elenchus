package edge

// edge.go implements the edge-level adversarial pass (spec/EDGE.md §2–§4): the schema and prompt for
// one model call per in-scope finding→recommendation edge, the mechanical admission check that decides
// whether a returned defeater COUNTS (§3), and the cross-edge template rule plus rollup that lift a
// method-level defeater to the root and reduce the rest to a per-edge `open`/`unchallenged` verdict.
//
// The pass may do exactly two things to an edge: REFUTE it (an admitted defeater ⇒ `open`) or decline
// to (`unchallenged`). It may never certify that the recommendation follows — the axis boundary in
// CLAUDE.md § The axis boundary applied to inference: reasoning kills an edge with a counter-world but
// can never confirm one. So the schema below has no field in which the model can assert soundness, and
// `unchallenged` means "no admitted defeater this pass", never "the inference is sound."
//
// The model call itself is NOT here — it rides the same backend seam and callSchema as faithJudge, in
// assay.go, because that is where the network lives. This package is pure: schema, prompt text, and the
// three code-side decisions (admit, template, rollup), each unit-testable with no backend.

import (
	"regexp"
	"strings"
)

// Defeater says: the finding is true, and the recommendation still does not follow. For a
// `practical` edge there is one allowed reason — the report itself says the recommended action has
// a downside (a cost, e.g. F-BATCH: small batches help product performance but reduce individual
// effectiveness), and in some setting that downside wins.
//
//   - `World` is the model's prose describing that setting. Not checked beyond non-empty; this is
//     what the reader judges.
//   - `Anchor` is the sentence in the report that states the downside, copied verbatim. `Admit`
//     checks it: it must be a norm-substring of this edge's finding text or one of its verified
//     quotes, else it is rejected. The model can only copy it, not write it.
//   - `Settles` is what observation would decide it. Non-empty only.
//   - `Kind` is a closed enum (see `kinds`) — what the `World` names.
//   - `CriticalQuestion` is a closed enum per scheme (see `schemeCQs`); `practical` has only
//     `side_effects`.
//
// The check guarantees the downside is really in the report; it does not guarantee the `World`
// built on it is right — R-ACCESS in run 5 quoted a sentence about friction decreasing and called
// it a cost.
type Defeater struct {
	World            string `json:"world"`
	Kind             string `json:"kind"`   // population | condition | definition
	Anchor           string `json:"anchor"` // the cost clause, verbatim in this edge's finding or verified quotes
	Settles          string `json:"settles"`
	CriticalQuestion string `json:"critical_question"` // which of the scheme's fixed CQs this answers
}

// Result is one edge sample's schema-enforced answer. `NoneAdmitted` true is the model's report that it
// asked the scheme's questions and none opened a concrete world; the axis boundary forbids reading it
// as confirmation. When NoneAdmitted is false, `Defeater` carries the offer the admission check judges.
type Result struct {
	Warrant             string   `json:"warrant"`
	NoneAdmitted        bool     `json:"none_admitted"`
	Defeater            Defeater `json:"defeater"`
	QuestionsConsidered []string `json:"questions_considered"`
}

// Verdicts a rolled-up edge can carry.
const (
	Open         = "open"         // an admitted defeater stands: F holds, yet a concrete world makes R fail
	Unchallenged = "unchallenged" // no admitted defeater (none returned, or all rejected)
)

// kinds is the closed set a defeater's `kind` must name (spec/EDGE.md §3 step 2). It carries the
// concreteness the note demands — a world that gestures without naming a population, condition or
// definition has no kind to declare and fails admission. `competing_goal` is deliberately absent: a
// world that swaps in a goal the report's audience does not hold varies the addressee, which the
// domain rule (§3 step 5) puts outside the report's domain, so it is not an admissible kind.
var kinds = map[string]bool{
	"population": true,
	"condition":  true,
	"definition": true,
}

// cq is one of a scheme's fixed critical questions: `Slug` is the closed enum value the model returns,
// `Question` the text supplied verbatim in the prompt (spec/EDGE.md §2). The model asks the scheme's
// questions, not questions of its own, so both the enum and the prompt list are authored here.
type cq struct {
	Slug     string
	Question string
}

// schemeCQs maps each argument scheme (spec/EDGE.md §1) to its fixed s, taken verbatim
// from the consultant-report scheme list in docs/todo/adversary-design-2026-09-14.md § Operational form.
// The scheme names what the inference from finding to recommendation IS; its CQs are the only questions
// the edge call may ask. `survey`/`trend`/`classification` are defined for completeness though the dora
// and master-plan refuter corpora exercise only `practical` and `example` (spec/EDGE.md §5(a)).
var schemeCQs = map[string][]cq{
	// practical carries side_effects only (spec/EDGE.md §2, run 5, 15 Sept). The goal-naming CQs
	// (goal_held, feasible, goal_conflict) are removed — each is answerable for any practical
	// recommendation by inventing an addressee with other goals, a world outside the report's domain
	// (§3 step 5). alt_means is removed (run 3): a cheaper route M' defeats only "do M rather than M'",
	// a comparative no recommendation makes. means_mismatch is removed (run 4): its admitted worlds were
	// implementation or feasibility questions — the reader's, not an inference the model can settle. The
	// one CQ that remains attacks the F→R gap from inside the report: side_effects names a cost of M in
	// the finding's own text, which is checkable against the report.
	"practical": {
		{"side_effects", "Does the recommended action carry a cost the finding or its quotes already name?"},
	},
	"survey": {
		{"position_to_know", "Are the respondents in a position to know what the claim reports?"},
		{"sample_population", "Is the surveyed sample the population the recommendation is made to?"},
		{"question_asked", "Was the question actually asked the claim now being drawn from it?"},
	},
	"example": {
		{"typical_case", "Is the cited case typical of the class the rule generalises to?"},
		{"counter_cases", "How many counter-cases were looked for before generalising?"},
	},
	"trend": {
		{"indicator_tracks", "Does the indicator actually track the thing it is read as signalling?"},
		{"extrapolates", "Does the trend extrapolate to the range the conclusion needs?"},
	},
	"classification": {
		{"definition_matches", "Does the definition used here match the one that carries the conclusion?"},
	},
}

// KnownScheme reports whether s is a scheme this pass can run — i.e. one schemeCQs types. A finding
// tagged with a scheme outside the closed set never reaches the parser (argument.go validates it), so
// this guards only a direct caller.
func KnownScheme(s string) bool { _, ok := schemeCQs[s]; return ok }

// Admit runs the mechanical, no-NLP admission check of spec/EDGE.md §3 steps 1–3 (the cross-edge
// template rule of step 4 runs separately, in Resolve). `findingAndQuotes` is the prose `anchor` must
// land in — THIS edge's own finding text and each of its verified quotes, the premise under attack. It
// returns ok, and on failure the step that rejected the offer — "world", "settles", "kind", or
// "anchor" — so a run logs why each offered defeater was discarded and states how many were offered
// against how many admitted. The check is form-only: it never judges whether the world is a GOOD
// defeater, only that the model named a concrete referent AND that referent — the cost clause
// side_effects names — appears in this edge's own finding or its quotes (§3 step 3, run 5). The anchor
// is confined to the premise being defeated: never the recommendation (a referent in R while the world
// names a goal R never states, run 3), and never another finding's text (F-ACCESS anchored on F-DATA,
// run 4). If the finding and its quotes name no cost, no concrete world opens and the offer is rejected.
func Admit(d Defeater, findingAndQuotes []string) (ok bool, failedStep string) {
	if strings.TrimSpace(d.World) == "" {
		return false, "world"
	}
	if strings.TrimSpace(d.Settles) == "" {
		return false, "settles"
	}
	if !kinds[d.Kind] {
		return false, "kind"
	}
	// step 3: `anchor` is the cost clause side_effects names, verbatim in this edge's own finding or its
	// quotes — never the recommendation, never another finding.
	anchor := strings.TrimSpace(d.Anchor)
	if anchor == "" || !anchorInReport(anchor, findingAndQuotes) {
		return false, "anchor"
	}
	return true, ""
}

var anchorWS = regexp.MustCompile(`\s+`)
var anchorPunct = strings.NewReplacer("’", "'", "‘", "'", "“", `"`, "”", `"`, "—", "-", "–", "-")

// norm lowercases, folds curly punctuation and dashes to ASCII, and collapses whitespace — the same
// normalisation quoteInPassage (assay.go) runs, so an anchor matches its world despite a wrap or a
// curly apostrophe but a changed word still fails. The lower-casing also makes the anchor match
// case-insensitively, which is what the template rule (Resolve) counts on.
func norm(s string) string {
	return anchorWS.ReplaceAllString(anchorPunct.Replace(strings.ToLower(strings.TrimSpace(s))), " ")
}

// anchorInReport reports whether the anchor appears verbatim (modulo `norm`) inside any of the given
// report texts — this edge's own finding and its verified quotes — step 3 of the admission check. The
// model must point at a concrete referent the finding under attack itself names, not one it coined in
// its own world; this confirms the pointer lands in the premise being defeated.
func anchorInReport(anchor string, reportText []string) bool {
	a := norm(anchor)
	if a == "" {
		return false
	}
	for _, t := range reportText {
		if strings.Contains(norm(t), a) {
			return true
		}
	}
	return false
}

// EdgeResult is one edge's modal outcome over the N samples, before the template rule. `Verdict` is Open
// or Unchallenged; on Open, `Defeater` carries the modal admitted defeater (its world and anchor feed
// the template rule and the recommendation's reason line). FindingID/RecID name the F→R edge.
type EdgeResult struct {
	FindingID string
	RecID     string
	Verdict   string
	Defeater  Defeater
}

// MethodNote is a defeater the template rule lifted from the edges to the root (spec/EDGE.md §3 rule 4):
// a world whose anchor recurs across more than one edge is a property of the report's EVIDENCE TYPE, not
// of any one inference, so it is recorded once at the root and removed from every edge. `Edges` names the
// finding ids it defeated, in input order.
type MethodNote struct {
	Anchor  string
	World   string
	Settles string
	Edges   []string
}

// Resolve applies the cross-edge template rule and returns the post-template state (spec/EDGE.md §3 rule
// 4, §4). An admitted defeater whose anchor (case-insensitive, via norm) recurs on more than one Open
// edge is method-level: it becomes one MethodNote at the root and every edge it hit is reduced to
// Unchallenged (none-admitted for this pass). The returned `verdict` and `world` maps are keyed by
// finding id and hold the FINAL per-edge verdict and, for a still-Open edge, the defeater world its
// reason line names. dora's correlation-as-cause is the expected lift — one world defeats each
// associational edge, so it belongs at the root once, not on all eight.
func Resolve(edges []EdgeResult) (verdict, world map[string]string, methods []MethodNote) {
	verdict = make(map[string]string, len(edges))
	world = make(map[string]string, len(edges))

	// Count each Open edge's anchor across edges; an anchor on more than one edge is method-level.
	type group struct {
		display string // the first raw anchor seen, for the MethodNote
		note    MethodNote
		edges   []string // finding ids carrying this anchor
	}
	groups := map[string]*group{}
	var order []string
	for _, e := range edges {
		verdict[e.FindingID] = e.Verdict
		if e.Verdict != Open {
			continue
		}
		key := norm(e.Defeater.Anchor)
		if key == "" {
			continue
		}
		g := groups[key]
		if g == nil {
			g = &group{display: e.Defeater.Anchor,
				note: MethodNote{Anchor: e.Defeater.Anchor, World: e.Defeater.World, Settles: e.Defeater.Settles}}
			groups[key] = g
			order = append(order, key)
		}
		g.edges = append(g.edges, e.FindingID)
	}

	methodKeys := map[string]bool{}
	for _, key := range order {
		if g := groups[key]; len(g.edges) > 1 {
			methodKeys[key] = true
			g.note.Edges = g.edges
			methods = append(methods, g.note)
		}
	}

	for _, e := range edges {
		if e.Verdict != Open {
			continue
		}
		if methodKeys[norm(e.Defeater.Anchor)] {
			verdict[e.FindingID] = Unchallenged // lifted to the root; the edge no longer carries it
			continue
		}
		world[e.FindingID] = e.Defeater.World
	}
	return verdict, world, methods
}

// FixedMethodNote is the root method note the pass emits from CODE (spec/EDGE.md §3 rule 4, §5) for a
// report whose in-scope edges are all `practical`: the correlation→intervention gap is a property of
// the evidence type, not of any one inference, so it is stated once at the root rather than attacked
// per edge. With practical's CQ set reduced to side_effects only (§2) the model no longer offers the
// correlation-as-cause world per edge, so the template rule (Resolve) is no longer expected to lift it;
// this fixed note carries dora's method-level point instead, produced by the pass, never by
// the model. It rides RootMethods, rendered where MethodLine's lifted defeaters render.
const FixedMethodNote = "method: findings are associational; recommendations are interventions."

// MethodLine renders one lifted defeater as the root line spec/EDGE.md §4 prescribes: the world as the
// claim and what would settle it, tagged as an evidence-type defect that defeats every edge of its kind.
func (m MethodNote) MethodLine() string {
	line := "method: " + m.World +
		" — defeats every edge of this evidence type; settled against the report's evidence type, not any one inference."
	if strings.TrimSpace(m.Settles) != "" {
		line += " Settles it: " + m.Settles + "."
	}
	return line
}
