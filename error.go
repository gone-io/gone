package gone

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime"
)

// Error normal error
type Error interface {
	error
	Msg() string
	SetMsg(msg string)
	Code() int

	GetStatusCode() int
}

// InnerError which has stack, and which is used for Internal error
type InnerError interface {
	Error
	Stack() []byte
}

// BusinessError which has data, and which is used for Business error
type BusinessError interface {
	Error
	Data() any
}

// BError Business error implementation
type BError struct {
	err  Error
	data any
}

func (e *BError) SetMsg(msg string) {
	e.err.SetMsg(msg)
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
func (e *BError) GetStatusCode() int {
	return e.err.GetStatusCode()
}

func (e *BError) Data() any {
	return e.data
}

type defaultErr struct {
	code       int
	msg        string
	statusCode int
}

func (e *defaultErr) Error() string {
	return fmt.Sprintf("GoneError(code=%v); %s", e.Code(), e.Msg())
}

func (e *defaultErr) Msg() string {
	return e.msg
}
func (e *defaultErr) SetMsg(msg string) {
	e.msg = msg
}

func (e *defaultErr) Code() int {
	return e.code
}

func (e *defaultErr) GetStatusCode() int {
	return e.statusCode
}

// NewError creates a new Error instance with the specified error code, message and HTTP status code.
// Parameters:
//   - code: application-specific error code
//   - msg: error message
//   - statusCode: HTTP status code to return
func NewError(code int, msg string, statusCode int) Error {
	return &defaultErr{code: code, msg: msg, statusCode: statusCode}
}

// NewParameterError creates a parameter validation error with HTTP 400 Bad Request status.
// Parameters:
//   - msg: error message
//   - ext: optional error code (defaults to http.StatusBadRequest if not provided)
func NewParameterError(msg string, ext ...int) Error {
	var code = http.StatusBadRequest
	if len(ext) > 0 {
		code = ext[0]
	}
	return NewError(code, msg, http.StatusBadRequest)
}

// NewBusinessError creates a business error with a message, optional error code and data.
// Parameters:
//   - msg: error message
//   - ext: optional parameters:
//   - ext[0]: error code (int)
//   - ext[1]: additional error data (any type)
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
	return &BError{err: NewError(code, msg, http.StatusOK), data: data}
}

// ToError converts any input to an Error type.
// If input is nil, returns nil.
// If input is already an Error type, returns it directly.
// For other types (error, string, or any other), wraps them in a new InnerError with stack trace.
func ToError(input any) Error {
	if input == nil {
		return nil
	}
	switch input := input.(type) {
	case Error:
		return input
	case error:
		return NewInnerErrorSkip(input.Error(), http.StatusInternalServerError, 2)
	case string:
		return NewInnerErrorSkip(input, http.StatusInternalServerError, 2)
	default:
		return NewInnerErrorSkip(fmt.Sprintf("%v", input), http.StatusInternalServerError, 2)
	}
}

// ToErrorWithMsg converts any input to an Error type with an additional message prefix.
// If input is nil, returns nil.
// If msg is not empty, prepends it to the error message in format "msg: original_error_msg".
// Uses ToError internally to handle the input conversion.
func ToErrorWithMsg(input any, msg string) Error {
	if input == nil {
		return nil
	}
	err := ToError(input)
	if msg != "" {
		err.SetMsg(fmt.Sprintf("%s: %s", msg, err.Msg()))
	}
	return err
}

type iError struct {
	*defaultErr
	trace []byte
}

func (e *iError) Error() string {
	msg := e.defaultErr.Error()
	return fmt.Sprintf("%s\n\n%s", msg, e.trace)
}

func (e *iError) Stack() []byte {
	return e.trace
}

// NewInnerError creates a new InnerError with message and code, skipping one stack frame.
// Parameters:
//   - msg: error message
//   - code: error code
//
// Returns Error interface implementation with stack trace
func NewInnerError(msg string, code int) Error {
	return NewInnerErrorSkip(msg, code, 1)
}

// NewInnerErrorWithParams creates a new InnerError with formatted message and stack trace.
// Parameters:
//   - code: error code
//   - format: format string for error message
//   - params: parameters to format the message string
func NewInnerErrorWithParams(code int, format string, params ...any) Error {
	return NewInnerErrorSkip(fmt.Sprintf(format, params...), code, 2)
}

// PanicTrace captures and formats a stack trace for error reporting.
// Parameters:
//   - kb: size of stack buffer in KB (actual size will be kb * 1024 bytes)
//   - skip: number of stack frames to skip from the top
//
// Returns formatted stack trace as bytes, trimmed to relevant section starting from caller
// and excluding goroutine headers.
func PanicTrace(kb int, skip int) []byte {
	stack := make([]byte, kb<<10) //4KB
	length := runtime.Stack(stack, true)

	_, filename, fileLine, ok := runtime.Caller(skip)
	start := 0
	if ok {
		start = bytes.Index(stack, []byte(fmt.Sprintf("%s:%d", filename, fileLine)))
		stack = stack[start:length]
	}

	line := []byte("\n")
	start = bytes.Index(stack, line) + 1
	stack = stack[start:]
	end := bytes.LastIndex(stack, line)
	if end != -1 {
		stack = stack[:end]
	}

	e := []byte("\ngoroutine ")
	end = bytes.Index(stack, e)
	if end != -1 {
		stack = stack[:end]
	}
	stack = bytes.TrimRight(stack, "\n")
	return stack
}

// NewInnerErrorSkip creates a new InnerError with stack trace, skipping the specified number of stack frames.
// Parameters:
//   - msg: error message
//   - code: error code
//   - skip: number of stack frames to skip when capturing stack trace
func NewInnerErrorSkip(msg string, code int, skip int) Error {
	return &iError{
		defaultErr: &defaultErr{code: code, msg: msg, statusCode: http.StatusInternalServerError},
		trace:      PanicTrace(2, skip),
	}
}

// Error Code：1001~1999 used for gone framework.
const (
	GonerNameNotFound   = 1001
	GonerTypeNotFound   = 1002
	CircularDependency  = 1003
	GonerTypeNotMatch   = 1004
	ConfigError         = 1005
	NotSupport          = 1006
	LoadedError         = 1007
	FailInstall         = 1008
	InjectError         = 1009
	ProviderError       = 1010
	StartError          = 1011
	DbRollForPanicError = 1012
	PanicError          = 1013
)
