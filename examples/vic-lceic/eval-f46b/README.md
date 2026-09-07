# eval-f46b — single-claim rerun of F46b

F46b is the ABC-submission claim `...and produced 52 internal projects with the majority of
production in Victoria.` (cites `submission:41`, route `evidence`). It is the worked example the
faithfulness judge prompt now carries: the source says the 52 productions were **based in**
Victoria, not that the **majority of production** happened there — so its verdict is `contradicted`
and its stakes line contrasts the two.

This cell reruns F46b alone to inspect the leaf the new judge writes: the two ≤12-word plain
restatements `report_says` / `source_says` (replacing the old free-form `so_what`) and the stakes
line `renderStakes` assembles from them — `"The report says X. The source only says Y."`.

## Run

```sh
export ANTHROPIC_API_KEY=sk-ant-...
examples/vic-lceic/eval-f46b/run.sh          # ~$0.10, one Sonnet judge call
```

`run.sh` prints the leaf and the SUMMARY, and writes `claims-f46b.faithfulness.jsonl` +
`usage.jsonl` here. Corpus is `sources/` (hearings + submissions + qon); `hearings-all.txt` is set
aside for the run so passage IDs match the full run.
