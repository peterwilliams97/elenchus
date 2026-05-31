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
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	defaultModel = "claude-sonnet-4-6"
	apiURL       = "https://api.anthropic.com/v1/messages"
	maxTokens    = 1500
)

var httpClient = &http.Client{Timeout: 150 * time.Second}

type cfg struct {
	model      string
	apiKey     string
	maxRounds  int
	maxClaims  int
	verbose    bool
	noColor    bool
	asMarkdown bool
	// call is the API dispatch function. When nil, callClaude is used (production).
	// Tests set this to a stub to avoid network calls.
	call func(system, prompt string, withTools bool) (string, error)
}

func main() {
	var c cfg
	var src, text string
	var ev, audit bool
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
	flag.Parse()

	c.apiKey = os.Getenv("ANTHROPIC_API_KEY")
	if c.apiKey == "" {
		fatal("set ANTHROPIC_API_KEY in your environment first.")
	}
	input := readInput(text)

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
}

// ── runners ──────────────────────────────────────────────────────────────────

func (c cfg) runSubstance(input string) {
	claims := c.decompose(input)
	results := make([]substance, len(claims))
	for i, cl := range claims {
		if c.verbose {
			fmt.Println(c.bold(fmt.Sprintf("▸ claim %d/%d: %s", i+1, len(claims), cl)))
		}
		results[i] = c.assayClaim(cl)
	}
	if c.asMarkdown {
		fmt.Print(mdSubstance(results))
	} else {
		c.termSubstance(results)
	}
}

func (c cfg) runFaithfulness(input, src string) {
	claims := splitSummary(input)
	results := make([]faith, len(claims))
	for i, cl := range claims {
		results[i] = c.faithClaim(cl, src)
	}
	if c.asMarkdown {
		fmt.Print(mdFaith(results))
	} else {
		c.termFaith(results)
	}
}

// intendedProposition returns what_source_actually_says when the faithfulness
// pass flagged the claim as partial or overstated and populated that field.
// Otherwise it returns the original claim unchanged. Using this in both
// runEvidence and the audit grounding pass keeps the two paths in sync.
func intendedProposition(fc faith, claim string) string {
	if fc.SourceSays != "" && (fc.Verdict == "partial" || fc.Verdict == "overstated") {
		return fc.SourceSays
	}
	return claim
}

func (c cfg) runEvidence(input, src string) {
	claims := splitSummary(input)
	results := make([]evidence, len(claims))
	for i, cl := range claims {
		// Skip evidence grounding if over the cap.
		if c.maxClaims > 0 && i >= c.maxClaims {
			results[i] = evidence{Claim: cl, Verdict: "skipped (over cap)", Finding: ""}
			continue
		}
		proposition := cl
		if src != "" {
			// Reconstruct the asserted proposition via the faithfulness pass so
			// we ground what the speaker actually meant, not a literalized paraphrase.
			fc := c.faithClaim(cl, src)
			proposition = intendedProposition(fc, cl)
		}
		r := c.evidenceClaim(proposition)
		r.Claim = cl // keep original label for display
		results[i] = r
	}
	if c.asMarkdown {
		fmt.Print(mdEvidence(results))
	} else {
		c.termEvidence(results)
	}
}

// computeAudit runs faithfulness, substance, and evidence for each claim and
// returns the three result slices. The evidence pass grounds the INTENDED
// proposition (via intendedProposition) rather than the literal summary claim,
// so the audit grounding column agrees with -evidence -source.
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
func (c cfg) runAudit(input, src string) {
	claims := splitSummary(input) // shared decomposition across all three modes
	if !c.asMarkdown {
		// Print progress ahead of the blocking calls so the user sees motion.
		for i := range claims {
			fmt.Println(c.grey(fmt.Sprintf("auditing %d/%d…", i+1, len(claims))))
		}
	}
	fs, ss, es := c.computeAudit(claims, src)
	fmt.Print(mdAudit(claims, fs, ss, es)) // audit is always markdown
}

// ── stages ───────────────────────────────────────────────────────────────────

// decompose is the first stage: break an input blob into independently checkable claims. The rest
// of the stages operate on these individual claims.
func (c cfg) decompose(text string) []string {
	var arr []string
	if err := c.callJSON(decomposeSys, "TEXT:\n"+text, false, &arr); err != nil {
		return nil
	}
	return arr
}

