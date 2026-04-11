package models

import (
	"backend/internal/domain/entities"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type TaskRow struct {
	UUID        uuid.UUID          `db:"task_id"`
	TypeNm      entities.TaskType  `db:"task_type_nm"`
	State       entities.TaskState `db:"task_state"`
	Attempts    int                `db:"attempts"`
	MaxAttempts int                `db:"max_attempts"`
	RetryAt     time.Time          `db:"retry_at"`
	Attribute   []byte             `db:"attribute"`
	CreatedAt   time.Time          `db:"created_at"`
	UpdatedAt   time.Time          `db:"updated_at"`
}

func NewTaskRow(t *entities.Task) (*TaskRow, error) {
	raw, err := json.Marshal(t.Attribute())
	if err != nil {
		return nil, fmt.Errorf("marshal attribute: %w", err)
	}

	return &TaskRow{
		UUID:        t.UUID(),
		TypeNm:      t.TypeNm(),
		State:       t.State(),
		Attempts:    t.Attempts(),
		MaxAttempts: t.MaxAttempts(),
		RetryAt:     t.RetryAt(),
		Attribute:   raw,
		CreatedAt:   t.CreatedAt(),
		UpdatedAt:   t.UpdatedAt(),
	}, nil
}

func (r *TaskRow) ToEntity() (*entities.Task, error) {
	var attr entities.TaskAttribute
	if len(r.Attribute) > 0 {
		if err := json.Unmarshal(r.Attribute, &attr); err != nil {
			return nil, fmt.Errorf("unmarshal attribute: %w", err)
		}
	}

	return entities.NewTask(entities.WithTaskRestoreSpec(entities.TaskRestoreSpecification{
		UUID:        r.UUID,
		TypeNm:      r.TypeNm,
		State:       r.State,
		Attempts:    r.Attempts,
		MaxAttempts: r.MaxAttempts,
		Attribute:   attr,
		RetryAt:     r.RetryAt,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	})), nil
}
