package main

// Command crossexam parses flags and wires the packages together per spec/CLI.md.
//
// This build wires substance mode only (PLAN.md §1 walking skeleton):
// `./crossexam claims.txt` decomposes the input, runs the producer–critic loop,
// and prints per-claim verdicts to stdout with the SUMMARY/USAGE accounting to
// stderr. Faithfulness, grounding, -audit, -disagreements, and -md are not wired
// yet; selecting them is a fatal error, never a silent no-op.

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/peterwilliams97/elenchus2/internal/claims"
	"github.com/peterwilliams97/elenchus2/internal/client"
	"github.com/peterwilliams97/elenchus2/internal/modes"
	"github.com/peterwilliams97/elenchus2/internal/render"
)

// defaultInput is used when no -text, file arg, or piped stdin yields content
// (spec/CLI.md, Input forms).
const defaultInput = `1. The future of work will happen inside Codex or Claude Code.
2. Every company will have one super-agent inside their Slack.
3. SaaS is not dead — I would buy SaaS stocks right now.
4. PMs will thrive in the AI era.`

// runOpts is the resolved per-run configuration handed to run().
type runOpts struct {
	input     string
	fixture   string
	model     string
	maxRounds int
	chainDir  string
	quiet     bool
}

func main() {
	// flag.XxxVar binds each flag into a named var, not a *T (spec/CLI.md reads
	// defaults from the flag.XxxVar calls). One var per flag — a single flags
	// struct would breach the ≤6-field rule.
	var (
		model     string
		source    string
		text      string
		chainDir  string
		evidence  bool
		audit     bool
		md        bool
		quiet     bool
		maxRounds int
	)
	flag.StringVar(&model, "model", envOr("ANTHROPIC_MODEL", "claude-sonnet-4-6"), "model ID passed to the API")
	flag.StringVar(&source, "source", "", "source transcript → faithfulness mode (not wired in this build)")
	flag.BoolVar(&evidence, "evidence", false, "evidence-grounding mode (not wired in this build)")
	flag.BoolVar(&audit, "audit", false, "all three modes cross-tabulated (not wired in this build)")
	flag.BoolVar(&md, "md", false, "markdown output (not wired in this build)")
	flag.StringVar(&text, "text", "", "inline input text; takes priority over file arg and stdin")
	flag.IntVar(&maxRounds, "max-rounds", 2, "producer–critic rounds per claim")
	flag.BoolVar(&quiet, "quiet", false, "suppress per-case lines; SUMMARY is still written")
	flag.StringVar(&chainDir, "chain-dir", "", "directory for the Tier-2 JSONL chain; default eval/<stamp>-<model>/")
	flag.Parse()

	// This build is substance only. Reject the modes/flags it does not wire,
	// loudly — never silently ignore a requested mode.
	if source != "" || evidence || audit || md {
		fatal("this build wires substance mode only (PLAN §1 walking skeleton); -source/-evidence/-audit/-md are not implemented yet")
	}

	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		fatal("set ANTHROPIC_API_KEY in your environment first.")
	}

	arg := flag.Arg(0)
	input, err := resolveInput(text, arg, os.Stdin, isCharDevice(os.Stdin))
	if err != nil {
		fatal(err.Error())
	}

	if maxRounds < 1 {
		maxRounds = 1
	}

	c := client.New(client.Config{APIKey: os.Getenv("ANTHROPIC_API_KEY"), Model: model})
	opts := runOpts{
		input:     input,
		fixture:   fixtureName(arg),
		model:     model,
		maxRounds: maxRounds,
		chainDir:  chainDir,
		quiet:     quiet,
	}
	if err := run(c, opts, os.Stdout, os.Stderr); err != nil {
		fatal(err.Error())
	}
}

// run executes a substance run: decompose, grade each claim, write the chain
// (best-effort), then render results to stdout and accounting to stderr.
func run(c *client.Client, o runOpts, stdout, stderr io.Writer) error {
	start := time.Now()

	cl, err := modes.Decompose(c, o.input)
	if err != nil {
		return fmt.Errorf("decompose failed: %w", err)
	}

	results := make([]claims.Substance, 0, len(cl))
	elapsed := make([]time.Duration, 0, len(cl))
	for i, claim := range cl {
		t0 := time.Now()
		r := modes.AssayClaim(c, claim, o.maxRounds)
		dt := time.Since(t0)
		results = append(results, r)
		elapsed = append(elapsed, dt)
		if !o.quiet {
			emit(stderr, progressDone(i+1, len(cl), r, dt))
		}
	}

	chainFile := writeChain(stderr, o.chainDir, o.fixture, o.model, "substance", results, elapsed)

	emit(stdout, render.TermSubstance(results))
	wall := time.Since(start)
	emit(stderr, render.Summary(o.fixture, "substance", results, wall, chainFile))
	emit(stderr, render.UsageLine(o.model, c.Usage(), wall))
	return nil
}

// progressDone formats the Tier-1 per-case completion line (spec/CLI.md).
// ✓ for any real verdict,
// ✗ for "error".
func progressDone(idx, total int, r claims.Substance, dt time.Duration) string {
	mark := "✓"
	if r.Verdict == claims.Error {
		mark = "✗"
	}
	return fmt.Sprintf("[%d/%d] %s %-11s %q  %.1fs\n",
		idx, total, mark, r.Verdict, truncate(r.Claim, 60), dt.Seconds())
}

// emit writes s to w, ignoring write errors — stdout/stderr write failures are
// not actionable mid-run.
func emit(w io.Writer, s string) { _, _ = io.WriteString(w, s) }

// resolveInput applies the input priority: -text > file arg > piped stdin >
// defaultInput (spec/CLI.md, Input forms). stdinIsTTY true means stdin is a
// character device (a terminal), so it is not read.
func resolveInput(text, argPath string, stdin io.Reader, stdinIsTTY bool) (string, error) {
	if text != "" {
		return text, nil
	}
	if argPath != "" {
		b, err := os.ReadFile(argPath)
		if err != nil {
			return "", fmt.Errorf("cannot read %s: %w", argPath, err)
		}
		return string(b), nil
	}
	if stdin != nil && !stdinIsTTY {
		b, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("read stdin: %w", err)
		}
		if s := strings.TrimSpace(string(b)); s != "" {
			return s, nil
		}
	}
	return defaultInput, nil
}

// fixtureName derives the base name used in SUMMARY and chain filenames
// (spec/CLI.md): the file arg's base without extension, or "stdin".
func fixtureName(arg string) string {
	if arg == "" {
		return "stdin"
	}
	return strings.TrimSuffix(filepath.Base(arg), filepath.Ext(arg))
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func isCharDevice(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return true // treat as a terminal: don't block reading a broken stdin
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// truncate cuts s to at most n runes (not bytes), so a multibyte char — the
// sample claims have em-dashes — is never split into invalid UTF-8.
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

func fatal(msg string) {
	emit(os.Stderr, msg+"\n")
	os.Exit(1)
}