func (c cfg) assayClaim(claim string) substance {
	current, rounds, steel := claim, 0, ""
	var last substanceJSON
	for rounds < c.maxRounds {
		var p producerJSON
		if err := c.callJSON(producerSys, "CLAIM:\n"+current, false, &p); err != nil {
			return substance{Claim: claim, Verdict: "error", Reason: err.Error()}
		}
		if rounds == 0 {
			steel = p.Steelman
		}
		u := "CLAIM:\n" + current + "\n\nPRODUCER STEELMAN:\n" + p.Steelman + "\n\nPRODUCER CONDITIONS:\n" + p.Conditions
		if err := c.callJSON(substanceCriticSys, u, false, &last); err != nil {
			return substance{Claim: claim, Verdict: "error", Reason: err.Error()}
		}
		rounds++
		// Only continue to the next round when the critic wants one AND the
		// claim didn't survive purely through condition laundering.
		if last.NeedsAnother && last.SurvivingClaim != "" && rounds < c.maxRounds && !last.SurvivesOnlyByConditions {
			current = last.SurvivingClaim
			continue
		}
		break
	}
	// Downgrade at loop exit if the final round survived only by laundered
	// conditions. A claim that needs invented qualifiers to survive is not partial.
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

func (c cfg) faithClaim(claim, src string) faith {
	var d defenderJSON
	if err := c.callJSON(faithDefenderSys, "SUMMARY CLAIM:\n"+claim+"\n\nSOURCE:\n"+src, false, &d); err != nil {
		return faith{Claim: claim, Verdict: "error"}
	}
	quotes := "(none)"
	if len(d.Quotes) > 0 {
		quotes = "- " + strings.Join(d.Quotes, "\n- ")
	}
	u := fmt.Sprintf("SUMMARY CLAIM:\n%s\n\nDEFENDER FOUND SUPPORT: %v\nDEFENDER QUOTES:\n%s\n\nSOURCE:\n%s",
		claim, d.Found, quotes, src)
	var fj faithJSON
	if err := c.callJSON(faithCriticSys, u, false, &fj); err != nil {
		return faith{Claim: claim, Verdict: "error"}
	}
	return faith{Claim: claim, Verdict: fj.Verdict, Evidence: fj.Evidence, SourceSays: fj.SourceSays}
}

func (c cfg) evidenceClaim(claim string) evidence {
	var e evidenceJSON
	if err := c.callJSON(evidenceSys, "CLAIM:\n"+claim, true, &e); err != nil {
		return evidence{Claim: claim, Verdict: "error", Finding: err.Error()}
	}
	out := evidence{Claim: claim, Verdict: e.Verdict, Finding: e.Finding}
	for _, s := range e.Sources {
		out.Sources = append(out.Sources, source{s.Title, s.URL})
	}
	return out
}

// ── prompts (single source of truth — carry any prompt fixes here) ───────────
// NOTE: the two calibration fixes in flight on the Python side (faithfulness "literalization"
// mode + grounding the intended proposition; substance condition-laundering penalty + loop-stop)
// land in faithCriticSys / substanceCriticSys / assayClaim. They are NOT yet applied here.

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

func (c cfg) callJSON(system, prompt string, withTools bool, v any) error {
	dispatch := c.call
	if dispatch == nil {
		dispatch = c.callClaude
	}
	out, err := dispatch(system, prompt, withTools)
	if err != nil {
		return err
	}
	if err := unmarshalLoose(out, v); err == nil {
		return nil
	}
	strict := system + "\n\nReturn ONLY raw JSON. No prose, no markdown, no backticks. First character must be { or [."
	out2, err := dispatch(strict, prompt, withTools)
	if err != nil {
		return err
	}
	return unmarshalLoose(out2, v)
}

// callClaude makes a raw API call to Claude and returns the full text response.
// `system` and `prompt` are passed directly to the API. The caller is responsible for any prompt
// engineering,
// `withTools`==true enables tool use (e.g. web search) when available for the model.
func (c cfg) callClaude(system, prompt string, withTools bool) (string, error) {
	req := apiReq{Model: c.model, MaxTokens: maxTokens, System: system,
		Messages: []apiMsg{{Role: "user", Content: prompt}}}
	if withTools {
		req.Tools = []apiTool{{Type: "web_search_20250305", Name: "web_search", MaxUses: 5}}
	}
	body, _ := json.Marshal(req)

	if c.verbose {
		fmt.Println(c.cyan("┌─ call · " + c.model + tern(withTools, " (web search)", "")))
		fmt.Println(c.grey("│ system:\n│   " + strings.ReplaceAll(system, "\n", "\n│   ")))
		fmt.Println(c.grey("│ user:\n│   " + strings.ReplaceAll(prompt, "\n", "\n│   ")))
	}

	httpReq, _ := http.NewRequest("POST", apiURL, bytes.NewReader(body))
	httpReq.Header.Set("content-type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	var ar apiResp
	if err := json.Unmarshal(raw, &ar); err != nil {
		return "", fmt.Errorf("unreadable response: %.200s", string(raw))
	}
	if ar.Error != nil {
		return "", fmt.Errorf("api error: %s", ar.Error.Message)
	}
	var sb strings.Builder
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
		}
	}
	if c.verbose {
		fmt.Println(c.cyan("├─ response:"))
		fmt.Println(c.grey("│ " + strings.ReplaceAll(sb.String(), "\n", "\n│ ")))
		fmt.Println(c.cyan("└─"))
	}
	return sb.String(), nil
}

type apiReq struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system,omitempty"`
	Messages  []apiMsg  `json:"messages"`
	Tools     []apiTool `json:"tools,omitempty"`
}
type apiMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type apiTool struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	MaxUses int    `json:"max_uses,omitempty"`
}
type apiResp struct {
	Content []apiBlock `json:"content"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error"`
}
type apiBlock struct {
	Type  string          `json:"type"`
	Text  string          `json:"text"`
	Input json.RawMessage `json:"input"`
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
type faith struct{ Claim, Verdict, Evidence, SourceSays string }
type source struct{ Title, URL string }
type evidence struct {
	Claim, Verdict, Finding string
	Sources                 []source
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
		b.WriteString(fmt.Sprintf("| %d | %s | %s | %s |\n", i+1, mdCell(r.Claim), mdV(r.Verdict), mdCell(why)))
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
		b.WriteString(fmt.Sprintf("| %d | %s | %s | %s |\n", i+1, mdCell(r.Claim), mdV(r.Verdict), mdCell(r.SourceSays)))
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
		b.WriteString(fmt.Sprintf("| %d | %s | %s | %s | %s |\n",
			i+1, mdCell(r.Claim), mdV(r.Verdict), mdCell(r.Finding), strings.Join(srcs, "; ")))
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
