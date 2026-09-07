package main

// assay — a dialectical filter for claims.
//
// Three modes, one per sense of "does this hold up":
//   default          substance     is the claim well-formed / falsifiable?
//   -source FILE     faithfulness  did the source actually say it?
//   -evidence        grounding     is it true, per external evidence (web search)?
//   -audit -source F  all three, emitted as a cross-tabulation
//
// Output is a colour terminal view by default, or clean markdown tables with -md
// (always markdown for -audit).

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"assay/internal/backend"
	"assay/internal/backend/anthropic"
	"assay/internal/backend/ollama"
	"assay/internal/brief"
	"assay/internal/embed"
	"assay/internal/retrieve"
	"assay/internal/tree"
)

const defaultModel = "claude-sonnet-4-6"

// retrievedSource is a URL actually fetched during a web_search_tool_result round trip — distinct
// from source, which is what the model claims it used in its JSON response. Aliased to backend.Source
// so the chain schema and the backend seam name one type.
type retrievedSource = backend.Source

// ── config ────────────────────────────────────────────────────────────────────

type cfg struct {
	model        string
	apiKey       string
	maxRounds    int
	maxClaims    int
	repeat       int // -n: run each claim this many times and report the modal verdict + agreement
	verbose      bool
	noColor      bool
	asMarkdown   bool
	showProgress bool
	quiet        bool
	usageOut     string
	chainFile    string              // Tier-2 JSONL destination; set by runners before case loop
	renderMode   string              // stdout renderer: "brief" (default), "full", or "tree"
	treeAll      bool                // -tree=full: expand every node rather than only Needs-you branches
	retrieveMode string              // "bm25" (default) retrieves per-claim passages; "none" sends the full corpus
	maxTokens    int                 // -max-tokens: retrieval token budget per claim (fused bm25+embed ranking)
	floor        float64             // -floor: top-passage cosine below this ⇒ absent from code, no model call
	embed        bool                // -embed: add the nomic-embed-text ranker, fused with BM25 by RRF
	oracle       map[string][]string // -retrieve=oracle: claim-id → fixed gold passage ids
	index        *retrieve.Index     // built once from the source corpus when retrieveMode != "none"
	auditPath    string              // full-table sink; every run writes it, whichever renderer stdout gets
	treeHTMLPath string              // eval/<stamp>/tree.html sink, written alongside audit.md
	usage        *usageCounters
	tally        *runTally
	// cachedSource is the stable prefix (e.g. the source transcript) placed in a Request's Cached
	// field, which the Anthropic backend turns into an ephemeral cache block ahead of the variable
	// prompt and the Ollama backend folds into the prompt. Set per-call by callJSON/callJSONSourced
	// from their `cached` argument; "" sends no prefix.
	cachedSource string
	// reqSchema/reqSchemaName/reqTemp are the per-call structured-output controls dispatch threads into
	// the backend Request, set by callSchema on its value copy (like cachedSource). Empty schema and a
	// nil temp leave a call unconstrained at the provider default.
	reqSchema     json.RawMessage
	reqSchemaName string
	reqTemp       *float64
	// backend is the LLM provider dispatch uses when `call` is nil (production). Set in main from
	// -backend; nil under -from, where no model call is made.
	backend backend.Backend
	// backendName is the provider id stamped into the chain and header ("anthropic"|"ollama").
	backendName string
	// call is the low-level test seam. When non-nil, dispatch uses it instead of `backend`, so the
	// behaviour tests drive the real mode runners with canned JSON and no network. Returns the
	// response text, any retrieved sources, and an error.
	call func(system, prompt string, withTools bool) (string, []retrievedSource, error)
}

// runTally tracks verified/errored counts across a single assay run for the SUMMARY block.
type runTally struct {
	mu       sync.Mutex
	total    int
	verified int // cases that got a real verdict (not "error")
	counts   map[string]int
}

func newRunTally(total int) *runTally {
	return &runTally{total: total, counts: make(map[string]int)}
}

func (rt *runTally) record(verdict string) {
	rt.mu.Lock()
	if verdict == "error" {
		rt.counts["error"]++
	} else {
		rt.verified++
		rt.counts[verdict]++
	}
	rt.mu.Unlock()
}

func (rt *runTally) snapshot() (total, verified int, counts map[string]int) {
	rt.mu.Lock()
	cp := make(map[string]int, len(rt.counts))
	for k, v := range rt.counts {
		cp[k] = v
	}
	total = rt.total
	verified = rt.verified
	rt.mu.Unlock()
	return total, verified, cp
}

func main() {
	var c cfg
	var src, text, chainDir, fromChain string
	var ev, audit, full bool
	var treeF treeFlag
	var backendName, ollamaURL string
	var think bool
	flag.StringVar(&c.model, "model", envOr("ANTHROPIC_MODEL", defaultModel), "model id")
	flag.StringVar(&backendName, "backend", "anthropic", "LLM backend: anthropic|ollama")
	flag.StringVar(&ollamaURL, "ollama-url", envOr("OLLAMA_HOST", ollama.DefaultBaseURL), "ollama server base URL")
	flag.BoolVar(&think, "think", false, "ollama: emit the model's reasoning block (default off; on needs a higher token cap)")
	var speakers bool
	var embedModel, oracleFile string
	flag.StringVar(&c.retrieveMode, "retrieve", "bm25", "per-claim passage retrieval: bm25|none|oracle (none sends the full corpus; oracle reads -oracle)")
	flag.StringVar(&oracleFile, "oracle", "", "oracle retrieval: JSON map of claim-id → [passage-id]; the judge sees exactly those passages")
	flag.IntVar(&c.maxTokens, "max-tokens", retrieveTokenCap, "retrieval token budget per claim (fused bm25+embed ranking)")
	flag.Float64Var(&c.floor, "floor", 0, "retrieval floor: top passage cosine below this ⇒ verdict absent, no model call (0=off)")
	flag.BoolVar(&c.embed, "embed", true, "add the nomic-embed-text ranker fused with BM25 by RRF (needs local ollama or a committed cache)")
	flag.StringVar(&embedModel, "embed-model", embed.DefaultModel, "embedding model for the semantic ranker")
	flag.BoolVar(&speakers, "speakers", false, "print the distinct speakers + roles found in -source, then exit")
	flag.StringVar(&src, "source", "", "transcript file or corpus dir → faithfulness mode")
	flag.BoolVar(&ev, "evidence", false, "evidence-grounding mode (web search)")
	flag.BoolVar(&audit, "audit", false, "run all three modes and emit a cross-tab (needs -source)")
	flag.StringVar(&text, "text", "", "inline input instead of a file")
	flag.BoolVar(&c.asMarkdown, "md", false, "emit markdown tables")
	flag.BoolVar(&c.verbose, "v", false, "verbose: show every API call")
	flag.BoolVar(&c.verbose, "verbose", false, "verbose: show every API call")
	flag.BoolVar(&c.noColor, "no-color", false, "disable ANSI colour")
	flag.IntVar(&c.maxRounds, "max-rounds", 2, "producer-critic rounds per claim (substance)")
	flag.IntVar(&c.maxClaims, "max-claims", 0, "bound evidence grounding (0 = unlimited)")
	flag.IntVar(&c.repeat, "n", 1, "repeat each claim N times, show modal verdict + agreement (faithfulness)")
	flag.BoolVar(&c.showProgress, "progress", true, "show per-claim progress on stderr (default on)")
	flag.BoolVar(&c.quiet, "quiet", false, "suppress per-case lines + heartbeat; keeps SUMMARY and writes Tier-2")
	flag.StringVar(&c.usageOut, "usage-out", "", "append one JSON record per run to this file")
	flag.StringVar(&chainDir, "chain-dir", "", "directory for Tier-2 JSONL verification chain (default: eval/<stamp>/)")
	flag.BoolVar(&full, "full", false, "print the full table to stdout instead of the brief report")
	flag.Var(&treeF, "tree", "print the tree report to stdout; -tree=full expands every node")
	flag.StringVar(&fromChain, "from", "", "render brief/tree/audit from a saved chain JSONL (no model calls)")
	flag.Parse()

	// -full and -tree select different stdout renderers; refuse to guess which the caller meant.
	if full && treeF.on {
		fatal("-full and -tree select different stdout renderers; choose one")
	}
	switch {
	case treeF.on:
		c.renderMode = "tree"
		c.treeAll = treeF.all
	case full:
		c.renderMode = "full"
	default:
		c.renderMode = "brief"
	}

	// -from replays a saved chain with no model calls, so it needs neither an API key nor a new
	// chain directory. It renders straight from the JSONL and returns.
	if fromChain != "" {
		c.usage = newUsageCounters()
		c.runFromChain(fromChain, flag.Arg(0))
		return
	}

	// -speakers is a corpus inspection: split the source into passages and print the distinct speakers
	// and the role each was tagged, then exit. No model call, so no key needed.
	if speakers {
		if src == "" {
			fatal("-speakers needs -source CORPUS")
		}
		ix, err := retrieve.Load(src)
		if err != nil {
			fatal("load corpus: " + err.Error())
		}
		printSpeakers(ix)
		return
	}

	// Wire the backend before any model call. Only Anthropic needs a key; Ollama talks to a local
	// server, so requiring ANTHROPIC_API_KEY there would be a false gate.
	switch backendName {
	case "anthropic":
		c.apiKey = os.Getenv("ANTHROPIC_API_KEY")
		if c.apiKey == "" {
			fatal("set ANTHROPIC_API_KEY in your environment first (or pass -backend ollama).")
		}
		c.backend = anthropic.New(c.model, c.apiKey, nil)
	case "ollama":
		c.backend = ollama.New(c.model, ollamaURL, think, nil)
	default:
		fatal("unknown -backend " + backendName + ": use anthropic or ollama")
	}
	c.backendName = backendName
	c.usage = newUsageCounters()
	input := readInput(text)

	// Derive the fixture base name for Tier-2 JSONL naming.
	fixtureName := "stdin"
	if a := flag.Arg(0); a != "" {
		fixtureName = strings.TrimSuffix(filepath.Base(a), filepath.Ext(a))
	}

	// Resolve the chain directory: explicit flag > default eval/<stamp>/.
	if chainDir == "" {
		stamp := time.Now().Format("20060102-1504")
		chainDir = filepath.Join("eval", stamp+"-"+c.model)
	}
	if err := os.MkdirAll(chainDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "warning: cannot create chain-dir %q: %v\n", chainDir, err)
		chainDir = ""
	}

	// Build the mode suffix for the JSONL filename.
	modeSuffix := "substance"
	switch {
	case audit:
		modeSuffix = "audit"
	case ev:
		modeSuffix = "grounding"
	case src != "":
		modeSuffix = "faithfulness"
	}
	if chainDir != "" {
		c.chainFile = filepath.Join(chainDir, fixtureName+"."+modeSuffix+".jsonl")
		c.auditPath = filepath.Join(chainDir, "audit.md")
		c.treeHTMLPath = filepath.Join(chainDir, "tree.html")
	}

	// Build the retrieval index once from the source corpus, so every claim retrieves from the same
	// passage set. On failure (e.g. a source with no speaker turns) fall back to full-corpus rather
	// than aborting — retrieval is an optimisation, not a precondition.
	if c.retrieveMode == "oracle" {
		if oracleFile == "" {
			fatal("-retrieve=oracle needs -oracle FILE (claim-id → passage-id map)")
		}
		m, err := loadOracle(oracleFile)
		if err != nil {
			fatal("load oracle: " + err.Error())
		}
		c.oracle = m
	}
	if c.retrieveMode != "none" && src != "" {
		ix, err := retrieve.Load(src)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: retrieval disabled (%v); using full corpus\n", err)
			c.retrieveMode = "none"
		} else {
			c.index = ix
			// Attach the semantic ranker for bm25 mode only (oracle uses a fixed passage set, so it
			// needs no ranker). An embedder is passed only when -embed is on; with it off, or when the
			// embedder cannot reach ollama, retrieval stays BM25-only rather than aborting.
			if c.embed && c.retrieveMode == "bm25" {
				var emb retrieve.Embedder = embed.New(embedModel, ollamaURL, nil)
				if err := ix.AttachEmbeddings(emb, embedCacheDir(src)); err != nil {
					fmt.Fprintf(os.Stderr, "warning: embeddings disabled (%v); BM25-only retrieval\n", err)
				}
			}
			if c.verbose {
				fmt.Fprintf(os.Stderr, "[retrieve] mode=%s %d passages, max-tokens=%d, embeddings=%v\n",
					c.retrieveMode, len(ix.Passages), c.maxTokens, ix.HasEmbeddings())
			}
		}
	}

	stopHeartbeat := func() {}
	if !c.quiet && c.showProgress {
		stopHeartbeat = startHeartbeat(c.model, c.usage, func() *runTally { return c.tally })
	}

	switch {
	case audit:
		if src == "" {
			fatal("-audit needs -source TRANSCRIPT")
		}
		c.runAudit(input, mustRead(src))
	case ev:
		evSrc := ""
		if src != "" {
			evSrc = mustRead(src)
		}
		c.runEvidence(input, evSrc)
	case src != "":
		c.runFaithfulness(input, src)
	default:
		c.runSubstance(input)
	}

	stopHeartbeat()

	// Print SUMMARY block to stderr (always, even in quiet mode — spec requires it).
	if c.tally != nil {
		printSummary(fixtureName, modeSuffix, c.model, c.tally, c.usage, c.chainFile)
	}

	printUsageLine(c.model, c.usage, c.usageOut)
}

