package auth

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	autherrors "backend/internal/errors"
	"backend/internal/service/auth/mocks"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// helpers

func strPtr(s string) *string { return &s }

func newHashedUser(username, password string) *entities.UserInfo {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return entities.NewUserInfo(entities.WithUserInfoInitSpec(entities.UserInfoInitSpec{
		ID:       uuid.New(),
		Username: username,
		Password: string(hashed),
		Email:    username + "@example.com",
		Phone:    "+79001234567",
	}))
}

func newActiveRefreshToken(userID uuid.UUID, raw string) *entities.UserRefreshToken {
	h := hashToken(raw)
	return entities.NewUserRefreshToken(entities.WithUserRefreshTokenInitSpec(entities.UserRefreshTokenInitSpec{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: h,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}))
}

func newExpiredRefreshToken(userID uuid.UUID, raw string) *entities.UserRefreshToken {
	h := hashToken(raw)
	return entities.NewUserRefreshToken(entities.WithUserRefreshTokenInitSpec(entities.UserRefreshTokenInitSpec{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: h,
		ExpiresAt: time.Now().Add(-time.Hour),
	}))
}

func newVerificationCode(userID uuid.UUID, code string, codeType entities.VerificationCodeType, expired, used bool) *entities.UserVerificationCode {
	h := hashToken(code)
	ttl := time.Now().Add(10 * time.Minute)
	if expired {
		ttl = time.Now().Add(-time.Minute)
	}
	vc := entities.NewUserVerificationCode(entities.WithUserVerificationCodeInitSpec(entities.UserVerificationCodeInitSpec{
		ID:        uuid.New(),
		UserID:    userID,
		CodeHash:  h,
		CodeType:  codeType,
		ExpiresAt: ttl,
	}))
	if used {
		vc.MarkUsed()
	}
	return vc
}

// ──────────────────────────────────────────────────────────────
// Register
// ──────────────────────────────────────────────────────────────

func TestService_Register(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		prepareMocks func(userRepo *mocks.UserInfoRepository, txm *mocks.TransactionManager)
		info         entities.UserInfoInitSpec
		wantErr      error
	}{
		{
			name: "success",
			prepareMocks: func(userRepo *mocks.UserInfoRepository, txm *mocks.TransactionManager) {
				txm.On("Do", mock.Anything, mock.Anything).
					Return(func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) })
				userRepo.On("Get", mock.Anything, dto.UserInfoFilter{Username: strPtr("user")}, false).
					Return(nil, nil)
				userRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.UserInfo")).
					Return(nil)
			},
			info:    entities.UserInfoInitSpec{Username: "user", Password: "password"},
			wantErr: nil,
		},
		{
			name: "user already exists",
			prepareMocks: func(userRepo *mocks.UserInfoRepository, txm *mocks.TransactionManager) {
				txm.On("Do", mock.Anything, mock.Anything).
					Return(func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) })
				userRepo.On("Get", mock.Anything, dto.UserInfoFilter{Username: strPtr("user")}, false).
					Return(&entities.UserInfo{}, nil)
			},
			info:    entities.UserInfoInitSpec{Username: "user", Password: "password"},
			wantErr: autherrors.ErrUserAlreadyExists,
		},
		{
			name: "create returns error",
			prepareMocks: func(userRepo *mocks.UserInfoRepository, txm *mocks.TransactionManager) {
				txm.On("Do", mock.Anything, mock.Anything).
					Return(func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) })
				userRepo.On("Get", mock.Anything, mock.Anything, false).Return(nil, nil)
				userRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			info:    entities.UserInfoInitSpec{Username: "user", Password: "password"},
			wantErr: errors.New("db error"),
		},
		{
			name: "password is hashed before create",
			prepareMocks: func(userRepo *mocks.UserInfoRepository, txm *mocks.TransactionManager) {
				txm.On("Do", mock.Anything, mock.Anything).
					Return(func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) })
				userRepo.On("Get", mock.Anything, mock.Anything, false).Return(nil, nil)
				userRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *entities.UserInfo) bool {
					return u.Password() != "secret" && bcrypt.CompareHashAndPassword([]byte(u.Password()), []byte("secret")) == nil
				})).Return(nil)
			},
			info:    entities.UserInfoInitSpec{Username: "user", Password: "secret"},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := mocks.NewUserInfoRepository(t)
			txm := mocks.NewTransactionManager(t)
			tt.prepareMocks(userRepo, txm)

			s := NewService(&Config{TransactionManager: txm, UserInfoRepository: userRepo})
			err := s.Register(ctx, tt.info)

			if tt.wantErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ──────────────────────────────────────────────────────────────
