package retrieve

// segment_test.go pins the segmenter seam (segment.go, spec/TREE.md § The segmenter seam): the seam
// leaves every existing corpus's passage set byte-identical (the plain segmenter is additive, reached
// only by cited/), and the plain segmenter splits each held cited/ web-page capture into ≥3 paragraph
// passages with the cited/<stem> id base a route=evidence claim's cite resolves to.

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestSegmenterGoldenUnchanged is refuter (a): the seam alters nothing for the parsers it did not
// touch. Each existing corpus root is loaded and its passage set hashed field-for-field; the golden
// hashes were captured on the pre-seam code, so a mismatch means the refactor (splitFile → splitData,
// the added cited/ route) perturbed a corpus it must not. The plain segmenter is reached only by a
// cited/ path, absent from all of these, so its addition cannot show here — that is the point.
func TestSegmenterGoldenUnchanged(t *testing.T) {
	golden := []struct {
		root string
		n    int
		hash string
	}{
		{"../../examples/vic-lceic/sources/hearings", 1435, "e8c0b9bc4bc16829b0af1873a711c541523693febb5e583aacea5a4ce7cc1445"},
		{"../../examples/vic-lceic/sources/submissions", 691, "233e6e55314af60899824dc2c8213cb852bd0c0ddf755e4e6a86897d143b3242"},
		{"../../examples/vic-lceic/sources/qon", 39, "bbf7976a7907e48a450a52466990ffcf17bdf364025ee6f34ffae89d160557d0"},
		{"../../examples/dora-2026/sources/report.txt", 1072, "69a49e8e042644b8241ffe7c8e6f959e65cf9c40df762731d8277fb417df9f7c"},
		{"../../examples/quocirca-2026/sources/report.txt", 53, "0cb67bae1d6b41960bb699ea22cded74c641b8c467c6a4f3468dd52a8945c284"},
		{"../../examples/ai-index-2026-coding/sources/report.txt", 18, "0123ffde4076a600f38d12da663ce95cfa802dc27fa7c1f34d03ad3df41a9a18"},
		{"../../examples/ai-index-2026-coding/sources/leaderboards", 17, "aa6ebb80419f353aebe53317a2ba063bd24b828acf7a6dc7626ec090401ac8d1"},
		{"../../examples/ai-index-2026-coding/sources/papers", 5, "5d694227c7359b00a942f6474feb923b9ae2bc660acd75a9480fdecb79dd9ba1"},
	}
	// master-plan is deliberately absent: its sources/ is Markdown with no .txt, so it runs retrieve=none
	// (the whole file as one corpus passage) and never reaches a segmenter — nothing here to keep stable.
	checked := 0
	for _, g := range golden {
		if _, err := os.Stat(g.root); err != nil {
			t.Errorf("%s: corpus not present (%v) — the golden refuter cannot run without it", g.root, err)
			continue
		}
		ix, err := Load(g.root)
		if err != nil {
			t.Errorf("Load(%s): %v", g.root, err)
			continue
		}
		if len(ix.Passages) != g.n {
			t.Errorf("%s: passage count %d, golden %d", g.root, len(ix.Passages), g.n)
		}
		if got := passageSetHash(ix.Passages); got != g.hash {
			t.Errorf("%s: passage set hash %s, golden %s — the seam perturbed this corpus", g.root, got, g.hash)
		}
		checked++
	}
	if checked == 0 {
		t.Fatal("no corpus roots checked — an empty golden run is a failed run, not a clean one")
	}
}

// TestPlainSegmenterCited is refuter (b): every held cited/ capture yields ≥3 paragraph passages with
// the cited/<stem> id base (so a claim's cite id cited/<stem>.txt resolves to it, extension dropped)
// and Source=cited. The "figure sentence lands in exactly one passage" sub-check is demonstrated on
// two captures the MANIFEST marks fig=present with a verbatim anchor quote (07 pacing, 06 wemustactnow);
// it is not asserted for the other held captures because the MANIFEST records their cited figure as
// figure-only or absent-in-extraction — the number lives in an interactive chart, not the static text,
// so there is no verbatim sentence to land. That gap is the corpus's, recorded in sources/MANIFEST.md
// § Slice-3 cited sources, not the segmenter's.
func TestPlainSegmenterCited(t *testing.T) {
	const dir = "../../examples/tai-europe-2026/sources/cited"
	// The 16 HELD cited captures (08b Bloomberg 403 and 09 The Information login-wall are held on disk
	// but stay out of this set — the walls make them not usably in the corpus, MANIFEST § Slice-3).
	held := []string{
		"01-epoch.ai", "02-metr.org", "03-anthropic.com", "04-epoch.ai", "06-wemustactnow.ai",
		"07-pacingthefrontier.com", "08a-epoch.ai", "10-epoch.ai", "11-internationalaisafetyreport.org",
		"12-redwoodresearch.org", "13-forbes.com", "14-epoch.ai-aicompanies", "16a-europe2031.ai",
		"16b-epoch.ai-supercomputers", "17-epoch.ai-fdc-hub", "20-foxphilip.substack.com",
	}
	if _, err := os.Stat(dir); err != nil {
		t.Skipf("tai-europe cited corpus not present: %v", err)
	}
	for _, stem := range held {
		path := dir + "/" + stem + ".txt"
		ix, err := Load(path)
		if err != nil {
			t.Errorf("Load(%s): %v", path, err)
			continue
		}
		if len(ix.Passages) < 3 {
			t.Errorf("%s: %d passages, want ≥3 (a capture that splits into < 3 is not usably retrievable)", stem, len(ix.Passages))
		}
		wantBase := "cited/" + stem
		for _, p := range ix.Passages {
			if passageBase(p.ID) != wantBase {
				t.Errorf("%s: passage id %q base is not %q", stem, p.ID, wantBase)
			}
			if p.Source != SourceCited {
				t.Errorf("%s: passage %q Source=%q, want %q", stem, p.ID, p.Source, SourceCited)
			}
		}
	}

	// fig=present anchor sentences (MANIFEST § Slice-3 cited sources) each land in exactly one passage.
	present := map[string]string{
		"07-pacingthefrontier.com": "could be close to automating AI research",
		"06-wemustactnow.ai":       "unprecedented transformation of our economy",
	}
	for stem, quote := range present {
		ix, err := Load(dir + "/" + stem + ".txt")
		if err != nil {
			t.Errorf("Load(%s): %v", stem, err)
			continue
		}
		n := 0
		for _, p := range ix.Passages {
			if strings.Contains(p.Text, quote) {
				n++
			}
		}
		if n != 1 {
			t.Errorf("%s: anchor quote %q lands in %d passages, want exactly 1 (0 = split across a paragraph break)", stem, quote, n)
		}
	}
}

// ── test helpers ───────────────────────────────────────────────────────────────

// passageSetHash hashes a passage slice field-by-field with NUL separators, so the golden refuter
// pins every field of every passage in order — a reordering or a single changed byte breaks it.
func passageSetHash(ps []Passage) string {
	h := sha256.New()
	for _, p := range ps {
		fmt.Fprintf(h, "%s\x00%s\x00%s\x00%s\x00%s\x00%s\x00%s\x00%s\x00%s\x00%s\n",
			p.ID, p.Date, p.Session, p.Speaker, p.Role, p.Text, p.Context, p.ContextSpeaker, p.ContextRole, p.Source)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
