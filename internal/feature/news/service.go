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
	Priority           int      `json:"priority" jsonschema:"description=Market impact priority from 1 (highest market urgency) to 10 (routine)"`
	ValueScore         int      `json:"value_score" jsonschema:"description=Fundamental value investing impact score strictly between -10 and +10"`
	ImpactDirection    string   `json:"impact_direction" jsonschema:"enum=Bullish,enum=Bearish,enum=Neutral,description=Directional impact on underlying business intrinsic value"`
	InvestmentTakeaway string   `json:"investment_takeaway" jsonschema:"description=1-2 sentence actionable takeaway for a disciplined long-term value investor"`
	Tickers            []string `json:"tickers" jsonschema:"description=List of 4-letter IDX stock ticker symbols (e.g. ['BBRI', 'BBCA']). Empty array if none."`
	Sector             string   `json:"sector" jsonschema:"enum=A. Energy,enum=B. Basic Materials,enum=C. Industrials,enum=D. Consumer Non-Cyclicals,enum=E. Consumer Cyclicals,enum=F. Healthcare,enum=G. Financials,enum=H. Properties and Real Estate,enum=I. Technology,enum=J. Infrastructures,enum=K. Transportation and Logistic,enum=Macroeconomics,description=Official IDX Industrial Classification (IDX-IC) Primary Sector"`
	Subsector          string   `json:"subsector" jsonschema:"enum=A1. Oil, Gas, and Coal,enum=A2. Alternative Energy,enum=B1. Basic Materials,enum=C1. Industrial Goods,enum=C2. Industrial Services,enum=C3. Multi-sector Holdings,enum=D1. Food and Staples Retailing,enum=D2. Food and Beverage,enum=D3. Tobacco,enum=D4. Nondurable Household Products,enum=E1. Automobiles and Components,enum=E2. Household Goods,enum=E3. Leisure Goods,enum=E4. Apparel and Luxury Goods,enum=E5. Consumer Services,enum=E6. Media and Entertainment,enum=E7. Retailing,enum=F1. Healthcare Equipment & Providers,enum=F2. Pharmaceuticals & Health Care Research,enum=G1. Banks,enum=G2. Financing Service,enum=G3. Investment Service,enum=G4. Insurance,enum=G5. Holding and Investment Companies,enum=H1. Properties & Real Estate,enum=I1. Software & IT Services,enum=I2. Technology Hardware & Equipment,enum=J1. Transportation Infrastructure,enum=J2. Heavy Constructions & Civil Engineering,enum=J3. Telecommunication,enum=J4. Utilities,enum=K1. Transportation,enum=K2. Logistics & Deliveries,enum=General Market & Policy,description=Official IDX-IC Subsector classification"`
	IsIndustryWide     bool     `json:"is_industry_wide" jsonschema:"description=True if the news affects an entire sector or macroeconomic policy rather than just one individual company"`
}

type BatchItemSummary struct {
	ArticleID          string   `json:"article_id" jsonschema:"description=The exact article ID matching the provided article"`
	Title              string   `json:"title" jsonschema:"description=An updated engaging and objective title capturing the essence of the article"`
	Summary            string   `json:"summary" jsonschema:"description=Concise 3-sentence summary highlighting financial facts figures and immediate market implications"`
	Priority           int      `json:"priority" jsonschema:"description=Market impact priority from 1 (highest market urgency) to 10 (routine)"`
	ValueScore         int      `json:"value_score" jsonschema:"description=Fundamental value investing impact score strictly between -10 and +10"`
	ImpactDirection    string   `json:"impact_direction" jsonschema:"enum=Bullish,enum=Bearish,enum=Neutral,description=Directional impact on underlying business intrinsic value"`
	InvestmentTakeaway string   `json:"investment_takeaway" jsonschema:"description=1-2 sentence actionable takeaway for a disciplined long-term value investor"`
	Tickers            []string `json:"tickers" jsonschema:"description=List of 4-letter IDX stock ticker symbols (e.g. ['BBRI', 'BBCA']). Empty array if none."`
	Sector             string   `json:"sector" jsonschema:"enum=A. Energy,enum=B. Basic Materials,enum=C. Industrials,enum=D. Consumer Non-Cyclicals,enum=E. Consumer Cyclicals,enum=F. Healthcare,enum=G. Financials,enum=H. Properties and Real Estate,enum=I. Technology,enum=J. Infrastructures,enum=K. Transportation and Logistic,enum=Macroeconomics,description=Official IDX Industrial Classification (IDX-IC) Primary Sector"`
	Subsector          string   `json:"subsector" jsonschema:"enum=A1. Oil, Gas, and Coal,enum=A2. Alternative Energy,enum=B1. Basic Materials,enum=C1. Industrial Goods,enum=C2. Industrial Services,enum=C3. Multi-sector Holdings,enum=D1. Food and Staples Retailing,enum=D2. Food and Beverage,enum=D3. Tobacco,enum=D4. Nondurable Household Products,enum=E1. Automobiles and Components,enum=E2. Household Goods,enum=E3. Leisure Goods,enum=E4. Apparel and Luxury Goods,enum=E5. Consumer Services,enum=E6. Media and Entertainment,enum=E7. Retailing,enum=F1. Healthcare Equipment & Providers,enum=F2. Pharmaceuticals & Health Care Research,enum=G1. Banks,enum=G2. Financing Service,enum=G3. Investment Service,enum=G4. Insurance,enum=G5. Holding and Investment Companies,enum=H1. Properties & Real Estate,enum=I1. Software & IT Services,enum=I2. Technology Hardware & Equipment,enum=J1. Transportation Infrastructure,enum=J2. Heavy Constructions & Civil Engineering,enum=J3. Telecommunication,enum=J4. Utilities,enum=K1. Transportation,enum=K2. Logistics & Deliveries,enum=General Market & Policy,description=Official IDX-IC Subsector classification"`
	IsIndustryWide     bool     `json:"is_industry_wide" jsonschema:"description=True if the news affects an entire sector or macroeconomic policy rather than just one individual company"`
}

