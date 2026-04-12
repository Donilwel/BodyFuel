package executor

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"backend/pkg/logging"
	"backend/pkg/notifications/apns"
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ── mocks ──────────────────────────────────────────────────────────────────

type mockTxManager struct{}

func (m *mockTxManager) Do(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type mockTasksRepo struct{ mock.Mock }

func (m *mockTasksRepo) Get(ctx context.Context, f dto.TasksFilter, withBlock bool) (*entities.Task, error) {
	args := m.Called(ctx, f, withBlock)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Task), args.Error(1)
}

func (m *mockTasksRepo) Delete(ctx context.Context, ids []uuid.UUID) error {
	return m.Called(ctx, ids).Error(0)
}

func (m *mockTasksRepo) Update(ctx context.Context, t *entities.Task) error {
	return m.Called(ctx, t).Error(0)
}

type mockUserInfoRepo struct{ mock.Mock }

func (m *mockUserInfoRepo) Get(ctx context.Context, f dto.UserInfoFilter, withBlock bool) (*entities.UserInfo, error) {
	args := m.Called(ctx, f, withBlock)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.UserInfo), args.Error(1)
}

type mockEmailClient struct{ mock.Mock }

func (m *mockEmailClient) SendEmail(to, subject, body string) error {
	return m.Called(to, subject, body).Error(0)
}

type mockSMSClient struct{ mock.Mock }

func (m *mockSMSClient) SendSMS(to, body string) error {
	return m.Called(to, body).Error(0)
}

type mockPushClient struct{ mock.Mock }

func (m *mockPushClient) Send(deviceToken string, p apns.Payload) error {
	return m.Called(deviceToken, p).Error(0)
}

// ── helpers ────────────────────────────────────────────────────────────────

func newTaskWithAttr(taskType entities.TaskType, attr entities.TaskAttribute) *entities.Task {
	return entities.NewTask(entities.WithTaskInitSpec(entities.TaskInitSpec{
		TypeNm:      taskType,
		MaxAttempts: 3,
		Attribute:   attr,
	}))
}

func newVerifiedUser(userID uuid.UUID) *entities.UserInfo {
	now := time.Now()
	return entities.NewUserInfo(entities.WithUserInfoRestoreSpec(entities.UserInfoRestoreSpec{
		ID:              userID,
		Email:           "user@example.com",
		Phone:           "+79991234567",
		EmailVerifiedAt: &now,
		PhoneVerifiedAt: &now,
	}))
}

func newUnverifiedUser(userID uuid.UUID) *entities.UserInfo {
	return entities.NewUserInfo(entities.WithUserInfoRestoreSpec(entities.UserInfoRestoreSpec{
		ID:    userID,
		Email: "user@example.com",
		Phone: "+79991234567",
	}))
}

func newService(tasksRepo *mockTasksRepo, userInfoRepo *mockUserInfoRepo, email *mockEmailClient, sms *mockSMSClient, push *mockPushClient) *Service {
	return &Service{
		txm:             &mockTxManager{},
		tasksRepository: tasksRepo,
		userInfoRepo:    userInfoRepo,
		emailClient:     email,
		smsClient:       sms,
		pushClient:      push,
		queryDelay:      time.Minute,
		log:             logging.GetLoggerFromContext(context.Background()),
	}
}

// ── handleEmailTask ────────────────────────────────────────────────────────

func TestHandleEmailTask_Success(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	task := newTaskWithAttr(entities.TaskTypeSendCodeOnEmail, entities.TaskAttribute{
		UserID:  userID,
		Email:   "test@example.com",
		Subject: "Code",
		Body:    "Your code: 123456",
	})

	emailMock := &mockEmailClient{}
	emailMock.On("SendEmail", "test@example.com", "Code", "Your code: 123456").Return(nil)

	svc := newService(&mockTasksRepo{}, nil, emailMock, nil, nil)
	err := svc.handleEmailTask(ctx, task)

	assert.NoError(t, err)
	emailMock.AssertExpectations(t)
}

func TestHandleEmailTask_EmptyEmail_Error(t *testing.T) {
	ctx := context.Background()
	task := newTaskWithAttr(entities.TaskTypeSendCodeOnEmail, entities.TaskAttribute{
		Email: "",
	})

	svc := newService(&mockTasksRepo{}, nil, &mockEmailClient{}, nil, nil)
	err := svc.handleEmailTask(ctx, task)
	assert.Error(t, err)
}

