// Command assay parses flags and wires the packages together per spec/CLI.md.
//
// It owns flag parsing and run-state assembly only; the evaluation logic lives
// in internal/modes, the API in internal/client, and output in internal/render.
package main

// main is a no-op stub. A main package cannot link without it, so it exists
// only to keep `go build ./...` green this session. All flag parsing and
// wiring is the cmd/assay session's work (PLAN.md §1.4); none is added here.
func main() {}
