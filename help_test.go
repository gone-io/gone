package gone

import (
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func Test_FuncName(t *testing.T) {
	name := GetFuncName(Test_FuncName)
	println(name)
	assert.Equal(t, name, "github.com/gone-io/gone.Test_FuncName")
	fn := func() {}

	println(GetFuncName(fn))

	assert.Equal(t, GetFuncName(fn), "github.com/gone-io/gone.Test_FuncName.func1")
}

type XX interface {
	Get()
}

var XXPtr *XX
var XXType = reflect.TypeOf(XXPtr).Elem()

func Test_getInterfaceType(t *testing.T) {
	interfaceType := GetInterfaceType(new(XX))
	assert.Equal(t, interfaceType, XXType)
}

func forText(in struct {
	a Point `gone:"point-a"`
	b Point `gone:"point-b"`
}) int {
	println(in.a.GetIndex(), in.b.GetIndex())
	return in.a.GetIndex() + in.b.GetIndex()
}

func TestInjectWrapFn(t *testing.T) {
	heaven :=
		New(func(cemetery Cemetery) error {
			cemetery.
				Bury(&Point{Index: 1}, GonerId("point-a")).
				Bury(&Point{Index: 2}, GonerId("point-b")).
				Bury(&Point{Index: 3}, GonerId("point-c"))

			return nil
		})

	flag := 0
	heaven.AfterStart(func(cemetery Cemetery) error {
		fn, err := InjectWrapFn(cemetery, forText)
		assert.Nil(t, err)

		results := ExecuteInjectWrapFn(fn)
		assert.Equal(t, 3, results[0])

		flag = 1

		return nil
	})
	heaven.Install()
	heaven.Start()
	assert.Equal(t, 1, flag)
}

type testData struct {
	a Point `gone:"point-a"`
}

func TestInjectWrapFn1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	theCemetery := NewMockCemetery(ctrl)
	injectFailed := errors.New("inject failed")
	theCemetery.EXPECT().ReviveOne(gomock.Any()).Return(nil, injectFailed)

	type args struct {
		cemetery Cemetery
		fn       any
	}

	tests := []struct {
		name    string
		args    args
		want    *reflect.Value
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "fn is not func",
			args: args{
				cemetery: theCemetery,
				fn:       "not func",
			},
			want: nil,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "fn must be a function")
			},
		},
		{
			name: "parameters count of fn is gt 1",
			args: args{
				cemetery: theCemetery,
				fn:       func(x struct{}, y struct{}) {},
			},
			want: nil,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "fn only support one input parameter or no input parameter")
			},
		},
		{
			name: "fn input parameter must be a struct ",
			args: args{
				cemetery: theCemetery,
				fn:       func(x int) {},
			},
			want: nil,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "fn input parameter must be a struct")
			},
		},
		{
			name: "fn input parameter must be a anonymous struct",
			args: args{
				cemetery: theCemetery,
				fn:       func(x testData) {},
			},
			want: nil,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "fn input parameter must be a anonymous struct")
			},
		},
		{
			name: "inject failed",
			args: args{
				cemetery: theCemetery,
				fn: func(x struct {
					cemetery Cemetery `gone:"gone-cemetery"`
				}) {
				},
			},
			want: nil,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "inject failed")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InjectWrapFn(tt.args.cemetery, tt.args.fn)
			if !tt.wantErr(t, err, fmt.Sprintf("InjectWrapFn(%v, %v)", tt.args.cemetery, tt.args.fn)) {
				return
			}
			assert.Equalf(t, tt.want, got, "InjectWrapFn(%v, %v)", tt.args.cemetery, tt.args.fn)
		})
	}
}
