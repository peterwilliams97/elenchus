#!/usr/bin/env python3
"""Deterministic route pre-pass for the LCEIC report's 55 Findings.

For each Finding: pull the body paragraph it is drawn from, mark the sentence
carrying the Finding's subject, list the footnotes in that paragraph's section,
and run three fixed regex classes over the prose and footnotes. NO MODEL decides
any route.

  testimony : the Committee heard | told the Committee | witnesses | submitted
              that | evidence to the inquiry
  opinion   : the Committee believes/considers | notes with concern |
              is disappointed | calls for | recommends | should
  world     : a footnote to a dataset / annual report / statute, no witness named

Suggested route = the single class of the phrases found; "mixed" if >1 class,
"unclear" if none. The deciding phrase is printed. Route scope = the Finding's
own paragraph(s) + the footnotes those paragraphs cite.

Usage: build_routes_prep.py report-raw.txt routes-prep.md
"""
import re
import sys
import unicodedata

RAW, OUT = sys.argv[1], sys.argv[2]

# ── Finding table (id, page, section, verbatim text) — from report-summary.md ──
# section is the report's own §; findings sharing a § form a cluster over shared prose.
FINDINGS = [
    ("F1", 11, "2.1.1", "The Victorian cultural and creative industries are economically significant, contributing $40.5 billion to the Victorian economy in 2022–23."),
    ("F2", 11, "2.1.1", "The Victorian cultural and creative industries are an important employer, employing more than 320,000 Victorians or almost 9 per cent of total employment in the state."),
    ("F3", 11, "2.1.1", "The Victorian cultural and creative industries provide significant economic stimulus to the state by attracting visitation and tourism."),
    ("F4", 11, "2.1.1", "Victoria is a significant cultural audience, which rated higher for cultural participation than all other jurisdictions surveyed, including New York, Sweden and the United Kingdom."),
    ("F5", 14, "2.1.2", "Engagement with the Victorian cultural and creative industries brings communities together, breaks down barriers between different groups within society and encourages greater communication and cohesion."),
    ("F6", 14, "2.1.2", "There is a growing body of evidence that recognises the crucial role arts and creativity can play in promoting positive mental health and wellbeing outcomes."),
    ("F7", 18, "2.2.1", "The COVID-19 pandemic severely damaged the Victorian cultural and creative industries. It significantly limited opportunities to present and engage with works publicly, and consequently restricted practitioners' income."),
    ("F8", 18, "2.2.1", "The COVID-19 pandemic took away critical training opportunities from those looking to enter the Victorian cultural and creative industries during this time."),
    ("F9", 18, "2.2.1", "Exclusion from COVID-19 financial support programs exacerbated the financial issues experienced by many in the industry and created ongoing skill gaps in the industry."),
    ("F10", 18, "2.2.1", "Children and young people's creative participation and attendance between 2017–18 and 2021–22 has significantly declined."),
    ("F11", 18, "2.2.1", "Several live music venues have closed following the COVID-19 pandemic due to various pressures, such as rising costs and inconsistent financial support."),
    ("F12", 20, "2.2.2", "The COVID-19 pandemic led to worsening mental health for many in the Victorian cultural and creative industries, particularly amongst children and young people."),
    ("F13", 21, "2.2.3", "Since the COVID-19 pandemic, the costs to produce and present work have significantly increased."),
    ("F14", 23, "2.2.4", "It has never been harder for Victorians to make a living in the cultural and creative industries."),
    ("F15", 23, "2.2.4", "Insecure or low wages impacts retention in the industries and discourages people from pursuing a career in the industries."),
    ("F16", 26, "2.2.5", "Many people who want to engage with the Victorian cultural and creative industries cannot afford to do so due to the cost of living crisis."),
    ("F17", 26, "2.2.5", "Ticket prices are a key barrier for people wanting to engage with the Victorian cultural and creative industries."),
    ("F18", 26, "2.2.5", "The COVID-19 pandemic and the cost of living crisis has accelerated the decline of volunteerism in the Victorian cultural and creative industries."),
    ("F19", 27, "2.2.6", "The COVID-19 pandemic exposed and intensified existing systemic issues within the Victorian cultural and creative industries, such as barriers to participation for women and gender diverse people, particularly those with caring responsibilities and long-term financial precarity."),
    ("F20", 37, "3.2.1", "Failing to index Victorian cultural and creative industries funding has led to a decrease in financial support in real terms."),
    ("F21", 38, "3.2.2", "Grant and reporting requirements can be very onerous, particularly on individuals and organisations who are made up of volunteers or part-time staff."),
    ("F22", 40, "3.2.3", "Unpredictable grant programs cause instability in the industries and prevents forward planning and investments in longer-term projects and strategies."),
    ("F23", 41, "3.2.4", "Infrequent grant programs can negatively impact the industries by leaving greater numbers of applicants without government support for longer periods of time."),
    ("F24", 42, "3.2.5", "A low maximum grant cap of $20,000 in some grant programs limits what recipients can achieve and may necessitate multiple grant applications to fund a project, consuming significant time and resources of applicants."),
    ("F25", 43, "3.2.6", "There needs to be a balance between investment in infrastructure and investment in artists. Without this balance, Victoria runs the risk of being dependent on interstate and international artists to fill these venues."),
    ("F26", 44, "3.2.7", "By requiring a developed concept or project, grant programs exclude creatives who are unable to undertake significant unpaid ideation work."),
    ("F27", 45, "3.2.8", "The role that each level of government plays could be better defined to assist the industries in understanding where to go for relevant investment and support."),
    ("F28", 46, "3.2.9", "People should not need to leave regional Victoria to engage with its cultural and creative industries."),
    ("F29", 57, "3.3.5", "Victoria receives its fair share of funding from Creative Australia."),
    ("F30", 57, "3.3.5", "It is difficult to determine whether Victoria receives its fair share of other funding captured in the cultural funding by government dataset, as state and territory breakdowns are not released."),
    ("F31", 59, "3.3.6", "Regional Victoria does not receive its fair share of funding from Creative Australia."),
    ("F32", 59, "3.3.6", "It is difficult to determine whether regional Victoria receives its fair share of other funding captured in the cultural funding by government dataset, as remoteness level or geography breakdowns are not released."),
    ("F33", 69, "4.3.1", "No state or territory has an ABC headcount that accurately reflects its population size; an overrepresentation or underrepresentation of ABC employee headcount exists to varying degrees in all states and territories."),
    ("F34", 69, "4.3.1", "The ABC's decision to relocate its Ultimo office to Parramatta was influenced by multiple external factors, including legal obligations under the ABC Enterprise Agreement 2022–2025, budget constraints that made New South Wales the most cost-effective option, the Parramatta office's proximity to key state departments and agencies, and a strategic goal to better reflect New South Wales' population demographics."),
    ("F35", 69, "4.3.1", "It is disappointing that both of Australia's national broadcasters the ABC and SBS expanded their presence to Western Sydney instead of Victoria."),
    ("F36", 69, "4.3.1", "The ABC's impact and ability to culturally represent any particular state goes beyond the headcount or location of ABC offices."),
    ("F37", 74, "4.3.2", "The Committee calls on the Victorian Government to continue to advocate to the Federal Government and the ABC for the return of a Victorian 7:30 Report."),
    ("F38", 74, "4.3.2", "ABC spending is unevenly distributed across all states and territories when compared to their population sizes. No state or territory receives funding proportionate to its share of Australia's population."),
    ("F39", 74, "4.3.2", "The ABC's content production decisions are shaped by state screen agency policies and funding availability. External pressures, including rising production costs and a 14% real-term reduction in the ABC's budget over the past decade, have constrained the broadcaster's capacity to create content and ability to expand its footprint in the nation's states and territories."),
    ("F40", 76, "4.4.1", "No state or territory reflects an SBS FTE employee number that is proportionate to its share of the Australia's population."),
    ("F41", 76, "4.4.1", "The geographic distribution of SBS staff does not reflect the availability of its content. While a significant portion of SBS's workforce is based in New South Wales, its services are national and platform-agnostic, ensuring all Australians, regardless of location, can access and benefit from SBS content."),
    ("F42", 78, "4.4.1", "The SBS's decision to relocate to Western Sydney was determined by funding requirements set by the Federal Government."),
    ("F43", 78, "4.4.1", "It is unclear why Victoria was not considered as an option for the Federal Government's relocation feasibility study for the SBS."),
    ("F44", 80, "4.4.2", "SBS provides significant financial and creative investment to Victoria."),
    ("F45", 80, "4.4.2", "SBS's content remains nationally accessible regardless of the production location."),
    ("F46", 82, "4.5.1", "Between 2020 and 2025, the ABC commissioned 75 external and 52 internal projects where the majority of production took place in Victoria, demonstrating its vital role in supporting the state's screen industry."),
    ("F47", 83, "4.5.1", "The ABC partners with major Victorian cultural institutions to create educational content and showcase Victorian talent."),
    ("F48", 83, "4.5.1", "Through initiatives like VicScreen internship and Indigenous placements, the ABC actively develops Victoria creative and cultural workforce."),
    ("F49", 84, "4.5.1", "The ABC delivers 16 regional radio programs in Victoria which produces over 112 hours of weekly local content across 9 stations, ensuring broad community representation."),
    ("F50", 85, "4.5.1", "One of the main challenges for the ABC in producing content in regional Victoria is the lack of independent production sector proposals for shows to be made in regional Victoria."),
    ("F51", 85, "4.5.1", "The cost of producing content in regional Victoria requires significant financial investment."),
    ("F52", 85, "4.5.1", "Producing content in regional areas delivers significant benefits, including social benefits, such as better representation and celebration of local communities and economic benefits, like job creation and increased local spending."),
    ("F53", 87, "4.5.2", "SBS's receives about one-third of the ABC's funding, which restricts its ability to produce and commission content for and in Victoria."),
    ("F54", 87, "4.5.2", "SBS sustains substantial Melbourne-based operations, producing language programs that reach 3.42 million Victorian viewers monthly, while supporting diverse communities through partnerships with VicScreen and local creatives."),
    ("F55", 87, "4.5.2", "The SBS deeply engages with Victoria's multilingual and multicultural communities by supporting community events and major cultural events in Victoria."),
]

