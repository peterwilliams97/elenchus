package manifest

// manifest_test pins the filename→canonical-id derivation and the Render/LoadHeld round trip. It uses
// a temp tree of EMPTY files — Scan reads filenames, never content — so it fabricates no source text.

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestScanDerivesCanonicalIDs(t *testing.T) {
	root := t.TempDir()
	touch := func(rel string) {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	touch("hearings/2025-02-27/1_yarra-city-council.txt")
	touch("submissions/09.-ana-a-new-approach-redacted.txt")
	touch("submissions/19.-theatre-network-australia.txt")
	touch("submissions/33.1-public-galleries-association-of-victoria_redacted.txt")
	touch("qon/abc-2025-03-21.txt")

	docs, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	got := make([]string, len(docs))
	for i, d := range docs {
		got[i] = d.ID
	}
	sort.Strings(got)
	want := []string{
		"hearing:2025-02-27/1_yarra-city-council",
		"qon:abc/2025-03-21",
		"submission:19",
		"submission:33/attachment-1",
		"submission:9",
	}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("ids: got %v, want %v", got, want)
	}

	// Render → LoadHeld must round-trip the id set exactly.
	mpath := filepath.Join(root, "MANIFEST.md")
	if err := os.WriteFile(mpath, []byte(Render(docs)), 0o644); err != nil {
		t.Fatal(err)
	}
	held, err := LoadHeld(mpath)
	if err != nil {
		t.Fatalf("LoadHeld: %v", err)
	}
	if len(held) != len(want) {
		t.Fatalf("round-trip size: got %d, want %d", len(held), len(want))
	}
	for _, id := range want {
		if !held[id] {
			t.Errorf("round-trip lost %s", id)
		}
	}

	// A missing manifest is an empty set, not an error.
	if s, err := LoadHeld(filepath.Join(root, "nope.md")); err != nil || len(s) != 0 {
		t.Errorf("missing manifest should be empty/no-error, got %d, %v", len(s), err)
	}
}
