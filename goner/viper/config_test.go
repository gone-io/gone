package gone_viper

import (
	"fmt"
	"github.com/gone-io/gone"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func Test_configure_Get(t *testing.T) {
	gone.Prepare(Priest).Test(func(conf gone.Configure) {
		var str string

		err := conf.Get("test", &str, "my-test")
		assert.Nil(t, err)

		assert.Equal(t, "my-test", str)

		var integer int
		err = conf.Get("setting.value", &integer, "100")
		assert.Nil(t, err)
		assert.Equal(t, 2000, integer)

		err = conf.Get("setting.value", integer, "100")
		assert.Error(t, err)
	})
}

func Test_getConf(t *testing.T) {
	vConf := viper.New()
	vConf.SetConfigFile("testdata/config/default.yaml")
	err := vConf.ReadInConfig()
	assert.Nil(t, err)

	var s string
	var i int
	var b bool
	var i64 int64
	var d time.Duration
	var u32 uint32
	var f32 float32

	type Person struct {
		Name         string
		Age          int
		Sex          string
		Introduction string `mapstructure:"self-introduction"`
	}

	var p Person
	var l []*Person

	type args struct {
		key   string
		v     reflect.Value
		vConf *viper.Viper
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "string",
			args: args{
				key:   "setting.name",
				v:     reflect.ValueOf(&s).Elem(),
				vConf: vConf,
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return assert.Equal(t, "default", s)
			},
		},
		{
			name: "int",
			args: args{
				key:   "setting.value",
				v:     reflect.ValueOf(&i).Elem(),
				vConf: vConf,
			},
			wantErr: func(t assert.TestingT, err error, x ...interface{}) bool {
				assert.Nil(t, err)
				return assert.Equal(t, 100, i)
			},
		},
		{
			name: "bool",
			args: args{
				key:   "setting.isOK",
				v:     reflect.ValueOf(&b).Elem(),
				vConf: vConf,
			},
			wantErr: func(t assert.TestingT, err error, x ...interface{}) bool {
				assert.Nil(t, err)
				return assert.Equal(t, true, b)
			},
		},
		{
			name: "int64",
			args: args{
				key:   "setting.value",
				v:     reflect.ValueOf(&i64).Elem(),
				vConf: vConf,
			},
			wantErr: func(t assert.TestingT, err error, x ...interface{}) bool {
				assert.Nil(t, err)
				return assert.Equal(t, int64(100), i64)
			},
		},
		{
			name: "duration",
			args: args{
				key:   "setting.time",
				v:     reflect.ValueOf(&d).Elem(),
				vConf: vConf,
			},
			wantErr: func(t assert.TestingT, err error, x ...interface{}) bool {
				assert.Nil(t, err)
				return assert.Equal(t, 10*time.Hour, d)
			},
		},
		{
			name: "uint32",
			args: args{
				key:   "setting.value",
				v:     reflect.ValueOf(&u32).Elem(),
				vConf: vConf,
			},
			wantErr: func(t assert.TestingT, err error, x ...interface{}) bool {
				assert.Nil(t, err)
				return assert.Equal(t, uint32(100), u32)
			},
		},
		{
			name: "float32",
			args: args{
				key:   "setting.value",
				v:     reflect.ValueOf(&f32).Elem(),
				vConf: vConf,
			},
			wantErr: func(t assert.TestingT, err error, x ...interface{}) bool {
				assert.Nil(t, err)
				return assert.Equal(t, float32(100), f32)
			},
		},
		{
			name: "struct",
			args: args{
				key:   "setting.person",
				v:     reflect.ValueOf(&p).Elem(),
				vConf: vConf,
			},
			wantErr: func(t assert.TestingT, err error, x ...interface{}) bool {
				assert.Nil(t, err)
				return assert.Equal(t, "dapeng", p.Name) &&
					assert.Equal(t, 18, p.Age) &&
					assert.Equal(t, "male", p.Sex) &&
					assert.Equal(t, "i am dapeng", p.Introduction)
			},
		},
		{
			name: "slice",
			args: args{
				key:   "setting.list",
				v:     reflect.ValueOf(&l).Elem(),
				vConf: vConf,
			},
			wantErr: func(t assert.TestingT, err error, x ...interface{}) bool {
				return assert.Nil(t, err) &&
					assert.Equal(t, 2, len(l)) &&
					assert.Equal(t, "dapeng", l[0].Name) &&
					assert.Equal(t, 18, l[0].Age) &&
					assert.Equal(t, "male", l[0].Sex) &&
					assert.Equal(t, "degfy", l[1].Name) &&
					assert.Equal(t, 20, l[1].Age) &&
					assert.Equal(t, "male", l[1].Sex)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, getConf(tt.args.key, tt.args.v, tt.args.vConf), fmt.Sprintf("getConf(%v, %v, %v)", tt.args.key, tt.args.v, tt.args.vConf))
		})
	}
}

