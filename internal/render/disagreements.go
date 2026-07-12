// SPDX-License-Identifier: Apache-2.0

package render

import (
	"fmt"
	"strings"

	"github.com/peterwilliams97/elenchus2/internal/claims"
)

// This file implements the -disagreements render (REVIEWER.md Q1, decision d029):
// a projection over what -audit already computes. It reads only the three column
// verdicts and the claim text per row — the same data mdAudit draws as its table —
// and emits, in plain English, only the claims whose three columns disagree. No
// model is called and no orchestration struct gains a field; AuditRow below is a
// render input the caller fills, not shared run-state.

// AuditRow is the render input for one audited claim: its text and the three audit
// column verdicts as strings, exactly as audit produced them.
type AuditRow struct {
	Claim     string // the claim text, shown verbatim
	Faith     string // faithfulness verdict
	Substance string // substance verdict
	Grounding string // grounding verdict
}

// tier is the severity bucket a verdict maps to. It reuses the spec's terminal
// colour routing (spec/BEHAVIOR.md, "Terminal color routing", vcolor): green=PASS,
// yellow=WEAK, red=FAIL, grey=NULL. The three modes speak three vocabularies, so a
// tier is the only way to compare their verdicts to one another.
type tier int

const (
	tierNull tier = iota // grey: unverifiable, error, skipped, and any unknown verdict
	tierWeak             // yellow: partial, overstated, mixed
	tierFail             // red: hollow, absent, refuted, contradicted
	tierPass             // green: faithful, substantive, supported
)

func (t tier) String() string {
	switch t {
	case tierPass:
		return "PASS"
	case tierFail:
		return "FAIL"
	case tierWeak:
		return "WEAK"
	default:
		return "NULL"
	}
}

// tierOf maps a verdict string to its tier, following vcolor exactly. Unknown
// verdicts fall through to NULL, matching vcolor's "all others" grey bucket.
func tierOf(verdict string) tier {
	switch claims.Verdict(verdict) {
	case claims.Faithful, claims.Substantive, claims.Supported:
		return tierPass
	case claims.Partial, claims.Overstated, claims.Mixed:
		return tierWeak
	case claims.Hollow, claims.Absent, claims.Refuted, claims.Contradicted:
		return tierFail
	default:
		// unverifiable, error, "skipped (over cap)", and anything unrecognised.
		return tierNull
	}
}

// surfaces reports whether a claim's three columns disagree in the sense d029
// fixed: at least one column is PASS and at least one is FAIL. WEAK and NULL never
// make a claim surface on their own — an all-weak or weak/null row is unsettled,
// not contested, and a row with only FAILs is a clean reject, not a question.
func surfaces(faith, substance, grounding tier) bool {
	hasPass := faith == tierPass || substance == tierPass || grounding == tierPass
	hasFail := faith == tierFail || substance == tierFail || grounding == tierFail
	return hasPass && hasFail
}

// disagree applies the surface rule to a row's verdict strings.
func disagree(r AuditRow) bool {
	return surfaces(tierOf(r.Faith), tierOf(r.Substance), tierOf(r.Grounding))
}

// Per-column clauses for the conflict sentence. Only PASS and FAIL verdicts carry
// a clause; WEAK and NULL columns are not part of the conflict and contribute
// nothing. Because a surfaced row always has >=1 PASS column and >=1 FAIL column,
// and every PASS/FAIL verdict below has a clause, both halves of the sentence are
// always non-empty.
var (
	faithPass = map[claims.Verdict]string{claims.Faithful: "the source really says it"}
	faithFail = map[claims.Verdict]string{
		claims.Absent:       "the source does not say it",
		claims.Contradicted: "the source says the opposite",
	}
	substancePass = map[claims.Verdict]string{claims.Substantive: "it survives scrutiny"}
	substanceFail = map[claims.Verdict]string{claims.Hollow: "it is hollow under scrutiny"}
	groundingPass = map[claims.Verdict]string{claims.Supported: "the evidence backs it"}
	groundingFail = map[claims.Verdict]string{claims.Refuted: "the evidence contradicts it"}
)

// question builds the plain-English question for one surfaced row: the PASS clauses
// joined with "and", then "but", then the FAIL clauses, closed with the fixed
// reviewer prompt.
func question(r AuditRow) string {
	cols := []struct {
		verdict    string
		pass, fail map[claims.Verdict]string
	}{
		{r.Faith, faithPass, faithFail},
		{r.Substance, substancePass, substanceFail},
		{r.Grounding, groundingPass, groundingFail},
	}
	var pass, fail []string
	for _, c := range cols {
		v := claims.Verdict(c.verdict)
		if cl, ok := c.pass[v]; ok {
			pass = append(pass, cl)
		}
		if cl, ok := c.fail[v]; ok {
			fail = append(fail, cl)
		}
	}
	sentence := capitalizeFirst(strings.Join(pass, " and ")) + ", but " + strings.Join(fail, " and ") + "."
	return sentence + "\n  Keep it, qualify it, or cut it?"
}

// Disagreements renders only the claims whose three columns disagree (>=1 PASS and
// >=1 FAIL), each as one plain-English question plus its three verdicts. Rows are
// numbered by their original 1-based position so a reviewer can find them in the
// full audit; non-disagreeing rows are omitted. Returns "" when nothing surfaces.
func Disagreements(rows []AuditRow) string {
	var b strings.Builder
	for i, r := range rows {
		if !disagree(r) {
			continue
		}
		if b.Len() > 0 {
			b.WriteString("\n\n")
		}
		fmt.Fprintf(&b, "Claim %d [%s · %s · %s]\n  %s\n  %q",
			i+1, r.Faith, r.Substance, r.Grounding, question(r), r.Claim)
	}
	return b.String()
}

// capitalizeFirst upper-cases the first rune of s, leaving the rest unchanged.
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = []rune(strings.ToUpper(string(r[0])))[0]
	return string(r)
}
