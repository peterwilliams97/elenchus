# routes-prep.md — regex route pre-pass (no model)

Prepared for a human to decide each Finding's route. Every field below is drawn from the report's own body text and footnotes (pdftotext of the source PDF); the **suggested route** is computed by regex only — see the class definitions — never by a model. Fill the **human route** and **why** columns yourself. Compare against the route currently declared in `claims-machine.txt` / `routes.md`.

**Regex classes.** *testimony* = the Committee heard · told the Committee · witnesses · submitted that · evidence to the inquiry. *opinion* = the Committee believes/considers · notes with concern · is disappointed · calls for · recommends · should. *world* = a footnote citing a dataset / annual report / statute with no witness named. Route scope = the Finding's own drawing paragraph + the footnotes it cites. Suggested = the single class found; **mixed** if more than one, **unclear** if none.

**No model guess is filled in.** The task allows a separate model-proposed column for the `unclear` rows; it is deliberately left out here so the regex signal and the human's own call stay uncontaminated (the disagreement baseline the rigour-map track depends on). Ask for model guesses as a clearly-separate column if you want them.

**Provenance.** Source = *Inquiry into the cultural and creative industries in Victoria*, LCEIC Final report, June 2025 (the local PDF recorded in `PROVENANCE.md` / `report-summary.md`), text layer via `pdftotext report.pdf report-raw.txt`. Finding text, § and printed page are the report's own, transcribed in `report-summary.md`. Every paragraph and footnote below is machine-extracted from that text layer — nothing is authored here.

**Extraction limits to keep in mind.** (1) Findings that share one subsection (e.g. F47–F52 under §4.5.1) can surface the *same* lead paragraph/sentence — the § is the real drawing unit, so read the finding's own subject against the subsection. (2) Chapter-4 findings often read `unclear` because the report attributes with *the ABC highlighted* / *SBS emphasised* / *the Committee received evidence* — none of which are in the five fixed testimony phrases; that is the regex declining to guess, not an absence of a source. (3) A few footnote markers embedded in tables/figures do not resolve and show as *(none)*; the § still cites them.

## Summary table (unclear & mixed first)

| id | suggested | deciding phrase | current (claims-machine) | human route |
|----|-----------|-----------------|--------------------------|-------------|
| F7 | **unclear** |  | source |  |
| F22 | **unclear** |  | source |  |
| F29 | **unclear** |  | evidence |  |
| F33 | **unclear** |  | evidence |  |
| F34 | **unclear** |  | source |  |
| F35 | **unclear** |  | evaluative |  |
| F36 | **unclear** |  | source |  |
| F37 | **unclear** |  | evaluative |  |
| F39 | **unclear** |  | source/evidence |  |
| F40 | **unclear** |  | evidence |  |
| F41 | **unclear** |  | source |  |
| F42 | **unclear** |  | source |  |
| F44 | **unclear** |  | source |  |
| F46 | **unclear** |  | evidence/evaluative |  |
| F47 | **unclear** |  | source |  |
| F48 | **unclear** |  | source |  |
| F49 | **unclear** |  | evidence |  |
| F50 | **unclear** |  | source |  |
| F51 | **unclear** |  | evaluative |  |
| F52 | **unclear** |  | source |  |
| F53 | **unclear** |  | evidence/source |  |
| F54 | **unclear** |  | source/evidence |  |
| F55 | **unclear** |  | source |  |
| F1 | **mixed** | testimony:the Committee heard; world:footnote: Creative Victoria, Econ | evidence |  |
| F6 | **mixed** | testimony:the Committee heard; world:footnote: Victorian Department of | evidence |  |
| F16 | **mixed** | testimony:the Committee heard; world:footnote: Creative Victoria, Audi | source |  |
| F18 | **mixed** | testimony:the Committee heard; world:footnote: Creative Victoria, Audi | source |  |
| F31 | **mixed** | testimony:the Committee heard; world:footnote: Creative Australia, Ann | evidence |  |
| F32 | **mixed** | testimony:the Committee heard; world:footnote: Creative Australia, Ann | data-gap |  |
| F17 | **world** | footnote: Creative Victoria, Audience Atlas Victoria 2024: Exploring t | source |  |
| F2 | **testimony** | the Committee heard | evidence |  |
| F3 | **testimony** | the Committee heard | source |  |
| F4 | **testimony** | the Committee heard | evidence |  |
| F5 | **testimony** | the Committee heard | source |  |
| F8 | **testimony** | the Committee heard | source |  |
| F9 | **testimony** | told the Committee | source |  |
| F10 | **testimony** | told the Committee | evidence |  |
| F11 | **testimony** | told the Committee | source |  |
| F12 | **testimony** | the Committee heard | source |  |
| F13 | **testimony** | the Committee heard | source |  |
| F14 | **testimony** | the Committee heard | evaluative |  |
| F15 | **testimony** | the Committee heard | source |  |
| F19 | **testimony** | the Committee heard | source |  |
| F20 | **testimony** | the Committee heard | source |  |
| F21 | **testimony** | the Committee heard | source |  |
| F23 | **testimony** | the Committee heard | source |  |
| F24 | **testimony** | the Committee heard | source |  |
| F25 | **testimony** | the Committee heard | evaluative |  |
| F26 | **testimony** | the Committee heard | source |  |
| F27 | **testimony** | the Committee heard | evaluative |  |
| F30 | **testimony** | told the Committee | data-gap |  |
| F38 | **testimony** | the Committee heard | evidence |  |
| F43 | **testimony** | the Committee heard | evaluative |  |
| F45 | **testimony** | the Committee heard | source |  |
| F28 | **opinion** | should | evaluative |  |

---

## Per-finding detail

### F1 — §2.1.1, printed p11

**Finding (verbatim):** The Victorian cultural and creative industries are economically significant, contributing $40.5 billion to the Victorian economy in 2022–23.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> The Committee heard that ‘the creative industries are a major contributor to local economies and to our national economy’. **〔In 2022–23, ‘the Victorian creative industries contributed $40.5 billion to the Victorian economy’.〕** See Figure 2.1. The industries economic contribution has ‘increased at an average annual rate of 4.8% over the last five years’.

**Footnotes cited in that paragraph:**
- [1] (named-source) Parliamentary Budget Office, Submission 15, pp. 17–18.
- [2] (named-source) Claire Febey, Chief Executive, Creative Victoria, public hearing, Melbourne, 12 March 2025, Transcript of evidence, p. 1.
- [3] (world) Creative Victoria, Economic contribution data: Victoria’s creative economy <https://creative.vic.gov.au/resources/datainsights/victorias-creative-economy/economic-impact> accessed 15 May 2025.

**Suggested route (regex): `mixed`** — deciding: testimony:the Committee heard; world:footnote: Creative Victoria, Economic contribution data: Victoria’s creative eco
**current claims-machine route:** evidence
**human route:** ______   **why:** ______

### F2 — §2.1.1, printed p11

**Finding (verbatim):** The Victorian cultural and creative industries are an important employer, employing more than 320,000 Victorians or almost 9 per cent of total employment in the state.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> **〔The industries are a significant employer The creative economy employs ‘more than 320,000 Victorians ... that is almost 9 per cent of total employment in the state’.〕** The Committee heard that the industries ‘create a sort of ecology of economic activity around them’. Maggie Collins, Executive Director, Association of Artist Managers, stated: Every single artist is a business. They are either a sole trader, partnership, proprietary limited or trust, or whatever it is; they all file a tax return. They are all businesses and they all interact with other businesses – like a manager who is a sole trader or partnership or proprietary limited, or a publicist who is that, or an agent who is booking their shows, or a graphic designer who is doing their work, or anyone else – and in this ecosystem all of these businesses are trying to interact with one another and lift each other’s work and businesses up.