// ── progress helpers ─────────────────────────────────────────────────────────

func (c cfg) progressEnabled() bool { return !c.quiet && c.showProgress }

// progressStart records the start time and sets the heartbeat label. It does NOT print a pre-call
// line — the completion line printed by progressDone covers liveness (60s heartbeat fills the gap).
func (c cfg) progressStart(i, n int, mode string) time.Time {
	if c.usage != nil {
		c.usage.setLabel(fmt.Sprintf("%s %d/%d", mode, i+1, n))
	}
	return time.Now()
}

// progressDone prints one completion line per case to stderr (Tier 1). Always emitted unless quiet.
// ✓ = real verdict; ✗ = error / no verdict.
func (c cfg) progressDone(i, n int, verdict, claim string, start time.Time) {
	if c.tally != nil {
		c.tally.record(verdict)
	}
	if !c.progressEnabled() {
		return
	}
	elapsed := time.Since(start).Round(10 * time.Millisecond)
	sym := "✓"
	if verdict == "error" {
		sym = "✗"
	}
	label := claim
	if len(label) > 60 {
		label = label[:57] + "…"
	}
	fmt.Fprintf(os.Stderr, "[%3d/%d] %s %-13s %q  %s\n", i+1, n, sym, verdict, label, elapsed)
}

// printSummary emits the final SUMMARY block to stderr after a run completes.
func printSummary(fixture, mode, model string, rt *runTally, u *usageCounters, chainFile string) {
	total, verified, counts := rt.snapshot()
	errored := total - verified
	_, in, out, cr, cc, ws, _, elapsed := u.snapshot()
	cost, _ := estimateCost(model, in, out, cr, cc, ws)

	// Build verdict breakdown line (skip "error" — covered by errored count).
	var parts []string
	verdictOrder := []string{"supported", "mixed", "refuted", "unverifiable",
		"faithful", "partial", "overstated", "absent", "contradicted",
		"substantive", "hollow", "skipped (over cap)"}
	for _, v := range verdictOrder {
		if n := counts[v]; n > 0 {
			parts = append(parts, fmt.Sprintf("%s %d", v, n))
		}
	}
	// Any verdict not in the ordered list still gets printed.
	inOrder := make(map[string]bool)
	for _, v := range verdictOrder {
		inOrder[v] = true
	}
	for v, n := range counts {
		if !inOrder[v] && v != "error" {
			parts = append(parts, fmt.Sprintf("%s %d", v, n))
		}
	}
	if counts["error"] > 0 {
		parts = append(parts, fmt.Sprintf("error %d", counts["error"]))
	}

	schemaRetries, quoteRejects := u.extras()

	fmt.Fprintf(os.Stderr, "\nSUMMARY %s %s\n", fixture, mode)
	fmt.Fprintf(os.Stderr, "  cases %d · verified %d · errored %d\n", total, verified, errored)
	fmt.Fprintf(os.Stderr, "  %s\n", strings.Join(parts, " · "))
	fmt.Fprintf(os.Stderr, "  wall %s · est_usd %s · cache_hit %.0f%% (read %d / created %d)\n",
		elapsed.Round(time.Second), cost, 100*u.cacheHitRate(), cr, cc)
	fmt.Fprintf(os.Stderr, "  schema_retries %d · quote_rejects %d\n", schemaRetries, quoteRejects)
	if chainFile != "" {
		fmt.Fprintf(os.Stderr, "  detail: %s\n", chainFile)
	}
}

// ── runners ──────────────────────────────────────────────────────────────────

func (c *cfg) runSubstance(input string) {
	claims, err := c.decompose(input)
	if err != nil {
		fatal("decompose failed: " + err.Error())
	}
	c.tally = newRunTally(len(claims))
	results := make([]substance, len(claims))
	rows := make([]brief.Row, len(claims))
	vs := make([]string, len(claims))
	details := make(map[string]tree.Leaf, len(claims))
	for i, cl := range claims {
		if c.verbose {
			fmt.Println(c.bold(fmt.Sprintf("▸ claim %d/%d: %s", i+1, len(claims), cl)))
		}
		t := c.progressStart(i, len(claims), "substance")
		results[i] = c.assayClaim(cl)
		c.appendChain(substanceChainRecord(i, len(claims), cl, results[i], t))
		c.progressDone(i, len(claims), results[i].Verdict, cl, t)
		reason := results[i].Reason
		if reason == "" && results[i].SurvivingClaim != "" {
			reason = "survives as: " + results[i].SurvivingClaim
		}
		id := fmt.Sprintf("c%d", i+1)
		rows[i] = brief.Row{ID: id, Text: cl,
			Substance: results[i].Verdict, SubstanceReason: reason}
		details[id] = tree.Leaf{Reason: reason}
		vs[i] = results[i].Verdict
	}
	c.present(rows, tally(vs), mdSubstance(results), func() { c.termSubstance(results) }, details)
}

// claimLine is one parsed claim from a claims file: its id, its §-heading path, and its text.
type claimLine struct{ id, path, text string }

func (c *cfg) runFaithfulness(input, srcPath string) {
	raw := splitSummary(input)
	c.tally = newRunTally(len(raw))
	results := make([]faith, len(raw))
	rows := make([]brief.Row, len(raw))
	vs := make([]string, len(raw))
	details := make(map[string]tree.Leaf, len(raw))

	// Retrieve every claim's passages first, then run claims sharing passages consecutively over one
	// shared (union) prefix, so the judge's cached prefix stays hot across a group (2c). A claim the
	// retrieval floor suppressed gets no group and no model call.
	parsed := make([]claimLine, len(raw))
	claimPassages := make([][]retrieve.Passage, len(raw))
	below := make([]bool, len(raw))
	var fullSrc []retrieve.Passage
	var eligible []int
	for i, item := range raw {
		id, path, text := parseClaimLine(item)
		if id == "" {
			id = fmt.Sprintf("c%d", i+1)
		}
		parsed[i] = claimLine{id: id, path: path, text: text}
		claimPassages[i], below[i] = c.passagesForClaim(id, text, path, srcPath, &fullSrc)
		if !below[i] {
			eligible = append(eligible, i)
		}
	}

	emit := func(i int, chosen faith, spread string, pids []string, t time.Time) {
		results[i] = chosen
		rec := faithChainRecord(i, len(raw), parsed[i].text, chosen, t)
		rec.Spread = spread
		rec.Passages = pids
		c.appendChain(rec)
		c.progressDone(i, len(raw), chosen.Verdict, parsed[i].text, t)
		rows[i] = brief.Row{ID: parsed[i].id, Path: parsed[i].path, Text: parsed[i].text,
			Faith: chosen.Verdict, FaithReason: chosen.Evidence, Spread: spread,
			Gap: chosen.Gap, SoWhat: chosen.SoWhat}
		details[parsed[i].id] = tree.Leaf{Reason: chosen.Evidence, Quotes: chosen.Quotes}
		vs[i] = chosen.Verdict
	}

	// Floor-suppressed claims: "absent" from code, no model call.
	for i, isBelow := range below {
		if isBelow {
			emit(i, faith{Verdict: "absent", Evidence: "no passage cleared the retrieval floor (retrieved=0)"},
				"floor", nil, time.Now())
		}
	}

	// Grouped judge calls: within a group every claim sees the same union prefix, so calls after the
	// first read the cache instead of re-billing the passages.
	for _, group := range groupBySharedPassages(claimPassages, eligible) {
		shared := unionPassages(claimPassages, group)
		pids := retrieve.IDs(shared)
		for _, i := range group {
			t := c.progressStart(i, len(raw), "faithfulness")
			chosen, spread := c.faithJudgeRepeat(parsed[i].text, shared)
			emit(i, chosen, spread, pids, t)
		}
	}
	c.present(rows, tally(vs), mdFaith(results), func() { c.termFaith(results) }, details)
}

// groupBySharedPassages partitions eligible claim indices into groups whose passage sets overlap by
// at least half (|A∩B| ≥ ½·min(|A|,|B|)), so a group's claims can share one cached prefix. A greedy
// seed-and-attach pass against each seed; claims that match nothing form singleton groups. Order is
// stable (seeds and members in ascending index) so a run is deterministic.
func groupBySharedPassages(claimPassages [][]retrieve.Passage, eligible []int) [][]int {
	sets := make(map[int]map[string]bool, len(eligible))
	for _, i := range eligible {
		s := make(map[string]bool, len(claimPassages[i]))
		for _, p := range claimPassages[i] {
			s[p.ID] = true
		}
		sets[i] = s
	}
	used := make(map[int]bool, len(eligible))
	var groups [][]int
	for _, seed := range eligible {
		if used[seed] {
			continue
		}
		group := []int{seed}
		used[seed] = true
		for _, other := range eligible {
			if used[other] {
				continue
			}
			if sharedAtLeastHalf(sets[seed], sets[other]) {
				group = append(group, other)
				used[other] = true
			}
		}
		groups = append(groups, group)
	}
	return groups
}

// sharedAtLeastHalf reports whether a and b overlap by at least half the smaller set.
func sharedAtLeastHalf(a, b map[string]bool) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	inter := 0
	for id := range a {
		if b[id] {
			inter++
		}
	}
	smaller := min(len(a), len(b))
	return inter*2 >= smaller
}

// unionPassages returns the deduplicated passages across a group's claims, in first-seen order, so
// every claim in the group is judged against one identical prefix.
func unionPassages(claimPassages [][]retrieve.Passage, group []int) []retrieve.Passage {
	seen := map[string]bool{}
	var out []retrieve.Passage
	for _, i := range group {
		for _, p := range claimPassages[i] {
			if !seen[p.ID] {
				seen[p.ID] = true
				out = append(out, p)
			}
		}
	}
	return out
}

// retrieveTokenCap bounds the passages sent to the judge — enough that the answer is present but
// small enough to keep the judge focused. Raised to 10000 when each witness passage grew to carry its
// preceding question as context: the larger passages need a proportionally larger budget, and 10000 is
// the knee of the coverage curve (~25 passages/claim), where the refuter reaches its 15/18 ceiling.
const retrieveTokenCap = 10000

// passagesForClaim returns the passages the faithfulness judge sees for one claim, and whether the
// retrieval floor suppressed all of them. With retrieval off it is the whole corpus as a single
// "corpus" passage (loaded once into *fullSrc); with retrieval on it is the fused bm25+embed top
// passages up to the token budget for the claim text plus its §-heading hint. When `below` is true
// nothing cleared -floor and the caller records "absent" from code without a model call.
func (c cfg) passagesForClaim(id, text, path, srcPath string, fullSrc *[]retrieve.Passage) (ps []retrieve.Passage, below bool) {
	if c.retrieveMode == "oracle" {
		// Oracle: the judge sees exactly the gold passages mapped to this claim id (no ranking).
		return c.index.ByIDs(c.oracle[id]), false
	}
	if c.index == nil {
		if len(*fullSrc) == 0 {
			*fullSrc = []retrieve.Passage{{ID: "corpus", Text: readCorpus(srcPath)}}
		}
		return *fullSrc, false
	}
	q := claimQuery(text, path)
	res := c.index.Retrieve(q, c.maxTokens, c.floor)
	if res.Below {
		return nil, true
	}
	return res.Passages, false
}

// loadOracle reads the -retrieve=oracle map: a JSON object of claim-id → list of passage-id.
func loadOracle(path string) (map[string][]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m map[string][]string
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	if len(m) == 0 {
		return nil, fmt.Errorf("oracle file %s has no entries", path)
	}
	return m, nil
}

// claimQuery builds the retrieval query for a claim: its text, the §-heading labels as a hint, and —
// when the claim explicitly attributes itself to a named person ("According to Jane Doe", "Jane Doe
// of X", "Jane Doe said") — that person's surname as a hard filter, so their turns rank first.
func claimQuery(text, path string) retrieve.Query {
	return retrieve.Query{Text: text, Hint: hintFromPath(path), Witness: namedWitness(text)}
}

