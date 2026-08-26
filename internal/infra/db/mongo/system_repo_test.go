package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/system"
)

type MockSystemRepo struct {
	LastRuns map[string]*system.LastRun
	Err      error
}

func NewMockSystemRepo() *MockSystemRepo {
	return &MockSystemRepo{
		LastRuns: make(map[string]*system.LastRun),
	}
}

func (m *MockSystemRepo) GetLastRun(ctx context.Context, scriptName string) (*system.LastRun, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.LastRuns[scriptName], nil
}

func (m *MockSystemRepo) SaveLastRun(ctx context.Context, lastRun *system.LastRun) error {
	if m.Err != nil {
		return m.Err
	}
	m.LastRuns[lastRun.ScriptName] = lastRun
	return nil
}

func TestSystemRepository_InterfaceCompliance(t *testing.T) {
	var _ system.Repository = (*SystemRepository)(nil)
	var _ system.Repository = (*MockSystemRepo)(nil)
}

func TestSystem_MockRepo(t *testing.T) {
	repo := NewMockSystemRepo()
	ctx := context.Background()

	now := time.Now()
	lr := &system.LastRun{
		ScriptName: "idx-announcement",
		LastRunAt:  now,
		Metadata:   map[string]any{"latest_date": now},
	}

	if err := repo.SaveLastRun(ctx, lr); err != nil {
		t.Fatalf("SaveLastRun failed: %v", err)
	}

	found, err := repo.GetLastRun(ctx, "idx-announcement")
	if err != nil || found == nil {
		t.Fatalf("GetLastRun failed: %v, found=%+v", err, found)
	}

	if found.ScriptName != "idx-announcement" {
		t.Errorf("Expected scriptName idx-announcement, got %s", found.ScriptName)
	}
}
