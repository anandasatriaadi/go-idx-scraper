package financialreport

import (
	"context"

	"github.com/anandasatriaadi/go-idx-scraper/internal/domain/financialreport"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

type Service struct {
	repo   financialreport.Repository
	logger *zap.Logger
}

func NewService(repo financialreport.Repository, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

func (s *Service) Create(ctx context.Context, f *financialreport.FinancialReport) error {
	return s.repo.Create(ctx, f)
}

func (s *Service) FindAll(ctx context.Context, filter interface{}, opts ...options.Lister[options.FindOptions]) ([]*financialreport.FinancialReport, error) {
	return s.repo.FindAll(ctx, filter, opts...)
}
