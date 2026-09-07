package main

// judge_test covers the Step-2 faithfulness judge: the output schema is valid, the prompt keeps the
// literalization + verbatim-quote rules, the code-side grounding check rejects a non-verbatim quote
// (the bendigo#t15 canary — a Sonnet paraphrase), the judge drops rejected quotes and counts them,
// claims sharing passages are grouped for cache reuse, and a schema-invalid answer is retried once.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"assay/internal/retrieve"
)

func TestJudgeSchemaValid(t *testing.T) {
	if !json.Valid([]byte(judgeSchema)) {
		t.Fatal("judgeSchema is not valid JSON")
	}
	var s struct {
		Required   []string `json:"required"`
		Properties map[string]json.RawMessage
	}
	if err := json.Unmarshal([]byte(judgeSchema), &s); err != nil {
		t.Fatalf("judgeSchema decode: %v", err)
	}
	for _, f := range []string{"verdict", "gap", "evidence", "so_what", "reason"} {
		if _, ok := s.Properties[f]; !ok {
			t.Errorf("judgeSchema missing property %q", f)
		}
	}
	if len(s.Required) < 5 {
		t.Errorf("judgeSchema should require the core fields, got %v", s.Required)
	}
}

func TestFaithJudgeSysKeepsLiteralizationAndVerbatim(t *testing.T) {
	for _, want := range []string{"Literalization", "VERBATIM", "passage_id"} {
		if !strings.Contains(faithJudgeSys, want) {
			t.Errorf("faithJudgeSys missing %q", want)
		}
	}
}

// TestFaithJudgeSysScopeDiscipline pins the scope rule and its two worked negatives (the training
// -opportunities claim and the regional-Victoria-vs-regional-Australia claim) are in the judge prompt.
func TestFaithJudgeSysScopeDiscipline(t *testing.T) {
	for _, want := range []string{"SUBJECT, SCOPE, and DIRECTION", "SCOPE DISCIPLINE", "gap=scope",
		"training opportunities", "wrong scope (place)"} {
		if !strings.Contains(faithJudgeSys, want) {
			t.Errorf("faithJudgeSys missing scope-discipline text %q", want)
		}
	}
}

// TestQuoteInPassageVerbatimVsParaphrase is the grounding-check canary: the exact span verifies, but
// the Sonnet paraphrase ("arts" for the source's "creative") does not, and a question-context quote
// verifies against the passage's visible Q text.
func TestQuoteInPassageVerbatimVsParaphrase(t *testing.T) {
	p := retrieve.Passage{
		ID:      "2025-03-13/6_bendigo#t15",
		Text:    "The creative sector had shut down, and as a mother I had no viable career prospects.",
		Context: "Gaelle BROAD: How did the pandemic affect your career?",
	}
	if !quoteInPassage("The creative sector had shut down, and as a mother I had no viable career prospects.", p) {
		t.Error("verbatim answer quote should verify")
	}
	if quoteInPassage("The arts sector had shut down, and as a mother I had no viable career prospects.", p) {
		t.Error("paraphrase (arts≠creative) must NOT verify — this is the grounding check")
	}
	if !quoteInPassage("How did the pandemic affect your career?", p) {
		t.Error("verbatim question-context quote should verify against visible Q text")
	}
	if quoteInPassage("", p) {
		t.Error("empty quote must never verify")
	}
}

