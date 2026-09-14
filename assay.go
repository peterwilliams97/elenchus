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
	"archive/zip"
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"assay/internal/adjudicate"
	"assay/internal/backend"
	"assay/internal/backend/anthropic"
	"assay/internal/backend/ollama"
	"assay/internal/brief"
	"assay/internal/edge"
	"assay/internal/embed"
	"assay/internal/manifest"
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
	model          string
	apiKey         string
	maxRounds      int
	maxClaims      int
	repeat         int // -n: run each claim this many times and report the modal verdict + agreement
	verbose        bool
	noColor        bool
	asMarkdown     bool
	showProgress   bool
	quiet          bool
	fresh          bool // -fresh: re-judge every claim instead of resuming from an existing chain
	usageOut       string
	chainFile      string              // Tier-2 JSONL destination; set by runners before case loop
	substanceOn    bool                // -axis substance: also run the substance axis in the -source corpus run
	substanceFile  string              // <chain-dir>/<fixture>.substance.jsonl, when substanceOn (spec/SUBSTANCE-CORPUS.md)
	ceScoped       bool                // -ce-scoped: substance critic uses the Counterexample-scope variant prompt
	narrowBoundary bool                // -narrowing-boundary: verdict keys on how far the surviving claim was narrowed, not axis severity
	asOf           bool                // -as-of: judge the claim as of its date, barring the critic from citing the realised outcome
	renderMode     string              // stdout renderer: "brief" (default), "full", or "tree"
	treeAll        bool                // -tree=full: expand every node rather than only Needs-you branches
	retrieveMode   string              // "bm25" (default) retrieves per-claim passages; "none" sends the full corpus
	maxTokens      int                 // -max-tokens: retrieval token budget per claim (fused bm25+embed ranking)
	floor          float64             // -floor: top-passage cosine below this ⇒ absent from code, no model call
	embed          bool                // -embed: add the nomic-embed-text ranker, fused with BM25 by RRF
	oracle         map[string][]string // -retrieve=oracle: claim-id → fixed gold passage ids
	held           map[string]bool     // -manifest: doc ids the corpus holds; a claim citing a missing one is unverifiable
	docLabels      map[string]string   // -manifest: canonical doc id → witness/author, for a quote's provenance line
	refs           map[string]string   // -refs: claim id → report section ref (§…), for the leaf's report line
	index          *retrieve.Index     // built once from the source corpus when retrieveMode != "none"
	auditPath      string              // full-table sink; every run writes it, whichever renderer stdout gets
	treeHTMLPath   string              // eval/<stamp>/tree.html sink, written alongside audit.md
	argumentFile   string              // -argument: argument.txt path; when set, -from renders the argument tree
	indexHTMLPath  string              // index.html sink (the argument-tree page), alongside audit.md, when -argument is set
	review         bool                // -review: build the self-contained review site (spec/SERVE.md)
	edge           bool                // -edge: run the edge-level adversarial pass over in-scope F→R edges (spec/EDGE.md)
	edgeChainDir   string              // -chain-dir for the edge pass; default testing/chains/edge-<stamp>-<model>
	zip            bool                // -zip: write site.zip beside site/ after -review builds it
	siteDir        string              // site/ sink under the example dir, when -review is set
	sourcesDir     string              // -manifest's parent dir; the on-disk PDF tree the site copies from
	reports        []manifest.Report   // -manifest report excerpts: each with its file, page offset, sections, and claim-id prefixes
	rootTitle      string              // report title for the root block's line 1, from the claims file "# title:" header
	rootDate       string              // report date for the root block's line 1, from the claims file "# date:" header
	sourceDocs     int                 // M: documents held per the manifest; the root block's "checked against M" figure
	singleSource   bool                // -manifest single_source: report is its own only source; leaf 'absent' → 'uncorroborated'
	runs           int                 // chains merged under -from; >1 adds the root block's stability line. 0/1 = single run
	usage          *usageCounters
	tally          *runTally
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
	// serve is a subcommand with its own flag set — the review page it serves needs a real HTTP
	// origin, not the global claim-assay flags. Dispatch before flag.Parse so -port/-no-open don't
	// collide with the assay flags of the same shape.
	if len(os.Args) > 1 && os.Args[1] == "serve" {
		runServe(os.Args[2:])
		return
	}

	var c cfg
	var src, text, chainDir, fromChain, axisList string
	var ev, audit, full bool
	var treeF treeFlag
	var backendName, ollamaURL string
	var think bool
	flag.StringVar(&c.model, "model", envOr("ANTHROPIC_MODEL", defaultModel), "model id")
	flag.StringVar(&backendName, "backend", "anthropic", "LLM backend: anthropic|ollama")
	flag.StringVar(&ollamaURL, "ollama-url", envOr("OLLAMA_HOST", ollama.DefaultBaseURL), "ollama server base URL")
	flag.BoolVar(&think, "think", false, "ollama: emit the model's reasoning block (default off; on needs a higher token cap)")
	var speakers bool
	var embedModel, oracleFile, manifestFile, makeManifest, refsFile, backfillChain string
	flag.StringVar(&manifestFile, "manifest", "", "MANIFEST.md of held documents; a claim citing a doc not in it is 'unverifiable' (no model call)")
	flag.StringVar(&makeManifest, "make-manifest", "", "scan this sources root, write <root>/MANIFEST.md, and exit")
	flag.StringVar(&refsFile, "refs", "", "claims-machine.txt whose ref=§ per claim id annotates each rendered leaf with its report section")
	flag.StringVar(&backfillChain, "backfill-passages", "", "no-model: rewrite this chain JSONL in place, adding quote_passages by verbatim-matching each quote against the record's own passages (needs -source); then exit")
	flag.StringVar(&c.retrieveMode, "retrieve", "bm25", "per-claim passage retrieval: bm25|none|oracle (none sends the full corpus; oracle reads -oracle)")
	flag.StringVar(&oracleFile, "oracle", "", "oracle retrieval: JSON map of claim-id → [passage-id]; the judge sees exactly those passages")
	flag.IntVar(&c.maxTokens, "max-tokens", retrieveTokenCap, "retrieval token budget per claim (fused bm25+embed ranking)")
	flag.Float64Var(&c.floor, "floor", 0, "retrieval floor: top passage cosine below this ⇒ verdict absent, no model call (0=off)")
	flag.BoolVar(&c.embed, "embed", true, "add the nomic-embed-text ranker fused with BM25 by RRF (needs local ollama or a committed cache)")
	flag.StringVar(&embedModel, "embed-model", embed.DefaultModel, "embedding model for the semantic ranker")
	flag.BoolVar(&speakers, "speakers", false, "print the distinct speakers + roles found in -source, then exit")
	flag.StringVar(&src, "source", "", "transcript file or corpus dir → faithfulness mode")
	flag.StringVar(&axisList, "axis", "faithfulness", "with -source: axes to run, comma-list of faithfulness|substance (spec/SUBSTANCE-CORPUS.md)")
	flag.BoolVar(&ev, "evidence", false, "evidence-grounding mode (web search)")
	flag.BoolVar(&audit, "audit", false, "run all three modes and emit a cross-tab (needs -source)")
	flag.StringVar(&text, "text", "", "inline input instead of a file")
	flag.BoolVar(&c.asMarkdown, "md", false, "emit markdown tables")
	flag.BoolVar(&c.verbose, "v", false, "verbose: show every API call")
	flag.BoolVar(&c.verbose, "verbose", false, "verbose: show every API call")
	flag.BoolVar(&c.noColor, "no-color", false, "disable ANSI colour")
	flag.IntVar(&c.maxRounds, "max-rounds", 2, "producer-critic rounds per claim (substance)")
	flag.BoolVar(&c.ceScoped, "ce-scoped", false, "substance: scope the Counterexample axis (variant prompt; default off) — fatal only for an in-scope instance of a universal claim; a constructed hypothetical only weakens; an observed-outcome counterexample is left to grounding")
	flag.BoolVar(&c.narrowBoundary, "narrowing-boundary", false, "substance: key the verdict on how far the surviving claim was narrowed, not axis severity (variant prompt; default off) — substantive = the claim as stated or trivially qualified, partial = materially narrower, hollow = no defensible core; whether the claim is TRUE is left to grounding")
	flag.BoolVar(&c.asOf, "as-of", false, "substance: judge the claim as of the date it was made (variant prompt; default off) — the critic may use only what a careful reader could have known then and must not cite a prediction's realised outcome; that belongs to the evidence axis")
	flag.IntVar(&c.maxClaims, "max-claims", 0, "bound evidence grounding (0 = unlimited)")
	flag.IntVar(&c.repeat, "n", 1, "repeat each claim N times, show modal verdict + agreement (faithfulness)")
	flag.BoolVar(&c.showProgress, "progress", true, "show per-claim progress on stderr (default on)")
	flag.BoolVar(&c.quiet, "quiet", false, "suppress per-case lines + heartbeat; keeps SUMMARY and writes Tier-2")
	flag.BoolVar(&c.fresh, "fresh", false, "faithfulness: ignore any existing chain and re-judge every claim (default resumes)")
	flag.StringVar(&c.usageOut, "usage-out", "", "append one JSON record per run to this file")
	flag.StringVar(&chainDir, "chain-dir", "", "directory for Tier-2 JSONL verification chain (default: eval/<stamp>/)")
	flag.BoolVar(&full, "full", false, "print the full table to stdout instead of the brief report")
	flag.Var(&treeF, "tree", "print the tree report to stdout; -tree=full expands every node")
	flag.StringVar(&fromChain, "from", "", "render brief/tree/audit from a saved chain JSONL (no model calls); a comma-list of chains merges them leaf-by-leaf")
	flag.StringVar(&c.argumentFile, "argument", "", "with -from: render the argument tree keyed by this argument.txt (spec/ARGUMENT.md) into index.html, instead of the section-path tree")
	flag.BoolVar(&c.review, "review", false, "with -argument -manifest: build a self-contained review site (site/index.html + site/review.html + site/sources/ with a copy of every linked PDF; links relative to site/) under the example dir — see spec/SERVE.md")
	flag.BoolVar(&c.edge, "edge", false, "with -from -argument: run the edge-level adversarial pass over in-scope F→R edges (one model call each, spec/EDGE.md) and roll the open/unchallenged verdict into the recommendations; off by default")
	flag.BoolVar(&c.zip, "zip", false, "with -review: also write site.zip beside site/ (the whole built site, one downloadable archive)")
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

	// -manifest: load the held-document set. A claim citing a document not in it is judged
	// "unverifiable" from code, no model call — distinct from "absent" (cited doc present, claim not
	// found in it). Loaded before -from too: the root block's "checked against M source documents" is
	// M = the held count, and reading a manifest costs nothing and needs no key.
	if manifestFile != "" {
		held, err := manifest.LoadHeld(manifestFile)
		if err != nil {
			fatal("load manifest: " + err.Error())
		}
		c.held = held
		c.sourceDocs = len(held)
		labels, err := manifest.LoadLabels(manifestFile)
		if err != nil {
			fatal("load manifest labels: " + err.Error())
		}
		c.docLabels = labels                      // canonical doc id → witness/author, for a rendered quote's provenance
		c.sourcesDir = filepath.Dir(manifestFile) // MANIFEST.md sits at sources/; its dir is the href prefix
		reports, err := manifest.LoadReports(manifestFile)
		if err != nil {
			fatal("load manifest report-link rules: " + err.Error())
		}
		c.reports = reports // one report excerpt per PDF; ReportFor routes each leaf to its own by claim-id prefix
		single, err := manifest.LoadSingleSource(manifestFile)
		if err != nil {
			fatal("load manifest single_source: " + err.Error())
		}
		c.singleSource = single // report is its own only source: leaf 'absent' renders 'uncorroborated'
	}

	// -review builds the PDF-linked review page from the rendered argument page; both links and the
	// left iframe need the sources tree the manifest lives in, so it only runs alongside -argument and
	// -manifest.
	if c.review && (c.argumentFile == "" || manifestFile == "") {
		fatal("-review needs -argument (the tree to link) and -manifest (the sources tree the links open)")
	}
	// -zip archives the built site, so it means nothing without -review to build one.
	if c.zip && !c.review {
		fatal("-zip needs -review (there is no site to archive without it)")
	}

	// -refs: load the report section (§) per claim id from claims-machine.txt, so each rendered leaf
	// names the section it rests on. No model, no key; used only on the render paths.
	if refsFile != "" {
		refs, err := loadRefs(refsFile)
		if err != nil {
			fatal("load refs: " + err.Error())
		}
		c.refs = refs
	}

	// -backfill-passages rewrites a saved chain in place, adding each quote's passage id by matching it
	// against the record's own passages — no model call. It needs the corpus (-source), then exits.
	if backfillChain != "" {
		c.backfillPassages(backfillChain, src)
		return
	}

	// -from replays a saved chain with no model calls, so it renders straight from the JSONL and
	// returns. The one exception is -edge: the edge-level adversarial pass (spec/EDGE.md) makes one
	// model call per in-scope F→R edge, so -from -edge wires a backend + key first and writes its own
	// chain under testing/chains/. Without -edge the render is unchanged and needs no key.
	if fromChain != "" {
		c.usage = newUsageCounters()
		if c.edge {
			if c.argumentFile == "" {
				fatal("-edge needs -argument (the tree whose F→R edges it attacks)")
			}
			c.wireBackend(backendName, ollamaURL, think)
			c.edgeChainDir = chainDir // "" ⇒ the edge pass defaults to testing/chains/edge-<stamp>-<model>
		}
		c.runFromChain(fromChain, flag.Arg(0))
		return
	}

	// -make-manifest scans a sources root and writes MANIFEST.md, then exits. No model, no key.
	if makeManifest != "" {
		docs, err := manifest.Scan(makeManifest)
		if err != nil {
			fatal("scan sources: " + err.Error())
		}
		if len(docs) == 0 {
			fatal("no documents found under " + makeManifest)
		}
		out := filepath.Join(makeManifest, "MANIFEST.md")
		if err := os.WriteFile(out, []byte(manifest.Render(docs)), 0o644); err != nil {
			fatal("write manifest: " + err.Error())
		}
		fmt.Fprintf(os.Stderr, "wrote %s (%d documents)\n", out, len(docs))
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

	c.wireBackend(backendName, ollamaURL, think)
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
	// -axis adds the substance axis to the -source corpus run (spec/SUBSTANCE-CORPUS.md). It is inert
	// outside that path: the single-file modes select their axis directly.
	for _, a := range strings.Split(axisList, ",") {
		if strings.EqualFold(strings.TrimSpace(a), "substance") {
			c.substanceOn = true
		}
	}
	if chainDir != "" {
		c.chainFile = filepath.Join(chainDir, fixtureName+"."+modeSuffix+".jsonl")
		c.auditPath = filepath.Join(chainDir, "audit.md")
		c.treeHTMLPath = filepath.Join(chainDir, "tree.html")
		if c.substanceOn && src != "" {
			c.substanceFile = filepath.Join(chainDir, fixtureName+".substance.jsonl")
		}
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
		ix, err := loadCorpusIndex(src)
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

// ── serve ────────────────────────────────────────────────────────────────────

// runServe starts a loopback static file server over a self-contained review site and opens
// index.html at its root — the argument-tree page, which links review.html and the source report.
// The site links into its own `sources/` tree with root-relative hrefs
// (`sources/report.pdf?p=53#page=53`), so serving the site directory whole puts index.html at
// /index.html and every link at /sources/…. Serving over HTTP — rather than opening the file://
// path — is what lets a PDF viewer honour the `#page=N` fragment and load the `sources/…` iframe
// target. `<dir>` is the site built by `assay -review` (spec/SERVE.md).
func runServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	var (
		port   int
		noOpen bool
	)
	fs.IntVar(&port, "port", 8080, "listen port on 127.0.0.1")
	fs.BoolVar(&noOpen, "no-open", false, "print and serve, but don't open a browser")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: assay serve [-port N] [-no-open] <site-dir>")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		fs.Usage()
		os.Exit(2)
	}
	dir := fs.Arg(0)
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		fatal("serve: not a directory: " + dir)
	}

	// index.html sits at the site root; warn but keep serving if it is missing, so a site whose
	// review.html is still reachable directly works even without the argument-tree page.
	indexURL := fmt.Sprintf("http://127.0.0.1:%d/index.html", port)
	if _, err := os.Stat(filepath.Join(dir, "index.html")); err != nil {
		fmt.Fprintf(os.Stderr, "warning: index.html not found under %s; serving anyway\n", dir)
	}

	// Bind before printing, so a taken port fails now rather than after we claim a URL that never
	// answers. Loopback only — this serves local render output, not the network.
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		fatal("serve: " + err.Error())
	}

	fmt.Println(indexURL)
	if !noOpen {
		if err := openBrowser(indexURL); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not open browser: %v\n", err)
		}
	}
	if err := http.Serve(ln, serveHandler(dir)); err != nil {
		fatal("serve: " + err.Error())
	}
}

