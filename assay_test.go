package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
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

// TestUsageAccumulatorSumsCorrectly drives callJSON through the stub seam and asserts the
// usageCounters accumulate correctly from canned API-response-shaped JSON. Because the stub
// bypasses callClaude entirely, we drive the accumulator directly to test the math.
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

	stub := func(system, prompt string, withTools bool) (string, error) {
		switch {
		case strings.Contains(system, "You are the Defender"):
			return `{"found":true,"quotes":["q"],"best_case":"direct"}`, nil
		case strings.Contains(system, "You are the Faithfulness Critic"):
			return `{"findings":[],"verdict":"faithful","evidence":"e","what_source_actually_says":null}`, nil
		case strings.Contains(system, "You are the Evidence Grounder"):
			return `{"verdict":"supported","finding":"found","sources":[]}`, nil
		default:
			return `[]`, nil
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

// ── transport-retry tests ─────────────────────────────────────────────────────
//
// These drive callClaude via a stubbed http.RoundTripper (cfg.httpClient), not
// the cfg.call seam, so the retry loop and status-code handling are exercised
// directly. retryBase is set to 0 so retries are instant.

// roundTripFunc adapts a function to the http.RoundTripper interface.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// fakeResp builds a minimal *http.Response with the given status, body, and headers.
func fakeResp(status int, body string, headers map[string]string) *http.Response {
	h := http.Header{}
	for k, v := range headers {
		h.Set(k, v)
	}
	return &http.Response{StatusCode: status, Header: h, Body: io.NopCloser(strings.NewReader(body))}
}

// okBody is a representative Anthropic API success response (real field layout).
const okBody = `{"id":"msg_01","type":"message","role":"assistant",` +
	`"content":[{"type":"text","text":"hello"}],` +
	`"model":"claude-sonnet-4-6","stop_reason":"end_turn","stop_sequence":null,` +
	`"usage":{"input_tokens":10,"output_tokens":5,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}`

// maxTokensBody is a response that was cut off mid-output.
const maxTokensBody = `{"id":"msg_02","type":"message","role":"assistant",` +
	`"content":[{"type":"text","text":"partial {"}],` +
	`"model":"claude-sonnet-4-6","stop_reason":"max_tokens","stop_sequence":null,` +
	`"usage":{"input_tokens":10,"output_tokens":1500,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}`

// TestCallClaude429ThenSuccess confirms that a 429 followed by a 200 succeeds
// after one retry and returns the expected content.
func TestCallClaude429ThenSuccess(t *testing.T) {
	retryBase = 0
	t.Cleanup(func() { retryBase = time.Second })

	attempts := 0
	c := cfg{
		model:  "claude-sonnet-4-6",
		apiKey: "test-key",
		httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			attempts++
			if attempts == 1 {
				return fakeResp(429, `{"error":{"type":"rate_limit_error","message":"rate limited"}}`,
					map[string]string{"Retry-After": "0"}), nil
			}
			return fakeResp(200, okBody, nil), nil
		})},
	}
	got, err := c.callClaude("sys", "prompt", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hello" {
		t.Errorf("want %q, got %q", "hello", got)
	}
	if attempts != 2 {
		t.Errorf("want 2 attempts, got %d", attempts)
	}
}

// TestCallClaude529PersistentFails confirms that persistent 529s exhaust retries
// and return a clean error. retryBase=0 keeps the test instant.
func TestCallClaude529PersistentFails(t *testing.T) {
	retryBase = 0
	t.Cleanup(func() { retryBase = time.Second })

	attempts := 0
	c := cfg{
		model:  "claude-sonnet-4-6",
		apiKey: "test-key",
		httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			attempts++
			return fakeResp(529, `{"error":{"type":"overloaded_error","message":"overloaded"}}`, nil), nil
		})},
	}
	_, err := c.callClaude("sys", "prompt", false)
	if err == nil {
		t.Fatal("expected error on persistent 529, got nil")
	}
	if attempts != retryMaxAttempts {
		t.Errorf("want %d attempts, got %d", retryMaxAttempts, attempts)
	}
}

// TestCallClaudeMaxTokensTruncation confirms that a max_tokens stop_reason is
// returned as a named error and that usage is recorded before the error is returned.
func TestCallClaudeMaxTokensTruncation(t *testing.T) {
	u := newUsageCounters()
	c := cfg{
		model:  "claude-sonnet-4-6",
		apiKey: "test-key",
		usage:  u,
		httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return fakeResp(200, maxTokensBody, nil), nil
		})},
	}
	_, err := c.callClaude("sys", "prompt", false)
	if err == nil {
		t.Fatal("expected truncation error, got nil")
	}
	if !strings.Contains(err.Error(), "max_tokens") {
		t.Errorf("error should mention max_tokens, got: %v", err)
	}
	// Usage must have been accumulated before the error was returned.
	calls, in, out, _, _, _, _, _ := u.snapshot()
	if calls != 1 || in != 10 || out != 1500 {
		t.Errorf("usage not recorded before truncation error: calls=%d in=%d out=%d", calls, in, out)
	}
}

// TestDecomposeReturnsError confirms that a callJSON failure propagates out of
// decompose as a non-nil error rather than a silent nil slice.
func TestDecomposeReturnsError(t *testing.T) {
	c := cfg{
		call: func(system, prompt string, withTools bool) (string, error) {
			return "", fmt.Errorf("API down")
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

	stub := func(system, prompt string, withTools bool) (string, error) {
		switch {
		case strings.Contains(system, "You are the Evidence Grounder"):
			return `{"verdict":"supported","finding":"found","sources":[]}`, nil
		default:
			return `[]`, nil
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
