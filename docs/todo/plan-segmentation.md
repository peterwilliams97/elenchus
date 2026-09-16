# plan-segmentation.md — the retrieval segmentation seam

Plan of record for the cross-cut behind roadmap items 2, 6, 7. Assessed 2026-09-12.

The cross-cut those items share is narrower than "`retrieve.go` only knows hearings." Everything
downstream of a `retrieve.Passage` is already corpus-agnostic (`Index`, `build`, `bm25Scores`,
`Search`, `tokenize`, `Format`, `ByIDs`). Three of the four parsers are already generic prose
splitters — `submissionPassages`, `qonPassages`, and `reportPassages` (`retrieve.go:202`) tag
`Source` and leave `Role`/`Context` empty; only `splitFile` (`retrieve.go:302`) is Hansard-specific
(speaker turns, `parseRoster`, `classify`, Q/A `Context`). Role already degrades cleanly: `Format`
(`retrieve.go:533`) branches on `Source`, so role-less passages render fine and the judge never sees
a missing role.

The one coupled point is `passagesForFile` (`retrieve.go:181`): it picks a parser by hardcoded path
substrings (`/submissions/`, `/qon/`, `report.txt`, else Hansard). That implicit switch is the only
thing you edit to add a corpus type.

## The seam (item 7 — do first)

Replace the substring switch with an explicit ordered registry:

```go
type Segmenter interface {
    Match(path string) bool                           // does this segmenter own the file?
    Split(path string, data []byte) ([]Passage, error)
}
```

`passagesForFile` reads the file once, walks a registered `[]Segmenter` (first `Match` wins), Hansard
last as default. The four existing parsers become four `Segmenter` values; registration order =
today's switch order. `Index`/`Search`/`Format` are untouched. This isolates the Hansard
role/context logic inside one segmenter rather than generalising it.

It is a **behaviour-altering-nothing change**, so (per `../elenchus_material/` master plan) the
refuter is near-free: golden-test the vic-lceic corpus — capture the current `[]Passage`, refactor,
assert byte-identical output. The old code is the test.

## After the seam

- **Golden refuter pins gitignored corpora — make it portable.** `TestSegmenterGoldenUnchanged`
  (`internal/retrieve/segment_test.go`), built with the seam, hashes each existing corpus's passage
  set against a constant to prove the refactor is byte-identical. The corpora it reads are gitignored
  (`.gitignore:73`, `:77` — the don't-commit-copyrighted-docs rule), so the constants are machine- and
  disk-specific: a corpus re-fetched or re-extracted anywhere drifts the hash with no code change, and
  the failure reads as "the seam perturbed this corpus" when it means "this machine's corpus differs
  from the capture machine's". It already bit once — dora's on-disk `report.txt` had drifted
  (`85573b…` → `69a49e…`) while `reportPassages`/`paraSplit` were provably untouched by the seam; the
  dora constant was bumped 2026-09-16 (PW-approved) rather than treated as a regression, and the other
  seven corpora still matched. Fix: either compare pre-seam vs post-seam on the same inputs (the git
  diff already shows the report parsers unchanged, so the byte-identical claim is a diff, not a
  hash), or hash a small *committed* fixture rather than the gitignored corpus. Decide
  committed-fixture vs same-machine-diff, then rework the test.
- **Item 6 (other reports)** — closer to done than first stated: a plain report already flows through
  `reportPassages` via the `report.txt` suffix. The real blocker is the page-number assumption —
  `reportPassages` maps `printed page = form-feed index + 1`, valid only "for a report whose PDF has
  no front-matter offset" (`retrieve.go:199`). A report with a cover/TOC (likely
  `examples/quocirca-2026/`) makes the `#page=N` links in the `-review` site (`spec/SERVE.md`) off by
  the offset. Fix: make the page offset a segmenter parameter (or read it from the manifest). Plus
  ordinary prep — a `MANIFEST.md` and decomposed claims for the new report (real inputs, never
  synthesised).
- **Item 2 (critique of code)** — the seam gives it a home, not the work: a code corpus needs a
  segmenter whose passages are Go declarations (`go/ast`, one passage per top-level decl). Genuinely
  new — a different producer (code, not hearings), so a different partition with its own segmenter.
- **Retrieve the whole section by id for id-keyed documents** — a retrieval-strategy lesson from the
  `tai-europe-2026` slice-1 faithfulness run (`examples/tai-europe-2026/PLAN.md` § Adjudicated 16
  Sept (PW)). When the source is keyed by the same ids as the claims — there, each Part A objective
  claim is checked against its own-id Part B implementation — BM25 top-k over the whole source surfaced
  each objective's "WHY IT MATTERS" prose but not its ACTION blocks, and PW's read overturned four
  `faithful`-should-have-been verdicts (IO-3, IO-4, IO-5, O1.2) to false non-faithful. Candidacy was
  not the problem (the whole 128-pp Part B was in the pool); top-k *selection* was. Fix: for an
  id-keyed corpus, a retrieve route that returns the whole section for the claim's id (an oracle-by-id
  segmenter, or a manifest-declared id→page-range map feeding `-retrieve oracle`) rather than BM25
  top-k. Not a `retrieve.go` change to make blind — it needs a corpus that declares the id→section
  mapping in its `MANIFEST.md`, which `tai-europe-2026` already does (`sections:` table).

## Sequence

7 (seam + golden refuter) → 6 (report-offset parameter + quocirca manifest/claims) → 2 (Go-decl
segmenter). One refactor unblocks the routing for all three; 6 and 2 then differ only in which
segmenter they add. If the seam is written up as a contract, it belongs in a new `spec/RETRIEVE.md`
(there is none today — retrieval is specified inside `spec/CLI.md` §Retrieval).
