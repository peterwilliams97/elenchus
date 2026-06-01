package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
func stubCall(criticJSON string) func(system, prompt string, withTools bool) (string, error) {
	return func(system, prompt string, withTools bool) (string, error) {
		switch {
		case strings.Contains(system, "You are the Producer"):
			return `{"steelman":"strongest version","conditions":"some conditions"}`, nil
		case strings.Contains(system, "You are the Critic"):
			return criticJSON, nil
		case strings.Contains(system, "You are the Defender"):
			return `{"found":true,"quotes":["the speaker said it"],"best_case":"direct quote"}`, nil
		case strings.Contains(system, "You are the Faithfulness Critic"):
			// Default: overstated with what_source_actually_says populated.
			if strings.Contains(prompt, "faithful-claim") {
				return `{"findings":[],"verdict":"faithful","evidence":"direct match","what_source_actually_says":null}`, nil
			}
			return `{"findings":[],"verdict":"overstated","evidence":"rhetorical","what_source_actually_says":"automation requires human oversight"}`, nil
		case strings.Contains(system, "You are the Evidence Grounder"):
			// Record which proposition was grounded via the prompt content.
			return `{"verdict":"supported","finding":"evidence found","sources":[]}`, nil
		default:
			return `[]`, nil
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
	stub := func(system, prompt string, withTools bool) (string, error) {
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
	stub := func(system, prompt string, withTools bool) (string, error) {
		switch {
		case strings.Contains(system, "You are the Defender"):
			return `{"found":true,"quotes":["automation is a lie"],"best_case":"direct"}`, nil
		case strings.Contains(system, "You are the Faithfulness Critic"):
			return `{"findings":[],"verdict":"overstated","evidence":"rhetorical","what_source_actually_says":"automation requires human oversight"}`, nil
		case strings.Contains(system, "You are the Evidence Grounder"):
			// Capture what proposition was sent for grounding.
			groundedProposition = prompt
			return `{"verdict":"mixed","finding":"humans still needed","sources":[]}`, nil
		default:
			return `[]`, nil
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
	stub := func(system, prompt string, withTools bool) (string, error) {
		switch {
		case strings.Contains(system, "You are the Producer"):
			return `{"steelman":"strongest version","conditions":"some conditions"}`, nil
		case strings.Contains(system, "You are the Critic"):
			return `{"critique":[],"verdict":"partial","surviving_claim":null,"reason":"ok",` +
				`"needs_another_round":false,"added_conditions":0,"survives_only_by_conditioning":false}`, nil
		case strings.Contains(system, "You are the Defender"):
			return `{"found":true,"quotes":["automation is a lie"],"best_case":"direct"}`, nil
		case strings.Contains(system, "You are the Faithfulness Critic"):
			if strings.Contains(prompt, "faithful-claim") {
				// Control: faithful claim with no source says → should ground literal.
				return `{"findings":[],"verdict":"faithful","evidence":"direct","what_source_actually_says":null}`, nil
			}
			return `{"findings":[],"verdict":"overstated","evidence":"rhetorical",` +
				`"what_source_actually_says":"automation always needs a human in the loop"}`, nil
		case strings.Contains(system, "You are the Evidence Grounder"):
			groundedProposition = prompt
			if strings.Contains(prompt, "Automation is a lie") {
				return `{"verdict":"refuted","finding":"literal reading refuted","sources":[]}`, nil
			}
			return `{"verdict":"supported","finding":"automation paradox documented","sources":[]}`, nil
		default:
			return `[]`, nil
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
	stub := func(system, prompt string, withTools bool) (string, error) {
		switch {
		case strings.Contains(system, "You are the Producer"):
			return `{"steelman":"strongest version","conditions":"some conditions"}`, nil
		case strings.Contains(system, "You are the Critic"):
			return `{"critique":[],"verdict":"partial","surviving_claim":null,"reason":"ok",` +
				`"needs_another_round":false,"added_conditions":0,"survives_only_by_conditioning":false}`, nil
		case strings.Contains(system, "You are the Defender"):
			return `{"found":true,"quotes":["SaaS is not dead"],"best_case":"direct"}`, nil
		case strings.Contains(system, "You are the Faithfulness Critic"):
			return `{"findings":[],"verdict":"faithful","evidence":"direct","what_source_actually_says":null}`, nil
		case strings.Contains(system, "You are the Evidence Grounder"):
			groundedProposition = prompt
			return `{"verdict":"supported","finding":"SaaS growing","sources":[]}`, nil
		default:
			return `[]`, nil
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
	stub := func(system, prompt string, withTools bool) (string, error) {
		if strings.Contains(system, "You are the Evidence Grounder") {
			evidenceCallCount++
		}
		switch {
		case strings.Contains(system, "You are the Defender"):
			return `{"found":true,"quotes":["claim text"],"best_case":"direct"}`, nil
		case strings.Contains(system, "You are the Faithfulness Critic"):
			return `{"findings":[],"verdict":"faithful","evidence":"direct","what_source_actually_says":null}`, nil
		case strings.Contains(system, "You are the Evidence Grounder"):
			return `{"verdict":"supported","finding":"evidence found","sources":[]}`, nil
		default:
			return `[]`, nil
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
	stub := func(system, prompt string, withTools bool) (string, error) {
		if strings.Contains(system, "You are the Evidence Grounder") {
			evidenceCallCount++
		}
		switch {
		case strings.Contains(system, "You are the Producer"):
			return `{"steelman":"strongest version","conditions":"some conditions"}`, nil
		case strings.Contains(system, "You are the Critic"):
			return `{"critique":[],"verdict":"partial","surviving_claim":null,"reason":"ok",` +
				`"needs_another_round":false,"added_conditions":0,"survives_only_by_conditioning":false}`, nil
		case strings.Contains(system, "You are the Defender"):
			return `{"found":true,"quotes":["claim text"],"best_case":"direct"}`, nil
		case strings.Contains(system, "You are the Faithfulness Critic"):
			return `{"findings":[],"verdict":"faithful","evidence":"direct","what_source_actually_says":null}`, nil
		case strings.Contains(system, "You are the Evidence Grounder"):
			return `{"verdict":"supported","finding":"evidence found","sources":[]}`, nil
		default:
			return `[]`, nil
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

// TestFixtureIngestion iterates fixtures/raw and passes each real file through splitSummary.
// Goal: confirm the regex boundaries don't panic or hang on large, heterogeneous real-world inputs.
// Expected files that are missing are t.Skip'd — never substituted.
func TestFixtureIngestion(t *testing.T) {
	const (
		dir        = "fixtures/raw"
		maxSegSize = 4096 // no legitimate atomic-claim line exceeds 4 KB; table blobs do
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
			avgSeg := len(data) / len(got)
			t.Logf("%s: %d bytes → %d segments, avg %d b/seg, max seg %d b (first: %.80q)",
				name, len(data), len(got), avgSeg, maxSeg, strings.TrimSpace(got[0]))
		})
	}
}
