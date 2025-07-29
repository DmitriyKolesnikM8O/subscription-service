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
		sub entity.Subscription,
	) (*entity.Subscription, error)

	GetSubscriptionByID(ctx context.Context, id uuid.UUID) (*entity.Subscription, error)
	UpdateSubscription(
		ctx context.Context,
		id uuid.UUID,
		sub entity.Subscription,
	) error
	DeleteSubscription(ctx context.Context, id uuid.UUID) error
	ListSubscriptionsByUser(ctx context.Context, userID uuid.UUID) ([]entity.Subscription, error)
	CalculateTotalCost(
		ctx context.Context,
		userID *uuid.UUID,
		serviceName *string,
		startDate, endDate time.Time,
	) (int, error)
}
