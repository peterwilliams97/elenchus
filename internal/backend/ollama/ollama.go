package ollama

// ollama is the local-model backend: a non-streaming POST to a running Ollama server's /api/chat. It
// implements backend.Backend. Two differences from the Claude backend are load-bearing:
//   - No prompt caching and no server-side web search. Request.Cached is folded into the prompt;
//     SupportsWebSearch is false, so evidence grounding cannot retrieve on this backend (the right
//     result there is "unverifiable", never a fabricated citation).
//   - Reasoning models (qwen3.x and kin) emit a thinking block that consumes the token budget before
//     any answer. Think defaults false so the budget goes to the answer; the -think flag turns it on
//     for callers who want the reasoning trace and have raised the token cap to afford it.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"assay/internal/backend"
)

// DefaultBaseURL is the address of a local Ollama server; override via -ollama-url / OLLAMA_HOST.
const DefaultBaseURL = "http://localhost:11434"

var defaultHTTP = &http.Client{Timeout: 600 * time.Second} // local generation on a 27B model is slow

// Client is a configured Ollama backend. think carries the -think flag through to the request; a
// nil http uses the package default.
type Client struct {
	model   string
	baseURL string
	think   bool
	http    *http.Client
}

// New builds an Ollama backend. Empty baseURL falls back to DefaultBaseURL; a nil httpc uses the
// package default client.
func New(model, baseURL string, think bool, httpc *http.Client) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{model: model, baseURL: baseURL, think: think, http: httpc}
}

func (c *Client) Name() string            { return "ollama" }
func (c *Client) Model() string           { return c.model }
func (c *Client) SupportsWebSearch() bool { return false }

// Complete makes one /api/chat call. Cached is prepended to the prompt (Ollama has no cache seam);
// WithTools is ignored (no web search). A done_reason of "length" means the answer hit the token cap
// mid-stream — surfaced as an error so a truncated case collapses to "error", never a silent partial.
func (c *Client) Complete(req backend.Request) (backend.Response, error) {
	user := req.Prompt
	if req.Cached != "" {
		user = req.Cached + "\n\n" + req.Prompt
	}
	msgs := make([]chatMsg, 0, 2)
	if req.System != "" {
		msgs = append(msgs, chatMsg{Role: "system", Content: req.System})
	}
	msgs = append(msgs, chatMsg{Role: "user", Content: user})
	// num_ctx sized to the prompt: ~4 bytes/token, rounded up to a 2048 multiple, clamped to
	// [4096, 32768], so the whole prompt fits without over-allocating KV cache on a short one.
	temp := 0.0
	if req.Temperature != nil {
		temp = *req.Temperature
	}
	opts := chatOptions{NumPredict: backend.MaxTokens, Temperature: temp, NumCtx: numCtxFor(req.System, user)}
	creq := chatReq{
		Model:     c.model,
		Messages:  msgs,
		Stream:    false,
		Think:     c.think,
		KeepAlive: keepAlive, // hold the model + its KV cache hot across a grouped run
		Options:   opts,
	}
	if len(req.Schema) > 0 {
		creq.Format = req.Schema // Ollama constrains the answer to this JSON schema
	}
	body, _ := json.Marshal(creq)

	client := defaultHTTP
	if c.http != nil {
		client = c.http
	}
	httpReq, _ := http.NewRequest("POST", c.baseURL+"/api/chat", bytes.NewReader(body))
	httpReq.Header.Set("content-type", "application/json")
	resp, err := client.Do(httpReq)
	if err != nil {
		return backend.Response{}, err
	}
	raw, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	var cr chatResp
	if err := json.Unmarshal(raw, &cr); err != nil {
		return backend.Response{}, fmt.Errorf("unreadable ollama response: %.200s", string(raw))
	}
	if cr.Error != "" {
		return backend.Response{}, fmt.Errorf("ollama error: %s", cr.Error)
	}

	// prompt_eval_count / eval_count are Ollama's input/output token counts; there is no cache or
	// web-search billing, so those stay zero.
	out := backend.Response{
		Text: cr.Message.Content,
		Usage: backend.Usage{
			InputTokens:  cr.PromptEvalCount,
			OutputTokens: cr.EvalCount,
		},
	}
	if cr.DoneReason == "length" {
		return out, fmt.Errorf(
			"response truncated: done_reason=length (num_predict=%d); raise backend.MaxTokens or disable -think",
			backend.MaxTokens)
	}
	return out, nil
}

// keepAlive holds the model and its KV cache resident between calls, so a run that groups claims
// sharing passages keeps the cached prefix hot instead of paying a cold reload per claim.
const keepAlive = "30m"

// numCtxFor sizes the context window to the prompt: ~4 bytes/token, rounded up to a 2048 multiple and
// clamped to [4096, 32768]. Sizing it to the prompt keeps a short call cheap while never truncating a
// long grouped one.
func numCtxFor(system, user string) int {
	tokens := (len(system) + len(user)) / 4
	ctx := ((tokens+512)/2048 + 1) * 2048 // headroom for the answer, then round up
	return min(max(ctx, 4096), 32768)
}

type chatReq struct {
	Model     string          `json:"model"`
	Messages  []chatMsg       `json:"messages"`
	Stream    bool            `json:"stream"`
	Think     bool            `json:"think"`
	KeepAlive string          `json:"keep_alive,omitempty"`
	Format    json.RawMessage `json:"format,omitempty"`
	Options   chatOptions     `json:"options"`
}
type chatMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type chatOptions struct {
	NumPredict  int     `json:"num_predict"`
	Temperature float64 `json:"temperature"`
	NumCtx      int     `json:"num_ctx,omitempty"`
}
type chatResp struct {
	Message         chatMsg `json:"message"`
	DoneReason      string  `json:"done_reason"`
	PromptEvalCount int     `json:"prompt_eval_count"`
	EvalCount       int     `json:"eval_count"`
	Error           string  `json:"error"`
}
