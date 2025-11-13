package gone

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestNewError(t *testing.T) {
	tests := []struct {
		name       string
		code       int
		msg        string
		statusCode int
		wantErr    string
	}{
		{
			name:       "basic error",
			code:       1001,
			msg:        "test error",
			statusCode: http.StatusBadRequest,
			wantErr:    "GoneError(code=1001); test error",
		},
		{
			name:       "empty message",
			code:       1002,
			msg:        "",
			statusCode: http.StatusInternalServerError,
			wantErr:    "GoneError(code=1002); ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewError(tt.code, tt.msg, tt.statusCode)
			if err.Error() != tt.wantErr {
				t.Errorf("NewError() error = %v, wantErr %v", err.Error(), tt.wantErr)
			}
			if err.Code() != tt.code {
				t.Errorf("NewError() code = %v, want %v", err.Code(), tt.code)
			}
			if err.GetStatusCode() != tt.statusCode {
				t.Errorf("NewError() statusCode = %v, want %v", err.GetStatusCode(), tt.statusCode)
			}
		})
	}
}

func TestNewParameterError(t *testing.T) {
	tests := []struct {
		name    string
		msg     string
		ext     []int
		wantErr string
	}{
		{
			name:    "basic parameter error",
			msg:     "invalid parameter",
			ext:     nil,
			wantErr: "GoneError(code=400); invalid parameter",
		},
		{
			name:    "parameter error with custom code",
			msg:     "custom error",
			ext:     []int{1001},
			wantErr: "GoneError(code=1001); custom error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewParameterError(tt.msg, tt.ext...)
			if err.Error() != tt.wantErr {
				t.Errorf("NewParameterError() error = %v, wantErr %v", err.Error(), tt.wantErr)
			}
			if err.GetStatusCode() != http.StatusBadRequest {
				t.Errorf("NewParameterError() statusCode = %v, want %v", err.GetStatusCode(), http.StatusBadRequest)
			}
		})
	}
}

func TestNewBusinessError(t *testing.T) {
	tests := []struct {
		name    string
		msg     string
		ext     []any
		wantErr string
	}{
		{
			name:    "basic business error",
			msg:     "business error",
			ext:     nil,
			wantErr: "GoneError(code=0); business error",
		},
		{
			name:    "business error with code",
			msg:     "business error with code",
			ext:     []any{1001},
			wantErr: "GoneError(code=1001); business error with code",
		},
		{
			name:    "business error with code and data",
			msg:     "business error with data",
			ext:     []any{1002, "data"},
			wantErr: "GoneError(code=1002); business error with data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewBusinessError(tt.msg, tt.ext...)
			if err.Error() != tt.wantErr {
				t.Errorf("NewBusinessError() error = %v, wantErr %v", err.Error(), tt.wantErr)
			}
			if err.GetStatusCode() != http.StatusOK {
				t.Errorf("NewBusinessError() statusCode = %v, want %v", err.GetStatusCode(), http.StatusOK)
			}

			// Test Data() if provided
			if len(tt.ext) > 1 {
				if err.Data() != tt.ext[1] {
					t.Errorf("NewBusinessError() data = %v, want %v", err.Data(), tt.ext[1])
				}
			}
		})
	}
}

func TestToError(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		wantNil  bool
		wantType string
	}{
		{
			name:    "nil input",
			input:   nil,
			wantNil: true,
		},
		{
			name:     "error interface",
			input:    fmt.Errorf("test error"),
			wantType: "*gone.iError",
		},
		{
			name:     "string input",
			input:    "test error",
			wantType: "*gone.iError",
		},
		{
			name:     "any input",
			input:    123,
			wantType: "*gone.iError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ToError(tt.input)
			if tt.wantNil {
				if err != nil {
					t.Errorf("ToError() = %v, want nil", err)
				}
				return
			}

			if err == nil {
				t.Error("ToError() = nil, want error")
				return
			}

			// Check if it's an InnerError and contains stack trace
			if innerErr, ok := err.(InnerError); ok {
				if len(innerErr.Stack()) == 0 {
					t.Error("ToError() stack trace is empty")
				}
			}
		})
	}
}

func TestToErrorWithMsg(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		msg     string
		wantNil bool
		wantMsg string
	}{
		{
			name:    "nil input",
			input:   nil,
			msg:     "prefix",
			wantNil: true,
		},
		{
			name:    "with prefix",
			input:   "test error",
			msg:     "prefix",
			wantMsg: "prefix",
		},
		{
			name:    "without prefix",
			input:   "test error",
			msg:     "",
			wantMsg: "test error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ToErrorWithMsg(tt.input, tt.msg)
			if tt.wantNil {
				if err != nil {
					t.Errorf("ToErrorWithMsg() = %v, want nil", err)
				}
				return
			}

			if err == nil {
				t.Error("ToErrorWithMsg() = nil, want error")
				return
			}

			if err.Msg() != tt.wantMsg {
				t.Errorf("ToErrorWithMsg() message = %v, want %v", err.Msg(), tt.wantMsg)
			}
		})
	}
}

