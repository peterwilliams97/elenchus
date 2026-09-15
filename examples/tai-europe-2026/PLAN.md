# PLAN — A Transformative AI Strategy for Europe (KIRA Center, v1.0, Sept 2026)

Corpus catalogue entry: `docs/todo/Consultant & institutional reports.md` § H (priority pick #1).
Template: `examples/mit-2026/PLAN.md` (single-source internal-consistency config). **Fetch done for
the report PDF only** (SHA-256 + version in `sources/MANIFEST.md`); no other fetch, no code, no
commit — PW reads this first.

Report: 206 pp, CC BY 4.0, no numbered reference list, 795 inline citation links across 222 hosts
(count reproduced from the held file — MANIFEST). The largest, most citation-dense artifact in the
corpus. This plan does **not** scope the whole document. It scopes **three slices only**; everything
outside them is explicitly OUT OF SCOPE (see last section).

## Slice 1 — the 11 objectives, stated twice (Part A claim, Part B implementation)

The document indexes both parts by the same objective IDs, so each objective appears once as a
claim-plus-recommendation (Part A) and once as a worked implementation (Part B). Checking whether the
Part B detail supports the Part A claim for the same ID is a directly scorable internal-consistency
task with 11 instances, and the document supplies the mapping itself (H.2). Both statements located,
printed page-referenced from the two tables of contents (report.txt:170–199 Part A, :213–246 Part B):

| ID   | Objective                          | Part A p | Part B p |
|------|------------------------------------|----------|----|
| IO-1 | Create a Member State Alliance for Supply Chain Security | 28 | 79 |
| IO-2 | Make Europe's institutions ready to act in a world with transformative AI | 31 | 90 |
| IO-3 | Secure Europe's share of global AI compute | 36 | 99 |
| IO-4 | Ensure resilience to AI crises | 40 | 108 |
| IO-5 | Make Europe the global leader in AI assurance technology | 45 | 118 |
| O1.1 | Secure ongoing access to frontier AI systems | 50 | 127 |
| O1.2 | Protect European assets | 53 | 133 |
| O2.1 | Make Europe the best place to build and scale high-growth companies | 57 | 141 |
| O2.2 | Develop indispensable assets across the AI value chain | 60 | 162 |
| O3.1 | Ensure prioritised and targeted use of EU rules to address risks from highly capable AI | 64 | 169 |
| O3.2 | Prepare for labour market impacts | 68 | 169 |

- **TOC quirk to verify at build:** Part B lists O3.1 and O3.2 both at p169; one is a TOC compression
  artefact. Confirm the O3.2 body start page against the section heading before page-referencing its
  leaves.
- **Representation (reconfigured to FAITHFULNESS):** Part B is the held SOURCE
  (`sources/report-partB.txt`, printed pp77–204); Part A is the TARGET. One leaf per ID — the Part A
  objective claim (`ref=IO-N pXX`, `src=report-partB.txt`, `route=source`) — and the judge checks
  whether the Part B "WHY IT MATTERS" block SUPPORTS that claim. Part A is deliberately not in
  `sources/`: holding it would let retrieval ground a Part A claim on its own wording. Decomposition
  is NOT in this plan.

### Built 2026-09-15 (Slice 1 only), reconfigured to faithfulness

`report.txt` split on `\f`, printed pp77–204 kept as `sources/report-partB.txt` (128 pages, 488,236
bytes). `claims-objectives.txt` now holds the **11 Part A leaves only** (IO-1-A … O32-A,
`src=report-partB.txt`, `route=source`); `argument.txt` is root → 11 objective nodes → one leaf each.
`sources/MANIFEST.md` lists `report-partB.txt` as the only held source, states the source/target
roles, and keeps `single_source: true` (claim and truth-maker are two parts of one report). Offset 0
verified. Each leaf = the framing paragraph under a Part A objective heading; its truth-maker = the
Part B "WHY IT MATTERS" block (pp79–169). No judge run (no model call, no `-review`).

- **O3.2 TOC quirk resolved.** Both O3.1 and O3.2 Part B genuinely begin on p169; not a compression
  artefact. O3.1's Part B is a four-line stub — "there are no detailed recommendations for this
  objective" — and O3.2 follows on the same page (its recommendations run from p170). p169 is the
  correct `ref` for both.

### Pre-registration — faithfulness verdict expected per Part A leaf

Written before any judge run (discipline: pre-register the attempt). Prediction: Part B supports six
Part A claims in full (IO-1, IO-2, IO-5, O2.1, O2.2, O3.2 → `faithful`), supports four only in part
(a Part A clause with no Part B truth-maker → `partial`, clause quoted below), and does not support
O3.1 at all (`unsupported`/`absent`).

**The miss that matters: a `faithful` verdict on O3.1.** Part B declines the objective outright, so a
judge that scores O3.1 `faithful` on topical overlap (both name the EU AI Act / GPAI Code) has
laundered abdication into support — the single result that would most discredit this slice.

- **O3.1 — `unsupported`.** Part A asserts an active mandate: "Europe must ensure that their
  enforcement is prioritised, targeted, proportionate, driven by frontier AI expertise, and insulated
  from political pressures." Part B (p169): "there are no detailed recommendations for this
  objective." The truth-maker declines the very enforcement-shaping the claim asserts.
- **IO-4 — `partial`, prevention clause dropped.** Part A clause with no Part B support: "overhauling
  Europe's ability to prevent and respond to AI crises and emergencies." Part B narrows to resilience
  only — "Many of these risks cannot be fully managed upstream" — so the *prevent* leg is unsupported.
- **O1.1 — `partial`, backstop clause dropped.** Part A clause with no Part B support: "underwritten
  by contractual guarantees backstopped by physical leverage." Part B keeps the 'compute for access'
  deal but substitutes "Funding a European fast follower can reduce the downside if access is
  restricted" for the physical-leverage backstop.
- **O1.2 — `partial`, inbound-screening clause dropped.** Part A clause with no Part B support:
  "protect its AI and semiconductor assets through investment screenings with sufficient enforcement
  capacity, while ensuring that at-risk European companies can flourish without foreign capital." Part
  B leads instead with the outbound concern ("how much of its own capital and technology is flowing
  into foreign AI and semiconductor industries").
- **IO-3 — `partial`, community-interest clause dropped.** Part A clause with no Part B support: "A
  'European Way' of building AI compute should combine speed with respect for local community
  interests." Part B keeps only the geopolitical-weight / revocable-access framing.

## Slice 2 — headline positive-control claim chain (compute: Europe vs one Malaysian site)

The catalogue (H.7) flags this as the known-good path: executive summary, body and footnote all agree,
and the footnote shows its arithmetic. Use it to calibrate — a run that fails to reconcile it is
miscalibrated. Three passages, one claim, page-referenced (printed pages from the page footers):

- **Executive summary** (report.txt:382–383, printed p6): "Europe hosts only three times as much AI
  computing capacity as a single data centre site being completed in Malaysia this year."
- **Body, IO-3 lead-in** (report.txt:942–943, printed p24): "By the end of 2026, a single data centre
  site in Malaysia will reach roughly one third of the AI compute capacity of Europe, including the UK
  and Norway."
- **Footnote 2** (report.txt:951, printed p24): "The Nusajaya site in Malaysia will reach 0.662 GW in
  2026 … compared to 2.1 GW for all of Europe." → 2.1 / 0.662 = 3.17 ≈ "three times" / "one third".
- **Represent as** three leaves (`ref` p6, p24, p24-fn2) on ONE internal-consistency node; expected
  verdict `faithful`/consistent across all three. This is the positive control, not a defect probe.
- A second in-document positive control exists (Figure 1, Epoch index, discloses the metric that did
  NOT accelerate — H.6) but it is not a claim *chain*; note it, don't scope it here.

## Slice 3 — empirical section on AI progress and Europe's position

Section: **"Why Europe needs to prepare for transformative AI"**, printed pp20–26 (report.txt:934 body
heading region), three subsections — *AI progress is extraordinarily rapid* (p21), *Near-term
transformative AI is plausible* (p23), *Europe risks marginalisation* (p24). This is the citation-dense
core the whole urgency framing rests on.

- **Link inventory (from the PDF's own URI annotations, PDF pages 21–25):** 61 links. Per-page:
  p21=9, p22=16, p23=7, p24=17, p25=12. Top hosts across the section: epoch.ai (14), anthropic.com (5),
  metr.org (5), pacingthefrontier.com (3), europe2031.ai (2), arxiv.org (2), forethought.org (2),
  theinformation.com (2), esrb.europa.eu (2), euractiv.com (2), media.defense.gov (2); singletons incl.
  internationalaisafetyreport.org, wemustactnow.ai, rand.org, openai.com, forbes.com, fortune.com,
  bloomberg.com, reuters.com, aisi.gov.uk, blog.google, redwoodresearch.org.
- **~20 load-bearing citations to HOLD** (anchor sentence → host; capture the exact link target at
  extraction, mark each `evidence`-route `unverifiable` for grounding until a later fetch pass):

  1. Epoch Capabilities Index growth (+8.3 then +15.5 pts/yr), Fig 1 → epoch.ai
  2. METR task-horizon doubling (~7-month) → metr.org
  3. Coding-agent success-rate chart (Fig 2, "Source: Anthropic") → anthropic.com — INTERESTED PARTY (H.6)
  4. Frontier capability / RE-bench-style automation claim → metr.org / epoch.ai
  5. "capability growth is 'jagged'" / RL-on-verifiable-outcomes caveat → (acknowledge-then-dismiss, H.7)
  6. July 2026 ~200 economists incl. 16 Nobel laureates statement → wemustactnow.ai — UNVERIFIED (H.8)
  7. Frontier-lab-employees statement (same month) → (host to capture) — UNVERIFIED (H.8)
  8. OpenAI 2.3× projected → run rate $20B→$40B in 8 months → theinformation.com / fortune.com
  9. Anthropic 4× optimistic → already 5.2× by May 2026 → theinformation.com
  10. AI capex 2026 ~€650B, +70%/yr → (host to capture)
  11. International AI Safety Report risk list → internationalaisafetyreport.org
  12. OpenAI-agents hidden-message-board / Hugging Face hack → (host to capture) — UNVERIFIED (H.8)
  13. None of the ten most valuable AI companies are European → (host to capture)
  14. European model developers' run rate < 2% of OpenAI+Anthropic combined → (host to capture)
  15. OpenAI+Anthropic combined revenue ~25× in two years, $4B→>$100B → (host to capture)
  16. EU ~5% of global AI compute vs 75% US, 15% China → epoch.ai / europe2031.ai
  17. Malaysia site ≈ one third of Europe (Nusajaya 0.662 vs 2.1 GW) → epoch.ai — SHARED WITH SLICE 2
  18. Two US data centres (Stargate Abilene + New Carlisle, 2.4 GW) exceed Europe → (host to capture)
  19. 2025: Europe received 5–6% of global private AI investment → (host to capture)
  20. Market pays ~100× more for US frontier than European fast follower → (host to capture)

- Items 16 and 17 route through europe2031.ai / epoch.ai — the compute figures the catalogue flags as
  sitting inside the authors' own citation orbit (H.5). Hold them; the grounding pass is where the
  provenance-loop test lives, and it is not this plan.
- "(host to capture)" = the anchor is a compound sentence whose specific link target must be read off
  the PDF annotation at extraction, not guessed. Do NOT invent a host.

## Out of scope (explicit)

Everything not in the three slices above, including: the other 8 objectives-worth of Part A/B body and
all §-recommendation decomposition beyond the 11 objective pairs; the compute (pp189–196) and robotics
(pp197–204) deep dives; the costed European-frontier-AI-project hypothetical (pp174–187); the
attribution-trap contributor test (H.3); the AI-provenance re-verification pass over all 795 links
(H.4) and the citation-loop provenance test (H.5) except where slice-3 items 16–17 touch it; Figure 2
interested-party grounding (H.6) beyond holding the link; the acknowledge-then-dismiss inference step
(H.7) beyond holding item 5; and any model-calling / `-review` / `-edge` run. No claims files,
`argument.txt`, decomposition, or run sequence are authored here — those are the next plan's work.
