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
