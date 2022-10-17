package gone

import (
	"errors"
	"fmt"
	"reflect"
)

var GonerIdIsExistedError = errors.New("goner id is existed")

type CannotFoundGonerByIdError struct {
	Id    GonerId
	trace []byte
}

func (e *CannotFoundGonerByIdError) Error() string {
	return fmt.Sprintf("cannot found the Goner by Id(%s)\n%s", e.Id, e.trace)
}
func newCannotFoundGonerById(id GonerId) *CannotFoundGonerByIdError {
	return &CannotFoundGonerByIdError{Id: id, trace: PanicTrace(2)}
}

type NotCompatibleGonerError struct {
	Id    GonerId
	Type  reflect.Type
	trace []byte
}

func (e *NotCompatibleGonerError) Error() string {
	if e.Id != "" {
		return fmt.Sprintf("Id(%s) goner is not compatible\n%s", e.Id, e.trace)
	}
	return fmt.Sprintf("Type(%s) goner is not compatible\n%s", e.Type.Name(), e.trace)
}

func newNotCompatibleGonerError(id GonerId) *NotCompatibleGonerError {
	return &NotCompatibleGonerError{Id: id, trace: PanicTrace(2)}
}

func newNotCompatibleGonerErrorByType(t reflect.Type) *NotCompatibleGonerError {
	return &NotCompatibleGonerError{Type: t, trace: PanicTrace(2)}
}
