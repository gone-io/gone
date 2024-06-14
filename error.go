package gone

import (
	"fmt"
	"net/http"
	"reflect"
)

// BError Business error implementation
type BError struct {
	err  Error
	data any
}

func (e *BError) Msg() string {
	return e.err.Msg()
}
func (e *BError) Code() int {
	return e.err.Code()
}
func (e *BError) Error() string {
	return e.err.Error()
}
func (e *BError) Data() any {
	return e.data
}

type defaultErr struct {
	code int
	msg  string
}

func (e *defaultErr) Error() string {
	return fmt.Sprintf("GoneError(code=%v):%s", e.Code(), e.Msg())
}

func (e *defaultErr) Msg() string {
	return e.msg
}
func (e *defaultErr) Code() int {
	return e.code
}

// NewError create a error
func NewError(code int, msg string) Error {
	return &defaultErr{code: code, msg: msg}
}

// NewParameterError create a Parameter error
func NewParameterError(msg string, ext ...int) Error {
	var code = http.StatusBadRequest
	if len(ext) > 0 {
		code = ext[0]
	}
	return NewError(code, msg)
}

// NewBusinessError create a business error
func NewBusinessError(msg string, ext ...any) BusinessError {
	var code = 0
	var data any = nil
	if len(ext) > 0 {
		i, ok := ext[0].(int)
		if ok {
			code = i
		}
	}
	if len(ext) > 1 {
		data = ext[1]
	}

	return &BError{err: NewError(code, msg), data: data}
}

// ToError translate any type to An Error
func ToError(input any) Error {
	if input == nil {
		return nil
	}
	switch input.(type) {
	case Error:
		return input.(Error)
	case error:
		return NewInnerError(input.(error).Error(), http.StatusInternalServerError)
	case string:
		return NewInnerError(input.(string), http.StatusInternalServerError)
	default:
		return NewInnerError(fmt.Sprintf("%v", input), http.StatusInternalServerError)
	}
}

type iError struct {
	*defaultErr
	trace []byte
}

func (e *iError) Error() string {
	msg := e.defaultErr.Error()
	return fmt.Sprintf("%s\n%s", msg, e.trace)
}

func (e *iError) Stack() []byte {
	return e.trace
}

func NewInnerError(msg string, code int) Error {
	return NewInnerErrorSkip(msg, code, 1)
}

func NewInnerErrorSkip(msg string, code int, skip int) Error {
	return &iError{defaultErr: &defaultErr{code: code, msg: msg}, trace: PanicTrace(2, skip)}
}

// Error Code：1001~1999 used for gone framework.
const (
	// GonerIdIsExisted Goner for the GonerId is existed.
	GonerIdIsExisted = 1001 + iota

	// CannotFoundGonerById cannot find the Goner by the GonerId.
	CannotFoundGonerById

	// CannotFoundGonerByType cannot find the Goner by the Type.
	CannotFoundGonerByType

	//NotCompatible Goner is not compatible with the Type.
	NotCompatible

	//ReplaceBuryIdParamEmpty Cemetery.ReplaceBury error for the GonerId is empty.
	ReplaceBuryIdParamEmpty

	//StartError Gone Start flow error.
	StartError

	//StopError Gone Stop flow error.
	StopError

	//DbRollForPanic error in rollback of DB transaction  for panic.
	DbRollForPanic

	//MustHaveGonerId error for the GonerId is empty.
	MustHaveGonerId

	//InjectError error for dependence injection error
	InjectError
)

func GonerIdIsExistedError(id GonerId) Error {
	return NewInnerErrorSkip(fmt.Sprintf("Goner Id(%s) is existed", id), GonerIdIsExisted, 2)
}

func CannotFoundGonerByIdError(id GonerId) Error {
	return NewInnerErrorSkip(fmt.Sprintf("Cannot found the Goner by Id(%s)", id), CannotFoundGonerById, 2)
}

func CannotFoundGonerByTypeError(t reflect.Type) Error {
	return NewInnerErrorSkip(fmt.Sprintf("Cannot found the Goner by Type(%s)", t.Name()), CannotFoundGonerByType, 2)
}

func NotCompatibleError(a reflect.Type, b reflect.Type) Error {
	return NewInnerErrorSkip(fmt.Sprintf("Not compatible: %s/%s vs %s/%s", a.PkgPath(), a.Name(), b.PkgPath(), b.Name()), NotCompatible, 2)
}

func ReplaceBuryIdParamEmptyError() Error {
	return NewInnerErrorSkip("ReplaceBury id cannot be empty", ReplaceBuryIdParamEmpty, 2)
}
