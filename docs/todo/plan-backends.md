# plan-backends.md — multi-provider LLM backends

Add non-Anthropic providers behind the existing `backend.Backend` seam. Serves the cross-model
calibration goal (`SESSION.md` 3c/W7): a *different model family* breaks the shared Producer/Critic
blind spot (`CLAUDE.md` axis boundary — both roles share one `c.model`) far better than another
Anthropic model. Roadmap item 3.3. Assessed 2026-09-12.

## The seam this builds on

Every provider implements the 4-method `backend.Backend` interface (`internal/backend/backend.go:59`):
`Name`, `Model`, `SupportsWebSearch`, `Complete(Request) (Response, error)`. Mode runners call
`Complete` and never name a provider (`spec/CLI.md` §Backends is the contract). A new provider is a
subpackage plus one `case` in the dispatch at `assay.go:319-328`. No new dependency — `anthropic.go`
(263 lines) hand-rolls raw HTTP/JSON against the API URL; new backends do the same (`one wrapper per
dependency`). Test via the injected `*http.Client` RoundTripper seam (`New(model, apiKey, httpc)`),
no live calls.

## Key decision: one OpenAI-compatible backend, not N vendor backends

Most of the field ships an OpenAI-compatible HTTP API, so a single
`internal/backend/openaicompat` parameterised by `{baseURL, apiKey, model}` reaches all of these
with no per-vendor code:

| Provider                       | How                                   | Notes |
|--------------------------------|---------------------------------------|-------|
| OpenAI                         | native                                | base case |
| **DeepSeek** (V3/R1)           | `api.deepseek.com`, OpenAI-compatible | strongest open Chinese model, drop-in |
| **Qwen** (Alibaba, Qwen-Max/2.5) | DashScope OpenAI-compat endpoint    | |
| Moonshot (Kimi), Zhipu (GLM)   | OpenAI-compat endpoints               | |
| **Gemini**                     | Google's OpenAI-compat layer (`…/v1beta/openai/`) | works; native has better schema support |
| OpenRouter / Groq / local vLLM | OpenAI-compat                         | reach everything through one gateway |

One ~1-day backend covers OpenAI + the Chinese models + Gemini-via-compat. The `-backend` flag gains
an `openaicompat` case reading a base-URL/model env pair.

## Field parity (OpenAI-shaped)

- `Schema`/`SchemaName` → `response_format:{type:"json_schema",strict:true}` — *easier* than
  Anthropic's forced-tool-call because the response is JSON text, no `tool_use` unwrap.
- `Temperature` → `temperature`. `Cached` folds into the prompt (OpenAI auto-caches ≥1024-token
  prefixes; no explicit field).
- Usage: `prompt_tokens`→`InputTokens`, `completion_tokens`→`OutputTokens`, `cached_tokens`→
  `CacheRead`, `CacheCreate`=0.

## The load-bearing gotcha: strict schema vs. JSON mode

`faithJudge` → `callSchema` needs the verdict to come back schema-valid (right enum, right fields).
Providers split:

- **OpenAI + native Gemini** do *constrained decoding* — the model cannot emit an off-spec verdict.
- **DeepSeek, Qwen, most compat endpoints, Gemini-via-compat** offer only JSON *mode* — valid JSON,
  not schema-enforced. A verdict could come back `"yes"` instead of the enum.

The repo already tolerates messy JSON (`extractJSON`, `unmarshalLoose`) but does **not** yet validate
the verdict enum — that is `SESSION.md` **W10** ("verdict-enum validation … expected red today"). So
a non-strict backend **promotes W10 from nice-to-have to prerequisite**: normalize-or-reject the
verdict before trusting a DeepSeek/Qwen run. Do W10 alongside this, or first.

## Other gotchas

1. Reasoning models (o-series, gpt-5 reasoning, DeepSeek-R1) use `max_completion_tokens`, reserve
   output budget for hidden reasoning, and may reject `temperature`; a 1500-token `MaxTokens` cap
   (`backend.go:12`) truncates them. Target a standard chat model first (gpt-4.1 / gpt-5 non-reasoning,
   DeepSeek-V3).
2. Web search / `SupportsWebSearch()` is the one non-trivial piece: `crossCheckEvidence` compares
   model-reported citations against actually-fetched URLs (Anthropic exposes these via
   `web_search_tool_result` blocks; OpenAI/Gemini annotations carry different provenance semantics).
   Ship the first cut with `SupportsWebSearch() bool { return false }` — the schema-enforced judge
   path is the main one, and deferring grounding matches the project's treatment of it as the weakest
   column.
3. Key/model plumbing: needs an `OPENAI_API_KEY` (and per-provider) analog to
   `ANTHROPIC_API_KEY`/`setkey.sh`, and a model-env name alongside `ANTHROPIC_MODEL`.
4. Data egress: any non-Anthropic backend sends the source corpus to that provider — fine for the
   public vic-lceic hearings, a per-corpus trust call for anything sensitive, no different in kind for
   a Chinese vs. US endpoint. Not a code issue; a decision to record per corpus.
5. **Naming.** "Codex" as a model is retired; this is an OpenAI *chat-model* backend
   (gpt-4.1 / gpt-5 / o-series). "Codex" now names OpenAI's CLI/agent, unrelated to this seam.

## Scope & sequence

- **W10 (verdict-enum validation)** first or alongside — the prerequisite for trusting any
  non-strict backend.
- **`openaicompat` backend** (~1 day): unlocks OpenAI + DeepSeek + Qwen + Gemini-compat at once. New
  `internal/backend/openaicompat/{*.go,*_test.go}` + one dispatch case + `spec/CLI.md` §Backends update
  in the same commit.
- **Native Gemini backend** (+~1 day): only if you want Gemini's real constrained decoding or
  Google-Search grounding fidelity — needs a JSON-Schema→Gemini-schema-dialect translation (Gemini
  accepts an OpenAPI subset, not full JSON Schema). Otherwise skip; use the compat layer.
