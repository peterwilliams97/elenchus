# Corpus manifest — documents held under sources/

Every source document the corpus holds, in the report's own naming. The backticked id is
canonical: a claim's `src` column names the document a claim rests on with these ids, and a
faithfulness run marks a claim `unverifiable` (no model call) when a cited id is not listed here.

This corpus holds two AI Index 2026 excerpts, both assay TARGETS and held SOURCES: the coding section
(`report.pdf`, printed pp100–102, SEC/SW/SB/TB/VC claims) and the labor-impact section
(`report-productivity.pdf`, printed pp219–221, EP claims). Alongside them: seven Wayback captures of
the three benchmark leaderboards the coding section cites (see "Leaderboard captures" below) and seven
academic study PDFs the labor section cites (see "Studies" below). Two automated checks are therefore
possible:

- INTERNAL CONSISTENCY, against `report.pdf` at a claim's `ref` page — does the section actually say
  this; do the prose figures agree with the figure they restate. This is the only check for
  apparatus and evaluative claims, and the truth-maker for the coding section's own text.
- GROUNDING a `route=benchmark` score against a held leaderboard capture. A capture grounds the
  leaderboard's state AT ITS OWN timestamp, which brackets — never exactly reproduces — the report's
  cited snapshot ("as of February 2026" / "early 2026"). So a capture can confirm a specific value
  where it is present (e.g. SWE-bench 76.80, Terminal-Bench 77.3%, Vibe Code Bench 57.57%) and
  document drift where the leaderboard has since moved; it cannot reconstruct a state that falls
  between two captures. `route=benchmark` stays the route; the truth-maker is now held, subject to
  that timestamp gap. No judge run has been performed (no `evidence/` chains).

## Report excerpts (2)

- `report.pdf` — Stanford HAI, *Artificial Intelligence Index Report 2026*, Chapter 2 Technical
  Performance, §2.5 Performance in Specific Domains, the "Software" (coding) subsection, printed
  pages 100–102 (the SWE-bench / Terminal-Bench / Vibe Code Bench benchmarks). Extracted from the
  425-page full report as a 3-page excerpt; `report.txt` is its pdftotext. The SEC/SW/SB/TB/VC claims.
- `report-productivity.pdf` — same report, Chapter 4 Economy, §AI's Labor Impact / §Productivity
  Trends, printed pages 219–221 (Figures 4.4.27 and 4.4.28, the micro- and macro-level productivity
  studies). A 3-page excerpt; `report-productivity.txt` is its pdftotext. The EP claims.

## Leaderboard captures (7) — Wayback snapshots of the three cited benchmark leaderboards

Held under `leaderboards/`. Each is a Blink-saved MHTML of an `web.archive.org` capture; the
canonical held id is the `.txt` (the main leaderboard table extracted from the MHTML, main table
only — the `.mhtml` beside it is the raw capture, as `report.pdf`/`report.txt` pair). A capture
grounds the leaderboard's state at its capture timestamp, so the "bears on" column names the claims
whose truth-maker it supplies AND says whether the capture actually contains the value the report
cites — a capture on the wrong side of a change confirms the leaderboard's existence and shape but
not the report's specific number.

Provenance caveats that travel with these captures:
- **SWE-bench** loads its table via JS; the MHTML froze whichever filter tab was rendered. The
  `tablinks`-active marker and the visible rows can disagree, so each `.txt` records the marker and
  says to read the actual filter off the Release/score columns, not the tab label.
- **Vals.ai** renders its ranking as a Pareto-curve chart (SVG), not an HTML table, so the extracted
  text is the page's own **Takeaways + Results** prose (which names the top models and their scores),
  not a full per-model table. The full ranked bar values are not in the static DOM.
- **`mini-SWE-agent-v2`** — the exact filter the report says Figure 2.5.1 used (SB9) — appears in no
  capture as literal text; swebench.com exposes a `mini-SWE-agent` filter without the `-v2` suffix.

Each capture is registered as its own held id (the extracted `.txt`) so LoadHeld picks it up; the URL,
capture timestamp, and the claims it bears on follow.

