package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"assay/internal/backend"
	"assay/internal/backend/fake"
)

func TestSplitSummaryNewlines(t *testing.T) {
	in := "1. The future of work happens in Claude Code.\n2. Every company has a super-agent.\n3. SaaS is not dead."
	got := splitSummary(in)
	if len(got) != 3 {
		t.Fatalf("want 3, got %d: %#v", len(got), got)
	}
	if got[0] != "The future of work happens in Claude Code." {
		t.Errorf("leading marker not stripped: %q", got[0])
	}
}

func TestSplitSummaryRunTogether(t *testing.T) {
	in := `1. Future of work in Claude Code.2. A super-agent in Slack.3. SaaS is not dead—"buy SaaS stocks right now."4. PMs will thrive.`
	got := splitSummary(in)
	if len(got) != 4 {
		t.Fatalf("want 4, got %d: %#v", len(got), got)
	}
	if !strings.Contains(got[3], "PMs will thrive") {
		t.Errorf("last item wrong: %q", got[3])
	}
}

func TestSplitSummaryBullets(t *testing.T) {
	in := "- The future of work happens in Claude Code.\n- Every company has a super-agent.\n- SaaS is not dead."
	got := splitSummary(in)
	if len(got) != 3 {
		t.Fatalf("want 3, got %d: %#v", len(got), got)
	}
	if got[0] != "The future of work happens in Claude Code." {
		t.Errorf("leading bullet not stripped: %q", got[0])
	}
}

func TestSplitSummarySingleLine(t *testing.T) {
	in := "Automation is a lie."
	got := splitSummary(in)
	if len(got) != 1 || got[0] != in {
		t.Errorf("single line: want [%q], got %#v", in, got)
	}
}

func TestSplitSummaryBlankLines(t *testing.T) {
	in := "1. First claim.\n\n2. Second claim.\n\n3. Third claim."
	got := splitSummary(in)
	if len(got) != 3 {
		t.Fatalf("want 3, got %d: %#v", len(got), got)
	}
}

