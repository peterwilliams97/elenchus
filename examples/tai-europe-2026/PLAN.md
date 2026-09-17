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

| ID   | Part A p | Part B p | Objective                          |
|------|----------|----------|------------------------------------|
| IO-1 |       28 |       79 | Create a Member State Alliance for Supply Chain Security |
| IO-2 |       31 |       90 | Make Europe's institutions ready to act in a world with transformative AI |
| IO-3 |       36 |       99 | Secure Europe's share of global AI compute |
| IO-4 |       40 |      108 | Ensure resilience to AI crises |
| IO-5 |       45 |      118 | Make Europe the global leader in AI assurance technology |
| O1.1 |       50 |      127 | Secure ongoing access to frontier AI systems |
| O1.2 |       53 |      133 | Protect European assets |
| O2.1 |       57 |      141 | Make Europe the best place to build and scale high-growth companies |
| O2.2 |       60 |      162 | Develop indispensable assets across the AI value chain |
| O3.1 |       64 |      169 | Ensure prioritised and targeted use of EU rules to address risks from highly capable AI |
| O3.2 |       68 |      169 | Prepare for labour market impacts |

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

**The miss that matters: a `faithful` verdict on O3.1.** Part B declines the objective outright, so 
a judge that scores O3.1 `faithful` on topical overlap (both name the EU AI Act / GPAI Code) has
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

### Result — 15 Sept, Sonnet N=3

One judge run, `-backend anthropic -model claude-sonnet-4-6 -retrieve bm25 -n 3` against
`sources/report-partB.txt` (the whole held Part B, 128 pp, in the pool for every leaf). Chain:
`evidence/2026-09-15-slice1-faith/sonnet/claims-objectives.faithfulness.clean.jsonl` (the 11 real
leaves; see the comment-line bug note below). Verdicts, spread `k/N`, and the critic's named clause
(one line each, from `critic_finding`):

| ID   | Pre-reg     | Verdict     | Spread | Critic's named clause                                                                 |
|------|-------------|-------------|--------|---------------------------------------------------------------------------------------|
| IO-1 | faithful    | overstated  | 3/3    | "so far failed to translate" overstates source's "dispersed, never leveraged in tandem" — no active-failure framing |
| IO-2 | faithful    | unsupported | 3/3    | "decisions under uncertainty / time pressure" and "without extraordinary capacity, none of the other objectives" not in Part B |
| IO-3 | partial     | overstated  | 2/3    | "'European Way' combining speed with respect for local community interests" — source lists community concerns only as a political success condition |
| IO-4 | partial     | partial     | 3/3    | "overhauling Europe's ability to **prevent** and respond" added as co-equal pillar beyond source's resist/absorb/recover/adapt |
| IO-5 | faithful    | partial     | 3/3    | "national security vs commercial/scientific" distinction and "lead the underlying AI technology" not framed that way in Part B |
| O1.1 | partial     | partial     | 3/3    | adds "**sovereignty**" and frames the episode as "export controls on Anthropic's most capable models" — not the predicted physical-leverage backstop |
| O1.2 | partial     | partial     | 3/3    | "flourish **without foreign capital**" contradicts source's anti-capture fund, which mobilises public AND private capital — not the predicted inbound-screening clause |
| O2.1 | faithful    | absent      | 3/3    | no passage makes the three-part firms-grow-fast / retain-in-Europe / supply-side-reform argument |
| O2.2 | faithful    | unsupported | 3/3    | opening premise — "experts disagree which layer of the value chain creates most value" — not in Part B |
| O3.1 | unsupported | absent      | 3/3    | no passage states Europe is well-positioned to lead AI safety via the AI Act / GPAI Code |
| O3.2 | faithful    | unsupported | 3/3    | source says nothing linking Europe's share of AI value creation to its capacity to absorb labour-market shocks |

**Scored against the pre-registration: 5/11.**

- **The miss that would matter did not occur.** O3.1 came back `absent` (3/3), as registered — the
  judge did not launder Part B's abdication ("there are no detailed recommendations for this
  objective") into a `faithful` verdict on topical AI-Act/GPAI overlap.
- **The four predicted gaps all fired** (IO-4, O1.1, O1.2, IO-3 → all non-faithful), but two on a
  **different clause** than predicted. IO-4 (prevent leg) and IO-3 (European Way / local community)
  fired on the predicted clause; O1.1 fired on "sovereignty" + the Anthropic-episode framing rather
  than the physical-leverage backstop, and O1.2 on "without foreign capital" rather than the
  inbound-vs-outbound screening asymmetry.
