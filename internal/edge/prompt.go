package edge

// prompt.go holds the edge call's schema and prompt text (spec/EDGE.md §2), kept apart from the
// admission and rollup logic in edge.go. The schema is built PER SCHEME because each scheme's
// critical_question enum is that scheme's fixed CQ list (schemeCQs) — a defeater must answer one of the
// scheme's own questions, so the closed set the model may return is scheme-specific. The prompt reads
// the finding, the source quotes the report-tree pass already verified, the recommendation, and the
// scheme's CQs verbatim, and asks for a warrant plus at most one concrete defeater; it can never assert
// the recommendation follows.

import (
	"encoding/json"
	"fmt"
	"strings"
)

// SchemaName is the forced-tool name Anthropic uses for the edge call (backend.Request.SchemaName),
// the edge-pass counterpart of faithJudge's "faith_verdict".
const SchemaName = "edge_defeater"

// Schema returns the JSON schema constraining one edge sample for `scheme`, or ("", false) for an
// unknown scheme. The shape follows judgeSchema's proven strict-tool discipline (every field required,
// additionalProperties false): `defeater` is always an object and `none_admitted` is the authoritative
// "no defeater" signal, rather than a nullable object a strict tool may reject — the spec's `object |
// null` semantics preserved without the union type. `critical_question` and every `questions_considered`
// item are constrained to this scheme's CQ slugs, so a defeater cannot answer a question the scheme
// does not pose. There is no field for asserting the inference is sound (spec/EDGE.md §2, the axis
// boundary).
func Schema(scheme string) (json.RawMessage, bool) {
	cqs, ok := schemeCQs[scheme]
	if !ok {
		return nil, false
	}
	slugs := make([]string, len(cqs))
	for i, q := range cqs {
		slugs[i] = q.Slug
	}
	enum := jsonStringEnum(slugs)
	s := fmt.Sprintf(`{
  "type":"object",
  "properties":{
    "warrant":{"type":"string"},
    "none_admitted":{"type":"boolean"},
    "defeater":{
      "type":"object",
      "properties":{
        "world":{"type":"string"},
        "kind":{"type":"string","enum":["population","condition","definition"]},
        "anchor":{"type":"string"},
        "settles":{"type":"string"},
        "critical_question":{"type":"string","enum":%s}
      },
      "required":["world","kind","anchor","settles","critical_question"],
      "additionalProperties":false
    },
    "questions_considered":{"type":"array","items":{"type":"string","enum":%s}}
  },
  "required":["warrant","none_admitted","defeater","questions_considered"],
  "additionalProperties":false
}`, enum, enum)
	return json.RawMessage(s), true
}

// jsonStringEnum renders a []string as a JSON array literal for embedding in a schema.
func jsonStringEnum(vals []string) string {
	b, _ := json.Marshal(vals)
	return string(b)
}

// System is the fixed edge-attacker system prompt. It states the one move (reconstruct the warrant,
// then attack it with the scheme's questions), the entry condition for a defeater (a concrete named
// world whose anchor is copied from the report — finding, recommendation or source quotes — and what
// would settle it, the name-not-count rule as the admission gate), and the axis boundary the schema
// has no room to break: the model may refute an edge or decline, never certify that the recommendation
// follows.
const System = `You are the Edge Attacker. You are given a FINDING that a report established, the
verified SOURCE QUOTES the report's own faithfulness check already grounded for it, and a
RECOMMENDATION the report rests on that finding. The finding STANDS — do not re-litigate whether it is
faithful to its source; that is settled. Your one question is whether the RECOMMENDATION actually
FOLLOWS from the finding.

First reconstruct the WARRANT: the implicit premise W such that "for the recommendation to follow from
the finding, you need W." State W in one sentence.

Then attack W using ONLY the scheme's CRITICAL QUESTIONS, listed in the user message. Ask the scheme's
questions; do not invent your own. You are looking for a DEFEATER: a concrete state of the world,
plausible in the report's domain, in which the finding STILL HOLDS yet the recommendation FAILS.

A defeater is admitted only if it is CONCRETE. It must:
- name a specific referent — a population, a condition, or a definition (the "kind");
- set "anchor" to that referent copied VERBATIM from the FINDING, the RECOMMENDATION, or the SOURCE
  QUOTES — the concrete thing already in the report that the world turns on, NOT a phrase you coin in
  the "world" you write. A world built on a referent the report never names is not admitted.
- say what source or observation would SETTLE it ("settles").
"It might not generalise", "could be equivocating", "the sample may be unrepresentative" are NOT
defeaters — they name nothing and settle nothing. The anchor must be a referent the report itself
names; if the report raises no such referent, no concrete world opens and you decline.

DOMAIN RULE. The world must hold FOR THE AUDIENCE THE REPORT ADDRESSES. The report's genre
stipulates its audience and the goals that audience holds. A world that gives the organisation
DIFFERENT goals or constraints — a competing goal the report never puts on its reader, a binding
obligation outside what it addresses — is not a defeater of the inference: it varies the addressee
rather than attacking the step from finding to recommendation. Do not offer one. This is why the
scheme's questions do not ask whether the goal is held or conflicts with another goal; ask only
whether a cheaper means exists or the action carries a self-defeating side effect.

If, after asking the scheme's questions, no concrete world opens, set none_admitted=true, leave the
defeater fields "" (kind "condition" as a placeholder), and list the questions you considered in
questions_considered. That is an honest report of failure to refute.

You may NEVER assert that the recommendation follows. There is no verdict for "the inference is sound":
reasoning can kill an edge with a counter-world but can never confirm one — positive grounding of "R
follows from F" needs a truth-maker (an intervention estimate, a typicality count), which is a
retrieval, not a deduction. Refute the edge, or decline. Those are the only two moves.`

// User assembles the edge call's user message: the scheme, the finding, the verified quotes, the
// recommendation, and the scheme's critical questions verbatim (spec/EDGE.md §2). `quotes` are the
// already-verified source spans for the finding's leaves; an edge with no verified quote still runs —
// the finding stands on its faithfulness verdict — with a marker in their place.
func User(scheme, finding string, quotes []string, rec string) string {
	q := "(none — the finding stands on its faithfulness verdict)"
	if len(quotes) > 0 {
		q = "- " + strings.Join(quotes, "\n- ")
	}
	var cqLines strings.Builder
	for _, c := range schemeCQs[scheme] {
		fmt.Fprintf(&cqLines, "- %s (%s)\n", c.Question, c.Slug)
	}
	return fmt.Sprintf(`SCHEME: %s

FINDING (stands):
%s

VERIFIED SOURCE QUOTES:
%s

RECOMMENDATION (does it follow?):
%s

CRITICAL QUESTIONS for this scheme — ask these, and name which one a defeater answers in
critical_question:
%s`, scheme, finding, q, rec, strings.TrimRight(cqLines.String(), "\n"))
}