// witnessRe pulls an attributed speaker from a claim: a capitalised "First Last" (optionally "First
// Middle Last") introduced by "according to", or followed by "of"/"said"/"told"/"argued"/"noted".
// Returns "" when the claim makes no attribution — the common case, where retrieval falls back to the
// fused ranking with no hard filter.
var witnessRe = regexp.MustCompile(`(?:According to|according to) ([A-Z][a-z]+(?: [A-Z][a-z]+){1,2})|([A-Z][a-z]+(?: [A-Z][a-z]+){1,2}) (?:of|said|told|argued|noted|stated)`)

func namedWitness(text string) string {
	m := witnessRe.FindStringSubmatch(text)
	if m == nil {
		return ""
	}
	name := m[1]
	if name == "" {
		name = m[2]
	}
	fields := strings.Fields(name)
	if len(fields) == 0 {
		return ""
	}
	return fields[len(fields)-1] // surname
}

// embedCacheDir is where a corpus's embedding caches live: a .embcache directory beside the corpus,
// so the committed vectors travel with the transcripts they were computed from. For a single file it
// sits in that file's directory.
func embedCacheDir(src string) string {
	if info, err := os.Stat(src); err == nil && info.IsDir() {
		return filepath.Join(src, ".embcache")
	}
	return filepath.Join(filepath.Dir(src), ".embcache")
}

// hintFromPath turns a claim's "key=Label/key=Label" §-path into a plain-text query hint (the Labels),
// so retrieval is steered by the section a claim sits under as well as its own words.
func hintFromPath(path string) string {
	if path == "" {
		return ""
	}
	var labels []string
	for _, seg := range strings.Split(path, "/") {
		if i := strings.Index(seg, "="); i >= 0 {
			labels = append(labels, seg[i+1:])
		} else {
			labels = append(labels, seg)
		}
	}
	return strings.Join(labels, " ")
}

// readCorpus reads a full-corpus source for retrieval-off mode: a single file, or every .txt under a
// directory concatenated in sorted order (deterministic).
func readCorpus(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		fatal("read source " + path + ": " + err.Error())
	}
	if !info.IsDir() {
		return mustRead(path)
	}
	var files []string
	filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".txt") {
			files = append(files, p)
		}
		return nil
	})
	sort.Strings(files)
	var b strings.Builder
	for _, f := range files {
		b.WriteString(mustRead(f))
		b.WriteString("\n\n")
	}
	return b.String()
}

// printSpeakers prints the distinct speakers found in a corpus with their assigned role, questioners
// and chair before witnesses — the -speakers output.
func printSpeakers(ix *retrieve.Index) {
	sp := ix.Speakers()
	fmt.Printf("%d passages · %d distinct speakers\n", len(ix.Passages), len(sp))
	for _, s := range sp {
		fmt.Printf("  %-10s %4d turns  %s\n", s.Role, s.Turns, s.Speaker)
	}
}

// intendedProposition returns what_source_actually_says when the faithfulness pass flagged the
// claim as partial or overstated and populated that field.
// Otherwise it returns the original claim unchanged. Using this in both runEvidence and the audit
// grounding pass keeps the two paths in sync.
func intendedProposition(fc faith, claim string) string {
	if fc.SourceSays != "" && (fc.Verdict == "partial" || fc.Verdict == "overstated") {
		return fc.SourceSays
	}
	return claim
}

// runEvidence grounds each claim via web search. If a source transcript is provided, it first runs
// the faithfulness pass to reconstruct the intended proposition, so we ground what the speaker
// actually meant, not a literalized paraphrase. The audit grounding pass shares this same logic via
// intendedProposition, so the audit's grounding column agrees with -evidence -source.
func (c *cfg) runEvidence(input, src string) {
	claims := splitSummary(input)
	c.tally = newRunTally(len(claims))
	results := make([]evidence, len(claims))
	for i, cl := range claims {
		// Skip evidence grounding if over the cap.
		if c.maxClaims > 0 && i >= c.maxClaims {
			results[i] = evidence{Claim: cl, Verdict: "skipped (over cap)", Finding: ""}
			c.tally.record("skipped (over cap)")
			continue
		}
		t := c.progressStart(i, len(claims), "evidence")
		proposition := cl
		if src != "" {
			// Reconstruct the asserted proposition via the faithfulness pass so we ground what the
			// speaker actually meant, not a literalized paraphrase.
			fc := c.faithClaim(cl, src)
			proposition = intendedProposition(fc, cl)
		}
		r := c.evidenceClaim(proposition)
		r.Claim = cl // keep original label for display
		results[i] = r
		c.appendChain(evidenceChainRecord(i, len(claims), cl, results[i], t))
		c.progressDone(i, len(claims), results[i].Verdict, cl, t)
	}
	if c.asMarkdown {
		fmt.Print(mdEvidence(results))
	} else {
		c.termEvidence(results)
	}
}

// computeAudit runs faithfulness, substance, and evidence for each claim and returns the three
// result slices. The evidence pass grounds the INTENDED proposition (via intendedProposition)
// rather than the literal summary claim, so the audit grounding column agrees with
// -evidence -source.
func (c cfg) computeAudit(claims []string, src string) ([]faith, []substance, []evidence) {
	fs := make([]faith, len(claims))
	ss := make([]substance, len(claims))
	es := make([]evidence, len(claims))
	for i, cl := range claims {
		fs[i] = c.faithClaim(cl, src)
		ss[i] = c.assayClaim(cl)
		// Skip evidence grounding if over the cap.
		if c.maxClaims > 0 && i >= c.maxClaims {
			es[i] = evidence{Claim: cl, Verdict: "skipped (over cap)", Finding: ""}
		} else {
			proposition := intendedProposition(fs[i], cl)
			r := c.evidenceClaim(proposition)
			r.Claim = cl // keep original label for display
			es[i] = r
		}
	}
	return fs, ss, es
}

// runAudit performs a full audit by running faithfulness, substance, and evidence passes in
// sequence, then cross-tabulating the results. It shares the same decomposition as runSubstance,
// but skips the producer-critic loop and just runs one substance pass per claim, since the audit is
// more about cross-filter patterns than squeezing out every last drop of rigor from each claim.
func (c *cfg) runAudit(input, src string) {
	claims := splitSummary(input)
	n := len(claims)
	c.tally = newRunTally(n)
	fs := make([]faith, n)
	ss := make([]substance, n)
	es := make([]evidence, n)
	for i, cl := range claims {
		t := c.progressStart(i, n, "audit")
		fs[i] = c.faithClaim(cl, src)
		ss[i] = c.assayClaim(cl)
		if c.maxClaims > 0 && i >= c.maxClaims {
			es[i] = evidence{Claim: cl, Verdict: "skipped (over cap)"}
		} else {
			proposition := intendedProposition(fs[i], cl)
			r := c.evidenceClaim(proposition)
			r.Claim = cl
			es[i] = r
		}
		verdictSummary := fmt.Sprintf("faith=%s sub=%s ev=%s",
			fs[i].Verdict, ss[i].Verdict, es[i].Verdict)
		c.appendChain(auditChainRecord(i, n, cl, fs[i], ss[i], es[i], t))
		c.progressDone(i, n, verdictSummary, cl, t)
	}
	fmt.Print(mdAudit(claims, fs, ss, es)) // audit is always markdown
}

// ── stages ───────────────────────────────────────────────────────────────────

// decompose is the first stage: break an input blob into independently checkable claims. The rest
// of the stages operate on these individual claims.
func (c cfg) decompose(text string) ([]string, error) {
	var arr []string
	if err := c.callJSON(decomposeSys, "", "TEXT:\n"+text, false, &arr); err != nil {
		return nil, err
	}
	return arr, nil
}

// assayClaim is the substance pass: given a claim, run it through producer-critic rounds to see if
// it can be substantiated.
func (c cfg) assayClaim(claim string) substance {
	current, rounds, steel := claim, 0, ""
	var last substanceJSON
	for rounds < c.maxRounds {
		var p producerJSON
		if err := c.callJSON(producerSys, "", "CLAIM:\n"+current, false, &p); err != nil {
			return substance{Claim: claim, Verdict: "error", Reason: err.Error()}
		}
		if rounds == 0 {
			steel = p.Steelman
		}
		u := "CLAIM:\n" + current + "\n\nPRODUCER STEELMAN:\n" + p.Steelman + "\n\nPRODUCER CONDITIONS:\n" + p.Conditions
		if err := c.callJSON(substanceCriticSys, "", u, false, &last); err != nil {
			return substance{Claim: claim, Verdict: "error", Reason: err.Error()}
		}
		rounds++
		// Only continue to the next round when the critic wants one AND the
		// claim didn't survive purely through condition laundering.
		if last.NeedsAnother && last.SurvivingClaim != "" &&
			rounds < c.maxRounds &&
			!last.SurvivesOnlyByConditions {
			current = last.SurvivingClaim
			continue
		}
		break
	}
	// Downgrade at loop exit if the final round survived only by laundered conditions. A claim that
	// needs invented qualifiers to survive is not partial.
	if last.SurvivesOnlyByConditions {
		last.Verdict = "hollow"
		last.Reason = last.Reason + " Survives only by conditions the speaker never stated."
	}
	out := substance{Claim: claim, Steelman: steel, Verdict: last.Verdict,
		SurvivingClaim: last.SurvivingClaim, Reason: last.Reason, Rounds: rounds}
	for _, ci := range last.Critique {
		out.Critique = append(out.Critique, critiqueItem{ci.Axis, ci.Finding, ci.Severity})
	}
	return out
}

// faithClaim is the faithfulness pass: given a claim and a source transcript, check whether the
// claim faithfully represents what the speaker said in the source. This is a two-step process:
// first the defender identifies supporting evidence in the source, then the critic evaluates the
// faithfulness of the claim based on this evidence.
func (c cfg) faithClaim(claim, src string) faith {
	// The source transcript is identical across every claim in a run, so it rides in the cached
	// prefix (ahead of the claim) rather than being re-sent inline. The two passes use different
	// system prompts, so each caches its own (system + source) prefix once and reads it thereafter.
	cached := "SOURCE:\n" + src
	var d defenderJSON
	if err := c.callJSON(faithDefenderSys, cached, "SUMMARY CLAIM:\n"+claim, false, &d); err != nil {
		return faith{Claim: claim, Verdict: "error"}
	}
	quotes := "(none)"
	if len(d.Quotes) > 0 {
		quotes = "- " + strings.Join(d.Quotes, "\n- ")
	}
	u := fmt.Sprintf("SUMMARY CLAIM:\n%s\n\nDEFENDER FOUND SUPPORT: %v\nDEFENDER QUOTES:\n%s",
		claim, d.Found, quotes)
	var fj faithJSON
	if err := c.callJSON(faithCriticSys, cached, u, false, &fj); err != nil {
		return faith{Claim: claim, Verdict: "error"}
	}
	return faith{Claim: claim, Verdict: fj.Verdict, Evidence: fj.Evidence,
		SourceSays: fj.SourceSays, Gap: fj.Gap, SoWhat: fj.SoWhat, Quotes: d.Quotes}
}

// judgeSampleTemp is the sampling temperature for repeat (-n>1) judge calls, so the N samples can
// disagree and the spread is meaningful; a single call runs at 0 (deterministic). Both backends
// accept it (sonnet-4-6 permits temperature; ollama sets it in options).
const judgeSampleTemp = 0.7

// faithJudge is the single schema-enforced faithfulness pass over retrieved passages: one model call
// returns the verdict and cites its evidence by {passage_id, quote}, then this code verifies each
// quote is a verbatim substring of the passage it names — the grounding check, no second model call.
// A quote that fails is dropped and counted; the passages are the cached prefix so a run that groups
// claims sharing passages reuses it. temp is nil for a deterministic call, non-nil for repeat sampling.
func (c cfg) faithJudge(claim string, passages []retrieve.Passage, temp *float64) faith {
	cached := "PASSAGES:\n" + retrieve.Format(passages)
	var j judgeJSON
	if err := c.callSchema(faithJudgeSys, cached, "SUMMARY CLAIM:\n"+claim, json.RawMessage(judgeSchema), "faith_verdict", temp, &j); err != nil {
		return faith{Claim: claim, Verdict: "error", Evidence: err.Error()}
	}
	byID := make(map[string]retrieve.Passage, len(passages))
	for _, p := range passages {
		byID[p.ID] = p
	}
	var verified []string
	rejects := 0
	for _, e := range j.Evidence {
		p, ok := byID[e.PassageID]
		if ok && quoteInPassage(e.Quote, p) {
			verified = append(verified, e.Quote)
		} else {
			rejects++ // a paraphrase presented as verbatim, or an id the judge invented
		}
	}
	if rejects > 0 && c.usage != nil {
		c.usage.addQuoteReject(rejects)
	}
	return faith{Claim: claim, Verdict: j.Verdict, Evidence: j.Reason,
		SourceSays: j.SourceSays, Gap: j.Gap, SoWhat: j.SoWhat, Quotes: verified}
}

