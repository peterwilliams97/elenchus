#!/usr/bin/env python3
"""
assay.py — a dialectical filter, run from the terminal.

Pipeline:
  1. DECOMPOSE (grammar)   extract atomic, independently-evaluable claims
  2. DIALECTIC (logic)     per claim, run a producer<->critic loop:
       - Producer steelmans the claim, BLIND to the critique axes
         (pattern rule: the producer must not see the critic's instructions)
       - Critic attacks on FIXED AXES, returns structured findings + verdict
       - Loop on the surviving claim until the critic raises no new
         fatal/weakening finding, or --max-rounds is reached (convergence)
  3. FILTER (rhetoric)     keep the substantive residue; show what dissolved

Every Claude call is printed AS IT IS MADE: stage, system persona, user
prompt, and the response streamed token-by-token. Calls run sequentially so
the terminal output stays readable.

Usage:
  export ANTHROPIC_API_KEY=sk-ant-...
  python3 assay.py                      # runs the built-in example
  python3 assay.py path/to/prose.txt    # assay a file
  python3 assay.py --text "Our AI-first transformation will ..."
  echo "some prose" | python3 assay.py  # read stdin
  python3 assay.py --model claude-opus-4-8 --max-rounds 3 --no-color

Requires:  pip install anthropic
"""

import argparse
import json
import os
import re
import sys
import textwrap

try:
    import anthropic
except ImportError:
    sys.exit("Missing dependency. Install with:  pip install anthropic")

# ── config ───────────────────────────────────────────────────────────────────

MODEL = os.environ.get("ANTHROPIC_MODEL", "claude-sonnet-4-6")
MAX_TOKENS = 1024

AXES = [
    "Evidence",
    "Hidden premise",
    "Falsifiability",
    "Equivocation",
    "Base rate / magnitude",
    "Counterexample",
    "Causality vs correlation",
]

DEFAULT_INPUT = (
    "Our AI-first transformation will unlock unprecedented synergies across the "
    "organization. By leveraging best-in-class machine learning, we will become "
    "the market leader within 18 months. Customer-centricity is in our DNA, and "
    "the data shows that engaged customers spend 3x more. We must move fast "
    "because the window is closing. Empowering our people to think outside the "
    "box will build a culture of innovation that competitors simply cannot "
    "replicate."
)

# ── terminal colour ────────────────────────────────────────────────────────────

USE_COLOR = sys.stdout.isatty()


def c(code, s):
    return f"\033[{code}m{s}\033[0m" if USE_COLOR else s


def dim(s):      return c("2", s)
def bold(s):     return c("1", s)
def green(s):    return c("32", s)
def yellow(s):   return c("33", s)
def red(s):      return c("31", s)
def cyan(s):     return c("36", s)
def grey(s):     return c("90", s)


VERDICT_COLOR = {
    "substantive": green,
    "partial": yellow,
    "hollow": red,
    "error": grey,
}
SEV_COLOR = {"fatal": red, "weakens": yellow, "clears": green}

GROUND_COLOR = {
    "grounded": green,
    "partial": yellow,
    "overstated": yellow,
    "unsupported": red,
    "contradicted": red,
    "error": grey,
}

CALL_NO = 0
VERBOSE = False  # -v/--verbose: show every call's prompts + raw streamed response

# ── Claude plumbing ──────────────────────────────────────────────────────────

client = None  # constructed in main(), after arg parsing


def call_claude(system, prompt, label):
    """One streaming call. In verbose mode the whole exchange is printed as it
    happens; in default mode the call is silent (the dialectic results are
    rendered by the stage functions instead)."""
    global CALL_NO
    CALL_NO += 1
    if VERBOSE:
        print()
        print(cyan(f"┌─ call #{CALL_NO} · {label} · {MODEL}"))
        print(grey("│ system:"))
        for line in system.splitlines():
            print(grey("│   " + line))
        print(grey("│ user:"))
        for line in prompt.splitlines():
            print(grey("│   " + line))
        print(cyan("├─ response:"))
        sys.stdout.write("│ ")
        sys.stdout.flush()

    parts = []
    with client.messages.stream(
        model=MODEL,
        max_tokens=MAX_TOKENS,
        system=system,
        messages=[{"role": "user", "content": prompt}],
    ) as stream:
        for chunk in stream.text_stream:
            if VERBOSE:
                # keep the left gutter on newlines so streamed output stays aligned
                sys.stdout.write(chunk.replace("\n", "\n│ "))
                sys.stdout.flush()
            parts.append(chunk)
    if VERBOSE:
        print()
        print(cyan("└─"))
    return "".join(parts)


