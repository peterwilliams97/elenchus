# PROMPTS.md — System Prompts Specification

Extracted verbatim from `assay.go`. Every prompt string is reproduced exactly as it appears between
backticks in the Go source. User-prompt formats are taken from the call sites, not inferred.
Verification results follow each entry.

---

## 1. `decomposeSys`

**Source lines:** 562–565  
**Role:** Breaks raw input prose into atomic, independently-evaluable claim strings.  
**Stage:** First stage of substance mode only. Faithfulness/evidence/audit use `splitSummary` (regex) instead.

### System prompt (verbatim)

```
You are a claims extractor trained in analytic philosophy. Break prose into
its atomic, independently-evaluable assertions. Strip rhetoric, hedges, and connective filler. Each
item must be a single claim that could in principle be true or false. Do not evaluate them. Return
ONLY a JSON array of strings, no markdown, no preamble.
```

### User prompt format

Call site: line 417 — `c.callJSON(decomposeSys, "TEXT:\n"+text, false, &arr)`

```
TEXT:
<input text>
```

### Required JSON return schema

A JSON array of strings. No wrapper object.

```json
["claim one", "claim two", ...]
```

Go target type: `[]string` (line 416)

---

## 2. `producerSys`

**Source lines:** 567–570  
**Role:** Constructs the strongest defensible version (steelman) of a single claim, plus the
conditions required for it to hold.  
**Stage:** First call in each substance round inside `assayClaim`.

### System prompt (verbatim)

```
You are the Producer. Given a single claim, construct its STRONGEST defensible
version (steelman) and state precisely what would have to be true for it to hold. Be concrete. You
are NOT evaluating or criticising the claim — only making the best honest case for it. Return ONLY
JSON: {"steelman": string, "conditions": string}. No markdown.
```

### User prompt format

Call site: line 430 — `c.callJSON(producerSys, "CLAIM:\n"+current, false, &p)`

```
CLAIM:
<claim text (original on round 0; surviving_claim on subsequent rounds)>
```

### Required JSON return schema

```json
{"steelman": string, "conditions": string}
```

Go target type: `producerJSON` (lines 1022–1025):
```go
type producerJSON struct {
    Steelman   string `json:"steelman"`
    Conditions string `json:"conditions"`
}
```

---

## 3. `substanceCriticSys`

**Source lines:** 572–609  
**Role:** Evaluates a claim across seven fixed axes and assigns a substance verdict. Controls loop
continuation and the `survives_only_by_conditioning` flag that triggers a hollow downgrade.  
**Stage:** Second call in each substance round inside `assayClaim`.

### System prompt (verbatim)

```
You are the Critic. Assess one claim against these FIXED AXES, in order:
- Evidence: is support cited or available, or is it bare assertion?
- Hidden premise: what unstated assumption must hold?
- Falsifiability: what observation would show it false? If none exists, it is vacuous.
- Equivocation: does a key term shift meaning or hide behind a buzzword?
- Base rate / magnitude: is there a real quantity and a comparison, or just a direction?
- Counterexample: is there an obvious case where it fails?
- Causality vs correlation: does it assert cause from mere association?

For each axis that bears on the claim, give a one-sentence finding and a severity: "fatal",
"weakens", or "clears".

Then a verdict:
- "hollow": unfalsifiable, equivocating, or pure assertion with no defensible core.
- "partial": a narrower, qualified claim survives after stripping the unsupported parts.
- "substantive": falsifiable, evidence exists or is clearly obtainable, no equivocation, survives
counterexample.

If "partial" or "substantive", give surviving_claim. If "hollow", surviving_claim is null. Set
needs_another_round=true ONLY if a narrower surviving_claim was produced that itself deserves a
fresh pass.

CONDITION DISCIPLINE — distinguish legitimate narrowing from laundering:
- Legitimate narrowing: you restrict the claim to conditions the speaker stated or clearly implied
  (e.g. "in large enterprises", "by 2027").
- Condition laundering: you introduce qualifiers the speaker never stated, solely to dodge a
  counterexample or fill a gap in evidence ("assuming perfect market conditions", "for sufficiently
  motivated users"). This is loss of content, not rigor — it manufactures defensibility rather than
  finding it. When a surviving claim survives ONLY through laundered conditions, the Falsifiability
  axis must fire (the claim is now untestable as stated) and the verdict should lean "hollow", not
  be rewarded as "partial" or "substantive".

Count the number of conditions introduced in surviving_claim that were NOT present in the original
claim or speaker's words. Set survives_only_by_conditioning=true if the claim would revert to
"hollow" without those added conditions.

Return ONLY JSON:
{"critique":[{"axis":string,"finding":string,"severity":"fatal"|"weakens"|"clears"}],"verdict":"substantive"|"partial"|"hollow","surviving_claim":string|null,"reason":string,"needs_another_round":boolean,"added_conditions":integer,"survives_only_by_conditioning":boolean}
```

