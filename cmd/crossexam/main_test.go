package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/peterwilliams97/elenchus2/internal/client"
)

const (
	prod   = `{"steelman":"S","conditions":"C"}`
	critOK = `{"critique":[{"axis":"Evidence","finding":"f","severity":"clears"}],` +
		`"verdict":"substantive","surviving_claim":"narrowed","reason":"R",` +
		`"needs_another_round":false,"added_conditions":0,"survives_only_by_conditioning":false}`
)

// Full cmd-level wiring through the client fake: decompose → assay → render to
// stdout, SUMMARY/USAGE to stderr, and a chain JSONL file written. No network.
func TestRunWiring(t *testing.T) {
	c := client.New(client.Config{APIKey: "k", Model: "m", HTTP: client.Stub(`["claim one"]`, prod, critOK)})
	dir := t.TempDir()
	var stdout, stderr strings.Builder
	err := run(c, runOpts{input: "some prose", fixture: "fix", model: "m", maxRounds: 2, chainDir: dir}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !strings.Contains(stdout.String(), "substantive") {
		t.Errorf("stdout missing verdict:\n%s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "SUMMARY fix substance") {
		t.Errorf("stderr missing SUMMARY:\n%s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "USAGE") {
		t.Errorf("stderr missing USAGE:\n%s", stderr.String())
	}
	files, _ := filepath.Glob(filepath.Join(dir, "*.substance.jsonl"))
	if len(files) != 1 {
		t.Fatalf("expected one chain file, found %v", files)
	}
	data, _ := os.ReadFile(files[0])
	if !strings.Contains(string(data), `"verdict":"substantive"`) {
		t.Errorf("chain file missing verdict:\n%s", string(data))
	}
}

// A decompose failure is a fatal-able error returned to main (spec/CLI.md).
func TestRunDecomposeError(t *testing.T) {
	c := client.New(client.Config{APIKey: "k", Model: "m", HTTP: client.Stub("not json")})
	var stdout, stderr strings.Builder
	err := run(c, runOpts{input: "x", fixture: "f", model: "m", maxRounds: 2, chainDir: t.TempDir()}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected decompose error")
	}
}

func TestResolveInputPriority(t *testing.T) {
	// -text wins over everything.
	if got, _ := resolveInput("inline", "", strings.NewReader("piped"), false); got != "inline" {
		t.Errorf("text priority: got %q", got)
	}
	// No text/arg, stdin is a TTY → fall back to the hardcoded default input.
	if got, _ := resolveInput("", "", nil, true); got != defaultInput {
		t.Errorf("default fallback: got %q", got)
	}
	// Piped stdin is used when no text/arg.
	if got, _ := resolveInput("", "", strings.NewReader("from pipe\n"), false); got != "from pipe" {
		t.Errorf("stdin: got %q", got)
	}
}

func TestTruncateRuneSafe(t *testing.T) {
	// The em-dash is the 60th rune (bytes 59–61), so it straddles a byte-60 cut:
	// a byte slice would split it into invalid UTF-8. Cutting by runes keeps it
	// whole as the last rune.
	s := strings.Repeat("a", 59) + "—tail"
	got := truncate(s, 60)
	if !utf8.ValidString(got) {
		t.Errorf("truncate produced invalid UTF-8: %q", got)
	}
	if r := []rune(got); len(r) != 60 {
		t.Errorf("truncate returned %d runes, want 60", len(r))
	}
	if !strings.HasSuffix(got, "—") {
		t.Errorf("multibyte rune at the cut was not kept whole: %q", got)
	}
}

func TestFixtureName(t *testing.T) {
	if got := fixtureName("examples/dan_shipper/dan_summary.txt"); got != "dan_summary" {
		t.Errorf("fixtureName = %q, want dan_summary", got)
	}
	if got := fixtureName(""); got != "stdin" {
		t.Errorf("fixtureName(\"\") = %q, want stdin", got)
	}
}