type BatchSummariesOutput struct {
	Summaries []BatchItemSummary `json:"summaries" jsonschema:"description=List of structured evaluations for each article in the batch"`
}

func (s *Service) Summarize(ctx context.Context, ids []bson.ObjectID) error {
	s.logger.Info("Starting news batch summarization", zap.Int("num_ids", len(ids)))
	if s.cfg == nil {
		return fmt.Errorf("config not loaded")
	}
	client := openrouter.NewClient(s.cfg.OpenrouterApiKey)

	batchSchema, err := jsonschema.GenerateSchemaForType(BatchSummariesOutput{})
	if err != nil {
		return fmt.Errorf("GenerateSchemaForType error: %w", err)
	}

	const batchSize = 8

	for i := 0; i < len(ids); i += batchSize {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}
		chunkIDs := ids[i:end]

		var validArticles []*News
		var promptArticles strings.Builder

		for _, id := range chunkIDs {
			n, err := s.repo.FindByID(ctx, id)
			if err != nil {
				s.logger.Warn("Error finding news by ID", zap.String("id", id.Hex()), zap.Error(err))
				continue
			}
			if n.Status == StatusSummarized && n.Summary != "" {
				s.logger.Debug("Article already summarized, skipping", zap.String("id", id.Hex()))
				continue
			}

			// Trim content if exceptionally large to protect token budget
			content := strings.TrimSpace(n.Content)
			if len(content) > 3500 {
				content = content[:3500] + "..."
			}

			validArticles = append(validArticles, n)
			promptArticles.WriteString(fmt.Sprintf("\n==============================\n[Article ID: %s]\nOriginal Title: %s\nContent:\n%s\n==============================\n", n.ID.Hex(), n.Title, content))
		}

		if len(validArticles) == 0 {
			continue
		}

		s.logger.Info("Submitting batch summarization to Gemini 3.7 Flash",
			zap.Int("batch_index", (i/batchSize)+1),
			zap.Int("total_batches", (len(ids)+batchSize-1)/batchSize),
			zap.Int("articles_in_batch", len(validArticles)),
		)

		prompt := fmt.Sprintf(`You are an expert Personal Investment Manager and seasoned Value Investor adhering strictly to the fundamental principles of Benjamin Graham, Warren Buffett, and Charlie Munger.

Analyze each of the following financial news articles with super objective, disciplined rigor. Evaluate whether each event impacts intrinsic business value, economic moats, capital allocation, balance sheet durability, or free cash flow generation.

Provide a structured evaluation for EVERY article in the batch, returning an array of summaries with the matching article_id.

Guidelines per article:
- article_id: The exact article ID string provided.
- title: Clear, concise, and professional headline.
- summary: Exactly 3 sentences. Cover core facts, financial metrics, and operational impact.
- priority: Integer from 1 (highest market urgency) to 10 (routine noise).
- value_score: Integer from -10 to +10 (-10 = severe capital destruction, 0 = neutral noise, +10 = massive moat widening).
- impact_direction: Exactly "Bullish", "Bearish", or "Neutral".
- investment_takeaway: 1-2 sentence bottom-line takeaway for a disciplined value investor.
- tickers: Array of uppercase 4-letter IDX stock tickers (e.g. ["BBRI", "BBCA"]). Empty array [] if none.
- sector: Official IDX-IC Sector (e.g. "G. Financials", "A. Energy", "Macroeconomics").
- subsector: Official IDX-IC Subsector (e.g. "G1. Banks", "A1. Oil, Gas, and Coal", "General Market & Policy").
- is_industry_wide: Boolean true if the news affects an entire sector or macro economy rather than an isolated company.

ARTICLES TO EVALUATE:
%s`, promptArticles.String())

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
					Name:   "batch_summaries",
					Schema: batchSchema,
					Strict: true,
				},
			},
		}

		res, err := client.CreateChatCompletion(ctx, request)
		if err != nil {
			s.logger.Error("Error in batch chat completion", zap.Error(err))
			// Mark batch articles as failed for future retry
			for _, art := range validArticles {
				_ = s.repo.UpdateByID(ctx, art.ID, bson.M{"$set": bson.M{"status": StatusFailed}})
			}
			continue
		}

		if len(res.Choices) == 0 {
			s.logger.Warn("Empty choices returned from batch completion")
			continue
		}

		var batchOutput BatchSummariesOutput
		if err := json.Unmarshal([]byte(res.Choices[0].Message.Content.Text), &batchOutput); err != nil {
			s.logger.Error("Error unmarshaling batch output", zap.Error(err))
			continue
		}

		// Map summaries by article ID
		summariesMap := make(map[string]BatchItemSummary)
		for _, sm := range batchOutput.Summaries {
			summariesMap[sm.ArticleID] = sm
		}

		// Persist each evaluated summary
		for _, art := range validArticles {
			sm, found := summariesMap[art.ID.Hex()]
			if !found {
				// Try partial match if id hex prefix matched
				for k, v := range summariesMap {
					if strings.Contains(k, art.ID.Hex()) || strings.Contains(art.ID.Hex(), k) {
						sm = v
						found = true
						break
					}
				}
			}

			if !found {
				s.logger.Warn("Article not found in batch output", zap.String("id", art.ID.Hex()))
				_ = s.repo.UpdateByID(ctx, art.ID, bson.M{"$set": bson.M{"status": StatusPending}})
				continue
			}

			update := bson.M{
				"$set": bson.M{
					"title":               sm.Title,
					"summary":             sm.Summary,
					"priority":            sm.Priority,
					"value_score":         sm.ValueScore,
					"impact_direction":    sm.ImpactDirection,
					"investment_takeaway": sm.InvestmentTakeaway,
					"tickers":             sm.Tickers,
					"sector":              sm.Sector,
					"subsector":           sm.Subsector,
					"industry":            sm.Subsector,
					"is_industry_wide":    sm.IsIndustryWide,
					"status":              StatusSummarized,
					"updated_at":          time.Now(),
				},
			}

			if err := s.repo.UpdateByID(ctx, art.ID, update); err != nil {
				s.logger.Error("Failed to update summarized news", zap.String("id", art.ID.Hex()), zap.Error(err))
			} else {
				s.logger.Info("Summarized & saved",
					zap.String("id", art.ID.Hex()),
					zap.String("title", sm.Title),
					zap.Int("value_score", sm.ValueScore),
					zap.String("impact", sm.ImpactDirection),
					zap.Strings("tickers", sm.Tickers),
				)
			}
		}
	}

	s.logger.Info("Completed batch news summarization successfully")
	return nil
}

