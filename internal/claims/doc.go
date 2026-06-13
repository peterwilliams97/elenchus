// Package claims holds the claim, fragment, and verdict types together with
// enum validation, JSON extraction, and decompose/split input handling.
//
// It is pure: no network, no global state. Verdict-enum validation happens at
// this boundary so malformed verdicts never reach orchestration or render.
package claims
