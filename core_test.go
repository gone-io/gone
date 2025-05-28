package gone

import (
	"errors"
	"fmt"
	"go.uber.org/mock/gomock"
	"reflect"
	"testing"
)

func Test_core_GetGonerByName(t *testing.T) {
	type X struct {
		Flag
		ID int
	}

	NewApp().
		Load(&X{ID: 1001}, Name("x")).
		Run(func(k GonerKeeper) {
			name := k.GetGonerByName("x")
			if name == nil {
				t.Error("name is nil")
				return
			}
			if x, ok := name.(*X); !ok {
				t.Error("name is not *X")
				return
			} else if x.ID != 1001 {
				t.Error("name.ID is not 1001")
				return
			}

			y := k.GetGonerByName("y")
			if y != nil {
				t.Error("y is not nil")
				return
			}
		})
}

func Test_core_GetGonerByType(t *testing.T) {
	type X struct {
		Flag
		ID int
	}

	t.Run("suc", func(t *testing.T) {
		NewApp().
			Load(&X{ID: 1001}, Name("x")).
			Load(WrapFunctionProvider(func(tagConf string, param struct{}) (*X, error) {
				return &X{ID: 1002}, nil
			})).
			Run(func(k GonerKeeper) {
				x := k.GetGonerByType(reflect.TypeOf(&X{}))
				if x == nil {
					t.Error("x is nil")
					return
				}
				if x.(*X).ID != 1001 {
					t.Error("x.ID is not 1001")
					return
				}
			})
	})

	t.Run("provide err", func(t *testing.T) {
		NewApp().
			Load(WrapFunctionProvider(func(tagConf string, param struct{}) (*X, error) {
				return nil, errors.New("provide err")
			})).
			Run(func(k GonerKeeper) {
				err := SafeExecute(func() error {
					_ = k.GetGonerByType(reflect.TypeOf(&X{}))
					return nil
				})
				if err == nil {
					t.Errorf("err is nil")
				}
			})
	})

	t.Run("not found", func(t *testing.T) {
		NewApp().
			Run(func(k GonerKeeper) {
				g := k.GetGonerByType(reflect.TypeOf(&X{}))
				if g != nil {
					t.Error("g is not nil")
				}
			})
	})

}

func Test_core_InjectStruct(t *testing.T) {
	type dep struct {
		Flag
	}

	type X struct {
		x *dep `gone:"*"`
	}

	t.Run("suc", func(t *testing.T) {
		d := &dep{}
		NewApp().
			Load(d).
			Run(func(i StructInjector) {
				var x = &X{}
				err := i.InjectStruct(x)
				if err != nil {
					t.Errorf("err is not nil")
					return
				}
				if x.x != d {
					t.Errorf("x.x is not d")
				}
			})
	})

	t.Run("goner is not ptr", func(t *testing.T) {
		NewApp().
			Run(func(i StructInjector) {
				err := i.InjectStruct("test")
				if err == nil {
					t.Errorf("err is nil")
				}
			})
	})
	t.Run("goner is not ptr to struct", func(t *testing.T) {
		var test string

		NewApp().
			Run(func(i StructInjector) {
				err := i.InjectStruct(&test)
				if err == nil {
					t.Errorf("err is nil")
				}
			})
	})
}

