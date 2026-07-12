// SPDX-License-Identifier: Apache-2.0

package modes

import (
	"strings"
	"testing"

	"github.com/peterwilliams97/elenchus2/internal/claims"
	"github.com/peterwilliams97/elenchus2/internal/client"
)

func newClient(texts ...string) *client.Client {
	return client.New(client.Config{APIKey: "k", Model: "m", HTTP: client.Stub(texts...)})
}

const (
	prod   = `{"steelman":"S","conditions":"C"}`
	critOK = `{"critique":[{"axis":"Evidence","finding":"f","severity":"clears"}],` +
		`"verdict":"substantive","surviving_claim":"narrowed","reason":"R",` +
		`"needs_another_round":false,"added_conditions":0,"survives_only_by_conditioning":false}`
)

// Proof-1 wiring test: Decompose then AssayClaim, driven entirely through the
// client fake — a substance verdict comes out, no network.
func TestDecomposeThenAssayWiring(t *testing.T) {
	c := newClient(`["The future of work happens in Claude Code."]`, prod, critOK)
	cl, err := Decompose(c, "The future of work happens in Claude Code.")
	if err != nil {
		t.Fatalf("Decompose: %v", err)
	}
	if len(cl) != 1 {
		t.Fatalf("decomposed into %d claims, want 1", len(cl))
	}
	got := AssayClaim(c, cl[0], 2)
	if got.Verdict != claims.Substantive {
		t.Errorf("verdict = %q, want substantive", got.Verdict)
	}
	if !claims.ValidSubstanceVerdict(got.Verdict) {
		t.Errorf("verdict %q is not a valid substance verdict", got.Verdict)
	}
}

// Example C (spec/BEHAVIOR.md): needs_another_round=false on round 0 → single
// round, exactly two API calls (producer + critic).
func TestAssaySingleRoundExit(t *testing.T) {
	c := newClient(prod, critOK)
	got := AssayClaim(c, "claim", 2)
	if got.Verdict != claims.Substantive {
		t.Errorf("verdict = %q, want substantive", got.Verdict)
	}
	if got.Detail.Rounds != 1 {
		t.Errorf("rounds = %d, want 1", got.Detail.Rounds)
	}
	if c.Usage().Calls != 2 {
		t.Errorf("calls = %d, want 2 (one producer + one critic)", c.Usage().Calls)
	}
}

// Example A (spec/BEHAVIOR.md): round 0 asks for another round and produces a
// surviving claim; at maxRounds=2 the loop runs round 1 then stops. The verdict
// is round 1's, and rounds == 2.
func TestAssayStopsAtMaxRounds(t *testing.T) {
	crit0 := `{"critique":[],"verdict":"partial","surviving_claim":"narrowed X","reason":"r0",` +
		`"needs_another_round":true,"added_conditions":0,"survives_only_by_conditioning":false}`
	crit1 := `{"critique":[],"verdict":"substantive","surviving_claim":"narrowed X","reason":"r1",` +
		`"needs_another_round":true,"added_conditions":0,"survives_only_by_conditioning":false}`
	c := newClient(prod, crit0, prod, crit1)
	got := AssayClaim(c, "claim", 2)
	if got.Verdict != claims.Substantive {
		t.Errorf("verdict = %q, want substantive (round 1's)", got.Verdict)
	}
	if got.Reason != "r1" {
		t.Errorf("reason = %q, want r1 (round 1's)", got.Reason)
	}
	if got.Detail.Rounds != 2 {
		t.Errorf("rounds = %d, want 2", got.Detail.Rounds)
	}
	if c.Usage().Calls != 4 {
		t.Errorf("calls = %d, want 4 (two rounds × producer+critic)", c.Usage().Calls)
	}
}

// Example B (spec/BEHAVIOR.md): survives_only_by_conditioning=true forces the
// verdict to hollow at loop exit and appends the laundering note to the reason,
// regardless of the verdict the critic returned.
func TestAssayConditionLaunderingDowngrade(t *testing.T) {
	crit := `{"critique":[],"verdict":"substantive","surviving_claim":"narrowed only via invented qualifier",` +
		`"reason":"looks fine","needs_another_round":true,"added_conditions":2,` +
		`"survives_only_by_conditioning":true}`
	c := newClient(prod, crit)
	got := AssayClaim(c, "claim", 2)
	if got.Verdict != claims.Hollow {
		t.Errorf("verdict = %q, want hollow (downgraded)", got.Verdict)
	}
	if !strings.Contains(got.Reason, "Survives only by conditions the speaker never stated.") {
		t.Errorf("reason missing laundering note: %q", got.Reason)
	}
	// The downgrade fires immediately: only round 0 ran (the continue is blocked
	// by survives_only_by_conditioning), so exactly two calls.
	if c.Usage().Calls != 2 {
		t.Errorf("calls = %d, want 2 (downgrade breaks immediately)", c.Usage().Calls)
	}
}

// An unknown verdict string from the critic must not pass through as a verified
// win — it is routed to "error".
func TestAssayUnknownVerdictRoutesToError(t *testing.T) {
	crit := `{"critique":[],"verdict":"bogus","surviving_claim":null,"reason":"r",` +
		`"needs_another_round":false,"added_conditions":0,"survives_only_by_conditioning":false}`
	c := newClient(prod, crit)
	got := AssayClaim(c, "claim", 2)
	if got.Verdict != claims.Error {
		t.Errorf("verdict = %q, want error (unknown verdict must not pass as a win)", got.Verdict)
	}
}

// A failed API call routes the claim to verdict "error", not a panic.
func TestAssayAPIErrorRoutesToError(t *testing.T) {
	// "nope" never parses as producer JSON; CallJSON's one retry also fails.
	c := newClient("nope")
	got := AssayClaim(c, "claim", 2)
	if got.Verdict != claims.Error {
		t.Errorf("verdict = %q, want error", got.Verdict)
	}
}
