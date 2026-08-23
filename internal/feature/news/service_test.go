package news

import (
	"context"
	"testing"

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
