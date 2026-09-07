package retrieve

// rank.go adds the second (semantic) ranker behind faithfulness retrieval and fuses it with BM25.
// The full-corpus judge quotes evidence spread across many hearings and often phrased without a
// claim's terms (a bare "Absolutely inadequate.", a cross-reference to another witness) — exactly
// where BM25 alone drops to rank>40 or score 0. A nomic-embed-text cosine ranker reaches some of
// that evidence; the two rankings are fused so a passage strong in either surfaces.
//
// Fusion is reciprocal-rank fusion combined by MAX, not the canonical sum. The measured evidence is
// complementary — the same passage is often strong in one ranker and weak in the other (a source
// span at BM25 rank 3 sits at cosine rank 147; another at cosine rank 6 sits at BM25 rank 46).
// Summing the reciprocal ranks rewards consensus and so penalises a true single-ranker hit — it
// dropped a passage the BM25-only retriever had covered. Taking the max (equivalently: ordering by
// each passage's BEST rank across the two rankers) keeps every ranker's strong hits near the top,
// which is what coverage needs. See rrfK.
//
// The embeddings are expensive to compute but stable, so they are cached to disk keyed by a hash of
// the corpus text (Load's passage order is deterministic). The first run with an embedder attached
// computes and writes the cache; every later run — the refuter and go test included — reads it with
// no model call. With neither a cache nor an embedder, Retrieve degrades to BM25-only rather than
// failing: the semantic ranker is an improvement, not a precondition.

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// rrfK is the reciprocal-rank-fusion constant. 60 is the value from Cormack et al. 2009. Under the
// max combiner (see the file comment) a passage's fused score is 1/(rrfK + its best rank across the
// two rankers), so ordering by score is ordering by best rank; rrfK only shapes the score's scale,
// not the order, and is kept at the standard value for continuity with the summed form.
const rrfK = 60.0

// Embedder produces one unit vector per text. It is implemented by internal/embed's ollama client;
// retrieve names it as an interface so the package carries no backend dependency and tests can pass a
// stub. Model is the id the cache is keyed on, so a model swap invalidates a stale cache.
type Embedder interface {
	Embed(texts []string) ([][]float32, error)
	Model() string
}

// Query is one retrieval request. Text is the claim, Hint the §-heading labels. Witness (a surname)
// and Session (a file-stem substring), when non-empty, name a speaker or hearing the claim explicitly
// attributes to — their passages are hard-filtered to the front ahead of the fused ranking.
type Query struct {
	Text    string
	Hint    string
	Witness string
	Session string
}

func (q Query) queryText() string { return strings.TrimSpace(q.Text + " " + q.Hint) }

// Result is a retrieval outcome. TopCosine is the best passage's semantic similarity to the query
// (0 when no embeddings); Below is set when TopCosine is under the caller's floor, the signal to
// return "absent" from code with no model call. Ranks maps every returned passage id to its 1-based
// fused rank, for the chain and the refuter's miss diagnosis.
type Result struct {
	Passages  []Passage
	TopCosine float64
	Below     bool
	Ranks     map[string]int
}

// AttachEmbeddings gives the index a semantic ranker. It loads corpus vectors from cacheDir when a
// cache for this corpus hash and model exists; otherwise, with emb != nil, it computes them and
// writes the cache. It also opens the query-vector cache (queries-<model>.json) for read, and for
// write-through when emb != nil. With no cache and no embedder it is a no-op — retrieval stays
// BM25-only. cacheDir is created if it does not exist.
func (ix *Index) AttachEmbeddings(emb Embedder, cacheDir string) error {
	model := DefaultEmbedModel
	if emb != nil {
		model = emb.Model()
	}
	ix.embModel = model
	ix.embedder = emb
	ix.cacheDir = cacheDir
	ix.hash = ix.corpusHash()

	corpusPath := filepath.Join(cacheDir, fmt.Sprintf("corpus-%s-%s.gz", model, ix.hash[:12]))
	if vecs, ok := loadCorpusCache(corpusPath, ix.hash, len(ix.Passages)); ok {
		ix.embed = vecs
	} else if emb != nil {
		texts := make([]string, len(ix.Passages))
		for i, p := range ix.Passages {
			texts[i] = p.searchText() // embed the question + answer, matching what BM25 and the judge see
		}
		vecs, err := emb.Embed(texts)
		if err != nil {
			return fmt.Errorf("embed corpus: %w", err)
		}
		ix.embed = vecs
		if err := saveCorpusCache(corpusPath, ix.hash, vecs); err != nil {
			return fmt.Errorf("write corpus cache: %w", err)
		}
	} else {
		return nil // no cache, no embedder: BM25-only, not an error
	}

	ix.queryPath = filepath.Join(cacheDir, fmt.Sprintf("queries-%s.json", model))
	ix.queryVecs = loadQueryCache(ix.queryPath)
	return nil
}

// HasEmbeddings reports whether a semantic ranker is available (cache loaded or embedder attached).
func (ix *Index) HasEmbeddings() bool { return len(ix.embed) == len(ix.Passages) && len(ix.embed) > 0 }

