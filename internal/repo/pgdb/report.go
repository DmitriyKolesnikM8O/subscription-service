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
) (int, error) {
	qb := r.psql.
		Select("COALESCE(SUM(svc.price), 0)").
		From("subscriptions s").
		Join("services svc ON s.service_id = svc.id").
		Where("s.start_date >= ?", startDate).
		Where("(s.end_date IS NULL OR s.end_date <= ?)", endDate)

	if userID != nil {
		qb = qb.Where("s.user_id = ?", *userID)
	}

	if serviceName != nil {
		qb = qb.Where("svc.name = ?", *serviceName)
	}

	sql, args, err := qb.ToSql()
	if err != nil {
		return 0, fmt.Errorf("ReportRepo.GetTotalCost - sql build: %w", err)
	}

	var total int
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("ReportRepo.GetTotalCost - query exec: %w", err)
	}

	return total, nil
}
