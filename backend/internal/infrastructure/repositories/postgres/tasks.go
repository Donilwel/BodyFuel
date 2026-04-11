package postgres

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"backend/internal/infrastructure/repositories/postgres/builders"
	"backend/internal/infrastructure/repositories/postgres/models"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const (
	queryTaskCreate = `INSERT INTO bodyfuel.tasks (
		task_id, task_type_nm, task_state, max_attempts, attempts,
		retry_at, created_at, updated_at, attribute
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	queryTaskUpdate = `UPDATE bodyfuel.tasks SET
		task_type_nm = :task_type_nm,
		task_state   = :task_state,
		max_attempts = :max_attempts,
		attempts     = :attempts,
		retry_at     = :retry_at,
		updated_at   = :updated_at,
		attribute    = :attribute
		WHERE task_id = :task_id`
)

type TasksRepo struct {
	getter dbClientGetter
}

func NewTasksRepository(db *sqlx.DB) *TasksRepo {
	return &TasksRepo{getter: dbClientGetter{db: db}}
}

func (r *TasksRepo) Create(ctx context.Context, task *entities.Task) error {
	row, err := models.NewTaskRow(task)
	if err != nil {
		return fmt.Errorf("new task row: %w", err)
	}

	_, err = r.getter.Get(ctx).ExecContext(ctx, queryTaskCreate,
		row.UUID,
		row.TypeNm,
		row.State,
		row.MaxAttempts,
		row.Attempts,
		row.RetryAt,
		row.CreatedAt,
		row.UpdatedAt,
		row.Attribute,
	)
	if err != nil {
		return fmt.Errorf("exec context: %w", err)
	}

	return nil
}

func (r *TasksRepo) Get(ctx context.Context, f dto.TasksFilter, withBlock bool) (*entities.Task, error) {
	b := builders.NewTasksSelectBuilder().
		WithFilterSpecification(builders.NewTasksFilterSpecification(f)).
		Limit(1)

	if withBlock {
		b = b.WithBlock()
	}

	query, args, err := b.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	var row models.TaskRow
	if err = r.getter.Get(ctx).GetContext(ctx, &row, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("get context: %w", err)
	}

	return row.ToEntity()
}

func (r *TasksRepo) List(ctx context.Context, f dto.TasksFilter, withBlock bool) ([]*entities.Task, error) {
	b := builders.NewTasksSelectBuilder().
		WithFilterSpecification(builders.NewTasksFilterSpecification(f))

	if withBlock {
		b = b.WithBlock()
	}

	query, args, err := b.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	var rows []models.TaskRow
	if err = r.getter.Get(ctx).SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("select context: %w", err)
	}

	result := make([]*entities.Task, len(rows))
	for i := range rows {
		e, err := rows[i].ToEntity()
		if err != nil {
			return nil, fmt.Errorf("to entity: %w", err)
		}
		result[i] = e
	}

	return result, nil
}

func (r *TasksRepo) Update(ctx context.Context, task *entities.Task) error {
	row, err := models.NewTaskRow(task)
	if err != nil {
		return fmt.Errorf("new task row: %w", err)
	}

	res, err := r.getter.Get(ctx).NamedExecContext(ctx, queryTaskUpdate, row)
	if err != nil {
		return fmt.Errorf("named exec context: %w", err)
	}

	ar, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if ar == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

func (r *TasksRepo) Delete(ctx context.Context, ids []uuid.UUID) error {
	query, args, err := builders.NewTasksDeleteBuilder().WithID(ids).ToSql()
	if err != nil {
		return fmt.Errorf("build sql: %w", err)
	}

	if _, err = r.getter.Get(ctx).ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("exec context: %w", err)
	}

	return nil
}
