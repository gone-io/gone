package gone

import (
	"reflect"
	"unsafe"
)

func NewCemetery() Cemetery {
	return &cemetery{
		Logger:  &defaultLogger{},
		tombMap: make(map[GonerId]Tomb),
	}
}

type cemetery struct {
	Logger `inject:"-"`

	tombMap map[GonerId]Tomb
	tombs   []Tomb
}

func (c *cemetery) Bury(goner Goner, id GonerId) Tomb {
	t := NewTomb(goner)
	if id != "" {
		_, ok := c.tombMap[id]
		if ok {
			panic(GonerIdIsExistedError)
		}

		c.tombMap[id] = t.SetId(id)
	}
	c.tombs = append(c.tombs, t)
	return t
}

func (c *cemetery) ReplaceBury(Goner, GonerId) Tomb {
	//todo
	return nil
}

const goneTag = "gone"
const matchAll = "*"

func parseGoneTagId(tag string) (id GonerId, extend string) {
	//todo
	return
}

func (c *cemetery) reviveOne(tomb Tomb) (err error) {
	goner := tomb.GetGoner()

	gonerType := reflect.TypeOf(goner).Elem()
	gonerValue := reflect.ValueOf(goner).Elem()

	for i := 0; i < gonerValue.NumField(); i++ {
		field := gonerType.Field(i)
		tag := field.Tag.Get(goneTag)
		if tag == "" {
			continue
		}

		id, extConfig := parseGoneTagId(tag)
		t := field.Type
		v := gonerValue.Field(i)

		if !field.IsExported() {
			//黑魔法：让非导出字段可以访问
			v = reflect.NewAt(t, unsafe.Pointer(v.UnsafeAddr())).Elem()
		}

		//gone标记的是slice
		if t.Kind() == reflect.Slice {
			if id == matchAll {
				tombs := c.GetTomByType(t)
				print(tombs)
				//todo：设置数组
			} else {
				tomb := c.GetTomById(id)
				if tomb == nil {
					err = newCannotFoundGonerById(id)
					return
				}

				goner := tomb.GetGoner()
				builder, ok := goner.(Builder)
				if ok {
					err = builder.Build(extConfig, v)
					if err != nil {
						return
					}
				} else if IsCompatible(goner, t) {
					//todo: 设置数组
					v.Set(reflect.ValueOf(goner))
				} else {
					err = newNotCompatibleGonerError(id)
					return
				}
			}
		}

		//todo 结构体指针 或者 接口

		c.Infof("field field:%v, type:%v", field, v)
	}

	after, ok := goner.(ReviveAfter)
	if ok {
		err = after.After(c, tomb)
	}
	return
}

func (c *cemetery) revive() error {
	for _, tomb := range c.tombs {
		err := c.reviveOne(tomb)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *cemetery) GetTomById(id GonerId) Tomb {
	return c.tombMap[id]
}

func (c *cemetery) GetTomByType(t reflect.Type) (tombs []Tomb) {
	for _, tomb := range c.tombs {
		if IsCompatible(tomb.GetGoner(), t) {
			tombs = append(tombs, tomb)
		}
	}
	return
}

func IsCompatible(goner Goner, t reflect.Type) bool {
	gonerType := reflect.TypeOf(goner)
	if isInterface(gonerType) {
		return gonerType.Implements(t)
	}
	return gonerType == t
}

func isInterface(t reflect.Type) bool {
	return t.Kind() == reflect.Interface
}
