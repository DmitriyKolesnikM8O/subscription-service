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
	userID *uuid.UUID, // Меняем на указатель (может быть nil)
	serviceName *string, // Также делаем указатель
	startDate, endDate time.Time,
) (int, error) {
	qb := r.psql.
		Select("COALESCE(SUM(price), 0)").
		From("subscriptions").
		Where("start_date >= ?", startDate).
		Where("(end_date IS NULL OR end_date <= ?)", endDate)

	if userID != nil {
		qb = qb.Where("user_id = ?", *userID)
	}

	if serviceName != nil {
		qb = qb.Where("service_name = ?", *serviceName)
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
