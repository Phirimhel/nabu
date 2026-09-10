package domain

import "time"

type Player struct {
	ID                 string
	TelegramID         int64
	BottleBalance      int64 `json:"bottleBalance"`
	LastPassiveClaimAt time.Time
}
