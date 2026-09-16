package retrieve

// segment.go is the segmenter seam: a Segmenter turns one held document into passages, so the parser
// a corpus file gets is chosen by an explicit registry rather than a wall of hardcoded path
// substrings. spec/TREE.md § The segmenter seam is the contract. Only two segmenters are named here —
// the Hansard default (committee transcripts, speaker turns) and the plain paragraph splitter used for
// cited/ web-page captures; the submission/qon/leaderboard/report parsers keep their own functions in
// retrieve.go and passagesForFile still routes to them, so this file adds the seam without disturbing
// their byte-identical output.

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Segmenter turns one held document (its path and bytes already read) into passages. Split takes the
// bytes rather than re-reading, so passagesForFile reads a file once and hands the data on.
type Segmenter interface {
	Split(path string, data []byte) ([]Passage, error)
}

// HansardSegmenter is the default: it wraps splitData (splitFile's body) — the speaker-turn parser,
// unchanged. A committee transcript with no cited/ or submission routing lands here.
type HansardSegmenter struct{}

func (HansardSegmenter) Split(path string, data []byte) ([]Passage, error) {
	return splitData(path, data)
}

// plainMergeFloor is the rune length below which a paragraph is a fragment (a wrapped heading, a nav
// item, a stray line from an HTML-to-text strip) rather than a passage of its own. A fragment is held
// and merged into the next paragraph so it does not become a one-line passage; "~200 chars" per the
// spec.
const plainMergeFloor = 200

// PlainSegmenter splits a document into paragraph passages on blank lines, merging each fragment under
// plainMergeFloor runes into the NEXT paragraph — a fetched web-page capture strips to a mass of short
// nav/heading lines around the prose, and attaching them forward keeps the prose whole rather than
// scattering it across sub-passages. It is used for any held id under cited/ (a route=evidence claim's
// external truth-maker, no speaker turns). The id base is "cited/<stem>" so a claim's cite id
// cited/<stem>.txt resolves to it with the extension dropped (assay.go citedExternalBases).
type PlainSegmenter struct{}

func (PlainSegmenter) Split(path string, data []byte) ([]Passage, error) {
	stem := strings.TrimSuffix(filepath.Base(path), ".txt")
	base := "cited/" + stem
	text := strings.ReplaceAll(string(data), "\f", "\n\n")

	var out []Passage
	turn := 0
	var buf []string
	emit := func(paras []string) {
		p := strings.TrimSpace(strings.Join(paras, "\n\n"))
		if p == "" {
			return
		}
		out = append(out, Passage{
			ID: fmt.Sprintf("%s#p%d", base, turn), Session: stem,
			Speaker: stem, Role: SourceCited, Source: SourceCited, Text: p,
		})
		turn++
	}
	for _, para := range paraSplit.Split(text, -1) {
		p := strings.TrimSpace(para)
		if p == "" {
			continue
		}
		buf = append(buf, p)
		if len([]rune(p)) < plainMergeFloor {
			continue // a fragment: hold it and attach to the next substantial paragraph
		}
		emit(buf)
		buf = nil
	}
	emit(buf) // trailing fragments (no following paragraph to merge into) become one final passage
	return out, nil
}
