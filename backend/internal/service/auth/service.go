package auth

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"backend/internal/errors"
	"backend/pkg/JWT"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	refreshTokenLength   = 64
	verificationCodeLen  = 6
	verificationCodeTTL  = 10 * time.Minute
	refreshTokenTTL      = 30 * 24 * time.Hour
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type (
	UserInfoRepository interface {
		Get(ctx context.Context, f dto.UserInfoFilter, withBlock bool) (*entities.UserInfo, error)
		Create(ctx context.Context, userInfo *entities.UserInfo) error
		Update(ctx context.Context, userInfo *entities.UserInfo) error
	}

	UserRefreshTokensRepository interface {
		Create(ctx context.Context, t *entities.UserRefreshToken) error
		Get(ctx context.Context, f dto.UserRefreshTokenFilter) (*entities.UserRefreshToken, error)
		Delete(ctx context.Context, f dto.UserRefreshTokenFilter) error
		DeleteByUser(ctx context.Context, userID uuid.UUID) error
	}

	UserVerificationCodesRepository interface {
		Create(ctx context.Context, c *entities.UserVerificationCode) error
		GetLatest(ctx context.Context, f dto.UserVerificationCodeFilter) (*entities.UserVerificationCode, error)
		MarkUsed(ctx context.Context, id interface{}) error
	}

	TasksRepository interface {
		Create(ctx context.Context, t *entities.Task) error
	}

	TransactionManager interface {
		Do(ctx context.Context, fn func(ctx context.Context) error) (err error)
	}
)

type Service struct {
	txm                   TransactionManager
	userInfoRepo          UserInfoRepository
	refreshTokensRepo     UserRefreshTokensRepository
	verificationCodesRepo UserVerificationCodesRepository
	tasksRepo             TasksRepository
}

type Config struct {
	TransactionManager          TransactionManager
	UserInfoRepository          UserInfoRepository
	UserRefreshTokensRepository UserRefreshTokensRepository
	VerificationCodesRepository UserVerificationCodesRepository
	TasksRepository             TasksRepository
}

func NewService(c *Config) *Service {
	return &Service{
		txm:                   c.TransactionManager,
		userInfoRepo:          c.UserInfoRepository,
		refreshTokensRepo:     c.UserRefreshTokensRepository,
		verificationCodesRepo: c.VerificationCodesRepository,
		tasksRepo:             c.TasksRepository,
	}
}

func (u *Service) hashesPassword(user *entities.UserInfoInitSpec) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("%s: %v", errors.ErrHashedPassword, err)
	}
	user.Password = string(hashedPassword)
	return nil
}

func (u *Service) Register(ctx context.Context, info entities.UserInfoInitSpec) error {
	return u.txm.Do(ctx, func(ctx context.Context) error {
		existingUser, _ := u.userInfoRepo.Get(ctx, dto.UserInfoFilter{Username: &info.Username}, false)
		if existingUser != nil {
			return fmt.Errorf("register: %w", errors.ErrUserAlreadyExists)
		}

		if err := u.hashesPassword(&info); err != nil {
			return fmt.Errorf("register: %w", err)
		}

		if err := u.userInfoRepo.Create(ctx, entities.NewUserInfo(entities.WithUserInfoInitSpec(info))); err != nil {
			return fmt.Errorf("register: %w", err)
		}

		return nil
	})
}

