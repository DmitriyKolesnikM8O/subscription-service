package pgdb

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)

type ReportRepo struct {
	pool *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewReportRepo(pg *pgxpool.Pool) *ReportRepo {
	return &ReportRepo{
		pool: pg,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *ReportRepo) GetTotalCost(
	ctx context.Context,
	userID *uuid.UUID,
	serviceName *string,
	startDate, endDate time.Time,
) ([]struct {
	Price     int
	StartDate time.Time
	EndDate   *time.Time
}, error) {
	qb := r.psql.
		Select("svc.price", "s.start_date", "s.end_date").
		From("subscriptions s").
		Join("services svc ON s.service_id = svc.id").
		Where("s.start_date <= ?", endDate).
		Where("(s.end_date IS NULL OR s.end_date >= ?)", startDate)

	if userID != nil {
		qb = qb.Where("s.user_id = ?", *userID)
	}

	if serviceName != nil {
		qb = qb.Where("svc.name = ?", *serviceName)
	}

	sql, args, err := qb.ToSql()
	if err != nil {
		return nil, fmt.Errorf("ReportRepo.GetTotalCost - sql build: %w", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("ReportRepo.GetTotalCostData - query exec: %w", err)
	}
	defer rows.Close()

	var costData []struct {
		Price     int
		StartDate time.Time
		EndDate   *time.Time
	}
	for rows.Next() {
		var data struct {
			Price     int
			StartDate time.Time
			EndDate   *time.Time
		}
		if err := rows.Scan(&data.Price, &data.StartDate, &data.EndDate); err != nil {
			return nil, fmt.Errorf("ReportRepo.GetTotalCostData - row scan: %w", err)
		}
		costData = append(costData, data)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ReportRepo.GetTotalCostData - rows error: %w", err)
	}

	return costData, nil
}