func Test_core_InjectFuncParameters(t *testing.T) {

	type X struct {
		ID int
	}

	type args struct {
		fn           any
		injectBefore FuncInjectHook
		injectAfter  FuncInjectHook
	}
	tests := []struct {
		name string

		load    func(loader Loader) error
		args    args
		wantErr bool
	}{
		{
			name: "fn is not a func",
			load: func(loader Loader) error {
				return nil
			},
			args: args{
				fn:           "test",
				injectBefore: nil,
				injectAfter:  nil,
			},
			wantErr: true,
		},
		{
			name: "provide value err",
			load: func(loader Loader) error {
				return nil
			},
			args: args{
				fn: func(in struct {
					x *X `gone:"*"`
				}) {
					t.Error("should not be called")
				},
				injectBefore: nil,
				injectAfter:  nil,
			},
			wantErr: true,
		},
		{
			name: "provide value suc",
			load: func(loader Loader) error {
				return loader.Load(WrapFunctionProvider(func(tagConf string, param struct{}) (*X, error) {
					return &X{ID: 1001}, nil
				}))
			},
			args: args{
				fn: func(in struct {
					x *X `gone:"*"`
				}) {
				},
				injectBefore: nil,
				injectAfter:  nil,
			},
			wantErr: false,
		},
		{
			name: "provide value err with struct pointer",
			load: func(loader Loader) error {
				return nil
			},
			args: args{
				fn: func(in *struct {
					x *X `gone:"*"`
				}) {
					t.Error("should not be called")
				},
				injectBefore: nil,
				injectAfter:  nil,
			},
			wantErr: true,
		},
		{
			name: "provide value suc with struct pointer",
			load: func(loader Loader) error {
				return loader.Load(WrapFunctionProvider(func(tagConf string, param struct{}) (*X, error) {
					return &X{ID: 1001}, nil
				}))
			},
			args: args{
				fn: func(in *struct {
					x *X `gone:"*"`
				}) {
					if in.x == nil {
						t.Error("in.x is nil")
						return
					}
					if in.x.ID != 1001 {
						t.Error("in.x.ID is not 1001")
						return
					}
				},
				injectBefore: nil,
				injectAfter:  nil,
			},
			wantErr: false,
		},
		{
			name: "use hook",
			load: func(loader Loader) error {
				return loader.Load(WrapFunctionProvider(func(tagConf string, param struct{}) (*X, error) {
					return &X{ID: 1001}, nil
				}))
			},
			args: args{
				fn: func(
					x string,
					in *struct {
						x *X `gone:"*"`
					},
					y int,
				) {
					if x != "test" {
						t.Errorf("x is not test")
						return
					}
					if in.x == nil {
						t.Error("in.x is nil")
						return
					}
					if in.x.ID != 1001 {
						t.Error("in.x.ID is not 1001")
						return
					}
					if y != 1002 {
						t.Error("y is not 1002")
						return
					}
				},
				injectBefore: func(pt reflect.Type, i int, injected bool) any {
					if i == 0 {
						return reflect.ValueOf("test")
					}
					return nil
				},
				injectAfter: func(pt reflect.Type, i int, injected bool) any {
					if i != 2 {
						return nil
					}
					return reflect.ValueOf(1002)
				},
			},
			wantErr: false,
		},
		{
			name: "injected err",
			load: func(loader Loader) error {
				return loader.Load(WrapFunctionProvider(func(tagConf string, param struct{}) (*X, error) {
					return &X{ID: 1001}, nil
				}))
			},
			args: args{
				fn: func(
					x string,
					in *struct {
						x X `gone:"*"`
					},
					y int,
				) {
					t.Error("should not be called")
				},
				injectBefore: nil,
				injectAfter:  nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			NewApp(tt.load).
				Run(func(fj FuncInjector) {
					_, err := fj.InjectFuncParameters(tt.args.fn, tt.args.injectBefore, tt.args.injectAfter)
					if (err != nil) != tt.wantErr {
						t.Errorf("InjectFuncParameters() error = %v, wantErr %v", err, tt.wantErr)
						return
					}
				})
		})
	}
}

func Test_core_InjectWrapFunc(t *testing.T) {
	type X struct {
		ID int
	}

	var x = &X{ID: 1001}

	type args struct {
		fn           any
		injectBefore FuncInjectHook
		injectAfter  FuncInjectHook
	}
	tests := []struct {
		name    string
		load    func(loader Loader) error
		args    args
		want    []any
		wantErr bool
	}{
		{
			name: "return nil",
			load: func(loader Loader) error {
				return nil
			},
			args: args{
				fn: func(in struct{}) (error, *X) {
					return nil, nil
				},
			},
			want: []any{nil, nil},
		},
		{
			name: "return nil",
			load: func(loader Loader) error {
				return nil
			},
			args: args{
				fn: func(in struct{}) (error, *X) {
					var z *X
					return nil, z
				},
			},
			want: []any{nil, nil},
		},
		{
			name: "return not nil",
			load: func(loader Loader) error {
				return nil
			},
			args: args{
				fn: func(in struct{}) (error, *X) {
					return nil, x
				},
			},
			want: []any{nil, x},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			NewApp(tt.load).
				Run(func(fj FuncInjector) {
					fn, err := fj.InjectWrapFunc(tt.args.fn, tt.args.injectBefore, tt.args.injectAfter)
					if (err != nil) != tt.wantErr {
						t.Errorf("InjectFuncParameters() error = %v, wantErr %v", err, tt.wantErr)
						return
					}
					if tt.want != nil {
						got := fn()

						if !reflect.DeepEqual(got, tt.want) {
							t.Errorf("InjectWrapFunc() got = %v, want %v", got, tt.want)
						}
					}
				})
		})
	}
}

