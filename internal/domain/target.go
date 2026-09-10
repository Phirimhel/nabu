package domain

import "time"

type Target struct {
	ID                   string     `json:"id"`
	Name                 string     `json:"name"`
	PhotoURL             string     `json:"photoUrl"`
	Description          string     `json:"description"`
	BottleCount          int64      `json:"bottleCount"`
	TotalBottlesReceived int64      `json:"totalBottlesReceived"`
	IsLocked             bool       `json:"isLocked"`
	LockUntil            *time.Time `json:"lockUntil,omitempty"`
	CreatedAt            time.Time  `json:"createdAt"`
}