func TestHandleEmailTask_NotificationBlocked_EmailNotVerified(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	task := newTaskWithAttr(entities.TaskTypeSendNotificationEmail, entities.TaskAttribute{
		UserID: userID,
		Email:  "test@example.com",
		Body:   "You have a new workout",
	})

	userInfoRepo := &mockUserInfoRepo{}
	userInfoRepo.On("Get", mock.Anything, mock.Anything, false).
		Return(newUnverifiedUser(userID), nil)

	emailMock := &mockEmailClient{} // SendEmail must NOT be called
	svc := newService(&mockTasksRepo{}, userInfoRepo, emailMock, nil, nil)
	err := svc.handleEmailTask(ctx, task)

	assert.NoError(t, err) // nil = delete the task silently
	emailMock.AssertNotCalled(t, "SendEmail")
}

func TestHandleEmailTask_NotificationAllowed_EmailVerified(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	task := newTaskWithAttr(entities.TaskTypeSendNotificationEmail, entities.TaskAttribute{
		UserID:  userID,
		Email:   "test@example.com",
		Subject: "New workout",
		Body:    "Ready!",
	})

	userInfoRepo := &mockUserInfoRepo{}
	userInfoRepo.On("Get", mock.Anything, mock.Anything, false).
		Return(newVerifiedUser(userID), nil)

	emailMock := &mockEmailClient{}
	emailMock.On("SendEmail", "test@example.com", "New workout", "Ready!").Return(nil)

	svc := newService(&mockTasksRepo{}, userInfoRepo, emailMock, nil, nil)
	err := svc.handleEmailTask(ctx, task)

	assert.NoError(t, err)
	emailMock.AssertExpectations(t)
}

func TestHandleEmailTask_DefaultSubject(t *testing.T) {
	ctx := context.Background()
	task := newTaskWithAttr(entities.TaskTypeSendCodeOnEmail, entities.TaskAttribute{
		Email:   "x@example.com",
		Subject: "", // empty → default "BodyFuel"
		Body:    "hello",
	})

	emailMock := &mockEmailClient{}
	emailMock.On("SendEmail", "x@example.com", "BodyFuel", "hello").Return(nil)

	svc := newService(&mockTasksRepo{}, nil, emailMock, nil, nil)
	err := svc.handleEmailTask(ctx, task)
	assert.NoError(t, err)
	emailMock.AssertExpectations(t)
}

// ── handleSMSTask ──────────────────────────────────────────────────────────

func TestHandleSMSTask_Success(t *testing.T) {
	ctx := context.Background()
	task := newTaskWithAttr(entities.TaskTypeSendCodeOnPhone, entities.TaskAttribute{
		Phone: "+79991234567",
		Body:  "Your code",
		Code:  "654321",
	})

	smsMock := &mockSMSClient{}
	smsMock.On("SendSMS", "+79991234567", "Your code Ваш код: 654321").Return(nil)

	svc := newService(&mockTasksRepo{}, nil, nil, smsMock, nil)
	err := svc.handleSMSTask(ctx, task)

	assert.NoError(t, err)
	smsMock.AssertExpectations(t)
}

func TestHandleSMSTask_EmptyPhone_Error(t *testing.T) {
	ctx := context.Background()
	task := newTaskWithAttr(entities.TaskTypeSendCodeOnPhone, entities.TaskAttribute{Phone: ""})
	svc := newService(&mockTasksRepo{}, nil, nil, &mockSMSClient{}, nil)
	err := svc.handleSMSTask(ctx, task)
	assert.Error(t, err)
}

func TestHandleSMSTask_NotificationBlocked_PhoneNotVerified(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	task := newTaskWithAttr(entities.TaskTypeSendNotificationPhone, entities.TaskAttribute{
		UserID: userID,
		Phone:  "+79991234567",
		Body:   "You have a new workout",
	})

	userInfoRepo := &mockUserInfoRepo{}
	userInfoRepo.On("Get", mock.Anything, mock.Anything, false).
		Return(newUnverifiedUser(userID), nil)

	smsMock := &mockSMSClient{}
	svc := newService(&mockTasksRepo{}, userInfoRepo, nil, smsMock, nil)
	err := svc.handleSMSTask(ctx, task)

	assert.NoError(t, err)
	smsMock.AssertNotCalled(t, "SendSMS")
}

// ── handlePushTask ─────────────────────────────────────────────────────────

