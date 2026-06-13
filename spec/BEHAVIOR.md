# BEHAVIOR.md — Orchestration Behavioral Contract

Extracted from `assay.go`. Describes what the program does, not how the Go is structured. Each rule
is traceable to a source line range and includes a GIVEN/WHEN/THEN example concrete enough to
become a test.

---

## Mode routing

Source: lines 168–184 (the `switch` in `main`).

| Flags present              | Mode         | Runner |
|----------------------------|--------------|--------|
| neither                    | substance    | `runSubstance` |
| `-source FILE`             | faithfulness | `runFaithfulness` |
| `-evidence`                | evidence     | `runEvidence` |
| `-evidence -source FILE`   | evidence (with proposition substitution) | `runEvidence` |
| `-audit -source FILE`      | audit (all three) | `runAudit` |
| `-audit` without `-source` | **fatal**    | `os.Exit(1)` |

---

## Mode 1 — Substance (`runSubstance`)

Source: lines 269–290 (`runSubstance`), 415–421 (`decompose`), 425–463 (`assayClaim`)

### Call sequence

1. `decompose(input)` → `[]string` of atomic claims.
   - Calls `callJSON(decomposeSys, "TEXT:\n"+input, false, &arr)`.
   - On error: `fatal("decompose failed: …")`.

2. For each claim: `assayClaim(claim)` → `substance`.

### `assayClaim` loop (lines 428–450)

```
current = claim
rounds  = 0
for rounds < maxRounds:
    call producerSys  → {steelman, conditions}
    call substanceCriticSys with CLAIM + STEELMAN + CONDITIONS → substanceJSON
    rounds++
    if last.NeedsAnother
       AND last.SurvivingClaim != ""
       AND rounds < maxRounds
       AND NOT last.SurvivesOnlyByConditions:
        current = last.SurvivingClaim
        continue
    break
```

**Continue condition** (all four must be true, lines 443–448):
1. `needs_another_round == true`
2. `surviving_claim != ""`
3. `rounds < maxRounds` (checked *after* increment, so at `maxRounds=2`: only round 0 can continue)
4. `survives_only_by_conditioning == false`

**Post-loop hollow downgrade** (lines 451–456):
If `last.SurvivesOnlyByConditions == true` after the loop exits (regardless of cause), force
`verdict = "hollow"` and append `" Survives only by conditions the speaker never stated."` to
`reason`.

### GIVEN/WHEN/THEN examples

**Example A — loop terminates at maxRounds**

GIVEN `maxRounds=2`, round 0 returns `needs_another_round=true, surviving_claim="narrowed X", survives_only_by_conditioning=false`
WHEN round 1 completes (rounds is now 2, equals maxRounds)
THEN the loop breaks; no third call is made; verdict is whatever round 1 returned.

**Example B — condition-laundering downgrade**

GIVEN a claim where the critic's only surviving form requires an invented qualifier
WHEN `survives_only_by_conditioning=true` on any round
THEN the loop breaks immediately (does not continue), AND at loop exit `verdict` is overwritten to `"hollow"`.

**Example C — single-round early exit**

GIVEN `needs_another_round=false` on round 0
WHEN round 0 completes
THEN the `if` block is not entered; the loop breaks; total API calls for this claim = 2 (one producer + one critic).

---

## Mode 2 — Faithfulness (`runFaithfulness`)

Source: lines 292–307 (`runFaithfulness`), 469–485 (`faithClaim`)

### Call sequence

1. `splitSummary(input)` → `[]string` (regex-based, NOT decompose).

2. For each claim: `faithClaim(claim, src)` → `faith`.
   a. `callJSON(faithDefenderSys, "SUMMARY CLAIM:\n"+claim+"\n\nSOURCE:\n"+src, false, &d)` → `{found, quotes}`.
   b. Format quotes as `"- q1\n- q2"` (or `"(none)"` if empty, line 474).
   c. `callJSON(faithCriticSys, assembled_prompt, false, &fj)` → `{findings, verdict, evidence, what_source_actually_says}`.
   d. Return `faith{Claim, Verdict: fj.Verdict, Evidence: fj.Evidence, SourceSays: fj.SourceSays}`.

3. No loop; each claim is always exactly two API calls.

### GIVEN/WHEN/THEN

**Example D — defender finds no support**

GIVEN a claim not present in the source
WHEN defender returns `{found: false, quotes: []}`
THEN critic prompt contains `DEFENDER FOUND SUPPORT: false\nDEFENDER QUOTES:\n(none)`; critic verdict is expected to be `"absent"` or `"contradicted"`; `SourceSays` field will be empty.

**Example E — overstated with source reconstruction**

GIVEN the source hedged ("might") but the claim is absolute
WHEN critic returns `verdict="overstated", what_source_actually_says="the speaker said it might…"`
THEN `faith.SourceSays` is populated with that string; it will be used as the proposition in evidence mode if `-source` is combined with `-evidence`.

