package announcement

import (
	"context"

	"github.com/anandasatriaadi/go-idx-scraper/internal/domain/announcement"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

type Service struct {
	repo   announcement.Repository
	logger *zap.Logger
}

func NewService(repo announcement.Repository, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

func (s *Service) Create(ctx context.Context, a *announcement.Announcement) error {
	return s.repo.Create(ctx, a)
}

func (s *Service) CreateMany(ctx context.Context, announcements []*announcement.Announcement) error {
	for _, a := range announcements {
		if err := s.repo.Create(ctx, a); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) FindAll(ctx context.Context, filter interface{}, opts ...options.Lister[options.FindOptions]) ([]*announcement.Announcement, error) {
	return s.repo.FindAll(ctx, filter, opts...)
}

func (s *Service) FindByID(ctx context.Context, id string) (*announcement.Announcement, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Exists(ctx context.Context, id string) (bool, error) {
	return s.repo.Exists(ctx, id)
}
