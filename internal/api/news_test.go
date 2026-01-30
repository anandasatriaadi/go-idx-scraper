package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/api"
	"github.com/anandasatriaadi/go-idx-scraper/internal/db/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// MockNewsRepository satisfies api.NewsRepository
type MockNewsRepository struct {
	Data map[string]*model.News
}

func (m *MockNewsRepository) FindAll(ctx context.Context, filter bson.M, opts ...options.Lister[options.FindOptions]) ([]*model.News, error) {
	var results []*model.News
	for _, n := range m.Data {
		// Basic filtering logic for test
		if dateFilter, ok := filter["date"].(bson.M); ok {
			if gte, ok := dateFilter["$gte"].(time.Time); ok {
				if n.Date.Before(gte) {
					continue
				}
			}
			if lte, ok := dateFilter["$lte"].(time.Time); ok {
				if n.Date.After(lte) {
					continue
				}
			}
		}

		if priority, ok := filter["priority"].(int); ok {
			if n.Priority != priority {
				continue
			}
		}

		if source, ok := filter["link"].(string); ok {
			if n.Link != source {
				continue
			}
		}

		results = append(results, n)
	}
	return results, nil
}

func (m *MockNewsRepository) FindByID(ctx context.Context, id bson.ObjectID) (*model.News, error) {
	if n, ok := m.Data[id.Hex()]; ok {
		return n, nil
	}
	return nil, errors.New("not found")
}

func TestNewsRoutes_GetAll(t *testing.T) {
	id1, _ := bson.ObjectIDFromHex("65c0e1234567890123456789")
	id2, _ := bson.ObjectIDFromHex("65c0e1234567890123456790")

	mockRepo := &MockNewsRepository{
		Data: map[string]*model.News{
			id1.Hex(): {
				Id:       id1,
				Title:    "News 1",
				Date:     time.Date(2024, 2, 1, 10, 0, 0, 0, time.UTC),
				Priority: 10,
			},
			id2.Hex(): {
				Id:       id2,
				Title:    "News 2",
				Date:     time.Date(2024, 2, 5, 10, 0, 0, 0, time.UTC),
				Priority: 1,
			},
		},
	}

	r := api.NewsRoutes(mockRepo)
	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", res.Status)
	}

	var news []model.News
	if err := json.NewDecoder(res.Body).Decode(&news); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(news) != 2 {
		t.Errorf("Expected 2 news items, got %d", len(news))
	}
}

func TestNewsRoutes_FilterByDate(t *testing.T) {
	mockRepo := &MockNewsRepository{
		Data: map[string]*model.News{
			"1": {Id: bson.NewObjectID(), Title: "Old News", Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
			"2": {Id: bson.NewObjectID(), Title: "Recent News", Date: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)},
		},
	}

	r := api.NewsRoutes(mockRepo)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Filter >= 2024-01-15
	url := ts.URL + "/?date_gte=" + time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	res, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer res.Body.Close()

	var news []model.News
	if err := json.NewDecoder(res.Body).Decode(&news); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(news) != 1 {
		t.Errorf("Expected 1 news item, got %d", len(news))
	}
	if len(news) > 0 && news[0].Title != "Recent News" {
		t.Errorf("Expected 'Recent News', got '%s'", news[0].Title)
	}
}

func TestNewsRoutes_GetByID(t *testing.T) {
	id := bson.NewObjectID()
	mockRepo := &MockNewsRepository{
		Data: map[string]*model.News{
			id.Hex(): {Id: id, Title: "Target News"},
		},
	}

	r := api.NewsRoutes(mockRepo)
	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/" + id.Hex())
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", res.Status)
	}

	var n model.News
	if err := json.NewDecoder(res.Body).Decode(&n); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if n.Title != "Target News" {
		t.Errorf("Expected title 'Target News', got '%s'", n.Title)
	}
}
