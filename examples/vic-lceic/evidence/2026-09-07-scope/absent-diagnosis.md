# Why F10, F33, F54b came back "absent" — traced to the report's own citations

No model. Each finding's citation is quoted verbatim from `claims-machine.txt` (which transcribes the
report's footnotes), then checked against our fetched corpus (15 hearing transcripts + 42 written
submissions). Verdict for all three: **the cited truth-maker is outside our corpus, so "absent" is a
fact about our corpus, not a fault in the report or the judge.** None is a retrieval miss (the cited
content is not in the corpus to rank); none is unsourced.

| claim | report's citation (footnote · page) | cited source in our corpus? | classification |
|---|---|---|---|
| **F10** | §2.2.1 p16, fn46–47,50 — "ABS data via Theatre Network Australia, **Submission 19 Attachment 1** pp12–14" | **No.** We fetched Submission 19's main PDF; its Attachment 1 (the ABS participation/attendance series) has **no link on the submissions page** and was not harvested. The children/young-people decline appears nowhere in the corpus. | cited source **outside our corpus** → absent is about us. (The attachment is a real doc, just not published on the listing we harvested.) |
| **F33** | §4.3.1 p62, fn28–29 — "ABC, **response to questions on notice** received 21 Mar 2025 pp12–13"; also "ABC, Submission 41" | **Primary: No** — a QoN response is neither a hearing transcript nor a submission. **Secondary: partial** — Submission 41 *is* in the corpus and gives Victoria's ABC headcount ("937 paid active employees", line 43) but not the cross-state headcount-vs-population comparison the finding makes; that table is in the QoN response. | cited primary source (QoN response) **outside our corpus**; the in-corpus submission doesn't carry the comparative claim → absent is about us. |
| **F54b** | §4.5.2 p85, fn138–139 — "SBS, **Submission 42 p5**"; "SBS, **response to QoN** 10 Apr 2025 pp4–5" | **No.** The 3.42 million / ~49% viewer-reach figure (OzTAM VOZ) is **not in our Submission 42 text** (it carries budget figures — $334.9m, $159m, $55.8m — not audience reach) and appears nowhere in the corpus. The figure is the QoN response's (fn139); the "Submission 42 p5" cite does not resolve to it in the published PDF. | cited source **outside/absent from our corpus** → absent is about us. |

## Consequence

To turn any of these three from "absent" into a checkable faithfulness case, the corpus needs the
document the report actually rests them on: TNA Submission 19 **Attachment 1**, and the ABC and SBS
**responses to questions on notice** (21 Mar and 10 Apr 2025). Those are a distinct document class
(QoN responses + submission attachments) we have not fetched. Until then, "absent" is the correct and
honest verdict for our corpus — the judge is not wrong, and neither is the report.
