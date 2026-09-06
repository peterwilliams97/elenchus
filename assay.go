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
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"assay/internal/brief"
	"assay/internal/tree"
)

const (
	defaultModel     = "claude-sonnet-4-6"
	apiURL           = "https://api.anthropic.com/v1/messages"
	maxTokens        = 1500
	retryMaxAttempts = 4 // 1 initial + 3 retries on 429/503/529
)

var (
	httpClient = &http.Client{Timeout: 150 * time.Second}
	retryBase  = time.Second // overridden to 0 in tests for instant retry
)

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
	chainFile    string // Tier-2 JSONL destination; set by runners before case loop
	renderMode   string // stdout renderer: "brief" (default), "full", or "tree"
	treeAll      bool   // -tree=full: expand every node rather than only Needs-you branches
	auditPath    string // full-table sink; every run writes it, whichever renderer stdout gets
	treeHTMLPath string // eval/<stamp>/tree.html sink, written alongside audit.md
	usage        *usageCounters
	tally        *runTally
	// cachedSource is the stable prefix (e.g. the source transcript) that callClaude places in its
	// own content block with cache_control ephemeral, ahead of the variable prompt. Set per-call by
	// callJSON/callJSONSourced from their `cached` argument; "" sends a single prompt block. The
	// stub (`call`) ignores it — caching is only meaningful against the real API.
	cachedSource string
	// call is the API dispatch function. When nil, callClaude is used (production).
	// Returns the response text, any retrieved sources (from web_search_tool_result blocks), and an
	// error.
	// Tests set this to a stub to avoid network calls.
	call func(system, prompt string, withTools bool) (string, []retrievedSource, error)
	// httpClient overrides the package-level httpClient. Tests inject a custom RoundTripper here to
	// exercise the transport-retry path without network calls.
	httpClient *http.Client
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
	flag.StringVar(&c.model, "model", envOr("ANTHROPIC_MODEL", defaultModel), "model id")
	flag.StringVar(&src, "source", "", "transcript file → faithfulness mode")
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

	c.apiKey = os.Getenv("ANTHROPIC_API_KEY")
	if c.apiKey == "" {
		fatal("set ANTHROPIC_API_KEY in your environment first.")
	}
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
		c.runFaithfulness(input, mustRead(src))
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

	fmt.Fprintf(os.Stderr, "\nSUMMARY %s %s\n", fixture, mode)
	fmt.Fprintf(os.Stderr, "  cases %d · verified %d · errored %d\n", total, verified, errored)
	fmt.Fprintf(os.Stderr, "  %s\n", strings.Join(parts, " · "))
	fmt.Fprintf(os.Stderr, "  wall %s · est_usd %s\n", elapsed.Round(time.Second), cost)
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