- **The six predicted `faithful` all came back non-faithful** (IO-1, IO-2, IO-5, O2.1, O2.2, O3.2 →
  overstated / unsupported / absent). This is the pre-registration's cost: the prediction that Part B
  would fully support the majority of Part A claims was wrong across the board.

**Reading — for PW's adjudication, not decided here.** Every one of the 11 verdicts names a Part A
clause the critic reports absent from the retrieved Part B passages. Whether each is a **real gap**
(Part A over-claims relative to its own Part B implementation) or a **clause-level false positive**
(the critic penalising a paraphrase or a compression the same document does support elsewhere) is a
per-leaf human call. Prior single-source-consistency precedent (DORA, ~40% flag precision) says a
material share of these will be false positives on read; the verdicts are a worklist, not a finding.
Adjudication drafts land in `evidence/2026-09-15-slice1-faith/adjudications.txt`.

**Run notes.**
- **Comment-line bug.** `splitSummary` did not skip `#`-prefixed lines, so the run judged all 33
  header-comment and `# === ID ===` section-heading lines in `claims-objectives.txt` as claims (44
  judged, only 11 real). The `.clean.jsonl` chain used above drops them to the 11 leaves; the raw
  `claims-objectives.faithfulness.jsonl` carries all 44. Fix the parser before the next run.
- **Cost: $6.26 gross, ~$1.50 attributable to the 11 real leaves** — the rest was spent judging the
  33 comment/heading lines above.
- **Retrieval did not starve any leaf:** all 11 expected Part B pages, and in fact the whole held
  Part B (128 pp), were in the pool for every claim, so no verdict is a retrieval miss.

### Adjudicated 16 Sept (PW)

PW read all 11 leaves against the held Part B and filed a verdict per leaf in `adjudications.txt`
(example root; read by the `-review` overlay, spec/SERVE.md § Adjudications). PW upheld six leaves and
overturned five to `faithful` (table below).

Two agreement counts once differed by the vocabulary the compare runs over: against the machine's
**raw** verdicts PW agrees on **6/11** (every upheld leaf), but the string-equality compare against
the `single_source` display spelling scored PW's `absent` on O2.1 and O3.1 as disagreements
(`single_source` mode renames `absent` to `uncorroborated`, so it derives as *weakened* not *failed*,
spec/ARGUMENT.md), giving 4/11 — a vocabulary artifact, not a read PW changed. **Resolved:** `Agree`
scores against the pre-remap raw verdict (the `machine` map in `presentArgument`, `assay.go`, is
captured before the remap) and `canonVerdict` folds `uncorroborated`→`absent`
(`internal/adjudicate/adjudicate.go`), so both counts reconcile at **6/11**. The remaining question is
display, not counting — the card still shows `uncorroborated` while the human files `absent` — a UX
call parked in `docs/todo/design-notes.md` § Single-source adjudication.

| ID   | Machine     | PW          | Reason (one line, from `adjudications.txt`)                                            |
|------|-------------|-------------|---------------------------------------------------------------------------------------|
| IO-1 | overstated  | faithful    | "failed to translate" reads as "not yet", not an active failure; Part B p79 ("dispersed", "never leveraged in tandem") carries the same state |
| IO-2 | unsupported | unsupported | the "any of the other objectives" dependency and the "under uncertainty / time pressure" framing are A's, absent from Part B p90 |
| IO-3 | overstated  | faithful    | Part B IO-3 ACTIONs (municipal revenue-sharing, community-infrastructure investment) implement the community-respect commitment; judge saw only the "why it matters" block |
| IO-4 | partial     | faithful    | Part B IO-4 ACTIONs (CBRN plan, loss-of-control emergency plans, Crisis Coordination Hub) implement both prevent and respond; judge saw only the "why it matters" block |
| IO-5 | partial     | faithful    | Part B IO-5 actions (RAND-tiered weight security, export security, SL5 data centre) implement secure-access-controls; clause 2 was the judge misreading A's own referent |
| O1.1 | partial     | partial     | A names Anthropic and export controls; Part B O1.1 describes only access "suddenly withdrawn by foreign governments" — A is more specific than B |
| O1.2 | partial     | faithful    | Part B O1.2 ACTION (anti-capture facility via national public banks, "to keep a strategic asset under European control") is the without-foreign-capital alternative; judge read "private" as possibly foreign |
| O2.1 | absent      | absent      | the unpredictability-of-value rationale is not in Part B §O2.1 (pp141–161); B argues breadth as a "coherent package", not from where value lands |
| O2.2 | unsupported | unsupported | the "experts disagree which layer creates most value" premise is not in Part B §O2.2 (pp162–168); B states the conclusion without it |
| O3.1 | absent      | absent      | Part B p169 declines the objective outright — "there are no detailed recommendations for this objective" |
| O3.2 | unsupported | unsupported | the link from Europe's share of AI value creation to its shock-absorption capacity is not in Part B §O3.2 (pp169–173); it is A's own |

