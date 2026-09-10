package application

import (
	"context"
	"strings"

	"nabutilivanie/internal/domain"
)

type TargetService struct{ targets TargetRepository }

func NewTargetService(targets TargetRepository) *TargetService {
	return &TargetService{targets: targets}
}

func (s *TargetService) Create(ctx context.Context, name, photoURL, description string) (domain.Target, error) {
	return s.targets.CreateTarget(ctx, strings.TrimSpace(name), photoURL, description)
}
func (s *TargetService) Leaderboard(ctx context.Context, limit int) ([]domain.Target, error) {
	return s.targets.Leaderboard(ctx, limit)
}
func (s *TargetService) Search(ctx context.Context, query string, limit int) ([]domain.Target, error) {
	return s.targets.Search(ctx, strings.TrimSpace(query), limit)
}
func (s *TargetService) Nabutilit(ctx context.Context, telegramID int64, targetID, comment string) (domain.Target, error) {
	return s.targets.Nabutilit(ctx, telegramID, targetID, strings.TrimSpace(comment))
}
