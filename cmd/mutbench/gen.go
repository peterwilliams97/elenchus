// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/peterwilliams97/elenchus2/internal/mutbench"
)

// loadClaims reads a fixture's claims.jsonl.
func loadClaims(dir string) ([]mutbench.Claim, error) {
	path := filepath.Join(dir, "claims.jsonl")
	f, err := os.Open(path) // #nosec G304 -- operator-supplied fixture path
	if err != nil {
		return nil, fmt.Errorf("open fixture: %w", err)
	}
	defer func() { _ = f.Close() }()

	var out []mutbench.Claim
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if len(sc.Bytes()) == 0 {
			continue
		}
		var c mutbench.Claim
		if err := json.Unmarshal(sc.Bytes(), &c); err != nil {
			return nil, fmt.Errorf("%s line %d: %w", path, len(out)+1, err)
		}
		out = append(out, c)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%s: no claims", path)
	}
	return out, nil
}

// generate writes one input file per variant plus the manifest. Mutant and clean
// twin are both written: the clean twin is the over-catch measurement (A1.2),
// not an afterthought.
func generate(fixtureDir, outDir string, seed int, version string) error {
	claims, err := loadClaims(fixtureDir)
	if err != nil {
		return err
	}
	fixture := filepath.Base(fixtureDir)
	rows, err := mutbench.Generate(claims, fixture, seed, version)
	if err != nil {
		return err
	}
	inputs := filepath.Join(outDir, "inputs")
	if err := os.MkdirAll(inputs, 0o750); err != nil {
		return fmt.Errorf("mkdir %s: %w", inputs, err)
	}
	for _, r := range rows {
		if err := writeInput(inputs, r.AnchorClaimID+".mutant.txt", r.MutantText); err != nil {
			return err
		}
		if err := writeInput(inputs, r.AnchorClaimID+".clean.txt", r.CleanText); err != nil {
			return err
		}
	}
	if err := writeManifest(filepath.Join(outDir, "manifest.jsonl"), rows); err != nil {
		return err
	}
	scorable, quarantined := 0, 0
	for _, r := range rows {
		if mutbench.Scorable(r) {
			scorable++
		} else {
			quarantined++
		}
	}
	fmt.Fprintf(os.Stderr, "[gen] %d mutants + %d clean twins -> %s\n", len(rows), len(rows), inputs)
	fmt.Fprintf(os.Stderr, "[gen] %d scorable, %d quarantined (confound:arith, A2.1)\n", scorable, quarantined)
	fmt.Fprintf(os.Stderr, "[gen] injector: %s  seed: %d\n", version, seed)
	return nil
}

func writeInput(dir, name, text string) error {
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(text+"\n"), 0o600); err != nil {
		return fmt.Errorf("write %s: %w", p, err)
	}
	return nil
}

func writeManifest(path string, rows []mutbench.Row) error {
	f, err := os.Create(path) // #nosec G304 -- operator-supplied output path
	if err != nil {
		return fmt.Errorf("create manifest: %w", err)
	}
	defer func() { _ = f.Close() }()
	enc := json.NewEncoder(f)
	for _, r := range rows {
		if err := enc.Encode(r); err != nil {
			return fmt.Errorf("write manifest row %s: %w", r.MutantID, err)
		}
	}
	return nil
}
