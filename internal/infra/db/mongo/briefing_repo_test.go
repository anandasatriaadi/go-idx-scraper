package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/news"
)

type MockBriefingRepo struct {
	Briefing *news.Briefing
	List     []*news.Briefing
	Err      error
}

func (m *MockBriefingRepo) Create(ctx context.Context, b *news.Briefing) error {
	m.Briefing = b
	return m.Err
}
func (m *MockBriefingRepo) FindByDate(ctx context.Context, date time.Time) (*news.Briefing, error) {
	return m.Briefing, m.Err
}
func (m *MockBriefingRepo) FindLatest(ctx context.Context) (*news.Briefing, error) {
	return m.Briefing, m.Err
}
func (m *MockBriefingRepo) FindRecent(ctx context.Context, limit int) ([]*news.Briefing, error) {
	return m.List, m.Err
}

func TestBriefingStruct(t *testing.T) {
	b := &news.Briefing{
		ID:    "briefing-test",
		Title: "Morning Briefing Test",
	}
	if b.Title != "Morning Briefing Test" {
		t.Errorf("Unexpected title: %s", b.Title)
	}

	var _ news.BriefingRepository = (*MockBriefingRepo)(nil)
}

func TestNewsRepository_InterfaceCompliance(t *testing.T) {
	var _ news.Repository = (*NewsRepository)(nil)
}

func TestBriefingRepository_InterfaceCompliance(t *testing.T) {
	var _ news.BriefingRepository = (*BriefingRepository)(nil)
}