def _strip_fences(t):
    return re.sub(r"```(?:json)?", "", t).strip()


def _extract_balanced(t):
    """Return the first balanced {...} or [...] block, respecting strings."""
    ib, ia = t.find("{"), t.find("[")
    if ib == -1:
        start = ia
    elif ia == -1:
        start = ib
    else:
        start = min(ia, ib)
    if start == -1:
        return None
    open_ch = t[start]
    close_ch = "}" if open_ch == "{" else "]"
    depth = 0
    in_str = esc = False
    for i in range(start, len(t)):
        ch = t[i]
        if in_str:
            if esc:
                esc = False
            elif ch == "\\":
                esc = True
            elif ch == '"':
                in_str = False
            continue
        if ch == '"':
            in_str = True
        elif ch == open_ch:
            depth += 1
        elif ch == close_ch:
            depth -= 1
            if depth == 0:
                return t[start : i + 1]
    return t[start:]


def parse_json(text):
    cleaned = _strip_fences(text or "")

    def attempt(s):
        return json.loads(re.sub(r",\s*([}\]])", r"\1", s))  # drop trailing commas

    try:
        return attempt(cleaned)
    except Exception:
        sliced = _extract_balanced(cleaned)
        if sliced:
            return attempt(sliced)
        raise ValueError("Could not parse JSON from model output")


def call_json(system, prompt, label, retries=1):
    out = call_claude(system, prompt, label)
    try:
        return parse_json(out)
    except Exception:
        if retries > 0:
            strict = (
                system
                + "\n\nCRITICAL: Output ONLY raw JSON. No prose, no markdown, "
                "no backticks. Your first character must be { or [."
            )
            return call_json(strict, prompt, label + " (retry)", retries - 1)
        raise


# ── pipeline stages ──────────────────────────────────────────────────────────

def decompose(source):
    system = (
        "You are a claims extractor trained in analytic philosophy. Break prose "
        "into its atomic, independently-evaluable assertions. Strip rhetoric, "
        "hedges, and connective filler. Each item must be a single claim that "
        "could in principle be true or false. Do not evaluate them. Return ONLY "
        "a JSON array of strings, no markdown, no preamble."
    )
    arr = call_json(system, f"TEXT:\n{source}", "decompose")
    return [s for s in arr if isinstance(s, str)] if isinstance(arr, list) else []


def produce(claim):
    """Producer — blind to the critique axes."""
    system = (
        "You are the Producer. Given a single claim, construct its STRONGEST "
        "defensible version (steelman) and state precisely what would have to be "
        "true for it to hold. Be concrete. You are NOT evaluating or criticising "
        'the claim — only making the best honest case for it. Return ONLY JSON: '
        '{"steelman": string, "conditions": string}. No markdown.'
    )
    return call_json(system, f"CLAIM:\n{claim}", "producer (steelman)")


