package gone

import (
	"reflect"
	"strings"
	"unsafe"
)

func NewCemetery() Cemetery {
	return &cemetery{
		Logger:  &defaultLogger{},
		tombMap: make(map[GonerId]Tomb),
	}
}

type cemetery struct {
	GonerFlag

	Logger `gone:"gone-logger"`

	tombMap map[GonerId]Tomb
	tombs   []Tomb
}

func (c *cemetery) bury(goner Goner, ids ...GonerId) Tomb {
	t := NewTomb(goner)
	var id GonerId
	if len(ids) > 0 {
		id = ids[0]
	}

	if id != "" {
		_, ok := c.tombMap[id]
		if ok {
			panic(GonerIdIsExistedError(id))
		}

		c.tombMap[id] = t.SetId(id)
	}
	c.tombs = append(c.tombs, t)
	return t
}

func (c *cemetery) Bury(goner Goner, ids ...GonerId) Cemetery {
	c.bury(goner, ids...)
	return c
}

func (c *cemetery) ReplaceBury(goner Goner, id GonerId) Cemetery {
	if id == "" {
		return c
	}

	oldTomb := c.tombMap[id]
	replaceTomb := NewTomb(goner).SetId(id)
	c.tombMap[id] = replaceTomb

	buried := oldTomb != nil
	var oldGoner Goner
	if buried {
		oldGoner = oldTomb.GetGoner()
		for i := 0; i < len(c.tombs); i++ {
			itemGoner := c.tombs[i].GetGoner()
			if itemGoner == oldGoner {
				c.tombs = append(c.tombs[:i], c.tombs[i+1:]...)
			}
		}
	}

	c.tombs = append(c.tombs, replaceTomb)
	_, err := c.reviveOne(replaceTomb)

	c.replaceTombsGonerField(id, goner, oldGoner, buried)

	if err != nil {
		panic(err)
	}
	return c
}