func TestNewInnerError(t *testing.T) {
	err := NewInnerError("test error", http.StatusInternalServerError)
	if err == nil {
		t.Fatal("NewInnerError() = nil, want error")
	}

	innerErr, ok := err.(InnerError)
	if !ok {
		t.Fatal("NewInnerError() did not return InnerError interface")
	}

	if !strings.Contains(string(innerErr.Stack()), "error_test.go") {
		t.Error("NewInnerError() stack trace does not contain test file")
	}
}

func TestPanicTrace(t *testing.T) {
	trace := PanicTrace(4, 1)
	if len(trace) == 0 {
		t.Error("PanicTrace() returned empty trace")
	}

	if strings.Contains(string(trace), "error_test.go") {
		t.Error("PanicTrace() does not contain test file")
	}

	trace = PanicTrace(4, 0)
	if !strings.Contains(string(trace), "error_test.go") {
		t.Error("PanicTrace() does not contain test file")
	}
}

func Test_iError_Error(t *testing.T) {
	type fields struct {
		defaultErr *defaultErr
		trace      []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test error",
			fields: fields{
				defaultErr: &defaultErr{
					code: 100,
				},
				trace: []byte("test trace"),
			},
			want: "GoneError(code=100); \n\ntest trace",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &iError{
				defaultErr: tt.fields.defaultErr,
				trace:      tt.fields.trace,
			}
			if got := e.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBError_SetMsg(t *testing.T) {
	type fields struct {
		err  Error
		data any
	}
	type args struct {
		msg string
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		want     string
		wantCode int
	}{
		{
			name: "test error",
			fields: fields{
				err: &iError{
					defaultErr: &defaultErr{
						code: 100,
					},
				},
			},
			args:     args{msg: "test error"},
			want:     "test error",
			wantCode: 100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &BError{
				err:  tt.fields.err,
				data: tt.fields.data,
			}
			e.SetMsg(tt.args.msg)
			if got := e.Msg(); got != tt.want {
				t.Errorf("SetMsg(%v) = %v, want %v", tt.args.msg, got, tt.want)
			}
			if got := e.Code(); got != tt.wantCode {
				t.Errorf("Code() after SetMsg(%v) = %v, want %v", tt.args.msg, got, tt.wantCode)
			}
		})
	}
}

func TestToErrorf(t *testing.T) {
	t.Run("input nil", func(t *testing.T) {
		err := ToErrorf(nil, "test error")
		if err != nil {
			t.Error("must be nil")
		}
	})

	t.Run("input error", func(t *testing.T) {
		var err = errors.New("test error")
		newErr := ToErrorf(err, "my-test")

		if !errors.Is(newErr, err) {
			t.Error("must be input error")
		}

		var gErr Error
		if !errors.As(newErr, &gErr) {
			t.Error("must be GoneError")
		}

		if gErr.GetStatusCode() != http.StatusInternalServerError {
			t.Error("must be 500")
		}

		if gErr.Msg() != "my-test" {
			t.Error("must be my-test")
		}

		gErr.SetMsg("my-test2")
		if gErr.Msg() != "my-test2" {
			t.Error("must be my-test2")
		}

	})
}

func TestWrapError(t *testing.T) {
	t.Run("wrap err", func(t *testing.T) {
		tests := []struct {
			name string
			err  error
		}{
			{
				name: "wrap normal error",
				err:  fmt.Errorf("normal error"),
			},
			{
				name: "wrap gone InnerError",
				err:  NewInnerError("test error", http.StatusInternalServerError),
			},
			{
				name: "wrap gone BusinessError",
				err:  NewBusinessError("test error", http.StatusOK),
			},
			{
				name: "wrap gone ParameterError",
				err:  NewParameterError("test error", http.StatusBadRequest),
			},
			{
				name: "wrap gone Error",
				err:  NewError(1101, "test error", http.StatusBadRequest),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				wrapError := WrapError(tt.err, "wrap error")
				if !errors.Is(wrapError, tt.err) {
					t.Error("must be input error")
				}
			})
		}
	})
	t.Run("wrap none error", func(t *testing.T) {
		wrapError := WrapError("some thing", "wrap error")

		msg := wrapError.Msg()
		if msg != "wrap error" {
			t.Error("must be wrap error")
		}
	})
}
