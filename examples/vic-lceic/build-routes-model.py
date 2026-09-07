#!/usr/bin/env python3
"""Model-guess route pass for the LCEIC Findings the regex could not decide.

One Sonnet call per row (the 23 `unclear` + 6 `mixed` rows from routes-prep.md,
carried in routes-model-input.jsonl), each given the Finding and its drawing
paragraph and asked for a route + the phrase it relied on. Writes routes-model.md
with a three-way table `id | regex | model | human` (human left blank) plus the
per-row model answers. This is the model's read, kept SEPARATE from the regex
pass — do not merge it into routes-prep.md / routes.md until the human column is in.

The key is never handled here: run it from the terminal that has it, e.g.

    source ./setkey.sh && python3 examples/vic-lceic/build-routes-model.py

Resumable: results stream to routes-model.jsonl as each row returns, and a rerun
skips ids already answered there. ~29 calls, roughly $0.30 on claude-sonnet-4-6.

Usage: build-routes-model.py [--model M] [--limit N] [--dry-run]
"""
import json
import os
import sys
import time
import urllib.request

HERE = os.path.dirname(os.path.abspath(__file__))
IN = os.path.join(HERE, "routes-model-input.jsonl")
JSONL = os.path.join(HERE, "routes-model.jsonl")
OUT = os.path.join(HERE, "routes-model.md")

API_URL = "https://api.anthropic.com/v1/messages"
MODEL = "claude-sonnet-4-6"
LIMIT = None
DRY = False
args = sys.argv[1:]
while args:
    a = args.pop(0)
    if a == "--model":
        MODEL = args.pop(0)
    elif a == "--limit":
        LIMIT = int(args.pop(0))
    elif a == "--dry-run":
        DRY = True
    else:
        sys.exit(f"unknown arg: {a}")

# Regex-class vocabulary, so the model column compares like-for-like with the
# regex column. The claims-machine route each maps to is named for the reader.
ROUTES = ["testimony", "opinion", "world", "data-gap"]
SYSTEM = (
    "You route a single Finding from a Victorian parliamentary committee report into exactly one "
    "class, judging ONLY from the Finding text and the report paragraph it is drawn from. Return the "
    "class and quote the shortest phrase (verbatim, from the Finding or the paragraph) you relied on.\n\n"
    "Classes:\n"
    "- testimony: the Finding compresses what witnesses or submitters said to the Committee "
    "(maps to route `source`; checkable by faithfulness against the transcripts/submissions).\n"
    "- opinion: the Finding is the Committee's own value judgment or call to act — disappointment, "
    "'should', a call for something (maps to route `evaluative`; not a fact about the world).\n"
    "- world: the Finding is a factual claim about the world that needs an external truth-maker — a "
    "dataset, annual report, statute, or official record (maps to route `evidence`).\n"
    "- data-gap: the Finding is a claim about a dataset's own limits, e.g. that a breakdown is not "
    "released (maps to route `data-gap`).\n\n"
    "If the paragraph both quotes a witness AND rests the number on an external dataset, choose by "
    "what the Finding's own subject asserts: a headline number/fact about the world is `world` even "
    "when a witness voiced it; a claim that only stakeholders hold is `testimony`. Do not hedge — "
    "pick the single best class."
)

TOOL = {
    "name": "route",
    "description": "Return the route for this Finding as JSON.",
    "input_schema": {
        "type": "object",
        "properties": {
            "route": {"type": "string", "enum": ROUTES},
            "phrase": {"type": "string",
                       "description": "verbatim phrase from the Finding or paragraph you relied on"},
            "reason": {"type": "string", "description": "one clause, why"},
        },
        "required": ["route", "phrase", "reason"],
        "additionalProperties": False,
    },
    "strict": True,
}


def prompt_for(row):
    return (f"Finding {row['id']} (§{row['section']}, printed p{row['page']}):\n{row['finding']}\n\n"
            f"Report paragraph it is drawn from:\n{row['paragraph']}\n\n"
            "Route this Finding.")


def call(row, key):
    body = json.dumps({
        "model": MODEL,
        "max_tokens": 400,
        "system": SYSTEM,
        "tools": [TOOL],
        "tool_choice": {"type": "tool", "name": "route"},
        "messages": [{"role": "user", "content": prompt_for(row)}],
    }).encode()
    req = urllib.request.Request(API_URL, data=body, method="POST")
    req.add_header("content-type", "application/json")
    req.add_header("x-api-key", key)
    req.add_header("anthropic-version", "2023-06-01")
    for attempt in range(4):
        try:
            with urllib.request.urlopen(req, timeout=150) as r:
                resp = json.loads(r.read())
            break
        except urllib.error.HTTPError as e:
            if e.code in (429, 500, 503, 529) and attempt < 3:
                time.sleep(2 ** attempt)
                continue
            return {"error": f"HTTP {e.code}: {e.read()[:200].decode(errors='replace')}"}
        except Exception as e:  # noqa: BLE001 — surface any transport error into the row
            return {"error": str(e)}
    for b in resp.get("content", []):
        if b.get("type") == "tool_use" and b.get("name") == "route":
            u = resp.get("usage", {})
            return {**b["input"],
                    "in_tok": u.get("input_tokens"), "out_tok": u.get("output_tokens")}
    return {"error": f"no tool_use in response: {json.dumps(resp)[:200]}"}


