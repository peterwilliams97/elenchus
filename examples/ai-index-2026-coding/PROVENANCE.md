# Provenance — ai-index-2026-coding

Records the report under analysis, the exact sections extracted, and which of their citations are
held. The corpus now holds **two extracted sections** of the same report, with different grounding
profiles:

- **Coding section** (Ch.2 §2.5 Software, pp100–102) — internal-consistency assay against the report
  itself, now with the three benchmark leaderboards held as seven Wayback captures under
  `sources/leaderboards/` (they bracket the report's Feb-2026 / early-2026 snapshot). See "Leaderboard
  captures" below.
- **Productivity section** (Ch.4 §4.4 Productivity Trends, pp219–221) — added later; it cites named
  academic studies, and 7 of them were fetched as open PDFs, so several of its claims are genuinely
  groundable. See "Second section" below.

## Report under analysis (both TARGET and only held SOURCE)

- Stanford HAI, *Artificial Intelligence Index Report 2026*, published 2026-06-30 (PDF ModDate),
  425 pages. Chapter 2 Technical Performance, §2.5 "Performance in Specific Domains" → the
  **"Software"** (coding) subsection.
- **Section extracted:** printed pages **100–102** (= PDF pages 100–102 of the full report; the
  running footer number equals the PDF page in this range). Covers the Software intro and the three
  coding benchmarks: SWE-bench (Figure 2.5.1), Terminal-Bench (Figure 2.5.2), Vibe Code Bench
  (Figure 2.5.3). Page 100 also carries the §2.5 parent intro (which frames all four domains —
  coding, mathematics, finance, legal); page 103 begins Mathematics and is not included.
- Source PDF: `/Users/peterw/code/personal/elenchus_material/ai_index_report_2026.pdf`
  (37,885,501 bytes). `sources/report.pdf` is pages 100–102 extracted with
  `pdfseparate -f 100 -l 102` then `pdfunite` (poppler; no gs/qpdf on this machine). The extract is
  ~39 MB because pdfunite copies the parent's full resource dictionaries; it is a valid 3-page PDF
  with an intact text layer (`sources/report.txt` is its pdftotext). Held for analysis, not
  redistribution — gitignored via `.gitignore`; the derived `claims-machine.txt` / `argument.txt`
  and this `MANIFEST.md` are tracked.

## Citations in this section — 3 leaderboards held (7 captures), snapshot dates bracketed

Every citation in the Software subsection is a **live benchmark leaderboard** (not an arXiv paper, so
`sources/papers/` stays empty here). All three are now held as Wayback captures under
`sources/leaderboards/` — held for analysis, not redistribution, and gitignored like the rest of
`sources/`. A capture grounds the leaderboard's state at its own timestamp; because the report cites
a snapshot ("as of February 2026" / "early 2026") the captures bracket rather than reproduce, the
"bears on" note in `sources/MANIFEST.md` says whether each capture actually carries the report's
cited value.

| Leaderboard | Locator | Captures held | Carries the report's cited value? |
|---|---|---|---|
| SWE-bench Leaderboard | `https://www.swebench.com/index.html` (p101 fn21) | `swebench-{20260205,20260226,20260403}.txt` | Yes — `20260226` top row Claude 4.5 Opus (high reasoning) 76.80 = SB6; the 74–72 cluster = SB5/SB7. `20260205` is pre-2.0 (no 76.80). |
| Terminal-Bench 2.0 Leaderboard | `https://www.tbench.ai/leaderboard/terminal-bench/2.0` (p102 fn22) | `terminal-bench-{20260220,20260310}.txt` | Yes for 77.3% (`20260310`, dated 2026-02-24); the 20% Feb-2025 baseline is in no held capture. |
| Vals.ai — Vibe Code Bench | `https://www.vals.ai/benchmarks/vibe-code` (p102 fn23) | `vibe-code-20260216.txt` (v1.0), `vibe-code-v1.1-20260331.txt` (v1.1) | Partly — `20260331` confirms Claude Opus 4.6 (Nonthinking) 57.57% but not the "leads" ordering; the report's early-v1.1 snapshot falls in a capture gap (below). |

These captures are the truth-makers behind the `route=benchmark` claims. The route is unchanged; the
truth-maker is now held, so a future judge run can ground each score where its capture carries the
value and report `unverifiable` where the report's snapshot falls between captures (the Vibe Code
Bench v1.1 leader ordering). No judge run has been performed.

## Derived files

- `claims-machine.txt` — 30 atomic claims (9 benchmark, 11 evidence, 9 evaluative, 1 data-gap). Every
  number and every year-on-year figure is a claim; interpretive sentences are `route=evaluative`.
  Two internal-consistency flags recorded in `src=` notes: the Vibe Code Bench prose leader (56.5%)
  vs the highest bar plotted in Figure 2.5.3 (57.57%), and the prose model name "GPT 5.3 Codex
  (41.4%)" vs the figure legend's "GPT 5.2 Codex" with no 41.4% bar (VC4, VC6). A judge run over
  report.pdf is meant to adjudicate these.
