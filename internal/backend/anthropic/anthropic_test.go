package anthropic

// anthropic_test pins the Claude backend's transport and request shaping: retry/backoff on
// 429/503/529, the max_tokens truncation error (with usage still billed), web_search_tool_result
// parsing, and the cache_control layout of the outgoing request. These are white-box tests — they
// set the unexported retryBase and inspect the unexported apiReq — because that is exactly the
// machinery Complete owns. They drive Complete through a stubbed RoundTripper, never a network.

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"assay/internal/backend"
)

// TestComplete429ThenSuccess confirms a 429 followed by a 200 succeeds after one retry and returns
// the expected content.
func TestComplete429ThenSuccess(t *testing.T) {
	retryBase = 0
	t.Cleanup(func() { retryBase = time.Second })

	attempts := 0
	cl := New("claude-sonnet-4-6", "test-key",
		&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			attempts++
			if attempts == 1 {
				return fakeResp(429, `{"error":{"type":"rate_limit_error","message":"rate limited"}}`,
					map[string]string{"Retry-After": "0"}), nil
			}
			return fakeResp(200, okBody, nil), nil
		})})
	resp, err := cl.Complete(backend.Request{System: "sys", Prompt: "prompt"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Text != "hello" {
		t.Errorf("want %q, got %q", "hello", resp.Text)
	}
	if attempts != 2 {
		t.Errorf("want 2 attempts, got %d", attempts)
	}
}

// TestComplete529PersistentFails confirms persistent 529s exhaust retries and return a clean error.
func TestComplete529PersistentFails(t *testing.T) {
	retryBase = 0
	t.Cleanup(func() { retryBase = time.Second })

	attempts := 0
	cl := New("claude-sonnet-4-6", "test-key",
		&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			attempts++
			return fakeResp(529, `{"error":{"type":"overloaded_error","message":"overloaded"}}`, nil), nil
		})})
	if _, err := cl.Complete(backend.Request{System: "sys", Prompt: "prompt"}); err == nil {
		t.Fatal("expected error on persistent 529, got nil")
	}
	if attempts != retryMaxAttempts {
		t.Errorf("want %d attempts, got %d", retryMaxAttempts, attempts)
	}
}

// TestCompleteMaxTokensTruncation confirms a max_tokens stop_reason is returned as a named error and
// that usage is populated on the response before the error is returned.
func TestCompleteMaxTokensTruncation(t *testing.T) {
	cl := New("claude-sonnet-4-6", "test-key",
		&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return fakeResp(200, maxTokensBody, nil), nil
		})})
	resp, err := cl.Complete(backend.Request{System: "sys", Prompt: "prompt"})
	if err == nil {
		t.Fatal("expected truncation error, got nil")
	}
	if !strings.Contains(err.Error(), "max_tokens") {
		t.Errorf("error should mention max_tokens, got: %v", err)
	}
	// Usage must be filled before the error is returned — a truncated response was still billed.
	if resp.Usage.InputTokens != 10 || resp.Usage.OutputTokens != 1500 {
		t.Errorf("usage not recorded before truncation error: in=%d out=%d",
			resp.Usage.InputTokens, resp.Usage.OutputTokens)
	}
}

// TestWebSearchToolResultParsesRetrievedSources loads the live-capture fixture and asserts Complete
// extracts at least one Source from the web_search_tool_result block.
func TestWebSearchToolResultParsesRetrievedSources(t *testing.T) {
	body, err := os.ReadFile("../../../testdata/web_search_live.json")
	if err != nil {
		t.Skipf("live fixture not available: %v", err)
	}
	cl := New("claude-sonnet-4-6", "test-key",
		&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return fakeResp(200, string(body), nil), nil
		})})
	resp, err := cl.Complete(backend.Request{System: "sys", Prompt: "prompt", WithTools: true})
	if err != nil {
		t.Fatalf("Complete error: %v", err)
	}
	if len(resp.Sources) == 0 {
		t.Errorf("want at least 1 Source from web_search_tool_result block, got 0")
	}
}