**Totals.** Five machine verdicts overturned to `faithful` (IO-1, IO-3, IO-4, IO-5, O1.2); six upheld
(IO-2, O1.1, O2.1, O2.2, O3.1, O3.2).

- **Overturns, by cause.** Four are retrieval-scope (IO-3, IO-4, IO-5, O1.2): BM25 fed the judge each
  objective's "WHY IT MATTERS" prose, not its ACTION blocks, so the implementation the Part A claim
  promises sat outside the judged passages. One is phrasing (IO-1): the critic read "failed to
  translate" as an active-failure assertion the source does not make, where PW reads it as "not yet".
- **Upheld, in three kinds.** Rationale stated in A and absent from B (IO-2, O2.1, O2.2, O3.2) — A
  supplies a premise or dependency B never argues; factual specificity A adds over B (O1.1 — Anthropic
  / export controls where B says only "withdrawn by foreign governments"); objective declined with no
  implementation (O3.1).
- **The retrieval-scope cause is a tool lesson, filed as a todo, not fixed here** —
  `docs/todo/plan-segmentation.md` § After the seam. It refines the Run-notes line above: the whole
  Part B was in the candidate *pool*, but BM25 top-k *selected* each objective's prose over its own
  ACTION bullets, so the four are top-k selection misses, not candidacy misses. For a document keyed
  by the same ids as its claims, retrieval should return the whole section for the id rather than
  BM25 top-k.

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

### Built 2026-09-16 (Slice 2)

No judge run (no model call, no `-review`). Three leaves in `claims-slice2.txt` — `CMP-EXEC` (p6),
`CMP-BODY` (p24), `CMP-FN2` (p24 footnote 2) — hung as one internal-consistency node (`CMP`) under
`ROOT` in `argument-slice2.txt`. Source is the whole report: `sources/report.txt` added to
`MANIFEST.md` as a held Slice-2 source (`src=report.txt`, `route=source`), with a "Roles for Slice
2" section stating the internal-consistency config; `single_source: true` already holds. All three
statements printed verbatim from `report.txt`, page-referenced, hyphenation joined and
footnote-reference digits dropped per the Slice-1 text convention.

- **Footnote-2 text differs from the quote above.** The verbatim footnote 2 (report.txt:950–952)
  reads: "OpenAI Stargate Abilene and Anthropic-Amazon New Carlisle will together reach 2.4 GW
  facility power in 2026, compared to 2.1 GW for all of Europe. The Nusajaya site in Malaysia will
  reach 0.662 GW in 2026, with 0.34 GW already in operation." The "compared to 2.1 GW for all of
  Europe" clause attaches to the two US data centres (2.4 GW), and Nusajaya's 0.662 GW is a separate
  sentence — the elision "0.662 GW … compared to 2.1 GW" in this plan's bullet reattaches it. Both
  operands sit in the one footnote, so 2.1 / 0.662 = 3.17 still reconciles; `CMP-FN2` carries the
  footnote verbatim, not the elided form.

### Pre-registration — verdict expected per leaf

Written before any judge run (discipline: pre-register the attempt). Prediction: **all three
`faithful`**. The two framings reconcile through footnote 2's arithmetic — 2.1 GW (all of Europe) /
0.662 GW (Nusajaya, Malaysia) = 3.17, which is both "three times as much" (CMP-EXEC) and "roughly
one third" (CMP-BODY). This is the positive control: **a non-faithful on ANY leaf is a calibration
miss** (the critic penalising a paraphrase or arithmetic the same document supports), **not a
document finding** — the report is internally consistent here by construction (H.7).

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

### Built 2026-09-16 (Slice 3 fixtures) — no judge run

`claims-slice3.txt` (20 leaves) and `argument-slice3.txt` written; no model call, no `-review`. Each
leaf is one Slice-3 cited-source row: the report's sentence VERBATIM (reconstructed from `report.txt`,
not from the MANIFEST anchor column — items 11, 12, 14 were paraphrased there and are fixed here),
page-referenced, `route=evidence`, `cites` naming the held cited page (`cited/<stem>.txt`). The
20-leaf set uses the MANIFEST's row granularity: item 8 splits into `08a` (Epoch, present) + `08b`
(Bloomberg, walled), item 16 into `16a` (europe2031, present) + `16b` (Epoch, figure-only); items
14≡15 and 17≡18 each share one target and one leaf. The argument tree's three nodes are the report's
own subsection headings (`report.txt:164-166`): AIPROG (5 leaves), NEARTERM (8), EUROPE (7).

