package postgres

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	errs "backend/internal/errors"
	"backend/internal/infrastructure/repositories/postgres/builders"
	"backend/internal/infrastructure/repositories/postgres/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
)

const (
	queryCreateUserWeight = `INSERT INTO bodyfuel.user_weight (
                                   "id",
                                   "id_user",
                                    "weight",
                                    "date") VALUES ($1, $2, $3, $4)`

	queryUpdateUserWeight = `UPDATE bodyfuel.user_weight SET
									"id_user" = :id_user,
									"weight" = :weight,
									"date" = :date
									WHERE id=:id`
)

type UserWeightRepo struct {
	getter dbClientGetter
}

func NewUserWeightRepository(db *sqlx.DB) *UserWeightRepo {
	return &UserWeightRepo{getter: dbClientGetter{db: db}}
}

func (r *UserWeightRepo) Get(ctx context.Context, f dto.UserWeightFilter, withBlock bool) (*entities.UserWeight, error) {
	selectBuilder := builders.NewUserWeightSelectBuilder().WithFilterSpecification(builders.NewUserWeightFilterSpecification(f))
	if withBlock {
		selectBuilder = selectBuilder.WithBlock()
	}
	query, args, err := selectBuilder.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	var row models.UserWeightRow
	if err := r.getter.Get(ctx).GetContext(ctx, &row, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrUserWeightNotFound
		}
		return nil, fmt.Errorf("get context: %w", err)
	}

	return row.ToEntity(), nil
}

func (r *UserWeightRepo) List(ctx context.Context, f dto.UserWeightFilter, withBlock bool) ([]*entities.UserWeight, error) {
	var rows []*models.UserWeightRow

	selectBuilder := builders.NewUserWeightSelectBuilder()
	if withBlock {
		selectBuilder = selectBuilder.WithBlock()
	}

	query, args, err := selectBuilder.WithFilterSpecification(builders.NewUserWeightFilterSpecification(f)).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	if err = r.getter.Get(ctx).SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("select context: %w", err)
	}
	result := make([]*entities.UserWeight, len(rows))
	for i := range rows {
		result[i] = rows[i].ToEntity()
	}

	return result, nil
}

func (r *UserWeightRepo) Create(ctx context.Context, userWeight *entities.UserWeight) error {
	row := models.NewUserWeightRow(userWeight)

	_, err := r.getter.Get(ctx).ExecContext(ctx, queryCreateUserWeight,
		row.ID,
		row.UserId,
		row.Weight,
		row.Date,
	)
	if err != nil {
		return fmt.Errorf("exec context: %w", err)
	}
	return nil
}

func (r *UserWeightRepo) Update(ctx context.Context, userWeight *entities.UserWeight) error {
	row := models.NewUserWeightRow(userWeight)

	res, err := r.getter.Get(ctx).NamedExecContext(ctx, queryUpdateUserWeight, row)
	if err != nil {
		return fmt.Errorf("exec context: %w", err)
	}

	rowAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", errs.ErrUserWeightNotFound)
	}

	if rowAffected == 0 {
		return fmt.Errorf("rows affected: %w", err)
	}

	return nil
}

func (r *UserWeightRepo) Delete(ctx context.Context, f dto.UserWeightFilter) error {
	deleteBuilder := builders.NewUserWeightDeleteBuilder().WithFilterSpecification(builders.NewUserWeightFilterSpecification(f))
	query, args, err := deleteBuilder.ToSQL()
	if err != nil {
		return fmt.Errorf("build sql: %w", err)
	}

	result, err := r.getter.Get(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec context: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no rows deleted: %w", errs.ErrUserWeightAlreadyDeleted)
	}

	return nil
}