- `leaderboards/swebench-20260205.txt` — SWE-bench Leaderboard, `https://www.swebench.com/index.html`
  (p101 fn21), Wayback capture 2026-02-05 19:46:02 UTC. Bears on SB1 (definition). Pre-2.0 state:
  harness 1.x, top Claude 4.5 Opus medium 74.40; does NOT contain the report's 76.80 (SWE-bench 2.0
  landed ~Feb 17), so it brackets SB5/SB6/SB7 from below rather than confirming them.
- `leaderboards/swebench-20260226.txt` — SWE-bench Leaderboard, same URL, Wayback capture 2026-02-26
  12:34:30 UTC. Bears on SB5, SB6, SB7. SWE-bench 2.0 table: top row Claude 4.5 Opus (high reasoning)
  76.80 (2026-02-17, harness 2.0.0) = SB6 exactly; Gemini 3 Flash / MiniMax M2.5 75.80, Claude Opus
  4.6 75.60, then a 74–72 cluster = SB5 (low-to-mid 70s) and SB7 (others 70–76). The in-February
  capture that carries the report's cited values.
- `leaderboards/swebench-20260403.txt` — SWE-bench Leaderboard, same URL, Wayback capture 2026-04-03
  (Verified tab active). Bears on SB5, SB6, SB7 (drift). Later Verified table, top still Claude 4.5
  Opus (high reasoning) 76.80 — the cited leader held two months on.
- `leaderboards/terminal-bench-20260220.txt` — Terminal-Bench 2.0 Leaderboard,
  `https://www.tbench.ai/leaderboard/terminal-bench/2.0` (p102 fn22), Wayback capture 2026-02-20
  18:00:20 UTC. Bears on TB1, TB2 (definition). Top Simple Codex / GPT-5.3-Codex 75.1%; does NOT yet
  contain 77.3% (that entry is dated 2026-02-24), so it brackets TB4 from below.
- `leaderboards/terminal-bench-20260310.txt` — Terminal-Bench 2.0 Leaderboard, same URL, Wayback
  capture 2026-03-10 16:13:00 UTC. Bears on TB4. Contains the report's 77.3% (Droid / GPT-5.3-Codex,
  rank 2, dated 2026-02-24) — the peak in late Feb before Gemini 3.1 Pro (78.4%, Mar 2) overtook it.
  Grounds TB4's "77.3% in early 2026"; the "20% in February 2025" baseline is in no held capture (all
  captures are 2026).
- `leaderboards/vibe-code-20260216.txt` — Vals.ai Vibe Code Bench,
  `https://www.vals.ai/benchmarks/vibe-code` (p102 fn23), Wayback capture 2026-02-16 (updated
  2/5/2026). Bears on VC1, VC2 (definition). This is Vibe Code Bench v1.0, not the v1.1 the report
  cites (Figure 2.5.3): GPT 5.2 leads 41.31%, Claude Opus 4.6 absent. A prior benchmark version — does
  NOT bear on the v1.1 scores VC4–VC8.
- `leaderboards/vibe-code-v1.1-20260331.txt` — Vals.ai Vibe Code Bench, same URL, Wayback capture
  2026-03-31 16:20:10 UTC (updated 3/20/2026). Bears on VC4 (value). v1.1: confirms Claude Opus 4.6
  (Nonthinking) = 57.57% — the value Figure 2.5.3 plots and VC4's INTERNAL flag pins against the 56.5%
  prose. But here GPT 5.4 (67.42%) and GPT 5.3 Codex (61.77%) lead it, so the ranking claims (VC4
  "leads", VC5 GPT 5.2 ~47%) reflect an earlier v1.1 state — see the gap below.

### Vibe Code Bench capture gap — the report's v1.1 leader snapshot is not held

There is **no capture between 2026-02-17 and 2026-03-30**. The 2026-02-16 capture is still v1.0
(GPT 5.2 leading at 41.31%); by the 2026-03-31 capture v1.1 already had GPT 5.4 and GPT 5.3 Codex on
top. The report's Figure 2.5.3 shows an early-v1.1 state — Claude Opus 4.6 (Nonthinking) LEADING at
57.57% (prose 56.5%), GPT 5.2 second at ~47%, before GPT 5.3 Codex / GPT 5.4 existed on this board.
That state falls inside the gap: the `20260331` capture corroborates the 57.57% value but not the
"leads" ranking, and no held capture reproduces the report's snapshot. So VC4's "leads" and VC5 are
grounded only on the value, not the ordering (noted on F-VC in `argument.txt`).

