package model

import "time"

type Model struct {
	ID        string
	Type      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
