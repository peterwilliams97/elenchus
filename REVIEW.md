# Code review: the substance check

A review of crossexam's substance check, 2026-06-14. I changed no code; I only read it and ran it.
crossexam is meant to run four checks on a claim: substance, faithfulness, grounding, and audit.
Only substance is built; the other three don't exist yet. Substance asks whether a claim holds up
or is empty.

I read every Go file in full, including `internal/claims/claims.go`. I ran the toolchain. I ran the
binary against a real two-claim file with a live API key.

## What was checked

- Every Go file under `cmd/` and `internal/`, each read in full.
- `go build`, `go vet`, `go test -count=1`, `golangci-lint run`. Results below.
- A live `./crossexam` run on a two-line claims file, with `ANTHROPIC_API_KEY` set.
- Each file against the spec it implements: BEHAVIOR.md, CLI.md, PROMPTS.md.
- The tests against the BEHAVIOR.md GIVEN/WHEN/THEN examples. I broke each load-bearing test on
  purpose to confirm it fails.

## Toolchain — all green

Run with the matching Go binary:

```
go build ./...   → exit 0
go vet ./...     → exit 0
go test ./...    → ok (all 5 packages), exit 0
golangci-lint    → 0 issues, exit 0
```

Coverage: claims 100.0% · modes 92.5% · render 88.3% · client 88.2% · cmd 53.8%. The cmd figure is
low because cmd is `main()`, which is mostly wiring with the unbuilt features stubbed out. The rest
is well covered. No TODO/FIXME/!@#$ markers in the reviewed code.

I ran it on two claims: "Remote work boosts productivity by 40%" and "AI will replace most software
engineers within five years". Both came back `hollow`, downgraded through
`survives_only_by_conditioning`. The laundering note was appended to each reason. The SUMMARY and
USAGE blocks rendered. A well-formed `two.substance.jsonl` chain was written. The substance check
runs correctly from one end to the other.

## Environment note (not a code defect)

`go build ./...` fails out of the box here: `compile: version "go1.26.3" does not match go tool
version "go1.26.4"`. The cause is two Go installs. `/usr/local/bin/go` is Homebrew's 1.26.4, but a
`GOROOT=/usr/local/go` env var points it at a separate 1.26.3 toolchain, so the 1.26.4 driver runs
the 1.26.3 compiler. Everything passes when run with the matching binary (`/usr/local/go/bin/go`,
whose GOROOT already matches) or with `GOROOT` unset. This is a machine-setup issue, not a repo
problem. I record it so the "build is broken" symptom isn't read as a code fault.

## Verified against spec — correct, nothing to fix

- `modes/substance.go` AssayClaim vs BEHAVIOR.md Mode 1. The loop, the four-part continue condition
  (needs_another_round AND surviving!="" AND rounds<maxRounds-after-increment AND NOT
  survives_only_by_conditioning), the post-loop hollow downgrade with the appended note, and the
  unknown-verdict→error routing all match. Examples A/B/C each have a test that counts API calls.
- `client/client.go` vs BEHAVIOR.md. Retryable set {429,503,529}, 4 attempts, Retry-After honored,
  full-jitter backoff capped at 30s, truncation→error with usage billed before the stop_reason
  check, parse-then-retry-once-then-error. Examples I/J/K/L/N tested. maxTokens=1500 matches. A 120s
  client timeout means a stalled call can't hang the run. The loose-JSON comma stripper is
  quote-aware: a comma inside a string value is left intact (tested).
- `modes/prompts.go` vs PROMPTS.md §1–3. decomposeSys, producerSys, and substanceCriticSys are
  verbatim, including line breaks and the final JSON schema line.
- `cmd` vs CLI.md. Input priority (-text > file arg > piped stdin > defaultInput), fixtureName,
  -model default ($ANTHROPIC_MODEL or claude-sonnet-4-6), the SUMMARY/USAGE/Tier-1/chain formats,
  the verdictOrder, and the chain dir/filename (`eval/<stamp>-<model>/<fixture>.<mode>.jsonl`) all
  match.
- `render/disagreements.go`. The tier mapping and the surface rule match BEHAVIOR.md vcolor.
  Exhaustively tested: 64 triples plus an independent oracle. The live substance run doesn't
  exercise this path, but it is correct.

## Findings — file · problem · why it matters

