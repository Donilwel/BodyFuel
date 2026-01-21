package entities

import (
	"github.com/google/uuid"
	"time"
)

type UserInfo struct {
	id        uuid.UUID
	username  string
	name      string
	surname   string
	password  string
	email     string
	phone     string
	createdAt time.Time
}

func (u *UserInfo) ID() uuid.UUID {
	return u.id
}

func (u *UserInfo) Username() string {
	return u.username
}

func (u *UserInfo) Name() string {
	return u.name
}

func (u *UserInfo) Surname() string {
	return u.surname
}

func (u *UserInfo) Password() string {
	return u.password
}

func (u *UserInfo) Email() string {
	return u.email
}

func (u *UserInfo) Phone() string {
	return u.phone
}

func (u *UserInfo) CreatedAt() time.Time {
	return u.createdAt
}

type UserInfoOption func(u *UserInfo)

func NewUserInfo(opt UserInfoOption) *UserInfo {
	u := new(UserInfo)

	opt(u)

	return u
}

type UserInfoRestoreSpec struct {
	ID        uuid.UUID
	Username  string
	Name      string
	Surname   string
	Password  string
	Email     string
	Phone     string
	CreatedAt time.Time
}

type UserInfoInitSpec struct {
	ID        uuid.UUID
	Username  string
	Name      string
	Surname   string
	Password  string
	Email     string
	Phone     string
	CreatedAt time.Time
}

type UserAuthInitSpec struct {
	Username string
	Password string
}

func WithUserInfoRestoreSpec(spec UserInfoRestoreSpec) UserInfoOption {
	return func(u *UserInfo) {
		u.id = spec.ID
		u.username = spec.Username
		u.name = spec.Name
		u.surname = spec.Surname
		u.password = spec.Password
		u.email = spec.Email
		u.phone = spec.Phone
		u.createdAt = spec.CreatedAt
	}
}

func WithUserInfoInitSpec(s UserInfoInitSpec) UserInfoOption {
	return func(u *UserInfo) {
		u.id = s.ID
		u.username = s.Username
		u.name = s.Name
		u.surname = s.Surname
		u.password = s.Password
		u.email = s.Email
		u.phone = s.Phone
		u.createdAt = s.CreatedAt
	}
}

type UserInfoUpdateParams struct {
	Username *string
	Name     *string
	Surname  *string
	Email    *string
	Phone    *string
}

func (ui *UserInfo) Update(p UserInfoUpdateParams) {
	if p.Username != nil {
		ui.username = *p.Username
	}
	if p.Name != nil {
		ui.name = *p.Name
	}
	if p.Surname != nil {
		ui.surname = *p.Surname
	}
	if p.Email != nil {
		ui.email = *p.Email
	}
	if p.Phone != nil {
		ui.phone = *p.Phone
	}
}