## Studies (7) — academic papers the labor section cites, fetched to `sources/papers/`

Held truth-makers for the `study-held` EP claims: a study's own PDF settles whether the section's
figure for it (Figure 4.4.27 / 4.4.28) reports the study's result faithfully. The other cited studies
are gated or landing-page-only and stay `unverifiable` (see "Not held" below and PROVENANCE.md). The
backticked id is `paper:<file-stem>`; the label gives author-year and the EP claim it bears on.

- `paper:cui-2025` — Cui et al. 2025, GitHub Copilot field experiments (economics.mit.edu draft).
  Bears on EP8 (developers +26% pull requests).
- `paper:becker-2025-metr` — Becker et al. 2025 (METR), arXiv:2507.09089v2. Bears on EP12, EP13
  (experienced open-source developers 19% slower; slowdown not replicated in a later study).
- `paper:ju-aral-2025` — Ju & Aral 2025, arXiv:2503.18238v3. Bears on EP9 (marketing output +50%).
- `paper:shen-tamkin-2026` — Shen & Tamkin 2026, arXiv:2601.20245v2 (prose miscites "2025"; the
  reference list dates it 2026, EP14's INTERNAL flag). Bears on EP14 (learning penalties).
- `paper:frank-2026` — Frank et al. 2026, arXiv:2601.02554v1. Bears on EP24 (AI-exposed occupations'
  decline pre-dates ChatGPT).
- `paper:brynjolfsson-chandar-chen-2025-canaries` — Brynjolfsson, Chandar & Chen 2025, "Canaries in
  the Coal Mine" (digitaleconomy.stanford.edu). Bears on EP25 (entry-level employment −15–16%).
- `paper:hosseini-lichtinger-2025` — Hosseini & Lichtinger 2025, SSRN 5425555. Bears on EP26
  (seniority-biased employment change).

## Not held — named so their absence is `unverifiable`, never silently `absent`

- **Full AI Index 2026 report** — the other 422 pages, including the rest of Chapter 2 and the
  mathematics / finance / legal subsections that share §2.5 with coding, are not held; a claim that
  reaches beyond the excerpted pages cannot be checked here.
- **The report's exact benchmark snapshots** — for the reasons above, no held capture reproduces the
  precise leaderboard state the report read (SWE-bench Verified `mini-SWE-agent-v2` top-10 "as of
  February 2026"; Terminal-Bench's 20%→77.3% time series; the early-v1.1 Vibe Code Bench ordering).
  The captures bracket these; they do not recover them.

## Report-link resolution

These lines steer the `-review` site's report deep-links (spec/SERVE.md), and are NOT held documents
— none begins with a backticked id, so LoadHeld and LoadLabels ignore them. This corpus holds two
report excerpts, so the rules are a `reports:` list, one entry per excerpt. Each entry's `offset:` is
the printed→PDF page offset (the excerpt's page 1 is the report's printed page 100 / 219, so
offset −99 / −218), `prefixes:` names the claim-id prefixes whose leaves open that excerpt, and
`sections:` maps each NAMED heading to the printed page it starts on, so a claim citing "§Software" or
"§Productivity Trends" resolves to a page. A leaf routes to its excerpt by the alphabetic head of its
claim id (SB6 → SB → `report.pdf`; EP8 → EP → `report-productivity.pdf`).

`single_source: false`: this corpus no longer holds a single report excerpt, so a leaf `absent` is a
plain grounding miss (no held source carries the claim), not the "said once, not repeated" that the
single-report shape reads as `uncorroborated`. The leaderboard captures and the study PDFs are
grounding truth-makers for the `route=benchmark` and `study-held` claims; a score's or study-figure's
grounding is settled by its capture / paper (above). No judge run has been performed (no `evidence/`
chains); whether a benchmark or study grounding counts as corroboration is a judge-run decision, left
open here.

single_source: false

reports:
  - file: report.pdf
    offset: -99
    prefixes: SEC SW SB TB VC
    sections:
      2.5 Performance in Specific Domains: 100
      Software: 100
      SWE-bench: 100
      Terminal-Bench: 101
      Vibe Code Bench: 102
  - file: report-productivity.pdf
    offset: -218
    prefixes: EP
    sections:
      AI's Labor Impact: 219
      Productivity Trends: 219
