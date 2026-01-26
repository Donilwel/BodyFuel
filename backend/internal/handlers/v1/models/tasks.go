package models

import (
	"github.com/google/uuid"
	"time"
)

type TaskCreateResponseModel struct {
	UUID        uuid.UUID `json:"uuid"`
	TypeNm      string    `json:"task_nm"`
	State       string    `json:"state"`
	MaxAttempts int       `json:"max_attempts"`
	Attempts    int       `json:"attempts"`
	RetryAt     time.Time `json:"retry_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Attribute   any       `json:"attribute"`
}