**Deviation from this prompt's "src= blank" for the four unverifiable leaves.** A blank `cites` does
NOT render `unverifiable` — `missingCites("")` returns nil, so the leaf would be judged against the
pooled corpus (`assay.go:3506`, `:674`). `unverifiable` (no model call) fires only for a cited id that
is NAMED but not in the held manifest (`assay.go:743-750`). So items 5, 19, 08b, 09 each cite their
(un)available truth-maker and are left out of the held set: `08b`/`09` cite the walled file (fetched
but not usably in the corpus), `05`/`19` cite an `uncited/…` id no held document answers to. Flagged
for PW: revert to blank only if the intended semantics changed.

### Run prerequisites — the grounding run is NOT executable as-is (three gaps)

Verified 2026-09-16 against HEAD. Until all three close, the command below renders **20/20
unverifiable** at best, or silently under-runs — a partial run, not a smaller pass.

1. **Cited ids are not held.** MANIFEST § Slice-3 cited sources lists the 18 files in a markdown
   TABLE; `manifest.LoadHeld` (`internal/manifest/manifest.go:145`, regex `:141`) reads only
   backticked `- ` + "\`id\`" bullets, so `c.held` contains none of them → every `cites` leaf is
   `unverifiable`. Fix: register the 16 grounding ids as bullets (`- ` + "\`cited/01-epoch.ai.txt\`" …),
   leaving `08b`/`09`/`uncited/*` deliberately OUT so those four stay unverifiable.
2. **No parser for web-page captures.** `retrieve.passagesForFile` (`internal/retrieve/retrieve.go:182`)
   routes a `cited/` file to `splitFile`, which emits a passage only on a Hansard speaker line. The
   captures have none (3 accidental glossary `NOTE:`/`HPIM:`/`CBRN:` matches across 18 files), so the
   index gets ~3 garbage passages and 15 files contribute nothing. Fix: a `citedPassages` parser
   (paragraph split, id base `cited/<stem>`, a `SourceCited` role), a `/cited/` case in
   `passagesForFile`, and a spec line in `spec/TREE.md § Cite-scoped retrieval` beside the
   `leaderboards/` shape it copies.
3. **No cite-base mapping for `cited/`.** `citedExternalBases` (`assay.go:1006`) maps only
   `paper:<stem>` and `<x>.txt` → `<x>`; a `cited/<stem>.txt` cite resolves to base `cited/<stem>`
   only once the parser (gap 2) mints that base. These two are the same change and land together.

Gaps 2–3 are Go code on a branch whose `assay.go` already carries an unrelated in-progress diff — a
separate topic/branch, not folded into this fixture commit. This is a hand-off, not "should work".

### Pre-registration — verdict expected per leaf (written before any judge run)

Driven by the MANIFEST `fig` column: **present-as-text → `faithful`** (the cited page states the
number) **unless flagged**; **figure-only → `absent`** (the datum renders only in an interactive
chart / data tool, so it is not in the static extraction — a stated limit of the fixture, read as
"checked, not located in text", not a document defect); **absent-in-extraction → `absent`**;
**not-held → `unverifiable`** (no model call).

| id | item | fig | expected | read first if NOT as expected |
|----|------|-----|----------|-------------------------------|
| c01-epoch | 1 | figure-only | absent | — |
| c02-metr | 2 | figure-only | absent | — |
| c03-anthropic | 3 | present | **flag** | Fig 2 is "Source: Anthropic" — the interested party (H.6); a non-faithful here is the finding |
| c04-epoch | 4 | present | faithful | — |
| c05-jagged | 5 | uncited | unverifiable | — |
| c06-wemustactnow | 6 | present | **flag** | the ~200-economist statement is post-cutoff / catalogue-unverified (H.8) |
| c07-pacing | 7 | present | **flag** | the 1,000+-employee statement is catalogue-unverified (H.8) |
| c08a-epoch | 8 | present | faithful | — |
| c08b-bloomberg | 8 | hand-fetch | unverifiable | — |
| c09-theinfo | 9 | hand-fetch | unverifiable | — |
| c10-epoch | 10 | figure-only | absent | — |
| c11-iasr | 11 | present | faithful | — |
| c12-redwood | 12 | present | **flag** | the OpenAI/Hugging Face incident is post-cutoff / catalogue-unverified (H.8) |
| c13-forbes | 13 | present | faithful | ("none European" is derived from the list, not stated) |
| c14-epoch | 14/15 | figure-only | absent | — |
| c16a-europe2031 | 16 | present | **flag** | europe2031.ai sits in the authors' own citation orbit (H.5) |
| c16b-epoch | 16 | figure-only | absent | — |
| c17-epoch-fdc | 17/18 | figure-only | absent | — |
| c19-privateinv | 19 | uncited | unverifiable | — |
| c20-foxphilip | 20 | absent-in-extraction | absent | if the 100× figure IS in the page on a hand-read, that's a finding (extraction miss, not a citation gap) |