# ── attribution-phrase regexes ────────────────────────────────────────────────
TESTIMONY = [
    (re.compile(r"the Committee heard", re.I), "the Committee heard"),
    (re.compile(r"told the Committee", re.I), "told the Committee"),
    (re.compile(r"\bwitnesses\b", re.I), "witnesses"),
    (re.compile(r"submitted that", re.I), "submitted that"),
    (re.compile(r"evidence to the [Ii]nquiry", re.I), "evidence to the inquiry"),
]
OPINION = [
    (re.compile(r"the Committee believes", re.I), "the Committee believes"),
    (re.compile(r"the Committee considers", re.I), "the Committee considers"),
    (re.compile(r"notes with concern", re.I), "notes with concern"),
    (re.compile(r"is disappointed|it is disappointing", re.I), "is disappointed/disappointing"),
    (re.compile(r"[Tt]he Committee calls (for|on)", re.I), "the Committee calls for/on"),
    (re.compile(r"\brecommends\b", re.I), "recommends"),
    (re.compile(r"\bshould\b", re.I), "should"),
]
WORLD_FN = re.compile(
    r"Annual Report|\bdataset\b|\bdata\b|Australian Bureau of Statistics|\bABS\b|"
    r"\bAct\s+\d{4}|\bCensus\b|Audience Atlas|OzTAM|\bVOZ\b|Productivity Commission|"
    r"Enterprise Agreement|Budget Office|cultural funding by government|"
    r"Government cultural funding|Artists as Workers|\bsurvey\b|\bReport\b|"
    r"<https?://|accessed \d{1,2}", re.I)
