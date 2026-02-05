package announcement

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

type Service struct {
	repo   Repository
	logger *zap.Logger
}

func NewService(repo Repository, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

func (s *Service) Create(ctx context.Context, a *Announcement) error {
	return s.repo.Create(ctx, a)
}

func (s *Service) CreateMany(ctx context.Context, announcements []*Announcement) error {
	for _, a := range announcements {
		if err := s.repo.Create(ctx, a); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) FindAll(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) ([]*Announcement, error) {
	return s.repo.FindAll(ctx, filter, opts...)
}

func (s *Service) FindByID(ctx context.Context, id string) (*Announcement, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Exists(ctx context.Context, id string) (bool, error) {
	return s.repo.Exists(ctx, id)
}
