package anthropic

// anthropic is the Claude backend: a non-streaming POST to the Messages API, with prompt caching and
// server-side web search kept intact from assay's original inline client. It implements
// backend.Backend; the mode runners never construct or name it directly — main wires it behind the
// backend seam. Caching is intrinsic (a non-empty Request.Cached becomes an ephemeral prefix block),
// so there is no flag to disable it.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"assay/internal/backend"
)

const (
	apiURL           = "https://api.anthropic.com/v1/messages"
	retryMaxAttempts = 4 // 1 initial + 3 retries on 429/503/529
)

var (
	defaultHTTP = &http.Client{Timeout: 150 * time.Second}
	retryBase   = time.Second // overridden to 0 in tests for instant retry
)

// Client is a configured Claude backend. http is nil in production (defaultHTTP is used); tests
// inject a RoundTripper here to drive the retry and parse paths without a network.
type Client struct {
	model  string
	apiKey string
	http   *http.Client
}

// New builds a Claude backend. A nil httpc uses the package default 150s client; tests pass a stub.
func New(model, apiKey string, httpc *http.Client) *Client {
	return &Client{model: model, apiKey: apiKey, http: httpc}
}

func (c *Client) Name() string            { return "anthropic" }
func (c *Client) Model() string           { return c.model }
func (c *Client) SupportsWebSearch() bool { return true }

// Complete makes one Messages API call. A non-empty Cached becomes its own ephemeral content block
// ahead of the prompt, so a repeated source corpus is billed once and read thereafter; the system
// prompt is cached too. Usage is filled before any max_tokens error so the caller can bill a
// truncated response.
func (c *Client) Complete(req backend.Request) (backend.Response, error) {
	content := make([]textBlock, 0, 2)
	if req.Cached != "" {
		content = append(content, textBlock{Type: "text", Text: req.Cached, CacheControl: ephemeral})
	}
	content = append(content, textBlock{Type: "text", Text: req.Prompt})
	ar := apiReq{Model: c.model, MaxTokens: backend.MaxTokens,
		Messages: []apiMsg{{Role: "user", Content: content}}}
	if req.System != "" {
		// The system prompt is stable per mode, so cache it too — one breakpoint, reused every call.
		ar.System = []textBlock{{Type: "text", Text: req.System, CacheControl: ephemeral}}
	}
	if req.Temperature != nil {
		ar.Temperature = req.Temperature
	}
	if req.WithTools {
		ar.Tools = []apiTool{{Type: "web_search_20250305", Name: "web_search", MaxUses: 5}}
	}
	// Schema constrains the answer to JSON via a forced strict tool call — the response then arrives
	// as a tool_use block whose `input` is schema-valid JSON (no beta header; forced tool_choice is
	// supported on sonnet-4-6 / opus-4-8). Mutually exclusive with web search in practice: the judge
	// uses a schema, evidence grounding uses tools.
	if len(req.Schema) > 0 {
		name := req.SchemaName
		if name == "" {
			name = "emit"
		}
		ar.Tools = []apiTool{{Name: name, Description: "Return the result as JSON.", InputSchema: req.Schema, Strict: true}}
		ar.ToolChoice = &toolChoice{Type: "tool", Name: name}
	}
	body, _ := json.Marshal(ar)

	client := defaultHTTP
	if c.http != nil {
		client = c.http
	}

	var raw []byte
	for attempt := 0; ; attempt++ {
		httpReq, _ := http.NewRequest("POST", apiURL, bytes.NewReader(body))
		httpReq.Header.Set("content-type", "application/json")
		httpReq.Header.Set("x-api-key", c.apiKey)
		httpReq.Header.Set("anthropic-version", "2023-06-01")

		resp, err := client.Do(httpReq)
		if err != nil {
			return backend.Response{}, err
		}
		raw, _ = io.ReadAll(resp.Body)
		resp.Body.Close() // explicit close before any retry; do not defer across iterations

		if !retryable(resp.StatusCode) || attempt >= retryMaxAttempts-1 {
			break
		}
		time.Sleep(retryDelay(resp.Header.Get("Retry-After"), attempt))
	}

	var resp apiResp
	if err := json.Unmarshal(raw, &resp); err != nil {
		return backend.Response{}, fmt.Errorf("unreadable response: %.200s", string(raw))
	}
	if resp.Error != nil {
		return backend.Response{}, fmt.Errorf("api error: %s", resp.Error.Message)
	}

	// Bill usage before checking stop_reason — a truncated response was still billed.
	out := backend.Response{}
	if resp.Usage != nil {
		webSearches := 0
		if resp.Usage.ServerToolUse != nil {
			webSearches = resp.Usage.ServerToolUse.WebSearchRequests
		}
		// Fall back to counting web_search tool_use blocks if server_tool_use absent.
		if webSearches == 0 {
			for _, b := range resp.Content {
				if b.Type == "server_tool_use" || b.Type == "tool_use" {
					var name struct {
						Name string `json:"name"`
					}
					if json.Unmarshal(b.Input, &name) == nil && name.Name == "web_search" {
						webSearches++
					}
				}
			}
		}
		out.Usage = backend.Usage{
			InputTokens:  resp.Usage.InputTokens,
			OutputTokens: resp.Usage.OutputTokens,
			CacheRead:    resp.Usage.CacheReadInputTokens,
			CacheCreate:  resp.Usage.CacheCreationInputTokens,
			WebSearches:  webSearches,
		}
	}

	if resp.StopReason == "max_tokens" {
		return out, fmt.Errorf(
			"response truncated: stop_reason=max_tokens (limit=%d tokens); raise backend.MaxTokens",
			backend.MaxTokens)
	}

	var sb strings.Builder
	for _, b := range resp.Content {
		switch b.Type {
		case "text":
			sb.WriteString(b.Text)
		case "tool_use":
			// A forced schema tool call returns the answer as the tool input (schema-valid JSON), not
			// as text; surface it as the response text so callJSON parses it exactly as before.
			sb.Write(b.Input)
		case "web_search_tool_result":
			// b.Content is the raw JSON value of the "content" field: either a []web_search_result
			// array or a web_search_tool_result_error object. Unmarshal directly into a slice; an
			// error object (not an array) fails silently and leaves Sources empty, which the caller's
			// grounding cross-check reads as "nothing retrieved".
			var results []struct {
				Type  string `json:"type"`
				URL   string `json:"url"`
				Title string `json:"title"`
			}
			if json.Unmarshal(b.Content, &results) == nil {
				for _, r := range results {
					if r.Type == "web_search_result" {
						out.Sources = append(out.Sources, backend.Source{Title: r.Title, URL: r.URL})
					}
				}
			}
		}
	}
	out.Text = sb.String()
	return out, nil
}

