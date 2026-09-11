# `assay -review` + `assay serve` — spec

## Why

`review.html` is a two-pane page: a left `<iframe>` PDF viewer and, on the right, the argument tree
with every report section and quote provenance rewritten as a link that drives the iframe. The links
open PDFs at a page (`sources/report.pdf?p=53#page=53`). Opened as a `file://` path the browser
resolves them, but PDF viewers don't reliably honour the `#page=N` fragment over `file://` and some
refuse cross-origin `file://` iframes. Serving the page over HTTP gives it a real origin, so the
fragment lands on the right page and the iframe loads.

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
        review.html       # two-pane PDF-linked argument page
        argument.html     # the same argument tree, no iframe
        sources/
          report.pdf
          hearings/…/*.pdf
          submissions/*.pdf
          qon/*.pdf

Every href in `review.html` is **site-relative** — `sources/report.pdf?p=53#page=53`, never
`../sources/…` — and each links to a PDF copied into `site/sources/` at the same relative path. Four
link classes are copied: the report, and each linked hearing, submission and qon document. A PDF that
cannot be read is warned and counted `missing`, never synthesized; a non-zero `missing` means a link
class is broken.

Build command (run from `current/`, the render dir; flags **before** the positional claims file):

    assay -from current.faithfulness.jsonl,current-2026-09-10.faithfulness.jsonl \
      -argument ../argument.txt -manifest ../sources/MANIFEST.md -refs ../claims-machine.txt \
      -review \
      claims.txt

It reports one line: link tally per class plus `pdfs copied=N missing=0`.

`argument.html` carries no PDF links (it is the plain tree), so it needs no copies; it ships in the
site as the iframe-free reading of the same page.

## `assay serve` — serving the site

    assay serve [-port N] [-no-open] <site-dir>

- `<site-dir>` is the site `-review` built, e.g. `examples/vic-lceic/site`. It is served whole, so
  `review.html` answers at `/review.html` and each `sources/…` link at `/sources/…`.
- `-port` — listen port on `127.0.0.1`, default `8080`.
- `-no-open` — don't open the browser (still prints the URL, still serves).

### Behaviour

1. Bind `127.0.0.1:<port>` (loopback only — never `0.0.0.0`). If the port is taken, report it and
   exit non-zero.
2. Serve `<site-dir>` as static files (`http.FileServer`, rooted there — no path escapes it).
3. Print `http://127.0.0.1:<port>/review.html`. Warn to stderr, non-fatally, if `review.html` is
   absent under `<site-dir>` — the server still runs so `argument.html` works.
4. Unless `-no-open`, open that URL in the default browser: `open` (macOS), `xdg-open` (Linux),
   `cmd /c start` (Windows). A missing opener is a warning, not a failure.
5. Block, serving, until interrupted.

### What it does not do

- No write access, no upload, no directory outside `<site-dir>`.
- No `?p=N#page=N` rewriting — the fragment is harmless over HTTP and left as-is.

## Refuters

- **Every href in `site/review.html` resolves to a file under `site/`.** `TestWriteSiteHrefsResolve`
  drives `writeSite` with a fake sources tree (one PDF per link class) and asserts each `sources/`
  href lands on a copied file and both pages sit at the site root — an unresolved href is exactly the
  break a viewer would hit as a 404.
- **The site serves whole over HTTP.** `TestServeReviewLinksResolve` writes a site holding
  `review.html` (with a `sources/x.pdf` link) and `sources/x.pdf`, starts the handler, and checks
  `GET /review.html` (200, body carries the link) and `GET /sources/x.pdf` (200, byte-identical). A
  404 on either fails — the link class breaking.
