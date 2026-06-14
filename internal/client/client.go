// Package client is the Anthropic API boundary: HTTP transport, retry, and
// response parsing (including usage accounting and truncation handling).
//
// It is the ONLY place a test fake for the API lives (see Stub / DoerFunc in
// fake.go); no other package stubs the network. Callers pass a system+user
// prompt in and get parsed JSON out via CallJSON.
package client

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
)

const (
	defaultBaseURL   = "https://api.anthropic.com/v1/messages"
	anthropicVersion = "2023-06-01"
	maxTokens        = 1500 // frozen at the spec value (spec/CLI.md, maxTokens constant)
	retryMaxAttempts = 4    // 1 initial + 3 retries (spec/BEHAVIOR.md)
	retryCap         = 30 * time.Second
	httpTimeout      = 120 * time.Second // a stalled call must not hang the run

	// jsonRetrySuffix is appended to the system prompt for the single retry after
	// a parse failure (spec/BEHAVIOR.md, JSON parse + retry rule).
	jsonRetrySuffix = "\n\nReturn ONLY raw JSON. No prose, no markdown, no backticks. First character must be { or [."
)

// Doer is the HTTP seam. The real client uses *http.Client; tests inject a fake.
// This is the only network seam in the codebase.
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

// Usage accumulates token and call accounting across a run.
type Usage struct {
	Calls             int
	InputTokens       int
	OutputTokens      int
	CacheReadTokens   int
	CacheCreateTokens int
}

// Config is the immutable construction input for a Client.
type Config struct {
	APIKey  string
	Model   string
	BaseURL string // defaults to the Anthropic messages endpoint
	HTTP    Doer   // defaults to http.DefaultClient
}

// Client calls the Anthropic Messages API. It is not safe for concurrent use; a
// substance run grades claims sequentially.
type Client struct {
	apiKey    string
	model     string
	baseURL   string
	http      Doer
	retryBase time.Duration
	usage     Usage
}

// New builds a Client from cfg, applying defaults for BaseURL and HTTP.
func New(cfg Config) *Client {
	c := &Client{
		apiKey:    cfg.APIKey,
		model:     cfg.Model,
		baseURL:   cfg.BaseURL,
		http:      cfg.HTTP,
		retryBase: time.Second,
	}
	if c.baseURL == "" {
		c.baseURL = defaultBaseURL
	}
	if c.http == nil {
		c.http = &http.Client{Timeout: httpTimeout}
	}
	return c
}

// Usage returns the accumulated accounting so far.
func (c *Client) Usage() Usage { return c.usage }

// Model returns the model ID this client calls.
func (c *Client) Model() string { return c.model }

// --- wire types ----------------------------------------------------------------

type apiRequest struct {
	Model     string       `json:"model"`
	MaxTokens int          `json:"max_tokens"`
	System    string       `json:"system"`
	Messages  []apiMessage `json:"messages"`
}

type apiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type apiResponse struct {
	Content    []apiContent `json:"content"`
	StopReason string       `json:"stop_reason"`
	Usage      apiUsage     `json:"usage"`
}

type apiContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type apiUsage struct {
	InputTokens       int `json:"input_tokens"`
	OutputTokens      int `json:"output_tokens"`
	CacheReadTokens   int `json:"cache_read_input_tokens"`
	CacheCreateTokens int `json:"cache_creation_input_tokens"`
}

// CallJSON sends system+user to the API and unmarshals the JSON response into v.
// On a parse failure it retries once with an augmented system prompt; a second
// parse failure returns an error (the caller records verdict "error").
func (c *Client) CallJSON(system, user string, v any) error {
	raw, err := c.callClaude(system, user)
	if err != nil {
		return err
	}
	if err := unmarshalLoose(raw, v); err == nil {
		return nil
	}
	raw, err = c.callClaude(system+jsonRetrySuffix, user)
	if err != nil {
		return err
	}
	if err := unmarshalLoose(raw, v); err != nil {
		return fmt.Errorf("response not valid JSON after retry: %w", err)
	}
	return nil
}

