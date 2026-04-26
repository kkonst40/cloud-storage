package user

import (
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
	Create(user domain.User) (domain.User, error)
	IsExistsByName(name string) (bool, error)
	IsExistsByID(id int64) (bool, error)
	ByName(name string) (domain.User, error)
	ById(userId int64) (domain.User, error)
}

func New(userRepo Repository) *Service {
	return &Service{
		userRepo: userRepo,
	}
}

func (s *Service) CreateUser(name, pwd string) (domain.User, error) {
	const op = "CreateUser"

	exists, err := s.userRepo.IsExistsByName(name)
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

	user, err = s.userRepo.Create(user)
	if err != nil {
		if errors.Is(err, storage.ErrDuplicate) {
			return domain.User{}, ErrAlreadyExists
		}

		return domain.User{}, errs.Wrap(pkg, op, err)
	}

	return user, nil
}

func (s *Service) UserByName(name string) (domain.User, error) {
	const op = "UserEmail"

	u, err := s.userRepo.ByName(name)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return domain.User{}, ErrNotFound
		}

		return domain.User{}, errs.Wrap(pkg, op, err)
	}

	return u, nil
}

func (s *Service) UserById(userId int64) (domain.User, error) {
	const op = "UserById"

	user, err := s.userRepo.ById(userId)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return domain.User{}, ErrNotFound
		}

		return domain.User{}, errs.Wrap(pkg, op, err)
	}

	return user, nil
}
