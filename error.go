package gone

import (
	"fmt"
	"reflect"
)

type defaultErr struct {
	code int
	msg  string
}

func (e *defaultErr) Error() string {
	return fmt.Sprintf("GoneError(code=%v):%s", e.Code, e.Msg)
}

func (e *defaultErr) Msg() string {
	return e.msg
}
func (e *defaultErr) Code() int {
	return e.code
}

func NewError(code int, msg string) Error {
	return &defaultErr{code: code, msg: msg}
}

type iError struct {
	*defaultErr
	trace []byte
}

func (e *iError) Error() string {
	msg := e.defaultErr.Error()
	if e.trace == nil {
		return msg
	}
	return fmt.Sprintf("%s\n%s", msg, e.trace)
}

func (e *iError) Stack() []byte {
	return e.trace
}

func NewInnerError(code int, msg string) Error {
	return &iError{defaultErr: &defaultErr{code: code, msg: msg}, trace: PanicTrace(2)}
}

const (
	GonerIdIsExisted = 1001 + iota
	CannotFoundGonerById
	CannotFoundGonerByType
	NotCompatible
	ReplaceBuryIdParamEmpty
	GonerDoNotInitCompletely
)

func GonerIdIsExistedError(id GonerId) Error {
	return NewInnerError(GonerIdIsExisted, fmt.Sprintf("Goner Id(%s) is existed", id))
}

func CannotFoundGonerByIdError(id GonerId) Error {
	return NewInnerError(CannotFoundGonerById, fmt.Sprintf("Cannot found the Goner by Id(%s)", id))
}

func CannotFoundGonerByTypeError(t reflect.Type) Error {
	return NewInnerError(CannotFoundGonerByType, fmt.Sprintf("Cannot found the Goner by Type(%s)", t.Name()))
}

func NotCompatibleError(a reflect.Type, b reflect.Type) Error {
	return NewInnerError(NotCompatible, fmt.Sprintf("Not compatible: %s/%s vs %s/%s", a.PkgPath(), a.Name(), b.PkgPath(), b.Name()))
}

func ReplaceBuryIdParamEmptyError() Error {
	return NewInnerError(ReplaceBuryIdParamEmpty, "ReplaceBury id cannot be empty")
}

func GonerDoNotInitCompletelyError(id GonerId) Error {
	return NewInnerError(GonerDoNotInitCompletely, fmt.Sprintf("gone(%s) do not init completely", id))
}
