package application

import (
	"context"
	"time"

	"nabutilivanie/internal/domain"
)

type TargetRepository interface {
	CreateTarget(context.Context, string, string, string) (domain.Target, error)
	Leaderboard(context.Context, int) ([]domain.Target, error)
	Search(context.Context, string, int) ([]domain.Target, error)
	Nabutilit(context.Context, int64, string, string) (domain.Target, error)
}

type PlayerRepository interface {
	Player(context.Context, int64) (domain.Player, error)
	Credit(context.Context, int64, int64) (domain.Player, error)
	ClaimPassive(context.Context, int64) (domain.Player, int64, error)
}

type PaymentRepository interface {
	ProcessPayment(context.Context, string, string, int64, int64, any) (bool, domain.Player, error)
}

type ActiveClaimCooldown interface {
	Acquire(context.Context, string, time.Duration) (acquired bool, retryAfter time.Duration, err error)
}