// serveHandler is the static file handler rooted at `dir`, split out so a test can drive it through
// httptest without binding a port or opening a browser. http.Dir confines every request to `dir`;
// a path escaping it (../) is rejected by the FileServer, not served.
func serveHandler(dir string) http.Handler { return http.FileServer(http.Dir(dir)) }

// openBrowser opens `url` in the platform default browser: `open` on macOS, `xdg-open` on Linux,
// `cmd /c start` on Windows. A missing opener returns an error for the caller to downgrade to a
// warning — the server is useful without a browser having launched.
func openBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "windows":
		return exec.Command("cmd", "/c", "start", "", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
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
		"faithful", "partial", "overstated", "absent", "contradicted", "unsupported",
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

	schemaRetries, quoteRejects, noQuoteDowngrades := u.extras()

	fmt.Fprintf(os.Stderr, "\nSUMMARY %s %s\n", fixture, mode)
	fmt.Fprintf(os.Stderr, "  cases %d · verified %d · errored %d\n", total, verified, errored)
	fmt.Fprintf(os.Stderr, "  %s\n", strings.Join(parts, " · "))
	fmt.Fprintf(os.Stderr, "  wall %s · est_usd %s · cache_hit %.0f%% (read %d / created %d)\n",
		elapsed.Round(time.Second), cost, 100*u.cacheHitRate(), cr, cc)
	fmt.Fprintf(os.Stderr, "  schema_retries %d · quote_rejects %d · no_quote_downgrades %d · plain_retries %d\n",
		schemaRetries, quoteRejects, noQuoteDowngrades, u.plainRetriesN())
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

// claimLine is one parsed claim from a claims file: its id, its §-heading path, its text, and an
// optional cites field (document ids the finding rests on, checked against the manifest).
type claimLine struct{ id, path, text, cites, route string }

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
	unverif := make([][]string, len(raw)) // per claim: cited doc ids not in the manifest, if any
	for i, item := range raw {
		id, path, text, cites, route := parseClaimLine(item)
		if id == "" {
			id = fmt.Sprintf("c%d", i+1)
		}
		parsed[i] = claimLine{id: id, path: path, text: text, cites: cites, route: route}
		// A claim whose cited truth-maker is not held cannot be checked: mark it unverifiable and skip
		// retrieval and the judge entirely.
		if miss := c.missingCites(cites); len(miss) > 0 {
			unverif[i] = miss
			continue
		}
		claimPassages[i], below[i] = c.passagesForClaim(id, text, path, cites, srcPath, &fullSrc)
		if !below[i] {
			eligible = append(eligible, i)
		}
	}

	emit := func(i int, chosen faith, spread string, samples, pids []string, t time.Time) {
		results[i] = chosen
		rec := faithChainRecord(i, len(raw), parsed[i].text, chosen, t)
		rec.Spread = spread
		rec.Samples = samples
		rec.Passages = pids
		rec.Route = parsed[i].route
		c.appendChain(rec)
		c.progressDone(i, len(raw), chosen.Verdict, parsed[i].text, t)
		_, dissent := spreadFromSamples(samples)
		rows[i] = brief.Row{ID: parsed[i].id, Path: parsed[i].path, Text: parsed[i].text,
			Faith: chosen.Verdict, FaithReason: chosen.Evidence, Spread: spread, Dissent: dissent,
			Gap: chosen.Gap, SoWhat: chosen.SoWhat, Route: parsed[i].route, Section: c.refs[parsed[i].id]}
		details[parsed[i].id] = tree.Leaf{Reason: chosen.Evidence, Quotes: buildLeafQuotes(chosen.Quotes, chosen.QuotePassages, c.docLabels, c.reports)}
		vs[i] = chosen.Verdict
	}

	// reuse reconstructs a claim's result from an existing chain record — no model call, no re-append —
	// so a resumed run keeps the work an interrupted one already did.
	reuse := func(i int, rec chainRecord) {
		var fd faithDetail
		_ = json.Unmarshal(rec.Detail, &fd)
		results[i] = faith{Claim: rec.Claim, Verdict: rec.Verdict, Evidence: fd.CriticFinding,
			ReportSays: fd.ReportSays, SourceSays: fd.SourceSays, Gap: fd.Gap, SoWhat: fd.SoWhat,
			Quotes: fd.Quotes, QuoteSources: fd.QuoteSources}
		if c.tally != nil {
			c.tally.record(rec.Verdict)
		}
		spread, dissent := spreadForRecord(rec)
		rows[i] = brief.Row{ID: parsed[i].id, Path: parsed[i].path, Text: parsed[i].text,
			Faith: rec.Verdict, FaithReason: fd.CriticFinding, Spread: spread, Dissent: dissent,
			Gap: fd.Gap, SoWhat: fd.SoWhat, Route: routeOr(rec.Route, parsed[i].route),
			Section: c.refs[parsed[i].id]}
		details[parsed[i].id] = tree.Leaf{Reason: fd.CriticFinding, Quotes: buildLeafQuotes(fd.Quotes, fd.QuotePassages, c.docLabels, c.reports)}
		vs[i] = rec.Verdict
	}

	// Resume: unless -fresh, reuse any claim already judged in the chain so an interrupted run (an OOM
	// kill mid-group) continues instead of re-paying. Grouping completes claims out of idx order, so
	// the chain may have gaps — read whatever records are present.
	done := map[int]bool{}
	if c.chainFile != "" {
		if c.fresh {
			_ = os.Truncate(c.chainFile, 0)
		} else {
			for idx, rec := range readChainSparse(c.chainFile) {
				if 0 <= idx && idx < len(raw) {
					reuse(idx, rec)
					done[idx] = true
				}
			}
			if len(done) > 0 {
				fmt.Fprintf(os.Stderr, "[resume] %d/%d claims already in %s; judging the rest\n",
					len(done), len(raw), c.chainFile)
			}
		}
	}

	// Unverifiable claims: a cited document is not in the manifest, so-what = fetch it. No model call.
	for i, miss := range unverif {
		if len(miss) > 0 && !done[i] {
			list := strings.Join(miss, ", ")
			emit(i, faith{Verdict: "unverifiable",
				Evidence: "cited document(s) not in corpus: " + list, SoWhat: "fetch " + list},
				"manifest", nil, nil, time.Now())
		}
	}

	// Floor-suppressed claims: "absent" from code, no model call.
	for i, isBelow := range below {
		if isBelow && !done[i] {
			emit(i, faith{Verdict: "absent", Evidence: "no passage cleared the retrieval floor (retrieved=0)"},
				"floor", nil, nil, time.Now())
		}
	}

	// Grouped judge calls: within a group every claim sees the same union prefix, so calls after the
	// first read the cache instead of re-billing the passages.
	for _, group := range groupBySharedPassages(claimPassages, eligible) {
		shared := unionPassages(claimPassages, group)
		for _, i := range group {
			if done[i] {
				continue
			}
			// Exclude the claim's own paragraph here, on the shared union, not before grouping: the
			// union re-pools passages across the group, so a claim's source paragraph re-enters via a
			// group-mate's retrieval unless it is dropped from what THIS claim is judged against. A no-op
			// for a non-report corpus (dropOwnParagraph keeps every non-report passage), so group cache
			// reuse is unaffected there; for the report corpus, same-page claims still share a prefix
			// that differs only by each claim's own dropped paragraph.
			// The claim's own excerpt (which report PDF its leaf routes to) scopes the self-exclusion, so a
			// two-excerpt corpus drops the source paragraph from the right one; "" for a single-report or
			// no-manifest run, which matches any excerpt.
			excerpt := ""
			if rep, ok := manifest.ReportFor(c.reports, parsed[i].id); ok {
				excerpt = reportStem(rep.File)
			}
			ps := dropOwnParagraph(parsed[i].text, excerpt, pageOfPath(parsed[i].path), shared)
			// Cite-scoping must survive the group union too: a claim citing an external document retrieved
			// only that document's passages, but the union re-pools report and other-doc passages from its
			// group-mates' whole-corpus retrieval, so the report restatement re-enters unless it is dropped
			// from what THIS claim is judged against (the same hazard dropOwnParagraph handles for the
			// self-paragraph). Restrict the union to the cited bases (spec/TREE.md § Cite-scoped retrieval).
			if bases := c.citedExternalBases(parsed[i].cites); len(bases) > 0 {
				ps = keepBases(ps, bases)
			}
			t := c.progressStart(i, len(raw), "faithfulness")
			chosen, spread, samples := c.faithJudgeRepeat(parsed[i].text, ps)
			emit(i, chosen, spread, samples, retrieve.IDs(ps), t)
		}
	}
	if c.substanceOn {
		c.runSubstanceAxis(parsed, rows)
	}
	c.present(rows, tally(vs), mdFaith(results), func() { c.termFaith(results) }, details)
}

// runSubstanceAxis runs the substance pass (assayClaim, source-independent) over every corpus leaf and
// writes a second `.substance.jsonl` beside the faithfulness chain, filling each row's Substance verdict
// so the argument tree can roll a faithful-but-hollow leaf up as `weakened` (spec/SUBSTANCE-CORPUS.md).
// It resumes from an existing substance chain the same way the faithfulness pass does — reuse a leaf
// already judged unless -fresh — so an interrupted run does not re-pay. `rows` is indexed 1:1 with
// `parsed`; a row already carrying the reused verdict is mutated in place.
func (c *cfg) runSubstanceAxis(parsed []claimLine, rows []brief.Row) {
	done := map[int]bool{}
	if c.substanceFile != "" {
		if c.fresh {
			_ = os.Truncate(c.substanceFile, 0)
		} else {
			for idx, rec := range readChainSparse(c.substanceFile) {
				if 0 <= idx && idx < len(rows) {
					var sd substanceDetail
					_ = json.Unmarshal(rec.Detail, &sd)
					rows[idx].Substance, rows[idx].SubstanceReason = rec.Verdict, sd.Reason
					done[idx] = true
				}
			}
		}
	}
	for i := range parsed {
		if done[i] {
			continue
		}
		t := c.progressStart(i, len(parsed), "substance")
		s := c.assayClaim(parsed[i].text)
		c.appendChainTo(c.substanceFile, substanceChainRecord(i, len(parsed), parsed[i].text, s, t))
		reason := s.Reason
		if reason == "" && s.SurvivingClaim != "" {
			reason = "survives as: " + s.SurvivingClaim
		}
		rows[i].Substance, rows[i].SubstanceReason = s.Verdict, reason
		c.progressDone(i, len(parsed), s.Verdict, parsed[i].text, t)
	}
}

// substanceSibling maps a faithfulness chain path to its substance sibling: the same file with the
// `.faithfulness.jsonl` suffix swapped for `.substance.jsonl` (spec/SUBSTANCE-CORPUS.md). A path that
// does not carry the faithfulness suffix yields "" — no sibling to look for.
func substanceSibling(faithChain string) string {
	const suf = ".faithfulness.jsonl"
	if !strings.HasSuffix(faithChain, suf) {
		return ""
	}
	return strings.TrimSuffix(faithChain, suf) + ".substance.jsonl"
}

// attachSubstanceOverlay fills each row's Substance verdict/reason from a substance chain, matched by
// idx and guarded on the claim text so a mismatched sibling never mislabels a leaf. A missing file or a
// text mismatch leaves the rows untouched — the overlay is optional (spec/SUBSTANCE-CORPUS.md).
func attachSubstanceOverlay(rows []brief.Row, texts []string, path string) {
	if path == "" {
		return
	}
	for idx, rec := range readChainSparse(path) {
		if idx < 0 || idx >= len(rows) || strings.TrimSpace(rec.Claim) != strings.TrimSpace(texts[idx]) {
			continue
		}
		var sd substanceDetail
		_ = json.Unmarshal(rec.Detail, &sd)
		reason := sd.Reason
		if reason == "" && sd.SurvivingClaim != "" {
			reason = "survives as: " + sd.SurvivingClaim
		}
		rows[idx].Substance, rows[idx].SubstanceReason = rec.Verdict, reason
	}
}

// readChainSparse reads whatever records a chain file holds, keyed by idx, tolerating gaps (a resumed
// faithfulness run completes claims out of idx order, so the partial chain need not be contiguous). A
// missing or unreadable file yields an empty map — a fresh start.
func readChainSparse(path string) map[int]chainRecord {
	out := map[int]chainRecord{}
	b, err := os.ReadFile(path)
	if err != nil {
		return out
	}
	for _, ln := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		if ln = strings.TrimSpace(ln); ln == "" {
			continue
		}
		var r chainRecord
		if json.Unmarshal([]byte(ln), &r) == nil && r.Mode == "faithfulness" {
			out[r.Idx] = r
		}
	}
	return out
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
func (c cfg) passagesForClaim(id, text, path, cites, srcPath string, fullSrc *[]retrieve.Passage) (ps []retrieve.Passage, below bool) {
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
	// A claim that cites a held document OTHER than the report is grounded against those documents
	// alone, with the report excluded (spec/TREE.md § Cite-scoped retrieval): the report restating a
	// benchmark figure cannot corroborate a claim about that figure, so retrieval draws from the cited
	// leaderboard/paper. A claim citing only the report (or nothing) is judged against the report as
	// before (single-source rules).
	if bases := c.citedExternalBases(cites); len(bases) > 0 {
		res := c.index.RetrieveFrom(q, c.maxTokens, c.floor, bases)
		if res.Below {
			return nil, true
		}
		return res.Passages, false
	}
	res := c.index.Retrieve(q, c.maxTokens, c.floor)
	if res.Below {
		return nil, true
	}
	return res.Passages, false
}

// citedExternalBases maps a claim's cites field to the passage-id bases (retrieve.passageBase) of the
// cited documents that are NOT the report — the set RetrieveFrom restricts to. It is empty when the
// claim cites only report excerpts (or nothing), the signal to judge against the report itself. A report
// cite is any id matching a manifest report excerpt's PDF filename (`report.pdf`,
// `report-productivity.pdf`); the two external shapes this corpus uses map to their passage bases:
// `paper:<stem>` → `papers/<stem>` and a `<dir>/<stem>.txt` leaderboard id → `<dir>/<stem>`.
func (c cfg) citedExternalBases(cites string) map[string]bool {
	reportFiles := map[string]bool{}
	for _, r := range c.reports {
		reportFiles[r.File] = true
	}
	bases := map[string]bool{}
	for _, id := range citedDocs(cites) {
		switch {
		case reportFiles[id]:
			continue
		case strings.HasPrefix(id, "paper:"):
			bases["papers/"+strings.TrimPrefix(id, "paper:")] = true
		case strings.HasSuffix(id, ".txt"):
			bases[strings.TrimSuffix(id, ".txt")] = true
		}
	}
	return bases
}

