# EXPECTED — well-formed (positive controls)

The six defect probes under `examples/destructive/` hunt false *passes* — a claim the critic wrongly
lets through as `substantive`. This directory is their complement: three claims that a working
substance critic **must** pass. There is no defect here, so there is no `DEFECT.md` — the target is a
false *negative*, the critic collapsing a sound claim it should certify.

Each fragment is deliberately built with the four properties that leave the substance axes nothing to
bite: it is **scoped** (a stated population and horizon), **quantified** (a real number and a
comparison), **falsifiable** (a named observation would settle it), and it **defines its key term
inline** (so Equivocation and Hidden-premise have no unstated meaning to attack). None is universal in
form, so Counterexample can at most weaken.

## The three fragments

- **`true.txt`** — a scoped, quantified, checkable fact (the US federal minimum wage floor). True.
- **`false.txt`** — a 2007-vintage prediction that touchscreen-only handsets stay under 10% of global
  smartphone shipments through 2015. Well-formed and false — a **construction, not a real quotation**,
  and deliberately *not* Ballmer's "no significant market share" wording (that real quote is the
  separate `../laundering/` fixture). Its falsity is grounding's to establish, not substance's.
- **`conditional.txt`** — a scoped conditional (sub-replacement fertility with no net migration →
  population decline). Well-formed and defensible; the obvious counterexample (immigration) is closed
  off in the antecedent.

## The core fragment scored in each fixture

Each fixture decomposes into several atoms, but only one carries the proposition the control is
about; the rest are the definitional scaffolding the fixture states inline so the axes have nothing
to bite. `nb_tally.py` scores **that one core atom**, not the best verdict across all atoms — a
`best`-across-atoms score lets a trivially-true definition ("touchscreen-only smartphones are
handsets with no physical keyboard", certified `substantive`) stand in for the fragment and mask a
`hollow` core.

- **`true.txt`** — core is the wage fact: *"the federal minimum wage has been $7.25 … since July
  2009"* (matched by `7.25`). The definitional atoms (FLSA sets it; applies to covered non-exempt
  workers) are scaffolding.
- **`false.txt`** — core is the **prediction**: *"Touchscreen-only smartphones will account for less
  than 10% of global smartphone unit shipments in calendar year 2015"* (matched by `will account
  for less than 10%`, which excludes the separate causal atom about tactile-key preference). The
  "handsets with no physical keyboard" atom is the definition, not the claim.
- **`conditional.txt`** — core is the **conditional**: *"a total fertility rate below 2.1 sustained
  for a full generation with no net migration causes a country's population to fall"* (matched by
  `below 2.1`). The "TFR is the average number of children a woman bears" atom is the definition.

## Envelope (pre-registered)

- **All three must reach `substantive`** under a critic whose verdict keys on how far the claim had to
  be narrowed to defend it (the `-narrowing-boundary` variant, `assay.go`
  `substanceCriticSysNarrowingBoundary`). Each survives *as stated*, so nothing is narrowed.
- **Substance must not decide the false one.** `false.txt` is well-formed; substance should still pass
  it `substantive`. Only the **grounding** pass (`-evidence`, by retrieval) is licensed to return
  `refuted` on it — that split is the whole point of the axis boundary in `../../../CLAUDE.md`
  (reasoning can refute a self-contradiction but never confirm or refute how the world turned out).
  A run where substance sinks `false.txt` to `hollow`/`partial` on the Counterexample of the real 2015
  outcome is substance laundering grounding's job, and is the failure this control watches for.
- **Under the default boundary these are expected to FAIL.** The sonnet anatomy in
  `../../../docs/todo/destructive-sonnet-2026-09-13.md` established that the default verdict line
  ("no equivocation, survives counterexample") is unreachable for any general claim, because
  Equivocation and Hidden-premise weaken every bare claim of this form. So a `substantive` verdict on
  these three is evidence about the **boundary rule**, not about the fragments — a default-boundary
  collapse is the known negative, not a fixture bug.

## What a clean run means (and does not mean)

A clean run says only: *the working boundary certified these three on this model, this time.* It does
not say the critic "understands well-formedness." Per `../README.md`, these fixtures are public and may
eventually be recognised rather than reasoned about. Date every result.

## Calibration results

Not yet run. The refuter recipe (N=5, default vs `-narrowing-boundary`, on all three fragments plus the
honest motte, the bailey and unfalsifiable-dress) is in
`../../../docs/todo/destructive-sonnet-2026-09-13.md` § Refuter — narrowing-boundary variant.
