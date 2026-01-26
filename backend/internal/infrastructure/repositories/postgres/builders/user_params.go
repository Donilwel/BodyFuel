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
	ID                  *uuid.UUID
	UserID              *uuid.UUID
	Height              *int
	Photo               *string
	Wants               *entities.Want
	TargetWorkoutsWeeks *int
	TargetCaloriesDaily *int
	TargetWeight        *float64
	Lifestyle           *entities.Lifestyle
}

func NewUserParamsFilterSpecification(f dto.UserParamsFilter) *UserParamsFilterSpecification {
	s := &UserParamsFilterSpecification{
		ID:                  f.ID,
		UserID:              f.UserID,
		Height:              f.Height,
		Photo:               f.Photo,
		Wants:               f.Wants,
		TargetWorkoutsWeeks: f.TargetWorkoutsWeeks,
		TargetCaloriesDaily: f.TargetCaloriesDaily,
		TargetWeight:        f.TargetWeight,
		Lifestyle:           f.Lifestyle,
	}

	return s
}

func (spec *UserParamsFilterSpecification) Predicates() []sq.Sqlizer {
	var predicates []sq.Sqlizer

	if v := spec.ID; v != nil {
		predicates = append(predicates, sq.Eq{"p.id": v})
	}

	if v := spec.UserID; v != nil {
		predicates = append(predicates, sq.Eq{"p.id_user": v})
	}

	if v := spec.Height; v != nil {
		predicates = append(predicates, sq.Eq{"p.height": v})
	}

	if v := spec.Photo; v != nil {
		predicates = append(predicates, sq.Eq{"p.photo": v})
	}

	if v := spec.Wants; v != nil {
		predicates = append(predicates, sq.Eq{"p.wants": v})
	}

	if v := spec.Lifestyle; v != nil {
		predicates = append(predicates, sq.Eq{"p.lifestyle": v})
	}

	if v := spec.TargetWorkoutsWeeks; v != nil {
		predicates = append(predicates, sq.Eq{"p.target_workouts_weeks": v})
	}

	if v := spec.TargetWeight; v != nil {
		predicates = append(predicates, sq.Eq{"p.target_weight": v})
	}

	if v := spec.TargetCaloriesDaily; v != nil {
		predicates = append(predicates, sq.Eq{"p.target_calories_daily": v})
	}

	return predicates
}

type UserParamsSelectBuilder struct {
	b sq.SelectBuilder
}

func NewUserParamsSelectBuilder() *UserParamsSelectBuilder {
	latestWeightSubquery := sq.Select(
		"uw.id_user",
		"uw.weight",
		"uw.date",
	).
		From("bodyfuel.user_weight uw").
		InnerJoin(
			"(SELECT id_user, MAX(date) as max_date FROM bodyfuel.user_weight GROUP BY id_user) latest" +
				" ON uw.id_user = latest.id_user AND uw.date = latest.max_date")

	subquerySQL, _, _ := latestWeightSubquery.ToSql()

	b := newQueryBuilder().Select(
		"p.id",
		"p.id_user",
		"p.height",
		"p.photo",
		"p.wants",
		"p.lifestyle",
		"p.target_workouts_weeks",
		"p.target_weight",
		"p.target_calories_daily",
		"w.weight as current_weight",
	).
		From(userParamsTable + " p").
		LeftJoin("(" + subquerySQL + ") w ON p.id_user = w.id_user")

	return &UserParamsSelectBuilder{b: b}
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
	deleteBuilder := newDeleteQueryBuilder().
		Delete(userParamsTable + " p")

	return &UserParamsDeleteBuilder{b: deleteBuilder}
}

func (a *UserParamsDeleteBuilder) WithFilterSpecification(spec *UserParamsFilterSpecification) *UserParamsDeleteBuilder {
	a.b = ApplyFilter(a.b, spec)

	return a
}

func (a *UserParamsDeleteBuilder) ToSQL() (query string, args []any, err error) {
	return a.b.ToSql()
}
