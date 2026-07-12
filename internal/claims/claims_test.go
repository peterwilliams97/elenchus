// SPDX-License-Identifier: Apache-2.0

package claims

import "testing"

func TestValidSubstanceVerdict(t *testing.T) {
	for _, v := range []string{Substantive, Partial, Hollow, Error} {
		if !ValidSubstanceVerdict(v) {
			t.Errorf("%q should be a valid substance verdict", v)
		}
	}
	// Reject the empty string, near-misses, other modes' verdicts, and the
	// "substantial" misspelling the v1 critic once emitted (PLAN.md §2).
	for _, v := range []string{"", "substantial", "supported", "faithful", "hollow ", "Hollow"} {
		if ValidSubstanceVerdict(v) {
			t.Errorf("%q should NOT be a valid substance verdict", v)
		}
	}
}
