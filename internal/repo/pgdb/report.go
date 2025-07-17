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

func (r *ReportRepo) GetTotalCost(ctx context.Context, userID uuid.UUID, serviceName string, startDate, endDate time.Time) (int, error) {
	sql, args, err := r.psql.
		Select("COALESCE(SUM(price), 0)").
		From("subscriptions").
		Where("user_id = ?", userID).
		Where("start_date >= ?", startDate).
		Where("(end_date IS NULL OR end_date <= ?)", endDate).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("ReportRepo.GetTotalCost: %v", err)
	}

	var total int
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("ReportRepo.GetTotalCost: %v", err)
	}

	return total, nil
}