// TestWebSearchToolResultErrorSkipped loads the error fixture (content is an object, not an array)
// and asserts Complete returns 0 Sources without panicking.
func TestWebSearchToolResultErrorSkipped(t *testing.T) {
	body, err := os.ReadFile("../../../testdata/web_search_error.json")
	if err != nil {
		t.Fatalf("error fixture missing: %v", err)
	}
	cl := New("claude-sonnet-4-6", "test-key",
		&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return fakeResp(200, string(body), nil), nil
		})})
	resp, err := cl.Complete(backend.Request{System: "sys", Prompt: "prompt", WithTools: true})
	if err != nil {
		t.Fatalf("Complete error: %v", err)
	}
	if len(resp.Sources) != 0 {
		t.Errorf("want 0 Sources for error block, got %d", len(resp.Sources))
	}
}

// TestCompleteCacheControl verifies the request puts Cached in its own content block marked
// cache_control ephemeral, ahead of the prompt block, and caches the system prompt too.
func TestCompleteCacheControl(t *testing.T) {
	var body []byte
	cl := New("claude-sonnet-4-6", "k",
		&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			body, _ = io.ReadAll(r.Body)
			return fakeResp(200, okBody, nil), nil
		})})
	if _, err := cl.Complete(backend.Request{System: "SYSTEM", Prompt: "CLAIM:\nx", Cached: "SOURCE:\nbig corpus"}); err != nil {
		t.Fatalf("Complete: %v", err)
	}
	var req apiReq
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatalf("unmarshal req: %v", err)
	}
	if len(req.System) != 1 || req.System[0].CacheControl == nil {
		t.Errorf("system prompt not cache-marked: %+v", req.System)
	}
	if len(req.Messages) != 1 || len(req.Messages[0].Content) != 2 {
		t.Fatalf("want 2 content blocks, got %+v", req.Messages)
	}
	blocks := req.Messages[0].Content
	if blocks[0].CacheControl == nil || !strings.Contains(blocks[0].Text, "big corpus") {
		t.Errorf("first block should be the cached source: %+v", blocks[0])
	}
	if blocks[1].CacheControl != nil || !strings.Contains(blocks[1].Text, "CLAIM") {
		t.Errorf("second block should be the uncached claim: %+v", blocks[1])
	}
}

// TestCompleteNoCacheSingleBlock confirms that with no Cached the user message is a single uncached
// block while the system prompt is still cached.
func TestCompleteNoCacheSingleBlock(t *testing.T) {
	var body []byte
	cl := New("claude-sonnet-4-6", "k",
		&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			body, _ = io.ReadAll(r.Body)
			return fakeResp(200, okBody, nil), nil
		})})
	if _, err := cl.Complete(backend.Request{System: "SYSTEM", Prompt: "CLAIM:\nx"}); err != nil {
		t.Fatalf("Complete: %v", err)
	}
	var req apiReq
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatalf("unmarshal req: %v", err)
	}
	if len(req.Messages[0].Content) != 1 || req.Messages[0].Content[0].CacheControl != nil {
		t.Errorf("want one uncached block, got %+v", req.Messages[0].Content)
	}
}

// ── test helpers ───────────────────────────────────────────────────────────────

// roundTripFunc adapts a function to the http.RoundTripper interface.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// fakeResp builds a minimal *http.Response with the given status, body, and headers.
func fakeResp(status int, body string, headers map[string]string) *http.Response {
	h := http.Header{}
	for k, v := range headers {
		h.Set(k, v)
	}
	return &http.Response{StatusCode: status, Header: h, Body: io.NopCloser(strings.NewReader(body))}
}

// okBody is a representative Anthropic API success response (real field layout).
const okBody = `{"id":"msg_01","type":"message","role":"assistant",` +
	`"content":[{"type":"text","text":"hello"}],` +
	`"model":"claude-sonnet-4-6","stop_reason":"end_turn","stop_sequence":null,` +
	`"usage":{"input_tokens":10,"output_tokens":5,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}`

// maxTokensBody is a response that was cut off mid-output.
const maxTokensBody = `{"id":"msg_02","type":"message","role":"assistant",` +
	`"content":[{"type":"text","text":"partial {"}],` +
	`"model":"claude-sonnet-4-6","stop_reason":"max_tokens","stop_sequence":null,` +
	`"usage":{"input_tokens":10,"output_tokens":1500,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}`
