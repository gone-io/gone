package gone

import (
	"fmt"
	"reflect"
)

type Error struct {
	Code  int
	Msg   string
	trace []byte
}

func (e *Error) Error() string {
	msg := fmt.Sprintf("GoneError(code=%v):%s", e.Code, e.Msg)
	if e.trace == nil {
		return msg
	}
	return fmt.Sprintf("%s\n%s", msg, e.trace)
}

func NewError(code int, msg string) *Error {
	return &Error{Code: code, Msg: msg, trace: PanicTrace(2)}
}

const (
	GonerIdIsExisted = 1001 + iota
	CannotFoundGonerById
	CannotFoundGonerByType
	NotCompatible
	ReplaceBuryIdParamEmpty
)

func GonerIdIsExistedError(id GonerId) *Error {
	return NewError(GonerIdIsExisted, fmt.Sprintf("Goner Id(%s) is existed", id))
}

func CannotFoundGonerByIdError(id GonerId) *Error {
	return NewError(CannotFoundGonerById, fmt.Sprintf("Cannot found the Goner by Id(%s)", id))
}

func CannotFoundGonerByTypeError(t reflect.Type) *Error {
	return NewError(CannotFoundGonerByType, fmt.Sprintf("Cannot found the Goner by Type(%s)", t.Name()))
}

func NotCompatibleError(a reflect.Type, b reflect.Type) *Error {
	return NewError(NotCompatible, fmt.Sprintf("Not compatible: %s/%s vs %s/%s", a.PkgPath(), a.Name(), b.PkgPath(), b.Name()))
}

func ReplaceBuryIdParamEmptyError() *Error {
	return NewError(ReplaceBuryIdParamEmpty, "ReplaceBury id cannot be empty")
}
