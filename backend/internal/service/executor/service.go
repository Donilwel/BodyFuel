package executor

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"backend/pkg/logging"
	"backend/pkg/notifications/apns"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	moduleFieldName    = "module"
	executorModuleName = "executor"

	taskTimeout = 15 * time.Second
)

type (
	TransactionManager interface {
		Do(ctx context.Context, fn func(ctx context.Context) error) error
	}

	TasksRepository interface {
		Get(ctx context.Context, f dto.TasksFilter, withBlock bool) (*entities.Task, error)
		Delete(ctx context.Context, ids []uuid.UUID) error
		Update(ctx context.Context, t *entities.Task) error
	}

	UserInfoRepository interface {
		Get(ctx context.Context, f dto.UserInfoFilter, withBlock bool) (*entities.UserInfo, error)
	}

	EmailClient interface {
		SendEmail(to, subject, body string) error
	}

	SMSClient interface {
		SendSMS(to, body string) error
	}

	PushClient interface {
		Send(deviceToken string, p apns.Payload) error
	}
)

type handleTaskFunc func(ctx context.Context, t *entities.Task) error

type Config struct {
	TransactionManager TransactionManager
	TasksRepository    TasksRepository
	UserInfoRepository UserInfoRepository
	EmailClient        EmailClient
	SMSClient          SMSClient
	PushClient         PushClient
	QueryDelay         time.Duration
}

type Service struct {
	txm             TransactionManager
	tasksRepository TasksRepository
	userInfoRepo    UserInfoRepository
	emailClient     EmailClient
	smsClient       SMSClient
	pushClient      PushClient
	queryDelay      time.Duration

	cancelFn context.CancelFunc
	wg       sync.WaitGroup

	log logging.Entry
}

func NewService(cfg *Config) *Service {
	return &Service{
		txm:             cfg.TransactionManager,
		tasksRepository: cfg.TasksRepository,
		userInfoRepo:    cfg.UserInfoRepository,
		emailClient:     cfg.EmailClient,
		smsClient:       cfg.SMSClient,
		pushClient:      cfg.PushClient,
		queryDelay:      cfg.QueryDelay,
	}
}

func (s *Service) Run() error {
	ctx, cancelFn := context.WithCancel(context.Background())
	s.cancelFn = cancelFn

	s.log = logging.GetLoggerFromContext(ctx).WithFields(logging.Fields{
		moduleFieldName: executorModuleName,
	})

	s.wg.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.log.Errorf("Recovered in executor service: %v; stack: %s", r, debug.Stack())
			}
			s.wg.Done()
		}()
		s.run(ctx)
	}()

	s.log.Infof("Started task executor service")

	return nil
}

func (s *Service) run(ctx context.Context) {
	ticker := time.NewTicker(s.queryDelay)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Обрабатываем задачи пока они есть
			for {
				if err := s.processTask(ctx); err != nil {
					if !errors.Is(err, sql.ErrNoRows) {
						s.log.Errorf("Process task: %v", err)
					}
					break
				}
			}
		}
	}
}

func (s *Service) processTask(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, taskTimeout)
	defer cancel()

	return s.txm.Do(ctx, func(ctx context.Context) error {
		now := time.Now()
		task, err := s.tasksRepository.Get(ctx, dto.TasksFilter{
			States:  []entities.TaskState{entities.TaskStateRunning},
			RetryAt: &now,
		}, true)
		if err != nil {
			return err
		}

		return s.handleTask(ctx, task)
	})
}