// faithJudgeRepeat runs faithJudge c.repeat times and returns the modal result plus a "k/N" agreement
// string ("" when N=1). N=1 is deterministic (temperature 0); N>1 samples at judgeSampleTemp so the
// spread reflects real judge instability, with the cached passage prefix reused across the repeats.
func (c cfg) faithJudgeRepeat(claim string, passages []retrieve.Passage) (faith, string) {
	n := c.repeat
	if n < 1 {
		n = 1
	}
	if n == 1 {
		zero := 0.0
		return c.faithJudge(claim, passages, &zero), ""
	}
	t := judgeSampleTemp
	counts := make(map[string]int, n)
	rep := make(map[string]faith, n)
	for k := 0; k < n; k++ {
		r := c.faithJudge(claim, passages, &t)
		counts[r.Verdict]++
		if _, seen := rep[r.Verdict]; !seen {
			rep[r.Verdict] = r
		}
	}
	modal := modalVerdict(counts)
	return rep[modal], fmt.Sprintf("%d/%d", counts[modal], n)
}

// quoteInPassage reports whether quote appears verbatim in the passage the judge cited — checked over
// the passage's full visible text (question context + answer), normalised for whitespace and curly
// punctuation only. This is the grounding check: a paraphrase ("arts sector" for the source's
// "creative sector") is not a substring and is rejected. An empty quote never verifies.
func quoteInPassage(quote string, p retrieve.Passage) bool {
	q := normQuote(quote)
	if q == "" {
		return false
	}
	visible := p.Text
	if p.Context != "" {
		visible = p.Context + "\n\n" + p.Text
	}
	return strings.Contains(normQuote(visible), q)
}

var quoteWS = regexp.MustCompile(`\s+`)
var quotePunct = strings.NewReplacer("’", "'", "‘", "'", "“", `"`, "”", `"`, "—", "-", "–", "-")

// normQuote lowercases, folds curly punctuation/dashes to ASCII, and collapses whitespace, so a quote
// matches its passage despite a line wrap or a curly apostrophe. It does NOT strip words, so a changed
// word still fails the substring test — that is the point of the grounding check.
func normQuote(s string) string {
	return quoteWS.ReplaceAllString(quotePunct.Replace(strings.ToLower(strings.TrimSpace(s))), " ")
}

// faithRepeat runs faithClaim c.repeat times (once when repeat ≤ 1) and returns the modal result
// plus a "k/N" agreement string, "" when N=1. Repeat sampling measures how stable the critic's
// verdict is on one claim — a claim that comes back "partial 2/3" is one the human should not read
// as settled. The returned faith is a run that actually produced the modal verdict, so its reason
// and quotes match the verdict shown.
func (c cfg) faithRepeat(claim, src string) (faith, string) {
	n := c.repeat
	if n < 1 {
		n = 1
	}
	if n == 1 {
		return c.faithClaim(claim, src), ""
	}
	counts := make(map[string]int, n)
	rep := make(map[string]faith, n)
	for k := 0; k < n; k++ {
		r := c.faithClaim(claim, src)
		counts[r.Verdict]++
		if _, seen := rep[r.Verdict]; !seen {
			rep[r.Verdict] = r
		}
	}
	modal := modalVerdict(counts)
	return rep[modal], fmt.Sprintf("%d/%d", counts[modal], n)
}

// faithVerdictOrder ranks faithfulness verdicts worst-first; modalVerdict breaks a count tie by it,
// so a split surfaces the verdict a human is likelier to need to look at rather than a random one.
var faithVerdictOrder = []string{"contradicted", "absent", "overstated", "partial", "faithful", "error"}

// modalVerdict returns the most frequent verdict in `counts`, breaking ties by faithVerdictOrder
// (worst first). It is total over any non-empty map.
func modalVerdict(counts map[string]int) string {
	rank := make(map[string]int, len(faithVerdictOrder))
	for i, v := range faithVerdictOrder {
		rank[v] = i
	}
	best, bestN, bestRank := "", -1, 1<<31
	for v, n := range counts {
		r, ok := rank[v]
		if !ok {
			r = len(faithVerdictOrder) // unknown verdicts sort last
		}
		if n > bestN || (n == bestN && r < bestRank) {
			best, bestN, bestRank = v, n, r
		}
	}
	return best
}

// evidenceClaim is the grounding pass: check the claim against current evidence via web search. If
// a source transcript is provided, it first runs the faithfulness pass to reconstruct the intended
// proposition, so we ground what the speaker actually meant, not a literalized paraphrase. The
// audit grounding pass shares this same logic via intendedProposition, so the audit's grounding
// column agrees with -evidence -source.
func (c cfg) evidenceClaim(claim string) evidence {
	var e evidenceJSON
	rs, err := c.callJSONSourced(evidenceSys, "", "CLAIM:\n"+claim, true, &e)
	if err != nil {
		return evidence{Claim: claim, Verdict: "error", Finding: err.Error()}
	}
	out := evidence{
		Claim: claim, Verdict: e.Verdict, Finding: e.Finding,
		RetrievedSources: rs,
	}
	for _, s := range e.Sources {
		out.Sources = append(out.Sources, source{s.Title, s.URL})
	}
	return crossCheckEvidence(out)
}

// crossCheckEvidence downgrades a verdict to "unverifiable" when no model-claimed source URL
// matches the set of URLs actually retrieved during the web_search_tool_result round trip.
// This closes the grounding-integrity gap: the axis boundary requires a real truth-maker for
// supported/mixed/refuted — positive grounding cannot be confirmed from parametric knowledge alone.
// Only "unverifiable" and "error" are exempt (they make no external-evidence assertion).
func crossCheckEvidence(e evidence) evidence {
	if e.Verdict == "unverifiable" || e.Verdict == "error" || e.DowngradeReason != "" {
		return e
	}
	if len(e.RetrievedSources) == 0 {
		e.OriginalVerdict = e.Verdict
		e.Verdict = "unverifiable"
		e.DowngradeReason = "no sources retrieved; verdict is model self-report"
		return e
	}
	if len(e.Sources) == 0 {
		e.OriginalVerdict = e.Verdict
		e.Verdict = "unverifiable"
		e.DowngradeReason = "no URLs cited in response"
		return e
	}
	retrieved := make(map[string]bool, len(e.RetrievedSources))
	for _, r := range e.RetrievedSources {
		retrieved[normalizeURL(r.URL)] = true
	}
	matched := 0
	for _, s := range e.Sources {
		if retrieved[normalizeURL(s.URL)] {
			matched++
		}
	}
	e.SourcesVerified = matched
	if matched == 0 {
		e.OriginalVerdict = e.Verdict
		e.Verdict = "unverifiable"
		e.DowngradeReason = "claimed sources not present in retrieval"
	}
	return e
}

// normalizeURL returns a canonical host+path string for URL comparison.
// Strips scheme, query, and fragment; lowercases host; trims trailing slash from path.
// Matching on host+path catches "cited a real article it actually read" without allowing
// "cited some other page on the same domain" to pass.
func normalizeURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return strings.ToLower(rawURL)
	}
	return strings.ToLower(u.Host) + strings.TrimRight(u.Path, "/")
}

// ── prompts (single source of truth — carry any prompt fixes here) ───────────

const decomposeSys = `You are a claims extractor trained in analytic philosophy. Break prose into
its atomic, independently-evaluable assertions. Strip rhetoric, hedges, and connective filler. Each
item must be a single claim that could in principle be true or false. Do not evaluate them. Return
ONLY a JSON array of strings, no markdown, no preamble.`

const producerSys = `You are the Producer. Given a single claim, construct its STRONGEST defensible
version (steelman) and state precisely what would have to be true for it to hold. Be concrete. You
are NOT evaluating or criticising the claim — only making the best honest case for it. Return ONLY
JSON: {"steelman": string, "conditions": string}. No markdown.`

const substanceCriticSys = `You are the Critic. Assess one claim against these FIXED AXES, in order:
- Evidence: is support cited or available, or is it bare assertion?
- Hidden premise: what unstated assumption must hold?
- Falsifiability: what observation would show it false? If none exists, it is vacuous.
- Equivocation: does a key term shift meaning or hide behind a buzzword?
- Base rate / magnitude: is there a real quantity and a comparison, or just a direction?
- Counterexample: is there an obvious case where it fails?
- Causality vs correlation: does it assert cause from mere association?

For each axis that bears on the claim, give a one-sentence finding and a severity: "fatal",
"weakens", or "clears".

Then a verdict:
- "hollow": unfalsifiable, equivocating, or pure assertion with no defensible core.
- "partial": a narrower, qualified claim survives after stripping the unsupported parts.
- "substantive": falsifiable, evidence exists or is clearly obtainable, no equivocation, survives
counterexample.

If "partial" or "substantive", give surviving_claim. If "hollow", surviving_claim is null. Set
needs_another_round=true ONLY if a narrower surviving_claim was produced that itself deserves a
fresh pass.

CONDITION DISCIPLINE — distinguish legitimate narrowing from laundering:
- Legitimate narrowing: you restrict the claim to conditions the speaker stated or clearly implied
  (e.g. "in large enterprises", "by 2027").
- Condition laundering: you introduce qualifiers the speaker never stated, solely to dodge a
  counterexample or fill a gap in evidence ("assuming perfect market conditions", "for sufficiently
  motivated users"). This is loss of content, not rigor — it manufactures defensibility rather than
  finding it. When a surviving claim survives ONLY through laundered conditions, the Falsifiability
  axis must fire (the claim is now untestable as stated) and the verdict should lean "hollow", not
  be rewarded as "partial" or "substantive".

Count the number of conditions introduced in surviving_claim that were NOT present in the original
claim or speaker's words. Set survives_only_by_conditioning=true if the claim would revert to
"hollow" without those added conditions.

Return ONLY JSON:
{"critique":[{"axis":string,"finding":string,"severity":"fatal"|"weakens"|"clears"}],"verdict":"substantive"|"partial"|"hollow","surviving_claim":string|null,"reason":string,"needs_another_round":boolean,"added_conditions":integer,"survives_only_by_conditioning":boolean}`

const faithDefenderSys = `You are the Defender. You are given a SUMMARY CLAIM and a SOURCE
transcript. Find the STRONGEST evidence in the SOURCE that the speaker actually asserts this claim.
Quote spans VERBATIM from the SOURCE only — never paraphrase, never use outside knowledge. If there
is no support, set found=false and quotes=[]. Return ONLY JSON:
{"found": boolean, "quotes": [string], "best_case": string}. No markdown.`

const faithCriticSys = `You are the Faithfulness Critic. Decide whether the SUMMARY CLAIM faithfully
 represents what the speaker said in the SOURCE, using the Defender's cited quotes. You are checking
 sense-preservation, not truth: your only question is whether the summary reports the speaker
 accurately, never whether the speaker was right. Watch for the ways a summary distorts a source:
- Fabrication: the claim is simply not in the source.
- Overstatement: the source hedged or qualified it; the summary made it absolute.
- Distortion: the meaning was changed.
- Context-stripping: a conditional or hypothetical presented as an unconditional belief.
- Misattribution: the speaker was quoting or steelmanning someone else, and the summary attributes
  it as the speaker's own view.
- Cherry-pick: present but unrepresentative of the source's stance.
- Literalization: the summary states as a sincere literal assertion something the speaker meant as
  provocation, hyperbole, or irony. The words may appear in the source but the asserted proposition
  does not — the speaker's force or register was rhetorical, not declarative. When this occurs the
  verdict is NOT "faithful"; use "overstated", and ALWAYS populate what_source_actually_says with
  the proposition the speaker actually asserted (i.e. what they meant, not what they said literally).

Verify the Defender's quotes actually appear in the source; do not take the Defender's word for it.

Verdict:
- "faithful": the summary reports the claim as the speaker stated it, including the register and
  force with which they stated it.
- "partial": the source supports a weaker/narrower version.
- "overstated": same in kind but the summary strengthened it — or literalized a rhetorical/ironic
  claim (see Literalization above).
- "absent": not in the source.
- "contradicted": the source says the opposite.

For "partial"/"overstated", give what_source_actually_says (the faithful version, including the
correct register).

GAP — classify why the summary overreaches, as the single field that most changes what a reader
would do. This decides whether a "partial" claim reaches the reader or stays buried, so err toward
naming a gap rather than "none":
- "scope": true only of a narrower population, place, or category than the summary implies.
- "denominator": true only of a specific fraction, share, or subtotal — the summary drops the base
  it is a fraction of (e.g. a per-capita share of one programme that is itself 3% of total funding,
  reported as overall fairness).
- "timerange": true only within a limited period the summary drops.
- "attribution": the support is about a different actor, programme, or body than the one named.
- "other": a real gap that is none of the above.
- "none": the narrowing is benign — no reader would act differently knowing it. Use "none" for
  "faithful", and for verdicts other than "partial" unless a gap genuinely applies.

SO WHAT — if a reader who believed the summary would act on a false impression, state in <=20 words
what they would get wrong. Empty string when the summary is faithful or the gap is benign.

Return ONLY JSON:
{"findings":[{"mode":string,"finding":string}],"verdict":"faithful"|"partial"|"overstated"|"absent"|"contradicted","evidence":string,"what_source_actually_says":string|null,"gap":"none"|"scope"|"denominator"|"timerange"|"attribution"|"other","so_what":string}`