def main():
    rows = [json.loads(l) for l in open(IN, encoding="utf-8") if l.strip()]
    if LIMIT:
        rows = rows[:LIMIT]
    if DRY:
        print(SYSTEM + "\n\n" + "=" * 60 + "\n" + prompt_for(rows[0]))
        print(f"\n[dry-run] would call {MODEL} for {len(rows)} rows")
        return
    key = os.environ.get("ANTHROPIC_API_KEY")
    if not key:
        sys.exit("ANTHROPIC_API_KEY not set — run: source ./setkey.sh && python3 " + sys.argv[0])

    done = {}
    if os.path.exists(JSONL):
        for l in open(JSONL, encoding="utf-8"):
            if l.strip():
                d = json.loads(l)
                done[d["id"]] = d
    with open(JSONL, "a", encoding="utf-8") as jf:
        for i, row in enumerate(rows, 1):
            if row["id"] in done:
                print(f"[{i}/{len(rows)}] {row['id']} cached", file=sys.stderr)
                continue
            res = call(row, key)
            rec = {"id": row["id"], "regex_route": row["regex_route"], **res}
            jf.write(json.dumps(rec, ensure_ascii=False) + "\n")
            jf.flush()
            done[row["id"]] = rec
            print(f"[{i}/{len(rows)}] {row['id']} -> {res.get('route', res.get('error'))}",
                  file=sys.stderr)

    render([json.loads(l) for l in open(JSONL, encoding="utf-8") if l.strip()], rows)


CLAIMS = {"testimony": "source", "opinion": "evaluative", "world": "evidence", "data-gap": "data-gap"}


def render(results, rows):
    by_id = {r["id"]: r for r in results}
    rowmap = {r["id"]: r for r in rows}
    order = [r["id"] for r in rows]
    o = ["# routes-model.md — model-guess routes (Sonnet), regex-undecided rows only\n",
         f"Model: `{MODEL}`. One call per row over the 23 `unclear` + 6 `mixed` Findings from "
         "`routes-prep.md`, each given only the Finding and its drawing paragraph. This is the "
         "**model's** read; the regex pass in `routes-prep.md` is unchanged. **Do not merge either "
         "into `routes.md` / `claims-machine.txt` until the human route column is filled** — the "
         "point is a three-way disagreement record, not an averaged answer.\n",
         "Model class → claims-machine route: testimony→`source`, opinion→`evaluative`, "
         "world→`evidence`, data-gap→`data-gap`.\n",
         "## Three-way table\n",
         "| id | regex | model | model→claims | phrase model relied on | human |",
         "|----|-------|-------|--------------|------------------------|-------|"]
    for fid in order:
        r = by_id.get(fid, {})
        rt = r.get("route", "—")
        err = r.get("error")
        cell = f"**{rt}**" if not err else "_error_"
        claims = CLAIMS.get(rt, "—") if not err else "—"
        ph = (r.get("phrase", err or "")).replace("|", "\\|")[:60]
        o.append(f"| {fid} | {r.get('regex_route','?')} | {cell} | {claims} | {ph} |  |")
    o.append("\n---\n\n## Per-row model answers\n")
    for fid in order:
        r = by_id.get(fid, {})
        src = rowmap[fid]
        o.append(f"### {fid} — §{src['section']}, p{src['page']}  (regex: {src['regex_route']})\n")
        o.append(f"**Finding:** {src['finding']}\n")
        if r.get("error"):
            o.append(f"_model error: {r['error']}_\n")
            continue
        o.append(f"**Model route:** `{r.get('route')}`  → claims-machine `{CLAIMS.get(r.get('route'),'?')}`")
        o.append(f"**Phrase relied on:** “{r.get('phrase','')}”")
        o.append(f"**Reason:** {r.get('reason','')}\n")
    with open(OUT, "w", encoding="utf-8") as f:
        f.write("\n".join(o) + "\n")
    ok = sum(1 for r in results if not r.get("error"))
    print(f"wrote {OUT}: {ok}/{len(results)} routed, "
          f"{len(results)-ok} errors", file=sys.stderr)


if __name__ == "__main__":
    main()
