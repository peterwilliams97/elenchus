package ollama

// ollama_test pins the /api/chat request shaping and response parsing: think defaults off and the
// flag turns it on, Cached folds into the user message, usage maps from prompt_eval_count/eval_count,
// and a done_reason of "length" surfaces as an error. All through a stubbed RoundTripper, no server.

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"assay/internal/backend"
)

// TestThinkDefaultFalse confirms a client built with think=false sends "think":false and caps output
// at backend.MaxTokens.
func TestThinkDefaultFalse(t *testing.T) {
	req := captureRequest(t, false, backend.Request{System: "SYS", Prompt: "Q"})
	if req.Think {
		t.Errorf("want think=false by default, got true")
	}
	if req.Options.NumPredict != backend.MaxTokens {
		t.Errorf("num_predict: want %d, got %d", backend.MaxTokens, req.Options.NumPredict)
	}
	if req.Stream {
		t.Errorf("want stream=false")
	}
}

// TestThinkTrue confirms the -think flag reaches the request as "think":true.
func TestThinkTrue(t *testing.T) {
	req := captureRequest(t, true, backend.Request{Prompt: "Q"})
	if !req.Think {
		t.Errorf("want think=true when flag set, got false")
	}
}

// TestSystemAndCachedFold confirms System becomes a system message and Cached is prepended to the
// user message (Ollama has no cache seam, so the corpus must ride in the prompt).
func TestSystemAndCachedFold(t *testing.T) {
	req := captureRequest(t, false, backend.Request{System: "SYS", Cached: "CORPUS", Prompt: "QUESTION"})
	if len(req.Messages) != 2 || req.Messages[0].Role != "system" || req.Messages[0].Content != "SYS" {
		t.Fatalf("want [system,user], got %+v", req.Messages)
	}
	user := req.Messages[1].Content
	if !strings.Contains(user, "CORPUS") || !strings.Contains(user, "QUESTION") {
		t.Errorf("user message should carry both corpus and question: %q", user)
	}
}

// TestSchemaAndKeepAliveAndCtx confirms a Request.Schema reaches the request as `format`, keep_alive
// holds the model resident, num_ctx is sized to the prompt (not the default), and Temperature is
// forwarded into options.
func TestSchemaAndKeepAliveAndCtx(t *testing.T) {
	temp := 0.7
	schema := json.RawMessage(`{"type":"object","properties":{"verdict":{"type":"string"}}}`)
	long := strings.Repeat("passage text ", 2000) // ~26KB → sizes num_ctx above the 4096 floor
	req := captureRequest(t, false, backend.Request{
		System: "SYS", Cached: long, Prompt: "CLAIM", Schema: schema, Temperature: &temp,
	})
	if len(req.Format) == 0 || !strings.Contains(string(req.Format), "verdict") {
		t.Errorf("schema not sent as format: %q", string(req.Format))
	}
	if req.KeepAlive != "30m" {
		t.Errorf("keep_alive: want 30m, got %q", req.KeepAlive)
	}
	if req.Options.Temperature != 0.7 {
		t.Errorf("temperature not forwarded: %v", req.Options.Temperature)
	}
	if req.Options.NumCtx <= 4096 {
		t.Errorf("num_ctx should be sized above the floor for a long prompt, got %d", req.Options.NumCtx)
	}
}

// TestNumCtxFloor confirms a short prompt still gets at least the 4096 floor.
func TestNumCtxFloor(t *testing.T) {
	req := captureRequest(t, false, backend.Request{Prompt: "hi"})
	if req.Options.NumCtx != 4096 {
		t.Errorf("short prompt num_ctx: want 4096 floor, got %d", req.Options.NumCtx)
	}
}

// TestParsesUsageAndText maps prompt_eval_count/eval_count to input/output tokens and returns the
// assistant content; cache and web-search figures stay zero.
func TestParsesUsageAndText(t *testing.T) {
	cl := New("m", "", false, stub(`{"message":{"role":"assistant","content":"hi there"},`+
		`"prompt_eval_count":7,"eval_count":3,"done":true,"done_reason":"stop"}`))
	resp, err := cl.Complete(backend.Request{Prompt: "Q"})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if resp.Text != "hi there" {
		t.Errorf("text: got %q", resp.Text)
	}
	if resp.Usage.InputTokens != 7 || resp.Usage.OutputTokens != 3 {
		t.Errorf("usage: in=%d out=%d", resp.Usage.InputTokens, resp.Usage.OutputTokens)
	}
	if resp.Usage.CacheRead != 0 || resp.Usage.WebSearches != 0 {
		t.Errorf("cache/web should be zero on ollama: %+v", resp.Usage)
	}
}

// TestLengthTruncationErrors confirms done_reason=length is surfaced as an error, with the partial
// text and usage still returned so the caller can bill it.
func TestLengthTruncationErrors(t *testing.T) {
	cl := New("m", "", false, stub(`{"message":{"content":"partial"},`+
		`"prompt_eval_count":5,"eval_count":1500,"done":true,"done_reason":"length"}`))
	resp, err := cl.Complete(backend.Request{Prompt: "Q"})
	if err == nil || !strings.Contains(err.Error(), "length") {
		t.Fatalf("want length-truncation error, got %v", err)
	}
	if resp.Usage.OutputTokens != 1500 {
		t.Errorf("usage should be billed on truncation, got out=%d", resp.Usage.OutputTokens)
	}
}

func TestSupportsWebSearchFalse(t *testing.T) {
	if New("m", "", false, nil).SupportsWebSearch() {
		t.Errorf("ollama has no web search; SupportsWebSearch must be false")
	}
}

// ── test helpers ───────────────────────────────────────────────────────────────

// captureRequest runs one Complete against a stub that records the outgoing /api/chat body, and
// returns it decoded. The stub replies with a minimal successful chat response.
func captureRequest(t *testing.T, think bool, req backend.Request) chatReq {
	t.Helper()
	var got chatReq
	httpc := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("bad request body: %v", err)
		}
		return okResp(), nil
	})}
	if _, err := New("m", "", think, httpc).Complete(req); err != nil {
		t.Fatalf("Complete: %v", err)
	}
	return got
}

// stub returns an http.Client that replies to every request with the given JSON body and 200.
func stub(body string) *http.Client {
	return &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}, nil
	})}
}

func okResp() *http.Response {
	return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(
		`{"message":{"role":"assistant","content":"ok"},"prompt_eval_count":1,"eval_count":1,"done":true,"done_reason":"stop"}`))}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
