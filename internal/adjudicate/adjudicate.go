package adjudicate

// adjudicate.go reads a human's leaf-by-leaf verdicts (examples/<x>/adjudications.txt) and measures
// where the machine judge agreed with them, so the argument page can show the human verdict beside the
// machine's and the index can report the agreement count (spec/SERVE.md § Adjudications). It authors no
// verdict: a machine verdict is the judge's pooled Faith, a human verdict is a line in the file, and
// `Agree` only compares the two strings for equality.

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"
)

// Adjudication is one human-judged leaf: the claim `ID`, the human's `Verdict`, the verbatim `Quote`
// they grounded it on, the report `Page` (0 when the line gives none), the adjudicator's `Initials`,
// and a one-line `Reason`.
type Adjudication struct {
	ID       string
	Verdict  string
	Quote    string
	Page     int
	Initials string
	Reason   string
}

// Load reads adjudications from `path` — one leaf per line, `id | verdict | quote | page | initials |
// reason`, with `#` comments and blank lines skipped. A missing file is not an error (a corpus need not
// carry adjudications): it returns nil, nil. A present line without exactly six `|`-fields is an error,
// so a hand-edit that drops a `|` fails loudly rather than mis-parsing a five-field line as a verdict.
func Load(path string) ([]Adjudication, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []Adjudication
	for n, ln := range strings.Split(string(data), "\n") {
		s := strings.TrimSpace(ln)
		if s == "" || strings.HasPrefix(s, "#") {
			continue
		}
		f := strings.Split(s, "|")
		if len(f) != 6 {
			return nil, fmt.Errorf("adjudications %s line %d: want 6 |-fields, got %d", path, n+1, len(f))
		}
		for i := range f {
			f[i] = strings.TrimSpace(f[i])
		}
		page := 0
		if f[3] != "" {
			v, perr := strconv.Atoi(strings.TrimPrefix(f[3], "p"))
			if perr != nil {
				return nil, fmt.Errorf("adjudications %s line %d: page %q is not a number", path, n+1, f[3])
			}
			page = v
		}
		out = append(out, Adjudication{ID: f[0], Verdict: f[1], Quote: f[2], Page: page, Initials: f[4], Reason: f[5]})
	}
	return out, nil
}

// Disagreement names one leaf where the human and the judge differ: the `ID`, the two verdicts, and the
// human's one-line `Reason`, so the index can say why the human read it otherwise.
type Disagreement struct {
	ID      string
	Human   string
	Machine string
	Reason  string
}

// Result is the agreement tally over the adjudicated leaves: `N` adjudications, `Agreed` of them
// matching the machine verdict, and one `Disagreements` entry per mismatch in file order.
type Result struct {
	N             int
	Agreed        int
	Disagreements []Disagreement
}

// Overlay bundles the adjudications keyed by id (for the per-leaf display) with the agreement `Result`
// (for the index summary), so a renderer takes one value — or nil, when the corpus has no adjudications.
type Overlay struct {
	By     map[string]Adjudication
	Result Result
}

// Agree compares each adjudication's verdict with the machine's verdict for the same leaf
// (`machine[id]`), counting matches and listing the mismatches in file order. `machine` is every judged
// leaf's verdict (assay.go builds it from the run's rows, id → pooled Faith), NOT only the argument-tree
// leaves — so an adjudication of a judged claim counts even when that claim sits off the argument tree.
// An adjudication whose id names no judged leaf (absent from `machine`) is IGNORED — not counted in `N`,
// not a disagreement — so a claim split or dropped since `adjudications.txt` was written falls out
// silently rather than showing a spurious `(no verdict)` disagreement. `N` is therefore the
// adjudications that hit a judged leaf, not the file's line count. It authors nothing — string equality
// is the whole test.
func Agree(machine map[string]string, adjs []Adjudication) Result {
	var r Result
	for _, a := range adjs {
		m, ok := machine[a.ID]
		if !ok {
			continue // adjudicates no rendered leaf: gated out
		}
		r.N++
		if m == a.Verdict {
			r.Agreed++
			continue
		}
		r.Disagreements = append(r.Disagreements, Disagreement{ID: a.ID, Human: a.Verdict, Machine: m, Reason: a.Reason})
	}
	return r
}

// NewOverlay builds the render overlay from the adjudications and the machine verdicts. It returns nil
// when there are no adjudications, so a corpus without a file renders exactly as it did before the
// feature existed.
func NewOverlay(machine map[string]string, adjs []Adjudication) *Overlay {
	if len(adjs) == 0 {
		return nil
	}
	by := make(map[string]Adjudication, len(adjs))
	for _, a := range adjs {
		by[a.ID] = a
	}
	return &Overlay{By: by, Result: Agree(machine, adjs)}
}
