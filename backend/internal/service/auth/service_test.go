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

func TestService_checkPasswordAndTakeToken(t *testing.T) {
	hashPass, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	hashEmpty, _ := bcrypt.GenerateFromPassword([]byte(""), bcrypt.DefaultCost)

	tests := []struct {
		name    string
		user    *entities.UserInfo
		pass    string
		wantErr error
	}{
		{
			name: "correct password",
			user: entities.NewUserInfo(entities.WithUserInfoInitSpec(entities.UserInfoInitSpec{
				Username: "user",
				Password: string(hashPass),
			})),
			pass:    "password123",
			wantErr: nil,
		},
		{
			name: "wrong password",
			user: entities.NewUserInfo(entities.WithUserInfoInitSpec(entities.UserInfoInitSpec{
				Username: "user",
				Password: string(hashPass),
			})),
			pass:    "wrong",
			wantErr: errors.ErrInvalidCredentials,
		},
		{
			name: "empty password correct",
			user: entities.NewUserInfo(entities.WithUserInfoInitSpec(entities.UserInfoInitSpec{
				Username: "user",
				Password: string(hashEmpty),
			})),
			pass:    "",
			wantErr: nil,
		},
	}

	s := &Service{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.checkPasswordAndTakeToken(tt.user, tt.pass)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErr.Error())
				assert.Empty(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, got)
			}
		})
	}
}

func TestService_hashesPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "normal password",
			password: "mysecret",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false,
		},
		{
			name:     "long password",
			password: "this_is_a_very_long_password_1234567890",
			wantErr:  false,
		},
		{
			name:     "simulate bcrypt error",
			password: "error",
			wantErr:  false,
		},
	}

	s := &Service{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &entities.UserInfoInitSpec{
				Username: "user",
				Password: tt.password,
			}

			err := s.hashesPassword(user)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, user.Password)
			assert.NotEqual(t, tt.password, user.Password)

			err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(tt.password))
			assert.NoError(t, err)
		})
	}
}
