package helper

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"github.com/anandasatriaadi/go-idx-scraper/internal/db/model"
	"github.com/revrost/go-openrouter"
	"github.com/revrost/go-openrouter/jsonschema"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"
)

type NewsSummary struct {
	Title    string `json:"title"`
	Summary  string `json:"summary"`
	Priority int    `json:"priority"`
}

// SummarizeNews takes a list of ObjectIDs, queries the news, summarizes using AI, and updates Title, Summary, and Priority
func SummarizeNews(ctx context.Context, logger *zap.Logger, ids []bson.ObjectID, repo *model.NewsRepository) error {
	logger.Info("Starting news summarization", zap.Int("num_ids", len(ids)))
	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("config not loaded")
	}
	client := openrouter.NewClient(cfg.OpenrouterApiKey)

	schema, err := jsonschema.GenerateSchemaForType(NewsSummary{})
	if err != nil {
		return fmt.Errorf("GenerateSchemaForType error: %v", err)
	}

	for _, id := range ids {
		logger.Info("Processing news ID", zap.String("id", id.Hex()))
		news, err := repo.FindByID(ctx, id)
		if err != nil {
			logger.Error("Error finding news by ID", zap.String("id", id.Hex()), zap.Error(err))
			continue
		}

		request := openrouter.ChatCompletionRequest{
			Model: "google/gemini-2.5-flash-lite",
			Messages: []openrouter.ChatCompletionMessage{
				{
					Role:    openrouter.ChatMessageRoleUser,
					Content: openrouter.Content{Text: fmt.Sprintf("You are a seasoned investor constantly monitoring news for stories that could impact financial markets. Your role is to analyze news articles and provide the following:\n\n- An updated, engaging title that captures the essence of the article.\n- A concise summary in exactly 3 sentences, highlighting key points, implications for the market, and any relevant data. The summary must also reflect whether the news is worth the time to read for scoping possibilities (i.e., beneficial for identifying potential investment opportunities or not).\n- A priority score from 1 to 10, where 1 indicates the highest potential impact on markets and 10 the lowest. Base this on factors like the scale of the event, affected sectors, and immediacy.\n\nSummarize the provided news article accordingly. Here is the news you need to summarize: ```%s```", news.Content)},
				},
			},
			ResponseFormat: &openrouter.ChatCompletionResponseFormat{
				Type: openrouter.ChatCompletionResponseFormatTypeJSONSchema,
				JSONSchema: &openrouter.ChatCompletionResponseFormatJSONSchema{
					Name:   "news_summary",
					Schema: schema,
					Strict: true,
				},
			},
		}

		res, err := client.CreateChatCompletion(ctx, request)
		if err != nil {
			logger.Error("Error creating chat completion for ID", zap.String("id", id.Hex()), zap.Error(err))
			continue
		}

		var summary NewsSummary
		if err := json.Unmarshal([]byte(res.Choices[0].Message.Content.Text), &summary); err != nil {
			logger.Error("Error unmarshaling response for ID", zap.String("id", id.Hex()), zap.Error(err))
			continue
		}
		logger.Info("Successfully summarized news", zap.String("id", id.Hex()), zap.String("title", summary.Title), zap.Int("priority", summary.Priority))

		// Update the news document
		update := map[string]interface{}{
			"$set": map[string]interface{}{
				"title":    summary.Title,
				"summary":  summary.Summary,
				"priority": summary.Priority,
			},
		}
		_, err = repo.UpdateByID(ctx, id, update)
		if err != nil {
			logger.Error("Error updating news for ID", zap.String("id", id.Hex()), zap.Error(err))
		} else {
			logger.Info("Successfully updated news", zap.String("id", id.Hex()))
		}
	}
	logger.Info("Completed news summarization")

	return nil
}
