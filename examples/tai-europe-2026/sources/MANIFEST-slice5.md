# Corpus manifest — Slice 5, claim 2 (cited-source faithfulness: why not build / why not open-weight)

Every source document the Slice-5 corpus holds, in the report’s own naming. The backticked id is
canonical: a claim’s `cites` column names the document a claim rests on with these ids, and a
faithfulness run marks a claim `unverifiable` (no model call) when a cited id is not listed here.

This is the **Slice-5** manifest. Slice 5, claim 2 (PLAN.md §Slice 5) is the “why not build / why not
open-weight” step of O1.1 (“Secure ongoing access to frontier AI systems”, Part A pp50–53): the five
sourced claims by which the report argues Europe should not count on its own frontier build or on
open-weight models, and so should take the Leverage Approach. One `evidence`-route leaf per row in
`../claims-slice5.txt`. It grounds each leaf against the report’s OWN external citation — a fetched web
page — not another part of the report. `MANIFEST.md` stays the Slice-1/2 manifest, `MANIFEST-slice3.md`
the Slice-3 one, `MANIFEST-slice4.md` the Slice-4 one; this file is loaded only for a Slice-5 run.

## Roles for Slice 5 (faithfulness config)

The report is the TARGET; the cited web pages are the SOURCES (as in Slice 3/4). This inverts Slice 1/2’s
held-report arrangement: the report is the thing under test and is NOT held, while each external
truth-maker it cites IS (when it could be fetched).

- **TARGET (assayed, not held):** the report’s sentences (`../claims-slice5.txt`, verbatim from
  `report.txt`). Not in this manifest.
- **SOURCE (held):** the fetched cited pages held under `sources/cited/` as `4x-2N-<host>.txt` (stdlib
  HTMLParser strip of the raw `.html`). The check is cite-scoped faithfulness — each leaf is judged
  against ONLY its cited page’s passages (spec/TREE.md § Cite-scoped retrieval).

**No `single_source` line.** Claim and truth-maker are two DIFFERENT documents (report vs. cited page),
so the single-source judge rules are off: a leaf `absent` here means “the held cited page does not carry
this claim”, a genuine grounding miss against a second document — not “said once, not repeated”.
`LoadSingleSource` returns false for this file.

Structurally weakest verdict: `evidence` grounding proves only that the cited page was retrieved and
carries the claim, never that the report’s inferential use of it is sound (CLAUDE.md § axis boundary).
Read `faithful` here as “the cited page says it”, nothing larger.

## Groundability — only ONE of the five truth-makers is usably held

The payload of this slice is that the anti-build / anti-open-weight case rests on citations a fetch pass
mostly cannot ground. Four of five leaves render `unverifiable` this run; only w02 is held (and its held
page is a portal, so it is pre-registered `absent`, not `faithful`).

- **w01 open-weight lag → epoch.ai/data-insights/open-closed-eci-gap — COULD NOT OBTAIN.** epoch.ai is
  unreachable from this session: the filtering proxy returns `502 CONNECT tunnel failed` to `curl` and
  `ENOTFOUND` to WebFetch. Not fetched, not held; cited-but-not-held → `unverifiable`. **PW hand-fetch.**
  Pre-registration: the page reportedly puts the open-vs-closed ECI gap at ~“four months”; the report
  says “several months”. Direction is ambiguous (four IS several) — flag only if the source figure
  actually contradicts “several”.
- **w03 China restriction → reuters.com/…/beijing-…-2026-07-07 — WALL.** HTTP 401, DataDome bot-wall
  (“Please enable JS and disable any ad blocker”). Fetched as a 771-byte stub, not usable; left out of
  the held set → `unverifiable`.
- **w04 €790bn cost, w05 half-hearted → UNCITED.** No URI annotation on the PDF at these loci (printed
  p175; von der Leyen, hyperscaler-capex and Manhattan-Project are the only three links on the page, and
  none overlaps the cost figure or the half-hearted sentence — verified with pypdf by rect y). They cite
  an `uncited/…` id no held document answers to → `unverifiable`. w05 is the report’s own normative
  judgement; expecting it to ground on an external source would be a category error.

**€800bn vs €790bn (flagged to PW).** This slice was prompted with “the €800bn three-year cost”. The
scoped pp174–187 text states “approximately **€790 billion** over the first three years” (report.txt:9332,
printed p175). €800bn is the page-73/75 summary / Table-1 rounding. w04 carries the p175 €790bn wording
verbatim; do not “correct” it to €800bn.