// dropOwnParagraph removes, from a report claim's retrieved passages, the ONE passage the claim was
// decomposed from — its own paragraph — so the claim is judged against the REST of the report rather
// than the sentence it was lifted from, which would confirm every claim trivially. The source paragraph
// is the report passage on the claim's own printed page (its §) that shares the most distinct words with
// the claim: an atomic claim overlaps the paragraph it came from more than any other, and word overlap
// finds it where a substring test cannot (a decomposed claim drops clauses and punctuation, so it is
// rarely a verbatim substring). Paragraph granularity, not the whole page, so a qualifier that sits in
// the same § but a DIFFERENT paragraph (the K37 shape) survives and can still ground the claim — the
// page-wide exclusion this replaces dropped every same-§ passage, so a fact restated one paragraph over
// read `absent`. It still keeps the corpus's cross-page restatement design (sources/MANIFEST.md): a
// fact stated on p2 and restated on p4 corroborates itself across the boundary. With no restatement
// anywhere the source paragraph is the only match dropped and the rest neither pin the claim down nor
// repeat it, so the verdict is `uncorroborated` (absent), never a distortion — spec/TREE.md § Single-
// source judging has a direction. It fires only for report-as-its-own-source passages (the report id
// shape, reportExcerpt) and only when the claim's page is known (>0): for a hearing/submission corpus
// the passage that carries a claim is the grounding target and must be kept. `claimExcerpt` scopes the
// exclusion to the claim's own report excerpt (manifest.ReportFor by its claim id), so a two-excerpt
// corpus drops the source paragraph from the RIGHT PDF and never a same-page paragraph of the other;
// "" (no excerpt resolved) matches any excerpt, the single-report behaviour.
func dropOwnParagraph(claim, claimExcerpt string, claimPage int, ps []retrieve.Passage) []retrieve.Passage {
	if claimPage <= 0 {
		return ps
	}
	claimTerms := map[string]bool{}
	for _, w := range runTokens(claim) {
		claimTerms[w] = true
	}
	own, ownScore := -1, 0
	for i, p := range ps {
		stem, page, isReport := reportExcerpt(p.ID)
		if !isReport || page != claimPage || (claimExcerpt != "" && stem != claimExcerpt) {
			continue
		}
		seen := map[string]bool{}
		n := 0
		for _, w := range runTokens(p.Text) {
			if claimTerms[w] && !seen[w] {
				seen[w] = true
				n++
			}
		}
		if n > ownScore {
			own, ownScore = i, n
		}
	}
	if own < 0 {
		return ps
	}
	// A fresh slice, not ps[:0]: the caller reuses one shared union across every claim in a group, so an
	// in-place filter would drop this claim's paragraph from what the NEXT claim is judged against.
	kept := make([]retrieve.Passage, 0, len(ps)-1)
	for i, p := range ps {
		if i == own {
			continue
		}
		kept = append(kept, p)
	}
	return kept
}

// keepBases returns the passages whose id base (retrieve.passageBase, the part before "#") is in
// `bases` — the cite-scoped subset of a group's shared union. First-seen order is preserved. Its
// counterpart is dropOwnParagraph: both filter the union down to what one claim may be judged against,
// dropOwnParagraph removing the claim's own paragraph, keepBases removing every document the claim did
// not cite (spec/TREE.md § Cite-scoped retrieval).
func keepBases(ps []retrieve.Passage, bases map[string]bool) []retrieve.Passage {
	kept := make([]retrieve.Passage, 0, len(ps))
	for _, p := range ps {
		if base, _, _ := strings.Cut(p.ID, "#"); bases[base] {
			kept = append(kept, p)
		}
	}
	return kept
}

// pageOfPath reads the printed page a claim sits on from the leading "p<N>=" segment of its §-path
// (claims-machine-full.txt encodes it there), returning 0 when the path carries no page — the signal
// dropOwnPage reads as "no self-page to exclude".
func pageOfPath(path string) int {
	seg := path
	if i := strings.IndexByte(seg, '/'); i >= 0 {
		seg = seg[:i]
	}
	key, _, _ := strings.Cut(seg, "=")
	if !strings.HasPrefix(key, "p") {
		return 0
	}
	n, err := strconv.Atoi(key[1:])
	if err != nil {
		return 0
	}
	return n
}

// reportExcerpt parses a report passage id "<excerpt>#p<N>#<n>" into its excerpt file stem and printed
// page, ok false for any other id shape. A report id alone carries a SECOND "#" (the paragraph ordinal
// after the page); a hearing "<date>/<session>#t<n>" or submission "submission-9#p3" has one, so
// dropOwnParagraph never mistakes one for a report page match. The stem is what scopes the exclusion to
// the claim's own excerpt in a multi-excerpt corpus ("report", "report-productivity").
func reportExcerpt(id string) (stem string, page int, ok bool) {
	base, frag, cut := strings.Cut(id, "#")
	if !cut || base == "" {
		return "", 0, false
	}
	pageTok, idx, cut := strings.Cut(frag, "#")
	if !cut || !strings.HasPrefix(pageTok, "p") {
		return "", 0, false
	}
	if _, err := strconv.Atoi(idx); err != nil {
		return "", 0, false
	}
	n, err := strconv.Atoi(pageTok[1:])
	if err != nil {
		return "", 0, false
	}
	return base, n, true
}

// reportStem is the excerpt file stem of a report PDF filename, the key reportExcerpt and CanonicalDoc
// share: "report-productivity.pdf" → "report-productivity". "" (no report claims the claim's id) stays
// "", which dropOwnParagraph reads as "match any excerpt".
func reportStem(file string) string { return strings.TrimSuffix(file, filepath.Ext(file)) }

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

// loadCorpusIndex builds the retrieval index from -source, which may be a comma-separated list of
// roots (e.g. the hearings dir and the submissions dir) merged into one index. A single root keeps the
// existing single-Load path.
func loadCorpusIndex(src string) (*retrieve.Index, error) {
	if strings.Contains(src, ",") {
		var paths []string
		for _, p := range strings.Split(src, ",") {
			if p = strings.TrimSpace(p); p != "" {
				paths = append(paths, p)
			}
		}
		return retrieve.LoadMany(paths)
	}
	return retrieve.Load(src)
}

// embedCacheDir is where a corpus's embedding caches live: a .embcache directory beside the corpus,
// so the committed vectors travel with the transcripts they were computed from. For a comma-separated
// -source it keys off the first root; the corpus hash keeps a combined corpus's cache distinct.
func embedCacheDir(src string) string {
	src = strings.TrimSpace(strings.SplitN(src, ",", 2)[0])
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
		criticSys := substanceCriticSys
		// The three variants are separate and mutually exclusive; none includes another's change.
		// Precedence when more than one flag is set: -narrowing-boundary, then -ce-scoped, then -as-of.
		// Each is the default prompt with one surgical change, so a run names exactly which produced its
		// verdicts.
		switch {
		case c.narrowBoundary:
			criticSys = substanceCriticSysNarrowingBoundary // verdict keys on narrowing distance, not axis severity
		case c.ceScoped:
			criticSys = substanceCriticSysCEScoped // default prompt + the Counterexample-scope rule
		case c.asOf:
			criticSys = substanceCriticSysAsOf // default prompt + the as-of rule, barring hindsight
		}
		if err := c.callJSON(criticSys, "", u, false, &last); err != nil {
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
	user := "SUMMARY CLAIM:\n" + claim
	byID := make(map[string]retrieve.Passage, len(passages))
	for _, p := range passages {
		byID[p.ID] = p
	}
	// ground verifies each cited quote is a verbatim substring of the passage it names — the grounding
	// check, no model call — returning the surviving quotes, their sources, and the reject count.
	ground := func(jj judgeJSON) (verified, sources, passageIDs []string, rejects int) {
		for _, e := range jj.Evidence {
			if p, ok := byID[e.PassageID]; ok && quoteInPassage(e.Quote, p) {
				verified = append(verified, e.Quote)
				sources = append(sources, p.Source)   // "hearing" | "submission", for the evidence-origin table
				passageIDs = append(passageIDs, p.ID) // the passage the quote was cited from, for provenance
			} else {
				rejects++ // a paraphrase presented as verbatim, or an id the judge invented
			}
		}
		return
	}

	// A single_source corpus judges the report against itself, so the "source" passages are other parts
	// of the same report — the direction rules below tell overstatement from a benign restatement, which
	// the hearing-transcript framing has no reason to. Off (the whole tree runs against held sources),
	// the base prompt is unchanged.
	sys := faithJudgeSys
	if c.singleSource {
		sys += faithJudgeSingleSourceRules
	}

	var j judgeJSON
	if err := c.callSchema(sys, cached, user, json.RawMessage(judgeSchema), "faith_verdict", temp, &j); err != nil {
		return faith{Claim: claim, Verdict: "error", Evidence: err.Error()}
	}
	verified, sources, passageIDs, rejects := ground(j)

	// report_says/source_says must be plain restatements the reader can act on (the check reads the
	// grounded quotes, so it runs after grounding). A violation gets one retry (counted), steered to
	// rephrase in everyday words; whatever the retry returns is kept.
	if reportSaysBad(j.ReportSays, claim, verified) || sourceSaysBad(j.SourceSays, claim, verified) {
		if c.usage != nil {
			c.usage.addPlainRetry()
		}
		steer := user + "\n\nYour report_says and source_says must not copy the report's or a quote's " +
			"wording, and source_says must not contrast or negate: rephrase in everyday words, as if to " +
			"someone who hasn't read the report."
		var j2 judgeJSON
		if err := c.callSchema(sys, cached, steer, json.RawMessage(judgeSchema), "faith_verdict", temp, &j2); err == nil {
			j = j2
			verified, sources, passageIDs, rejects = ground(j)
		}
	}
	if rejects > 0 && c.usage != nil {
		c.usage.addQuoteReject(rejects)
	}
	verdict := c.groundVerdict(j.Verdict, verified)
	// Assemble the stakes line from the FINAL verdict, so a grounding downgrade (contradicted→absent,
	// faithful/partial/overstated→unsupported) selects the matching template.
	soWhat := renderStakes(verdict, j.ReportSays, j.SourceSays)
	return faith{Claim: claim, Verdict: verdict, Evidence: j.Reason, ReportSays: j.ReportSays,
		SourceSays: j.SourceSays, Gap: j.Gap, SoWhat: soWhat, Quotes: verified, QuoteSources: sources,
		QuotePassages: passageIDs}
}

// renderStakes assembles the stakes line the tree leaf and root block show, from the judge's two
// plain restatements and the FINAL (post-grounding) verdict. partial/overstated/contradicted put the
// two halves side by side; absent/unsupported say no held source carries it; "faithful" (and any
// non-verdict) render nothing. An empty report_says yields "" — there is nothing to contrast.
func renderStakes(verdict, reportSays, sourceSays string) string {
	rs := strings.TrimSpace(reportSays)
	if rs == "" {
		return ""
	}
	switch verdict {
	case "partial", "overstated", "contradicted":
		ss := strings.TrimSpace(sourceSays)
		if ss == "" {
			return ""
		}
		return "The report says " + rs + ". The source only says " + ss + "."
	case "absent", "unsupported":
		return "The report says " + rs + ". No held source says this."
	default: // faithful, error, and any unknown verdict
		return ""
	}
}

// reportSaysBad reports whether report_says fails its discipline: a banned or imported word
// (plainBadWord), or a run of 3+ words lifted verbatim from the CLAIM — report_says is the summary's
// own framing restated, so copying the claim's wording is the failure. It is NOT checked against the
// quotes: echoing a source phrase in report_says is not the concern there.
func reportSaysBad(s, claim string, quotes []string) bool {
	return plainBadWord(s, claim, quotes) != "" || verbatimRun(s, []string{claim}, 3) != ""
}

// sourceSaysBad reports whether source_says fails its discipline: a banned or imported word
// (plainBadWord); a run of 5+ words lifted verbatim from a QUOTE — a short source phrase like "based
// in Victoria" is the source's own and belongs here, so only a long lift counts as copying; or a
// contrast/negation word (contrastWord). The contrast is renderStakes's job ("The source only says
// …"); the field states what the source says and nothing else.
func sourceSaysBad(s, claim string, quotes []string) bool {
	return plainBadWord(s, claim, quotes) != "" || verbatimRun(s, quotes, 5) != "" || contrastWord(s) != ""
}

// contrastWord returns the first contrast or negation term in s — "not", "only", "just", "instead",
// "but", or the phrase "rather than" — or "" when s carries none. source_says must state what the
// source supports without setting it against the report; the code frame supplies the contrast.
func contrastWord(s string) string {
	banned := map[string]bool{"not": true, "only": true, "just": true, "instead": true, "but": true}
	words := splitWords(s)
	for i, w := range words {
		if banned[w] {
			return w
		}
		if w == "rather" && i+1 < len(words) && words[i+1] == "than" {
			return "rather than"
		}
	}
	return ""
}

// verbatimRun returns the first run of n consecutive words in s that also appears, as a contiguous
// word sequence, in any of refs (case-insensitive, numbers kept as words) — or "" when s copies no
// such run. n is the shortest run that signals lifted phrasing rather than unavoidable shared-noun
// overlap; a longer copied run necessarily contains an n-word one, so checking n-grams suffices.
func verbatimRun(s string, refs []string, n int) string {
	if n < 1 {
		return ""
	}
	var refTokens [][]string
	for _, r := range refs {
		refTokens = append(refTokens, runTokens(r))
	}
	words := runTokens(s)
	for i := 0; i+n <= len(words); i++ {
		run := words[i : i+n]
		for _, ref := range refTokens {
			for j := 0; j+n <= len(ref); j++ {
				if slices.Equal(ref[j:j+n], run) {
					return strings.Join(run, " ")
				}
			}
		}
	}
	return ""
}

// runTokens lowercases text and splits it into alphanumeric words (digits kept, so "52 internal
// projects" is three words), keeping internal apostrophes. The verbatim-run check needs numbers as
// words — a copied "$80 million on 52" is exactly the lifted phrasing it looks for — where the
// syllable allowlist (splitWords) drops them.
func runTokens(text string) []string {
	var out []string
	for _, tok := range strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '\''
	}) {
		if w := strings.Trim(tok, "'"); w != "" {
			out = append(out, w)
		}
	}
	return out
}

// plainBadWord returns the first word in s that disqualifies it as a plain restatement — "reader" or
// "would" outright, or any word of four or more syllables that appears in neither the claim nor a
// cited quote — or "" when s is clean. The allowlist is the vocabulary the source itself used, so a
// long domain word ("Victoria", "production") passes while imported jargon ("unrepresentative",
// "criterion") does not.
func plainBadWord(s, claim string, quotes []string) string {
	allow := map[string]bool{}
	addWords := func(text string) {
		for _, w := range splitWords(text) {
			allow[w] = true
		}
	}
	addWords(claim)
	for _, q := range quotes {
		addWords(q)
	}
	for _, w := range splitWords(s) {
		if w == "reader" || w == "would" {
			return w
		}
		if syllables(w) >= 4 && !allow[w] {
			return w
		}
	}
	return ""
}

// splitWords lowercases text and splits it into letter-only words, keeping internal apostrophes so
// "creative's" is one word. It is the tokeniser both the restatement check and its allowlist share.
func splitWords(text string) []string {
	var out []string
	for _, tok := range strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && r != '\''
	}) {
		if w := strings.Trim(tok, "'"); w != "" {
			out = append(out, w)
		}
	}
	return out
}

// syllables estimates an English word's syllable count by counting vowel groups and dropping a silent
// trailing "e", floored at 1. It is a heuristic — the restatement check only needs the ≥4 boundary,
// where over-counting a rare word costs one retry, never a wrong verdict.
func syllables(word string) int {
	n, prevVowel := 0, false
	for _, r := range word {
		v := strings.ContainsRune("aeiouy", r)
		if v && !prevVowel {
			n++
		}
		prevVowel = v
	}
	if strings.HasSuffix(word, "e") && n > 1 {
		n--
	}
	if n == 0 {
		return 1
	}
	return n
}

// groundVerdict enforces that every verdict except "absent" is backed by at least one verified quote.
// With none surviving the grounding check: "contradicted" (which asserts the source says the opposite)
// downgrades to "absent"; "faithful"/"partial"/"overstated" (which assert the source supports some
// version of the claim) downgrade to "unsupported" — no verdict, routed to needs-you (brief.Qualify
// tier 0). Each downgrade is counted in usage. "absent" needs no quote (it asserts nothing is there);
// "error"/"unsupported" are already non-verdicts. The rule lives in code, not the prompt, so a fluent
// but ungrounded judgment cannot talk its way to a verdict.
func (c cfg) groundVerdict(verdict string, verified []string) string {
	if len(verified) > 0 {
		return verdict
	}
	switch verdict {
	case "contradicted":
		if c.usage != nil {
			c.usage.addNoQuoteDowngrade()
		}
		return "absent"
	case "faithful", "partial", "overstated":
		if c.usage != nil {
			c.usage.addNoQuoteDowngrade()
		}
		return "unsupported"
	}
	return verdict
}