// judgeSchema constrains the faithfulness judge's answer. Field order matches the prompt: verdict,
// gap, evidence[], so_what, reason, then what_source_actually_says. All are required (Anthropic strict
// tools and the Ollama format both accept "" for the free-text fields), and evidence cites the source
// by passage id and a verbatim quote, so code can verify the quote without a second model call.
const judgeSchema = `{
  "type":"object",
  "properties":{
    "verdict":{"type":"string","enum":["faithful","partial","overstated","absent","contradicted"]},
    "gap":{"type":"string","enum":["none","scope","denominator","timerange","attribution","other"]},
    "evidence":{"type":"array","items":{"type":"object","properties":{"passage_id":{"type":"string"},"quote":{"type":"string"}},"required":["passage_id","quote"],"additionalProperties":false}},
    "so_what":{"type":"string"},
    "reason":{"type":"string"},
    "what_source_actually_says":{"type":"string"}
  },
  "required":["verdict","gap","evidence","so_what","reason","what_source_actually_says"],
  "additionalProperties":false
}`

// faithJudgeSys is the single schema-enforced faithfulness judge, replacing the defender+critic pair
// for retrieval-fed runs: one call decides the verdict AND cites its evidence by passage id, so the
// quotes are verified in code (a substring check, no second model call) rather than trusted.
const faithJudgeSys = `You are the Faithfulness Judge. You are given PASSAGES from a hearing
transcript — each headed by an id line "[<id> · <date> · <speaker> (<role>)]", and a witness passage
shows the question that prompted it ("Q — …") above the answer ("A — …"). You are also given a
SUMMARY CLAIM. Decide whether the SUMMARY CLAIM faithfully represents what a witness actually said in
the PASSAGES. You check sense-preservation, not truth: your only question is whether the summary
reports the speaker accurately, never whether the speaker was right.

Watch for the ways a summary distorts a source:
- Fabrication: the claim is simply not in the passages.
- Overstatement: the source hedged or qualified it; the summary made it absolute.
- Distortion: the meaning was changed.
- Context-stripping: a conditional or hypothetical presented as an unconditional belief.
- Misattribution: the words are a questioner's (role questioner/chair), or the speaker was quoting or
  steelmanning someone else, and the summary attributes it as a witness's own view.
- Cherry-pick: present but unrepresentative of the source's stance.
- Literalization: the summary states as a sincere literal assertion something the speaker meant as
  provocation, hyperbole, or irony. The words may appear in the source but the asserted proposition
  does not — the speaker's force or register was rhetorical, not declarative. When this occurs the
  verdict is NOT "faithful"; use "overstated", and ALWAYS populate what_source_actually_says.

EVIDENCE: cite the passages that decide the verdict. Each evidence item is {passage_id, quote} where
passage_id is the id from a header line and quote is copied VERBATIM from that passage — never
paraphrase, never edit, never merge across passages. If nothing supports the claim, return an empty
evidence array and verdict "absent".

VERDICT:
- "faithful": the summary reports the claim as the speaker stated it, including register and force.
- "partial": the source supports a weaker/narrower version.
- "overstated": same in kind but the summary strengthened it — or literalized a rhetorical claim.
- "absent": not in the passages.
- "contradicted": the source says the opposite.
For "partial"/"overstated", give what_source_actually_says (the faithful version, correct register);
otherwise set it to "".

GAP — the single field that most changes what a reader would do, err toward naming one over "none":
"scope" (narrower population/place/category), "denominator" (a fraction/share whose base is dropped),
"timerange" (a limited period dropped), "attribution" (support is about a different actor/programme/
body), "other", or "none" (benign narrowing; use for "faithful").

SO_WHAT: if a reader who believed the summary would act on a false impression, state in <=20 words
what they would get wrong; "" when faithful or the gap is benign. REASON: <=40 words, why this verdict.`

const evidenceSys = `You are the Evidence Grounder. Decide whether the CLAIM is TRUE, using web
search to find real, current evidence — the actual truth-makers, not anyone's assertion that it is
true. Search for data, primary sources, and credible reporting; weigh what you find. Then judge:
- "supported": credible evidence backs the claim.
- "mixed": evidence cuts both ways, or supports only a qualified version.
- "refuted": credible evidence contradicts the claim.
- "unverifiable": a prediction, opinion, or otherwise not checkable against current evidence.

Keep finding to one sentence. List the sources you actually used, with real URLs from your search
results. Return ONLY JSON after searching:
{"verdict":"supported"|"mixed"|"refuted"|"unverifiable","finding":string,"sources":[{"title":string,"url":string}]}`

const defaultInput = `1. The future of work will happen inside Codex or Claude Code.
2. Every company will have one super-agent inside their Slack.
3. SaaS is not dead — I would buy SaaS stocks right now.
4. PMs will thrive in the AI era.`

// ── API ──────────────────────────────────────────────────────────────────────

func (c cfg) callJSON(system, cached, prompt string, withTools bool, v any) error {
	_, err := c.callJSONSourced(system, cached, prompt, withTools, v)
	return err
}

// callJSONSourced is callJSON with retrieved sources threaded through. Used by evidenceClaim,
// which is the only caller that needs to cross-check model-claimed URLs against actual retrieval.
// `cached` is the stable prefix (source corpus) the backend reuses ahead of `prompt`; pass "" when
// the call has no reusable prefix.
func (c cfg) callJSONSourced(system, cached, prompt string, withTools bool, v any) ([]retrievedSource, error) {
	c.cachedSource = cached // value copy; read by dispatch when building the backend Request
	out, rs, err := c.dispatch(system, prompt, withTools)
	if err != nil {
		return nil, err
	}
	if unmarshalLoose(out, v) == nil {
		return rs, nil
	}
	strict := system +
		"\n\nReturn ONLY raw JSON. No prose, no markdown, no backticks. " +
		"First character must be { or [."
	out2, rs2, err := c.dispatch(strict, prompt, withTools)
	if err != nil {
		return nil, err
	}
	return rs2, unmarshalLoose(out2, v)
}

// callSchema makes one structured-output judge call: the backend constrains the answer to `schema`
// (Anthropic via a forced strict tool, Ollama via `format`), `cached` is the reusable passage prefix,
// and `temp` sets sampling (nil = provider default). On a parse failure — schema-invalid JSON despite
// the constraint — it retries once, counting the retry in usage, then returns whatever the retry
// parsed. The retry itself is a real billed call, so its tokens land in usage automatically.
func (c cfg) callSchema(system, cached, prompt string, schema json.RawMessage, name string, temp *float64, v any) error {
	c.cachedSource = cached
	c.reqSchema = schema
	c.reqSchemaName = name
	c.reqTemp = temp
	out, _, err := c.dispatch(system, prompt, false)
	if err != nil {
		return err
	}
	if unmarshalLoose(out, v) == nil {
		return nil
	}
	if c.usage != nil {
		c.usage.addSchemaRetry()
	}
	out2, _, err := c.dispatch(system, prompt, false)
	if err != nil {
		return err
	}
	return unmarshalLoose(out2, v)
}

// dispatch runs one model call. When the test seam `call` is set it is used directly (canned JSON,
// no network); otherwise the configured backend's Complete is invoked, its usage accumulated, and —
// under -v — the request and response are traced. Usage is added even on an error return, because a
// truncated (max_tokens / length) response was still billed.
func (c cfg) dispatch(system, prompt string, withTools bool) (string, []retrievedSource, error) {
	if c.call != nil {
		return c.call(system, prompt, withTools)
	}
	if c.verbose {
		fmt.Println(c.cyan("┌─ call · " + c.backendName + " · " + c.model + tern(withTools, " (web search)", "")))
		fmt.Println(c.grey("│ system:\n│   " + strings.ReplaceAll(system, "\n", "\n│   ")))
		if c.cachedSource != "" {
			fmt.Printf(c.grey("│ cached prefix: %d bytes\n"), len(c.cachedSource))
		}
		fmt.Println(c.grey("│ user:\n│   " + strings.ReplaceAll(prompt, "\n", "\n│   ")))
	}
	resp, err := c.backend.Complete(backend.Request{
		System: system, Prompt: prompt, Cached: c.cachedSource, WithTools: withTools,
		Schema: c.reqSchema, SchemaName: c.reqSchemaName, Temperature: c.reqTemp,
	})
	if c.usage != nil {
		u := resp.Usage
		c.usage.add(u.InputTokens, u.OutputTokens, u.CacheRead, u.CacheCreate, u.WebSearches)
		if c.verbose {
			fmt.Fprintf(os.Stderr, "[call] in=%d out=%d cache_read=%d cache_create=%d\n",
				u.InputTokens, u.OutputTokens, u.CacheRead, u.CacheCreate)
		}
	}
	if c.verbose {
		fmt.Println(c.cyan("├─ response:"))
		fmt.Println(c.grey("│ " + strings.ReplaceAll(resp.Text, "\n", "\n│ ")))
		if len(resp.Sources) > 0 {
			fmt.Printf(c.grey("│ retrieved %d source(s)\n"), len(resp.Sources))
		}
		fmt.Println(c.cyan("└─"))
	}
	return resp.Text, resp.Sources, err
}

// ── JSON extraction ──────────────────────────────────────────────────────────

var (
	fenceRe         = regexp.MustCompile("```(?:json)?")
	trailingCommaRe = regexp.MustCompile(`,(\s*[}\]])`)
)

func extractJSON(s string) string {
	s = strings.TrimSpace(fenceRe.ReplaceAllString(s, ""))
	start := -1
	for i, c := range s {
		if c == '{' || c == '[' {
			start = i
			break
		}
	}
	if start < 0 {
		return s
	}
	open := s[start]
	close := byte('}')
	if open == '[' {
		close = ']'
	}
	depth, inStr, esc := 0, false, false
	for i := start; i < len(s); i++ {
		c := s[i]
		if inStr {
			switch {
			case esc:
				esc = false
			case c == '\\':
				esc = true
			case c == '"':
				inStr = false
			}
			continue
		}
		switch c {
		case '"':
			inStr = true
		case open:
			depth++
		case close:
			depth--
			if depth == 0 {
				return s[start : i+1]
			}
		}
	}
	return s[start:]
}

func unmarshalLoose(s string, v any) error {
	js := trailingCommaRe.ReplaceAllString(extractJSON(s), "$1")
	return json.Unmarshal([]byte(js), v)
}

// ── claim splitting ──────────────────────────────────────────────────────────

var (
	leadingMarkerRe = regexp.MustCompile(`^\s*(?:\d{1,2}[.)]|[-*\x{2022}])\s+`)
	runOnMarkerRe   = regexp.MustCompile(`(?:^|[\s.?!")\x{201d}])(\d{1,2}[.)])\s+`)
)

func splitSummary(text string) []string {
	text = strings.TrimSpace(text)
	var items []string
	for _, ln := range strings.Split(text, "\n") {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		items = append(items, strings.TrimSpace(leadingMarkerRe.ReplaceAllString(ln, "")))
	}
	if len(items) >= 2 {
		return items
	}
	// single blob: split run-together numbered markers
	locs := runOnMarkerRe.FindAllStringIndex(text, -1)
	if len(locs) >= 2 {
		var out []string
		for i, loc := range locs {
			end := len(text)
			if i+1 < len(locs) {
				end = locs[i+1][0]
			}
			if seg := strings.TrimSpace(text[loc[1]:end]); seg != "" {
				out = append(out, seg)
			}
		}
		if len(out) >= 2 {
			return out
		}
	}
	if text == "" {
		return nil
	}
	return []string{text}
}

