#!/usr/bin/env python3
# Tally the ce-scoped refuter. Usage: ce_tally.py <BASE>  where BASE has old/ and variant/ subtrees:
#   BASE/<arm>/motte/run-*/*.substance.jsonl
#   BASE/<arm>/ballmer/run-*/*.substance.jsonl
#   BASE/<arm>/unfalsifiable/run-*/*.substance.jsonl
import json,glob,re,sys,collections
BASE=sys.argv[1]
honest=re.compile(r"(system that learns|learns from feedback|adapts? its output).*intellig",re.I)
bailey=re.compile(r"genuinely intelligent|\bis intelligent\b",re.I)
rank={"substantive":2,"partial":1,"hollow":0,"error":-1}
def run_verdicts(arm,probe):
    out=[]
    for d in sorted(glob.glob(f"{BASE}/{arm}/{probe}/run-*")):
        recs=[json.loads(l) for f in glob.glob(f"{d}/*.substance.jsonl") for l in open(f)]
        out.append(recs)
    return out
def best(recs,pred):
    frs=[r for r in recs if pred(r["claim"])]
    if not frs: return None
    return max(frs,key=lambda r:rank.get(r["verdict"],-1))["verdict"]
def modal(recs,pred):
    vs=[r["verdict"] for r in recs if pred(r["claim"])]
    return collections.Counter(vs).most_common(1)[0][0] if vs else None
for arm in ("old","variant"):
    print(f"\n===== {arm} =====")
    # honest motte: best verdict per run; count runs with substantive-or-partial
    m=run_verdicts(arm,"motte")
    hv=[best(r,honest.search) for r in m]
    bv=[best(r,bailey.search) for r in m]
    hp=sum(1 for v in hv if v in ("substantive","partial"))
    hs=sum(1 for v in hv if v=="substantive")
    print(f"honest motte  runs={len(m)}  per-run={hv}  subst-or-partial={hp}/{len(m)}  substantive={hs}")
    print(f"bailey        modal={collections.Counter([v for v in bv if v]).most_common()}  per-run={bv}")
    # ballmer substance
    b=run_verdicts(arm,"ballmer")
    bl=[max(r,key=lambda x:rank.get(x['verdict'],-1))['verdict'] if r else None for r in b]
    bsp=sum(1 for r in b for x in r if x['verdict']=='substantive')  # any substantive fragment
    # per-run best on the core Ballmer prediction fragment (the 'no significant share' one)
    core=re.compile(r"significant.*share|market share|no chance",re.I)
    bcore=[best(r,core.search) or (max(r,key=lambda x:rank.get(x['verdict'],-1))['verdict'] if r else None) for r in b]
    bsub=sum(1 for v in bcore if v=="substantive")
    print(f"ballmer(core) runs={len(b)}  per-run={bcore}  substantive={bsub}/{len(b)}")
    # unfalsifiable
    u=run_verdicts(arm,"unfalsifiable")
    um=[modal(r,lambda c:True) for r in u]
    uleak=sum(1 for r in u for x in r if x['verdict']=='substantive')
    print(f"unfalsifiable per-run-modal={um}  substantive-leak-fragments={uleak}")