// faithJudgeRepeat runs faithJudge c.repeat times and returns the modal result, a "k/N" agreement
// string ("" when N=1), and every sample's verdict in draw order (nil when N=1). N=1 is deterministic
// (temperature 0); N>1 samples at judgeSampleTemp so the spread reflects real judge instability, with
// the cached passage prefix reused across the repeats. The ordered list is what the chain persists so
// -from can rebuild the spread and name the dissent without re-running the judge.
func (c cfg) faithJudgeRepeat(claim string, passages []retrieve.Passage) (faith, string, []string) {
	n := c.repeat
	if n < 1 {
		n = 1
	}
	if n == 1 {
		zero := 0.0
		return c.faithJudge(claim, passages, &zero), "", nil
	}
	t := judgeSampleTemp
	counts := make(map[string]int, n)
	rep := make(map[string]faith, n)
	samples := make([]string, 0, n)
	for k := 0; k < n; k++ {
		r := c.faithJudge(claim, passages, &t)
		counts[r.Verdict]++
		samples = append(samples, r.Verdict)
		if _, seen := rep[r.Verdict]; !seen {
			rep[r.Verdict] = r
		}
	}
	modal := modalVerdict(counts)
	return rep[modal], fmt.Sprintf("%d/%d", counts[modal], n), samples
}

// spreadFromSamples derives the "k/N" agreement fraction for the modal verdict and the dissent — the
// minority verdicts in first-seen order, comma-joined, "" when unanimous — from the recorded list of
// sample verdicts. It reads the stored list rather than any pre-baked spread string, so -from renders
// exactly what the live run drew. A list shorter than 2 has no spread to show.
func spreadFromSamples(samples []string) (spread, dissent string) {
	if len(samples) < 2 {
		return "", ""
	}
	counts := make(map[string]int, len(samples))
	order := make([]string, 0, len(samples))
	for _, v := range samples {
		if _, seen := counts[v]; !seen {
			order = append(order, v)
		}
		counts[v]++
	}
	modal := modalVerdict(counts)
	var d []string
	for _, v := range order {
		if v != modal {
			d = append(d, v)
		}
	}
	return fmt.Sprintf("%d/%d", counts[modal], len(samples)), strings.Join(d, ", ")
}

// spreadForRecord gives the spread fraction and dissent a rendered row should show for a saved chain
// record. It rebuilds both from the recorded `samples` list when present, so -from names the dissent
// behind a split; a record written before `samples` existed falls back to the pre-baked `spread`
// string with no dissent — the back-compat path.
func spreadForRecord(rec chainRecord) (spread, dissent string) {
	if len(rec.Samples) > 0 {
		return spreadFromSamples(rec.Samples)
	}
	return rec.Spread, ""
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

// resolveQuoteProv renders a verified quote's provenance suffix from its passage id, using only the
// manifest labels — no corpus, so the render needs the chain plus MANIFEST.md and nothing more. It is
// "— <doc>, <witness>, <locator>" when the id maps to a held document, else the
// "(passage <id>, unresolved)" marker (the quote is never dropped). An empty id — the backfill could
// not pin the quote to exactly one passage — renders "(passage unresolved)". With no labels loaded
// (no -manifest) it returns "", so the leaf shows the bare quote unchanged.
func resolveQuoteProv(pid string, labels map[string]string, reports []manifest.Report) string {
	if len(labels) == 0 {
		return ""
	}
	if pid == "" {
		return "(passage unresolved)"
	}
	doc, loc, ok := manifest.CanonicalDoc(pid, reports)
	if !ok {
		return "(passage " + pid + ", unresolved)"
	}
	witness, held := labels[doc]
	if !held {
		return "(passage " + pid + ", unresolved)"
	}
	if witness == "" {
		return fmt.Sprintf("— %s, %s", doc, loc)
	}
	return fmt.Sprintf("— %s, %s, %s", doc, witness, loc)
}

// buildLeafQuotes pairs each verified quote with its resolved provenance suffix for a tree leaf.
// `passageIDs` is parallel to `quotes`; a missing entry (a shorter or absent list) leaves that quote
// with no id, which resolveQuoteProv renders as unresolved rather than dropping it.
func buildLeafQuotes(quotes, passageIDs []string, labels map[string]string, reports []manifest.Report) []tree.Quote {
	out := make([]tree.Quote, len(quotes))
	for i, q := range quotes {
		pid := ""
		if i < len(passageIDs) {
			pid = passageIDs[i]
		}
		out[i] = tree.Quote{Text: q, Prov: resolveQuoteProv(pid, labels, reports)}
	}
	return out
}

// review link/scan patterns. The three provenance forms are the resolveQuoteProv suffixes rendered
// into the argument page as plain text; each trailing (\n|<) is the delimiter the match stops at (Go
// has no look-ahead) and is re-emitted unchanged. subFilePDF pulls the submission number and
// attachment flag from a PDF filename so the redaction suffix and zero-padding need not be guessed.
// reReportLink captures a whole §-ref — the label (numbered "2.1.1" or named "Executive summary") in
// group 2 and its trailing printed page in group 3 — so reportPDFPage can resolve a page for either.
var (
	reReportLink = regexp.MustCompile(`report: (§([^<\n]+?) p(\d+))`)
	reLeafIDSpan = regexp.MustCompile(`<span class="id">([A-Za-z]+\d+)`)
	reSubLink    = regexp.MustCompile(`— (submission:(\d+)(/attachment-1)?), (.+?), p\.(\d+)(\n|<)`)
	reQonLink    = regexp.MustCompile(`— (qon:([a-z]+)/(\d{4}-\d{2}-\d{2})), (.+?), (\d{4}-\d{2}-\d{2}), p\.(\d+)(\n|<)`)
	reHearLink   = regexp.MustCompile(`— (hearing:[^\s,]+), (.+?), (line \d+)(\n|<)`)
	subFilePDF   = regexp.MustCompile(`^(\d+)\.(\d+)?-`)
	reStyleBlock = regexp.MustCompile(`(?s)<style>(.*?)</style>`)
	reBodyBlock  = regexp.MustCompile(`(?s)<body>(.*?)</body>`)
)

// reportPDFPage resolves a review report-link's PDF page from a §-ref label and its inline printed
// page, applying the manifest's rules. A numbered label ("2.1.1") is self-locating: its page is the
// inline one. A named label ("Executive summary") is looked up in the manifest's sections table, since
// the code cannot know where a named heading sits; a named label absent from the table returns
// ok=false so the ref is left unlinked rather than pointed at a guessed page. The printed page is then
// offset to the PDF page (report_page_offset), the single place the offset is applied.
func reportPDFPage(label string, inlinePrinted, offset int, sections map[string]int) (int, bool) {
	printed := inlinePrinted
	if !numberedSection(label) {
		p, ok := sections[strings.TrimSpace(label)]
		if !ok {
			return 0, false
		}
		printed = p
	}
	return printed + offset, true
}

// numberedSection reports whether a §-ref label is a section NUMBER (digits and dots, e.g. "2.1.1")
// rather than a section NAME ("Executive summary") — the two resolve their page differently.
func numberedSection(label string) bool {
	label = strings.TrimSpace(label)
	if label == "" {
		return false
	}
	for _, r := range label {
		if !('0' <= r && r <= '9') && r != '.' {
			return false
		}
	}
	return true
}

// renderReview turns the rendered argument page into review.html: a left <iframe name="doc"> PDF
// viewer and, on the right, the argument body with its plain-text report sections and quote
// provenances rewritten as links that drive the iframe (a bare `<a target="doc">`, so the page
// carries no JS). Each page link carries a distinct `?p=<page>` query BEFORE the `#page` fragment so
// Chrome reloads the iframe on every click — a fragment-only change leaves the URL identical and the
// PDF viewer does not re-fetch. Hearing PDFs carry no page anchor (the transcript locator is a turn
// index, not a page), so they link to the document. `srcPrefix` is the href prefix baked into the
// page (`sources`, so links are site-relative); `scanDir` is the on-disk sources tree whose
// submissions/*.pdf are scanned to recover each submission's real filename — kept separate so the
// links can point at the site while the scan reads the real corpus. `title` is the report name that
// becomes review.html's <title>. `reports` is the manifest's report excerpts: a claim §-ref opens the
// excerpt its claim-id prefix routes to (manifest.ReportFor), with that excerpt's page offset and
// named-section table — so a two-report corpus links each claim to its own PDF. It is the Go port of
// the retired current/build-review.py, and returns a one-line link tally for the caller to report.
func renderReview(page, srcPrefix, scanDir, title string, reports []manifest.Report) (out, counts string) {
	// submission number (+attachment flag) → pdf filename, scanned from disk so the redaction suffix
	// and zero-padding don't have to be guessed.
	type subKey struct {
		num int
		att bool
	}
	subMap := map[subKey]string{}
	entries, _ := os.ReadDir(filepath.Join(scanDir, "submissions"))
	for _, e := range entries {
		fn := e.Name()
		if !strings.HasSuffix(fn, ".pdf") {
			continue
		}
		if m := subFilePDF.FindStringSubmatch(fn); m != nil {
			n, _ := strconv.Atoi(m[1])
			subMap[subKey{n, m[2] == "1"}] = fn
		}
	}

	var nReport, nSub, nQon, nHear, nSubMissing int

	style := ""
	if m := reStyleBlock.FindStringSubmatch(page); m != nil {
		style = m[1]
	}
	body := ""
	if m := reBodyBlock.FindStringSubmatch(page); m != nil {
		body = m[1]
	}
	// The two-pane page exists to click the provenance links, so its cards start open — otherwise every
	// link is hidden behind a disclosure triangle. index.html keeps them closed; only review.html opens.
	body = strings.ReplaceAll(body, `<details class=`, `<details open class=`)

	// 1. claim § → its report excerpt's PDF at the page. Which excerpt a §-ref opens is decided by the
	// claim id of the leaf card the ref sits in (manifest.ReportFor), found as the nearest preceding
	// id span — so a multi-report corpus links each claim to its own PDF. The page comes from that
	// report's offset + named-section table (reportPDFPage), not from code; a named section with no
	// table entry, or an id no report claims, leaves the ref as plain text rather than guessing a page.
	idSpans := reLeafIDSpan.FindAllStringSubmatchIndex(body, -1)
	leafIDBefore := func(pos int) string {
		id := ""
		for _, s := range idSpans {
			if s[0] >= pos {
				break
			}
			id = body[s[2]:s[3]]
		}
		return id
	}
	var b1 strings.Builder
	prev := 0
	for _, loc := range reReportLink.FindAllStringSubmatchIndex(body, -1) {
		b1.WriteString(body[prev:loc[0]])
		prev = loc[1]
		whole := body[loc[2]:loc[3]]
		label := body[loc[4]:loc[5]]
		inline, _ := strconv.Atoi(body[loc[6]:loc[7]])
		rep, ok := manifest.ReportFor(reports, leafIDBefore(loc[0]))
		if ok {
			if pdfPage, pok := reportPDFPage(label, inline, rep.Offset, rep.Sections); pok {
				nReport++
				fmt.Fprintf(&b1, `report: <a href="%s/%s?p=%d#page=%d" target="doc">%s</a>`,
					srcPrefix, rep.File, pdfPage, pdfPage, whole)
				continue
			}
		}
		b1.WriteString(body[loc[0]:loc[1]]) // no report claims the id, or named section absent from the table
	}
	b1.WriteString(body[prev:])
	body = b1.String()

	// 2a. submission quote provenance → submission PDF by page.
	body = reSubLink.ReplaceAllStringFunc(body, func(s string) string {
		m := reSubLink.FindStringSubmatch(s)
		doc, num, att, witness, pageNo, delim := m[1], m[2], m[3], m[4], m[5], m[6]
		n, _ := strconv.Atoi(num)
		fn, ok := subMap[subKey{n, att == "/attachment-1"}]
		if !ok {
			nSubMissing++
			return s
		}
		nSub++
		href := fmt.Sprintf("%s/submissions/%s?p=%s#page=%s", srcPrefix, fn, pageNo, pageNo)
		return fmt.Sprintf(`— <a href="%s" target="doc">%s, %s, p.%s</a>%s`, href, doc, witness, pageNo, delim)
	})

	// 2b. qon quote provenance → qon PDF by page (id qon:abc/2025-03-21 → abc-2025-03-21.pdf).
	body = reQonLink.ReplaceAllStringFunc(body, func(s string) string {
		m := reQonLink.FindStringSubmatch(s)
		doc, org, date, witness, locDate, pageNo, delim := m[1], m[2], m[3], m[4], m[5], m[6], m[7]
		nQon++
		href := fmt.Sprintf("%s/qon/%s-%s.pdf?p=%s#page=%s", srcPrefix, org, date, pageNo, pageNo)
		return fmt.Sprintf(`— <a href="%s" target="doc">%s, %s, %s, p.%s</a>%s`, href, doc, witness, locDate, pageNo, delim)
	})

	// 2c. hearing quote provenance → hearing PDF (document only; the locator is a turn index, not a
	// page, so no ?p/#page and no per-click reload is needed — every quote in one hearing is one doc).
	body = reHearLink.ReplaceAllStringFunc(body, func(s string) string {
		m := reHearLink.FindStringSubmatch(s)
		doc, witness, line, delim := m[1], m[2], m[3], m[4]
		nHear++
		href := fmt.Sprintf("%s/hearings/%s.pdf", srcPrefix, strings.SplitN(doc, ":", 2)[1])
		return fmt.Sprintf(`— <a href="%s" target="doc">%s, %s, %s</a>%s`, href, doc, witness, line, delim)
	})

	nUnresolved := strings.Count(body, "(passage unresolved)")
	total := nReport + nSub + nQon + nHear

	// The left pane lands on the first report excerpt; every clicked link re-points it, and each report
	// excerpt is copied into site/sources/ because its links carry it (referencedPDFs).
	iframeSrc := "report.pdf"
	if len(reports) > 0 && reports[0].File != "" {
		iframeSrc = reports[0].File
	}

	out = fmt.Sprintf(`<!doctype html>
<html lang="en"><head><meta charset="utf-8"><title>%s</title>
<link rel="icon" href="favicon.svg" type="image/svg+xml">
<style>
html,body{height:100%%;margin:0}
.split{display:flex;height:100vh}
#doc{flex:1;min-width:0;height:100%%;border:0;border-right:1px solid #ccc}
.arg{flex:1;min-width:0;height:100%%;overflow:auto;box-sizing:border-box;padding:0 1rem}
%s
.arg a{color:#0645ad}
</style>
</head><body>
<div class="split">
<iframe id="doc" name="doc" src="%s/%s" title="source document"></iframe>
<div class="arg">
%s
</div>
</div>
</body></html>
`, html.EscapeString(title), style, srcPrefix, iframeSrc, body)

	counts = fmt.Sprintf("report=%d submission=%d qon=%d hearing=%d total=%d; unresolved=%d",
		nReport, nSub, nQon, nHear, total, nUnresolved)
	if nSubMissing > 0 {
		counts += fmt.Sprintf("; submission files missing=%d", nSubMissing)
	}
	return out, counts
}

// faviconSVG is the site mark: an argument-tree glyph — one node above three, edges fanning down — in
// the pages' text colour (#000). No brand and no text; it is the same tree the page draws, shrunk to
// a favicon. writeSite writes it into the site and both index.html and review.html link it
// (spec/SERVE.md).
const faviconSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="#000" stroke="#000" stroke-width="1.3" stroke-linecap="round">
<line x1="12" y1="5" x2="5" y2="19"/>
<line x1="12" y1="5" x2="12" y2="19"/>
<line x1="12" y1="5" x2="19" y2="19"/>
<circle cx="12" cy="5" r="2.4"/>
<circle cx="5" cy="19" r="2.4"/>
<circle cx="12" cy="19" r="2.4"/>
<circle cx="19" cy="19" r="2.4"/>
</svg>
`

// siteREADME is the reader's-manual dropped at the site root by writeSite: how to open the two pages,
// how to serve them over HTTP when the browser blocks the file:// PDF, and one line on what each page
// is. It ships inside the site (and the -zip archive) so an unzipped copy is self-explanatory with no
// reference back to this repo (spec/SERVE.md).
const siteREADME = "# Review site\n" + `
Two pages, self-contained. ` + "`index.html`" + ` is the argument tree — the thesis, the counts, and
every claim with its faithfulness verdict. ` + "`review.html`" + ` is that same tree in a two-pane
layout: the tree on the right, the source PDF on the left, so a claim's page link lands on the page
it rests on.

## Open it

1. Unzip, then open ` + "`index.html`" + ` in Chrome or Edge.
2. If ` + "`review.html`" + `'s left pane is blank, the browser is refusing the ` + "`file://`" + ` PDF —
   serve the folder over HTTP instead, then open ` + "`http://127.0.0.1:8080/review.html`" + `:

       python3 -m http.server 8080     # run inside this folder

   or, with the assay binary from the parent folder:

       assay serve site
`

// sourceCountsLine summarises the held corpus for the page's opening sentence — hearing transcripts,
// submissions and answers to questions on notice — counted from the manifest's held set by the same
// canonical-id prefixes manifest.Render groups by (manifest.go:123). An empty category is dropped so
// a corpus with no qon does not read "0 answers"; an empty set yields "" and the caller omits the
// sentence rather than claiming to hold nothing. presentArgument passes the result into ArgumentPage,
// keeping the corpus-specific counting out of the corpus-agnostic tree package.
func sourceCountsLine(held map[string]bool) string {
	var hearings, submissions, qon int
	for id := range held {
		switch {
		case strings.HasPrefix(id, "hearing:"):
			hearings++
		case strings.HasPrefix(id, "submission:"):
			submissions++
		case strings.HasPrefix(id, "qon:"):
			qon++
		}
	}
	var parts []string
	if hearings > 0 {
		parts = append(parts, fmt.Sprintf("%d hearing transcripts", hearings))
	}
	if submissions > 0 {
		parts = append(parts, fmt.Sprintf("%d submissions", submissions))
	}
	if qon > 0 {
		parts = append(parts, fmt.Sprintf("%d answers to questions on notice", qon))
	}
	return strings.Join(parts, ", ")
}

// reSitePDF matches a PDF the rendered site page reaches — an `href` link or the left iframe `src` —
// under the site-relative `sources/` prefix, capturing the path with its `?p=…#page=…` suffix
// stripped. That path is both the read source (under `sourcesDir`) and the write destination (under
// the site's `sources/`), which is what makes the site relocatable as a unit.
var reSitePDF = regexp.MustCompile(`(?:href|src)="sources/([^"?#]*\.pdf)`)

// buildSite writes the self-contained review site under `siteDir`: `review.html` and `argument.html`
// at its root and `sources/` holding a copy of every PDF `review.html` links to (report.pdf, and each
// hearing, submission and qon document — the four link classes). Every href is site-relative
// (`sources/…`, not `../sources/…`), so the tree serves over HTTP and moves as one unit; `assay serve
// <siteDir>` then answers `index.html` at the root. `page` is the rendered argument-tree page (the
// same page written as index.html); `sourcesDir` is the on-disk sources tree the copies are read
// from. `title` is the report name (ArgumentTitle) that becomes review.html's <title>. Returns
// renderReview's link tally plus the copied/missing PDF count. See spec/SERVE.md.
func buildSite(page, sourcesDir, siteDir, title string, reports []manifest.Report) (string, error) {
	review, linkCounts := renderReview(page, "sources", sourcesDir, title, reports)
	copied, missing, err := writeSite(review, page, sourcesDir, siteDir)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s; pdfs copied=%d missing=%d", linkCounts, copied, missing), nil
}

// writeSite copies every PDF `reviewHTML` references into `siteDir/sources/`, writes `pageHTML` as
// index.html (the argument-tree page verbatim) and `reviewHTML` as review.html at the site root, and
// writes favicon.svg and README.md beside them. It is split from buildSite so a refuter can drive it with a
// hand-written page and a fake sources tree, and assert every href resolves under the site — the
// property the whole site exists to hold. A PDF that cannot be copied is counted as missing and
// warned, never synthesized: a broken link is reported, not papered over.
func writeSite(reviewHTML, pageHTML, sourcesDir, siteDir string) (copied, missing int, err error) {
	if err = os.MkdirAll(siteDir, 0o755); err != nil {
		return 0, 0, err
	}
	for _, rel := range referencedPDFs(reviewHTML) {
		if cpErr := copyFile(filepath.Join(sourcesDir, rel), filepath.Join(siteDir, "sources", rel)); cpErr != nil {
			fmt.Fprintf(os.Stderr, "warning: site: cannot copy %s: %v\n", rel, cpErr)
			missing++
			continue
		}
		copied++
	}
	for _, f := range []struct{ name, body string }{
		{"index.html", pageHTML},
		{"review.html", reviewHTML},
		{"favicon.svg", faviconSVG},
		{"README.md", siteREADME},
	} {
		if err = os.WriteFile(filepath.Join(siteDir, f.name), []byte(f.body), 0o644); err != nil {
			return copied, missing, err
		}
	}
	return copied, missing, nil
}

// referencedPDFs returns the distinct `sources/`-relative PDF paths the site page links to or embeds,
// sorted for a stable copy order. Each path is the relative locator shared by the read source and the
// write destination — see reSitePDF.
func referencedPDFs(html string) []string {
	seen := map[string]bool{}
	var out []string
	for _, m := range reSitePDF.FindAllStringSubmatch(html, -1) {
		if !seen[m[1]] {
			seen[m[1]] = true
			out = append(out, m[1])
		}
	}
	sort.Strings(out)
	return out
}

// copyFile streams `src` to `dst`, creating dst's parent and truncating any existing dst so a rebuild
// is idempotent. It streams rather than reading whole because the source PDFs run to megabytes.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// zipSite writes `siteDir`.zip beside the built site: every file under `siteDir`, entry-named under the
// site's own base dir (site/index.html, …), so unzipping restores a `site/` folder `assay serve site`
// can serve. Called after writeSite when -zip is set. Returns the archive path and the file count so a
// zero-file archive — the empty-run failure the whole tool guards against — is visible, not silent.
func zipSite(siteDir string) (string, int, error) {
	zipPath := siteDir + ".zip"
	f, err := os.Create(zipPath)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	base := filepath.Base(siteDir)
	n := 0
	walkErr := filepath.Walk(siteDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, err := filepath.Rel(siteDir, path)
		if err != nil {
			return err
		}
		w, err := zw.Create(filepath.ToSlash(filepath.Join(base, rel)))
		if err != nil {
			return err
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		if _, err := io.Copy(w, in); err != nil {
			return err
		}
		n++
		return nil
	})
	if walkErr != nil {
		return "", n, walkErr
	}
	if err := zw.Close(); err != nil {
		return "", n, err
	}
	if n == 0 {
		return zipPath, 0, fmt.Errorf("zipSite: %s held no files to archive", siteDir)
	}
	return zipPath, n, nil
}

// loadRefs reads the per-claim report section from claims-machine.txt: the pipe-delimited "ref=§…"
// field, keyed by claim id, for the leaf's "report:" line. Comment (#) and pipe-less lines are
// skipped; a line with no ref= field contributes no entry.
func loadRefs(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	refs := map[string]string{}
	for _, ln := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(strings.TrimSpace(ln), "#") || !strings.Contains(ln, "|") {
			continue
		}
		fields := strings.Split(ln, "|")
		id := strings.TrimSpace(fields[0])
		if id == "" {
			continue
		}
		for _, f := range fields[1:] {
			if ref, ok := strings.CutPrefix(strings.TrimSpace(f), "ref="); ok {
				refs[id] = strings.TrimSpace(ref)
				break
			}
		}
	}
	return refs, nil
}

