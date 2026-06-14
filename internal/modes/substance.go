// SPDX-License-Identifier: Apache-2.0

package modes

import (
	"github.com/peterwilliams97/elenchus2/internal/claims"
	"github.com/peterwilliams97/elenchus2/internal/client"
)

// --- JSON shapes the model returns (parsed by the code right here) -------------

type producerJSON struct {
	Steelman   string `json:"steelman"`
	Conditions string `json:"conditions"`
}

type axisJSON struct {
	Axis     string `json:"axis"`
	Finding  string `json:"finding"`
	Severity string `json:"severity"`
}

// substanceJSON is the critic's response shape — a per-call JSON DTO, flat.
type substanceJSON struct {
	Critique                   []axisJSON `json:"critique"`
	Verdict                    string     `json:"verdict"`
	SurvivingClaim             *string    `json:"surviving_claim"` // null on "hollow"
	Reason                     string     `json:"reason"`
	NeedsAnotherRound          bool       `json:"needs_another_round"`
	AddedConditions            int        `json:"added_conditions"`
	SurvivesOnlyByConditioning bool       `json:"survives_only_by_conditioning"`
}

func (s substanceJSON) surviving() string {
	if s.SurvivingClaim == nil {
		return ""
	}
	return *s.SurvivingClaim
}

// Decompose breaks raw input prose into atomic claim strings — the first stage
// of substance mode (spec/BEHAVIOR.md Mode 1).
func Decompose(c *client.Client, input string) ([]string, error) {
	var arr []string
	if err := c.CallJSON(decomposeSys, "TEXT:\n"+input, &arr); err != nil {
		return nil, err
	}
	return arr, nil
}

// AssayClaim runs the producer–critic loop on one claim (spec/BEHAVIOR.md Mode 1,
// assayClaim). maxRounds bounds the loop; the continue requires all four of
// {needs_another_round, a non-empty surviving_claim, rounds<maxRounds, NOT
// survives_only_by_conditioning}. After the loop, survives_only_by_conditioning
// forces the verdict to hollow.
func AssayClaim(c *client.Client, claim string, maxRounds int) claims.Substance {
	current := claim
	rounds := 0
	var last substanceJSON
	var steelman string

	for rounds < maxRounds {
		var p producerJSON
		if err := c.CallJSON(producerSys, "CLAIM:\n"+current, &p); err != nil {
			return errSubstance(claim, err)
		}
		steelman = p.Steelman

		u := "CLAIM:\n" + current +
			"\n\nPRODUCER STEELMAN:\n" + p.Steelman +
			"\n\nPRODUCER CONDITIONS:\n" + p.Conditions

		last = substanceJSON{}
		if err := c.CallJSON(substanceCriticSys, u, &last); err != nil {
			return errSubstance(claim, err)
		}
		rounds++

		if last.NeedsAnotherRound && last.surviving() != "" && rounds < maxRounds && !last.SurvivesOnlyByConditioning {
			current = last.surviving()
			continue
		}
		break
	}

	verdict := last.Verdict
	reason := last.Reason
	if last.SurvivesOnlyByConditioning {
		verdict = claims.Hollow
		reason += " Survives only by conditions the speaker never stated."
	}
	if !claims.ValidSubstanceVerdict(verdict) {
		return claims.Substance{Claim: claim, Verdict: claims.Error, Reason: "critic returned unknown verdict " + verdict}
	}

	return claims.Substance{
		Claim:   claim,
		Verdict: verdict,
		Reason:  reason,
		Detail: claims.SubstanceDetail{
			Steelman:        steelman,
			Critique:        toAxes(last.Critique),
			SurvivingClaim:  last.surviving(),
			AddedConditions: last.AddedConditions,
			Rounds:          rounds,
		},
	}
}

func errSubstance(claim string, err error) claims.Substance {
	return claims.Substance{Claim: claim, Verdict: claims.Error, Reason: err.Error()}
}

func toAxes(in []axisJSON) []claims.Axis {
	if len(in) == 0 {
		return nil
	}
	out := make([]claims.Axis, len(in))
	for i, a := range in {
		out[i] = claims.Axis{Axis: a.Axis, Finding: a.Finding, Severity: a.Severity}
	}
	return out
}