def critique(claim, steelman, conditions):
    """Critic — works only from the fixed axes."""
    system = (
        "You are the Critic. Assess one claim against these FIXED AXES, in order:\n"
        "- Evidence: is support cited or available, or is it bare assertion?\n"
        "- Hidden premise: what unstated assumption must hold?\n"
        "- Falsifiability: what observation would show it false? If none exists, it is vacuous.\n"
        "- Equivocation: does a key term shift meaning or hide behind a buzzword?\n"
        "- Base rate / magnitude: is there a real quantity and a comparison, or just a direction?\n"
        "- Counterexample: is there an obvious case where it fails?\n"
        "- Causality vs correlation: does it assert cause from mere association?\n\n"
        "For each axis that bears on the claim, give a one-sentence finding and a "
        'severity: "fatal" (this axis alone guts the claim), "weakens" (survives '
        'only in narrower form), "clears" (no problem here).\n\n'
        "Then deliver a verdict:\n"
        '- "hollow": unfalsifiable, equivocating, or pure assertion with no '
        "defensible core.\n"
        '- "partial": a narrower, qualified claim survives after stripping the '
        "unsupported parts.\n"
        '- "substantive": falsifiable, evidence exists or is clearly obtainable, '
        "no equivocation, survives counterexample.\n\n"
        'If "partial" or "substantive", give the surviving_claim (the exact '
        'narrowed/defensible version). If "hollow", surviving_claim is null.\n'
        "Set needs_another_round=true ONLY if a narrower surviving_claim was "
        "produced that itself deserves a fresh pass.\n\n"
        "Return ONLY JSON:\n"
        '{"critique":[{"axis":string,"finding":string,"severity":"fatal"|'
        '"weakens"|"clears"}],"verdict":"substantive"|"partial"|"hollow",'
        '"surviving_claim":string|null,"reason":string,'
        '"needs_another_round":boolean}'
    )
    user = (
        f"CLAIM:\n{claim}\n\nPRODUCER STEELMAN:\n{steelman}\n\n"
        f"PRODUCER CONDITIONS:\n{conditions}"
    )
    return call_json(system, user, "critic")


def _render_finding(f):
    sev = f.get("severity", "weakens")
    axis = f.get("axis", "")
    finding = textwrap.fill(
        f.get("finding", ""), width=100, subsequent_indent=" " * 40
    )
    print("      " + SEV_COLOR.get(sev, yellow)(f"{sev:<8}") + dim(f"{axis:<26}") + finding)


def render_round(n, steelman, cr):
    """Default-mode view of one producer<->critic round: the dialectical moves a
    human would make and read, with the prompt machinery left out."""
    print(bold(f"    round {n}"))
    if steelman:
        print(dim("      steelman  ") + textwrap.fill(steelman, width=100, subsequent_indent=" " * 16))
    for f in cr.get("critique") or []:
        _render_finding(f)
    verdict = cr.get("verdict", "?")
    print("      " + VERDICT_COLOR.get(verdict, grey)(f"verdict: {verdict.upper()}"), flush=True)
    if cr.get("reason"):
        print(dim(textwrap.fill(cr["reason"], width=100,
                                initial_indent=" " * 8, subsequent_indent=" " * 8)), flush=True)
    sc = cr.get("surviving_claim")
    if sc and verdict != "substantive":
        print(yellow(textwrap.fill(f'survives as: "{sc}"', width=100,
                                   initial_indent=" " * 6, subsequent_indent=" " * 8)), flush=True)


def assay_claim(claim, max_rounds):
    """One claim through the producer<->critic loop. Never raises."""
    try:
        current = claim
        rounds = 0
        last = None
        steelman_shown = None
        while rounds < max_rounds:
            p = produce(current)
            if rounds == 0:
                steelman_shown = p.get("steelman")
            cr = critique(current, p.get("steelman", ""), p.get("conditions", ""))
            last = cr
            rounds += 1
            if not VERBOSE:
                render_round(rounds, p.get("steelman"), cr)
            if (
                cr.get("needs_another_round")
                and cr.get("surviving_claim")
                and rounds < max_rounds
            ):
                current = cr["surviving_claim"]
                if not VERBOSE:
                    print(dim("    → re-running on the narrowed claim"))
                continue
            break
        result = {"claim": claim, "steelman": steelman_shown, "rounds": rounds}
        result.update(last or {})
        return result
    except Exception as e:
        return {
            "claim": claim,
            "steelman": None,
            "rounds": 0,
            "critique": [],
            "verdict": "error",
            "surviving_claim": None,
            "reason": f"assay failed: {e}",
        }


# ── report ───────────────────────────────────────────────────────────────────