NAMED_SRC = re.compile(r"Transcript of evidence|Submission\s+\d+", re.I)


def is_world(c):
    """A footnote that names a dataset / annual report / statute / web source and
    no witness (no Transcript-of-evidence, no Submission N)."""
    return bool(WORLD_FN.search(c)) and not NAMED_SRC.search(c)


def norm(s):
    s = unicodedata.normalize("NFKC", s)
    s = s.replace("\f", "").replace("‑", "-")
    return s


with open(RAW, encoding="utf-8") as f:
    LINES = [norm(l.rstrip("\n")) for l in f]
N = len(LINES)

RUNNING = (
    "Legislative Council Economy and Infrastructure Committee",
    "Inquiry into the cultural and creative industries in Victoria",
)
CH = re.compile(r"^Chapter\s+\d")
NUMONLY = re.compile(r"^\d{1,3}$")
SECNUM = re.compile(r"^\d\.\d+(\.\d+)?$")
FIND = re.compile(r"^FINDING\s+(\d+):?\s*(.*)$")
REC = re.compile(r"^RECOMMENDATION\b")
CAP = re.compile(r"^(Figure|Table|Source:|Notes:|Case Study|Box\b)")


def looks_citation(s):
    return ("," in s) and re.search(
        r"Transcript of evidence|Submission\s+\d+|Annual Report|p\.\s*\d+|"
        r"Report|dataset|Atlas|Agreement|Commission|Bureau|Act\s+\d{4}|"
        r"Census|OzTAM|Committee|Government|<https?", s)


