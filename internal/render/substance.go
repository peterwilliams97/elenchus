// SPDX-License-Identifier: Apache-2.0

package render

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/peterwilliams97/elenchus2/internal/claims"
	"github.com/peterwilliams97/elenchus2/internal/client"
)

// TermSubstance renders the per-claim substance results for stdout and returns
// the text (the caller writes it, matching Disagreements). The spec pins only
// the markdown table schema and the SUMMARY/USAGE formats; this plain terminal
// layout is a minimal readable rendering of the same columns (# / verdict /
// claim / why).
func TermSubstance(results []claims.Substance) string {
	var b strings.Builder
	for i, r := range results {
		fmt.Fprintf(&b, "[%d] %-11s %s\n", i+1, r.Verdict, r.Claim)
		if r.Reason != "" {
			fmt.Fprintf(&b, "    why: %s\n", r.Reason)
		}
	}
	return b.String()
}

// verdictOrder is the fixed display order for the SUMMARY verdict line
// (spec/CLI.md). Unknown verdicts sort after these; "error" is always last.
var verdictOrder = []string{
	claims.Supported, claims.Mixed, claims.Refuted, claims.Unverifiable,
	claims.Faithful, claims.Partial, claims.Overstated, claims.Absent, claims.Contradicted,
	claims.Substantive, claims.Hollow, claims.Skipped,
}

// Summary renders the SUMMARY block for stderr. errored counts "error" verdicts,
// which never increment verified. The detail: line is omitted when no chain file
// was written. est_usd is reported as n/a: the price table lives in v1 assay.go
// and is not ported in this build, so the model is treated as unpriced
// (spec/CLI.md: est_usd is null when the model is not in the table).
func Summary(fixture, mode string, results []claims.Substance, wall time.Duration, chainFile string) string {
	counts := map[string]int{}
	for _, r := range results {
		counts[r.Verdict]++
	}
	total := len(results)
	errored := counts[claims.Error]
	verified := total - errored

	var b strings.Builder
	fmt.Fprintf(&b, "SUMMARY %s %s\n", fixture, mode)
	fmt.Fprintf(&b, "  cases %d · verified %d · errored %d\n", total, verified, errored)
	fmt.Fprintf(&b, "  %s\n", verdictLine(counts))
	fmt.Fprintf(&b, "  wall %.1fs · est_usd n/a\n", wall.Seconds())
	if chainFile != "" {
		fmt.Fprintf(&b, "  detail: %s\n", chainFile)
	}
	return b.String()
}

// verdictLine renders the per-verdict counts in verdictOrder, with any unknown
// verdicts (sorted) after the ordered ones and "error" always last.
func verdictLine(counts map[string]int) string {
	seen := map[string]bool{}
	var parts []string
	add := func(v string) {
		if n, ok := counts[v]; ok && !seen[v] {
			parts = append(parts, fmt.Sprintf("%s %d", v, n))
			seen[v] = true
		}
	}
	for _, v := range verdictOrder {
		add(v)
	}
	var unknown []string // verdicts not in verdictOrder, excluding error (added last)
	for v := range counts {
		if v != claims.Error && !seen[v] {
			unknown = append(unknown, v)
		}
	}
	sort.Strings(unknown)
	for _, v := range unknown {
		add(v)
	}
	add(claims.Error)
	if len(parts) == 0 {
		return "(none)"
	}
	return strings.Join(parts, " · ")
}

// UsageLine renders the USAGE accounting line for stderr. web_searches is 0 in
// substance mode (no web tool). est_usd is n/a (see Summary).
func UsageLine(model string, u client.Usage, wall time.Duration) string {
	return fmt.Sprintf(
		"USAGE model=%s calls=%d in=%d out=%d cache_read=%d cache_create=%d web_searches=0 wall=%.1fs est_usd=n/a\n",
		model, u.Calls, u.InputTokens, u.OutputTokens, u.CacheReadTokens, u.CacheCreateTokens, wall.Seconds())
}
