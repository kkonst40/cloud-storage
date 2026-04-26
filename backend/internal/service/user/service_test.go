package user

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/kkonst40/cloud-storage/backend/internal/config"
	"github.com/kkonst40/cloud-storage/backend/internal/domain"
	"github.com/kkonst40/cloud-storage/backend/internal/storage"
	"github.com/kkonst40/cloud-storage/backend/internal/storage/user"
)

type testService struct {
	service *Service
	db      *sql.DB
}

func TestUserService_CreateUser(t *testing.T) {
	userService := userTestService(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			panic(err)
		}
	}(userService.db)

	user := domain.User{
		Username: "test@example.ru",
		Password: "1234",
	}

	u1, err := userService.service.CreateUser(user.Username, user.Password)
	if err != nil {
		t.Errorf("error while create new user: %v", err)
	}
	defer func(db *sql.DB, userId int64) {
		err := deleteTestUser(db, userId)
		if err != nil {
			t.Errorf("error while delete test user: %v", err)
		}
	}(userService.db, u1.ID)

	if u1.Username != user.Username {
		t.Errorf("email does not match")
	}
}

func TestUserService_CreateUserDuplicate(t *testing.T) {
	userService := userTestService(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			panic(err)
		}
	}(userService.db)

	user := domain.User{
		Username: "test@example.ru",
		Password: "1234",
	}

	u1, err := userService.service.CreateUser(user.Username, user.Password)
	if err != nil {
		t.Errorf("error while create new user: %v", err)
	}
	defer func(db *sql.DB, userId int64) {
		err := deleteTestUser(db, userId)
		if err != nil {
			t.Errorf("error while delete test user: %v", err)
		}
	}(userService.db, u1.ID)

	// check when trying to create duplicate user
	u2, err := userService.service.CreateUser(user.Username, user.Password)
	if err != nil {
		if !errors.Is(err, ErrAlreadyExists) {
			t.Errorf("error while create duplicate user: %v", err)
		}
	} else {
		deleteTestUser(userService.db, u2.ID)
		t.Error("create duplicate user not allowed")
	}
}

func TestUserService_UserByName(t *testing.T) {
	userService := userTestService(t)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			panic(err)
		}
	}(userService.db)

	name := "test@example.ru"

	u1, err := userService.service.CreateUser(name, "1234")
	if err != nil {
		t.Errorf("error while create new user: %v", err)
	}
	defer func(db *sql.DB, userId int64) {
		err := deleteTestUser(db, userId)
		if err != nil {
			t.Errorf("error while delete test user: %v", err)
		}
	}(userService.db, u1.ID)

	u2, err := userService.service.UserByName(name)
	if err != nil {
		t.Errorf("error while get user by email: %v", err)
	}

	if u2.ID != u1.ID {
		t.Errorf("user by email not match")
	}
}

func deleteTestUser(db *sql.DB, userId int64) error {
	stmt, err := db.Prepare("DELETE FROM users WHERE id = $1")
	if err != nil {
		return err
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			return
		}
	}(stmt)
	_, err = stmt.Exec(userId)

	return err
}

func userTestService(t *testing.T) *testService {
	dir, err := findProjectRoot()
	if err != nil {
		t.Fatal(err)
	}

	cfg := config.MustNew(filepath.Join(dir, ".env.dev"))
	dbClient, err := storage.NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}

	db := dbClient.DB()

	return &testService{
		service: New(user.NewRepository(db)),
		db:      db,
	}
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}
