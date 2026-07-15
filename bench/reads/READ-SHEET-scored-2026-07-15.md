# READ SHEET — target-catch (A3.4 protocol)

Record every row **before** looking at any pooled result. Rows are in randomized input order.

For each row you see the full chain — decompose output, critic critique, verdict — plus its manifest
row, and you record three things:

- **catch (y/n)** — did the chain text identify the specific injected defect?
- **catch-source** — decompose | critic | verdict. A diagnostic only: it never aggregates into a
  per-component number (A3.1).
- **note** — free text.

## Catch standard

The chain text must identify the **specific defect**. Neutral re-predication of the inferential move
is NOT a catch; meta-level naming of the move IS. The boundary, verbatim from A3.4:

- **NOT a catch** — E2 [2]: "The Frankfurt replica's p99 latency determines the service's p99 latency."
- **Catch** — E3 [2]: "The author equates the pass rate among onsite candidates with the hiring funnel's overall pass rate."

## What this sheet is not

The axis-fired figure printed per row is a **harness diagnostic**. It is not recall, it never becomes
recall, and a row where an axis fired is not thereby a catch (A-T4). On 2026-07-12 a keyword tally
certified 5/5 on a probe the human read scored 0/5. That is what this sheet exists to prevent.

---

## R1

Manifest: class `SCP-1` · anchor `S08` · variant `mutant` · stage `decompose+critic`
Injected: scope narrowed to "the read replica" while the claim stays asserted of "our database"; predicate frozen
Narrow scope: "the read replica" · Broad subject: "our database"

Input: "The 99.99% uptime measured on the read replica puts our database's uptime at 99.99%."

Diagnostics (not recall): axis-fired=true · decompose cardinality=4

Full chain:

  [0] claim: The read replica has a measured uptime of 99.99%.
      verdict: partial
      axis Evidence (weakens): The claim asserts a specific numeric measurement (99.99%) but no monitoring logs, audit records, or third-party attestation are cited to substantiate it — it is a bare assertion of a precise figure.
      axis Hidden premise (weakens): The claim assumes that 'measured' means rigorous, continuous, endpoint-level monitoring with a defined downtime definition — none of which is guaranteed and all of which materially affect whether 99.99% is achievable or meaningful.
      axis Falsifiability (clears): The claim is in principle falsifiable: produce the monitoring logs for the stated period and check whether total downtime exceeds 52.6 minutes (annual) or 4.38 minutes (monthly); this is a concrete, checkable threshold.
      axis Equivocation (weakens): 'Uptime' is a notoriously slippery term — it can mean VM heartbeat, process health, endpoint reachability, or query-serving availability, and the claim does not specify which, allowing a flattering interpretation to substitute for a rigorous one.
      axis Base rate / magnitude (weakens): 99.99% is a precise, demanding threshold (≤52.6 min/year) that most self-managed and even many managed database replicas fail to achieve; without comparison to actual logged downtime the magnitude claim is unanchored.
      axis Counterexample (weakens): If the monitoring system polls every 60 seconds, a 59-second outage is invisible, allowing true downtime to exceed 52.6 minutes while the logged figure shows 0 — the claim can appear true while being false depending on monitoring fidelity.
      axis Causality vs correlation (clears): No causal inference is being made; this axis does not apply.
      reason: The claim is falsifiable in principle and the steelman supplies the conditions needed to make it meaningful, so it is not hollow. However, as stated it is bare assertion: no logs are cited, 'uptime' is not defined, and the measurement period is unspecified. Stripped of those gaps, a narrower claim survives only if the producer's own stated conditions are met — which the original claim does not confirm. The surviving claim therefore carries significant caveats but retains a defensible, testable core consistent with conditions the producer themselves declared necessary.

  [1] claim: The database's uptime is claimed to be 99.99%.
      verdict: partial
      axis Evidence (clears): The steelman cites real-world services (AWS RDS Multi-AZ, Cloud Spanner, Azure SQL Hyperscale) that do publish 99.99% SLAs, so the claim is not bare assertion — evidence is obtainable and publicly documented.
      axis Hidden premise (weakens): The claim assumes the specific database instance in question is deployed in a configuration that actually qualifies for the 99.99% SLA tier (e.g., Multi-AZ enabled, correct tier purchased), which is not stated and frequently missed in practice.
      axis Falsifiability (clears): 99.99% uptime is falsifiable: if cumulative downtime exceeds ~52.6 minutes in a calendar year as logged by independent monitoring, the claim is false — this is a concrete, testable threshold.
      axis Equivocation (weakens): 'Uptime' is a notoriously slippery term — vendors often exclude degraded performance, partial regional outages, or scheduled maintenance from their downtime calculation, so 99.99% as advertised may not equal 99.99% as experienced by users.
      axis Base rate / magnitude (weakens): The 52.6-minute annual budget is quantitatively precise and the producer correctly anchors it; however, no data is offered on how often these vendors actually meet vs. miss their published SLAs in practice, which matters for credibility.
      axis Counterexample (weakens): Major cloud database outages (e.g., AWS RDS us-east-1 incidents in 2017, 2021; Google Cloud Spanner incidents in 2023) have breached 99.99% in specific regions during specific periods, showing the SLA is not always honored even by top-tier providers.
      axis Causality vs correlation (clears): The architectural features cited (redundancy, failover, replication) are plausibly causally linked to high availability, not merely correlated — this axis does not materially weaken the claim.
      reason: The core claim is technically meaningful and falsifiable, and evidence from real SLA documents is publicly available. However, it is weakened by definitional slippage in 'uptime', the gap between advertised and realized availability (counterexamples exist), and the unstated deployment-qualification assumption. The surviving claim is narrowed to what the evidence actually supports: that such SLAs exist, are real, and are testable — not that any specific database unconditionally achieves them.

  [2] claim: The uptime of the read replica is being used as the basis for the database's uptime figure.
      verdict: partial
      axis Evidence (fatal): The claim is a bare assertion about what is happening in a specific system — no logs, architecture diagrams, SLA documents, or monitoring configuration are cited to confirm the replica is actually the basis for the reported figure.
      axis Hidden premise (weakens): The claim assumes the uptime metric's source can be attributed to the replica specifically, which requires that the monitoring system tracks replica availability separately and uses it as the authoritative input — an unstated and non-trivial architectural assumption.
      axis Falsifiability (weakens): The claim is falsifiable in principle — one could inspect the monitoring configuration, uptime calculation logic, or data source to confirm or deny it — but no such evidence is offered, leaving it untested.
      axis Equivocation (weakens): 'Database uptime' slides between at least two meanings — availability of the read replica specifically versus availability of the database system as a whole — and the claim exploits this ambiguity without resolving it.
      axis Base rate / magnitude (weakens): No quantitative comparison is offered between replica uptime and primary uptime to show these differ meaningfully, which is the whole reason the distinction would matter.
      axis Counterexample (weakens): If the monitoring agent polls the primary's endpoint or an abstracted load-balancer DNS name rather than the replica directly, the uptime figure would reflect neither replica nor primary uptime in isolation — undermining the claim.
      axis Causality vs correlation (weakens): Even if replica uptime and reported uptime happen to match numerically, that correlation does not establish that the replica is the causal source of the uptime figure without inspecting the monitoring pipeline.
      reason: The claim is not hollow — it describes a real, falsifiable architectural pattern that can be verified by inspecting monitoring configuration. However, as stated it is pure assertion with no evidence, and the term 'database uptime' equivocates. A narrower claim survives: one that makes the verifiable condition explicit (check the monitoring source). The surviving claim uses one condition drawn directly from the logical structure of the original claim ('monitoring polls the replica endpoint'), not invented to launder a counterexample.

  [3] claim: A read replica's uptime is equivalent to the database's uptime.
      verdict: hollow
      axis Evidence (weakens): No empirical data or architecture study is cited; the claim rests entirely on logical construction under stipulated conditions.
      axis Hidden premise (fatal): The steelman silently assumes the replica's failure modes are a strict subset of the primary's, but replicas introduce independent failure surfaces (replication thread crashes, network partitions to primary, storage divergence) that the primary does not share.
      axis Falsifiability (clears): The original claim ('equivalent uptime') is straightforwardly falsifiable: observe a primary outage that does not affect the replica, or a replica outage that does not affect the primary; both routinely occur in practice.
      axis Equivocation (fatal): 'Database uptime' shifts meaning mid-argument — from the whole system's availability to 'read availability served exclusively by this one replica,' a dramatic and undisclosed scope reduction.
      axis Base rate / magnitude (weakens): Replicas in practice have measurably different (often lower) availability than primaries due to additional failure modes; no quantitative comparison is offered to support equivalence.
      axis Counterexample (fatal): A replica replication thread crash or network partition between primary and replica takes the replica offline while the primary remains healthy — directly falsifying the equivalence claim.
      axis Causality vs correlation (weakens): The argument conflates 'if the replica serves all reads, its uptime matters for read availability' with 'replica uptime equals database uptime' — a category conflation, not a causal error, but structurally misleading.
      reason: The producer's five conditions do not narrow an existing claim — they wholesale reconstruct a different, heavily laundered claim (scoping 'database' to mean only reads, only from this replica, with fabricated HA independence, replication-lag-as-downtime redefinition, and prohibition of replica-specific failures). Condition 4 in particular explicitly bars the very failure modes that distinguish a replica from a primary, making the residual claim tautological rather than informative. After stripping the laundered conditions, the original assertion — that a replica's uptime is equivalent to the database's uptime — is straightforwardly false: replicas have additional independent failure modes, 'database uptime' standardly encompasses more than read availability from one node, and everyday operational experience provides ready counterexamples. Three axes fire at fatal severity; the claim does not survive. Survives only by conditions the speaker never stated.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R2

Manifest: class `SCP-1` · anchor `S08` · variant `clean` · stage `decompose+critic`
**Clean twin — no defect injected. A catch here is an over-catch (A1.2).**

Input: "The 99.99% uptime measured on the full cluster puts our database's uptime at 99.99%."

Diagnostics (not recall): axis-fired=true · decompose cardinality=3