def print_method(max_rounds):
    print(bold("\nTHE ASSAY — a dialectical filter"))
    print(dim("decompose → producer–critic → filter\n"))
    print(dim("fixed critique axes:  " + "  ·  ".join(AXES)))
    print(
        dim(
            "convergence rule:     producer steelmans (blind to axes) → critic "
            f"attacks on axes → loop on the surviving claim until no new "
            f"fatal/weakening finding, or {max_rounds} rounds."
        )
    )


def print_report(results):
    survivors = [r for r in results if r["verdict"] in ("substantive", "partial")]
    dissolved = [r for r in results if r["verdict"] == "hollow"]
    errors = [r for r in results if r["verdict"] == "error"]

    print(bold("\n" + "═" * 70))
    print(
        bold(
            f"{green(str(len(survivors)))} of {len(results)} claims survive the "
            f"dialectic · {red(str(len(dissolved)))} dissolve"
            + (f" · {grey(str(len(errors)) + ' unparsed')}" if errors else "")
        )
    )
    print(bold("═" * 70))

    # In default mode the per-round detail was already shown live, so the
    # report is just the headline + residue. Verbose mode prints the full
    # per-claim breakdown here since its live output was raw JSON.
    for r in (results if VERBOSE else []):
        col = VERDICT_COLOR.get(r["verdict"], grey)
        glyph = {"substantive": "✓", "partial": "≈", "hollow": "✕", "error": "?"}.get(
            r["verdict"], "?"
        )
        print("\n" + col(f"{glyph} [{r['verdict'].upper()}] ") + r["claim"])
        if r.get("surviving_claim") and r["verdict"] != "substantive":
            print(yellow(f"    survives as: “{r['surviving_claim']}”"))
        for cr in r.get("critique") or []:
            sev = cr.get("severity", "weakens")
            sc = SEV_COLOR.get(sev, yellow)
            print(
                "    "
                + sc(f"{sev:<8}")
                + dim(f"{cr.get('axis',''):<24}")
                + cr.get("finding", "")
            )
        if r.get("reason"):
            print(dim(f"    verdict — {r['reason']}"))
        if r.get("rounds", 0) > 1:
            print(grey(f"    converged in {r['rounds']} rounds"))

    if survivors:
        print(green(bold("\nThe residue:")))
        for r in survivors:
            print(green("  — " + (r.get("surviving_claim") or r["claim"])))
    print()


# ── grounding mode (is summary A grounded by source B?) ──────────────────────

def split_summary(text):
    """Split an authored summary into its claims as written, preserving the
    author's items (handles numbered lists, including run-together ones)."""
    text = text.strip()
    parts = re.split(r'(?:(?<=[\.\?\!"\)\u201d])|^|\n)\s*\d{1,2}[\.\)]\s*', text)
    parts = [p.strip() for p in parts if p.strip()]
    if len(parts) >= 2:
        return parts
    parts = [p.strip() for p in re.split(r'(?:^|\n)\s*[-*\u2022]\s+', text) if p.strip()]
    if len(parts) >= 2:
        return parts
    lines = [l.strip() for l in text.splitlines() if l.strip()]
    return lines if len(lines) >= 2 else [text]


def ground_defender(claim, source):
    """Find the strongest verbatim support for the claim in the source only."""
    system = (
        "You are the Defender. You are given a SUMMARY CLAIM and a SOURCE "
        "transcript. Find the STRONGEST evidence in the SOURCE that the speaker "
        "actually asserts this claim. Quote spans VERBATIM from the SOURCE only — "
        "never paraphrase, never use outside knowledge. If there is no support, "
        'set found=false and quotes=[]. Return ONLY JSON: '
        '{"found": boolean, "quotes": [string], "best_case": string}. No markdown.'
    )
    user = f"SUMMARY CLAIM:\n{claim}\n\nSOURCE:\n{source}"
    return call_json(system, user, "grounding defender")


