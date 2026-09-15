# Corpus manifest — documents held under sources/

Every source document the corpus holds, in the report's own naming. The backticked id is
canonical: a claim's `src`/`cites` column names the documents a claim rests on with these ids, and
a faithfulness run marks a claim `unverifiable` (no model call) when a cited id is not listed here.

This corpus holds exactly one document: the report itself. It is both the assay TARGET and the only
held SOURCE, which makes the only automated check an INTERNAL-CONSISTENCY one — a claim's text is
verified against report.pdf at its `ref` page (does the report actually say this; do restated figures
agree with the appendix tables they come from). It does NOT ground the survey figures: their
truth-makers are the three surveys' microdata and The Tech's own writeup, which are not held (see "Not
held" below), so every `route=survey` claim is `unverifiable` for grounding by design, not by
omission. The design intent is precisely internal contradiction: the §3 recommendation body is checked
against the report's own Appendix C survey tables — identical in shape to dora-2026, not external
grounding.

## Report (1)

- `report.pdf` — *Report of MIT's Ad Hoc Committee on AI Use in Teaching, Learning, and Research
  Training*, MIT, 13 August 2026, co-chairs Klopfer & Madden, 38 printed pages plus appendices. One
  file; its appendices (A committee process, B syllabus-policy menu, C survey results) are companions
  bundled inside it, not separate PDFs. The graphs and statistical summaries in Appendix C were
  Codex-generated (Appendix A disclosure), so some appendix figures live in charts that may not
  survive PDF→text extraction — the extraction risk noted under the pre-registration below.

## Not held — named so their absence is `unverifiable`, never silently `absent`

These are the truth-makers the report's `survey`- and `evidence`-route claims actually rest on. None
is in this corpus; a grounding run against them cannot be performed, and any conclusion depending on
one is inconclusive, not clean.

- **The three surveys' microdata** — the response-level data behind every `route=survey` figure:
  the Spring 2026 AI Usage and Attitudes Survey (n=1,632), the Spring 2026 Quality of Life Survey
  (n≈8,200), and the Fall 2025 Tech Survey (n=1,002). The report summarizes these in Appendix C; the
  underlying data and the fitted breakdowns are not held. Without them a survey figure can be checked
  only for internal consistency against the appendix text, never for truth.
- **The Tech's own published writeup** — the Fall 2025 student-newspaper article at
  thetech.com/2025/11/25/llm-survey-results that Appendix C.3 summarizes. Not held; the C.3 figures
  are checkable only against the report's own summary of them.
- **Cited third-party sources** — the external references the report's `evidence`-route claims quote,
  each a separate publication whose quoted figure cannot be verified without it: the Brookings
  "pro-worker AI" piece (Acemoglu/Autor/Johnson, §2.7), the PsyArXiv "cognitive surrender" preprint
  (§2.7), the Pew "Americans and AI 2026" report (§3.3.9), the Wall Street Journal interview with
  Rafella Sadun (§3.1.3, paywalled), and the Healthy Minds Network 2024 national report (§3.2).
- **MIT internal pages** — the IS&T/Parley resources and pricing behind §3.3.7 (Parley $30/month
  credits, $200/month commercial tiers), and other mit.edu URLs behind SSO. Not held; institutional
  figures are taken as stated, not grounded.

## Report-link resolution

These lines steer the `-review` site's report deep-links (spec/SERVE.md), and are NOT held documents —
none begins with a backticked id, so LoadHeld and LoadLabels ignore them. `report_page_offset` is the
printed→PDF page offset: **2** here, because this report's printed footer page is two behind its page
in report.pdf (printed p3, §1.1 The Landscape, is PDF page 5 — a two-page front matter of cover plus
contents precedes printed p1). A claim's `ref` numbered label (`3.1.8`, `2.4`) is self-locating: its
printed page is the `pNN` in the ref, offset to the PDF page here. A NAMED appendix label (`C.2 MIT
Guidance`) cannot be placed from its name, so `sections:` maps each named label used in the claims
file to the printed page it sits on. `single_source: true` declares that this corpus holds the report
as its own only source, so a leaf `absent` is not a grounding miss against some other held document —
there is none — but a claim the report states once and no second passage repeats; the tree renders it
`uncorroborated` and derives it as weakened, not failed.

report_page_offset: 2

single_source: true

sections:
  A. Committee Process: 29
  B. Sample AI Policy: 29
  C.1 Spring 2026 AI Usage and Attitudes Survey: 31
  C.1 Overall Usage: 31
  C.1 Student Usage: 32
  C.1 AI Attitudes: 32
  C.1 Research Use: 33
  C.1 AI Concerns: 33
  C.2 Spring 2026 Quality of Life Survey: 34
  C.2 Frequency of AI Tool Use: 34
  C.2 Pessimistic/Optimistic: 35
  C.2 Replaceable/Capable: 35
  C.2 Worsens/Improves Work: 36
  C.2 Inefficient/Efficient: 36
  C.2 Unreliable/Reliable: 37
  C.2 MIT Guidance: 37
  C.3 Fall 2025 Tech Survey: 38