# ── footnote list ordered by line: [(line, num, citation)] ────────────────────
# A footnote block = a bare-number line, then (after blanks) a citation that may
# wrap across several lines until the next blank line. FN_LINES holds every line
# index inside a block so prose extraction can drop them (else the citation text,
# e.g. a wrapped ABS URL, leaks into the body).
FN_LINES = set()


def build_footnotes():
    out = []
    i = 0
    while i < N:
        if NUMONLY.match(LINES[i]):
            num = int(LINES[i])
            j = i + 1
            while j < N and LINES[j].strip() == "":
                j += 1
            if j < N and looks_citation(LINES[j]) and 1 <= num <= 250:
                # consume the citation (to next blank), collecting its lines
                k = j
                cit = []
                while k < N and LINES[k].strip() != "":
                    cit.append(LINES[k].strip())
                    FN_LINES.add(k)
                    k += 1
                FN_LINES.add(i)
                out.append((i, num, " ".join(cit)))
                i = k
                continue
        i += 1
    return out


FN = build_footnotes()

# A stray citation/epigraph-source line (no footnote number), e.g. a case-study
# "Kate Larsen, Submission 13, p. 2." — skip it from prose too.
STRAY_CIT = re.compile(
    r"^[A-Z][\w’'-]+ .*(Transcript of evidence|Submission\s+\d+),?\s*(pp?\.\s*[\d–-]+)?\.?\s*$")
CHART_LABEL = re.compile(r"^[\d.,%$ ]+%?$|%$")
URLISH = re.compile(r"<https?://|www\.|\.gov\.au|accessed \d{1,2}", re.I)
# A real prose paragraph carries a quote, an attribution verb, or a run of prose.
ATTR_VERB = re.compile(
    r"the Committee|told|stated|said|submission|emphasi[sz]ed|noted|explained|"
    r"highlighted|argued|according to|witness|submitted", re.I)
LC_RUN = re.compile(r"(?:\b[a-z][a-z]+\b[ ,]+){5,}")
HANSARD = re.compile(r"[A-Z]{3,}:|\b[A-Z][a-z]+ [A-Z]{3,}\b")  # Hansard speaker tag (David DAVIS:)


def is_prose(p):
    """A real body sentence, not table/figure/Hansard debris. Requires an
    attribution verb or a run of lowercase prose; rejects paragraphs that are
    mostly capitalised tokens or carry an ALL-CAPS speaker tag."""
    if len(content_words(p)) < 6:
        return False
    if not (ATTR_VERB.search(p) or LC_RUN.search(p)):
        return False
    if HANSARD.search(p):
        return False
    toks = p.split()
    caps = sum(1 for t in toks if t[:1].isupper())
    return caps <= 0.5 * len(toks)


def fn_lookup(num, near_line):
    """The footnote block with this number nearest at/after near_line (same page
    bottom); fall back to nearest before."""
    after = [(l, c) for (l, n, c) in FN if n == num and l >= near_line - 3]
    if after:
        return min(after, key=lambda t: t[0])[1]
    before = [(l, c) for (l, n, c) in FN if n == num and l < near_line]
    return max(before, key=lambda t: t[0])[1] if before else None


# ── body FINDING boxes: {num: line} (handles split "FINDING\nNN:") ────────────
def body_find_lines():
    d = {}
    for i in range(1372, N):
        m = FIND.match(LINES[i])
        if m and m.group(2).strip():
            d.setdefault(int(m.group(1)), i)
        elif LINES[i].strip() == "FINDING" and i + 1 < N:
            m2 = re.match(r"^(\d+):", LINES[i + 1])
            if m2:
                d.setdefault(int(m2.group(1)), i)
    return d


