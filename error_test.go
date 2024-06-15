package gone

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func Test_iError_Error(t *testing.T) {
	innerError := NewInnerError("test", 100)
	s := innerError.Error()

	assert.True(t, strings.Contains(s, "Test_iError_Error"))
}

func TestNewBusinessError(t *testing.T) {
	type args struct {
		msg string
		ext []any
	}

	var data = map[string]any{}
	tests := []struct {
		name string
		args args
		want BError
	}{
		{
			name: "single parameter",
			args: args{
				msg: "error",
				ext: []any{},
			},
			want: BError{
				err: NewError(0, "error"),
			},
		},
		{
			name: "two parameters",
			args: args{
				msg: "error",
				ext: []any{100},
			},
			want: BError{
				err: NewError(100, "error"),
			},
		},
		{
			name: "three parameters",
			args: args{
				msg: "error",
				ext: []any{100, data},
			},
			want: BError{
				err:  NewError(100, "error"),
				data: data,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want,
				*(NewBusinessError(tt.args.msg, tt.args.ext...).(*BError)),
				"NewBusinessError(%v, %v)",
				tt.args.msg,
				tt.args.ext,
			)
		})
	}

	businessError := NewBusinessError("error", 100, data)
	assert.Equal(t, "error", businessError.Msg())
	assert.Equal(t, 100, businessError.Code())
	assert.Equal(t, data, businessError.Data())
	assert.Equal(t, "GoneError(code=100); error", businessError.Error())
}

func TestNewParameterError(t *testing.T) {
	type args struct {
		msg string
		ext []int
	}
	tests := []struct {
		name string
		args args
		want defaultErr
	}{
		{
			name: "single parameter",
			args: args{
				msg: "error",
				ext: []int{},
			},
			want: defaultErr{msg: "error", code: 400},
		},
		{
			name: "single parameter",
			args: args{
				msg: "error",
				ext: []int{401},
			},
			want: defaultErr{msg: "error", code: 401},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t,
				tt.want,
				*NewParameterError(tt.args.msg, tt.args.ext...).(*defaultErr),
				"NewParameterError(%v, %v)",
				tt.args.msg,
				tt.args.ext,
			)
		})
	}
}

func TestToError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want Error
	}{
		{
			name: "err = nil",
			args: args{
				err: nil,
			},
			want: nil,
		},
		{
			name: "err is Error",
			args: args{
				err: NewBusinessError("error", 100),
			},
			want: NewBusinessError("error", 100),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, ToError(tt.args.err), "ToError(%v)", tt.args.err)
		})
	}

	t.Run("err is not Error", func(t *testing.T) {
		err := errors.New("error")
		innerError := ToError(err)
		assert.Equal(t, 500, innerError.Code())
		assert.Equal(t, "error", innerError.Msg())
		assert.NotNil(t, innerError.(InnerError).Stack())
	})
	t.Run("err is string", func(t *testing.T) {
		toError := ToError("error")
		assert.Equal(t, 500, toError.Code())
		assert.Equal(t, "error", toError.Msg())
	})

	t.Run("err is int", func(t *testing.T) {
		toError := ToError(100)
		assert.Equal(t, 500, toError.Code())
		assert.Equal(t, "100", toError.Msg())
	})
}
