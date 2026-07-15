// SPDX-License-Identifier: Apache-2.0

package mutbench

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

// Axis mirrors one critique entry in the chain JSONL (cmd/crossexam/chain.go).
type Axis struct {
	Axis     string `json:"axis"`
	Finding  string `json:"finding"`
	Severity string `json:"severity"`
}

// ChainDetail mirrors the chain's detail object. Only the fields the read sheet
// and the axis-fired diagnostic need are declared; the rest is ignored by
// encoding/json.
type ChainDetail struct {
	CritiqueByAxis []Axis `json:"critique_by_axis"`
	SurvivingClaim string `json:"surviving_claim"`
	Reason         string `json:"reason"`
}

// ChainRec is one record of crossexam's Tier-2 chain: one graded claim.
type ChainRec struct {
	Idx     int         `json:"idx"`
	Total   int         `json:"total"`
	Claim   string      `json:"claim"`
	Verdict string      `json:"verdict"`
	Detail  ChainDetail `json:"detail"`
}

// ReadChain parses a chain JSONL stream into its records, in file order.
func ReadChain(r io.Reader) ([]ChainRec, error) {
	var out []ChainRec
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		if len(sc.Bytes()) == 0 {
			continue
		}
		var rec ChainRec
		if err := json.Unmarshal(sc.Bytes(), &rec); err != nil {
			return nil, fmt.Errorf("chain line %d: %w", len(out)+1, err)
		}
		out = append(out, rec)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read chain: %w", err)
	}
	return out, nil
}

// AxisFired reports whether any axis emitted a non-empty critique on any claim
// derived from this input (A3.2).
//
// This is a harness diagnostic. It is NOT recall and must never be reported as,
// aggregated with, or converted into recall (A-T4). An axis fires on nearly
// every claim, so a high axis-fired rate says almost nothing about whether the
// injected defect was caught — that is target-catch, and target-catch is a human
// read. The 2026-07-12 run is the standing reminder: a keyword tally certified
// 5/5 where the human read showed 0/5.
func AxisFired(recs []ChainRec) bool {
	for _, r := range recs {
		if len(r.Detail.CritiqueByAxis) > 0 {
			return true
		}
	}
	return false
}

// Cardinality is how many claims decompose returned for this input. Recorded per
// input as a diagnostic, ungated (A3.2/A3.3).
func Cardinality(recs []ChainRec) int { return len(recs) }