// callClaude performs one logical API call with HTTP retry on 429/503/529, then
// returns the concatenated response text. Usage is accumulated before the
// truncation check, so a truncated (billed) response still counts.
func (c *Client) callClaude(system, user string) (string, error) {
	reqBody, err := json.Marshal(apiRequest{
		Model:     c.model,
		MaxTokens: maxTokens,
		System:    system,
		Messages:  []apiMessage{{Role: "user", Content: user}},
	})
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	for attempt := 0; ; attempt++ {
		req, err := http.NewRequest(http.MethodPost, c.baseURL, bytes.NewReader(reqBody))
		if err != nil {
			return "", fmt.Errorf("build request: %w", err)
		}
		req.Header.Set("x-api-key", c.apiKey)
		req.Header.Set("anthropic-version", anthropicVersion)
		req.Header.Set("content-type", "application/json")

		resp, err := c.http.Do(req)
		if err != nil {
			return "", fmt.Errorf("http do: %w", err)
		}
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return "", fmt.Errorf("read response body: %w", err)
		}

		if retryable(resp.StatusCode) && attempt < retryMaxAttempts-1 {
			time.Sleep(c.retryDelay(resp.Header, attempt))
			continue
		}
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("anthropic api: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
		}

		var ar apiResponse
		if err := json.Unmarshal(body, &ar); err != nil {
			return "", fmt.Errorf("decode response: %w", err)
		}
		c.usage.Calls++
		c.usage.InputTokens += ar.Usage.InputTokens
		c.usage.OutputTokens += ar.Usage.OutputTokens
		c.usage.CacheReadTokens += ar.Usage.CacheReadTokens
		c.usage.CacheCreateTokens += ar.Usage.CacheCreateTokens

		if ar.StopReason == "max_tokens" {
			return "", fmt.Errorf("response truncated: stop_reason=max_tokens (limit=%d tokens); raise maxTokens constant", maxTokens)
		}
		return textOf(ar), nil
	}
}

func retryable(status int) bool {
	return status == http.StatusTooManyRequests || // 429
		status == http.StatusServiceUnavailable || // 503
		status == 529 // overloaded
}

// retryDelay honors a positive Retry-After header (integer seconds); otherwise
// full-jitter exponential backoff in [0, min(retryBase<<attempt, 30s)].
func (c *Client) retryDelay(h http.Header, attempt int) time.Duration {
	if ra := strings.TrimSpace(h.Get("Retry-After")); ra != "" {
		if secs, err := strconv.Atoi(ra); err == nil && secs > 0 {
			return time.Duration(secs) * time.Second
		}
	}
	hi := c.retryBase << attempt
	if hi > retryCap {
		hi = retryCap
	}
	if hi <= 0 {
		return 0
	}
	return time.Duration(rand.Int63n(int64(hi)))
}

func textOf(ar apiResponse) string {
	var b strings.Builder
	for _, c := range ar.Content {
		if c.Type == "text" {
			b.WriteString(c.Text)
		}
	}
	return b.String()
}

// unmarshalLoose strips markdown fences and any trailing comma before a } or ],
// then unmarshals. CallJSON's parse-retry is the backstop for anything else.
func unmarshalLoose(s string, v any) error {
	return json.Unmarshal([]byte(stripTrailingCommas(extractJSON(s))), v)
}

func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSpace(s)
		s = strings.TrimSuffix(s, "```")
		s = strings.TrimSpace(s)
	}
	return s
}

// stripTrailingCommas removes a comma that sits immediately before a } or ]
// (ignoring whitespace) when it is outside any quoted string. JSON has no trailing
// commas, but models emit them; a naive regex also struck commas inside string
// values (e.g. "60%,}"), so this scans quote state. Structural bytes are all ASCII
// (<0x80) and never collide with a UTF-8 multibyte sequence, so byte scanning is
// safe.
func stripTrailingCommas(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	inStr, esc := false, false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if inStr {
			inStr, esc = nextStringState(c, esc)
			b.WriteByte(c)
			continue
		}
		if c == ',' && commaBeforeBracket(s, i+1) {
			continue // drop the structural trailing comma
		}
		if c == '"' {
			inStr = true
		}
		b.WriteByte(c)
	}
	return b.String()
}

// nextStringState advances the in-string scan, given the current byte and whether
// it is escaped: returns whether still inside the string and whether the next byte
// is escaped.
func nextStringState(c byte, esc bool) (inStr, nextEsc bool) {
	switch {
	case esc:
		return true, false
	case c == '\\':
		return true, true
	case c == '"':
		return false, false
	}
	return true, false
}

// commaBeforeBracket reports whether the next non-whitespace byte at or after j is
// } or ].
func commaBeforeBracket(s string, j int) bool {
	for j < len(s) && (s[j] == ' ' || s[j] == '\t' || s[j] == '\n' || s[j] == '\r') {
		j++
	}
	return j < len(s) && (s[j] == '}' || s[j] == ']')
}
