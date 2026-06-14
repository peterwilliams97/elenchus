// SPDX-License-Identifier: Apache-2.0

package client

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

// helper: a Doer returning a fixed sequence of (status, apiResponse) pairs.
func seq(items ...*http.Response) Doer {
	i := 0
	return DoerFunc(func(*http.Request) (*http.Response, error) {
		r := items[i]
		if i < len(items)-1 {
			i++
		}
		return r, nil
	})
}

func resp(status int, ar apiResponse) *http.Response {
	body, _ := json.Marshal(ar)
	return &http.Response{StatusCode: status, Body: io.NopCloser(bytes.NewReader(body)), Header: make(http.Header)}
}

func TestCallJSONHappyPath(t *testing.T) {
	c := New(Config{APIKey: "k", Model: "m", HTTP: Stub(`{"verdict":"hollow","reason":"x"}`)})
	var out struct {
		Verdict string `json:"verdict"`
		Reason  string `json:"reason"`
	}
	if err := c.CallJSON("sys", "user", &out); err != nil {
		t.Fatalf("CallJSON: %v", err)
	}
	if out.Verdict != "hollow" || out.Reason != "x" {
		t.Errorf("parsed %+v, want hollow/x", out)
	}
	if c.Usage().Calls != 1 {
		t.Errorf("calls = %d, want 1", c.Usage().Calls)
	}
}

func TestCallJSONFenceStripped(t *testing.T) {
	c := New(Config{APIKey: "k", Model: "m", HTTP: Stub("```json\n{\"verdict\":\"partial\"}\n```")})
	var out struct {
		Verdict string `json:"verdict"`
	}
	if err := c.CallJSON("sys", "user", &out); err != nil {
		t.Fatalf("CallJSON: %v", err)
	}
	if out.Verdict != "partial" {
		t.Errorf("verdict = %q, want partial", out.Verdict)
	}
}

func TestCallJSONValueWithCommaBrace(t *testing.T) {
	// A value containing ",}" must survive — the old trailing-comma regex ran
	// inside quoted strings and corrupted it.
	c := New(Config{APIKey: "k", Model: "m", HTTP: Stub(`{"reason":"holds at 60%,} per the data"}`)})
	var out struct {
		Reason string `json:"reason"`
	}
	if err := c.CallJSON("sys", "user", &out); err != nil {
		t.Fatalf("CallJSON: %v", err)
	}
	if out.Reason != "holds at 60%,} per the data" {
		t.Errorf("reason corrupted: %q", out.Reason)
	}
}

func TestCallJSONStructuralTrailingComma(t *testing.T) {
	// A real trailing comma before } is stripped and parses on the first call —
	// no retry needed.
	c := New(Config{APIKey: "k", Model: "m", HTTP: Stub(`{"verdict":"hollow",}`)})
	var out struct {
		Verdict string `json:"verdict"`
	}
	if err := c.CallJSON("sys", "user", &out); err != nil {
		t.Fatalf("CallJSON: %v", err)
	}
	if out.Verdict != "hollow" {
		t.Errorf("verdict = %q, want hollow", out.Verdict)
	}
	if c.Usage().Calls != 1 {
		t.Errorf("calls = %d, want 1 (stripped inline, no retry)", c.Usage().Calls)
	}
}

func TestCallJSONParseRetry(t *testing.T) {
	// First response is not JSON; the loose parse fails and CallJSON retries once
	// with an augmented system prompt; the second response parses.
	c := New(Config{APIKey: "k", Model: "m", HTTP: Stub("I cannot comply.", `{"verdict":"substantive"}`)})
	var out struct {
		Verdict string `json:"verdict"`
	}
	if err := c.CallJSON("sys", "user", &out); err != nil {
		t.Fatalf("CallJSON: %v", err)
	}
	if out.Verdict != "substantive" {
		t.Errorf("verdict = %q, want substantive", out.Verdict)
	}
	if c.Usage().Calls != 2 {
		t.Errorf("calls = %d, want 2 (one parse-failure + one retry)", c.Usage().Calls)
	}
}

