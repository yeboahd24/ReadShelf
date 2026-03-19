package nomic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

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
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

type hfRequest struct {
	Inputs string `json:"inputs"`
}

func (e *embedder) Embed(ctx context.Context, text string) (pgvector.Vector, error) {
	var lastErr error

	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			// Longer backoff for cold starts: 5s, 10s
			backoff := time.Duration(attempt) * 5 * time.Second
			log.Printf("embed: retry %d after %s", attempt, backoff)
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return pgvector.Vector{}, ctx.Err()
			}
		}

		vec, err := e.doEmbed(ctx, text)
		if err == nil {
			return vec, nil
		}
		lastErr = err

		// Only retry on transient errors (502, 503, 504, timeouts).
		if !isRetryable(err) {
			return pgvector.Vector{}, err
		}
	}

	return pgvector.Vector{}, fmt.Errorf("embed failed after 3 attempts: %w", lastErr)
}

func (e *embedder) doEmbed(ctx context.Context, text string) (pgvector.Vector, error) {
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
		return pgvector.Vector{}, &apiError{
			statusCode: resp.StatusCode,
			message:    string(respBody),
		}
	}

	var embedding []float32
	if err := json.NewDecoder(resp.Body).Decode(&embedding); err != nil {
		return pgvector.Vector{}, fmt.Errorf("decode embedding: %w", err)
	}

	return pgvector.NewVector(embedding), nil
}

type apiError struct {
	statusCode int
	message    string
}

func (e *apiError) Error() string {
	return fmt.Sprintf("huggingface API error (%d): %s", e.statusCode, e.message)
}

func isRetryable(err error) bool {
	if ae, ok := err.(*apiError); ok {
		switch ae.statusCode {
		case 502, 503, 504, 429:
			return true
		}
	}
	// Treat timeouts and network errors as retryable.
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "context deadline exceeded") ||
			strings.Contains(errMsg, "Client.Timeout") ||
			strings.Contains(errMsg, "connection refused") {
			return true
		}
	}
	return false
}