- `argument.txt` — the section as an argument tree: root = the Software thesis (evaluation shifting
  to end-to-end delivery, §Software p100); three benchmark subsections (SB / TB / VC) as the
  recommendation tier; one finding each (F-SB / F-TB / F-VC); score-claim leaves. Apparatus
  (benchmark definitions) and evaluative asides are excluded and listed at the file's foot.
- No judge run (no `evidence/` chains, no `site/`).

---

# Second section — Economy §4.4 Productivity Trends (AI and developer productivity)

## Section extracted

- Stanford HAI, *Artificial Intelligence Index Report 2026*, Chapter 4 Economy, §4.4 Jobs →
  **"AI's Labor Impact"** → **"Productivity Trends"** subsection.
- **Section extracted:** printed pages **219–221** (= PDF pages 219–221 of the full report; the
  running footer number equals the PDF page in this range, verified). Covers the "AI's Labor Impact"
  intro, "Micro-level Studies" (prose + Figure 4.4.27, the per-study micro table) and "Macro-level
  Studies" (prose + Figure 4.4.28, the macro productivity/employment table).
- **Boundary:** page 221 also *begins* the next subsection, **"Workforce Impact"** (employment /
  junior-worker displacement). That thread is out of scope — it is about employment, not the
  productivity thesis — so its prose and its charts (Figures 4.4.29–4.4.31, pp222–226) are not
  decomposed. Three employment studies that appear inside Figure 4.4.28 on p221 (Frank et al.,
  Brynjolfsson "canaries", Hosseini/Lichtinger) are recorded as claims EP24–EP26 but are not attached
  to the argument tree (argument-productivity.txt "NOT in this tree").
- Source PDF: `/Users/peterw/code/personal/elenchus_material/ai_index_report_2026.pdf`.
  `sources/report-productivity.pdf` is pages 219–221 extracted with `pdfseparate -f 219 -l 221` then
  `pdfunite` (poppler). ~39 MB because pdfunite copies the parent's full resource dictionaries; it is
  a valid 3-page PDF with an intact text layer (`sources/report-productivity.txt` is its pdftotext).
  Held for analysis, not redistribution — gitignored via the repo `.gitignore` (`sources/*`); the
  derived `claims-machine.txt` (EP* claims) and `argument-productivity.txt` are tracked.

## Citations in this section — 7 held, 9 not held

Unlike the coding section (whose citations are all live leaderboards), this section cites named
academic studies by author-year. The fetch rule (arXiv / open PDF → `sources/papers/<id>.pdf`;
everything else → "not held" with a reason) resolves to **7 fetched**. Each held paper's first page
was checked to be the document its title claims. The 7 held papers make the corresponding
`route=evidence` claims groundable; the 9 not-held studies ground to `unverifiable`, as the benchmark
claims in the coding section do.

### Held (7) — fetched to `sources/papers/`