---

## Mode 3 — Evidence (`runEvidence`)

Source: lines 324–354 (`runEvidence`), 492–506 (`evidenceClaim`), 313–318 (`intendedProposition`)

### Call sequence

1. `splitSummary(input)` → `[]string`.

2. For each claim at index `i`:
   a. If `maxClaims > 0 && i >= maxClaims`: record verdict `"skipped (over cap)"`, skip to next.
   b. If `src != ""`: run `faithClaim(claim, src)` → `fc`; set `proposition = intendedProposition(fc, claim)`.
   c. Else: `proposition = claim`.
   d. `evidenceClaim(proposition)` → `evidence` struct (then `r.Claim = cl` to keep original label).
   e. `crossCheckEvidence(out)` → possibly downgraded evidence.

### Proposition substitution rule (lines 313–318)

`intendedProposition(fc, claim)` returns `fc.SourceSays` when:
- `fc.SourceSays != ""`  **AND**
- `fc.Verdict == "partial"` **OR** `fc.Verdict == "overstated"`

Otherwise returns `claim` unchanged.

The substitution grounds what the speaker actually meant, not a literalized paraphrase of the
summary. The displayed `Claim` field is always the original claim string (line 344 `r.Claim = cl`).

### `evidenceClaim` internals (lines 492–506)

1. `callJSONSourced(evidenceSys, "CLAIM:\n"+proposition, withTools=true, &e)` → `([]retrievedSource, error)`
2. Build `evidence` struct from `e.Verdict`, `e.Finding`, model-cited `e.Sources`, and `rs` (actually retrieved URLs).
3. Call `crossCheckEvidence(out)` before returning.

### `crossCheckEvidence` — downgrade rules (lines 513–546)

Checks URL *provenance* only (was the page retrieved), never content support.
Exempt from downgrade: verdicts `"unverifiable"` and `"error"` (line 514).

| Condition | `DowngradeReason` |
|---|---|
| `len(RetrievedSources) == 0` | `"no sources retrieved; verdict is model self-report"` |
| `len(Sources) == 0` (model cited no URLs) | `"no URLs cited in response"` |
| none of cited URLs match retrieved set | `"claimed sources not present in retrieval"` |

In all three cases: `OriginalVerdict = prior Verdict`, `Verdict = "unverifiable"`.

URL matching uses `normalizeURL` (lines 552–558): lowercase host + path, scheme/query/fragment
stripped, trailing slash removed. Matching is host+path — same domain but different path does NOT
match.

### GIVEN/WHEN/THEN

**Example F — proposition substitution**

GIVEN `-evidence -source transcript.txt summary.txt`, claim "X will dominate", source says "X might grow"
WHEN faithClaim returns `{Verdict:"overstated", SourceSays:"X might grow"}`
THEN `evidenceClaim` receives `"X might grow"` as its claim, not `"X will dominate"`. The evidence
row's displayed `Claim` column still shows `"X will dominate"`.

**Example G — crossCheck downgrade**

GIVEN the model returns `{verdict:"supported", sources:[{url:"https://example.com/page"}]}`
AND the `web_search_tool_result` block contained `[{url:"https://other.com/page"}]`
WHEN `crossCheckEvidence` runs
THEN `normalizeURL("example.com/page")` != `normalizeURL("other.com/page")`, `matched=0`, verdict
becomes `"unverifiable"`, `DowngradeReason="claimed sources not present in retrieval"`,
`OriginalVerdict="supported"`.

**Example H — no retrieval at all**

GIVEN the API call succeeded but no `web_search_tool_result` blocks appeared
WHEN `crossCheckEvidence` runs with `RetrievedSources=[]`
THEN verdict becomes `"unverifiable"` regardless of what the model returned.

---

## Mode 4 — Audit (`runAudit`)

Source: lines 384–409

### Call sequence

1. `splitSummary(input)` → `[]string`.
2. For each claim: run `faithClaim`, `assayClaim`, and `evidenceClaim` (with `intendedProposition`)
   in sequence.
3. Output is always markdown (`mdAudit`), regardless of `-md` flag.
4. `progressDone` receives a combined verdict string: `"faith=<v> sub=<v> ev=<v>"` (line 403).

**Note:** Audit uses `assayClaim` directly (no decompose step), same as faithfulness/evidence.

---

## JSON parse + retry rule

Source: lines 675–695 (`callJSONSourced`)

1. Call API → raw text.
2. `unmarshalLoose(out, v)` — strips markdown fences, removes trailing commas, then `json.Unmarshal`.
3. If parse succeeds → return.
4. If parse fails → **retry once** with a modified system prompt:
   ```
   <original system> + "\n\nReturn ONLY raw JSON. No prose, no markdown, no backticks. First character must be { or [."
   ```
5. If second parse fails → return the error (caller records verdict `"error"`).

There is no third attempt. The retry uses the same `prompt` but a different `system`.

### GIVEN/WHEN/THEN

