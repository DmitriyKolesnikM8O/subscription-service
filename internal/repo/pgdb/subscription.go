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

func (r *SubscriptionRepo) CreateSubscription(ctx context.Context, sub entity.Subscription) (uuid.UUID, error) {
	sql, args, err := r.psql.
		Insert("subscriptions").
		Columns("service_name", "price", "user_id", "start_date", "end_date").
		Values(sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return uuid.Nil, fmt.Errorf("SubscriptionRepo.CreateSubscription - sql build: %v", err)
	}

	var id uuid.UUID
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return uuid.Nil, repoerrs.ErrAlreadyExists
		}
		return uuid.Nil, fmt.Errorf("SubscriptionRepo.CreateSubscription - query exec: %v", err)
	}

	return id, nil
}

func (r *SubscriptionRepo) GetSubscriptionByID(ctx context.Context, id uuid.UUID) (entity.Subscription, error) {
	sql, args, err := r.psql.
		Select("*").
		From("subscriptions").
		Where("id = ?", id).
		ToSql()
	if err != nil {
		return entity.Subscription{}, fmt.Errorf("SubscriptionRepo.GetSubscriptionByID - sql build: %v", err)
	}

	var sub entity.Subscription
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&sub.StartDate,
		&sub.EndDate,
		&sub.CreatedAt,
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
	sql, args, err := r.psql.
		Update("subscriptions").
		Set("service_name", sub.ServiceName).
		Set("price", sub.Price).
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
		Select("*").
		From("subscriptions").
		Where("user_id = ?", userID).
		OrderBy("start_date DESC").
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
			&sub.ServiceName,
			&sub.Price,
			&sub.UserID,
			&sub.StartDate,
			&sub.EndDate,
			&sub.CreatedAt,
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
