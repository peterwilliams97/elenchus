package embed

// embed.go is the local embedding client: a batched POST to a running Ollama server's /api/embed,
// used only to build the second (semantic) ranker behind faithfulness retrieval. It is deliberately
// separate from the judge backend (internal/backend/ollama) — embeddings always come from a local
// ollama with nomic-embed-text, whatever model the judge itself runs on. The vectors it returns are
// L2-normalised so a downstream dot product is a cosine; retrieval never has to normalise again.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"
)

// DefaultModel is the embedding model. nomic-embed-text is 768-dim and pulls in well under a second
// per batch on this hardware; it is the one the retrieval cache is keyed on.
const DefaultModel = "nomic-embed-text"

// DefaultBaseURL is the local Ollama server address, used when New is given an empty baseURL. It
// mirrors internal/backend/ollama.DefaultBaseURL — kept here so this package carries no ollama import.
const DefaultBaseURL = "http://localhost:11434"

// batchSize bounds one /api/embed request. The corpus is ~1000 passages; batching keeps the request
// bodies small without a round trip per passage.
const batchSize = 64

var defaultHTTP = &http.Client{Timeout: 300 * time.Second}

// Client is a configured ollama embedding endpoint.
type Client struct {
	model   string
	baseURL string
	http    *http.Client
}

// New builds an embedding client. Empty baseURL/model fall back to the package defaults; a nil httpc
// uses the package default client.
func New(model, baseURL string, httpc *http.Client) *Client {
	if model == "" {
		model = DefaultModel
	}
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	if httpc == nil {
		httpc = defaultHTTP
	}
	return &Client{model: model, baseURL: baseURL, http: httpc}
}

// Model reports the embedding model id — the value the retrieval cache is keyed on, so a model swap
// invalidates a stale cache rather than silently mixing vector spaces.
func (c *Client) Model() string { return c.model }

// Embed returns one L2-normalised vector per input text, in input order. It batches the inputs and
// fails the whole call if any batch fails — a partial embedding set would silently degrade the
// semantic ranker, which is worse than falling back to BM25-only.
func (c *Client) Embed(texts []string) ([][]float32, error) {
	out := make([][]float32, 0, len(texts))
	for start := 0; start < len(texts); start += batchSize {
		end := min(start+batchSize, len(texts))
		vecs, err := c.embedBatch(texts[start:end])
		if err != nil {
			return nil, fmt.Errorf("embed batch [%d:%d): %w", start, end, err)
		}
		out = append(out, vecs...)
	}
	if len(out) != len(texts) {
		return nil, fmt.Errorf("embed count mismatch: got %d vectors for %d texts", len(out), len(texts))
	}
	return out, nil
}

func (c *Client) embedBatch(texts []string) ([][]float32, error) {
	body, _ := json.Marshal(embedReq{Model: c.model, Input: texts})
	httpReq, _ := http.NewRequest("POST", c.baseURL+"/api/embed", bytes.NewReader(body))
	httpReq.Header.Set("content-type", "application/json")
	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, err
	}
	raw, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	var er embedResp
	if err := json.Unmarshal(raw, &er); err != nil {
		return nil, fmt.Errorf("unreadable ollama embed response: %.200s", string(raw))
	}
	if er.Error != "" {
		return nil, fmt.Errorf("ollama embed error: %s", er.Error)
	}
	if len(er.Embeddings) != len(texts) {
		return nil, fmt.Errorf("ollama returned %d embeddings for %d inputs", len(er.Embeddings), len(texts))
	}
	for i := range er.Embeddings {
		normalize(er.Embeddings[i])
	}
	return er.Embeddings, nil
}

// normalize scales a vector to unit L2 length in place, so a later dot product is a cosine. A
// zero-length vector (all zeros) is left as-is rather than dividing by zero.
func normalize(v []float32) {
	var sum float64
	for _, x := range v {
		sum += float64(x) * float64(x)
	}
	if sum == 0 {
		return
	}
	inv := float32(1 / math.Sqrt(sum))
	for i := range v {
		v[i] *= inv
	}
}

type embedReq struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}
type embedResp struct {
	Embeddings [][]float32 `json:"embeddings"`
	Error      string      `json:"error"`
}
