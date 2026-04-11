package builders

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

type TasksFilterSpecification struct {
	IDs     []uuid.UUID
	TypeNms []string
	States  []entities.TaskState
	RetryAt *time.Time
	Attempts *int
}

func NewTasksFilterSpecification(f dto.TasksFilter) *TasksFilterSpecification {
	types := make([]string, len(f.Types))
	for i, t := range f.Types {
		types[i] = t.String()
	}

	return &TasksFilterSpecification{
		IDs:      f.IDs,
		TypeNms:  types,
		States:   f.States,
		RetryAt:  f.RetryAt,
		Attempts: f.Attempts,
	}
}

func (s *TasksFilterSpecification) Predicates() []sq.Sqlizer {
	var predicates []sq.Sqlizer

	if len(s.IDs) != 0 {
		predicates = append(predicates, sq.Eq{"t.task_id": s.IDs})
	}

	if len(s.TypeNms) != 0 {
		predicates = append(predicates, sq.Eq{"t.task_type_nm": s.TypeNms})
	}

	if len(s.States) != 0 {
		predicates = append(predicates, sq.Eq{"t.task_state": s.States})
	}

	if s.Attempts != nil {
		predicates = append(predicates, sq.Eq{"t.attempts": *s.Attempts})
	}

	if s.RetryAt != nil {
		predicates = append(predicates, sq.LtOrEq{"t.retry_at": *s.RetryAt})
	}

	return predicates
}

type TasksSelectBuilder struct {
	b       sq.SelectBuilder
	orderBy []string
}

var tasksSelectBuilder = newQueryBuilder().Select(
	"t.task_id",
	"t.task_type_nm",
	"t.task_state",
	"t.attempts",
	"t.max_attempts",
	"t.retry_at",
	"t.created_at",
	"t.updated_at",
	"t.attribute",
).From("bodyfuel.tasks t")

func NewTasksSelectBuilder() *TasksSelectBuilder {
	return &TasksSelectBuilder{b: tasksSelectBuilder}
}

func (b *TasksSelectBuilder) OrderByTyped(conds ...OrderBy) *TasksSelectBuilder {
	for _, cond := range conds {
		b.orderBy = append(b.orderBy, fmt.Sprintf("%s %s", cond.Field, cond.Direction))
	}
	return b
}

func (b *TasksSelectBuilder) WithFilterSpecification(s *TasksFilterSpecification) *TasksSelectBuilder {
	b.b = ApplyFilter(b.b, s)
	return b
}

func (b *TasksSelectBuilder) Limit(limit int) *TasksSelectBuilder {
	if limit > 0 {
		b.b = b.b.Limit(uint64(limit))
	}
	return b
}

func (b *TasksSelectBuilder) Offset(offset int) *TasksSelectBuilder {
	if offset > 0 {
		b.b = b.b.Offset(uint64(offset))
	}
	return b
}

func (b *TasksSelectBuilder) WithBlock() *TasksSelectBuilder {
	b.b = b.b.Suffix("FOR UPDATE OF t SKIP LOCKED")
	return b
}

func (b *TasksSelectBuilder) ToSql() (string, []interface{}, error) {
	return b.b.ToSql()
}

type TasksDeleteBuilder struct {
	b sq.DeleteBuilder
}

func NewTasksDeleteBuilder() *TasksDeleteBuilder {
	return &TasksDeleteBuilder{
		b: newDeleteQueryBuilder().Delete("bodyfuel.tasks"),
	}
}

func (b *TasksDeleteBuilder) WithID(ids []uuid.UUID) *TasksDeleteBuilder {
	b.b = b.b.Where(sq.Eq{"task_id": ids})
	return b
}

func (b *TasksDeleteBuilder) ToSql() (string, []interface{}, error) {
	return b.b.ToSql()
}
