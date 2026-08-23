package news

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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
	Title              string   `json:"title" jsonschema:"description=An updated engaging and objective title capturing the essence of the article"`
	Summary            string   `json:"summary" jsonschema:"description=Concise 3-sentence summary highlighting financial facts figures and immediate market implications"`
	Priority           int      `json:"priority" jsonschema:"description=Market impact priority from 1 (highest market urgency/impact) to 10 (routine/low market impact)"`
	ValueScore         int      `json:"value_score" jsonschema:"description=Fundamental value investing impact score strictly between -10 and +10 (-10 is severe fundamental impairment, 0 is neutral/macro noise, +10 is massive fundamental value creation)"`
	ImpactDirection    string   `json:"impact_direction" jsonschema:"enum=Bullish,enum=Bearish,enum=Neutral,description=Directional impact on underlying business intrinsic value"`
	InvestmentTakeaway string   `json:"investment_takeaway" jsonschema:"description=1-2 sentence actionable takeaway strictly from the perspective of a disciplined long-term value investor focusing on moat, capital allocation, cash flows, and intrinsic value"`
	Tickers            []string `json:"tickers" jsonschema:"description=List of 4-letter IDX stock ticker symbols explicitly mentioned or directly affected (e.g. ['BBRI', 'BBCA']). Empty array if none."`
	Industry           string   `json:"industry" jsonschema:"description=Primary industry or sector classification (e.g. Banking, Poultry, Mining, Energy, Consumer Goods, Infrastructure, Technology, Macroeconomics)"`
	IsIndustryWide     bool     `json:"is_industry_wide" jsonschema:"description=True if the news affects an entire industry sector or macroeconomic policy rather than just one individual company"`
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
- Tickers: Array of uppercase 4-letter Indonesian stock tickers explicitly mentioned or impacted (e.g. ["BBRI", "BBCA"]). Empty array [] if no specific company is mentioned.
- Industry: Sector category (e.g., "Banking", "Poultry", "Mining", "Energy", "Consumer Goods", "Technology", "Infrastructure", "Macroeconomics").
- IsIndustryWide: Boolean true if the news affects the whole sector or macro economy rather than an isolated company.

Article:
"""
%s
"""`, n.Content)

		request := openrouter.ChatCompletionRequest{
			Model: "google/gemini-3.7-flash",
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
				zap.Strings("tickers", summary.Tickers),
				zap.String("industry", summary.Industry),
			)

			// Update the news document
			update := bson.M{
				"$set": bson.M{
					"title":               summary.Title,
					"summary":             summary.Summary,
					"priority":            summary.Priority,
					"value_score":         summary.ValueScore,
					"impact_direction":    summary.ImpactDirection,
					"investment_takeaway": summary.InvestmentTakeaway,
					"tickers":             summary.Tickers,
					"industry":            summary.Industry,
					"is_industry_wide":    summary.IsIndustryWide,
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

type BriefingSchemaOutput struct {
	Title            string            `json:"title" jsonschema:"description=Engaging title for today's market briefing"`
	MacroPulse       string            `json:"macro_pulse" jsonschema:"description=2-3 sentence summary of broader market sentiment and macro developments"`
	BullishLookout   []BriefingItem    `json:"bullish_lookout" jsonschema:"description=Top high-conviction companies with positive fundamental catalysts, expanding moats, or earnings growth"`
	BearishLookout   []BriefingItem    `json:"bearish_lookout" jsonschema:"description=Companies facing serious fundamental risks, governance issues, or severe headwind alerts"`
	SectorHighlights []SectorHighlight `json:"sector_highlights" jsonschema:"description=Industry-wide and sector developments grouped by industry"`
	ActionPlan       string            `json:"action_plan" jsonschema:"description=Disciplined, concrete 2-3 sentence action plan for long-term value investors entering today's session"`
}

func (s *Service) GenerateDailyBriefing(ctx context.Context, targetDate time.Time, bRepo BriefingRepository) (*Briefing, error) {
	s.logger.Info("Starting Daily Briefing generation", zap.Time("date", targetDate))
	if s.cfg == nil {
		return nil, fmt.Errorf("config not loaded")
	}

	// Calculate 24h window (yesterday 00:00 to targetDate 23:59:59)
	startWindow := targetDate.AddDate(0, 0, -1)
	startOfWindow := time.Date(startWindow.Year(), startWindow.Month(), startWindow.Day(), 0, 0, 0, 0, targetDate.Location())
	endOfWindow := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 23, 59, 59, 999999999, targetDate.Location())

	filter := bson.M{
		"created_at": bson.M{
			"$gte": startOfWindow,
			"$lte": endOfWindow,
		},
	}
	newsList, err := s.repo.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("fetching 24h news: %w", err)
	}

	if len(newsList) == 0 {
		s.logger.Warn("No news found in 24h window for briefing", zap.Time("start", startOfWindow), zap.Time("end", endOfWindow))
		return nil, nil
	}

	s.logger.Info("Aggregating news for briefing", zap.Int("count", len(newsList)))

	// Format news summaries for LLM prompt
	var sb strings.Builder
	for i, n := range newsList {
		sb.WriteString(fmt.Sprintf("%d. [%s] (Tickers: %v, Score: %+d, Direction: %s, Sector: %s)\nTitle: %s\nSummary: %s\nTakeaway: %s\n\n",
			i+1, n.Date.Format("2006-01-02"), n.Tickers, n.ValueScore, n.ImpactDirection, n.Industry, n.Title, n.Summary, n.InvestmentTakeaway))
	}

	client := openrouter.NewClient(s.cfg.OpenrouterApiKey)
	schema, err := jsonschema.GenerateSchemaForType(BriefingSchemaOutput{})
	if err != nil {
		return nil, fmt.Errorf("GenerateSchemaForType for briefing: %w", err)
	}

	prompt := fmt.Sprintf(`You are an elite Personal Investment Manager preparing the Daily Morning Market Intelligence Briefing for a disciplined Value Investor (Graham-Buffett-Munger school).