**Footnotes cited in that paragraph:**
- [6] (named-source) Claire Febey, Transcript of evidence, p. 1.
- [7] (named-source) Claire Febey, Transcript of evidence, p. 10.
- [8] (named-source) Parliamentary Budget Office, Submission 15, p. 18.

**Suggested route (regex): `testimony`** — deciding: the Committee heard
**current claims-machine route:** evidence
**human route:** ______   **why:** ______

### F3 — §2.1.1, printed p11

**Finding (verbatim):** The Victorian cultural and creative industries provide significant economic stimulus to the state by attracting visitation and tourism.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> **〔The industries drive tourism The Committee heard that the Victorian cultural and creative industries provide ‘huge economic stimulus to the state in terms of locals and people traveling to’ engage with the industries.〕** Joshua Lowe, Theatre Network Australia, emphasised this point by asking: I want you to imagine Melbourne without Hosier Lane, without the arts centre, the NGV, the comedy festival, the Fringe Festival, concerts at Rod Laver, RISING lighting up the dead of winter, Fed Square alive with performances. Why would people visit the city? How would Melbourne be different to any other city?

**Footnotes cited in that paragraph:**
- [10] (named-source) Tim Stitz, Board Member, Theatre Network Australia, public hearing, Melbourne, 12 March 2025, Transcript of evidence, p. 23.
- [11] (named-source) Parliamentary Budget Office, Submission 15, p. 19.

**Suggested route (regex): `testimony`** — deciding: the Committee heard
**current claims-machine route:** source
**human route:** ______   **why:** ______

### F4 — §2.1.1, printed p11

**Finding (verbatim):** Victoria is a significant cultural audience, which rated higher for cultural participation than all other jurisdictions surveyed, including New York, Sweden and the United Kingdom.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> The Committee heard that ‘Victoria is one of Australia’s most significant cultural audiences’. **〔Claire Febey of Creative Victoria, stated that ‘Victoria rated higher for cultural participation than all other jurisdictions where this research has been undertaken.〕** This includes New York City, Sweden and the UK’.

**Footnotes cited in that paragraph:**
- [14] (named-source) Australian Museums and Galleries Association (AMaGA), Submission 35, p. 3.
- [15] (named-source) Claire Febey, Transcript of evidence, p. 1.

**Suggested route (regex): `testimony`** — deciding: the Committee heard
**current claims-machine route:** evidence
**human route:** ______   **why:** ______

### F5 — §2.1.2, printed p14

**Finding (verbatim):** Engagement with the Victorian cultural and creative industries brings communities together, breaks down barriers between different groups within society and encourages greater communication and cohesion.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> Social significance Social cohesion is ‘essential for healthy people, societies and economies’. Yet, research warns that Australia ‘faces real and immediate social cohesion challenges’. **〔The Committee heard that engagement with the cultural and creative industries can significantly improve social cohesion by counteracting ‘loneliness, isolation and disrupted connections’.〕** Katrina Cornwell, Co‐Artistic Director, Rawcus Theatre Company, told the Committee:

**Footnotes cited in that paragraph:**
- [20] (other) A New Approach, Belong, Trust, Connect: Policy Opportunities for social cohesion through arts and culture March 2025, p. 11.
- [21] (other) A New Approach, Belong, Trust, Connect: Policy Opportunities for social cohesion through arts and culture, p. 11.
- [22] (other) A New Approach, Belong, Trust, Connect: Policy Opportunities for social cohesion through arts and culture, p. 11.

**Suggested route (regex): `testimony`** — deciding: the Committee heard
**current claims-machine route:** source
**human route:** ______   **why:** ______

### F6 — §2.1.2, printed p14

**Finding (verbatim):** There is a growing body of evidence that recognises the crucial role arts and creativity can play in promoting positive mental health and wellbeing outcomes.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> isolation and loneliness have detrimental physical and mental health consequences’. Concerning data from the Victorian Mental Health and Wellbeing Commission shows that rates of mental distress and mental illness are increasing – with 1 in 5 Victorians experiencing a mental health condition each year. Additionally, the Royal Commission into Victoria’s Mental Health System identified that the economic cost of poor mental health in Victoria is $14.2 billion annually. **〔The Committee heard that ‘there is a growing body of evidence that recognises the crucial role arts and creativity can play in promoting positive mental health and wellbeing outcomes’.〕** Additionally, ‘Australians are increasingly acting on that belief and attending arts and cultural events to improve their wellbeing’. Anne Robertson, Executive Officer, Public Galleries Association of Victoria, told the Committee: The research out of Western Australia – its catchphrase is ‘You don’t have to be good at the arts for the arts to be good for you.’ They recommend 2 hours of any kind of art activity a week, and it can improve your mental health by 30 per cent.

**Footnotes cited in that paragraph:**
- [27] (world) Victorian Department of Health, Mental illness and mental wellbeing, <https://www.health.vic.gov.au/your-health-report-ofthe-chief-health-officer-victoria-2018/mental-health/mental-illness-and> accessed 15 May 2025.
- [28] (named-source) Caitlin Dullard, Transcript of evidence, p. 16.
- [29] (world) Royal Commission into Victoria’s Mental Health System, Interim Report, p. 361.
- [30] (named-source) Theatre Network Australia, Submission 19, Attachment 1, p. 6.
- [31] (other) A New Approach, Belong, Trust, Connect: Policy Opportunities for social cohesion through arts and culture, p. 11.
- [32] (named-source) Dr Ashley Robertson, Transcript of evidence, p. 49.

**Suggested route (regex): `mixed`** — deciding: testimony:the Committee heard; world:footnote: Victorian Department of Health, Mental illness and mental wellbeing, <
**current claims-machine route:** evidence
**human route:** ______   **why:** ______

### F7 — §2.2.1, printed p18

**Finding (verbatim):** The COVID-19 pandemic severely damaged the Victorian cultural and creative industries. It significantly limited opportunities to present and engage with works publicly, and consequently restricted practitioners' income.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> **〔Loss of audiences and income Australia’s arts, cultural and creative sector was among the first and hardest‐hit casualties of the COVID‐19 pandemic – nowhere more so than in Victoria.〕**

