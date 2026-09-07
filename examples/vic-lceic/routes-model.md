# routes-model.md — model-guess routes (Sonnet), regex-undecided rows only

Model: `claude-sonnet-4-6`. One call per row over the 23 `unclear` + 6 `mixed` Findings from `routes-prep.md`, each given only the Finding and its drawing paragraph. This is the **model's** read; the regex pass in `routes-prep.md` is unchanged. **Do not merge either into `routes.md` / `claims-machine.txt` until the human route column is filled** — the point is a three-way disagreement record, not an averaged answer.

Model class → claims-machine route: testimony→`source`, opinion→`evaluative`, world→`evidence`, data-gap→`data-gap`.

## Three-way table

| id | regex | model | model→claims | phrase model relied on | human |
|----|-------|-------|--------------|------------------------|-------|
| F1 | mixed | **world** | evidence | the Victorian creative industries contributed $40.5 billion  |  |
| F6 | mixed | **testimony** | source | The Committee heard that 'there is a growing body of evidenc |  |
| F7 | unclear | **world** | evidence | Australia's arts, cultural and creative sector was among the |  |
| F16 | mixed | **testimony** | source | The Committee heard that many who want to engage with the Vi |  |
| F18 | mixed | **world** | evidence | 41%, or an estimated 2.10 million people, report engaging wi |  |
| F22 | unclear | **testimony** | source | operating on a four‐year grant cycle |  |
| F29 | unclear | **world** | evidence | In 2023–24, Victoria received $28.7 million, or 27% of fundi |  |
| F31 | mixed | **world** | evidence | Out of the $237.4 Creative Australia invested in 2023–24, on |  |
| F32 | mixed | **data-gap** | data-gap | remoteness level or geography breakdowns are not released |  |
| F33 | unclear | **world** | evidence | a breakdown of ABC full‐time equivalent (FTE) staff nationwi |  |
| F34 | unclear | **testimony** | source | the Committee received evidence from the ABC providing great |  |
| F35 | unclear | **opinion** | evaluative | It is disappointing that both of Australia's national broadc |  |
| F36 | unclear | **testimony** | source | the ABC's impact does go beyond that headcount |  |
| F37 | unclear | **opinion** | evaluative | calls on the Victorian Government to continue to advocate |  |
| F39 | unclear | **world** | evidence | 14% real-term reduction in the ABC's budget over the past de |  |
| F40 | unclear | **world** | evidence | No state or territory reflects an SBS FTE employee number th |  |
| F41 | unclear | **testimony** | source | we look at the available talent |  |
| F42 | unclear | **testimony** | source | SBS's decision to relocate to Western Sydney was determined  |  |
| F44 | unclear | **testimony** | source | SBS's [spending] in Victoria is particularly significant |  |
| F46 | unclear | **world** | evidence | commissioned 75 projects with the independent sector where t |  |
| F47 | unclear | **world** | evidence | commissioned 75 projects with the independent sector where t |  |
| F48 | unclear | **testimony** | source | have made a significant contribution to the development of t |  |
| F49 | unclear | **world** | evidence | 16 regional radio programs in Victoria which produces over 1 |  |
| F50 | unclear | **testimony** | source | lack of independent production sector proposals for shows to |  |
| F51 | unclear | **world** | evidence | commissioned 75 projects with the independent sector where t |  |
| F52 | unclear | **opinion** | evaluative | delivers significant benefits, including social benefits … a |  |
| F53 | unclear | **world** | evidence | SBS's receives about one-third of the ABC's funding |  |
| F54 | unclear | **world** | evidence | producing language programs that reach 3.42 million Victoria |  |
| F55 | unclear | **testimony** | source | Clare O'Neil emphasised that SBS 'simply do not have compara |  |

---

## Per-row model answers

### F1 — §2.1.1, p11  (regex: mixed)

**Finding:** The Victorian cultural and creative industries are economically significant, contributing $40.5 billion to the Victorian economy in 2022–23.

