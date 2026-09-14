# Review site

Two pages, self-contained. `index.html` is the argument tree — the thesis, the counts, and
every claim with its faithfulness verdict. `review.html` is that same tree in a two-pane
layout: the tree on the right, the source PDF on the left, so a claim's page link lands on the page
it rests on.

## Open it

1. Unzip, then open `index.html` in Chrome or Edge.
2. If `review.html`'s left pane is blank, the browser is refusing the `file://` PDF —
   serve the folder over HTTP instead, then open `http://127.0.0.1:8080/review.html`:

       python3 -m http.server 8080     # run inside this folder

   or, with the assay binary from the parent folder:

       assay serve site
