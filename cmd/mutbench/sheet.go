// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/peterwilliams97/elenchus2/internal/mutbench"
)

// sheetHeader is the A3.4 read protocol, verbatim where the amendment is
// verbatim. The calibration boundary pair is quoted exactly: it is the whole
// catch standard, and paraphrasing it would move the standard.
const sheetHeader = `# READ SHEET — target-catch (A3.4 protocol)

Record every row **before** looking at any pooled result. Rows are in randomized input order.

For each row you see the full chain — decompose output, critic critique, verdict — plus its manifest
row, and you record three things:

- **catch (y/n)** — did the chain text identify the specific injected defect?
- **catch-source** — decompose | critic | verdict. A diagnostic only: it never aggregates into a
  per-component number (A3.1).
- **note** — free text.

## Catch standard

The chain text must identify the **specific defect**. Neutral re-predication of the inferential move
is NOT a catch; meta-level naming of the move IS. The boundary, verbatim from A3.4:

- **NOT a catch** — E2 [2]: "The Frankfurt replica's p99 latency determines the service's p99 latency."
- **Catch** — E3 [2]: "The author equates the pass rate among onsite candidates with the hiring funnel's overall pass rate."

## What this sheet is not

The axis-fired figure printed per row is a **harness diagnostic**. It is not recall, it never becomes
recall, and a row where an axis fired is not thereby a catch (A-T4). On 2026-07-12 a keyword tally
certified 5/5 on a probe the human read scored 0/5. That is what this sheet exists to prevent.

---
`

func readManifest(path string) ([]mutbench.Row, error) {
	f, err := os.Open(path) // #nosec G304 -- operator-supplied manifest path
	if err != nil {
		return nil, fmt.Errorf("open manifest: %w", err)
	}
	defer func() { _ = f.Close() }()
	var rows []mutbench.Row
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if len(sc.Bytes()) == 0 {
			continue
		}
		var r mutbench.Row
		if err := json.Unmarshal(sc.Bytes(), &r); err != nil {
			return nil, fmt.Errorf("manifest line %d: %w", len(rows)+1, err)
		}
		rows = append(rows, r)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	return rows, nil
}

// blindOrder returns manifest rows in a deterministic, non-obvious order keyed by
// the mutant hash. Deterministic beats random here: the sheet is reproducible and
// the ordering still leaks nothing, since the hash is not correlated with class.
func blindOrder(rows []mutbench.Row) []mutbench.Row {
	out := append([]mutbench.Row(nil), rows...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].MutantSHA256 < out[j].MutantSHA256 })
	return out
}

func loadChain(dir, claimID, variant string) ([]mutbench.ChainRec, error) {
	p := filepath.Join(dir, fmt.Sprintf("%s.%s.substance.jsonl", claimID, variant))
	f, err := os.Open(p) // #nosec G304 -- operator-supplied chain dir
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return mutbench.ReadChain(f)
}

func sheet(manifestPath, chainDir, outPath string) error {
	rows, err := readManifest(manifestPath)
	if err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString(sheetHeader)

	n, missing := 0, 0
	for _, r := range blindOrder(rows) {
		for _, variant := range []string{"mutant", "clean"} {
			recs, err := loadChain(chainDir, r.AnchorClaimID, variant)
			if err != nil {
				missing++
				continue
			}
			n++
			writeRow(&b, n, r, variant, recs)
		}
	}
	if n == 0 {
		return fmt.Errorf("no chains found under %s: run crossexam over the generated inputs first (a scored run needs separate authorization)", chainDir)
	}
	if err := os.WriteFile(outPath, []byte(b.String()), 0o600); err != nil {
		return fmt.Errorf("write sheet: %w", err)
	}
	fmt.Fprintf(os.Stderr, "[sheet] %d rows -> %s\n", n, outPath)
	if missing > 0 {
		// Never silently truncate: a short sheet must announce what it dropped.
		fmt.Fprintf(os.Stderr, "[sheet] WARNING: %d inputs had no chain and are absent from the sheet\n", missing)
	}
	return nil
}

func writeRow(b *strings.Builder, n int, r mutbench.Row, variant string, recs []mutbench.ChainRec) {
	fmt.Fprintf(b, "\n## R%d\n\n", n)
	fmt.Fprintf(b, "Manifest: class `%s` · anchor `%s` · variant `%s` · stage `%s`\n",
		r.DefectClass, r.AnchorClaimID, variant, r.StageUnderTest)
	if !mutbench.Scorable(r) {
		fmt.Fprintf(b, "Tags: `%s` — **quarantined: read it, but it never pools into the class number (A2.1)**\n",
			strings.Join(r.Tags, "`, `"))
	}
	if r.DefectClass == "SCP-1" && variant == "mutant" {
		fmt.Fprintf(b, "Injected: %s\n", r.Note)
		fmt.Fprintf(b, "Narrow scope: %q · Broad subject: %q\n", r.Scopes.Narrow, r.Scopes.Broad)
	}
	if r.DefectClass == "MFC" && variant == "mutant" {
		fmt.Fprintf(b, "Injected: %s\n", r.Note)
	}
	if variant == "clean" {
		b.WriteString("**Clean twin — no defect injected. A catch here is an over-catch (A1.2).**\n")
	}
	fmt.Fprintf(b, "\nInput: %q\n", inputText(r, variant))
	fmt.Fprintf(b, "\nDiagnostics (not recall): axis-fired=%v · decompose cardinality=%d\n\n",
		mutbench.AxisFired(recs), mutbench.Cardinality(recs))
	b.WriteString("Full chain:\n\n")
	for _, rec := range recs {
		fmt.Fprintf(b, "  [%d] claim: %s\n", rec.Idx, rec.Claim)
		fmt.Fprintf(b, "      verdict: %s\n", rec.Verdict)
		for _, a := range rec.Detail.CritiqueByAxis {
			fmt.Fprintf(b, "      axis %s (%s): %s\n", a.Axis, a.Severity, a.Finding)
		}
		if rec.Detail.Reason != "" {
			fmt.Fprintf(b, "      reason: %s\n", rec.Detail.Reason)
		}
		b.WriteString("\n")
	}
	b.WriteString("catch (y/n): ______   catch-source (decompose|critic|verdict): ______\n\n")
	b.WriteString("note: ____________________________________________\n")
}

func inputText(r mutbench.Row, variant string) string {
	if variant == "clean" {
		return r.CleanText
	}
	return r.MutantText
}
