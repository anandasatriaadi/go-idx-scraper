package news

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/revrost/go-openrouter/jsonschema"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

type MockRepository struct {
	NewsList []*News
	OneNews  *News
	Err      error
}

func (m *MockRepository) Create(ctx context.Context, n *News) error { return m.Err }
func (m *MockRepository) FindAll(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) ([]*News, error) {
	return m.NewsList, m.Err
}
func (m *MockRepository) FindByID(ctx context.Context, id bson.ObjectID) (*News, error) {
	return m.OneNews, m.Err
}
func (m *MockRepository) UpdateByID(ctx context.Context, id bson.ObjectID, update any) error {
	return m.Err
}
func (m *MockRepository) ExistsByLink(ctx context.Context, link string) (bool, error) {
	return false, m.Err
}

func TestService_Create(t *testing.T) {
	mock := &MockRepository{}
	svc := NewService(mock, zap.NewNop(), nil)
	err := svc.Create(context.Background(), &News{Title: "Test"})
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}
}

func TestService_FindAll(t *testing.T) {
	mock := &MockRepository{NewsList: []*News{{Title: "Test"}}}
	svc := NewService(mock, zap.NewNop(), nil)
	res, err := svc.FindAll(context.Background(), bson.M{})
	if err != nil {
		t.Errorf("FindAll failed: %v", err)
	}
	if len(res) != 1 {
		t.Errorf("Expected 1 news, got %d", len(res))
	}
}

func TestService_FindByID(t *testing.T) {
	id := bson.NewObjectID()
	mock := &MockRepository{OneNews: &News{ID: id, Title: "Test"}}
	svc := NewService(mock, zap.NewNop(), nil)
	res, err := svc.FindByID(context.Background(), id)
	if err != nil {
		t.Errorf("FindByID failed: %v", err)
	}
	if res.ID != id {
		t.Errorf("Expected ID %v, got %v", id, res.ID)
	}
}

func TestNewsEntity_Fields(t *testing.T) {
	id := bson.NewObjectID()
	n := &News{
		ID:                 id,
		Title:              "Sample News",
		Summary:            "3 sentence summary.",
		Content:            "Full markdown content.",
		Priority:           5,
		ValueScore:         7,
		ImpactDirection:    "Bullish",
		InvestmentTakeaway: "Strong cash flow and expansion potential.",
	}

	if n.ValueScore != 7 {
		t.Errorf("Expected ValueScore 7, got %d", n.ValueScore)
	}
	if n.ImpactDirection != "Bullish" {
		t.Errorf("Expected ImpactDirection 'Bullish', got '%s'", n.ImpactDirection)
	}
	if n.InvestmentTakeaway != "Strong cash flow and expansion potential." {
		t.Errorf("Expected InvestmentTakeaway to match, got '%s'", n.InvestmentTakeaway)
	}
}

func TestNewsSummary_SchemaAndPrompt(t *testing.T) {
	schema, err := jsonschema.GenerateSchemaForType(NewsSummary{})
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}
	if schema == nil {
		t.Fatal("Expected non-nil schema")
	}

	jsonSample := `{
		"title": "Emiten BBRI Perkuat Pendanaan",
		"summary": "BBRI membukukan pertumbuhan laba bersih dan peningkatan margin bunga bersih. Pertumbuhan ini didorong oleh segmen mikro yang solid. Manajemen mempertahankan rasio dividen tinggi.",
		"priority": 2,
		"value_score": 8,
		"impact_direction": "Bullish",
		"investment_takeaway": "Fundamental kuat dengan moat kokoh di pembiayaan mikro dan valuasi menarik."
	}`

	var summary NewsSummary
	if err := json.Unmarshal([]byte(jsonSample), &summary); err != nil {
		t.Fatalf("Failed to unmarshal NewsSummary: %v", err)
	}

	if summary.ValueScore != 8 {
		t.Errorf("Expected ValueScore 8, got %d", summary.ValueScore)
	}
	if summary.ImpactDirection != "Bullish" {
		t.Errorf("Expected ImpactDirection 'Bullish', got '%s'", summary.ImpactDirection)
	}
	if summary.InvestmentTakeaway == "" {
		t.Errorf("Expected non-empty InvestmentTakeaway")
	}
}

func TestNewsSummary_IDXICClassification(t *testing.T) {
	schema, err := jsonschema.GenerateSchemaForType(NewsSummary{})
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}
	if schema == nil {
		t.Fatal("Expected non-nil schema")
	}

	jsonSample := `{
		"title": "Bank Mandiri Catat Pertumbuhan Kredit 13%",
		"summary": "Kredit BMRI tumbuh kuat ditopang segmen korporasi dan komersial. Kualitas aset terjaga dengan NPL rendah. Laba bersih semester I meningkat signifikan.",
		"priority": 3,
		"value_score": 7,
		"impact_direction": "Bullish",
		"investment_takeaway": "Kinerja fundamental solid dengan profitabilitas tinggi dan solvabilitas kokoh.",
		"tickers": ["BMRI"],
		"sector": "G. Financials",
		"subsector": "G1. Banks",
		"is_industry_wide": false
	}`

	var summary NewsSummary
	if err := json.Unmarshal([]byte(jsonSample), &summary); err != nil {
		t.Fatalf("Failed to unmarshal NewsSummary: %v", err)
	}

	if summary.Sector != "G. Financials" {
		t.Errorf("Expected sector 'G. Financials', got '%s'", summary.Sector)
	}
	if summary.Subsector != "G1. Banks" {
		t.Errorf("Expected subsector 'G1. Banks', got '%s'", summary.Subsector)
	}
}