func TestCallJSONParseRetryExhausted(t *testing.T) {
	c := New(Config{APIKey: "k", Model: "m", HTTP: Stub("nope", "still not json")})
	var out struct {
		V string `json:"v"`
	}
	if err := c.CallJSON("sys", "user", &out); err == nil {
		t.Fatal("expected error after the single retry also fails to parse")
	}
}

func TestTruncationIsError(t *testing.T) {
	c := New(Config{APIKey: "k", Model: "m", HTTP: seq(
		resp(200, apiResponse{
			Content:    []apiContent{{Type: "text", Text: `{"verdict":"hollow"}`}},
			StopReason: "max_tokens",
			Usage:      apiUsage{InputTokens: 10, OutputTokens: 1500},
		}),
	)})
	var out struct {
		V string `json:"verdict"`
	}
	err := c.CallJSON("sys", "user", &out)
	if err == nil || !strings.Contains(err.Error(), "truncated") {
		t.Fatalf("want truncation error, got %v", err)
	}
	// Usage is billed even on truncation (spec: accumulate before the stop_reason check).
	if c.Usage().OutputTokens != 1500 {
		t.Errorf("output tokens = %d, want 1500 (billed despite truncation)", c.Usage().OutputTokens)
	}
}

func TestHTTPRetryOn429(t *testing.T) {
	good := apiResponse{Content: []apiContent{{Type: "text", Text: `{"verdict":"partial"}`}}, StopReason: "end_turn"}
	c := New(Config{APIKey: "k", Model: "m", HTTP: seq(
		resp(429, apiResponse{}),
		resp(200, good),
	)})
	c.retryBase = 0 // collapse backoff to zero in tests
	var out struct {
		V string `json:"verdict"`
	}
	if err := c.CallJSON("sys", "user", &out); err != nil {
		t.Fatalf("CallJSON after a 429 retry: %v", err)
	}
	if out.V != "partial" {
		t.Errorf("verdict = %q, want partial", out.V)
	}
}

func TestNonRetryableStatusIsError(t *testing.T) {
	c := New(Config{APIKey: "k", Model: "m", HTTP: seq(resp(400, apiResponse{}))})
	c.retryBase = 0
	var out struct {
		V string `json:"v"`
	}
	if err := c.CallJSON("sys", "user", &out); err == nil {
		t.Fatal("expected error on a non-retryable 400")
	}
}

func TestRequestShape(t *testing.T) {
	// The fake inspects the outbound request: correct URL, headers, and body.
	var got *http.Request
	var gotBody []byte
	doer := DoerFunc(func(r *http.Request) (*http.Response, error) {
		got = r
		gotBody, _ = io.ReadAll(r.Body)
		return resp(200, apiResponse{Content: []apiContent{{Type: "text", Text: `{}`}}, StopReason: "end_turn"}), nil
	})
	c := New(Config{APIKey: "secret", Model: "claude-sonnet-4-6", HTTP: doer})
	var out struct{}
	if err := c.CallJSON("SYS", "USER", &out); err != nil {
		t.Fatalf("CallJSON: %v", err)
	}
	if got.Method != http.MethodPost {
		t.Errorf("method = %s, want POST", got.Method)
	}
	if got.Header.Get("x-api-key") != "secret" {
		t.Errorf("x-api-key = %q", got.Header.Get("x-api-key"))
	}
	if got.Header.Get("anthropic-version") == "" {
		t.Error("missing anthropic-version header")
	}
	var req apiRequest
	if err := json.Unmarshal(gotBody, &req); err != nil {
		t.Fatalf("request body not JSON: %v", err)
	}
	if req.Model != "claude-sonnet-4-6" || req.System != "SYS" || req.MaxTokens != maxTokens {
		t.Errorf("request = %+v", req)
	}
	if len(req.Messages) != 1 || req.Messages[0].Role != "user" || req.Messages[0].Content != "USER" {
		t.Errorf("messages = %+v", req.Messages)
	}
}
