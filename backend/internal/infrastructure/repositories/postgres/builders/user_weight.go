package builders

import (
	"backend/internal/dto"
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"time"
)

const (
	userWeightTable = "bodyfuel.user_weight"
)

type UserWeightFilterSpecification struct {
	ID        *uuid.UUID
	UserID    *uuid.UUID
	Weight    *float64
	CreatedAt *time.Time
}

func NewUserWeightFilterSpecification(f dto.UserWeightFilter) *UserWeightFilterSpecification {
	s := &UserWeightFilterSpecification{
		ID:        f.ID,
		UserID:    f.UserID,
		Weight:    f.Weight,
		CreatedAt: f.CreatedAt,
	}

	return s
}

func (spec *UserWeightFilterSpecification) Predicates() []sq.Sqlizer {
	var predicates []sq.Sqlizer

	if v := spec.ID; v != nil {
		predicates = append(predicates, sq.Eq{"user_weight.id": v})
	}

	if v := spec.UserID; v != nil {
		predicates = append(predicates, sq.Eq{"user_weight.id_user": v})
	}

	if v := spec.Weight; v != nil {
		predicates = append(predicates, sq.Eq{"user_weight.weight": v})
	}

	if v := spec.CreatedAt; v != nil {
		predicates = append(predicates, sq.Eq{"user_weight.date": v})
	}

	return predicates
}

type UserWeightSelectBuilder struct {
	b sq.SelectBuilder
}

func NewUserWeightSelectBuilder() *UserWeightSelectBuilder {
	selectBuilder := newQueryBuilder().Select(
		"user_weight.id",
		"user_weight.id_user",
		"user_weight.weight",
		"user_weight.date",
	).From(userWeightTable)

	return &UserWeightSelectBuilder{b: selectBuilder}
}

func (a *UserWeightSelectBuilder) WithFilterSpecification(spec *UserWeightFilterSpecification) *UserWeightSelectBuilder {
	a.b = ApplyFilter(a.b, spec)

	return a
}

func (a *UserWeightSelectBuilder) Limit(limit int) *UserWeightSelectBuilder {
	if limit > 0 {
		a.b = a.b.Limit(uint64(limit))
	}

	return a
}

func (a *UserWeightSelectBuilder) Offset(offset int) *UserWeightSelectBuilder {
	a.b = a.b.Offset(uint64(offset))

	return a
}

func (a *UserWeightSelectBuilder) WithBlock() *UserWeightSelectBuilder {
	a.b = a.b.Suffix("FOR UPDATE OF user_weight")

	return a
}

func (a *UserWeightSelectBuilder) ToSQL() (query string, args []any, err error) {
	return a.b.ToSql()
}

type UserWeightDeleteBuilder struct {
	b sq.DeleteBuilder
}

func NewUserWeightDeleteBuilder() *UserWeightDeleteBuilder {
	deleteBuilder := newDeleteQueryBuilder().
		Delete(userWeightTable)

	return &UserWeightDeleteBuilder{b: deleteBuilder}
}

func (a *UserWeightDeleteBuilder) WithFilterSpecification(spec *UserWeightFilterSpecification) *UserWeightDeleteBuilder {
	a.b = ApplyFilter(a.b, spec)

	return a
}

func (a *UserWeightDeleteBuilder) ToSQL() (query string, args []any, err error) {
	return a.b.ToSql()
}
