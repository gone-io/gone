package gone

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

func newCemetery() Cemetery {
	return &cemetery{
		SimpleLogger: &defaultLogger{},
		tombMap:      make(map[GonerId]Tomb),
	}
}

type cemetery struct {
	Flag
	SimpleLogger `gone:"gone-logger"`
	tombMap      map[GonerId]Tomb
	tombs        []Tomb
}

func (c *cemetery) SetLogger(logger SimpleLogger) SetLoggerError {
	c.SimpleLogger = logger
	return nil
}

func GetGoneDefaultId(goner Goner) GonerId {
	elem := reflect.TypeOf(goner).Elem()
	pkgName := fmt.Sprintf("%s/%s", elem.PkgPath(), elem.Name())
	i, ok := goner.(identity)
	if ok {
		return GonerId(fmt.Sprintf("%s#%s", pkgName, i.GetId()))
	}
	return GonerId(fmt.Sprintf("%s#%d", pkgName, reflect.ValueOf(goner).Elem().UnsafeAddr()))
}

func (c *cemetery) bury(goner Goner, ids ...GonerId) Tomb {
	t := NewTomb(goner)
	var id GonerId
	if len(ids) > 0 {
		id = ids[0]
	}
	if id == "" {
		id = GetGoneDefaultId(goner)
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

func (c *cemetery) ReplaceBury(goner Goner, id GonerId) (err error) {
	if id == "" {
		err = ReplaceBuryIdParamEmptyError()
		return
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
	_, err = c.reviveOneFromTomb(replaceTomb)
	c.replaceTombsGonerField(id, goner, oldGoner, buried)
	return
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
					var iErr InnerError
					if errors.As(err, &iErr) {
						c.Errorf("inner Error: %s(code=%d)\n%s", iErr.Msg(), iErr.Code(), iErr.Stack())
					}
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

func (c *cemetery) setFieldValue(v reflect.Value, ref any) error {
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
		deps = append(deps, tomb)
		if tomb == nil {
			err = CannotFoundGonerByIdError(id)
			return
		}

		goner := tomb.GetGoner()
		if isCompatible(field.Type, goner) {
			err = c.setFieldValue(v, goner)
			suc = err == nil
			return
		}

		//如果不兼容，检查Goner是否为Vampire；对Vampire启动吸血行为
		if builder, ok := goner.(Vampire); ok {
			if !tomb.GonerIsRevive() {
				_, err = c.reviveOneAndItsDeps(tomb)
				if err != nil {
					return
				}
			}
			err = builder.Suck(extConfig, v)
			suc = err == nil
			return
		}

		if builder, ok := goner.(Vampire2); ok {
			if !tomb.GonerIsRevive() {
				_, err = c.reviveOneAndItsDeps(tomb)
				if err != nil {
					return
				}
			}
			err = builder.Suck(extConfig, v, field)
			suc = err == nil
			return
		}

		err = NotCompatibleError(field.Type, reflect.TypeOf(goner).Elem())
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

func (c *cemetery) reviveDependence(tomb Tomb) (deps []Tomb, err error) {
	deps, err = c.reviveOneAndItsDeps(tomb)
	if err != nil {
		return
	}

	err = c.prophesy(append(deps, tomb)...)
	return
}

func (c *cemetery) reviveOneAndItsDeps(tomb Tomb) (deps []Tomb, err error) {
	deps, err = c.reviveOneFromTomb(tomb)
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
			tmpDeps, err = c.reviveOneAndItsDeps(tomb)
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

func (c *cemetery) ReviveOne(goner any) (deps []Tomb, err error) {
	gonerType := reflect.TypeOf(goner).Elem()
	gonerValue := reflect.ValueOf(goner).Elem()

	goneTypeName := gonerType.PkgPath()
	if goneTypeName == "" {
		goneTypeName = gonerType.Name()
	} else {
		goneTypeName = goneTypeName + "/" + gonerType.Name()
	}

	if goneTypeName == "" {
		goneTypeName = "[Anonymous Goner]"
	}
	c.Infof("Revive %s", goneTypeName)

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
	return
}

func (c *cemetery) reviveOneFromTomb(tomb Tomb) (deps []Tomb, err error) {
	goner := tomb.GetGoner()

	deps, err = c.ReviveOne(goner)

	tomb.GonerIsRevive(true)
	return
}

func (c *cemetery) ReviveAllFromTombs() error {
	for _, tomb := range c.tombs {
		_, err := c.reviveOneFromTomb(tomb)
		if err != nil {
			return err
		}
	}
	return c.prophesy()
}

var obsessionPtr *Prophet
var obsessionType = reflect.TypeOf(obsessionPtr).Elem()

var obsessionPtr2 *Prophet2
var obsessionType2 = reflect.TypeOf(obsessionPtr2).Elem()

func (c *cemetery) prophesy(deps ...Tomb) error {
	var tombs []Tomb
	if len(deps) > 0 {
		tombs = Tombs(deps).GetTomByType(obsessionType)
	} else {
		tombs = c.GetTomByType(obsessionType)
	}

	for _, tomb := range tombs {
		obsession := tomb.GetGoner().(Prophet)
		err := obsession.AfterRevive()
		if err != nil {
			return err
		}
	}

	// deal with Prophet2
	if len(deps) > 0 {
		tombs = Tombs(deps).GetTomByType(obsessionType2)
	} else {
		tombs = c.GetTomByType(obsessionType2)
	}

	for _, tomb := range tombs {
		obsession := tomb.GetGoner().(Prophet2)
		err := obsession.AfterRevive()
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
