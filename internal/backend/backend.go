package backend

// backend defines the one seam every LLM provider implements, so assay's mode runners (substance,
// faithfulness, evidence, audit) call a single Complete method and never name a provider. spec/CLI.md
// §"Backends" is the contract. The package holds only the request/response shapes and the interface;
// each concrete provider lives in a subpackage (anthropic, ollama) and the test double in fake.

import "encoding/json"

// MaxTokens caps the output of every model call — Anthropic's max_tokens and Ollama's num_predict.
// Raise it if critiques truncate; a cutoff collapses a case to an error rather than a silent partial.
const MaxTokens = 1500

// Request is one model call. Cached is a stable prefix (e.g. a source transcript) that a provider
// may bill once and reuse across calls; a provider without prompt caching folds it into the prompt.
// WithTools asks for web search — honoured only by a backend whose SupportsWebSearch reports true.
//
// Schema, when set, constrains the response to that JSON schema: Anthropic via a forced strict tool
// call, Ollama via the `format` field. SchemaName is the tool name Anthropic forces. Temperature,
// when non-nil, is sent to the provider (nil leaves the provider default) — the faithfulness judge
// sends 0 for a single deterministic call and >0 for repeat sampling.
type Request struct {
	System      string
	Prompt      string
	Cached      string
	WithTools   bool
	Schema      json.RawMessage
	SchemaName  string
	Temperature *float64
}

// Source is a URL actually fetched during a web_search round trip — the provenance record the
// evidence cross-check compares against the model's self-reported citations. Field names match the
// legacy chain schema (JSON keys Title, URL), so a chain written before this type stays readable.
type Source struct{ Title, URL string }

// Usage is the per-call token accounting a run accumulates. Cache and web-search figures are zero on
// a provider that has neither (Ollama), which keeps the cost line honest rather than absent.
type Usage struct {
	InputTokens  int
	OutputTokens int
	CacheRead    int
	CacheCreate  int
	WebSearches  int
}

// Response carries the model's text, any retrieved sources, and the call's usage. Usage is populated
// even when Complete returns an error for a truncated (max_tokens / length) response — the call was
// still billed, so the caller accumulates it before surfacing the error.
type Response struct {
	Text    string
	Sources []Source
	Usage   Usage
}

// Backend is one LLM provider. Name is the short id stamped into the chain and header (anthropic,
// ollama, fake); Model is the provider-specific model id. SupportsWebSearch gates evidence-grounding
// retrieval — a false here means positive grounding is unreachable, never that it is faked.
type Backend interface {
	Name() string
	Model() string
	SupportsWebSearch() bool
	Complete(Request) (Response, error)
}