// BM25Order and EmbedOrder return every passage in each single ranker's order, for offline diagnosis
// of where a fused rank came from. BM25Order places score-0 passages last (their relative order by
// id); EmbedOrder is by descending cosine, or the corpus order when no embeddings are loaded.
func (ix *Index) BM25Order(q Query) []Passage {
	scores := ix.bm25Scores(tokenize(q.queryText()))
	return ix.orderBy(argsortDesc(scores))
}

func (ix *Index) EmbedOrder(q Query) []Passage {
	if !ix.HasEmbeddings() {
		return append([]Passage(nil), ix.Passages...)
	}
	return ix.orderBy(argsortDesc(ix.cosines(q.queryText())))
}

func (ix *Index) orderBy(order []int) []Passage {
	out := make([]Passage, len(order))
	for i, idx := range order {
		out[i] = ix.Passages[idx]
	}
	return out
}

// FusedRanking returns every passage in final fused order — the hard filter, then descending RRF —
// for offline diagnosis of what retrieval ranked where. Retrieve returns a token-bounded prefix of
// this; the refuter uses the full order to report the rank of a passage that fell outside the budget.
func (ix *Index) FusedRanking(q Query) []Passage {
	order, _ := ix.fusedOrder(q)
	out := make([]Passage, len(order))
	for i, idx := range order {
		out[i] = ix.Passages[idx]
	}
	return out
}

// Retrieve ranks the corpus for a query by reciprocal-rank fusion of BM25 and cosine, floats any
// witness/session-named passages to the front, and returns passages up to the token budget. floor
// (when >0 and embeddings are present) sets Below on the result if the best passage's cosine is under
// it — the caller then records "absent" with no model call. A zero/negative maxTokens is unbounded.
func (ix *Index) Retrieve(q Query, maxTokens int, floor float64) Result {
	order, topCos := ix.fusedOrder(q)

	res := Result{TopCosine: topCos, Ranks: map[string]int{}}
	if ix.HasEmbeddings() && floor > 0 && topCos < floor {
		res.Below = true
		return res // nothing clears the floor: absent, no passages, no model call
	}

	tokens := 0
	for rank, i := range order {
		if maxTokens > 0 && tokens+ix.docLen[i] > maxTokens && len(res.Passages) > 0 {
			break
		}
		res.Passages = append(res.Passages, ix.Passages[i])
		res.Ranks[ix.Passages[i].ID] = rank + 1
		tokens += ix.docLen[i]
	}
	return res
}

// fusedOrder returns passage indices in final retrieval order and the best passage's cosine. The
// order is: hard-filtered (witness/session-named) passages first, then everything else, each block by
// descending fused score (1/(rrfK+best rank), the max combiner) with a passage-id tie-break for
// determinism. When no embeddings are loaded it is BM25 rank alone (still with the hard filter), so
// the method degrades cleanly.
func (ix *Index) fusedOrder(q Query) (order []int, topCosine float64) {
	bm25Rank := ix.bm25Ranks(q.queryText())

	var cos []float64
	embRank := map[int]int{}
	if ix.HasEmbeddings() {
		cos = ix.cosines(q.queryText())
		for r, i := range argsortDesc(cos) {
			embRank[i] = r + 1
			if r == 0 {
				topCosine = cos[i]
			}
		}
	}

	type fused struct {
		i     int
		score float64
		front bool
	}
	scored := make([]fused, len(ix.Passages))
	for i := range ix.Passages {
		var s float64
		if r, ok := bm25Rank[i]; ok {
			s = 1 / (rrfK + float64(r))
		}
		if r, ok := embRank[i]; ok {
			if e := 1 / (rrfK + float64(r)); e > s {
				s = e
			}
		}
		scored[i] = fused{i: i, score: s, front: ix.namedByQuery(i, q)}
	}
	sort.SliceStable(scored, func(a, b int) bool {
		if scored[a].front != scored[b].front {
			return scored[a].front // hard-filtered passages first
		}
		if scored[a].score != scored[b].score {
			return scored[a].score > scored[b].score
		}
		return ix.Passages[scored[a].i].ID < ix.Passages[scored[b].i].ID
	})

	order = make([]int, len(scored))
	for k, f := range scored {
		order[k] = f.i
	}
	return order, topCosine
}

// namedByQuery reports whether passage i is spoken by the witness the query names, or comes from the
// session it names — the hard-filter test. Witness matches the speaker's final surname token
// case-insensitively; Session matches as a case-insensitive substring of the session file stem.
func (ix *Index) namedByQuery(i int, q Query) bool {
	p := ix.Passages[i]
	if q.Witness != "" {
		fields := strings.Fields(p.Speaker)
		if len(fields) > 0 && strings.EqualFold(fields[len(fields)-1], q.Witness) {
			return true
		}
	}
	if q.Session != "" && strings.Contains(strings.ToLower(p.Session), strings.ToLower(q.Session)) {
		return true
	}
	return false
}

