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
	"time"

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
	for _, f := range []string{"verdict", "gap", "evidence", "report_says", "source_says", "reason"} {
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
		`"report_says":"the country parts are funded fairly","source_says":"the country parts get about a dollar each","reason":"narrower than claimed"}`
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
	if got.SourceSays != "the country parts get about a dollar each" {
		t.Errorf("source_says not carried: %q", got.SourceSays)
	}
	if got.ReportSays != "the country parts are funded fairly" {
		t.Errorf("report_says not carried: %q", got.ReportSays)
	}
	// partial + a verified quote → the two-clause stakes line.
	wantSoWhat := "The report says the country parts are funded fairly. " +
		"The source only says the country parts get about a dollar each."
	if got.SoWhat != wantSoWhat {
		t.Errorf("stakes line not assembled:\n got %q\nwant %q", got.SoWhat, wantSoWhat)
	}
	if r := c.usage.plainRetriesN(); r != 0 {
		t.Errorf("paraphrased restatements should trigger no plain retry, got %d", r)
	}
}

// TestFaithJudgeChainRecordF46b is the golden test on F46b's chain record: the judge's contradicted
// verdict, its two verbatim quotes, and its report_says/source_says restatements produce the exact
// leaf a rerun writes — the two ≤12-word halves and the stakes line renderStakes assembles from them.
func TestFaithJudgeChainRecordF46b(t *testing.T) {
	passages := []retrieve.Passage{
		{ID: "submission-41#p2", Source: "submission",
			Text: "In addition to these external projects, over the same period the ABC spent $80 million on 52 internal projects."},
		{ID: "2025-02-27/4_abc#t14", Source: "hearing",
			Text: "Additionally, over the same period the ABC invested over $80 million in 52 internal productions based in Victoria, delivering a further 507 hours."},
	}
	const judge = `{"verdict":"contradicted","gap":"scope","evidence":[` +
		`{"passage_id":"submission-41#p2","quote":"In addition to these external projects, over the same period the ABC spent $80 million on 52 internal projects."},` +
		`{"passage_id":"2025-02-27/4_abc#t14","quote":"Additionally, over the same period the ABC invested over $80 million in 52 internal productions based in Victoria, delivering a further 507 hours."}],` +
		`"report_says":"most of those 52 shows were mainly made in Victoria",` +
		`"source_says":"those shows were only located in Victoria",` +
		`"reason":"The sources describe the 52 shows as headquartered in the state, never as a majority of production."}`
	stub := func(system, prompt string, withTools bool) (string, []retrievedSource, error) {
		return judge, nil, nil
	}
	c := cfg{call: stub, usage: newUsageCounters()}
	const claim = "...and produced 52 internal projects with the majority of production in Victoria."
	got := c.faithJudge(claim, passages, nil)

	rec := faithChainRecord(53, 69, claim, got, time.Now())
	var fd faithDetail
	if err := json.Unmarshal(rec.Detail, &fd); err != nil {
		t.Fatalf("detail decode: %v", err)
	}
	if rec.Verdict != "contradicted" {
		t.Errorf("verdict = %q, want contradicted", rec.Verdict)
	}
	if fd.ReportSays != "most of those 52 shows were mainly made in Victoria" {
		t.Errorf("report_says = %q", fd.ReportSays)
	}
	if fd.SourceSays != "those shows were only located in Victoria" {
		t.Errorf("source_says = %q", fd.SourceSays)
	}
	wantSoWhat := "The report says most of those 52 shows were mainly made in Victoria. " +
		"The source only says those shows were only located in Victoria."
	if fd.SoWhat != wantSoWhat {
		t.Errorf("so_what:\n got %q\nwant %q", fd.SoWhat, wantSoWhat)
	}
	// The restatements paraphrase — no banned word, no ≥4-syllable import, no copied 3-word run — so
	// the judge is called once, no plain retry.
	if r := c.usage.plainRetriesN(); r != 0 {
		t.Errorf("clean restatement should trigger no plain retry, got %d", r)
	}
}

// TestRenderStakes pins the three templates: partial/overstated/contradicted contrast the two halves,
// absent/unsupported say no held source carries it, faithful (and any non-verdict) render nothing, and
// a missing half yields "" — there is nothing to contrast.
func TestRenderStakes(t *testing.T) {
	for _, tc := range []struct{ verdict, want string }{
		{"partial", "The report says R. The source only says S."},
		{"overstated", "The report says R. The source only says S."},
		{"contradicted", "The report says R. The source only says S."},
		{"absent", "The report says R. No held source says this."},
		{"unsupported", "The report says R. No held source says this."},
		{"faithful", ""},
		{"error", ""},
	} {
		if got := renderStakes(tc.verdict, "R", "S"); got != tc.want {
			t.Errorf("renderStakes(%q) = %q, want %q", tc.verdict, got, tc.want)
		}
	}
	if got := renderStakes("partial", "", "S"); got != "" {
		t.Errorf("empty report_says should yield \"\", got %q", got)
	}
	if got := renderStakes("partial", "R", ""); got != "" {
		t.Errorf("partial with empty source_says should yield \"\", got %q", got)
	}
}

