package storage

import (
	"errors"
)

var (
	ErrNotFound  = errors.New("sql: no rows in result set")
	ErrDuplicate = errors.New("sql: duplicate not allowed")
)
