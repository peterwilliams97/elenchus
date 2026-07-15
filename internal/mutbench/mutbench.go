// SPDX-License-Identifier: Apache-2.0

// Package mutbench generates seeded mutants from a claims fixture and scores the
// verification chains crossexam writes for them.
//
// It is pure: no network, no API key, no clock, no randomness. Injection is slot
// substitution over authored sites, so the same (fixture, class, seed, version)
// yields a byte-identical mutant (MUTATION_BENCH.md I3). The only thing that
// touches the Anthropic API is crossexam itself; mutbench reads the chain JSONL
// it leaves behind.
//
// What this package deliberately does NOT do: convert axis-fired into recall.
// axis-fired is a harness diagnostic and nothing else (A-T4). Recall is
// target-catch, which is a human read under the A3.4 protocol, and no code here
// computes it.
package mutbench

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

// Variant selects which authored slot value a claim renders with.
type Variant string

const (
	// Clean renders the honest twin: the scope that already covers the broad
	// subject, or the criterion left in place.
	Clean Variant = "clean"
	// Mutant renders the injected defect: the narrow scope, or the criterion
	// deleted.
	Mutant Variant = "mutant"
)

// ConfoundArith quarantines the E1 checkout template, whose embedded 1:50 < 2:00
// step is a derivable sub-claim (A2.1). Tagged rows are reported as their own
// diagnostic row and never pooled into a class number.
const ConfoundArith = "confound:arith"

// Slot is one authored injection site: the two values a template renders with.
type Slot struct {
	Clean  string `json:"clean"`
	Mutant string `json:"mutant"`
}

// Scopes records an SCP-1 claim's narrow measured scope and broad asserted
// subject, so the A3.4 read sheet can state what the reader is looking for
// without the reader re-deriving it. Empty for MFC.
type Scopes struct {
	Narrow string `json:"narrow"`
	Broad  string `json:"broad"`
}

// Claim is one fixture row (bench/fixtures/<F>/claims.jsonl). A per-file DTO: it
// mirrors the fixture schema, so it stays flat.
type Claim struct {
	ID       string          `json:"id"`
	Class    string          `json:"class"`
	Template string          `json:"template"`
	Slots    map[string]Slot `json:"slots"`
	Scopes   Scopes          `json:"scopes"`
	Tags     []string        `json:"tags"`
}

// Row is one manifest record (MUTATION_BENCH.md §4). A per-call JSON DTO, flat
// by design.
//
// AnchorClaimID is the fixture claim ID (A3.2). CounterpartClaimID is retained
// from §4 for v2 document-mode defects and is always empty in v1: the anchor is
// the input, and the reader maps manifest row to chain by reading.
type Row struct {
	MutantID           string   `json:"mutant_id"`
	BaseFixture        string   `json:"base_fixture"`
	DefectClass        string   `json:"defect_class"`
	AnchorClaimID      string   `json:"anchor_claim_id"`
	CounterpartClaimID string   `json:"counterpart_claim_id"`
	Seed               int      `json:"seed"`
	InjectorVersion    string   `json:"injector_version"`
	BaseSHA256         string   `json:"base_sha256"`
	MutantSHA256       string   `json:"mutant_sha256"`
	MutantText         string   `json:"mutant_text"`
	CleanText          string   `json:"clean_text"`
	Scopes             Scopes   `json:"scopes"`
	Tags               []string `json:"tags"`
	StageUnderTest     string   `json:"stage_under_test"`
	Note               string   `json:"note"`
}

var slotRe = regexp.MustCompile(`\{([A-Z_]+)\}`)

// Render substitutes every {SLOT} in the claim's template with the value for v.
// A slot with no authored value is an error: rendering "{SCOPE}" into a mutant
// would ship a broken input that gets graded as if it were prose.
func Render(c Claim, v Variant) (string, error) {
	var missing []string
	out := slotRe.ReplaceAllStringFunc(c.Template, func(m string) string {
		name := strings.Trim(m, "{}")
		s, ok := c.Slots[name]
		if !ok {
			missing = append(missing, name)
			return m
		}
		if v == Clean {
			return s.Clean
		}
		return s.Mutant
	})
	if len(missing) > 0 {
		return "", fmt.Errorf("claim %s: template names slot(s) %v with no authored value", c.ID, missing)
	}
	return out, nil
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// Generate renders one manifest Row per claim. Deterministic: no RNG is
// consulted. seed is recorded so a future generator that selects among several
// authored variants stays re-runnable under I3; v1 substitution has nothing to
// select, and that is a property worth stating rather than implying.
func Generate(claims []Claim, fixture string, seed int, version string) ([]Row, error) {
	rows := make([]Row, 0, len(claims))
	for _, c := range claims {
		clean, err := Render(c, Clean)
		if err != nil {
			return nil, fmt.Errorf("render clean: %w", err)
		}
		mut, err := Render(c, Mutant)
		if err != nil {
			return nil, fmt.Errorf("render mutant: %w", err)
		}
		if clean == mut {
			return nil, fmt.Errorf("claim %s: clean and mutant render identically; injection is a no-op", c.ID)
		}
		mh := sha256Hex(mut)
		rows = append(rows, Row{
			MutantID:        fmt.Sprintf("%s.%s.%s.s%d.%s", fixture, c.Class, c.ID, seed, mh[:8]),
			BaseFixture:     fixture,
			DefectClass:     c.Class,
			AnchorClaimID:   c.ID,
			Seed:            seed,
			InjectorVersion: version,
			BaseSHA256:      sha256Hex(clean),
			MutantSHA256:    mh,
			MutantText:      mut,
			CleanText:       clean,
			Scopes:          c.Scopes,
			Tags:            c.Tags,
			StageUnderTest:  "decompose+critic",
			Note:            noteFor(c),
		})
	}
	return rows, nil
}

func noteFor(c Claim) string {
	switch c.Class {
	case "SCP-1":
		return fmt.Sprintf("scope narrowed to %q while the claim stays asserted of %q; predicate frozen",
			c.Scopes.Narrow, c.Scopes.Broad)
	case "MFC":
		return "kill condition deleted"
	default:
		return ""
	}
}

// Scorable reports whether a row may enter its class number. A confounded row is
// generated and read like any other, but reported on its own line (A2.1).
func Scorable(r Row) bool {
	for _, t := range r.Tags {
		if t == ConfoundArith {
			return false
		}
	}
	return true
}
