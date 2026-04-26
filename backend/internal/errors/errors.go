package errors

import "fmt"

func Wrap(pkg, op string, err error) error {
	return fmt.Errorf("%v -> %v: %w", pkg, op, err)
}
