package retrieve

// retrieve_test runs against the real committed LCEIC corpus (examples/vic-lceic/sources/hearings) —
// no synthetic transcript, per the repo's no-fabricated-inputs rule. It pins the two things a
// downstream judge relies on: that speaker turns split with correct role tags (a witness is not a
// questioner), and that BM25 retrieval is deterministic, bounded by k, and capped by tokens.

import (
	"strings"
	"testing"
)

const corpus = "../../examples/vic-lceic/sources/hearings"

func loadCorpus(t *testing.T) *Index {
	t.Helper()
	ix, err := Load(corpus)
	if err != nil {
		t.Fatalf("Load(%s): %v", corpus, err)
	}
	if len(ix.Passages) < 1000 {
		t.Fatalf("want a substantial corpus, got %d passages", len(ix.Passages))
	}
	return ix
}

// TestRolesFromRealCorpus pins the classification the A1 attack depends on: the Chair and committee
// MPs are questioners, the org representatives are witnesses. Names are read from the transcripts.
func TestRolesFromRealCorpus(t *testing.T) {
	ix := loadCorpus(t)
	role := map[string]string{}
	for _, s := range ix.Speakers() {
		role[s.Speaker] = s.Role
	}
	// Chair and known LCEIC members (from the MEMBERS header) must be questioner-side.
	for _, q := range []struct{ name, want string }{
		{"The CHAIR", RoleChair},
		{"David DAVIS", RoleQuestioner},
		{"Richard WELCH", RoleQuestioner},
		{"Gaelle BROAD", RoleQuestioner},
		{"Tom McINTOSH", RoleQuestioner},
	} {
		if role[q.name] != q.want {
			t.Errorf("%s: want %s, got %q", q.name, q.want, role[q.name])
		}
	}
	// Org representatives must be witnesses, never questioners — this is the distinction that lets a
	// judge reject a claim quoting a question as testimony.
	for _, w := range []string{"Vicky GUGLIELMO", "Claire FEBEY", "Kate FIELDING", "Craig BARRIE"} {
		if role[w] != RoleWitness {
			t.Errorf("%s: want witness, got %q", w, role[w])
		}
	}
}

// TestSearchDeterministicBoundedCapped pins that Search returns the same ids twice, honours k, and
// respects the token cap.
func TestSearchDeterministicBoundedCapped(t *testing.T) {
	ix := loadCorpus(t)
	const query = "ticket prices are a key barrier for people wanting to engage cost of living"

	a := ix.Search(query, 8, 8000)
	b := ix.Search(query, 8, 8000)
	if len(a) == 0 {
		t.Fatal("query returned no passages")
	}
	if len(a) > 8 {
		t.Errorf("k=8 exceeded: got %d", len(a))
	}
	if idsOf(a) != idsOf(b) {
		t.Errorf("non-deterministic: %v vs %v", idsOf(a), idsOf(b))
	}

	// Token cap: the selected passages' total length stays within the cap (allowing the one-passage
	// floor when a single passage is itself large).
	tokens := 0
	for _, p := range a {
		tokens += len(tokenize(p.Text))
	}
	if len(a) > 1 && tokens > 8000 {
		t.Errorf("token cap exceeded with %d passages: %d tokens", len(a), tokens)
	}

	// Relevance sanity: the top hit for a ticket-price query mentions tickets or prices.
	top := strings.ToLower(a[0].Text)
	if !strings.Contains(top, "ticket") && !strings.Contains(top, "price") {
		t.Errorf("top passage unrelated to query: %q", firstN(a[0].Text, 80))
	}
}

// TestSearchTokenCapForcesFewer confirms a tight token cap returns fewer than k passages.
func TestSearchTokenCapForcesFewer(t *testing.T) {
	ix := loadCorpus(t)
	full := ix.Search("funding creative australia regional victoria fair share", 8, 8000)
	tight := ix.Search("funding creative australia regional victoria fair share", 8, 120)
	if len(tight) >= len(full) || len(tight) == 0 {
		t.Errorf("tight cap should return fewer (but >0): tight=%d full=%d", len(tight), len(full))
	}
}

// ── test helpers ───────────────────────────────────────────────────────────────

func idsOf(ps []Passage) string {
	s := ""
	for _, p := range ps {
		s += p.ID + ";"
	}
	return s
}

func firstN(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
