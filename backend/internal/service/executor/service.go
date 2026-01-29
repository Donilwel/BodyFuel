package executor

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"backend/pkg/logging"
	"context"
	"github.com/google/uuid"
	"sync"
	"time"
)

const (
	moduleFieldName    = "module"
	executorModuleName = "executor"

	handleTaskErrorType      = "handle_task"
	getPendingTasksErrorType = "get_pending_tasks"

	banTimeout = 10 * time.Second

	executorTaskFailedTotalName = "executor_tasks_failed_total"
	bannedUsersTotalCounterName = "actually_banned_users_total"
	banDurationName             = "ban_duration_seconds"

	executorTaskFailedTotalHelp = "Total number of failed tasks by the executor service"
	bannedUsersTotalCounterHelp = "Total number of banned users"
	banDurationHelp             = "Duration of user ban/unban operations"
)

type (
	TransactionManager interface {
		Do(ctx context.Context, fn func(ctx context.Context) error) error
	}

	TasksRepository interface {
		Get(ctx context.Context, f dto.TasksFilter, withBlock bool) (*entities.Task, error)
		Delete(ctx context.Context, ids []uuid.UUID) error
		Update(ctx context.Context, f *entities.Task) error
	}
)

type handleTaskFunc func(ctx context.Context, t *entities.Task) error

type Config struct {
	TransactionManager TransactionManager
	//TasksRepository    TasksRepository

	QueryDelay time.Duration
}

type Service struct {
	txm TransactionManager
	//tasksRepository TasksRepository
	queryDelay time.Duration

	cancelFn context.CancelFunc
	wg       sync.WaitGroup

	log logging.Entry
}

func NewService(cfg *Config) *Service {
	return &Service{
		txm: cfg.TransactionManager,
		//tasksRepository: cfg.TasksRepository,
		queryDelay: cfg.QueryDelay,
	}
}

//func (s *Service) Run() error {
//	ctx, cancelFn := context.WithCancel(context.Background())
//	s.cancelFn = cancelFn
//
//	s.log = logging.GetLoggerFromContext(ctx).WithFields(logging.Fields{
//		moduleFieldName: executorModuleName,
//	})
//
//	s.wg.Add(1)
//	go func() {
//		defer func() {
//			if r := recover(); r != nil {
//				s.log.Errorf("Recovered in %s service: %v; stack trace: %s", executorModuleName, r, debug.Stack())
//			}
//			s.wg.Done()
//		}()
//
//		s.run(ctx)
//	}()
//
//	s.log.Infof("Started task executor service")
//
//	return nil
//}
//
//func (s *Service) run(ctx context.Context) {
//	for {
//		select {
//		case <-ctx.Done():
//			return
//		case <-time.After(s.queryDelay):
//			if err := s.processTask(ctx); err != nil {
//				s.log.Errorf("Process pending tasks: %v", err)
//			}
//		}
//	}
//}
//
//func (s *Service) processTask(ctx context.Context) error {
//	ctx, cancel := context.WithTimeout(ctx, banTimeout)
//	defer cancel()
//
//	s.log.Infof("Executing task %s started")
//
//	if err := s.handleTask(ctx, &entities.Task{}); err != nil {
//		s.log.Errorf("Handle task: %v", err)
//
//		return err
//	}
//	return nil
//}
//
//func (s *Service) handleTask(ctx context.Context, t *entities.Task) error {
//	return nil // TODO: заглушка пока что
//}

////
//func (s *Service) processTasks(ctx context.Context) error {
//	taskFound := true
//	for taskFound {
//		if err := s.txm.Do(ctx, func(ctx context.Context) error {
//			task, err := s.tasksRepository.Get(ctx, dto.TasksFilter{
//				States: []entities.TaskState{entities.TaskStateRunning},
//				Types: []entities.TaskType{entities.TaskTypeSendCodeOnEmail, entities.TaskTypeSendCodeOnPhone,
//					entities.TaskTypeSendNotificationEmail, entities.TaskTypeSendNotificationPhone},
//				RetryAt:        ptr.To(time.Now()),
//				ClusterIsReady: ptr.To(true),
//			}, true)
//			if err != nil {
//				if errors.Is(err, errs.ErrTaskNotFound()) {
//					taskFound = false
//					return nil
//				}
//
//				s.log.Errorf("Get pending tasks: %v", err)
//
//				return fmt.Errorf("get pending tasks: %w", err)
//			}
//			s.processTask(ctx, task)
//
//			return nil
//		}); err != nil {
//			return fmt.Errorf("task transaction failed: %w", err)
//		}
//	}
//
//	return nil
//}
//
//func (s *Service) handleTask(ctx context.Context, t *entities.Task) error {
//	var fn handleTaskFunc
//	switch t.TypeNm() {
//	case entities.TaskTypeBanUser:
//		fn = s.handleTaskBanUser
//	case entities.TaskTypeUnbanUser:
//		fn = s.handleTaskUnbanUser
//	default:
//		return errs.ErrUnsupportedTaskType().WithMetadata(map[string]any{
//			"task_type": t.TypeNm(),
//		})
//	}
//
//	execErr := fn(ctx, t)
//	switch {
//	case execErr == nil:
//		return nil
//	case errors.Is(execErr, context.Canceled):
//		s.log.Errorf("Termination timeout for task %s", t.UUID())
//	}
//
//	t.CalculateNextRetryAt()
//
//	if t.IsLimitAttemptsExceeded() {
//		t.Failed()
//
//		if err := s.tasksRepository.Update(ctx, t); err != nil {
//			return fmt.Errorf("update task: %w", err)
//		}
//		s.log.Errorf("Task %s failed: %v", t.UUID(), execErr)
//
//		return nil
//	}
//
//	return nil
//}
//
//func (s *Service
//func (s *Service) handleTaskBanUser(ctx context.Context, t *entities.Task) error {
//	attr := t.Attribute().(*entities.TaskBanUserAttribute)
//
//	if err := s.gpBanClient.Ban(ctx, attr); err != nil {
//		return fmt.Errorf("ban user: %w", err)
//	}
//
//	if err := s.tasksRepository.Delete(ctx, []uuid.UUID{t.UUID()}); err != nil {
//		return fmt.Errorf("delete task: %w", err)
//	}
//
//	s.log.Infof("Task %s successfully completed.  %s ", t.UUID(), attr.Username)
//
//	return nil
//}
//
//func (s *Service) handleTaskUnbanUser(ctx context.Context, t *entities.Task) error {
//	attr := t.Attribute().(*entities.TaskBanUserAttribute)
//
//	if err := s.gpBanClient.Unban(ctx, attr); err != nil {
//		return fmt.Errorf("ban user: %w", err)
//	}
//
//	if err := s.tasksRepository.Delete(ctx, []uuid.UUID{t.UUID()}); err != nil {
//		return fmt.Errorf("delete task: %w", err)
//	}
//	s.log.Infof("Task %s successfully completed. User %s unbanned", t.UUID(), attr.Username)
//
//	return nil
//}
//
//func (s *Service) taskType(t entities.TaskType) string {
//	switch t {
//	case entities.TaskTypeBanUser:
//		return "ban"
//	case entities.TaskTypeUnbanUser:
//		return "unban"
//	default:
//		return "unknown"
//	}
//}
//

func (s *Service) Close() error {
	if s.cancelFn != nil {
		s.cancelFn()
	}

	s.wg.Wait()

	s.log.Info("Stopped task executor service")

	return nil
}