// ── result types ─────────────────────────────────────────────────────────────

type critiqueItem struct{ Axis, Finding, Severity string }
type substance struct {
	Claim, Steelman, Verdict, SurvivingClaim, Reason string
	Critique                                         []critiqueItem
	NeedsAnother                                     bool
	Rounds                                           int
}
type faith struct {
	Claim, Verdict, Evidence, SourceSays string
	Gap                                  string   // {none,scope,denominator,timerange,attribution,other}
	SoWhat                               string   // ≤20 words: what a reader who believed the summary gets wrong
	Quotes                               []string // defender's verbatim source spans, for the tree leaf
}
type source struct{ Title, URL string }
type evidence struct {
	Claim, Verdict, Finding string
	Sources                 []source
	RetrievedSources        []retrievedSource
	SourcesVerified         int
	DowngradeReason         string
	OriginalVerdict         string
}

// JSON shims (tagged) ----------------------------------------------------------

type substanceJSON struct {
	Steelman string `json:"steelman"`
	Critique []struct {
		Axis     string `json:"axis"`
		Finding  string `json:"finding"`
		Severity string `json:"severity"`
	} `json:"critique"`
	Verdict                  string `json:"verdict"`
	SurvivingClaim           string `json:"surviving_claim"`
	Reason                   string `json:"reason"`
	NeedsAnother             bool   `json:"needs_another_round"`
	AddedConditions          int    `json:"added_conditions"`
	SurvivesOnlyByConditions bool   `json:"survives_only_by_conditioning"`
}
type producerJSON struct {
	Steelman   string `json:"steelman"`
	Conditions string `json:"conditions"`
}
type defenderJSON struct {
	Found  bool     `json:"found"`
	Quotes []string `json:"quotes"`
}
type faithJSON struct {
	Findings []struct {
		Mode    string `json:"mode"`
		Finding string `json:"finding"`
	} `json:"findings"`
	Verdict    string `json:"verdict"`
	Evidence   string `json:"evidence"`
	SourceSays string `json:"what_source_actually_says"`
	Gap        string `json:"gap"`
	SoWhat     string `json:"so_what"`
}
type evidenceJSON struct {
	Verdict string `json:"verdict"`
	Finding string `json:"finding"`
	Sources []struct {
		Title string `json:"title"`
		URL   string `json:"url"`
	} `json:"sources"`
}

// judgeJSON is the single faithfulness judge's schema-enforced output (judgeSchema). Evidence cites
// the source by passage id and a verbatim quote, so the grounding check is a code-side substring test.
type judgeJSON struct {
	Verdict  string `json:"verdict"`
	Gap      string `json:"gap"`
	Evidence []struct {
		PassageID string `json:"passage_id"`
		Quote     string `json:"quote"`
	} `json:"evidence"`
	SoWhat     string `json:"so_what"`
	Reason     string `json:"reason"`
	SourceSays string `json:"what_source_actually_says"`
}

// ── markdown rendering ───────────────────────────────────────────────────────

func mdCell(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.TrimSpace(s)
}

var badVerdict = map[string]bool{
	"hollow":       true,
	"absent":       true,
	"refuted":      true,
	"contradicted": true,
	"error":        true}

func mdV(v string) string {
	if badVerdict[v] {
		return "**" + v + "**"
	}
	return v
}

func tally(vs []string) string {
	order := []string{}
	seen := map[string]int{}
	for _, v := range vs {
		if seen[v] == 0 {
			order = append(order, v)
		}
		seen[v]++
	}
	parts := make([]string, 0, len(order))
	for _, v := range order {
		parts = append(parts, fmt.Sprintf("%d %s", seen[v], v))
	}
	return strings.Join(parts, " · ")
}

func mdSubstance(rs []substance) string {
	var b strings.Builder
	b.WriteString("## Substance — is the claim well-formed?\n\n")
	b.WriteString("| # | Claim | Verdict | Why |\n|---|---|---|---|\n")
	vs := []string{}
	for i, r := range rs {
		why := r.Reason
		if why == "" && r.SurvivingClaim != "" {
			why = "survives as: " + r.SurvivingClaim
		}
		b.WriteString(fmt.Sprintf("| %d | %s | %s | %s |\n",
			i+1, mdCell(r.Claim), mdV(r.Verdict), mdCell(why)))
		vs = append(vs, r.Verdict)
	}
	b.WriteString("\n**" + tally(vs) + "**\n")
	return b.String()
}

func mdFaith(rs []faith) string {
	var b strings.Builder
	b.WriteString("## Faithfulness — did the source actually say it?\n\n")
	b.WriteString("| # | Claim | Verdict | Source actually says |\n|---|---|---|---|\n")
	vs := []string{}
	for i, r := range rs {
		b.WriteString(fmt.Sprintf("| %d | %s | %s | %s |\n",
			i+1, mdCell(r.Claim), mdV(r.Verdict), mdCell(r.SourceSays)))
		vs = append(vs, r.Verdict)
	}
	b.WriteString("\n**" + tally(vs) + "**\n")
	return b.String()
}

func mdEvidence(rs []evidence) string {
	var b strings.Builder
	b.WriteString("## Grounding — is it true, per external evidence?\n\n")
	b.WriteString("| # | Claim | Verdict | Finding | Sources |\n|---|---|---|---|---|\n")
	vs := []string{}
	for i, r := range rs {
		srcs := make([]string, 0, len(r.Sources))
		for _, s := range r.Sources {
			srcs = append(srcs, fmt.Sprintf("[%s](%s)", mdCell(s.Title), s.URL))
		}
		finding := mdCell(r.Finding)
		if r.DowngradeReason != "" {
			finding += " *(was: " + r.OriginalVerdict + "; " + r.DowngradeReason + ")*"
		}
		b.WriteString(fmt.Sprintf("| %d | %s | %s | %s | %s |\n",
			i+1, mdCell(r.Claim), mdV(r.Verdict), finding, strings.Join(srcs, "; ")))
		vs = append(vs, r.Verdict)
	}
	b.WriteString("\n**" + tally(vs) + "**\n")
	return b.String()
}

func mdAudit(claims []string, fs []faith, ss []substance, es []evidence) string {
	var b strings.Builder
	b.WriteString("## Audit — three filters, cross-tabulated\n\n")
	b.WriteString("| # | Claim | Faithful? | Substantive? | Grounded? |\n|---|---|---|---|---|\n")
	for i, c := range claims {
		b.WriteString(fmt.Sprintf("| %d | %s | %s | %s | %s |\n",
			i+1, mdCell(c), mdV(fs[i].Verdict), mdV(ss[i].Verdict), mdV(es[i].Verdict)))
	}
	b.WriteString("\n**Pattern-reading.**\n\n")
	b.WriteString("| Pattern | Reading |\n|---|---|\n")
	b.WriteString("| faithful + substantive/partial + supported | **Signal** — a real, checkable claim that holds |\n")
	b.WriteString("| faithful + substantive/partial + mixed/unverifiable | **Genuine bet** — well-formed and honestly attributed, not yet settled |\n")
	b.WriteString("| faithful + substantive/partial + **refuted** | **Anti-signal** — checkable and wrong |\n")
	b.WriteString("| **overstated**/**absent** + anything | **Summarizer's noise** — go back to the source |\n")
	b.WriteString("| faithful + **hollow** | **Speaker's noise** — vacuous, but accurately reported |\n")
	return b.String()
}

// ── terminal rendering ───────────────────────────────────────────────────────

func (c cfg) col(code, s string) string {
	if c.noColor || c.asMarkdown {
		return s
	}
	return "\033[" + code + "m" + s + "\033[0m"
}
func (c cfg) bold(s string) string   { return c.col("1", s) }
func (c cfg) grey(s string) string   { return c.col("90", s) }
func (c cfg) cyan(s string) string   { return c.col("36", s) }
func (c cfg) green(s string) string  { return c.col("32", s) }
func (c cfg) yellow(s string) string { return c.col("33", s) }
func (c cfg) red(s string) string    { return c.col("31", s) }

func (c cfg) vcolor(v string) func(string) string {
	switch v {
	case "substantive", "faithful", "supported":
		return c.green
	case "partial", "overstated", "mixed":
		return c.yellow
	case "hollow", "absent", "refuted", "contradicted":
		return c.red
	default:
		return c.grey
	}
}

func (c cfg) termLine(verdict, claim, why string) {
	col := c.vcolor(verdict)
	fmt.Println("\n" + col(fmt.Sprintf("%-13s", strings.ToUpper(verdict))) + claim)
	if why != "" {
		fmt.Println(c.grey("  " + why))
	}
}

func (c cfg) termSubstance(rs []substance) {
	vs := []string{}
	for _, r := range rs {
		why := r.Reason
		if r.SurvivingClaim != "" && r.Verdict != "substantive" {
			why = "survives as: " + r.SurvivingClaim
		}
		c.termLine(r.Verdict, r.Claim, why)
		vs = append(vs, r.Verdict)
	}
	fmt.Println(c.bold("\n" + tally(vs)))
}

func (c cfg) termFaith(rs []faith) {
	vs := []string{}
	for _, r := range rs {
		c.termLine(r.Verdict, r.Claim, tern(r.SourceSays != "", "source says: "+r.SourceSays, ""))
		vs = append(vs, r.Verdict)
	}
	fmt.Println(c.bold("\n" + tally(vs)))
}

func (c cfg) termEvidence(rs []evidence) {
	vs := []string{}
	for _, r := range rs {
		c.termLine(r.Verdict, r.Claim, r.Finding)
		for _, s := range r.Sources {
			fmt.Println(c.grey("  · " + s.Title + " — " + s.URL))
		}
		if r.DowngradeReason != "" {
			fmt.Println(c.yellow("  ↓ downgraded from " + r.OriginalVerdict + ": " + r.DowngradeReason))
		}
		vs = append(vs, r.Verdict)
	}
	fmt.Println(c.bold("\n" + tally(vs)))
}

func tern(c bool, a, b string) string {
	if c {
		return a
	}
	return b
}

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func mustRead(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		fatal("cannot read " + path + ": " + err.Error())
	}
	return string(b)
}

// treeFlag backs -tree, which takes no value (default expansion) or "=full" (expand every node).
// IsBoolFlag lets `-tree` stand alone; Set still receives "full" for `-tree=full`.
type treeFlag struct{ on, all bool }

func (f *treeFlag) String() string   { return "" }
func (f *treeFlag) IsBoolFlag() bool { return true }
func (f *treeFlag) Set(v string) error {
	f.on = true
	switch v {
	case "", "true":
		f.all = false
	case "full":
		f.all = true
	default:
		return fmt.Errorf("-tree takes no value or =full, got %q", v)
	}
	return nil
}

// parseClaimLine pulls an optional "<id>\t<path>\t<text>" prefix off a claim line so the tree can
// key claims to the source document's headings (path carries "/"-separated key=label segments).
// A line with no tabs is a bare claim: no id, no path, the whole line is the text.
func parseClaimLine(raw string) (id, path, text string) {
	if parts := strings.SplitN(raw, "\t", 3); len(parts) == 3 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), strings.TrimSpace(parts[2])
	}
	return "", "", raw
}

// present routes one run's verdicts to the chosen stdout renderer and always writes the full table
// to auditPath, so the eval directory holds the complete result regardless of what stdout showed.
// `mdTable` is the markdown full table (also what audit.md gets); `termTable` prints the colour full
// table for `-full` without `-md`.
func (c *cfg) present(rows []brief.Row, counts, mdTable string, termTable func(), details map[string]tree.Leaf) {
	if c.auditPath != "" {
		if err := os.WriteFile(c.auditPath, []byte(mdTable), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "warning: cannot write %s: %v\n", c.auditPath, err)
		}
	}
	if c.treeHTMLPath != "" {
		htmlDoc := tree.RenderHTML(rows, details, c.headerLine())
		if err := os.WriteFile(c.treeHTMLPath, []byte(htmlDoc), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "warning: cannot write %s: %v\n", c.treeHTMLPath, err)
		} else {
			fmt.Fprintf(os.Stderr, "tree (html): %s\n", c.treeHTMLPath)
		}
	}
	fmt.Println(c.headerLine())
	// The value line (docs/VALUE.md): how many executive-summary lines this run changes.
	fmt.Printf("changes %d summary lines\n", brief.OpenedBranches(rows))
	switch c.renderMode {
	case "full":
		if c.asMarkdown {
			fmt.Print(mdTable)
		} else {
			termTable()
		}
	case "tree":
		fmt.Print(tree.Render(rows, c.treeAll, c.auditPath))
	default:
		s, _ := brief.Brief(rows, counts, c.auditPath)
		fmt.Print(s)
	}
}

