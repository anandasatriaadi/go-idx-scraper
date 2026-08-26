package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/anandasatriaadi/go-idx-scraper/internal/feature/announcement"
)

type MockAnnouncementRepo struct {
	Announcements map[string]*announcement.Announcement
	Err           error
}

func NewMockAnnouncementRepo() *MockAnnouncementRepo {
	return &MockAnnouncementRepo{
		Announcements: make(map[string]*announcement.Announcement),
	}
}

func (m *MockAnnouncementRepo) Create(ctx context.Context, model *announcement.Announcement) error {
	if m.Err != nil {
		return m.Err
	}
	m.Announcements[model.ID] = model
	return nil
}

func (m *MockAnnouncementRepo) FindByID(ctx context.Context, id string) (*announcement.Announcement, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Announcements[id], nil
}

func (m *MockAnnouncementRepo) Exists(ctx context.Context, id string) (bool, error) {
	if m.Err != nil {
		return false, m.Err
	}
	_, ok := m.Announcements[id]
	return ok, nil
}

func (m *MockAnnouncementRepo) FindRecent(ctx context.Context, limit int) ([]*announcement.Announcement, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	var list []*announcement.Announcement
	for _, a := range m.Announcements {
		list = append(list, a)
		if limit > 0 && len(list) >= limit {
			break
		}
	}
	return list, nil
}

func (m *MockAnnouncementRepo) GetLatestCreatedDate(ctx context.Context) (*time.Time, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	var latest *time.Time
	for _, a := range m.Announcements {
		if a.CreatedDate != nil {
			if latest == nil || a.CreatedDate.After(*latest) {
				latest = a.CreatedDate
			}
		}
	}
	return latest, nil
}

func (m *MockAnnouncementRepo) FindExistingIDs(ctx context.Context, limit int) (map[string]bool, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	res := make(map[string]bool)
	for id := range m.Announcements {
		res[id] = true
		if limit > 0 && len(res) >= limit {
			break
		}
	}
	return res, nil
}

func TestAnnouncementRepository_InterfaceCompliance(t *testing.T) {
	var _ announcement.Repository = (*AnnouncementRepository)(nil)
	var _ announcement.Repository = (*MockAnnouncementRepo)(nil)
}

func TestAnnouncement_MockRepo(t *testing.T) {
	repo := NewMockAnnouncementRepo()
	ctx := context.Background()

	now := time.Now()
	kode := "BBRI"
	ann := &announcement.Announcement{
		ID:          "ann-123",
		KodeEmiten:  &kode,
		CreatedDate: &now,
	}

	if err := repo.Create(ctx, ann); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	exists, err := repo.Exists(ctx, "ann-123")
	if err != nil || !exists {
		t.Fatalf("Expected Exists=true, got %v, err=%v", exists, err)
	}

	found, err := repo.FindByID(ctx, "ann-123")
	if err != nil || found == nil || found.ID != "ann-123" {
		t.Fatalf("FindByID failed: %v, found=%+v", err, found)
	}
}