**Model route:** `world`  → claims-machine `evidence`
**Phrase relied on:** “the Victorian creative industries contributed $40.5 billion to the Victorian economy”
**Reason:** The Finding makes a specific, dated economic measurement about the sector's contribution that requires an external dataset or official economic account as its truth-maker, not merely the Committee's judgment or a witness's opinion.

### F6 — §2.1.2, p14  (regex: mixed)

**Finding:** There is a growing body of evidence that recognises the crucial role arts and creativity can play in promoting positive mental health and wellbeing outcomes.

**Model route:** `testimony`  → claims-machine `source`
**Phrase relied on:** “The Committee heard that 'there is a growing body of evidence that recognises the crucial role arts and creativity can play in promoting positive mental health and wellbeing outcomes'”
**Reason:** The Finding directly compresses what a witness/submitter told the Committee, as shown by the explicit attribution "The Committee heard that" in the paragraph, making it a compression of witness testimony rather than the Committee's own factual or evaluative claim.

### F7 — §2.2.1, p18  (regex: unclear)

**Finding:** The COVID-19 pandemic severely damaged the Victorian cultural and creative industries. It significantly limited opportunities to present and engage with works publicly, and consequently restricted practitioners' income.

**Model route:** `world`  → claims-machine `evidence`
**Phrase relied on:** “Australia's arts, cultural and creative sector was among the first and hardest‐hit casualties of the COVID‐19 pandemic”
**Reason:** The Finding makes a factual claim about the real-world impact of COVID-19 on Victorian cultural and creative industries — verifiable against economic data, industry reports, or official records — rather than compressing witness testimony or expressing the Committee's own value judgment.

### F16 — §2.2.5, p26  (regex: mixed)

**Finding:** Many people who want to engage with the Victorian cultural and creative industries cannot afford to do so due to the cost of living crisis.

**Model route:** `testimony`  → claims-machine `source`
**Phrase relied on:** “The Committee heard that many who want to engage with the Victorian cultural and creative industries, cannot afford to do so due to the current cost of living crisis.”
**Reason:** The Finding directly compresses what witnesses/submitters told the Committee ("The Committee heard that…"), not the external Audience Atlas dataset figures that follow in the paragraph.

### F18 — §2.2.5, p26  (regex: mixed)

**Finding:** The COVID-19 pandemic and the cost of living crisis has accelerated the decline of volunteerism in the Victorian cultural and creative industries.

**Model route:** `world`  → claims-machine `evidence`
**Phrase relied on:** “41%, or an estimated 2.10 million people, report engaging with the Victorian cultural and creative industries less, compared to five years ago”
**Reason:** The Finding's core claim about accelerating decline rests on a quantified, externally published dataset (Audience Atlas Victoria 2024), making it a factual claim about the world requiring that external source as its truth-maker.

### F22 — §3.2.3, p40  (regex: unclear)

**Finding:** Unpredictable grant programs cause instability in the industries and prevents forward planning and investments in longer-term projects and strategies.

**Model route:** `testimony`  → claims-machine `source`
**Phrase relied on:** “operating on a four‐year grant cycle”
**Reason:** The Finding compresses what submitters/witnesses told the Committee, as evidenced by the quoted phrases drawn directly from their evidence (e.g. "operating on a four-year grant cycle", "direct valuable resources away from program delivery"), making it a compression of witness testimony rather than the Committee's own judgment or an externally verifiable fact.

### F29 — §3.3.5, p57  (regex: unclear)

**Finding:** Victoria receives its fair share of funding from Creative Australia.

**Model route:** `world`  → claims-machine `evidence`
**Phrase relied on:** “In 2023–24, Victoria received $28.7 million, or 27% of funding”
**Reason:** The finding rests on a concrete, externally verifiable funding figure drawn from Creative Australia's published investment breakdown, making it a factual claim about the world checkable against that official dataset.

### F31 — §3.3.6, p59  (regex: mixed)

**Finding:** Regional Victoria does not receive its fair share of funding from Creative Australia.