// backfillPassages rewrites a chain JSONL in place, adding each faithfulness record's quote_passages —
// the passage id every verified quote was cited from — with NO model call. It recovers, for chains
// judged before ground kept e.PassageID, the id the live judge now persists, by verbatim-substring-
// matching each quote against the record's OWN passages (loaded from -source, via the same quoteInPassage
// grounding check). A quote matching exactly one passage records that id; a quote matching zero or more
// than one is left "" (unresolved) — never guessed. Non-faithfulness chains and quote-less records pass
// through unchanged. The record is re-marshalled through chainRecord/faithDetail, whose field order the
// live writer already uses, so the on-disk diff is exactly the added quote_passages.
func (c *cfg) backfillPassages(chainPath, src string) {
	if src == "" {
		fatal("-backfill-passages needs -source (the corpus to match quotes against)")
	}
	ix, err := loadCorpusIndex(src)
	if err != nil {
		fatal("load corpus: " + err.Error())
	}
	byID := make(map[string]retrieve.Passage, len(ix.Passages))
	for _, p := range ix.Passages {
		byID[p.ID] = p
	}
	data, err := os.ReadFile(chainPath)
	if err != nil {
		fatal("read chain: " + err.Error())
	}
	var out []string
	var recs, quotes, resolved int
	for _, ln := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		if strings.TrimSpace(ln) == "" {
			continue
		}
		var rec chainRecord
		if err := json.Unmarshal([]byte(ln), &rec); err != nil {
			fatal("parse chain line: " + err.Error())
		}
		if rec.Mode == "faithfulness" {
			var det faithDetail
			if err := json.Unmarshal(rec.Detail, &det); err != nil {
				fatal("parse detail: " + err.Error())
			}
			if len(det.Quotes) > 0 {
				recs++
				det.QuotePassages = make([]string, len(det.Quotes))
				for i, q := range det.Quotes {
					quotes++
					pid := uniquePassageFor(q, rec.Passages, byID)
					det.QuotePassages[i] = pid
					if pid != "" {
						resolved++
					}
				}
				raw, _ := json.Marshal(det)
				rec.Detail = raw
			}
		}
		raw, err := json.Marshal(rec)
		if err != nil {
			fatal("marshal record: " + err.Error())
		}
		out = append(out, string(raw))
	}
	if quotes == 0 {
		fatal("backfill matched 0 quotes — not a faithfulness chain with quotes?") // zero-output rule
	}
	if err := os.WriteFile(chainPath, []byte(strings.Join(out, "\n")+"\n"), 0o644); err != nil {
		fatal("write chain: " + err.Error())
	}
	fmt.Fprintf(os.Stderr, "backfilled %s: %d records, %d quotes → %d resolved, %d unresolved\n",
		chainPath, recs, quotes, resolved, quotes-resolved)
}

