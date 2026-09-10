package application

import (
	"context"
	"time"

	"nabutilivanie/internal/domain"
)

const ActiveClaimCooldownDuration = 40 * time.Second

type PlayerService struct {
	players  PlayerRepository
	cooldown ActiveClaimCooldown
}

func NewPlayerService(players PlayerRepository, cooldown ActiveClaimCooldown) *PlayerService {
	return &PlayerService{players: players, cooldown: cooldown}
}
func (s *PlayerService) Get(ctx context.Context, telegramID int64) (domain.Player, error) {
	return s.players.Player(ctx, telegramID)
}
func (s *PlayerService) ClaimActive(ctx context.Context, telegramID int64) (domain.Player, time.Duration, error) {
	p, err := s.players.Player(ctx, telegramID)
	if err != nil {
		return domain.Player{}, 0, err
	}
	ok, retryAfter, err := s.cooldown.Acquire(ctx, "active-claim:"+p.ID, ActiveClaimCooldownDuration)
	if err != nil || !ok {
		return p, retryAfter, err
	}
	p, err = s.players.Credit(ctx, telegramID, 1)
	return p, 0, err
}
func (s *PlayerService) ClaimPassive(ctx context.Context, telegramID int64) (domain.Player, int64, error) {
	if _, err := s.players.Player(ctx, telegramID); err != nil {
		return domain.Player{}, 0, err
	}
	return s.players.ClaimPassive(ctx, telegramID)
}
