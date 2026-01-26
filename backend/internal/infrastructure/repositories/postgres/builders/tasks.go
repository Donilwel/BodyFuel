package builders

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"time"
)

type TasksFilterSpecification struct {
	IDs        []uuid.UUID
	TypeNms    []string
	ClusterNms []string
	Attempts   *int
	RetryAt    *time.Time
	States     []entities.TaskState
	//MetaStatus     *entities.MetaStatus
	ClusterIsReady *bool
}

func NewTasksFilterSpecification(f dto.TasksFilter) *TasksFilterSpecification {
	types := make([]string, len(f.Types))
	for i, t := range f.Types {
		types[i] = t.String()
	}
	return &TasksFilterSpecification{
		IDs:     f.IDs,
		TypeNms: types,
		//ClusterNms:     f.ClusterNms,
		Attempts: f.Attempts,
		States:   f.States,
		//MetaStatus:     f.MetaStatus,
		ClusterIsReady: f.ClusterIsReady,
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

	if len(s.ClusterNms) != 0 {
		predicates = append(predicates, sq.Eq{"t.task_cluster_nm": s.ClusterNms})
	}

	if len(s.States) != 0 {
		predicates = append(predicates, sq.Eq{"t.task_state": s.States})
	}

	if s.Attempts != nil {
		predicates = append(predicates, sq.Eq{"t.attempts": *s.Attempts})
	}

	if s.RetryAt != nil {
		predicates = append(predicates, sq.LtOrEq{"t.retry_at": s.RetryAt})
	}
	if s.ClusterIsReady != nil {
		predicates = append(predicates, sq.Eq{"cl.is_ready": *s.ClusterIsReady})
	}

	return predicates
}

type TasksSelectBuilder struct {
	b       sq.SelectBuilder
	orderBy []string
}

var tasksSelectBuilder = newQueryBuilder().Select(
	"task_id",
	"task_type_nm",
	"task_cluster_nm",
	"task_state",
	"attempts",
	"max_attempts",
	"retry_at",
	"created_at",
	"updated_at",
	"attribute",
	"cl.cluster_nm \"cluster.cluster_nm\"",
	"cl.execution_mode \"cluster.execution_mode\"",
	"cl.version \"cluster.version\"",
	"cl.host \"cluster.host\"",
	"cl.last_scrape_at \"cluster.last_scrape_at\"",
	"cl.is_ready \"cluster.is_ready\"",
	"cl.meta_status \"cluster.meta_status\"",
).From("raskolnikov2.tasks t").Join("raskolnikov2.clusters cl ON cl.cluster_nm = t.task_cluster_nm")

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
	b.b = b.b.Suffix("FOR UPDATE of t")

	return b
}

func (b *TasksSelectBuilder) ToSql() (string, []interface{}, error) {
	return b.b.ToSql()
}

type TasksDeleteBuilder struct {
	b sq.DeleteBuilder
}

func NewTasksDeleteBuilder() *TasksDeleteBuilder {
	deleteBuilder := newDeleteQueryBuilder().Delete("bodyfuel.tasks")

	return &TasksDeleteBuilder{b: deleteBuilder}
}

func (b *TasksDeleteBuilder) WithID(ids []uuid.UUID) *TasksDeleteBuilder {
	b.b = b.b.Where(sq.Eq{"task_id": ids})

	return b
}

func (b *TasksDeleteBuilder) ToSql() (string, []interface{}, error) {
	return b.b.ToSql()
}
