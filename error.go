package gone

import (
	"errors"
	"fmt"
)

var GonerIdIsExistedError = errors.New("goner id is existed")

type CannotFoundGonerByIdError struct {
	Id GonerId
}

func (e *CannotFoundGonerByIdError) Error() string {
	return fmt.Sprintf("cannot found the Goner by Id(%s)", e.Id)
}
func newCannotFoundGonerById(id GonerId) *CannotFoundGonerByIdError {
	return &CannotFoundGonerByIdError{Id: id}
}

type NotCompatibleGonerError struct {
	Id GonerId
}

func (e *NotCompatibleGonerError) Error() string {
	return fmt.Sprintf("Id(%s) goner is not compatible", e.Id)
}

func newNotCompatibleGonerError(id GonerId) *NotCompatibleGonerError {
	return &NotCompatibleGonerError{Id: id}
}