- `internal/client/client_test.go` · `TestNonRetryableStatusIsError` has no teeth. It asserts that a
  400 yields an error but never counts HTTP attempts. The `seq` helper returns 400 on every call, so
  even if `retryable()` were changed to include 400, the loop would exhaust its retries and still
  return a 400 error. The test passes either way. I checked this: adding 400 to `retryable()` still
  passes the test. · Why it matters: this is the only test guarding "non-retryable codes break
  immediately." A regression that retried a hard 4xx would 4× the latency and cost on every
  permanent failure, and nothing would catch it. One-line fix: assert the attempt count is 1, the
  way `TestRetryExhaustion` asserts it is `retryMaxAttempts`.

- `claims.Axis` (claims.go) · `axisJSON` (modes/substance.go) · `axisJSON` (cmd/chain.go) · the same
  `{axis, finding, severity}` shape is declared three times, with two converters (`toAxes`,
  `axesOut`) moving data between them. · Why it matters: three copies of one shape can drift apart,
  and `claims.Axis` has no JSON tags so it can't be the wire type directly. Collapse to one tagged
  type once the chain DTO is allowed to share it. Cleanup, not a bug.

- `cmd/crossexam/main.go` · out-of-scope flags are rejected inconsistently. `-source/-evidence/
  -audit/-md` are declared and rejected with a clear "this build runs substance mode only" message.
  But `-v`/`-verbose`, `-no-color`, `-progress`, `-max-claims`, and `-usage-out` (all in CLI.md) are
  not declared at all, so passing one gives Go's generic `flag provided but not defined` and exit 2.
  `-usage-out` and `-no-color` would apply to a substance run. · Why it matters: a user reading
  CLI.md will try these. Four give a helpful message; five give a bare parser error. Either reject
  them the same way or note in the main() comment which flags this build does not accept. Low
  severity — loud either way, just uneven.

- `cmd/crossexam/main.go` · `maxRounds < 1` is silently clamped to 1. · Why it matters: this isn't
  in the spec, and a user who passes `-max-rounds 0` gets one round with no warning. Harmless, but
  the clamp should be a documented choice, not a silent one. Trivial.

- `cmd/crossexam/main.go` · the mode name `"substance"` is a bare string literal, passed twice — to
  `writeChain` and `render.Summary`. · Why it matters: the mode set (substance/faithfulness/
  grounding/audit) is the kind of closed, switched-on vocabulary the repo's standing rule and
  degradation log say should be one typed definition. It is the same shape as the verdict-literal
  debt already recorded. No impact today, since only one value exists. A consolidation note for when
  the other modes land, not a bug.

## Known, documented scope cuts (listed so they are not mistaken for gaps)

- `render/substance.go` Summary/UsageLine emit `est_usd n/a` where CLI.md shows a dollar amount and
  a `(rates YYYY-MM-DD)` suffix. The price table is not ported in this build. The code comment and
  CLI.md ("est_usd is null when the model is not in the table") both acknowledge it.
- The 60-second heartbeat (BEHAVIOR.md §Heartbeat, CLI.md `-progress`) is not implemented. Left out
  of this first end-to-end build.
- `render/substance.go` TermSubstance emits no ANSI colour, though spec/doc.go calls substance
  non-`-md` output "ANSI terminal." The function comment says so plainly ("a minimal readable
  rendering"). The `mdSubstance` markdown table is likewise not built (`-md` is rejected).
- `internal/claims/claims.go` verdict consts are deliberately untyped `string` rather than `Verdict`,
  with a comment explaining why: one spelling serves both the string result path and the
  Verdict-typed render maps without conversions. This meets the real goal of the standing rule — no
  re-spelled literals — even though it doesn't type the consts themselves. Defensible; flagged for
  the record.

## Tests vs BEHAVIOR.md examples

Every substance-relevant example is tested: A (TestAssayStopsAtMaxRounds), B
(TestAssayConditionLaunderingDowngrade), C (TestAssaySingleRoundExit), I (TestCallJSONFenceStripped),
J (TestCallJSONParseRetry), K (TestRetryAfterHonored), L (TestRetryExhaustion), M (TestSummaryBlock),
N (TestTruncationIsError). Examples D–H are faithfulness and evidence, not in this build, so
correctly no test. The load-bearing loop tests (A/B/C) all assert call counts, so the continue and
downgrade logic can't be silently inverted. The one test without teeth is
`TestNonRetryableStatusIsError`, above.

## Bottom line

The live substance path is correct against BEHAVIOR.md, CLI.md, and PROMPTS.md, and it runs from one
end to the other. The one finding with real consequence is the toothless
`TestNonRetryableStatusIsError`. The rest is one cleanup (the triplicated axis shape), small
consistency notes, and recorded scope. None of it blocks committing this work.
