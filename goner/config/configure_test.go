package config

import (
	"fmt"
	"github.com/magiconair/properties"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func Test_getFromProperties(t *testing.T) {
	var b bool
	var f func()
	var duration time.Duration

	type Person struct {
		Name string `properties:"name"`
		Age  int    `properties:"age"`
	}

	var list []Person

	type args struct {
		key         string
		value       any
		defaultVale string
		props       *properties.Properties
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "passed value for `value`",
			args: args{
				key:         "test",
				value:       "test",
				defaultVale: "test",
				props:       properties.MustLoadString(`test=test`),
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err, i...)
			},
		},
		{
			name: "value is bool",
			args: args{
				key:         "test",
				value:       &b,
				defaultVale: "test",
				props:       properties.MustLoadString(`test=true`),
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.NoError(t, err, i...) &&
					assert.Equal(t, true, b)
			},
		},
		{
			name: "value is unsupported type",
			args: args{
				key:         "test",
				value:       &f,
				defaultVale: "test",
				props:       properties.MustLoadString(`test=test`),
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err, i...)
			},
		},
		{
			name: "value is duration",
			args: args{
				key:         "test",
				value:       &duration,
				defaultVale: "10s",
				props:       properties.MustLoadString(`test=`),
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.NoError(t, err, i...) &&
					assert.Equal(t, 10*time.Second, duration)
			},
		},
		{
			name: "value is slice, and decode error",
			args: args{
				key:         "test",
				value:       &list,
				defaultVale: "test",
				props:       properties.MustLoadString(`test[0].age=10i`),
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err, i...)
			},
		},
		{
			name: "value is slice, and decode suc",
			args: args{
				key:         "test",
				value:       &list,
				defaultVale: "test",
				props:       properties.MustLoadString("test[0].age=10\ntest[0].name=dapeng\ntest[1].name=degfy\ntest[1].age=20"),
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.NoError(t, err, i...) &&
					assert.Equal(t, 2, len(list)) &&
					assert.Equal(t, 10, list[0].Age) &&
					assert.Equal(t, "dapeng", list[0].Name) &&
					assert.Equal(t, "degfy", list[1].Name) &&
					assert.Equal(t, 20, list[1].Age)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, getFromProperties(tt.args.key, tt.args.value, tt.args.defaultVale, tt.args.props), fmt.Sprintf("getFromProperties(%v, %v, %v, %v)", tt.args.key, tt.args.value, tt.args.defaultVale, tt.args.props))
		})
	}
}

func Test_decodeSliceElement(t *testing.T) {
	type Person struct {
		Name string `properties:"name"`
		Age  int    `properties:"age"`
	}

	var list = make([]Person, 0)
	var listP = make([]*Person, 0)

	var p = &listP

	type args struct {
		sliceElementType reflect.Type
		k                string
		conf             *properties.Properties
		el               reflect.Value
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "struct slice suc",
			args: args{
				sliceElementType: reflect.TypeOf(Person{}),
				k:                "test",
				conf:             properties.MustLoadString("name=test\nage=20"),
				el:               reflect.ValueOf(&list).Elem(),
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.NoError(t, err, i...) &&
					assert.Equal(t, 1, len(list)) &&
					assert.Equal(t, "test", list[0].Name) &&
					assert.Equal(t, 20, list[0].Age)
			},
		},
		{
			name: "struct slice error",
			args: args{
				sliceElementType: reflect.TypeOf(Person{}),
				k:                "test",
				conf:             properties.MustLoadString("name=test\nage=20i"),
				el:               reflect.ValueOf(&list).Elem(),
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err, i...)
			},
		},
		{
			name: "struct pointer slice suc",
			args: args{
				sliceElementType: reflect.TypeOf(&Person{}),
				k:                "test",
				conf:             properties.MustLoadString("name=test\nage=20"),
				el:               reflect.ValueOf(&listP).Elem(),
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.NoError(t, err, i...) &&
					assert.Equal(t, 1, len(listP)) &&
					assert.Equal(t, "test", listP[0].Name) &&
					assert.Equal(t, 20, listP[0].Age)
			},
		},
		{
			name: "struct pointer slice err",
			args: args{
				sliceElementType: reflect.TypeOf(&Person{}),
				k:                "test",
				conf:             properties.MustLoadString("name=test\nage=20i"),
				el:               reflect.ValueOf(&listP).Elem(),
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err, i...)
			},
		},
		{
			name: "other pointer slice err",
			args: args{
				sliceElementType: reflect.TypeOf(p),
				k:                "test",
				conf:             properties.MustLoadString("name=test\nage=20i"),
				el:               reflect.ValueOf(&listP).Elem(),
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err, i...)
			},
		},
		{
			name: "other type",
			args: args{
				sliceElementType: reflect.TypeOf(func() {}),
				k:                "test",
				conf:             properties.MustLoadString("name=test\nage=20i"),
				el:               reflect.ValueOf(&listP).Elem(),
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err, i...)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, decodeSliceElement(tt.args.sliceElementType, tt.args.k, tt.args.conf, tt.args.el), fmt.Sprintf("decodeSliceElement(%v, %v, %v, %v)", tt.args.sliceElementType, tt.args.k, tt.args.conf, tt.args.el))
		})
	}
}
