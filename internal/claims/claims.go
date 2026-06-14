// SPDX-License-Identifier: Apache-2.0

// Package claims holds the claim and verdict types for the evaluation modes,
// together with enum validation at the boundary.
//
// It is pure: no network, no global state. Verdict-enum validation happens here
// so malformed verdicts never reach orchestration or render. This file carries
// the substance-mode types; faithfulness/evidence/audit types are added by their
// own sessions.
package claims

// Verdict is one grading verdict. The block below is the closed verdict
// vocabulary from spec/BEHAVIOR.md "Verdict enums and routing", verbatim, grouped
// by the mode that emits each; values shared across modes (partial, error) are
// defined once. These are the single source of each spelling — code switches on
// or matches the named const, never a re-spelled literal.
//
// The consts are deliberately left untyped (not `Substantive Verdict = …`) so the
// same name serves both the string-typed result path (Substance.Verdict) and the
// Verdict-typed maps/switch in render, without forcing conversions at every
// existing callsite. An untyped string const is assignable to both string and
// Verdict.
type Verdict string

const (
	// Substance mode: substantive | partial | hollow | error.
	Substantive = "substantive"
	Hollow      = "hollow"

	// Faithfulness mode: faithful | partial | overstated | absent | contradicted | error.
	Faithful     = "faithful"
	Overstated   = "overstated"
	Absent       = "absent"
	Contradicted = "contradicted"

	// Evidence mode: supported | mixed | refuted | unverifiable | skipped (over cap) | error.
	Supported    = "supported"
	Mixed        = "mixed"
	Refuted      = "refuted"
	Unverifiable = "unverifiable"
	Skipped      = "skipped (over cap)"

	// Shared across modes.
	Partial = "partial"
	Error   = "error"
)

// Axis is one critic finding: which axis fired, the one-sentence finding, and a
// severity of "fatal" | "weakens" | "clears" (spec/PROMPTS.md §3).
type Axis struct {
	Axis     string
	Finding  string
	Severity string
}

// SubstanceDetail is the per-claim verification detail recorded in the Tier-2
// chain (spec/CLI.md, substanceDetail). Reason lives on Substance.
type SubstanceDetail struct {
	Steelman        string
	Critique        []Axis
	SurvivingClaim  string
	AddedConditions int
	Rounds          int
}

// Substance is the result of grading one claim in substance mode.
type Substance struct {
	Claim   string
	Verdict string
	Reason  string
	Detail  SubstanceDetail
}

var substanceVerdicts = map[string]bool{
	Substantive: true,
	Partial:     true,
	Hollow:      true,
	Error:       true,
}

// ValidSubstanceVerdict reports whether v is a verdict substance mode can emit.
// Matching is exact: case, whitespace, and misspellings are rejected.
func ValidSubstanceVerdict(v string) bool { return substanceVerdicts[v] }
