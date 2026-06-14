package render

import (
	"strings"
	"testing"
	"time"

	"github.com/peterwilliams97/elenchus2/internal/claims"
	"github.com/peterwilliams97/elenchus2/internal/client"
)

func TestTermSubstance(t *testing.T) {
	out := TermSubstance([]claims.Substance{
		{Claim: "The future of work happens in Claude Code.", Verdict: "substantive", Reason: "checkable"},
		{Claim: "SaaS is not dead.", Verdict: "hollow", Reason: "unfalsifiable"},
	})
	for _, want := range []string{"substantive", "hollow", "The future of work happens in Claude Code.", "checkable"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestSummaryBlock(t *testing.T) {
	results := []claims.Substance{
		{Verdict: "substantive"}, {Verdict: "error"}, {Verdict: "hollow"},
	}
	out := Summary("dan_summary", "substance", results, 3*time.Second, "eval/x/dan_summary.substance.jsonl")
	for _, want := range []string{
		"SUMMARY dan_summary substance",
		"cases 3",
		"verified 2", // error is not a win
		"errored 1",
		"substantive 1",
		"hollow 1",
		"error 1",
		"detail: eval/x/dan_summary.substance.jsonl",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("SUMMARY missing %q:\n%s", want, out)
		}
	}
	// Verdict ordering: substantive precedes hollow precedes error.
	si, hi, ei := strings.Index(out, "substantive 1"), strings.Index(out, "hollow 1"), strings.Index(out, "error 1")
	if si >= hi || hi >= ei {
		t.Errorf("verdict ordering wrong (want substantive<hollow<error): %d %d %d\n%s", si, hi, ei, out)
	}
}

func TestSummaryNoChainOmitsDetail(t *testing.T) {
	out := Summary("stdin", "substance", []claims.Substance{{Verdict: "hollow"}}, time.Second, "")
	if strings.Contains(out, "detail:") {
		t.Errorf("detail: line should be omitted when no chain file was written:\n%s", out)
	}
}

func TestUsageLine(t *testing.T) {
	out := UsageLine("claude-sonnet-4-6", client.Usage{Calls: 5, InputTokens: 100, OutputTokens: 50}, 2*time.Second)
	for _, want := range []string{
		"USAGE", "model=claude-sonnet-4-6", "calls=5", "in=100", "out=50", "web_searches=0",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("USAGE missing %q:\n%s", want, out)
		}
	}
}
