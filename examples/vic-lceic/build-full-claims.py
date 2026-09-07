#!/usr/bin/env python3
"""Convert claims-machine.txt (pipe-delimited findings) → assay claims format (id \t path \t text \t
cites \t route). cites are the corpus-class documents (hearing/submission/attachment/qon) the report's
src= field names; external datasets (ABS, Creative Australia report, Audience Atlas, OzTAM, Throsby)
are NOT corpus-class and are excluded — they are the evidence-grounding mode's concern. route is
carried through verbatim from the source line's route= field (spec/TREE.md § Root node rendering)."""
import re, sys

SRC = "examples/vic-lceic/claims-machine.txt"
CH = {"2": "2=Chapter 2 — Overview of Victoria's cultural and creative industries",
      "3": "3=Chapter 3 — Government investment in Victoria's cultural and creative industries",
      "4": "4=Chapter 4 — National broadcasters"}

# witness surname / org keyword → hearing id (from the file's authoritative Appendix A.2 map).
HEAR = {
    "guglielmo": "2025-02-27/1_yarra-city-council", "blum": "2025-02-27/1_yarra-city-council",
    "yarra": "2025-02-27/1_yarra-city-council",
    "porter": "2025-02-27/2_regional-arts-victoria", "regional arts": "2025-02-27/2_regional-arts-victoria",
    "mullings": "2025-02-27/3_multicultural-arts-victoria", "multicultural arts": "2025-02-27/3_multicultural-arts-victoria",
    "tapley": "2025-02-27/4_abc", "gregson": "2025-02-27/4_abc",
    "febey": "2025-03-12/1_creative-victoria-and-vicscreen", "coffman": "2025-03-12/1_creative-victoria-and-vicscreen",
    "pitcher": "2025-03-12/1_creative-victoria-and-vicscreen", "creative victoria": "2025-03-12/1_creative-victoria-and-vicscreen",
    "vicscreen": "2025-03-12/1_creative-victoria-and-vicscreen",
    "barrie": "2025-03-12/2_community-music-victoria", "community music": "2025-03-12/2_community-music-victoria",
    "lowe": "2025-03-12/3_theatre-network-australia", "stitz": "2025-03-12/3_theatre-network-australia",
    "(tna)": "2025-03-12/3_theatre-network-australia",
    "collins": "2025-03-12/4_association-of-artist-managers-and-music-victoria",
    "packard": "2025-03-12/4_association-of-artist-managers-and-music-victoria",
    "music victoria": "2025-03-12/4_association-of-artist-managers-and-music-victoria",
    "irvine": "2025-03-12/5_sbs", "o'neil": "2025-03-12/5_sbs",
    "toulson": "2025-03-13/1_st-martins-and-theatre-works", "kostich": "2025-03-13/1_st-martins-and-theatre-works",
    "theatre works": "2025-03-13/1_st-martins-and-theatre-works", "st martins": "2025-03-13/1_st-martins-and-theatre-works",
    "allanson": "2025-03-13/2_arena-rawcus-lamama", "cornwell": "2025-03-13/2_arena-rawcus-lamama",
    "dullard": "2025-03-13/2_arena-rawcus-lamama", "(rawcus)": "2025-03-13/2_arena-rawcus-lamama",
    "la mama": "2025-03-13/2_arena-rawcus-lamama",
    "fielding": "2025-03-13/3_a-new-approach", "(a new approach)": "2025-03-13/3_a-new-approach",
    "anne robertson": "2025-03-13/4_public-galleries-association-of-victoria", "(pgav)": "2025-03-13/4_public-galleries-association-of-victoria",
    "camm": "2025-03-13/5_australian-museums-and-galleries-association-victoria",
    "ashley robertson": "2025-03-13/5_australian-museums-and-galleries-association-victoria",
    "amaga vic": "2025-03-13/5_australian-museums-and-galleries-association-victoria",
    "cunningham": "2025-03-13/6_bendigo-theatre-company-and-sertori-consulting", "vern wall": "2025-03-13/6_bendigo-theatre-company-and-sertori-consulting",
    "sertori": "2025-03-13/6_bendigo-theatre-company-and-sertori-consulting", "champion": "2025-03-13/6_bendigo-theatre-company-and-sertori-consulting",
    "bendigo": "2025-03-13/6_bendigo-theatre-company-and-sertori-consulting",
}
MONTHS = {"jan": "01", "feb": "02", "mar": "03", "apr": "04", "may": "05", "jun": "06",
          "jul": "07", "aug": "08", "sep": "09", "oct": "10", "nov": "11", "dec": "12"}
QON_ORG = [("abc", "abc"), ("sbs", "sbs"), ("creative victoria", "creative-victoria"),
           ("regional arts", "regional-arts-victoria"), ("community music", "community-music-victoria"),
           ("a new approach", "a-new-approach"), ("amaga", "amaga")]


def qon_date(src, i):
    # nearest "D Mon YYYY" or "Mon YYYY" after position i
    m = re.search(r"(\d{1,2})\s+([A-Za-z]{3})[a-z]*\.?\s+(\d{4})", src[i:i + 60])
    if m:
        return f"{m.group(3)}-{MONTHS.get(m.group(2).lower()[:3], '00')}-{int(m.group(1)):02d}"
    return "unknown"


def cites_for(src):
    low = src.lower()
    hearings = []
    for key, hid in HEAR.items():
        hc = "hearing:" + hid
        if key in low and hc not in hearings:
            hearings.append(hc)
    # submissions + attachments
    subs = []
    for m in re.finditer(r"submission (\d+)(?:\s+attachment\s+(\d+))?", low):
        n = int(m.group(1))
        subs.append(f"submission:{n}/attachment-{m.group(2)}" if m.group(2) else f"submission:{n}")
    # QoN responses
    qons = []
    for m in re.finditer(r"(response to (?:questions on notice|qon)|qons?\b|questions on notice)", low):
        # find org name in a window before the match
        w = low[max(0, m.start() - 40):m.start() + 10]
        org = next((slug for name, slug in QON_ORG if name in w), None)
        if org:
            qons.append(f"qon:{org}/{qon_date(low, m.start())}")
    out = list(dict.fromkeys(hearings + subs + qons))
    return out


def main():
    lines = open(SRC).read().split("\n")
    for ln in lines:
        if not re.match(r"^F\d", ln):
            continue
        parts = [p.strip() for p in ln.split("|")]
        if len(parts) < 5:
            continue
        cid = parts[0]
        route = parts[1][6:] if parts[1].startswith("route=") else parts[1]
        ref = parts[2]
        text = parts[3]
        src = parts[4][4:] if parts[4].startswith("src=") else parts[4]
        m = re.search(r"§(\d)\.(\d+)(?:\.(\d+))?", ref)
        if m:
            ch = m.group(1)
            sec = ".".join(x for x in m.groups() if x)
            path = f"{CH.get(ch, ch + '=Chapter ' + ch)}/{sec}=§{sec}"
        else:
            path = "unplaced"
        cites = " ".join(cites_for(src))
        sys.stdout.write(f"{cid}\t{path}\t{text}\t{cites}\t{route}\n")


if __name__ == "__main__":
    main()
