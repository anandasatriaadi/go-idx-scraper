package finreport

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
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

func (s *Service) Create(ctx context.Context, f *FinancialReport) error {
	return s.repo.Create(ctx, f)
}

func (s *Service) FindAll(ctx context.Context, filter bson.M, opts ...options.Lister[options.FindOptions]) ([]*FinancialReport, error) {
	return s.repo.FindAll(ctx, filter, opts...)
}