## Held cited ids (registered for `LoadHeld`)

Only the one capture that fetched HTTP 200 and is usable as text (a backticked `- ` bullet so
`manifest.LoadHeld` sees it). `fig`: `home-portal` = the cited URL is a landing/portal page that lists
publications but carries none of the claim’s body text (the substance is in the linked full-report PDF),
so the leaf is read as `absent` = “checked, not located on the cited page”, a citation-precision miss.

- `cited/4x-22-internationalaisafetyreport.org.txt` — International AI Safety Report home/portal (proliferation/misuse, p51) — home-portal
- `cited/4x-21-epoch.ai.txt` — Epoch open-vs-closed ECI gap (open-weight lag, p51) — present

## Not held — named so their absence is `unverifiable`, never silently `absent`

These entries are deliberately **not** `- ` backticked bullets: `manifest.LoadHeld`
(`internal/manifest/manifest.go:141`) registers every `^- ` + "`id`" line, so a not-held id must be
written as prose (as in `MANIFEST-slice4.md` § Not held), or it would be silently registered as held.

w03 cites `cited/4x-20-reuters.com.txt` — Reuters “Beijing looking at curbing overseas access to China’s
top AI models” (China restriction, p51), a **wall** (HTTP 401 DataDome); its 771-byte stub is kept for
provenance but left out of the held set. w01 cites `cited/4x-21-epoch.ai.txt` — the Epoch open-vs-closed
ECI-gap page (open-weight lag, p51), **unreachable** this session (proxy 502 / ENOTFOUND), not fetched at
all. w04/w05 cite `uncited/p175-cost-790bn` and `uncited/p175-half-hearted` — no URI annotation at those
loci. None of the four is in the held set, so each renders `unverifiable` (no model call: “cited
document(s) not in corpus”) rather than grounding on the pooled corpus.

| leaf | anchor sentence (verbatim, `report.txt` page) | URL (PDF URI annotation) | file | fig |
|------|-----------------------------------------------|--------------------------|------|-----|
| w01 | “Counting on open-weight models is also a risky bet: they lag behind the performance of frontier models by several months …” (p51) | https://epoch.ai/data-insights/open-closed-eci-gap | — (unreachable) | **unreachable** — proxy 502 / ENOTFOUND |
| w02 | “… governments are likely to restrict their proliferation as the potential to misuse them increases.” (p51) | https://internationalaisafetyreport.org/ | 4x-22-internationalaisafetyreport.org.html | home-portal (claim in linked PDF, not on cited URL) |
| w03 | “China is reportedly considering similar restrictions on its most advanced models.” (p51) | https://www.reuters.com/world/beijing-is-looking-curbing-overseas-access-chinas-top-ai-models-sources-say-2026-07-07/ | 4x-20-reuters.com.html | **wall** — HTTP 401 DataDome |
| w04 | “… a serious attempt … would likely cost Europe approximately €790 billion over the first three years of such a project …” (p175) | — (no URI annotation) | — | uncited |
| w05 | “A half-hearted attempt at a European frontier AI project … would lead Europe down the worst of all possible paths …” (p175) | — (no URI annotation) | — | uncited |

**SHA-256 (over held / fetched `.html`, = raw response bytes; fetched 2026-09-17 with `curl`, Chrome UA):**

```
b2c292fa72a03ff3875d49cfe1d91844891e35214cf53bff1bbb2c52bd9276dd  4x-22-internationalaisafetyreport.org.html   (HTTP 200, HELD)
6ecb94a5de877979ca650f75b541e1261f29ecad6363cb29a62a606989855a34  4x-20-reuters.com.html                       (HTTP 401 wall, not held)
a366f8c4f78f0544b591ce6683c53e4ac86cd521aa3babf4a93e9fc1b44a2959  4x-21-epoch.ai.html                          (HTTP 200, HELD, hand-fetched 17 Sept via --resolve)
```

## Report-link resolution

`report_page_offset` 0 (VERIFIED Slice 1): printed page N sits on PDF page N. The three URLs above were
read off the PDF’s own URI annotation with pypdf (PDF pages 51 and 175, by rect y-coordinate against the
text line — never guessed). The cost (p175) and half-hearted (p175) loci carry no URI annotation.

report_page_offset: 0   # VERIFIED 2026-09-15 (Slice 1); printed page N == PDF page N.
