package postgres

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"backend/internal/infrastructure/repositories/postgres/models"
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var psq = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

const (
	queryUpsertUserDevice = `INSERT INTO bodyfuel.user_devices (id, user_id, device_token, platform, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, device_token) DO UPDATE SET
			platform   = EXCLUDED.platform,
			updated_at = EXCLUDED.updated_at`

	queryDeleteUserDevice = `DELETE FROM bodyfuel.user_devices WHERE id = $1 AND user_id = $2`
)

type UserDevicesRepo struct {
	getter dbClientGetter
}

func NewUserDevicesRepository(db *sqlx.DB) *UserDevicesRepo {
	return &UserDevicesRepo{getter: dbClientGetter{db: db}}
}

func (r *UserDevicesRepo) Upsert(ctx context.Context, device *entities.UserDevice) error {
	row := models.NewUserDeviceRow(device)

	_, err := r.getter.Get(ctx).ExecContext(ctx, queryUpsertUserDevice,
		row.ID,
		row.UserID,
		row.DeviceToken,
		row.Platform,
		row.CreatedAt,
		row.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("exec context: %w", err)
	}

	return nil
}

func (r *UserDevicesRepo) List(ctx context.Context, f dto.UserDeviceFilter) ([]*entities.UserDevice, error) {
	q := psq.Select(
		"id", "user_id", "device_token", "platform", "created_at", "updated_at",
	).From("bodyfuel.user_devices")

	if f.UserID != nil {
		q = q.Where(sq.Eq{"user_id": *f.UserID})
	}
	if f.ID != nil {
		q = q.Where(sq.Eq{"id": *f.ID})
	}
	if f.DeviceToken != nil {
		q = q.Where(sq.Eq{"device_token": *f.DeviceToken})
	}

	query, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	var rows []models.UserDeviceRow
	if err = r.getter.Get(ctx).SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("select context: %w", err)
	}

	result := make([]*entities.UserDevice, len(rows))
	for i := range rows {
		result[i] = rows[i].ToEntity()
	}

	return result, nil
}

func (r *UserDevicesRepo) Delete(ctx context.Context, id, userID uuid.UUID) error {
	_, err := r.getter.Get(ctx).ExecContext(ctx, queryDeleteUserDevice, id, userID)
	if err != nil {
		return fmt.Errorf("exec context: %w", err)
	}

	return nil
}