func (c *cfg) runFaithfulness(input, src string) {
	raw := splitSummary(input)
	c.tally = newRunTally(len(raw))
	results := make([]faith, len(raw))
	rows := make([]brief.Row, len(raw))
	vs := make([]string, len(raw))
	details := make(map[string]tree.Leaf, len(raw))
	for i, item := range raw {
		id, path, text := parseClaimLine(item)
		if id == "" {
			id = fmt.Sprintf("c%d", i+1)
		}
		t := c.progressStart(i, len(raw), "faithfulness")
		chosen, spread := c.faithRepeat(text, src)
		results[i] = chosen
		rec := faithChainRecord(i, len(raw), text, chosen, t)
		rec.Spread = spread
		c.appendChain(rec)
		c.progressDone(i, len(raw), chosen.Verdict, text, t)
		rows[i] = brief.Row{ID: id, Path: path, Text: text,
			Faith: chosen.Verdict, FaithReason: chosen.Evidence, Spread: spread}
		details[id] = tree.Leaf{Reason: chosen.Evidence, Quotes: chosen.Quotes}
		vs[i] = chosen.Verdict
	}
	c.present(rows, tally(vs), mdFaith(results), func() { c.termFaith(results) }, details)
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
		SourceSays: fj.SourceSays, Quotes: d.Quotes}
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
correct register). Return ONLY JSON:
{"findings":[{"mode":string,"finding":string}],"verdict":"faithful"|"partial"|"overstated"|"absent"|"contradicted","evidence":string,"what_source_actually_says":string|null}`

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
// `cached` is the stable prefix (source corpus) callClaude puts in its own cache_control block,
// ahead of `prompt`; pass "" when the call has no reusable prefix.
func (c cfg) callJSONSourced(system, cached, prompt string, withTools bool, v any) ([]retrievedSource, error) {
	c.cachedSource = cached // value copy; only callClaude reads it, the stub ignores it
	dispatch := c.call
	if dispatch == nil {
		dispatch = c.callClaude
	}
	out, rs, err := dispatch(system, prompt, withTools)
	if err != nil {
		return nil, err
	}
	if unmarshalLoose(out, v) == nil {
		return rs, nil
	}
	strict := system +
		"\n\nReturn ONLY raw JSON. No prose, no markdown, no backticks. " +
		"First character must be { or [."
	out2, rs2, err := dispatch(strict, prompt, withTools)
	if err != nil {
		return nil, err
	}
	return rs2, unmarshalLoose(out2, v)
}

// callClaude makes a raw API call to Claude and returns the full text response.
// `system` and `prompt` are passed directly to the API. The caller is responsible for any prompt
// engineering,
// `withTools`==true enables tool use (e.g. web search) when available for the model.
func (c cfg) callClaude(system, prompt string, withTools bool) (string, []retrievedSource, error) {
	// Build the user message. A non-empty cachedSource becomes its own block, marked ephemeral and
	// placed ahead of the claim so it forms the cacheable prefix reused across every claim in a run.
	content := make([]textBlock, 0, 2)
	if c.cachedSource != "" {
		content = append(content, textBlock{Type: "text", Text: c.cachedSource, CacheControl: ephemeral})
	}
	content = append(content, textBlock{Type: "text", Text: prompt})
	req := apiReq{Model: c.model, MaxTokens: maxTokens,
		Messages: []apiMsg{{Role: "user", Content: content}}}
	if system != "" {
		// The system prompt is stable per mode, so cache it too — one breakpoint, reused every call.
		req.System = []textBlock{{Type: "text", Text: system, CacheControl: ephemeral}}
	}
	if withTools {
		req.Tools = []apiTool{{Type: "web_search_20250305", Name: "web_search", MaxUses: 5}}
	}
	body, _ := json.Marshal(req)

	if c.verbose {
		fmt.Println(c.cyan("┌─ call · " + c.model + tern(withTools, " (web search)", "")))
		fmt.Println(c.grey("│ system:\n│   " + strings.ReplaceAll(system, "\n", "\n│   ")))
		if c.cachedSource != "" {
			fmt.Printf(c.grey("│ cached prefix: %d bytes\n"), len(c.cachedSource))
		}
		fmt.Println(c.grey("│ user:\n│   " + strings.ReplaceAll(prompt, "\n", "\n│   ")))
	}

	client := httpClient
	if c.httpClient != nil {
		client = c.httpClient
	}

	var raw []byte
	for attempt := 0; ; attempt++ {
		httpReq, _ := http.NewRequest("POST", apiURL, bytes.NewReader(body))
		httpReq.Header.Set("content-type", "application/json")
		httpReq.Header.Set("x-api-key", c.apiKey)
		httpReq.Header.Set("anthropic-version", "2023-06-01")

		resp, err := client.Do(httpReq)
		if err != nil {
			return "", nil, err
		}
		raw, _ = io.ReadAll(resp.Body)
		resp.Body.Close() // explicit close before any retry; do not defer across iterations

		if !retryable(resp.StatusCode) || attempt >= retryMaxAttempts-1 {
			break
		}
		time.Sleep(retryDelay(resp.Header.Get("Retry-After"), attempt))
	}

	var ar apiResp
	if err := json.Unmarshal(raw, &ar); err != nil {
		return "", nil, fmt.Errorf("unreadable response: %.200s", string(raw))
	}
	if ar.Error != nil {
		return "", nil, fmt.Errorf("api error: %s", ar.Error.Message)
	}

	// Accumulate usage before checking stop_reason — a truncated response was still billed.
	if c.usage != nil && ar.Usage != nil {
		webSearches := 0
		if ar.Usage.ServerToolUse != nil {
			webSearches = ar.Usage.ServerToolUse.WebSearchRequests
		}
		// Fall back to counting web_search tool_use blocks if server_tool_use absent.
		if webSearches == 0 {
			for _, b := range ar.Content {
				if b.Type == "server_tool_use" || b.Type == "tool_use" {
					var name struct {
						Name string `json:"name"`
					}
					if json.Unmarshal(b.Input, &name) == nil && name.Name == "web_search" {
						webSearches++
					}
				}
			}
		}
		c.usage.add(ar.Usage.InputTokens, ar.Usage.OutputTokens,
			ar.Usage.CacheReadInputTokens, ar.Usage.CacheCreationInputTokens, webSearches)
		if c.verbose {
			fmt.Fprintf(os.Stderr, "[call] in=%d out=%d cache_read=%d cache_create=%d\n",
				ar.Usage.InputTokens, ar.Usage.OutputTokens,
				ar.Usage.CacheReadInputTokens, ar.Usage.CacheCreationInputTokens)
		}
	}

	if ar.StopReason == "max_tokens" {
		return "", nil, fmt.Errorf(
			"response truncated: stop_reason=max_tokens (limit=%d tokens); raise maxTokens constant",
			maxTokens)
	}

	var (
		sb        strings.Builder
		retrieved []retrievedSource
	)
	for _, b := range ar.Content {
		switch b.Type {
		case "text":
			sb.WriteString(b.Text)
		case "server_tool_use":
			if c.verbose {
				var in struct {
					Query string `json:"query"`
				}
				_ = json.Unmarshal(b.Input, &in)
				fmt.Println(c.grey("│   searched: " + in.Query))
			}
		case "web_search_tool_result":
			// b.Content is the raw JSON value of the "content" field: either a []web_search_result
			// array or a web_search_tool_result_error object.
			// Unmarshal directly into a slice; an error object (not an array) fails silently and
			// lands in the "no sources retrieved" downgrade path.
			var results []struct {
				Type  string `json:"type"`
				URL   string `json:"url"`
				Title string `json:"title"`
			}
			if json.Unmarshal(b.Content, &results) == nil {
				for _, r := range results {
					if r.Type == "web_search_result" {
						retrieved = append(retrieved, retrievedSource{Title: r.Title, URL: r.URL})
					}
				}
			}
		}
	}
	if c.verbose {
		fmt.Println(c.cyan("├─ response:"))
		fmt.Println(c.grey("│ " + strings.ReplaceAll(sb.String(), "\n", "\n│ ")))
		if len(retrieved) > 0 {
			fmt.Printf(c.grey("│ retrieved %d source(s)\n"), len(retrieved))
		}
		fmt.Println(c.cyan("└─"))
	}
	return sb.String(), retrieved, nil
}

// retryable reports whether an HTTP status code warrants a retry.
func retryable(code int) bool { return code == 429 || code == 503 || code == 529 }

// retryDelay returns how long to wait before the next attempt. It honours the Retry-After header
// (integer seconds) when present; otherwise uses full-jitter exponential backoff capped at 30 s.
// retryBase==0 (tests) always returns 0.
func retryDelay(retryAfter string, attempt int) time.Duration {
	if s, err := strconv.Atoi(strings.TrimSpace(retryAfter)); err == nil && s > 0 {
		return time.Duration(s) * time.Second
	}
	d := retryBase << uint(attempt) // 1 s, 2 s, 4 s, …
	if d > 30*time.Second {
		d = 30 * time.Second
	}
	// rand.Int63n(n+1) with n==0 returns 0, so retryBase==0 sleeps for 0.
	return time.Duration(rand.Int63n(int64(d) + 1)) // full jitter: [0, d]
}

type apiReq struct {
	Model     string      `json:"model"`
	MaxTokens int         `json:"max_tokens"`
	System    []textBlock `json:"system,omitempty"`
	Messages  []apiMsg    `json:"messages"`
	Tools     []apiTool   `json:"tools,omitempty"`
}
type apiMsg struct {
	Role    string      `json:"role"`
	Content []textBlock `json:"content"`
}

// textBlock is one content block. A non-nil CacheControl marks the prefix up to and including this
// block as cacheable, so a repeated source corpus is billed once at cache-write rates and then read.
type textBlock struct {
	Type         string        `json:"type"` // always "text"
	Text         string        `json:"text"`
	CacheControl *cacheControl `json:"cache_control,omitempty"`
}
type cacheControl struct {
	Type string `json:"type"` // "ephemeral"
}

// ephemeral is the shared marker for every cache breakpoint; the API caps a request at four.
var ephemeral = &cacheControl{Type: "ephemeral"}
type apiTool struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	MaxUses int    `json:"max_uses,omitempty"`
}
type apiUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	ServerToolUse            *struct {
		WebSearchRequests int `json:"web_search_requests"`
	} `json:"server_tool_use"`
}
type apiResp struct {
	Content    []apiBlock `json:"content"`
	StopReason string     `json:"stop_reason"`
	Usage      *apiUsage  `json:"usage"`
	Error      *struct {
		Message string `json:"message"`
	} `json:"error"`
}
type apiBlock struct {
	Type    string          `json:"type"`
	Text    string          `json:"text"`
	Input   json.RawMessage `json:"input"`
	Content json.RawMessage `json:"content"` // populated for web_search_tool_result blocks
}

// retrievedSource is a URL actually fetched during a web_search_tool_result round trip — distinct
// from source, which is what the model claims it used in its JSON response.
type retrievedSource struct{ Title, URL string }

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
}
type evidenceJSON struct {
	Verdict string `json:"verdict"`
	Finding string `json:"finding"`
	Sources []struct {
		Title string `json:"title"`
		URL   string `json:"url"`
	} `json:"sources"`
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
	return fmt.Sprintf("model %s · calls %d · %s · wall %s",
		c.model, calls, cost, elapsed.Round(time.Second))
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
	Claim    string          `json:"claim"`
	Verdict  string          `json:"verdict"`
	Spread   string          `json:"spread,omitempty"` // "k/N" agreement when -n>1, else ""
	ElapsedS float64         `json:"elapsed_s"`
	Detail   json.RawMessage `json:"detail"`
}

// appendChain appends one JSON record to c.chainFile. Errors are logged to
// stderr and silently dropped — chain failures must never abort the run.
func (c *cfg) appendChain(rec chainRecord) {
	if c.chainFile == "" {
		return
	}
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
				SourceSays: det.SourceSays, Quotes: det.Quotes}
			rows[i] = brief.Row{ID: ids[i], Path: paths[i], Text: texts[i],
				Faith: r.Verdict, FaithReason: det.CriticFinding, Spread: r.Spread}
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
				SourceSays: det.Faith.SourceSays, Quotes: det.Faith.Quotes}
			sr[i] = substance{Claim: r.Claim, Verdict: sv, Reason: det.Substance.Reason,
				SurvivingClaim: det.Substance.SurvivingClaim, Steelman: det.Substance.Steelman}
			er[i] = evidence{Claim: r.Claim, Verdict: evv, Finding: det.Evidence.Finding,
				Sources: det.Evidence.Sources, DowngradeReason: det.Evidence.DowngradeReason,
				OriginalVerdict: det.Evidence.OriginalVerdict}
			rows[i] = brief.Row{ID: ids[i], Path: paths[i], Text: texts[i],
				Faith: fv, Substance: sv, Grounding: evv,
				FaithReason: det.Faith.CriticFinding, GroundReason: det.Evidence.Finding}
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

// usageCounters accumulates API usage across all callClaude invocations.
type usageCounters struct {
	mu                sync.Mutex
	calls             int
	inputTokens       int
	outputTokens      int
	cacheReadTokens   int
	cacheCreateTokens int
	webSearches       int
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
