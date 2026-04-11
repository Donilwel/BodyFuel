package models

import (
	"backend/internal/domain/entities"
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

type TaskResponse struct {
	UUID        uuid.UUID `json:"uuid"`
	TypeNm      string    `json:"type_nm"`
	State       string    `json:"state"`
	MaxAttempts int       `json:"max_attempts"`
	Attempts    int       `json:"attempts"`
	RetryAt     time.Time `json:"retry_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Attribute   any       `json:"attribute"`
}

func NewTaskResponse(t *entities.Task) TaskResponse {
	return TaskResponse{
		UUID:        t.UUID(),
		TypeNm:      t.TypeNm().String(),
		State:       t.State().String(),
		MaxAttempts: t.MaxAttempts(),
		Attempts:    t.Attempts(),
		RetryAt:     t.RetryAt(),
		CreatedAt:   t.CreatedAt(),
		UpdatedAt:   t.UpdatedAt(),
		Attribute:   t.Attribute(),
	}
}

func NewTasksResponse(tasks []*entities.Task) []TaskResponse {
	resp := make([]TaskResponse, len(tasks))
	for i, t := range tasks {
		resp[i] = NewTaskResponse(t)
	}
	return resp
}
