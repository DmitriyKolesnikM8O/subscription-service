package service

import (
	"context"
	"time"

	"github.com/DmitriyKolesnikM8O/subscription-service/internal/entity"
	"github.com/google/uuid"
)

//go:generate mockgen -source=service.go -destination=mocks/service.go -package=mocks
type SubscriptionService interface {
	CreateSubscription(
		ctx context.Context,
		serviceName string,
		price int,
		userID uuid.UUID,
		startDate time.Time,
		endDate *time.Time,
	) (*entity.Subscription, error)

	GetSubscriptionByID(ctx context.Context, id uuid.UUID) (*entity.Subscription, error)
	UpdateSubscription(
		ctx context.Context,
		id uuid.UUID,
		serviceName string,
		price int,
		endDate *time.Time,
	) error
	DeleteSubscription(ctx context.Context, id uuid.UUID) error
	ListSubscriptionsByUser(ctx context.Context, userID uuid.UUID) ([]entity.Subscription, error)
	CalculateTotalCost(
		ctx context.Context,
		userID uuid.UUID,
		serviceName string,
		startDate, endDate time.Time,
	) (int, error)
}