BOX = body_find_lines()

# ── section heading line for each § (nearest heading before the cluster) ──────
def section_start(sec, before_line):
    """Line of the body heading whose number == sec, nearest before before_line.
    Threshold 1000 skips the TOC (<420) and the front-matter findings list, whose
    ranges carry no standalone section-number lines."""
    cands = [i for i in range(1000, before_line)
             if LINES[i].strip() == sec]
    return cands[-1] if cands else None


STOP = set("the a an of to in and or for with is are was were be been that this these those "
           "on at by from as it its their they them our we us have has had can could may might "
           "not do does no more most all any into over per such other which who whom".split())


def content_words(t):
    return [w for w in re.findall(r"[a-z0-9$%]+", t.lower()) if w not in STOP and len(w) > 2]


def prose_paragraphs(lo, hi):
    """Clean prose paragraphs within [lo,hi): list of (text, marker_nums)."""
    paras = []
    buf = []

    def flush():
        if not buf:
            return
        raw = " ".join(buf)
        raw = re.sub(r"\s+", " ", raw).strip()
        # markers: digits glued to a letter/quote/paren, or to a letter+period
        # a footnote marker is a 1-3 digit run glued (no space) to a letter, a
        # closing quote/paren, sentence punctuation, or a letter+colon (a colon
        # introducing a block quote: "examples include:109"); never a decimal (%,
        # a preceding digit) — so $40.5, 4.8%, 7:30 and 2021–22 are left alone.
        marks = []
        for pat in (r"(?:[A-Za-z’”\)\?\!])(\d{1,3})(?![\d%])",
                    r"(?:[A-Za-z’”\)\?\!]\.)(\d{1,3})(?![\d%])",
                    r"(?:[A-Za-z’”]:)(\d{1,3})(?![\d%])"):
            for m in re.finditer(pat, raw):
                marks.append(int(m.group(1)))
        clean = raw
        for pat in (r"([A-Za-z’”\)\?\!])(\d{1,3})(?![\d%])",
                    r"([A-Za-z’”\)\?\!]\.)(\d{1,3})(?![\d%])",
                    r"([A-Za-z’”]:)(\d{1,3})(?![\d%])"):
            clean = re.sub(pat, r"\1", clean)
        if len(content_words(clean)) >= 3:
            paras.append((clean.strip(), marks))
        buf.clear()

    for i in range(lo, hi):
        l = LINES[i]
        s = l.strip()
        if s == "":
            flush()
            continue
        if (i in FN_LINES or s in RUNNING or CH.match(s) or NUMONLY.match(s)
                or SECNUM.match(s) or FIND.match(l) or REC.match(l) or CAP.match(s)
                or s == "FINDING" or s.startswith("•")
                or STRAY_CIT.match(s) or CHART_LABEL.match(s) or URLISH.search(s)):
            flush()
            continue
        buf.append(s)
    flush()
    return paras


def sentences(t):
    return [x.strip() for x in re.split(r"(?<=[.:”’])\s+(?=[A-Z‘“])", t) if x.strip()]


def classify(prose, fns):
    hits = {}
    for rx, name in TESTIMONY:
        if rx.search(prose):
            hits.setdefault("testimony", name)
    for rx, name in OPINION:
        if rx.search(prose):
            hits.setdefault("opinion", name)
    for c in fns:
        if is_world(c):
            hits.setdefault("world", "footnote: " + c[:70])
            break
    if not hits:
        return "unclear", "", hits
    if len(hits) > 1:
        return "mixed", "; ".join(f"{k}:{v}" for k, v in hits.items()), hits
    (k, v), = hits.items()
    return k, v, hits


# ── per-finding extraction ────────────────────────────────────────────────────
# group cluster ranges: for each § the shared prose is [section_start, first box].
by_sec = {}
for fid, pg, sec, txt in FINDINGS:
    by_sec.setdefault(sec, []).append(fid)

