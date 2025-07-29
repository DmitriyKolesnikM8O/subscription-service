package pgdb

import (
	"context"
	"errors"
	"fmt"

	"github.com/DmitriyKolesnikM8O/subscription-service/internal/entity"
	"github.com/DmitriyKolesnikM8O/subscription-service/internal/repo/repoerrs"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type SubscriptionRepo struct {
	pool *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewSubscriptionRepo(pg *pgxpool.Pool) *SubscriptionRepo {
	return &SubscriptionRepo{
		pool: pg,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *SubscriptionRepo) CreateSubscription(ctx context.Context, sub entity.Subscription) (*entity.Subscription, error) {
	var serviceID uuid.UUID
	err := r.pool.QueryRow(ctx, "SELECT id FROM services WHERE name = $1 AND price = $2", sub.Service.Name, sub.Service.Price).Scan(&serviceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			sql, args, err := r.psql.
				Insert("services").
				Columns("name", "price").
				Values(sub.Service.Name, sub.Service.Price).
				Suffix("RETURNING id").
				ToSql()
			if err != nil {
				return nil, fmt.Errorf("SubscriptionRepo.CreateSubscription - service sql build: %v", err)
			}
			err = r.pool.QueryRow(ctx, sql, args...).Scan(&serviceID)
			if err != nil {
				return nil, fmt.Errorf("SubscriptionRepo.CreateSubscription - service query exec: %v", err)
			}
		} else {
			return nil, fmt.Errorf("SubscriptionRepo.CreateSubscription - service query: %v", err)
		}
	}

	sub.Service.ID = serviceID

	sql, args, err := r.psql.
		Insert("subscriptions").
		Columns("id", "service_id", "user_id", "start_date", "end_date", "created_at").
		Values(uuid.New(), serviceID, sub.UserID, sub.StartDate, sub.EndDate, "NOW()").
		Suffix("RETURNING id, created_at").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("SubscriptionRepo.CreateSubscription - subscription sql build: %v", err)
	}

	err = r.pool.QueryRow(ctx, sql, args...).Scan(&sub.ID, &sub.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, repoerrs.ErrAlreadyExists
		}
		return nil, fmt.Errorf("SubscriptionRepo.CreateSubscription - subscription query exec: %v", err)
	}

	return &sub, nil
}
func (r *SubscriptionRepo) GetSubscriptionByID(ctx context.Context, id uuid.UUID) (entity.Subscription, error) {
	sql, args, err := r.psql.
		Select("s.id", "s.user_id", "s.start_date", "s.end_date", "s.created_at", "svc.id", "svc.name", "svc.price").
		From("subscriptions s").
		Join("services svc ON s.service_id = svc.id").
		Where("s.id = ?", id).
		ToSql()
	if err != nil {
		return entity.Subscription{}, fmt.Errorf("SubscriptionRepo.GetSubscriptionByID - sql build: %v", err)
	}

	var sub entity.Subscription
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&sub.ID,
		&sub.UserID,
		&sub.StartDate,
		&sub.EndDate,
		&sub.CreatedAt,
		&sub.Service.ID,
		&sub.Service.Name,
		&sub.Service.Price,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Subscription{}, repoerrs.ErrNotFound
		}
		return entity.Subscription{}, fmt.Errorf("SubscriptionRepo.GetSubscriptionByID - query exec: %v", err)
	}

	return sub, nil
}

func (r *SubscriptionRepo) UpdateSubscription(ctx context.Context, sub entity.Subscription) error {
	// First, try to find an existing service with the same name and price.
	// If not found, create a new service.
	var serviceID uuid.UUID
	err := r.pool.QueryRow(ctx, "SELECT id FROM services WHERE name = $1 AND price = $2", sub.Service.Name, sub.Service.Price).Scan(&serviceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Service not found, create a new one
			sql, args, err := r.psql.
				Insert("services").
				Columns("name", "price").
				Values(sub.Service.Name, sub.Service.Price).
				Suffix("RETURNING id").
				ToSql()
			if err != nil {
				return fmt.Errorf("SubscriptionRepo.UpdateSubscription - service sql build: %v", err)
			}
			err = r.pool.QueryRow(ctx, sql, args...).Scan(&serviceID)
			if err != nil {
				return fmt.Errorf("SubscriptionRepo.UpdateSubscription - service query exec: %v", err)
			}
		} else {
			return fmt.Errorf("SubscriptionRepo.UpdateSubscription - service query: %v", err)
		}
	}

	sql, args, err := r.psql.
		Update("subscriptions").
		Set("service_id", serviceID).
		Set("user_id", sub.UserID).
		Set("start_date", sub.StartDate).
		Set("end_date", sub.EndDate).
		Where("id = ?", sub.ID).
		ToSql()
	if err != nil {
		return fmt.Errorf("SubscriptionRepo.UpdateSubscription - sql build: %v", err)
	}

	result, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("SubscriptionRepo.UpdateSubscription - query exec: %v", err)
	}

	if result.RowsAffected() == 0 {
		return repoerrs.ErrNotFound
	}

	return nil
}

func (r *SubscriptionRepo) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	sql, args, err := r.psql.
		Delete("subscriptions").
		Where("id = ?", id).
		ToSql()
	if err != nil {
		return fmt.Errorf("SubscriptionRepo.DeleteSubscription - sql build: %v", err)
	}

	result, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("SubscriptionRepo.DeleteSubscription - query exec: %v", err)
	}

	if result.RowsAffected() == 0 {
		return repoerrs.ErrNotFound
	}

	return nil
}

func (r *SubscriptionRepo) ListSubscriptions(ctx context.Context, userID uuid.UUID) ([]entity.Subscription, error) {
	sql, args, err := r.psql.
		Select("s.id", "s.user_id", "s.start_date", "s.end_date", "s.created_at", "svc.id", "svc.name", "svc.price").
		From("subscriptions s").
		Join("services svc ON s.service_id = svc.id").
		Where("s.user_id = ?", userID).
		OrderBy("s.start_date DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("SubscriptionRepo.ListSubscriptions - sql build: %v", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("SubscriptionRepo.ListSubscriptions - query exec: %v", err)
	}
	defer rows.Close()

	var subscriptions []entity.Subscription
	for rows.Next() {
		var sub entity.Subscription
		if err := rows.Scan(
			&sub.ID,
			&sub.UserID,
			&sub.StartDate,
			&sub.EndDate,
			&sub.CreatedAt,
			&sub.Service.ID,
			&sub.Service.Name,
			&sub.Service.Price,
		); err != nil {
			return nil, fmt.Errorf("SubscriptionRepo.ListSubscriptions - row scan: %v", err)
		}
		subscriptions = append(subscriptions, sub)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("SubscriptionRepo.ListSubscriptions - rows error: %v", err)
	}

	return subscriptions, nil
}
