package retrieve

// retrieve turns a corpus of committee-hearing transcripts into speaker-turn passages and ranks them
// against a claim with BM25 — deterministically, with no model call — so a faithfulness judge sees
// only the passages a claim is about, not the whole 25K-token corpus. spec/CLI.md §"Retrieval" is the
// contract.
//
// A transcript is Hansard-style: a header naming the committee MEMBERS (the MPs) and the Chair, then a
// body of turns, each opening with a speaker line — "Vicky GUGLIELMO:", "The CHAIR:". A turn runs from
// one speaker line to the next. Every passage is tagged with its date, session (source file), speaker,
// and role. Role matters downstream: a claim that quotes a questioner's question as if it were witness
// testimony is a distortion, and only a role tag lets the judge catch it — so a speaker whose surname
// is on the committee roster, or who is the Chair, is tagged `questioner`, everyone else `witness`.

import (
	"bufio"
	"bytes"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Role classifies who is speaking. questioner and chair are the committee side (their words are
// questions, not testimony); witness is everyone appearing before it.
const (
	RoleWitness    = "witness"
	RoleQuestioner = "questioner"
	RoleChair      = "chair"
)

// Passage is one speaker turn. ID is stable across runs ("<date>/<file>#t<n>") so it can be recorded
// in the chain and matched back to the source. A witness turn also carries, tagged separately, the
// questioner turn that immediately precedes it (Context/ContextSpeaker/ContextRole): a bare answer
// ("Absolutely inadequate.") is meaningless without the question it answers, and the question is
// where the topical terms live. Text stays the witness's own words alone, so a quote still matches the
// answer precisely; ranking and the judge see the question via searchText and Format.
type Passage struct {
	ID             string
	Date           string
	Session        string // source file stem, e.g. "1_yarra-city-council"
	Speaker        string
	Role           string
	Text           string
	Context        string // the immediately preceding questioner/chair turn, "" if none
	ContextSpeaker string
	ContextRole    string
	Source         string // "hearing" (speaker turns) or "submission" (paragraphs); reports evidence origin
}

// Source values distinguish a hearing-transcript passage (speaker turns) from a written-submission
// passage (paragraphs). The faithfulness table reports which kind a verified quote came from.
const (
	SourceHearing    = "hearing"
	SourceSubmission = "submission"
	SourceQoN        = "qon" // a response to questions on notice (paragraphs, like a submission)
)

// searchText is what ranking and embedding see: the question (when present) then the answer. Text
// alone is what a quote is matched against, so keeping Context out of Text preserves precise quote
// attribution while letting the question's terms lift the passage's rank.
func (p Passage) searchText() string {
	if p.Context == "" {
		return p.Text
	}
	return p.Context + "\n\n" + p.Text
}

// Index is a searchable corpus of passages plus the BM25 statistics over them. The embedding fields
// are populated by AttachEmbeddings (rank.go) and are the second, semantic ranker; they stay nil for
// a BM25-only index.
type Index struct {
	Passages []Passage
	docTerms []map[string]int // term frequencies per passage
	docLen   []int            // token count per passage
	df       map[string]int   // document frequency per term
	avgLen   float64

	embed     [][]float32          // per-passage unit vector, aligned to Passages; nil if none
	embModel  string               // embedding model id the caches are keyed on
	embedder  Embedder             // live query embedder; nil in cache-only (offline) mode
	cacheDir  string               // where corpus/query caches live
	hash      string               // corpus hash the vectors were built for
	queryPath string               // query-vector cache file
	queryVecs map[string][]float32 // query text hash → unit vector
}

const (
	bm25K1 = 1.5
	bm25B  = 0.75
)

// speakerLine matches a turn opener: optional indent, a name whose final surname token carries a run
// of ≥2 uppercase letters (Hansard caps surnames — "McINTOSH", "WELCH", "GUGLIELMO"), or the literal
// "The CHAIR", followed by a colon. The ≥2-uppercase test is what separates a real speaker line from
// an ordinary "Word:" mid-sentence.
var speakerLine = regexp.MustCompile(`^ {0,8}(The CHAIR|[A-Z][A-Za-z.'’-]*(?: [A-Z][A-Za-z.'’-]*)*[A-Z]{2}[A-Za-z.'’-]*):\s`)

// rosterName pulls "Firstname Surname" entries from the MEMBERS header block. `[A-Z][a-zA-Z]+` keeps
// internal caps as one token so "McIntosh"/"McArthur" survive whole (a `[a-z]+` surname class would
// split them and lose the match). Two names share a line (a two-column layout), so it is applied with
// FindAllString.
var rosterName = regexp.MustCompile(`[A-Z][a-zA-Z]+(?: [A-Z][a-zA-Z]+)+`)

// firstSpeaker (multiline) locates where the header ends and the body begins — the roster must be read
// only up to here, or witness self-introductions ("My name is Vicky Guglielmo") leak in as questioners.
var firstSpeaker = regexp.MustCompile(`(?m)` + speakerLine.String())

var wordRe = regexp.MustCompile(`[a-z0-9]+`)

// Load walks path (a file or a directory tree) and splits every .txt into speaker-turn passages,
// returning a BM25 index over them. The directory layout carries provenance: the parent directory is
// the hearing date, the file stem is the session.
func Load(path string) (*Index, error) {
	var files []string
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		err = filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && strings.HasSuffix(p, ".txt") {
				files = append(files, p)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		sort.Strings(files) // deterministic passage order across runs
	} else {
		files = []string{path}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no .txt files under %s", path)
	}

	ix := &Index{df: map[string]int{}}
	for _, f := range files {
		ps, err := passagesForFile(f)
		if err != nil {
			return nil, err
		}
		ix.Passages = append(ix.Passages, ps...)
	}
	if len(ix.Passages) == 0 {
		return nil, fmt.Errorf("no passages parsed from %s (no speaker turns or paragraphs matched)", path)
	}
	ix.build()
	return ix, nil
}

// LoadMany loads several corpus roots into one index — e.g. the hearing transcripts and the written
// submissions — so a claim retrieves across both. Passage order is deterministic: roots in argument
// order, files sorted within each.
func LoadMany(paths []string) (*Index, error) {
	ix := &Index{df: map[string]int{}}
	for _, root := range paths {
		part, err := Load(root)
		if err != nil {
			return nil, err
		}
		ix.Passages = append(ix.Passages, part.Passages...)
	}
	if len(ix.Passages) == 0 {
		return nil, fmt.Errorf("no passages parsed from %v", paths)
	}
	ix.build()
	return ix, nil
}

// passagesForFile routes a corpus file to its parser: a written submission (path under a
// "submissions" directory) splits on paragraphs, everything else on Hansard speaker turns.
func passagesForFile(path string) ([]Passage, error) {
	switch {
	case strings.Contains(path, "/submissions/"):
		return submissionPassages(path)
	case strings.Contains(path, "/qon/"):
		return qonPassages(path)
	default:
		return splitFile(path)
	}
}

// qonPassages splits a response to questions on notice into paragraph passages, tagged Source=qon. A
// QoN response is not Hansard (no speaker turns) and not a submission; it is the document a report
// footnote cites as "response to questions on notice", so its passages must be retrievable and
// distinguishable in the evidence-origin table. The filename stem "<org>-<date>" becomes the speaker.
func qonPassages(path string) ([]Passage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	stem := strings.TrimSuffix(filepath.Base(path), ".txt")
	org := stem
	if m := qonName.FindStringSubmatch(stem); m != nil {
		org = strings.ToUpper(m[1]) + " (QoN " + m[2] + ")"
	}
	text := strings.ReplaceAll(string(data), "\f", "\n\n")
	var out []Passage
	turn := 0
	for _, para := range paraSplit.Split(text, -1) {
		p := strings.TrimSpace(para)
		if len([]rune(p)) < 40 {
			continue
		}
		out = append(out, Passage{
			ID: fmt.Sprintf("qon-%s#p%d", stem, turn), Session: stem,
			Speaker: org, Role: SourceQoN, Source: SourceQoN, Text: p,
		})
		turn++
	}
	return out, nil
}

// qonName matches a QoN filename stem "<org>-<YYYY-MM-DD>".
var qonName = regexp.MustCompile(`^(.+)-(\d{4}-\d{2}-\d{2})$`)

// subName pulls the submission number and organisation from a filename stem like
// "09.-ana-a-new-approach-redacted" → ("09", "ana a new approach") or "01.1-...-redacted" → ("01.1", …).
var subName = regexp.MustCompile(`^(\d+(?:\.\d+)?)[.\-]+(.+?)(?:[_-]redacted)?$`)

// paraSplit separates paragraphs on a blank line (a form feed, from pdftotext page breaks, counts).
var paraSplit = regexp.MustCompile(`\n\s*\n`)

// submissionPassages splits a written-submission .txt into paragraph passages, each tagged with the
// submission number and organisation from the filename. A submission has no speaker turns, so there is
// no role or question context — the passage is a paragraph, and Source marks it a submission so the
// judge and the table can tell it from hearing testimony. Paragraphs under 40 runes (page numbers,
// stray headers from pdftotext) are dropped.
func submissionPassages(path string) ([]Passage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	stem := strings.TrimSuffix(filepath.Base(path), ".txt")
	number, org := stem, ""
	if m := subName.FindStringSubmatch(stem); m != nil {
		number = m[1]
		org = strings.ReplaceAll(m[2], "-", " ")
	}
	text := strings.ReplaceAll(string(data), "\f", "\n\n")
	var out []Passage
	turn := 0
	for _, para := range paraSplit.Split(text, -1) {
		p := strings.TrimSpace(para)
		if len([]rune(p)) < 40 {
			continue
		}
		out = append(out, Passage{
			ID:      fmt.Sprintf("submission-%s#p%d", number, turn),
			Session: stem, Speaker: org, Role: SourceSubmission, Source: SourceSubmission, Text: p,
		})
		turn++
	}
	return out, nil
}

// splitFile parses one transcript into passages: read the roster, then accumulate lines into the
// current turn until the next speaker line.
func splitFile(path string) ([]Passage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	date := filepath.Base(filepath.Dir(path))
	session := strings.TrimSuffix(filepath.Base(path), ".txt")
	roster := parseRoster(string(data))

	var (
		out     []Passage
		speaker string
		role    string
		buf     []string
		turn    int
	)
	flush := func() {
		if speaker == "" {
			return
		}
		text := strings.TrimSpace(strings.Join(buf, "\n"))
		if text == "" {
			return
		}
		out = append(out, Passage{
			ID:   fmt.Sprintf("%s/%s#t%d", date, session, turn),
			Date: date, Session: session, Speaker: speaker, Role: role, Text: text,
			Source: SourceHearing,
		})
		turn++
	}

	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if m := speakerLine.FindStringSubmatch(line); m != nil {
			flush()
			speaker = strings.TrimSpace(m[1])
			role = classify(speaker, roster)
			// The remainder of the speaker line after "Name:" is the first line of the turn.
			buf = []string{strings.TrimSpace(line[len(m[0]):])}
			continue
		}
		if speaker != "" {
			buf = append(buf, line)
		}
	}
	flush()
	// Attach each witness turn's immediately preceding questioner/chair turn as tagged context. The
	// turns in `out` are one session in transcript order, so out[i-1] is the immediately preceding turn.
	for i := range out {
		if out[i].Role != RoleWitness || i == 0 {
			continue
		}
		if p := out[i-1]; p.Role == RoleQuestioner || p.Role == RoleChair {
			out[i].Context = p.Text
			out[i].ContextSpeaker = p.Speaker
			out[i].ContextRole = p.Role
		}
	}
	return out, sc.Err()
}

// parseRoster extracts the committee members' surnames from the MEMBERS / PARTICIPATING MEMBERS header
// block — the authoritative set of questioners for that hearing. Returns uppercased surnames.
func parseRoster(doc string) map[string]bool {
	set := map[string]bool{}
	start := strings.Index(doc, "MEMBERS")
	if start < 0 {
		return set
	}
	// The roster runs from the first "MEMBERS" to where the body starts — the first speaker line, or a
	// "WITNESSES" header, whichever comes first. Bounding here is what keeps witness names (listed in
	// the WITNESSES block and stated in the body) out of the questioner set.
	rest := doc[start:]
	end := len(rest)
	if loc := firstSpeaker.FindStringIndex(rest); loc != nil && loc[0] < end {
		end = loc[0]
	}
	if w := strings.Index(rest, "WITNESS"); w >= 0 && w < end {
		end = w // matches both "WITNESSES" and the singular "WITNESS (via videoconference)"
	}
	for _, name := range rosterName.FindAllString(rest[:end], -1) {
		fields := strings.Fields(name)
		surname := strings.ToUpper(fields[len(fields)-1])
		set[surname] = true
	}
	return set
}

// classify tags a speaker. The Chair is always a questioner; a speaker whose surname is on the roster
// is a questioner; everyone else is a witness. Surname match is on the final name token, uppercased,
// because Hansard renders it in caps ("McINTOSH" → "MCINTOSH", roster "Tom McIntosh" → "MCINTOSH").
func classify(speaker string, roster map[string]bool) string {
	if strings.Contains(strings.ToUpper(speaker), "CHAIR") {
		return RoleChair // "The CHAIR", "The DEPUTY CHAIR"
	}
	fields := strings.Fields(speaker)
	if len(fields) == 0 {
		return RoleWitness
	}
	surname := strings.ToUpper(fields[len(fields)-1])
	if roster[surname] {
		return RoleQuestioner
	}
	return RoleWitness
}

// build computes the BM25 statistics over the loaded passages.
func (ix *Index) build() {
	ix.docTerms = make([]map[string]int, len(ix.Passages))
	ix.docLen = make([]int, len(ix.Passages))
	total := 0
	for i, p := range ix.Passages {
		tf := map[string]int{}
		n := 0
		for _, tok := range tokenize(p.searchText()) {
			tf[tok]++
			n++
		}
		ix.docTerms[i] = tf
		ix.docLen[i] = n
		total += n
		for term := range tf {
			ix.df[term]++
		}
	}
	if len(ix.Passages) > 0 {
		ix.avgLen = float64(total) / float64(len(ix.Passages))
	}
}

// bm25Scores returns the BM25 score of every passage against the query terms, in passage order. It is
// the shared scoring both Search (top-k, token-capped) and bm25Ranks (fusion) call, so the two never
// diverge on how a passage is scored.
func (ix *Index) bm25Scores(qTerms []string) []float64 {
	n := float64(len(ix.Passages))
	out := make([]float64, len(ix.Passages))
	for i := range ix.Passages {
		var s float64
		dl := float64(ix.docLen[i])
		for _, qt := range qTerms {
			tf := float64(ix.docTerms[i][qt])
			if tf == 0 {
				continue
			}
			df := float64(ix.df[qt])
			idf := math.Log((n-df+0.5)/(df+0.5) + 1)
			s += idf * (tf * (bm25K1 + 1)) / (tf + bm25K1*(1-bm25B+bm25B*dl/ix.avgLen))
		}
		out[i] = s
	}
	return out
}

// scored pairs a passage index with its BM25 score for ranking.
type scored struct {
	i     int
	score float64
}

// Search returns the top-k passages for a query, in rank order, stopping early once the running token
// budget (maxTokens, approx by word count) would be exceeded. Ties break by passage ID so the result
// is deterministic. A zero or negative k or maxTokens disables that bound.
func (ix *Index) Search(query string, k, maxTokens int) []Passage {
	scores := ix.bm25Scores(tokenize(query))
	ranked := make([]scored, len(ix.Passages))
	for i := range ix.Passages {
		ranked[i] = scored{i, scores[i]}
	}
	sort.SliceStable(ranked, func(a, b int) bool {
		if ranked[a].score != ranked[b].score {
			return ranked[a].score > ranked[b].score
		}
		return ix.Passages[ranked[a].i].ID < ix.Passages[ranked[b].i].ID
	})

	var out []Passage
	tokens := 0
	for _, r := range ranked {
		if r.score == 0 {
			break // no query term present — nothing to retrieve beyond here
		}
		if k > 0 && len(out) >= k {
			break
		}
		if maxTokens > 0 && tokens+ix.docLen[r.i] > maxTokens && len(out) > 0 {
			break // token cap reached; keep at least one passage
		}
		out = append(out, ix.Passages[r.i])
		tokens += ix.docLen[r.i]
	}
	return out
}

// SpeakerInfo is one distinct speaker and the role the corpus assigned them.
type SpeakerInfo struct {
	Speaker string
	Role    string
	Turns   int
}

// Speakers returns the distinct speakers across the corpus, questioners/chair first then witnesses,
// each alphabetical — the roster the speaker-detection check prints.
func (ix *Index) Speakers() []SpeakerInfo {
	seen := map[string]*SpeakerInfo{}
	for _, p := range ix.Passages {
		key := p.Speaker
		if s, ok := seen[key]; ok {
			s.Turns++
		} else {
			seen[key] = &SpeakerInfo{Speaker: p.Speaker, Role: p.Role, Turns: 1}
		}
	}
	out := make([]SpeakerInfo, 0, len(seen))
	for _, s := range seen {
		out = append(out, *s)
	}
	rank := map[string]int{RoleChair: 0, RoleQuestioner: 1, RoleWitness: 2}
	sort.Slice(out, func(a, b int) bool {
		if rank[out[a].Role] != rank[out[b].Role] {
			return rank[out[a].Role] < rank[out[b].Role]
		}
		return out[a].Speaker < out[b].Speaker
	})
	return out
}

// Format renders passages as the judge sees them: a provenance header per passage (id, date, speaker,
// role) above its text, so the model can tell a witness answer from a questioner's question.
func Format(ps []Passage) string {
	var b strings.Builder
	for _, p := range ps {
		if p.Source == SourceSubmission {
			// A written submission: a paragraph, tagged with its number and organisation so the judge
			// reads it as a submission, not hearing testimony.
			fmt.Fprintf(&b, "[%s · written submission · %s]\n%s\n\n", p.ID, p.Speaker, p.Text)
			continue
		}
		if p.Source == SourceQoN {
			fmt.Fprintf(&b, "[%s · response to questions on notice · %s]\n%s\n\n", p.ID, p.Speaker, p.Text)
			continue
		}
		fmt.Fprintf(&b, "[%s · %s · %s (%s)]\n", p.ID, p.Date, p.Speaker, p.Role)
		if p.Context != "" {
			// The question is shown first, tagged as the questioner's, so the judge reads the witness
			// answer in the context it was given — and never mistakes the question for testimony.
			fmt.Fprintf(&b, "Q — %s (%s): %s\nA — %s (%s): %s\n\n",
				p.ContextSpeaker, p.ContextRole, p.Context, p.Speaker, p.Role, p.Text)
			continue
		}
		fmt.Fprintf(&b, "%s\n\n", p.Text)
	}
	return strings.TrimRight(b.String(), "\n")
}

// ByIDs returns the passages with the given ids, in the order requested, skipping any id not in the
// corpus. It backs the oracle retrieval mode — feeding the judge a fixed, hand-specified passage set
// (the gold evidence) instead of a ranked one, to isolate judge quality from retrieval quality.
func (ix *Index) ByIDs(ids []string) []Passage {
	byID := make(map[string]Passage, len(ix.Passages))
	for _, p := range ix.Passages {
		byID[p.ID] = p
	}
	out := make([]Passage, 0, len(ids))
	for _, id := range ids {
		if p, ok := byID[id]; ok {
			out = append(out, p)
		}
	}
	return out
}

// IDs returns the passage ids in order, for the chain record.
func IDs(ps []Passage) []string {
	out := make([]string, len(ps))
	for i, p := range ps {
		out[i] = p.ID
	}
	return out
}

func tokenize(s string) []string { return wordRe.FindAllString(strings.ToLower(s), -1) }
