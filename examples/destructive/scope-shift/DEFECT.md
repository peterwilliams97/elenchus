# DEFECT — scope-shift (target axis: **Equivocation**)

> This is an **openly-constructed adversarial probe**, not a real artifact. The construction
> *is* the ground truth: the defect below was engineered in deliberately. It does not stand in
> for anyone's real claim. (See `../README.md` for why that is not a no-fabrication violation.)

## The engineered defect

A **scope-shift**: the *subject* a claim is predicated of silently widens from the narrow subject it
was measured on to the broad subject it is asserted of — with the predicate held fixed.

- **Narrow (where the evidence is):** the **Oxford Street flagship** — one store — posted a median
  checkout time of 1 minute 50 seconds last week. True, concrete, checkable.
- **Broad (where the claim lands):** "**The company** is fast: **our** median checkout time is under
  two minutes." The same predicate ("median checkout time under two minutes") is now asserted of the
  whole company on evidence about one store.

The possessive *our* does the widening work: it reads as company-wide while resting on a single-store
measurement. Pin the subject to the flagship and the claim is true and well-formed; pin it to the
company and it is unsupported. The passage never fixes which — so no single claim is falsifiable *as
stated*. That is the defect.

## The predicate is deliberately frozen

"Fast" means one thing throughout — **median checkout time under two minutes** — stated in sentence 1
and repeated verbatim in sentence 3. Nothing about the *predicate* moves: not its sense (it is not a
word retreating to a weaker meaning) and not its strength (sentence 3 does **not** upgrade the median
to a universal "no shopper waits longer" — that would be a second, unengineered defect). The **only**
motion in the passage is the subject: *flagship → company*. A grade of `hollow`/`partial` is therefore
attributable to the scope-shift and nothing else.

## How this differs from motte-and-bailey (the sibling probe)

Both live on the Equivocation axis; they are opposite motions, and the difference is the point:

- **motte-and-bailey** is a **lexical/predicate** fault: a *term's sense* retreats strong → trivial,
  and it happens **under challenge** — the classic form anticipates an interlocutor ("once you accept
  X, you've already granted Y"). Direction: a broad claim *defended* by narrowing.
- **scope-shift** is a **referential** fault: the *subject of predication* widens narrow → broad, and
  it needs **no challenge and no interlocutor** — the defect is fully present in a monologue. Direction:
  narrow evidence *expanded* to a broad claim.

Stated as a trichotomy — the fault here is **referential** (what the claim is *about* changes), not
**inferential** (a missing load-bearing premise — that is `../hidden-premise/`) and not **lexical** (a
word's *meaning* changes — that is `../motte-and-bailey/`).

## Axes it is deliberately clean on

- **Evidence** — the one number (flagship median 1:50) is real, uncited-magnitude-free, and not in
  dispute; it is honest scaffolding.
- **Base rate / magnitude** — no quantity is compared against a gamed class. (One store standing for a
  company *is* the scope-shift, viewed from the sampling angle — not a separate defect; see adjacent
  catches.)
- **Causality vs correlation** — no cause is inferred from an association.
- **Falsifiability (vacuity sense)** — each subject-pinned claim *is* observable and checkable; the
  fault is an unfixed subject, not an untestable predicate.

## Most likely *adjacent* catches (acceptable signal — but read the discriminator)

- **Hidden premise** — the unstated "the flagship represents the company" is the scope-shift seen from
  the premise angle.
- **Base rate / reference class** — "one store generalized to all stores" is the same move seen from
  the sampling angle.
- **Counterexample** — a single slow branch store refutes "the company is fast."

A finding on any of these counts as the probe *working* **only if the stated reasoning names the
subject substitution** (measured of the flagship, asserted of the company). A `hollow` reached by
"unsupported / no evidence" while treating the subject as fixed is *right-answer-wrong-reason*, not a
catch — see `EXPECTED.md`'s pre-registered discriminator.

## How this probe was constructed

Took the pure part-for-whole substitution and bound it to a single defined metric (median checkout
time) so the predicate could be frozen and repeated verbatim across the widening, leaving the subject
as the only moving part. Earnings-call phrasing (the possessive "our") was chosen because that is how
the move actually occurs in the wild — a single flagship's number narrated as the company's. Every
flourish that would trip another axis (a strengthened universal claim, a gamed comparison, a causal
story) was deliberately excluded to isolate the signal.