records = []
for fid, pg, sec, txt in FINDINGS:
    num = int(fid[1:])
    box = BOX.get(num)
    first_num = int(by_sec[sec][0][1:])
    first_box = BOX.get(first_num, box)
    sec_start = section_start(sec, first_box) if first_box else None
    lo = sec_start + 1 if sec_start else (first_box - 80 if first_box else 0)
    hi = first_box if first_box else lo + 80
    paras = prose_paragraphs(lo, hi)          # the subsection's evidence paragraphs
    cand = [pm for pm in paras if is_prose(pm[0])] or paras
    fwords = set(content_words(txt))
    # subject sentence = highest-overlap sentence across the whole subsection;
    # the drawing paragraph is the prose paragraph that contains it.
    best_para, best_marks, subj, subj_sc = None, [], "", -1
    for p, marks in cand:
        for s in sentences(p):
            sc = len(fwords & set(content_words(s)))
            if sc > subj_sc:
                best_para, best_marks, subj, subj_sc = p, marks, s, sc
    if best_para is None:
        records.append((fid, pg, sec, txt, "", "", [], "unclear", "", "NO-PROSE"))
        continue
    para, marks = best_para, best_marks
    # footnotes cited by the drawing paragraph's markers (route scope)
    fns, seen = [], set()
    for mnum in marks:
        if mnum in seen:
            continue
        seen.add(mnum)
        c = fn_lookup(mnum, lo)
        if c:
            fns.append((mnum, c))
    route, phrase, hits = classify(para, [c for _, c in fns])
    records.append((fid, pg, sec, txt, para, subj, fns, route, phrase, ""))

# ── emit routes-prep.md ───────────────────────────────────────────────────────
ORDER = {"unclear": 0, "mixed": 1, "world": 2, "testimony": 3, "opinion": 4}
out = []
out.append("# routes-prep.md — regex route pre-pass (no model)\n")
out.append(
    "Prepared for a human to decide each Finding's route. Every field below is drawn from the "
    "report's own body text and footnotes (pdftotext of the source PDF); the **suggested route** "
    "is computed by regex only — see the class definitions — never by a model. Fill the **human "
    "route** and **why** columns yourself. Compare against the route currently declared in "
    "`claims-machine.txt` / `routes.md`.\n")
out.append("**Regex classes.** "
           "*testimony* = the Committee heard · told the Committee · witnesses · submitted that · "
           "evidence to the inquiry. "
           "*opinion* = the Committee believes/considers · notes with concern · is disappointed · "
           "calls for · recommends · should. "
           "*world* = a footnote citing a dataset / annual report / statute with no witness named. "
           "Route scope = the Finding's own drawing paragraph + the footnotes it cites. "
           "Suggested = the single class found; **mixed** if more than one, **unclear** if none.\n")
out.append(
    "**No model guess is filled in.** The task allows a separate model-proposed column for the "
    "`unclear` rows; it is deliberately left out here so the regex signal and the human's own call "
    "stay uncontaminated (the disagreement baseline the rigour-map track depends on). Ask for model "
    "guesses as a clearly-separate column if you want them.\n")
out.append(
    "**Provenance.** Source = *Inquiry into the cultural and creative industries in Victoria*, LCEIC "
    "Final report, June 2025 (the local PDF recorded in `PROVENANCE.md` / `report-summary.md`), text "
    "layer via `pdftotext report.pdf report-raw.txt`. Finding text, § and printed page are the "
    "report's own, transcribed in `report-summary.md`. Every paragraph and footnote below is machine-"
    "extracted from that text layer — nothing is authored here.\n")
out.append(
    "**Extraction limits to keep in mind.** (1) Findings that share one subsection (e.g. F47–F52 "
    "under §4.5.1) can surface the *same* lead paragraph/sentence — the § is the real drawing unit, "
    "so read the finding's own subject against the subsection. (2) Chapter-4 findings often read "
    "`unclear` because the report attributes with *the ABC highlighted* / *SBS emphasised* / *the "
    "Committee received evidence* — none of which are in the five fixed testimony phrases; that is "
    "the regex declining to guess, not an absence of a source. (3) A few footnote markers embedded "
    "in tables/figures do not resolve and show as *(none)*; the § still cites them.\n")

