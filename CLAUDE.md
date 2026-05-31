# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`assay.py` — a single-file CLI that runs prose through a three-stage dialectical filter using the
Claude API: **decompose** (extract atomic claims) → **dialectic** (producer↔critic loop per claim)
→ **filter** (surface the substantive residue).

Three operating modes:

- **Dialectic** (default) — assay prose for logical integrity
- **Faithfulness** (`--source`) — check whether a summary accurately represents a source transcript
- **Evidence-grounding** (`--evidence`) — check each claim against external evidence via web search

## Running it

```sh
# install dependency
pip install anthropic

# set key (or use setup.sh, but note it contains a live key — don't commit changes to it)
export ANTHROPIC_API_KEY=sk-ant-...

# run with built-in example
python3 assay.py

# run on a file
python3 assay.py path/to/prose.txt

# run on inline text
python3 assay.py --text "Our AI-first strategy will..."

# pipe stdin
echo "some prose" | python3 assay.py

# faithfulness mode: does SUMMARY faithfully represent SOURCE?
python3 assay.py summary.txt --source transcript.txt

# evidence-grounding mode: is each claim actually true?
python3 assay.py claims.txt --evidence

# options
python3 assay.py --model claude-opus-4-8 --max-rounds 3 --no-color --verbose
```

`ANTHROPIC_MODEL` env var overrides the default model (`claude-sonnet-4-6`).

## Architecture

Everything lives in `assay.py`. The call graph by mode:

**Dialectic (default)**
```
main()
  read_source()          # args.text | args.path | stdin | DEFAULT_INPUT
  decompose(source)      # → list[str] of atomic claims via call_json()
  for claim in claims:
    assay_claim(claim, max_rounds)
      produce(claim)     # steelman, blind to critique axes
      critique(claim, steelman, conditions)  # 7 fixed axes → verdict + surviving_claim
      loop if needs_another_round and rounds < max_rounds
  print_report(results)
```

**Faithfulness (`--source TRANSCRIPT`)**
```
run_faithfulness(summary, source)
  split_summary(summary_text)   # splits numbered/bulleted lists or lines
  for claim:
    faithfulness_claim(claim, source)
      faithfulness_defender()   # find strongest verbatim support in source
      faithfulness_critic()     # judge accuracy: faithful|partial|overstated|absent|contradicted
  print_faithfulness_report(results)
```

**Evidence-grounding (`--evidence`)**
```
run_evidence(text)
  split_summary(text)
  for claim:
    evidence_claim(claim)
      evidence_ground()         # web search via call_claude_tools(); verdict: supported|mixed|refuted|unverifiable
  print_evidence_report(results)
```

**Key design constraint:** the Producer prompt deliberately does NOT include the critique axes
(Evidence, Hidden premise, Falsifiability, Equivocation, Base rate/magnitude, Counterexample,
Causality vs correlation). This blindness is intentional — the Producer must steelman without
knowing how it will be attacked.

**API plumbing:** `call_claude()` streams from the Anthropic API. `call_json()` wraps it with
retry logic — on JSON parse failure it appends a strict JSON-only instruction and retries once.
`parse_json()` handles markdown fences and trailing commas before calling `json.loads`.
`call_claude_tools()` is used only by evidence-grounding mode; it enables the `web_search_20250305`
tool and reads `resp.content` blocks (non-streaming).

**Verdicts:**
- Dialectic: `"substantive"` | `"partial"` | `"hollow"` | `"error"`. Only `substantive` and
  `partial` appear in the final residue. `partial` claims carry a `surviving_claim` with the
  narrowed defensible version.
- Faithfulness: `"faithful"` | `"partial"` | `"overstated"` | `"absent"` | `"contradicted"`
- Evidence: `"supported"` | `"mixed"` | `"refuted"` | `"unverifiable"`

**`--verbose`** (`VERBOSE` global) controls whether streaming API calls print their full
exchange to the terminal. Off by default; in default mode the dialectic stages render their
own structured output instead. In verbose mode `print_report` renders the full per-claim
breakdown; in default mode it only shows the headline + residue (per-round detail was already
shown live).