| Study (report's cite) | File | Source URL | Backs |
|---|---|---|---|
| Cui et al. 2025 | `cui-2025.pdf` | economics.mit.edu/…/draft_copilot_experiments.pdf | EP8 (Copilot +26% PRs) |
| Becker et al. 2025 (METR) | `becker-2025-metr.pdf` | arXiv:2507.09089v2 | EP12, EP13 (devs 19% slower) |
| Ju & Aral 2025 | `ju-aral-2025.pdf` | arXiv:2503.18238v3 | EP9 (marketing +50%) |
| Shen & Tamkin 2026 | `shen-tamkin-2026.pdf` | arXiv:2601.20245v2 | EP14 (learning penalties) |
| Frank et al. 2026 | `frank-2026.pdf` | arXiv:2601.02554v1 | EP24 (AI-exposed decline pre-ChatGPT) |
| Brynjolfsson, Chandar & Chen 2025 | `brynjolfsson-chandar-chen-2025-canaries.pdf` | digitaleconomy.stanford.edu/…/CanariesintheCoalMine_Nov25.pdf | EP25 (canaries; -15–16%) |
| Hosseini & Lichtinger 2025 | `hosseini-lichtinger-2025.pdf` | alejandrobarros.com/…/ssrn-5425555.pdf | EP26 (seniority-biased change) |

Frank et al.'s arXiv id is truncated in the report's own reference list ("https://arxiv.org/" + a
stray page number); it was resolved by title via the arXiv API to 2601.02554 and fetched.

### Not held (9) — recorded with reason

| Study (report's cite) | Locator | Type | Reason not held | Backs |
|---|---|---|---|---|
| Brynjolfsson, Li & Raymond 2025 "Generative AI at Work" | doi.org/10.1093/qje/qjae044 | Journal (QJE) | Gated journal DOI, no open PDF | EP7 (support +14–15%) |
| Reimers & Waldfogel 2026 | NBER WP 34777 (doi.org/10.3386/w34777) | NBER working paper | Gated | EP16 (authors +200%) |
| Ho Choi & Xie 2025 | gsb.stanford.edu/…/human-ai-accounting-early-evidence-field | GSB working-paper page | Landing page; no open PDF link found | EP17 (accountants +55%) |
| Aldasoro et al. 2026 | cepr.org/voxeu/columns/how-ai-affecting-productivity-and-jobs-europe | VoxEU column | HTML column, not the paper PDF | EP19 (Euro firms +4%) |
| Brynjolfsson 2026 "AI productivity take-off" | ft.com/content/4b51d0b4-… | News article | Paywalled | EP20, EP21 (US 2.7%, J-curve) |
| Filippucci et al. 2025 (OECD) | doi.org/10.1787/a5319ab5-en | OECD publication | Gated DOI | EP22 (G7 +0.2–1.3pp) |
| Yotzov et al. 2026 | NBER WP 34836 (doi.org/10.3386/w34836) | NBER working paper | Gated | EP23 (6,000 execs) |
| St. Louis Fed 2025 (Bick, Blandin, Deming) | stlouisfed.org/on-the-economy/2025/nov/state-generative-ai-adoption-2025 | Fed blog page | HTML page, not a fixed PDF | EP27 (+1.1–1.3%) |
| Penn Wharton Budget Model 2025 (Arnon) | budgetmodel.wharton.upenn.edu/p/2025-09-08-… | Budget-model report page | HTML report page | EP28 (+0.01pp TFP) |

## Internal-consistency flags (for a future judge run over report-productivity.pdf)

Recorded in the EP claims' `src=` notes; a judge run against report-productivity.pdf is meant to
adjudicate them:

- **EP18** — the macro paragraph's parenthetical points at "(Figure 4.2)", but the macro-studies
  table it introduces is Figure 4.4.28. Apparent stale/typo cross-reference.
- **EP22** — prose gives a single "0.2 to 1.3 percentage points" OECD range; Figure 4.4.28 splits it
  by country (US/UK +0.4 to +1.3; Italy/Japan +0.2 to +0.8). The prose range takes the floor of one
  group and the ceiling of the other.
- **EP23** — prose says Yotzov et al. found "minimal realized productivity gains"; the same row in
  Figure 4.4.28 lists a "+1.4% projected productivity boost". Realized vs projected, presented side
  by side.
- **EP14** — prose cites "Shen and Tamkin, 2025"; the report's own reference list dates the paper
  2026 (arXiv:2601.20245v2, "February 3, 2026"). Claim ids follow the prose; the file dates match.

## Derived files

- `claims-machine.txt` — extended with 28 atomic claims prefixed **EP** (EP1–EP28): 4 intro-framing,
  6 micro study-results (+2 micro framing), 5 macro (+3 macro framing/interpretation), 5 macro-table
  employment/productivity entries, and the flags above. Study results are `route=evidence` with the
  cited paper as truth-maker (`study-held` / `study-not-held` in `src=`); interpretive sentences are
  `route=evaluative`.
- `argument-productivity.txt` — the section as an argument tree. Root = the section's own thesis
  (§Productivity Trends p219, EP15: productivity effects are context dependent). Two root children =
  the two levels of evidence (Micro / Macro); Micro carries two findings (F-GAIN / F-DRAG) because
  the thesis is two-sided. 14 leaves; framing, evaluative, and employment claims are listed at the
  file's foot as excluded.
- **No judge run** (no `evidence/` chains, no `site/`), per the task.

## Deferred — MANIFEST.md wiring (a decision for whoever runs the judge)

`sources/MANIFEST.md` was **not** modified. It is calibrated for the coding section alone —
`single_source: true`, `report_page_offset: -99`, and a `sections:` map for pp100–102 — and a judge
run against the coding claims depends on those values. Wiring the productivity section for its own
judge run needs choices the manifest's current single-report shape does not express, and making them
blind would risk the working coding run:

- a second report file (`report-productivity.pdf`) with its own printed→PDF offset (**−218**: its
  page 1 is printed page 219) and its own `sections:` map (AI's Labor Impact / Productivity Trends →
  219, Figure 4.4.27 → 220, Figure 4.4.28 → 221);
- registering the 7 held papers as held truth-makers so the `study-held` EP claims can ground against
  them (this section is **not** `single_source` — that flag is coding-only);
- confirming the tool supports multiple report excerpts with per-file offsets in one manifest.

Left to the human / the judge-run session by design (this task did no judge runs).