**Totals expected:** 4 `faithful` (c04, c08a, c11, c13), 5 `flag` (c03, c06, c07, c12, c16a — a
non-faithful verdict on any of these is the finding to read first), 7 `absent` (c01, c02, c10, c14,
c16b, c17, c20; c20 doubles as a hand-read prompt), 4 `unverifiable` (c05, c08b, c09, c19). The flag
leaves are NOT pre-registered `faithful`: they are where the cited page most plausibly does NOT carry
what the report claims (interested-party chart, post-cutoff statements, a citation-orbit compute
figure), so a `faithful` there is the result to distrust, and a non-faithful is the slice's payload.

> **Data addendum (17 Sept) supersedes the `absent` rows for c14 and c16b below** — see next section.

### Data addendum (17 Sept) — Epoch chart CSVs re-pre-register c16b / c14 / c08a

PW fetched Epoch's underlying chart DATA (`sources/cited/data/*.csv`, SHA-256 in
`MANIFEST-slice3.md`) — the numbers the `figure-only` pages render only as interactive charts. c16b,
c14, c08a now cite the CSV first (primary) and keep their page id as secondary
(`claims-slice3.txt`). The `fig`-column logic that pre-registered c14/c16b `absent` (figure-only → not
in the static strip) no longer applies: the datum IS in the held CSV. Re-pre-registered, written
before any judge run:

| id | truth-maker CSV | expected | basis / caveat |
|----|-----------------|----------|----------------|
| c08a-epoch | `14-ai_companies_revenue_reports.csv` | faithful | OpenAI run-rate rows ($20bn → $40bn, first 8 months 2026) present in the export |
| c14-epoch | `14-ai_companies_revenue_reports.csv` | faithful | OpenAI+Anthropic combined run-rate and European-developer rows present; ~25× / $4bn→$100bn checkable |
| c16b-epoch | `ai_supercomputers.csv` | **faithful for EU ≈5% and US ≈75% (by MW); NOT for China 15%** | China is 11% by MW, 7% by H100e — the report's 15% overstates it. Shares computed by PW over the 482 `Status=Existing` systems by Power Capacity (MW): US 73.9 / China 11.2 / EU 5.5 |

c16b is a **compound** claim (three shares); the CSV grounds two and refutes the third, so a uniform
`faithful` is the wrong result — the China 15% is the overstatement to surface. Falsifier: recompute
the by-MW country shares over `Status=Existing` rows of `ai_supercomputers.csv`; if China lands near
15% (not ~11%) the pre-registration is wrong and c16b is faithful after all.

**Fourth run prerequisite (CSV-specific), on top of gaps 1–3 above.** `retrieve.Load` ingests `.txt`
only (`internal/retrieve/retrieve.go:131`); `citedExternalBases` (`assay.go:1007`) maps only
`paper:`/`.txt`; the plain segmenter splits on blank lines a CSV lacks, and CSV records span multiple
physical lines (quoted fields wrap). Needs a CSV rule — `encoding/csv`, header + one data row per
passage, id base `cited/data/<file>` — plus a `.csv` case in `Load`'s walker and in
`citedExternalBases`. Go code on its own branch, not folded into this fixture edit; until it lands the
three re-pointed leaves render `absent` on their secondary page cite, not `faithful`.

### Judge command — N=3 Sonnet against `sources/cited/` (blocked on the three prerequisites)

Do not run until the prerequisites above close; PW runs it (never call the model API from a session).
Flags go BEFORE the positional claims file. `-source sources/cited` is the fetched-citation corpus;
`-manifest` gives the held set that decides `unverifiable`; `-refs` carries the report page per leaf
for the `-review` deep-links.

```sh
# run from examples/tai-europe-2026/
source ../../setkey.sh && ../../assay -backend anthropic -model claude-sonnet-4-6 \
  -retrieve bm25 -n 3 \
  -manifest sources/MANIFEST.md \
  -source sources/cited \
  -argument argument-slice3.txt -refs claims-slice3.txt \
  -chain-dir evidence/2026-09-16-slice3/sonnet-cited \
  claims-slice3.txt
```

