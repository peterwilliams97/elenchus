# Ollama local model selection

_Written 2026-09-07. Registry figures cross-verified by a second session (elenchus-76)._

## Machine

| Fact               | Value                                   | Source |
|--------------------|-----------------------------------------|--------|
| Chip               | Apple M5 Pro                            | `system_profiler SPDisplaysDataType` |
| Unified memory     | 48 GiB (51,539,607,552 bytes)           | `sysctl hw.memsize` |
| GPU memory budget  | ~34 GB usable (macOS reserves the rest) | default `iogpu.wired_limit` ≈ 75% |
| Ollama             | 0.32.5 (`/opt/homebrew/bin/ollama`)     | `ollama --version` |
| Pre-existing model | `gemma3:4b` (3.3 GB) | `ollama list`    |

## Candidate matrix

Sizes are download size from the registry manifests (`registry.ollama.ai/v2/library/<m>/manifests/<tag>`).
"Fits?" = weights + a 16K fp16 KV cache + compute buffers against the ~34 GB budget.

| Tag                      | Exists | Size on disk | Quant     | Fits 34 GB w/ 16K KV | Notes |
|--------------------------|--------|---------|----------------|----------------------|-----|
| `qwen3.6:27b-q8_0`       | yes    | 30.0 GB | Q8_0           | **No** — 30 GB + KV + buffers overruns ~34 GB | dense 27B |
| `qwen3.6:27b-q4_K_M`     | yes    | 17.4 GB | Q4_K_M         | **Yes**, comfortable (~20 GB total) | dense 27B — **selected** |
| `qwen3.6:35b-a3b-q8_0`   | yes    | 38.7 GB | Q8_0           | No — exceeds budget outright | MoE 35B/3B active |
| `qwen3.6:35b-a3b-q4_K_M` | yes    | 23.9 GB | Q4_K_M         | Yes (~26 GB) | MoE 35B/3B active — runner-up |
| `gemma4:26b-a4b`         | **no** | —       | —              | —            | 404, does not exist |
| `gpt-oss:20b`            | yes    | 13.8 GB | MXFP4 (native) | Yes, easily   | MoE ~21B/3.6B active |

Note: the default `qwen3.6:27b` tag (17.8 GB) is Q4-class; `-q4_K_M` (17.4 GB) is the explicit pin used here.

## Decision

**Pull `qwen3.6:27b-q4_K_M`** (17.4 GB), the stated first-choice model at the quant that fits.

Rationale, following the requested priority order:

1. `qwen3.6:27b` is the top-priority model. The rule was "Q8 if it fits, else Q4_K_M." Q8_0 is
   30.0 GB of weights; adding a 16K KV cache and ollama's compute/graph buffers pushes it past the
   ~34 GB usable ceiling with no margin, so it does **not** safely fit → Q4_K_M.
2. Q4_K_M leaves ~14 GB of headroom, which allows a large context window and keeps the model fully
   GPU-resident (no CPU spill).
3. Dense 27B is preferred over the MoE runner-up for a faithfulness/grounded-QA task, where output
   quality matters more than the MoE's higher tok/s.

**Runner-up:** `qwen3.6:35b-a3b-q4_K_M` (23.9 GB) — MoE with 3B active params, materially faster
generation, at some quality cost versus the dense 27B. `gpt-oss:20b` (13.8 GB) is the lightweight
fallback.

## Correction to an earlier claim

Earlier in this session I stated `qwen3.6:27b` did not exist and that Qwen had no 27B size. That
was wrong — it was based on training-cutoff knowledge. The registry confirms `qwen3.6:27b` (dense)
and `qwen3.6:35b-a3b` (MoE) both exist; they postdate the training cutoff. `gemma4` does not exist;
`gemma3:27b` does.

## Caveat: "the refuter set"

