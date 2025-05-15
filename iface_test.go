package gone

import (
	"reflect"
	"testing"
)

func TestFlag_goneFlag(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Flag{}
			g.goneFlag()
		})
	}
}

func Test_actionType_String(t *testing.T) {
	tests := []struct {
		name string
		t    actionType
		want string
	}{
		{
			name: "fillAction",
			t:    fillAction,
			want: "fill fields",
		},
		{
			name: "initAction",
			t:    initAction,
			want: "initialize",
		},
		{
			name: "unknown",
			t:    actionType(10),
			want: "unknown",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.t.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dependency_String(t *testing.T) {
	type fields struct {
		coffin *coffin
		action actionType
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "with name",
			fields: fields{
				coffin: &coffin{
					name: "test",
				},
				action: fillAction,
			},
			want: "<fill fields of \"test\">",
		},
		{
			name: "without name",
			fields: fields{
				coffin: &coffin{
					goner: &struct{}{},
				},
				action: fillAction,
			},
			want: "<fill fields of \"*struct{}\">",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := dependency{
				coffin: tt.fields.coffin,
				action: tt.fields.action,
			}
			if got := d.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_filedHasOption(t *testing.T) {
	type args struct {
		filed      *reflect.StructField
		tagName    string
		optionName string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "has option",
			args: args{
				filed:      &reflect.StructField{Tag: reflect.StructTag(`gone:"name,option"`)},
				tagName:    "gone",
				optionName: "option",
			},
			want: true,
		},
		{
			name: "has not option",
			args: args{
				filed:      &reflect.StructField{Tag: reflect.StructTag(`gone:"name"`)},
				tagName:    "gone",
				optionName: "option",
			},
			want: false,
		},
		{
			name: "has option with empty tag",
			args: args{
				filed:      &reflect.StructField{Tag: reflect.StructTag(`gone:""`)},
				tagName:    "gone",
				optionName: "option",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filedHasOption(tt.args.filed, tt.args.tagName, tt.args.optionName); got != tt.want {
				t.Errorf("filedHasOption() = %v, want %v", got, tt.want)
			}
		})
	}
}
