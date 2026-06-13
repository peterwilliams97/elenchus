# CLI.md — Flags, Env Vars, Input Forms, Output Formats

Extracted from `assay.go`. All flag defaults are read from the `flag.XxxVar` calls in `main`
(lines 109–123). Env-var behavior is from lines 109 and 126–129.

---

## Flags

| Flag | Type | Default | Source line | Description |
|---|---|---|---|---|
| `-model` | string | `$ANTHROPIC_MODEL` or `"claude-sonnet-4-6"` | 109 | Model ID passed to the API |
| `-source FILE` | string | `""` | 110 | Source transcript → faithfulness mode; combined with `-evidence` for proposition substitution |
| `-evidence` | bool | false | 111 | Evidence-grounding mode (enables web search) |
| `-audit` | bool | false | 112 | All three modes cross-tabulated; requires `-source` |
| `-text STRING` | string | `""` | 113 | Inline input text; takes priority over positional file arg and stdin |
| `-md` | bool | false | 114 | Emit markdown tables instead of ANSI terminal output |
| `-v` / `-verbose` | bool | false | 115–116 | Verbose: print every API call to stdout |
| `-no-color` | bool | false | 117 | Disable ANSI colour codes |
| `-max-rounds N` | int | 2 | 118 | Producer–critic rounds per claim (substance mode only) |
| `-max-claims N` | int | 0 | 119 | Cap the number of claims graded in evidence/audit (0 = unlimited) |
| `-progress` | bool | true | 120 | Per-claim progress lines to stderr (Tier 1) and 60 s heartbeat |
| `-quiet` | bool | false | 121 | Suppress per-case lines and heartbeat; SUMMARY block and Tier-2 JSONL are still written |
| `-usage-out FILE` | string | `""` | 122 | Append one JSON usage record per run to this file |
| `-chain-dir DIR` | string | `""` | 123 | Directory for Tier-2 JSONL verification chain; default is `eval/<stamp>-<model>/` |

### Flag interaction notes

- `-quiet` overrides `-progress`: `progressEnabled()` returns false when `quiet=true` (line 198).
- `-audit` without `-source` calls `fatal("-audit needs -source TRANSCRIPT")` (lines 170–172).
- `-md` is implicit for audit: `runAudit` always calls `mdAudit` regardless of the flag (line 408).
- `-v`/`-verbose` are aliases; both set the same `c.verbose` field (lines 115–116).

---

## Environment variables

| Variable | Required | Source line | Effect |
|---|---|---|---|
| `ANTHROPIC_API_KEY` | **yes** | 126–129 | API key; fatal if unset |
| `ANTHROPIC_MODEL` | no | 109 | Overrides the default model; superseded by `-model` flag |

---

## Input forms

Source: `readInput` (lines 1264–1278), priority order:

1. `-text STRING` — inline string passed directly (highest priority).
2. Positional argument — `flag.Arg(0)` is read as a file path via `mustRead`.
3. stdin — used only if stdin is not a character device (i.e., piped or redirected); empty stdin
   after trimming is ignored.
4. **Fallback** — if none of the above yields content, the hardcoded `defaultInput` constant is used
   (lines 661–664):
   ```
   1. The future of work will happen inside Codex or Claude Code.
   2. Every company will have one super-agent inside their Slack.
   3. SaaS is not dead — I would buy SaaS stocks right now.
   4. PMs will thrive in the AI era.
   ```

### Fixture name derivation (lines 134–136)

The base name used in SUMMARY output and Tier-2 JSONL filenames:
- Positional arg present: `strings.TrimSuffix(filepath.Base(arg), filepath.Ext(arg))`
- No positional arg: `"stdin"`

---

## Output formats

### stdout — results

| Mode | `-md` | Output |
|---|---|---|
| substance | false | ANSI terminal (`termSubstance`) |
| substance | true | Markdown table (`mdSubstance`) |
| faithfulness | false | ANSI terminal (`termFaith`) |
| faithfulness | true | Markdown table (`mdFaith`) |
| evidence | false | ANSI terminal (`termEvidence`) |
| evidence | true | Markdown table (`mdEvidence`) |
| audit | any | Always markdown (`mdAudit`), ignores `-md` flag |

Markdown table schemas:

**Substance** (lines 1092–1108):
```
## Substance — is the claim well-formed?

| # | Claim | Verdict | Why |
```

**Faithfulness** (lines 1110–1122):
```
## Faithfulness — did the source actually say it?

| # | Claim | Verdict | Source actually says |
```

**Evidence** (lines 1124–1143):
```
## Grounding — is it true, per external evidence?

| # | Claim | Verdict | Finding | Sources |
```
When a verdict was downgraded by `crossCheckEvidence`, the Finding cell appends:
`*(was: <OriginalVerdict>; <DowngradeReason>)*`

