package announcement

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

type MockRepository struct {
	Announcements []*Announcement
	One           *Announcement
	ExistsRes     bool
	Err           error
}

func (m *MockRepository) Create(ctx context.Context, a *Announcement) error { return m.Err }
func (m *MockRepository) FindAll(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) ([]*Announcement, error) {
	return m.Announcements, m.Err
}
func (m *MockRepository) FindByID(ctx context.Context, id string) (*Announcement, error) {
	return m.One, m.Err
}
func (m *MockRepository) Exists(ctx context.Context, id string) (bool, error) {
	return m.ExistsRes, m.Err
}

func TestService_Create(t *testing.T) {
	mock := &MockRepository{}
	svc := NewService(mock, zap.NewNop())
	err := svc.Create(context.Background(), &Announcement{ID: "1"})
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}
}

func TestService_FindAll(t *testing.T) {
	mock := &MockRepository{Announcements: []*Announcement{{ID: "1"}}}
	svc := NewService(mock, zap.NewNop())
	res, err := svc.FindAll(context.Background(), nil)
	if err != nil {
		t.Errorf("FindAll failed: %v", err)
	}
	if len(res) != 1 {
		t.Errorf("Expected 1 announcement, got %d", len(res))
	}
}