// TestFaithJudgeRejectsNonVerbatimQuotes drives the real faithJudge with a stub returning one verbatim
// quote, one paraphrase, and one invented passage id. Only the verbatim quote survives; the two bad
// ones are counted as quote rejects.
func TestFaithJudgeRejectsNonVerbatimQuotes(t *testing.T) {
	passages := []retrieve.Passage{
		{ID: "p1", Text: "regional Victorians receive $1.06 per capita in federal arts funding."},
		{ID: "p2", Text: "Creative Australia's annual report shows the shortfall."},
	}
	const judge = `{"verdict":"partial","gap":"denominator",` +
		`"evidence":[` +
		`{"passage_id":"p1","quote":"regional Victorians receive $1.06 per capita in federal arts funding."},` +
		`{"passage_id":"p1","quote":"regional Victorians receive $2.00 per capita"},` +
		`{"passage_id":"p9","quote":"Creative Australia's annual report shows the shortfall."}],` +
		`"so_what":"reader overstates the share","reason":"narrower than claimed","what_source_actually_says":"regional gets $1.06 per capita"}`
	stub := func(system, prompt string, withTools bool) (string, []retrievedSource, error) {
		return judge, nil, nil
	}
	c := cfg{call: stub, usage: newUsageCounters()}
	zero := 0.0
	got := c.faithJudge("Regional Victoria gets its fair share.", passages, &zero)

	if got.Verdict != "partial" || got.Gap != "denominator" {
		t.Errorf("verdict/gap not parsed: %q/%q", got.Verdict, got.Gap)
	}
	if len(got.Quotes) != 1 || got.Quotes[0] != passages[0].Text {
		t.Errorf("only the verbatim quote should survive, got %v", got.Quotes)
	}
	if _, rejects, _ := c.usage.extras(); rejects != 2 {
		t.Errorf("want 2 quote rejects (paraphrase + invented id), got %d", rejects)
	}
	if got.SourceSays != "regional gets $1.06 per capita" {
		t.Errorf("what_source_actually_says not carried: %q", got.SourceSays)
	}
}

// TestGroundVerdictDowngrade pins the code-side rule: every verdict except absent needs at least one
// verified quote. With none, contradicted→absent and faithful/partial/overstated→unsupported, each
// counted; a surviving quote leaves the verdict untouched; absent alone is never downgraded.
func TestGroundVerdictDowngrade(t *testing.T) {
	for _, tc := range []struct {
		verdict  string
		verified []string
		want     string
		counted  bool
	}{
		{"contradicted", nil, "absent", true},
		{"faithful", nil, "unsupported", true},
		{"partial", nil, "unsupported", true},
		{"overstated", nil, "unsupported", true},
		{"faithful", []string{"a real quote"}, "faithful", false},
		{"contradicted", []string{"a real quote"}, "contradicted", false},
		{"overstated", []string{"a real quote"}, "overstated", false},
		{"absent", nil, "absent", false}, // absent is the only verdict that needs no quote
	} {
		c := cfg{usage: newUsageCounters()}
		got := c.groundVerdict(tc.verdict, tc.verified)
		if got != tc.want {
			t.Errorf("groundVerdict(%q, %d quotes) = %q, want %q", tc.verdict, len(tc.verified), got, tc.want)
		}
		if _, _, dg := c.usage.extras(); (dg == 1) != tc.counted {
			t.Errorf("groundVerdict(%q): downgrade counted=%v, want %v", tc.verdict, dg == 1, tc.counted)
		}
	}
}

// TestMissingCites pins the manifest pre-check: a claim citing a document not in the held set reports
// exactly the missing ids; a claim whose cites are all held, or which cites nothing, reports none.
func TestMissingCites(t *testing.T) {
	c := cfg{held: map[string]bool{"submission:41": true, "hearing:2025-03-12/5_sbs": true}}
	if got := c.missingCites("submission:41 hearing:2025-03-12/5_sbs"); len(got) != 0 {
		t.Errorf("all-held should have no missing, got %v", got)
	}
	got := c.missingCites("submission:41, qon:abc/2025-03-21, submission:19/attachment-1")
	if strings.Join(got, "|") != "qon:abc/2025-03-21|submission:19/attachment-1" {
		t.Errorf("missing cites wrong: %v", got)
	}
	if got := c.missingCites(""); got != nil {
		t.Errorf("no cites → nil, got %v", got)
	}
	// With no manifest loaded, the check is disabled entirely.
	if got := (cfg{}).missingCites("qon:abc/2025-03-21"); got != nil {
		t.Errorf("no manifest → no check, got %v", got)
	}
}

