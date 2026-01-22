package auth

import (
	"backend/internal/domain/entities"
	"backend/internal/dto"
	"backend/internal/errors"
	"backend/internal/service/auth/mocks"
	"context"
	"golang.org/x/crypto/bcrypt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func strPtr(s string) *string {
	return &s
}

func TestService_Register(t *testing.T) {
	ctx := context.Background()

	type fields struct {
		prepareMocks func(
			userRepo *mocks.UserInfoRepository,
			txm *mocks.TransactionManager,
		)
	}

	type args struct {
		info entities.UserInfoInitSpec
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "success register",
			fields: fields{
				prepareMocks: func(
					userRepo *mocks.UserInfoRepository,
					txm *mocks.TransactionManager,
				) {
					txm.
						On("Do", mock.Anything, mock.Anything).
						Return(func(ctx context.Context, fn func(ctx context.Context) error) error {
							return fn(ctx)
						})

					userRepo.
						On("Get", mock.Anything,
							dto.UserInfoFilter{Username: strPtr("user")},
							false,
						).
						Return(nil, nil)

					userRepo.
						On("Create", mock.Anything, mock.AnythingOfType("*entities.UserInfo")).
						Return(nil)
				},
			},
			args: args{
				info: entities.UserInfoInitSpec{
					Username: "user",
					Password: "password",
				},
			},
			wantErr: nil,
		},
		{
			name: "user already exists",
			fields: fields{
				prepareMocks: func(
					userRepo *mocks.UserInfoRepository,
					txm *mocks.TransactionManager,
				) {
					txm.
						On("Do", mock.Anything, mock.Anything).
						Return(func(ctx context.Context, fn func(ctx context.Context) error) error {
							return fn(ctx)
						})

					userRepo.
						On("Get", mock.Anything,
							dto.UserInfoFilter{Username: strPtr("user")},
							false,
						).
						Return(&entities.UserInfo{}, nil)
				},
			},
			args: args{
				info: entities.UserInfoInitSpec{
					Username: "user",
					Password: "password",
				},
			},
			wantErr: errors.ErrUserAlreadyExists,
		},
		{
			name: "repo create error",
			fields: fields{
				prepareMocks: func(userRepo *mocks.UserInfoRepository, txm *mocks.TransactionManager) {
					txm.On("Do", mock.Anything, mock.Anything).
						Return(func(ctx context.Context, fn func(ctx context.Context) error) error {
							return fn(ctx)
						})

					userRepo.
						On("Get", mock.Anything, dto.UserInfoFilter{Username: strPtr("user")}, false).
						Return(nil, nil)

					userRepo.
						On("Create", mock.Anything, mock.Anything).
						Return(errors.ErrHashedPassword)
				},
			},
			args: args{
				info: entities.UserInfoInitSpec{
					Username: "user",
					Password: "password",
				},
			},
			wantErr: errors.ErrHashedPassword,
		},
		{
			name: "transaction manager error",
			fields: fields{
				prepareMocks: func(userRepo *mocks.UserInfoRepository, txm *mocks.TransactionManager) {
					txm.
						On("Do", mock.Anything, mock.Anything).
						Return(errors.ErrHashedPassword)
				},
			},
			args: args{
				info: entities.UserInfoInitSpec{
					Username: "user",
					Password: "password",
				},
			},
			wantErr: errors.ErrHashedPassword,
		},
		{
			name: "password is hashed before create",
			fields: fields{
				prepareMocks: func(userRepo *mocks.UserInfoRepository, txm *mocks.TransactionManager) {
					txm.On("Do", mock.Anything, mock.Anything).
						Return(func(ctx context.Context, fn func(ctx context.Context) error) error {
							return fn(ctx)
						})

					userRepo.
						On("Get", mock.Anything, dto.UserInfoFilter{Username: strPtr("user")}, false).
						Return(nil, nil)

					userRepo.
						On("Create", mock.Anything, mock.MatchedBy(func(u *entities.UserInfo) bool {
							return u.Password() != "password"
						})).
						Return(nil)
				},
			},
			args: args{
				info: entities.UserInfoInitSpec{
					Username: "user",
					Password: "password",
				},
			},
			wantErr: nil,
		},
		{
			name: "username with spaces",
			fields: fields{
				prepareMocks: func(userRepo *mocks.UserInfoRepository, txm *mocks.TransactionManager) {
					txm.On("Do", mock.Anything, mock.Anything).
						Return(func(ctx context.Context, fn func(ctx context.Context) error) error {
							return fn(ctx)
						})

					userRepo.
						On("Get", mock.Anything, dto.UserInfoFilter{Username: strPtr(" user ")}, false).
						Return(nil, nil)

					userRepo.
						On("Create", mock.Anything, mock.Anything).
						Return(nil)
				},
			},
			args: args{
				info: entities.UserInfoInitSpec{
					Username: " user ",
					Password: "password",
				},
			},
			wantErr: nil,
		},
		{
			name: "empty password",
			fields: fields{
				prepareMocks: func(userRepo *mocks.UserInfoRepository, txm *mocks.TransactionManager) {
					txm.On("Do", mock.Anything, mock.Anything).
						Return(func(ctx context.Context, fn func(ctx context.Context) error) error {
							return fn(ctx)
						})

					userRepo.
						On("Get", mock.Anything, dto.UserInfoFilter{Username: strPtr("user")}, false).
						Return(nil, nil)

					userRepo.
						On("Create", mock.Anything, mock.Anything).
						Return(nil)
				},
			},
			args: args{
				info: entities.UserInfoInitSpec{
					Username: "user",
					Password: "",
				},
			},
			wantErr: nil,
		},
		{
			name: "create not called if user exists",
			fields: fields{
				prepareMocks: func(userRepo *mocks.UserInfoRepository, txm *mocks.TransactionManager) {
					txm.On("Do", mock.Anything, mock.Anything).
						Return(func(ctx context.Context, fn func(ctx context.Context) error) error {
							return fn(ctx)
						})

					userRepo.
						On("Get", mock.Anything, dto.UserInfoFilter{Username: strPtr("user")}, false).
						Return(&entities.UserInfo{}, nil)
				},
			},
			args: args{
				info: entities.UserInfoInitSpec{
					Username: "user",
					Password: "password",
				},
			},
			wantErr: errors.ErrUserAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := mocks.NewUserInfoRepository(t)
			txm := mocks.NewTransactionManager(t)

			if tt.fields.prepareMocks != nil {
				tt.fields.prepareMocks(userRepo, txm)
			}

			s := NewService(&Config{
				TransactionManager: txm,
				UserInfoRepository: userRepo,
			})

			err := s.Register(ctx, tt.args.info)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_Login(t *testing.T) {
	ctx := context.Background()

	hashedPass, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	info := entities.UserInfoInitSpec{
		Username: "user",
		Password: string(hashedPass),
	}
	user := entities.NewUserInfo(entities.WithUserInfoInitSpec(info))

	type fields struct {
		prepareMocks func(userRepo *mocks.UserInfoRepository)
	}

	type args struct {
		auth entities.UserAuthInitSpec
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "success login",
			fields: fields{
				prepareMocks: func(userRepo *mocks.UserInfoRepository) {
					userRepo.
						On("Get", mock.Anything,
							dto.UserInfoFilter{Username: strPtr("user")},
							false,
						).
						Return(user, nil)
				},
			},
			args: args{
				auth: entities.UserAuthInitSpec{
					Username: "user",
					Password: "password",
				},
			},
			wantErr: nil,
		},
		{
			name: "invalid password",
			fields: fields{
				prepareMocks: func(userRepo *mocks.UserInfoRepository) {
					userRepo.
						On("Get", mock.Anything,
							dto.UserInfoFilter{Username: strPtr("user")},
							false,
						).
						Return(user, nil)
				},
			},
			args: args{
				auth: entities.UserAuthInitSpec{
					Username: "user",
					Password: "wrong",
				},
			},
			wantErr: errors.ErrInvalidCredentials,
		},
		{
			name: "user not found",
			fields: fields{
				prepareMocks: func(userRepo *mocks.UserInfoRepository) {
					userRepo.
						On("Get", mock.Anything,
							dto.UserInfoFilter{Username: strPtr("user")},
							false,
						).
						Return(nil, errors.ErrUserInfoNotFound)
				},
			},
			args: args{
				auth: entities.UserAuthInitSpec{
					Username: "user",
					Password: "password",
				},
			},
			wantErr: errors.ErrUserInfoNotFound,
		},
		{
			name: "repo error",
			fields: fields{
				prepareMocks: func(userRepo *mocks.UserInfoRepository) {
					userRepo.
						On("Get", mock.Anything, dto.UserInfoFilter{Username: strPtr("user")}, false).
						Return(nil, errors.ErrHashedPassword)
				},
			},
			args: args{
				auth: entities.UserAuthInitSpec{
					Username: "user",
					Password: "password",
				},
			},
			wantErr: errors.ErrHashedPassword,
		},
		{
			name: "empty password",
			fields: fields{
				prepareMocks: func(userRepo *mocks.UserInfoRepository) {
					userRepo.
						On("Get", mock.Anything, dto.UserInfoFilter{Username: strPtr("user")}, false).
						Return(user, nil)
				},
			},
			args: args{
				auth: entities.UserAuthInitSpec{
					Username: "user",
					Password: "",
				},
			},
			wantErr: errors.ErrInvalidCredentials,
		},
		{
			name: "empty username",
			fields: fields{
				prepareMocks: func(userRepo *mocks.UserInfoRepository) {
					userRepo.
						On("Get", mock.Anything, dto.UserInfoFilter{Username: strPtr("")}, false).
						Return(nil, errors.ErrUserInfoNotFound)
				},
			},
			args: args{
				auth: entities.UserAuthInitSpec{
					Username: "",
					Password: "password",
				},
			},
			wantErr: errors.ErrUserInfoNotFound,
		},
		{
			name: "invalid stored hash",
			fields: fields{
				prepareMocks: func(userRepo *mocks.UserInfoRepository) {
					badUser := entities.NewUserInfo(entities.WithUserInfoInitSpec(entities.UserInfoInitSpec{
						Username: "user",
						Password: "not-a-hash",
					}))

					userRepo.
						On("Get", mock.Anything, dto.UserInfoFilter{Username: strPtr("user")}, false).
						Return(badUser, nil)
				},
			},
			args: args{
				auth: entities.UserAuthInitSpec{
					Username: "user",
					Password: "password",
				},
			},
			wantErr: errors.ErrInvalidCredentials,
		},
		{
			name: "second login returns token",
			fields: fields{
				prepareMocks: func(userRepo *mocks.UserInfoRepository) {
					userRepo.
						On("Get", mock.Anything, dto.UserInfoFilter{Username: strPtr("user")}, false).
						Return(user, nil)
				},
			},
			args: args{
				auth: entities.UserAuthInitSpec{
					Username: "user",
					Password: "password",
				},
			},
			wantErr: nil,
		},
		{
			name: "username with spaces",
			fields: fields{
				prepareMocks: func(userRepo *mocks.UserInfoRepository) {
					userRepo.
						On("Get", mock.Anything, dto.UserInfoFilter{Username: strPtr(" user ")}, false).
						Return(nil, errors.ErrUserInfoNotFound)
				},
			},
			args: args{
				auth: entities.UserAuthInitSpec{
					Username: " user ",
					Password: "password",
				},
			},
			wantErr: errors.ErrUserInfoNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := mocks.NewUserInfoRepository(t)

			if tt.fields.prepareMocks != nil {
				tt.fields.prepareMocks(userRepo)
			}

			s := NewService(&Config{
				UserInfoRepository: userRepo,
			})

			token, err := s.Login(ctx, tt.args.auth)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}