func (c *cemetery) replaceTombsGonerField(id GonerId, newGoner, oldGoner Goner, buried bool) {
	for _, tomb := range c.tombs {
		goner := tomb.GetGoner()

		gonerType := reflect.TypeOf(goner).Elem()
		gonerValue := reflect.ValueOf(goner).Elem()

		for i := 0; i < gonerValue.NumField(); i++ {
			field := gonerType.Field(i)
			tag := field.Tag.Get(goneTag)
			if tag == "" {
				continue
			}

			v := gonerValue.Field(i)
			if !field.IsExported() {
				//黑魔法：让非导出字段可以访问
				v = reflect.NewAt(field.Type, unsafe.Pointer(v.UnsafeAddr())).Elem()
			}

			if buried && v.Interface() == oldGoner {
				v.Set(reflect.ValueOf(newGoner))
				continue
			}
			oldId, _ := parseGoneTagId(tag)
			if oldId == id {
				_, _, err := c.reviveFieldById(tag, field, v)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

const goneTag = "gone"
const anonymous = "*"

func parseGoneTagId(tag string) (id GonerId, extend string) {
	list := strings.SplitN(tag, ",", 2)
	switch len(list) {
	case 0:
		return
	case 1:
		id = GonerId(list[0])
	default:
		id, extend = GonerId(list[0]), list[1]
	}
	return
}

// 兼容：t类型可以装下goner
func isCompatible(t reflect.Type, goner Goner) bool {
	gonerType := reflect.TypeOf(goner)

	switch t.Kind() {
	case reflect.Interface:
		return gonerType.Implements(t)
	case reflect.Struct:
		return gonerType.Elem() == t
	default:
		return gonerType == t
	}
}

func (c *cemetery) setFieldValue(v reflect.Value, ref interface{}) error {
	t := v.Type()

	switch t.Kind() {
	case reflect.Interface, reflect.Pointer, reflect.Slice, reflect.Map:
		v.Set(reflect.ValueOf(ref))
	default:
		v.Set(reflect.ValueOf(ref).Elem())
	}
	return nil
}

func (c *cemetery) reviveFieldById(tag string, field reflect.StructField, v reflect.Value) (deps []Tomb, suc bool, err error) {
	id, extConfig := parseGoneTagId(tag)
	if id != anonymous {
		tomb := c.GetTomById(id)
		if tomb == nil {
			err = CannotFoundGonerByIdError(id)
			return
		}

		goner := tomb.GetGoner()
		builder, ok := goner.(Vampire)
		if ok {
			err = builder.Suck(extConfig, v)
			if err != nil {
				return
			}
		} else {
			if !isCompatible(field.Type, goner) {
				err = NotCompatibleError(field.Type, reflect.TypeOf(goner).Elem())
				return
			}

			err = c.setFieldValue(v, goner)
			if err != nil {
				return
			}
		}
		suc = true
		deps = append(deps, tomb)
	}
	return
}

func (c *cemetery) reviveFieldByType(field reflect.StructField, v reflect.Value) (deps []Tomb, suc bool, err error) {
	tombs := c.GetTomByType(field.Type)
	if len(tombs) > 0 {
		if len(tombs) > 1 {
			c.Warnf("more than one goner was found, use first one!")
		}
		tomb := tombs[0]
		err = c.setFieldValue(v, tomb.GetGoner())
		if err != nil {
			return
		}
		suc = true
		deps = append(deps, tomb)
	}
	return
}

func (c *cemetery) reviveSpecialTypeFields(field reflect.StructField, v reflect.Value) (deps []Tomb, suc bool, err error) {
	t := field.Type
	switch t.Kind() {

	case reflect.Slice: //允许注入接口切片
		tombs := c.GetTomByType(t.Elem())
		for _, tomb := range tombs {
			if t.Elem().Kind() == reflect.Struct {
				v.Set(reflect.Append(v, reflect.ValueOf(tomb.GetGoner()).Elem()))
			} else {
				v.Set(reflect.Append(v, reflect.ValueOf(tomb.GetGoner())))
			}
			deps = append(deps, tomb)
		}
		suc = true

	case reflect.Map: //允许注入接口Map
		if t.Key().Kind() == reflect.String { //Map的key是string类型
			tombs := c.GetTomByType(t.Elem())

			m := reflect.MakeMap(t)
			for _, tomb := range tombs {
				id := tomb.GetId()
				if id != "" {
					if t.Elem().Kind() == reflect.Struct {
						m.SetMapIndex(reflect.ValueOf(id).Convert(t.Key()), reflect.ValueOf(tomb.GetGoner()).Elem())
					} else {
						m.SetMapIndex(reflect.ValueOf(id).Convert(t.Key()), reflect.ValueOf(tomb.GetGoner()))
					}
					deps = append(deps, tomb)
				}
			}
			v.Set(m)
			suc = true
		}
	default:
	}
	return
}

func (c *cemetery) reviveOneDep(tomb Tomb) (deps []Tomb, err error) {
	deps, err = c.reviveOne(tomb)
	if err != nil {
		return
	}

	m := make(map[Tomb]bool)
	for _, dep := range deps {
		m[dep] = true
	}

	for _, tomb := range deps {
		if !tomb.GonerIsRevive() {
			var tmpDeps []Tomb
			tmpDeps, err = c.reviveOneDep(tomb)
			if err != nil {
				return
			}
			for _, dep := range tmpDeps {
				m[dep] = true
			}

		}
	}
	deps = make([]Tomb, 0, len(m))
	for dep := range m {
		deps = append(deps, dep)
	}
	return
}

func (c *cemetery) reviveOne(tomb Tomb) (deps []Tomb, err error) {
	goner := tomb.GetGoner()

	gonerType := reflect.TypeOf(goner).Elem()
	gonerValue := reflect.ValueOf(goner).Elem()

	for i := 0; i < gonerValue.NumField(); i++ {
		field := gonerType.Field(i)
		tag := field.Tag.Get(goneTag)
		if tag == "" {
			continue
		}

		v := gonerValue.Field(i)
		if !field.IsExported() {
			//黑魔法：让非导出字段可以访问
			v = reflect.NewAt(field.Type, unsafe.Pointer(v.UnsafeAddr())).Elem()
		}

		//如果已经存在值，不再注入
		if !v.IsZero() {
			continue
		}

		var suc bool
		var tmpDeps []Tomb

		// 根据Id匹配
		if tmpDeps, suc, err = c.reviveFieldById(tag, field, v); err != nil {
			return
		} else if suc {
			deps = append(deps, tmpDeps...)
			continue
		}

		// 根据类型匹配
		if tmpDeps, suc, err = c.reviveFieldByType(field, v); err != nil {
			return
		} else if suc {
			deps = append(deps, tmpDeps...)
			continue
		}

		// 特殊类型处理
		if tmpDeps, suc, err = c.reviveSpecialTypeFields(field, v); err != nil {
			return
		} else if suc {
			deps = append(deps, tmpDeps...)
			continue
		}

		return deps, CannotFoundGonerByTypeError(field.Type)
	}

	tomb.GonerIsRevive(true)

	after, ok := goner.(ReviveAfter)
	if ok {
		err = after.After(c, tomb)
	}
	return
}

func (c *cemetery) revive() error {
	for _, tomb := range c.tombs {
		_, err := c.reviveOne(tomb)
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
	return Tombs(c.tombs).GetTomByType(t)
}
