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

// TestCanonicalDoc pins the passage-id → canonical-doc-id inverse for every passage shape (hearing,
// submission + attachment, qon, report) and that a malformed id resolves to ok=false. The doc ids must
// be exactly those Scan mints (TestScanDerivesCanonicalIDs), or a resolved quote would miss its
// manifest entry — except report.pdf, whose id is the passage-prefix convention, not a Scan output.
func TestCanonicalDoc(t *testing.T) {
	cases := []struct {
		pid, doc, loc string
		ok            bool
	}{
		{"2025-03-13/4_public-galleries-association-of-victoria#t48",
			"hearing:2025-03-13/4_public-galleries-association-of-victoria", "line 48", true},
		{"submission-19#p9", "submission:19", "p.9", true},
		{"submission-09#p3", "submission:9", "p.3", true}, // zero-padded number folds to 9
		{"submission-33.1#p5", "submission:33/attachment-1", "p.5", true},
		{"qon-abc-2025-03-21#p2", "qon:abc/2025-03-21", "p.2", true},
		{"report#p2#0", "report.pdf", "p.2", true},   // report as its own source; paragraph index dropped
		{"report#p13#4", "report.pdf", "p.13", true}, // locator is the page alone
		{"no-hash-fragment", "", "", false},
		{"submission-notanumber#p1", "", "", false},
	}
	for _, c := range cases {
		doc, loc, ok := CanonicalDoc(c.pid)
		if ok != c.ok || doc != c.doc || loc != c.loc {
			t.Errorf("CanonicalDoc(%q) = (%q,%q,%v), want (%q,%q,%v)",
				c.pid, doc, loc, ok, c.doc, c.loc, c.ok)
		}
	}
}

// TestLoadLabels pins that the witness/author is the last " — "-segment of each manifest bullet, keyed
// by the canonical id, and that a doc absent from the manifest yields no label (the render's unresolved
// path). It writes a manifest from Scan so the fixture is the real Render output, not hand-shaped text.
func TestLoadLabels(t *testing.T) {
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
	touch("hearings/2025-03-13/4_public-galleries-association-of-victoria.txt")
	touch("submissions/19.-theatre-network-australia.txt")
	docs, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	mpath := filepath.Join(root, "MANIFEST.md")
	if err := os.WriteFile(mpath, []byte(Render(docs)), 0o644); err != nil {
		t.Fatal(err)
	}
	labels, err := LoadLabels(mpath)
	if err != nil {
		t.Fatal(err)
	}
	if got := labels["hearing:2025-03-13/4_public-galleries-association-of-victoria"]; got != "Public Galleries Association Of Victoria" {
		t.Errorf("hearing label: %q", got)
	}
	if got := labels["submission:19"]; got != "Theatre Network Australia" {
		t.Errorf("submission label: %q", got)
	}
	if _, ok := labels["submission:99"]; ok {
		t.Errorf("absent doc should have no label")
	}
	if s, err := LoadLabels(filepath.Join(root, "nope.md")); err != nil || len(s) != 0 {
		t.Errorf("missing manifest should be empty/no-error, got %d, %v", len(s), err)
	}
}

// TestLoadReportLinks pins the report-link rule parser: the printed→PDF offset and the named-section
// page table, both read from manifest prose that is not a held-document bullet (so LoadHeld ignores
// them). The manifest text here is config, not a source document — no fabricated corpus. It also
// checks a numbered-only manifest (LCEIC shape) yields its offset and an empty table, and that a
// missing file is the zero offset with no error.
func TestLoadReportLinks(t *testing.T) {
	dir := t.TempDir()
	named := filepath.Join(dir, "NAMED.md")
	if err := os.WriteFile(named, []byte(`# m
- `+"`report.pdf`"+` — Quocirca report

report_page_offset: 0

sections:
  Executive summary: 2
  Key findings: 4
`), 0o644); err != nil {
		t.Fatal(err)
	}
	off, secs, err := LoadReportLinks(named)
	if err != nil || off != 0 || secs["Executive summary"] != 2 || secs["Key findings"] != 4 {
		t.Fatalf("named manifest: off=%d secs=%v err=%v", off, secs, err)
	}
	// The section rows must not be read as held documents.
	if held, _ := LoadHeld(named); len(held) != 1 || !held["report.pdf"] {
		t.Fatalf("section rows leaked into held set: %v", held)
	}

	numbered := filepath.Join(dir, "NUM.md")
	if err := os.WriteFile(numbered, []byte("# m\n\nreport_page_offset: 18\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if off, secs, _ := LoadReportLinks(numbered); off != 18 || len(secs) != 0 {
		t.Fatalf("numbered manifest: off=%d secs=%v, want 18 and empty", off, secs)
	}

	if off, secs, err := LoadReportLinks(filepath.Join(dir, "nope.md")); err != nil || off != 0 || len(secs) != 0 {
		t.Fatalf("missing manifest: off=%d secs=%v err=%v, want 0/empty/nil", off, secs, err)
	}
}

// TestLoadSingleSource pins the single_source declaration parser: a manifest carrying `single_source:
// true` reads true, one without it reads false (the LCEIC multi-source shape), and a missing file reads
// false with no error. The flag is what makes a leaf `absent` render `uncorroborated` downstream.
func TestLoadSingleSource(t *testing.T) {
	dir := t.TempDir()
	with := filepath.Join(dir, "SINGLE.md")
	if err := os.WriteFile(with, []byte("# m\n\nsingle_source: true\n\nreport_page_offset: 0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if ok, err := LoadSingleSource(with); err != nil || !ok {
		t.Fatalf("single_source manifest: ok=%v err=%v, want true/nil", ok, err)
	}

	without := filepath.Join(dir, "MULTI.md")
	if err := os.WriteFile(without, []byte("# m\n\nreport_page_offset: 18\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if ok, err := LoadSingleSource(without); err != nil || ok {
		t.Fatalf("multi-source manifest: ok=%v err=%v, want false/nil", ok, err)
	}

	if ok, err := LoadSingleSource(filepath.Join(dir, "nope.md")); err != nil || ok {
		t.Fatalf("missing manifest: ok=%v err=%v, want false/nil", ok, err)
	}
}