func Test_core_Check(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	analyzer := NewMockiDependenceAnalyzer(controller)
	mockiKeeper := NewMockiKeeper(controller)
	logger := NewMockLogger(controller)

	s := &core{
		logger:              logger,
		iKeeper:             mockiKeeper,
		iDependenceAnalyzer: analyzer,
	}

	var x struct {
		Flag
	}

	tests := []struct {
		name    string
		setUp   func() func()
		want    []dependency
		wantErr bool
	}{
		{
			name: "err",
			setUp: func() func() {
				analyzer.EXPECT().checkCircularDepsAndGetBestInitOrder().Return(nil, nil, errors.New("err"))
				return func() {}
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "circular deps",
			setUp: func() func() {
				analyzer.EXPECT().checkCircularDepsAndGetBestInitOrder().Return([]dependency{{
					action: initAction,
					coffin: newCoffin(&struct{}{}),
				}, {
					action: initAction,
					coffin: newCoffin(&struct{}{}),
				}}, nil, nil)
				return func() {}
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "print debug info && success",
			setUp: func() func() {
				analyzer.EXPECT().checkCircularDepsAndGetBestInitOrder().Return(nil, []dependency{{
					action: initAction,
					coffin: newCoffin(&x),
				}}, nil)
				logger.EXPECT().GetLevel().Return(DebugLevel)
				logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).Times(1)

				mockiKeeper.EXPECT().getAllCoffins().Return([]*coffin{
					newCoffin(&x),
				})
				return func() {}
			},
			want: []dependency{
				{action: initAction, coffin: newCoffin(&x)},
				{action: fillAction, coffin: newCoffin(&x)},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer tt.setUp()()
			got, err := s.Check()
			if (err != nil) != tt.wantErr {
				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Check() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_core_Install(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	analyzer := NewMockiDependenceAnalyzer(controller)
	mockiKeeper := NewMockiKeeper(controller)
	logger := NewMockLogger(controller)
	mockiInstaller := NewMockiInstaller(controller)

	s := &core{
		logger:              logger,
		iKeeper:             mockiKeeper,
		iDependenceAnalyzer: analyzer,
		iInstaller:          mockiInstaller,
	}

	var x struct {
		Flag
	}

	tests := []struct {
		name    string
		setUp   func() func()
		wantErr bool
	}{
		{
			name: "err",
			setUp: func() func() {
				analyzer.EXPECT().checkCircularDepsAndGetBestInitOrder().Return(nil, nil, errors.New("err"))
				return func() {}
			},
			wantErr: true,
		},

		{
			name: "print debug info && success",
			setUp: func() func() {
				analyzer.EXPECT().checkCircularDepsAndGetBestInitOrder().Return(nil, []dependency{{
					action: initAction,
					coffin: newCoffin(&x),
				}}, nil)
				logger.EXPECT().GetLevel().Return(DebugLevel)
				logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).Times(1)

				mockiKeeper.EXPECT().getAllCoffins().Return([]*coffin{
					newCoffin(&x),
				})
				mockiInstaller.EXPECT().safeFillOne(gomock.Any()).Return(nil)
				mockiInstaller.EXPECT().safeInitOne(gomock.Any()).Return(nil)
				return func() {}
			},
			wantErr: false,
		},
		{
			name: "init error",
			setUp: func() func() {
				analyzer.EXPECT().checkCircularDepsAndGetBestInitOrder().Return(nil, []dependency{{
					action: initAction,
					coffin: newCoffin(&x),
				}}, nil)

				logger.EXPECT().GetLevel().Return(InfoLevel)
				logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

				mockiKeeper.EXPECT().getAllCoffins().Return([]*coffin{
					newCoffin(&x),
				})
				mockiInstaller.EXPECT().safeInitOne(gomock.Any()).Return(errors.New("err"))

				return func() {}
			},
			wantErr: true,
		},
		{
			name: "init error",
			setUp: func() func() {
				analyzer.EXPECT().checkCircularDepsAndGetBestInitOrder().Return(nil, []dependency{{
					action: initAction,
					coffin: newCoffin(&x),
				}}, nil)

				logger.EXPECT().GetLevel().Return(InfoLevel)
				logger.EXPECT().Debugf(gomock.Any(), gomock.Any())

				mockiKeeper.EXPECT().getAllCoffins().Return([]*coffin{
					newCoffin(&x),
				})
				mockiInstaller.EXPECT().safeFillOne(gomock.Any()).Return(errors.New("err"))
				mockiInstaller.EXPECT().safeInitOne(gomock.Any()).Return(nil)

				return func() {}
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer tt.setUp()()
			err := s.Install()
			if (err != nil) != tt.wantErr {
				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_core_GetGonerByPattern(t *testing.T) {
	type X struct {
		Flag
		Id int
	}

	t.Run("success", func(t *testing.T) {
		NewApp().
			Load(&X{Id: 1}, Name("x1")).
			Load(&X{Id: 2}, Name("x2")).
			Load(&X{Id: 3}, Name("x311")).
			Load(&X{Id: 4}, Name("y3")).
			Run(func(k Keeper) {
				goners := k.GetGonerByPattern(reflect.TypeOf(&X{}), "x*")
				if !reflect.DeepEqual(goners, []any{&X{Id: 1}, &X{Id: 2}, &X{Id: 3}}) {
					t.Errorf("GetGonerByPattern error")
				}
			})
	})

	t.Run("panic", func(t *testing.T) {
		provider := WrapFunctionProvider(func(tagConf string, param struct{}) (*X, error) {
			return nil, errors.New("error")
		})

		err := SafeExecute(func() error {
			NewApp().
				Load(provider, Name("x110")).
				Run(func(k Keeper) {
					goners := k.GetGonerByPattern(reflect.TypeOf(&X{}), "x*")
					fmt.Printf("%#v", goners)
				})
			return nil
		})
		if err == nil {
			t.Errorf("should be error")
		}
	})
}