## Slice 4 — the leverage premise (what Europe controls in the chip supply chain, and what it buys)

IO-1 ("Create a Member State Alliance for Supply Chain Security", Part A pp28–30, Part B pp79–89) and
the compute deep dive (pp189–196) rest one recommendation — form an Alliance around Europe's
supply-chain assets — on a single premise: that Europe controls chokepoints in the AI chip supply
chain, that those chokepoints have no ready substitutes, and that this is leverage worth organising
around. Slice 4 grounds every SOURCED factual claim on that premise against the report's own external
citation (route=evidence, cite-scoped, held under `sources/cited/4x-*`), the same config as Slice 3.
This pulls the compute deep dive IN from the original out-of-scope list (below); robotics stays out.

- **Link inventory** (PDF URI annotations read with pypdf, by rect y-coordinate against the text line
  — never guessed): p28 = 9 links, p190 = 14, p191 = 10, p192 = 16, p193 = 19; Part B pp79–89 carries
  only eur-lex / state.gov / government.nl policy-instrument links, no factual chokepoint citations, so
  no Part B leaf. The chokepoint facts live in IO-1 Part A p28 and the deep dive pp190–192.

### Built 2026-09-16 (Slice 4 fixtures) — no judge run, no commit

`claims-slice4.txt` (13 leaves), `argument-slice4.txt` (root → 5 premise nodes), `sources/MANIFEST-slice4.md`
(no `single_source`). 14 cited pages fetched with `curl` (Chrome UA) into `sources/cited/4x-*.html`,
stripped to `.txt` (stdlib HTMLParser); 12 held HTTP 200, 2 walled (euractiv 403 Cloudflare, Bloomberg
403). SHA-256 over each held `.html` in MANIFEST-slice4. The five premise nodes:

- **CONTROL** (2) — Europe holds chokepoints: assets enumeration (s01, chipexplorer) + EUV made only
  by the Dutch ASML (s02, dwarkesh).
- **SUBST** (4) — no ready substitutes: HBM sold out (s03), TSMC 2nm sold out (s04), inference demand
  ~10x vs supply ~3.4x (s05), GPUs "sold out across all GPU types" (s06).
- **SHARE** (2) — Europe hosts ~5% of global AI compute (s07a europe2031, s07b epoch).
- **DEPEND** (4) — access can be cut: Glasswing (s08), US export directive (s09), ENISA delay (s10),
  no preferential chip access (s11).
- **LEVERAGE** (1) — indirect powers turn assets into leverage (s12, antonleicht).

**Two scope gaps flagged (not fabricated to close):** (1) the report's strongest "share of the layer"
claim — "ASML (Netherlands) and suppliers such as Zeiss (Germany) or Trumpf (Germany) hold monopolies
over relevant [layers]", Figure 10 (report.txt:2818, O2.2, ~p60) — sits OUTSIDE the scoped pages. This
slice grounds the assets *enumeration* and the EUV-monopoly *step* but not that layer-*monopoly*
assertion, which is what the leverage thesis most leans on; its own citation is a separate fetch.
(2) **ASM** (ASM International), named in the task, does not appear anywhere in `report.txt` — no leaf.

### Pre-registration — verdict expected per leaf, and what a non-faithful means for the strategy

Written before any judge run. Expected verdict is driven by the MANIFEST `fig` column: **present →
`faithful`** unless flagged; **figure-only / headline-only → `absent`** (datum not in the static
extraction — a stated fixture limit, read as "checked, not located", not a document defect);
**walled → `unverifiable`** (no model call). The **flag** leaves are where the cited page most
plausibly does NOT carry what the report claims; a `faithful` there is the result to distrust, a
non-faithful is the slice's payload.

