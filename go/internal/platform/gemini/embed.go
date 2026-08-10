package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// EmbeddingDims is the fixed dimension for text-embedding-004 / embedding-001.
const EmbeddingDims = 768

const defaultEmbeddingModel = "text-embedding-004"

// Embedder produces dense vectors for RAG indexing / search.
type Embedder interface {
	EmbedTexts(ctx context.Context, texts []string) ([][]float32, error)
}

func (c *Client) effectiveEmbeddingModel() string {
	if c != nil && strings.TrimSpace(c.EmbeddingModel) != "" {
		return strings.TrimSpace(c.EmbeddingModel)
	}
	return defaultEmbeddingModel
}

// EmbedTexts embeds each string via Gemini embedContent (one request per text).
// Empty inputs yield a zero vector of EmbeddingDims to keep ordinals stable.
func (c *Client) EmbedTexts(ctx context.Context, texts []string) ([][]float32, error) {
	if !c.Configured() {
		return nil, fmt.Errorf("gemini_not_configured")
	}
	out := make([][]float32, len(texts))
	model := c.effectiveEmbeddingModel()
	for i, text := range texts {
		if strings.TrimSpace(text) == "" {
			out[i] = make([]float32, EmbeddingDims)
			continue
		}
		vec, err := c.embedOne(ctx, model, text)
		if err != nil {
			return nil, err
		}
		if len(vec) != EmbeddingDims {
			return nil, fmt.Errorf("gemini_embedding_dims: got %d want %d", len(vec), EmbeddingDims)
		}
		out[i] = vec
	}
	return out, nil
}

func (c *Client) embedOne(ctx context.Context, model, text string) ([]float32, error) {
	body := map[string]any{
		"content": map[string]any{
			"parts": []map[string]string{{"text": text}},
		},
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:embedContent?key=%s",
		model, c.APIKey,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini_request: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("gemini_http_%d: %s", resp.StatusCode, truncate(string(respBody), 300))
	}
	var parsed struct {
		Embedding struct {
			Values []float64 `json:"values"`
		} `json:"embedding"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("gemini_embed_parse: %w", err)
	}
	if len(parsed.Embedding.Values) == 0 {
		return nil, fmt.Errorf("gemini_embed_empty")
	}
	out := make([]float32, len(parsed.Embedding.Values))
	for i, v := range parsed.Embedding.Values {
		out[i] = float32(v)
	}
	return out, nil
}

// ExtractPlainTextFromPDF asks Gemini to return the plain text of a PDF (for RAG chunking).
func (c *Client) ExtractPlainTextFromPDF(ctx context.Context, data []byte) (string, error) {
	system := `Tu extrais le texte intégral d'un document PDF vétérinaire (guides, posologies, fiches techniques).
Réponds UNIQUEMENT avec le texte brut, sans markdown ni préambule. Conserve titres et listes en texte simple.`
	return c.GenerateTextWithMedia(ctx, system, "Extrais tout le texte de ce PDF.", "application/pdf", data, 0.1)
}
