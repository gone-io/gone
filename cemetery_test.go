package gone

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type XPoint interface {
	GetX() int
	GetY() int
	GetIndex() int
}
type Point struct {
	Flag
	x int
	y int

	Index int
}

func (p *Point) GetX() int {
	return p.x
}
func (p *Point) GetY() int {
	return p.y
}

func (p *Point) GetIndex() int {
	return p.Index
}

func Test_cemetery_revive(t *testing.T) {
	type fields struct {
		goneList []Tomb
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "inject field is Struct",
			fields: fields{
				goneList: []Tomb{
					NewTomb(&struct {
						Flag
						a Point `gone:"point-a"`
						b Point `gone:"point-b"`
					}{}).SetId("line"),
					NewTomb(&Point{x: 1, y: 2}).SetId("point-a"),
					NewTomb(&Point{x: 1, y: 2}).SetId("point-b"),
				},
			},
			wantErr: false,
		},
		{
			name: "inject field is Struct pointer",
			fields: fields{
				goneList: []Tomb{
					NewTomb(&struct {
						Flag
						a *Point `gone:"point-a"`
						b *Point `gone:"point-b"`
					}{}).SetId("line"),
					NewTomb(&Point{x: 1, y: 2}).SetId("point-a"),
					NewTomb(&Point{x: -1, y: -2}).SetId("point-b"),
				},
			},
			wantErr: false,
		},

		{
			name: "inject field is interface",
			fields: fields{
				goneList: []Tomb{
					NewTomb(&struct {
						Flag
						a XPoint `gone:"point-a"`
						b XPoint `gone:"point-b"`
					}{}).SetId("line"),
					NewTomb(&Point{x: 1, y: 2}).SetId("point-a"),
					NewTomb(&Point{x: -1, y: -2}).SetId("point-b"),
				},
			},
			wantErr: false,
		},

		{
			name: "inject field(with id) is interface slice",
			fields: fields{
				goneList: []Tomb{
					NewTomb(&struct {
						Flag
						c []XPoint `gone:"point-a"`
					}{}).SetId("line"),
					NewTomb(&Point{x: 1, y: 2}).SetId("point-a"),
					NewTomb(&Point{x: -1, y: -2}).SetId("point-b"),
				},
			},
			wantErr: true,
		},

		{
			name: "inject field is interface slice",
			fields: fields{
				goneList: []Tomb{
					NewTomb(&struct {
						Flag
						c []XPoint `gone:"*"`
					}{}).SetId("line"),
					NewTomb(&Point{x: 1, y: 2}).SetId("point-a"),
					NewTomb(&Point{x: -1, y: -2}).SetId("point-b"),
				},
			},
			wantErr: false,
		},

		{
			name: "inject field is struct pointer slice",
			fields: fields{
				goneList: []Tomb{
					NewTomb(&struct {
						Flag
						c []*Point `gone:"*"`
					}{}).SetId("line"),
					NewTomb(&Point{x: 1, y: 2}).SetId("point-a"),
					NewTomb(&Point{x: -1, y: -2}).SetId("point-b"),
				},
			},
			wantErr: false,
		},

		{
			name: "inject field is struct slice",
			fields: fields{
				goneList: []Tomb{
					NewTomb(&struct {
						Flag
						c []Point `gone:"*"`
					}{}).SetId("line"),
					NewTomb(&Point{x: 1, y: 2}).SetId("point-a"),
					NewTomb(&Point{x: -1, y: -2}).SetId("point-b"),
				},
			},
			wantErr: false,
		},
		{
			name: "inject field is map[string]interface",
			fields: fields{
				goneList: []Tomb{
					NewTomb(&struct {
						Flag
						c map[string]XPoint `gone:"*"`
					}{}).SetId("line"),
					NewTomb(&Point{x: 1, y: 2}).SetId("point-a"),
					NewTomb(&Point{x: -1, y: -2}).SetId("point-b"),
				},
			},
			wantErr: false,
		},

		{
			name: "inject field is map[string]interface",
			fields: fields{
				goneList: []Tomb{
					NewTomb(&struct {
						Flag
						c map[GonerId]XPoint `gone:"*"`
					}{}).SetId("line"),
					NewTomb(&Point{x: 1, y: 2}).SetId("point-a"),
					NewTomb(&Point{x: -1, y: -2}).SetId("point-b"),
				},
			},
			wantErr: false,
		},

		{
			name: "inject field is map[string]pointer",
			fields: fields{
				goneList: []Tomb{
					NewTomb(&struct {
						Flag
						c map[GonerId]*Point `gone:"*"`
					}{}).SetId("line"),
					NewTomb(&Point{x: 1, y: 2}).SetId("point-a"),
					NewTomb(&Point{x: -1, y: -2}).SetId("point-b"),
				},
			},
			wantErr: false,
		},

		{
			name: "inject field is map[string]struct",
			fields: fields{
				goneList: []Tomb{
					NewTomb(&struct {
						Flag
						c map[GonerId]Point `gone:"*"`
					}{}).SetId("line"),
					NewTomb(&Point{x: 1, y: 2}).SetId("point-a"),
					NewTomb(&Point{x: -1, y: -2}).SetId("point-b"),
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &cemetery{
				SimpleLogger: &defaultLogger{},
				tombMap:      make(map[GonerId]Tomb),
			}

			for _, tomb := range tt.fields.goneList {
				c.Bury(tomb.GetGoner(), tomb.GetId())
			}

			if err := c.ReviveAllFromTombs(); (err != nil) != tt.wantErr {
				t.Errorf("ReviveAllFromTombs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type TestLogger struct {
	Flag
	defaultLogger
	X int
}

type ZeroPoint struct {
	Flag
}

func (p *ZeroPoint) GetX() int {
	return 0
}
func (p *ZeroPoint) GetY() int {
	return 0
}

func (p *ZeroPoint) GetIndex() int {
	return 0
}

func Test_cemetery_ReplaceBury(t *testing.T) {
	t.Run("replace has default value field", func(t *testing.T) {
		c := &cemetery{
			SimpleLogger: &defaultLogger{},
			tombMap:      make(map[GonerId]Tomb),
		}
		c.Bury(c, IdGoneCemetery)

		err := c.ReviveAllFromTombs()
		assert.Nil(t, err)

		logger := TestLogger{X: 100}

		c.ReplaceBury(&logger, IdGoneLogger)

		assert.Equal(t, c.SimpleLogger, &logger)
	})

	t.Run("replace revived field", func(t *testing.T) {
		c := &cemetery{
			SimpleLogger: &defaultLogger{},
			tombMap:      make(map[GonerId]Tomb),
		}
		const line GonerId = "the-line"
		type Line struct {
			Flag
			A XPoint `gone:"point-a"`
			B XPoint `gone:"point-b"`
			c XPoint `gone:"point-a"`
		}

		c.Bury(c, IdGoneCemetery).
			Bury(&Line{}, line).
			Bury(&Point{x: -1, y: -2}, GonerId("point-a")).
			Bury(&Point{x: 1, y: 2}, GonerId("point-b"))

		err := c.ReviveAllFromTombs()
		assert.Nil(t, err)

		err = c.ReplaceBury(&ZeroPoint{}, GonerId("point-a"))
		assert.Nil(t, err)

		tomb := c.GetTomById(line)
		goner := tomb.GetGoner().(*Line)
		assert.Equal(t, goner.A.GetIndex(), 0)
		assert.Equal(t, goner.A.GetX(), 0)
		assert.Equal(t, goner.A.GetY(), 0)
	})

	t.Run("replace revived with empty goneId", func(t *testing.T) {
		c := newCemetery()
		err := c.ReplaceBury(&ZeroPoint{}, "")
		assert.Equal(t, err.(Error).Code(), ReplaceBuryIdParamEmpty)
	})

	t.Run("replace revived failed", func(t *testing.T) {
		c := &cemetery{
			SimpleLogger: &defaultLogger{},
			tombMap:      make(map[GonerId]Tomb),
		}
		const line GonerId = "the-line"
		type Line struct {
			Flag
			A XPoint `gone:"point-a"`
			B XPoint `gone:"point-b"`
			c XPoint `gone:"point-a"`
			z XPoint `gone:"point-z"`
		}

		c.Bury(c, IdGoneCemetery).
			Bury(&Line{
				z: &Point{},
			}, line).
			Bury(&Point{x: -1, y: -2}, GonerId("point-a")).
			Bury(&Point{x: 1, y: 2}, GonerId("point-b"))

		err := c.ReviveAllFromTombs()
		assert.Nil(t, err)

		err = c.ReplaceBury(&Line{}, line)
		assert.NotNil(t, err)
	})
}

func Test_cemetery_SetLogger(t *testing.T) {
	c := cemetery{}

	type TestLogger struct {
		defaultLogger
	}

	logger := &TestLogger{}

	err := c.SetLogger(logger)
	assert.Nil(t, err)
	assert.Equal(t, c.SimpleLogger, logger)
}

type identityGoner struct {
	Flag
}

func (i *identityGoner) GetId() GonerId {
	return "identityGoner"
}

func TestGetGoneDefaultId(t *testing.T) {
	type args struct {
		goner Goner
	}
	tests := []struct {
		name string
		args args
		want GonerId
	}{
		{
			name: "identityGoner",
			args: args{goner: &identityGoner{}},
			want: "github.com/gone-io/gone/identityGoner#identityGoner",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, GetGoneDefaultId(tt.args.goner), "GetGoneDefaultId(%v)", tt.args.goner)
		})
	}
}

func Test_cemetery_bury(t *testing.T) {
	t.Run("GonerIdIsExistedError", func(t *testing.T) {
		executed := false
		func() {
			defer func() {
				a := recover()

				assert.Equal(t, a.(Error).Code(), GonerIdIsExisted)
			}()

			c := newCemetery()
			c.Bury(&Point{x: 1, y: 2}, GonerId("point-a"))
			executed = true
			c.Bury(&Point{x: 1, y: 2}, GonerId("point-a"))
		}()
		assert.True(t, executed)
	})
}

func Test_parseGoneTagId(t *testing.T) {
	id, _ := parseGoneTagId("")
	assert.Equal(t, string(id), "")

	id, _ = parseGoneTagId("xxx")
	assert.Equal(t, string(id), "xxx")

	id, ext := parseGoneTagId("xxx,2222,2222333")
	assert.Equal(t, string(id), "xxx")
	assert.Equal(t, "2222,2222333", ext)
}

type vampire1 struct {
	Flag
}

func (g *vampire1) Suck(conf string, v reflect.Value) SuckError {
	v.SetString(conf)
	return nil
}

type vampire2 struct {
	Flag
}

func (g *vampire2) Suck(conf string, v reflect.Value, field reflect.StructField) error {
	v.SetString(conf + ":" + field.Name)
	return nil
}

func Test_cemetery_reviveFieldById(t *testing.T) {
	Prepare(func(cemetery Cemetery) error {
		cemetery.
			BuryOnce(&vampire1{}, GonerId("v1")).
			BuryOnce(&vampire1{}, GonerId("v1")).
			BuryOnce(&vampire2{}, GonerId("v2")).
			BuryOnce(&vampire2{}, GonerId("v2"))
		return nil
	}).AfterStart(func(in struct {
		test1 string      `gone:"v1,xxxx"`
		test2 string      `gone:"v2,xxxx"`
		test3 []*vampire2 `gone:"*"`
		test4 []*vampire1 `gone:"*"`
	}) {
		assert.Equal(t, "xxxx", in.test1)
		assert.Equal(t, "xxxx:test2", in.test2)
		assert.Equal(t, 1, len(in.test3))
		assert.Equal(t, 1, len(in.test4))
	}).Run()

}

func Test_cemetery_checkRevive(t *testing.T) {
	type Line struct {
		Flag
		PointA *Point `gone:"point-a"`
		PointB *Point `gone:"point-b"`
	}

	test := NewBuryMockCemeteryForTest()
	test.
		Bury(&Point{x: 1, y: 2, Index: 1}, GonerId("point-a")).
		Bury(&Point{x: 1, y: 2, Index: 2}, GonerId("point-b")).
		Bury(&Line{}, GonerId("line"))

	theTomb := test.GetTomById(GonerId("line"))

	err := test.(*cemetery).checkRevive(theTomb)
	assert.Nil(t, err)
}

func Test_cemetery_InjectFuncParameters(t *testing.T) {
	Prepare().Test(func(cemetery Cemetery) {
		t.Run("fn is not a func", func(t *testing.T) {
			_, err := cemetery.InjectFuncParameters("", nil, nil)
			assert.Error(t, err)
			assert.Equal(t, err.(Error).Code(), NotCompatible)
		})

		t.Run("fn has two parameters", func(t *testing.T) {
			executed := false
			fn := func(cemetery Cemetery, in struct {
				cemetery Cemetery `gone:"*"`
			}) {
				assert.NotNil(t, in.cemetery)
				assert.NotNil(t, cemetery)
				assert.Equal(t, cemetery, in.cemetery)
				executed = true
			}

			args, err := cemetery.InjectFuncParameters(fn, nil, nil)
			assert.Nil(t, err)

			reflect.ValueOf(fn).Call(args)
			assert.True(t, executed)
		})

		t.Run("fn parameter is not struct", func(t *testing.T) {
			fn := func(int) {}
			_, err := cemetery.InjectFuncParameters(fn, nil, nil)
			assert.Error(t, err)
			assert.Equal(t, err.(Error).Code(), NotCompatible)
		})

		t.Run("filter some parameters", func(t *testing.T) {
			executed := false
			fn := func(
				cemetery Cemetery,
				in struct {
					cemetery Cemetery `gone:"*"`
				},
				test bool,
			) {
				assert.NotNil(t, in.cemetery)
				assert.NotNil(t, cemetery)
				assert.Equal(t, cemetery, in.cemetery)
				assert.Equal(t, true, test)
				executed = true
			}

			args, err := cemetery.InjectFuncParameters(fn, func(pt reflect.Type, i int) any {
				if i == 2 {
					return true
				}
				return nil
			}, func(pt reflect.Type, i int) {
				assert.Equal(t, 1, i)
			})

			assert.Nil(t, err)
			assert.Equal(t, 3, len(args))
			reflect.ValueOf(fn).Call(args)
			assert.True(t, executed)
		})

		t.Run("Revive Failed", func(t *testing.T) {
			fn := func(in struct {
				cemetery Cemetery `gone:"xxxxx"`
			}) {
			}
			_, err := cemetery.InjectFuncParameters(fn, nil, nil)
			assert.Error(t, err)
			assert.Equal(t, CannotFoundGonerById, err.(Error).Code())
		})
	})
}

func Test_cemetery_BuryOnce(t *testing.T) {
	test := NewBuryMockCemeteryForTest()
	defer func() {
		err := recover()
		assert.Error(t, err.(Error))
	}()
	type X struct {
		Flag
	}
	test.BuryOnce(&X{})
}
