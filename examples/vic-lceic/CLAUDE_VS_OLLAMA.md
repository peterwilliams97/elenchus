# LCEIC 8-claim refuter — Claude vs Ollama

Same 8 claims, same N=3, same faithfulness critic. Only the engine and what it sees differ.

| model | id | pricing | wall | sees |
|---|---|---|---|---|
| `claude-sonnet-4-6` | 48 calls | $3.59 | 8m48s | **full corpus** (25K tok/call) |
| `qwen3.6:27b-q4_K_M` (ollama) | 48 calls | $0 local | 15m15s | **BM25 top-8 passages** (~6.5K tok/call) |

**Agreement: 3/8.**

| # | id | claim | Claude (full corpus) | Ollama (retrieved) | |
|---|----|-------|----------------------|--------------------|---|
| 1 | F8 | COVID-19 took away critical training opportunities | faithful 2/3 | absent 3/3 | ✗ |
| 2 | F12 | COVID-19 led to worsening mental health in the ind | faithful 3/3 | partial 3/3 | ✗ |
| 3 | F17 | Ticket prices are a key barrier for people wanting | faithful 3/3 | faithful 3/3 | ✓ |
| 4 | F29 | Victoria receives its fair share of funding from C | partial 2/3 | faithful 3/3 | ✗ |
| 5 | F31 | Regional Victoria does not receive its fair share  | partial 2/3 | contradicted 3/3 | ✗ |
| 6 | M1 | Regional Victoria receives its fair share of fundi | contradicted 3/3 | contradicted 3/3 | ✓ |
| 7 | M2 | Witnesses said ticket prices have fallen since 201 | absent 3/3 | contradicted 3/3 | ✗ |
| 8 | M3 | According to Claire Febey of Creative Victoria, CO | absent 3/3 | absent 3/3 | ✓ |

F8/F12/F17/F29/F31 are real; M1/M2/M3 are planted mutations that should fail.
- **Mutations:** both engines reject all three (M1 contradicted both; M3 absent both; M2 absent/contradicted — both "not supported").
- **Real claims:** only F17 agrees. F17 is the one claim whose evidence retrieval fully covered (RETRIEVAL_REFUTER.md); F8/F12/F29/F31, whose evidence retrieval dropped, all diverge — F8 collapses to `absent` because the local judge simply never sees the supporting turns.
- **Direction isn't uniform:** retrieval makes Ollama harsher on F8/F12/F31 (less evidence → reject) but more lenient on F29 (the few retrieved passages happened to support it), so a small passage set is not a safe proxy for the corpus in either direction.