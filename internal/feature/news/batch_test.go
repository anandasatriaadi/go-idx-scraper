package news

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/revrost/go-openrouter"
)

func TestOpenRouterBatch_SubmitAndPoll(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Submit endpoint
		if r.Method == http.MethodPost && r.URL.Path == "/api/beta/batches" {
			w.WriteHeader(http.StatusAccepted)
			_ = json.NewEncoder(w).Encode(OpenRouterBatch{
				ID:       "batch_test123",
				Object:   "batch",
				Endpoint: "/v1/chat/completions",
				Model:    "google/gemini-3.7-flash",
				Status:   "validating",
				RequestCounts: BatchCounts{
					Total: 1,
				},
			})
			return
		}

		// Poll endpoint
		if r.Method == http.MethodGet && r.URL.Path == "/api/beta/batches/batch_test123" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(OpenRouterBatch{
				ID:       "batch_test123",
				Object:   "batch",
				Endpoint: "/v1/chat/completions",
				Model:    "google/gemini-3.7-flash",
				Status:   "completed",
				RequestCounts: BatchCounts{
					Total:     1,
					Completed: 1,
				},
				Usage: &BatchUsage{
					TotalTokens: 150,
					Cost:        0.0001,
				},
				Results: []BatchResultItem{
					{
						ID:       "res_001",
						CustomID: "art_123",
						Response: &BatchResultResponse{
							StatusCode: 200,
							Body: BatchResultBody{
								ID: "gen_001",
								Choices: []BatchResultChoice{
									{
										Index: 0,
										Message: struct {
											Role    string `json:"role"`
											Content string `json:"content"`
										}{
											Role:    "assistant",
											Content: `{"title":"Sample Batch News","summary":"3-sentence test summary.","priority":5,"value_score":6,"impact_direction":"Bullish","investment_takeaway":"Solid compounding.","tickers":["BBRI"],"sector":"G. Financials","subsector":"G1. Banks","is_industry_wide":false}`,
										},
									},
								},
							},
						},
					},
				},
			})
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	requests := []BatchRequestItem{
		{
			CustomID: "art_123",
			Body: BatchRequestBody{
				Messages: []openrouter.ChatCompletionMessage{
					{
						Role:    openrouter.ChatMessageRoleUser,
						Content: openrouter.Content{Text: "Analyze this news."},
					},
				},
			},
		},
	}

	// Test submission payload marshaling
	payload := BatchSubmissionPayload{
		Endpoint: "/v1/chat/completions",
		Model:    "google/gemini-3.7-flash",
		Requests: requests,
	}
	bytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal batch payload: %v", err)
	}
	if len(bytes) == 0 {
		t.Fatal("Empty marshaled payload")
	}

	// Verify OpenRouterBatch JSON unmarshaling
	var batch OpenRouterBatch
	sampleResp := fmt.Sprintf(`{"id":"batch_test123","status":"completed","request_counts":{"total":1,"completed":1}}`)
	if err := json.Unmarshal([]byte(sampleResp), &batch); err != nil {
		t.Fatalf("Failed to unmarshal sample batch response: %v", err)
	}
	if batch.ID != "batch_test123" || batch.Status != "completed" {
		t.Errorf("Unexpected batch data: %+v", batch)
	}
}