func TestHandlePushTask_Success(t *testing.T) {
	ctx := context.Background()
	task := newTaskWithAttr(entities.TaskTypeSendPushNotification, entities.TaskAttribute{
		DeviceToken: "device-abc",
		Title:       "Workout",
		Body:        "Your workout is ready!",
	})

	pushMock := &mockPushClient{}
	pushMock.On("Send", "device-abc", apns.Payload{Title: "Workout", Body: "Your workout is ready!"}).Return(nil)

	svc := newService(&mockTasksRepo{}, nil, nil, nil, pushMock)
	err := svc.handlePushTask(ctx, task)

	assert.NoError(t, err)
	pushMock.AssertExpectations(t)
}

func TestHandlePushTask_EmptyDeviceToken_Error(t *testing.T) {
	ctx := context.Background()
	task := newTaskWithAttr(entities.TaskTypeSendPushNotification, entities.TaskAttribute{DeviceToken: ""})
	svc := newService(&mockTasksRepo{}, nil, nil, nil, &mockPushClient{})
	err := svc.handlePushTask(ctx, task)
	assert.Error(t, err)
}

func TestHandlePushTask_DefaultTitle(t *testing.T) {
	ctx := context.Background()
	task := newTaskWithAttr(entities.TaskTypeSendPushNotification, entities.TaskAttribute{
		DeviceToken: "tok",
		Title:       "", // empty → "BodyFuel"
		Body:        "msg",
	})

	pushMock := &mockPushClient{}
	pushMock.On("Send", "tok", apns.Payload{Title: "BodyFuel", Body: "msg"}).Return(nil)

	svc := newService(&mockTasksRepo{}, nil, nil, nil, pushMock)
	err := svc.handlePushTask(ctx, task)
	assert.NoError(t, err)
	pushMock.AssertExpectations(t)
}

// ── handleTask routing ─────────────────────────────────────────────────────

func TestHandleTask_DeletesOnSuccess(t *testing.T) {
	ctx := context.Background()
	task := newTaskWithAttr(entities.TaskTypeSendCodeOnEmail, entities.TaskAttribute{
		Email: "x@example.com",
		Body:  "code",
	})

	emailMock := &mockEmailClient{}
	emailMock.On("SendEmail", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	tasksRepo := &mockTasksRepo{}
	tasksRepo.On("Delete", mock.Anything, mock.Anything).Return(nil)

	svc := newService(tasksRepo, nil, emailMock, nil, nil)
	err := svc.handleTask(ctx, task)

	assert.NoError(t, err)
	tasksRepo.AssertCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestHandleTask_UpdatesOnFailure(t *testing.T) {
	ctx := context.Background()
	task := newTaskWithAttr(entities.TaskTypeSendCodeOnEmail, entities.TaskAttribute{
		Email: "x@example.com",
		Body:  "code",
	})

	emailMock := &mockEmailClient{}
	emailMock.On("SendEmail", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("smtp error"))

	tasksRepo := &mockTasksRepo{}
	tasksRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Task")).Return(nil)

	svc := newService(tasksRepo, nil, emailMock, nil, nil)
	err := svc.handleTask(ctx, task)

	assert.NoError(t, err)
	tasksRepo.AssertCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestHandleTask_UnknownType_Deleted(t *testing.T) {
	ctx := context.Background()
	task := entities.NewTask(entities.WithTaskInitSpec(entities.TaskInitSpec{
		TypeNm:      "unknown_task_type",
		MaxAttempts: 3,
		Attribute:   entities.TaskAttribute{},
	}))

	tasksRepo := &mockTasksRepo{}
	tasksRepo.On("Delete", mock.Anything, mock.Anything).Return(nil)

	svc := newService(tasksRepo, nil, nil, nil, nil)
	err := svc.handleTask(ctx, task)

	assert.NoError(t, err)
	tasksRepo.AssertCalled(t, "Delete", mock.Anything, mock.Anything)
}

// ── processTask ────────────────────────────────────────────────────────────

func TestProcessTask_NoRows_StopsLoop(t *testing.T) {
	ctx := context.Background()

	tasksRepo := &mockTasksRepo{}
	tasksRepo.On("Get", mock.Anything, mock.Anything, true).Return(nil, sql.ErrNoRows)

	svc := newService(tasksRepo, nil, nil, nil, nil)
	err := svc.processTask(ctx)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, sql.ErrNoRows))
}