func TestGroupBySharedPassages(t *testing.T) {
	ps := func(ids ...string) []retrieve.Passage {
		out := make([]retrieve.Passage, len(ids))
		for i, id := range ids {
			out[i] = retrieve.Passage{ID: id}
		}
		return out
	}
	// claim 0 and 1 share 2 of 3 (≥half); claim 2 is disjoint.
	claimPassages := [][]retrieve.Passage{
		ps("a", "b", "c"),
		ps("a", "b", "z"),
		ps("x", "y"),
	}
	groups := groupBySharedPassages(claimPassages, []int{0, 1, 2})
	if len(groups) != 2 {
		t.Fatalf("want 2 groups, got %d: %v", len(groups), groups)
	}
	if len(groups[0]) != 2 || groups[0][0] != 0 || groups[0][1] != 1 {
		t.Errorf("group 0 should be claims [0 1], got %v", groups[0])
	}
	if len(groups[1]) != 1 || groups[1][0] != 2 {
		t.Errorf("group 1 should be the singleton [2], got %v", groups[1])
	}
	// The union of the shared group is deduplicated in first-seen order.
	union := unionPassages(claimPassages, groups[0])
	if got := retrieve.IDs(union); strings.Join(got, ",") != "a,b,c,z" {
		t.Errorf("union order wrong: %v", got)
	}
}

// TestReadChainSparseToleratesGaps pins that the resume reader returns whatever faithfulness records
// are present, keyed by idx, even when they are non-contiguous (grouping completes claims out of
// order) — and ignores blank lines, non-faithfulness records, and a missing file.
func TestReadChainSparseToleratesGaps(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "c.faithfulness.jsonl")
	// idx 0 and 2 present (idx 1 missing — a gap); one substance record that must be ignored.
	lines := []string{
		`{"idx":0,"mode":"faithfulness","claim":"a","verdict":"partial","spread":"3/3","detail":{"quotes":["q"],"critic_finding":"r"}}`,
		``,
		`{"idx":2,"mode":"faithfulness","claim":"c","verdict":"absent","detail":{}}`,
		`{"idx":9,"mode":"substance","claim":"x","verdict":"hollow","detail":{}}`,
	}
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		t.Fatal(err)
	}
	got := readChainSparse(path)
	if len(got) != 2 {
		t.Fatalf("want 2 faithfulness records (gap at idx1, substance ignored), got %d", len(got))
	}
	if got[0].Verdict != "partial" || got[2].Verdict != "absent" {
		t.Errorf("records not keyed by idx: %+v", got)
	}
	if _, ok := got[1]; ok {
		t.Error("idx 1 should be absent (the gap)")
	}
	if r := readChainSparse(filepath.Join(dir, "nope.jsonl")); len(r) != 0 {
		t.Errorf("missing file should yield empty map, got %d", len(r))
	}
}

// TestSchemaRetryCountedOnce drives callSchema with a stub that returns unparseable text first and
// valid JSON on the retry, asserting exactly one schema retry is counted.
func TestSchemaRetryCountedOnce(t *testing.T) {
	n := 0
	stub := func(system, prompt string, withTools bool) (string, []retrievedSource, error) {
		n++
		if n == 1 {
			return "not json at all", nil, nil
		}
		return `{"verdict":"absent","gap":"none","evidence":[],"so_what":"","reason":"nope","what_source_actually_says":""}`, nil, nil
	}
	c := cfg{call: stub, usage: newUsageCounters()}
	var j judgeJSON
	if err := c.callSchema(faithJudgeSys, "cached", "claim", json.RawMessage(judgeSchema), "faith_verdict", nil, &j); err != nil {
		t.Fatalf("callSchema: %v", err)
	}
	if j.Verdict != "absent" {
		t.Errorf("retry result not parsed, got %q", j.Verdict)
	}
	if retries, _, _ := c.usage.extras(); retries != 1 {
		t.Errorf("want 1 schema retry, got %d", retries)
	}
	if n != 2 {
		t.Errorf("want exactly 2 dispatches (1 + 1 retry), got %d", n)
	}
}