// retryable reports whether an HTTP status code warrants a retry.
func retryable(code int) bool { return code == 429 || code == 503 || code == 529 }

// retryDelay returns how long to wait before the next attempt. It honours the Retry-After header
// (integer seconds) when present; otherwise uses full-jitter exponential backoff capped at 30 s.
// retryBase==0 (tests) always returns 0.
func retryDelay(retryAfter string, attempt int) time.Duration {
	if s, err := strconv.Atoi(strings.TrimSpace(retryAfter)); err == nil && s > 0 {
		return time.Duration(s) * time.Second
	}
	d := min(retryBase<<uint(attempt), 30*time.Second) // 1 s, 2 s, 4 s, … capped at 30 s
	// rand.Int63n(n+1) with n==0 returns 0, so retryBase==0 sleeps for 0.
	return time.Duration(rand.Int63n(int64(d) + 1)) // full jitter: [0, d]
}

type apiReq struct {
	Model       string      `json:"model"`
	MaxTokens   int         `json:"max_tokens"`
	System      []textBlock `json:"system,omitempty"`
	Messages    []apiMsg    `json:"messages"`
	Tools       []apiTool   `json:"tools,omitempty"`
	ToolChoice  *toolChoice `json:"tool_choice,omitempty"`
	Temperature *float64    `json:"temperature,omitempty"`
}
type toolChoice struct {
	Type string `json:"type"` // "tool"
	Name string `json:"name"`
}
type apiMsg struct {
	Role    string      `json:"role"`
	Content []textBlock `json:"content"`
}

// textBlock is one content block. A non-nil CacheControl marks the prefix up to and including this
// block as cacheable, so a repeated source corpus is billed once at cache-write rates and then read.
type textBlock struct {
	Type         string        `json:"type"` // always "text"
	Text         string        `json:"text"`
	CacheControl *cacheControl `json:"cache_control,omitempty"`
}
type cacheControl struct {
	Type string `json:"type"` // "ephemeral"
}

// ephemeral is the shared marker for every cache breakpoint; the API caps a request at four.
var ephemeral = &cacheControl{Type: "ephemeral"}

type apiTool struct {
	Type        string          `json:"type,omitempty"` // "web_search_20250305"; empty for a custom schema tool
	Name        string          `json:"name"`
	MaxUses     int             `json:"max_uses,omitempty"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"input_schema,omitempty"`
	Strict      bool            `json:"strict,omitempty"`
}
type apiUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	ServerToolUse            *struct {
		WebSearchRequests int `json:"web_search_requests"`
	} `json:"server_tool_use"`
}
type apiResp struct {
	Content    []apiBlock `json:"content"`
	StopReason string     `json:"stop_reason"`
	Usage      *apiUsage  `json:"usage"`
	Error      *struct {
		Message string `json:"message"`
	} `json:"error"`
}
type apiBlock struct {
	Type    string          `json:"type"`
	Text    string          `json:"text"`
	Input   json.RawMessage `json:"input"`
	Content json.RawMessage `json:"content"` // populated for web_search_tool_result blocks
}