// TestPlainBadWord pins the restatement check: a restatement built from the claim's and quotes'
// vocabulary passes; "reader" and "would" are caught outright; a ≥4-syllable word the source never
// used is caught.
func TestPlainBadWord(t *testing.T) {
	claim := "regional Victoria gets its fair share of production funding"
	quotes := []string{"52 internal productions based in Victoria"}
	if w := plainBadWord("productions based in Victoria", claim, quotes); w != "" {
		t.Errorf("clean restatement flagged %q", w)
	}
	if plainBadWord("the reader misreads it", claim, quotes) != "reader" {
		t.Error(`"reader" not caught`)
	}
	if plainBadWord("a summary would imply more", claim, quotes) != "would" {
		t.Error(`"would" not caught`)
	}
	if plainBadWord("disproportionate concentration in Victoria", claim, quotes) == "" {
		t.Error("imported ≥4-syllable word not caught")
	}
}

// TestFaithJudgePlainRestatementRetry drives faithJudge with a stub that returns a "reader"/"would"
// restatement first and a plain one on the retry, asserting exactly one plain retry is counted and
// the retry's restatement is the one kept.
func TestFaithJudgePlainRestatementRetry(t *testing.T) {
	passages := []retrieve.Passage{{ID: "p1", Source: "submission",
		Text: "The ABC spent $80 million on 52 internal projects based in Victoria."}}
	const q = `{"passage_id":"p1","quote":"The ABC spent $80 million on 52 internal projects based in Victoria."}`
	bad := `{"verdict":"partial","gap":"scope","evidence":[` + q + `],` +
		`"report_says":"the reader would infer a majority","source_says":"projects based in Victoria","reason":"x"}`
	good := `{"verdict":"partial","gap":"scope","evidence":[` + q + `],` +
		`"report_says":"most work was done in Victoria","source_says":"the shows only sat in the state","reason":"x"}`
	n := 0
	stub := func(system, prompt string, withTools bool) (string, []retrievedSource, error) {
		n++
		if n == 1 {
			return bad, nil, nil
		}
		return good, nil, nil
	}
	c := cfg{call: stub, usage: newUsageCounters()}
	got := c.faithJudge("claim about Victoria projects", passages, nil)
	if n != 2 {
		t.Errorf("want 2 dispatches (1 + plain retry), got %d", n)
	}
	if c.usage.plainRetriesN() != 1 {
		t.Errorf("want 1 plain retry, got %d", c.usage.plainRetriesN())
	}
	if got.ReportSays != "most work was done in Victoria" {
		t.Errorf("retry result not kept: %q", got.ReportSays)
	}
}

// TestVerbatimRun pins the no-copy check: a 3+-word run lifted from the claim or a quote is caught
// (numbers count as words), a 2-word overlap is allowed, and a genuine paraphrase passes.
func TestVerbatimRun(t *testing.T) {
	claim := "the ABC produced 52 internal projects with the majority of production in Victoria"
	quotes := []string{"52 internal productions based in Victoria"}
	if verbatimRun("it had the majority of production locally", claim, quotes) == "" {
		t.Error(`3-word run "majority of production" copied from the claim not caught`)
	}
	if verbatimRun("the shows were based in Victoria", claim, quotes) == "" {
		t.Error(`3-word run "based in Victoria" copied from a quote not caught`)
	}
	if verbatimRun("those 52 internal projects were local", claim, quotes) == "" {
		t.Error(`numeric 3-word run "52 internal projects" not caught`)
	}
	if w := verbatimRun("most shows were made in Victoria", claim, quotes); w != "" {
		t.Errorf("a 2-word overlap should be allowed, flagged %q", w)
	}
	if w := verbatimRun("the country was funded fairly", claim, quotes); w != "" {
		t.Errorf("a genuine paraphrase should pass, flagged %q", w)
	}
}

// TestFaithJudgeVerbatimCopyRetry drives faithJudge with a stub whose first restatement lifts a
// 3-word run from a verified quote (no banned or long word — only the copy is wrong) and paraphrases
// on the retry, asserting the copied run alone triggers exactly one plain retry and the paraphrase is
// kept.
func TestFaithJudgeVerbatimCopyRetry(t *testing.T) {
	passages := []retrieve.Passage{{ID: "p1", Source: "submission",
		Text: "52 internal productions based in Victoria, delivering a further 507 hours."}}
	const q = `{"passage_id":"p1","quote":"52 internal productions based in Victoria, delivering a further 507 hours."}`
	copyJSON := `{"verdict":"partial","gap":"scope","evidence":[` + q + `],` +
		`"report_says":"the shows were based in Victoria","source_says":"the shows sat in the state","reason":"x"}`
	clean := `{"verdict":"partial","gap":"scope","evidence":[` + q + `],` +
		`"report_says":"the shows were mostly made locally","source_says":"the shows only sat in the state","reason":"x"}`
	n := 0
	stub := func(system, prompt string, withTools bool) (string, []retrievedSource, error) {
		n++
		if n == 1 {
			return copyJSON, nil, nil
		}
		return clean, nil, nil
	}
	c := cfg{call: stub, usage: newUsageCounters()}
	got := c.faithJudge("a claim about the shows", passages, nil)
	if n != 2 {
		t.Errorf("want 2 dispatches (copy + steered retry), got %d", n)
	}
	if c.usage.plainRetriesN() != 1 {
		t.Errorf("want 1 plain retry for the copied run, got %d", c.usage.plainRetriesN())
	}
	if got.ReportSays != "the shows were mostly made locally" {
		t.Errorf("retry paraphrase not kept: %q", got.ReportSays)
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
		return `{"verdict":"absent","gap":"none","evidence":[],"report_says":"","source_says":"","reason":"nope"}`, nil, nil
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