func (u *Service) Login(ctx context.Context, ua entities.UserAuthInitSpec) (TokenPair, error) {
	user, err := u.userInfoRepo.Get(ctx, dto.UserInfoFilter{Username: &ua.Username}, false)
	if err != nil {
		return TokenPair{}, fmt.Errorf("login: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password()), []byte(ua.Password)); err != nil {
		return TokenPair{}, fmt.Errorf("login: %w", errors.ErrInvalidCredentials)
	}

	accessToken, err := JWT.GenerateJWT(user)
	if err != nil {
		return TokenPair{}, fmt.Errorf("login: %w: %v", errors.ErrTokenGeneration, err)
	}

	rawRefresh, err := u.issueRefreshToken(ctx, user.ID())
	if err != nil {
		return TokenPair{}, fmt.Errorf("login: %w", err)
	}

	return TokenPair{AccessToken: accessToken, RefreshToken: rawRefresh}, nil
}

func (u *Service) Refresh(ctx context.Context, rawToken string) (TokenPair, error) {
	hash := hashToken(rawToken)
	record, err := u.refreshTokensRepo.Get(ctx, dto.UserRefreshTokenFilter{TokenHash: &hash})
	if err != nil {
		return TokenPair{}, fmt.Errorf("refresh: token not found: %w", errors.ErrInvalidCredentials)
	}
	if record.IsExpired() {
		_ = u.refreshTokensRepo.Delete(ctx, dto.UserRefreshTokenFilter{TokenHash: &hash})
		return TokenPair{}, fmt.Errorf("refresh: %w", errors.ErrTokenExpired)
	}

	userID := record.UserID()
	user, err := u.userInfoRepo.Get(ctx, dto.UserInfoFilter{ID: &userID}, false)
	if err != nil {
		return TokenPair{}, fmt.Errorf("refresh: %w", err)
	}

	// rotate: delete old token, issue new pair
	_ = u.refreshTokensRepo.Delete(ctx, dto.UserRefreshTokenFilter{TokenHash: &hash})

	accessToken, err := JWT.GenerateJWT(user)
	if err != nil {
		return TokenPair{}, fmt.Errorf("refresh: %w: %v", errors.ErrTokenGeneration, err)
	}

	rawRefresh, err := u.issueRefreshToken(ctx, user.ID())
	if err != nil {
		return TokenPair{}, fmt.Errorf("refresh: %w", err)
	}

	return TokenPair{AccessToken: accessToken, RefreshToken: rawRefresh}, nil
}

func (u *Service) SendVerificationCode(ctx context.Context, userID uuid.UUID, codeType entities.VerificationCodeType) error {
	user, err := u.userInfoRepo.Get(ctx, dto.UserInfoFilter{ID: &userID}, false)
	if err != nil {
		return fmt.Errorf("send verification code: %w", err)
	}

	code, codeHash, err := generateVerificationCode()
	if err != nil {
		return fmt.Errorf("send verification code: %w", err)
	}

	vcEntity := entities.NewUserVerificationCode(entities.WithUserVerificationCodeInitSpec(entities.UserVerificationCodeInitSpec{
		ID:        uuid.New(),
		UserID:    userID,
		CodeHash:  codeHash,
		CodeType:  codeType,
		ExpiresAt: time.Now().Add(verificationCodeTTL),
	}))

	if err := u.verificationCodesRepo.Create(ctx, vcEntity); err != nil {
		return fmt.Errorf("send verification code: %w", err)
	}

	var taskType entities.TaskType
	var taskAttr entities.TaskAttribute

	switch codeType {
	case entities.VerificationCodeEmail:
		taskType = entities.TaskTypeSendCodeOnEmail
		taskAttr = entities.TaskAttribute{
			UserID:  userID,
			Email:   user.Email(),
			Subject: "BodyFuel — подтверждение email",
			Body:    fmt.Sprintf("Ваш код подтверждения: %s", code),
			Code:    code,
		}
	case entities.VerificationCodePhone:
		taskType = entities.TaskTypeSendCodeOnPhone
		taskAttr = entities.TaskAttribute{
			UserID: userID,
			Phone:  user.Phone(),
			Body:   fmt.Sprintf("BodyFuel: ваш код подтверждения %s", code),
			Code:   code,
		}
	default:
		return fmt.Errorf("send verification code: unknown code type %s", codeType)
	}

	task := entities.NewTask(entities.WithTaskInitSpec(entities.TaskInitSpec{
		TypeNm:      taskType,
		MaxAttempts: 5,
		Attribute:   taskAttr,
	}))

	if err := u.tasksRepo.Create(ctx, task); err != nil {
		return fmt.Errorf("send verification code: create task: %w", err)
	}

	return nil
}

func (u *Service) VerifyCode(ctx context.Context, userID uuid.UUID, code string, codeType entities.VerificationCodeType) error {
	record, err := u.verificationCodesRepo.GetLatest(ctx, dto.UserVerificationCodeFilter{
		UserID:   &userID,
		CodeType: &codeType,
	})
	if err != nil {
		return fmt.Errorf("verify code: %w", errors.ErrInvalidVerificationCode)
	}
	if record.IsExpired() {
		return fmt.Errorf("verify code: %w", errors.ErrVerificationCodeExpired)
	}
	if record.IsUsed() {
		return fmt.Errorf("verify code: %w", errors.ErrVerificationCodeAlreadyUsed)
	}

	inputHash := hashToken(code)
	if inputHash != record.CodeHash() {
		return fmt.Errorf("verify code: %w", errors.ErrInvalidVerificationCode)
	}

	if err := u.verificationCodesRepo.MarkUsed(ctx, record.ID()); err != nil {
		return fmt.Errorf("verify code: mark used: %w", err)
	}

	return nil
}

func (u *Service) SendRecoveryCode(ctx context.Context, email string) error {
	user, err := u.userInfoRepo.Get(ctx, dto.UserInfoFilter{Email: &email}, false)
	if err != nil {
		// don't reveal if user exists
		return nil
	}

	code, codeHash, err := generateVerificationCode()
	if err != nil {
		return fmt.Errorf("send recovery code: %w", err)
	}

	codeType := entities.VerificationCodeRecover
	vcEntity := entities.NewUserVerificationCode(entities.WithUserVerificationCodeInitSpec(entities.UserVerificationCodeInitSpec{
		ID:        uuid.New(),
		UserID:    user.ID(),
		CodeHash:  codeHash,
		CodeType:  codeType,
		ExpiresAt: time.Now().Add(verificationCodeTTL),
	}))

	if err := u.verificationCodesRepo.Create(ctx, vcEntity); err != nil {
		return fmt.Errorf("send recovery code: %w", err)
	}

	task := entities.NewTask(entities.WithTaskInitSpec(entities.TaskInitSpec{
		TypeNm:      entities.TaskTypeSendCodeOnEmail,
		MaxAttempts: 5,
		Attribute: entities.TaskAttribute{
			UserID:  user.ID(),
			Email:   email,
			Subject: "BodyFuel — восстановление пароля",
			Body:    fmt.Sprintf("Ваш код для сброса пароля: %s", code),
			Code:    code,
		},
	}))

	if err := u.tasksRepo.Create(ctx, task); err != nil {
		return fmt.Errorf("send recovery code: create task: %w", err)
	}

	return nil
}

func (u *Service) ResetPassword(ctx context.Context, email, code, newPassword string) error {
	user, err := u.userInfoRepo.Get(ctx, dto.UserInfoFilter{Email: &email}, false)
	if err != nil {
		return fmt.Errorf("reset password: %w", errors.ErrInvalidCredentials)
	}

	codeType := entities.VerificationCodeRecover
	resetUserID := user.ID()
	record, err := u.verificationCodesRepo.GetLatest(ctx, dto.UserVerificationCodeFilter{
		UserID:   &resetUserID,
		CodeType: &codeType,
	})
	if err != nil {
		return fmt.Errorf("reset password: %w", errors.ErrInvalidVerificationCode)
	}
	if record.IsExpired() {
		return fmt.Errorf("reset password: %w", errors.ErrVerificationCodeExpired)
	}
	if record.IsUsed() {
		return fmt.Errorf("reset password: %w", errors.ErrVerificationCodeAlreadyUsed)
	}

	inputHash := hashToken(code)
	if inputHash != record.CodeHash() {
		return fmt.Errorf("reset password: %w", errors.ErrInvalidVerificationCode)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("reset password: hash: %w", err)
	}

	newPass := string(hashed)
	user.Update(entities.UserInfoUpdateParams{Password: &newPass})
	if err := u.userInfoRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("reset password: update: %w", err)
	}

	if err := u.verificationCodesRepo.MarkUsed(ctx, record.ID()); err != nil {
		return fmt.Errorf("reset password: mark used: %w", err)
	}

	// invalidate all refresh tokens after password reset
	_ = u.refreshTokensRepo.DeleteByUser(ctx, user.ID())

	return nil
}

// issueRefreshToken generates a random token, stores its hash, returns raw token.
func (u *Service) issueRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	raw, err := generateRandomHex(refreshTokenLength)
	if err != nil {
		return "", fmt.Errorf("issue refresh token: %w", err)
	}

	hash := hashToken(raw)
	record := entities.NewUserRefreshToken(entities.WithUserRefreshTokenInitSpec(entities.UserRefreshTokenInitSpec{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(refreshTokenTTL),
	}))

	if err := u.refreshTokensRepo.Create(ctx, record); err != nil {
		return "", fmt.Errorf("issue refresh token: %w", err)
	}

	return raw, nil
}

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func generateRandomHex(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func generateVerificationCode() (code string, hash string, err error) {
	digits := make([]byte, verificationCodeLen)
	for i := range digits {
		n, e := rand.Int(rand.Reader, big.NewInt(10))
		if e != nil {
			return "", "", e
		}
		digits[i] = byte('0') + byte(n.Int64())
	}
	code = string(digits)
	hash = hashToken(code)
	return code, hash, nil
}
