// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/peterwilliams97/elenchus2/internal/claims"
)

// chainRecord is the per-claim Tier-2 record (spec/CLI.md) — a per-call JSON DTO,
// flat.
type chainRecord struct {
	Idx      int             `json:"idx"`
	Total    int             `json:"total"`
	Mode     string          `json:"mode"`
	Claim    string          `json:"claim"`
	Verdict  string          `json:"verdict"`
	ElapsedS float64         `json:"elapsed_s"`
	Detail   substanceDetail `json:"detail"`
}

// substanceDetail mirrors the spec/CLI.md substanceDetail schema.
type substanceDetail struct {
	Steelman        string     `json:"steelman"`
	CritiqueByAxis  []axisJSON `json:"critique_by_axis"`
	SurvivingClaim  string     `json:"surviving_claim"`
	AddedConditions int        `json:"added_conditions"`
	Rounds          int        `json:"rounds"`
	Reason          string     `json:"reason"`
}

type axisJSON struct {
	Axis     string `json:"axis"`
	Finding  string `json:"finding"`
	Severity string `json:"severity"`
}

// writeChain writes one JSONL record per claim to <dir>/<fixture>.<mode>.jsonl,
// where dir defaults to eval/<stamp>-<model>/. It is best-effort: any failure is
// logged to stderr and an empty path is returned (so the SUMMARY detail: line is
// omitted), never aborting the run (spec/CLI.md).
func writeChain(stderr io.Writer, dir, fixture, model, mode string, results []claims.Substance, elapsed []time.Duration) string {
	if dir == "" {
		stamp := time.Now().Format("20060102-1504")
		dir = filepath.Join("eval", stamp+"-"+model)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		emit(stderr, fmt.Sprintf("[chain] mkdir %s: %v\n", dir, err))
		return ""
	}
	path := filepath.Join(dir, fixture+"."+mode+".jsonl")
	f, err := os.Create(path)
	if err != nil {
		emit(stderr, fmt.Sprintf("[chain] create %s: %v\n", path, err))
		return ""
	}
	defer func() { _ = f.Close() }()

	enc := json.NewEncoder(f)
	for i, r := range results {
		rec := chainRecord{
			Idx: i, Total: len(results), Mode: mode,
			Claim: r.Claim, Verdict: r.Verdict, ElapsedS: elapsed[i].Seconds(),
			Detail: substanceDetail{
				Steelman:        r.Detail.Steelman,
				CritiqueByAxis:  axesOut(r.Detail.Critique),
				SurvivingClaim:  r.Detail.SurvivingClaim,
				AddedConditions: r.Detail.AddedConditions,
				Rounds:          r.Detail.Rounds,
				Reason:          r.Reason,
			},
		}
		if err := enc.Encode(rec); err != nil {
			emit(stderr, fmt.Sprintf("[chain] write %s idx=%d: %v\n", path, i, err))
			return ""
		}
	}
	return path
}

func axesOut(in []claims.Axis) []axisJSON {
	if len(in) == 0 {
		return nil
	}
	out := make([]axisJSON, len(in))
	for i, a := range in {
		out[i] = axisJSON{Axis: a.Axis, Finding: a.Finding, Severity: a.Severity}
	}
	return out
}