The task asked to run "a 2K-token faithfulness-style prompt from the refuter set." No such prompt
set exists on disk. The files named `*-refuter-*.txt` under `papercutsoftware/ipp/xtmp/` and
`security-review-tooling/xtmp/` are Go **security-review** test output (falsifiers for code
findings), not faithfulness/RAG prompts. "The refuter set" as a source of faithfulness prompts is
an unbound referent.

Rather than invent one, the benchmark below uses a self-contained **grounded-QA (faithfulness)
probe**: a ~2K-token passage plus ten questions, of which two (season-ticket price, FY2025
patronage) are deliberately **not answerable** from the passage. A faithful model answers the eight
supported questions and returns "NOT STATED IN CONTEXT" for the two unsupported ones. The prompt
lives at `docs/faithfulness_prompt.txt`. This measures the same throughput the step asked for; if a
canonical refuter/faithfulness prompt set is meant to exist, point me at it and I'll re-run against
it.

## Benchmark

Model as pulled (`ollama show`): 27.8B params, architecture `qwen35`, 262144 (256K) max
context, Q4_K_M. Run via `/api/generate`, `num_ctx: 16384`, `temperature: 0`, streaming off.
Prompt: `docs/faithfulness_prompt.txt`.

| Metric | Value |
|---|---|
| Prompt tokens | 1,318 |
| Prompt-eval (prefill) | **364.9 tok/s** |
| Generation tokens | 2,226 (incl. `qwen35` reasoning tokens) |
| Generation (decode) | **15.1 tok/s** |
| Wall time (first call, incl. model load) | 158.8 s |
| Resident memory (`ollama ps`) | 17 GB, **100% GPU**, no CPU spill |

Notes:
- The passage tokenised to 1,318 tokens, somewhat under the "2K-token" target — reported as
  measured, not padded. Decode rate is unaffected by the shortfall.
- `qwen35` is a reasoning model: the 2,226 generation tokens are mostly an internal reasoning
  trace that ollama strips from `.response`; the 15.1 tok/s is the true decode rate and dominates
  wall time. Prefill (prompt processing) runs 24× faster at 364.9 tok/s.
- Resident footprint of 17 GB at 100% GPU confirms the fit analysis — weights + 16K KV stay well
  inside the ~34 GB budget.

### Faithfulness result: 10/10

| Q | Expected | Model | OK |
|---|---|---|---|
| 1 | 14 stations, 21.6 km | 14 stations, 21.6 km | ✓ |
| 2 | Series 70, 140/unit | Series 70, 140/unit | ✓ |
| 3 | 3.60, cap 4.80 | 3.60, 4.80 | ✓ |
| 4 | 6 min peak / 12 min off-peak | 6 / 12 min | ✓ |
| 5 | Ostmann Traction, 2017 | Ostmann Traction, 2017 | ✓ |
| 6 | 41.7M, 88% | 41.7M, 88% | ✓ |
| 7 | Aldergrove West & Millfield, lifts unfunded | correct | ✓ |
| 8 | **NOT STATED** (FY2025 patronage) | NOT STATED IN CONTEXT | ✓ |
| 9 | **NOT STATED** (Zone 1 season ticket) | NOT STATED IN CONTEXT | ✓ |
| 10 | Green Line | Green Line | ✓ |

Both faithfulness traps (Q8, Q9 — facts absent from the passage) were correctly refused; no
fabricated numbers. All eight supported questions answered correctly.

### Reproduce

```bash
cd docs
jq -Rs '{model:"qwen3.6:27b-q4_K_M", prompt:., stream:false,
         options:{num_ctx:16384, temperature:0}}' faithfulness_prompt.txt \
| curl -s http://127.0.0.1:11434/api/generate -d @- \
| jq '{prompt_tok:.prompt_eval_count, prompt_tok_s:(.prompt_eval_count/(.prompt_eval_duration/1e9)),
       gen_tok:.eval_count, gen_tok_s:(.eval_count/(.eval_duration/1e9)), response}'
```

Requires `ollama serve` running **outside** the command sandbox (a sandboxed server hits
`x509: OSStatus -26276` on the registry, and localhost may be blocked for the client).