**Footnotes cited in that paragraph:**
- _(none resolved from this paragraph's markers)_

**Suggested route (regex): `unclear`** — deciding: (no class matched)
**current claims-machine route:** source
**human route:** ______   **why:** ______

### F8 — §2.2.1, printed p18

**Finding (verbatim):** The COVID-19 pandemic took away critical training opportunities from those looking to enter the Victorian cultural and creative industries during this time.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> **〔The Committee heard that the COVID‐19 pandemic ‘shut down’ the cultural and creative industries.〕** Conor Cunningham, President, Bendigo Theatre Company, told the Committee that ‘it completely stopped everything we were doing’. He stated: I think we had two productions ready to be staged at that point. The youth theatre were doing Cats. We ourselves were doing a production of We Will Rock You. Both productions came to a complete standstill, unfortunately. I mean, as grown‐ups, the adults were able to deal with the impact of that, but unfortunately our younger people felt the impact of that more because of the work they had put in, the investment of time and love that they gave to that production, only for it to be taken away from them in such a sudden moment.

**Footnotes cited in that paragraph:**
- [35] (named-source) Dr Ashley Robertson, Transcript of evidence, p. 49.
- [36] (named-source) Conor Cunningham, President, Bendigo Theatre Company, public hearing, Melbourne, 12 March 2025, Transcript of evidence, p. 59.
- [37] (named-source) Conor Cunningham, Transcript of evidence, p. 59.

**Suggested route (regex): `testimony`** — deciding: the Committee heard
**current claims-machine route:** source
**human route:** ______   **why:** ______

### F9 — §2.2.1, printed p18

**Finding (verbatim):** Exclusion from COVID-19 financial support programs exacerbated the financial issues experienced by many in the industry and created ongoing skill gaps in the industry.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> Furthermore, many individuals who work in the cultural and creative industries were ineligible to apply for targeted COVID‐19 financial support programs provided by Commonwealth and State Governments. Ineligibility exacerbated the financial issues experienced by those in the industries and meant that they did not benefit from the same payments many individuals in other sectors benefited from during the COVID‐19 pandemic. The impact of being excluded from COVID‐19 financial support programs are still being felt by many in the industries who are yet to recover financially. **〔Exclusion from COVID‐19 financial support programs also has created skill gaps in the industries.〕** Dianne Toulson of Theatre Works, told the Committee: There is that ongoing effect of people being nervous about staying in an industry that is so volatile, and parts of the arts sector were not supported through COVID, so they have disappeared. The lack of technical people that can do technical work – they have left the industry because they were not supported. All of those impacts put financial pressure on us.

**Footnotes cited in that paragraph:**
- [43] (named-source) Dianne Toulson, Transcript of Evidence, p. 6.

**Suggested route (regex): `testimony`** — deciding: told the Committee
**current claims-machine route:** source
**human route:** ______   **why:** ______

### F10 — §2.2.1, printed p18

**Finding (verbatim):** Children and young people's creative participation and attendance between 2017–18 and 2021–22 has significantly declined.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> Conversely, those creative organisations who did quality for financial support programs reported that such payments allowed them to ‘keep staff on and pay artists’. Indeed, some organisations told the Committee that their ‘budgets over COVID were bigger than they are now’. Decreased audience engagement with the cultural and creative industries also had a negative impact upon the audiences themselves, particularly children and young people. **〔The Australian Bureau of Statistics ‘reported a concerning decrease’ in children and young people’s creative participation and attendance between 2017–18 and 2021–22.46〕**

**Footnotes cited in that paragraph:**
- [44] (other) ABC, hearing, response to questions on notice received 21 March 2025, p. 3.
- [45] (named-source) Caitlin Dullard, Transcript of evidence, p. 14.

**Suggested route (regex): `testimony`** — deciding: told the Committee
**current claims-machine route:** evidence
**human route:** ______   **why:** ______

### F11 — §2.2.1, printed p18

**Finding (verbatim):** Several live music venues have closed following the COVID-19 pandemic due to various pressures, such as rising costs and inconsistent financial support.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> **〔Furthermore, many individuals who work in the cultural and creative industries were ineligible to apply for targeted COVID‐19 financial support programs provided by Commonwealth and State Governments.〕** Ineligibility exacerbated the financial issues experienced by those in the industries and meant that they did not benefit from the same payments many individuals in other sectors benefited from during the COVID‐19 pandemic. The impact of being excluded from COVID‐19 financial support programs are still being felt by many in the industries who are yet to recover financially. Exclusion from COVID‐19 financial support programs also has created skill gaps in the industries. Dianne Toulson of Theatre Works, told the Committee: There is that ongoing effect of people being nervous about staying in an industry that is so volatile, and parts of the arts sector were not supported through COVID, so they have disappeared. The lack of technical people that can do technical work – they have left the industry because they were not supported. All of those impacts put financial pressure on us.

**Footnotes cited in that paragraph:**
- [43] (named-source) Dianne Toulson, Transcript of Evidence, p. 6.

**Suggested route (regex): `testimony`** — deciding: told the Committee
**current claims-machine route:** source
**human route:** ______   **why:** ______

### F12 — §2.2.2, printed p20

**Finding (verbatim):** The COVID-19 pandemic led to worsening mental health for many in the Victorian cultural and creative industries, particularly amongst children and young people.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> **〔The Committee heard that the COVID‐19 pandemic had a ‘profound’ impact upon mental health in the Victorian cultural and creative industries.〕** Nadja Kostich of St Martins Youth Arts Centre, spoke about how worsening mental health during the COVID‐19 pandemic led to increased staff attrition: A lot of staff mental health issues came through and we were then not able to afford to replace certain key members of the staff that we are able to do only now with this small bridging funding. Our executive director we lost during COVID; an inclusion coordinator we lost and we have not been able to regain. The cost of living and the cost of everything has made it very difficult to rebuild the team.

**Footnotes cited in that paragraph:**
- [55] (named-source) Dianne Toulson, Transcript of evidence, p. 6.
- [56] (named-source) Nadja Kostich, Transcript of evidence, pp. 6–7.

**Suggested route (regex): `testimony`** — deciding: the Committee heard
**current claims-machine route:** source
**human route:** ______   **why:** ______

### F13 — §2.2.3, printed p21

**Finding (verbatim):** Since the COVID-19 pandemic, the costs to produce and present work have significantly increased.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> **〔Significant cost increases The Committee heard that across the industry the ‘costs to produce and present work have now significantly increased’.〕** Vern Wall, General Committee Member, Bendigo Theatre Company, stated that ‘in the last eight to 10 years the costs roughly have increased by 80 to 100 per cent’. He stated: The only way it appears now that we can actually balance our books is as well as having our volunteers provide all their time to create the production, we then have to create fundraising opportunities to get additional funds. And without that funding, the impact on us is that the only way we can meet these increasing costs is obviously by increasing our ticket prices. That then, conversely, reduces our audience numbers due to the lack of affordability for them to attend our shows.

**Footnotes cited in that paragraph:**
- [61] (named-source) Joshua Lowe, Transcript of evidence, p. 21.
- [62] (named-source) Vern Wall, General Committee Member, Bendigo Theatre Company, public hearing, Melbourne, 13 March 2025, Transcript of evidence, p. 57.
- [63] (named-source) Vern Wall, Transcript of evidence, p. 57.

**Suggested route (regex): `testimony`** — deciding: the Committee heard
**current claims-machine route:** source
**human route:** ______   **why:** ______

### F14 — §2.2.4, printed p23

**Finding (verbatim):** It has never been harder for Victorians to make a living in the cultural and creative industries.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> **〔Harder to make a living in the industries The Committee heard that ‘it’s never been harder for Australian artists to make a living’,70 with ‘overall artist income barely exceeding the minimum wage’.〕** Low incomes not only impact those currently working within the cultural and creative industries but also discourages others from pursuing a creative career. An economic study of professional artists in Australia highlighted that overall, the incomes of those in the industries fell 26% below the national average and 45% below

**Footnotes cited in that paragraph:**
- [71] (named-source) National Exhibitions Touring Support (NETS), Submission 28, p. 17.

**Suggested route (regex): `testimony`** — deciding: the Committee heard
**current claims-machine route:** evaluative
**human route:** ______   **why:** ______

### F15 — §2.2.4, printed p23

**Finding (verbatim):** Insecure or low wages impacts retention in the industries and discourages people from pursuing a career in the industries.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> Harder to make a living in the industries The Committee heard that ‘it’s never been harder for Australian artists to make a living’,70 with ‘overall artist income barely exceeding the minimum wage’. **〔Low incomes not only impact those currently working within the cultural and creative industries but also discourages others from pursuing a creative career.〕** An economic study of professional artists in Australia highlighted that overall, the incomes of those in the industries fell 26% below the national average and 45% below

**Footnotes cited in that paragraph:**
- [71] (named-source) National Exhibitions Touring Support (NETS), Submission 28, p. 17.

**Suggested route (regex): `testimony`** — deciding: the Committee heard
**current claims-machine route:** source
**human route:** ______   **why:** ______

### F16 — §2.2.5, printed p26

**Finding (verbatim):** Many people who want to engage with the Victorian cultural and creative industries cannot afford to do so due to the cost of living crisis.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> **〔The Committee heard that many who want to engage with the Victorian cultural and creative industries, cannot afford to do so due to the current cost of living crisis.〕** Audience Atlas Victoria 2024 found that 41%, or an estimated 2.10 million people, report engaging with the Victorian cultural and creative industries less, compared to five years ago. See Figure 2.3.

**Footnotes cited in that paragraph:**
- [82] (world) Creative Victoria, Audience Atlas Victoria 2024: Exploring the market for arts and culture in Victoria, report prepared by Morris Hargreaves Mcintyre, October 2024, p. 21.

**Suggested route (regex): `mixed`** — deciding: testimony:the Committee heard; world:footnote: Creative Victoria, Audience Atlas Victoria 2024: Exploring the market 
**current claims-machine route:** source
**human route:** ______   **why:** ______

### F17 — §2.2.5, printed p26

**Finding (verbatim):** Ticket prices are a key barrier for people wanting to engage with the Victorian cultural and creative industries.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> Of the 2.1 million people who attended fewer Victorian cultural and creative industries events, 53% cited that cost of living means they have cut back on everything, 47% cited that ticket prices were too high and 38% cited specifically that cost of living has limited their arts engagement. See Figure 2.4. **〔Multiple stakeholders echoed this finding by identifying ticket prices as a key barrier for people wanting to engage with the industries.〕** Musica Viva Australia stated: Many Australians are having to make tough decisions about their spending and are looking for free or cheap forms of entertainment.

**Footnotes cited in that paragraph:**
- [83] (world) Creative Victoria, Audience Atlas Victoria 2024: Exploring the market for arts and culture in Victoria, report prepared by Morris Hargreaves Mcintyre, October 2024, p. 24.
- [84] (named-source) Musica Viva Australia, Submission 23, p.7

**Suggested route (regex): `world`** — deciding: footnote: Creative Victoria, Audience Atlas Victoria 2024: Exploring the market 
**current claims-machine route:** source
**human route:** ______   **why:** ______

### F18 — §2.2.5, printed p26

**Finding (verbatim):** The COVID-19 pandemic and the cost of living crisis has accelerated the decline of volunteerism in the Victorian cultural and creative industries.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> **〔The Committee heard that many who want to engage with the Victorian cultural and creative industries, cannot afford to do so due to the current cost of living crisis.〕** Audience Atlas Victoria 2024 found that 41%, or an estimated 2.10 million people, report engaging with the Victorian cultural and creative industries less, compared to five years ago. See Figure 2.3.

**Footnotes cited in that paragraph:**
- [82] (world) Creative Victoria, Audience Atlas Victoria 2024: Exploring the market for arts and culture in Victoria, report prepared by Morris Hargreaves Mcintyre, October 2024, p. 21.

**Suggested route (regex): `mixed`** — deciding: testimony:the Committee heard; world:footnote: Creative Victoria, Audience Atlas Victoria 2024: Exploring the market 
**current claims-machine route:** source
**human route:** ______   **why:** ______

### F19 — §2.2.6, printed p27

**Finding (verbatim):** The COVID-19 pandemic exposed and intensified existing systemic issues within the Victorian cultural and creative industries, such as barriers to participation for women and gender diverse people, particularly those with caring responsibilities and long-term financial precarity.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> Similarly, Kate Larsen, Arts, Cultural and Non‐profit Consultant and Writer, in their submission to this Inquiry, noted that ‘studies of creative labour have been around for nearly two decades, stressing low pay, lack of pensions or health insurance, no paid holidays, precarity, self‐exploitation and, indeed, exploitation’. **〔The Committee heard that that challenges persist for women and gender diverse people that exacerbate these issues of financial precarity and insecurity, particularly for those juggling creative work with caring responsibilities.〕**

**Footnotes cited in that paragraph:**
- [92] (named-source) Kate Larsen, Submission 13, p. 3.

**Suggested route (regex): `testimony`** — deciding: the Committee heard
**current claims-machine route:** source
**human route:** ______   **why:** ______

### F20 — §3.2.1, printed p37

**Finding (verbatim):** Failing to index Victorian cultural and creative industries funding has led to a decrease in financial support in real terms.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> **〔Funding has not been indexed The Committee heard that Victorian cultural and creative industries funding is not always indexed to account for CPI increases.〕** As such, funding doesn’t stretch nearly as far as it may have done previously. The impacts of a lack of indexation are more intensely felt in the current cost of living crisis. Tim Stitz of Theatre Network Australia, told the Committee: Stagnation of funding levels... it is a net decrease. It does become harder ... It just becomes too hard for those organisations to survive.

**Footnotes cited in that paragraph:**
- [13] (named-source) Tim Stitz, Transcript of evidence, p. 23.

**Suggested route (regex): `testimony`** — deciding: the Committee heard
**current claims-machine route:** source
**human route:** ______   **why:** ______

### F21 — §3.2.2, printed p38

**Finding (verbatim):** Grant and reporting requirements can be very onerous, particularly on individuals and organisations who are made up of volunteers or part-time staff.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> **〔Grant and reporting requirements are onerous Many, if not all, stakeholders working in the Victorian cultural and creative industries who engaged with the Inquiry had experience applying for a grant program.〕** The Committee heard that ‘the cycle of funding applications, reporting and acquitting’ is one that takes up significant time and resources of the industries. Indeed, some stakeholders described these requirements as so onerous that it ‘sidelines its core business’. Nadja Kostich of St Martins Youth Arts Centre, told the Committee: We find ourselves now, through sheer necessity, ostensibly becoming centres for fundraising and advocacy who do art on the side, such is the struggle to make ends meet.

**Footnotes cited in that paragraph:**
- [20] (named-source) Nadja Kostich, Transcript of evidence, p. 2.
- [21] (named-source) Caitlin Dullard, Transcript of evidence, p. 17.
- [22] (named-source) Nadja Kostich, Transcript of evidence, p. 2.

**Suggested route (regex): `testimony`** — deciding: the Committee heard
**current claims-machine route:** source
**human route:** ______   **why:** ______

### F22 — §3.2.3, printed p40

**Finding (verbatim):** Unpredictable grant programs cause instability in the industries and prevents forward planning and investments in longer-term projects and strategies.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> Unpredictability also means that organisations who receive grant funding must ‘direct valuable resources away from program delivery’ in a constant pursuit of future grant funding. This is because many organisations face ‘a very strong reality that its budgets will not allow it to continue’ if it does not receive further grant funding in the next round. **〔As such, many organisations are restricted to only ‘operating on a four‐year grant cycle’ – preventing valuable forward planning and longer‐term investment.〕**

**Footnotes cited in that paragraph:**
- [33] (named-source) Arena Theatre Company, Submission 25, p. 2.
- [34] (named-source) Caitlin Dullard, Transcript of evidence, p. 19.
- [35] (named-source) Dr Ashley Robertson, Transcript of evidence, p. 49.

**Suggested route (regex): `unclear`** — deciding: (no class matched)
**current claims-machine route:** source
**human route:** ______   **why:** ______

### F23 — §3.2.4, printed p41

**Finding (verbatim):** Infrequent grant programs can negatively impact the industries by leaving greater numbers of applicants without government support for longer periods of time.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> **〔Grant opportunities are limited The Committee heard that grant programs are incredibly competitive, with some grant programs awarding funding to ‘as low as 4% of applicants’.〕** Stakeholders also expressed concern that there are fewer grant rounds compared to previous years, further limiting opportunities to successfully secure grant funding. Joshua Lowe of Theatre Network Australia, stated: There have been just two rounds of project investment since the beginning of the pandemic, when there used to be two a year.

**Footnotes cited in that paragraph:**
- [38] (named-source) Martin Jackson, Submission 36, p. 1.
- [39] (named-source) Joshua Lowe, Transcript of evidence, p. 21.

**Suggested route (regex): `testimony`** — deciding: the Committee heard
**current claims-machine route:** source
**human route:** ______   **why:** ______

### F24 — §3.2.5, printed p42

**Finding (verbatim):** A low maximum grant cap of $20,000 in some grant programs limits what recipients can achieve and may necessitate multiple grant applications to fund a project, consuming significant time and resources of applicants.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> Grant caps are too low The Committee heard that grant funding is often capped at a maximum of $20,000. This cap is ‘well below that of other states, [for example] in Western Australia it is capped at $80,000’. **〔Those grant programs which are capped at $20,000 limits what recipients can achieve.〕** The Committee heard that $20,000 ‘is quite a small amount of money to put a project on’,46 particularly in current economic circumstances. Katrina Cornwell of Rawcus Theatre Company, told the Committee:

**Footnotes cited in that paragraph:**
- [45] (named-source) Joshua Lowe, Transcript of evidence, p. 21.

**Suggested route (regex): `testimony`** — deciding: the Committee heard
**current claims-machine route:** source
**human route:** ______   **why:** ______

### F25 — §3.2.6, printed p43

**Finding (verbatim):** There needs to be a balance between investment in infrastructure and investment in artists. Without this balance, Victoria runs the risk of being dependent on interstate and international artists to fill these venues.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> **〔Grant programs place too great a focus on infrastructure, rather than artists The Committee heard that whilst investment in Victorian cultural and creative infrastructure is necessary, such investment cannot come at the expense of the artists who will fill such infrastructure with their work.〕** Joshua Lowe of Theatre Network Australia, told the Committee: If you look particularly in Victoria over the last decade or so at the investment in arts performance venues, the graph looks like this: it just goes up. We have got significant investment in the arts precinct in Southbank happening, which is partly driving that, but we have to make sure that those incredible venues are not empty. We have to think about what goes in those venues, what goes on the stages and what gets put on the walls. Unless we want it to be exclusively interstate and international artists, we need to make sure that people can have sustainable careers and that they have the resources to make that art here in Victoria. I imagine it is very attractive for a government to build a building and cut a ribbon – I would be delighted if I got to cut a ribbon – but it is not as exciting to invest in something that remains hidden until the performance or the exhibition happens. I think it is really important to have a balance of both. As Tim was saying, particularly in somewhere like regional Victoria, we want to ensure that the venues are connected with the community, that those local creatives can make and present work in their community, but also that that venue is resourced so that it can do proper audience development, it can reach out to audiences, it can make them feel

**Footnotes cited in that paragraph:**
- _(none resolved from this paragraph's markers)_

**Suggested route (regex): `testimony`** — deciding: the Committee heard
**current claims-machine route:** evaluative
**human route:** ______   **why:** ______

### F26 — §3.2.7, printed p44

**Finding (verbatim):** By requiring a developed concept or project, grant programs exclude creatives who are unable to undertake significant unpaid ideation work.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> Grants programs don’t cover the ideation process The Committee heard that grant programs cover a concept or project, but rarely compensate the ideation process that comes before. As a result, those in the Victorian cultural and creative industries must put in significant unpaid work to develop a project to the stage where it can qualify for grant funding. **〔This is exclusionary – as many creatives are unable to undertake the significant unpaid ideation work, particularly in a cost of living crisis.〕** Megan Champion of Sertori Consulting, told the Committee: Too often funding models prioritise tangible outcomes over the development of ideas. If we truly want a sustainable creative industry, we must invest in the three Ps: the person, the process and the product. This means supporting creatives not just in production but in education, incubation and cross‐industry collaboration. In order to get the funding, you need to have a project. You need to give a concept, an example of a project. All of the work that goes into that takes time, and then when you receive the funding, if you get the funding, you are only paid from when that funding starts. So that entire ideation process that comes behind that concept is not paid for. We used to have, built into funding structures, something called research and development, and that is not there anymore. That is what happens with projects like my project and like Rhayven’s project – we put these out there because we know that they are important, and we will only get a small portion of the funding. Then we have to make of the rest of it ourselves, and that means us going unpaid.

**Footnotes cited in that paragraph:**
- [51] (named-source) Megan Champion, Transcript of evidence, pp. 56, 62.

**Suggested route (regex): `testimony`** — deciding: the Committee heard
**current claims-machine route:** source
**human route:** ______   **why:** ______

### F27 — §3.2.8, printed p45

**Finding (verbatim):** The role that each level of government plays could be better defined to assist the industries in understanding where to go for relevant investment and support.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> **〔Confusion around what level of government funds what The Committee heard that many do not have clarity as to what role each level of government plays in funding the cultural and creative industries.〕** Kate Feilding, Chief Executive Officer, A New Approach, told the Committee:

**Footnotes cited in that paragraph:**
- _(none resolved from this paragraph's markers)_

**Suggested route (regex): `testimony`** — deciding: the Committee heard
**current claims-machine route:** evaluative
**human route:** ______   **why:** ______

### F28 — §3.2.9, printed p46

**Finding (verbatim):** People should not need to leave regional Victoria to engage with its cultural and creative industries.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> into Melbourne to access the Victorian cultural and creative industries. **〔Instead, the focus should be on developing the Victorian cultural and creative industries in Regional Victoria:〕** There have been programs that I am aware of that bus kids to Arts Centre Melbourne. That is fantastic; kids should have the experience of coming and seeing what metro kids can see any time of the year. But those communities should also have the ability to see things and to go to their local theatre to see things, whether that is music, dance, circus or whatever it is.

**Footnotes cited in that paragraph:**
- [55] (named-source) Tim Stitz, Transcript of evidence, p. 24.

**Suggested route (regex): `opinion`** — deciding: should
**current claims-machine route:** evaluative
**human route:** ______   **why:** ______

### F29 — §3.3.5, printed p57

**Finding (verbatim):** Victoria receives its fair share of funding from Creative Australia.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> Does Victoria get its fair share of federal funding? Creative Australia does break down its yearly investment by location. In 2023–24, Victoria received $28.7 million, or 27% of funding. **〔Victoria accounts for approximately Investment overview figures 25% of Australia’s total population, leading some stakeholders to conclude that Victoria receives its fair share of funding from Creative Australia.〕** See Figure 3.8.

**Footnotes cited in that paragraph:**
- [82] (named-source) Clare O’Neil, Transcript of evidence, p. 44.

**Suggested route (regex): `unclear`** — deciding: (no class matched)
**current claims-machine route:** evidence
**human route:** ______   **why:** ______

### F30 — §3.3.5, printed p57

**Finding (verbatim):** It is difficult to determine whether Victoria receives its fair share of other funding captured in the cultural funding by government dataset, as state and territory breakdowns are not released.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> The cultural funding by government dataset does not release a state and territory breakdown. As such, this dataset does not allow an understanding of how this total annual funding is distributed amongst the states and territories. Kate Fielding told the Committee: **〔If the cultural funding by government dataset was released, cut by state and territory data in terms of where federal government is sending its cultural expenditure on a state and territory basis, we would be able to have a data‐informed conversation about that in a coherent way.〕** But because that data is not released at the moment, we are really left with a very partial view, as you are putting together – getting bits of information and trying to see the whole picture from that.

**Footnotes cited in that paragraph:**
- [86] (named-source) Kate Fielding, Transcript of evidence, p. 27.

**Suggested route (regex): `testimony`** — deciding: told the Committee
**current claims-machine route:** data-gap
**human route:** ______   **why:** ______

### F31 — §3.3.6, printed p59

**Finding (verbatim):** Regional Victoria does not receive its fair share of funding from Creative Australia.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> **〔Regional Victoria is not getting its fair share The Committee heard that regional Victoria is not receiving its fair share of cultural and creative industry funding.〕** The Committee heard that ‘Creative Australia funds all Australian creatives but does not necessarily review its funding for an equitable spread between metro and regional’. The cultural funding by government dataset does not release a breakdown of funding based upon remoteness level or geography. As discussed above, further data transparency is required to ascertain whether regional Australia, and specifically regional Victoria, receives its fair share of federal funding. Creative Australia does break down funding by region. Out of the $237.4 Creative Australia invested in 2023–24, only $28.4 (approximately 12%) was invested in regional Australia. See Figure 3.9.

**Footnotes cited in that paragraph:**
- [87] (named-source) Jo Porter, Transcript of evidence, p. 12.
- [88] (world) Creative Australia, Annual Report 2023–24, p. 24.

**Suggested route (regex): `mixed`** — deciding: testimony:the Committee heard; world:footnote: Creative Australia, Annual Report 2023–24, p. 24.
**current claims-machine route:** evidence
**human route:** ______   **why:** ______

### F32 — §3.3.6, printed p59

**Finding (verbatim):** It is difficult to determine whether regional Victoria receives its fair share of other funding captured in the cultural funding by government dataset, as remoteness level or geography breakdowns are not released.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> Regional Victoria is not getting its fair share The Committee heard that regional Victoria is not receiving its fair share of cultural and creative industry funding. The Committee heard that ‘Creative Australia funds all Australian creatives but does not necessarily review its funding for an equitable spread between metro and regional’. **〔The cultural funding by government dataset does not release a breakdown of funding based upon remoteness level or geography.〕** As discussed above, further data transparency is required to ascertain whether regional Australia, and specifically regional Victoria, receives its fair share of federal funding. Creative Australia does break down funding by region. Out of the $237.4 Creative Australia invested in 2023–24, only $28.4 (approximately 12%) was invested in regional Australia. See Figure 3.9.

**Footnotes cited in that paragraph:**
- [87] (named-source) Jo Porter, Transcript of evidence, p. 12.
- [88] (world) Creative Australia, Annual Report 2023–24, p. 24.

**Suggested route (regex): `mixed`** — deciding: testimony:the Committee heard; world:footnote: Creative Australia, Annual Report 2023–24, p. 24.
**current claims-machine route:** data-gap
**human route:** ______   **why:** ______

### F33 — §4.3.1, printed p69

**Finding (verbatim):** No state or territory has an ABC headcount that accurately reflects its population size; an overrepresentation or underrepresentation of ABC employee headcount exists to varying degrees in all states and territories.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> **〔ABC headcount in Victoria The Committee received evidence from the ABC that provided ‘a breakdown of ABC full‐time equivalent (FTE) staff nationwide along with the staffing costs associated with each state and territory’, shown below in Table 4.1.11〕**

**Footnotes cited in that paragraph:**
- _(none resolved from this paragraph's markers)_

**Suggested route (regex): `unclear`** — deciding: (no class matched)
**current claims-machine route:** evidence
**human route:** ______   **why:** ______

### F34 — §4.3.1, printed p69

**Finding (verbatim):** The ABC's decision to relocate its Ultimo office to Parramatta was influenced by multiple external factors, including legal obligations under the ABC Enterprise Agreement 2022–2025, budget constraints that made New South Wales the most cost-effective option, the Parramatta office's proximity to key state departments and agencies, and a strategic goal to better reflect New South Wales' population demographics.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> Questioned on the cost of the move to Parramatta, Mark Tapley explained that ‘the net cost ... was borne by the ABC.’ **〔He stated that ‘[t]here is also just a budget constraint,’ in terms of the ABC’s decision to have an office in Parramatta.〕** The Committee received evidence from the ABC providing greater detail on the costing of the Parramatta relocation: The Sydney Accommodation Project brings together the elements of work required to deliver the new ABC site in Parramatta along with the staged refurbishment and restack of the Ultimo site, allowing the ABC to sub‐let up to 7 floors of the Ultimo tower. There will be nil additional cost to the taxpayer in the delivery of the project, as the costs for delivery are supported via the sale of aging assets and the sub‐leasing of up to seven floors of the ABC’s Ultimo facility. Total project spend to date is $66.4m at February 2025, with some $0.8m still to be incurred in finalising the project within the budget of $67.2m. Of this, Parramatta spend to February 2025 is $39.1m with some $0.4m still to be incurred, mostly within the Technology fit out component of the work.

**Footnotes cited in that paragraph:**
- [36] (named-source) Mark Tapley, Transcript of evidence, p. 40.
- [37] (named-source) Mark Tapley, Transcript of evidence, p. 40.
- [38] (other) ABC, hearing, response to questions on notice received 21 March 2025, p. 15.

**Suggested route (regex): `unclear`** — deciding: (no class matched)
**current claims-machine route:** source
**human route:** ______   **why:** ______

### F35 — §4.3.1, printed p69

**Finding (verbatim):** It is disappointing that both of Australia's national broadcasters the ABC and SBS expanded their presence to Western Sydney instead of Victoria.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> and we took the opportunity to lease space in Parramatta for up to 300 people. **〔We have moved the Sydney newsroom out there as well as ABC Sydney radio, the idea being that we decentralise out of Ultimo and we connect better with western Sydney, where I think 11 per cent of the Australian population lives.〕**

**Footnotes cited in that paragraph:**
- _(none resolved from this paragraph's markers)_

**Suggested route (regex): `unclear`** — deciding: (no class matched)
**current claims-machine route:** evaluative
**human route:** ______   **why:** ______

### F36 — §4.3.1, printed p69

**Finding (verbatim):** The ABC's impact and ability to culturally represent any particular state goes beyond the headcount or location of ABC offices.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> Mark Tapley further explained the ABC’s consideration of alternative locations to Parramatta, noting that ‘there is an issue that arises around redundancies if you are trying to move people.’ Regarding potential relocation to Victoria, he warned that ‘there would be significant redundancy costs involved in moving people holus‐bolus’. **〔Mark Tapley emphasised that ‘the ABC’s impact does go beyond that headcount’, adding that the organisation could ‘use the money and invest it in partnership with local cultural institutions and the independent production sector.〕** So we can have an impact around the country that is beyond just the headcount in the relevant cities’.

**Footnotes cited in that paragraph:**
- [39] (named-source) Mark Tapley, Transcript of evidence, p. 38.
- [40] (named-source) Mark Tapley, Transcript of evidence, p. 40.
- [41] (named-source) Mark Tapley, Transcript of evidence, p. 38.

**Suggested route (regex): `unclear`** — deciding: (no class matched)
**current claims-machine route:** source
**human route:** ______   **why:** ______

### F37 — §4.3.2, printed p74

**Finding (verbatim):** The Committee calls on the Victorian Government to continue to advocate to the Federal Government and the ABC for the return of a Victorian 7:30 Report.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> **〔Sacha Gregson drew the Committee’s attention towards The Newsreader which is ‘a really successful Victorian production with Werner Film Productions’ and reinforced how government partnerships shaped the show’s success.〕** She stated: It is an excellent example of a partnership that involves federal funding and VicScreen funding. We are incredibly grateful for our partnership with VicScreen, who are fundamental to closing finance on a number of our scripted and children’s content ... The Newsreader has won a swag of awards across the two series that have been broadcast so far, and we are part way through series 3 at the moment. The program is also broadcasting on the BBC and would have prominence in other territories on other

**Footnotes cited in that paragraph:**
- [67] (named-source) Sacha Gregson, Transcript of evidence, p. 33.

**Suggested route (regex): `unclear`** — deciding: (no class matched)
**current claims-machine route:** evaluative
**human route:** ______   **why:** ______

### F38 — §4.3.2, printed p74

**Finding (verbatim):** ABC spending is unevenly distributed across all states and territories when compared to their population sizes. No state or territory receives funding proportionate to its share of Australia's population.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> of ABC’s expenditure fromgreater 2020–2024 ($243.57 million). Victoria is the second most populous state and represents 26% of the nation’s total population, meaning that Victoria received a higher allocation of funding relative to its population. Similarly, New South Wales received disproportionately high funding relative to its population share. **〔In contrast, states like Queensland and Western Australia were allocated disproportionately lower funding levels compared to their respective population shares.〕** The Committee heard that differences in content production partially ‘come down to policy setting’ and that all ‘states are active with their various screen agencies trying to attract some production into other states’.

**Footnotes cited in that paragraph:**
- [46] (named-source) Mark Tapley, Transcript of evidence, p. 31.

**Suggested route (regex): `testimony`** — deciding: the Committee heard
**current claims-machine route:** evidence
**human route:** ______   **why:** ______

### F39 — §4.3.2, printed p74

**Finding (verbatim):** The ABC's content production decisions are shaped by state screen agency policies and funding availability. External pressures, including rising production costs and a 14% real-term reduction in the ABC's budget over the past decade, have constrained the broadcaster's capacity to create content and ability to expand its footprint in the nation's states and territories.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> **〔ABC spending on content production in Victoria From 2019–20 to 2023–24, the ABC spent a total of $728 million on internal and external co‐commissioned productions in Australia ‘with the total yearly average of $145 million’.〕** More specifically, the ABC allocated $243.57 million in that period to Victoria with the total yearly average for Victoria being $48.71 million. Table 4.4 ‘attributes expenditure to the state or territory where the majority of production activity and expenditure took place’.

**Footnotes cited in that paragraph:**
- [43] (other) ABC, hearing, response to questions on notice received 21 March 2025, p. 3.
- [44] (other) ABC, hearing, response to questions on notice received 21 March 2025, p. 3.
- [45] (other) ABC, hearing, response to questions on notice received 21 March 2025, p. 3.

**Suggested route (regex): `unclear`** — deciding: (no class matched)
**current claims-machine route:** source/evidence
**human route:** ______   **why:** ______

### F40 — §4.4.1, printed p76

**Finding (verbatim):** No state or territory reflects an SBS FTE employee number that is proportionate to its share of the Australia's population.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> **〔population by state and territory SBS FTE employees〕**

**Footnotes cited in that paragraph:**
- _(none resolved from this paragraph's markers)_

**Suggested route (regex): `unclear`** — deciding: (no class matched)
**current claims-machine route:** evidence
**human route:** ______   **why:** ______

### F41 — §4.4.1, printed p76

**Finding (verbatim):** The geographic distribution of SBS staff does not reflect the availability of its content. While a significant portion of SBS's workforce is based in New South Wales, its services are national and platform-agnostic, ensuring all Australians, regardless of location, can access and benefit from SBS content.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> There is the capacity to serve our audiences. **〔For example, in relation to our language services, there will be some language groups that are more highly represented in New South Wales and some that are more highly represented in Victoria, and so when we are looking to staff those services, we look at the available talent ...〕**

**Footnotes cited in that paragraph:**
- _(none resolved from this paragraph's markers)_

**Suggested route (regex): `unclear`** — deciding: (no class matched)
**current claims-machine route:** source
**human route:** ______   **why:** ______

### F42 — §4.4.1, printed p78

**Finding (verbatim):** The SBS's decision to relocate to Western Sydney was determined by funding requirements set by the Federal Government.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> and territories, compared to the distribution of Australia’s population (as of 30 September 2024, the most recent available data). **〔Outside of ‘SBS’s Sydney headquarters, the Melbourne office is the largest of its interstate offices’. 14% of SBS’s FTE employees are based in Victoria, which is disproportionately low compared to the State’s 26% share of the national population.〕** In contrast, 82% of SBS’S FTE employees are located in New South Wales, significantly exceeding the state’s 32% share of the

**Footnotes cited in that paragraph:**
- [72] (named-source) SBS, Submission 41, p. 2

**Suggested route (regex): `unclear`** — deciding: (no class matched)
**current claims-machine route:** source
**human route:** ______   **why:** ______

### F43 — §4.4.1, printed p78

**Finding (verbatim):** It is unclear why Victoria was not considered as an option for the Federal Government's relocation feasibility study for the SBS.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> **〔SBS headcount in Victoria The Committee heard that the SBS employed 192 FTE in Victoria.〕** These employees are spread across divisions, including NITV, TV & online content, news and current affairs, audio and language content, and more’. Table 4.7 outlines how SBS’s FTE employees are distributed across the nation’s states and territories.

**Footnotes cited in that paragraph:**
- [71] (named-source) SBS, Submission 41, p. 27

**Suggested route (regex): `testimony`** — deciding: the Committee heard
**current claims-machine route:** evaluative
**human route:** ______   **why:** ______

### F44 — §4.4.2, printed p80

**Finding (verbatim):** SBS provides significant financial and creative investment to Victoria.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> ‘When looking at premium Australian drama,’ **〔Clare O’Neil of SBS, noted that ‘SBS’s [spending] in Victoria is particularly significant’, with:〕**

**Footnotes cited in that paragraph:**
- _(none resolved from this paragraph's markers)_

**Suggested route (regex): `unclear`** — deciding: (no class matched)
**current claims-machine route:** source
**human route:** ______   **why:** ______

### F45 — §4.4.2, printed p80

**Finding (verbatim):** SBS's content remains nationally accessible regardless of the production location.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> **〔SBS spending on content production in Victoria The Committee heard that ‘[between] July 2021 and February 2025, SBS contributed to projects with a total combined budget of $55.8m’ in Victoria, funded through:〕**

**Footnotes cited in that paragraph:**
- [91] (named-source) Clare O’Neil, Transcript of evidence, p. 41.

**Suggested route (regex): `testimony`** — deciding: the Committee heard
**current claims-machine route:** source
**human route:** ______   **why:** ______

### F46 — §4.5.1, printed p82

**Finding (verbatim):** Between 2020 and 2025, the ABC commissioned 75 external and 52 internal projects where the majority of production took place in Victoria, demonstrating its vital role in supporting the state's screen industry.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> How is Victoria represented in ABC’s content As highlighted in Section 4.3.2, the ABC has produced numerous internal and external commissions in Victoria. **〔From 2020–21 to January 2025, the ABC has ‘commissioned 75 projects with the independent sector where the majority of production [took] place in Victoria’.〕** During the same time period, the ABC also produced 52 internal projects where the majority of production took place in Victoria, examples include:

**Footnotes cited in that paragraph:**
- _(none resolved from this paragraph's markers)_

**Suggested route (regex): `unclear`** — deciding: (no class matched)
**current claims-machine route:** evidence/evaluative
**human route:** ______   **why:** ______

### F47 — §4.5.1, printed p83

**Finding (verbatim):** The ABC partners with major Victorian cultural institutions to create educational content and showcase Victorian talent.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> **〔How is Victoria represented in ABC’s content As highlighted in Section 4.3.2, the ABC has produced numerous internal and external commissions in Victoria.〕** From 2020–21 to January 2025, the ABC has ‘commissioned 75 projects with the independent sector where the majority of production [took] place in Victoria’. During the same time period, the ABC also produced 52 internal projects where the majority of production took place in Victoria, examples include:

**Footnotes cited in that paragraph:**
- _(none resolved from this paragraph's markers)_

**Suggested route (regex): `unclear`** — deciding: (no class matched)
**current claims-machine route:** source
**human route:** ______   **why:** ______

### F48 — §4.5.1, printed p83

**Finding (verbatim):** Through initiatives like VicScreen internship and Indigenous placements, the ABC actively develops Victoria creative and cultural workforce.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> Mark Tapley of the ABC, emphasised that these projects ‘have made a significant contribution to the development of the sector, supporting employment and opportunity for talented Victorian creative workers’. He added: **〔They have allowed us to create award‐winning shows such as The Newsreader, Utopia and the internationally acclaimed Fisk, all of which were shot in Victoria and draw upon the state’s considerable creative talent, as were shows like Hard Quiz with Tom Gleeson, The Weekly with Charlie Pickering, Aunty Donna’s Coffee Cafe, Gold Diggers and Shaun Micallef’s Eve of Destruction.〕**

**Footnotes cited in that paragraph:**
- _(none resolved from this paragraph's markers)_

**Suggested route (regex): `unclear`** — deciding: (no class matched)
**current claims-machine route:** source
**human route:** ______   **why:** ______

### F49 — §4.5.1, printed p84

**Finding (verbatim):** The ABC delivers 16 regional radio programs in Victoria which produces over 112 hours of weekly local content across 9 stations, ensuring broad community representation.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> **〔How is Victoria represented in ABC’s content As highlighted in Section 4.3.2, the ABC has produced numerous internal and external commissions in Victoria.〕** From 2020–21 to January 2025, the ABC has ‘commissioned 75 projects with the independent sector where the majority of production [took] place in Victoria’. During the same time period, the ABC also produced 52 internal projects where the majority of production took place in Victoria, examples include:

**Footnotes cited in that paragraph:**
- _(none resolved from this paragraph's markers)_

**Suggested route (regex): `unclear`** — deciding: (no class matched)
**current claims-machine route:** evidence
**human route:** ______   **why:** ______

### F50 — §4.5.1, printed p85

**Finding (verbatim):** One of the main challenges for the ABC in producing content in regional Victoria is the lack of independent production sector proposals for shows to be made in regional Victoria.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> How is Victoria represented in ABC’s content As highlighted in Section 4.3.2, the ABC has produced numerous internal and external commissions in Victoria. **〔From 2020–21 to January 2025, the ABC has ‘commissioned 75 projects with the independent sector where the majority of production [took] place in Victoria’.〕** During the same time period, the ABC also produced 52 internal projects where the majority of production took place in Victoria, examples include:

**Footnotes cited in that paragraph:**
- _(none resolved from this paragraph's markers)_

**Suggested route (regex): `unclear`** — deciding: (no class matched)
**current claims-machine route:** source
**human route:** ______   **why:** ______

### F51 — §4.5.1, printed p85

**Finding (verbatim):** The cost of producing content in regional Victoria requires significant financial investment.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> **〔How is Victoria represented in ABC’s content As highlighted in Section 4.3.2, the ABC has produced numerous internal and external commissions in Victoria.〕** From 2020–21 to January 2025, the ABC has ‘commissioned 75 projects with the independent sector where the majority of production [took] place in Victoria’. During the same time period, the ABC also produced 52 internal projects where the majority of production took place in Victoria, examples include:

**Footnotes cited in that paragraph:**
- _(none resolved from this paragraph's markers)_

**Suggested route (regex): `unclear`** — deciding: (no class matched)
**current claims-machine route:** evaluative
**human route:** ______   **why:** ______

### F52 — §4.5.1, printed p85

**Finding (verbatim):** Producing content in regional areas delivers significant benefits, including social benefits, such as better representation and celebration of local communities and economic benefits, like job creation and increased local spending.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> **〔How is Victoria represented in ABC’s content As highlighted in Section 4.3.2, the ABC has produced numerous internal and external commissions in Victoria.〕** From 2020–21 to January 2025, the ABC has ‘commissioned 75 projects with the independent sector where the majority of production [took] place in Victoria’. During the same time period, the ABC also produced 52 internal projects where the majority of production took place in Victoria, examples include:

**Footnotes cited in that paragraph:**
- _(none resolved from this paragraph's markers)_

**Suggested route (regex): `unclear`** — deciding: (no class matched)
**current claims-machine route:** source
**human route:** ______   **why:** ______

### F53 — §4.5.2, printed p87

**Finding (verbatim):** SBS's receives about one-third of the ABC's funding, which restricts its ability to produce and commission content for and in Victoria.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> **〔Clare O’Neil emphasised that SBS ‘simply do not have comparable scale’ to the ABC, but ‘the significant commissioning investment it makes in Victoria reflects its strong engagement and commitment to the growth and sustainability of the sector’.〕** Clare O’Neil highlighted other examples of SBS’s Victorian made productions: In terms of non‐drama, we have documentary series Meet the Neighbours; episodes of the Great Australian Walks; Little J and Big Cuz, a First Nations Australian kids animation which has won Logies, awards from the Australian Teachers of Media and been nominated for AACTA’s award for best children’s program, all made in Victoria. And of course we have Yokayi Footy and Yokayi Footy Shorts, a celebration of Aussie Rules football from a First Nations perspective, and we have more in the pipeline.

**Footnotes cited in that paragraph:**
- _(none resolved from this paragraph's markers)_

**Suggested route (regex): `unclear`** — deciding: (no class matched)
**current claims-machine route:** evidence/source
**human route:** ______   **why:** ______

### F54 — §4.5.2, printed p87

**Finding (verbatim):** SBS sustains substantial Melbourne-based operations, producing language programs that reach 3.42 million Victorian viewers monthly, while supporting diverse communities through partnerships with VicScreen and local creatives.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> Clare O’Neil emphasised that SBS ‘simply do not have comparable scale’ to the ABC, but ‘the significant commissioning investment it makes in Victoria reflects its strong engagement and commitment to the growth and sustainability of the sector’. **〔Clare O’Neil highlighted other examples of SBS’s Victorian made productions:〕** In terms of non‐drama, we have documentary series Meet the Neighbours; episodes of the Great Australian Walks; Little J and Big Cuz, a First Nations Australian kids animation which has won Logies, awards from the Australian Teachers of Media and been nominated for AACTA’s award for best children’s program, all made in Victoria. And of course we have Yokayi Footy and Yokayi Footy Shorts, a celebration of Aussie Rules football from a First Nations perspective, and we have more in the pipeline.

**Footnotes cited in that paragraph:**
- _(none resolved from this paragraph's markers)_

**Suggested route (regex): `unclear`** — deciding: (no class matched)
**current claims-machine route:** source/evidence
**human route:** ______   **why:** ______

### F55 — §4.5.2, printed p87

**Finding (verbatim):** The SBS deeply engages with Victoria's multilingual and multicultural communities by supporting community events and major cultural events in Victoria.

**Report paragraph drawn from** (subject sentence in **〔 〕**):

> **〔Clare O’Neil emphasised that SBS ‘simply do not have comparable scale’ to the ABC, but ‘the significant commissioning investment it makes in Victoria reflects its strong engagement and commitment to the growth and sustainability of the sector’.〕** Clare O’Neil highlighted other examples of SBS’s Victorian made productions: In terms of non‐drama, we have documentary series Meet the Neighbours; episodes of the Great Australian Walks; Little J and Big Cuz, a First Nations Australian kids animation which has won Logies, awards from the Australian Teachers of Media and been nominated for AACTA’s award for best children’s program, all made in Victoria. And of course we have Yokayi Footy and Yokayi Footy Shorts, a celebration of Aussie Rules football from a First Nations perspective, and we have more in the pipeline.

**Footnotes cited in that paragraph:**
- _(none resolved from this paragraph's markers)_

**Suggested route (regex): `unclear`** — deciding: (no class matched)
**current claims-machine route:** source
**human route:** ______   **why:** ______