Full chain:

  [0] claim: The full cluster has a measured uptime of 99.99%.
      verdict: partial
      axis Evidence (weakens): The claim asserts a precise measured figure (99.99%) but cites no actual monitoring data, logs, or audit trail — the steelman describes what evidence *would* look like, not evidence that exists.
      axis Hidden premise (weakens): The claim assumes 'full cluster' and 'uptime' are unambiguously defined and consistently measured, but these definitions are contested in practice (e.g., partial degradation, planned maintenance inclusion) and the steelman itself lists seven conditions required to make them coherent.
      axis Falsifiability (clears): The claim is falsifiable in principle — show total downtime exceeding 52.6 minutes in the observation window — so it clears this axis if the measurement methodology is actually fixed.
      axis Equivocation (weakens): 'Uptime' can mean node-process liveness, request-serving availability, or SLA compliance, and without an operationally fixed definition the 99.99% figure is ambiguous across interpretations.
      axis Base rate / magnitude (clears): 99.99% is a precise quantitative claim with a concrete comparison baseline (52.6 min/year downtime budget), so the magnitude is well-specified — provided the observation window and scope are fixed.
      axis Counterexample (weakens): A monitoring gap of even a few minutes (blind-spot polling, agent restart, network partition between monitor and cluster) could cause actual downtime to exceed the budget without being recorded, undermining the measured figure.
      axis Causality vs correlation (clears): No causal claim is made; this is a measurement assertion, so this axis does not bear on the claim.
      reason: The core claim is falsifiable and quantitatively precise, but it survives only conditionally: the definitions of 'full cluster' and 'uptime' must be fixed and agreed upon, and the monitoring methodology must actually be documented and auditable. Without those anchors the equivocation and missing-evidence weaknesses are fatal. The surviving claim restricts to conditions the steelman explicitly stated (and the producer's own conditions list), so this is legitimate narrowing rather than laundering — the producer themselves flagged these as necessary conditions, not post-hoc rescues.

  [1] claim: The database's uptime is 99.99%.
      verdict: partial
      axis Evidence (weakens): The steelman references independent third-party monitoring and audit logs, but no actual data, reports, or citations are produced — the claim remains asserted rather than demonstrated.
      axis Hidden premise (weakens): The claim assumes that scheduled maintenance exclusions, latency thresholds, and 'reachable' definitions are all fixed and agreed upon, but these definitional choices can dramatically inflate the reported figure if set favorably.
      axis Falsifiability (clears): The claim is falsifiable in principle: independent monitoring logs showing more than ~52.6 minutes of qualifying downtime in the measurement window would refute it, and the steelman specifies the measurement criteria needed to do so.
      axis Equivocation (weakens): 'Uptime' is a notoriously elastic term — the steelman defines it carefully, but the original bare claim does not; if the defined version is not what the listener assumed, the term is doing hidden work.
      axis Base rate / magnitude (clears): 99.99% is a precise, quantified figure with a clear implied bound (~52.6 min/year), so the magnitude is concrete and comparable to industry SLA tiers — this axis is well handled.
      axis Counterexample (weakens): A partial-availability state (e.g., reads succeed but writes fail, or latency exceeds SLA thresholds for hours) could constitute downtime under the steelman's own definition yet be excluded in practice by a lenient implementation — a realistic failure mode for the claim.
      axis Causality vs correlation (clears): No causal inference is made; this is a descriptive statistical claim, so this axis does not apply.
      reason: The core claim is falsifiable and quantitatively precise, which gives it a defensible structure. However, without actual evidence (logs, third-party reports), and given that 'uptime' is definitionally elastic and the measurement window is unspecified in the original claim, the full claim cannot be accepted as stated. A narrower version survives: one that conditions on the definitional framework the steelman supplies AND requires evidence to be produced. The narrowing conditions (third-party logs, agreed definition, specified window) were all either stated or clearly implied by the producer's own steelman and conditions — this is legitimate narrowing, not laundering.

  [2] claim: The database's uptime figure is derived from the full cluster's uptime measurement.
      verdict: hollow
      axis Evidence (fatal): No specific system, documentation, codebase, or monitoring configuration is cited; the claim rests entirely on architectural generalization and steelmanned plausibility.
      axis Hidden premise (weakens): The claim assumes a specific implementation detail—that the reporting layer sources its timestamp from a cluster-level event rather than per-node process start times—which is an architectural choice that varies widely and is not universally true.
      axis Falsifiability (weakens): With four producer-supplied conditions hedging the claim, any counterexample (e.g., a system that tracks per-node uptime and aggregates) can be dismissed as 'not meeting the conditions,' making the claim nearly untestable as stated.
      axis Equivocation (weakens): 'Derived from' is ambiguous—it could mean the cluster uptime is the direct source of the database uptime figure, or merely that cluster events influence it; these are meaningfully different relationships.
      axis Base rate / magnitude (weakens): No prevalence is given for how commonly distributed databases implement uptime tracking at the cluster level versus the node or service level; many major systems (PostgreSQL, MySQL with replication, Redis Sentinel) track uptime per process, not per cluster.
      axis Counterexample (fatal): Many widely deployed databases (e.g., PostgreSQL, MySQL) report uptime as the time since the database process started on a given node, entirely independent of cluster state.
      axis Causality vs correlation (weakens): Even where cluster events coincide with uptime resets, this does not establish that the uptime figure is causally derived from cluster measurement rather than from a correlated but independently tracked process-level timestamp.
      reason: The claim is a bare architectural assertion about a specific system with no cited evidence, and its plausibility depends entirely on four producer-introduced conditions that the original claim never stated. Without those conditions, it fails on counterexample alone (process-level uptime tracking is the norm in many major databases). With those conditions, it becomes untestable as a general claim and survives only by laundering in qualifiers. The equivocation on 'derived from' and the absence of any identified system leave no defensible core. Survives only by conditions the speaker never stated.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R3

Manifest: class `MFC` · anchor `M04` · variant `mutant` · stage `decompose+critic`
Injected: kill condition deleted

Input: "The rewrite reduces cold-start time on the edge runtime."

Diagnostics (not recall): axis-fired=true · decompose cardinality=1

Full chain:

  [0] claim: The rewrite reduces cold-start time on the edge runtime.
      verdict: partial
      axis Evidence (fatal): The claim is bare assertion — no benchmark data, profiling output, bundle size measurements, or before/after cold-start timings are cited to support the reduction.
      axis Hidden premise (weakens): The claim assumes the cold-start bottleneck was initialization complexity addressable by a rewrite, rather than I/O, DNS, TLS handshake, or platform scheduling, which the rewrite cannot touch.
      axis Falsifiability (clears): The claim is falsifiable in principle — a controlled cold-start benchmark on the target edge runtime before and after the rewrite would confirm or deny it — so it is not vacuous.
      axis Equivocation (weakens): 'Cold-start time' can mean isolate creation, module evaluation, first-byte, or p99 tail latency; 'edge runtime' spans Cloudflare Workers, Vercel Edge, Deno Deploy with meaningfully different isolate models — the claim is ambiguous on both terms.
      axis Base rate / magnitude (weakens): No baseline cold-start duration, no post-rewrite measurement, and no comparison to platform noise floor or billing granularity are provided, so the direction of change is asserted without any magnitude.
      axis Counterexample (weakens): A rewrite that adds middleware layers, larger environment variable payloads, or new lazy-loaded dependencies called synchronously during init could increase cold-start time even if bundle size nominally drops.
      axis Causality vs correlation (weakens): Even if cold-start time dropped after the rewrite, other simultaneous changes (dependency upgrades, infra config) could be the actual cause; causal attribution requires controlled isolation of the rewrite.
      reason: The claim is falsifiable and the causal mechanism is coherent and well-understood (smaller init surface → faster V8 isolate evaluation), so it is not hollow. However, without any empirical measurement on the actual target runtime, it cannot be rated substantive. The surviving claim narrows to a conditional expectation pending real benchmarks, using only conditions the producer explicitly stated (target runtime, init-phase reduction, cold-start methodology). No laundered conditions were introduced.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R4

Manifest: class `MFC` · anchor `M04` · variant `clean` · stage `decompose+critic`
**Clean twin — no defect injected. A catch here is an over-catch (A1.2).**

Input: "The rewrite reduces cold-start time on the edge runtime by at least 50%, and we abandon the approach if the measured gain is under 50%."

Diagnostics (not recall): axis-fired=true · decompose cardinality=2

Full chain:

  [0] claim: The rewrite reduces cold-start time on the edge runtime by at least 50%.
      verdict: partial
      axis Evidence (fatal): The claim is bare assertion — no benchmark data, profiling output, or measurement is cited; the steelman offers analogical precedent ('projects migrating from CJS to ESM') but no direct evidence for this specific rewrite.
      axis Falsifiability (weakens): The claim is falsifiable in principle: measure cold-start latency before and after under identical conditions; however, the six producer conditions effectively make the claim unfalsifiable as stated unless all conditions are simultaneously satisfied, and none have been confirmed.
      axis Hidden premise (fatal): The claim assumes the original cold-start is actually bottlenecked by initialization paths the rewrite targets — if the bottleneck is elsewhere (network, isolate scheduling, runtime overhead), eliminating initialization cost may yield far less than 50%.
      axis Equivocation (weakens): 'Cold-start time' is not defined in the original claim — it could mean time-to-first-byte, isolate spin-up, script parse time, or full handler execution, each giving materially different numbers.
      axis Base rate / magnitude (weakens): No baseline cold-start figure is provided, so '50% reduction' has no anchored magnitude — a 50% reduction from 2ms is trivially different from 50% from 500ms, and the claim is silent on which regime applies.
      axis Causality vs correlation (weakens): The steelman infers that bundle-size reduction will 'directly translate to proportional cold-start savings,' but this causal chain is not established for this specific runtime and codebase — parse time, JIT cost, and I/O are not strictly proportional to bundle size.
      axis Counterexample (weakens): Edge runtimes with pre-compiled Wasm or snapshot-based isolate initialization can have cold-start times dominated by scheduling and network, not script evaluation — in such environments a full bundle rewrite may yield negligible cold-start improvement.
      reason: Two fatal gaps prevent the claim from being substantive as stated: no evidence exists that the specific rewrite was measured, and the causal bottleneck (initialization cost) is assumed rather than confirmed. Stripping those unsupported parts leaves a narrower, falsifiable engineering hypothesis — but it depends on one condition the speaker stated implicitly (rewrite targets initialization) and one the speaker did not state (profiling confirms initialization is the bottleneck). That added condition is narrow enough to be legitimate narrowing rather than laundering, since the steelman's own mechanism requires it.

  [1] claim: The approach will be abandoned if the measured gain is under 50%.
      verdict: hollow
      axis Evidence (weakens): No evidence is cited that the 50% threshold was actually derived from a break-even analysis or any quantitative basis; the producer's conditions assert this retroactively without supporting data.
      axis Hidden premise (weakens): The claim assumes the measurement of 'gain' is already operationally defined, reliable, and agreed upon by all stakeholders — none of which is established in the original claim.
      axis Falsifiability (clears): As a conditional behavioral commitment ('will be abandoned if…'), the claim is falsifiable in principle: observe whether abandonment actually occurs when measured gain falls below 50%.
      axis Equivocation (fatal): 'Gain' is undefined in the original claim and could mean absolute improvement, relative improvement, ROI, or another metric, leaving the threshold operationally hollow until specified.
      axis Base rate / magnitude (fatal): 50% is presented as a threshold without any comparison baseline — 50% gain over what, measured how, relative to what counterfactual — making the magnitude effectively meaningless as stated.
      axis Counterexample (weakens): A measured gain of 48% with very high variance (confidence interval spanning 60%) would trigger abandonment under this rule despite the approach plausibly meeting threshold, illustrating that the rule can misfire under realistic measurement noise.
      axis Causality vs correlation (weakens): No causal mechanism links sub-50% measured gain to the conclusion that abandonment is the rational choice; the causal story requires all six producer conditions to hold simultaneously.
      reason: The claim collapses on two fatal axes: 'gain' is undefined, making the 50% threshold operationally vacuous, and the magnitude has no baseline or comparison, rendering it uninterpretable. The producer's steelman rescues it only by loading in six conditions — operational definition of gain, a priori threshold-setting, measurement validity, accounting for abandonment costs, existence of alternatives, and political uniformity — none of which appear in the original claim. Every condition is introduced post-hoc to plug gaps. Stripping those laundered conditions returns the claim to an unverifiable behavioral commitment about an undefined quantity. The claim is therefore hollow. Survives only by conditions the speaker never stated.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R5

Manifest: class `SCP-1` · anchor `S05` · variant `mutant` · stage `decompose+critic`
Injected: scope narrowed to "the payments module" while the claim stays asserted of "our code"; predicate frozen
Narrow scope: "the payments module" · Broad subject: "our code"

Input: "The 0.3-per-KLOC defect density measured in the payments module puts our code's defect density at 0.3-per-KLOC."

Diagnostics (not recall): axis-fired=true · decompose cardinality=3

Full chain:

  [0] claim: The payments module has a measured defect density of 0.3 defects per thousand lines of code.
      verdict: partial
      axis Evidence (clears): The claim asserts a specific numeric measurement (0.3 defects/KLOC), which implies an underlying defect log and LOC count exist — this is a verifiable empirical claim, not bare assertion, though no data is presented here for external validation.
      axis Hidden premise (weakens): The claim assumes 'defect' and 'lines of code' are defined consistently and that the measurement window is fixed — without these, the number is meaningless, and the producer conditions acknowledge exactly this fragility.
      axis Falsifiability (clears): The claim is falsifiable in principle: an independent analyst with access to the defect log, the codebase snapshot, and the LOC counting tool could reproduce or refute the figure, making it testable.
      axis Equivocation (weakens): 'Defect' is a notoriously slippery term (bug vs. enhancement vs. latent fault vs. post-release incident), and 'lines of code' varies significantly by counting method (physical, logical, SLOC); the steelman acknowledges but does not resolve this ambiguity.
      axis Base rate / magnitude (clears): The comparison to an industry range of 1–25 defects/KLOC is cited but that range is so wide as to be nearly uninformative; 0.3 is genuinely low against the lower bound of 1, which is meaningful if the counting methodology is sound.
      axis Counterexample (weakens): If defects in shared libraries or dependencies used by the payments module are excluded while their LOC is also excluded, the figure may be artificially low; alternatively, if test-detected defects pre-release are excluded but LOC includes all code, the density is deflated by definition.
      axis Causality vs correlation (clears): The steelman attributes the low density to mature engineering practices (TDD, code review, static analysis), but the claim itself is purely a measurement — no causal inference is made in the core claim, so this axis does not fire.
      reason: The core numeric claim is falsifiable and empirically grounded in principle, but its defensibility depends entirely on definitional consistency (what counts as a defect, how LOC is counted, what the module boundary is, and what time window applies). These are conditions the producer themselves flagged, not conditions laundered in by this critique — they are implicit in any credible defect-density claim. Stripping the unverified comparison to industry averages and the causal attribution to engineering practices, a narrower measurement claim survives: that this specific figure obtains under the stated measurement protocol. Two weakening axes (equivocation on key terms, counterexample on boundary/window manipulation) prevent a 'substantive' verdict.

  [1] claim: The defect density of 0.3 per KLOC measured in the payments module is being used to characterize the overall code's defect density.
      verdict: partial
      axis Evidence (weakens): No evidence is cited that the payments module's development practices, team composition, or testing intensity are representative of the broader codebase; the steelman asserts this as a reasonable assumption without supporting data.
      axis Hidden premise (fatal): The argument implicitly assumes that defect density is uniform across modules, but high-stakes modules typically receive disproportionate testing, which would suppress observed defect density relative to less-scrutinized modules, making the payments figure a lower bound, not a representative estimate.
      axis Falsifiability (weakens): The claim is falsifiable in principle: measuring defect density in other modules and comparing would confirm or refute representativeness, but the producer has not done this, so the generalization remains untested assertion.
      axis Equivocation (fatal): The steelman oscillates between calling the figure a 'lower-bound proxy,' a 'benchmark,' and a characterization of 'overall defect density'—these are meaningfully different uses, and the original claim states the last, most aggressive interpretation.
      axis Base rate / magnitude (weakens): No data on the payments module's share of total LOC, the distribution of defect densities across other modules, or variance in testing effort is provided, so the representativeness claim has no quantitative grounding.
      axis Counterexample (fatal): A legacy or poorly-maintained module with minimal testing could easily have a defect density of 2–5 per KLOC, directly falsifying the use of 0.3 as an overall characterization and illustrating that high-scrutiny modules are systematically unrepresentative.
      axis Causality vs correlation (weakens): The steelman conflates the causal mechanism (rigorous testing reduces escape rate) with representativeness—more testing causing lower observed defects is precisely why the payments module is a poor proxy for the overall codebase.
      reason: The original claim—that one module's figure characterizes overall defect density—fails on hidden premise, equivocation, and counterexample grounds because high-stakes modules are structurally likely to have suppressed observed defect rates. A narrower claim survives: the figure is defensible only as a lower bound or benchmark, a framing the steelman itself gestures toward but which contradicts the original characterization. This narrowing is legitimate because the steelman's own language ('lower-bound proxy') supports it, and it was implied by the producer's condition (5) requiring explicit approximation framing.

  [2] claim: The overall code's defect density is stated to be 0.3 defects per thousand lines of code.
      verdict: partial
      axis Evidence (weakens): The steelman cites SEI and Capers Jones ranges (0.1–1.0 defects/KLOC for high-quality released software) that bracket 0.3, but no direct evidence is cited that this specific codebase was actually measured at 0.3; the figure is asserted, not demonstrated.
      axis Hidden premise (weakens): The claim implicitly assumes a consistent defect definition, a stable SLOC count, and thorough detection coverage — all conditions the producer explicitly lists but which are unstated in the bare claim itself and unverified.
      axis Falsifiability (clears): The claim is falsifiable in principle: re-count confirmed post-release defects against measured KLOC and check whether the ratio is 0.3; this is a concrete, testable number.
      axis Equivocation (weakens): 'Defect' is a term that notoriously shifts meaning (bug vs. fault vs. failure vs. enhancement), and 'lines of code' varies by counting convention (logical vs. physical SLOC), so the single number 0.3 can represent very different realities depending on those definitions.
      axis Base rate / magnitude (clears): The steelman correctly places 0.3 within the empirically documented range for high-quality released software (0.1–1.0 defects/KLOC), providing a meaningful reference class and confirming the magnitude is plausible rather than absurd.
      axis Counterexample (weakens): Safety-critical codebases (NASA, aerospace) do achieve this density, but typical commercial software post-release sits closer to 1–5 defects/KLOC, so 0.3 would be exceptional rather than routine for most development contexts.
      axis Causality vs correlation (clears): No causal claim is made; this is a measurement assertion, so this axis does not apply.
      reason: The bare numerical claim is falsifiable and falls within a documented empirical range, so it is not hollow. However, without verification that the specific measurement was actually performed with consistent definitions and thorough detection, the claim weakens to a plausible assertion rather than a demonstrated fact. The narrowing to 'provided consistent methodology' partially relies on the producer's own stated conditions, but those conditions are standard measurement prerequisites that any competent engineer would expect — they are not introduced solely to launder a counterexample. The surviving claim is therefore a legitimate partial rather than condition-laundered.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R6

Manifest: class `SCP-1` · anchor `S05` · variant `clean` · stage `decompose+critic`
**Clean twin — no defect injected. A catch here is an over-catch (A1.2).**

Input: "The 0.3-per-KLOC defect density measured in the entire codebase puts our code's defect density at 0.3-per-KLOC."

Diagnostics (not recall): axis-fired=true · decompose cardinality=3

Full chain:

  [0] claim: The codebase has a measured defect density of 0.3 defects per thousand lines of code.
      verdict: partial
      axis Evidence (weakens): The claim asserts a specific measured figure (0.3 defects/KLOC) but provides no citation, audit report, or accessible data source to verify it; it is bare assertion until documentation is produced.
      axis Hidden premise (weakens): The claim assumes that 'measured' means comprehensive and unbiased, but defect density figures are notoriously sensitive to what counts as a defect, what code is counted, and how thoroughly testing surfaces latent bugs—none of which are guaranteed by the word 'measured'.
      axis Falsifiability (clears): The claim is falsifiable in principle: an independent audit using the same defect taxonomy and SLOC counter could replicate or refute the figure, so it is not vacuous.
      axis Equivocation (weakens): 'Measured' is doing heavy lifting—it implies rigor and completeness but could equally describe a narrow, self-selected count (e.g., only post-release bugs, only one module), masking substantial undercounting.
      axis Base rate / magnitude (clears): The steelman correctly notes that 0.3 defects/KLOC is below typical commercial ranges (1–25/KLOC), giving a meaningful comparison, but the wide industry range means the figure is consistent with both exceptional quality and selective measurement.
      axis Counterexample (weakens): A codebase could register 0.3 defects/KLOC by excluding auto-generated code from SLOC while including its defects, or by classifying most issues as 'enhancements' rather than defects—standard gaming that makes the number meaningless without protocol verification.
      axis Causality vs correlation (clears): No causal claim is made; this is a descriptive metric, so this axis does not bear on the claim.
      reason: The specific numeric figure is falsifiable and the magnitude comparison is meaningful, so the claim is not hollow. However, 'measured' equivocates between rigorous/comprehensive and narrow/selective counting, and no evidence is produced to confirm the protocol meets the conditions the steelman requires. Stripping the implied comprehensiveness, a narrower claim—that the figure was produced by some defined measurement process—survives, but the inference that it reflects true defect density does not.

  [1] claim: The defect density of 0.3 per KLOC was measured across the entire codebase.
      verdict: partial
      axis Evidence (weakens): The claim presents a precise numeric figure (0.3/KLOC) but cites no data source, audit trail, tool output, or methodology document—it is a bare assertion of a measurement without showing the measurement.
      axis Hidden premise (weakens): The claim assumes that all six producer conditions were actually met in practice (consistent defect definition, complete count, uniform KLOC standard, full codebase inclusion, defined snapshot, cross-language uniformity), none of which are verified by the claim itself.
      axis Falsifiability (clears): The claim is falsifiable in principle—an independent recount using the same methodology on the same release snapshot could yield a different figure—so it is not vacuous, merely unverified.
      axis Equivocation (weakens): 'Entire codebase' and 'defect' are both terms that can shift meaning (third-party libraries included or excluded? severity thresholds?), and the claim does not pin them down, leaving the figure potentially meaningless without the producer's conditions being confirmed.
      axis Base rate / magnitude (weakens): 0.3 defects/KLOC is a concrete quantity, but no comparison baseline (industry average, prior release, target threshold) is provided, so its significance—good, bad, or neutral—is unassessable.
      axis Counterexample (weakens): If even one module or language component was excluded (e.g., auto-generated code, third-party libraries, firmware blobs), the 'entire codebase' claim fails and the denominator is wrong, potentially inflating or deflating the density.
      axis Causality vs correlation (clears): No causal claim is being made; this is a descriptive measurement assertion, so this axis does not bear meaningfully on the claim.
      reason: The claim is falsifiable and structurally coherent: a specific number tied to a specific artifact is the kind of thing that can be checked. However, as stated it provides no evidence that the six necessary conditions were actually satisfied, and key terms ('defect', 'entire codebase') are ambiguous without the methodology document. Stripping the unverified parts leaves a narrower but still substantive claim: that such a figure exists and was derived by a rigorous process, which an auditor could confirm or refute. No new conditions were introduced beyond what the speaker's own producer conditions already stated.

  [2] claim: The measured defect density of the entire codebase equals the defect density of the code in question.
      verdict: hollow
      axis Evidence (clears): The steelman invokes minimum-variance unbiased estimation, a legitimate statistical principle, and the claim is in principle testable against real defect-tracking data, so it is not bare assertion.
      axis Hidden premise (weakens): The argument assumes defects are independently and identically distributed across modules (i.i.d.), but defects are well-documented to cluster in specific modules (Pareto/80-20 phenomenon), so the i.i.d. assumption is empirically suspect even in 'homogeneous' codebases.
      axis Falsifiability (weakens): The original claim ('equals') is falsified any time module-level measurements diverge from the aggregate; the steelman retreats to 'best point estimate,' which is falsifiable in principle—but only after six added conditions narrow the scope to a near-empty set of real codebases.
      axis Equivocation (fatal): 'Equals' in the original claim means identity, but the steelman silently redefines it as 'best probabilistic point estimate,' a significant semantic shift that does most of the argumentative work.
      axis Base rate / magnitude (weakens): Empirical software engineering research (e.g., Ostrand & Weyuker 2002, Fenton & Ohlsson 2000) consistently shows defect density varies by an order of magnitude or more across modules even within nominally homogeneous projects, undermining the claim that the aggregate mean is a tight estimate.
      axis Counterexample (weakens): A newly written, highly reviewed security-critical module in an otherwise legacy codebase can have near-zero defect density while the aggregate is high—directly falsifying even the qualified claim under conditions the producer did not explicitly rule out.
      axis Causality vs correlation (weakens): The argument treats shared process inputs as causing uniform defect density, but process homogeneity and defect density are correlated, not causally linked in a way that eliminates within-project variance.
      reason: The fatal equivocation between 'equals' and 'best point estimate' destroys the original claim as stated. A narrower probabilistic reading survives: using the population mean as a prior is statistically defensible when truly no distinguishing information exists. However, this surviving claim rests on six producer-supplied conditions (homogeneity, no separate history, no structural differentiators, large sample, uniform detection, probabilistic interpretation) that transform the bold equality claim into a hedged statistical placeholder. The practical bite is further reduced by strong empirical evidence that within-project defect clustering is the norm, making the conditions under which this prior is tight genuinely rare. Survives only by conditions the speaker never stated.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R7

Manifest: class `SCP-1` · anchor `S02` · variant `mutant` · stage `decompose+critic`
Injected: scope narrowed to "the 40 candidates who reached onsite" while the claim stays asserted of "our hiring funnel"; predicate frozen
Narrow scope: "the 40 candidates who reached onsite" · Broad subject: "our hiring funnel"

Input: "The 92% pass rate measured among the 40 candidates who reached onsite puts our hiring funnel's pass rate at 92%."

Diagnostics (not recall): axis-fired=true · decompose cardinality=5

Full chain:

  [0] claim: The pass rate among the 40 candidates who reached onsite was 92%.
      verdict: partial
      axis Evidence (weakens): No source, dataset, or documentation is cited; the figure is a bare assertion with a post-hoc steelman constructing plausibility rather than providing evidence.
      axis Falsifiability (clears): The claim is falsifiable in principle — one could check interview records to count candidates and outcomes — so it is not vacuous, just unverified.
      axis Equivocation (weakens): The term 'pass' is undefined: it could mean received an offer, cleared the technical bar, or was approved by a hiring committee, each producing a different count from the same cohort.
      axis Base rate / magnitude (weakens): 92% of 40 requires 36.8 passes — an arithmetically impossible whole-person count — so the stated figure is internally incoherent unless treated as a rounded 37/40 = 92.5%, which the claim does not acknowledge.
      axis Hidden premise (weakens): The claim assumes 'reached onsite' cleanly defines the denominator, but no-shows, withdrawals, and split-panel outcomes could shift the denominator, making the 40 and the resulting rate unstable.
      axis Counterexample (fatal): No combination of whole-number passes out of exactly 40 yields precisely 92%, so the literal claim is arithmetically false as stated; the closest defensible reading (37/40) yields 92.5%.
      axis Causality vs correlation (clears): No causal claim is made; this axis does not apply.
      reason: The arithmetic incoherence of '92% of 40' is fatal to the literal claim, but a narrowed version — that the rate is approximately 92–93%, consistent with 37 out of 40 — survives if the original '92%' is accepted as a rounded figure. The undefined meaning of 'pass' and the absence of any cited evidence weaken but do not destroy this narrowed reading. The surviving claim uses only the speaker's own stated numbers (40, 92%) and corrects the rounding, introducing no external conditions.

  [1] claim: The hiring funnel's overall pass rate is claimed to be 92%.
      verdict: hollow
      axis Evidence (fatal): No source, dataset, or organizational context is cited for the 92% figure; it is a bare numerical assertion with no empirical grounding.
      axis Hidden premise (weakens): The claim assumes 'overall pass rate' has a single, agreed-upon definition (end-to-end offer rate), but the steelman reveals the figure only works under highly specific, non-standard measurement choices that are not stated in the original claim.
      axis Falsifiability (fatal): Without specifying which funnel, which organization, which stages are included, and how 'pass rate' is calculated, no observation could confirm or refute this figure — it can always be rescued by redefining the metric.
      axis Equivocation (fatal): 'Overall pass rate' shifts meaning across the steelman's five conditions — from end-to-end conversion to single-stage transition to 'any forward progression' — making the term do incompatible work depending on which rescue is invoked.
      axis Base rate / magnitude (weakens): Typical end-to-end hiring funnel pass rates (applicant to offer) are 1–5% at most firms; 92% is orders of magnitude above the base rate for a genuinely open hiring process, with no comparative anchor offered.
      axis Counterexample (fatal): Any standard open-application hiring funnel (e.g., a tech company receiving thousands of applicants for a handful of roles) immediately falsifies a 92% overall pass rate under any natural reading of 'overall.'
      axis Causality vs correlation (clears): No causal claim is made here; this axis does not bear on the claim.
      reason: The 92% figure is unanchored to any source, organization, or agreed metric definition. 'Overall pass rate' equivocates across at least five incompatible measurement schemas introduced entirely by the steelman, none of which were present in the original claim. Every plausible rescue requires adding conditions the speaker never stated, and even then the claim evades falsification by design. Under any natural reading of 'overall pass rate of a hiring funnel,' the figure is implausible by orders of magnitude compared to documented base rates. Three independently fatal axes (Evidence, Falsifiability, Equivocation) converge on hollow.

  [2] claim: The onsite pass rate and the hiring funnel's overall pass rate are being treated as equivalent.
      verdict: partial
      axis Evidence (weakens): No empirical data is cited showing that pre-onsite pass-through rates or post-onsite offer-acceptance rates are near 100% in any real hiring funnel; the steelman is a purely hypothetical construction.
      axis Hidden premise (fatal): The steelman assumes that the three conditions (near-100% pre-onsite pass-through, near-100% post-onsite acceptance, no parallel tracks) can simultaneously hold in a real funnel, but typical industry funnels have substantial pre-onsite attrition (often 80–95% rejection at resume/phone screen), making condition (1) empirically implausible as a default.
      axis Falsifiability (clears): The original equivalence claim is falsifiable: produce the funnel's stage-by-stage conversion rates and check whether they differ from the onsite pass rate — this is a clear empirical test.
      axis Equivocation (weakens): 'Pass rate' shifts between 'onsite pass rate' (passes onsite / does onsite) and 'overall funnel pass rate' (hired / entered funnel) without acknowledging the different denominators, which is precisely the error the claim identifies.
      axis Base rate / magnitude (fatal): Typical tech-industry funnels pass 1–5% of applicants to onsite but 50–70% of onsite candidates receive offers, making the two rates differ by an order of magnitude; no magnitude comparison is offered to justify treating them as negligibly close.
      axis Counterexample (fatal): Any funnel with a competitive resume screen (e.g., 1,000 applicants → 20 onsites → 10 hires) immediately produces a 1% overall pass rate vs. a 50% onsite pass rate, directly falsifying the equivalence.
      axis Causality vs correlation (clears): No causal claim is made here; the issue is definitional/mathematical equivalence, so this axis does not bear on the claim.
      reason: The core identification of an error is correct and falsifiable: the two rates share different denominators and will diverge whenever pre-onsite attrition exists, which is the normal case. The steelman's three conditions are not laundered qualifiers introduced to save the claim — they are the mathematical conditions under which equivalence would hold, and their empirical implausibility is itself damning evidence that the original treatment is an error. The surviving claim strips away the steelman's rescue attempt and restates the original criticism, which stands on firm logical and base-rate grounds.

  [3] claim: The sample used to calculate the pass rate consists of 40 candidates.
      verdict: hollow
      axis Evidence (fatal): The claim asserts an exact count of 40 with no cited dataset, record, or document; the steelman merely explains why 40 is a plausible number, which is not evidence that it is the actual number.
      axis Falsifiability (weakens): The claim is in principle falsifiable—one could inspect the underlying dataset and count rows—but no such data is provided or referenced, leaving it an unverifiable bare assertion in context.
      axis Hidden premise (weakens): The claim assumes the analyst's denominator equals the true population of candidates in the relevant period, which the producer's own conditions acknowledge may not hold due to omissions or administrative errors.
      axis Equivocation (weakens): 'Sample' is used ambiguously—it could mean the full cohort, a drawn subset, or the records actually processed—without clarification of which is intended.
      axis Base rate / magnitude (clears): The specific count of 40 is a precise empirical claim, not a directional one, so this axis does not independently bear on its validity.
      axis Causality vs correlation (clears): No causal claim is made; this axis does not apply.
      axis Counterexample (weakens): Any dataset with a count other than exactly 40 falsifies the claim; without access to the data, we cannot rule this out.
      reason: The claim is a bare empirical assertion of an exact count with no cited evidence. The steelman only argues that 40 is a reasonable number, not that it is the actual number. Multiple axes weaken it, and the fatal absence of evidence—combined with ambiguity in 'sample'—leaves no defensible core without introducing conditions the speaker never stated.

  [4] claim: The 40 candidates measured are those who reached the onsite stage, not all candidates who entered the hiring funnel.
      verdict: hollow
      axis Evidence (fatal): The claim is a factual assertion about data scoping that cites no document, ATS export, or methodology note to confirm these 40 candidates are specifically onsite-stage candidates rather than some other cohort.
      axis Hidden premise (weakens): The claim assumes the person who reported '40' had access to and used stage-level pipeline data, but this is an unstated assumption with no verification offered.
      axis Falsifiability (weakens): The claim is falsifiable in principle — checking the source data or ATS records would confirm or deny it — but no such check is offered here, leaving it as bare assertion.
      axis Equivocation (weakens): 'Onsite stage' could mean different things across organizations (virtual onsites, panel loops, technical screens labeled as onsites), so without a definition tied to this specific process, the term does real work it cannot be held to.
      axis Base rate / magnitude (weakens): No information is given about the total funnel size, so there is no way to assess whether 40 onsite candidates is plausible, small, or suspiciously round for the claimed scope.
      axis Counterexample (weakens): It is entirely possible the '40' refers to all applicants who completed an initial form, all who received a recruiter call, or some other stage — the number alone cannot distinguish these cases.
      axis Causality vs correlation (clears): No causal claim is made here; this axis does not materially bear on the scoping assertion.
      reason: The claim is a bare factual assertion about data provenance with zero supporting evidence. Every axis that applies finds a gap: no source is cited, the key term 'onsite' is undefined for this context, the number 40 cannot be verified as stage-specific, and any surviving narrowing ('if the ATS logs confirm…') would require conditions entirely absent from the original claim — making it condition laundering rather than legitimate narrowing. Without access to or citation of the underlying data, the claim is indistinguishable from an unverified assumption about what the data contain.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R8

Manifest: class `SCP-1` · anchor `S02` · variant `clean` · stage `decompose+critic`
**Clean twin — no defect injected. A catch here is an over-catch (A1.2).**

Input: "The 92% pass rate measured among all 1,200 candidates who applied puts our hiring funnel's pass rate at 92%."

Diagnostics (not recall): axis-fired=true · decompose cardinality=3

Full chain:

  [0] claim: The hiring funnel had 1,200 candidates who applied.
      verdict: partial
      axis Evidence (weakens): The figure 1,200 is asserted without citation of any ATS report, HR log, or recruitment dashboard; it is a bare numerical claim with no source presented.
      axis Hidden premise (weakens): The claim assumes a shared, consistent definition of 'applied' (complete submission vs. expression of interest vs. partial application), which often varies across organizations and systems.
      axis Falsifiability (clears): The claim is straightforwardly falsifiable: pull the ATS report or HR log for the relevant hiring period and check whether the unique applicant count equals 1,200.
      axis Equivocation (weakens): 'Applied' can mean anything from clicking 'Express Interest' to submitting a fully complete application package; the term is not pinned down in the original claim.
      axis Base rate / magnitude (weakens): No context is given about the role, industry, or campaign scope, so 1,200 cannot be evaluated as plausible, inflated, or understated.
      axis Counterexample (weakens): If the ATS counted duplicate submissions, incomplete inquiries, or re-opened applications, the true unique-applicant figure could differ materially from 1,200.
      axis Causality vs correlation (clears): No causal claim is made; this axis does not bear on a straightforward count.
      reason: The claim is a concrete, falsifiable numerical assertion about a recordable event — exactly the kind of claim that can be verified or refuted by pulling a report. It does not collapse into vagueness. However, without a cited source and without a fixed definition of 'applied', the raw assertion is unverified. A narrower surviving claim holds: the figure is checkable, and if an auditable source corroborates 1,200 unique complete submissions, the claim becomes fully substantive. The weaknesses are evidentiary and definitional, not structural.

  [1] claim: The pass rate measured among all 1,200 candidates was 92%.
      verdict: substantive
      axis Evidence (weakens): The claim is a precise quantitative assertion (1,104/1,200 = 92%) that is fully verifiable in principle from administrative records, but no source, audit trail, or independent verification is cited here.
      axis Hidden premise (weakens): The computation assumes all 1,200 registered candidates are the correct denominator — if any were excluded post-registration (withdrew, were disqualified, did not sit), the denominator and resulting rate could differ.
      axis Falsifiability (clears): The claim is straightforwardly falsifiable: access to the candidate roster and score records would either confirm or refute it, making it a well-formed empirical assertion.
      axis Equivocation (clears): 'Pass rate among all 1,200 candidates' could hide ambiguity about whether partial completers or absentees are counted as failures or excluded, but the steelman explicitly resolves this by including all 1,200 in the denominator.
      axis Base rate / magnitude (clears): A specific numerator (1,104), denominator (1,200), and resulting proportion (92%) are all provided, so magnitude is fully specified; no comparison benchmark is given, but none is required for the claim as stated.
      axis Counterexample (clears): No logical counterexample exists to a straightforward arithmetic proportion claim; the only threat is factual inaccuracy in the underlying counts, which is an evidence issue, not a structural flaw.
      axis Causality vs correlation (clears): The claim is purely descriptive and proportional — no causal inference is made, so this axis does not apply.
      reason: The claim is a well-formed, falsifiable arithmetic assertion with a specific numerator, denominator, and defined operation. The only genuine weaknesses are absence of a cited source and the implicit assumption that the denominator is correctly bounded — both addressable by inspection of records, not structural flaws in the claim itself. No equivocation, no causal overreach, no unfalsifiability.

  [2] claim: A pass rate measured among all applicants constitutes the hiring funnel's pass rate.
      verdict: partial
      axis Evidence (clears): The claim is a definitional/mathematical assertion about funnel metrics, not an empirical one, so direct evidence is less relevant than logical coherence; the steelman correctly identifies that cohort-based funnel analytics does use total entrants as denominator.
      axis Hidden premise (weakens): The claim silently assumes that 'all applicants' and 'top-of-funnel entrants' are coextensive—a substantive assumption that fails whenever pre-application sourcing, referrals, or invite-only pipelines exist, meaning the funnel may begin before or above the formal application step.
      axis Falsifiability (weakens): Because the producer has loaded five conditions that together make the claim true by construction, the original unqualified claim ('a pass rate measured among all applicants constitutes the funnel's pass rate') is empirically untestable as stated—it collapses into a tautology under those conditions and is simply false without them.
      axis Equivocation (fatal): 'Pass rate' slides between a within-stage conditional rate (passers / those who reached the stage) and a cumulative funnel conversion rate (passers / all who entered the funnel), and the claim's validity depends entirely on which definition is in use, yet the original claim does not specify.
      axis Base rate / magnitude (clears): Not applicable to a definitional claim about denominators; no magnitude comparison is required.
      axis Counterexample (fatal): A recruiter measuring résumé-screen pass rate only over applicants who were not auto-rejected by ATS (a standard industry practice) produces a within-stage conditional rate over a sub-population, not all applicants—directly falsifying the claim that the all-applicant figure 'constitutes' the funnel pass rate.
      axis Causality vs correlation (clears): The claim is definitional, not causal; this axis does not apply.
      reason: The unqualified original claim is fatally equivocal: 'pass rate' has two standard meanings (within-stage conditional vs. cumulative funnel conversion) and the claim is only true under one of them. A clear counterexample—any funnel with a pre-screen or ATS auto-rejection layer—shows the all-applicant denominator does not automatically equal the funnel's pass rate. A narrower claim survives only when the producer's own definitional conditions (application = funnel entry, cumulative not conditional rate) are granted, but those conditions were stated by the producer, not embedded in the original claim. The narrowing is therefore legitimate (restricted to the producer's stated framing) rather than laundered.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R9

Manifest: class `SCP-1` · anchor `S06` · variant `mutant` · stage `decompose+critic`
Injected: scope narrowed to "the Sydney datacenter" while the claim stays asserted of "our API"; predicate frozen
Narrow scope: "the Sydney datacenter" · Broad subject: "our API"

Input: "The 120-millisecond median response time measured in the Sydney datacenter puts our API's median response time at 120-millisecond."

Diagnostics (not recall): axis-fired=true · decompose cardinality=3

Full chain:

  [0] claim: The Sydney datacenter measured a median response time of 120 milliseconds.
      verdict: partial
      axis Evidence (weakens): The claim asserts a specific numeric measurement (120ms median) but cites no source, dataset, monitoring tool, or methodology — it is bare assertion.
      axis Hidden premise (weakens): The claim assumes 'response time' has a single, agreed-upon definition (e.g., TTFB vs. full payload vs. round-trip from a specific client), but this is unstated and non-trivial.
      axis Falsifiability (clears): The claim is in principle falsifiable: re-running the same measurement with documented methodology against the Sydney datacenter could confirm or refute 120ms; no structural barrier to testing exists.
      axis Equivocation (weakens): 'Response time' is a buzzword that can mean TTFB, full payload, application-layer latency, or round-trip from an unspecified origin — the term does real hiding work here.
      axis Base rate / magnitude (weakens): 120ms is a concrete number, but without a comparison baseline (e.g., SLA target, industry norm, or other datacenters) it provides a direction but no context for whether it is good, bad, or typical.
      axis Counterexample (weakens): A measurement taken during off-peak hours from a co-located client would trivially yield a different median, showing the figure is not stable without scope constraints.
      axis Causality vs correlation (clears): No causal claim is made; the claim is purely descriptive/observational, so this axis does not apply.
      reason: The core numeric claim is falsifiable and specific enough to survive scrutiny, but it is weakened by undefined measurement methodology, undefined client geography, and lack of cited evidence. The narrowing here is legitimate: the original claim implies a real measurement event, and restricting it to 'a specific documented measurement run' does not introduce new conditions so much as clarify what any credible interpretation of the claim already requires. Without source documentation and a definition of 'response time', the 120ms figure cannot be validated or generalized, but it is not hollow — it makes a testable assertion about a concrete datacenter at a concrete time.

  [1] claim: The API's median response time is 120 milliseconds.
      verdict: partial
      axis Evidence (weakens): The claim presents a precise numeric figure (120ms) but cites no source, measurement methodology, sample size, or observation window — it is a bare assertion of a specific statistic.
      axis Hidden premise (weakens): The claim silently assumes 'median' means p50 latency measured consistently at a defined boundary under representative load — none of which is stated, and any deviation (e.g., mean mislabeled as median, server-side vs. client-side measurement) would invalidate the figure.
      axis Falsifiability (clears): The claim is falsifiable in principle: re-running the p50 measurement on production traffic with a sufficiently large sample would either confirm or refute 120ms, making this empirically testable.
      axis Equivocation (weakens): 'Median response time' could be confused with mean, mode, or a percentile other than p50, and 'response time' could mean server-side processing time, client-perceived round-trip, or time-to-first-byte — the term is underspecified but not fundamentally equivocal.
      axis Base rate / magnitude (weakens): 120ms is a concrete number, but without context (what type of API, what endpoints, what load level) there is no way to assess whether it is plausible, impressive, or alarming relative to any baseline.
      axis Counterexample (weakens): The figure would fail immediately if measured during low-traffic periods, on a subset of fast endpoints only, or using mean rather than median — all common measurement errors that could produce a misleadingly low number.
      axis Causality vs correlation (clears): No causal claim is made; this is a descriptive statistic, so this axis does not apply.
      reason: The claim asserts a precise, falsifiable numeric statistic, which gives it a defensible empirical core. However, as stated it is a bare assertion lacking cited evidence, a defined measurement boundary, and confirmation of representative sampling. Stripping those gaps, a narrower claim survives: the figure is meaningful if and only if the measurement conditions the producer steelman describes are actually met. The narrowing here is legitimate — it restricts to conditions the producer explicitly stated, not conditions introduced by the critic to rescue the claim.

  [2] claim: The API's median response time measurement is based on data from the Sydney datacenter.
      verdict: hollow
      axis Evidence (fatal): No citation, documentation, or source is provided to support the claim that Sydney is the designated benchmarking location; it is bare assertion.
      axis Hidden premise (weakens): The claim assumes the API provider has made an explicit, documented methodological choice to use Sydney as the canonical measurement point, which is an entirely unstated and unverified organizational decision.
      axis Falsifiability (weakens): The claim is falsifiable in principle — one could inspect the provider's documentation or measurement methodology — but no such inspection is referenced, leaving it currently untestable as asserted.
      axis Equivocation (weakens): 'Median response time' could mean client-perceived latency, server-side processing time, or a composite metric, and 'based on data from Sydney' could mean measured at Sydney, routed through Sydney, or benchmarked against Sydney — these distinctions are unexplored.
      axis Base rate / magnitude (weakens): No data is offered showing Sydney's latency distribution actually approximates a global median, nor is any comparison with other datacenters provided to justify its representative status.
      axis Counterexample (weakens): Many API providers compute median response time as an aggregate across all regions or from their primary (often US-based) datacenter, making Sydney-as-canonical an atypical and easily falsifiable choice.
      axis Causality vs correlation (weakens): No causal or methodological chain is established between Sydney's operational characteristics and its designation as the measurement source; the steelman's reasoning is speculative.
      reason: The claim is a bare factual assertion about a specific, verifiable organizational and methodological decision, but zero evidence is provided. The steelman elaborates plausible reasons Sydney could be chosen, but plausibility is not evidence. Every supporting axis — evidence, hidden premise, falsifiability, equivocation, base rate, and counterexample — fires against it. No narrowed version of the claim survives without introducing conditions the speaker never stated, which would constitute laundering rather than legitimate narrowing.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R10

Manifest: class `SCP-1` · anchor `S06` · variant `clean` · stage `decompose+critic`
**Clean twin — no defect injected. A catch here is an over-catch (A1.2).**

Input: "The 120-millisecond median response time measured in all six datacenters puts our API's median response time at 120-millisecond."

Diagnostics (not recall): axis-fired=true · decompose cardinality=3

Full chain:

  [0] claim: The API's median response time is 120 milliseconds.
      verdict: partial
      axis Evidence (weakens): The claim asserts a precise numeric figure (120ms) but cites no measurement data, methodology documentation, or source — it is a bare assertion of a statistic.
      axis Hidden premise (weakens): The claim assumes a stable, well-characterized latency distribution exists; if the distribution is multimodal or highly variable (e.g., bimodal due to cache hits vs. misses), the median alone is misleading without that context.
      axis Falsifiability (clears): The claim is eminently falsifiable: collect a sufficiently large sample of API response times and compute the 50th percentile — if it differs materially from 120ms, the claim is false.
      axis Equivocation (weakens): 'Response time' is ambiguous — it can mean server-side processing time, end-to-end latency including network transit, or time-to-first-byte — without fixing this, the number is uninterpretable and the claim shifts meaning depending on context.
      axis Base rate / magnitude (weakens): 120ms is a specific, quantified figure, but no comparison is offered (e.g., vs. SLA target, industry benchmark, or prior period) to establish whether this is good, bad, or typical for this type of API.
      axis Counterexample (weakens): Under high-load conditions, off-hours, or different request-type mixes, the measured median could shift substantially — the claim implicitly assumes a fixed operating context that may not hold.
      axis Causality vs correlation (clears): No causal inference is made; this is a descriptive statistic, so this axis does not bear on the claim.
      reason: The core claim is falsifiable and numerically specific, which is a genuine epistemic virtue. However, the equivocation on 'response time' is a real defect — the number is unverifiable without knowing what is being measured. The claim survives in a narrowed form only if the measurement boundary is fixed. The narrowing condition ('defined methodology') partially reflects the producer's own steelman conditions, so it is not fully laundered, but it is one step removed from the original bare claim.

  [1] claim: The median response time was measured across six datacenters.
      verdict: partial
      axis Evidence (weakens): The claim asserts a specific measurement was performed but cites no data, study, report, or verifiable source to confirm it actually happened.
      axis Hidden premise (weakens): The claim assumes 'measured across six datacenters' means a methodologically consistent, comparable measurement at each site, which is unstated and non-trivial to guarantee.
      axis Falsifiability (clears): The claim is falsifiable in principle: one could inspect the measurement logs or methodology to confirm or deny that exactly six datacenters were included and that median was computed correctly.
      axis Equivocation (weakens): 'Measured across' is slightly ambiguous — it could mean one aggregate measurement spanning six sites or six separate measurements — but this is a minor ambiguity, not a fatal shift in meaning.
      axis Base rate / magnitude (clears): The claim makes no assertion about the magnitude or significance of the result, only about the methodology scope, so no base-rate comparison is required.
      axis Counterexample (clears): No logical counterexample defeats the claim; it is a factual assertion about a procedure, not a general rule.
      axis Causality vs correlation (clears): No causal claim is made; the statement is purely descriptive of a measurement procedure.
      reason: The claim is a straightforward empirical assertion about a measurement procedure that is falsifiable and free of equivocation or causal overreach. Its weaknesses are absence of cited evidence and mild ambiguity in 'measured across.' The steelman's producer conditions are reasonable elaborations of what the claim implicitly requires, not external laundering. The surviving claim narrows only to conditions the producer explicitly stated, so no condition laundering occurs. It does not rise to 'substantive' solely because no evidence is cited or available in the claim itself.

  [2] claim: The 120-millisecond figure is derived from measurements taken in all six datacenters.
      verdict: hollow
      axis Evidence (fatal): The claim that the figure is 'derived from measurements taken in all six datacenters' is a bare assertion — no methodology document, instrumentation log, or data provenance record is cited to confirm this.
      axis Hidden premise (weakens): The claim silently assumes that 'all six datacenters' were simultaneously active, instrumented, and equally weighted in the aggregation, none of which is established.
      axis Falsifiability (weakens): The claim is in principle falsifiable — one could inspect instrumentation logs to verify whether data from all six sites contributed — but without access to those records it currently functions as an untestable assertion.
      axis Equivocation (weakens): 'Measurements taken in all six datacenters' could mean per-site instrumentation or simply a central collector receiving traffic from all six, two very different things that the claim does not distinguish.
      axis Base rate / magnitude (weakens): The claim makes no statement about what statistical summary (mean, median, p99) the 120ms figure represents, nor how the six-site data was aggregated, leaving the number's meaning unclear.
      axis Counterexample (weakens): If any one of the six datacenters was offline or excluded during the measurement window, the claim fails entirely, and there is no evidence ruling this out.
      axis Causality vs correlation (clears): No causal inference is asserted here; this axis does not materially bear on the claim.
      reason: The claim rests entirely on an unverified assertion about data provenance. No documentation, instrumentation record, or methodology is cited. The key term 'measurements taken in' equivocates between per-site instrumentation and central aggregation. The claim is in principle falsifiable but currently unfalsifiable in practice due to absent evidence. The steelman conditions the producer supplies are precisely what would need to be verified — their enumeration underscores, rather than resolves, the evidential gap. No narrower version of the claim survives without introducing conditions the speaker never established.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R11

Manifest: class `MFC` · anchor `M07` · variant `mutant` · stage `decompose+critic`
Injected: kill condition deleted

Input: "Sharding the queue raises sustained throughput on the worker pool."

Diagnostics (not recall): axis-fired=true · decompose cardinality=1

Full chain:

  [0] claim: Sharding the queue raises sustained throughput on the worker pool.
      verdict: partial
      axis Evidence (clears): The steelman cites Disruptor, Kafka consumer groups, and partitioned thread-pool designs as empirical support for 2–10× gains, which are real systems with documented benchmarks, making this more than bare assertion.
      axis Hidden premise (weakens): The claim assumes lock contention on the queue is the actual throughput ceiling; if the bottleneck is downstream processing, I/O, or memory bandwidth, sharding the queue yields no sustained throughput gain.
      axis Falsifiability (clears): The claim is falsifiable: measure sustained task completion rate before and after sharding under high worker counts with a contention-bound workload; a failure to increase throughput refutes it.
      axis Equivocation (clears): 'Sustained throughput' could mean aggregate task completion rate or per-worker rate, but the steelman consistently uses the aggregate sense, so no meaningful equivocation is present.
      axis Base rate / magnitude (weakens): The 2–10× figure is plausible but wide; the actual gain is highly sensitive to worker count, task granularity, and contention severity, so the range is more illustrative than predictive.
      axis Counterexample (weakens): If tasks are coarse-grained (long processing time relative to dequeue time), sharding adds structural complexity with negligible throughput benefit because queue contention was never the binding constraint.
      axis Causality vs correlation (clears): The mechanistic argument (lock serialization → contention → throughput ceiling → sharding removes serialization) is causal and structurally sound, not merely correlational.
      reason: The core mechanism is sound and empirically grounded, but the unqualified claim fails when the queue is not the bottleneck or tasks are coarse-grained. The two conditions retained in the surviving claim were explicitly stated by the producer, not introduced by the critic, so this is legitimate narrowing rather than condition laundering.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R12

Manifest: class `MFC` · anchor `M07` · variant `clean` · stage `decompose+critic`
**Clean twin — no defect injected. A catch here is an over-catch (A1.2).**

Input: "Sharding the queue raises sustained throughput on the worker pool by at least 35%, and we abandon the approach if the measured gain is under 35%."

Diagnostics (not recall): axis-fired=true · decompose cardinality=2

Full chain:

  [0] claim: Sharding the queue raises sustained throughput on the worker pool by at least 35%.
      verdict: partial
      axis Evidence (weakens): The steelman offers a plausible analytic model and cites lock-convoy research directionally, but no empirical study or benchmark is cited that actually demonstrates ≥35% throughput gain specifically under these threshold conditions; the arithmetic illustration is model-internal, not externally validated.
      axis Hidden premise (weakens): The argument assumes freed stall time is immediately and fully converted to productive dispatch work, which requires that the CPU pipeline, memory subsystem, and scheduler are all idle during stall periods and can absorb the released capacity without introducing new latency—a non-trivial assumption not guaranteed by the stated conditions.
      axis Falsifiability (clears): The claim is genuinely falsifiable: run a controlled benchmark satisfying all stated thresholds, measure sustained task dispatch rate before and after sharding, and check whether the gain is ≥35%; this is a concrete, observable test.
      axis Equivocation (weakens): 'Roughly even load distribution' and 'downstream headroom' are imprecise enough to be interpreted post-hoc to explain away any shortfall, but the producer conditions do attempt to quantify them (≤2:1 skew, genuine spare capacity), partially resolving the ambiguity.
      axis Base rate / magnitude (weakens): The ≥35% magnitude claim rests on a specific stall-elimination efficiency assumption (80% of stall removed) that is asserted without empirical backing; actual lock-elimination efficiency varies widely in practice and can be substantially lower due to cache-coherence traffic and false sharing introduced by sharding itself.
      axis Counterexample (weakens): A system where 25% stall is lock contention but 60% of wall-clock is I/O wait (still satisfying the ≥25% contention threshold) would see minimal throughput gain from sharding, violating the ≥35% prediction while technically meeting all stated conditions.
      axis Causality vs correlation (clears): Producer condition 1 explicitly requires a causal confirmation test (lock-removed benchmark or contention injection), which is methodologically sound and elevates the contention attribution from correlation to causation—this axis is largely addressed.
      reason: The core mechanism is sound and the claim is falsifiable, but the precise ≥35% magnitude threshold is fragile: it depends on an 80% stall-elimination efficiency figure that is asserted rather than evidenced, and the counterexample of a mixed I/O/lock-stall system shows the stated conditions do not fully exclude cases where the gain falls short. Stripping the hard numerical floor leaves a defensible directional claim with the producer's own qualifying conditions folded in. The added conditions in the surviving claim (CPU/lock-bound path, ≤2:1 skew, genuine sustained headroom, causal validation) were all stated or clearly implied by the producer, so no laundering is introduced.

  [1] claim: The approach is abandoned if the measured throughput gain is under 35%.
      verdict: partial
      axis Evidence (weakens): No evidence is provided that the 35% threshold was derived from empirical cost-benefit analysis rather than chosen arbitrarily; the steelman asserts principled derivation but supplies no data.
      axis Hidden premise (weakens): The claim assumes that a 35% benchmark gain reliably predicts net production benefit after integration overhead, variance, and regression risk—a non-trivial translation that is simply asserted, not demonstrated.
      axis Falsifiability (clears): As a procedural rule ('the approach is abandoned if X'), the claim is falsifiable in principle: one could observe whether an approach yielding under 35% gain was in fact abandoned, making it testable against process records.
      axis Equivocation (weakens): 'Throughput' is unspecified—it could mean transactions per second, data volume, latency-adjusted throughput, or other metrics—allowing the threshold to shift interpretation depending on which measure is convenient.
      axis Base rate / magnitude (weakens): The claim states a specific magnitude (35%) but offers no baseline comparison to industry norms, prior system performance, or the distribution of gains typically seen in similar optimization work, so the threshold's appropriateness is unanchored.
      axis Counterexample (weakens): A 34% throughput gain that also eliminates a critical bottleneck or reduces tail latency dramatically could rationally justify continuation, showing the single-threshold rule is not universally sound.
      axis Causality vs correlation (weakens): No causal mechanism links a sub-35% benchmark result to net negative or negligible production value; the producer asserts this relationship but does not establish it.
      reason: The claim has a falsifiable procedural core—a decision rule tied to a numeric threshold—that survives scrutiny as a process description. However, the justification that 35% is the correct or principled threshold is unsupported, the term 'throughput' equivocates, and the single-metric rule is vulnerable to counterexample. The surviving claim is narrowed to the procedural fact of the rule's existence and application, stripped of the normative claim that 35% is the right threshold for the stated reasons.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R13

Manifest: class `MFC` · anchor `M02` · variant `mutant` · stage `decompose+critic`
Injected: kill condition deleted

Input: "The new index cuts p95 query time on the production shard."

Diagnostics (not recall): axis-fired=true · decompose cardinality=1

Full chain:

  [0] claim: The new index cuts p95 query time on the production shard.
      verdict: partial
      axis Evidence (weakens): The claim is asserted without citing any benchmark, EXPLAIN plan, or before/after measurement from the actual production shard; the steelman references 'analogous workloads' but that is not evidence for this specific system.
      axis Hidden premise (weakens): The claim assumes p95 latency is currently driven by the scan pattern the index targets, rather than by lock contention, network jitter, or application-layer bottlenecks — an untested assumption about where tail latency actually originates.
      axis Falsifiability (clears): The core claim is concretely falsifiable: measure p95 query time on the production shard before and after index creation under comparable load; if p95 does not decrease, the claim is false.
      axis Equivocation (clears): 'Cuts' is directionally unambiguous and 'p95 query time' is a well-defined statistic, so no meaningful term-shifting is present.
      axis Base rate / magnitude (weakens): No magnitude or comparison baseline is given for the specific shard; citing 40–80% from analogous workloads is suggestive but not a measurement of this system.
      axis Counterexample (weakens): If the production shard's p95 is driven by write-heavy lock contention or if the query planner ignores the index due to stale statistics, the new index could fail to reduce or even worsen p95 — a realistic failure mode the producer's own conditions acknowledge.
      axis Causality vs correlation (clears): The causal mechanism (index eliminates scan overhead → lower I/O → lower tail latency) is theoretically sound for the stipulated conditions, so this is not a mere correlation claim.
      reason: The causal logic is sound and the claim is falsifiable, but it currently rests on unverified assumptions about what drives p95 on this specific shard and lacks any before/after measurement. Stripping the unsupported certainty leaves a conditional but defensible prediction. The surviving conditions are largely those the producer explicitly stated, not laundered additions, so the narrowing is legitimate rather than evasive.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R14

Manifest: class `MFC` · anchor `M02` · variant `clean` · stage `decompose+critic`
**Clean twin — no defect injected. A catch here is an over-catch (A1.2).**

Input: "The new index cuts p95 query time on the production shard by at least 25%, and we abandon the approach if the measured gain is under 25%."

Diagnostics (not recall): axis-fired=true · decompose cardinality=2

Full chain:

  [0] claim: The new index cuts p95 query time on the production shard by at least 25%
      verdict: partial
      axis Evidence (weakens): The steelman cites 'standard database benchmarking literature and common operational experience' but names no specific study, benchmark, or measurement from this actual system — the 25% figure is asserted without pre/post data from the production shard.
      axis Falsifiability (clears): The claim is straightforwardly falsifiable: measure p95 query latency on the production shard before and after index deployment under representative load and check whether the reduction meets the 25% threshold.
      axis Hidden premise (weakens): The claim assumes the query planner will consistently choose the new index for the specific query shapes constituting the p95 tail — a non-trivial assumption that can fail due to stale statistics, parameter sniffing, or planner conservatism.
      axis Equivocation (clears): 'P95 query time' is unambiguous and 'production shard' is specific; no meaningful term shift is present.
      axis Base rate / magnitude (weakens): The 25% threshold is stated as a concrete, measurable quantity with a directional comparison, and the steelman's cited range (40–70% in analogous cases) gives a plausible magnitude anchor, though no shard-specific baseline latency or row-count data is provided.
      axis Counterexample (weakens): If p95 latency is bottlenecked by lock contention, network round-trips, or write amplification from index maintenance rather than scan cost, the index could provide zero or negative improvement — a realistic failure mode explicitly acknowledged in the producer's own conditions.
      axis Causality vs correlation (clears): The mechanism (index seek replacing scan → fewer rows examined → lower latency) is a well-understood causal chain, not a mere correlation, so no causality problem exists if the planner actually uses the index.
      reason: The claim is falsifiable and the causal mechanism is sound, but no shard-specific pre/post evidence is yet presented, and multiple realistic failure modes (planner choice, contention, write amplification) could prevent the threshold from being reached. The producer's own conditions acknowledge these risks. The core directional claim survives but cannot be confirmed without measurement; the 25% floor is plausible but unverified. Conditions (a) and (b) in the surviving claim were explicitly stated by the producer, so this is legitimate narrowing rather than laundering.

  [1] claim: The approach is abandoned if the measured gain is under 25%
      verdict: partial
      axis Evidence (weakens): The 30–50% lab-to-production erosion figure is asserted without citation; no empirical basis is provided for why 25% specifically is the right cutoff rather than 20% or 30%.
      axis Hidden premise (weakens): The rule assumes the metric being measured is the correct proxy for real-world value, and that a single threshold applies uniformly regardless of domain, cost of alternatives, or strategic context.
      axis Falsifiability (clears): As stated ('the approach is abandoned if measured gain is under 25%'), the claim is a concrete, testable behavioral rule: one can observe whether an approach with a sub-25% measured gain was in fact halted, making it falsifiable.
      axis Equivocation (weakens): 'Gain' is undefined — gain in what metric, measured how, over what baseline — and 'abandoned' could range from formal closure to informal deprioritization, creating ambiguity that undermines enforceability.
      axis Base rate / magnitude (weakens): No comparison class is offered: what fraction of approaches actually fall below 25%, what is the distribution of real gains in this domain, and whether 25% represents a meaningful upper tail or median are all unaddressed.
      axis Counterexample (weakens): An approach yielding 22% gain in accuracy but 40% gain in cost-efficiency, or one that is the only viable option in a constrained environment, would likely not be abandoned despite the rule — showing the threshold is not truly unconditional.
      axis Causality vs correlation (weakens): The steelman argues sub-25% lab gains cause near-zero real-world outcomes via the erosion mechanism, but this causal chain is asserted, not demonstrated for the specific domain.
      reason: The core behavioral claim — a bright-line stopping rule at 25% — is falsifiable and operationally coherent once key terms are pinned down. However, the justification for 25% specifically lacks evidence, the metric and 'abandoned' definitions are equivocal, and obvious counterexamples (multi-metric tradeoffs, constrained alternatives) show the rule is not unconditional as stated. A narrower claim survives: the rule holds as a documented, pre-registered, single-metric policy with adequate statistical power, but the strong framing ('ensures resources are only committed to meaningful improvements') and the specific 25% figure remain insufficiently supported.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R15

Manifest: class `MFC` · anchor `M08` · variant `mutant` · stage `decompose+critic`
Injected: kill condition deleted

Input: "The new allocator reduces GC pause time on the render service."

Diagnostics (not recall): axis-fired=true · decompose cardinality=1

Full chain:

  [0] claim: The new allocator reduces GC pause time on the render service.
      verdict: partial
      axis Evidence (weakens): The steelman provides a mechanistic rationale and cites relevant techniques (TLABs, size-segregated free lists), but no actual benchmark data, production metrics, or measurements are presented — it remains a plausibility argument, not demonstrated fact.
      axis Hidden premise (weakens): The claim assumes that heap fragmentation and TLAB design are the binding constraints on GC pause time for this specific service; if pauses are dominated by other factors (e.g., reference scanning, finalization, write barriers), the allocator change may have negligible effect.
      axis Falsifiability (clears): The claim is falsifiable: P50/P95/P99 GC pause distributions measured under A/B conditions (old vs. new allocator, same service, same load) would confirm or refute it, and the producer conditions explicitly describe this test — so the claim is not vacuous.
      axis Equivocation (weakens): 'Reduces GC pause time' could mean any of frequency, mean duration, tail latency, or total pause budget — without specifying which metric, the claim is slightly slippery but not fatally so given common industry usage of P-percentile pause metrics.
      axis Base rate / magnitude (weakens): No magnitude or baseline is given — 'reduces' could mean 1% or 50%; without a claimed effect size or comparison baseline, the practical significance is unassessable.
      axis Counterexample (weakens): GC pauses on allocation-heavy services can be dominated by concurrent marking or card-table scanning rather than allocation rate, in which case allocator improvements do not reduce pause time — a real class of cases where this fails.
      axis Causality vs correlation (weakens): The claim asserts causation (the allocator causes the reduction), and the producer conditions properly call for controlled isolation of the variable; as stated without data the causal link is asserted rather than established.
      reason: The claim is falsifiable and the mechanistic rationale is coherent, but it currently rests on assertion rather than presented evidence. Multiple axes weaken it — no data, no magnitude, no baseline, potential confounders — but none are fatal because the producer conditions already specify the exact tests needed to substantiate it. The surviving claim is the forward-looking, testable version: the allocator is expected to reduce pauses, and this can and should be confirmed by the described measurements. The narrowing to 'expected, pending measurement' is legitimate because the original claim is an engineering prediction about a deployed change, not a post-hoc report of measured results.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R16

Manifest: class `MFC` · anchor `M08` · variant `clean` · stage `decompose+critic`
**Clean twin — no defect injected. A catch here is an over-catch (A1.2).**

Input: "The new allocator reduces GC pause time on the render service by at least 45%, and we abandon the approach if the measured gain is under 45%."

Diagnostics (not recall): axis-fired=true · decompose cardinality=2

Full chain:

  [0] claim: The new allocator reduces GC pause time on the render service by at least 45%.
      verdict: partial
      axis Evidence (weakens): The claim asserts a precise quantitative threshold (≥45%) but no benchmark data, measurement methodology, or source is cited; the steelman offers only a plausibility argument, not evidence.
      axis Hidden premise (weakens): The claim assumes the render service's GC pauses are primarily driven by allocation-side pathologies (fragmentation, promotion failures) rather than reference-graph complexity or finalizer pressure—an untested premise that could fully negate the benefit.
      axis Falsifiability (clears): The claim is falsifiable in principle: a controlled A/B benchmark on the render service measuring p99 or mean GC pause under matched load would confirm or refute it, and the producer's own conditions describe exactly such a test.
      axis Equivocation (weakens): 'GC pause time' could mean mean, p50, p99, or p999 wall-clock duration, and the choice materially changes the reported reduction; the claim does not pin this down.
      axis Base rate / magnitude (weakens): 45% is a specific, testable magnitude with a clear direction, but without a baseline figure (e.g., current p99 pause of X ms) the practical significance cannot be assessed.
      axis Counterexample (weakens): If the render service's GC pauses are dominated by reference traversal depth or large retained object graphs rather than allocation churn, an allocator change would yield little improvement—a plausible failure case given typical rendering workloads.
      axis Causality vs correlation (clears): The mechanism (allocator → heap locality → shorter GC pauses) is causally coherent and the producer's conditions include experimental controls (same JVM flags, heap size, runtime version), so this axis is adequately addressed if those controls are honored.
      reason: The claim is falsifiable and mechanistically coherent, but the precise 45% figure is unsubstantiated by cited evidence, the key metric (mean vs. p99) is unspecified, and the benefit depends on an unverified assumption about the dominant cause of GC pauses. Stripping the unsupported precision and flagging the metric ambiguity leaves a narrower but defensible claim: the threshold and direction survive if the producer's measurement conditions are actually met. One condition ('pauses primarily caused by allocation-side fragmentation') was implied by the steelman rather than stated in the original claim, but it is logically required by the mechanism and not introduced solely to dodge a counterexample.

  [1] claim: The approach will be abandoned if the measured gain is under 45%.
      verdict: hollow
      axis Evidence (fatal): No documentation, pre-registration record, or contractual instrument is cited to confirm the decision rule exists; this is a bare assertion of a binding commitment.
      axis Hidden premise (weakens): The claim silently assumes that the parties with authority to continue or terminate are actually bound by this rule and have no override capacity — an assumption that is rarely true in practice and is unverified here.
      axis Falsifiability (weakens): The claim is in principle falsifiable — observe a sub-45% measured gain and check whether abandonment occurs — but only if 'measured gain' and 'abandoned' are unambiguously defined, which they are not confirmed to be.
      axis Equivocation (weakens): 'Abandoned' is undefined: it could mean formal termination, indefinite pause, rebranding, or a mere statement of intent, each carrying very different consequences and verifiability.
      axis Base rate / magnitude (weakens): The 45% threshold is stated as a number, which is concrete, but no baseline, metric definition, or derivation is provided to assess whether it is a meaningful or arbitrary cutoff.
      axis Counterexample (weakens): Organizational decision rules routinely fail to bind when results are inconvenient; a sub-45% result could be met with metric redefinition, scope narrowing, or senior override rather than abandonment.
      axis Causality vs correlation (clears): The claim is normative/procedural rather than causal-correlational, so this axis does not materially apply.
      reason: The claim asserts a binding, enforceable stopping rule but provides zero evidence that any such rule has been formally documented, that 'abandoned' has a concrete meaning, or that decision-makers are actually constrained by it. Every condition required to make the claim substantive — pre-registration, metric definition, measurement integrity, override prohibition, concrete abandonment definition — is supplied by the steelman rather than the original claim, meaning the claim survives only through conditions the speaker never stated. Stripped of those laundered conditions, the original assertion is an unverifiable promise about future organizational behavior with no evidentiary anchor. Survives only by conditions the speaker never stated.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R17

Manifest: class `MFC` · anchor `M03` · variant `mutant` · stage `decompose+critic`
Injected: kill condition deleted

Input: "Batching writes raises throughput on the ingest path."

Diagnostics (not recall): axis-fired=true · decompose cardinality=1

Full chain:

  [0] claim: Batching writes raises throughput on the ingest path.
      verdict: partial
      axis Evidence (clears): The mechanism is well-established in systems literature and empirically documented across Kafka, RocksDB, Linux I/O, and TCP stacks—this is not bare assertion.
      axis Hidden premise (weakens): The claim assumes that the ingest path's bottleneck is in per-write fixed overhead; if the bottleneck is CPU serialization cost or downstream consumer throughput, batching on the ingest path may not raise end-to-end throughput.
      axis Falsifiability (clears): The claim is falsifiable: a benchmark on a real ingest path (e.g., Kafka producer with batch.size=1 vs. batch.size=65536) would show measurable records-per-second differences, and the claim would be false if throughput did not increase.
      axis Equivocation (weakens): 'Throughput' could mean ingest-path throughput (records written per second at the producer/broker) or end-to-end pipeline throughput; the claim implicitly restricts to the ingest path, which is defensible but should be noted.
      axis Base rate / magnitude (weakens): The steelman cites 'orders of magnitude' for sequential vs. random writes on SSDs/logs, but the actual throughput gain from batching alone varies widely (2x–100x) depending on hardware and workload; no single magnitude is universally true.
      axis Counterexample (weakens): At very low ingest rates, batches time-out before filling, so per-record overhead savings are minimal and throughput gain is negligible or zero—producer condition 4 acknowledges this but it limits the claim's generality.
      axis Causality vs correlation (clears): The causal mechanism (amortizing fixed costs over more records) is well-specified and directionally correct, not merely correlational.
      reason: The core causal claim is well-evidenced and falsifiable, but it fails at very low arrival rates and when the bottleneck lies outside the ingest path. The surviving claim restricts to conditions the producer explicitly stated, making this legitimate narrowing rather than laundering.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R18

Manifest: class `MFC` · anchor `M03` · variant `clean` · stage `decompose+critic`
**Clean twin — no defect injected. A catch here is an over-catch (A1.2).**

Input: "Batching writes raises throughput on the ingest path by at least 40%, and we abandon the approach if the measured gain is under 40%."

Diagnostics (not recall): axis-fired=true · decompose cardinality=2

Full chain:

  [0] claim: Batching writes raises throughput on the ingest path by at least 40%.
      verdict: partial
      axis Evidence (weakens): The steelman cites real, named systems (Kafka, PostgreSQL WAL, RocksDB, DynamoDB/Bigtable) with plausible ranges, but the specific 40% floor is not directly sourced — the cited figures range from 2× to 16×, and the 40% is inferred as a 'conservative lower bound', not a measured minimum across a defined population of workloads.
      axis Hidden premise (weakens): The claim implicitly assumes the baseline system is not already batching internally (e.g., OS page cache coalescing, driver-level buffering), so that user-space batching adds a genuinely new amortization layer; if the storage layer already batches, the marginal gain may be far smaller.
      axis Falsifiability (fatal): The claim is testable in principle — run a controlled benchmark comparing single-write vs. batched-write throughput on a specific ingest path — but the producer's six conditions so tightly scope the domain that any failure case can be dismissed as out-of-scope, making the universal '≥40%' claim effectively unfalsifiable as stated.
      axis Equivocation (weakens): 'Throughput on the ingest path' is ambiguous: it could mean records/sec at the producer, bytes/sec at the storage layer, or end-to-end pipeline throughput, and the steelman itself acknowledges the metric shifts depending on where the new bottleneck lands after batching.
      axis Base rate / magnitude (weakens): The cited systems show gains from 40% to 1600%, which is an enormous range; asserting 40% as a universal lower bound requires knowing the distribution of real-world workloads, which is not provided, so the floor is an assertion rather than a derived statistic.
      axis Counterexample (fatal): For large-record, CPU-bound, or already-coalesced workloads (e.g., writing 10 MB blobs sequentially to a local NVMe SSD without fsync), batching adds negligible throughput and may even hurt latency without meaningful throughput gain, directly refuting the universal ≥40% claim.
      axis Causality vs correlation (clears): The causal mechanism (amortizing fixed per-op costs) is well-reasoned and supported by systems literature, so this axis does not weaken the directional claim.
      reason: Two fatal issues prevent the universal claim from standing: (1) the six producer conditions make the original unqualified claim unfalsifiable in practice, and (2) obvious counterexamples (large records, CPU-bound paths, already-coalescing storage) break the universal floor. However, the directional claim and the 40% threshold are defensible within the conditions the producer explicitly stated, supported by named empirical benchmarks, and causally grounded in well-understood systems mechanics. Stripping the universality leaves a narrower but substantive claim. The surviving claim reuses only conditions the producer explicitly stated — no new conditions are introduced.

  [1] claim: The approach is abandoned if the measured gain is under 40%.
      verdict: partial
      axis Evidence (weakens): No citation, empirical basis, or domain context is provided to justify why 40% specifically is the threshold; the steelman asserts it is principled but offers only post-hoc rationalization.
      axis Hidden premise (weakens): The claim assumes the baseline, metric definition, and measurement reliability are all fixed and uncontested, but none of those preconditions are established in the original claim.
      axis Falsifiability (weakens): The stopping rule itself is falsifiable in principle—one could observe whether abandonment is triggered at sub-40% gain—but the claim is vacuous without specifying gain of what, measured how, over what period, making it untestable as stated.
      axis Equivocation (fatal): 'Gain' is undefined: it could mean revenue, efficiency, accuracy, speed, or any other metric, and shifting its meaning makes the 40% figure meaningless without anchoring.
      axis Base rate / magnitude (weakens): 40% is asserted as meaningfully high enough to filter noise and low enough to be achievable, but no domain comparison or cost-benefit data is offered to validate this specific number.
      axis Counterexample (weakens): In domains with very high switching costs or irreversibility (e.g., infrastructure decisions), a 39% gain might still warrant adoption over abandonment, breaking the binary rule.
      axis Causality vs correlation (clears): No causal claim is being made here; this axis does not materially bear on a decision-rule claim.
      reason: The fatal equivocation on 'gain' prevents the claim from standing as stated. Once 'gain' is anchored to a specific, pre-defined metric, the core idea—that a concrete numeric threshold can serve as a stopping rule—is falsifiable and defensible in principle. However, the 40% figure itself has no evidential support and cannot be rescued from that; the surviving claim strips it to the structural rule while flagging the threshold as unvalidated. The surviving claim does not introduce laundered conditions beyond what the producer's own steelman conditions already stated, so this is legitimate narrowing.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R19

Manifest: class `SCP-1` · anchor `S07` · variant `mutant` · stage `decompose+critic`
Injected: scope narrowed to "the pilot team" while the claim stays asserted of "our tooling"; predicate frozen
Narrow scope: "the pilot team" · Broad subject: "our tooling"

Input: "The 64% adoption rate measured on the pilot team puts our tooling's adoption rate at 64%."

Diagnostics (not recall): axis-fired=true · decompose cardinality=1

Full chain:

  [0] claim: The pilot team had a measured tooling adoption rate of 64%.
      verdict: partial
      axis Evidence (weakens): The claim asserts a precise figure (64%) but provides no citation, log excerpt, methodology document, or any external corroboration — it is bare assertion dressed in specificity.
      axis Hidden premise (weakens): The claim assumes 'adoption' was defined consistently and objectively before measurement; if adoption was defined post-hoc or loosely, the 64% figure is arbitrary.
      axis Falsifiability (clears): The claim is falsifiable in principle: one could inspect the usage logs or survey data and verify whether ~64% of the fixed denominator met the adoption threshold during the stated window.
      axis Equivocation (weakens): 'Adoption' is a well-known chameleon term — it can mean installation, first use, regular use, or workflow integration — and without a fixed definition the number is uninterpretable.
      axis Base rate / magnitude (weakens): No comparison is offered: 64% relative to industry norms, prior cohorts, or a target rate is unknown, so whether this figure is good, bad, or typical cannot be assessed.
      axis Counterexample (weakens): If the denominator excluded non-compliant or absent members, or if adoption was measured only during the peak enthusiasm period, the same data could yield very different figures, undermining the claim's generalizability.
      axis Causality vs correlation (clears): The claim is purely descriptive/observational, not causal, so this axis does not materially bear on it.
      reason: The precise figure is specific enough to be falsifiable and plausible as a real measurement outcome, so the claim is not hollow. However, without disclosed methodology, a fixed definition of adoption, and any comparative baseline, the full claim cannot be taken at face value. A narrower version survives: there exists a recorded figure of 64% tied to some operationalized adoption definition, but its meaning and reliability hinge on methodological transparency the claim does not supply.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R20

Manifest: class `SCP-1` · anchor `S07` · variant `clean` · stage `decompose+critic`
**Clean twin — no defect injected. A catch here is an over-catch (A1.2).**

Input: "The 64% adoption rate measured on the whole engineering org puts our tooling's adoption rate at 64%."

Diagnostics (not recall): axis-fired=true · decompose cardinality=2

Full chain:

  [0] claim: The engineering organization's tooling adoption rate is 64%.
      verdict: partial
      axis Evidence (weakens): The claim presents a precise numeric figure (64%) but cites no source, dataset, timestamp, or measurement system — it is a bare assertion dressed in steelman conditions added post-hoc.
      axis Hidden premise (weakens): The claim assumes 'adoption' has been consistently and operationally defined across the organization, but the definition of adoption (active use vs. installation vs. login) is notoriously variable and is admitted as needing specification in the producer's own conditions.
      axis Falsifiability (clears): As stated, the claim is falsifiable in principle — one could audit usage logs, compare against headcount, and verify the figure — so it clears this axis if the operational conditions are met.
      axis Equivocation (weakens): 'Adoption rate' is a term that slides between installation, activation, recurring use, and proficiency depending on context; the claim does not pin it down, and the producer's own steelman acknowledges this ambiguity requires resolution.
      axis Base rate / magnitude (weakens): 64% is a specific magnitude but has no comparison baseline (prior period, industry benchmark, target threshold), making it impossible to assess whether it is good, bad, or meaningful.
      axis Counterexample (weakens): If 'adoption' is measured as license activation rather than recurring use, the same organization might report 90%+ 'adoption' — demonstrating the figure is highly sensitive to definition, undermining its precision.
      axis Causality vs correlation (clears): No causal claim is made; this axis does not bear on the claim.
      reason: The claim has a falsifiable, numeric core that could in principle be verified through usage telemetry and headcount data. However, as stated it is underdetermined: 'adoption' is undefined, no measurement window is given, no source is cited, and no comparison context exists. The surviving claim is legitimate narrowing — it restates the claim with the operational constraints the producer themselves flagged as necessary, not invented qualifiers. Without those constraints the figure is meaningless, but with them it is a testable assertion. The absence of any actual cited evidence is the primary remaining weakness.

  [1] claim: The 64% adoption rate was measured across the whole engineering organization.
      verdict: hollow
      axis Evidence (fatal): The claim asserts a specific figure (64%) and a specific scope ('whole engineering organization') but cites no source, methodology document, or data system to verify either the numerator or denominator used.
      axis Hidden premise (weakens): The claim assumes a shared, unambiguous definition of 'engineering organization' exists and was applied consistently, but org boundaries (contractors, embedded teams, acquired units) are routinely contested and often silently trimmed.
      axis Falsifiability (weakens): The claim is in principle falsifiable—one could audit the headcount list used as denominator and check whether excluded subgroups exist—but no such audit is referenced, leaving it currently unverifiable.
      axis Equivocation (weakens): 'Whole engineering organization' is vague: it could mean all FTEs, all FTEs plus contractors, a single business unit's engineers, or engineers in one geography, and the claim does not pin this down.
      axis Base rate / magnitude (weakens): No total headcount figure is given, so 64% of an unknown N is uninterpretable—64% of 11 people and 64% of 4,000 carry very different statistical weight.
      axis Counterexample (weakens): Adoption metrics routinely use active-user bases or opted-in populations as the denominator while being labeled 'organization-wide,' a common inflation pattern that would make this claim false.
      axis Causality vs correlation (clears): No causal claim is made; this axis does not bear on a scoping/measurement assertion.
      reason: The claim is a bare assertion of both a specific number and a specific scope, with no cited methodology, no disclosed denominator, and an equivocal definition of 'whole engineering organization.' Every axis that fires weakens or fatally undermines it. Any narrowing that could rescue it—e.g., 'among the subset tracked by system X'—would introduce conditions entirely absent from the original claim, constituting condition laundering rather than legitimate narrowing. Without evidence of how the denominator was constructed and what population it covered, the 64% figure cannot be evaluated as organization-wide.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R21

Manifest: class `SCP-1` · anchor `S09` · variant `mutant` · stage `decompose+critic`
Injected: scope narrowed to "the users who completed onboarding" while the claim stays asserted of "our product"; predicate frozen
Narrow scope: "the users who completed onboarding" · Broad subject: "our product"

Input: "The 4.6-out-of-5 satisfaction score measured among the users who completed onboarding puts our product's satisfaction score at 4.6-out-of-5."

Diagnostics (not recall): axis-fired=true · decompose cardinality=4

Full chain:

  [0] claim: The satisfaction score measured among users who completed onboarding was 4.6 out of 5.
      verdict: partial
      axis Evidence (weakens): The claim cites a specific numeric score, implying a real measurement exists, but no source, sample size, confidence interval, or methodology is provided to verify it.
      axis Hidden premise (weakens): The claim assumes that users who completed onboarding are representative of users broadly, but completers are a self-selected high-engagement subset, so the score may not generalize even to all onboarding initiators.
      axis Falsifiability (clears): The claim is falsifiable in principle: access to the underlying survey data and methodology would confirm or refute the 4.6 figure, so it clears this axis.
      axis Equivocation (weakens): 'Satisfaction' is ambiguous — it could mean satisfaction with onboarding specifically, with the product, or with the overall experience — and the question wording is undisclosed.
      axis Base rate / magnitude (weakens): 4.6/5 is a precise number but lacks context: no sample size, no response rate, no comparison to industry benchmarks or prior periods, so its significance cannot be assessed.
      axis Counterexample (weakens): If response rate was low (e.g., only the most delighted completers responded), the 4.6 could be inflated by non-response bias, undermining the claim's validity.
      axis Causality vs correlation (clears): No causal claim is made; this is a descriptive measurement, so this axis does not materially apply.
      reason: The claim is not hollow — a specific numeric measurement is asserted and is in principle verifiable. However, multiple weakening factors (undisclosed sample size, response rate, potential self-selection among completers, ambiguous construct of 'satisfaction') prevent it from being fully substantive as stated. The surviving claim is the narrower, honest version: a recorded figure from a defined but potentially biased sample, pending methodological disclosure. The conditions in the surviving claim are drawn from what is logically required to evaluate the original claim's own precision, not introduced to launder counterexamples.

  [1] claim: The group measured was limited to users who completed onboarding.
      verdict: partial
      axis Evidence (weakens): The claim is a bare factual assertion about a methodological choice with no citation or data showing the group was actually limited this way or that the limitation was intentional rather than incidental.
      axis Hidden premise (weakens): The steelman assumes the research question was specifically about activated-user product performance, but if the question was about overall user value or acquisition ROI, the restriction is a flaw, not a feature.
      axis Falsifiability (clears): The core factual claim—that the measured group was limited to onboarding completers—is straightforwardly falsifiable by inspecting the dataset or methodology documentation, so it clears this axis.
      axis Equivocation (weakens): 'Completed onboarding' is doing heavy lifting: if the definition was post-hoc, inconsistently applied, or shifts between analyses, the restriction ceases to be a clean methodological choice and becomes a confound.
      axis Base rate / magnitude (weakens): No completion rate is given; if only a small or systematically skewed fraction of users completed onboarding, the measured group is not representative and the 'cleaner signal' rationale collapses.
      axis Counterexample (weakens): If onboarding completion correlates with user type (e.g., power users, enterprise accounts), the restricted cohort systematically excludes casual or lower-value users, making any generalization misleading even within the stated scope.
      axis Causality vs correlation (clears): No causal claim is made here; the claim is descriptive about measurement scope, so this axis does not bear materially.
      reason: The bare factual claim survives—it is falsifiable and not equivocating at its core—but the steelman's defense that this is 'methodologically sound' depends on several producer-stated conditions (pre-specified criterion, explicit scope communication, non-trivial and unbiased completion rate). The two conditions introduced in the surviving claim ('research question explicitly concerns activated users' and 'pre-specified and uniformly applied') were both present in the producer's own conditions, so this is legitimate narrowing rather than laundering. Without those conditions, the restriction is just as likely a survivorship-bias flaw as a sound design choice.

  [2] claim: The product's satisfaction score was reported as 4.6 out of 5.
      verdict: partial
      axis Evidence (weakens): The claim asserts a precise numerical score but cites no source, survey methodology, sample size, or documentation — it is a bare assertion of a measurement.
      axis Hidden premise (weakens): The claim assumes the score was derived from a representative, unbiased sample of genuine users, but this is entirely unstated and unverified.
      axis Falsifiability (clears): The claim is falsifiable in principle — one could inspect the raw survey data and check whether the mean equals 4.6 — so it is not vacuous.
      axis Equivocation (weakens): 'Reported' is doing quiet work: it could mean internally measured, selectively published, or cherry-picked from a filtered dataset, each with very different evidential weight.
      axis Base rate / magnitude (weakens): 4.6/5 sounds high, but without sample size, response rate, or comparison to industry benchmarks, the magnitude is unanchored and potentially misleading.
      axis Counterexample (weakens): Platforms routinely show inflated scores when only satisfied buyers are prompted to review, or when negative reviews are suppressed — the score could reflect bias rather than true satisfaction.
      axis Causality vs correlation (clears): No causal claim is made; this axis does not materially bear on a reported measurement claim.
      reason: The core factual claim — that a number was reported — is falsifiable and not logically incoherent. However, 'reported' conceals critical methodological assumptions, the evidence base is bare assertion, and the score's meaning is undermined without sample size or comparator. The steelman conditions are producer-supplied qualifiers, not independently verified facts. Stripping unverifiable embellishments, a narrow claim survives: a score was reported, but its reliability is unestablished.

  [3] claim: The satisfaction score of users who completed onboarding is being equated with the overall product satisfaction score.
      verdict: hollow
      axis Evidence (weakens): The steelman is entirely conditional — no evidence is cited that any of the five producer conditions actually hold in practice; the defense is hypothetical scaffolding, not documented fact.
      axis Hidden premise (weakens): The argument assumes that non-completers have experienced nothing worth measuring, but partial users often reveal real product failures (confusing UX, unmet promises) that are product-quality signals, not mere 'onboarding friction.'
      axis Falsifiability (weakens): The steelman is structured so that any counterexample can be deflected by asserting one of the five conditions was not met, making the defense untestable as a general claim about whether equating the scores is legitimate.
      axis Equivocation (fatal): 'Overall product satisfaction' shifts meaning mid-argument: it starts as satisfaction across all users and ends as satisfaction among users who experienced the product 'as designed' — two distinct populations with the same label.
      axis Base rate / magnitude (weakens): Condition (2) stipulates the completion rate must be high enough, but no threshold is given and no actual completion rate data is offered, leaving the magnitude of selection bias unquantified.
      axis Counterexample (weakens): If 60% of users drop during onboarding, reporting completers' score as 'overall product satisfaction' actively conceals a majority's negative experience — a straightforward failure case the conditions gesture at but do not empirically rule out.
      axis Causality vs correlation (weakens): The claim that higher satisfaction among completers reflects product quality rather than survivorship bias (self-selection of more motivated or compatible users) is asserted, not demonstrated.
      reason: The fatal equivocation on 'overall product satisfaction' (all-users vs. completers-only) is the core problem: the original claim is precisely that two different things are being labeled the same, and the steelman concedes this while rebranding it as legitimate. All five producer conditions are unverified hypotheticals that, if introduced as qualifiers to produce a surviving claim, would constitute condition laundering — the claim survives only by importing assumptions the speaker never established. Without demonstrated evidence that completers are not a self-selected minority, that stakeholders are explicitly informed, and that drop-off is separately tracked, the defense is a framework for how equating the scores could be acceptable, not a demonstration that it is. The claim as defended is hollow.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R22

Manifest: class `SCP-1` · anchor `S09` · variant `clean` · stage `decompose+critic`
**Clean twin — no defect injected. A catch here is an over-catch (A1.2).**

Input: "The 4.6-out-of-5 satisfaction score measured among all registered users puts our product's satisfaction score at 4.6-out-of-5."

Diagnostics (not recall): axis-fired=true · decompose cardinality=2

Full chain:

  [0] claim: The product has a satisfaction score of 4.6 out of 5.
      verdict: partial
      axis Evidence (weakens): The claim cites a specific numeric score but provides no source, sample size, methodology, or data provenance — it is a bare assertion of a number.
      axis Hidden premise (weakens): The claim assumes the survey instrument, sample, and aggregation method are valid and unbiased — none of which is established or verifiable from the claim itself.
      axis Falsifiability (weakens): The numeric claim is in principle falsifiable — one could audit the underlying survey data — but without source disclosure it cannot currently be tested, making it practically unfalsifiable as stated.
      axis Equivocation (weakens): 'Satisfaction score' could mean mean rating, median, Net Promoter Score conversion, or a proprietary weighted index, and '4.6 out of 5' has no fixed meaning without a defined rubric.
      axis Base rate / magnitude (weakens): 4.6/5 sounds high but without industry benchmarks or distribution data (e.g., what fraction are 5s vs. 1s) the magnitude claim — 'top tier of its category' — is unanchored.
      axis Counterexample (weakens): Many products achieve 4.6 through selection bias (only satisfied customers respond) or review gating, so the score alone does not establish genuine broad satisfaction.
      axis Causality vs correlation (clears): No causal inference is made; the claim is descriptive, so this axis does not materially apply.
      reason: The raw numeric claim is falsifiable in principle and could be substantive if methodology were disclosed. However, as stated, it lacks source, sample, and definitional grounding across five independent axes. The core fact — a number exists — survives, but all interpretive amplification ('overwhelming majority', 'top tier', 'diverse customer base') does not. The surviving claim is legitimately narrowed to the bare factual assertion plus the conditions the producer themselves conceded are necessary, which were explicitly stated in the producer's own conditions rather than introduced by the critic.

  [1] claim: The satisfaction score was measured among all registered users.
      verdict: partial
      axis Evidence (weakens): No citation, methodology document, or data source is provided to verify that all registered users were actually included; this is a bare assertion about a measurement process.
      axis Hidden premise (weakens): The claim assumes 100% response/capture rate or a valid imputation method for non-respondents, which is almost never achieved in practice and is left entirely unstated.
      axis Falsifiability (clears): The claim is falsifiable in principle — one could compare the registry size to the number of responses collected — but no such data is supplied here.
      axis Equivocation (weakens): 'Measured among' is ambiguous: it could mean 'delivered to all' (intended population) or 'collected from all' (actual respondents), and the distinction is critical to the claim's truth.
      axis Base rate / magnitude (weakens): No figures are given for registry size, survey reach, or response rate, so the scope of 'all' cannot be evaluated.
      axis Counterexample (weakens): Virtually every large-scale satisfaction survey excludes inactive accounts, uses sampling, or suffers significant non-response, making 'all registered users' empirically implausible without qualification.
      axis Causality vs correlation (clears): No causal claim is made; this axis does not apply.
      reason: The claim has a falsifiable core — one can check whether the measurement was designed to reach all registered users — but 'measured among all' most naturally implies complete capture, which is almost certainly false at face value. Stripping the overclaim leaves a defensible but weaker assertion about intended scope, not achieved coverage. This narrowing is grounded in the producer's own steelman language ('by design') rather than introduced externally.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R23

Manifest: class `SCP-1` · anchor `S10` · variant `mutant` · stage `decompose+critic`
Injected: scope narrowed to "the us-east edge node" while the claim stays asserted of "our CDN"; predicate frozen
Narrow scope: "the us-east edge node" · Broad subject: "our CDN"

Input: "The 88% cache hit rate measured on the us-east edge node puts our CDN's cache hit rate at 88%."

Diagnostics (not recall): axis-fired=true · decompose cardinality=2

Full chain:

  [0] claim: The us-east edge node has a cache hit rate of 88%.
      verdict: partial
      axis Evidence (weakens): The claim is a bare empirical assertion; no log excerpt, dashboard screenshot, telemetry reference, or query result is supplied — the steelman describes what evidence *would* look like, not evidence that exists.
      axis Hidden premise (weakens): The claim implicitly assumes 'cache hit rate' is computed on cacheable requests only, but if the denominator is all requests (including uncacheable ones), the 88% figure would be inflated or incomparable to industry benchmarks.
      axis Falsifiability (clears): The claim is directly falsifiable: pulling the us-east node's access logs or telemetry dashboard for the same window and applying the same formula would confirm or refute 88% — so it is not vacuous.
      axis Equivocation (weakens): 'Cache hit' is a term that can encompass TCP_HIT, stale-while-revalidate, conditional revalidation (304), or CDN-layer hits vs. origin-shield hits; without pinning the definition, the number is ambiguous.
      axis Base rate / magnitude (weakens): 88% is a concrete number but no comparison baseline is offered (e.g., historical average, peer nodes, SLO target), so its significance — good, bad, or typical — cannot be assessed.
      axis Counterexample (weakens): If the measurement window coincided with a traffic spike of highly repeated content (e.g., a viral asset), 88% could reflect an atypical interval rather than steady-state performance.
      axis Causality vs correlation (clears): No causal claim is made; the claim is purely descriptive/metric, so this axis does not apply.
      reason: The claim is empirically falsifiable and specific enough to verify, but it is weakened by the absence of actual cited evidence, an undefined hit taxonomy, an unspecified denominator population, and no baseline for contextual interpretation. Stripping those unsupported elements, a narrower claim survives: a stated metric from a named node, contingent on the definitional and sourcing conditions the producer themselves enumerate. Those conditions are not invented by the critic — they are explicitly stated by the producer — so this is legitimate narrowing, not laundering.

  [1] claim: The 88% cache hit rate measured on the us-east edge node is being used to represent the CDN's overall cache hit rate.
      verdict: partial
      axis Evidence (weakens): No data is provided showing that us-east's traffic composition, content distribution, or TTL policies actually match the global CDN mix; the 88% figure is asserted as representative without supporting cross-node comparison data.
      axis Hidden premise (weakens): The entire argument rests on the unstated assumption that us-east's request mix mirrors the global distribution—a premise that is empirically contestable and known to fail in practice when geographic content preferences, localized origin pulls, or differential cache-warming exist.
      axis Falsifiability (clears): The core claim—that the single-node figure represents overall CDN performance—is falsifiable: aggregate hit rates from all edge nodes would either confirm or refute it, making this a testable empirical question.
      axis Equivocation (weakens): 'Overall cache hit rate' slides between 'network-wide weighted average' and 'a defensible headline proxy figure,' which are meaningfully different claims with different accuracy thresholds.
      axis Base rate / magnitude (weakens): No variance data across other nodes is cited; without knowing whether peer nodes range from 70% to 95% or cluster tightly around 88%, the magnitude of the representational error is unknown.
      axis Counterexample (weakens): CDN nodes serving Asia-Pacific or Latin America frequently show markedly different hit rates due to distinct content popularity curves and higher origin-fetch latency tolerance, meaning a high-traffic US node routinely fails to represent global performance.
      axis Causality vs correlation (weakens): High traffic volume at us-east is correlated with statistical stability of that node's own metric, but stability of a single node's measurement does not cause or imply representativeness of the broader network.
      reason: The claim is falsifiable and not vacuous, so it is not hollow. However, every axis other than falsifiability weakens it: the representativeness assumption is unverified, the equivocation between 'proxy' and 'overall rate' is material, cross-node variance is unquantified, and geographic counterexamples are well-established. The surviving narrowed claim is the one the producer's own conditions describe, but those conditions are stipulated rather than evidenced. The claim therefore survives only as a conditional hypothesis, not as a demonstrated fact.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R24

Manifest: class `SCP-1` · anchor `S10` · variant `clean` · stage `decompose+critic`
**Clean twin — no defect injected. A catch here is an over-catch (A1.2).**

Input: "The 88% cache hit rate measured on our global edge fleet puts our CDN's cache hit rate at 88%."

Diagnostics (not recall): axis-fired=true · decompose cardinality=2

Full chain:

  [0] claim: The global edge fleet has a measured cache hit rate of 88%.
      verdict: partial
      axis Evidence (weakens): The claim asserts a precise measured figure (88%) but cites no data source, measurement window, methodology document, or audit trail — it is a bare number without provenance.
      axis Hidden premise (weakens): The claim assumes a consistent, unambiguous definition of 'cacheable request' as the denominator; if non-cacheable requests (auth, POST, bypass headers) are included or excluded differently, the 88% figure is incomparable across contexts.
      axis Falsifiability (clears): The claim is in principle falsifiable — one could pull aggregated edge logs and recompute the ratio — so it clears this axis structurally, though falsification requires access to proprietary instrumentation data.
      axis Equivocation (weakens): 'Cache hit rate' can mean TCP-level cache hit, byte-hit rate, object-hit rate, or shield/mid-tier hit excluded vs. included, creating meaningful ambiguity in the 88% figure without a stated definition.
      axis Base rate / magnitude (clears): 88% is a plausible figure for a mature CDN serving predominantly static content (industry benchmarks typically range 80–95%), so the magnitude is not implausible, but no comparison baseline or variance is provided.
      axis Counterexample (weakens): Fleets serving highly personalized, authenticated, or dynamic content routinely achieve 20–50% hit rates, so the figure is context-dependent and fails without knowing the workload composition.
      axis Causality vs correlation (clears): No causal claim is made; this is a descriptive measurement assertion, so this axis does not bear on the claim.
      reason: The claim is not hollow — 88% is a falsifiable, empirically computable figure — but it cannot stand as fully substantive without a cited source, a stated denominator definition, and workload context. Stripping those gaps leaves a narrower claim that survives: this is a specific operational metric that is in-principle verifiable, at a plausible magnitude, with the caveat that 'cache hit rate' must be defined consistently and the workload must be one where such rates are achievable. The conditions in the surviving claim largely reflect qualifiers implied or stated by the producer (measurement methodology, defined window, cacheable-only denominator), not laundered additions.

  [1] claim: The CDN's cache hit rate is 88%.
      verdict: partial
      axis Evidence (weakens): The 88% figure is presented with a detailed steelman but no actual cited data, logs, or vendor report — it remains an asserted number whose source is described hypothetically rather than demonstrated.
      axis Hidden premise (weakens): The claim assumes the denominator excludes non-cacheable requests, but whether that exclusion is actually applied in the measurement is unstated and materially affects the figure — including non-cacheable requests could drop the effective rate significantly.
      axis Falsifiability (clears): The claim is highly falsifiable: pulling CDN access logs for the stated period and computing (total requests minus origin fetches) / total requests would either confirm or refute 88%, making this a testable empirical assertion.
      axis Equivocation (weakens): 'Cache hit' is a term with multiple vendor-specific definitions (e.g., some vendors count partial hits, shield-layer hits, or stale-while-revalidate responses as hits), so without pinning the definition the number could shift across reporting tools.
      axis Base rate / magnitude (clears): 88% is a concrete number with an implicit comparison class (industry CDN hit rates typically range 80–95% for static assets), so the magnitude is plausible and sits within a meaningful reference range.
      axis Counterexample (weakens): If the CDN serves a high proportion of personalized, authenticated, or dynamically generated content that is architecturally non-cacheable but included in the denominator, 88% would be an inflated or misleading figure.
      axis Causality vs correlation (clears): No causal claim is made; this is a descriptive performance metric, so this axis does not apply.
      reason: The core empirical assertion is falsifiable and numerically plausible, but three non-fatal weaknesses compound: the source data is asserted rather than cited, the definition of 'cache hit' is vendor-variable, and the denominator composition (cacheable-only vs. all requests) is not verified. These do not destroy the claim but require qualification. The surviving claim restricts to conditions the producer explicitly stated, so no condition laundering occurs.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R25

Manifest: class `MFC` · anchor `M01` · variant `mutant` · stage `decompose+critic`
Injected: kill condition deleted

Input: "Streaming the parser cuts peak memory on the 2 GB corpus."

Diagnostics (not recall): axis-fired=true · decompose cardinality=1

Full chain:

  [0] claim: Streaming the parser cuts peak memory on the 2 GB corpus.
      verdict: hollow
      axis Evidence (weakens): The claim cites 'empirical profiling on this specific corpus' as prospective evidence, but no actual measurements are provided — it remains a theoretical assertion awaiting verification.
      axis Hidden premise (weakens): The claim implicitly requires that the batch parser loads an O(N) in-memory representation AND that the downstream consumer releases records incrementally — both conditions the producer explicitly acknowledges but does not verify, making the claim contingent on up to 6 unstated (in the original claim) prerequisites.
      axis Falsifiability (clears): The core claim is falsifiable: one could profile both parsers on the 2 GB corpus and compare peak RSS, making it empirically testable and not vacuous.
      axis Equivocation (weakens): 'Streaming parser' could mean event-based (SAX-style) or a pipeline that still buffers large intermediate structures; without pinning the implementation, the term hides meaningful variation in memory behavior.
      axis Base rate / magnitude (clears): The steelman provides a concrete quantitative claim (peak RSS below 200 MB vs. above 2 GB, one to two orders of magnitude reduction), which is specific and checkable.
      axis Counterexample (weakens): If the streaming parser accumulates a full AST or the downstream consumer collects all records before processing (conditions 2 and 3), peak memory is not reduced — a realistic failure mode the producer acknowledges.
      axis Causality vs correlation (clears): The causal mechanism (incremental chunk processing prevents full-corpus allocation) is well-specified and architecturally sound, not merely associative.
      reason: The core causal mechanism is sound and the claim is falsifiable via profiling. However, the original bare claim ('streaming the parser cuts peak memory') survives only when at least three conditions hold that were not stated in the original claim: the batch baseline is genuinely O(N) in memory, the streaming parser does not accumulate full output, and the downstream consumer processes records incrementally. Without those conditions the claim can trivially fail. The surviving claim is narrowed to what is architecturally defensible, but the conditions are legitimate refinements implied by the producer's own steelman rather than laundered qualifiers introduced externally. Survives only by conditions the speaker never stated.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R26

Manifest: class `MFC` · anchor `M01` · variant `clean` · stage `decompose+critic`
**Clean twin — no defect injected. A catch here is an over-catch (A1.2).**

Input: "Streaming the parser cuts peak memory on the 2 GB corpus by at least 30%, and we abandon the approach if the measured gain is under 30%."

Diagnostics (not recall): axis-fired=true · decompose cardinality=2

Full chain:

  [0] claim: Streaming the parser cuts peak memory on the 2 GB corpus by at least 30%
      verdict: hollow
      axis Evidence (weakens): No measurement, benchmark, or citation is provided; the 30% figure is asserted with a hypothetical baseline ('say, 4–6 GB') rather than observed data on the actual corpus.
      axis Hidden premise (weakens): The claim silently assumes the corpus consists of many independently parseable records (not one monolithic structure), that the baseline parser fully materializes input + object graph, and that downstream consumers discard each record promptly — none of which is stated in the claim itself.
      axis Falsifiability (clears): The claim is testable in principle: run both parsers on the 2 GB corpus, measure peak RSS with a profiling tool, and check whether the streaming variant is ≥30% lower — so it is not vacuous.
      axis Equivocation (weakens): 'Peak memory' is not defined — it could mean RSS, virtual memory, heap allocations, or OS-reported peak, and the answer can differ substantially across definitions, but the claim does not equivocate by switching meanings mid-argument.
      axis Base rate / magnitude (weakens): The steelman provides a plausible magnitude argument (baseline 4–6 GB, streaming sub-1 GB), but the baseline is hypothetical and the 30% threshold, while likely achievable under the stated conditions, is unsupported by actual measurement.
      axis Counterexample (weakens): If the corpus is a single large JSON array or XML document with no discrete records, or if an aggregation step re-materializes the full dataset, streaming provides negligible peak-memory benefit — a straightforward failure case.
      axis Causality vs correlation (clears): The mechanism (incremental parse → bounded buffer → lower peak) is causally coherent and not merely correlational.
      reason: The causal mechanism is sound and the claim is falsifiable, but it survives only under conditions (multi-record corpus, full materialization baseline, bounded streaming buffer, prompt consumption) that the original bare claim does not state. Without those conditions the counterexample (monolithic structure, aggregation pipeline) readily defeats it. No fatal flaw prevents a qualified version from being defensible, but the claim as stated overreaches because it asserts a specific quantitative threshold without measurement and tacitly requires several structural preconditions. Survives only by conditions the speaker never stated.

  [1] claim: The approach is abandoned if the measured memory gain is under 30%
      verdict: partial
      axis Evidence (weakens): No citation, study, or empirical derivation is offered for why 30% specifically is the right threshold; the steelman asserts it 'reflects a pragmatic tradeoff' without showing any data tying this number to real system behavior or cost structures.
      axis Hidden premise (weakens): The claim assumes memory gain is the only metric that matters for the go/no-go decision, silently ignoring latency, energy, code simplicity, or compounding gains when combined with other optimizations — the producer conditions acknowledge but do not resolve this.
      axis Falsifiability (clears): The claim is falsifiable in principle: one could observe a team continuing with sub-30% gains or abandoning a 31%-gain approach for other reasons, making it a testable procedural rule rather than a vacuous statement.
      axis Equivocation (weakens): 'Measured memory gain' is ambiguous — it could mean peak allocation reduction, average working-set reduction, RSS, or GPU VRAM, each of which can yield very different percentages on the same optimization, so the threshold's meaning shifts depending on the measurement chosen.
      axis Base rate / magnitude (weakens): 30% is a specific, quantified threshold, which is better than a directional claim, but no baseline is given for what typical memory optimizations yield in this system class, so whether 30% is conservative or aggressive is unknowable.
      axis Counterexample (weakens): A 28% memory gain that also halves peak latency, or one that is part of a sequence of small improvements summing to >30%, would plausibly justify continuation yet would be abandoned under this rule as stated.
      axis Causality vs correlation (clears): The claim is a procedural rule, not a causal assertion, so this axis does not materially apply.
      reason: The core claim is a falsifiable procedural rule with a concrete threshold, which gives it genuine content. However, the equivocation on what 'measured memory gain' means, the absence of empirical grounding for the 30% figure, and the obvious counterexample of sub-30% gains bundled with other benefits all prevent it from being fully substantive. Stripping the unsupported parts leaves a narrower but defensible rule: when memory is the sole objective and the metric is pre-defined, the 30% cutoff functions as a reproducible stopping criterion. The two conditions added in the surviving claim (pre-specified metric, memory-only objective) were implied by the producer's own stated conditions (1) and (3) rather than invented externally, so this is legitimate narrowing rather than laundering.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R27

Manifest: class `SCP-1` · anchor `S01` · variant `mutant` · stage `decompose+critic`
Injected: scope narrowed to "the Frankfurt replica" while the claim stays asserted of "our service"; predicate frozen
Narrow scope: "the Frankfurt replica" · Broad subject: "our service"

Input: "The 40-millisecond p99 latency measured on the Frankfurt replica puts our service's p99 latency at 40-millisecond."

Diagnostics (not recall): axis-fired=true · decompose cardinality=3

Full chain:

  [0] claim: The Frankfurt replica has a measured p99 latency of 40 milliseconds.
      verdict: partial
      axis Evidence (fatal): No benchmark data, methodology report, or reproducible test artifact is cited; the claim rests entirely on assertion with no reference to any measurement record.
      axis Falsifiability (weakens): The claim is in principle falsifiable—re-run the benchmark and observe whether p99 lands at 40 ms—but without knowing the system, workload, or time of measurement, no one can actually reproduce or refute it.
      axis Hidden premise (weakens): The claim presupposes a specific, identifiable system whose Frankfurt replica exists and is accessible; without that referent the claim has no content.
      axis Equivocation (weakens): 'Latency' is ambiguous—client-perceived round-trip, server-side processing time, and one-way network latency can differ by 10–30 ms for Frankfurt-class distances, making '40 ms' uninterpretable without a fixed definition.
      axis Base rate / magnitude (weakens): 40 ms p99 is plausible for a Frankfurt cloud replica under normal load, but without a baseline or comparison (e.g., p50, other replicas, or a prior measurement) the figure conveys no meaningful signal about system health or performance.
      axis Counterexample (weakens): Under different load, time-of-day, or latency definition, the same replica could trivially yield a p99 of 25 ms or 120 ms, showing the figure is highly context-dependent and not a stable property of the replica.
      axis Causality vs correlation (clears): No causal inference is being made; this axis does not bear on the claim.
      reason: The claim is empirically falsifiable in principle and 40 ms is a physically plausible figure for Frankfurt-region infrastructure, so it is not hollow. However, the total absence of cited evidence, the undefined system referent, and the equivocation on latency definition prevent it from being substantive as stated. The surviving claim is narrowed only to conditions the producer explicitly stated (documented run, identified replica, fixed definition, sufficient sample), not to laundered conditions, so the narrowing is legitimate. Without those producer-stated conditions being met and documented, the claim cannot be evaluated.

  [1] claim: The Frankfurt replica's p99 latency determines the overall service's p99 latency.
      verdict: partial
      axis Evidence (clears): The steelman cites order-statistics theory, which is established mathematics; no empirical data is needed for a conditional mathematical claim of this form.
      axis Hidden premise (weakens): The core unstated assumption is that p99(max(X1,...,Xn)) equals max(p99(X1),...,p99(Xn)), which is only approximately true and holds exactly only under independence or specific dependence structures; with arbitrary dependence the equality can fail.
      axis Falsifiability (clears): Under the stated conditions the claim is testable: measure per-replica p99s and end-to-end p99 in a controlled scatter-gather system and check equality; a discrepancy would falsify it.
      axis Equivocation (weakens): 'Determines' could mean 'equals' or merely 'sets the floor/ceiling'; the steelman implies strict equality, but in practice p99(max) can exceed max(p99s) even when one replica dominates, making 'determines' imprecise.
      axis Base rate / magnitude (weakens): No quantification is given of how much Frankfurt's p99 exceeds others, so it is unclear whether the dominance is marginal or decisive; a razor-thin difference could be swamped by aggregation effects.
      axis Counterexample (weakens): If Frankfurt's p99 is only marginally higher than other replicas, joint tail events (multiple replicas simultaneously slow) can push the composite p99 above Frankfurt's p99, falsifying strict determination by Frankfurt alone.
      axis Causality vs correlation (clears): The claim is structural/mathematical rather than empirical-causal, so the causality axis does not meaningfully apply here.
      reason: The mathematical core survives but only as a lower-bound / approximation claim rather than an exact equality. The original claim overstates precision: p99(max) equals max(p99s) only under independence (and even then only asymptotically for large n); joint tail events can lift the composite p99 above Frankfurt's p99. The conditions the producer supplies are necessary but not sufficient for strict equality. Stripping 'determines' to 'approximately determines / lower-bounds' yields a defensible, falsifiable claim. The surviving claim relies on conditions already stated by the producer (fan-out-wait-for-all, Frankfurt dominance, independence/positive correlation), so no laundering is introduced.

  [2] claim: The service's p99 latency is 40 milliseconds.
      verdict: partial
      axis Evidence (weakens): The claim asserts a specific numeric figure (40 ms) but no cited measurement data, benchmark report, or observable source is provided in the claim itself—only a steelman describing what ideal evidence would look like.
      axis Hidden premise (weakens): The claim implicitly assumes the measurement conditions (window, load, infrastructure, timing point) are representative and stable, which the producer conditions acknowledge but do not verify—these are non-trivial assumptions that can easily fail in practice.
      axis Falsifiability (clears): The claim is falsifiable in principle: run a sufficiently large load test under the stated conditions and observe whether ≥1% of requests exceed 40 ms; this is a concrete, testable threshold.
      axis Equivocation (weakens): P99 latency is a well-defined statistical term and '40 milliseconds' is precise, but the measurement boundary ('service boundary') could shift between server-side, client-perceived, or partial-operation scope without changing the label—a real equivocation risk in practice.
      axis Base rate / magnitude (weakens): 40 ms is a concrete magnitude with an implicit comparison class (1% of requests exceed it), which is meaningful, but without a stated request volume or observation window the statistical stability of the estimate is unverified.
      axis Counterexample (weakens): A single cold-start event, GC pause, or dependency timeout during the measurement window that was filtered out could artificially depress the measured p99 below the true operational p99—common in practice.
      axis Causality vs correlation (clears): No causal claim is made; this is a descriptive statistical claim, so this axis does not bear materially on the assessment.
      reason: The claim is falsifiable and uses precise terminology, so it is not hollow. However, evidence is asserted rather than cited, the measurement boundary is susceptible to equivocation, and the statistical stability depends on an unverified request volume. The surviving claim is the original narrowed only to conditions the producer explicitly stated, not introduced by the critique, so conditioning is legitimate rather than laundered. The claim holds as a partial: meaningful and testable if the stated conditions are met, but unverified as stated.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R28

Manifest: class `SCP-1` · anchor `S01` · variant `clean` · stage `decompose+critic`
**Clean twin — no defect injected. A catch here is an over-catch (A1.2).**

Input: "The 40-millisecond p99 latency measured on all 14 production replicas puts our service's p99 latency at 40-millisecond."

Diagnostics (not recall): axis-fired=true · decompose cardinality=4

Full chain:

  [0] claim: The service has 14 production replicas.
      verdict: partial
      axis Evidence (weakens): The claim is a bare factual assertion with no cited deployment record, orchestration platform output, or snapshot to verify the count of 14.
      axis Hidden premise (weakens): The claim assumes a stable, single production environment where 'replicas' maps unambiguously to a running count, but multi-cluster or multi-region setups could make '14' an aggregate or a subset depending on scope.
      axis Falsifiability (clears): The claim is highly falsifiable: querying the orchestration platform (e.g., kubectl get replicaset) at any moment would confirm or refute it, making this a concrete, testable assertion.
      axis Equivocation (weakens): 'Replicas' could mean desired count in config versus actual running/ready count, and '14' could be stale if sourced from documentation rather than live state.
      axis Base rate / magnitude (clears): No comparison or context is given (e.g., expected count, historical count, or SLA requirement), but the claim makes only a point-in-time factual assertion about a specific number, not a comparative one, so this is not a primary concern.
      axis Counterexample (weakens): A scaling event, node failure, or rolling deployment in progress at observation time could mean the live count is transiently above or below 14, falsifying the claim at that moment.
      axis Causality vs correlation (clears): No causal structure is asserted; this is a direct factual count claim, so this axis does not apply.
      reason: The claim is in principle falsifiable and concrete, but weakened by lack of cited evidence, ambiguity between desired vs. actual replica count, potential scope ambiguity in 'production', and transient instability. Stripping unsupported precision leaves a narrower but defensible claim tied to a point-in-time live observation. The narrowing conditions (single environment, live query, ready state) were substantially present in the producer's own stated conditions, so this is legitimate narrowing rather than laundering.

  [1] claim: The p99 latency was measured on all 14 production replicas.
      verdict: partial
      axis Evidence (weakens): The claim asserts a specific measurement methodology ('all 14 replicas') but cites no monitoring dashboards, telemetry logs, or replica inventory records to confirm full coverage — it is bare assertion.
      axis Hidden premise (weakens): The claim assumes all 14 replicas had active, gap-free instrumentation during the measurement window — a non-trivial operational assumption that could silently fail if even one replica's agent was down or misconfigured.
      axis Falsifiability (clears): The claim is falsifiable in principle: telemetry coverage logs or a replica inventory audit could show fewer than 14 replicas reported, making this a testable factual assertion rather than a vacuous one.
      axis Equivocation (weakens): 'Measured on all 14 replicas' is ambiguous — it could mean p99 was computed per-replica and then aggregated, computed over a combined pool from all 14, or simply that all 14 were instrumented regardless of how the statistic was derived, which matters enormously for the validity of the figure.
      axis Base rate / magnitude (weakens): The claim makes no assertion about sample size per replica or total request volume, so whether 14-replica coverage actually produces a statistically robust p99 (vs. a sparse one) cannot be evaluated.
      axis Counterexample (weakens): A monitoring agent crash, a replica restart mid-window, or a network partition could silently exclude one or more replicas from the sample — common operational events that would make the claim false without anyone necessarily noticing.
      axis Causality vs correlation (clears): No causal inference is being made here; the claim is purely descriptive/factual, so this axis does not apply.
      reason: The claim is falsifiable and specific enough to be meaningful, but multiple weaknesses — no cited evidence, ambiguity in what 'measured on' means for the computation method, and realistic operational failure modes — prevent it from being fully substantive. The narrower surviving claim strips the assertion of certainty and flags the verification conditions that were already implied (not invented) by the producer's own stated conditions, so no condition laundering is occurring.

  [2] claim: The measured p99 latency is 40 milliseconds.
      verdict: partial
      axis Evidence (weakens): The claim asserts a specific numeric measurement, which is either verifiable from logs/telemetry or not — no evidence is cited in the claim itself, but the steelman describes a plausible measurement methodology that would produce citable evidence.
      axis Hidden premise (weakens): The claim assumes a single stable number represents the system's p99 latency, hiding the premise that latency is stationary across time, load, and deployment context — p99 varies with conditions and is not a fixed property.
      axis Falsifiability (clears): The claim is directly falsifiable: reproduce the measurement under the stated conditions and check whether the 99th percentile falls at 40ms — this is a concrete, testable assertion.
      axis Equivocation (weakens): The producer's own condition 8 acknowledges ambiguity in what '40ms' means — it could be the upper bound of a histogram bucket, the exact empirical rank value, or an interpolated estimate — these are meaningfully different and the claim does not specify.
      axis Base rate / magnitude (clears): 40ms is a specific magnitude with an implicit comparison class (is this good or bad for the system type?), but no comparison baseline or SLA context is provided — the claim is a measurement report, not an evaluation, so this axis barely applies.
      axis Counterexample (weakens): The measurement will differ if taken at a different time window, load level, or with a different p99 definition (e.g., bucket ceiling vs. exact rank), meaning the claim fails to generalize beyond the specific measurement window.
      axis Causality vs correlation (clears): No causal claim is made; this is a descriptive measurement assertion — this axis does not bear on the claim.
      reason: The claim is falsifiable and concerns a concrete, reportable measurement, so it is not hollow. However, it is weakened by definitional ambiguity in what '40ms p99' means (bucket ceiling vs. exact rank vs. interpolated), by the implicit assumption that the value is stationary or representative beyond the measurement window, and by the absence of cited evidence. Stripping those weaknesses leaves a narrower claim tied to a specific window and a precise p99 definition — both of which were partially present in the producer's own conditions, making the narrowing legitimate rather than laundered.

  [3] claim: The service's p99 latency is 40 milliseconds.
      verdict: partial
      axis Evidence (weakens): The claim is a bare numerical assertion with no cited measurement data, methodology, or source; even the steelman acknowledges these as aspirational conditions rather than demonstrated facts.
      axis Hidden premise (weakens): The claim silently assumes a specific measurement scope (which requests, which clients, which time window, which percentile algorithm) that entirely determines whether '40 ms' is meaningful or misleading.
      axis Falsifiability (clears): The claim is falsifiable in principle—run independent latency probes and observe whether ≥1% of requests exceed 40 ms—so it is not vacuous.
      axis Equivocation (weakens): 'p99 latency' can mean server-side, client-perceived, or a coarse-bucket approximation, and these can differ substantially; the claim does not fix the definition.
      axis Base rate / magnitude (weakens): 40 ms is a concrete number with a clear percentile rank (99th), but without a stated load level, infrastructure context, or comparison baseline it is uninterpretable as good, bad, or stable.
      axis Counterexample (weakens): A measurement taken only during low-traffic hours or excluding slow dependency calls could report p99=40 ms while true production p99 is far higher—an easily realised failure mode.
      axis Causality vs correlation (clears): No causal claim is made; this axis does not materially apply.
      reason: The claim is falsifiable and numerically precise, so it is not hollow. However, the undefined measurement scope (client vs. server, traffic mix, algorithm) and lack of cited evidence leave it unverifiable as stated. Narrowing it to 'under the conditions actually used' preserves a defensible core without introducing new conditions—it merely makes explicit the conditionality already embedded in any p99 statistic. The surviving claim remains worth a fresh pass only if the measurement conditions are specified.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R29

Manifest: class `SCP-1` · anchor `S04` · variant `mutant` · stage `decompose+critic`
Injected: scope narrowed to "our enterprise tier" while the claim stays asserted of "the business"; predicate frozen
Narrow scope: "our enterprise tier" · Broad subject: "the business"

Input: "The 87% renewal rate measured in our enterprise tier puts the business's renewal rate at 87%."

Diagnostics (not recall): axis-fired=true · decompose cardinality=3

Full chain:

  [0] claim: The enterprise tier has a measured renewal rate of 87%.
      verdict: partial
      axis Evidence (weakens): The claim asserts a specific measured figure (87%) but cites no data source, audit trail, or verifiable system of record — it is a bare numerical assertion dressed in measurement language.
      axis Hidden premise (weakens): The claim assumes a consistent, documented methodology exists and has been applied uniformly, but this is precisely what is unverified and what the producer's own conditions acknowledge as required.
      axis Falsifiability (clears): The claim is falsifiable in principle: pulling the underlying contract records and recomputing the renewal rate against a stated methodology would confirm or refute 87%; this axis clears.
      axis Equivocation (weakens): 'Renewal rate' is a term that shifts substantially depending on whether partial renewals, downgrades, multi-year rollovers, or logo-only vs. ARR-weighted counting are used, and none of these choices are fixed in the claim itself.
      axis Base rate / magnitude (weakens): 87% is a precise figure but no comparison benchmark is given (industry average, prior period, peer cohort), making it impossible to assess whether it is strong, weak, or unremarkable.
      axis Counterexample (weakens): If 'enterprise' is defined loosely or the measurement window cherry-picks a favorable period, the same underlying data could yield a materially different rate, illustrating that the figure is methodology-contingent.
      axis Causality vs correlation (clears): No causal inference is made; this is a descriptive statistic, so this axis does not apply.
      reason: The claim is falsifiable and numerically specific, which gives it a defensible core. However, it weakens substantially because 'renewal rate' is a methodologically slippery term, the data source and methodology are unverified in the claim as stated, and no comparative benchmark contextualizes 87%. Stripping out the unverified methodological assumptions, a narrower claim survives: the figure is real and checkable if — and only if — the producer's own listed conditions actually hold. Those conditions are stated by the producer, not invented by the critic, so this is legitimate narrowing rather than condition laundering.

  [1] claim: The business's overall renewal rate is 87%.
      verdict: partial
      axis Evidence (weakens): The claim states a precise figure (87%) but no source, dataset, audit trail, or citation is provided — it is a bare assertion of a number.
      axis Hidden premise (weakens): The claim assumes a single, unambiguous definition of 'renewal rate' is being used, but renewal rate can be calculated on a logo, revenue, or unit basis, each yielding materially different results from the same underlying data.
      axis Falsifiability (clears): The claim is falsifiable in principle — one could pull CRM or billing records for a defined period and recompute the ratio — so it is not vacuous.
      axis Equivocation (weakens): 'Renewal rate' is a genuine buzzword with multiple competing definitions (logo vs. revenue, gross vs. net, including or excluding expansions), and without specifying which is used, the number cannot be reproduced or compared.
      axis Base rate / magnitude (weakens): 87% is a specific magnitude, but without an industry benchmark, prior-period figure, or cohort context, there is no way to assess whether 87% represents strong, average, or poor retention for this business type.
      axis Counterexample (weakens): A business could report 87% by excluding churned mid-term customers, at-risk accounts offered special retention deals, or multi-year contracts not yet at renewal — all legitimate-seeming exclusions that inflate the headline figure.
      axis Causality vs correlation (clears): The claim is descriptive, not causal, so this axis does not apply.
      reason: The core claim — a specific percentage derived from a ratio of renewals to eligible contracts — is falsifiable and operationally meaningful. It survives once the definitional ambiguity (which methodology) and denominator-scoping issues are resolved, both of which the producer's own conditions acknowledge. The claim does not survive in its bare form because the term 'renewal rate' equivocates across incompatible definitions and no evidence is cited, but a narrower, auditable version holds. The narrowing conditions applied here (methodology consistency, denominator scoping) were explicitly stated by the producer, not introduced by the critic, so this is legitimate narrowing, not condition laundering.

  [2] claim: The enterprise tier renewal rate is being used to represent the business's overall renewal rate.
      verdict: partial
      axis Evidence (weakens): The 60-80% revenue concentration figure is a plausible industry heuristic but no specific evidence is cited tying it to this business; the claim rests on an assumed structural fact about this company's revenue mix.
      axis Hidden premise (weakens): The entire defense assumes enterprise revenue concentration is high enough here to justify the substitution — but if enterprise is, say, 40% of revenue, the proxy argument collapses entirely, and this assumption is stated as a condition rather than demonstrated.
      axis Falsifiability (clears): The steelman is testable — if enterprise revenue share is below the claimed threshold, or if stakeholders were not informed the metric was enterprise-specific, the defense fails — so the core logic is not vacuous.
      axis Equivocation (fatal): 'Overall renewal rate' slides between customer-count renewal and revenue-weighted renewal without acknowledgment; presenting the enterprise rate as 'overall' without disclosing this substitution is the very ambiguity under scrutiny, not a resolved one.
      axis Base rate / magnitude (weakens): No actual revenue split or tier-level renewal rates for this business are provided, so the magnitude of the representation error (or its absence) cannot be assessed.
      axis Counterexample (weakens): A business where enterprise is 40% of revenue but has a 95% renewal rate, while SMB at 60% has a 60% renewal rate, would make the enterprise-as-proxy deeply misleading — this is not a contrived scenario.
      axis Causality vs correlation (clears): No causal conflation issue; the claim is about representational accuracy, not causation.
      reason: The equivocation on 'overall renewal rate' is the fatal flaw: presenting a segment metric as an enterprise-wide metric without explicit disclosure is the problem being defended against, not solved by the steelman. The revenue-weighting argument is logically coherent but depends entirely on empirical conditions (revenue concentration, stakeholder disclosure) that are stated as producer-supplied assumptions rather than established facts. A narrow claim survives — the substitution can be legitimate under specific, verifiable conditions — but the original framing as 'overall' without qualification does not survive.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R30

Manifest: class `SCP-1` · anchor `S04` · variant `clean` · stage `decompose+critic`
**Clean twin — no defect injected. A catch here is an over-catch (A1.2).**

Input: "The 87% renewal rate measured in our whole customer base puts the business's renewal rate at 87%."

Diagnostics (not recall): axis-fired=true · decompose cardinality=2

Full chain:

  [0] claim: The business measured a renewal rate of 87% across its whole customer base.
      verdict: partial
      axis Evidence (weakens): The steelman asserts billing-system verification and cohort methodology, but these are producer-stated conditions, not independently confirmed facts — the underlying data, audit trail, and methodology documents are not cited or available to the critic.
      axis Hidden premise (weakens): The claim assumes 'renewal rate' is defined and applied consistently (e.g., denominator = eligible renewers only, not total customer base), but this is precisely the axis on which renewal-rate figures most commonly mislead — the steelman acknowledges this risk without resolving it.
      axis Falsifiability (clears): The core factual claim — that a measured renewal rate of 87% exists — is falsifiable: pull the billing records for the defined cohort period and recompute; if the result differs materially, the claim is false.
      axis Equivocation (weakens): 'Renewal rate' is a term with multiple valid definitions (logo retention, revenue retention, cohort-based, point-in-time), and without knowing which was actually applied, the figure 87% could represent materially different business realities.
      axis Base rate / magnitude (weakens): 87% is a specific, concrete number, but no industry benchmark or comparison period is provided, making it impossible to assess whether 87% is strong, average, or poor for this business type and segment mix.
      axis Counterexample (weakens): A business could report 87% renewal rate while excluding month-to-month customers, trial conversions, or involuntary churners — standard practices that can inflate the headline figure significantly without violating a loose definition of 'measurement'.
      axis Causality vs correlation (weakens): The steelman infers from 87% that there is strong product-market fit and effective customer success operations, but an 87% renewal rate could equally result from switching costs, contractual lock-in, or lack of alternatives — causal attribution is unsupported.
      reason: The bare factual claim — that a measured figure of 87% exists — is falsifiable and not inherently hollow. However, the inferential layer (strong PMF, effective CS, predictable revenue) is unsupported causal assertion, and the figure's meaningfulness depends entirely on definitional consistency that has been asserted but not demonstrated. Stripping the causal interpretation and flagging the definitional dependency produces a defensible but qualified claim. No conditions were laundered in — the qualifiers in the surviving claim directly mirror the producer's own stated conditions, which were in scope.

  [1] claim: The business's renewal rate is 87%.
      verdict: partial
      axis Evidence (weakens): The claim cites no data source, audit trail, or methodology; it is a bare numerical assertion with no supporting documentation presented.
      axis Hidden premise (weakens): The claim assumes a consistent, agreed-upon definition of 'renewal' (active recommitment vs. passive non-cancellation) and a stable denominator, neither of which is stated.
      axis Falsifiability (clears): The claim is falsifiable in principle: auditing CRM or billing records for a defined cohort and period would either confirm or refute 87%; the problem is methodological ambiguity, not unfalsifiability.
      axis Equivocation (weakens): 'Renewal rate' can mean customer count retention, revenue retention, or contract renewal rate, each yielding materially different figures from the same underlying data.
      axis Base rate / magnitude (weakens): 87% is a specific, directionally meaningful number, but without industry benchmarks or cohort size it is impossible to assess whether it is strong, average, or weak performance.
      axis Counterexample (weakens): If passive non-cancellation is counted as renewal, the figure could overstate true loyalty; if multi-year contracts are excluded from the denominator, the rate is artificially inflated.
      axis Causality vs correlation (clears): No causal claim is being made here; the axis does not apply to a descriptive statistic.
      reason: The core numerical claim is falsifiable and not internally incoherent, but it is fatally underspecified as stated. Equivocation on 'renewal rate' and the absence of any cited evidence prevent it from being substantive. A narrower claim survives only if the methodological conditions the Producer themselves stipulated are actually met — those conditions were stated by the speaker, not introduced by this analysis, so this is legitimate narrowing rather than laundering. If the conditions are not verifiably met, the figure is uninterpretable.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R31

Manifest: class `MFC` · anchor `M10` · variant `mutant` · stage `decompose+critic`
Injected: kill condition deleted

Input: "The connection pool reduces tail latency on the auth service."

Diagnostics (not recall): axis-fired=true · decompose cardinality=1

Full chain:

  [0] claim: The connection pool reduces tail latency on the auth service.
      verdict: partial
      axis Evidence (weakens): No empirical data, benchmark, or measurement is cited showing actual tail latency reduction from a connection pool on this specific auth service; the mechanism is plausible but the effect is asserted, not demonstrated.
      axis Hidden premise (weakens): The claim assumes that connection-establishment latency is the dominant tail-latency driver, but this must hold simultaneously with GC pauses, upstream dependency jitter, lock contention, and other sources not being larger contributors — a non-trivial empirical question.
      axis Falsifiability (clears): The claim is falsifiable: measure p95/p99 latency before and after deploying the pool under otherwise identical load; a lack of improvement would refute it.
      axis Equivocation (clears): 'Reduces tail latency' is not quantified — it could mean 1 ms or 100 ms improvement — but the directional claim is still meaningful and not equivocal in kind.
      axis Base rate / magnitude (weakens): No magnitude is given for the expected reduction (e.g., 'eliminates 2–6 RTTs ≈ 10–30 ms per tail event'), and no baseline tail latency is provided, making it impossible to assess practical significance.
      axis Counterexample (weakens): If the auth service's tail latency is dominated by GC stop-the-world pauses or database lock contention, a connection pool would leave tail latency essentially unchanged — a realistic scenario that defeats the claim.
      axis Causality vs correlation (clears): The mechanism (pool removes connection-setup from the critical path) is a direct causal chain, not merely a correlation, so this axis is satisfied.
      reason: The causal mechanism is sound and the claim is falsifiable, but it fails on magnitude (no quantification) and on the counterexample that other tail-latency sources may dominate. The producer's own conditions 3, 4, and 6 are not speaker-invented weasel words — they are genuine prerequisites for the mechanism to function, and they were explicitly stated by the producer, so including them is legitimate narrowing rather than laundering. The surviving claim is narrower but honest: the pool reduces the connection-setup component of tail latency, conditional on that component actually being present and not swamped by other sources.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R32

Manifest: class `MFC` · anchor `M10` · variant `clean` · stage `decompose+critic`
**Clean twin — no defect injected. A catch here is an over-catch (A1.2).**

Input: "The connection pool reduces tail latency on the auth service by at least 25%, and we abandon the approach if the measured gain is under 25%."

Diagnostics (not recall): axis-fired=true · decompose cardinality=2

Full chain:

  [0] claim: The connection pool reduces tail latency on the auth service by at least 25%.
      verdict: partial
      axis Evidence (weakens): The 25% figure is asserted with illustrative ranges (20–80ms handshakes, 50–150ms base latency) but no actual measured baseline or post-pool data from this auth service is cited; the steelman's numbers are hypothetical benchmarks, not evidence.
      axis Hidden premise (weakens): The claim implicitly assumes connection establishment is the dominant tail-latency contributor for this specific auth service, but if the bottleneck is downstream query execution, lock contention, or GC pauses, pooling provides little or no tail-latency relief.
      axis Falsifiability (clears): The claim is falsifiable in principle: measure p95/p99 latency under identical load before and after introducing the pool with the same instrumentation; a <25% reduction would falsify it.
      axis Equivocation (weakens): 'Tail latency' is not pinned to a specific percentile (p95 vs p99 vs p999 behave differently), and the improvement threshold could shift substantially depending on which percentile is chosen.
      axis Base rate / magnitude (weakens): The 25% floor is plausible in the regime where handshake overhead is 30%+ of total latency, but no empirical base rate for this service's current latency decomposition is provided, so the magnitude claim is unanchored.
      axis Counterexample (weakens): If the auth service's tail latency is dominated by credential-store query time (e.g., a slow DB lookup under lock contention), adding a connection pool eliminates a minor cost and the 25% threshold would not be met.
      axis Causality vs correlation (clears): The causal mechanism (eliminating per-request setup cost → reduced tail latency) is sound and direct; this is not a correlation claim.
      reason: The causal mechanism is sound and the claim is falsifiable, but the unconditional '≥25%' assertion fails without empirical evidence that connection setup is actually the dominant latency contributor for this specific service. The producer's own conditions expose that the 25% figure is contingent on a latency decomposition that has not been measured. Stripping unsupported universality leaves a conditional claim that is defensible: the threshold is achievable if and only if the handshake overhead is a sufficiently large fraction of current tail latency. The surviving condition ('connection establishment accounts for ≥30% of tail latency') was explicitly stated by the producer, so it is legitimate narrowing, not laundering.

  [1] claim: The approach will be abandoned if the measured gain is under 25%.
      verdict: partial
      axis Evidence (weakens): The steelman describes a pre-registration and break-even analysis as if they exist, but no evidence is cited that either document has actually been created or validated; this remains assertion.
      axis Hidden premise (weakens): The claim assumes that the measurement will be conducted faithfully and that no political, financial, or organizational pressure will cause the decision rule to be reinterpreted or ignored—a strong and unstated governance assumption.
      axis Falsifiability (clears): As stated the claim is falsifiable in principle: if the measured gain is under 25% and the approach continues, the claim is false—this is a genuine testable prediction.
      axis Equivocation (weakens): 'Measured gain' and 'abandoned' are each load-bearing terms whose definitions can shift (which metric? binding to whom?); the producer's conditions acknowledge this but do not resolve whether those definitions are actually locked in.
      axis Base rate / magnitude (weakens): The 25% threshold is specific and the break-even rationale is provided, but no base rate of similar decision rules being honored in comparable organizational contexts is supplied.
      axis Counterexample (weakens): Organizations routinely continue initiatives that fail pre-set thresholds due to sunk-cost pressure, stakeholder politics, or retroactive redefinition of metrics; this is a direct and common failure mode.
      axis Causality vs correlation (weakens): The causal chain (sub-threshold result → abandonment) depends entirely on enforcement mechanisms whose existence is asserted but unverified; the rule existing on paper does not cause abandonment.
      reason: The claim has a falsifiable, specific core—a quantitative decision rule with a named threshold. That core survives. However, its force depends on several conditions the producer themselves enumerate (pre-registration, binding authority, no override clauses) that are asserted rather than evidenced, and the counterexample of organizational non-compliance is pervasive. The claim is not hollow because the threshold and falsification condition are concrete, but substantial portions of its force rest on unverified governance assumptions, reducing it to a conditional.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R33

Manifest: class `MFC` · anchor `M05` · variant `mutant` · stage `decompose+critic`
Injected: kill condition deleted

Input: "Compression lowers egress cost on the media bucket."

Diagnostics (not recall): axis-fired=true · decompose cardinality=1

Full chain:

  [0] claim: Compression lowers egress cost on the media bucket.
      verdict: partial
      axis Evidence (clears): Egress pricing models for S3/GCS are publicly documented and per-GB billing is verifiable; compression ratios for WebP/AVIF/Opus vs. originals are well-established in the literature, so the causal chain has available empirical support.
      axis Hidden premise (weakens): The claim assumes egress is the dominant or a significant cost component—if the workload is read-light or assets are served primarily from a CDN with flat-rate or zero-origin-egress pricing (e.g., Cloudflare R2), the mechanism exists but the realized savings approach zero.
      axis Falsifiability (clears): The claim is clearly falsifiable: compare monthly egress invoices before and after applying compression to the same asset corpus under identical traffic; if cost does not decrease proportionally to size reduction, the claim fails.
      axis Equivocation (clears): 'Compression' is used broadly but the steelman appropriately distinguishes codec compression (WebP, Opus) from transfer encoding (gzip/Brotli), both of which reduce billed bytes—no meaningful equivocation remains given the producer's clarification.
      axis Base rate / magnitude (weakens): The steelman cites 30–80% size reduction and notes it 'compounds across millions of requests,' but provides no concrete dollar-per-TB egress rate or example invoice delta; the direction is right but the magnitude claim is asserted rather than grounded.
      axis Counterexample (weakens): Cloudflare R2 charges zero egress fees; assets already stored as JPEG progressive or MP4 with optimal encoding yield negligible further compression; flat-rate CDN contracts cap egress cost regardless of byte volume—these are real scenarios where the mechanism fails to produce savings.
      axis Causality vs correlation (clears): The causal mechanism is direct and well-specified: fewer bytes transferred → fewer GB billed at a per-GB rate → lower invoice; this is not a mere association.
      reason: The core causal mechanism is sound and falsifiable. Three axes merely weaken rather than fatally undercut the claim: the magnitude is underspecified, the hidden premise about egress billing model is real, and counterexamples (R2, flat-rate CDN) show non-trivial scope restrictions. However, these scope restrictions were explicitly stated by the producer in conditions 1–4, so narrowing to 'per-GB billed egress' is legitimate narrowing, not condition laundering. No new conditions were introduced beyond what the producer already stipulated.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R34

Manifest: class `MFC` · anchor `M05` · variant `clean` · stage `decompose+critic`
**Clean twin — no defect injected. A catch here is an over-catch (A1.2).**

Input: "Compression lowers egress cost on the media bucket by at least 20%, and we abandon the approach if the measured gain is under 20%."

Diagnostics (not recall): axis-fired=true · decompose cardinality=2

Full chain:

  [0] claim: Compression lowers egress cost on the media bucket by at least 20%.
      verdict: hollow
      axis Evidence (weakens): Compression ratios for specific format types (30–80% for JSON, SVG, etc.) are well-documented in literature and benchmarks, but no evidence is cited that these formats constitute the majority of egress in this specific 'media bucket'; the per-asset claim is supported, the aggregate claim is asserted.
      axis Hidden premise (fatal): The claim silently assumes the media bucket holds primarily compressible assets, yet 'media bucket' strongly connotes video, audio, and image files (JPEG, MP4, MP3) which are already compressed and yield near-zero savings.
      axis Falsifiability (clears): Given a specific bucket's egress logs and object-type distribution, one could measure pre- and post-compression transfer volumes and verify or refute the 20% threshold, so the claim is falsifiable in principle.
      axis Equivocation (fatal): 'Media bucket' is used loosely: in common cloud usage it implies binary media assets (already compressed), but the steelman redefines it to include text-adjacent formats; this definitional shift does significant load-bearing work.
      axis Base rate / magnitude (weakens): The 20% blended saving requires compressible assets to be a large fraction of egress volume, but for a typical media bucket dominated by video/audio, such assets are a small fraction, making 20% aggregate savings an outlier rather than a baseline expectation.
      axis Counterexample (fatal): A media bucket serving primarily H.264 video and JPEG images would see near-zero egress reduction from compression, directly falsifying the unqualified claim.
      axis Causality vs correlation (clears): The causal chain (compress → fewer bytes transferred → lower billed egress) is mechanistically sound given provider per-GB billing, no mere correlation issue here.
      reason: Three fatal issues undermine the unqualified claim: 'media bucket' typically implies already-compressed binary assets (equivocation), the assumption that compressible formats dominate egress is unstated and often false for media workloads (hidden premise), and a counterexample of a video/JPEG bucket trivially defeats it. The surviving claim is legitimate only after restricting to buckets explicitly verified to hold mostly compressible content—a condition the speaker implied via steelman but did not state in the claim itself. This is borderline laundering, but the speaker's own listed conditions (conditions 2 and 5) do supply this qualifier, so it is not fully introduced from outside. However, those conditions materially restrict the claim's applicability and must be explicit. Survives only by conditions the speaker never stated.

  [1] claim: The approach will be abandoned if the measured gain is under 20%.
      verdict: hollow
      axis Evidence (weakens): The claim is a forward-looking commitment with no evidence the threshold was actually pre-registered, that a cost-benefit analysis was performed, or that decision-makers are genuinely bound — all support comes from the steelman's asserted conditions, not demonstrated facts.
      axis Hidden premise (weakens): The claim assumes decision-makers will honor a pre-commitment rule even when facing sunk costs, political pressure, or ambiguous measurement results — a strong behavioral assumption that routinely fails in practice.
      axis Falsifiability (weakens): As a stated behavioral commitment ('will be abandoned'), it is in principle falsifiable: observe whether abandonment follows a sub-20% measured result; however, the five producer conditions together create enough interpretive escape routes (redefining the metric, disputing statistical power, invoking override mechanisms) that the commitment may be practically untestable.
      axis Equivocation (weakens): 'Measured gain' is doing heavy lifting — without a locked-in operational definition, the term can shift post-hoc to whichever metric clears the 20% bar, hollowing out the stopping rule.
      axis Base rate / magnitude (weakens): The 20% threshold is asserted as economically principled but no cost-benefit analysis is cited; whether 20% is the right cutoff versus 10% or 30% depends entirely on unstated cost structures.
      axis Counterexample (weakens): Numerous documented cases show organizations that pre-commit to stopping rules but renegotiate them when results are inconvenient (e.g., clinical trial interim analyses, corporate pilot programs), demonstrating the commitment is frequently not honored.
      axis Causality vs correlation (clears): No causal issue; the claim is a conditional commitment, not an empirical causal assertion.
      reason: The core logical structure — a binary stopping rule keyed to a quantitative threshold — is coherent and falsifiable in principle. However, every axis except causality fires as 'weakens': the threshold's justification is asserted not evidenced, 'measured gain' is equivocal without a locked definition, and the behavioral commitment is empirically fragile. The claim survives only as a narrower, conditional proposition, and the conditions that rescue it (pre-registration, locked metric, binding authority) are producer-stipulated, not demonstrated. This is borderline laundering but the producer explicitly stated the conditions, so it counts as legitimate narrowing rather than introduced conditions — yet the claim's practical force is substantially reduced. Survives only by conditions the speaker never stated.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R35

Manifest: class `SCP-1` · anchor `S03` · variant `mutant` · stage `decompose+critic`
Injected: scope narrowed to "the iOS build" while the claim stays asserted of "our app"; predicate frozen
Narrow scope: "the iOS build" · Broad subject: "our app"

Input: "The 99.4% crash-free rate measured on the iOS build puts our app's crash-free rate at 99.4%."

Diagnostics (not recall): axis-fired=true · decompose cardinality=1

Full chain:

  [0] claim: The iOS build has a measured crash-free rate of 99.4%.
      verdict: partial
      axis Evidence (weakens): The claim asserts a specific measured figure (99.4%) but provides no citation of the tool, dashboard, time window, or sample size — it is bare assertion at this stage, though the steelman describes what evidence would look like.
      axis Hidden premise (weakens): The claim assumes the crash-reporting SDK captures all crashes, but OOM kills, watchdog terminations, and background expirations are frequently excluded by default in tools like Crashlytics, meaning the true crash rate could be higher than reported.
      axis Falsifiability (clears): The claim is falsifiable: producing the crash-reporting dashboard for the specified build and window would either confirm or refute the 99.4% figure, so it is not vacuous.
      axis Equivocation (weakens): 'Crash-free rate' can mean session-based or user-based, and these yield materially different numbers (a power user who crashes once counts once in user-based but inflates session-based), so the figure is ambiguous without a stated definition.
      axis Base rate / magnitude (weakens): 99.4% is a concrete figure with implicit comparison to the complement (0.6% crash rate), but no industry benchmark or prior build figure is offered to contextualize whether this is good, average, or poor for the app category.
      axis Counterexample (weakens): If the measurement window coincided with a low-traffic period or a filtered user cohort (e.g., only users on the latest OS), the 99.4% would not generalize to the full user base.
      axis Causality vs correlation (clears): No causal claim is made — the figure is presented as a measurement, not an explanation, so this axis does not apply.
      reason: The specific numeric claim is falsifiable and plausible, but equivocation over 'crash-free rate' (session vs. user), unknown exclusions (OOM, watchdog), and the absence of any cited evidence reduce it from a clean factual assertion to a conditionally defensible one. The conditions introduced in the surviving claim largely mirror the producer's own stated conditions, not new laundered qualifiers, so the narrowing is legitimate rather than manufactured.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R36

Manifest: class `SCP-1` · anchor `S03` · variant `clean` · stage `decompose+critic`
**Clean twin — no defect injected. A catch here is an over-catch (A1.2).**

Input: "The 99.4% crash-free rate measured on all shipped platforms puts our app's crash-free rate at 99.4%."

Diagnostics (not recall): axis-fired=true · decompose cardinality=2

Full chain:

  [0] claim: The app has a measured crash-free rate of 99.4%.
      verdict: partial
      axis Evidence (weakens): The steelman asserts SDK-level instrumentation and a credible tool but provides no actual citation, dataset, or audit trail — the conditions describe what *would* make it valid, not confirmation that those conditions are met.
      axis Hidden premise (weakens): The claim assumes that the crash reporting SDK captures all crashes, but SDKs routinely miss OOM kills, ANRs (on Android), watchdog terminations, and out-of-process crashes, silently inflating the reported rate.
      axis Falsifiability (clears): The claim is falsifiable in principle: a different crash reporting tool or a raw device log audit on the same population and window should reproduce ≈99.4%; if it cannot be reproduced or audited, the figure is unverifiable rather than hollow.
      axis Equivocation (weakens): 'Crash-free' can mean crash-free sessions, crash-free users, or crash-free installs — these produce materially different numbers from the same underlying data, and the claim does not specify which unit is used.
      axis Base rate / magnitude (weakens): 99.4% is a concrete number, but without a stated sample size, time window, platform, or version range, the precision is cosmetic — 0.6% of 1,000 sessions (6 crashes) and 0.6% of 10,000,000 sessions carry entirely different reliability and operational significance.
      axis Counterexample (weakens): A known failure mode: if the SDK itself crashes or is excluded from a subset of builds/regions, those sessions are never reported, making 99.4% an undercount of crashes rather than a true rate — this is not hypothetical, it is a documented Crashlytics/Sentry limitation.
      axis Causality vs correlation (clears): No causal inference is being made; this is a descriptive measurement claim, so this axis does not apply.
      reason: The core claim — that a specific numeric quality metric was produced by standard tooling from real production data — is falsifiable and plausible. However, three independent weaknesses prevent 'substantive': (1) the metric unit (sessions vs. users) is unspecified and materially changes the number; (2) sample size and scope are unconfirmed, making the precision unverified; (3) SDK blind spots (OOM, ANR, watchdog) mean the true crash-free rate is likely lower than reported. No conditions were laundered in; the narrowing reflects ambiguities present in the original claim itself.

  [1] claim: The 99.4% crash-free rate was measured across all shipped platforms.
      verdict: partial
      axis Evidence (weakens): The claim asserts a precise numeric figure (99.4%) but provides no citation, dataset reference, or verifiable source — it is self-referentially supported only by the producer's own steelman conditions.
      axis Hidden premise (weakens): The claim assumes that the instrumented platforms are representative of the full user population, but if non-instrumented platforms have systematically worse crash rates, the 99.4% is misleading even if technically accurate within scope.
      axis Falsifiability (clears): Within the stated scope (instrumented platforms, consistent definition), the claim is falsifiable: auditing the telemetry pipeline and recalculating from raw session logs would confirm or refute it.
      axis Equivocation (weakens): 'Crash-free' is acknowledged as requiring a consistent, explicit definition, and the steelman stipulates this was done — but the actual definition is not stated here, leaving open whether ANRs, OOM kills, or watchdog terminations were included or excluded in ways that inflate the figure.
      axis Base rate / magnitude (weakens): 0.6% crash rate sounds small, but without session volume, affected-user count, platform coverage percentage, or industry benchmark comparison, the magnitude's significance is unassessable.
      axis Counterexample (weakens): If a major platform (e.g., a high-volume mobile OS version) was excluded from instrumentation due to known instability, the 99.4% would be a best-case cherry-picked subset — a structurally plausible failure mode the claim does not rule out.
      axis Causality vs correlation (clears): No causal claim is made; this is a measurement assertion, so this axis does not apply.
      reason: The claim survives in a narrowed form because it is falsifiable within its stated scope, the core conditions were articulated by the producer (not introduced by the critic), and no fatal flaw exists. However, the absence of a cited source, the unspecified crash definition, the lack of session-volume and benchmark context, and the structural risk of instrumentation selection bias all prevent it from being fully substantive. The surviving claim preserves the speaker's own explicit scoping language without adding new conditions.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R37

Manifest: class `MFC` · anchor `M06` · variant `mutant` · stage `decompose+critic`
Injected: kill condition deleted

Input: "The retry policy cuts transient failures on the payments API."

Diagnostics (not recall): axis-fired=true · decompose cardinality=1

Full chain:

  [0] claim: The retry policy reduces transient failures on the payments API.
      verdict: partial
      axis Evidence (weakens): The claim is asserted without citing measured before/after failure rates, logs, or any empirical data from the actual payments API deployment.
      axis Hidden premise (weakens): The claim assumes the failures being retried are genuinely transient and resolve within the retry window — if failures are persistent (e.g., downstream outage, misconfiguration), retries accomplish nothing and merely add latency.
      axis Falsifiability (clears): The claim is falsifiable in principle: compare final-failure rates (post-all-retries) before and after policy deployment on the same failure taxonomy; a null or negative result would refute it.
      axis Equivocation (weakens): 'Reduces transient failures' is ambiguous — it could mean fewer failures visible to the caller (retries absorb them), fewer total attempt-level failures, or a lower underlying error rate; these are very different quantities and the claim conflates them.
      axis Base rate / magnitude (weakens): No baseline failure rate, post-implementation rate, or effect size is given, so the claim establishes only a direction ('reduces') with no magnitude or statistical significance.
      axis Counterexample (weakens): If the payments API is experiencing a persistent overload, retries can trigger a retry storm that increases failure rates rather than reducing them — a well-documented failure mode that the claim ignores.
      axis Causality vs correlation (weakens): Even if failure rates dropped after deploying the retry policy, concurrent changes (infrastructure upgrades, dependency fixes) could be the actual cause; no controlled attribution is established.
      reason: The core mechanism is sound and falsifiable: retrying transient errors can convert attempt-level failures into caller-level successes. However, the original claim is too broad — it asserts reduction as a fact without evidence, elides the equivocation between attempt-failure and final-failure rates, and ignores the retry-storm counterexample. The surviving claim is narrowed to the conditions the producer explicitly stated, not laundered ones, so it remains defensible but demands empirical validation before it can be treated as established.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R38

Manifest: class `MFC` · anchor `M06` · variant `clean` · stage `decompose+critic`
**Clean twin — no defect injected. A catch here is an over-catch (A1.2).**

Input: "The retry policy cuts transient failures on the payments API by at least 60%, and we abandon the approach if the measured gain is under 60%."

Diagnostics (not recall): axis-fired=true · decompose cardinality=2

Full chain:

  [0] claim: The retry policy cuts transient failures on the payments API by at least 60%.
      verdict: partial
      axis Evidence (weakens): The steelman cites plausible industry ranges (70–90% transient error rates, specific backoff intervals) but no measured data from the actual payments API is provided; the 60% figure is asserted, not demonstrated.
      axis Hidden premise (weakens): The claim implicitly requires that the baseline is measured pre-retry and that ≥60% of failures are genuinely transient—neither of which is established for this specific system, and both are load-bearing assumptions.
      axis Falsifiability (clears): The claim is testable: deploy the retry policy behind a feature flag, compare failure rates before and after over matched traffic windows, and check whether the reduction meets 60%—this is a concrete empirical test.
      axis Equivocation (weakens): 'Transient failures' does real work here—if the denominator shifts to include hard errors (declined cards, fraud blocks), the 60% figure collapses; the producer conditions acknowledge this but the original claim does not scope it.
      axis Base rate / magnitude (weakens): The 60% threshold is directionally plausible given published transient-error distributions in cloud payments environments, but without this API's own baseline error composition the magnitude is unverified.
      axis Counterexample (weakens): If the transient error share is below 60% (e.g., dominated by hard declines or persistent downstream outages), the retry policy cannot achieve the claimed reduction regardless of configuration quality.
      axis Causality vs correlation (clears): Retry policies have a direct mechanistic causal path to reducing transient failures, so the causal link is structurally sound; the only question is magnitude, not direction.
      reason: The causal mechanism is sound and the claim is falsifiable, but the '60%+' magnitude depends entirely on the actual transient-error share of this API's baseline—a figure not yet measured or cited. Stripping the unverified magnitude claim leaves a narrower, testable version: the reduction in transient failures will meet or exceed 60% if and only if the transient fraction of baseline failures itself meets or exceeds 60%. The original claim's scope to 'transient failures' is stated in the producer's conditions and implied by any competent reading of retry logic, so the surviving condition (transient share ≥ 60%) is legitimate narrowing, not laundering—it was explicitly listed as producer condition #2.

  [1] claim: The approach will be abandoned if the measured gain is under 60%.
      verdict: partial
      axis Evidence (clears): The claim is structural/logical rather than empirical—it asserts that a governance design would produce falsifiability, which is evaluable by inspection of the design's properties rather than requiring external evidence; no citation is needed here.
      axis Hidden premise (weakens): The claim assumes that binding public pre-registration is actually sufficient to prevent motivated redefinition or non-compliance—but enforcement mechanisms and institutional willingness to act are assumed operative, not demonstrated.
      axis Falsifiability (clears): Within the stipulated governance conditions, a sub-60% result is a concrete, observable state of affairs that either does or does not trigger the defined abandonment actions, making the construct genuinely falsifiable in the Popperian sense.
      axis Equivocation (weakens): 'Binding' does real work here and could slide between legally enforceable, institutionally normative, or merely reputationally costly—the steelman acknowledges this range but does not pin it down, leaving the key load-bearing term partially equivocal.
      axis Base rate / magnitude (clears): No magnitude question applies: the claim is about logical and procedural structure, not a quantitative causal estimate, so this axis does not bear.
      axis Counterexample (weakens): A real-world counterexample is readily available: pre-registered clinical trial stopping rules are routinely renegotiated, ignored, or reframed post-hoc despite binding protocols—demonstrating that formal pre-definition does not guarantee that the trigger fires as written.
      axis Causality vs correlation (clears): The claim is conditional-logical rather than causal-empirical, so this axis does not bear meaningfully.
      reason: The core logical claim survives: a sufficiently constrained pre-registration regime does produce a falsifiable criterion in the design sense. The weaknesses are real but not fatal: 'binding' equivocates across a spectrum of enforceability, and real-world pre-registration precedents show the gap between procedural falsifiability and actual compliance. The surviving claim is therefore narrower—it is a claim about design-level falsifiability conditional on enforcement fidelity, not a guarantee of real-world compliance. All conditions in the surviving claim were explicitly stated by the speaker, so no laundering occurs.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R39

Manifest: class `MFC` · anchor `M09` · variant `mutant` · stage `decompose+critic`
Injected: kill condition deleted

Input: "Prefetching cuts page-load time on the catalog page."

Diagnostics (not recall): axis-fired=true · decompose cardinality=1

Full chain:

  [0] claim: Prefetching reduces page-load time on the catalog page.
      verdict: partial
      axis Evidence (weakens): The steelman cites plausible mechanism and named metrics (LCP, TTI) but provides no measured data from the actual catalog page; it is a mechanistic argument, not empirical evidence.
      axis Hidden premise (weakens): The claim assumes the catalog page's load time is currently bottlenecked on late resource discovery over the network rather than on CPU, scripting, or server-side rendering — an assumption that may not hold for any given catalog implementation.
      axis Falsifiability (clears): The claim is falsifiable: a controlled A/B test measuring LCP/TTI with and without prefetch directives on the catalog page would confirm or refute it, so it is not vacuous.
      axis Equivocation (weakens): 'Page-load time' shifts between perceived metrics (LCP, TTI) and total load time, and 'prefetching' covers both preload (critical path) and prefetch (idle-time speculation), which have meaningfully different effects.
      axis Base rate / magnitude (weakens): No magnitude or reference range is given — the claim says only 'reduces' without quantifying expected effect size or citing comparable catalog-page benchmarks, making the practical claim unanchored.
      axis Counterexample (weakens): On a bandwidth-constrained or CPU-bound device, prefetching can compete with in-flight critical resources and increase LCP, directly refuting the unqualified claim.
      axis Causality vs correlation (clears): The mechanism described is causal and well-established in browser networking literature, so causal inference here is appropriate if the conditions hold.
      reason: The core causal mechanism is sound and the claim is falsifiable via A/B testing, but the unqualified form fails because counterexamples (bandwidth contention, CPU-bound rendering) are real and the claim equivocates on both 'prefetching' and 'page-load time'. The surviving claim restricts to the speaker's own stated condition 6 (network bottleneck) and condition 3 (cache freshness), both of which were explicitly listed by the producer, so this is legitimate narrowing, not laundering. No magnitude claim survives.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R40

Manifest: class `MFC` · anchor `M09` · variant `clean` · stage `decompose+critic`
**Clean twin — no defect injected. A catch here is an over-catch (A1.2).**

Input: "Prefetching cuts page-load time on the catalog page by at least 30%, and we abandon the approach if the measured gain is under 30%."

Diagnostics (not recall): axis-fired=true · decompose cardinality=2

Full chain:

  [0] claim: Prefetching cuts page-load time on the catalog page by at least 30%
      verdict: partial
      axis Evidence (weakens): The steelman cites Google, Shopify, and 'field studies' for 20–50% reductions but provides no specific study, URL, or dataset tied to this catalog page or even a comparable one; the range merely brackets the claim rather than establishing it.
      axis Hidden premise (weakens): The claim assumes the catalog page's load is network-latency-bound and that prefetched assets sit on the critical path—neither is established for the actual page in question, and many catalog pages are JS-parse or render-bound, making prefetching largely irrelevant.
      axis Falsifiability (weakens): Without the six producer conditions, the bare claim ('cuts load time by at least 30%') is straightforwardly falsifiable via A/B measurement on the specific catalog page; the conditions introduced by the producer, however, are so numerous that any failure could be attributed to a violated condition.
      axis Equivocation (weakens): 'Page-load time' shifts between LCP, Time-to-Interactive, and full page load across the steelman, and the metric chosen can swing results by 30% or more on its own, obscuring whether the threshold is actually met.
      axis Base rate / magnitude (weakens): The cited 20–50% range from commerce pages provides a real comparison band, and 30% sits within it, giving the magnitude some grounding—though that band is wide enough to be nearly unfalsifiable as a reference.
      axis Counterexample (weakens): Catalog pages with heavy client-side rendering, lazy-loaded images, or dynamic personalization frequently show negligible gains from prefetching because the bottleneck is JS execution or server-side computation, not asset fetch latency.
      axis Causality vs correlation (weakens): The field studies cited conflate prefetching with other co-deployed optimizations (CDN changes, image compression, speculation rules bundles), so attributing a specific 30% delta to prefetching alone is not causally established.
      reason: The bare claim is too broad to be true universally—many catalog pages won't meet the network-bound, critical-path, and cache-hit conditions—but a narrower version tied to conditions the producer explicitly stated (not laundered in) does survive. The metric ambiguity and attribution-to-prefetching problem remain weakening factors even in the narrow form, but don't fully defeat it. The surviving claim uses only conditions the producer explicitly enumerated, so this is legitimate narrowing, not laundering.

  [1] claim: The approach is abandoned if the measured gain from prefetching is under 30%
      verdict: hollow
      axis Evidence (fatal): No citation, measurement data, or derivation is offered to show why 30% specifically was chosen; the steelman asserts it is 'principled' and 'empirical' but provides no actual evidence.
      axis Hidden premise (weakens): The claim assumes overhead costs are large enough and fixed enough that any sub-30% gain is net-negative, but this relationship between overhead magnitude and the threshold is asserted, not demonstrated.
      axis Falsifiability (weakens): The claim is falsifiable in principle—one could find a system that retains prefetching below 30% gain or abandons it above 30%—so it is not vacuous, but the producer conditions layer so many qualifiers that in practice the threshold becomes nearly unfalsifiable for any specific case.
      axis Equivocation (fatal): 'Gain' is left undefined—it could mean throughput, latency reduction, IPC improvement, or energy efficiency—and the choice of metric materially affects whether 30% is crossed, enabling selective measurement.
      axis Base rate / magnitude (weakens): 30% is presented as a meaningful engineering threshold with no comparison to industry norms, prior work, or cost-benefit analysis quantifying the overhead that the threshold is supposed to offset.
      axis Counterexample (weakens): A system with extremely low prefetch overhead (e.g., hardware stride prefetcher with negligible bandwidth cost) might rationally retain prefetching at a 10% gain, falsifying the universal abandonment rule at sub-30%.
      axis Causality vs correlation (weakens): No causal mechanism links the 30% figure to cost recovery; the argument is that costs 'are likely to outweigh benefits,' which is speculative association rather than a demonstrated cost model.
      reason: The two fatal axes—lack of evidence for the 30% figure and undefined 'gain' metric—combine with multiple weakening axes to remove any defensible core. The steelman's producer conditions attempt to rescue the claim by introducing extensive qualifiers (well-defined baseline, quantified overhead, correct metric, empirical derivation, consistent application, exhausted alternatives), none of which were present in the original claim. Stripped of those laundered conditions, the original claim is an arbitrary numerical threshold applied to an undefined metric with no supporting cost model. The surviving claim would require so many added conditions that it reverts to hollow without them. Survives only by conditions the speaker never stated.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R41

Manifest: class `SCP-1` · anchor `S11` · variant `mutant` · stage `decompose+critic`
Tags: `confound:arith` — **quarantined: read it, but it never pools into the class number (A2.1)**
Injected: scope narrowed to "our Oxford Street flagship" while the claim stays asserted of "the company"; predicate frozen
Narrow scope: "our Oxford Street flagship" · Broad subject: "the company"

Input: "The 1-minute-50 median checkout time measured at our Oxford Street flagship puts the company's median checkout time under two minutes."

Diagnostics (not recall): axis-fired=true · decompose cardinality=6

Full chain:

  [0] claim: The Oxford Street store is the company's flagship store.
      verdict: hollow
      axis Evidence (fatal): No citation, document, press release, or official company communication is provided to confirm the flagship designation; this is bare assertion.
      axis Hidden premise (weakens): The claim assumes that location on a busy shopping street (Oxford Street) itself confers flagship status, but many companies place flagship stores elsewhere (e.g., Bond Street, Fifth Avenue, or a home-city store) for brand-positioning reasons.
      axis Falsifiability (clears): The claim is falsifiable in principle — one could check official company communications, store size data, and investment records — so it is not vacuous, but no such check has been performed or cited.
      axis Equivocation (weakens): 'Flagship store' can mean formally designated by the company OR simply the largest/most prominent location; the claim conflates these two distinct senses without anchoring to either.
      axis Base rate / magnitude (weakens): No comparative data on store size, investment, or product range across the company's locations is offered to substantiate that this store is the primary one.
      axis Counterexample (weakens): Many major retailers (e.g., Apple, Nike, luxury brands) place their globally recognised flagship on streets other than Oxford Street even when they have an Oxford Street presence.
      axis Causality vs correlation (weakens): The steelman implies Oxford Street location causes flagship status, but high footfall correlates with commercial importance without determining corporate designation.
      reason: The claim is falsifiable and not inherently incoherent, but as stated it is unsupported bare assertion compounded by equivocation over what 'flagship' means. A narrower claim — conditional on verifiable official designation or measurable store-level metrics — survives, but the original claim in isolation has no evidential support and relies on an implicit (and contestable) assumption that Oxford Street presence equals flagship status. Survives only by conditions the speaker never stated.

  [1] claim: The Oxford Street store is located on Oxford Street.
      verdict: partial
      axis Evidence (weakens): No specific store, chain, or city is identified, so there is no concrete evidence to evaluate — the claim is a bare assertion about an unnamed referent.
      axis Hidden premise (weakens): The claim assumes 'Oxford Street store' refers to a real, specific, uniquely identifiable store, but without context this referent is undetermined.
      axis Falsifiability (clears): For any specific, identified store called 'the Oxford Street store,' the claim is trivially falsifiable by checking its postal address — so it is testable in principle.
      axis Equivocation (weakens): There are multiple streets named Oxford Street in different cities, and the claim does not specify which one, introducing mild ambiguity but not a fatal shift in meaning.
      axis Base rate / magnitude (clears): The naming convention of stores after streets is genuinely common, making the claim highly probable as a base-rate matter, but 'highly probable' is not 'certain.'
      axis Counterexample (weakens): A store could retain the 'Oxford Street' name after relocating, or could have been named after a founder or brand rather than its address, falsifying the identity claim in those cases.
      axis Causality vs correlation (clears): No causal inference is made; this is a straightforward identity/location claim, so this axis does not apply.
      reason: The claim is not hollow — it is falsifiable and grounded in a real and common naming convention — but it cannot be evaluated as fully substantive without a specific referent. The counterexamples (relocation, non-address-based naming) are real but rare, so a narrowed, probabilistic version survives. No conditions were laundered in; the narrowing merely makes the implicit referent explicit, which the original claim requires to be testable at all.

  [2] claim: The median checkout time at the Oxford Street flagship store was measured.
      verdict: hollow
      axis Evidence (fatal): The claim asserts a measurement was conducted but cites no data, study report, timestamp logs, or any verifiable source to confirm the measurement actually occurred.
      axis Falsifiability (weakens): The claim is in principle falsifiable — one could inspect whether records of such a measurement exist — but no such records are produced or referenced, leaving it an unverifiable assertion in practice.
      axis Hidden premise (weakens): The claim assumes a single, identifiable 'Oxford Street flagship store' exists and was accessible for structured observation, which is plausible but unconfirmed.
      axis Equivocation (weakens): 'Measured' is doing heavy lifting — it could mean a rigorous timed study or someone glancing at a clock; the steelman elaborates but the original claim is silent on methodology.
      axis Base rate / magnitude (weakens): No median value is actually reported, so the claim that a median was 'calculated' produces no checkable quantity whatsoever.
      axis Causality vs correlation (clears): No causal structure is implied; this axis does not materially bear on the claim.
      axis Counterexample (weakens): No counterexample is structurally possible against a pure existence claim ('a measurement was done'), but the absence of any output or record is effectively equivalent to one.
      reason: The claim is a bare procedural assertion — that a measurement happened — with no cited evidence, no reported value, and no methodology disclosed. The steelman elaborates what a valid measurement would require, but elaborating requirements does not constitute evidence that they were met. Every axis that bears on the claim weakens or fatally undermines it: there is no data, no number, no source, and the key term 'measured' is vague enough to mean almost anything. What survives is only the logical possibility that such a measurement could exist somewhere, which is not a substantive claim.

  [3] claim: The measured median checkout time at the Oxford Street flagship store is 1 minute and 50 seconds.
      verdict: partial
      axis Evidence (fatal): No source, dataset, study reference, or publication is cited; the steelman describes a methodology that could produce this number but offers no actual data confirming it was carried out or reported anywhere.
      axis Falsifiability (clears): The claim is in principle falsifiable — one could access the same POS logs or repeat the observational study and check whether the median equals 110 seconds — so it is not vacuous.
      axis Hidden premise (weakens): The claim assumes the study described in the steelman was actually conducted and that '1 minute 50 seconds' is the genuine output of that process, not a rounded, estimated, or fabricated figure.
      axis Equivocation (weakens): 'Checkout time' requires a fixed start and end point (scan-first-item to payment-confirmed vs. join-queue to receipt, for example), and without confirmation of the definition used, the precise figure is ambiguous across plausible interpretations.
      axis Base rate / magnitude (weakens): 110 seconds is a plausible order of magnitude for a retail checkout median, so the number is not inherently implausible, but plausibility is not evidence.
      axis Causality vs correlation (clears): No causal claim is made; this is a descriptive measurement claim, so this axis does not bear on it.
      axis Counterexample (clears): No counterexample defeats a specific empirical measurement claim of this type; either the data show 110 s or they do not.
      reason: The claim is falsifiable and the number is not implausible, so it is not hollow. However, it rests entirely on an unverified, uncited study. The methodology is coherent but asserted, not demonstrated. The surviving claim is narrower: the figure would be credible if and only if the study is real and its methodology matches the steelman's description — neither of which is established.

  [4] claim: The company's median checkout time is under two minutes.
      verdict: partial
      axis Evidence (weakens): The steelman cites internal telemetry logs across hundreds of thousands of transactions, which is a plausible and obtainable evidence base, but the data has not been independently verified or published—it remains a bare assertion that such logs exist and show what is claimed.
      axis Hidden premise (weakens): The claim assumes the measurement window definition (cart entry to confirmation render) captures what consumers and regulators would recognize as 'checkout time,' when competitors or critics might define it differently (e.g., including login or address-entry steps that precede cart review).
      axis Falsifiability (clears): The claim is straightforwardly falsifiable: access the telemetry dataset and compute the 50th-percentile completion time; if it is ≥120 seconds, the claim is false.
      axis Equivocation (clears): 'Median checkout time' is defined precisely in the producer conditions (true 50th percentile, consistent window, all segments), so there is no meaningful term-shifting.
      axis Base rate / magnitude (weakens): A specific threshold (120 seconds) and a specific statistic (50th percentile) over a large sample are provided, giving both a quantity and a comparison point; industry benchmarks for checkout times exist for external validation but are not cited.
      axis Counterexample (weakens): Guest checkouts, mobile users on slow connections, and transactions requiring address verification or 3DS authentication are known to inflate completion times and could push the median above two minutes if they constitute a large share of transactions.
      axis Causality vs correlation (clears): The claim is descriptive/statistical, not causal, so this axis does not apply.
      reason: The core claim is specific, falsifiable, and grounded in a plausible evidence base. However, the evidence is self-reported and unverified, the measurement-window definition is doing significant work that could be disputed, and known slow-path transaction types (guest checkout, 3DS, slow mobile connections) represent a real threat to the median figure. The claim survives in a narrowed form contingent on the producer conditions already stated by the speaker—no new conditions are laundered in—but it cannot be rated fully substantive without independent data access.

  [5] claim: The median checkout time at the Oxford Street flagship store is representative of the company's median checkout time.
      verdict: hollow
      axis Evidence (weakens): No empirical data is cited showing that Oxford Street's transaction mix, basket sizes, or customer demographics actually match the company-wide distribution; the steelman asserts it 'if' these conditions hold, not that they do.
      axis Hidden premise (fatal): The argument assumes Oxford Street is operationally typical, but flagship stores are by definition atypical — they are selected for high footfall, brand visibility, and premium location, all of which systematically differentiate them from the median store.
      axis Falsifiability (clears): The claim is falsifiable in principle: comparing Oxford Street's median checkout time to the distribution of store-level medians company-wide would confirm or refute it, so this axis clears.
      axis Equivocation (weakens): 'Representative' slides between 'statistically valid proxy' (a strong claim requiring distributional equivalence) and 'reasonable estimate' (a weak claim), without specifying which standard applies.
      axis Base rate / magnitude (weakens): No quantitative evidence is offered on how much Oxford Street's customer mix or checkout times deviate from the company-wide distribution, so the magnitude of any bias is unknown.
      axis Counterexample (fatal): Oxford Street serves a disproportionately tourist-heavy, international, and high-spend demographic with atypical peak hours and product mix — precisely the factors the steelman's own condition 4 acknowledges would invalidate the claim.
      axis Causality vs correlation (clears): No causal mechanism is invoked here; the argument is about statistical representativeness, so this axis does not materially bear on the claim.
      reason: Two fatal flaws sink the claim. First, the hidden-premise flaw is self-defeating: flagship stores are definitionally chosen for non-typical characteristics (high footfall, urban premium location, tourist concentration), making the assumption of typicality implausible on its face without contrary evidence. Second, the steelman's own conditions 2 and 4 — that the transaction mix must match the company-wide distribution and that no location-specific factors systematically skew times — are almost certainly violated at Oxford Street, as the producer effectively concedes. Any surviving claim would require adding conditions that Oxford Street demonstrably matches the company-wide mix across payment type, basket size, and customer type, none of which are established. Introducing those conditions would be laundering, not narrowing, since the speaker's original claim is unconditional. The equivocation on 'representative' further prevents a coherent partial claim from being extracted.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________

## R42

Manifest: class `SCP-1` · anchor `S11` · variant `clean` · stage `decompose+critic`
Tags: `confound:arith` — **quarantined: read it, but it never pools into the class number (A2.1)**
**Clean twin — no defect injected. A catch here is an over-catch (A1.2).**

Input: "The 1-minute-50 median checkout time measured at all 32 of our stores puts the company's median checkout time under two minutes."

Diagnostics (not recall): axis-fired=true · decompose cardinality=4

Full chain:

  [0] claim: The company measured checkout times at all 32 of its stores.
      verdict: partial
      axis Evidence (weakens): The claim is a bare assertion — no audit records, POS logs, or confirmation of data receipt from all 32 stores is cited or referenced.
      axis Hidden premise (weakens): The claim assumes that initiating measurement at all 32 stores is equivalent to successfully completing and recording valid data at all 32 stores — gaps from technical failures or staffing issues are silently excluded.
      axis Falsifiability (clears): The claim is falsifiable in principle: one could check whether records exist for all 32 stores with no omissions, making this testable rather than vacuous.
      axis Equivocation (weakens): 'Measured' is ambiguous — it could mean anything from a full systematic log to a supervisor glancing at a clock once, but no specific standard of measurement is defined.
      axis Base rate / magnitude (weakens): No time period, sample size per store, or quality threshold for the measurements is specified, making it unclear what 'measuring checkout times' actually produced in magnitude or completeness.
      axis Counterexample (weakens): It is common in multi-site audits for at least one location to have data dropout due to POS failure, staffing gaps, or connectivity issues, meaning 'all 32' is a strong claim that frequently fails in practice.
      axis Causality vs correlation (clears): No causal claim is made; this axis does not bear on a factual coverage claim.
      reason: The core factual claim — full coverage across exactly 32 stores — is falsifiable and not inherently incoherent, but it is unsupported by cited evidence, equivocates on what 'measured' means, and fails to account for the routine likelihood of partial data dropout. A narrower claim that the company undertook such a measurement effort survives, but the strong assertion of complete, verified coverage does not without corroborating documentation. The surviving claim uses no conditions not implied by the original; it simply strips the unverified completeness guarantee.

  [1] claim: The median checkout time measured across all 32 stores is 1 minute and 50 seconds.
      verdict: partial
      axis Evidence (weakens): The steelman describes a plausible methodology but the actual study data, sample size, and raw results are not cited or provided — this remains an assertion about what was done rather than demonstrated evidence.
      axis Hidden premise (weakens): The claim assumes that the measurement window chosen (days, shifts, seasons) is genuinely representative of ongoing operations; if data were collected only during, say, low-traffic morning periods, the median would not generalize to 'all 32 stores' in any meaningful sense.
      axis Falsifiability (clears): The claim is highly falsifiable: an independent audit re-pulling POS timestamp data from the same 32 stores and recalculating the median would either confirm or refute the 110-second figure.
      axis Equivocation (weakens): The producer conditions acknowledge 'checkout time' must be defined consistently, but in practice POS systems across stores may log different events (last scan vs. payment confirmation vs. drawer close), creating definitional drift that could shift the median non-trivially.
      axis Base rate / magnitude (clears): No comparison baseline is offered (e.g., industry norm, prior period) to contextualize whether 110 seconds is notable or expected, but the claim is purely descriptive rather than comparative, so this does not undermine its core content.
      axis Counterexample (weakens): If even one of the 32 stores had a system outage, data gap, or non-standard POS configuration during the measurement window, the 'all 32 stores' claim fails, and a corrected median over the remaining stores could differ materially.
      axis Causality vs correlation (clears): The claim is purely descriptive/statistical with no causal assertion; this axis does not apply.
      reason: The claim is specific, numerically precise, and structurally falsifiable, which clears the most fatal bars. However, it weakens on three axes: the underlying data are not produced for inspection, definitional consistency across 32 POS systems is assumed rather than demonstrated, and full-store coverage without gaps is unverified. A narrower surviving claim holds — the number is meaningful and checkable — but confidence is reduced pending audit. No conditions were introduced that the speaker did not already state; the surviving claim merely restates the producer's own conditions rather than adding laundered qualifiers.

  [2] claim: 1 minute and 50 seconds is under 2 minutes.
      verdict: substantive
      axis Evidence (clears): The arithmetic is directly verifiable: 1×60+50=110 seconds, 2×60=120 seconds, 110<120 — no citation needed for a definitional fact.
      axis Hidden premise (clears): Assumes the standard SI/common definition of 1 minute = 60 seconds, which is universally accepted and uncontroversial.
      axis Falsifiability (clears): The claim is falsified if 110 ≥ 120, which is arithmetically impossible under standard units, making it a necessary truth — not vacuous but analytically certain.
      axis Equivocation (clears): No key term shifts meaning; 'minute,' 'second,' and 'under' are used consistently and standardly.
      axis Base rate / magnitude (clears): The margin is precisely quantified at 10 seconds, so magnitude is stated exactly.
      axis Counterexample (clears): No counterexample exists within standard time units; the inequality 110<120 is unconditionally true.
      axis Causality vs correlation (clears): No causal claim is made; this is pure arithmetic identity, so the axis is inapplicable.
      reason: Every axis clears. The claim is an arithmetic necessity under universally accepted time unit definitions, is precisely quantified, uses no equivocation, and admits no counterexample. It is as substantive as a claim can be.

  [3] claim: The company's median checkout time is under 2 minutes.
      verdict: partial
      axis Evidence (weakens): The claim cites infrastructure investments as circumstantial support but presents no actual timing data, transaction logs, or third-party audit confirming the 2-minute median threshold has been measured and met.
      axis Hidden premise (weakens): The claim assumes that fast infrastructure (one-click, pre-filled fields) translates to fast user behavior, but user decision time, review pauses, and form corrections are behavioral variables decoupled from technical speed.
      axis Falsifiability (clears): The claim is fully falsifiable: pull a representative sample of timestamped completed transactions, compute the 50th percentile of elapsed time, and check whether it falls below 120 seconds — a concrete, executable test exists.
      axis Equivocation (weakens): 'Checkout time' could start at cart view, 'proceed to checkout' click, payment page load, or another trigger — without a fixed definition, the number is manipulable across definitions.
      axis Base rate / magnitude (weakens): No industry benchmark or comparative figure is offered; 2 minutes is a specific, testable threshold and not internally vague, but no context is given to assess whether it is extraordinary or ordinary for this business type.
      axis Counterexample (weakens): New customers without pre-filled data, mobile users on slow connections, and customers comparing products mid-checkout likely have substantially longer times, and if they form a large enough share, they would push the median above 2 minutes.
      axis Causality vs correlation (clears): No causal inference is being made — the claim is a descriptive statistic about a distribution, so this axis does not apply.
      reason: The claim is falsifiable and internally coherent, but survives only in a narrowed form that explicitly requires a fixed measurement definition and a representative sample — both absent from the original assertion. The infrastructure steelman is suggestive but not sufficient evidence; behavioral latency, new-customer share, and definitional ambiguity of 'checkout time' each materially weaken confidence. The narrowed claim retains a genuine, testable core but cannot be accepted at face value without transaction-log data.

catch (y/n): ______   catch-source (decompose|critic|verdict): ______

note: ____________________________________________