### User prompt format

Call site: lines 436–438 — `c.callJSON(substanceCriticSys, u, false, &last)`
where `u` is assembled at line 436:

```
CLAIM:
<current claim text>

PRODUCER STEELMAN:
<p.Steelman>

PRODUCER CONDITIONS:
<p.Conditions>
```

### Required JSON return schema

```json
{
  "critique": [
    {"axis": string, "finding": string, "severity": "fatal"|"weakens"|"clears"}
  ],
  "verdict": "substantive"|"partial"|"hollow",
  "surviving_claim": string|null,
  "reason": string,
  "needs_another_round": boolean,
  "added_conditions": integer,
  "survives_only_by_conditioning": boolean
}
```

Go target type: `substanceJSON` (lines 1008–1021)

---

## 4. `faithDefenderSys`

**Source lines:** 611–615  
**Role:** Finds the strongest verbatim evidence in the source transcript that the speaker actually
asserts the summary claim. Quotes only; no outside knowledge.  
**Stage:** First call in `faithClaim`.

### System prompt (verbatim)

```
You are the Defender. You are given a SUMMARY CLAIM and a SOURCE
transcript. Find the STRONGEST evidence in the SOURCE that the speaker actually asserts this claim.
Quote spans VERBATIM from the SOURCE only — never paraphrase, never use outside knowledge. If there
is no support, set found=false and quotes=[]. Return ONLY JSON:
{"found": boolean, "quotes": [string], "best_case": string}. No markdown.
```

### User prompt format

Call site: line 471 — `c.callJSON(faithDefenderSys, "SUMMARY CLAIM:\n"+claim+"\n\nSOURCE:\n"+src, false, &d)`

```
SUMMARY CLAIM:
<claim text>

SOURCE:
<source transcript text>
```

### Required JSON return schema

The prompt requests `best_case` but the Go struct does not read it:

```json
{"found": boolean, "quotes": [string], "best_case": string}
```

Go target type: `defenderJSON` (lines 1026–1029) — reads only `found` and `quotes`:
```go
type defenderJSON struct {
    Found  bool     `json:"found"`
    Quotes []string `json:"quotes"`
}
```

**Note:** `best_case` is in the prompt's schema but absent from the Go struct. The field is
generated by the model but silently discarded by the unmarshaller.

---

## 5. `faithCriticSys`

**Source lines:** 617–647  
**Role:** Given the defender's quoted evidence, decides whether the summary claim faithfully
represents the source. Identifies distortion mode and produces `what_source_actually_says` for
partial/overstated verdicts.  
**Stage:** Second call in `faithClaim`.

### System prompt (verbatim)

