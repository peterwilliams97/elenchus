# spec/EVIDENCE.md — the evidence-grounding cross-check

`-evidence` grounds each claim against web-retrieved truth-makers (`evidenceSys`), then
`crossCheckEvidence` (assay.go) forces `unverifiable` from code — no model call — whenever the
verdict rests on nothing a reader can re-check. It never *confirms* a verdict, only downgrades one,
per `CLAUDE.md` § The axis boundary.

Two gates, in order. A downgrade records the model's own verdict in `original_verdict` for audit and
keeps the cited `sources`; only `error` is exempt.

- **Horizon.** `horizon` ∈ {past, present, future} is a required schema field: the time by which the
  claim's truth is settled, relative to today. `horizon=future` ⇒ `unverifiable`, reason prefixed
  `forecast — projections are not evidence`. A projection of a future outcome is not an observation
  of it, however many forecasters agree; reasoning cannot ground an unobserved outcome.
- **Retrieval match.** A non-`unverifiable` verdict whose cited URLs are absent from the round-trip's
  retrieved set ⇒ `unverifiable` — the verdict was model self-report, not a read source.
