package entities

import (
	"github.com/google/uuid"
	"math/rand"
	"time"
)

type TaskState string

func (t TaskState) String() string {
	return string(t)
}

const (
	TaskStateRunning TaskState = "running"
	TaskStateFailed  TaskState = "failed"
)

type TaskType string

func (t TaskType) String() string {
	return string(t)
}

const (
	TaskTypeSendCodeOnEmail       TaskType = "send_code_email_task"
	TaskTypeSendCodeOnPhone       TaskType = "send_code_phone_task"
	TaskTypeSendNotificationEmail TaskType = "send_notification_email_task"
	TaskTypeSendNotificationPhone TaskType = "send_notification_phone_task"
)

type Task struct {
	uuid        uuid.UUID
	typeNm      TaskType
	message     string
	state       TaskState
	maxAttempts int
	attempts    int
	retryAt     time.Time
	createdAt   time.Time
	updatedAt   time.Time
	attribute   any

	calculateBackoffFn  func(n int) int
	baseBackoffDuration time.Duration
}

func (t *Task) IsLimitAttemptsExceeded() bool {
	return t.attempts >= t.maxAttempts
}

func (t *Task) UUID() uuid.UUID {
	return t.uuid
}

func (t *Task) TypeNm() TaskType {
	return t.typeNm
}

func (t *Task) Message() string {
	return t.message
}

func (t *Task) State() TaskState {
	return t.state
}

func (t *Task) MaxAttempts() int {
	return t.maxAttempts
}

func (t *Task) Attempts() int {
	return t.attempts
}

func (t *Task) IsAvailableForExecution() bool {
	return t.retryAt.Before(time.Now())
}

func (t *Task) RetryAt() time.Time {
	return t.retryAt
}

func (t *Task) CreatedAt() time.Time {
	return t.createdAt
}

func (t *Task) IsFailed() bool {
	return t.state == TaskStateFailed
}

func (t *Task) UpdatedAt() time.Time {
	return t.updatedAt
}

func (t *Task) Attribute() any {
	return t.attribute
}

func (t *Task) Failed() {
	t.state = TaskStateFailed
	t.updatedAt = time.Now()
}

func (t *Task) CalculateNextRetryAt() {
	t.attempts++

	if t.attempts > 1 {
		t.retryAt = t.retryAt.Add(t.baseBackoffDuration * time.Duration(t.calculateBackoffFn(t.attempts)))
	}

	t.retryAt = time.Now().Add(t.baseBackoffDuration)
	t.updatedAt = time.Now()
}

func (t *Task) SetState(s TaskState) {
	t.state = s
	t.updatedAt = time.Now()
}

func (t *Task) SetAttribute(a any) {
	t.attribute = a
	t.updatedAt = time.Now()
}

func (t *Task) Restart() {
	t.state = TaskStateRunning
	t.attempts = 0
	t.retryAt = time.Now()
	t.updatedAt = time.Now()
}

type TaskOption func(t *Task)

func NewTask(opt TaskOption) *Task {
	d := &Task{
		//cluster: c,
	}

	opt(d)

	return d
}

func WithTaskInitSpec(s TaskInitSpec) TaskOption {
	return func(t *Task) {
		switch s.TypeNm {
		case TaskTypeSendCodeOnEmail, TaskTypeSendCodeOnPhone:
			t.calculateBackoffFn = fibonacciBackoffCalculate
			t.baseBackoffDuration = 20 * time.Second
		case TaskTypeSendNotificationEmail, TaskTypeSendNotificationPhone:
			t.calculateBackoffFn = exponentialBackoffWithJitterCalculate
			t.baseBackoffDuration = 10 * time.Second
		default:
			t.calculateBackoffFn = linearBackoffCalculate
			t.baseBackoffDuration = 20 * time.Second
		}

		t.uuid = uuid.New()
		t.typeNm = s.TypeNm
		t.state = TaskStateRunning
		t.maxAttempts = s.MaxAttempts
		t.attempts = 0
		t.retryAt = time.Now()
		t.createdAt = time.Now()
		t.updatedAt = time.Now()
		t.attribute = s.Attribute
	}
}

type TaskInitSpec struct {
	TypeNm      TaskType
	Message     string
	MaxAttempts int
	Attribute   any
}

func WithTaskRestoreSpec(s TaskRestoreSpecification) TaskOption {
	return func(t *Task) {
		switch s.TypeNm {
		case TaskTypeSendCodeOnEmail, TaskTypeSendCodeOnPhone:
			t.calculateBackoffFn = exponentialBackoffCalculate
			t.baseBackoffDuration = 20 * time.Second
		case TaskTypeSendNotificationEmail, TaskTypeSendNotificationPhone:
			t.calculateBackoffFn = exponentialBackoffWithJitterCalculate
			t.baseBackoffDuration = 10 * time.Second
		default:
			t.calculateBackoffFn = linearBackoffCalculate
			t.baseBackoffDuration = 20 * time.Second
		}

		t.uuid = s.UUID
		t.typeNm = s.TypeNm
		//t.message = s.Message
		t.state = s.State
		t.maxAttempts = s.MaxAttempts
		t.attempts = s.Attempts
		t.retryAt = s.RetryAt
		t.createdAt = s.CreatedAt
		t.updatedAt = s.UpdatedAt
		t.attribute = s.Attribute
	}
}

type TaskRestoreSpecification struct {
	UUID        uuid.UUID
	TypeNm      TaskType
	ClusterNm   string
	State       TaskState
	MaxAttempts int
	Attempts    int
	RetryAt     time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Attribute   any
}

func fibonacciBackoffCalculate(n int) int {
	if n <= 1 {
		return n
	}

	a, b := 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}

	return b
}

func exponentialBackoffCalculate(n int) int {
	return 1 << (n - 1)
}

func linearBackoffCalculate(n int) int {
	return n
}

func exponentialBackoffWithJitterCalculate(n int) int {
	return 1<<(n-1) + rand.Intn(3)
}

type TaskAttribute struct {
	UserID uuid.UUID
	Method string
}
