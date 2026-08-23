package news

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/config"
	"github.com/revrost/go-openrouter"
	"github.com/revrost/go-openrouter/jsonschema"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"
)

type Service struct {
	repo   Repository
	logger *zap.Logger
	cfg    *config.Config
}

func NewService(repo Repository, logger *zap.Logger, cfg *config.Config) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
		cfg:    cfg,
	}
}

func (s *Service) Create(ctx context.Context, n *News) error {
	return s.repo.Create(ctx, n)
}

func (s *Service) FindAll(ctx context.Context, filter any) ([]*News, error) {
	return s.repo.FindAll(ctx, filter)
}

func (s *Service) FindByID(ctx context.Context, id bson.ObjectID) (*News, error) {
	return s.repo.FindByID(ctx, id)
}

type NewsSummary struct {
	Title              string `json:"title" jsonschema:"description=An updated engaging and objective title capturing the essence of the article"`
	Summary            string `json:"summary" jsonschema:"description=Concise 3-sentence summary highlighting financial facts figures and immediate market implications"`
	Priority           int    `json:"priority" jsonschema:"description=Market impact priority from 1 (highest market urgency/impact) to 10 (routine/low market impact)"`
	ValueScore         int    `json:"value_score" jsonschema:"description=Fundamental value investing impact score strictly between -10 and +10 (-10 is severe fundamental impairment, 0 is neutral/macro noise, +10 is massive fundamental value creation)"`
	ImpactDirection    string `json:"impact_direction" jsonschema:"enum=Bullish,enum=Bearish,enum=Neutral,description=Directional impact on underlying business intrinsic value"`
	InvestmentTakeaway string `json:"investment_takeaway" jsonschema:"description=1-2 sentence actionable takeaway strictly from the perspective of a disciplined long-term value investor focusing on moat, capital allocation, cash flows, and intrinsic value"`
}

func (s *Service) Summarize(ctx context.Context, ids []bson.ObjectID) error {
	s.logger.Info("Starting news summarization", zap.Int("num_ids", len(ids)))
	if s.cfg == nil {
		return fmt.Errorf("config not loaded")
	}
	client := openrouter.NewClient(s.cfg.OpenrouterApiKey)

	schema, err := jsonschema.GenerateSchemaForType(NewsSummary{})
	if err != nil {
		return fmt.Errorf("GenerateSchemaForType error: %v", err)
	}

	for _, id := range ids {
		s.logger.Info("Processing news ID", zap.String("id", id.Hex()))
		n, err := s.repo.FindByID(ctx, id)
		if err != nil {
			s.logger.Error("Error finding news by ID", zap.String("id", id.Hex()), zap.Error(err))
			continue
		}

		prompt := fmt.Sprintf(`You are an expert Personal Investment Manager and seasoned Value Investor adhering strictly to the fundamental principles of Benjamin Graham, Warren Buffett, and Charlie Munger.

Analyze the provided financial news article with super objective, disciplined rigor. Evaluate whether this event impacts intrinsic business value, economic moats, capital allocation, balance sheet durability, or free cash flow generation.

Provide your evaluation adhering to the following rules:
- Title: Clear, concise, and professional headline.
- Summary: Exactly 3 sentences. Cover core facts, financial metrics, and operational impact.
- Priority: Integer from 1 to 10 (1 = critical high-impact market event, 10 = routine noise).
- ValueScore: Integer from -10 to +10:
  * -10 to -1: Fundamental destruction (deteriorating moat, dilutive acquisitions, high debt risk, governance red flags).
  * 0: Neutral / Macro noise / Speculative price movements with no underlying business value change.
  * +1 to +10: Fundamental enhancement (widening moat, high ROIC reinvestment, robust organic growth, disciplined capital allocation).
- ImpactDirection: Exactly "Bullish", "Bearish", or "Neutral".
- InvestmentTakeaway: 1 to 2 sentences summarizing the bottom-line takeaway for a long-term value investor.

Article:
"""
%s
"""`, n.Content)

		request := openrouter.ChatCompletionRequest{
			Model: "google/gemini-2.5-flash",
			Messages: []openrouter.ChatCompletionMessage{
				{
					Role:    openrouter.ChatMessageRoleUser,
					Content: openrouter.Content{Text: prompt},
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
			s.logger.Error("Error creating chat completion for ID", zap.String("id", id.Hex()), zap.Error(err))
			continue
		}

		var summary NewsSummary
		if len(res.Choices) > 0 {
			if err := json.Unmarshal([]byte(res.Choices[0].Message.Content.Text), &summary); err != nil {
				s.logger.Error("Error unmarshaling response for ID", zap.String("id", id.Hex()), zap.Error(err))
				continue
			}

			s.logger.Info("Successfully summarized news",
				zap.String("id", id.Hex()),
				zap.String("title", summary.Title),
				zap.Int("priority", summary.Priority),
				zap.Int("value_score", summary.ValueScore),
				zap.String("impact_direction", summary.ImpactDirection),
			)

			// Update the news document
			update := map[string]any{
				"$set": map[string]any{
					"title":               summary.Title,
					"summary":             summary.Summary,
					"priority":            summary.Priority,
					"value_score":         summary.ValueScore,
					"impact_direction":    summary.ImpactDirection,
					"investment_takeaway": summary.InvestmentTakeaway,
					"updated_at":          time.Now(),
				},
			}
			err = s.repo.UpdateByID(ctx, id, update)
			if err != nil {
				s.logger.Error("Error updating news for ID", zap.String("id", id.Hex()), zap.Error(err))
			} else {
				s.logger.Info("Successfully updated news", zap.String("id", id.Hex()))
			}
		}
	}
	s.logger.Info("Completed news summarization")
	return nil
}
