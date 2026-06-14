package client

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// DoerFunc adapts a plain function to the Doer interface.
type DoerFunc func(*http.Request) (*http.Response, error)

// Do implements Doer.
func (f DoerFunc) Do(r *http.Request) (*http.Response, error) { return f(r) }

// Stub returns a Doer that answers successive calls with the given assistant
// texts, each wrapped as a 200 Messages API response (stop_reason "end_turn").
// It is the only network fake in the codebase; upstream packages' tests use it
// rather than stubbing the network themselves. Calls past the last text reuse
// the last one. Not safe for concurrent use — a run grades sequentially.
func Stub(texts ...string) Doer {
	i := 0
	return DoerFunc(func(*http.Request) (*http.Response, error) {
		text := ""
		if len(texts) > 0 {
			if i < len(texts) {
				text = texts[i]
			} else {
				text = texts[len(texts)-1]
			}
		}
		i++
		body, _ := json.Marshal(apiResponse{
			Content:    []apiContent{{Type: "text", Text: text}},
			StopReason: "end_turn",
			Usage:      apiUsage{InputTokens: 1, OutputTokens: 1},
		})
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	})
}
