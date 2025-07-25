package repo

import (
	"context"
	"time"

	"github.com/DmitriyKolesnikM8O/subscription-service/internal/repo/pgdb"

	"github.com/DmitriyKolesnikM8O/subscription-service/internal/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Subscription interface {
	CreateSubscription(ctx context.Context, sub entity.Subscription) (uuid.UUID, error)
	GetSubscriptionByID(ctx context.Context, id uuid.UUID) (entity.Subscription, error)
	UpdateSubscription(ctx context.Context, sub entity.Subscription) error
	DeleteSubscription(ctx context.Context, id uuid.UUID) error
	ListSubscriptions(ctx context.Context, userID uuid.UUID) ([]entity.Subscription, error)
}

type Report interface {
	GetTotalCost(ctx context.Context, userID *uuid.UUID, serviceName *string, startDate, endDate time.Time) (int, error)
}

type Repositories struct {
	Subscription
	Report
}

func NewRepositories(pg *pgxpool.Pool) *Repositories {
	return &Repositories{
		Subscription: pgdb.NewSubscriptionRepo(pg),
		Report:       pgdb.NewReportRepo(pg),
	}
}
