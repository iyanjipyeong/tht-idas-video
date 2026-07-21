package entity

import "time"

type Video struct {
	ID          UUID
	Title       string
	Description string
	Category    Tier
	VideoURL    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