**Audit** (lines 1146–1162):
```
## Audit — three filters, cross-tabulated

| # | Claim | Faithful? | Substantive? | Grounded? |
```
Followed by a fixed pattern-reading table.

### stderr — progress and accounting

All progress output goes to stderr. stdout is strictly results.

**Tier 1 — per-case completion line** (lines 209–228, `progressDone`):
```
[NNN/NNN] ✓/✗ <verdict>        "<claim up to 60 chars>"  <elapsed>
```
- `✓` for any real verdict; `✗` for `"error"`.
- Suppressed when `-quiet` is set.

**Heartbeat** (lines 1571–1572, every 60 s unless `-quiet`):
```
[heartbeat] elapsed=Xs calls=N in=N out=N web=N est=<cost> verified=N errored=N seen=N/N claim="..."
```

**SUMMARY block** (lines 258–264, always emitted, even under `-quiet`):
```
SUMMARY <fixture> <mode>
  cases N · verified N · errored N
  <verdict1> N · <verdict2> N ...
  wall Xs · est_usd $N.NNNN (rates YYYY-MM-DD)
  detail: <chainFile>
```
- `<mode>` is one of: `substance`, `faithfulness`, `grounding`, `audit` (lines 150–157).
- `detail:` line is omitted if no chain file was written.
- Verdict ordering follows `verdictOrder` (lines 1056–1060):
  `supported, mixed, refuted, unverifiable, faithful, partial, overstated, absent, contradicted, substantive, hollow, skipped (over cap)`.
  Unknown verdicts appear after the ordered ones; `"error"` always last.

**USAGE line** (lines 1599–1604, always emitted after SUMMARY):
```
USAGE model=<model> calls=N in=N out=N cache_read=N cache_create=N web_searches=N wall=Ns est_usd=<cost>
```

---

## Tier-2 JSONL verification chain

Source: lines 139–161 (chain-dir setup), 1280–1433 (record types and appenders)

**Directory:** `-chain-dir` value, or `eval/<stamp>-<model>/` where `<stamp>` is
`time.Now().Format("20060102-1504")` (lines 141–142).

**Filename:** `<fixture>.<mode>.jsonl` (line 160), where `<mode>` is `substance`, `faithfulness`,
`grounding`, or `audit`.

**Record envelope** (lines 1284–1292):
```json
{
  "idx": integer,
  "total": integer,
  "mode": string,
  "claim": string,
  "verdict": string,
  "elapsed_s": float,
  "detail": { ... mode-specific ... }
}
```

**Detail schemas by mode:**

*Substance* (`substanceDetail`, lines 1317–1324):
```json
{"steelman":string,"critique_by_axis":[...],"surviving_claim":string,"added_conditions":int,"rounds":int,"reason":string}
```

*Faithfulness* (`faithDetail`, lines 1326–1331):
```json
{"defender_support":string,"critic_finding":string,"distortion_type":string,"source_says":string}
```

*Grounding* (`evidenceDetail`, lines 1333–1341):
```json
{"finding":string,"sources":[...],"retrieved_sources":[...],"sources_verified":int,"downgrade_reason":string,"original_verdict":string,"error_cause":string}
```

*Audit* (`auditDetail`, lines 1343–1347): nested faith + substance + evidence detail objects under `"faith"`, `"substance"`, `"evidence"` keys.

Chain write failures are logged to stderr and silently dropped — they never abort the run (lines 1305–1314).

---

## `-usage-out FILE` JSON record schema

Source: lines 1607–1637

One record appended per run:
```json
{
  "model": string,
  "calls": integer,
  "input_tokens": integer,
  "output_tokens": integer,
  "cache_read_tokens": integer,
  "cache_create_tokens": integer,
  "web_searches": integer,
  "wall_seconds": float,
  "est_usd": string|null
}
```
`est_usd` is null when the model is not in the price table.

---

## Fatal conditions

Source: `fatal` (lines 1251–1254), `mustRead` (lines 1256–1261), `main`

| Condition | Message |
|---|---|
| `ANTHROPIC_API_KEY` unset | `"set ANTHROPIC_API_KEY in your environment first."` |
| `-audit` without `-source` | `"-audit needs -source TRANSCRIPT"` |
| Input file unreadable | `"cannot read <path>: <err>"` |
| `-source` file unreadable | same |
| `decompose` fails (substance mode) | `"decompose failed: <err>"` |

All fatal errors write to stderr and call `os.Exit(1)`.

---

## Verbose output (`-v` / `-verbose`)

Source: lines 709–713 (call header), 786–789 (search queries), 812–819 (response), lines 1300–1302 (chain)

When verbose, `callClaude` prints to stdout:
```
┌─ call · <model> [(web search)]
│ system:
│   <system prompt with newlines indented>
│ user:
│   <user prompt with newlines indented>
├─ response:
│ <response text>
│ retrieved N source(s)        ← only if sources were retrieved
└─
```
Chain append events are logged to stderr: `[chain] <file> idx=N verdict=<v>`.