```
You are the Faithfulness Critic. Decide whether the SUMMARY CLAIM faithfully
 represents what the speaker said in the SOURCE, using the Defender's cited quotes. You are checking
 sense-preservation, not truth: your only question is whether the summary reports the speaker
 accurately, never whether the speaker was right. Watch for the ways a summary distorts a source:
- Fabrication: the claim is simply not in the source.
- Overstatement: the source hedged or qualified it; the summary made it absolute.
- Distortion: the meaning was changed.
- Context-stripping: a conditional or hypothetical presented as an unconditional belief.
- Misattribution: the speaker was quoting or steelmanning someone else, and the summary attributes
  it as the speaker's own view.
- Cherry-pick: present but unrepresentative of the source's stance.
- Literalization: the summary states as a sincere literal assertion something the speaker meant as
  provocation, hyperbole, or irony. The words may appear in the source but the asserted proposition
  does not — the speaker's force or register was rhetorical, not declarative. When this occurs the
  verdict is NOT "faithful"; use "overstated", and ALWAYS populate what_source_actually_says with
  the proposition the speaker actually asserted (i.e. what they meant, not what they said literally).

Verify the Defender's quotes actually appear in the source; do not take the Defender's word for it.

Verdict:
- "faithful": the summary reports the claim as the speaker stated it, including the register and
  force with which they stated it.
- "partial": the source supports a weaker/narrower version.
- "overstated": same in kind but the summary strengthened it — or literalized a rhetorical/ironic
  claim (see Literalization above).
- "absent": not in the source.
- "contradicted": the source says the opposite.

For "partial"/"overstated", give what_source_actually_says (the faithful version, including the
correct register). Return ONLY JSON:
{"findings":[{"mode":string,"finding":string}],"verdict":"faithful"|"partial"|"overstated"|"absent"|"contradicted","evidence":string,"what_source_actually_says":string|null}
```

**Note:** The leading space before "represents" on the second line is verbatim from the Go source
(line 618). It is not a formatting artifact of this document.

### User prompt format

Call site: lines 478–479 — `fmt.Sprintf(...)` assembled as:

```
SUMMARY CLAIM:
<claim text>

DEFENDER FOUND SUPPORT: <true|false>
DEFENDER QUOTES:
- <quote 1>
- <quote 2>
...

SOURCE:
<source transcript text>
```

When the defender found no quotes, `DEFENDER QUOTES:` is followed by `(none)` (line 474).

### Required JSON return schema

```json
{
  "findings": [{"mode": string, "finding": string}],
  "verdict": "faithful"|"partial"|"overstated"|"absent"|"contradicted",
  "evidence": string,
  "what_source_actually_says": string|null
}
```

Go target type: `faithJSON` (lines 1030–1038). The JSON key for `SourceSays` is
`"what_source_actually_says"` (line 1037).

---

## 6. `evidenceSys`

**Source lines:** 649–659  
**Role:** Web-search grounding. Judges whether the claim is supported by real, retrievable evidence.
The only prompt that uses `withTools=true` (web_search_20250305).  
**Stage:** Single call in `evidenceClaim`. The verdict may be subsequently downgraded by
`crossCheckEvidence`.

### System prompt (verbatim)

```
You are the Evidence Grounder. Decide whether the CLAIM is TRUE, using web
search to find real, current evidence — the actual truth-makers, not anyone's assertion that it is
true. Search for data, primary sources, and credible reporting; weigh what you find. Then judge:
- "supported": credible evidence backs the claim.
- "mixed": evidence cuts both ways, or supports only a qualified version.
- "refuted": credible evidence contradicts the claim.
- "unverifiable": a prediction, opinion, or otherwise not checkable against current evidence.

Keep finding to one sentence. List the sources you actually used, with real URLs from your search
results. Return ONLY JSON after searching:
{"verdict":"supported"|"mixed"|"refuted"|"unverifiable","finding":string,"sources":[{"title":string,"url":string}]}
```

### User prompt format

Call site: line 494 — `c.callJSONSourced(evidenceSys, "CLAIM:\n"+claim, true, &e)`

```
CLAIM:
<proposition text (may differ from the original claim — see proposition substitution rule)>
```

### Required JSON return schema

```json
{
  "verdict": "supported"|"mixed"|"refuted"|"unverifiable",
  "finding": string,
  "sources": [{"title": string, "url": string}]
}
```

Go target type: `evidenceJSON` (lines 1039–1046)

---

## Verification

Post-write diff between each extracted prompt string above and the corresponding Go `const`
declaration. Run with:

```sh
# Example for decomposeSys — repeat for each prompt
grep -A5 'const decomposeSys' assay.go
```

Verification was performed manually against the source at the cited line numbers. All six prompts
match the Go source verbatim. The only noted discrepancy is structural, not textual: `best_case` in
`faithDefenderSys` is declared in the prompt schema but absent from the Go struct `defenderJSON`
(lines 1026–1029) — the field is generated but never read.