// headerLine is the one-line provenance stamp printed atop every brief/tree report and embedded in
// tree.html: the model, the API call count, the estimated cost, and wall time. Under -from every
// figure is zero because no model call was made, which is the point — the reader sees the render
// cost nothing.
func (c cfg) headerLine() string {
	calls, in, out, cr, cc, ws, _, elapsed := c.usage.snapshot()
	cost, _ := estimateCost(c.model, in, out, cr, cc, ws)
	be := c.backendName
	if be == "" {
		be = "—" // -from: rendered from a saved chain, no backend was wired
	}
	return fmt.Sprintf("model %s · backend %s · calls %d · %s · wall %s",
		c.model, be, calls, cost, elapsed.Round(time.Second))
}

func readInput(text string) string {
	if text != "" {
		return text
	}
	if a := flag.Arg(0); a != "" {
		return mustRead(a)
	}
	if fi, _ := os.Stdin.Stat(); fi != nil && fi.Mode()&os.ModeCharDevice == 0 {
		b, _ := io.ReadAll(os.Stdin)
		if s := strings.TrimSpace(string(b)); s != "" {
			return s
		}
	}
	return defaultInput
}

// ── Tier-2 verification chain (JSONL per case) ───────────────────────────────

// chainRecord is the common envelope written for every case. Mode-specific fields
// are embedded as a json.RawMessage under "detail" to keep the schema flat.
type chainRecord struct {
	Idx      int             `json:"idx"`
	Total    int             `json:"total"`
	Mode     string          `json:"mode"`
	Backend  string          `json:"backend,omitempty"` // provider that produced this verdict
	Claim    string          `json:"claim"`
	Verdict  string          `json:"verdict"`
	Spread   string          `json:"spread,omitempty"`   // "k/N" agreement when -n>1, else ""
	Passages []string        `json:"passages,omitempty"` // retrieved passage ids the judge saw (bm25 mode)
	ElapsedS float64         `json:"elapsed_s"`
	Detail   json.RawMessage `json:"detail"`
}

// appendChain appends one JSON record to c.chainFile. Errors are logged to
// stderr and silently dropped — chain failures must never abort the run.
func (c *cfg) appendChain(rec chainRecord) {
	if c.chainFile == "" {
		return
	}
	rec.Backend = c.backendName // stamp the producing backend on every record
	if c.verbose {
		fmt.Fprintf(os.Stderr, "[chain] %s idx=%d verdict=%s\n", c.chainFile, rec.Idx, rec.Verdict)
	}
	data, err := json.Marshal(rec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: chain marshal idx=%d: %v\n", rec.Idx, err)
		return
	}
	f, err := os.OpenFile(c.chainFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: chain open %q: %v\n", c.chainFile, err)
		return
	}
	defer f.Close()
	fmt.Fprintln(f, string(data))
}

type substanceDetail struct {
	Steelman        string         `json:"steelman"`
	CritiqueByAxis  []critiqueItem `json:"critique_by_axis"`
	SurvivingClaim  string         `json:"surviving_claim,omitempty"`
	AddedConditions int            `json:"added_conditions,omitempty"`
	Rounds          int            `json:"rounds"`
	Reason          string         `json:"reason,omitempty"`
}

type faithDetail struct {
	DefenderSupport string   `json:"defender_support,omitempty"`
	Quotes          []string `json:"quotes,omitempty"` // verbatim source spans the defender cited
	CriticFinding   string   `json:"critic_finding,omitempty"`
	DistortionType  string   `json:"distortion_type,omitempty"`
	SourceSays      string   `json:"source_says,omitempty"`
	Gap             string   `json:"gap,omitempty"`     // the judge's gap classification
	SoWhat          string   `json:"so_what,omitempty"` // the stakes line for a needs-you leaf
}

type evidenceDetail struct {
	Finding          string            `json:"finding,omitempty"`
	Sources          []source          `json:"sources,omitempty"`
	RetrievedSources []retrievedSource `json:"retrieved_sources,omitempty"`
	SourcesVerified  int               `json:"sources_verified,omitempty"`
	DowngradeReason  string            `json:"downgrade_reason,omitempty"`
	OriginalVerdict  string            `json:"original_verdict,omitempty"`
	ErrorCause       string            `json:"error_cause,omitempty"`
}

type auditDetail struct {
	Faith     faithDetail     `json:"faith"`
	Substance substanceDetail `json:"substance"`
	Evidence  evidenceDetail  `json:"evidence"`
}

func substanceChainRecord(i, total int, claim string, s substance, start time.Time) chainRecord {
	det := substanceDetail{
		Steelman:       s.Steelman,
		CritiqueByAxis: s.Critique,
		SurvivingClaim: s.SurvivingClaim,
		Rounds:         s.Rounds,
		Reason:         s.Reason,
	}
	raw, _ := json.Marshal(det)
	return chainRecord{
		Idx: i, Total: total, Mode: "substance",
		Claim: claim, Verdict: s.Verdict,
		ElapsedS: time.Since(start).Seconds(),
		Detail:   raw,
	}
}

func faithChainRecord(i, total int, claim string, f faith, start time.Time) chainRecord {
	det := faithDetail{
		Quotes:        f.Quotes,
		CriticFinding: f.Evidence,
		SourceSays:    f.SourceSays,
		Gap:           f.Gap,
		SoWhat:        f.SoWhat,
	}
	raw, _ := json.Marshal(det)
	return chainRecord{
		Idx: i, Total: total, Mode: "faithfulness",
		Claim: claim, Verdict: f.Verdict,
		ElapsedS: time.Since(start).Seconds(),
		Detail:   raw,
	}
}

func evidenceChainRecord(i, total int, claim string, e evidence, start time.Time) chainRecord {
	errCause := ""
	if e.Verdict == "error" {
		errCause = e.Finding
	}
	det := evidenceDetail{
		Finding:          e.Finding,
		Sources:          e.Sources,
		RetrievedSources: e.RetrievedSources,
		SourcesVerified:  e.SourcesVerified,
		DowngradeReason:  e.DowngradeReason,
		OriginalVerdict:  e.OriginalVerdict,
		ErrorCause:       errCause,
	}
	raw, _ := json.Marshal(det)
	return chainRecord{
		Idx: i, Total: total, Mode: "grounding",
		Claim: claim, Verdict: e.Verdict,
		ElapsedS: time.Since(start).Seconds(),
		Detail:   raw,
	}
}

func auditChainRecord(i, n int, claim string, f faith, s substance, e evidence, start time.Time) chainRecord {
	det := auditDetail{
		Faith: faithDetail{
			Quotes:        f.Quotes,
			CriticFinding: f.Evidence,
			SourceSays:    f.SourceSays,
			Gap:           f.Gap,
			SoWhat:        f.SoWhat,
		},
		Substance: substanceDetail{
			Steelman:       s.Steelman,
			CritiqueByAxis: s.Critique,
			SurvivingClaim: s.SurvivingClaim,
			Rounds:         s.Rounds,
			Reason:         s.Reason,
		},
		Evidence: evidenceDetail{
			Finding:          e.Finding,
			Sources:          e.Sources,
			RetrievedSources: e.RetrievedSources,
			SourcesVerified:  e.SourcesVerified,
			DowngradeReason:  e.DowngradeReason,
			OriginalVerdict:  e.OriginalVerdict,
		},
	}
	verdictSummary := fmt.Sprintf("faith=%s sub=%s ev=%s", f.Verdict, s.Verdict, e.Verdict)
	raw, _ := json.Marshal(det)
	return chainRecord{
		Idx: i, Total: n, Mode: "audit",
		Claim: claim, Verdict: verdictSummary,
		ElapsedS: time.Since(start).Seconds(),
		Detail:   raw,
	}
}

// ── -from: replay a saved chain ──────────────────────────────────────────────

// readChain reads a Tier-2 JSONL chain into records ordered by idx and returns the single mode they
// share. It fails on a gap in the idx sequence, an empty file, or mixed modes — any of which means
// the chain is not the coherent record of one run that -from expects to replay.
func readChain(path string) ([]chainRecord, string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	byIdx := make(map[int]chainRecord)
	maxIdx := -1
	for _, ln := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		if ln = strings.TrimSpace(ln); ln == "" {
			continue
		}
		var r chainRecord
		if err := json.Unmarshal([]byte(ln), &r); err != nil {
			return nil, "", fmt.Errorf("bad JSONL line: %w", err)
		}
		byIdx[r.Idx] = r
		if r.Idx > maxIdx {
			maxIdx = r.Idx
		}
	}
	if maxIdx < 0 {
		return nil, "", fmt.Errorf("empty chain")
	}
	recs := make([]chainRecord, 0, maxIdx+1)
	for i := 0; i <= maxIdx; i++ {
		r, ok := byIdx[i]
		if !ok {
			return nil, "", fmt.Errorf("chain missing idx %d", i)
		}
		recs = append(recs, r)
	}
	mode := recs[0].Mode
	for _, r := range recs {
		if r.Mode != mode {
			return nil, "", fmt.Errorf("mixed modes in chain: %q and %q", mode, r.Mode)
		}
	}
	return recs, mode, nil
}

// parseAuditVerdict splits an audit record's "faith=X sub=Y ev=Z" verdict string back into its three
// component verdicts. auditChainRecord is the sole writer of that format.
func parseAuditVerdict(s string) (f, sub, ev string) {
	for _, tok := range strings.Fields(s) {
		k, v, ok := strings.Cut(tok, "=")
		if !ok {
			continue
		}
		switch k {
		case "faith":
			f = v
		case "sub":
			sub = v
		case "ev":
			ev = v
		}
	}
	return f, sub, ev
}