// uniquePassageFor returns the id of the single passage among `passages` whose visible text contains
// `quote` verbatim (quoteInPassage). Zero or multiple matches return "" — the backfill records an
// unresolved marker rather than misattribute a quote that appears in two retrieved passages. A passage
// id absent from the corpus index is skipped.
func uniquePassageFor(quote string, passages []string, byID map[string]retrieve.Passage) string {
	hit, n := "", 0
	for _, pid := range passages {
		if p, ok := byID[pid]; ok && quoteInPassage(quote, p) {
			n++
			hit = pid
		}
	}
	if n == 1 {
		return hit
	}
	return ""
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
var faithVerdictOrder = []string{"unsupported", "contradicted", "absent", "unverifiable", "overstated", "partial", "faithful", "error"}

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
		Claim: claim, Verdict: e.Verdict, Finding: e.Finding, Horizon: e.Horizon,
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
//
// The horizon gate runs first and independently of retrieval: a claim whose truth is settled only
// in the future is a forecast, and no set of retrieved sources — however real and well-matched —
// can ground an outcome that has not happened. This is the axis boundary applied to time; reasoning
// cannot confirm an unobserved outcome. See spec/EVIDENCE.md.
func crossCheckEvidence(e evidence) evidence {
	if e.Horizon == "future" && e.Verdict != "error" {
		if e.Verdict != "unverifiable" {
			e.OriginalVerdict = e.Verdict
		}
		e.Verdict = "unverifiable"
		e.DowngradeReason = "forecast — projections are not evidence"
		return e
	}
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

// counterexampleScopeRule is the one rule the -ce-scoped variant adds to the Counterexample axis
// (spec: docs/todo/destructive-sonnet-2026-09-13.md § Fatal-axis anatomy). The default critic marks
// Counterexample fatal on any claim a case can be constructed against, which sinks well-formed
// tendencies, predictions and definitions ("a thermostat also adapts to feedback"); and it double-counts
// grounding by refuting a prediction with the world's actual outcome (the axis boundary in CLAUDE.md —
// only retrieval, not the armchair, may confirm how the world turned out). This block scopes the axis
// without touching any other. It is appended to the default prompt, so the variant is the default plus
// exactly this text.
const counterexampleScopeRule = `COUNTEREXAMPLE SCOPE — read this before scoring the Counterexample axis:
- A counterexample is "fatal" ONLY when the claim is UNIVERSAL in form (all / every / always / no /
  none, or an unqualified generalisation) AND the counterexample is an instance that falls WITHIN the
  claim's own stated scope. One such instance breaks a universal.
- If the claim is a TENDENCY, a PREDICTION, or a CONDITIONAL (not a universal), a CONSTRUCTED
  hypothetical counterexample only "weakens" — it never makes the verdict hollow on its own. A
  tendency is not refuted by exhibiting one imagined exception.
- A counterexample drawn from your OWN KNOWLEDGE OF HOW THE WORLD ACTUALLY TURNED OUT (an observed
  outcome, e.g. "the product later succeeded") is OUT OF SCOPE for this axis: whether a claim matches
  the world is the grounding pass's job, reached by retrieval, not yours to settle from memory. Note
  the observation in your finding if useful, but mark this axis "clears" and let grounding decide — do
  not record it as "fatal" or "weakens".
Every other axis is unchanged: Equivocation, Falsifiability, Hidden premise and the rest score exactly
as before, so an equivocating or unfalsifiable claim still collapses on its own axis.`

// substanceCriticSysCEScoped is the -ce-scoped variant: the default critic prompt plus the one
// Counterexample-scope rule, inserted just before the JSON contract so the schema line stays last.
var substanceCriticSysCEScoped = strings.Replace(
	substanceCriticSys, "\nReturn ONLY JSON:", "\n"+counterexampleScopeRule+"\n\nReturn ONLY JSON:", 1)

// substanceVerdictLinesDefault is the default critic's verdict block, keyed on axis severity — the
// exact needle the -narrowing-boundary variant replaces. The sonnet anatomy showed this boundary is
// unreachable for any general claim: "no equivocation, survives counterexample" cannot hold when
// Equivocation and Hidden-premise fire on every bare claim of the form, so `substantive` is never
// issued (docs/todo/destructive-sonnet-2026-09-13.md § Why substantive is never issued).
const substanceVerdictLinesDefault = `Then a verdict:
- "hollow": unfalsifiable, equivocating, or pure assertion with no defensible core.
- "partial": a narrower, qualified claim survives after stripping the unsupported parts.
- "substantive": falsifiable, evidence exists or is clearly obtainable, no equivocation, survives
counterexample.`

// narrowingBoundaryVerdicts is the verdict block the -narrowing-boundary variant swaps in. It keys the
// verdict on HOW FAR the surviving claim was narrowed rather than on which axes fired: a claim
// defensible as stated is `substantive` even when axes weaken it, and whether the claim is TRUE is left
// to grounding (the axis boundary in CLAUDE.md — only retrieval confirms how the world turned out). The
// critic states the surviving claim verbatim before the verdict it derives from it.
const narrowingBoundaryVerdicts = `First state the SURVIVING CLAIM verbatim — the strongest form of the
claim that withstands the axes above, narrowed no further than those findings force. THEN choose the
verdict by comparing that surviving claim to the claim AS STATED:
- "hollow": no defensible core — nothing survives that a reader could act on.
- "partial": a defensible claim survives, but it is materially narrower than the claim as stated.
- "substantive": the surviving claim IS the claim as stated, or the claim with only a trivial
qualification (a scope or unit the claim already implied). Decide this by how far the claim had to be
narrowed to defend it, NOT by whether an axis fired — a well-formed claim you happen to doubt is still
"substantive" here. Whether it is TRUE is the grounding pass's job, reached by retrieval, not settled on
this axis.`

// substanceCriticSysNarrowingBoundary is the -narrowing-boundary variant: every axis and its
// fatal/weakens/clears mark are unchanged; only the three verdict lines are replaced, and the JSON
// contract emits `surviving_claim` before `verdict` so the critic commits to the surviving claim first.
var substanceCriticSysNarrowingBoundary = strings.Replace(
	strings.Replace(substanceCriticSys, substanceVerdictLinesDefault, narrowingBoundaryVerdicts, 1),
	`"verdict":"substantive"|"partial"|"hollow","surviving_claim":string|null`,
	`"surviving_claim":string|null,"verdict":"substantive"|"partial"|"hollow"`, 1)

// asOfRule is the one rule the -as-of variant inserts before the axes. Intervention 2 item 2
// (docs/todo/destructive-sonnet-2026-09-13.md § Why false.txt is sunk) found the substance critic
// sinks a well-formed prediction by importing the realised outcome — on `false.txt` the fatal reason
// cites the actual 2010s shipment figures, the job CLAUDE.md's axis boundary reserves for grounding
// (reasoning may refute a self-contradiction but must never settle how the world turned out). This
// block bars that on predictions without touching any axis; it is spliced before the axis list so the
// critic reads it first.
const asOfRule = `Judge the claim as of the date it was made, using only what a careful reader could
have known then. Whether the prediction later came true is not a substance question; if you find
yourself citing the realised outcome, stop — that belongs to the evidence axis. Predictions are
judged on scope, falsifiability, and whether a mechanism is offered, not on hindsight.`

// substanceCriticSysAsOf is the -as-of variant: the default critic prompt with `asOfRule` inserted
// just before the first axis, so every axis and the verdict block are byte-unchanged and only the
// hindsight bar is added. A no-op strings.Replace (needle absent) would leave it equal to the default.
var substanceCriticSysAsOf = strings.Replace(
	substanceCriticSys, "- Evidence: is support", asOfRule+"\n\n- Evidence: is support", 1)

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
// gap, evidence[], report_says, source_says, then reason. All are required (Anthropic strict tools
// and the Ollama format both accept "" for the free-text fields), and evidence cites the source by
// passage id and a verbatim quote, so code can verify the quote without a second model call.
// report_says/source_says are the two ≤12-word plain restatements code assembles into the stakes
// line (renderStakes), replacing the old free-form so_what.
const judgeSchema = `{
  "type":"object",
  "properties":{
    "verdict":{"type":"string","enum":["faithful","partial","overstated","absent","contradicted"]},
    "gap":{"type":"string","enum":["none","scope","denominator","timerange","attribution","other"]},
    "evidence":{"type":"array","items":{"type":"object","properties":{"passage_id":{"type":"string"},"quote":{"type":"string"}},"required":["passage_id","quote"],"additionalProperties":false}},
    "report_says":{"type":"string"},
    "source_says":{"type":"string"},
    "reason":{"type":"string"}
  },
  "required":["verdict","gap","evidence","report_says","source_says","reason"],
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
  verdict is NOT "faithful"; use "overstated", and give source_says (below).

EVIDENCE: cite the passages that decide the verdict. Each evidence item is {passage_id, quote} where
passage_id is the id from a header line and quote is copied VERBATIM from that passage — never
paraphrase, never edit, never merge across passages. If nothing supports the claim, return an empty
evidence array and verdict "absent".

VERDICT:
- "faithful": the source states the claim's SUBJECT, SCOPE, and DIRECTION — the same thing, about the
  same population / place / quantity, moving the same way — including the register and force with which
  the speaker stated it. A statement that is only ADJACENT (a related but different proposition) or
  BROADER (true of a larger population, place, or category than the claim names) is "partial" at best,
  with gap="scope"; it is NOT "faithful".
- "partial": the source supports a weaker/narrower version, or only an adjacent/broader one (gap=scope).
- "overstated": same in kind but the summary strengthened it — or literalized a rhetorical claim.
- "absent": not in the passages.
- "contradicted": the source says the opposite.
For every verdict but "faithful", give report_says and source_says (below).

SCOPE DISCIPLINE — two worked negatives, so subject/scope/direction is checked, not waved through:
- "COVID-19 took away critical training opportunities from those looking to enter the industries" is
  NOT faithful to a source saying graduates who trained online during COVID came out under-skilled:
  that is an adjacent proposition (the training still happened, it was degraded), not the removal of
  training opportunities. Verdict: partial, gap=scope.
- A claim about "regional Victoria's" share of funding is NOT established by a source about regional
  AUSTRALIA, or about the nation as a whole: regional Australia is a broader place than regional
  Victoria. Same subject (funding share), wrong scope (place). Verdict: partial, gap=scope.

GAP — the single field that most changes what a reader would do, err toward naming one over "none":
"scope" (narrower population/place/category), "denominator" (a fraction/share whose base is dropped),
"timerange" (a limited period dropped), "attribution" (support is about a different actor/programme/
body), "other", or "none" (benign narrowing; use for "faithful").

REPORT_SAYS and SOURCE_SAYS — two plain restatements the reader compares side by side, each <=12
words, in ordinary words. report_says is the impression the SUMMARY CLAIM gives; source_says is what
the PASSAGES actually support. Rephrase them in everyday words,
as if to someone who hasn't read the report. Rules: no "reader", no "would"; no word longer than
three syllables unless it already appears in the claim or a quote; report_says must NOT copy a run of
three or more words straight from the claim (say the summary's point your own way — a source phrase
like "based in Victoria" is fine here). source_says states only what the source says, so it may reuse
the source's own short phrases but must NOT contrast or negate — no "not", "only", "just", "rather
than", "instead", "but" (the reader sees the contrast framed for them). Leave both "" only when the
verdict is "faithful"; for "absent" give report_says, source_says "".
Worked example — claim "...52 internal projects with the majority of production in Victoria",
quote "52 internal productions based in Victoria":
  report_says: most of the work on those 52 projects was done in Victoria
  source_says: the projects were based in Victoria
REASON: <=40 words, why this verdict.`

// faithJudgeSingleSourceRules is appended to faithJudgeSys only for a single_source run (the report is
// its own only source, so the PASSAGES are other parts of the same document, not a witness). The base
// prompt flags any narrowing or restatement as overstated/contradicted; here a report legitimately
// states a figure in the executive summary and again, at more or less detail, in the body. These rules
// give the comparison a direction so a faithful restatement is not scored as a distortion. spec/TREE.md
// § The rules that gate a verdict is the contract.
const faithJudgeSingleSourceRules = `

SINGLE-SOURCE DIRECTION — READ THIS BEFORE THE VERDICT RULES ABOVE; where the two conflict, THIS wins.
The PASSAGES are OTHER parts of the SAME report as the CLAIM, not an outside witness, so a claim and a
passage are two statements by one author about one dataset. This changes what counts as a distortion:
the SCOPE DISCIPLINE and "adjacent/broader is not faithful" rules above are for weighing a summary
against an INDEPENDENT source, and they DO NOT apply here. The comparison has a direction, and only one
direction is a distortion — the body pinning down what the claim inflated.

- The distortion verdicts ("overstated", "contradicted") require a passage that states the SAME fact
  with a NARROWER scope or a DIFFERENT value AND is the MORE DETAILED of the two. Absent such a passage,
  DO NOT reach them — with no conflicting passage retrieved, a claim the passages neither pin down nor
  repeat is "absent" (rendered "uncorroborated"), never "overstated" or "contradicted".
- A claim carrying MORE detail than a passage is NOT overstated by it. A passage that summarises,
  rounds, or drops a qualifier the claim keeps is the VAGUER statement; a summary that drops detail is
  not a conflict. A report that surveys four countries and elsewhere reports the pooled figure without
  renaming them has not contradicted the claim that names them. Verdict "faithful".
- Figures that NEST are consistent, NOT contradictory: "very" (55%) sits inside "very or somewhat"
  (85%); a component share sits inside the total that contains it. A smaller sub-figure you can see does
  NOT contradict a larger combined figure the claim states — seeing 55% "very" is positive evidence FOR
  an 85% "very or somewhat", not against it. NEVER return "contradicted" on a nested figure. Verdict
  "faithful" when the visible component nests inside the claim's total.
- COMPLEMENTARY percentages are consistent, NOT contradictory. Two shares that partition one population
  — "some degree of confidence" (70%) and "little or no trust" (30%) — are two faces of ONE 100% split,
  not two facts in conflict; seeing the 30% is positive evidence FOR the 70% claim, never against it.
  NEVER return "contradicted" because a claim's figure and a passage's figure sum to 100. Verdict
  "faithful".
- Figures WITHIN A POINT are the same figure rounded. Two figures within one percentage point of each
  other, or two shares of one split that fall within a point of summing to 100% (78% "not diminished"
  and a 21% probability of the reverse sum to 99%), differ only by rounding — consistent, never
  contradicted. Verdict "faithful".
- A claim at ONE LEVEL of a stated taxonomy does not conflict with the taxonomy's PARENT or a SIBLING
  level. When the report defines the hierarchy itself — the three factors of throughput, the two of
  instability, both under software delivery performance — a claim that places an item at the level the
  source places it does not contradict a passage that names only the parent or the sibling level.
  Reading "recovery time is a throughput factor" as contradicting a passage on instability's factors
  mistakes one level of a stated taxonomy for a conflict. Verdict "faithful" when the claim's placement
  matches the source's own taxonomy.
- A passage that contains the claim NEAR-VERBATIM — same words, same figure, same scope — is
  "faithful", whatever wording differs elsewhere.

So: find the more-detailed passage that narrows or changes the claim before reaching a distortion
verdict. A mere restatement, a rounding, a dropped qualifier, or a nested figure is "faithful".`

const evidenceSys = `You are the Evidence Grounder. Decide whether the CLAIM is TRUE, using web
search to find real, current evidence — the actual truth-makers, not anyone's assertion that it is
true. Search for data, primary sources, and credible reporting; weigh what you find. Then judge:
- "supported": credible evidence backs the claim.
- "mixed": evidence cuts both ways, or supports only a qualified version.
- "refuted": credible evidence contradicts the claim.
- "unverifiable": a prediction, opinion, or otherwise not checkable against current evidence.

Also set "horizon": the time by which the claim's truth is settled, relative to today:
- "past": already settled by events that have happened.
- "present": settled by the current state of the world.
- "future": settled only by an event or outcome that has not happened yet (a forecast, a target date).
A projection OF a future outcome is not an observation of it, no matter how many forecasters agree.

Keep finding to one sentence. List the sources you actually used, with real URLs from your search
results. Return ONLY JSON after searching:
{"verdict":"supported"|"mixed"|"refuted"|"unverifiable","horizon":"past"|"present"|"future","finding":string,"sources":[{"title":string,"url":string}]}`

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

// wireBackend selects the LLM provider for the run's model calls. Only Anthropic needs a key; Ollama
// talks to a local server, so requiring ANTHROPIC_API_KEY there would be a false gate. It sets
// c.backend and c.backendName, and is called by the main judge path and by -from -edge (whose edge
// pass is the one -from mode that makes model calls, spec/EDGE.md).
func (c *cfg) wireBackend(backendName, ollamaURL string, think bool) {
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
	// The evidence path (withTools) shares its output budget with web_search blocks, so it needs the
	// raised cap or the verdict JSON truncates before it is emitted; every other path keeps the default.
	maxTok := 0
	if withTools {
		maxTok = backend.WebSearchMaxTokens
	}
	resp, err := c.backend.Complete(backend.Request{
		System: system, Prompt: prompt, Cached: c.cachedSource, WithTools: withTools,
		Schema: c.reqSchema, SchemaName: c.reqSchemaName, Temperature: c.reqTemp,
		MaxTokens: maxTok,
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
	ReportSays                           string   // ≤12-word plain restatement of the summary's impression (judge)
	SoWhat                               string   // stakes line assembled by renderStakes from report/source_says
	Quotes                               []string // defender's verbatim source spans, for the tree leaf
	QuoteSources                         []string // parallel to Quotes: "hearing"|"submission" origin of each
	QuotePassages                        []string // parallel to Quotes: the passage id each quote was cited from
}
type source struct{ Title, URL string }
type evidence struct {
	Claim, Verdict, Finding string
	Horizon                 string // when the claim's truth is settled: past|present|future
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
	Horizon string `json:"horizon"`
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
	ReportSays string `json:"report_says"` // ≤12-word plain restatement of the summary's impression
	SourceSays string `json:"source_says"` // ≤12-word plain restatement of what the passages support
	Reason     string `json:"reason"`
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
// parseClaimLine splits a claims-file line: id, §-path, text, an optional 4th tab field of cites
// (space/comma-separated document ids checked against the manifest), and an optional 5th of route
// (evidence|source|evaluative|data-gap, per spec/TREE.md). Shorter lines leave the trailing fields "".
func parseClaimLine(raw string) (id, path, text, cites, route string) {
	parts := strings.SplitN(raw, "\t", 5)
	get := func(i int) string {
		if i < len(parts) {
			return strings.TrimSpace(parts[i])
		}
		return ""
	}
	if len(parts) < 3 {
		return "", "", raw, "", ""
	}
	return get(0), get(1), get(2), get(3), get(4)
}

// routeOr prefers the route recorded in the chain, falling back to the claims file's — so a chain
// written before the `route` field existed still renders correctly once the claims file carries it.
func routeOr(chainRoute, claimsRoute string) string {
	if strings.TrimSpace(chainRoute) != "" {
		return chainRoute
	}
	return claimsRoute
}

// loadClaimsFile reads a claims file, pulling any leading "# key: value" header lines the tree wants —
// "# title:" and "# date:" for the root block — and returning the claim lines with all "#"-comment
// lines removed, so a header never counts as a claim. The tab-delimited claim lines are untouched.
func loadClaimsFile(path string) (title, date string, items []string) {
	for _, ln := range strings.Split(mustRead(path), "\n") {
		t := strings.TrimSpace(ln)
		if t == "" {
			continue
		}
		if strings.HasPrefix(t, "#") {
			body := strings.TrimSpace(strings.TrimPrefix(t, "#"))
			if k, v, ok := strings.Cut(body, ":"); ok {
				switch strings.ToLower(strings.TrimSpace(k)) {
				case "title":
					title = strings.TrimSpace(v)
				case "date":
					date = strings.TrimSpace(v)
				}
			}
			continue
		}
		items = append(items, t)
	}
	return title, date, items
}

// citedDocs splits a cites field into document ids (space- or comma-separated).
func citedDocs(cites string) []string {
	fields := strings.FieldsFunc(cites, func(r rune) bool { return r == ',' || r == ' ' || r == '\t' })
	out := fields[:0]
	for _, f := range fields {
		if f = strings.TrimSpace(f); f != "" {
			out = append(out, f)
		}
	}
	return out
}

// missingCites returns the cited document ids not present in the held set, in order. Empty when every
// cited doc is held (or the claim cites nothing, or no manifest was loaded).
func (c cfg) missingCites(cites string) []string {
	if c.held == nil || cites == "" {
		return nil
	}
	var missing []string
	for _, id := range citedDocs(cites) {
		if !c.held[id] {
			missing = append(missing, id)
		}
	}
	return missing
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
	// The root summary block sits above the tree in both stdout (tree mode) and tree.html.
	rootBlock := tree.RootBlock(rows, tree.RootMeta{
		Title: c.rootTitle, Date: c.rootDate, SourceDocs: c.sourceDocs, Runs: c.runs})
	if c.treeHTMLPath != "" {
		htmlDoc := tree.RenderHTML(rows, details, c.headerLine(), rootBlock)
		if err := os.WriteFile(c.treeHTMLPath, []byte(htmlDoc), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "warning: cannot write %s: %v\n", c.treeHTMLPath, err)
		} else {
			fmt.Fprintf(os.Stderr, "tree (html): %s\n", c.treeHTMLPath)
		}
	}
	fmt.Println(c.headerLine())
	switch c.renderMode {
	case "full":
		fmt.Printf("changes %d summary lines\n", brief.OpenedBranches(rows))
		if c.asMarkdown {
			fmt.Print(mdTable)
		} else {
			termTable()
		}
	case "tree":
		fmt.Print(rootBlock) // its last line is the value line; it replaces the bare "changes N" above the tree
		fmt.Println()
		fmt.Print(tree.Render(rows, c.treeAll, c.auditPath))
	default:
		// The value line (docs/VALUE.md): how many executive-summary lines this run changes.
		fmt.Printf("changes %d summary lines\n", brief.OpenedBranches(rows))
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
	Route    string          `json:"route,omitempty"`    // what settles the claim: evidence|source|evaluative|data-gap
	Spread   string          `json:"spread,omitempty"`   // "k/N" agreement when -n>1, else ""
	Samples  []string        `json:"samples,omitempty"`  // every -n>1 sample's verdict, in draw order; -from reads spread and dissent from this
	Passages []string        `json:"passages,omitempty"` // retrieved passage ids the judge saw (bm25 mode)
	ElapsedS float64         `json:"elapsed_s"`
	Detail   json.RawMessage `json:"detail"`
}

// appendChain appends one JSON record to c.chainFile. Errors are logged to
// stderr and silently dropped — chain failures must never abort the run.
func (c *cfg) appendChain(rec chainRecord) { c.appendChainTo(c.chainFile, rec) }

// appendChainTo appends one record to a named chain file — the seam that lets the corpus run write a
// second `.substance.jsonl` beside the faithfulness chain (spec/SUBSTANCE-CORPUS.md) through the same
// backend-stamping and error handling. An empty path is a no-op.
func (c *cfg) appendChainTo(path string, rec chainRecord) {
	if path == "" {
		return
	}
	rec.Backend = c.backendName // stamp the producing backend on every record
	if c.verbose {
		fmt.Fprintf(os.Stderr, "[chain] %s idx=%d verdict=%s\n", path, rec.Idx, rec.Verdict)
	}
	data, err := json.Marshal(rec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: chain marshal idx=%d: %v\n", rec.Idx, err)
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: chain open %q: %v\n", path, err)
		return
	}
	defer f.Close()
	fmt.Fprintln(f, string(data))
}

// edgeDetail is the per-edge record the edge pass writes to its chain (spec/EDGE.md §2–§4): the
// reconstructed warrant, how many defeaters the N samples OFFERED against how many the admission check
// ADMITTED (their gap is the calibration signal, §3), the failed step for each rejected offer, the
// modal admitted defeater's fields (empty when the edge is unchallenged), and whether the template rule
// lifted this edge's defeater to the root as method-level.
type edgeDetail struct {
	Corpus           string   `json:"corpus,omitempty"` // example dir the edge is from; edge_tally reads this, not the filename
	FindingID        string   `json:"finding_id"`
	RecID            string   `json:"rec_id"`
	Scheme           string   `json:"scheme"`
	Warrant          string   `json:"warrant,omitempty"`
	Offered          int      `json:"offered"`
	Admitted         int      `json:"admitted"`
	Rejected         []string `json:"rejected,omitempty"` // one failed admission step per rejected offer
	World            string   `json:"world,omitempty"`
	Kind             string   `json:"kind,omitempty"`
	Anchor           string   `json:"anchor,omitempty"`
	Settles          string   `json:"settles,omitempty"`
	CriticalQuestion string   `json:"critical_question,omitempty"` // which of the scheme's fixed CQs the defeater answers
	MethodLevel      bool     `json:"method_level,omitempty"`      // lifted to the root by the template rule (§3 rule 4)
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
	Quotes          []string `json:"quotes,omitempty"`         // verbatim source spans the judge cited
	QuoteSources    []string `json:"quote_sources,omitempty"`  // parallel: "hearing"|"submission" per quote
	QuotePassages   []string `json:"quote_passages,omitempty"` // parallel: the passage id each quote was cited from ("" = unresolved)
	CriticFinding   string   `json:"critic_finding,omitempty"`
	DistortionType  string   `json:"distortion_type,omitempty"`
	ReportSays      string   `json:"report_says,omitempty"` // judge's ≤12-word plain restatement of the summary
	SourceSays      string   `json:"source_says,omitempty"` // judge's ≤12-word plain restatement of the source
	Gap             string   `json:"gap,omitempty"`         // the judge's gap classification
	SoWhat          string   `json:"so_what,omitempty"`     // stakes line assembled by renderStakes
}

type evidenceDetail struct {
	Finding          string            `json:"finding,omitempty"`
	Horizon          string            `json:"horizon,omitempty"` // past|present|future — settles the forecast gate
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
		QuoteSources:  f.QuoteSources,
		QuotePassages: f.QuotePassages,
		CriticFinding: f.Evidence,
		ReportSays:    f.ReportSays,
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
		Horizon:          e.Horizon,
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
			QuotePassages: f.QuotePassages,
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
			Horizon:          e.Horizon,
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

// mergeChains pools, per claim index, the sample verdicts from every chain and returns one record per
// index carrying the modal verdict over the pool, the pooled samples (so spreadForRecord recomputes
// the k/N fraction and dissent for free), and the detail of a chain that drew that modal verdict. It
// also returns each index's stability class (settled/wobble/contested) — the axis the multi-run root
// block and the contested Needs-you tier read. Every chain shares length and per-index claim text
// (the caller checks each chain against the claims file). Single-chain -from never reaches here.
func mergeChains(chains [][]chainRecord) (merged []chainRecord, classes []string) {
	n := len(chains[0])
	merged = make([]chainRecord, n)
	classes = make([]string, n)
	for i := 0; i < n; i++ {
		var pool []string
		counts := map[string]int{}
		for _, ch := range chains {
			for _, v := range recordSamples(ch[i]) {
				pool = append(pool, v)
				counts[v]++
			}
		}
		modal := modalVerdict(counts)
		rep := chains[0][i] // representative detail: the first chain that drew the modal verdict
		for _, ch := range chains {
			if ch[i].Verdict == modal {
				rep = ch[i]
				break
			}
		}
		rep.Verdict = modal
		rep.Samples = pool // drive spreadForRecord off the pool, not any one chain's pre-baked spread
		rep.Spread = ""
		merged[i] = rep
		classes[i] = stabilityClass(pool)
	}
	return merged, classes
}

// recordSamples is the sample verdicts one chain record contributes to the merge pool: its recorded
// per-sample list when it ran -n>1, else its single verdict counted once.
func recordSamples(r chainRecord) []string {
	if len(r.Samples) > 0 {
		return r.Samples
	}
	return []string{r.Verdict}
}

// stabilityClass names how the pooled verdicts for one leaf spread across the merged runs: settled
// (one verdict throughout), wobble (several verdicts on one side of the support divide, with a clear
// majority), or contested (the verdicts cross the divide — a run backs the claim and a run does not,
// or one could not check — or there is no majority at all, a tie the merge cannot resolve into one
// verdict). spec/TREE.md § Merging runs is the contract.
func stabilityClass(samples []string) string {
	if len(samples) == 0 {
		return ""
	}
	distinct := map[string]bool{}
	sides := map[string]bool{}
	for _, v := range samples {
		distinct[v] = true
		sides[verdictSide(v)] = true
	}
	switch {
	case len(distinct) == 1:
		return "settled"
	case isModalTie(samples):
		// No majority verdict, so the merge cannot name one — the sources do not settle this even when
		// every draw sits on one side. It reads as contested and renders "split a/b", never a verdict.
		return "contested"
	case len(sides) == 1:
		return "wobble"
	default:
		return "contested"
	}
}

// isModalTie reports a pool whose top verdict count is shared by two or more distinct verdicts — a
// merge with no majority. modalVerdict would still return one (worst-first), but that pick is an
// artifact of the tie-break, not a verdict the runs agreed on.
func isModalTie(samples []string) bool {
	_, tie := modalTie(samples)
	return tie
}

// modalTie returns the tied top verdicts (worst-first, by faithVerdictOrder) when the pool's highest
// count is shared, and reports whether such a tie exists. A single distinct verdict, or one clear
// winner, is not a tie.
func modalTie(samples []string) (tied []string, isTie bool) {
	counts := map[string]int{}
	for _, v := range samples {
		counts[v]++
	}
	top := 0
	for _, n := range counts {
		if n > top {
			top = n
		}
	}
	for v, n := range counts {
		if n == top {
			tied = append(tied, v)
		}
	}
	if len(tied) < 2 {
		return nil, false
	}
	rank := func(v string) int {
		for i, o := range faithVerdictOrder {
			if o == v {
				return i
			}
		}
		return len(faithVerdictOrder)
	}
	sort.SliceStable(tied, func(i, j int) bool { return rank(tied[i]) < rank(tied[j]) })
	return tied, true
}

// splitDescriptor is the "a/b" line a tied contested leaf shows instead of a verdict — the tied top
// verdicts joined worst-first. It is "" when the pool has a clear majority, so only a genuine tie
// carries one.
func splitDescriptor(samples []string) string {
	tied, tie := modalTie(samples)
	if !tie {
		return ""
	}
	return strings.Join(tied, "/")
}

// applySchemaGate forces a leaf's faithfulness verdict to unverifiable when the judge's reason failed
// the schema (empty or a raw tag), marking the row so the root block files it under schema failure and
// the brief surfaces it. Opinions are exempt: they are never judged, so their empty reason is expected,
// not a failure. Reports whether the gate fired. spec/TREE.md § The rules that gate a verdict.
func applySchemaGate(r *brief.Row, reason string) bool {
	if brief.IsOpinion(*r) || !brief.SchemaFailed(reason) {
		return false
	}
	r.Faith = "unverifiable"
	r.SchemaFail = true
	return true
}

// schemaLeafFlag is the reason line a schema-failed leaf shows in place of the unusable judge output:
// which failure it was, with the stranded tag quoted so a reader can see what leaked.
func schemaLeafFlag(reason string) string {
	if strings.TrimSpace(reason) == "" {
		return "schema failure: the judge returned no reason"
	}
	return "schema failure: the judge reason carried a raw tag (" + strings.TrimSpace(reason) + ")"
}

// verdictSide places a faithfulness verdict on the support divide the stability class turns on:
// "supported" (the source backs the claim) vs "not-supported" (it does not). unverifiable is its own
// side — "can't check" is neither support nor its refusal — so mixing it with a decided verdict reads
// as contested. Any other string (a substance/grounding verdict under a cross-mode merge, or "error")
// is its own side too, so a bare disagreement there reads as contested rather than being mislabelled.
func verdictSide(v string) string {
	switch v {
	case "faithful", "partial":
		return "supported"
	case "overstated", "absent", "contradicted", "unsupported":
		return "not-supported"
	default:
		return v
	}
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

// runFromChain reconstructs the verdict rows from one or more saved chains and renders them through
// the same stdout renderers a live run uses (brief / -tree / -full), making no model call. `chainSpec`
// is one path or a comma-list; several paths are merged leaf-by-leaf (mergeChains) so the render
// reports the agreement across full runs. The claims file is required: it supplies the ids and heading
// paths the tree needs, and cross-checking each record's claim text against it catches a chain paired
// with the wrong claims file.
func (c *cfg) runFromChain(chainSpec, claimsPath string) {
	if claimsPath == "" {
		fatal("-from needs the claims file as INPUT (its claims are checked against the chain)")
	}
	var chainPaths []string
	for _, p := range strings.Split(chainSpec, ",") {
		if p = strings.TrimSpace(p); p != "" {
			chainPaths = append(chainPaths, p)
		}
	}
	if len(chainPaths) == 0 {
		fatal("-from needs at least one chain path")
	}

	chains := make([][]chainRecord, len(chainPaths))
	mode := ""
	for k, p := range chainPaths {
		recs, m, err := readChain(p)
		if err != nil {
			fatal("read chain " + p + ": " + err.Error())
		}
		if k == 0 {
			mode = m
		} else if m != mode {
			fatal(fmt.Sprintf("chains disagree on mode: %q is %q, %q is %q", chainPaths[0], mode, p, m))
		}
		if len(recs) != len(chains[0]) && k > 0 {
			fatal(fmt.Sprintf("chains disagree on length: %q has %d, %q has %d",
				chainPaths[0], len(chains[0]), p, len(recs)))
		}
		chains[k] = recs
	}
	if len(chainPaths) > 1 && mode == "audit" {
		fatal("-from cannot merge audit chains: the verdict is a composite of three axes")
	}

	title, date, items := loadClaimsFile(claimsPath)
	c.rootTitle, c.rootDate = title, date
	if len(chains[0]) != len(items) {
		fatal(fmt.Sprintf("chain has %d records but %s has %d claims", len(chains[0]), claimsPath, len(items)))
	}
	ids := make([]string, len(items))
	paths := make([]string, len(items))
	texts := make([]string, len(items))
	routes := make([]string, len(items))
	for i, it := range items {
		id, path, text, _, route := parseClaimLine(it)
		if id == "" {
			id = fmt.Sprintf("c%d", i+1)
		}
		ids[i], paths[i], texts[i], routes[i] = id, path, text, route
	}
	// Every chain's claim text must match the claims file, so a merge cannot silently pool two runs
	// that were judging different claims.
	for k, recs := range chains {
		for i, r := range recs {
			if strings.TrimSpace(r.Claim) != strings.TrimSpace(texts[i]) {
				fatal(fmt.Sprintf("claim %d in %s does not match %s:\n  chain:  %q\n  claims: %q",
					i+1, chainPaths[k], claimsPath, r.Claim, texts[i]))
			}
		}
	}

	// One chain replays unchanged (no classes). Several merge leaf-by-leaf into one record set plus a
	// per-leaf stability class, and the root block reports the settled/wobble/contested split.
	recs := chains[0]
	classes := make([]string, len(recs))
	c.runs = len(chainPaths)
	if len(chainPaths) > 1 {
		recs, classes = mergeChains(chains)
	}

	// A -from render is a pure renderer: the chain dir belongs to the judge run that wrote it, so the
	// render leaves it byte-unchanged and never writes back into it. `audit.md` and `tree.html` already
	// sit in the chain dir as that judge run's own artifacts (`runJudge` writes them at chainDir); the
	// render re-derives none of them. Its only file output is the review site (spec/SERVE.md), under the
	// example dir when -review is set, where `buildSite` writes `index.html` and `review.html`.
	// Refuter: assay_test.go TestFromRenderLeavesChainDirUnchanged.
	if c.argumentFile != "" && c.review {
		// The site is a peer of the render dir under the example dir — sourcesDir is <example>/sources,
		// so its parent is the example dir the site sits beside. index.html goes into the site via
		// writeSite, never into the chain dir.
		c.siteDir = filepath.Join(filepath.Dir(c.sourcesDir), "site")
	}

	rows := make([]brief.Row, len(recs))
	details := make(map[string]tree.Leaf, len(recs))
	vs := make([]string, len(recs))

	var (
		counts    string
		mdTable   string
		termTable func()
	)
	switch mode {
	case "faithfulness":
		fr := make([]faith, len(recs))
		for i, r := range recs {
			var det faithDetail
			_ = json.Unmarshal(r.Detail, &det)
			fr[i] = faith{Claim: r.Claim, Verdict: r.Verdict, Evidence: det.CriticFinding,
				SourceSays: det.SourceSays, Gap: det.Gap, SoWhat: det.SoWhat, Quotes: det.Quotes}
			spread, dissent := spreadForRecord(r) // rebuild from the raw samples and name the dissent behind a split
			rows[i] = brief.Row{ID: ids[i], Path: paths[i], Text: texts[i],
				Faith: r.Verdict, FaithReason: det.CriticFinding, Spread: spread, Dissent: dissent,
				Gap: det.Gap, SoWhat: det.SoWhat, Route: routeOr(r.Route, routes[i])}
			leaf := tree.Leaf{Reason: det.CriticFinding, Quotes: buildLeafQuotes(det.Quotes, det.QuotePassages, c.docLabels, c.reports)}
			if applySchemaGate(&rows[i], det.CriticFinding) {
				leaf.Reason = schemaLeafFlag(det.CriticFinding)
			}
			details[ids[i]] = leaf
			vs[i] = rows[i].Faith // reflects the schema-gate override in the rollup, not just the tree
		}
		// A sibling <fixture>.substance.jsonl (spec/SUBSTANCE-CORPUS.md) overlays each leaf's substance
		// verdict, so the argument tree can roll a faithful-but-hollow leaf up as weakened. Absent, the
		// render is unchanged (a faithfulness-only run).
		attachSubstanceOverlay(rows, texts, substanceSibling(chainPaths[0]))
		counts, mdTable, termTable = tally(vs), mdFaith(fr), func() { c.termFaith(fr) }
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
				Substance: r.Verdict, SubstanceReason: reason, Route: routeOr(r.Route, routes[i])}
			details[ids[i]] = tree.Leaf{Reason: reason}
			vs[i] = r.Verdict
		}
		counts, mdTable, termTable = tally(vs), mdSubstance(sr), func() { c.termSubstance(sr) }
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
				Grounding: r.Verdict, GroundReason: det.Finding, Route: routeOr(r.Route, routes[i])}
			details[ids[i]] = tree.Leaf{Reason: det.Finding}
			vs[i] = r.Verdict
		}
		counts, mdTable, termTable = tally(vs), mdEvidence(er), func() { c.termEvidence(er) }
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
				Gap: det.Faith.Gap, SoWhat: det.Faith.SoWhat, Route: routeOr(r.Route, routes[i])}
			leaf := tree.Leaf{Reason: det.Faith.CriticFinding, Quotes: buildLeafQuotes(det.Faith.Quotes, det.Faith.QuotePassages, c.docLabels, c.reports)}
			if applySchemaGate(&rows[i], det.Faith.CriticFinding) {
				leaf.Reason = schemaLeafFlag(det.Faith.CriticFinding)
				fv = rows[i].Faith
			}
			details[ids[i]] = leaf
			fvs[i], svs[i], evs[i] = fv, sv, evv
		}
		counts = fmt.Sprintf("faith[%s] · sub[%s] · ground[%s]", tally(fvs), tally(svs), tally(evs))
		mdTable, termTable = mdAudit(texts, fr, sr, er), func() { fmt.Print(mdAudit(texts, fr, sr, er)) }
	default:
		fatal("unknown chain mode: " + mode)
	}

	for i := range rows {
		rows[i].Class = classes[i]           // "" under a single chain; settled/wobble/contested when merged
		rows[i].Section = c.refs[rows[i].ID] // -refs: the report section this claim rests on ("" if none)
		if classes[i] == "contested" {
			rows[i].Split = splitDescriptor(recordSamples(recs[i])) // "" unless the pool tied — then "a/b"
		}
	}
	if c.argumentFile != "" {
		c.presentArgument(rows, details, mdTable)
		return
	}
	c.present(rows, counts, mdTable, termTable, details)
}

// edgePair is one in-scope finding→recommendation edge: the finding node carries the scheme tag and the
// leaf quotes, its parent is the recommendation the edge licenses.
type edgePair struct {
	rec     *tree.ArgNode
	finding *tree.ArgNode
}

// inScopeEdges collects the load-bearing, stated F→R edges the pass may attack (spec/EDGE.md § Scope):
// a finding node carrying a scheme tag, on a stated (not `?`) edge, whose leaf-derived judgement has not
// already collapsed to `fails` — a collapsed premise has no F to hold, so attacking it spends a call to
// no effect. The scheme tag is authored only on in-scope edges, so its presence is the primary gate.
func inScopeEdges(root *tree.ArgNode) []edgePair {
	var out []edgePair
	var walk func(parent, n *tree.ArgNode)
	walk = func(parent, n *tree.ArgNode) {
		if n.Scheme != "" && parent != nil && !n.Query && n.Judgement() != "fails" {
			out = append(out, edgePair{rec: parent, finding: n})
		}
		for _, c := range n.Children {
			walk(n, c)
		}
	}
	walk(nil, root)
	return out
}

// edgeQuotes gathers the already-verified source spans for a finding's leaves — the `faithful`/`partial`
// evidence the report-tree pass grounded and quoteInPassage-verified (spec/EDGE.md §2). The edge reasons
// from what the source was shown to say, so only quotes on a faithful/partial leaf are passed; an
// opinion or collapsed leaf contributes none.
func edgeQuotes(finding *tree.ArgNode, byRow map[string]brief.Row, details map[string]tree.Leaf) []string {
	var qs []string
	for _, c := range finding.Children {
		if len(c.Children) != 0 {
			continue // structural child, not a claim leaf
		}
		r, ok := byRow[c.ID]
		if !ok || (r.Faith != "faithful" && r.Faith != "partial") {
			continue
		}
		for _, q := range details[c.ID].Quotes {
			if strings.TrimSpace(q.Text) != "" {
				qs = append(qs, q.Text)
			}
		}
	}
	return qs
}

// modalEdge returns the modal edge verdict over the N samples and its agreement count k (spec/EDGE.md
// §5). A tie, or a majority of `unchallenged`, declines to open — the pass opens an edge only on a
// genuine majority of admitted defeaters, never on a coin-flip. `error` samples count toward neither.
func modalEdge(samples []string) (verdict string, k int) {
	open, unch := 0, 0
	for _, s := range samples {
		switch s {
		case edge.Open:
			open++
		case edge.Unchallenged:
			unch++
		}
	}
	if open > unch {
		return edge.Open, open
	}
	return edge.Unchallenged, unch
}

// runEdgePass runs the edge-level adversarial pass over the built argument tree (spec/EDGE.md §2–§4):
// one schema-enforced model call per in-scope F→R edge (repeated -n times), the code-side admission
// check on each returned defeater, the cross-edge template rule that lifts a method-level defeater to
// the root, and the rollup that writes each edge's open/unchallenged verdict onto its finding node so
// the recommendation derives the worse of its leaf faithfulness and its edge. Every edge is recorded to
// a chain JSONL under testing/chains/ (warrant, offered/admitted counts, rejected steps, modal defeater,
// edge verdict). It mutates `root` in place; with -edge off it is never called, so the tree is unchanged.
func (c *cfg) runEdgePass(root *tree.ArgNode, rows []brief.Row, details map[string]tree.Leaf) {
	edges := inScopeEdges(root)
	if len(edges) == 0 {
		fmt.Fprintln(os.Stderr, "edge pass: no in-scope F→R edges (nothing tagged scheme= still stands)")
		return
	}
	byRow := make(map[string]brief.Row, len(rows))
	for _, r := range rows {
		byRow[r.ID] = r
	}
	n := c.repeat
	if n < 1 {
		n = 1
	}

	type edgeAgg struct {
		pair     edgePair
		verdict  string
		count    int
		defeater edge.Defeater
		warrant  string
		offered  int
		admitted int
		rejected []string
		samples  []string
	}
	var aggs []edgeAgg
	var results []edge.EdgeResult

	for _, p := range edges {
		schema, ok := edge.Schema(p.finding.Scheme)
		if !ok {
			fmt.Fprintf(os.Stderr, "edge pass: %s carries unknown scheme %q, skipped\n", p.finding.ID, p.finding.Scheme)
			continue
		}
		quotes := edgeQuotes(p.finding, byRow, details)
		user := edge.User(p.finding.Scheme, p.finding.Content, quotes, p.rec.Content)
		// The report text the anchor must land in (spec/EDGE.md §3 step 3): the finding, the
		// recommendation, and each verified quote — not the model's own world.
		reportText := append([]string{p.finding.Content, p.rec.Content}, quotes...)
		var samples, rejected []string
		var admittedDefs []edge.Defeater
		offered, admitted := 0, 0
		warrant := ""
		for s := 0; s < n; s++ {
			var temp *float64
			if n > 1 {
				t := judgeSampleTemp
				temp = &t
			}
			var r edge.Result
			if err := c.callSchema(edge.System, "", user, schema, edge.SchemaName, temp, &r); err != nil {
				samples = append(samples, "error")
				continue
			}
			if warrant == "" {
				warrant = r.Warrant
			}
			if r.NoneAdmitted {
				samples = append(samples, edge.Unchallenged)
				continue
			}
			offered++
			if adm, step := edge.Admit(r.Defeater, reportText...); adm {
				admitted++
				admittedDefs = append(admittedDefs, r.Defeater)
				samples = append(samples, edge.Open)
			} else {
				rejected = append(rejected, step)
				samples = append(samples, edge.Unchallenged)
			}
			if c.showProgress {
				fmt.Fprintf(os.Stderr, "  edge %s ← %s [%s] sample %d/%d\n", p.rec.ID, p.finding.ID, p.finding.Scheme, s+1, n)
			}
		}
		verdict, k := modalEdge(samples)
		var def edge.Defeater
		if verdict == edge.Open && len(admittedDefs) > 0 {
			def = admittedDefs[0]
		}
		aggs = append(aggs, edgeAgg{p, verdict, k, def, warrant, offered, admitted, rejected, samples})
		results = append(results, edge.EdgeResult{
			FindingID: p.finding.ID, RecID: p.rec.ID, Verdict: verdict, Defeater: def})
	}

	// The template rule (spec/EDGE.md §3 rule 4) runs across every edge once all are sampled: an admitted
	// defeater whose anchor recurs on more than one edge is method-level, lifted to the root and cleared
	// from its edges, which then read unchallenged.
	finalVerdict, world, methods := edge.Resolve(results)
	methodEdges := make(map[string]bool)
	for _, m := range methods {
		for _, e := range m.Edges {
			methodEdges[e] = true
		}
	}

	chainPath := c.edgeChainPath()
	// The corpus tag is the example directory the argument file sits in (examples/dora-2026 → dora-2026),
	// stamped on every edge record so edge_tally detects the corpus from the chain, not the chain filename
	// (which is the argument base, e.g. "argument", and names no corpus).
	corpus := filepath.Base(filepath.Dir(c.argumentFile))
	for i, a := range aggs {
		fv := finalVerdict[a.pair.finding.ID]
		det := edgeDetail{
			Corpus:    corpus,
			FindingID: a.pair.finding.ID, RecID: a.pair.rec.ID, Scheme: a.pair.finding.Scheme,
			Warrant: a.warrant, Offered: a.offered, Admitted: a.admitted, Rejected: a.rejected,
			MethodLevel: methodEdges[a.pair.finding.ID],
		}
		if a.verdict == edge.Open {
			det.World, det.Kind = a.defeater.World, a.defeater.Kind
			det.Anchor, det.Settles = a.defeater.Anchor, a.defeater.Settles
			det.CriticalQuestion = a.defeater.CriticalQuestion
		}
		raw, _ := json.Marshal(det)
		c.appendChainTo(chainPath, chainRecord{
			Idx: i + 1, Total: len(aggs), Mode: "edge",
			Claim:   fmt.Sprintf("%s <- %s", a.pair.rec.ID, a.pair.finding.ID),
			Verdict: fv, Spread: fmt.Sprintf("%d/%d", a.count, n), Samples: a.samples,
			Detail: raw,
		})
		a.pair.finding.EdgeVerdict = fv
		if w := world[a.pair.finding.ID]; w != "" {
			a.pair.finding.EdgeWorld = w
		}
	}
	for _, m := range methods {
		root.RootMethods = append(root.RootMethods, m.MethodLine())
	}
	if chainPath != "" {
		fmt.Fprintf(os.Stderr, "edge pass: %d in-scope edges, %d method-level lifted → %s\n",
			len(aggs), len(methods), chainPath)
	}
}

// edgeChainPath resolves the edge chain's destination and truncates any stale file at it. It honours an
// explicit -chain-dir; absent one it defaults under testing/chains/, beside the destructive-calibration
// runs the refuter reads (spec/EDGE.md §5). The file is named for the argument file, so dora and
// master-plan write distinct chains. Returns "" (chain writing off) only if the directory cannot be made.
func (c *cfg) edgeChainPath() string {
	dir := c.edgeChainDir
	if dir == "" {
		dir = filepath.Join("testing", "chains", "edge-"+time.Now().Format("20060102-1504")+"-"+c.model)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "warning: edge chain dir %q: %v\n", dir, err)
		return ""
	}
	base := strings.TrimSuffix(filepath.Base(c.argumentFile), filepath.Ext(c.argumentFile))
	path := filepath.Join(dir, base+".edge.jsonl")
	_ = os.Truncate(path, 0) // a re-run replaces its chain rather than appending a second pass
	return path
}

