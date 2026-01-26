package dto

import (
	"backend/internal/domain/entities"
	"github.com/google/uuid"
	"time"
)

type TasksFilter struct {
	IDs            []uuid.UUID
	Types          []entities.TaskType
	Message        *string
	Attempts       *int
	RetryAt        *time.Time
	States         []entities.TaskState
	ClusterIsReady *bool
}
