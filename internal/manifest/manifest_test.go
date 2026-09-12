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
// manifest entry — except a report excerpt PDF, whose id is the passage-prefix convention, not a Scan
// output. A report excerpt id (the second "#") resolves to its OWN PDF: with the two-excerpt reports
// list, "report" → report.pdf and "report-productivity" → report-productivity.pdf; with no list, each
// stem falls back to "<stem>.pdf".
func TestCanonicalDoc(t *testing.T) {
	reports := []Report{{File: "report.pdf"}, {File: "report-productivity.pdf"}}
	cases := []struct {
		pid, doc, loc string
		ok            bool
		reports       []Report
	}{
		{"2025-03-13/4_public-galleries-association-of-victoria#t48",
			"hearing:2025-03-13/4_public-galleries-association-of-victoria", "line 48", true, reports},
		{"submission-19#p9", "submission:19", "p.9", true, reports},
		{"submission-09#p3", "submission:9", "p.3", true, reports}, // zero-padded number folds to 9
		{"submission-33.1#p5", "submission:33/attachment-1", "p.5", true, reports},
		{"qon-abc-2025-03-21#p2", "qon:abc/2025-03-21", "p.2", true, reports},
		{"report#p2#0", "report.pdf", "p.2", true, reports},                              // base excerpt; paragraph index dropped
		{"report#p13#4", "report.pdf", "p.13", true, reports},                            // locator is the page alone
		{"report-productivity#p221#3", "report-productivity.pdf", "p.221", true, reports}, // second excerpt → its own PDF
		{"report#p2#0", "report.pdf", "p.2", true, nil},                                  // no reports list: "<stem>.pdf" fallback
		{"no-hash-fragment", "", "", false, reports},
		{"submission-notanumber#p1", "", "", false, reports},
	}
	for _, c := range cases {
		doc, loc, ok := CanonicalDoc(c.pid, c.reports)
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

// TestLoadReportsSingle pins the backward path: a flat single-report manifest (report_page_offset +
// top-level sections, the shape every existing corpus uses) reads as one report.pdf claiming every
// leaf, and a missing manifest yields one zero-offset report.pdf. The manifest text is config, not a
// held document — no fabricated corpus.
func TestLoadReportsSingle(t *testing.T) {
	dir := t.TempDir()
	flat := filepath.Join(dir, "FLAT.md")
	if err := os.WriteFile(flat, []byte("# m\n\nreport_page_offset: 18\n\nsections:\n  Executive summary: 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	reps, err := LoadReports(flat)
	if err != nil || len(reps) != 1 || reps[0].File != "report.pdf" || reps[0].Offset != 18 ||
		reps[0].Sections["Executive summary"] != 2 || len(reps[0].Prefixes) != 0 {
		t.Fatalf("flat manifest: reps=%+v err=%v", reps, err)
	}
	// A report with no prefixes claims every leaf.
	if r, ok := ReportFor(reps, "SB6"); !ok || r.File != "report.pdf" {
		t.Fatalf("single-report ReportFor: r=%+v ok=%v, want report.pdf/true", r, ok)
	}

	reps, err = LoadReports(filepath.Join(dir, "nope.md"))
	if err != nil || len(reps) != 1 || reps[0].File != "report.pdf" || reps[0].Offset != 0 {
		t.Fatalf("missing manifest: reps=%+v err=%v, want one zero-offset report.pdf", reps, err)
	}
}

// TestLoadReportsMulti pins the multi-report path: a `reports:` list reads as one report per entry,
// each with its own file, offset, prefixes, and sections, and ReportFor routes a claim id to the entry
// whose prefixes list its prefix. An id no entry claims returns ok=false, so the ref stays unlinked.
// The two entries mirror the ai-index-2026-coding wiring (report.pdf / report-productivity.pdf).
func TestLoadReportsMulti(t *testing.T) {
	dir := t.TempDir()
	multi := filepath.Join(dir, "MULTI.md")
	body := `# m
- ` + "`paper:cui-2025`" + ` — a held study

single_source: false

reports:
  - file: report.pdf
    offset: -99
    prefixes: SEC SW SB TB VC
    sections:
      Software: 100
  - file: report-productivity.pdf
    offset: -218
    prefixes: EP
    sections:
      Productivity Trends: 219
`
	if err := os.WriteFile(multi, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	reps, err := LoadReports(multi)
	if err != nil || len(reps) != 2 {
		t.Fatalf("multi manifest: reps=%+v err=%v, want 2 reports", reps, err)
	}
	if reps[0].File != "report.pdf" || reps[0].Offset != -99 || reps[0].Sections["Software"] != 100 ||
		len(reps[0].Prefixes) != 5 {
		t.Fatalf("report.pdf entry: %+v", reps[0])
	}
	if reps[1].File != "report-productivity.pdf" || reps[1].Offset != -218 ||
		reps[1].Sections["Productivity Trends"] != 219 || len(reps[1].Prefixes) != 1 {
		t.Fatalf("report-productivity.pdf entry: %+v", reps[1])
	}
	// Routing: SB → report.pdf, EP → report-productivity.pdf, an unclaimed prefix → not found.
	if r, ok := ReportFor(reps, "SB6"); !ok || r.File != "report.pdf" {
		t.Errorf("SB6 routed to %+v ok=%v, want report.pdf", r, ok)
	}
	if r, ok := ReportFor(reps, "EP8"); !ok || r.File != "report-productivity.pdf" {
		t.Errorf("EP8 routed to %+v ok=%v, want report-productivity.pdf", r, ok)
	}
	if _, ok := ReportFor(reps, "ZZ1"); ok {
		t.Errorf("ZZ1 (no report claims it) should not route")
	}
	// The held-document bullet is still read, and the reports list is not mistaken for one.
	if held, _ := LoadHeld(multi); !held["paper:cui-2025"] || len(held) != 1 {
		t.Errorf("held set leaked reports rows or dropped the study: %v", held)
	}
	if ok, _ := LoadSingleSource(multi); ok {
		t.Errorf("single_source: false must read false")
	}
}
