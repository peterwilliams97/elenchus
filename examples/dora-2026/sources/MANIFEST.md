# Corpus manifest — documents held under sources/

Every source document the corpus holds, in the report's own naming. The backticked id is
canonical: a claim's `src`/`cites` column names the documents a claim rests on with these ids, and
a faithfulness run marks a claim `unverifiable` (no model call) when a cited id is not listed here.

This corpus holds exactly one document: the report itself. It is both the assay TARGET and the only
held SOURCE, which makes the only automated check an INTERNAL-CONSISTENCY one — a claim's text is
verified against report.pdf at its `ref` page (does the report actually say this; do restated figures
agree across pages and chapters). It does NOT ground the survey figures: their truth-maker is the study
microdata and the fitted models, which are not held (see "Not held" below), so every `route=survey`
claim is `unverifiable` for grounding by design, not by omission.

## Report (1)

- `report.pdf` — Google Cloud / DORA, *State of AI-assisted Software Development* (the 2025 DORA
  Report), v.2025.2, 142 pp. Licensed under CC BY-NC-SA 4.0. Drawing on a global survey conducted
  June 13–July 21, 2025 (4,867 respondents) plus 78 in-depth interviews (July 2024–July 2025).

## Not held — named so their absence is `unverifiable`, never silently `absent`

These are the truth-makers the report's claims actually rest on. None is in this corpus; a grounding
run against them cannot be performed, and any conclusion depending on one is inconclusive, not clean.

- **Survey microdata + fitted models** — the 4,867-response dataset and the Bayesian regression and
  structural-equation / DAG models behind every `route=survey` figure (the 90% adoption rate, the
  cluster shares, the ~forty percentages, AND the "estimated effect" / "with a high degree of
  certainty ... amplified" moderation findings that constitute the DORA AI Capabilities Model). The
  report's Methodology chapter walks a *simplified toy example* (report.pdf p115–129), not the real
  fits; question wording is in the Appendix (p138–140) and at dora.dev/research/2025/questions (not
  held). Without the microdata a survey figure can be checked only for internal consistency, never for
  truth. Note the report's own framing: these are average comparisons, not causal effects (p36 fn20).
- **Interview corpus** — the 78 semi-structured interview transcripts underlying the qualitative
  claims and the pull-quotes. Recorded and transcribed by DORA; not held.
- **Cited third-party reports** — the external figures the report quotes: the 2025 Stack Overflow
  Developer Survey (84% / 76%), Atlassian's 2025 State of DevEx (99%), a 2025 LinkedIn report (88%),
  a McKinsey survey (78%), Stanford HAI's 2025 AI Index ($252.3B, 26%, 323%), and the METR study
  (19% / 20%). Each is a separate publication; the quoted figure cannot be verified without it.
- **Prior edition (2024 DORA Report)** — the source of the year-on-year baselines: the "DORA 2024
  anomaly", the 14.1%-increase 2024 adoption baseline, and the 1.5% throughput / 7.2% instability per
  25% AI adoption estimate. The cited baselines cannot be verified without it.
- **Vendor and named-case-study material** — the pilot results attributed to Adidas (20–30%, 50%
  "Happy Time"), Booking.com (up to 30% merge requests), Sabre (74% / 86% / 25%), and Wayfair.
  Contributor-attested, not independently held.

## Report-link resolution

These lines steer the `-review` site's report deep-links (spec/SERVE.md), and are NOT held documents
— none begins with a backticked id, so LoadHeld and LoadLabels ignore them. `report_page_offset` is
the printed→PDF page offset: 0 here, because this report's printed footer page equals its page in
report.pdf (printed p3 is PDF page 3). `sections:` maps each NAMED § heading to the printed page it
starts on, so a claim citing "§Executive summary" — which the code cannot place from the name alone —
resolves to a page. A named section spanning two pages is listed at the page it starts on.
`single_source: true` declares that this corpus holds the report as its own only source, so a leaf
`absent` is not a grounding miss against some other held document — there is none — but a claim the
report states once and no second document repeats; the tree renders it `uncorroborated` and derives it
as weakened, not failed.

report_page_offset: 0

single_source: true

sections:
  Executive summary: 3
  Key findings: 4
  Analysis and advice for technology leaders: 5
  Foreword: 8
  Software delivery performance factors: 13
  Look beyond software delivery performance: 14
  Finding commonality: 15
  Cluster 1: Foundational challenges: 16
  Cluster 2: The legacy bottleneck: 17
  Cluster 3: Constrained by process: 17
  Cluster 4: High impact, low cadence: 17
  Cluster 5: Stable and methodical: 18
  Cluster 6: Pragmatic performers: 18
  Cluster 7: Harmonious high-achiever: 18
  Software delivery performance levels: 19
  How do you compare?: 20
  Adoption: 24
  Experience: 25
  Time: 25
  Reflexive use: 26
  Reliance: 26
  Tasks: 27
  Surfaces: 29
  Mode of AI use: 29
  Individual productivity: 30
  Code quality: 30
  Trust: 31
  Final thoughts: 32
  AI is the new normal in software development: 34
  What is the impact of AI adoption?: 35
  Measuring AI adoption: 36
  The results this year: 38
  Software delivery instability: 41
  Changes in last year's patterns suggest adaptation: 42
  Conclusion: 43
  Authentic pride: 45
  Meaningful work: 45
  Need for cognition: 45
  Existential connection: 46
  Psychological ownership: 46
  Skill reprioritization: 46
  DORA AI Capabilities Model: 49
  AI capabilities: 50
  Clear and communicated AI stance: 52
  Healthy data ecosystems: 54
  AI-accessible internal data: 55
  Strong version control practices: 56
  Working in small batches: 58
  User-centric focus: 60
  Quality internal platforms: 62
  Clarify and socialize your AI policies: 63
  Center users' needs in product strategy: 64
  Our key findings: 66
  The platform landscape: 67
  A force multiplier for performance, well-being, and risk: 70
  The strategic imperative: 71
  Value stream management: 73
  How this appears in our 2025 findings: 77
  Looking beyond the tools to drive AI impact: 80
  The AI mirror: 80
  Organizations are systems, not sums of individuals: 81
  AI as mirror and multiplier: 86
  Choosing measurement frameworks to fit your organizational goals: 90
  Applying measurement frameworks in the age of AI: 93
  Focus on the user: 95
  Putting research into practice: 96
  Demographics and firmographics: 103
  Age: 103
  Gender: 104
  Roles: 105
  Role experience: 106
  Team experience: 106
  Where they work: 106
  Coding language: 107
  Industry: 108
  Size: 108
  Application criticality: 109
  Service age: 109
  Service users: 110
  Country: 111
  Interview participant demographics: 112
  Interviews: 112
  Localizations: 114
  Collect survey responses: 114
  Survey flow: 114
