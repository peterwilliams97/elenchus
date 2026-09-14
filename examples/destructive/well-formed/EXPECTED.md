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
- **`prediction.txt`** — a post-cutoff prediction that battery-electric cars pass 25% of new EU
  passenger-car sales by the end of calendar year 2029. Well-formed, and its truth is **unknowable
  now**: the horizon sits past any current model's training cutoff, so no realised outcome exists for
  a critic to import. That is the whole point — it retires the earlier `false.txt`, a 2015-horizon
  touchscreen prediction whose known outcome the substance critic kept importing (see the Envelope
  note below). A **construction, not a real quotation**.
- **`conditional.txt`** — a scoped conditional (sub-replacement fertility with no net migration →
  population decline). Well-formed and defensible; the obvious counterexample (immigration) is closed
  off in the antecedent.

## The core fragment scored in each fixture

Each fixture decomposes into several atoms, but only one carries the proposition the control is
about; the rest are the definitional scaffolding the fixture states inline so the axes have nothing
to bite. The refuter scores **that one core atom**, not the best verdict across all atoms — a
`best`-across-atoms score lets a trivially-true definition ("battery-electric vehicles are cars with
no internal-combustion engine", certified `substantive`) stand in for the fragment and mask a
`hollow` core.

- **`true.txt`** — core is the wage fact: *"the federal minimum wage has been $7.25 … since July
  2009"* (matched by `7.25`). The definitional atoms (FLSA sets it; applies to covered non-exempt
  workers) are scaffolding.
- **`prediction.txt`** — core is the **prediction**: *"Battery-electric vehicles will make up more
  than 25% of new passenger-car unit sales in the European Union by the end of calendar year 2029"*
  (matched by `more than 25%`, which excludes the separate causal atom about EU fleet-emission
  limits). The "passenger cars powered solely by an onboard battery" atom is the definition, not the
  claim.
- **`conditional.txt`** — core is the **conditional**: *"a total fertility rate below 2.1 sustained
  for a full generation with no net migration causes a country's population to fall"* (matched by
  `below 2.1`). The "TFR is the average number of children a woman bears" atom is the definition.

## Envelope (pre-registered)

- **`prediction.txt` reaches `partial` on substance and `unverifiable` on grounding — pass = the
  surviving claim keeps the specifics and strips only the certainty.** A forward-looking claim always
  affords a `weakens`-level counterexample (a policy reversal, a demand shock), and the default
  `substantive` branch requires *surviving a counterexample*, so the critic downgrades to `partial` by
  attaching the producer's own stated conditions — correct behaviour on an unconditional forecast, not
  a collapse. The pass criterion is therefore **≥4/5 runs where the surviving claim keeps the number
  (25%), scope (EU new passenger-car unit sales) and horizon (end-2029) verbatim and changes only the
  modality** (`will` → `likely` / `provided …`). `substantive` is **not** expected for any
  forward-looking claim; `hollow` is a **fail** — it would mean the specifics were stripped or the
  realised outcome imported. Its end-2029 horizon sits past any current model's training cutoff, so an
  `-evidence` retrieval finds no realised outcome to confirm or refute: expected `unverifiable`, **not**
  `refuted`.
- **Substance must not decide the prediction on hindsight.** The retired `false.txt` (a 2015-horizon
  touchscreen prediction) failed exactly here — the substance critic imported the realised 2015
  shipment figures and sank the core `hollow` 10/10, doing grounding's job (the axis boundary in
  `../../../CLAUDE.md`: reasoning can refute a self-contradiction but never confirm or refute how the
  world turned out). The defect is anatomised in
  `../../../docs/todo/destructive-sonnet-2026-09-13.md` §§ Intervention 2–3. A post-cutoff horizon
  removes the outcome the critic was reaching for, so a substance run that sinks `prediction.txt` to
  `hollow` has no realised result left to cite — a `hollow` core is the failure this control now watches
  for; `partial` with the specifics intact is the pass.
- **`true.txt` and `conditional.txt` are recorded on the default critic** at `substantive` 5/5 and
  3/5 (2 partial) respectively (`../../../docs/todo/destructive-sonnet-2026-09-13.md` § Intervention 2
  Results): a well-scoped, inline-defined claim passes the default `substanceCriticSys` with no
  boundary change, refuting the earlier reading that `substantive` is structurally unreachable for
  every general claim.

## What a clean run means (and does not mean)

A clean run says only: *the working boundary certified these three on this model, this time.* It does
not say the critic "understands well-formedness." Per `../README.md`, these fixtures are public and may
eventually be recognised rather than reasoned about. Date every result.

## Calibration results

- **`prediction.txt`** — run 2026-09-14, N=5, default prompt
  (`../../../docs/todo/destructive-sonnet-2026-09-13.md` § Prediction fixture refuter). Core prediction
  atom `partial` **5/5**, `substantive` 0/5. Against the criterion above (number/scope/horizon verbatim,
  only the modality stripped) this is a **PASS 5/5**: all five surviving claims keep `25%` / EU new
  passenger-car unit sales / end-2029 and change only `will` → `likely` / `provided …`, attaching
  producer-stated conditions with no laundering. No run cites a realised outcome — the post-cutoff
  horizon removed the hindsight import that sank the retired `false.txt` (`hollow` 10/10). The earlier
  `substantive`-≥4/5 gate is superseded by the `partial` criterion above.
- **`prediction.txt` grounding (horizon gate)** — run 2026-09-14, N=3, default `-evidence`
  (`../../../docs/todo/destructive-sonnet-2026-09-13.md` § Horizon-gate refuter, leg 1). Grounding
  `unverifiable` **3/3**, with `horizon=future` 3/3, `original_verdict=mixed` 3/3, and
  `downgrade_reason` = `forecast — projections are not evidence` 3/3. The end-2029 forecast is
  code-downgraded to `unverifiable` while the model's own `mixed` verdict is banked in
  `original_verdict` — **PASS**. Specificity control (leg 2): the 2007 Ballmer laundering claim
  (`horizon=past`) keeps its `refuted` 3/3 with the gate silent (`original_verdict` empty), so the gate
  fires on forecasts without blanketing settled claims.
- The `true.txt` / `conditional.txt` / retired-`false.txt` runs against the default and the
  `-narrowing-boundary` / `-as-of` variants are recorded in the same doc (Interventions 2–3); all three
  variants (`-ce-scoped`, `-narrowing-boundary`, `-as-of`) remain non-default recorded negatives.