func TestExtractJSON(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"```json\n[\"a\",\"b\"]\n```", `["a","b"]`},
		{`preamble {"v":"hollow","r":"has } brace in string"} trailing`, `{"v":"hollow","r":"has } brace in string"}`},
		{`text then {"a":1}`, `{"a":1}`},
	}
	for _, c := range cases {
		if got := extractJSON(c.in); got != c.want {
			t.Errorf("extractJSON(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestUnmarshalLooseTrailingComma(t *testing.T) {
	var v struct {
		Verdict string `json:"verdict"`
	}
	if err := unmarshalLoose(`{"verdict":"hollow",}`, &v); err != nil {
		t.Fatalf("err: %v", err)
	}
	if v.Verdict != "hollow" {
		t.Errorf("got %q", v.Verdict)
	}
}

// TestUnmarshalLooseConditionFields checks that the two new fields added to substanceJSON for
// condition-laundering detection round-trip correctly.
func TestUnmarshalLooseConditionFields(t *testing.T) {
	var v substanceJSON
	const raw = `{"verdict":"partial","reason":"survives narrowed","needs_another_round":false,` +
		`"added_conditions":3,"survives_only_by_conditioning":true,"surviving_claim":"narrowed","critique":[]}`
	if err := unmarshalLoose(raw, &v); err != nil {
		t.Fatalf("unmarshal err: %v", err)
	}
	if v.AddedConditions != 3 {
		t.Errorf("added_conditions: want 3, got %d", v.AddedConditions)
	}
	if !v.SurvivesOnlyByConditions {
		t.Error("survives_only_by_conditioning: want true")
	}
}

// stubCall returns canned JSON responses keyed on which system prompt is being called. It covers
// the producer, critic, faithfulness defender, faithfulness critic, and evidence grounder — enough
// for assayClaim, faithClaim, and runEvidence to run end-to-end without network access.
//
// The responses are controlled per test via the stubbedCritic field: callers can override what the
// substance critic returns to exercise specific paths.
func stubCall(criticJSON string) func(system, prompt string, withTools bool) (string, []retrievedSource, error) {
	return func(system, prompt string, withTools bool) (string, []retrievedSource, error) {
		switch {
		case strings.Contains(system, "You are the Producer"):
			return `{"steelman":"strongest version","conditions":"some conditions"}`, nil, nil
		case strings.Contains(system, "You are the Critic"):
			return criticJSON, nil, nil
		case strings.Contains(system, "You are the Defender"):
			return `{"found":true,"quotes":["the speaker said it"],"best_case":"direct quote"}`, nil, nil
		case strings.Contains(system, "You are the Faithfulness Critic"):
			// Default: overstated with what_source_actually_says populated.
			if strings.Contains(prompt, "faithful-claim") {
				return `{"findings":[],"verdict":"faithful","evidence":"direct match","what_source_actually_says":null}`, nil, nil
			}
			return `{"findings":[],"verdict":"overstated","evidence":"rhetorical","what_source_actually_says":"automation requires human oversight"}`, nil, nil
		case strings.Contains(system, "You are the Evidence Grounder"):
			// Return a matching retrieved source so the cross-check succeeds and verdict stands.
			rs := []retrievedSource{{Title: "stub source", URL: "https://stub.example.com/evidence"}}
			return `{"verdict":"supported","finding":"evidence found","sources":[{"title":"stub source","url":"https://stub.example.com/evidence"}]}`, rs, nil
		default:
			return `[]`, nil, nil
		}
	}
}

// TestConditionLaunderingDowngrade drives the REAL c.assayClaim with a stub
// that returns survives_only_by_conditioning=true. Asserts that assayClaim
// downgrades the verdict to hollow.
func TestConditionLaunderingDowngrade(t *testing.T) {
	const launderedCritic = `{"critique":[],"verdict":"partial","surviving_claim":"narrowed","reason":"survives narrowed.",` +
		`"needs_another_round":false,"added_conditions":3,"survives_only_by_conditioning":true}`
	c := cfg{maxRounds: 2, call: stubCall(launderedCritic)}
	result := c.assayClaim("Automation is a lie.")
	if result.Verdict != "hollow" {
		t.Errorf("want hollow after laundering downgrade, got %q", result.Verdict)
	}
	if !strings.Contains(result.Reason, "Survives only by conditions") {
		t.Errorf("reason should mention laundering, got %q", result.Reason)
	}
}

// TestConditionLaunderingLoopStop drives the REAL c.assayClaim with maxRounds=2 and a stub critic
// that wants another round but sets survives_only_by_conditioning=true.
// Asserts that the loop does NOT continue (only 1 round executes) and downgrades.
func TestConditionLaunderingLoopStop(t *testing.T) {
	// wants another round + laundered — loop must stop, not re-run producer.
	const launderedWantsMore = `{"critique":[],"verdict":"partial","surviving_claim":"narrowed","reason":"partial.",` +
		`"needs_another_round":true,"added_conditions":5,"survives_only_by_conditioning":true}`
	callCount := 0
	stub := func(system, prompt string, withTools bool) (string, []retrievedSource, error) {
		if strings.Contains(system, "You are the Critic") {
			callCount++
		}
		return stubCall(launderedWantsMore)(system, prompt, withTools)
	}
	c := cfg{maxRounds: 2, call: stub}
	result := c.assayClaim("Every company will have one super-agent.")
	if callCount != 1 {
		t.Errorf("critic should be called exactly once (loop stopped), got %d", callCount)
	}
	if result.Verdict != "hollow" {
		t.Errorf("want hollow, got %q", result.Verdict)
	}
}

// TestRunEvidencePropositionSubstitution drives the REAL c.runEvidence with a source string. The
// stub faithfulness critic returns overstated + what_source_actually_says.
// Asserts that the evidence grounder receives the intended proposition, not the literal claim.
func TestRunEvidencePropositionSubstitution(t *testing.T) {
	var groundedProposition string
	stub := func(system, prompt string, withTools bool) (string, []retrievedSource, error) {
		switch {
		case strings.Contains(system, "You are the Defender"):
			return `{"found":true,"quotes":["automation is a lie"],"best_case":"direct"}`, nil, nil
		case strings.Contains(system, "You are the Faithfulness Critic"):
			return `{"findings":[],"verdict":"overstated","evidence":"rhetorical","what_source_actually_says":"automation requires human oversight"}`, nil, nil
		case strings.Contains(system, "You are the Evidence Grounder"):
			// Capture what proposition was sent for grounding. Return a matching retrieved source
			// so the cross-check succeeds and the test keeps asserting proposition routing.
			groundedProposition = prompt
			rs := []retrievedSource{{Title: "oversight study", URL: "https://example.com/oversight"}}
			return `{"verdict":"mixed","finding":"humans still needed","sources":[{"title":"oversight study","url":"https://example.com/oversight"}]}`, rs, nil
		default:
			return `[]`, nil, nil
		}
	}
	c := cfg{call: stub}
	c.runEvidence("Automation is a lie.", "full transcript here")
	if !strings.Contains(groundedProposition, "automation requires human oversight") {
		t.Errorf("expected intended proposition to be grounded, got prompt: %q", groundedProposition)
	}
	if strings.Contains(groundedProposition, "Automation is a lie") {
		t.Errorf("literal claim should NOT be grounded when source provides intended proposition")
	}
}

// TestIntendedProposition is a pure-function unit test for the helper.
func TestIntendedProposition(t *testing.T) {
	for _, tc := range []struct {
		name  string
		fc    faith
		claim string
		want  string
	}{
		{"overstated with source says", faith{Verdict: "overstated", SourceSays: "intended prop"}, "literal", "intended prop"},
		{"partial with source says", faith{Verdict: "partial", SourceSays: "narrower prop"}, "literal", "narrower prop"},
		{"faithful — no substitution", faith{Verdict: "faithful", SourceSays: "same words"}, "literal", "literal"},
		{"overstated but empty source says", faith{Verdict: "overstated", SourceSays: ""}, "literal", "literal"},
		{"absent — no substitution", faith{Verdict: "absent"}, "literal", "literal"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := intendedProposition(tc.fc, tc.claim); got != tc.want {
				t.Errorf("want %q, got %q", tc.want, got)
			}
		})
	}
}

// TestAuditGroundsIntendedProposition drives the REAL c.computeAudit and asserts that the grounding
// column uses the intended proposition (what_source_actually_says) rather than the literal summary
// claim — reproducing the #10 case where the audit used to return refuted (literal reading) instead
// of the correct verdict.
func TestAuditGroundsIntendedProposition(t *testing.T) {
	// Stub returns:
	//   faithfulness → overstated, what_source_actually_says = intended proposition
	//   evidence grounding the LITERAL claim → refuted
	//   evidence grounding the INTENDED proposition → supported
	// The test asserts the audit grounding is supported (intended), not refuted (literal).
	var groundedProposition string
	stub := func(system, prompt string, withTools bool) (string, []retrievedSource, error) {
		switch {
		case strings.Contains(system, "You are the Producer"):
			return `{"steelman":"strongest version","conditions":"some conditions"}`, nil, nil
		case strings.Contains(system, "You are the Critic"):
			return `{"critique":[],"verdict":"partial","surviving_claim":null,"reason":"ok",` +
				`"needs_another_round":false,"added_conditions":0,"survives_only_by_conditioning":false}`, nil, nil
		case strings.Contains(system, "You are the Defender"):
			return `{"found":true,"quotes":["automation is a lie"],"best_case":"direct"}`, nil, nil
		case strings.Contains(system, "You are the Faithfulness Critic"):
			if strings.Contains(prompt, "faithful-claim") {
				return `{"findings":[],"verdict":"faithful","evidence":"direct","what_source_actually_says":null}`, nil, nil
			}
			return `{"findings":[],"verdict":"overstated","evidence":"rhetorical",` +
				`"what_source_actually_says":"automation always needs a human in the loop"}`, nil, nil
		case strings.Contains(system, "You are the Evidence Grounder"):
			groundedProposition = prompt
			if strings.Contains(prompt, "Automation is a lie") {
				// Literal claim path: return a retrieved source so refuted verdict stands.
				rs := []retrievedSource{{Title: "refutation source", URL: "https://example.com/refuted"}}
				return `{"verdict":"refuted","finding":"literal reading refuted","sources":[{"title":"refutation source","url":"https://example.com/refuted"}]}`, rs, nil
			}
			// Intended proposition path: return a retrieved source so supported verdict stands.
			rs := []retrievedSource{{Title: "paradox study", URL: "https://example.com/paradox"}}
			return `{"verdict":"supported","finding":"automation paradox documented","sources":[{"title":"paradox study","url":"https://example.com/paradox"}]}`, rs, nil
		default:
			return `[]`, nil, nil
		}
	}
	c := cfg{maxRounds: 1, call: stub}

	// Test claim: overstated → intendedProposition substitution should fire.
	claims := []string{"Automation is a lie."}
	_, _, es := c.computeAudit(claims, "full transcript")
	if es[0].Verdict != "supported" {
		t.Errorf("audit grounding: want supported (intended proposition), got %q", es[0].Verdict)
	}
	if es[0].Claim != "Automation is a lie." {
		t.Errorf("display claim should be original literal, got %q", es[0].Claim)
	}
	if !strings.Contains(groundedProposition, "automation always needs a human in the loop") {
		t.Errorf("intended proposition should be grounded, got prompt: %q", groundedProposition)
	}
}

// TestAuditFaithfulClaimGroundsLiteral is the control: a claim where faithfulness returns faithful
// (no SourceSays) must be grounded against the literal claim.
func TestAuditFaithfulClaimGroundsLiteral(t *testing.T) {
	var groundedProposition string
	stub := func(system, prompt string, withTools bool) (string, []retrievedSource, error) {
		switch {
		case strings.Contains(system, "You are the Producer"):
			return `{"steelman":"strongest version","conditions":"some conditions"}`, nil, nil
		case strings.Contains(system, "You are the Critic"):
			return `{"critique":[],"verdict":"partial","surviving_claim":null,"reason":"ok",` +
				`"needs_another_round":false,"added_conditions":0,"survives_only_by_conditioning":false}`, nil, nil
		case strings.Contains(system, "You are the Defender"):
			return `{"found":true,"quotes":["SaaS is not dead"],"best_case":"direct"}`, nil, nil
		case strings.Contains(system, "You are the Faithfulness Critic"):
			return `{"findings":[],"verdict":"faithful","evidence":"direct","what_source_actually_says":null}`, nil, nil
		case strings.Contains(system, "You are the Evidence Grounder"):
			groundedProposition = prompt
			rs := []retrievedSource{{Title: "SaaS market report", URL: "https://example.com/saas"}}
			return `{"verdict":"supported","finding":"SaaS growing","sources":[{"title":"SaaS market report","url":"https://example.com/saas"}]}`, rs, nil
		default:
			return `[]`, nil, nil
		}
	}
	c := cfg{maxRounds: 1, call: stub}
	claims := []string{"SaaS is not dead."}
	_, _, es := c.computeAudit(claims, "full transcript")
	if !strings.Contains(groundedProposition, "SaaS is not dead") {
		t.Errorf("faithful claim should ground literal, got prompt: %q", groundedProposition)
	}
	if es[0].Verdict != "supported" {
		t.Errorf("want supported, got %q", es[0].Verdict)
	}
}

// TestFaithCriticSysLiteralization confirms the Literalization distortion mode is present in the
// faithfulness critic prompt — it's the key fix for #10.
func TestFaithCriticSysLiteralization(t *testing.T) {
	for _, needle := range []string{
		"Literalization",
		"provocation, hyperbole, or irony",
		"including the register and",
		"what_source_actually_says",
	} {
		if !strings.Contains(faithCriticSys, needle) {
			t.Errorf("faithCriticSys missing %q", needle)
		}
	}
}

// TestSubstanceCriticSysConditionDiscipline confirms the condition-laundering instructions are
// present in the substance critic prompt — key fix for #1/#2.
func TestSubstanceCriticSysConditionDiscipline(t *testing.T) {
	for _, needle := range []string{
		"CONDITION DISCIPLINE",
		"Condition laundering",
		"Legitimate narrowing",
		"survives_only_by_conditioning",
		"added_conditions",
	} {
		if !strings.Contains(substanceCriticSys, needle) {
			t.Errorf("substanceCriticSys missing %q", needle)
		}
	}
}

func TestMarkdownSubstance(t *testing.T) {
	rs := []substance{
		{Claim: "CLIs are over", Verdict: "hollow", Reason: "unfalsifiable"},
		{Claim: "SaaS not dead", Verdict: "substantive", Reason: "checkable"},
	}
	out := mdSubstance(rs)
	for _, needle := range []string{
		"| # | Claim | Verdict | Why |",
		"**hollow**",
		"substantive",
		"1 hollow · 1 substantive"} {
		if !strings.Contains(out, needle) {
			t.Errorf("missing %q in:\n%s", needle, out)
		}
	}
}

func TestMarkdownAudit(t *testing.T) {
	claims := []string{"CLIs are over"}
	out := mdAudit(claims,
		[]faith{{Verdict: "overstated"}},
		[]substance{{Verdict: "partial"}},
		[]evidence{{Verdict: "refuted"}})
	for _, needle := range []string{
		"Faithful? | Substantive? | Grounded?",
		"**overstated**",
		"**refuted**",
		"Pattern-reading",
		"Anti-signal"} {
		if !strings.Contains(out, needle) {
			t.Errorf("missing %q in:\n%s", needle, out)
		}
	}
}

func TestMdCellPipeEscape(t *testing.T) {
	const in = `claim with | pipe and
newline`
	got := mdCell(in)
	if strings.Contains(got, "\n") {
		t.Error("mdCell should remove newlines")
	}
	if !strings.Contains(got, `\|`) {
		t.Error("mdCell should escape pipes")
	}
}

func TestTally(t *testing.T) {
	got := tally([]string{"hollow", "partial", "hollow", "substantive"})
	if !strings.Contains(got, "2 hollow") {
		t.Errorf("want '2 hollow' in %q", got)
	}
	// insertion order: hollow first
	if !strings.HasPrefix(got, "2 hollow") {
		t.Errorf("want hollow first: %q", got)
	}
}

// TestMaxClaimsCap drives the REAL c.runEvidence with maxClaims=2 and 5 input claims.
// Stubs cfg.call to count evidence-grounding invocations. Asserts exactly 2 grounding
// calls fire and the remaining 3 rows are marked "skipped (over cap)".
func TestMaxClaimsCap(t *testing.T) {
	evidenceCallCount := 0
	stub := func(system, prompt string, withTools bool) (string, []retrievedSource, error) {
		if strings.Contains(system, "You are the Evidence Grounder") {
			evidenceCallCount++
		}
		switch {
		case strings.Contains(system, "You are the Defender"):
			return `{"found":true,"quotes":["claim text"],"best_case":"direct"}`, nil, nil
		case strings.Contains(system, "You are the Faithfulness Critic"):
			return `{"findings":[],"verdict":"faithful","evidence":"direct","what_source_actually_says":null}`, nil, nil
		case strings.Contains(system, "You are the Evidence Grounder"):
			return `{"verdict":"supported","finding":"evidence found","sources":[]}`, nil, nil
		default:
			return `[]`, nil, nil
		}
	}
	c := cfg{maxClaims: 2, call: stub}
	input := "1. Claim A.\n2. Claim B.\n3. Claim C.\n4. Claim D.\n5. Claim E."
	c.runEvidence(input, "")
	if evidenceCallCount != 2 {
		t.Errorf("evidence grounder should be called exactly 2 times, got %d", evidenceCallCount)
	}
}

// TestMaxClaimsCapComputeAudit drives the REAL c.computeAudit with maxClaims=3 and
// N+3 claims. Stubs cfg.call to count evidence grounding calls. Asserts exactly 3
// calls fire and the remaining rows are marked "skipped (over cap)".
func TestMaxClaimsCapComputeAudit(t *testing.T) {
	evidenceCallCount := 0
	stub := func(system, prompt string, withTools bool) (string, []retrievedSource, error) {
		if strings.Contains(system, "You are the Evidence Grounder") {
			evidenceCallCount++
		}
		switch {
		case strings.Contains(system, "You are the Producer"):
			return `{"steelman":"strongest version","conditions":"some conditions"}`, nil, nil
		case strings.Contains(system, "You are the Critic"):
			return `{"critique":[],"verdict":"partial","surviving_claim":null,"reason":"ok",` +
				`"needs_another_round":false,"added_conditions":0,"survives_only_by_conditioning":false}`, nil, nil
		case strings.Contains(system, "You are the Defender"):
			return `{"found":true,"quotes":["claim text"],"best_case":"direct"}`, nil, nil
		case strings.Contains(system, "You are the Faithfulness Critic"):
			return `{"findings":[],"verdict":"faithful","evidence":"direct","what_source_actually_says":null}`, nil, nil
		case strings.Contains(system, "You are the Evidence Grounder"):
			rs := []retrievedSource{{Title: "evidence source", URL: "https://example.com/evidence"}}
			return `{"verdict":"supported","finding":"evidence found","sources":[{"title":"evidence source","url":"https://example.com/evidence"}]}`, rs, nil
		default:
			return `[]`, nil, nil
		}
	}
	c := cfg{maxRounds: 1, maxClaims: 3, call: stub}
	claims := []string{"Claim 1", "Claim 2", "Claim 3", "Claim 4", "Claim 5", "Claim 6"}
	_, _, es := c.computeAudit(claims, "")
	if evidenceCallCount != 3 {
		t.Errorf("evidence grounder should be called exactly 3 times, got %d", evidenceCallCount)
	}
	// Check that claims 0-2 were grounded and 3-5 are skipped
	for i := 0; i < 3; i++ {
		if es[i].Verdict != "supported" {
			t.Errorf("claim %d: want supported, got %q", i, es[i].Verdict)
		}
	}
	for i := 3; i < 6; i++ {
		if es[i].Verdict != "skipped (over cap)" {
			t.Errorf("claim %d: want 'skipped (over cap)', got %q", i, es[i].Verdict)
		}
	}
}

// ── crossCheckEvidence unit tests ────────────────────────────────────────────

// TestCrossCheckEvidenceClaimedPresent: a model-claimed URL that appears in the retrieved
// set should leave the verdict unchanged and record SourcesVerified=1.
func TestCrossCheckEvidenceClaimedPresent(t *testing.T) {
	e := evidence{
		Claim: "test claim", Verdict: "supported", Finding: "found",
		Sources:          []source{{Title: "Test Page", URL: "https://example.com/article"}},
		RetrievedSources: []retrievedSource{{Title: "Test Page", URL: "https://example.com/article"}},
	}
	got := crossCheckEvidence(e)
	if got.Verdict != "supported" {
		t.Errorf("want supported (cross-check passed), got %q", got.Verdict)
	}
	if got.SourcesVerified != 1 {
		t.Errorf("want SourcesVerified=1, got %d", got.SourcesVerified)
	}
	if got.DowngradeReason != "" {
		t.Errorf("want no downgrade, got reason %q", got.DowngradeReason)
	}
}

// TestCrossCheckEvidenceClaimedAbsent: a model-claimed URL absent from the retrieved set
// should be downgraded to "unverifiable".
func TestCrossCheckEvidenceClaimedAbsent(t *testing.T) {
	e := evidence{
		Claim: "test claim", Verdict: "supported", Finding: "found",
		Sources:          []source{{Title: "Claimed Page", URL: "https://claimed.example.com/article"}},
		RetrievedSources: []retrievedSource{{Title: "Other Page", URL: "https://other.example.com/stuff"}},
	}
	got := crossCheckEvidence(e)
	if got.Verdict != "unverifiable" {
		t.Errorf("want unverifiable (claimed URL absent from retrieval), got %q", got.Verdict)
	}
	if got.OriginalVerdict != "supported" {
		t.Errorf("want OriginalVerdict=supported, got %q", got.OriginalVerdict)
	}
	if !strings.Contains(got.DowngradeReason, "claimed sources not present in retrieval") {
		t.Errorf("want downgrade reason about claimed sources, got %q", got.DowngradeReason)
	}
}

// TestCrossCheckEvidenceZeroRetrieved: no web search results at all should downgrade
// any verdict that asserts external evidence.
func TestCrossCheckEvidenceZeroRetrieved(t *testing.T) {
	for _, verdict := range []string{"supported", "mixed", "refuted"} {
		e := evidence{
			Claim: "test claim", Verdict: verdict, Finding: "finding",
			Sources:          []source{{Title: "Some Page", URL: "https://example.com"}},
			RetrievedSources: nil,
		}
		got := crossCheckEvidence(e)
		if got.Verdict != "unverifiable" {
			t.Errorf("verdict=%s: want unverifiable (zero retrieved), got %q", verdict, got.Verdict)
		}
		if !strings.Contains(got.DowngradeReason, "no sources retrieved") {
			t.Errorf("verdict=%s: want 'no sources retrieved' in reason, got %q", verdict, got.DowngradeReason)
		}
		if got.OriginalVerdict != verdict {
			t.Errorf("verdict=%s: want OriginalVerdict=%s, got %q", verdict, verdict, got.OriginalVerdict)
		}
	}
}

// TestCrossCheckEvidenceZeroSources: retrieval succeeded but model cited no URLs — the
// truth-maker is missing even though a search happened.
func TestCrossCheckEvidenceZeroSources(t *testing.T) {
	e := evidence{
		Claim: "test claim", Verdict: "mixed", Finding: "cuts both ways",
		Sources:          nil,
		RetrievedSources: []retrievedSource{{Title: "Retrieved Page", URL: "https://example.com/data"}},
	}
	got := crossCheckEvidence(e)
	if got.Verdict != "unverifiable" {
		t.Errorf("want unverifiable (no URLs cited), got %q", got.Verdict)
	}
	if !strings.Contains(got.DowngradeReason, "no URLs cited in response") {
		t.Errorf("want 'no URLs cited in response' in reason, got %q", got.DowngradeReason)
	}
}

// TestCrossCheckEvidenceExemptions: "unverifiable" and "error" verdicts must pass
// through unchanged — they make no external-evidence assertion.
func TestCrossCheckEvidenceExemptions(t *testing.T) {
	for _, verdict := range []string{"unverifiable", "error"} {
		e := evidence{Claim: "c", Verdict: verdict, Finding: "f"}
		got := crossCheckEvidence(e)
		if got.Verdict != verdict {
			t.Errorf("%s should be exempt from cross-check, got %q", verdict, got.Verdict)
		}
		if got.DowngradeReason != "" {
			t.Errorf("%s: unexpected downgrade reason %q", verdict, got.DowngradeReason)
		}
	}
}

// ── normalizeURL unit tests ───────────────────────────────────────────────────

func TestNormalizeURL(t *testing.T) {
	cases := []struct{ in, want string }{
		{"https://example.com/article", "example.com/article"},
		{"http://example.com/article", "example.com/article"},
		{"https://example.com/article/", "example.com/article"},
		{"HTTPS://Example.COM/Article", "example.com/Article"},
		{"https://example.com/page?utm_source=x&ref=y", "example.com/page"},
		{"https://example.com/page#section", "example.com/page"},
		{"https://example.com", "example.com"},
		{"https://example.com/", "example.com"},
	}
	for _, tc := range cases {
		if got := normalizeURL(tc.in); got != tc.want {
			t.Errorf("normalizeURL(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// finitVerbRe matches common English finite verbs — used by the structural guard in
// TestFixtureIngestion to detect non-claim-shaped segments.
var finiteVerbRe = regexp.MustCompile(`(?i)\b(was|were|is|are|had|has|have|do|does|did|` +
	`drove|increased|decreased|grew|fell|rose|reached|expects?|believes?|saw|made|added|` +
	`continued|launched|closed|expanded|achieved|contributed|included|required|completed|` +
	`provides?|offers?|enabled|scaled|improved|extended|adopted|generated|brought|` +
	`maintained|executed|delivered|reported|gained|held|remained|declined|acquired|` +
	`boosted|converted|supported|showed|reflects?|represents?|will|would|should|may|might|can|could)\b`)

// isNonClaimShaped returns true when a segment is structurally a label, table cell, fragment,
// or bare number rather than a natural-language claim. The structural guard in TestFixtureIngestion
// uses this to catch fixtures that haven't been properly normalized.
func isNonClaimShaped(seg string) bool {
	words := strings.Fields(seg)
	if len(words) < 4 {
		return true
	}
	// All-uppercase (table column header or label)
	allCaps := true
	for _, w := range words {
		if strings.IndexFunc(w, func(r rune) bool { return r >= 'a' && r <= 'z' }) >= 0 {
			allCaps = false
			break
		}
	}
	if allCaps {
		return true
	}
	// No finite verb → bare noun phrase or label
	if !finiteVerbRe.MatchString(seg) {
		return true
	}
	return false
}

// TestFixtureIngestion iterates fixtures/raw and passes each real file through splitSummary.
// Goal: confirm the regex boundaries don't panic or hang on large, heterogeneous real-world inputs,
// and that fixture normalization has removed table cells and hard-wrapped fragments.
// Expected files that are missing are t.Skip'd — never substituted.
func TestFixtureIngestion(t *testing.T) {
	const (
		dir            = "fixtures/raw"
		maxSegSize     = 4096 // no legitimate atomic-claim line exceeds 4 KB; table blobs do
		maxNonClaimPct = 15   // >15% non-claim-shaped segments → fixture needs re-normalization
	)

	// The canonical set. Each missing file gets its own skip, not a test failure.
	expected := []string{
		"legal.md",
		"accounting.md",
		"sales.md",
		"marketing.md",
		"pm.md",
		"engineering.md",
		"contract.md",
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Skipf("fixtures/raw not present — run fetch script to populate")
	}

	for _, name := range expected {
		name := name
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(dir, name)
			data, err := os.ReadFile(path)
			if os.IsNotExist(err) {
				t.Skipf("%s not present — fetch failed or not yet attempted", name)
			}
			if err != nil {
				t.Fatalf("ReadFile %s: %v", path, err)
			}
			if len(data) == 0 {
				t.Skipf("%s is empty — fetch may have failed silently", name)
			}
			content := string(data)
			got := splitSummary(content)
			if len(got) == 0 {
				t.Errorf("%s: splitSummary returned 0 items on %d-byte input", name, len(data))
				return
			}
			// Sanity guard: no single segment may exceed maxSegSize.
			// A giant segment means the file contains table blobs or unbroken prose
			// that splitSummary can't decompose — the fixture is not usable as input.
			var maxSeg int
			for _, seg := range got {
				if len(seg) > maxSeg {
					maxSeg = len(seg)
				}
				if len(seg) > maxSegSize {
					t.Errorf("%s: segment of %d bytes exceeds %d-byte limit — fixture likely contains table blobs or unbroken runs; re-extract",
						name, len(seg), maxSegSize)
				}
			}
			var nonClaim int
			for _, seg := range got {
				if isNonClaimShaped(seg) {
					nonClaim++
				}
			}
			nonClaimPct := nonClaim * 100 / len(got)
			if nonClaimPct > maxNonClaimPct {
				t.Errorf("%s: %d%% of segments are non-claim-shaped (threshold %d%%) — fixture needs re-normalization",
					name, nonClaimPct, maxNonClaimPct)
			}
			avgSeg := len(data) / len(got)
			t.Logf("%s: %d bytes → %d segments, avg %d b/seg, max seg %d b, non-claim %d%% (first: %.80q)",
				name, len(data), len(got), avgSeg, maxSeg, nonClaimPct, strings.TrimSpace(got[0]))
		})
	}
}

// ── usage accounting tests ────────────────────────────────────────────────────

// TestUsageAccumulatorSumsCorrectly asserts the usageCounters accumulate correctly. Dispatch only
// bills usage on the backend path, not the cfg.call stub seam, so we drive the accumulator directly.
func TestUsageAccumulatorSumsCorrectly(t *testing.T) {
	u := newUsageCounters()
	u.add(100, 50, 10, 5, 1)
	u.add(200, 80, 20, 8, 2)

	calls, in, out, cacheRead, cacheCreate, webSearches, _, _ := u.snapshot()
	if calls != 2 {
		t.Errorf("calls: want 2, got %d", calls)
	}
	if in != 300 {
		t.Errorf("inputTokens: want 300, got %d", in)
	}
	if out != 130 {
		t.Errorf("outputTokens: want 130, got %d", out)
	}
	if cacheRead != 30 {
		t.Errorf("cacheReadTokens: want 30, got %d", cacheRead)
	}
	if cacheCreate != 13 {
		t.Errorf("cacheCreateTokens: want 13, got %d", cacheCreate)
	}
	if webSearches != 3 {
		t.Errorf("webSearches: want 3, got %d", webSearches)
	}
}

// TestStdoutPurity drives a stubbed runEvidence and asserts that stdout carries only the markdown
// table — no progress or USAGE lines — protecting -md harness integrity. stderr is not checked here
// since progress intentionally goes there.
func TestStdoutPurity(t *testing.T) {
	// Capture stdout.
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	stub := func(system, prompt string, withTools bool) (string, []retrievedSource, error) {
		switch {
		case strings.Contains(system, "You are the Defender"):
			return `{"found":true,"quotes":["q"],"best_case":"direct"}`, nil, nil
		case strings.Contains(system, "You are the Faithfulness Critic"):
			return `{"findings":[],"verdict":"faithful","evidence":"e","what_source_actually_says":null}`, nil, nil
		case strings.Contains(system, "You are the Evidence Grounder"):
			return `{"verdict":"supported","finding":"found","sources":[]}`, nil, nil
		default:
			return `[]`, nil, nil
		}
	}
	c := cfg{
		call:         stub,
		asMarkdown:   true,
		showProgress: true, // progress ON — but it must go to stderr, not stdout
		usage:        newUsageCounters(),
	}
	c.runEvidence("1. Claim one.\n2. Claim two.", "")

	w.Close()
	os.Stdout = origStdout

	captured, _ := io.ReadAll(r)
	out := string(captured)

	// stdout must contain the markdown table header
	if !strings.Contains(out, "## Grounding") {
		t.Errorf("stdout should contain markdown table, got: %.200q", out)
	}
	// stdout must NOT contain progress, per-case completion lines, or USAGE lines.
	for _, bad := range []string{"✓", "✗", "USAGE ", "[heartbeat]", "SUMMARY "} {
		if strings.Contains(out, bad) {
			t.Errorf("stdout contains progress/USAGE marker %q — violates stdout purity", bad)
		}
	}
}

// TestPriceLookupKnownModel checks that a known model returns a numeric estimate.
func TestPriceLookupKnownModel(t *testing.T) {
	est, ok := estimateCost("claude-sonnet-4-6", 1_000_000, 1_000_000, 0, 0, 0)
	if !ok {
		t.Errorf("want ok=true for known model, got false; est=%q", est)
	}
	if !strings.HasPrefix(est, "$") {
		t.Errorf("want dollar-prefixed estimate, got %q", est)
	}
	if !strings.Contains(est, priceTableDate) {
		t.Errorf("want price table date %q in estimate %q", priceTableDate, est)
	}
}

// TestPriceLookupUnknownModel checks that an unknown model returns tokens-present, cost flagged n/a.
func TestPriceLookupUnknownModel(t *testing.T) {
	est, ok := estimateCost("claude-unknown-9999", 500, 200, 0, 0, 0)
	if ok {
		t.Errorf("want ok=false for unknown model, got true; est=%q", est)
	}
	if !strings.Contains(est, "n/a") {
		t.Errorf("want 'n/a' in estimate for unknown model, got %q", est)
	}
	// Measured token counts must appear in the string so the caller can still see them.
	if !strings.Contains(est, "in=500") {
		t.Errorf("want input token count in estimate, got %q", est)
	}
}

// TestDecomposeReturnsError confirms that a callJSON failure propagates out of
// decompose as a non-nil error rather than a silent nil slice.
func TestDecomposeReturnsError(t *testing.T) {
	c := cfg{
		call: func(system, prompt string, withTools bool) (string, []retrievedSource, error) {
			return "", nil, fmt.Errorf("API down")
		},
	}
	_, err := c.decompose("some text")
	if err == nil {
		t.Fatal("expected decompose to return error, got nil")
	}
	if !strings.Contains(err.Error(), "API down") {
		t.Errorf("expected wrapped error to contain 'API down', got: %v", err)
	}
}

// TestPriceLookupUnsetPrices verifies that a model entry with InputPerMtok<0 prints the TODO
// sentinel rather than a fabricated number.
func TestPriceLookupUnsetPrices(t *testing.T) {
	// Temporarily install a model entry with unset prices.
	priceTable["__test_unset__"] = priceEntry{InputPerMtok: -1}
	defer delete(priceTable, "__test_unset__")

	est, ok := estimateCost("__test_unset__", 100, 50, 0, 0, 0)
	if ok {
		t.Errorf("want ok=false for unset prices, got true; est=%q", est)
	}
	if !strings.Contains(est, "TODO") {
		t.Errorf("want TODO sentinel for unset prices, got %q", est)
	}
}

// TestUsageOutFile verifies that printUsageLine appends a valid JSON record when -usage-out is set.
func TestUsageOutFile(t *testing.T) {
	tmp, err := os.CreateTemp(t.TempDir(), "usage*.jsonl")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	tmp.Close()

	u := newUsageCounters()
	u.add(1000, 500, 100, 50, 2)
	printUsageLine("claude-sonnet-4-6", u, tmp.Name())

	data, err := os.ReadFile(tmp.Name())
	if err != nil {
		t.Fatalf("read usage file: %v", err)
	}
	line := strings.TrimSpace(string(data))
	if line == "" {
		t.Fatal("usage file is empty")
	}
	var rec struct {
		Model        string  `json:"model"`
		Calls        int     `json:"calls"`
		InputTokens  int     `json:"input_tokens"`
		OutputTokens int     `json:"output_tokens"`
		WebSearches  int     `json:"web_searches"`
		EstUSD       *string `json:"est_usd"`
		WallSeconds  float64 `json:"wall_seconds"`
	}
	if err := json.Unmarshal([]byte(line), &rec); err != nil {
		t.Fatalf("unmarshal usage record: %v; raw=%q", err, line)
	}
	if rec.Model != "claude-sonnet-4-6" {
		t.Errorf("model: want claude-sonnet-4-6, got %q", rec.Model)
	}
	if rec.Calls != 1 {
		t.Errorf("calls: want 1, got %d", rec.Calls)
	}
	if rec.InputTokens != 1000 {
		t.Errorf("input_tokens: want 1000, got %d", rec.InputTokens)
	}
	if rec.EstUSD == nil {
		t.Error("est_usd should be non-null for known model")
	} else if !strings.HasPrefix(*rec.EstUSD, "$") {
		t.Errorf("est_usd should start with $, got %q", *rec.EstUSD)
	}
}

// ── Tier-1/Tier-2 reporting tests ─────────────────────────────────────────────

// TestProgressOneLine confirms that runEvidence emits exactly one stderr line per completed case
// (no pre-call line) and that the line format matches the tier-1 spec.
func TestProgressOneLine(t *testing.T) {
	origStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stderr = w

	stub := func(system, prompt string, withTools bool) (string, []retrievedSource, error) {
		switch {
		case strings.Contains(system, "You are the Evidence Grounder"):
			rs := []retrievedSource{{Title: "progress source", URL: "https://example.com/progress"}}
			return `{"verdict":"supported","finding":"found","sources":[{"title":"progress source","url":"https://example.com/progress"}]}`, rs, nil
		default:
			return `[]`, nil, nil
		}
	}
	c := cfg{
		call:         stub,
		showProgress: true,
		usage:        newUsageCounters(),
	}
	c.runEvidence("1. Claim one.\n2. Claim two.", "")

	w.Close()
	os.Stderr = origStderr

	captured, _ := io.ReadAll(r)
	lines := strings.Split(strings.TrimSpace(string(captured)), "\n")

	// Filter to just the per-case completion lines (start with "[").
	var caseLines []string
	for _, l := range lines {
		if strings.HasPrefix(l, "[") && (strings.Contains(l, "✓") || strings.Contains(l, "✗")) {
			caseLines = append(caseLines, l)
		}
	}
	if len(caseLines) != 2 {
		t.Errorf("want exactly 2 per-case lines, got %d; stderr:\n%s", len(caseLines), string(captured))
	}
	// No pre-call line: lines must not contain "claim N…" pattern.
	for _, l := range lines {
		if strings.Contains(l, "claim ") && strings.HasSuffix(strings.TrimSpace(l), "…") {
			t.Errorf("found pre-call line in stderr: %q", l)
		}
	}
	// Each case line must carry the verdict word.
	for _, l := range caseLines {
		if !strings.Contains(l, "supported") {
			t.Errorf("case line missing verdict: %q", l)
		}
	}
}

// TestRunTallyCorrect confirms verified/errored/total counts from a mix of real verdicts and errors.
func TestRunTallyCorrect(t *testing.T) {
	rt := newRunTally(5)
	rt.record("supported")
	rt.record("mixed")
	rt.record("error")
	rt.record("refuted")
	rt.record("error")

	total, verified, counts := rt.snapshot()
	if total != 5 {
		t.Errorf("total: want 5, got %d", total)
	}
	if verified != 3 {
		t.Errorf("verified: want 3, got %d", verified)
	}
	errored := total - verified
	if errored != 2 {
		t.Errorf("errored: want 2, got %d", errored)
	}
	if counts["error"] != 2 {
		t.Errorf("counts[error]: want 2, got %d", counts["error"])
	}
	if counts["supported"] != 1 {
		t.Errorf("counts[supported]: want 1, got %d", counts["supported"])
	}
}

// TestChainJSONLWritten verifies that appendChain writes a valid JSONL record for a grounding case,
// including error_cause on an errored case.
func TestChainJSONLWritten(t *testing.T) {
	dir := t.TempDir()
	chainPath := filepath.Join(dir, "test.grounding.jsonl")

	c := cfg{chainFile: chainPath, usage: newUsageCounters()}

	// Write a normal supported case.
	ev1 := evidence{Claim: "claim one", Verdict: "supported", Finding: "confirmed by data"}
	c.appendChain(evidenceChainRecord(0, 2, "claim one", ev1, time.Now().Add(-2*time.Second)))

	// Write an errored case.
	ev2 := evidence{Claim: "claim two", Verdict: "error", Finding: "network timeout"}
	c.appendChain(evidenceChainRecord(1, 2, "claim two", ev2, time.Now().Add(-1*time.Second)))

	data, err := os.ReadFile(chainPath)
	if err != nil {
		t.Fatalf("read chain file: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("want 2 JSONL records, got %d; file:\n%s", len(lines), string(data))
	}

	// First record: supported.
	var rec1 chainRecord
	if err := json.Unmarshal([]byte(lines[0]), &rec1); err != nil {
		t.Fatalf("unmarshal rec1: %v; raw=%q", err, lines[0])
	}
	if rec1.Verdict != "supported" {
		t.Errorf("rec1 verdict: want supported, got %q", rec1.Verdict)
	}
	if rec1.Mode != "grounding" {
		t.Errorf("rec1 mode: want grounding, got %q", rec1.Mode)
	}
	if rec1.ElapsedS <= 0 {
		t.Errorf("rec1 elapsed_s should be positive, got %f", rec1.ElapsedS)
	}

	// Second record: error — detail must carry error_cause.
	var rec2 chainRecord
	if err := json.Unmarshal([]byte(lines[1]), &rec2); err != nil {
		t.Fatalf("unmarshal rec2: %v; raw=%q", err, lines[1])
	}
	if rec2.Verdict != "error" {
		t.Errorf("rec2 verdict: want error, got %q", rec2.Verdict)
	}
	var det evidenceDetail
	if err := json.Unmarshal(rec2.Detail, &det); err != nil {
		t.Fatalf("unmarshal rec2 detail: %v", err)
	}
	if det.ErrorCause != "network timeout" {
		t.Errorf("rec2 error_cause: want 'network timeout', got %q", det.ErrorCause)
	}
}

// TestReadChainOrdersAndValidates covers the happy path (idx-ordered, one mode) plus the two refused
// shapes: a gap in the idx sequence and mixed modes.
func TestReadChainOrdersAndValidates(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "good.jsonl")
	os.WriteFile(good, []byte(
		`{"idx":1,"total":2,"mode":"faithfulness","claim":"b","verdict":"faithful","detail":{}}`+"\n"+
			`{"idx":0,"total":2,"mode":"faithfulness","claim":"a","verdict":"overstated","detail":{}}`+"\n"), 0o644)
	recs, mode, err := readChain(good)
	if err != nil {
		t.Fatalf("readChain: %v", err)
	}
	if mode != "faithfulness" || len(recs) != 2 || recs[0].Claim != "a" || recs[1].Claim != "b" {
		t.Errorf("bad reconstruction: mode=%q recs=%+v", mode, recs)
	}

	gap := filepath.Join(dir, "gap.jsonl")
	os.WriteFile(gap, []byte(
		`{"idx":0,"mode":"faithfulness","claim":"a","verdict":"faithful","detail":{}}`+"\n"+
			`{"idx":2,"mode":"faithfulness","claim":"c","verdict":"faithful","detail":{}}`+"\n"), 0o644)
	if _, _, err := readChain(gap); err == nil {
		t.Error("want error on idx gap, got nil")
	}

	mixed := filepath.Join(dir, "mixed.jsonl")
	os.WriteFile(mixed, []byte(
		`{"idx":0,"mode":"faithfulness","claim":"a","verdict":"faithful","detail":{}}`+"\n"+
			`{"idx":1,"mode":"substance","claim":"b","verdict":"hollow","detail":{}}`+"\n"), 0o644)
	if _, _, err := readChain(mixed); err == nil {
		t.Error("want error on mixed modes, got nil")
	}
}

// TestParseAuditVerdict pins the round-trip of auditChainRecord's "faith=X sub=Y ev=Z" verdict.
func TestParseAuditVerdict(t *testing.T) {
	f, s, e := parseAuditVerdict("faith=overstated sub=hollow ev=unverifiable")
	if f != "overstated" || s != "hollow" || e != "unverifiable" {
		t.Errorf("got faith=%q sub=%q ev=%q", f, s, e)
	}
}

// TestHeartbeatErroredIsFailuresNotRemainder pins the 1c fix: mid-run, errored counts actual failed
// cases, never the not-yet-processed remainder. With 10 total, 3 verified and 1 error seen, errored
// must be 1 (not 10-3=7) and seen must be 4.
func TestHeartbeatErroredIsFailuresNotRemainder(t *testing.T) {
	rt := newRunTally(10)
	rt.record("faithful")
	rt.record("partial")
	rt.record("faithful")
	rt.record("error")
	line := heartbeatLine("claude-sonnet-4-6", newUsageCounters(), rt)
	for _, want := range []string{"verified=3", "errored=1", "seen=4/10"} {
		if !strings.Contains(line, want) {
			t.Errorf("heartbeat line missing %q\n  got: %s", want, line)
		}
	}
}

// TestModalVerdict pins the count winner and the worst-first tie-break used when a repeat run splits.
func TestModalVerdict(t *testing.T) {
	if v := modalVerdict(map[string]int{"faithful": 1, "partial": 1}); v != "partial" {
		t.Errorf("a tie should break to the worse verdict (partial), got %q", v)
	}
	if v := modalVerdict(map[string]int{"faithful": 2, "partial": 1}); v != "faithful" {
		t.Errorf("majority faithful, got %q", v)
	}
}

// TestFaithRepeatSpread drives faithRepeat with a critic that returns partial, faithful, partial
// across three runs; the modal verdict must be partial with agreement 2/3, and the returned faith
// must be a run that produced partial (so its reason matches).
func TestFaithRepeatSpread(t *testing.T) {
	critics := []string{
		`{"findings":[],"verdict":"partial","evidence":"first-partial","what_source_actually_says":"x"}`,
		`{"findings":[],"verdict":"faithful","evidence":"the-faithful","what_source_actually_says":null}`,
		`{"findings":[],"verdict":"partial","evidence":"second-partial","what_source_actually_says":"y"}`,
	}
	ci := 0
	stub := func(system, prompt string, withTools bool) (string, []retrievedSource, error) {
		switch {
		case strings.Contains(system, "You are the Defender"):
			return `{"found":true,"quotes":["q"]}`, nil, nil
		case strings.Contains(system, "You are the Faithfulness Critic"):
			r := critics[ci%len(critics)]
			ci++
			return r, nil, nil
		}
		return "{}", nil, nil
	}
	c := cfg{repeat: 3, call: stub}
	got, spread := c.faithRepeat("claim", "src")
	if got.Verdict != "partial" {
		t.Errorf("modal verdict: want partial, got %q", got.Verdict)
	}
	if spread != "2/3" {
		t.Errorf("spread: want 2/3, got %q", spread)
	}
	if got.Evidence != "first-partial" {
		t.Errorf("returned run should be the first partial, got reason %q", got.Evidence)
	}
}

// TestSpreadFromSamples pins the reconstruction of the fraction and the dissent from a recorded
// sample list: a 2/3 split names its minority verdict, a unanimous run has no dissent, a lone sample
// (N=1) has no spread at all, and the modal is chosen worst-first on a count tie.
func TestSpreadFromSamples(t *testing.T) {
	cases := []struct {
		samples           []string
		wantSpread, wantD string
	}{
		{[]string{"partial", "partial", "faithful"}, "2/3", "faithful"},
		{[]string{"faithful", "faithful", "faithful"}, "3/3", ""},
		{[]string{"partial"}, "", ""},
		{nil, "", ""},
		{[]string{"faithful", "partial"}, "1/2", "faithful"}, // count tie → modal is worse (partial), faithful dissents
		{[]string{"faithful", "absent", "faithful"}, "2/3", "absent"},
	}
	for _, c := range cases {
		s, d := spreadFromSamples(c.samples)
		if s != c.wantSpread || d != c.wantD {
			t.Errorf("spreadFromSamples(%v) = (%q,%q), want (%q,%q)", c.samples, s, d, c.wantSpread, c.wantD)
		}
	}
}

// TestSpreadForRecord pins the -from rendering rule: a record carrying a samples list rebuilds its
// spread and dissent from that list, while a record written before samples existed falls back to the
// stored spread string with no dissent — the back-compat path that keeps old chains renderable.
func TestSpreadForRecord(t *testing.T) {
	withSamples := chainRecord{Verdict: "partial", Spread: "9/9", Samples: []string{"partial", "partial", "faithful"}}
	if s, d := spreadForRecord(withSamples); s != "2/3" || d != "faithful" {
		t.Errorf("record with samples: got (%q,%q), want (2/3,faithful) — spread must come from the list, not the stored %q", s, d, withSamples.Spread)
	}
	legacy := chainRecord{Verdict: "partial", Spread: "2/3"} // no Samples field
	if s, d := spreadForRecord(legacy); s != "2/3" || d != "" {
		t.Errorf("legacy record: got (%q,%q), want (2/3,\"\")", s, d)
	}
}

// TestStabilityClass pins the merged-run classifier: settled when every sample agrees, wobble when the
// verdict moves within one side of the support divide *and a majority still holds*, contested when it
// crosses the divide, mixes a decided verdict with unverifiable, or has no majority at all (a tie the
// merge cannot resolve into one verdict). spec/TREE.md § Merging runs is the contract.
func TestStabilityClass(t *testing.T) {
	cases := []struct {
		samples []string
		want    string
	}{
		{[]string{"faithful", "faithful", "faithful"}, "settled"},
		{[]string{"partial", "faithful", "faithful"}, "wobble"},               // one side, majority faithful
		{[]string{"absent", "contradicted", "unsupported"}, "contested"},      // one side, but no majority — a 3-way tie
		{[]string{"absent", "absent", "contradicted"}, "wobble"},              // one side, majority absent
		{[]string{"faithful", "faithful", "partial", "partial"}, "contested"}, // one side, 2/2 tie → split faithful/partial
		{[]string{"faithful", "faithful", "contradicted"}, "contested"},       // crosses the divide
		{[]string{"faithful", "unverifiable"}, "contested"},                   // decided vs can't-check
		{[]string{"unverifiable", "unverifiable"}, "settled"},
		{nil, ""},
	}
	for _, c := range cases {
		if got := stabilityClass(c.samples); got != c.want {
			t.Errorf("stabilityClass(%v) = %q, want %q", c.samples, got, c.want)
		}
	}
}

// TestSplitDescriptor pins the tie renderer: a pool with no majority yields its tied top verdicts
// worst-first as "a/b", and a pool with a clear winner yields "" (there is a verdict, so no split). A
// two-way even split and a three-way one-each split both tie; a majority does not.
func TestSplitDescriptor(t *testing.T) {
	cases := []struct {
		samples []string
		want    string
	}{
		{[]string{"faithful", "faithful", "partial", "partial"}, "partial/faithful"}, // worst-first
		{[]string{"absent", "contradicted", "unsupported"}, "unsupported/contradicted/absent"},
		{[]string{"faithful", "faithful", "partial"}, ""}, // majority faithful — not a tie
		{[]string{"faithful"}, ""},
		{nil, ""},
	}
	for _, c := range cases {
		if got := splitDescriptor(c.samples); got != c.want {
			t.Errorf("splitDescriptor(%v) = %q, want %q", c.samples, got, c.want)
		}
	}
}

// TestMergeChainsThreeClasses is the merge refuter: three chains over three leaves must produce
// exactly one settled, one wobble, one contested class, and each merged leaf's verdict must be the
// modal over the whole pool. It proves the merge distinguishes the three stability classes rather
// than collapsing them.
func TestMergeChainsThreeClasses(t *testing.T) {
	// Leaf 0 agrees across runs (settled); leaf 1 moves faithful↔partial, one side (wobble); leaf 2
	// crosses faithful↔contradicted (contested). N=1 chains, so each record contributes one sample.
	rec := func(idx int, claim, verdict string) chainRecord {
		return chainRecord{Idx: idx, Total: 3, Mode: "faithfulness", Claim: claim, Verdict: verdict, Detail: json.RawMessage(`{}`)}
	}
	chains := [][]chainRecord{
		{rec(0, "a", "faithful"), rec(1, "b", "faithful"), rec(2, "c", "faithful")},
		{rec(0, "a", "faithful"), rec(1, "b", "partial"), rec(2, "c", "faithful")},
		{rec(0, "a", "faithful"), rec(1, "b", "faithful"), rec(2, "c", "contradicted")},
	}
	merged, classes := mergeChains(chains)

	wantClass := []string{"settled", "wobble", "contested"}
	if strings.Join(classes, ",") != strings.Join(wantClass, ",") {
		t.Fatalf("classes = %v, want %v", classes, wantClass)
	}
	got := map[string]int{}
	for _, cl := range classes {
		got[cl]++
	}
	for _, want := range wantClass {
		if got[want] != 1 {
			t.Errorf("class %q appeared %d times, want exactly 1", want, got[want])
		}
	}
	// Modal verdict per leaf: faithful (3/3), faithful (2/3 over the pool), faithful (2/3 over the pool).
	wantVerdict := []string{"faithful", "faithful", "faithful"}
	for i, m := range merged {
		if m.Verdict != wantVerdict[i] {
			t.Errorf("leaf %d modal verdict = %q, want %q", i, m.Verdict, wantVerdict[i])
		}
		// The merged record carries the pooled samples so spreadForRecord recomputes the fraction.
		if s, _ := spreadForRecord(m); s != "2/3" && s != "3/3" {
			t.Errorf("leaf %d spread = %q, want a 3-sample fraction", i, s)
		}
	}
}

// TestFaithCriticSysGapAndSoWhat confirms the value-model additions (docs/VALUE.md) are in the
// faithfulness critic prompt: the structured gap field with all five gap classes plus none, and the
// so-what stakes line.
func TestFaithCriticSysGapAndSoWhat(t *testing.T) {
	for _, needle := range []string{
		`"gap":"none"|"scope"|"denominator"|"timerange"|"attribution"|"other"`,
		"SO WHAT", "<=20 words",
		`"so_what":string`,
	} {
		if !strings.Contains(faithCriticSys, needle) {
			t.Errorf("faithCriticSys missing %q", needle)
		}
	}
}

// TestFaithJudgeSysRestatementForm confirms the judge prompt pins report_says/source_says to two
// ≤12-word plain restatements, bans "reader"/"would", long imported words, and copied 3+-word runs,
// carries the "rephrase in everyday words" steer, and shows the compliant F46b worked example the
// code assembles into the stakes line (renderStakes).
func TestFaithJudgeSysRestatementForm(t *testing.T) {
	for _, needle := range []string{
		"REPORT_SAYS and SOURCE_SAYS", "<=12", "three syllables",
		`no "reader", no "would"`,
		"report_says must NOT copy a run of",
		"must NOT contrast or negate",
		`no "not", "only", "just"`,
		"as if to someone who hasn't read the report",
		"most of the work on those 52 projects was done in Victoria",
		"the projects were based in Victoria",
	} {
		if !strings.Contains(faithJudgeSys, needle) {
			t.Errorf("faithJudgeSys missing %q", needle)
		}
	}
}

// ── backend seam ──────────────────────────────────────────────────────────────

// TestDispatchThroughBackend confirms that with no cfg.call stub, dispatch routes through the
// configured backend and bills its usage — the production path the mode runners take.
func TestDispatchThroughBackend(t *testing.T) {
	u := newUsageCounters()
	be := fake.New("fake", "m", func(r backend.Request) (backend.Response, error) {
		if r.Prompt != "prompt" {
			t.Errorf("prompt not threaded: %q", r.Prompt)
		}
		return backend.Response{
			Text:  `{"claims":["a","b"]}`,
			Usage: backend.Usage{InputTokens: 4, OutputTokens: 2},
		}, nil
	})
	c := cfg{backend: be, backendName: "fake", usage: u}
	var got struct {
		Claims []string `json:"claims"`
	}
	if err := c.callJSON("sys", "", "prompt", false, &got); err != nil {
		t.Fatalf("callJSON: %v", err)
	}
	if len(got.Claims) != 2 {
		t.Fatalf("want 2 claims threaded from backend, got %+v", got.Claims)
	}
	calls, in, out, _, _, _, _, _ := u.snapshot()
	if calls != 1 || in != 4 || out != 2 {
		t.Errorf("backend usage not billed: calls=%d in=%d out=%d", calls, in, out)
	}
}

// TestHeaderLineNamesBackend confirms the provenance header stamps the backend, and falls back to a
// dash under -from where no backend was wired.
func TestHeaderLineNamesBackend(t *testing.T) {
	c := cfg{model: "m", backendName: "ollama", usage: newUsageCounters()}
	if !strings.Contains(c.headerLine(), "backend ollama") {
		t.Errorf("header should name the backend: %q", c.headerLine())
	}
	from := cfg{model: "m", usage: newUsageCounters()} // backendName "" → -from render
	if !strings.Contains(from.headerLine(), "backend —") {
		t.Errorf("from-render header should show a dash backend: %q", from.headerLine())
	}
}

// TestChainStampsBackend confirms appendChain records the producing backend on every case.
func TestChainStampsBackend(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x.faithfulness.jsonl")
	c := cfg{chainFile: p, backendName: "ollama"}
	c.appendChain(chainRecord{Idx: 0, Total: 1, Mode: "faithfulness", Claim: "a", Verdict: "faithful"})
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read chain: %v", err)
	}
	var rec chainRecord
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(data))), &rec); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if rec.Backend != "ollama" {
		t.Errorf("backend not stamped: %q", rec.Backend)
	}
}

