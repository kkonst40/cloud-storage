package user

import (
	"context"
	"errors"
	"testing"

	"github.com/kkonst40/cloud-storage/backend/internal/domain"
	errs "github.com/kkonst40/cloud-storage/backend/internal/errors"
	"github.com/kkonst40/cloud-storage/backend/internal/service/user/mocks"
	"github.com/kkonst40/cloud-storage/backend/internal/storage"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestService_UserByName(t *testing.T) {
	type mockBehavior func(r *mocks.MockRepository, ctx context.Context, name string)

	type testCase struct {
		name         string
		username     string
		mock         mockBehavior
		expectedUser domain.User
		expectedErr  error
		wantErr      bool
	}

	ctx := context.Background()
	testUser := domain.User{
		ID:       1,
		Username: "test_user",
	}
	unexpectedErr := errors.New("db connection failure")

	tests := []testCase{
		{
			name:     "Success",
			username: "test_user",
			mock: func(r *mocks.MockRepository, ctx context.Context, name string) {
				r.EXPECT().
					GetByName(ctx, name).
					Return(testUser, nil).
					Times(1)
			},
			expectedUser: testUser,
			wantErr:      false,
		},
		{
			name:     "User Not Found",
			username: "non_existent",
			mock: func(r *mocks.MockRepository, ctx context.Context, name string) {
				r.EXPECT().
					GetByName(ctx, name).
					Return(domain.User{}, storage.ErrNotFound).
					Times(1)
			},
			expectedUser: domain.User{},
			expectedErr:  ErrNotFound,
			wantErr:      true,
		},
		{
			name:     "Unexpected Repository Error",
			username: "test_user",
			mock: func(r *mocks.MockRepository, ctx context.Context, name string) {
				r.EXPECT().
					GetByName(ctx, name).
					Return(domain.User{}, unexpectedErr).
					Times(1)
			},
			expectedUser: domain.User{},
			expectedErr:  errs.Wrap("UserService", "UserByName", unexpectedErr),
			wantErr:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			tc.mock(mockRepo, ctx, tc.username)
			srv := New(mockRepo)

			resUser, err := srv.UserByName(ctx, tc.username)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.expectedErr != nil {
					//assert.ErrorIs(t, err, tc.expectedErr)
					assert.Equal(t, tc.expectedErr.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedUser, resUser)
			}
		})
	}
}

func TestService_UserById(t *testing.T) {
	type mockBehavior func(r *mocks.MockRepository, ctx context.Context, id int64)

	type testCase struct {
		name         string
		userID       int64
		mock         mockBehavior
		expectedUser domain.User
		expectedErr  error
		wantErr      bool
	}

	ctx := context.Background()
	testUser := domain.User{
		ID:       666,
		Username: "test_user",
	}
	unexpectedErr := errors.New("database timeout")

	tests := []testCase{
		{
			name:   "Success",
			userID: 666,
			mock: func(r *mocks.MockRepository, ctx context.Context, id int64) {
				r.EXPECT().
					GetById(ctx, id).
					Return(testUser, nil).
					Times(1)
			},
			expectedUser: testUser,
			wantErr:      false,
		},
		{
			name:   "User Not Found",
			userID: 999,
			mock: func(r *mocks.MockRepository, ctx context.Context, id int64) {
				r.EXPECT().
					GetById(ctx, id).
					Return(domain.User{}, storage.ErrNotFound).
					Times(1)
			},
			expectedUser: domain.User{},
			expectedErr:  ErrNotFound,
			wantErr:      true,
		},
		{
			name:   "Unexpected Repository Error",
			userID: 666,
			mock: func(r *mocks.MockRepository, ctx context.Context, id int64) {
				r.EXPECT().
					GetById(ctx, id).
					Return(domain.User{}, unexpectedErr).
					Times(1)
			},
			expectedUser: domain.User{},
			expectedErr:  errs.Wrap("UserService", "UserById", unexpectedErr),
			wantErr:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			tc.mock(mockRepo, ctx, tc.userID)
			srv := New(mockRepo)

			resUser, err := srv.UserById(ctx, tc.userID)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.expectedErr != nil {
					//assert.ErrorIs(t, err, tc.expectedErr)
					assert.Equal(t, tc.expectedErr.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedUser, resUser)
			}
		})
	}
}

func TestService_CreateUser(t *testing.T) {
	type mockBehavior func(r *mocks.MockRepository, ctx context.Context, expectedName string)

	type testCase struct {
		name         string
		username     string
		password     string
		mock         mockBehavior
		expectedUser domain.User
		expectedErr  error
		wantErr      bool
	}

	ctx := context.Background()
	unexpectedErr := errors.New("db failure")

	tests := []testCase{
		{
			name:     "Success",
			username: "new_user",
			password: "secret_password",
			mock: func(r *mocks.MockRepository, ctx context.Context, expectedName string) {
				r.EXPECT().
					Create(ctx, gomock.Cond(func(x any) bool {
						u, ok := x.(domain.User)
						if !ok {
							return false
						}

						return u.Username == expectedName && u.Password != ""
					})).
					DoAndReturn(func(ctx context.Context, u domain.User) (domain.User, error) {
						u.ID = 100
						return u, nil
					}).
					Times(1)
			},
			expectedUser: domain.User{ID: 100, Username: "new_user"},
			wantErr:      false,
		},
		{
			name:     "User Already Exists (Duplicate)",
			username: "existing_user",
			password: "password123",
			mock: func(r *mocks.MockRepository, ctx context.Context, expectedName string) {
				r.EXPECT().
					Create(ctx, gomock.Any()).
					Return(domain.User{}, storage.ErrDuplicate).
					Times(1)
			},
			expectedUser: domain.User{},
			expectedErr:  ErrAlreadyExists,
			wantErr:      true,
		},
		{
			name:     "Unexpected Repository Error",
			username: "failed_user",
			password: "password123",
			mock: func(r *mocks.MockRepository, ctx context.Context, expectedName string) {
				r.EXPECT().
					Create(ctx, gomock.Any()).
					Return(domain.User{}, unexpectedErr).
					Times(1)
			},
			expectedUser: domain.User{},
			expectedErr:  errs.Wrap("UserService", "CreateUser", unexpectedErr),
			wantErr:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			tc.mock(mockRepo, ctx, tc.username)
			srv := New(mockRepo)

			resUser, err := srv.CreateUser(ctx, tc.username, tc.password)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.expectedErr != nil {
					//assert.ErrorIs(t, err, tc.expectedErr)
					assert.Equal(t, tc.expectedErr.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedUser.ID, resUser.ID)
				assert.Equal(t, tc.expectedUser.Username, resUser.Username)
				assert.NotEmpty(t, resUser.Password)
			}
		})
	}
}
