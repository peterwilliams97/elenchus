# `assay -review` + `assay serve` — spec

## Why

The argument tree (`spec/ARGUMENT.md`) is read as one page — `index.html` — a proportional-font tree
of cards: the report's thesis, then one card per recommendation, each opening to its findings, its
claims, and the quotes those claims rest on. `review.html` is that same page in a two-pane layout: a
left `<iframe>` PDF viewer, the tree on the right, with every report section and quote provenance
rewritten as a link that drives the iframe. The links open PDFs at a page
(`sources/report.pdf?p=53#page=53`). Opened as a `file://` path the browser resolves them, but PDF
viewers don't reliably honour the `#page=N` fragment over `file://` and some refuse cross-origin
`file://` iframes. Serving the page over HTTP gives it a real origin, so the fragment lands on the
right page and the iframe loads.

For that to be portable, the page and everything it links must travel together. `-review` builds a
**self-contained site**: one directory holding the pages and a copy of every PDF they open, with
links relative to the site root — so it serves, zips, or moves as a unit, and `assay serve` needs
nothing outside it.

## `assay -review` — building the site

`-review` runs alongside `-argument` and `-manifest` on a render pass (`assay -from …`). It builds
the site under the **example dir** (the parent of the sources tree the manifest lives in):

    examples/vic-lceic/
      sources/            # the on-disk corpus (gitignored; PDFs read from here)
      site/               # ← built by -review (gitignored build output)
        index.html        # the argument-tree page (below)
        review.html       # index.html's tree in the right pane, the source PDF on the left
        favicon.svg       # the site mark (an argument-tree glyph), linked from both pages
        README.md         # how to open and serve the site (below)
        sources/
          report.pdf
          hearings/…/*.pdf
          submissions/*.pdf
          qon/*.pdf

Build command (run from `current/`, the render dir; flags **before** the positional claims file):

    assay -from current.faithfulness.jsonl,current-2026-09-10.faithfulness.jsonl \
      -argument ../argument.txt -manifest ../sources/MANIFEST.md -refs ../claims-machine.txt \
      -review \
      claims.txt

It reports one line: link tally per class plus `pdfs copied=N missing=0`.

`README.md` at the site root is the reader's manual — the two pages named in one line each, how to
open them (unzip, open `index.html` in Chrome or Edge), and how to serve them over HTTP
(`python3 -m http.server 8080`, or `assay serve site`) when the browser blocks the `file://` PDF and
`review.html`'s left pane comes up blank. It ships inside the site so an unzipped copy explains
itself with no reference back to this repo.

`-zip` (with `-review`) also writes `site.zip` beside `site/`: the whole built site archived under a
top-level `site/` entry, so unzipping restores a `site/` folder `assay serve site` can serve. It
reports `zip: <path> — N files`; an archive with no files is an error, not a silent empty download.

### `index.html` — the argument-tree page

One page replaces the earlier landing page and separate tree. It is plain HTML in a **proportional**
font, laid out top to bottom:

1. the report **title** (the argument file's `# title:` line);
2. two sentences — **what the page checks** and **which sources are held** (the manifest's held set,
   counted by category);
3. the **thesis** in a card, with the recommendation **tally** and the descriptive-base sentence
   beneath it (the root block of `spec/ARGUMENT.md`, split into a card rather than a text block);
4. one **card per root child** — the recommendations `R1`…`R11` in the argument file's order, then
   the descriptive `base` card. Each card's summary carries a **badge**, the node `id` (with a
   trailing `?` where the report does not establish the edge to its parent), the **proposition** in
   the report's words, and — when the node does not `hold` — its **deciding child** in small text.

Each card is a `<details>`. Opening a recommendation reveals its findings as nested cards with the
same badge; opening a finding reveals its claims, each a card whose badge is the claim's verdict and
whose summary carries the `k/N` agreement; opening a claim reveals the quotes it rests on, each with
its provenance. Nesting matches the argument file: root → recommendation → finding → claim → quotes.

**Badges are the one thing on the page with colour, and the colour is redundant with the word** —
every badge shows its class as text, so a colour-blind reader loses nothing. Six classes, one
colour-blind-safe hue each (Okabe–Ito):

| badge | on | colour |
|---|---|---|
| `holds` | a recommendation/finding whose load-bearing children all hold; a `faithful`+`settled` claim | green |
| `weakened` | a node standing on narrower ground; a `partial`/`overstated` claim, or a same-side `wobble` | orange |
| `open` | a node the report doesn't say what it rests on, or resting on a contested/unverifiable claim | blue |
| `fails` | a node a load-bearing child collapsed; a `contradicted`/`unsupported`/`absent` claim | vermillion |
| `opinion` | a node resting only on Committee value judgements; an `opinion` claim (never judged) | grey |
| `contested` | a claim leaf the runs could not settle — `contested`, or a tied `split` | reddish purple |

An internal node's badge is its **derived** judgement (`spec/ARGUMENT.md` § Internal judgement); a
claim's badge is its **pooled verdict**, coloured on the same severity scale, with the `contested`
hue reserved for leaves the runs split on. No judgement is authored — the tree computes every badge.

Cards are **closed by default**. One affordance opens them: an *Expand all* button, and an
`?open=all` query that opens every card on load — a single line of JavaScript wires both, and the
page is otherwise script-free.

`favicon.svg` is a plain mark — an argument-tree glyph, one node above three, in the page's text
colour, no brand and no text. Both `index.html` and `review.html` link it.

### Adjudications — the human verdict beside the machine's

A corpus may carry `examples/<x>/adjudications.txt`: a human's own verdicts on leaves they have read,
so the page can show where the judge agreed with a person and where it did not. It is read from the
example dir (the parent of the sources tree) whenever the argument page is built, and is optional — a
corpus without the file renders exactly as before.

One leaf per line, six `|`-delimited fields in order — no field may contain a `|`:

    id | verdict | quote | page | initials | reason
    E1 | faithful | Two-thirds (67%) of organisations report at least one print-related breach | p4 | PW | summary drops the country list; more specific is not overstated

`#` comment lines and blank lines are skipped; a present line without exactly six fields is an error,
not a silent mis-parse. `verdict` is the human's call in the same vocabulary as a leaf verdict
(`faithful`, `partial`, `overstated`, `contradicted`, …); `quote` is the verbatim span the human
grounded it on; `page` is the report page (`p4` or `4`, empty for none); `initials` name the
adjudicator; `reason` is one line.

The page uses the file two ways:

- **On the leaf.** An adjudicated leaf's card shows the human verdict and initials beside the machine
  badge — the reader sees both calls without leaving the leaf. No agreement colour on the leaf; the
  count lives in the index.
- **In the index.** A line beneath the `what` sentence reports `N leaves adjudicated, judge agreed on
  M`, then lists each disagreement — `<id> — human <hv>, judge <mv>: <reason>`. Agreement is exact
  string equality between the human verdict and the machine's **pooled faithfulness verdict** (the
  leaf's `Faith`, after the single-source `absent`→`uncorroborated` remap), so an evaluative leaf the
  tree badges `opinion` is still measured against the judge's underlying faithfulness call. A leaf the
  tree carries no verdict for counts as a disagreement (`(no verdict)`), never a silent agreement.

The page authors no verdict: the machine verdict is the judge's pooled `Faith`, the human verdict is a
line in the file, and agreement is only the two strings compared.

### `review.html` — the two-pane reading

`review.html` is `index.html`'s tree rewritten for the two-pane layout: the tree fills the right
pane, a PDF `<iframe>` the left. Every `report: §…` and every quote provenance becomes an
`<a target="doc">` that loads the PDF at the page. Its cards are **opened**, so the provenance links
are visible without a click — the two-pane page exists to click them. Each page link carries a
distinct `?p=<page>` query before the `#page` fragment so the viewer re-fetches on every click; a
hearing link carries no page anchor (the transcript locator is a turn index, not a page). Four link
classes are rewritten and copied into `site/sources/`: the report, and each linked hearing,
submission and qon document. A PDF that cannot be read is warned and counted `missing`, never
synthesized; a non-zero `missing` means a link class is broken.

Every href in `review.html` is **site-relative** — `sources/report.pdf?p=53#page=53`, never
`../sources/…` — and each links to a PDF copied into `site/sources/` at the same relative path.

### Multiple report excerpts

A report may be decomposed from more than one excerpt PDF — the AI Index 2026 corpus holds the coding
section (`report.pdf`, printed pp100–102) and the labor-impact section (`report-productivity.pdf`,
printed pp219–221) as two files. Each claim's report deep-link opens **its own** excerpt, so the
manifest carries a `reports:` list, one entry per excerpt:

    reports:
      - file: report.pdf
        offset: -99
        prefixes: SEC SW SB TB VC
        sections:
          Software: 100
      - file: report-productivity.pdf
        offset: -218
        prefixes: EP
        sections:
          Productivity Trends: 219

Each entry gives the PDF `file:`, its printed→PDF page `offset:`, the claim-id `prefixes:` whose leaves
belong to it (the alphabetic head of the id — `SB6` → `SB`), and its own named-section `sections:`
table. Which excerpt a §-ref opens is decided by the claim id of the leaf card the ref sits in
(`manifest.ReportFor`), so `SB6`'s §-ref opens `report.pdf` at `100 − 99 = 1` while `EP8`'s opens
`report-productivity.pdf` at `219 − 218 = 1`. A claim id no entry claims leaves the ref unlinked
rather than pointed at the wrong PDF; the left pane lands on the first entry. Every excerpt is copied
into `site/sources/` because its links carry it, so the site stays self-contained.

The flat single-report form — a top-level `report_page_offset:` and `sections:` table (`spec/CLI.md`,
`internal/manifest`) — is unchanged: it reads as one `report.pdf` with no `prefixes:`, so it claims
every leaf. A `reports:` list, when present, overrides the flat keys.

## `assay serve` — serving the site

    assay serve [-port N] [-no-open] <site-dir>

- `<site-dir>` is the site `-review` built, e.g. `examples/vic-lceic/site`. It is served whole, so
  `index.html` answers at `/index.html`, `review.html` at `/review.html`, and each `sources/…` link
  at `/sources/…`.
- `-port` — listen port on `127.0.0.1`, default `8080`.
- `-no-open` — don't open the browser (still prints the URL, still serves).

### Behaviour

1. Bind `127.0.0.1:<port>` (loopback only — never `0.0.0.0`). If the port is taken, report it and
   exit non-zero.
2. Serve `<site-dir>` as static files (`http.FileServer`, rooted there — no path escapes it).
3. Print `http://127.0.0.1:<port>/index.html`. Warn to stderr, non-fatally, if `index.html` is
   absent under `<site-dir>` — the server still runs so `review.html` is still reachable directly.
4. Unless `-no-open`, open that URL in the default browser: `open` (macOS), `xdg-open` (Linux),
   `cmd /c start` (Windows). A missing opener is a warning, not a failure.
5. Block, serving, until interrupted.

### What it does not do

- No write access, no upload, no directory outside `<site-dir>`.
- No `?p=N#page=N` rewriting — the fragment is harmless over HTTP and left as-is.

## Refuters

- **The page is a root card first, then the recommendation cards in order, nested to the argument
  file's depth.** `TestArgumentCardPage` (tree package) drives `ArgumentPage` with a thesis, two
  recommendations and a base child, and asserts the thesis card precedes `R1`, `R1` precedes `R2`, a
  claim card sits nested inside a finding card inside a recommendation card, and each node carries its
  derived badge class — the order and nesting a reader relies on to read the case top-down.
- **`site/index.html` is the page verbatim, and its links resolve under `site/`.** `TestWriteSitePage`
  drives `writeSite` and asserts `index.html` is the built page byte-for-byte, `favicon.svg` is
  written and linked, and `review.html` sits beside it — a page rewritten on the way to disk is a page
  no refuter checked.
- **A two-report corpus links each claim to its own excerpt PDF, at that excerpt's page offset.**
  `TestRenderReviewMultiReport` (assay package) drives `renderReview` with `report.pdf` (offset −99,
  prefix `SB`) and `report-productivity.pdf` (offset −218, prefix `EP`) and a page holding one leaf
  card per report, and asserts `SB6`'s §-ref opens `report.pdf` at page 1 and `EP8`'s opens
  `report-productivity.pdf` at page 1 — a leaf routed to the wrong file or page is the break a reader
  clicking the provenance would hit. `TestLoadReportsMulti` (manifest package) pins the `reports:`-list
  parse and `ReportFor`'s prefix routing; `TestLoadReportsSingle` pins the flat single-report fallback.
- **Every href in `site/review.html` resolves to a file under `site/`.** `TestWriteSiteHrefsResolve`
  drives `writeSite` with a fake sources tree (one PDF per link class) and asserts each `sources/`
  href lands on a copied file and both pages sit at the site root — an unresolved href is exactly the
  break a viewer would hit as a 404.
- **The site serves whole over HTTP.** `TestServeReviewLinksResolve` writes a site holding
  `review.html` (with a `sources/x.pdf` link) and `sources/x.pdf`, starts the handler, and checks
  `GET /review.html` (200, body carries the link) and `GET /sources/x.pdf` (200, byte-identical). A
  404 on either fails — the link class breaking.
- **Adjudications agree by exact verdict match, and disagreements keep their reason.**
  `TestAgreeThreeTwo` (adjudicate package) drives `Agree` with three adjudications, two matching the
  machine verdicts and one not, and asserts `N=3, Agreed=2` with the single disagreement carrying the
  human's reason; `TestLoad` pins the `|`-delimited parse, skipped comments, the page field, and that a
  missing file is not an error. `TestArgumentPageAdjudications` (tree package) drives `ArgumentPage`
  with an overlay and asserts the index reports the count and lists the disagreement, and an adjudicated
  leaf shows the human verdict beside the machine badge.