def ground_critic(claim, source, defender):
    """Judge whether the cited evidence supports the claim AS STATED."""
    system = (
        "You are the Grounding Critic. Decide whether the SUMMARY CLAIM is "
        "faithfully supported by the SOURCE, using the Defender's cited quotes. "
        "Watch specifically for the ways a summary distorts a source:\n"
        "- Fabrication: the claim is simply not in the source.\n"
        "- Overstatement: the source hedged or qualified it; the summary made it "
        "absolute ('kind of over' → 'over'; 'I might' → 'will').\n"
        "- Distortion: the meaning was changed.\n"
        "- Context-stripping: a conditional or hypothetical is presented as an "
        "unconditional belief.\n"
        "- Misattribution: the speaker was quoting or steelmanning someone else, "
        "and the summary attributes it as the speaker's own view.\n"
        "- Cherry-pick: technically present but unrepresentative of the source's "
        "overall stance.\n\n"
        "Verify the Defender's quotes actually appear to come from the source and "
        "actually support the claim; do not take the Defender's word for it.\n\n"
        "Verdict:\n"
        '- "grounded": the source asserts the claim as stated.\n'
        '- "partial": the source supports a weaker/narrower version.\n'
        '- "overstated": supported in kind but the summary strengthened it.\n'
        '- "unsupported": not present in the source.\n'
        '- "contradicted": the source says the opposite or materially against it.\n\n'
        "For 'partial' or 'overstated', give what_source_actually_says (the faithful "
        "version). Return ONLY JSON:\n"
        '{"findings":[{"mode":string,"finding":string}],"verdict":"grounded"|'
        '"partial"|"overstated"|"unsupported"|"contradicted","evidence":string,'
        '"what_source_actually_says":string|null,"note":string}'
    )
    quotes = "\n".join(f"- “{q}”" for q in defender.get("quotes", [])) or "(none)"
    user = (
        f"SUMMARY CLAIM:\n{claim}\n\nDEFENDER FOUND SUPPORT: {defender.get('found')}\n"
        f"DEFENDER QUOTES:\n{quotes}\n\nSOURCE:\n{source}"
    )
    return call_json(system, user, "grounding critic")


def render_grounding(claim, cr):
    verdict = cr.get("verdict", "?")
    col = GROUND_COLOR.get(verdict, grey)
    glyph = {"grounded": "✓", "partial": "≈", "overstated": "▲",
             "unsupported": "✕", "contradicted": "⚡", "error": "?"}.get(verdict, "?")
    print("\n" + col(f"{glyph} [{verdict.upper()}] ") + claim)
    for f in cr.get("findings") or []:
        mode = f.get("mode", "")
        print("      " + dim(f"{mode:<18}")
              + textwrap.fill(f.get("finding", ""), width=100, subsequent_indent=" " * 24))
    if cr.get("evidence"):
        print(dim(textwrap.fill("source: " + cr["evidence"], width=100,
                                initial_indent=" " * 6, subsequent_indent=" " * 6)))
    if cr.get("what_source_actually_says"):
        print(yellow(textwrap.fill('faithful version: "' + cr["what_source_actually_says"] + '"',
                                   width=100, initial_indent=" " * 6, subsequent_indent=" " * 8)))


def ground_claim(claim, source):
    """Never raises."""
    try:
        d = ground_defender(claim, source)
        cr = ground_critic(claim, source, d)
        cr["claim"] = claim
        return cr
    except Exception as e:
        return {"claim": claim, "findings": [], "verdict": "error",
                "evidence": None, "what_source_actually_says": None,
                "note": f"grounding failed: {e}"}


def print_ground_report(results):
    n = len(results)
    grounded = [r for r in results if r["verdict"] == "grounded"]
    failed = [r for r in results
              if r["verdict"] in ("overstated", "unsupported", "contradicted")]
    fully = len(grounded) == n and n > 0
    print(bold("\n" + "═" * 70))
    print(bold("Is the summary fully grounded by the source?  ")
          + (green("YES") if fully else red("NO")))
    print(
        f"{green(str(len(grounded)))} grounded · "
        f"{yellow(str(len([r for r in results if r['verdict']=='partial'])))} partial · "
        f"{yellow(str(len([r for r in results if r['verdict']=='overstated'])))} overstated · "
        f"{red(str(len([r for r in results if r['verdict']=='unsupported'])))} unsupported · "
        f"{red(str(len([r for r in results if r['verdict']=='contradicted'])))} contradicted"
        + (f" · {grey(str(len([r for r in results if r['verdict']=='error'])) + ' error')}"
           if any(r['verdict'] == 'error' for r in results) else "")
    )
    print(bold("═" * 70))
    if failed:
        print(red(bold("\nWhere the summary departs from the source:")))
        for r in failed:
            print(red(f"  {r['verdict'].upper()}: ") + r["claim"])
            if r.get("what_source_actually_says"):
                print(dim('      source actually supports: "'
                          + r["what_source_actually_says"] + '"'))
    print()


