# Corpus manifest — documents held under sources/

Every source document the corpus holds, in the report's own naming. The backticked id is
canonical: a claim's `src`/`cites` column names the documents a claim rests on with these ids, and
a faithfulness run marks a claim `unverifiable` (no model call) when a cited id is not listed here.

This corpus holds exactly one document: the report itself. It is both the assay TARGET and the only
held SOURCE, which makes the only automated check an INTERNAL-CONSISTENCY one — a claim's text is
verified against report.pdf at its `ref` page (does the report actually say this; do restated
figures agree across pages). It does NOT ground the survey figures: their truth-maker is the study
microdata, which is not held (see "Not held" below), so every `route=survey` claim is `unverifiable`
for grounding by design, not by omission.

## Report (1)

- `report.pdf` — Quocirca, *Print Security Landscape 2026 — How AI, Identity, and Quantum are reshaping the threat landscape*, RICOH excerpt, July 2026 (15 pp.)

## Not held — named so their absence is `unverifiable`, never silently `absent`

These are the truth-makers the report's claims actually rest on. None is in this corpus; a grounding
run against them cannot be performed, and any conclusion depending on one is inconclusive, not clean.

- **Survey microdata / methodology** — the Quocirca 2026 study underlying every `route=survey`
  figure (67% breach rate, £1m mean cost, the leader/laggard split, and the ~40 other percentages).
  Sample size, sampling frame, question wording, weighting, and the year-on-year panel (2024 / 2025
  baselines the report cites) are NOT stated in the excerpt. Without them a survey figure can be
  checked only for internal consistency, never for truth.
- **Print Security Maturity Index definition** — the rule that classifies an organisation as
  leader / follower / laggard ("implemented at least six areas from a list of security solutions",
  report.pdf p2). The list of areas and the follower/laggard cut-offs are not held, so the divide
  claims (56% vs 72% vs 73%) cannot be reconstructed.
- **Vendor briefings / scorecards** — the vendor-supplied material behind the Vendor Landscape
  (Figure 1) scores and the Ricoh profile's capability claims (@Remote, Streamline NX zero trust,
  TLS 1.3 / TPM 2.0, FIDO2, the natif.ai and MTI acquisitions, "200 countries"). Vendor-attested,
  not independently held.
- **Prior editions** — the 2024 and 2025 *Print Security Landscape* reports the excerpt compares
  against (e.g. "up from 34% in 2024", "56% in 2025 to 67% today", zero trust "37% previously").
  The cited baselines cannot be verified without them.
- **Full report** — this is the RICOH-licensed excerpt (15 pp.); the complete report, its full
  vendor comparison, and any un-excerpted figures are not held.

## Report-link resolution

These lines steer the `-review` site's report deep-links (spec/SERVE.md), and are NOT held documents
— none begins with a backticked id, so LoadHeld and LoadLabels ignore them. `report_page_offset` is
the printed→PDF page offset: 0 here, because this report's printed footer page equals its page in
report.pdf (printed p2 is PDF page 2). `sections:` maps each NAMED § heading to the printed page it
starts on, so a claim citing "§Executive summary" — which the code cannot place from the name alone —
resolves to a page. A numbered §-ref ("§2.1.1 p7", the LCEIC style) carries its own page and needs no
entry here. A named section spanning two pages is listed at the page it starts on. `single_source:
true` declares that this corpus holds the report as its own only source, so a leaf `absent` is not a
grounding miss against some other held document — there is none — but a claim the report states once
and no second document repeats; the tree renders it `uncorroborated` and derives it as weakened, not
failed.

report_page_offset: 0

single_source: true

sections:
  Executive summary: 2
  Print remains a growing security vulnerability: 2
  Mixed-fleet environments continue to present a risk vector: 2
  Identity is becoming central to print security: 2
  Widespread AI integration presents both risk and reassurance: 3
  Key findings: 4