// Login
// ──────────────────────────────────────────────────────────────

func TestService_Login(t *testing.T) {
	ctx := context.Background()
	user := newHashedUser("user", "password")

	tests := []struct {
		name         string
		prepareMocks func(userRepo *mocks.UserInfoRepository, refreshRepo *mocks.UserRefreshTokensRepository)
		auth         entities.UserAuthInitSpec
		wantErr      error
	}{
		{
			name: "success — returns token pair",
			prepareMocks: func(userRepo *mocks.UserInfoRepository, refreshRepo *mocks.UserRefreshTokensRepository) {
				userRepo.On("Get", mock.Anything, dto.UserInfoFilter{Username: strPtr("user")}, false).Return(user, nil)
				refreshRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.UserRefreshToken")).Return(nil)
			},
			auth:    entities.UserAuthInitSpec{Username: "user", Password: "password"},
			wantErr: nil,
		},
		{
			name: "wrong password",
			prepareMocks: func(userRepo *mocks.UserInfoRepository, refreshRepo *mocks.UserRefreshTokensRepository) {
				userRepo.On("Get", mock.Anything, dto.UserInfoFilter{Username: strPtr("user")}, false).Return(user, nil)
			},
			auth:    entities.UserAuthInitSpec{Username: "user", Password: "wrong"},
			wantErr: autherrors.ErrInvalidCredentials,
		},
		{
			name: "user not found",
			prepareMocks: func(userRepo *mocks.UserInfoRepository, refreshRepo *mocks.UserRefreshTokensRepository) {
				userRepo.On("Get", mock.Anything, mock.Anything, false).Return(nil, autherrors.ErrUserInfoNotFound)
			},
			auth:    entities.UserAuthInitSpec{Username: "nobody", Password: "pass"},
			wantErr: autherrors.ErrUserInfoNotFound,
		},
		{
			name: "refresh token create fails",
			prepareMocks: func(userRepo *mocks.UserInfoRepository, refreshRepo *mocks.UserRefreshTokensRepository) {
				userRepo.On("Get", mock.Anything, mock.Anything, false).Return(user, nil)
				refreshRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			auth:    entities.UserAuthInitSpec{Username: "user", Password: "password"},
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := mocks.NewUserInfoRepository(t)
			refreshRepo := mocks.NewUserRefreshTokensRepository(t)
			tt.prepareMocks(userRepo, refreshRepo)

			s := NewService(&Config{
				UserInfoRepository:          userRepo,
				UserRefreshTokensRepository: refreshRepo,
			})

			pair, err := s.Login(ctx, tt.auth)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Empty(t, pair.AccessToken)
				assert.Empty(t, pair.RefreshToken)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, pair.AccessToken)
				assert.NotEmpty(t, pair.RefreshToken)
			}
		})
	}
}

// ──────────────────────────────────────────────────────────────
// Refresh
// ──────────────────────────────────────────────────────────────

