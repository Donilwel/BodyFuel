package service

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"backend/internal/errors"
	"backend/pkg/JWT"
	"backend/pkg/logging"
	"context"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

type (
	UserInfoRepository interface {
		Get(ctx context.Context, f dto.UserInfoFilter, withBlock bool) (*entities.UserInfo, error)
		Create(ctx context.Context, userInfo *entities.UserInfo) error
	}

	TransactionManager interface {
		Do(ctx context.Context, fn func(ctx context.Context) error) (err error)
	}
)

type Service struct {
	txm          TransactionManager
	userInfoRepo UserInfoRepository
	log          logging.Entry
}

type Config struct {
	TransactionManager TransactionManager
	UserInfoRepository UserInfoRepository
	Log                logging.Entry
}

func NewService(c *Config) *Service {
	return &Service{
		txm:          c.TransactionManager,
		userInfoRepo: c.UserInfoRepository,
		log:          c.Log,
	}
}

func (u *Service) hashesPassword(user *entities.UserInfoInitSpec) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.ErrHashedPassword
	}
	user.Password = string(hashedPassword)
	return nil
}

func (u *Service) Register(ctx context.Context, info entities.UserInfoInitSpec) error {
	return u.txm.Do(ctx, func(ctx context.Context) error {
		existingUser, _ := u.userInfoRepo.Get(ctx, dto.UserInfoFilter{Username: &info.Username}, false)
		if existingUser != nil {
			return errors.ErrUserAlreadyExists
		}

		if err := u.hashesPassword(&info); err != nil {
			return fmt.Errorf("register: %w", err)
		}

		if err := u.userInfoRepo.Create(ctx, entities.NewUserInfo(entities.WithUserInfoInitSpec(info))); err != nil {
			return err
		}

		return nil
	})
}

func (u *Service) Login(ctx context.Context, ua entities.UserAuthInitSpec) (string, error) {
	user, err := u.userInfoRepo.Get(ctx, dto.UserInfoFilter{Username: &ua.Username}, false)
	if err != nil {
		return "", err
	}

	token, err := u.checkPasswordAndTakeToken(user, ua.Password)
	if err != nil {
		return "", fmt.Errorf("login: %w", err)
	}

	return token, nil
}

func (u *Service) checkPasswordAndTakeToken(user *entities.UserInfo, pass string) (string, error) {
	password := user.Password()
	if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(pass)); err != nil {
		return "", errors.ErrInvalidCredentials
	}

	token, err := JWT.GenerateJWT(user)
	if err != nil {
		return "", errors.ErrTokenGeneration
	}

	return token, nil
}