| leaf | node | fig | expected | if non-faithful, what it means for the strategy |
|------|------|-----|----------|--------------------------------------------------|
| s02-dwarkesh | CONTROL | present | **faithful — LOAD-BEARING** | if the "only the Dutch ASML makes EUV" chokepoint is overstated, IO-1's premier bargaining chip is weaker than claimed and the whole Alliance-leverage thesis loses its anchor |
| s01-chipexplorer | CONTROL | figure-only | absent | the report's single most-cited "what Europe controls" sentence grounds on an interactive data tool whose ASML/Zeiss/Trumpf data does not extract — already unconfirmable in text |
| s03-tomshardware | SUBST | present | faithful | if HBM is not a real sold-out bottleneck, the "no ready substitutes" pillar (IO-1's Alliance-secretariat leverage map) softens |
| s04-chosun | SUBST | headline-only | absent | headline confirms "TSMC 2nm sold out"; the "until 2028" specificity is not in the extraction — read as extraction miss, not citation gap |
| s05-epoch-crunch | SUBST | present | faithful | if demand is NOT outrunning supply ~10x vs ~3.4x, the "compute crunch → chokepoint = leverage" mechanism weakens |
| s06-semianalysis | SUBST | present | faithful | corroborates s05; a miss here is a scarcity-claim overreach |
| s07a-europe2031 | SHARE | present | **flag (H.5)** | europe2031 is the authors' own scenario paper; a faithful only proves the orbit is self-consistent. A non-faithful undercuts the compute-poverty diagnosis the whole deep dive rests on |
| s07b-epoch | SHARE | present | **flag (H.5)** | same 5% figure, second orbit source; distrust a faithful, read a non-faithful as the finding |
| s08-glasswing | DEPEND | present | **flag (H.6)** | Anthropic's own page is the truth-maker; a faithful proves only that the interested party says what the report says. If non-faithful, the urgency case (Europe gets cut off) leans on a source that doesn't support it |
| s09-fable-mythos | DEPEND | present | **flag (H.6)** | same interested-party caveat for the US-export-directive claim, the sharpest "access was actually withdrawn" evidence |
| s10-enisa | DEPEND | walled | unverifiable | both truth-makers (euractiv, Bloomberg) are 403 walls; the ENISA-delay claim cannot be grounded until hand-fetched |
| s11-europarl | DEPEND | present | faithful (partial risk) | the EP page is a debate agenda; "no preferential chip access" is the report's gloss, so an `overstated`/`partial` here is the report editorialising a citation |
| s12-antonleicht | LEVERAGE | present | partial | the essay grounds the leverage *framing*; the specific export-controls / investment-screening / ACI list is the report's own — expect the framing to hold and the instrument list to be absent |

**Totals expected:** 5 `faithful` (s03, s05, s06, s11 with partial risk; s02 as the load-bearing one),
4 `flag` (s07a, s07b, s08, s09 — a non-faithful on any is the payload), 2 `absent` (s01, s04),
1 `unverifiable` (s10), 1 `partial` (s12). The load-bearing claim is **s02**: it is the one leaf whose
non-faithful would discredit the Alliance recommendation outright, so it is pre-registered `faithful`
and is the result to read first if it comes back otherwise.

### Run prerequisites & judge command (blocked on the three Slice-3 gaps)

Same three gaps as Slice 3 (PLAN.md §Slice 3 "Run prerequisites"): register the `cited/4x-*` ids for
`LoadHeld`, add the `/cited/` paragraph parser in `internal/retrieve`, and map `cited/<stem>.txt` in
`citedExternalBases`. Until they close every leaf renders `unverifiable`. PW runs the command below
(never call the model API from a session). N=3 Sonnet, flags BEFORE the positional claims file:

```sh
# run from examples/tai-europe-2026/
source ../../setkey.sh && ../../assay -backend anthropic -model claude-sonnet-4-6 \
  -retrieve bm25 -n 3 \
  -manifest sources/MANIFEST-slice4.md \
  -source sources/cited \
  -argument argument-slice4.txt -refs claims-slice4.txt \
  -chain-dir evidence/2026-09-16-slice4/sonnet-cited \
  claims-slice4.txt
```

## Slice 5, claim 2 — why NOT build / why NOT open-weight (O1.1, p51 + pp174–187)

O1.1 (“Secure ongoing access to frontier AI systems”, Part A pp50–53) argues Europe should pursue
‘compute for access’ deals rather than count on its own frontier build or on open-weight models. Slice 5
grounds the five sourced claims behind that “why not build / why not open-weight” step against the
report’s OWN citations (route=evidence, cite-scoped, same config as Slice 3/4). This pulls the €790bn
cost and the half-hearted-is-worst claims IN from the costed-hypothetical pp174–187 (below, still OUT of
scope otherwise); the full costing table, coalition mechanics and robotics stay out.

- **Link inventory** (PDF URI annotations, pypdf by rect y): p51 = the open-weight sentence carries two
  links (epoch open-vs-closed ECI gap on “several months”; internationalaisafetyreport.org on the
  misuse/proliferation clause), the China sentence one (reuters). Printed **p175** (pp174–187 costed
  section): the €790bn figure and the half-hearted sentence carry **no** URI annotation — the page’s
  only three links (von der Leyen / hyperscaler-capex / Manhattan-Project) sit on other sentences.

### Built 2026-09-17 (Slice 5 fixtures) — no judge run, no commit

`claims-slice5.txt` (5 leaves), `argument-slice5.txt` (root → 3 legs: OPENWEIGHT, ACCESS, BUILD),
`sources/MANIFEST-slice5.md` (no `single_source`). Only ONE truth-maker is usably held — the payload of
this slice is that the anti-build / anti-open-weight case rests on citations a fetch pass mostly cannot
reach:

- **w02** → internationalaisafetyreport.org/ fetched HTTP 200 and **held**, but the bare home is a
  PORTAL (lists publications; the misuse/proliferation body is in the linked full-report PDF, not on the
  cited URL). SHA-256 in MANIFEST-slice5.
- **w01** → epoch.ai/data-insights/open-closed-eci-gap **could not be obtained**: epoch.ai is unreachable
  from this session (proxy `502 CONNECT tunnel failed` to curl, `ENOTFOUND` to WebFetch). Not fetched,
  cited-but-not-held → unverifiable. PW hand-fetch.
- **w03** → reuters.com/…/beijing-…-2026-07-07 is an HTTP 401 DataDome wall (771-byte stub kept, not
  held) → unverifiable.
- **w04/w05** → uncited (no URI annotation at p175) → unverifiable.

**Two flags to PW (not fabricated to close):** (1) **€800bn vs €790bn** — this prompt named the cost
“€800bn”, but the scoped pp174–187 text (report.txt:9332, printed p175) states “approximately **€790
billion** over the first three years”; €800bn is the page-73/75 summary / Table-1 rounding. w04 carries
the p175 €790bn wording verbatim. (2) w02’s cited URL is a portal, so the leaf is a citation-precision
case (the report cites a landing page, not the page carrying the claim), pre-registered `absent`.

### Pre-registration — verdict expected per leaf (written before any judge run)

| leaf | node | cite status | expected | read first if NOT as expected |
|------|------|-------------|----------|-------------------------------|
| w01-epoch-openlag | OPENWEIGHT | unreachable (proxy) | unverifiable | on PW’s hand-fetch: “several months” (report) vs the Epoch page’s ~“four months”. Direction is ambiguous (four IS several) — flag only if the source figure contradicts “several” |
| w02-iasr-proliferation | OPENWEIGHT | held (home-portal) | **absent** | a `faithful` here would mean the bare portal page actually carries the government-will-restrict-proliferation prediction — distrust it (citation-precision finding) |
| w03-reuters-china | ACCESS | wall (HTTP 401) | unverifiable | — |
| w04-cost-790bn | BUILD | uncited | unverifiable | — |
| w05-halfhearted | BUILD | uncited | unverifiable | the report’s own normative judgement; grounding it externally would be a category error |

**Totals expected:** 1 `absent` (w02), 4 `unverifiable` (w01, w03, w04, w05), 0 `faithful`. A `faithful`
on any leaf is the result to distrust; the slice’s payload is precisely that this case does not ground on
a fetch pass.

### Judge command — N=3 Sonnet against `sources/cited/` (blocked on the three Slice-3 prerequisites)

Same three gaps as Slice 3/4 (PLAN.md §Slice 3 “Run prerequisites”). Until they close every leaf renders
`unverifiable`. PW runs the command (never call the model API from a session). Flags BEFORE the positional
claims file:

```sh
# run from examples/tai-europe-2026/
source ../../setkey.sh && ../../assay -backend anthropic -model claude-sonnet-4-6 \
  -retrieve bm25 -n 3 \
  -manifest sources/MANIFEST-slice5.md \
  -source sources/cited \
  -argument argument-slice5.txt -refs claims-slice5.txt \
  -chain-dir evidence/2026-09-17-slice5/sonnet-cited \
  claims-slice5.txt
```

## Out of scope (explicit)

Everything not in the five slices above, including: the other 8 objectives-worth of Part A/B body and
all §-recommendation decomposition beyond the 11 objective pairs; the robotics
(pp197–204) deep dive; the rest of the costed European-frontier-AI-project hypothetical (pp174–187)
beyond Slice 5’s €790bn-cost and half-hearted leaves; the
attribution-trap contributor test (H.3); the AI-provenance re-verification pass over all 795 links
(H.4) and the citation-loop provenance test (H.5) except where slice-3 items 16–17 touch it; Figure 2
interested-party grounding (H.6) beyond holding the link; the acknowledge-then-dismiss inference step
(H.7) beyond holding item 5; and any model-calling / `-review` / `-edge` run. No claims files,
`argument.txt`, decomposition, or run sequence are authored here — those are the next plan's work.
