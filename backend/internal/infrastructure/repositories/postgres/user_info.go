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
	queryCreateUserInfo = `INSERT INTO bodyfuel.user_info (
                                   "id",
                                    "username",
                                    "name",
                                    "surname",
                                    "password",
                                    "email",
                                    "phone",
                                    "created_at") VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	queryUpdateUserInfo = `UPDATE bodyfuel.user_info SET
									username=:username,
									name=:name,
									surname=:surname,
									password=:password,
									email=:email,
									phone=:phone,
									created_at=:created_at
									WHERE id=:id`
)

type UserInfoRepo struct {
	getter dbClientGetter
}

func NewUserInfoRepository(db *sqlx.DB) *UserInfoRepo {
	return &UserInfoRepo{getter: dbClientGetter{db: db}}
}

func (r *UserInfoRepo) Get(ctx context.Context, f dto.UserInfoFilter, withBlock bool) (*entities.UserInfo, error) {
	selectBuilder := builders.NewUserInfoSelectBuilder().WithFilterSpecification(builders.NewUserInfoFilterSpecification(f))
	if withBlock {
		selectBuilder = selectBuilder.WithBlock()
	}
	query, args, err := selectBuilder.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	var row models.UserInfo
	if err := r.getter.Get(ctx).GetContext(ctx, &row, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrUserInfoNotFound
		}
		return nil, fmt.Errorf("get context: %w", err)
	}

	return row.ToEntity(), nil
}

func (r *UserInfoRepo) Create(ctx context.Context, userInfo *entities.UserInfo) error {
	row := models.NewUserInfoRow(userInfo)

	_, err := r.getter.Get(ctx).ExecContext(ctx, queryCreateUserInfo,
		row.ID,
		row.Username,
		row.Name,
		row.Surname,
		row.Password,
		row.Email,
		row.Phone,
		row.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("exec context: %w", err)
	}
	return nil
}

func (r *UserInfoRepo) Update(ctx context.Context, userInfo *entities.UserInfo) error {
	row := models.NewUserInfoRow(userInfo)

	res, err := r.getter.Get(ctx).NamedExecContext(ctx, queryUpdateUserInfo, row)
	if err != nil {
		return fmt.Errorf("exec context: %w", err)
	}

	rowAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", errs.ErrUserInfoNotFound)
	}

	if rowAffected == 0 {
		return fmt.Errorf("rows affected: %w", err)
	}

	return nil
}

func (r *UserInfoRepo) Delete(ctx context.Context, f dto.UserInfoFilter) error {
	deleteBuilder := builders.NewUserInfoDeleteBuilder().WithFilterSpecification(builders.NewUserInfoFilterSpecification(f))
	query, args, err := deleteBuilder.ToSQL()
	if err != nil {
		return fmt.Errorf("build sql: %w", err)
	}

	if _, err = r.getter.Get(ctx).ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("exec context: %w", err)
	}

	return nil
}