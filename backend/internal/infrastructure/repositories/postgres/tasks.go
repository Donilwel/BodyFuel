package postgres

//
//const (
//	queryTaskCUpdate = `UPDATE bodyfuel.tasks SET
//		task_type_nm=:task_type_nm,
//		task_state=:task_state,
//		task_cluster_nm=:task_cluster_nm,
//		"max_attempts"=:max_attempts,
//		"attempts"=:attempts,
//		"retry_at"=:retry_at,
//		"created_at"=:created_at,
//		"updated_at"=:updated_at,
//		"attribute"=:attribute
//		WHERE task_id=:task_id`
//
//	queryTaskCreate = `INSERT INTO bodyfuel.tasks (
//		task_id,
//		task_type_nm,
//		task_state,
//		task_cluster_nm,
//		"max_attempts",
//		"attempts",
//		"retry_at",
//		"created_at",
//		"updated_at",
//		"attribute"
//		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
//)
//
//type TasksRepo struct {
//	getter dbClientGetter
//}
//
//func NewTasksRepository(db *sqlx.DB) *TasksRepo {
//	return &TasksRepo{getter: dbClientGetter{db: db}}
//}
//
//func (r *TasksRepo) Get(ctx context.Context, f dto.TasksFilter, withBlock bool) (*entities.Task, error) {
//	b := builders.NewTasksSelectBuilder().
//		WithFilterSpecification(builders.NewTasksFilterSpecification(f))
//
//	if withBlock {
//		b = b.WithBlock()
//	}
//
//	query, args, err := b.ToSql()
//	if err != nil {
//		return nil, fmt.Errorf("build sql: %w", err)
//	}
//
//	var row models.TaskRow
//	if err = r.getter.Get(ctx).GetContext(ctx, &row, query, args...); err != nil {
//		if errors.Is(err, sql.ErrNoRows) {
//			return nil, errs.ErrTaskNotFound().WithMetadata(structToParams(f))
//		}
//
//		return nil, fmt.Errorf("get context: %w", err)
//	}
//
//	e, err := row.ToEntity()
//	if err != nil {
//		return nil, fmt.Errorf("to entity: %w", err)
//	}
//
//	return e, nil
//}
//
//func (r *TasksRepo) List(ctx context.Context, f dto.TasksFilter, withBlock bool) ([]*entities.Task, error) {
//	b := builders.NewTasksSelectBuilder().
//		WithFilterSpecification(builders.NewTasksFilterSpecification(f))
//
//	if withBlock {
//		b = b.WithBlock()
//	}
//
//	query, args, err := b.ToSql()
//	if err != nil {
//		return nil, fmt.Errorf("build sql: %w", err)
//	}
//
//	var rows []models.TaskRow
//
//	if err = r.getter.Get(ctx).SelectContext(ctx, &rows, query, args...); err != nil {
//		return nil, fmt.Errorf("select context: %w", err)
//	}
//
//	result := make([]*entities.Task, len(rows))
//	for i := range rows {
//		e, err := rows[i].ToEntity()
//		if err != nil {
//			return nil, fmt.Errorf("to entity: %w", err)
//		}
//
//		result[i] = e
//	}
//
//	return result, nil
//}
//
//func (r *TasksRepo) Create(ctx context.Context, t *entities.Task) error {
//	row, err := models.NewTaskRow(t)
//	if err != nil {
//		return fmt.Errorf("new task row: %w", err)
//	}
//
//	res, err := r.getter.Get(ctx).ExecContext(ctx, queryTaskCreate,
//		row.UUID,
//		row.TypeNm,
//		row.State,
//		row.ClusterNm,
//		row.MaxAttempts,
//		row.Attempts,
//		row.RetryAt,
//		row.CreatedAt,
//		row.UpdatedAt,
//		row.Attribute,
//	)
//	if err != nil {
//		return fmt.Errorf("exec context: %w", err)
//	}
//
//	ra, err := res.RowsAffected()
//	if err != nil {
//		return fmt.Errorf("rows affected: %w", err)
//	}
//
//	if ra == 0 {
//		return errs.ErrTasksNoRowsAffected().WithMetadata(structToParams(t))
//	}
//
//	return nil
//}
//
//func (r *TasksRepo) Update(ctx context.Context, t *entities.Task) error {
//	row, err := models.NewTaskRow(t)
//	if err != nil {
//		return fmt.Errorf("new task row: %w", err)
//	}
//
//	res, err := r.getter.Get(ctx).
//		NamedExecContext(ctx, queryTaskCUpdate, row)
//	if err != nil {
//		return fmt.Errorf("named exec context: %w", err)
//	}
//
//	ar, err := res.RowsAffected()
//	if err != nil {
//		return errs.ErrTaskNotFound().WithMetadata(structToParams(t))
//	}
//
//	if ar == 0 {
//		return errs.ErrTasksNoRowsAffected().WithMetadata(structToParams(t))
//	}
//
//	return nil
//}
//
//func (r *TasksRepo) Delete(ctx context.Context, ids []uuid.UUID) error {
//	query, args, err := builders.NewTasksDeleteBuilder().WithID(ids).ToSql()
//	if err != nil {
//		return fmt.Errorf("build sql: %w", err)
//	}
//
//	if _, err = r.getter.Get(ctx).ExecContext(ctx, query, args...); err != nil {
//		return fmt.Errorf("exec context: %w", err)
//	}
//
//	return nil
//}
