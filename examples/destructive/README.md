# Destructive testing — assay's published failure probes

This directory is where assay tries to break itself in public.

If you just cloned this repo and want to know whether to trust the tool, this is the most honest
thing to read. The other examples (`../dan_shipper`, `../url-length`) show assay *working*. These show
where it is *designed to fail*, what it actually does when pushed there, and how to tell a real catch
from a lucky one.

## What "destructive testing" means

Borrowed from materials engineering. You don't certify a steel beam by confirming it holds up a chair;
you load it until it bends and *record the load at which it bends*. The number you publish is the
failure envelope, not a pass. A beam with no published failure load is a beam nobody has actually
tested.

assay outputs **judgments**, and a judgment has no deterministic ground truth (see
[`../../TESTING.md`](../../TESTING.md)). So you cannot unit-test the verdicts the way you unit-test a
parser. What you *can* do is engineer inputs that target one weakness at a time, run each many times,
and publish the distribution of what comes back. That is what lives here.

## Why a claim-validation tool publishes its own failure probes

assay's entire value is that it externalises judgment you'd otherwise have to trust yourself to make.
A tool like that has one characteristic way to fail: **laundering confidence** — being right enough,
often enough, on the easy cases that you stop checking it on the hard ones. The defense against that
is not "trust the green checks." It's keeping a standing, public set of the cases that break it, so
the failure envelope is always visible next to the wins. A tool that only ships its successes is
asking for exactly the unexamined trust it claims to cure.

## The probes

Each probe is an **openly-constructed adversarial input** — fluent, hedged, plausible prose engineered
to carry **one** specific defect while staying clean on every other axis, so a flag on the target axis
is unambiguous signal. Six target one of assay's seven substance axes each; one targets the gaps
*between* the axes; and one — the priority — is real, not constructed.

| Probe | Targets | Defect |
|---|---|---|
| [`motte-and-bailey/`](motte-and-bailey/) | Equivocation | a key term retreats from a strong sense to a trivial one |
| [`reference-class/`](reference-class/) | Base rate / magnitude | a real number compared against a gamed reference class |
| [`hidden-premise/`](hidden-premise/) | Hidden premise | a conclusion valid only under an unstated load-bearing premise |
| [`unfalsifiable-dress/`](unfalsifiable-dress/) | Falsifiability | a claim no observation could disconfirm, in empirical dress |
| [`causal-narrative/`](causal-narrative/) | Causality vs correlation | a mechanism story laid over a single correlation |
| [`axis-gaps/`](axis-gaps/) | *(none — by design)* | category error, composition, survivorship — defects the seven axes don't name |
| [`laundering/`](laundering/) | **all three modes** | a **real** claim that is faithful + substantive + **false** at once |

Each probe directory contains:

- `claim.txt` — the input (for `laundering/`, also `summary.txt` + `PROVENANCE.md`)
- `DEFECT.md` — the engineered flaw, stated precisely; which axis should fire; which axes it is clean
  on; how it was built
- `EXPECTED.md` — the expected result as a **distribution**, not a golden verdict
- `results/` — committed calibration outputs, dated per model

### On "openly constructed" vs. the no-fabrication rule

assay's [`CLAUDE.md`](../../CLAUDE.md) forbids fabricating inputs — synthesizing a fake artifact and
passing it off as real. These probes do **not** violate that rule, because they are not standing in
for anything real: **the construction *is* the ground truth.** When `DEFECT.md` says "this contains a
motte-and-bailey," that is a fact about a thing we built on purpose, not a claim about the world that
could be wrong. The no-fabrication rule still bites in exactly one place — the laundering probe — where
the demonstration *requires* a real source, so a real source was [fetched and provenanced](laundering/PROVENANCE.md),
never invented. A constructed laundering claim would have the tool refuting a claim nobody made: worse
than no fixture at all.

## How to run one

You need the binary (`./build.sh` from the repo root) and `ANTHROPIC_API_KEY` set.

Single probe, by hand:

```sh
./assay -md examples/destructive/motte-and-bailey/claim.txt          # substance probes
```

The laundering probe is an audit (all three modes) and needs its source recreated first — it is
gitignored; see [`laundering/PROVENANCE.md`](laundering/PROVENANCE.md) for the one-liner:

```sh
./assay -md -audit -source examples/destructive/laundering/sources/ballmer_usatoday_2007.txt \
        examples/destructive/laundering/summary.txt
```

Calibrated N-run distribution (the way these are meant to be read):

```sh
cd examples/destructive
./run.sh                       # all probes, 10 runs each, on haiku
./run.sh motte-and-bailey 10   # one probe, 10 runs
./run.sh laundering 10 claude-sonnet-4-6
```

`run.sh` tallies the verdict distribution per probe, writes a dated summary into each `results/`, and
appends one line per probe to [`../../testing/calibration_log.jsonl`](../../testing/calibration_log.jsonl).

**These runs are calibration, never CI.** `run.sh` is deliberately *not* wired into `build.sh` or
`go test`. A flaky pass/fail gate on a non-deterministic judgment just trains everyone to ignore
failures (`../../TESTING.md`).

## How to read a distribution

Each run can come back differently — that is the point, not a bug. Read the **modal verdict** and the
**spread**, against the probe's `EXPECTED.md`:

- A probe "working" means its target axis fires in the **large majority** of runs (or the claim
  survives only as the honest narrow core). It does **not** mean every run agrees.
- The costly direction is the **false pass**: a defective claim rated `substantive` with no finding
  touching the engineered flaw. `EXPECTED.md` names the exact false-pass condition for each probe.
  One false pass on the laundering probe outweighs a hundred clean nominal runs.
- A clean run licenses one sentence only: *"the envelope held on this fixture, this model, this time."*
  Never "the critic is correct." The `axis-gaps/` probe has no "should fire" at all — a miss there is a
  **mapped limit**, recorded, not a bug.

## The training-data caveat (read this before trusting any clean result)

These probes are **public**, and several use famous defects and (in the laundering case) a famous
quote. Over time, a model may come to **recognise** a probe rather than **reason** about it — pattern-
matching "ah, the motte-and-bailey one" or "obviously the iPhone won." A clean result then measures
*recognition*, not the reasoning the probe was built to test, and the probe has quietly stopped testing
anything.

Two consequences, both load-bearing:

1. **Every result is dated and model-stamped.** A distribution from `claude-haiku-4-5` in June 2026 is
   a measurement of *that model at that time*, not a permanent property of assay. Re-run on new models.
2. **Treat a *too*-clean run on a famous probe with suspicion.** On `laundering/`, the diagnostic
   columns are faithfulness and substance (the hard, register-sensitive judgments), not the grounding
   refutation everyone "knows." When the easy column gets easier, lean on the hard ones.

The day these probes stop catching anything, the honest question is whether assay got better or whether
the probes got memorized. That question is part of the instrument.
