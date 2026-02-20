package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/news"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

type MockNewsRepository struct {
	NewsList []*news.News
	OneNews  *news.News
	Err      error
}

func (m *MockNewsRepository) Create(ctx context.Context, n *news.News) error { return m.Err }
func (m *MockNewsRepository) FindAll(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) ([]*news.News, error) {
	return m.NewsList, m.Err
}
func (m *MockNewsRepository) FindByID(ctx context.Context, id bson.ObjectID) (*news.News, error) {
	return m.OneNews, m.Err
}
func (m *MockNewsRepository) UpdateByID(ctx context.Context, id bson.ObjectID, update any) error {
	return m.Err
}

func TestNewsHandler_GetAll(t *testing.T) {
	mockRepo := &MockNewsRepository{
		NewsList: []*news.News{
			{Title: "News 1", Priority: 1},
			{Title: "News 2", Priority: 2},
		},
	}
	service := news.NewService(mockRepo, zap.NewNop(), nil)
	handler := NewNewsHandler(service)

	req := httptest.NewRequest("GET", "/?priority=1", nil)
	w := httptest.NewRecorder()

	handler.GetAll(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var results []*news.News
	json.NewDecoder(w.Body).Decode(&results)
	if len(results) != 2 {
		t.Errorf("Expected 2 news, got %d", len(results))
	}
}

func TestNewsHandler_GetByID(t *testing.T) {
	id := bson.NewObjectID()
	mockRepo := &MockNewsRepository{
		OneNews: &news.News{ID: id, Title: "Specific News"},
	}
	service := news.NewService(mockRepo, zap.NewNop(), nil)
	handler := NewNewsHandler(service)

	r := chi.NewRouter()
	r.Get("/{id}", handler.GetByID)

	req := httptest.NewRequest("GET", "/"+id.Hex(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result news.News
	json.NewDecoder(w.Body).Decode(&result)
	if result.Title != "Specific News" {
		t.Errorf("Expected title 'Specific News', got '%s'", result.Title)
	}
}