func (s *Service) handleTask(ctx context.Context, t *entities.Task) error {
	var fn handleTaskFunc

	switch t.TypeNm() {
	case entities.TaskTypeSendCodeOnEmail, entities.TaskTypeSendNotificationEmail:
		fn = s.handleEmailTask
	case entities.TaskTypeSendCodeOnPhone, entities.TaskTypeSendNotificationPhone:
		fn = s.handleSMSTask
	case entities.TaskTypeSendPushNotification:
		fn = s.handlePushTask
	default:
		s.log.Warnf("Unknown task type %q (id=%s), deleting", t.TypeNm(), t.UUID())
		return s.tasksRepository.Delete(ctx, []uuid.UUID{t.UUID()})
	}

	if err := fn(ctx, t); err != nil {
		s.log.Errorf("Handle task %s (%s): %v", t.UUID(), t.TypeNm(), err)

		t.CalculateNextRetryAt()

		if t.IsLimitAttemptsExceeded() {
			t.Failed()
			s.log.Errorf("Task %s exceeded max attempts, marking as failed", t.UUID())
		}

		return s.tasksRepository.Update(ctx, t)
	}

	return s.tasksRepository.Delete(ctx, []uuid.UUID{t.UUID()})
}

func (s *Service) handleEmailTask(ctx context.Context, t *entities.Task) error {
	attr, ok := t.Attribute().(entities.TaskAttribute)
	if !ok {
		return fmt.Errorf("invalid attribute type for email task")
	}

	if attr.Email == "" {
		return fmt.Errorf("email is empty")
	}

	// Notification tasks require verified email; verification code tasks always go through.
	if t.TypeNm() == entities.TaskTypeSendNotificationEmail && s.userInfoRepo != nil {
		user, err := s.userInfoRepo.Get(ctx, dto.UserInfoFilter{ID: &attr.UserID}, false)
		if err == nil && !user.IsEmailVerified() {
			s.log.Warnf("Skipping email notification for user %s: email not verified", attr.UserID)
			return nil // delete the task, no retry needed
		}
	}

	subject := attr.Subject
	if subject == "" {
		subject = "BodyFuel"
	}

	body := attr.Body
	if body == "" {
		body = attr.Message
	}
	if body == "" {
		body = string(t.Message())
	}

	return s.emailClient.SendEmail(attr.Email, subject, body)
}

func (s *Service) handleSMSTask(ctx context.Context, t *entities.Task) error {
	attr, ok := t.Attribute().(entities.TaskAttribute)
	if !ok {
		return fmt.Errorf("invalid attribute type for sms task")
	}

	if attr.Phone == "" {
		return fmt.Errorf("phone is empty")
	}

	// Notification tasks require verified phone; verification code tasks always go through.
	if t.TypeNm() == entities.TaskTypeSendNotificationPhone && s.userInfoRepo != nil {
		user, err := s.userInfoRepo.Get(ctx, dto.UserInfoFilter{ID: &attr.UserID}, false)
		if err == nil && !user.IsPhoneVerified() {
			s.log.Warnf("Skipping SMS notification for user %s: phone not verified", attr.UserID)
			return nil // delete the task, no retry needed
		}
	}

	body := attr.Body
	if body == "" {
		body = attr.Message
	}
	if body == "" {
		body = string(t.Message())
	}
	if attr.Code != "" {
		body = fmt.Sprintf("%s Ваш код: %s", body, attr.Code)
	}

	return s.smsClient.SendSMS(attr.Phone, body)
}

func (s *Service) handlePushTask(ctx context.Context, t *entities.Task) error {
	attr, ok := t.Attribute().(entities.TaskAttribute)
	if !ok {
		return fmt.Errorf("invalid attribute type for push task")
	}

	if attr.DeviceToken == "" {
		return fmt.Errorf("device token is empty")
	}

	title := attr.Title
	if title == "" {
		title = "BodyFuel"
	}

	body := attr.Body
	if body == "" {
		body = attr.Message
	}
	if body == "" {
		body = string(t.Message())
	}

	return s.pushClient.Send(attr.DeviceToken, apns.Payload{
		Title: title,
		Body:  body,
	})
}

func (s *Service) Close() error {
	if s.cancelFn != nil {
		s.cancelFn()
	}

	s.wg.Wait()

	s.log.Info("Stopped task executor service")

	return nil
}
