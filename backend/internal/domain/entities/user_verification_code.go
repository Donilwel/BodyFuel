package entities

import (
	"github.com/google/uuid"
	"time"
)

type VerificationCodeType string

const (
	VerificationCodeEmail    VerificationCodeType = "email"
	VerificationCodePhone    VerificationCodeType = "phone"
	VerificationCodeRecover  VerificationCodeType = "recover"
)

type UserVerificationCode struct {
	id        uuid.UUID
	userID    uuid.UUID
	codeHash  string
	codeType  VerificationCodeType
	expiresAt time.Time
	usedAt    *time.Time
	createdAt time.Time
}

func (c *UserVerificationCode) ID() uuid.UUID                 { return c.id }
func (c *UserVerificationCode) UserID() uuid.UUID              { return c.userID }
func (c *UserVerificationCode) CodeHash() string               { return c.codeHash }
func (c *UserVerificationCode) CodeType() VerificationCodeType { return c.codeType }
func (c *UserVerificationCode) ExpiresAt() time.Time           { return c.expiresAt }
func (c *UserVerificationCode) UsedAt() *time.Time             { return c.usedAt }
func (c *UserVerificationCode) CreatedAt() time.Time           { return c.createdAt }
func (c *UserVerificationCode) IsExpired() bool                { return time.Now().After(c.expiresAt) }
func (c *UserVerificationCode) IsUsed() bool                   { return c.usedAt != nil }

func (c *UserVerificationCode) MarkUsed() {
	now := time.Now()
	c.usedAt = &now
}

type UserVerificationCodeOption func(c *UserVerificationCode)

func NewUserVerificationCode(opt UserVerificationCodeOption) *UserVerificationCode {
	c := new(UserVerificationCode)
	opt(c)
	return c
}

type UserVerificationCodeInitSpec struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	CodeHash  string
	CodeType  VerificationCodeType
	ExpiresAt time.Time
}

type UserVerificationCodeRestoreSpec struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	CodeHash  string
	CodeType  VerificationCodeType
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

func WithUserVerificationCodeInitSpec(s UserVerificationCodeInitSpec) UserVerificationCodeOption {
	return func(c *UserVerificationCode) {
		c.id = s.ID
		c.userID = s.UserID
		c.codeHash = s.CodeHash
		c.codeType = s.CodeType
		c.expiresAt = s.ExpiresAt
		c.createdAt = time.Now()
	}
}

func WithUserVerificationCodeRestoreSpec(s UserVerificationCodeRestoreSpec) UserVerificationCodeOption {
	return func(c *UserVerificationCode) {
		c.id = s.ID
		c.userID = s.UserID
		c.codeHash = s.CodeHash
		c.codeType = s.CodeType
		c.expiresAt = s.ExpiresAt
		c.usedAt = s.UsedAt
		c.createdAt = s.CreatedAt
	}
}
