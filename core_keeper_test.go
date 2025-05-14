package gone

import (
	"errors"
	"reflect"
	"testing"
)

func Test_keeper_getByTypeAndPattern(t *testing.T) {
	k := newKeeper()
	type g struct {
		Flag
	}
	type g1 struct {
		Flag
	}

	_ = k.load(&g{}, Name("food-01"), OnlyForName())
	_ = k.load(&g{}, Name("food-02"))
	_ = k.load(&g{}, Name("food-03"))
	_ = k.load(&g1{}, Name("food-04"))

	type args struct {
		t       reflect.Type
		pattern string
	}
	tests := []struct {
		name        string
		args        args
		wantCoffins []*coffin
	}{
		{
			name: "getByTypeAndPattern",
			args: args{
				t:       reflect.TypeOf(&g{}),
				pattern: "food-0*",
			},
			wantCoffins: []*coffin{
				k.getByName("food-02"),
				k.getByName("food-03"),
			},
		},
		{
			name: "getByTypeAndPattern",
			args: args{
				t:       reflect.TypeOf(&g1{}),
				pattern: "food-0*",
			},
			wantCoffins: []*coffin{
				k.getByName("food-04"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := k
			if gotCoffins := s.getByTypeAndPattern(tt.args.t, tt.args.pattern); !reflect.DeepEqual(gotCoffins, tt.wantCoffins) {
				t.Errorf("getByTypeAndPattern() = %v, want %v", gotCoffins, tt.wantCoffins)
			}
		})
	}
}

func withErrOption() option {
	return option{
		apply: func(c *coffin) error {
			return errors.New("test error")
		},
	}
}

func Test_keeper_load(t *testing.T) {
	k := newKeeper()
	type g struct {
		Flag
		Name string
	}

	_ = k.load(&g{Name: "instance-01"}, Name("food-01"), IsDefault())

	type args struct {
		goner   Goner
		options []Option
	}
	tests := []struct {
		setUp   func() func()
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "gone is nil",
			args: args{
				goner:   nil,
				options: []Option{},
			},
			wantErr: true,
		},
		{
			name: "load with error",
			args: args{
				goner:   &g{},
				options: []Option{withErrOption()},
			},
			wantErr: true,
		},
		{
			name: "load with forceReplace",
			args: args{
				goner:   &g{Name: "replace-01"},
				options: []Option{ForceReplace(), Name("food-01")},
			},
			wantErr: false,
			setUp: func() func() {
				return func() {
					coffins := k.getByTypeAndPattern(reflect.TypeOf(&g{}), "*-01")
					if len(coffins) != 1 {
						t.Errorf("coffins len should be 1")
						return
					}
					if gg, ok := coffins[0].goner.(*g); !ok && gg.Name != "replace-01" {
						t.Errorf("coffins[0].goner should be *g and name should be replace-01")
					}
				}
			},
		},
		{
			name: "duplicated set to default",
			args: args{
				goner:   &g{Name: "duplicated-01"},
				options: []Option{IsDefault()},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setUp != nil {
				defer tt.setUp()()
			}
			s := k
			if err := s.load(tt.args.goner, tt.args.options...); (err != nil) != tt.wantErr {
				t.Errorf("load() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