# summary table first, unclear/mixed sorted to the top
out.append("## Summary table (unclear & mixed first)\n")
out.append("| id | suggested | deciding phrase | current (claims-machine) | human route |")
out.append("|----|-----------|-----------------|--------------------------|-------------|")
CURRENT = {  # route currently declared in claims-machine.txt / routes.md
    "F1":"evidence","F2":"evidence","F3":"source","F4":"evidence","F5":"source","F6":"evidence",
    "F7":"source","F8":"source","F9":"source","F10":"evidence","F11":"source","F12":"source",
    "F13":"source","F14":"evaluative","F15":"source","F16":"source","F17":"source","F18":"source",
    "F19":"source","F20":"source","F21":"source","F22":"source","F23":"source","F24":"source",
    "F25":"evaluative","F26":"source","F27":"evaluative","F28":"evaluative","F29":"evidence",
    "F30":"data-gap","F31":"evidence","F32":"data-gap","F33":"evidence","F34":"source",
    "F35":"evaluative","F36":"source","F37":"evaluative","F38":"evidence","F39":"source/evidence",
    "F40":"evidence","F41":"source","F42":"source","F43":"evaluative","F44":"source","F45":"source",
    "F46":"evidence/evaluative","F47":"source","F48":"source","F49":"evidence","F50":"source",
    "F51":"evaluative","F52":"source","F53":"evidence/source","F54":"source/evidence","F55":"source",
}
for r in sorted(records, key=lambda x: (ORDER.get(x[7], 9), int(x[0][1:]))):
    fid, pg, sec, txt, para, subj, fns, route, phrase, flag = r
    ph = (phrase or flag).replace("|", "\\|")
    out.append(f"| {fid} | **{route}** | {ph[:70]} | {CURRENT.get(fid,'?')} |  |")

out.append("\n---\n\n## Per-finding detail\n")
for r in sorted(records, key=lambda x: int(x[0][1:])):
    fid, pg, sec, txt, para, subj, fns, route, phrase, flag = r
    out.append(f"### {fid} — §{sec}, printed p{pg}\n")
    out.append(f"**Finding (verbatim):** {txt}\n")
    if flag == "NO-PROSE":
        out.append("_Drawing paragraph not auto-located (figure/table-adjacent); resolve by hand._\n")
    else:
        marked = para.replace(subj, f"**〔{subj}〕**", 1) if subj and subj in para else para
        out.append(f"**Report paragraph drawn from** (subject sentence in **〔 〕**):\n")
        out.append(f"> {marked}\n")
    out.append("**Footnotes cited in that paragraph:**")
    if fns:
        for mnum, c in fns:
            tag = "world" if is_world(c) else (
                "named-source" if NAMED_SRC.search(c) else "other")
            out.append(f"- [{mnum}] ({tag}) {c}")
    else:
        out.append("- _(none resolved from this paragraph's markers)_")
    out.append(f"\n**Suggested route (regex): `{route}`** — deciding: {phrase or '(no class matched)'}")
    out.append(f"**current claims-machine route:** {CURRENT.get(fid,'?')}")
    out.append("**human route:** ______   **why:** ______\n")

with open(OUT, "w", encoding="utf-8") as f:
    f.write("\n".join(out) + "\n")

# Optional 3rd arg: dump the unclear+mixed rows as JSONL, the input to the
# model-guess pass (build-routes-model.py). Bounded excerpts only.
if len(sys.argv) > 3:
    import json
    with open(sys.argv[3], "w", encoding="utf-8") as f:
        for r in sorted(records, key=lambda x: int(x[0][1:])):
            fid, pg, sec, txt, para, subj, fns, route, phrase, flag = r
            if route in ("unclear", "mixed"):
                f.write(json.dumps({
                    "id": fid, "section": sec, "page": pg, "finding": txt,
                    "paragraph": para, "regex_route": route,
                    "regex_phrase": phrase,
                }, ensure_ascii=False) + "\n")

# diagnostics
nop = [x[0] for x in records if x[9] == "NO-PROSE"]
from collections import Counter
cnt = Counter(x[7] for x in records)
print(f"findings: {len(records)}  routes: {dict(cnt)}", file=sys.stderr)
print(f"footnotes parsed: {len(FN)}  no-prose: {nop}", file=sys.stderr)
print(f"boxes: {sorted(BOX)}", file=sys.stderr)
