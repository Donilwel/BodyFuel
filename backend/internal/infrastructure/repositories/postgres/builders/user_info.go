package builders

import (
	"backend/internal/dto"
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"time"
)

const (
	userInfoTable = "bodyfuel.user_info"
)

type UserInfoFilterSpecification struct {
	ID        *uuid.UUID
	Username  *string
	Name      *string
	Surname   *string
	Password  *string
	Email     *string
	Phone     *string
	CreatedAt *time.Time
}

func NewUserInfoFilterSpecification(f dto.UserInfoFilter) *UserInfoFilterSpecification {
	s := &UserInfoFilterSpecification{
		ID:        f.ID,
		Username:  f.Username,
		Name:      f.Name,
		Surname:   f.Surname,
		Password:  f.Password,
		Email:     f.Email,
		Phone:     f.Phone,
		CreatedAt: f.CreatedAt,
	}

	return s
}

func (spec *UserInfoFilterSpecification) Predicates() []sq.Sqlizer {
	var predicates []sq.Sqlizer

	if v := spec.ID; v != nil {
		predicates = append(predicates, sq.Eq{"user_info.id": v})
	}

	if v := spec.Username; v != nil {
		predicates = append(predicates, sq.Eq{"user_info.username": v})
	}

	if v := spec.Name; v != nil {
		predicates = append(predicates, sq.Eq{"user_info.name": v})
	}

	if v := spec.Surname; v != nil {
		predicates = append(predicates, sq.Eq{"user_info.surname": v})
	}

	if v := spec.Password; v != nil {
		predicates = append(predicates, sq.Eq{"user_info.password": v})
	}

	if v := spec.Email; v != nil {
		predicates = append(predicates, sq.Eq{"user_info.email": v})
	}

	if v := spec.Phone; v != nil {
		predicates = append(predicates, sq.Eq{"user_info.phone": v})
	}

	if v := spec.CreatedAt; v != nil {
		predicates = append(predicates, sq.Eq{"user_info.created_at": v})
	}

	return predicates
}

type UserInfoSelectBuilder struct {
	b sq.SelectBuilder
}

func NewUserInfoSelectBuilder() *UserInfoSelectBuilder {
	selectBuilder := newQueryBuilder().Select(
		"user_info.id",
		"user_info.username",
		"user_info.name",
		"user_info.surname",
		"user_info.password",
		"user_info.email",
		"user_info.phone",
		"user_info.created_at",
	).From(userInfoTable)

	return &UserInfoSelectBuilder{b: selectBuilder}
}

func (a *UserInfoSelectBuilder) WithFilterSpecification(spec *UserInfoFilterSpecification) *UserInfoSelectBuilder {
	a.b = ApplyFilter(a.b, spec)

	return a
}

func (a *UserInfoSelectBuilder) Limit(limit int) *UserInfoSelectBuilder {
	if limit > 0 {
		a.b = a.b.Limit(uint64(limit))
	}

	return a
}

func (a *UserInfoSelectBuilder) Offset(offset int) *UserInfoSelectBuilder {
	a.b = a.b.Offset(uint64(offset))

	return a
}

func (a *UserInfoSelectBuilder) WithBlock() *UserInfoSelectBuilder {
	a.b = a.b.Suffix("FOR UPDATE OF user_info")

	return a
}

func (a *UserInfoSelectBuilder) ToSQL() (query string, args []any, err error) {
	return a.b.ToSql()
}

type UserInfoDeleteBuilder struct {
	b sq.DeleteBuilder
}

func NewUserInfoDeleteBuilder() *UserInfoDeleteBuilder {
	deleteBuilder := newQueryBuilder().
		Delete(userInfoTable)

	return &UserInfoDeleteBuilder{b: deleteBuilder}
}

func (a *UserInfoDeleteBuilder) WithFilterSpecification(spec *UserInfoFilterSpecification) *UserInfoDeleteBuilder {
	a.b = ApplyFilter(a.b, spec)

	return a
}

func (a *UserInfoDeleteBuilder) ToSQL() (query string, args []any, err error) {
	return a.b.ToSql()
}
