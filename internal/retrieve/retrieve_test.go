package retrieve

// retrieve_test runs against the real committed LCEIC corpus (examples/vic-lceic/sources/hearings) —
// no synthetic transcript, per the repo's no-fabricated-inputs rule. It pins what a downstream judge
// relies on: that speaker turns split with correct role tags (a witness is not a questioner), that
// BM25 Search is deterministic, bounded by k, and token-capped, and that the fused Retrieve orders by
// best rank, floats a named witness first, and floors on cosine.

import (
	"os"
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

// withEmbeddings loads the corpus and attaches the committed embedding cache (no ollama). Tests that
// need the semantic ranker call it and skip when the cache is absent, so the suite stays green on a
// checkout that has not yet seeded the cache with ASSAY_EMBED=1.
func withEmbeddings(t *testing.T) *Index {
	t.Helper()
	ix := loadCorpus(t)
	if err := ix.AttachEmbeddings(nil, "../../examples/vic-lceic/sources/hearings/.embcache"); err != nil {
		t.Fatalf("attach embeddings: %v", err)
	}
	if !ix.HasEmbeddings() {
		t.Skip("no committed embedding cache; seed with ASSAY_EMBED=1 go test -run Refuter .")
	}
	return ix
}

// TestRetrieveDeterministicAndCapped pins that fused Retrieve returns the same ids twice and stays
// within the token budget (allowing the one-passage floor).
func TestRetrieveDeterministicAndCapped(t *testing.T) {
	ix := withEmbeddings(t)
	q := Query{Text: "regional victoria fair share of funding creative australia"}
	a := ix.Retrieve(q, 8000, 0)
	b := ix.Retrieve(q, 8000, 0)
	if len(a.Passages) == 0 {
		t.Fatal("no passages retrieved")
	}
	if idsOf(a.Passages) != idsOf(b.Passages) {
		t.Errorf("non-deterministic retrieval")
	}
	tokens := 0
	for _, p := range a.Passages {
		tokens += len(tokenize(p.Text))
	}
	if len(a.Passages) > 1 && tokens > 8000 {
		t.Errorf("token budget exceeded: %d passages, %d tokens", len(a.Passages), tokens)
	}
}

// TestRetrieveWitnessHardFilter pins that naming a witness floats every one of their turns ahead of
// all other speakers in the fused order — the 1b hard filter.
func TestRetrieveWitnessHardFilter(t *testing.T) {
	ix := withEmbeddings(t)
	const surname = "GUGLIELMO" // Vicky Guglielmo, a witness with turns in the corpus
	q := Query{Text: "cost of living ticket prices barrier", Witness: surname}
	order := ix.FusedRanking(q)
	seenOther := false
	named := 0
	for _, p := range order {
		fields := strings.Fields(p.Speaker)
		isNamed := len(fields) > 0 && strings.EqualFold(fields[len(fields)-1], surname)
		if isNamed {
			named++
			if seenOther {
				t.Fatalf("named witness turn %s ranked after a non-witness turn", p.ID)
			}
		} else {
			seenOther = true
		}
	}
	if named == 0 {
		t.Fatalf("witness %s has no turns — pick a speaker who does", surname)
	}
}

// TestRetrieveFloorAbsent pins that an impossibly high cosine floor suppresses everything (Below set,
// no passages) while a zero floor returns the fused set — the 1d code-side "absent" path.
func TestRetrieveFloorAbsent(t *testing.T) {
	ix := withEmbeddings(t)
	q := Query{Text: "regional victoria fair share of funding"}
	if hi := ix.Retrieve(q, 8000, 1.01); !hi.Below || len(hi.Passages) != 0 {
		t.Errorf("floor 1.01 should suppress all: below=%v passages=%d", hi.Below, len(hi.Passages))
	}
	if lo := ix.Retrieve(q, 8000, 0); lo.Below || len(lo.Passages) == 0 {
		t.Errorf("floor 0 should return passages: below=%v passages=%d", lo.Below, len(lo.Passages))
	}
}

// TestByIDsOracle pins that ByIDs returns exactly the requested passages in order and skips an id not
// in the corpus — the oracle retrieval mode's contract.
func TestByIDsOracle(t *testing.T) {
	ix := loadCorpus(t)
	want := []string{"2025-02-27/1_yarra-city-council#t17", "2025-02-27/1_yarra-city-council#t14"}
	got := ix.ByIDs(append(want, "no-such/passage#t999"))
	if len(got) != 2 {
		t.Fatalf("want 2 passages (unknown id skipped), got %d", len(got))
	}
	for i, p := range got {
		if p.ID != want[i] {
			t.Errorf("order not preserved: pos %d want %s got %s", i, want[i], p.ID)
		}
	}
}

// TestLoadSubmissions pins the written-submission ingestion: paragraphs (not speaker turns), each
// tagged Source=submission with a "submission-<n>#p<k>" id and the organisation from the filename.
// Skips when the gitignored submissions are not present (a fresh clone without the fetched PDFs).
func TestLoadSubmissions(t *testing.T) {
	const subs = "../../examples/vic-lceic/sources/submissions"
	if _, err := os.Stat(subs); err != nil {
		t.Skip("no fetched submissions locally; run the Playwright harvest first")
	}
	ix, err := Load(subs)
	if err != nil {
		t.Fatalf("Load(%s): %v", subs, err)
	}
	if len(ix.Passages) < 100 {
		t.Fatalf("want many submission paragraphs, got %d", len(ix.Passages))
	}
	for _, p := range ix.Passages {
		if p.Source != SourceSubmission {
			t.Fatalf("passage %s not tagged submission: %q", p.ID, p.Source)
		}
		if !strings.HasPrefix(p.ID, "submission-") || !strings.Contains(p.ID, "#p") {
			t.Errorf("submission id malformed: %q", p.ID)
		}
		if p.Speaker == "" {
			t.Errorf("submission %s has no organisation", p.ID)
		}
	}
	// LoadMany merges hearings + submissions into one index carrying both sources.
	both, err := LoadMany([]string{corpus, subs})
	if err != nil {
		t.Fatalf("LoadMany: %v", err)
	}
	var nH, nS int
	for _, p := range both.Passages {
		switch p.Source {
		case SourceHearing:
			nH++
		case SourceSubmission:
			nS++
		}
	}
	if nH == 0 || nS == 0 {
		t.Errorf("combined corpus missing a source: hearing=%d submission=%d", nH, nS)
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
