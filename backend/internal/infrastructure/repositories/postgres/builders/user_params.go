package builders

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

const (
	userParamsTable = "bodyfuel.user_params"
)

type UserParamsFilterSpecification struct {
	ID        *uuid.UUID
	UserID    *uuid.UUID
	Height    *int
	Photo     *string
	Wants     *entities.Want
	Lifestyle *entities.Lifestyle
}

func NewUserParamsFilterSpecification(f dto.UserParamsFilter) *UserParamsFilterSpecification {
	s := &UserParamsFilterSpecification{
		ID:        f.ID,
		UserID:    f.UserID,
		Height:    f.Height,
		Photo:     f.Photo,
		Wants:     f.Wants,
		Lifestyle: f.Lifestyle,
	}

	return s
}

func (spec *UserParamsFilterSpecification) Predicates() []sq.Sqlizer {
	var predicates []sq.Sqlizer

	if v := spec.ID; v != nil {
		predicates = append(predicates, sq.Eq{"user_params.id": v})
	}

	if v := spec.UserID; v != nil {
		predicates = append(predicates, sq.Eq{"user_params.id_user": v})
	}

	if v := spec.Height; v != nil {
		predicates = append(predicates, sq.Eq{"user_params.height": v})
	}

	if v := spec.Photo; v != nil {
		predicates = append(predicates, sq.Eq{"user_params.photo": v})
	}

	if v := spec.Wants; v != nil {
		predicates = append(predicates, sq.Eq{"user_params.wants": v})
	}

	if v := spec.Lifestyle; v != nil {
		predicates = append(predicates, sq.Eq{"user_params.lifestyle": v})
	}

	return predicates
}

type UserParamsSelectBuilder struct {
	b sq.SelectBuilder
}

func NewUserParamsSelectBuilder() *UserParamsSelectBuilder {
	selectBuilder := newQueryBuilder().Select(
		"user_params.id",
		"user_params.id_user",
		"user_params.height",
		"user_params.photo",
		"user_params.wants",
		"user_params.lifestyle",
	).From(userParamsTable)

	return &UserParamsSelectBuilder{b: selectBuilder}
}

func (a *UserParamsSelectBuilder) WithFilterSpecification(spec *UserParamsFilterSpecification) *UserParamsSelectBuilder {
	a.b = ApplyFilter(a.b, spec)

	return a
}

func (a *UserParamsSelectBuilder) Limit(limit int) *UserParamsSelectBuilder {
	if limit > 0 {
		a.b = a.b.Limit(uint64(limit))
	}

	return a
}

func (a *UserParamsSelectBuilder) Offset(offset int) *UserParamsSelectBuilder {
	a.b = a.b.Offset(uint64(offset))

	return a
}

func (a *UserParamsSelectBuilder) WithBlock() *UserParamsSelectBuilder {
	a.b = a.b.Suffix("FOR UPDATE OF user_params")

	return a
}

func (a *UserParamsSelectBuilder) ToSQL() (query string, args []any, err error) {
	return a.b.ToSql()
}

type UserParamsDeleteBuilder struct {
	b sq.DeleteBuilder
}

func NewUserParamsDeleteBuilder() *UserParamsDeleteBuilder {
	deleteBuilder := newQueryBuilder().
		Delete(userParamsTable)

	return &UserParamsDeleteBuilder{b: deleteBuilder}
}

func (a *UserParamsDeleteBuilder) WithFilterSpecification(spec *UserParamsFilterSpecification) *UserParamsDeleteBuilder {
	a.b = ApplyFilter(a.b, spec)

	return a
}

func (a *UserParamsDeleteBuilder) ToSQL() (query string, args []any, err error) {
	return a.b.ToSql()
}
