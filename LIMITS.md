# LIMITS.md

## Operating envelope

| Region                                                 | Expected reliability | Basis | Tag |
|------------------------------------------------------------|------------------|---|---|
| Faithfulness with source present                           | Highest          | Truth-maker in the context window; Defender and Critic both quote source directly. | design-expectation |
| Refutation of internally contradictory claims              | High             | Reachable by reasoning alone; deduction, not retrieval. | design-expectation |
| `unverifiable` on unresolved predictions / armchair claims | High — by design | No direct unresolved-prediction probe run; reflexive canary `unverifiable` 5/5 is the only adjacent datum. | design-expectation (canary: haiku, 2026-06-11) |
| Resolved predictions grounded normally (not deflected to `unverifiable`) | Grounded on evidence | Ballmer iPhone forecast: `refuted` 10/10. The boundary tracks resolution — settled questions are assessed, not hidden behind `unverifiable`. | measured — haiku, 2026-06-10 |
| Relational defects (equivocation, motte-and-bailey)        | Partly dissolved by decomposition before grading | `decompose` splits motte from bailey into separate fragments; the equivocation move is graded piece-by-piece. Splitting grades the pieces — defective fragments were caught in all split runs — but the move between them is no longer graded as a move. The one false pass came from the single kept-whole run plus a silent critic; that leak belongs to the empty-critique signature. | measured — haiku, 2026-06-11 |
| Substance on sweeping forward predictions                  | Skews `hollow` (false-attack direction) | Bare Ballmer forecast `hollow` 10/10; central claim `hollow` 10/10 with full context. Fatal axis: Counterexample. Context does not relieve the verdict. | measured — haiku, 2026-06-11 |
| Substance on novel / specialist domains                    | Lowest — treat with most suspicion | Producer and Critic share the same model (W1); a blind spot in one role survives the other. | design-expectation |
| Positive grounding confirmation (`supported`)              | Bounded by retrieval quality; never armchair | `crossCheckEvidence` proves a URL was retrieved, not that the page backs the sentence (W2/W3). Structurally weakest verdict. | design-expectation |
| Tool's own headline claim | Untestable by the tool         | Reflexive canary `unverifiable` 5/5 — the only reachable check. | measured — haiku, 2026-06-11 |

---

## Known biases and bugs

**Forward predictions skew `hollow` via Counterexample.** Substance rates sweeping forward
predictions `hollow` because the Counterexample axis can always construct a counterexample against a
probabilistic forecast. The Ballmer iPhone prediction rated `hollow` 10/10 bare and `hollow` 10/10
with its full in-text argument. Evidence weakens but does not kill; Counterexample kills. The critic
does not distinguish a counterexample that defeats a universal from one that merely contests a
prediction. The planned `substanceCriticSys` fix was withdrawn after the probe falsified its
premise. Evidence: v1 examples/destructive/laundering/ (via spec/FINDINGS.md).

**Decomposition dissolves relational defects.** `decompose` splits a motte-and-bailey argument into
separate fragments and grades motte and bailey independently. Splitting grades the pieces —
defective fragments were caught in all split runs — but the move between them is no longer graded as
a move. The one false pass (motte run 10, 2026-06-11) came from the single kept-whole run plus a
silent critic; that leak belongs to the empty-critique signature.
Evidence: v1 testing/calibration_log.jsonl (via spec/FINDINGS.md), motte run 10.

**Referent collision on bare fragments.** Running substance on the repo's own claims showed the
critic reading "assay" as a chemical assay in 4 fragments, and rating "two of assay's three columns"
`hollow` because "columns" has no referent without surrounding context. Domain-polysemous terms
graded without a domain anchor collide their referents and skew toward `hollow`.
Evidence: v1 examples/reflexive/ (via spec/FINDINGS.md), 2026-06-11.

**Empty-critique false-pass signature.** Two of the three fragment-level false passes share one
signature: a defect fragment rated `substantive` with an empty critique — no axis fired (motte 1/63
and hidden-premise 1/36, 2026-06-11). The third leak, axis-gaps survivorship (1/70, 2026-06-11), had
a weak Hidden-premise finding and is classified as a mapped limit, not this signature. Any
`substantive` verdict where the critique list is empty or all severities are `"clears"` should be
treated with suspicion; no detection rule is implemented yet.
Evidence: v1 testing/calibration_log.jsonl (via spec/FINDINGS.md), motte run 10 and hidden-premise
run 8.

**Model-retyped URLs not matched to citation spans (W3).** `crossCheckEvidence` matches model-cited
URLs against retrieved URLs by host+path. This proves a URL was fetched; it does not prove the
specific sentence is backed by a specific span on that page. The API's citation block content is
never read, so this is URL-presence matching only, not span-level backing. Evidence:
assay.go:792–808 (via spec/FINDINGS.md).

**Non-enum verdict passes through unchecked (W10).** `unmarshalLoose` accepts any string into the
`Verdict` field; no caller validates enum membership. In hidden-premise run 3 (2026-06-11), the
model returned `"substantial"` — a non-enum near-miss — which passed through and rendered as-is.
Evidence: v1 testing/calibration_log.jsonl (via spec/FINDINGS.md), line 14.

---

## Permanent limits

The tool's headline claim — that seeing the three columns disagree causes a trained user to
recognise their skill as transferable — is a claim about effects on people. It cannot be settled by
reasoning, by web search, or by running crossexam on itself. The reflexive canary
(`unverifiable` 5/5, haiku, 2026-06-11) confirms only that grounding is not currently self-sealing.
The claim is a hypothesis awaiting field evidence.
