//go:generate mockgen -source=service.go -destination=mocks/mock_repository.go -package=mocks
package user

import (
	"context"
	"errors"

	"github.com/kkonst40/cloud-storage/backend/internal/domain"
	errs "github.com/kkonst40/cloud-storage/backend/internal/errors"
	"github.com/kkonst40/cloud-storage/backend/internal/service/password"
	"github.com/kkonst40/cloud-storage/backend/internal/storage"
)

var (
	ErrNotFound      = errors.New("user not found")
	ErrAlreadyExists = errors.New("user already exists")
)

const pkg = "UserService"

type Service struct {
	userRepo Repository
}

type Repository interface {
	GetByName(ctx context.Context, name string) (domain.User, error)
	GetById(ctx context.Context, userId int64) (domain.User, error)
	Create(ctx context.Context, user domain.User) (domain.User, error)
}

func New(userRepo Repository) *Service {
	return &Service{
		userRepo: userRepo,
	}
}

func (s *Service) UserByName(ctx context.Context, name string) (domain.User, error) {
	const op = "UserByName"

	user, err := s.userRepo.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return domain.User{}, ErrNotFound
		}

		return domain.User{}, errs.Wrap(pkg, op, err)
	}

	return user, nil
}

func (s *Service) UserById(ctx context.Context, userId int64) (domain.User, error) {
	const op = "UserById"

	user, err := s.userRepo.GetById(ctx, userId)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return domain.User{}, ErrNotFound
		}

		return domain.User{}, errs.Wrap(pkg, op, err)
	}

	return user, nil
}

func (s *Service) CreateUser(ctx context.Context, name, pwd string) (domain.User, error) {
	const op = "CreateUser"

	hashedPassword, err := password.GeneratePwdHash(pwd)
	if err != nil {
		return domain.User{}, errs.Wrap(pkg, op, err)
	}

	user := domain.User{
		Username: name,
		Password: hashedPassword,
	}

	user, err = s.userRepo.Create(ctx, user)
	if err != nil {
		if errors.Is(err, storage.ErrDuplicate) {
			return domain.User{}, ErrAlreadyExists
		}

		return domain.User{}, errs.Wrap(pkg, op, err)
	}

	return user, nil
}
