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
	Create(ctx context.Context, user domain.User) (domain.User, error)
	IsExistsByName(ctx context.Context, name string) (bool, error)
	IsExistsByID(ctx context.Context, id int64) (bool, error)
	ByName(ctx context.Context, name string) (domain.User, error)
	ById(ctx context.Context, userId int64) (domain.User, error)
}

func New(userRepo Repository) *Service {
	return &Service{
		userRepo: userRepo,
	}
}

func (s *Service) CreateUser(ctx context.Context, name, pwd string) (domain.User, error) {
	const op = "CreateUser"

	exists, err := s.userRepo.IsExistsByName(ctx, name)
	if err != nil {
		return domain.User{}, err
	}

	if exists {
		return domain.User{}, ErrAlreadyExists
	}

	hashedPassword, err := password.GeneratePwdHash(pwd)
	if err != nil {
		return domain.User{}, err
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

func (s *Service) UserByName(ctx context.Context, name string) (domain.User, error) {
	const op = "UserEmail"

	u, err := s.userRepo.ByName(ctx, name)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return domain.User{}, ErrNotFound
		}

		return domain.User{}, errs.Wrap(pkg, op, err)
	}

	return u, nil
}

func (s *Service) UserById(ctx context.Context, userId int64) (domain.User, error) {
	const op = "UserById"

	user, err := s.userRepo.ById(ctx, userId)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return domain.User{}, ErrNotFound
		}

		return domain.User{}, errs.Wrap(pkg, op, err)
	}

	return user, nil
}