// TestFromReplayBackendAgnostic is the refuter for "-from replays the chain identically on either
// backend": two chains with identical claims/verdicts but different producing backends reconstruct
// to the same ordered record sequence, and each keeps its own recorded backend.
func TestFromReplayBackendAgnostic(t *testing.T) {
	dir := t.TempDir()
	mk := func(be string) string {
		p := filepath.Join(dir, be+".faithfulness.jsonl")
		os.WriteFile(p, []byte(
			`{"idx":0,"total":2,"mode":"faithfulness","backend":"`+be+`","claim":"a","verdict":"faithful","detail":{}}`+"\n"+
				`{"idx":1,"total":2,"mode":"faithfulness","backend":"`+be+`","claim":"b","verdict":"overstated","detail":{}}`+"\n"), 0o644)
		return p
	}
	aRecs, aMode, aErr := readChain(mk("anthropic"))
	oRecs, oMode, oErr := readChain(mk("ollama"))
	if aErr != nil || oErr != nil {
		t.Fatalf("readChain errors: anthropic=%v ollama=%v", aErr, oErr)
	}
	if aMode != oMode || aMode != "faithfulness" {
		t.Errorf("modes differ: %q vs %q", aMode, oMode)
	}
	if len(aRecs) != len(oRecs) {
		t.Fatalf("record counts differ: %d vs %d", len(aRecs), len(oRecs))
	}
	for i := range aRecs {
		if aRecs[i].Claim != oRecs[i].Claim || aRecs[i].Verdict != oRecs[i].Verdict {
			t.Errorf("row %d differs across backends: %+v vs %+v", i, aRecs[i], oRecs[i])
		}
	}
	if aRecs[0].Backend != "anthropic" || oRecs[0].Backend != "ollama" {
		t.Errorf("producing backend not preserved: %q / %q", aRecs[0].Backend, oRecs[0].Backend)
	}
}
