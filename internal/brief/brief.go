package brief

// brief renders the default stdout report: a one-line rollup plus the "Needs you" rows a human
// should look at first. It is plain Go over the verdict rows — no model call reaches this package,
// so the selection is deterministic and auditable. spec/CLI.md §"Brief report (default)" is the
// contract; the tree package reuses Qualify here so tree expansion and the brief agree on which
// claims need attention.

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Row is one claim joined to whatever verdicts the run produced. A mode that did not run leaves its
// verdict field "". `Text` is the claim itself; the *Reason fields are the mode's own rationale,
// shown truncated for whichever tier selects the row.
type Row struct {
	ID, Path, Text                             string
	Faith, Substance, Grounding                string
	FaithReason, SubstanceReason, GroundReason string
	Spread                                     string // "k/N" verdict agreement under -n>1, else ""
}

// Sel is a selected row plus the tier that selected it (0 = highest priority).
type Sel struct {
	Row  Row
	Tier int
}

var numRe = regexp.MustCompile(`[0-9]`)

// Qualify applies the selection rule in spec/CLI.md, returning the highest (lowest-numbered) tier a
// claim matches. Tiers, in priority order:
//
//	0 (a) — faithfulness contradicted or absent
//	1 (b) — faithful AND grounding refuted (the laundering signature)
//	2 (c) — faithfulness overstated where the claim contains a number
//	3 (d) — grounding refuted
//
// A tier needing a signal the run did not produce (empty verdict) simply does not match, so a
// faithfulness-only run reaches only tiers a and c. Reused by the tree package for expansion.
func Qualify(faith, substance, grounding, text string) (tier int, ok bool) {
	switch {
	case faith == "contradicted" || faith == "absent":
		return 0, true
	case faith == "faithful" && grounding == "refuted":
		return 1, true
	case faith == "overstated" && numRe.MatchString(text):
		return 2, true
	case grounding == "refuted":
		return 3, true
	}
	return 0, false
}

// Selected returns the qualifying rows, sorted by tier, then claim id (natural order), then claim
// text — the last key makes the order total and deterministic under id ties.
func Selected(rows []Row) []Sel {
	var sels []Sel
	for _, r := range rows {
		if tier, ok := Qualify(r.Faith, r.Substance, r.Grounding, r.Text); ok {
			sels = append(sels, Sel{Row: r, Tier: tier})
		}
	}
	sort.SliceStable(sels, func(i, j int) bool {
		if sels[i].Tier != sels[j].Tier {
			return sels[i].Tier < sels[j].Tier
		}
		if c := idLess(sels[i].Row.ID, sels[j].Row.ID); c != 0 {
			return c < 0
		}
		return sels[i].Row.Text < sels[j].Row.Text
	})
	return sels
}

// reasonFor returns the rationale to display for the tier that selected the row: the faithfulness
// finding for the faithfulness tiers (a, c), the grounding finding for the grounding tiers (b, d).
func reasonFor(s Sel) string {
	switch s.Tier {
	case 0, 2:
		return s.Row.FaithReason
	default:
		return s.Row.GroundReason
	}
}

// Brief renders the default report for `rows`. `verdictCounts` is the line-1 rollup (already
// formatted for the mode(s) run, e.g. "faithful 7, overstated 3, contradicted 2"); `auditPath` is
// the full-table path named on the last line and in the overflow line.
func Brief(rows []Row, verdictCounts, auditPath string) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "%d claims — %s\n\n", len(rows), verdictCounts)

	sels := Selected(rows)
	fmt.Fprintf(&b, "Needs you (%d):\n", len(sels))
	shown := sels
	if len(shown) > 10 {
		shown = shown[:10]
	}
	for _, s := range shown {
		fmt.Fprintf(&b, "%s | %s | %s | %s\n",
			s.Row.ID, verdictLabel(s.Row), truncate(s.Row.Text, 60), truncate(reasonFor(s), 60))
	}
	if len(sels) > 10 {
		fmt.Fprintf(&b, "and %d more in %s\n", len(sels)-10, auditPath)
	}
	fmt.Fprintf(&b, "\nFull table: %s\n", auditPath)
	return b.String(), nil
}

// verdictLabel picks the verdict to show for a selected row: the faithfulness verdict when set,
// otherwise the grounding verdict. (Selection guarantees at least one is set.)
func verdictLabel(r Row) string {
	if r.Faith != "" {
		return r.Faith
	}
	return r.Grounding
}

func truncate(s string, n int) string {
	s = strings.ReplaceAll(strings.TrimSpace(s), "\n", " ")
	if len([]rune(s)) <= n {
		return s
	}
	return string([]rune(s)[:n-1]) + "…"
}

// idLess compares claim ids so that a numeric run sorts as a number (F2 < F10), with a trailing
// suffix (F2a < F2b) breaking the tie. Returns -1, 0, or 1. Purely lexical ids fall back to string
// comparison, which is still total.
func idLess(a, b string) int {
	an, asuf := splitID(a)
	bn, bsuf := splitID(b)
	switch {
	case an >= 0 && bn >= 0 && an != bn:
		if an < bn {
			return -1
		}
		return 1
	case an != bn: // one has a number, the other does not
		if an < bn {
			return -1
		}
		return 1
	}
	switch {
	case asuf < bsuf:
		return -1
	case asuf > bsuf:
		return 1
	}
	return 0
}

// splitID separates a leading alpha prefix + digits from a trailing suffix: "F12a" → 12, "a".
// A id with no digit run returns -1 and the whole string as the suffix.
func splitID(id string) (num int, suffix string) {
	i := 0
	for i < len(id) && (id[i] < '0' || '9' < id[i]) {
		i++
	}
	j := i
	for j < len(id) && '0' <= id[j] && id[j] <= '9' {
		j++
	}
	if i == j {
		return -1, id
	}
	n, err := strconv.Atoi(id[i:j])
	if err != nil {
		return -1, id
	}
	return n, id[j:]
}