Based on all the news collected over the past 24 hours below, synthesize an authoritative, super-objective Daily Briefing answering:
1. Executive Macro & Market Pulse.
2. Stocks to Watch / Buy Lookout (Companies with strong moats, positive catalysts, high ROIC potential, or attractive fundamental developments).
3. Stocks to Avoid / Risk Lookout (Companies with governance red flags, balance sheet damage, or regulatory headwinds).
4. Sector & Industry Highlights (Macro or industry-wide trends e.g. Banking, Poultry, Mining, Energy).
5. Value Investor Action Plan for today's market session.

24-Hour News Digest:
"""
%s
"""`, sb.String())

	request := openrouter.ChatCompletionRequest{
		Model: "google/gemini-3.7-flash",
		Messages: []openrouter.ChatCompletionMessage{
			{
				Role:    openrouter.ChatMessageRoleUser,
				Content: openrouter.Content{Text: prompt},
			},
		},
		ResponseFormat: &openrouter.ChatCompletionResponseFormat{
			Type: openrouter.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openrouter.ChatCompletionResponseFormatJSONSchema{
				Name:   "daily_market_briefing",
				Schema: schema,
				Strict: true,
			},
		},
	}

	res, err := client.CreateChatCompletion(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("openrouter briefing completion: %w", err)
	}

	if len(res.Choices) == 0 {
		return nil, fmt.Errorf("no response choices from openrouter")
	}

	var output BriefingSchemaOutput
	if err := json.Unmarshal([]byte(res.Choices[0].Message.Content.Text), &output); err != nil {
		return nil, fmt.Errorf("unmarshaling briefing output: %w", err)
	}

	// Format Raw Markdown for email/rendering
	rawMarkdown := formatBriefingMarkdown(output, targetDate)

	briefing := &Briefing{
		ID:               bson.NewObjectID(),
		Date:             targetDate,
		Title:            output.Title,
		MacroPulse:       output.MacroPulse,
		BullishLookout:   output.BullishLookout,
		BearishLookout:   output.BearishLookout,
		SectorHighlights: output.SectorHighlights,
		ActionPlan:       output.ActionPlan,
		RawMarkdown:      rawMarkdown,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if bRepo != nil {
		if err := bRepo.Create(ctx, briefing); err != nil {
			s.logger.Error("Failed to persist briefing in MongoDB", zap.Error(err))
		} else {
			s.logger.Info("Successfully saved Daily Briefing in MongoDB", zap.String("id", briefing.ID.Hex()))
		}
	}

	return briefing, nil
}

func formatBriefingMarkdown(b BriefingSchemaOutput, d time.Time) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", b.Title))
	sb.WriteString(fmt.Sprintf("**Date:** %s (7:00 AM GMT+8)\n\n", d.Format("Monday, 02 January 2006")))
	sb.WriteString("## 🌐 Executive Macro & Market Pulse\n")
	sb.WriteString(b.MacroPulse + "\n\n")

	sb.WriteString("## 🟢 Stocks on Lookout (Buy / Opportunities)\n")
	if len(b.BullishLookout) == 0 {
		sb.WriteString("_No high-conviction bullish candidates today._\n\n")
	} else {
		for _, item := range b.BullishLookout {
			sb.WriteString(fmt.Sprintf("### %s (%s) — Value Score: %+d\n", item.Ticker, item.IssuerName, item.ValueScore))
			sb.WriteString(fmt.Sprintf("**Headline:** %s\n\n", item.Headline))
			sb.WriteString(fmt.Sprintf("**Rationale:** %s\n\n", item.Rationale))
			sb.WriteString(fmt.Sprintf("**Takeaway:** %s\n\n", item.InvestmentTakeaway))
		}
	}

	sb.WriteString("## 🔴 Stocks on Lookout (Risk Alerts / Headwinds)\n")
	if len(b.BearishLookout) == 0 {
		sb.WriteString("_No major risk alerts today._\n\n")
	} else {
		for _, item := range b.BearishLookout {
			sb.WriteString(fmt.Sprintf("### %s (%s) — Value Score: %+d\n", item.Ticker, item.IssuerName, item.ValueScore))
			sb.WriteString(fmt.Sprintf("**Headline:** %s\n\n", item.Headline))
			sb.WriteString(fmt.Sprintf("**Rationale:** %s\n\n", item.Rationale))
			sb.WriteString(fmt.Sprintf("**Takeaway:** %s\n\n", item.InvestmentTakeaway))
		}
	}

	sb.WriteString("## 🏭 Sector & Industry Highlights\n")
	for _, sec := range b.SectorHighlights {
		sb.WriteString(fmt.Sprintf("- **%s** [%s]: %s\n", sec.Sector, sec.Sentiment, sec.Summary))
	}
	sb.WriteString("\n")

	sb.WriteString("## 🎯 Value Investor Action Plan\n")
	sb.WriteString(b.ActionPlan + "\n")

	return sb.String()
}
