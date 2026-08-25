package news

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/revrost/go-openrouter"
)

type BatchRequestBody struct {
	Messages       []openrouter.ChatCompletionMessage       `json:"messages"`
	ResponseFormat *openrouter.ChatCompletionResponseFormat `json:"response_format,omitempty"`
}

type BatchRequestItem struct {
	CustomID string           `json:"custom_id"`
	Body     BatchRequestBody `json:"body"`
}

type BatchSubmissionPayload struct {
	Endpoint string             `json:"endpoint"`
	Model    string             `json:"model"`
	Requests []BatchRequestItem `json:"requests"`
}

type BatchCounts struct {
	Total     int `json:"total"`
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
}

type BatchUsage struct {
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	Cost             float64 `json:"cost"`
	IsBYOK           bool    `json:"is_byok"`
}

type BatchResultChoice struct {
	Index   int `json:"index"`
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	FinishReason string `json:"finish_reason"`
}

type BatchResultBody struct {
	ID      string              `json:"id"`
	Choices []BatchResultChoice `json:"choices"`
}

type BatchResultResponse struct {
	StatusCode int             `json:"status_code"`
	RequestID  string          `json:"request_id"`
	Body       BatchResultBody `json:"body"`
}

type BatchResultItem struct {
	ID       string               `json:"id"`
	CustomID string               `json:"custom_id"`
	Response *BatchResultResponse `json:"response"`
	Error    any                  `json:"error"`
}

type OpenRouterBatch struct {
	ID               string            `json:"id"`
	Object           string            `json:"object"`
	Endpoint         string            `json:"endpoint"`
	Model            string            `json:"model"`
	CompletionWindow string            `json:"completion_window"`
	Status           string            `json:"status"`
	CreatedAt        int64             `json:"created_at"`
	FinalizedAt      *int64            `json:"finalized_at"`
	RequestCounts    BatchCounts       `json:"request_counts"`
	Usage            *BatchUsage       `json:"usage"`
	Results          []BatchResultItem `json:"results"`
	Error            any               `json:"error"`
}

// SubmitOpenRouterBatch submits an asynchronous batch request to OpenRouter's Batch API (50% pricing discount)
func SubmitOpenRouterBatch(ctx context.Context, apiKey string, model string, requests []BatchRequestItem) (*OpenRouterBatch, error) {
	if len(requests) == 0 {
		return nil, fmt.Errorf("no requests in batch")
	}

	payload := BatchSubmissionPayload{
		Endpoint: "/v1/chat/completions",
		Model:    model,
		Requests: requests,
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshaling batch payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://openrouter.ai/api/beta/batches", bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("creating http request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("submitting batch: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("unexpected status %d from batch API: %s", resp.StatusCode, string(respBytes))
	}

	var batch OpenRouterBatch
	if err := json.Unmarshal(respBytes, &batch); err != nil {
		return nil, fmt.Errorf("unmarshaling batch response: %w", err)
	}

	return &batch, nil
}

// GetOpenRouterBatch queries the status and results of a submitted batch
func GetOpenRouterBatch(ctx context.Context, apiKey string, batchID string) (*OpenRouterBatch, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://openrouter.ai/api/beta/batches/"+batchID, nil)
	if err != nil {
		return nil, fmt.Errorf("creating get batch request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("querying batch: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading batch response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d querying batch: %s", resp.StatusCode, string(respBytes))
	}

	var batch OpenRouterBatch
	if err := json.Unmarshal(respBytes, &batch); err != nil {
		return nil, fmt.Errorf("unmarshaling batch response: %w", err)
	}

	return &batch, nil
}