def run_grounding(summary_text, source, max_rounds):
    claims = split_summary(summary_text)
    print(bold(f"\nsummary contains {len(claims)} claims; checking each against "
               f"a {len(source.split())}-word source\n"))
    results = []
    for i, cl in enumerate(claims, 1):
        print(bold(f"▸ claim {i}/{len(claims)}: {cl}"))
        r = ground_claim(cl, source)
        results.append(r)
        if not VERBOSE:
            render_grounding(cl, r)
    print_ground_report(results)


# ── main ─────────────────────────────────────────────────────────────────────

def read_source(args):
    if args.text:
        return args.text
    if args.path:
        with open(args.path, "r", encoding="utf-8") as f:
            return f.read()
    if not sys.stdin.isatty():
        piped = sys.stdin.read().strip()
        if piped:
            return piped
    return DEFAULT_INPUT


def main():
    global USE_COLOR, MODEL, client, VERBOSE

    parser = argparse.ArgumentParser(description="Dialectical filter for prose.")
    parser.add_argument("path", nargs="?", help="text file to assay")
    parser.add_argument("--text", help="inline text to assay")
    parser.add_argument("--model", help=f"model id (default {MODEL})")
    parser.add_argument(
        "--max-rounds", type=int, default=2, help="producer–critic rounds per claim"
    )
    parser.add_argument(
        "--source",
        help="source/transcript file → grounding mode: check whether the summary "
        "(path/--text/stdin) is faithfully grounded by this source",
    )
    parser.add_argument(
        "-v", "--verbose", action="store_true",
        help="show every Claude call: prompts and raw streamed response",
    )
    parser.add_argument("--no-color", action="store_true", help="disable ANSI colour")
    args = parser.parse_args()

    VERBOSE = args.verbose
    if args.no_color:
        USE_COLOR = False
    if args.model:
        MODEL = args.model

    if not os.environ.get("ANTHROPIC_API_KEY"):
        sys.exit("Set ANTHROPIC_API_KEY in your environment first.")
    client = anthropic.Anthropic()

    # ── grounding mode ──
    if args.source:
        with open(args.source, "r", encoding="utf-8") as f:
            src = f.read()
        if args.text:
            summary = args.text
        elif args.path:
            with open(args.path, "r", encoding="utf-8") as f:
                summary = f.read()
        elif not sys.stdin.isatty():
            summary = sys.stdin.read()
        else:
            sys.exit(
                "Grounding mode needs a summary too. Provide it as a file, --text, "
                "or stdin, plus --source TRANSCRIPT."
            )
        print(bold("\nTHE ASSAY — grounding check"))
        print(dim("is summary A faithfully grounded by source B?"))
        run_grounding(summary, src, args.max_rounds)
        return

    source = read_source(args)
    print_method(args.max_rounds)
    print(dim("\nsource:\n  " + source.replace("\n", "\n  ")))

    print(bold("\n── stage 1: decompose ──"))
    claims = decompose(source)
    if not claims:
        sys.exit("No claims extracted.")
    print(bold(f"\nextracted {len(claims)} atomic claims:"))
    for i, cl in enumerate(claims, 1):
        print(f"  {i}. {cl}")

    print(bold("\n── stage 2: dialectic (producer ↔ critic) ──"))
    results = []
    for i, cl in enumerate(claims, 1):
        print(bold(f"\n▸ claim {i}/{len(claims)}: {cl}"))
        results.append(assay_claim(cl, args.max_rounds))

    print(bold("\n── stage 3: filter ──"))
    print_report(results)


if __name__ == "__main__":
    main()