func TestService_Refresh(t *testing.T) {
	ctx := context.Background()
	user := newHashedUser("user", "password")
	rawToken := "validrawtoken1234"

	tests := []struct {
		name         string
		rawToken     string
		prepareMocks func(userRepo *mocks.UserInfoRepository, refreshRepo *mocks.UserRefreshTokensRepository)
		wantErr      bool
	}{
		{
			name:     "success — token rotated",
			rawToken: rawToken,
			prepareMocks: func(userRepo *mocks.UserInfoRepository, refreshRepo *mocks.UserRefreshTokensRepository) {
				token := newActiveRefreshToken(user.ID(), rawToken)
				hash := hashToken(rawToken)
				tokenUserID := token.UserID()
				refreshRepo.On("Get", mock.Anything, dto.UserRefreshTokenFilter{TokenHash: &hash}).Return(token, nil)
				refreshRepo.On("Delete", mock.Anything, dto.UserRefreshTokenFilter{TokenHash: &hash}).Return(nil)
				userRepo.On("Get", mock.Anything, dto.UserInfoFilter{ID: &tokenUserID}, false).Return(user, nil)
				refreshRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.UserRefreshToken")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:     "token not found",
			rawToken: "unknowntoken",
			prepareMocks: func(userRepo *mocks.UserInfoRepository, refreshRepo *mocks.UserRefreshTokensRepository) {
				refreshRepo.On("Get", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
		{
			name:     "token expired",
			rawToken: rawToken,
			prepareMocks: func(userRepo *mocks.UserInfoRepository, refreshRepo *mocks.UserRefreshTokensRepository) {
				token := newExpiredRefreshToken(user.ID(), rawToken)
				hash := hashToken(rawToken)
				refreshRepo.On("Get", mock.Anything, dto.UserRefreshTokenFilter{TokenHash: &hash}).Return(token, nil)
				refreshRepo.On("Delete", mock.Anything, dto.UserRefreshTokenFilter{TokenHash: &hash}).Return(nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := mocks.NewUserInfoRepository(t)
			refreshRepo := mocks.NewUserRefreshTokensRepository(t)
			tt.prepareMocks(userRepo, refreshRepo)

			s := NewService(&Config{
				UserInfoRepository:          userRepo,
				UserRefreshTokensRepository: refreshRepo,
			})

			pair, err := s.Refresh(ctx, tt.rawToken)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, pair.AccessToken)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, pair.AccessToken)
				assert.NotEmpty(t, pair.RefreshToken)
			}
		})
	}
}

// ──────────────────────────────────────────────────────────────
// SendVerificationCode
// ──────────────────────────────────────────────────────────────

func TestService_SendVerificationCode(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	user := entities.NewUserInfo(entities.WithUserInfoInitSpec(entities.UserInfoInitSpec{
		ID: userID, Username: "user", Email: "user@example.com", Phone: "+79001234567",
	}))

	tests := []struct {
		name         string
		codeType     entities.VerificationCodeType
		prepareMocks func(userRepo *mocks.UserInfoRepository, vcRepo *mocks.UserVerificationCodesRepository, taskRepo *mocks.TasksRepository)
		wantErr      bool
	}{
		{
			name:     "email code sent successfully",
			codeType: entities.VerificationCodeEmail,
			prepareMocks: func(userRepo *mocks.UserInfoRepository, vcRepo *mocks.UserVerificationCodesRepository, taskRepo *mocks.TasksRepository) {
				userRepo.On("Get", mock.Anything, dto.UserInfoFilter{ID: &userID}, false).Return(user, nil)
				vcRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.UserVerificationCode")).Return(nil)
				taskRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Task")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:     "phone code sent successfully",
			codeType: entities.VerificationCodePhone,
			prepareMocks: func(userRepo *mocks.UserInfoRepository, vcRepo *mocks.UserVerificationCodesRepository, taskRepo *mocks.TasksRepository) {
				userRepo.On("Get", mock.Anything, dto.UserInfoFilter{ID: &userID}, false).Return(user, nil)
				vcRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.UserVerificationCode")).Return(nil)
				taskRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Task")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:     "user not found",
			codeType: entities.VerificationCodeEmail,
			prepareMocks: func(userRepo *mocks.UserInfoRepository, vcRepo *mocks.UserVerificationCodesRepository, taskRepo *mocks.TasksRepository) {
				userRepo.On("Get", mock.Anything, dto.UserInfoFilter{ID: &userID}, false).Return(nil, autherrors.ErrUserInfoNotFound)
			},
			wantErr: true,
		},
		{
			name:     "verification code repo error",
			codeType: entities.VerificationCodeEmail,
			prepareMocks: func(userRepo *mocks.UserInfoRepository, vcRepo *mocks.UserVerificationCodesRepository, taskRepo *mocks.TasksRepository) {
				userRepo.On("Get", mock.Anything, mock.Anything, false).Return(user, nil)
				vcRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := mocks.NewUserInfoRepository(t)
			vcRepo := mocks.NewUserVerificationCodesRepository(t)
			taskRepo := mocks.NewTasksRepository(t)
			tt.prepareMocks(userRepo, vcRepo, taskRepo)

			s := NewService(&Config{
				UserInfoRepository:          userRepo,
				VerificationCodesRepository: vcRepo,
				TasksRepository:             taskRepo,
			})

			err := s.SendVerificationCode(ctx, userID, tt.codeType)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ──────────────────────────────────────────────────────────────
// VerifyCode
// ──────────────────────────────────────────────────────────────

func TestService_VerifyCode(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	user := entities.NewUserInfo(entities.WithUserInfoInitSpec(entities.UserInfoInitSpec{ID: userID, Username: "user"}))
	validCode := "123456"
	codeType := entities.VerificationCodeEmail

	tests := []struct {
		name         string
		code         string
		prepareMocks func(userRepo *mocks.UserInfoRepository, vcRepo *mocks.UserVerificationCodesRepository)
		wantErr      error
	}{
		{
			name: "success — email verified",
			code: validCode,
			prepareMocks: func(userRepo *mocks.UserInfoRepository, vcRepo *mocks.UserVerificationCodesRepository) {
				vc := newVerificationCode(userID, validCode, codeType, false, false)
				vcRepo.On("GetLatest", mock.Anything, dto.UserVerificationCodeFilter{UserID: &userID, CodeType: &codeType}).Return(vc, nil)
				vcRepo.On("MarkUsed", mock.Anything, vc.ID()).Return(nil)
				userRepo.On("Get", mock.Anything, dto.UserInfoFilter{ID: &userID}, false).Return(user, nil)
				userRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *entities.UserInfo) bool {
					return u.IsEmailVerified()
				})).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "wrong code",
			code: "000000",
			prepareMocks: func(userRepo *mocks.UserInfoRepository, vcRepo *mocks.UserVerificationCodesRepository) {
				vc := newVerificationCode(userID, validCode, codeType, false, false)
				vcRepo.On("GetLatest", mock.Anything, mock.Anything).Return(vc, nil)
			},
			wantErr: autherrors.ErrInvalidVerificationCode,
		},
		{
			name: "code expired",
			code: validCode,
			prepareMocks: func(userRepo *mocks.UserInfoRepository, vcRepo *mocks.UserVerificationCodesRepository) {
				vc := newVerificationCode(userID, validCode, codeType, true, false)
				vcRepo.On("GetLatest", mock.Anything, mock.Anything).Return(vc, nil)
			},
			wantErr: autherrors.ErrVerificationCodeExpired,
		},
		{
			name: "code already used",
			code: validCode,
			prepareMocks: func(userRepo *mocks.UserInfoRepository, vcRepo *mocks.UserVerificationCodesRepository) {
				vc := newVerificationCode(userID, validCode, codeType, false, true)
				vcRepo.On("GetLatest", mock.Anything, mock.Anything).Return(vc, nil)
			},
			wantErr: autherrors.ErrVerificationCodeAlreadyUsed,
		},
		{
			name: "no code found",
			code: validCode,
			prepareMocks: func(userRepo *mocks.UserInfoRepository, vcRepo *mocks.UserVerificationCodesRepository) {
				vcRepo.On("GetLatest", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))
			},
			wantErr: autherrors.ErrInvalidVerificationCode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := mocks.NewUserInfoRepository(t)
			vcRepo := mocks.NewUserVerificationCodesRepository(t)
			tt.prepareMocks(userRepo, vcRepo)

			s := NewService(&Config{
				UserInfoRepository:          userRepo,
				VerificationCodesRepository: vcRepo,
			})

			err := s.VerifyCode(ctx, userID, tt.code, codeType)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ──────────────────────────────────────────────────────────────
// SendRecoveryCode
// ──────────────────────────────────────────────────────────────

func TestService_SendRecoveryCode(t *testing.T) {
	ctx := context.Background()
	email := "user@example.com"
	user := entities.NewUserInfo(entities.WithUserInfoInitSpec(entities.UserInfoInitSpec{
		ID: uuid.New(), Username: "user", Email: email,
	}))

	tests := []struct {
		name         string
		prepareMocks func(userRepo *mocks.UserInfoRepository, vcRepo *mocks.UserVerificationCodesRepository, taskRepo *mocks.TasksRepository)
		wantErr      bool
	}{
		{
			name: "success",
			prepareMocks: func(userRepo *mocks.UserInfoRepository, vcRepo *mocks.UserVerificationCodesRepository, taskRepo *mocks.TasksRepository) {
				userRepo.On("Get", mock.Anything, dto.UserInfoFilter{Email: &email}, false).Return(user, nil)
				vcRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.UserVerificationCode")).Return(nil)
				taskRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Task")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "user not found — silently returns nil (anti-enumeration)",
			prepareMocks: func(userRepo *mocks.UserInfoRepository, vcRepo *mocks.UserVerificationCodesRepository, taskRepo *mocks.TasksRepository) {
				userRepo.On("Get", mock.Anything, dto.UserInfoFilter{Email: &email}, false).Return(nil, autherrors.ErrUserInfoNotFound)
			},
			wantErr: false, // always 200
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := mocks.NewUserInfoRepository(t)
			vcRepo := mocks.NewUserVerificationCodesRepository(t)
			taskRepo := mocks.NewTasksRepository(t)
			tt.prepareMocks(userRepo, vcRepo, taskRepo)

			s := NewService(&Config{
				UserInfoRepository:          userRepo,
				VerificationCodesRepository: vcRepo,
				TasksRepository:             taskRepo,
			})

			err := s.SendRecoveryCode(ctx, email)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ──────────────────────────────────────────────────────────────
// ResetPassword
// ──────────────────────────────────────────────────────────────

func TestService_ResetPassword(t *testing.T) {
	ctx := context.Background()
	email := "user@example.com"
	user := entities.NewUserInfo(entities.WithUserInfoInitSpec(entities.UserInfoInitSpec{
		ID: uuid.New(), Username: "user", Email: email,
	}))
	validCode := "654321"
	codeType := entities.VerificationCodeRecover

	tests := []struct {
		name         string
		code         string
		prepareMocks func(userRepo *mocks.UserInfoRepository, vcRepo *mocks.UserVerificationCodesRepository, refreshRepo *mocks.UserRefreshTokensRepository)
		wantErr      error
	}{
		{
			name: "success — password updated, refresh tokens invalidated",
			code: validCode,
			prepareMocks: func(userRepo *mocks.UserInfoRepository, vcRepo *mocks.UserVerificationCodesRepository, refreshRepo *mocks.UserRefreshTokensRepository) {
				vc := newVerificationCode(user.ID(), validCode, codeType, false, false)
				userID := user.ID()
				userRepo.On("Get", mock.Anything, dto.UserInfoFilter{Email: &email}, false).Return(user, nil)
				vcRepo.On("GetLatest", mock.Anything, dto.UserVerificationCodeFilter{UserID: &userID, CodeType: &codeType}).Return(vc, nil)
				vcRepo.On("MarkUsed", mock.Anything, vc.ID()).Return(nil)
				userRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.UserInfo")).Return(nil)
				refreshRepo.On("DeleteByUser", mock.Anything, user.ID()).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "user not found",
			code: validCode,
			prepareMocks: func(userRepo *mocks.UserInfoRepository, vcRepo *mocks.UserVerificationCodesRepository, refreshRepo *mocks.UserRefreshTokensRepository) {
				userRepo.On("Get", mock.Anything, dto.UserInfoFilter{Email: &email}, false).Return(nil, autherrors.ErrUserInfoNotFound)
			},
			wantErr: autherrors.ErrInvalidCredentials,
		},
		{
			name: "invalid code",
			code: "000000",
			prepareMocks: func(userRepo *mocks.UserInfoRepository, vcRepo *mocks.UserVerificationCodesRepository, refreshRepo *mocks.UserRefreshTokensRepository) {
				vc := newVerificationCode(user.ID(), validCode, codeType, false, false)
				userID := user.ID()
				userRepo.On("Get", mock.Anything, dto.UserInfoFilter{Email: &email}, false).Return(user, nil)
				vcRepo.On("GetLatest", mock.Anything, dto.UserVerificationCodeFilter{UserID: &userID, CodeType: &codeType}).Return(vc, nil)
			},
			wantErr: autherrors.ErrInvalidVerificationCode,
		},
		{
			name: "code expired",
			code: validCode,
			prepareMocks: func(userRepo *mocks.UserInfoRepository, vcRepo *mocks.UserVerificationCodesRepository, refreshRepo *mocks.UserRefreshTokensRepository) {
				vc := newVerificationCode(user.ID(), validCode, codeType, true, false)
				userID := user.ID()
				userRepo.On("Get", mock.Anything, dto.UserInfoFilter{Email: &email}, false).Return(user, nil)
				vcRepo.On("GetLatest", mock.Anything, dto.UserVerificationCodeFilter{UserID: &userID, CodeType: &codeType}).Return(vc, nil)
			},
			wantErr: autherrors.ErrVerificationCodeExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := mocks.NewUserInfoRepository(t)
			vcRepo := mocks.NewUserVerificationCodesRepository(t)
			refreshRepo := mocks.NewUserRefreshTokensRepository(t)
			tt.prepareMocks(userRepo, vcRepo, refreshRepo)

			s := NewService(&Config{
				UserInfoRepository:          userRepo,
				VerificationCodesRepository: vcRepo,
				UserRefreshTokensRepository: refreshRepo,
			})

			err := s.ResetPassword(ctx, email, tt.code, "newpassword123")
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ──────────────────────────────────────────────────────────────
// hashesPassword
// ──────────────────────────────────────────────────────────────

func TestService_hashesPassword(t *testing.T) {
	s := &Service{}

	tests := []struct {
		name     string
		password string
	}{
		{"normal password", "mysecret"},
		{"empty password", ""},
		{"long password", "this_is_a_very_long_password_1234567890"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &entities.UserInfoInitSpec{Username: "user", Password: tt.password}
			err := s.hashesPassword(spec)
			assert.NoError(t, err)
			assert.NotEqual(t, tt.password, spec.Password)
			assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(spec.Password), []byte(tt.password)))
		})
	}
}
