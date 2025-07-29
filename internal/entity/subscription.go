package entity

import (
	"time"

	"github.com/google/uuid"
)

type Service struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Price int       `json:"price"`
}

type Subscription struct {
	ID        uuid.UUID  `json:"id"`
	Service   Service    `json:"service"`
	UserID    uuid.UUID  `json:"user_id"`
	StartDate time.Time  `json:"start_date"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}