// presentArgument renders the argument tree (spec/ARGUMENT.md) in place of the section-path tree: it
// writes the flat verdict table to auditPath (unchanged), then builds the tree from the argument file,
// hangs the merged rows on its leaves, derives each node's judgement bottom-up, and writes the
// argument-tree page as index.html (spec/SERVE.md). The root block — root proposition, tally, then one
// line per recommendation — is also printed to stdout, so a reader sees the top-level verdict without
// opening the page.
func (c *cfg) presentArgument(rows []brief.Row, details map[string]tree.Leaf, mdTable string) {
	if c.auditPath != "" {
		if err := os.WriteFile(c.auditPath, []byte(mdTable), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "warning: cannot write %s: %v\n", c.auditPath, err)
		}
	}
	// A single-source corpus (manifest single_source: true) holds the report as its own only source, so
	// a leaf 'absent' is not a grounding miss against some other held document — there is none — but a
	// claim the report states once and no second document repeats. Rename it 'uncorroborated' before it
	// reaches the tree, where it derives as weakened, not failed (leafJudgement, spec/ARGUMENT.md).
	if c.singleSource {
		for i := range rows {
			if rows[i].Faith == "absent" {
				rows[i].Faith = "uncorroborated"
			}
		}
	}
	argText := mustRead(c.argumentFile)
	root, err := tree.BuildArgument(argText, rows)
	if err != nil {
		fatal("argument tree: " + err.Error())
	}
	// -edge: attack the standing F→R inferences (spec/EDGE.md). It runs after the tree is built (it needs
	// each finding's leaf-derived judgement and verified quotes) and mutates the tree in place, so the
	// rollup below renders open/unchallenged edges. With -edge off this is never reached and the tree is
	// unchanged — the byte-identical guarantee.
	if c.edge {
		c.runEdgePass(root, rows, details)
	}
	// title names the page and its <h1>; the thesis (root.Content) stays the thesis card's alone. With
	// no "# title:" line the fall-back is the file name, not the thesis.
	title := tree.ArgumentTitle(argText)
	if title == "" {
		title = filepath.Base(c.argumentFile)
	}
	// The `what` sentence is corpus-specific (it counts the manifest's held set), so it is built here
	// and passed in — the tree package stays corpus-agnostic.
	what := "This page checks whether the report's findings say what its sources say, and which " +
		"recommendations that leaves standing."
	if counts := sourceCountsLine(c.held); counts != "" {
		what += " Sources held: " + counts + "."
	}
	// Adjudications (spec/SERVE.md § Adjudications): a human's own leaf verdicts, read from the example
	// dir (the parent of the sources tree), overlaid so the page shows the human call beside the
	// machine's and the index reports the agreement count. Missing file → nil overlay, page unchanged.
	adjPath := filepath.Join(filepath.Dir(c.sourcesDir), "adjudications.txt")
	adjs, err := adjudicate.Load(adjPath)
	if err != nil {
		fatal("adjudications: " + err.Error())
	}
	// The agreement count covers every adjudicated leaf that has a machine verdict, so an adjudication of
	// a judged claim counts even when that claim is not an argument-tree leaf (many DORA leaves are judged
	// but sit off the argument tree). `machine` is therefore every judged row's id → pooled Faith, after
	// the single-source remap above — not just the rendered leaves. The per-leaf DISPLAY stays gated to
	// the tree: only a rendered leaf draws a card, so only there does the human chip appear beside the
	// machine badge (adjudicate.Agree, spec/SERVE.md § Adjudications).
	machine := make(map[string]string, len(rows))
	for i := range rows {
		machine[rows[i].ID] = rows[i].Faith
	}
	adj := adjudicate.NewOverlay(machine, adjs)
	page, rootBlock := tree.ArgumentPage(root, details, title, what, c.singleSource, adj)
	if adj != nil {
		fmt.Fprintf(os.Stderr, "adjudications: %d leaves, judge agreed on %d\n", adj.Result.N, adj.Result.Agreed)
	}
	if c.indexHTMLPath != "" {
		if err := os.WriteFile(c.indexHTMLPath, []byte(page), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "warning: cannot write %s: %v\n", c.indexHTMLPath, err)
		} else {
			fmt.Fprintf(os.Stderr, "page (html): %s\n", c.indexHTMLPath)
		}
	}
	if c.siteDir != "" {
		if counts, err := buildSite(page, c.sourcesDir, c.siteDir, title, c.reports); err != nil {
			fmt.Fprintf(os.Stderr, "warning: cannot build site %s: %v\n", c.siteDir, err)
		} else {
			fmt.Fprintf(os.Stderr, "site: %s — %s\n", c.siteDir, counts)
			if c.zip {
				if zipPath, n, zErr := zipSite(c.siteDir); zErr != nil {
					fmt.Fprintf(os.Stderr, "warning: cannot zip site %s: %v\n", c.siteDir, zErr)
				} else {
					fmt.Fprintf(os.Stderr, "zip: %s — %d files\n", zipPath, n)
				}
			}
		}
	}
	fmt.Print(rootBlock)
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
	noQuoteDowngrades int // verdicts downgraded because no quote survived the grounding check
	plainRetries      int // judge calls retried because report_says/source_says was not a plain restatement
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
func (u *usageCounters) addPlainRetry()  { u.mu.Lock(); u.plainRetries++; u.mu.Unlock() }
func (u *usageCounters) plainRetriesN() int {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.plainRetries
}
func (u *usageCounters) addNoQuoteDowngrade() {
	u.mu.Lock()
	u.noQuoteDowngrades++
	u.mu.Unlock()
}
func (u *usageCounters) extras() (schemaRetries, quoteRejects, noQuoteDowngrades int) {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.schemaRetries, u.quoteRejects, u.noQuoteDowngrades
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