**Model route:** `world`  → claims-machine `evidence`
**Phrase relied on:** “Out of the $237.4 Creative Australia invested in 2023–24, only $28.4 (approximately 12%) was invested in regional Australia.”
**Reason:** The Finding rests on a concrete funding figure drawn from Creative Australia's own regional breakdown data — an external dataset — making it a factual claim about the world with an official record as its truth-maker.

### F32 — §3.3.6, p59  (regex: mixed)

**Finding:** It is difficult to determine whether regional Victoria receives its fair share of other funding captured in the cultural funding by government dataset, as remoteness level or geography breakdowns are not released.

**Model route:** `data-gap`  → claims-machine `data-gap`
**Phrase relied on:** “remoteness level or geography breakdowns are not released”
**Reason:** The Finding's core claim is that the cultural funding by government dataset does not publish a geographic/remoteness breakdown, which is a statement about the limits of an existing dataset rather than a factual claim about the world or a witness's view.

### F33 — §4.3.1, p69  (regex: unclear)

**Finding:** No state or territory has an ABC headcount that accurately reflects its population size; an overrepresentation or underrepresentation of ABC employee headcount exists to varying degrees in all states and territories.

**Model route:** `world`  → claims-machine `evidence`
**Phrase relied on:** “a breakdown of ABC full‐time equivalent (FTE) staff nationwide along with the staffing costs associated with each state and territory”
**Reason:** The Finding makes a factual claim about the distribution of ABC headcount across all states and territories relative to population, grounded in an external dataset (the ABC's own FTE staffing figures) rather than a witness's opinion or the Committee's value judgment.

### F34 — §4.3.1, p69  (regex: unclear)

**Finding:** The ABC's decision to relocate its Ultimo office to Parramatta was influenced by multiple external factors, including legal obligations under the ABC Enterprise Agreement 2022–2025, budget constraints that made New South Wales the most cost-effective option, the Parramatta office's proximity to key state departments and agencies, and a strategic goal to better reflect New South Wales' population demographics.

**Model route:** `testimony`  → claims-machine `source`
**Phrase relied on:** “the Committee received evidence from the ABC providing greater detail on the costing of the Parramatta relocation”
**Reason:** The Finding's characterisation of the multiple factors influencing the relocation decision — including the Enterprise Agreement, budget constraints, proximity to state agencies, and demographic goals — rests entirely on what the ABC and its representatives told the Committee, not on independently verifiable external datasets or the Committee's own value judgments.

### F35 — §4.3.1, p69  (regex: unclear)

**Finding:** It is disappointing that both of Australia's national broadcasters the ABC and SBS expanded their presence to Western Sydney instead of Victoria.

**Model route:** `opinion`  → claims-machine `evaluative`
**Phrase relied on:** “It is disappointing that both of Australia's national broadcasters the ABC and SBS expanded their presence to Western Sydney instead of Victoria.”
**Reason:** The Finding expresses the Committee's own value judgment ("it is disappointing") about a locational decision, making it an evaluative claim rather than a factual or testimonial one.

### F36 — §4.3.1, p69  (regex: unclear)

**Finding:** The ABC's impact and ability to culturally represent any particular state goes beyond the headcount or location of ABC offices.

**Model route:** `testimony`  → claims-machine `source`
**Phrase relied on:** “the ABC's impact does go beyond that headcount”
**Reason:** The Finding directly compresses what witness Mark Tapley told the Committee, paraphrasing his verbatim statement about the ABC's impact extending beyond headcount and office locations.

### F37 — §4.3.2, p74  (regex: unclear)

**Finding:** The Committee calls on the Victorian Government to continue to advocate to the Federal Government and the ABC for the return of a Victorian 7:30 Report.

**Model route:** `opinion`  → claims-machine `evaluative`
**Phrase relied on:** “calls on the Victorian Government to continue to advocate”
**Reason:** The Finding is the Committee's own directive — a call to action addressed to government — not a factual claim about the world or a compression of witness testimony.

### F39 — §4.3.2, p74  (regex: unclear)

**Finding:** The ABC's content production decisions are shaped by state screen agency policies and funding availability. External pressures, including rising production costs and a 14% real-term reduction in the ABC's budget over the past decade, have constrained the broadcaster's capacity to create content and ability to expand its footprint in the nation's states and territories.

**Model route:** `world`  → claims-machine `evidence`
**Phrase relied on:** “14% real-term reduction in the ABC's budget over the past decade”
**Reason:** This is a factual claim about the ABC's budget trajectory that requires an external truth-maker (budget records, official financial data) to verify, not merely a witness opinion or a gap in data.

### F40 — §4.4.1, p76  (regex: unclear)

**Finding:** No state or territory reflects an SBS FTE employee number that is proportionate to its share of the Australia's population.

**Model route:** `world`  → claims-machine `evidence`
**Phrase relied on:** “No state or territory reflects an SBS FTE employee number that is proportionate to its share of the Australia's population.”
**Reason:** The Finding makes a factual claim about the world — comparing SBS FTE employee counts by state/territory against population share — a claim verifiable against official workforce and census/ABS datasets, not merely what a witness said or a committee judgment.

### F41 — §4.4.1, p76  (regex: unclear)

**Finding:** The geographic distribution of SBS staff does not reflect the availability of its content. While a significant portion of SBS's workforce is based in New South Wales, its services are national and platform-agnostic, ensuring all Australians, regardless of location, can access and benefit from SBS content.

**Model route:** `testimony`  → claims-machine `source`
**Phrase relied on:** “we look at the available talent”
**Reason:** The Finding compresses what an SBS witness told the Committee about how the organisation staffs its language services relative to geographic talent pools — a claim only checkable against the transcripts/submissions.

### F42 — §4.4.1, p78  (regex: unclear)

**Finding:** The SBS's decision to relocate to Western Sydney was determined by funding requirements set by the Federal Government.

**Model route:** `testimony`  → claims-machine `source`
**Phrase relied on:** “SBS's decision to relocate to Western Sydney was determined by funding requirements set by the Federal Government”
**Reason:** The Finding asserts a causal explanation for SBS's relocation decision — a claim that could only have come from what SBS witnesses/submitters told the Committee, as the paragraph itself contains no external dataset or official record supporting this causal link, only staffing distribution figures.

### F44 — §4.4.2, p80  (regex: unclear)

**Finding:** SBS provides significant financial and creative investment to Victoria.

**Model route:** `testimony`  → claims-machine `source`
**Phrase relied on:** “SBS's [spending] in Victoria is particularly significant”
**Reason:** The Finding compresses what Clare O'Neil of SBS told the Committee, making it a witness claim checkable against the transcript rather than an external dataset or the Committee's own judgment.

### F46 — §4.5.1, p82  (regex: unclear)

**Finding:** Between 2020 and 2025, the ABC commissioned 75 external and 52 internal projects where the majority of production took place in Victoria, demonstrating its vital role in supporting the state's screen industry.

**Model route:** `world`  → claims-machine `evidence`
**Phrase relied on:** “commissioned 75 projects with the independent sector where the majority of production [took] place in Victoria”
**Reason:** The Finding asserts concrete production counts (75 external, 52 internal) drawn from ABC's own records — factual claims about the world requiring an external dataset or official record as truth-maker, not solely a stakeholder value judgment.

### F47 — §4.5.1, p83  (regex: unclear)

**Finding:** The ABC partners with major Victorian cultural institutions to create educational content and showcase Victorian talent.

**Model route:** `world`  → claims-machine `evidence`
**Phrase relied on:** “commissioned 75 projects with the independent sector where the majority of production [took] place in Victoria”
**Reason:** The Finding rests on a concrete production count (75 commissions, 52 internal projects) drawn from ABC records over a defined period, making it a factual claim about the world verifiable against an external dataset or official record.

### F48 — §4.5.1, p83  (regex: unclear)

**Finding:** Through initiatives like VicScreen internship and Indigenous placements, the ABC actively develops Victoria creative and cultural workforce.

**Model route:** `testimony`  → claims-machine `source`
**Phrase relied on:** “have made a significant contribution to the development of the sector, supporting employment and opportunity for talented Victorian creative workers”
**Reason:** The Finding compresses what ABC witness Mark Tapley told the Committee about the impact of these initiatives, relying entirely on his testimony rather than an external dataset or the Committee's own value judgment.

### F49 — §4.5.1, p84  (regex: unclear)

**Finding:** The ABC delivers 16 regional radio programs in Victoria which produces over 112 hours of weekly local content across 9 stations, ensuring broad community representation.

**Model route:** `world`  → claims-machine `evidence`
**Phrase relied on:** “16 regional radio programs in Victoria which produces over 112 hours of weekly local content across 9 stations”
**Reason:** The Finding states concrete, verifiable operational facts (number of programs, hours of content, number of stations) that require an external truth-maker such as ABC's own records or annual reports to confirm.

### F50 — §4.5.1, p85  (regex: unclear)

**Finding:** One of the main challenges for the ABC in producing content in regional Victoria is the lack of independent production sector proposals for shows to be made in regional Victoria.

**Model route:** `testimony`  → claims-machine `source`
**Phrase relied on:** “lack of independent production sector proposals for shows to be made in regional Victoria”
**Reason:** The Finding restates a challenge as characterised by the ABC to the Committee — it reflects what the ABC told the inquiry, not an independently verifiable external dataset or the Committee's own value judgment.

### F51 — §4.5.1, p85  (regex: unclear)

**Finding:** The cost of producing content in regional Victoria requires significant financial investment.

**Model route:** `world`  → claims-machine `evidence`
**Phrase relied on:** “commissioned 75 projects with the independent sector where the majority of production [took] place in Victoria”
**Reason:** The Finding makes a factual claim about production costs in regional Victoria, grounded in concrete output/investment figures (75 external commissions, 52 internal projects) drawn from ABC records — an external dataset — not merely a witness's opinion.

### F52 — §4.5.1, p85  (regex: unclear)

**Finding:** Producing content in regional areas delivers significant benefits, including social benefits, such as better representation and celebration of local communities and economic benefits, like job creation and increased local spending.

**Model route:** `opinion`  → claims-machine `evaluative`
**Phrase relied on:** “delivers significant benefits, including social benefits … and economic benefits”
**Reason:** The Finding is the Committee's own value judgment that regional content production is beneficial, not a measurable fact drawn from an external dataset nor a direct compression of witness testimony.

### F53 — §4.5.2, p87  (regex: unclear)

**Finding:** SBS's receives about one-third of the ABC's funding, which restricts its ability to produce and commission content for and in Victoria.

**Model route:** `world`  → claims-machine `evidence`
**Phrase relied on:** “SBS's receives about one-third of the ABC's funding”
**Reason:** The headline claim is a concrete funding ratio between two public broadcasters — a factual claim about the world verifiable against budget/appropriation records, not merely a witness opinion or value judgment.

### F54 — §4.5.2, p87  (regex: unclear)

**Finding:** SBS sustains substantial Melbourne-based operations, producing language programs that reach 3.42 million Victorian viewers monthly, while supporting diverse communities through partnerships with VicScreen and local creatives.

**Model route:** `world`  → claims-machine `evidence`
**Phrase relied on:** “producing language programs that reach 3.42 million Victorian viewers monthly”
**Reason:** The headline figure of 3.42 million Victorian viewers monthly is a factual claim about audience reach that requires an external data source (e.g., ratings or audience measurement data) as its truth-maker, and is not found in the witness testimony in the paragraph, which contains no such number.

### F55 — §4.5.2, p87  (regex: unclear)

**Finding:** The SBS deeply engages with Victoria's multilingual and multicultural communities by supporting community events and major cultural events in Victoria.

**Model route:** `testimony`  → claims-machine `source`
**Phrase relied on:** “Clare O'Neil emphasised that SBS 'simply do not have comparable scale' to the ABC, but 'the significant commissioning investment it makes in Victoria reflects its strong engagement and commitment to the growth and sustainability of the sector'”
**Reason:** The Finding compresses what a witness (Clare O'Neil) told the Committee about SBS's engagement with Victorian communities, resting entirely on her submitted characterisation rather than an independently verifiable external dataset or the Committee's own value judgment.

