# PREFLIGHT-2026-07-15 — Task 1 gate result (A2.2)

Gate result record. G1 is scored here (deterministic). **G2 is not evaluated** — it is a human read,
and the sheet is `READ-SHEET-2026-07-15.md`.

## Provenance

| Field | Value |
|---|---|
| Model | `claude-sonnet-4-6` (pinned via `-model`, never the default) |
| crossexam commit | `9bba615` (HEAD at run time) |
| Dirty tree | `false` |
| Binary | `./crossexam`, built 2026-07-12; no Go source has changed since |
| Flags | `-model claude-sonnet-4-6 -max-rounds 1 -quiet -chain-dir …` |
| Inputs | 8 (E1/E2/E3 × mutant+clean, M1 × mutant+clean) |
| Chains | scratchpad, not committed (gitignored per `.gitignore`) |
| Date | 2026-07-15 |

`-max-rounds 1` was used to keep the call budget on `decompose`, which is the only thing G1 and G2
measure. It cannot affect the result: `maxRounds` bounds `AssayClaim` only
(`internal/modes/substance.go:56-84`) and `Decompose` is called once per run, before the loop
(`cmd/crossexam/main.go:110`).

## G1 — cardinality-1 rate: **1/8 = 0.12. FAIL** (threshold ≥ 0.90)

| Input | Claims returned | G1 |
|---|---|---|
| E1-clean | 4 | FAIL |
| E1-mutant | 5 | FAIL |
| E2-clean | 4 | FAIL |
| E2-mutant | 3 | FAIL |
| E3-clean | 3 | FAIL |
| E3-mutant | 3 | FAIL |
| M1-clean | 2 | FAIL |
| M1-mutant | 1 | PASS |

**The single-sentence premise is false.** `decompose` splits a single indivisible sentence with one
matrix verb into 3–5 atomic claims. SESSION.md Item A's shape — built specifically so that no clause
boundary exists to cut — does not survive `decompose`, because `decompose` does not need a clause
boundary. It re-predicates. This kills the anchoring scheme in bench/PHASE1-PLAN.md §3(a): there is
no 1:1 mutant → graded-claim map, and per-claim invocation does not create one.

The one PASS is diagnostic, not encouraging: `M1-mutant` is the **criterion-deleted** MFC mutant, a
bare single clause with nothing left to split. Its clean twin `M1-clean` split 2 ways, separating the
claim from its kill condition — exactly the MFC anchoring problem flagged in PHASE1-PLAN.md §9.1. So
MFC's mutant anchors and its clean twin does not, which breaks the over-catch measurement, not the
recall one.

## The unexpected result: `decompose` split the move but did **not** dissolve it

This is the finding that matters, and it contradicts the expectation this whole gate was built on.
The 2026-07-12 run concluded `decompose` "dissolves the move before the critic sees it". On
single-sentence input it does not. It **isolates the transfer as its own atomic claim**:

- **E2-mutant** → `[2] The Frankfurt replica's p99 latency determines the service's p99 latency.`
- **E1-mutant** → `[4] The median checkout time of 1 minute 50 seconds measured at the Oxford Street flagship is the basis for the claim that the company's median checkout time is under two minutes.`
- **E3-mutant** → `[2] The author equates the pass rate among onsite candidates with the hiring funnel's overall pass rate.`

E3 is the striking one: `decompose` **named the equivocation itself**, unprompted, before any critic
saw it. Its prompt says only "break prose into atomic, independently-evaluable assertions … Do not
evaluate them" (`internal/modes/prompts.go:11-14`).

So the pre-registered cause recorded on 2026-07-12 — "decompose splits the flagship measurement from
the company conclusion, so the critic never sees the widening as one move" — **does not reproduce on
single-sentence input.** The widening is preserved, as a dedicated claim naming both scopes. Whether
the *critic* then catches it is the open question, and it is untouched by this gate.

Both readings of that are live and this gate cannot choose between them:

1. `decompose` is doing more of the work than the architecture claims, and the transfer claim is a
   better SCP-1 anchor than the fused sentence ever was.
2. `decompose` is a second grader in disguise. If it names equivocations, then a benchmark scoring
   "the critic" on post-decompose claims is not measuring what its result row says it measures — and
   E3's `[2]` is a finding emitted by a component the scoring path treats as mere preprocessing.

Reading 2 is the one that should worry us: it is the laundering risk relocated, not removed.

## Noise in the output, recorded not smoothed

`E1-mutant` also produced `[1] The Oxford Street store is located on Oxford Street.` — a vacuous
fragment invented from the possessive. This is the decontextualized-fragment skew already known from
PLAN.md §3(c). It inflates claim counts and would inflate any per-claim over-catch denominator.

## Consequences

- **G1 FAIL → A2.2 failure handling applies:** the anchoring scheme and A1.2 re-open before any other
  task runs. Phase 2 does not start.
- **G2 as worded is ill-posed.** It asks whether both scopes survive "in the returned claim",
  singular. No SCP-1 input returned a single claim. The read sheet asks the nearest well-posed
  question — both scopes in the returned claim **set** — and flags that reformulation as requiring
  ratification.
- **PHASE1-PLAN.md §3(a) is falsified** by its own pre-flight, which is what it was for. It cost 25
  calls to learn, before any harness existed.

## Not done

- G2 not evaluated (human read; sheet emitted).
- No Phase 2 code written.
- Nothing pushed.