func TestBriefingEntity_Fields(t *testing.T) {
	id := bson.NewObjectID()
	now := time.Now()
	b := &Briefing{
		ID:         id,
		Date:       now,
		Title:      "Daily Market Briefing",
		MacroPulse: "Positive macro sentiment.",
		BullishLookout: []BriefingItem{
			{
				Ticker:             "BBRI",
				IssuerName:         "PT Bank Rakyat Indonesia Tbk",
				Headline:           "Expanding Net Interest Margin",
				Rationale:          "Strong micro loan disbursement.",
				ValueScore:         8,
				InvestmentTakeaway: "Attractive valuation and solid moat.",
			},
		},
		BearishLookout: []BriefingItem{
			{
				Ticker:             "ASBI",
				IssuerName:         "PT Asuransi Bintang Tbk",
				Headline:           "Embezzlement Investigation",
				Rationale:          "Internal controls failure.",
				ValueScore:         -7,
				InvestmentTakeaway: "Avoid until balance sheet risk is cleared.",
			},
		},
		SectorHighlights: []SectorHighlight{
			{
				Sector:    "Banking",
				Summary:   "Liquidity tightening persists.",
				Sentiment: "Neutral",
			},
		},
		ActionPlan: "Accumulate high ROE banks on weakness.",
	}

	if len(b.BullishLookout) != 1 || b.BullishLookout[0].Ticker != "BBRI" {
		t.Errorf("Expected BullishLookout with BBRI")
	}
	if len(b.BearishLookout) != 1 || b.BearishLookout[0].Ticker != "ASBI" {
		t.Errorf("Expected BearishLookout with ASBI")
	}
	if len(b.SectorHighlights) != 1 || b.SectorHighlights[0].Sector != "Banking" {
		t.Errorf("Expected SectorHighlights with Banking")
	}
}

func TestNewsSummary_TickersAndIndustry(t *testing.T) {
	schema, err := jsonschema.GenerateSchemaForType(NewsSummary{})
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}
	if schema == nil {
		t.Fatal("Expected non-nil schema")
	}

	jsonSample := `{
		"title": "Kinerja Emiten Unggas Menguat",
		"summary": "Emiten peternakan ayam mencatat pemulihan margin di semester II. Permintaan stabil menopang pertumbuhan laba. Efisiensi pakan menjadi faktor pendukung utama.",
		"priority": 4,
		"value_score": 6,
		"impact_direction": "Bullish",
		"investment_takeaway": "CPIN dan JPFA memiliki posisi pasar dominan dan efisiensi rantai pasok terintegrasi.",
		"tickers": ["CPIN", "JPFA", "MAIN"],
		"sector": "D. Consumer Non-Cyclicals",
		"subsector": "D2. Food and Beverage",
		"is_industry_wide": true
	}`

	var summary NewsSummary
	if err := json.Unmarshal([]byte(jsonSample), &summary); err != nil {
		t.Fatalf("Failed to unmarshal NewsSummary: %v", err)
	}

	if len(summary.Tickers) != 3 || summary.Tickers[0] != "CPIN" {
		t.Errorf("Expected tickers with CPIN, got %v", summary.Tickers)
	}
	if summary.Sector != "D. Consumer Non-Cyclicals" {
		t.Errorf("Expected sector 'D. Consumer Non-Cyclicals', got '%s'", summary.Sector)
	}
	if summary.Subsector != "D2. Food and Beverage" {
		t.Errorf("Expected subsector 'D2. Food and Beverage', got '%s'", summary.Subsector)
	}
	if !summary.IsIndustryWide {
		t.Errorf("Expected is_industry_wide to be true")
	}
}

func TestDailyBriefing_SchemaGeneration(t *testing.T) {
	schema, err := jsonschema.GenerateSchemaForType(BriefingSchemaOutput{})
	if err != nil {
		t.Fatalf("Failed to generate BriefingSchemaOutput schema: %v", err)
	}
	if schema == nil {
		t.Fatal("Expected non-nil schema")
	}

	jsonSample := `{
		"title": "Morning Market Intelligence Briefing - 24 August 2026",
		"macro_pulse": "IHSG consolidates as foreign flows stabilize and commodity prices rebound.",
		"bullish_lookout": [
			{
				"ticker": "BBRI",
				"issuer_name": "PT Bank Rakyat Indonesia Tbk",
				"headline": "Micro Lending Expansion",
				"rationale": "High NIM and robust capital adequacy ratio.",
				"value_score": 8,
				"investment_takeaway": "Long-term compounding opportunity."
			}
		],
		"bearish_lookout": [
			{
				"ticker": "ASBI",
				"issuer_name": "PT Asuransi Bintang Tbk",
				"headline": "Embezzlement Scandal",
				"rationale": "Governance breakdown.",
				"value_score": -7,
				"investment_takeaway": "High governance risk; avoid."
			}
		],
		"sector_highlights": [
			{
				"sector": "Banking",
				"summary": "Liquidity remains tight but top tier banks maintain pricing power.",
				"sentiment": "Neutral"
			}
		],
		"action_plan": "Focus on high-quality big cap banks and poultry leaders with widening moats."
	}`

	var output BriefingSchemaOutput
	if err := json.Unmarshal([]byte(jsonSample), &output); err != nil {
		t.Fatalf("Failed to unmarshal BriefingSchemaOutput: %v", err)
	}

	if output.Title == "" || len(output.BullishLookout) != 1 {
		t.Errorf("Unexpected unmarshaled briefing: %+v", output)
	}
}