// bm25Ranks returns the 1-based BM25 rank of every passage with a nonzero score. A passage that
// shares no query term (score 0) is absent from the map — it contributes no BM25 term to fusion,
// leaving the semantic ranker to reach it.
func (ix *Index) bm25Ranks(query string) map[int]int {
	scores := ix.bm25Scores(tokenize(query))
	idx := make([]int, len(scores))
	for i := range idx {
		idx[i] = i
	}
	sort.SliceStable(idx, func(a, b int) bool {
		if scores[idx[a]] != scores[idx[b]] {
			return scores[idx[a]] > scores[idx[b]]
		}
		return ix.Passages[idx[a]].ID < ix.Passages[idx[b]].ID
	})
	out := map[int]int{}
	for r, i := range idx {
		if scores[i] == 0 {
			break
		}
		out[i] = r + 1
	}
	return out
}

// cosines returns cosine(query, passage) for every passage. Vectors are already unit length, so the
// cosine is a dot product. Requires HasEmbeddings and an embedder or cached query vector; a query
// with no vector yields all-zero cosines (the caller then relies on BM25).
func (ix *Index) cosines(query string) []float64 {
	qv := ix.queryVector(query)
	out := make([]float64, len(ix.Passages))
	if qv == nil {
		return out
	}
	for i, pv := range ix.embed {
		var s float64
		for d := range pv {
			s += float64(pv[d]) * float64(qv[d])
		}
		out[i] = s
	}
	return out
}

// queryVector returns the (unit) embedding of a query, from the on-disk query cache when present, or
// by embedding it live and writing through to the cache. With no embedder and no cached vector it
// returns nil, and cosine ranking is skipped for that query.
func (ix *Index) queryVector(query string) []float32 {
	key := sha256Hex(ix.embModel + "\x00" + query)
	if v, ok := ix.queryVecs[key]; ok {
		return v
	}
	if ix.embedder == nil {
		return nil
	}
	vecs, err := ix.embedder.Embed([]string{query})
	if err != nil || len(vecs) != 1 {
		return nil
	}
	if ix.queryVecs == nil {
		ix.queryVecs = map[string][]float32{}
	}
	ix.queryVecs[key] = vecs[0]
	saveQueryCache(ix.queryPath, ix.queryVecs) // write-through; errors are non-fatal (cache is an optimisation)
	return vecs[0]
}

// corpusHash is a sha256 over every passage id and text in Load order — the key the embedding cache
// is stored under, so an edited corpus misses the cache and is recomputed rather than served stale.
func (ix *Index) corpusHash() string {
	h := sha256.New()
	for _, p := range ix.Passages {
		fmt.Fprintf(h, "%s\x00%s\n", p.ID, p.searchText()) // searchText so attaching context invalidates the cache
	}
	return hex.EncodeToString(h.Sum(nil))
}

// ── cache I/O ────────────────────────────────────────────────────────────────

// corpusCache is the on-disk form: the hash it was built for (checked on load so a hash collision in
// the truncated filename cannot serve wrong vectors) and one vector per passage in Load order.
type corpusCache struct {
	Hash string
	Vecs [][]float32
}

func loadCorpusCache(path, hash string, want int) ([][]float32, bool) {
	f, err := os.Open(path)
	if err != nil {
		return nil, false
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, false
	}
	defer gz.Close()
	var c corpusCache
	if err := gob.NewDecoder(gz).Decode(&c); err != nil {
		return nil, false
	}
	if c.Hash != hash || len(c.Vecs) != want {
		return nil, false // stale: corpus changed under this cache
	}
	return c.Vecs, true
}

func saveCorpusCache(path, hash string, vecs [][]float32) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	if err := gob.NewEncoder(gz).Encode(corpusCache{Hash: hash, Vecs: vecs}); err != nil {
		gz.Close()
		return err
	}
	return gz.Close()
}

func loadQueryCache(path string) map[string][]float32 {
	data, err := os.ReadFile(path)
	if err != nil {
		return map[string][]float32{}
	}
	var m map[string][]float32
	if json.Unmarshal(data, &m) != nil || m == nil {
		return map[string][]float32{}
	}
	return m
}

func saveQueryCache(path string, m map[string][]float32) {
	if path == "" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	data, err := json.Marshal(m)
	if err != nil {
		return
	}
	os.WriteFile(path, data, 0o644)
}

// ── small helpers ────────────────────────────────────────────────────────────

// DefaultEmbedModel names the model the cache is keyed on when no embedder is attached (cache-only
// reads). It must match internal/embed.DefaultModel; kept here so retrieve carries no embed import.
const DefaultEmbedModel = "nomic-embed-text"

// argsortDesc returns indices of v ordered by descending value, ties by index for determinism.
func argsortDesc(v []float64) []int {
	idx := make([]int, len(v))
	for i := range idx {
		idx[i] = i
	}
	sort.SliceStable(idx, func(a, b int) bool {
		if v[idx[a]] != v[idx[b]] {
			return v[idx[a]] > v[idx[b]]
		}
		return idx[a] < idx[b]
	})
	return idx
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
