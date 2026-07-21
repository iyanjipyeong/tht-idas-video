package entity

import "time"

type User struct {
	ID        UUID
	Name      string
	Email     string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
