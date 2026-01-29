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
	queryCreateUserParams = `INSERT INTO bodyfuel.user_params (
                                   "id",
                                   "id_user",
                                    "height",
                                    "photo",
                                    "wants",
                                    "target_workouts_weeks",
                                    "target_calories_daily",
                                    "target_weight",
                                    "lifestyle") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	queryUpdateUserParams = `UPDATE bodyfuel.user_params SET
									"id_user" = :id_user,
									"height" = :height,
									"photo" = :photo,
									"wants" = :wants,
									"target_workouts_weeks" = :target_workouts_weeks,
                                    "target_calories_daily" = :target_calories_daily,
                                    "target_weight" = :target_weight,
									"lifestyle" = :lifestyle
									WHERE id=:id`
)

type UserParamsRepo struct {
	getter dbClientGetter
}

func NewUserParamsRepository(db *sqlx.DB) *UserParamsRepo {
	return &UserParamsRepo{getter: dbClientGetter{db: db}}
}

func (r *UserParamsRepo) Get(ctx context.Context, f dto.UserParamsFilter, withBlock bool) (*entities.UserParams, error) {
	selectBuilder := builders.NewUserParamsSelectBuilder().WithFilterSpecification(builders.NewUserParamsFilterSpecification(f))
	if withBlock {
		selectBuilder = selectBuilder.WithBlock()
	}
	query, args, err := selectBuilder.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	var row models.UserParamsRow
	if err := r.getter.Get(ctx).GetContext(ctx, &row, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrUserParamsNotFound
		}
		return nil, fmt.Errorf("get context: %w", err)
	}

	return row.ToEntity(), nil
}

func (r *UserParamsRepo) List(ctx context.Context, f dto.UserParamsFilter, withBlock bool) ([]*entities.UserParams, error) {
	var rows []*models.UserParamsRow

	selectBuilder := builders.NewUserParamsSelectBuilder()
	if withBlock {
		selectBuilder = selectBuilder.WithBlock()
	}

	query, args, err := selectBuilder.WithFilterSpecification(builders.NewUserParamsFilterSpecification(f)).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	if err = r.getter.Get(ctx).SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("select context: %w", err)
	}
	result := make([]*entities.UserParams, len(rows))
	for i := range rows {
		result[i] = rows[i].ToEntity()
	}

	return result, nil
}

func (r *UserParamsRepo) Create(ctx context.Context, userParams *entities.UserParams) error {
	row := models.NewUserParamsRow(userParams)

	_, err := r.getter.Get(ctx).ExecContext(ctx, queryCreateUserParams,
		row.ID,
		row.UserId,
		row.Height,
		row.Photo,
		row.Wants,
		row.TargetWorkoutsWeeks,
		row.TargetCaloriesDaily,
		row.TargetWeight,
		row.Lifestyle,
	)
	if err != nil {
		return fmt.Errorf("exec context: %w", err)
	}
	return nil
}

func (r *UserParamsRepo) Update(ctx context.Context, userParams *entities.UserParams) error {
	row := models.NewUserParamsRow(userParams)

	res, err := r.getter.Get(ctx).NamedExecContext(ctx, queryUpdateUserParams, row)
	if err != nil {
		return fmt.Errorf("exec context: %w", err)
	}

	rowAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", errs.ErrUserParamsNotFound)
	}

	if rowAffected == 0 {
		return fmt.Errorf("rows affected: %w", err)
	}

	return nil
}

func (r *UserParamsRepo) Delete(ctx context.Context, f dto.UserParamsFilter) error {
	deleteBuilder := builders.NewUserParamsDeleteBuilder().WithFilterSpecification(builders.NewUserParamsFilterSpecification(f))
	query, args, err := deleteBuilder.ToSQL()
	if err != nil {
		return fmt.Errorf("build sql: %w", err)
	}
	fmt.Println(query)

	result, err := r.getter.Get(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec context: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no rows deleted: %w", errs.ErrUserParamsAlreadyDeleted)
	}

	return nil
}