// SummarizeViaOpenRouterBatchAPI submits unsummarized articles to OpenRouter's asynchronous Batch API
// to receive the 50% discount on token pricing and process large backfills without connection timeouts.
func (s *Service) SummarizeViaOpenRouterBatchAPI(ctx context.Context, ids []bson.ObjectID) error {
	s.logger.Info("Starting OpenRouter Batch API summarization (50% token discount)", zap.Int("num_ids", len(ids)))
	if s.cfg == nil || s.cfg.OpenrouterApiKey == "" {
		return fmt.Errorf("openrouter API key not configured")
	}

	schema, err := jsonschema.GenerateSchemaForType(NewsSummary{})
	if err != nil {
		return fmt.Errorf("generating schema: %w", err)
	}

	var batchRequests []BatchRequestItem

	for _, id := range ids {
		n, err := s.repo.FindByID(ctx, id)
		if err != nil || n == nil {
			continue
		}
		if n.Status == StatusSummarized && n.Summary != "" {
			continue
		}

		prompt := fmt.Sprintf(`You are an expert Personal Investment Manager and seasoned Value Investor adhering strictly to the fundamental principles of Benjamin Graham, Warren Buffett, and Charlie Munger.

Analyze the provided financial news article with super objective, disciplined rigor. Evaluate whether this event impacts intrinsic business value, economic moats, capital allocation, balance sheet durability, or free cash flow generation.

Provide your evaluation adhering to the following rules:
- Title: Clear, concise, and professional headline.
- Summary: Exactly 3 sentences. Cover core facts, financial metrics, and operational impact.
- Priority: Integer from 1 to 10 (1 = critical high-impact market event, 10 = routine noise).
- ValueScore: Integer from -10 to +10 (-10 = severe capital destruction, 0 = neutral noise, +10 = massive moat widening).
- ImpactDirection: Exactly "Bullish", "Bearish", or "Neutral".
- InvestmentTakeaway: 1 to 2 sentences summarizing the bottom-line takeaway for a long-term value investor.
- Tickers: Array of uppercase 4-letter Indonesian stock tickers explicitly mentioned or impacted (e.g. ["BBRI", "BBCA"]). Empty array [] if no specific company is mentioned.
- Sector: Official IDX-IC Sector (e.g. "G. Financials", "A. Energy", "Macroeconomics").
- Subsector: Official IDX-IC Subsector (e.g. "G1. Banks", "A1. Oil, Gas, and Coal", "General Market & Policy").
- IsIndustryWide: Boolean true if the news affects the whole sector or macro economy rather than an isolated company.

Article:
"""
%s
"""`, n.Content)

		batchRequests = append(batchRequests, BatchRequestItem{
			CustomID: n.ID.Hex(),
			Body: BatchRequestBody{
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
			},
		})
	}

	if len(batchRequests) == 0 {
		s.logger.Info("All requested articles are already summarized")
		return nil
	}

	s.logger.Info("Submitting batch to OpenRouter Batch API (50% pricing discount)", zap.Int("requests_count", len(batchRequests)))
	batch, err := SubmitOpenRouterBatch(ctx, s.cfg.OpenrouterApiKey, "google/gemini-3.7-flash", batchRequests)
	if err != nil {
		return fmt.Errorf("submitting openrouter batch: %w", err)
	}

	s.logger.Info("OpenRouter Batch successfully queued",
		zap.String("batch_id", batch.ID),
		zap.String("status", batch.Status),
		zap.Int("total_requests", batch.RequestCounts.Total),
	)

	// Poll for batch completion
	ticker := time.NewTicker(6 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			current, err := GetOpenRouterBatch(ctx, s.cfg.OpenrouterApiKey, batch.ID)
			if err != nil {
				s.logger.Warn("Failed to poll batch status", zap.String("batch_id", batch.ID), zap.Error(err))
				continue
			}

			s.logger.Info("Batch progress",
				zap.String("batch_id", current.ID),
				zap.String("status", current.Status),
				zap.Int("completed", current.RequestCounts.Completed),
				zap.Int("total", current.RequestCounts.Total),
			)

			if current.Status == "completed" {
				s.logger.Info("Batch completed! Processing results...",
					zap.String("batch_id", current.ID),
					zap.Any("usage", current.Usage),
				)

				successCount := 0
				for _, resItem := range current.Results {
					artID, err := bson.ObjectIDFromHex(resItem.CustomID)
					if err != nil {
						continue
					}

					if resItem.Response != nil && len(resItem.Response.Body.Choices) > 0 {
						var summary NewsSummary
						text := resItem.Response.Body.Choices[0].Message.Content
						if err := json.Unmarshal([]byte(text), &summary); err == nil {
							update := bson.M{
								"$set": bson.M{
									"title":               summary.Title,
									"summary":             summary.Summary,
									"priority":            summary.Priority,
									"value_score":         summary.ValueScore,
									"impact_direction":    summary.ImpactDirection,
									"investment_takeaway": summary.InvestmentTakeaway,
									"tickers":             summary.Tickers,
									"sector":              summary.Sector,
									"subsector":           summary.Subsector,
									"industry":            summary.Subsector,
									"is_industry_wide":    summary.IsIndustryWide,
									"status":              StatusSummarized,
									"updated_at":          time.Now(),
								},
							}
							if err := s.repo.UpdateByID(ctx, artID, update); err == nil {
								successCount++
							}
						}
					} else if resItem.Error != nil {
						_ = s.repo.UpdateByID(ctx, artID, bson.M{"$set": bson.M{"status": StatusFailed}})
					}
				}

				s.logger.Info("Finished updating batch articles in MongoDB",
					zap.Int("success_count", successCount),
					zap.Int("total_batch", len(current.Results)),
				)
				return nil
			}

			if current.Status == "failed" || current.Status == "expired" || current.Status == "cancelled" {
				return fmt.Errorf("batch terminated with status: %s", current.Status)
			}
		}
	}
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
		sec := n.Sector
		if sec == "" {
			sec = n.Industry
		}
		sb.WriteString(fmt.Sprintf("%d. [%s] (Tickers: %v, Score: %+d, Direction: %s, Sector: %s)\nTitle: %s\nSummary: %s\nTakeaway: %s\n\n",
			i+1, n.Date.Format("2006-01-02"), n.Tickers, n.ValueScore, n.ImpactDirection, sec, n.Title, n.Summary, n.InvestmentTakeaway))
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
