package gone

import (
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

func (c *cemetery) bury(goner Goner, options ...GonerOption) Tomb {
	t := NewTomb(goner)
	var id GonerId

	for _, option := range options {
		switch option.(type) {
		case GonerId:
			id = option.(GonerId)
		case IsDefault:
			t.SetDefault(bool(option.(IsDefault)))
		}
	}

	if id == "" {
		id = GetGoneDefaultId(goner)
	}

	_, ok := c.tombMap[id]
	if ok {
		panic(GonerIdIsExistedError(id))
	}

	c.tombMap[id] = t.SetId(id)
	c.tombs = append(c.tombs, t)
	return t
}

func (c *cemetery) Bury(goner Goner, options ...GonerOption) Cemetery {
	c.bury(goner, options...)
	return c
}

func (c *cemetery) BuryOnce(goner Goner, options ...GonerOption) Cemetery {
	var id GonerId

	for _, option := range options {
		switch option.(type) {
		case GonerId:
			id = option.(GonerId)
		}
	}
	if id == "" {
		panic(NewInnerError("GonerId is empty, must have gonerId option", MustHaveGonerId))
	}

	if nil == c.GetTomById(id) {
		c.Bury(goner, options...)
	}
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
	if err != nil {
		return err
	}
	return c.replaceTombsGonerField(id, goner, oldGoner, buried)
}

func (c *cemetery) replaceTombsGonerField(id GonerId, newGoner, oldGoner Goner, buried bool) error {
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
					return err
				}
			}
		}
	}
	return nil
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
		if IsCompatible(field.Type, goner) {
			c.setFieldValue(v, goner)
			suc = true
			return
		}

		if suc, err = c.reviveByVampire(goner, tomb, extConfig, v); err != nil || suc {
			return
		}

		if suc, err = c.reviveByVampire2(goner, tomb, extConfig, v, field); err != nil || suc {
			return
		}

		err = NotCompatibleError(field.Type, reflect.TypeOf(goner).Elem())
	}
	return
}

func (c *cemetery) checkRevive(tomb Tomb) error {
	if !tomb.GonerIsRevive() {
		_, err := c.reviveOneAndItsDeps(tomb)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *cemetery) reviveByVampire(goner Goner, tomb Tomb, extConfig string, v reflect.Value) (suc bool, err error) {
	if builder, ok := goner.(Vampire); ok {
		err = c.checkRevive(tomb)
		if err != nil {
			return
		}

		err = builder.Suck(extConfig, v)
		return err == nil, err
	}
	return false, nil
}

func (c *cemetery) reviveByVampire2(goner Goner, tomb Tomb, extConfig string, v reflect.Value, field reflect.StructField) (suc bool, err error) {
	if builder, ok := goner.(Vampire2); ok {
		err = c.checkRevive(tomb)
		if err != nil {
			return
		}
		err = builder.Suck(extConfig, v, field)
		return err == nil, err
	}
	return false, nil
}

func (c *cemetery) reviveFieldByType(field reflect.StructField, v reflect.Value, goneTypeName string) (deps []Tomb, suc bool, err error) {
	container := c.getGonerContainerByType(field.Type, fmt.Sprintf("%s.%s", goneTypeName, field.Name))
	if container != nil {
		c.setFieldValue(v, container.GetGoner())
		suc = true
		deps = append(deps, container)
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
	}
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
		if tmpDeps, suc, err = c.reviveFieldByType(field, v, goneTypeName); err != nil {
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

func (c *cemetery) getGonerContainerByType(t reflect.Type, name string) Tomb {
	tombs := c.GetTomByType(t)
	if len(tombs) > 0 {
		var container Tomb

		for _, t := range tombs {
			if t.IsDefault() {
				container = t
				break
			}
		}
		if container == nil {
			container = tombs[0]
			if len(tombs) > 1 {
				c.Warnf(fmt.Sprintf("inject %s more than one goner was found and no default, used the first!", name))
			}
		}
		return container
	}
	return nil
}

func (c *cemetery) InjectFuncParameters(fn any, injectBefore func(pt reflect.Type, i int) any, injectAfter func(pt reflect.Type, i int, obj *any)) (args []any, err error) {
	ft := reflect.TypeOf(fn)
	if ft.Kind() != reflect.Func {
		return nil, NewInnerError("fn must be a function", NotCompatible)
	}

	in := ft.NumIn()

	getOnlyOne := func(pt reflect.Type, i int) Goner {
		container := c.getGonerContainerByType(pt, fmt.Sprintf("%d parameter of %s", i, GetFuncName(fn)))
		if container != nil {
			return container.GetGoner()
		}
		return nil
	}

	for i := 0; i < in; i++ {
		pt := ft.In(i)
		if injectBefore != nil {
			x := injectBefore(pt, i)
			if x != nil {
				args = append(args, x)
				continue
			}
		}

		x := getOnlyOne(pt, i+1)
		if x != nil {
			args = append(args, x)
			continue
		}

		if pt.Kind() != reflect.Struct {
			err = NewInnerError(fmt.Sprintf("%dth parameter of %s must be a struct", i+1, GetFuncName(fn)), NotCompatible)
			return
		}

		parameter := reflect.New(pt)
		goner := parameter.Interface()
		_, err = c.ReviveOne(goner)
		if err != nil {
			return
		}
		obj := parameter.Elem().Interface()
		args = append(args, obj)
		if injectAfter != nil {
			injectAfter(pt, i, &obj)
		}
	}
	return
}