func Test_getDefault(t *testing.T) {
	type args struct {
		v           reflect.Value
		defaultVale string
	}
	var s string
	var i int
	var b bool
	var i64 int64
	var d time.Duration
	var u32 uint32
	var f32 float32

	type Person struct {
		Name         string
		Age          int
		Sex          string
		Introduction string `mapstructure:"self-introduction"`
	}

	var p Person
	var l []*Person

	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "string",
			args: args{
				v:           reflect.ValueOf(&s).Elem(),
				defaultVale: "default",
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return assert.Equal(t, "default", s)
			},
		},
		{
			name: "int",
			args: args{
				v:           reflect.ValueOf(&i).Elem(),
				defaultVale: "100",
			},
			wantErr: func(t assert.TestingT, err error, x ...interface{}) bool {
				assert.Nil(t, err)
				return assert.Equal(t, 100, i)
			},
		},
		{
			name: "bool",
			args: args{
				v:           reflect.ValueOf(&b).Elem(),
				defaultVale: "true",
			},
			wantErr: func(t assert.TestingT, err error, x ...interface{}) bool {
				assert.Nil(t, err)
				return assert.Equal(t, true, b)
			},
		},
		{
			name: "int64",
			args: args{
				v:           reflect.ValueOf(&i64).Elem(),
				defaultVale: "100",
			},
			wantErr: func(t assert.TestingT, err error, x ...interface{}) bool {
				assert.Nil(t, err)
				return assert.Equal(t, int64(100), i64)
			},
		},
		{
			name: "duration",
			args: args{
				v:           reflect.ValueOf(&d).Elem(),
				defaultVale: "10h",
			},
			wantErr: func(t assert.TestingT, err error, x ...interface{}) bool {
				assert.Nil(t, err)
				return assert.Equal(t, 10*time.Hour, d)
			},
		},
		{
			name: "uint32",
			args: args{
				v:           reflect.ValueOf(&u32).Elem(),
				defaultVale: "100",
			},
			wantErr: func(t assert.TestingT, err error, x ...interface{}) bool {
				assert.Nil(t, err)
				return assert.Equal(t, uint32(100), u32)
			},
		},
		{
			name: "float32",
			args: args{
				v:           reflect.ValueOf(&f32).Elem(),
				defaultVale: "100",
			},
			wantErr: func(t assert.TestingT, err error, x ...interface{}) bool {
				assert.Nil(t, err)
				return assert.Equal(t, float32(100), f32)
			},
		},
		{
			name: "struct",
			args: args{
				v:           reflect.ValueOf(&p).Elem(),
				defaultVale: "{\n    \"name\":\"dapeng\",\n    \"age\": 18,\n    \"sex\":\"male\",\n    \"self-introduction\":\"i am dapeng\"\n}",
			},
			wantErr: func(t assert.TestingT, err error, x ...interface{}) bool {
				assert.Nil(t, err)
				return assert.Equal(t, "dapeng", p.Name) &&
					assert.Equal(t, 18, p.Age) &&
					assert.Equal(t, "male", p.Sex) &&
					assert.Equal(t, "i am dapeng", p.Introduction)
			},
		},
		{
			name: "slice",
			args: args{
				v:           reflect.ValueOf(&l).Elem(),
				defaultVale: "[{\n    \"name\":\"dapeng\",\n    \"age\": 18,\n    \"sex\":\"male\"\n},{\n    \"name\":\"degfy\",\n    \"age\": 20,\n    \"sex\":\"male\"\n}]",
			},
			wantErr: func(t assert.TestingT, err error, x ...interface{}) bool {
				return assert.Nil(t, err) &&
					assert.Equal(t, 2, len(l)) &&
					assert.Equal(t, "dapeng", l[0].Name) &&
					assert.Equal(t, 18, l[0].Age) &&
					assert.Equal(t, "male", l[0].Sex) &&
					assert.Equal(t, "degfy", l[1].Name) &&
					assert.Equal(t, 20, l[1].Age) &&
					assert.Equal(t, "male", l[1].Sex)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, getDefault(tt.args.v, tt.args.defaultVale), fmt.Sprintf("getDefault(%v, %v)", tt.args.v, tt.args.defaultVale))
		})
	}
}
