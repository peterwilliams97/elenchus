# Phase 2 Fixture Collection Report

**Date:** 2026-05-31
**Status:** HALTED — No real sources obtained
**Reason:** Environment constraints + CLAUDE.md discipline

## Attempted Sources

I attempted to fetch REAL documents for the 7 test functions according to the specification:

| Function                             | Source Attempted        | Result | Why Failed |
|--------------------------------------|-------------------------|--------|-----------|
| legal                                | OLC opinions (justice.gov)   | ✗ | PDFs require extraction; HTML pages 404 or JavaScript-heavy |
| accounting                           | SEC EDGAR 10-K MD&A          | ✗ | Requires PDF parsing or complex HTML extraction |
| sales                                | Pitch decks / press releases | ✗ | Image-heavy or behind JS rendering |
| marketing                            | BusinessWire / newsroom press releases | ✗ | Require JS rendering or scraping |
| pm                                   | PEP 585 / PEP 8 from peps.python.org or raw GitHub | ✗ | Fetch returns redirects or 302 responses |
| engineering                          | Cloudflare blog postmortem   | ✗ | JavaScript-heavy extraction; yields noise not clean text |
| contract                             | SEC EDGAR 8-K + Exhibit 10   | ✗ | PDFs and complex HTML structure |

## Discipline Applied

Per CLAUDE.md "Hard rule: never fabricate inputs":

> NEVER synthesize a fixture, test input, sample document, dataset, or "example" of a real artifact.
> If a real one is required and cannot be fetched or obtained, STOP and report
> "could not obtain real <X>" for that item. Do not substitute a fabricated stand-in.

**This is the correct choice.** A populated results table built on fabricated inputs would:
1. Invert the tool's entire purpose (a claim validator running on fake claims)
2. Make calibration signals invalid
3. Hide real bugs that would surface against actual messy prose

## What Was Completed

**Phase 1 (COMPLETE):**
- Added `-max-claims N` flag to bound evidence grounding
- Seam tests verify the cap works and are mutation-tested
- All tests pass; build succeeds

**Phase 2 (INCOMPLETE):**
- No real sources successfully fetched
- No fixtures created
- No assay runs possible

## Next Steps

To complete Phase 2, either:
1. Manually provide real documents (copy/paste from sources)
2. Use a different environment with better web access or tooling (lynx, wkhtmltohtml, etc.)
3. Accept that the fixture test is not possible without real source material
4. Use different test strategies (e.g., unit tests of individual parsing/grounding logic)

## Recommendation

Report Phase 1 as complete and ship the `-max-claims` feature. Phase 2's fixture collection is
blocked by environment constraints, not code issues. The evidence-cap feature is valuable
independently of the calibration exercise.
