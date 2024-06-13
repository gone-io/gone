package properties

import (
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/magiconair/properties"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_parseKeyFromProperties(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	props := NewMockPropertiesConfigure(controller)
	props.EXPECT().FilterStripPrefix(gomock.Any()).Return(properties.LoadMap(map[string]string{
		"test": "test",
	})).AnyTimes()
	props.EXPECT().Decode(gomock.Any()).Return(errors.New("err")).AnyTimes()

	var m = map[string]any{}
	var s1 []string
	var s2 []*string

	type X struct {
		Test  string
		Test2 []string
	}

	var x = X{}
	var x1 = make([]X, 0)
	var x2 = make([]*X, 0)
	var dd time.Duration

	type args struct {
		key         string
		value       any
		defaultVale string
		props       Configure
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "type of value must be ptr",
			args: args{
				key:         "test",
				value:       "test",
				defaultVale: "",
				props:       props,
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err, i...)
			},
		}, {
			name: "not support",
			args: args{
				key:         "test",
				value:       &m,
				defaultVale: "",
				props:       props,
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err, i...)
			},
		}, {
			name: "not support",
			args: args{
				key:         "test",
				value:       &s1,
				defaultVale: "",
				props:       props,
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err, i...)
			},
		}, {
			name: "not support",
			args: args{
				key:         "test",
				value:       &s2,
				defaultVale: "",
				props:       props,
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err, i...)
			},
		}, {
			name: "Decode error",
			args: args{
				key:         "test",
				value:       &x,
				defaultVale: "",
				props:       props,
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err, i...)
			},
		}, {
			name: "Decode error2",
			args: args{
				key:         "test",
				value:       &x1,
				defaultVale: "",
				props:       props,
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err, i...)
			},
		}, {
			name: "Decode error2",
			args: args{
				key:         "test",
				value:       &x2,
				defaultVale: "",
				props:       props,
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err, i...)
			},
		}, {
			name: "Decode error2",
			args: args{
				key:         "test",
				value:       &dd,
				defaultVale: "xxx",
				props:       props,
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err, i...)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, parseKeyFromProperties(tt.args.key, tt.args.value, tt.args.defaultVale, tt.args.props), fmt.Sprintf("parseKeyFromProperties(%v, %v, %v, %v)", tt.args.key, tt.args.value, tt.args.defaultVale, tt.args.props))
		})
	}
}