// runFromChain reconstructs the verdict rows from a saved chain and renders them through the same
// stdout renderers a live run uses (brief / -tree / -full), making no model call. The claims file is
// required: it supplies the ids and heading paths the tree needs, and cross-checking each record's
// claim text against it catches a chain paired with the wrong claims file.
func (c *cfg) runFromChain(chainPath, claimsPath string) {
	if claimsPath == "" {
		fatal("-from needs the claims file as INPUT (its claims are checked against the chain)")
	}
	recs, mode, err := readChain(chainPath)
	if err != nil {
		fatal("read chain " + chainPath + ": " + err.Error())
	}
	items := splitSummary(mustRead(claimsPath))
	if len(recs) != len(items) {
		fatal(fmt.Sprintf("chain has %d records but %s has %d claims", len(recs), claimsPath, len(items)))
	}
	ids := make([]string, len(items))
	paths := make([]string, len(items))
	texts := make([]string, len(items))
	for i, it := range items {
		id, path, text := parseClaimLine(it)
		if id == "" {
			id = fmt.Sprintf("c%d", i+1)
		}
		ids[i], paths[i], texts[i] = id, path, text
	}
	for i, r := range recs {
		if strings.TrimSpace(r.Claim) != strings.TrimSpace(texts[i]) {
			fatal(fmt.Sprintf("claim %d in the chain does not match %s:\n  chain:  %q\n  claims: %q",
				i+1, claimsPath, r.Claim, texts[i]))
		}
	}

	dir := filepath.Dir(chainPath)
	c.auditPath = filepath.Join(dir, "audit.md")
	c.treeHTMLPath = filepath.Join(dir, "tree.html")

	rows := make([]brief.Row, len(recs))
	details := make(map[string]tree.Leaf, len(recs))
	vs := make([]string, len(recs))

	switch mode {
	case "faithfulness":
		fr := make([]faith, len(recs))
		for i, r := range recs {
			var det faithDetail
			_ = json.Unmarshal(r.Detail, &det)
			fr[i] = faith{Claim: r.Claim, Verdict: r.Verdict, Evidence: det.CriticFinding,
				SourceSays: det.SourceSays, Gap: det.Gap, SoWhat: det.SoWhat, Quotes: det.Quotes}
			rows[i] = brief.Row{ID: ids[i], Path: paths[i], Text: texts[i],
				Faith: r.Verdict, FaithReason: det.CriticFinding, Spread: r.Spread,
				Gap: det.Gap, SoWhat: det.SoWhat}
			details[ids[i]] = tree.Leaf{Reason: det.CriticFinding, Quotes: det.Quotes}
			vs[i] = r.Verdict
		}
		c.present(rows, tally(vs), mdFaith(fr), func() { c.termFaith(fr) }, details)
	case "substance":
		sr := make([]substance, len(recs))
		for i, r := range recs {
			var det substanceDetail
			_ = json.Unmarshal(r.Detail, &det)
			reason := det.Reason
			if reason == "" && det.SurvivingClaim != "" {
				reason = "survives as: " + det.SurvivingClaim
			}
			sr[i] = substance{Claim: r.Claim, Verdict: r.Verdict, Reason: det.Reason,
				SurvivingClaim: det.SurvivingClaim, Steelman: det.Steelman, Rounds: det.Rounds}
			rows[i] = brief.Row{ID: ids[i], Path: paths[i], Text: texts[i],
				Substance: r.Verdict, SubstanceReason: reason}
			details[ids[i]] = tree.Leaf{Reason: reason}
			vs[i] = r.Verdict
		}
		c.present(rows, tally(vs), mdSubstance(sr), func() { c.termSubstance(sr) }, details)
	case "grounding":
		er := make([]evidence, len(recs))
		for i, r := range recs {
			var det evidenceDetail
			_ = json.Unmarshal(r.Detail, &det)
			er[i] = evidence{Claim: r.Claim, Verdict: r.Verdict, Finding: det.Finding,
				Sources: det.Sources, RetrievedSources: det.RetrievedSources,
				SourcesVerified: det.SourcesVerified, DowngradeReason: det.DowngradeReason,
				OriginalVerdict: det.OriginalVerdict}
			rows[i] = brief.Row{ID: ids[i], Path: paths[i], Text: texts[i],
				Grounding: r.Verdict, GroundReason: det.Finding}
			details[ids[i]] = tree.Leaf{Reason: det.Finding}
			vs[i] = r.Verdict
		}
		c.present(rows, tally(vs), mdEvidence(er), func() { c.termEvidence(er) }, details)
	case "audit":
		fr := make([]faith, len(recs))
		sr := make([]substance, len(recs))
		er := make([]evidence, len(recs))
		fvs := make([]string, len(recs))
		svs := make([]string, len(recs))
		evs := make([]string, len(recs))
		for i, r := range recs {
			var det auditDetail
			_ = json.Unmarshal(r.Detail, &det)
			fv, sv, evv := parseAuditVerdict(r.Verdict)
			fr[i] = faith{Claim: r.Claim, Verdict: fv, Evidence: det.Faith.CriticFinding,
				SourceSays: det.Faith.SourceSays, Gap: det.Faith.Gap, SoWhat: det.Faith.SoWhat,
				Quotes: det.Faith.Quotes}
			sr[i] = substance{Claim: r.Claim, Verdict: sv, Reason: det.Substance.Reason,
				SurvivingClaim: det.Substance.SurvivingClaim, Steelman: det.Substance.Steelman}
			er[i] = evidence{Claim: r.Claim, Verdict: evv, Finding: det.Evidence.Finding,
				Sources: det.Evidence.Sources, DowngradeReason: det.Evidence.DowngradeReason,
				OriginalVerdict: det.Evidence.OriginalVerdict}
			rows[i] = brief.Row{ID: ids[i], Path: paths[i], Text: texts[i],
				Faith: fv, Substance: sv, Grounding: evv,
				FaithReason: det.Faith.CriticFinding, GroundReason: det.Evidence.Finding,
				Gap: det.Faith.Gap, SoWhat: det.Faith.SoWhat}
			details[ids[i]] = tree.Leaf{Reason: det.Faith.CriticFinding, Quotes: det.Faith.Quotes}
			fvs[i], svs[i], evs[i] = fv, sv, evv
		}
		counts := fmt.Sprintf("faith[%s] · sub[%s] · ground[%s]", tally(fvs), tally(svs), tally(evs))
		c.present(rows, counts, mdAudit(texts, fr, sr, er), func() { fmt.Print(mdAudit(texts, fr, sr, er)) }, details)
	default:
		fatal("unknown chain mode: " + mode)
	}
}

// ── usage accounting ─────────────────────────────────────────────────────────

// priceEntry holds per-million-token rates and per-search cost for one model.
// Source: https://www.anthropic.com/pricing (retrieved 2026-06-01).
// Set a field to -1 to mark it unknown (triggers est_usd=n/a for that model).
type priceEntry struct {
	InputPerMtok       float64 // $ per 1M input tokens
	OutputPerMtok      float64 // $ per 1M output tokens
	CacheReadPerMtok   float64 // $ per 1M cache-read tokens
	CacheCreatePerMtok float64 // $ per 1M cache-creation tokens
	WebSearchPer1k     float64 // $ per 1k web-search requests
}

// priceTable maps model id → rates.
// Source: https://www.anthropic.com/pricing (retrieved 2026-06-01).
// TODO: update rates whenever Anthropic revises pricing.
var priceTable = map[string]priceEntry{
	"claude-opus-4-8": {
		InputPerMtok: 15.00, OutputPerMtok: 75.00,
		CacheReadPerMtok: 1.50, CacheCreatePerMtok: 18.75,
		WebSearchPer1k: 10.00,
	},
	"claude-sonnet-4-6": {
		InputPerMtok: 3.00, OutputPerMtok: 15.00,
		CacheReadPerMtok: 0.30, CacheCreatePerMtok: 3.75,
		WebSearchPer1k: 10.00,
	},
	"claude-haiku-4-5-20251001": {
		InputPerMtok: 0.80, OutputPerMtok: 4.00,
		CacheReadPerMtok: 0.08, CacheCreatePerMtok: 1.00,
		WebSearchPer1k: 10.00,
	},
}

// priceTableDate is stamped alongside any dollar figure in output.
const priceTableDate = "2026-06-01"

// usageCounters accumulates model usage across every backend call in a run.
type usageCounters struct {
	mu                sync.Mutex
	calls             int
	inputTokens       int
	outputTokens      int
	cacheReadTokens   int
	cacheCreateTokens int
	webSearches       int
	schemaRetries     int // judge calls that came back schema-invalid and were retried once
	quoteRejects      int // evidence quotes rejected as non-verbatim by the grounding check
	start             time.Time
	currentClaim      string // label of the in-flight claim for heartbeat display
}

func newUsageCounters() *usageCounters {
	return &usageCounters{start: time.Now()}
}

func (u *usageCounters) add(in, out, cacheRead, cacheCreate, webSearches int) {
	u.mu.Lock()
	u.calls++
	u.inputTokens += in
	u.outputTokens += out
	u.cacheReadTokens += cacheRead
	u.cacheCreateTokens += cacheCreate
	u.webSearches += webSearches
	u.mu.Unlock()
}

func (u *usageCounters) setLabel(label string) {
	u.mu.Lock()
	u.currentClaim = label
	u.mu.Unlock()
}

func (u *usageCounters) addSchemaRetry() { u.mu.Lock(); u.schemaRetries++; u.mu.Unlock() }
func (u *usageCounters) extras() (schemaRetries, quoteRejects int) {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.schemaRetries, u.quoteRejects
}
func (u *usageCounters) addQuoteReject(n int) {
	u.mu.Lock()
	u.quoteRejects += n
	u.mu.Unlock()
}

// cacheHitRate is cache-read tokens over all cacheable input (read + created), 0 when nothing was
// cacheable. It is the per-backend prompt-cache effectiveness the SUMMARY reports; Ollama bills no
// cache tokens, so it reads 0 there even when keep_alive kept the KV cache warm.
func (u *usageCounters) cacheHitRate() float64 {
	u.mu.Lock()
	defer u.mu.Unlock()
	denom := u.cacheReadTokens + u.cacheCreateTokens
	if denom == 0 {
		return 0
	}
	return float64(u.cacheReadTokens) / float64(denom)
}

// snapshot returns a consistent read without holding the lock across formatting.
func (u *usageCounters) snapshot() (
	calls, in, out, cacheRead, cacheCreate, webSearches int,
	label string,
	elapsed time.Duration,
) {
	u.mu.Lock()
	calls = u.calls
	in = u.inputTokens
	out = u.outputTokens
	cacheRead = u.cacheReadTokens
	cacheCreate = u.cacheCreateTokens
	webSearches = u.webSearches
	label = u.currentClaim
	elapsed = time.Since(u.start)
	u.mu.Unlock()
	return
}

// estimateCost returns (estUSD string, ok bool). When prices are unknown or
// unset, ok is false and the string carries the reason sentinel.
func estimateCost(model string, in, out, cacheRead, cacheCreate, webSearches int) (string, bool) {
	p, found := priceTable[model]
	if !found {
		return fmt.Sprintf("n/a(model not in price table) in=%d out=%d cache_read=%d cache_create=%d",
			in, out, cacheRead, cacheCreate), false
	}
	if p.InputPerMtok < 0 {
		return "TODO(prices unset)", false
	}
	usd := float64(in)*p.InputPerMtok/1e6 +
		float64(out)*p.OutputPerMtok/1e6 +
		float64(cacheRead)*p.CacheReadPerMtok/1e6 +
		float64(cacheCreate)*p.CacheCreatePerMtok/1e6 +
		float64(webSearches)*p.WebSearchPer1k/1000
	return fmt.Sprintf("$%.4f (rates %s)", usd, priceTableDate), true
}

// heartbeatLine produces a one-line cumulative status for the 60s ticker.
// `rt` may be nil (e.g. when called before the runner initialises its tally).
func heartbeatLine(model string, u *usageCounters, rt *runTally) string {
	calls, in, out, cacheRead, cacheCreate, webSearches, label, elapsed := u.snapshot()
	cost, _ := estimateCost(model, in, out, cacheRead, cacheCreate, webSearches)
	tallyPart := ""
	if rt != nil {
		total, verified, counts := rt.snapshot()
		// errored is failures observed so far, not the not-yet-processed remainder. Cases that have
		// not run yet are simply absent from `seen`, never counted as errors.
		errored := counts["error"]
		seen := verified + errored // verified already includes any "skipped (over cap)" cases
		tallyPart = fmt.Sprintf(" verified=%d errored=%d seen=%d/%d", verified, errored, seen, total)
	}
	return fmt.Sprintf("[heartbeat] elapsed=%s calls=%d in=%d out=%d web=%d est=%s%s claim=%q",
		elapsed.Round(time.Second), calls, in, out, webSearches, cost, tallyPart, label)
}

// startHeartbeat starts a background goroutine that prints a one-line status to stderr every 60
// seconds. Cancel it via the returned cancel func.
// rt is a pointer to the cfg's tally field — it may be nil at heartbeat start and populated later.
func startHeartbeat(model string, u *usageCounters, getTally func() *runTally) func() {
	done := make(chan struct{})
	go func() {
		t := time.NewTicker(60 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-done:
				return
			case <-t.C:
				fmt.Fprintln(os.Stderr, heartbeatLine(model, u, getTally()))
			}
		}
	}()
	return func() { close(done) }
}

// printUsageLine writes the final greppable USAGE line to stderr.
func printUsageLine(model string, u *usageCounters, usageOutFile string) {
	calls, in, out, cacheRead, cacheCreate, webSearches, _, elapsed := u.snapshot()
	cost, _ := estimateCost(model, in, out, cacheRead, cacheCreate, webSearches)
	line := fmt.Sprintf(
		"USAGE model=%s calls=%d in=%d out=%d cache_read=%d cache_create=%d web_searches=%d wall=%ds est_usd=%s",
		model, calls, in, out, cacheRead, cacheCreate, webSearches,
		int(elapsed.Seconds()), cost,
	)
	fmt.Fprintln(os.Stderr, line)

	if usageOutFile != "" {
		type usageRecord struct {
			Model             string  `json:"model"`
			Calls             int     `json:"calls"`
			InputTokens       int     `json:"input_tokens"`
			OutputTokens      int     `json:"output_tokens"`
			CacheReadTokens   int     `json:"cache_read_tokens"`
			CacheCreateTokens int     `json:"cache_create_tokens"`
			WebSearches       int     `json:"web_searches"`
			WallSeconds       float64 `json:"wall_seconds"`
			EstUSD            *string `json:"est_usd"`
		}
		estStr, ok := estimateCost(model, in, out, cacheRead, cacheCreate, webSearches)
		var estPtr *string
		if ok {
			estPtr = &estStr
		}
		rec := usageRecord{
			Model: model, Calls: calls,
			InputTokens: in, OutputTokens: out,
			CacheReadTokens: cacheRead, CacheCreateTokens: cacheCreate,
			WebSearches: webSearches, WallSeconds: elapsed.Seconds(),
			EstUSD: estPtr,
		}
		data, _ := json.Marshal(rec)
		f, err := os.OpenFile(usageOutFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: cannot open -usage-out %q: %v\n", usageOutFile, err)
			return
		}
		defer f.Close()
		fmt.Fprintln(f, string(data))
	}
}