**Example I — one parse failure allowed**

GIVEN the model returns `\`\`\`json\n{"verdict":"hollow"}\n\`\`\``
WHEN `unmarshalLoose` is called
THEN `extractJSON` strips the fence markers, parse succeeds; no retry occurs.

**Example J — retry triggered**

GIVEN the model returns `"The verdict is hollow."` (no JSON)
WHEN `unmarshalLoose` fails
THEN the system is augmented and the API is called a second time; if that response parses, success.

---

## HTTP retry / backoff policy

Source: lines 721–738 (`callClaude` loop), 824 (`retryable`), 829–838 (`retryDelay`), line 36 (`retryMaxAttempts=4`)

- **Retryable HTTP status codes:** 429, 503, 529.
- **Max attempts:** 4 (1 initial + 3 retries). Loop condition: `attempt >= retryMaxAttempts-1`.
- **Delay per retry:** Honor `Retry-After` header (integer seconds) if present and positive.
  Else: full-jitter exponential backoff — pick uniformly from `[0, min(retryBase << attempt, 30s)]`.
  - attempt 0→1: `[0, 1s]`
  - attempt 1→2: `[0, 2s]`
  - attempt 2→3: `[0, 4s]`
  - cap at 30 s.
- Non-retryable codes break immediately regardless of attempt count.
- `retryBase = 0` in tests (line 41), so all delays collapse to 0.

### GIVEN/WHEN/THEN

**Example K — 429 with Retry-After**

GIVEN the API returns HTTP 429 with `Retry-After: 10`
WHEN `retryDelay` is called
THEN sleep 10 s (exactly), then retry.

**Example L — exhausted retries**

GIVEN the API returns 503 on all four attempts
WHEN `attempt == retryMaxAttempts-1 == 3` on the fourth try
THEN the loop breaks; the 503 response body is parsed; if it parses as an error, that error is returned to the caller.

---

## Verdict enums and routing

Source: line 84 (`runTally.record`), lines 1181–1190 (`vcolor`), lines 1062–1064 (`badVerdict`)

### Substance mode
`"substantive"` | `"partial"` | `"hollow"` | `"error"`

### Faithfulness mode
`"faithful"` | `"partial"` | `"overstated"` | `"absent"` | `"contradicted"` | `"error"`

### Evidence mode
`"supported"` | `"mixed"` | `"refuted"` | `"unverifiable"` | `"skipped (over cap)"` | `"error"`

### Terminal color routing (lines 1181–1190)

| Color | Verdicts |
|---|---|
| green | `substantive`, `faithful`, `supported` |
| yellow | `partial`, `overstated`, `mixed` |
| red | `hollow`, `absent`, `refuted`, `contradicted` |
| grey | all others (including `unverifiable`, `error`, `skipped`) |

### `badVerdict` (markdown bold, lines 1062–1067)
Verdicts bolded in markdown output: `hollow`, `absent`, `refuted`, `contradicted`, `error`.

---

## Error verdict rule

Source: lines 82–91 (`runTally.record`)

- `"error"` increments `counts["error"]` but does NOT increment `verified`.
- Every other verdict increments `verified` AND `counts[verdict]`.
- `errored = total - verified` in the SUMMARY block (line 233).
- An error case is never counted as a win in any metric.

### GIVEN/WHEN/THEN

**Example M**

GIVEN a run of 3 claims, verdicts `["substantive", "error", "hollow"]`
WHEN `runTally.snapshot()` is called
THEN `total=3, verified=2, counts={"substantive":1,"hollow":1,"error":1}`, `errored=1`.

---

## `maxTokens` and truncation handling

Source: lines 35, 702, 771–775

- Every API call sets `max_tokens = 1500` (constant, line 35).
- If the response has `stop_reason == "max_tokens"` (line 771), `callClaude` returns:
  ```
  fmt.Errorf("response truncated: stop_reason=max_tokens (limit=%d tokens); raise maxTokens constant", maxTokens)
  ```
- This error propagates to `callJSONSourced`, which returns it; the caller records verdict `"error"`.
- Usage is still accumulated before checking stop_reason (line 748) — a truncated response was billed.

### GIVEN/WHEN/THEN

**Example N**

GIVEN the model's response would require 2000 tokens but `max_tokens=1500`
WHEN the API returns `stop_reason="max_tokens"`
THEN the claim's verdict is `"error"`, reason is the truncation message, and the token cost of the
partial response is still counted in `usageCounters`.

---

## Heartbeat

Source: lines 1575–1593 (`startHeartbeat`), 1546–1573 (`heartbeatLine`)

A background goroutine fires every **60 seconds** and prints to stderr:
```
[heartbeat] elapsed=Xs calls=N in=N out=N web=N est=<cost> verified=N errored=N seen=N/N claim="..."
```
Suppressed when `-quiet` is set (line 164). The goroutine is stopped via a `done` channel closed by
the returned cancel func (called before `printSummary`).
