package nomic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/dominic/readshelf/internal/core/port/outbound"
	"github.com/pgvector/pgvector-go"
)

const defaultModel = "BAAI/bge-base-en-v1.5"

type embedder struct {
	apiKey string
	model  string
	client *http.Client
}

func NewEmbedder(apiKey string) outbound.Embedder {
	return NewEmbedderWithModel(apiKey, defaultModel)
}

func NewEmbedderWithModel(apiKey, model string) outbound.Embedder {
	return &embedder{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{},
	}
}

type hfRequest struct {
	Inputs string `json:"inputs"`
}

func (e *embedder) Embed(ctx context.Context, text string) (pgvector.Vector, error) {
	body, err := json.Marshal(hfRequest{Inputs: text})
	if err != nil {
		return pgvector.Vector{}, err
	}

	url := fmt.Sprintf("https://router.huggingface.co/hf-inference/models/%s", e.model)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return pgvector.Vector{}, err
	}
	req.Header.Set("Authorization", "Bearer "+e.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return pgvector.Vector{}, fmt.Errorf("huggingface request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return pgvector.Vector{}, fmt.Errorf("huggingface API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var embedding []float32
	if err := json.NewDecoder(resp.Body).Decode(&embedding); err != nil {
		return pgvector.Vector{}, fmt.Errorf("decode embedding: %w", err)
	}

	return pgvector.NewVector(embedding), nil
}
