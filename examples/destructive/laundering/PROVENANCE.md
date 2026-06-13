# PROVENANCE — laundering fixture source

This is the one probe in `examples/destructive/` whose ground truth is **not** a construction. The
laundering demonstration requires a claim that is *simultaneously* (a) faithfully attributable to a
real source, (b) well-formed and falsifiable, and (c) refutable by real-world evidence. Constructed
text cannot satisfy (a) and (c) at once — so per `CLAUDE.md`'s no-fabrication rule, the source here is
a **real public quotation**, fetched and verified, never invented.

## The source

- **Speaker:** Steve Ballmer (then CEO, Microsoft)
- **Occasion:** USA TODAY CEO Forum (sixth annual, with the University of Washington Business School),
  interviewed by USA TODAY's David Lieberman
- **Publication:** USA TODAY, **April 30, 2007**
- **Original article URL:** `https://www.usatoday.com/money/companies/management/2007-04-29-ballmer-ceo-forum_N.htm`
- **Retrieved / verified:** 2026-06-10

## Verifiable snippet (verbatim — this *is* the source)

> "There's no chance that the iPhone is going to get any significant market share. No chance. It's a
> $500 subsidized item. They may make a lot of money. But if you actually take a look at the 1.3
> billion phones that get sold, I'd prefer to have our software in 60% or 70% or 80% of them, than I
> would to have 2% or 3%, which is what Apple might get."

## Cross-source verification

The original USA TODAY article and the Internet Archive were not directly fetchable from the build
environment on the retrieval date. The passage above is nonetheless one of the most-documented quotes
in technology journalism and was confirmed **verbatim and consistent** across multiple independent
reproductions on 2026-06-10:

- MacDailyNews, "Microsoft's Ballmer: 'No chance Apple iPhone is going to get any significant market
  share'" (2007-04-30)
- libquotes.com / brainyquote.com / azquotes.com Ballmer quote pages
- Cult of Mac, "Today in Apple history: Steve Ballmer freaks out and stomps an iPhone"

The contiguous full passage (the "$500 subsidized … 2% or 3%" portion) was reproduced identically in
the initial source search and the MacDailyNews article. Only Ballmer's verbatim words are used; the
interviewer's exact phrasing was not verifiable and is therefore **not** reproduced.

## The refutation (the truth-maker for the grounding column)

"Significant market share" was decisively achieved. By 2013–2015 the iPhone held roughly **15–17% of
global smartphone shipments** (Statista / StatCounter) and captured the large majority of the
industry's profit — orders of magnitude beyond the "2% or 3%" Ballmer projected. The `-evidence`
column reaches this by web search; it is not an armchair deduction.

## How to run this probe (recreating the gitignored source)

The fetched source lives at `sources/ballmer_usatoday_2007.txt`, which is **gitignored** (real
third-party text: analysis, not redistribution — `CLAUDE.md`). After cloning you will not have it.
Recreate it from the verbatim snippet above — the whole source is that one public quotation:

```sh
mkdir -p examples/destructive/laundering/sources
cat > examples/destructive/laundering/sources/ballmer_usatoday_2007.txt <<'SRC'
USA TODAY CEO Forum (sixth annual, with the University of Washington Business School).
Microsoft CEO Steve Ballmer, interviewed by USA TODAY's David Lieberman. Published in
USA TODAY, April 30, 2007.

Ballmer, on Apple's iPhone:

"There's no chance that the iPhone is going to get any significant market share. No chance.
It's a $500 subsidized item. They may make a lot of money. But if you actually take a look
at the 1.3 billion phones that get sold, I'd prefer to have our software in 60% or 70% or
80% of them, than I would to have 2% or 3%, which is what Apple might get."
SRC
```

Then run the audit (see `../run.sh laundering` for the calibrated N-run version):

```sh
./crossexam -audit -source examples/destructive/laundering/sources/ballmer_usatoday_2007.txt \
        examples/destructive/laundering/summary.txt -md
```
