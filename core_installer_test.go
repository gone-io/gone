package gone

import (
	"errors"
	"go.uber.org/mock/gomock"
	"reflect"
	"testing"
)

func Test_installer_injectFieldAsSlice(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	analyzer := NewMockiDependenceAnalyzer(controller)
	logger := NewMockLogger(controller)

	ins := newInstaller(analyzer, logger)

	type g struct {
		Flag
	}

	var x struct {
		Flag
		List []any `gone:"*"`
	}
	field, _ := reflect.TypeOf(&x).Elem().FieldByName("List")
	v := reflect.ValueOf(&x).Elem().FieldByName("List")

	type args struct {
		extend     string
		depCoffins []*coffin
		field      reflect.StructField
		v          reflect.Value
		coName     string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "provider provide err",
			args: args{
				extend: "",
				depCoffins: []*coffin{
					newCoffin(&g1Provider{
						err: errors.New("err"),
					}),
				},
				field: reflect.StructField{
					Type: reflect.TypeOf([]*g1{}),
				},
				v:      reflect.ValueOf([]*g1{}),
				coName: "test-goner",
			},
			wantErr: true,
		},
		{
			name: "provider provide value not compatible",
			args: args{
				extend: "",
				depCoffins: []*coffin{
					newCoffin(&g1Provider{
						g1: &g1{},
					}),
				},
				field: reflect.StructField{
					Type: reflect.TypeOf([]*g{}),
				},
				v:      reflect.ValueOf([]*g1{}),
				coName: "test-goner",
			},
			wantErr: true,
		},
		{
			name: "ok",
			args: args{
				extend: "",
				depCoffins: []*coffin{
					newCoffin(&g1Provider{
						g1: &g1{},
					}),
					newCoffin(&g{}),
				},
				field: field,
				v:     v,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := ins
			if err := s.injectFieldAsSlice(tt.args.extend, tt.args.depCoffins, tt.args.field, tt.args.v, tt.args.coName); (err != nil) != tt.wantErr {
				t.Errorf("injectFieldAsSlice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

var _ StructFieldInjector = (*testInjector)(nil)

type testInjector struct {
	Flag
	v   any
	err error
}

func (t testInjector) GonerName() string {
	return "test-injector"
}

func (t testInjector) Inject(tagConf string, field reflect.StructField, fieldValue reflect.Value) (err error) {
	if t.err != nil {
		return t.err
	}
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(e.(string))
		}
	}()
	fieldValue.Set(reflect.ValueOf(t.v))
	return
}

func Test_installer_injectFieldAsNotSlice(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	analyzer := NewMockiDependenceAnalyzer(controller)
	logger := NewMockLogger(controller)

	ins := newInstaller(analyzer, logger)

	var x struct {
		Flag
		List []any `gone:"*"`
	}
	listField, _ := reflect.TypeOf(&x).Elem().FieldByName("List")
	listV := reflect.ValueOf(&x).Elem().FieldByName("List")

	type args struct {
		byName bool
		extend string
		depCo  *coffin
		field  reflect.StructField
		v      reflect.Value
		coName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "provider provide err",
			args: args{
				byName: false,
				extend: "",
				depCo: newCoffin(&g1Provider{
					err: errors.New("err"),
				}),

				field:  listField,
				v:      listV,
				coName: "test-goner",
			},
			wantErr: true,
		},

		{
			name: "provider provide suc",
			args: args{
				byName: false,
				extend: "",
				depCo:  newCoffin([]any{"test", "test2"}),

				field:  listField,
				v:      listV,
				coName: "test-goner",
			},
			wantErr: false,
		},
		{
			name: "injector suc",
			args: args{
				byName: false,
				extend: "",
				depCo: newCoffin(&testInjector{
					v:   []any{"test", "test2"},
					err: nil,
				}),

				field:  listField,
				v:      listV,
				coName: "test-goner",
			},
			wantErr: false,
		},
		{
			name: "injector err",
			args: args{
				byName: false,
				extend: "",
				depCo: newCoffin(&testInjector{
					v:   nil,
					err: errors.New("test err"),
				}),

				field:  listField,
				v:      listV,
				coName: "test-goner",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := ins
			if err := s.injectFieldAsNotSlice(tt.args.byName, tt.args.extend, tt.args.depCo, tt.args.field, tt.args.v, tt.args.coName); (err != nil) != tt.wantErr {
				t.Errorf("injectFieldAsNotSlice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type testErrBeforeInit struct {
	Flag
	g1  *g1 `gone:"*"`
	err error
}

func (t testErrBeforeInit) BeforeInit() error {
	return t.err
}

type testBeforeInitiatorNoError struct {
	Flag
}

func (t testBeforeInitiatorNoError) BeforeInit() {
}

func Test_installer_fillOne(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	analyzer := NewMockiDependenceAnalyzer(controller)
	logger := NewMockLogger(controller)

	ins := newInstaller(analyzer, logger)

	var x struct {
		Flag
		g1 *g1 `gone:"*"`
	}

	type args struct {
		co *coffin
	}
	tests := []struct {
		setUp   func() func()
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "init err",
			args: args{
				co: newCoffin(&testErrBeforeInit{err: errors.New("test")}),
			},
			wantErr: true,
		},
		{
			name: "is not struct pointer",
			args: args{
				co: newCoffin(&[]int{}),
			},
			wantErr: true,
		},
		{
			name: "ok",
			args: args{
				co: newCoffin(&x),
			},
			setUp: func() func() {
				analyzer.EXPECT().analyzerFieldDependencies(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)

				return func() {}
			},
			wantErr: false,
		},
		{
			name: "err",
			args: args{
				co: newCoffin(&x),
			},
			setUp: func() func() {
				analyzer.EXPECT().analyzerFieldDependencies(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("err"))

				return func() {}
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setUp != nil {
				defer tt.setUp()()
			}
			s := ins
			if err := s.fillOne(tt.args.co); (err != nil) != tt.wantErr {
				t.Errorf("fillOne() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_installer_doBeforeInit(t *testing.T) {

	type args struct {
		goner any
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "goner is BeforeInitiatorNoError",
			args: args{
				goner: &testBeforeInitiatorNoError{},
			},
			wantErr: false,
		},
		{
			name: "goner is BeforeInitiator return err",
			args: args{
				goner: &testErrBeforeInit{
					err: errors.New("test"),
				},
			},
			wantErr: true,
		},
		{
			name: "goner is BeforeInitiator return nil",
			args: args{
				goner: &testErrBeforeInit{
					err: nil,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &installer{}
			if err := s.doBeforeInit(tt.args.goner); (err != nil) != tt.wantErr {
				t.Errorf("doBeforeInit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_installer_safeInitOne(t *testing.T) {

	type args struct {
		c *coffin
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "coffin goner is Initiator",
			args: args{
				c: newCoffin(&testInitiator{}),
			},
			wantErr: false,
		},
		{
			name: "coffin goner is Initiator return err",
			args: args{
				c: newCoffin(&testInitiator{
					err: errors.New("test"),
				}),
			},
			wantErr: true,
		},
		{
			name: "coffin goner is InitiatorNoError",
			args: args{
				c: newCoffin(&testInitiatorNoError{}),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &installer{}
			if err := s.safeInitOne(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("safeInitOne() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
