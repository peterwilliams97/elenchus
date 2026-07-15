// SPDX-License-Identifier: Apache-2.0

// Command mutbench generates seeded mutants from a claims fixture and emits the
// A3.4 read sheet from the chains crossexam writes for them.
//
// It is a sibling binary, not a crossexam subcommand: crossexam has no
// subcommand dispatch — flag.Arg(0) is its input path — and adding one would
// change the input-resolution contract spec/CLI.md freezes (MUTATION_BENCH.md
// A1.4). Keeping the benchmark out of the shipped binary also makes §8's
// Goodhart guard structural: the graded pipeline cannot reach the benchmark.
//
//	mutbench gen   -fixture bench/fixtures/F-BASE-1 -out mutants/
//	mutbench sheet -manifest mutants/manifest.jsonl -chains eval/<dir> -out sheet.md
//
// gen and sheet are pure file->file: no network, no API key. Running the mutants
// through crossexam is a scored run and needs separate authorization (A3.7).
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const injectorVersion = "mutbench v0.1.0"

func main() {
	if len(os.Args) < 2 {
		fatal("usage: mutbench <gen|sheet> [flags]\n  gen    generate mutants + manifest from a fixture\n  sheet  emit the A3.4 read sheet from a manifest + chains")
	}
	var err error
	switch os.Args[1] {
	case "gen":
		err = runGen(os.Args[2:])
	case "sheet":
		err = runSheet(os.Args[2:])
	default:
		fatal(fmt.Sprintf("unknown subcommand %q; want gen or sheet", os.Args[1]))
	}
	if err != nil {
		fatal(err.Error())
	}
}

func runGen(args []string) error {
	fs := flag.NewFlagSet("gen", flag.ExitOnError)
	var fixture, out string
	var seed int
	fs.StringVar(&fixture, "fixture", "bench/fixtures/F-BASE-1", "fixture directory holding claims.jsonl")
	fs.StringVar(&out, "out", "mutants", "output directory for mutant inputs + manifest")
	fs.IntVar(&seed, "seed", 17, "generation seed, recorded in every manifest row")
	if err := fs.Parse(args); err != nil {
		return err
	}
	return generate(fixture, out, seed, stamp())
}

func runSheet(args []string) error {
	fs := flag.NewFlagSet("sheet", flag.ExitOnError)
	var manifest, chains, out string
	fs.StringVar(&manifest, "manifest", "mutants/manifest.jsonl", "manifest JSONL from gen")
	fs.StringVar(&chains, "chains", "", "directory of crossexam chain JSONL files (required)")
	fs.StringVar(&out, "out", "", "read sheet output path (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if chains == "" || out == "" {
		return fmt.Errorf("sheet: -chains and -out are both required")
	}
	return sheet(manifest, chains, out)
}

// stamp records the provenance every result row needs (A1.4): the commit the run
// was built from, and whether the tree was modified. A row from a dirty tree is
// not reproducible, so the flag is not optional.
func stamp() string {
	head := gitOut("rev-parse", "HEAD")
	if head == "" {
		return injectorVersion + " <no-git>"
	}
	dirty := ""
	if gitOut("status", "--porcelain") != "" {
		dirty = "-dirty"
	}
	return fmt.Sprintf("%s %s%s", injectorVersion, head[:min(7, len(head))], dirty)
}

func gitOut(args ...string) string {
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
