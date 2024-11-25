package gone

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unsafe"
)

func newCemetery() Cemetery {
	return &cemetery{
		Logger:      _defaultLogger,
		tombMap:     make(map[GonerId]Tomb),
		providerMap: make(map[reflect.Type]Tomb),
	}
}

type cemetery struct {
	Flag
	Logger      `gone:"gone-logger"`
	tombMap     map[GonerId]Tomb
	tombs       Tombs
	providerMap map[reflect.Type]Tomb
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

func (c *cemetery) buildTomb(goner Goner, options ...GonerOption) Tomb {
	newTomb := NewTomb(goner)
	for _, option := range options {
		switch option.(type) {
		case GonerId:
			newTomb.SetId(option.(GonerId))
		case defaultType:
			newTomb.SetDefault(option.(defaultType).t)
		case Order:
			newTomb.SetOrder(option.(Order))
		case provideType:
			if v, ok := goner.(Vampire2); ok {
				for _, t := range option.(provideType).t {
					if existed, ok := c.providerMap[t]; ok {
						c.Warnf("%s/%s is provided by %T, its provider will be replaced by %T", t.PkgPath(), t.Name(), existed, v)
					}
					c.providerMap[t] = newTomb
				}
			}
		}
	}
	return newTomb
}

func (c *cemetery) bury(goner Goner, options ...GonerOption) Tomb {
	theTomb := c.buildTomb(goner, options...)

	if theTomb.GetId() == "" {
		theTomb.SetId(GetGoneDefaultId(goner))
	}

	_, ok := c.tombMap[theTomb.GetId()]
	if ok {
		panic(GonerIdIsExistedError(theTomb.GetId()))
	}
	c.tombMap[theTomb.GetId()] = theTomb
	c.tombs = append(c.tombs, theTomb)
	return theTomb
}

func (c *cemetery) Bury(goner Goner, options ...GonerOption) Cemetery {
	c.bury(goner, options...)
	return c
}

func (c *cemetery) filterGonerIdFromOptions(options []GonerOption) GonerId {
	var id GonerId
loop:
	for _, option := range options {
		switch option.(type) {
		case GonerId:
			id = option.(GonerId)
			break loop
		}
	}
	return id
}

func (c *cemetery) BuryOnce(goner Goner, options ...GonerOption) Cemetery {
	var id = c.filterGonerIdFromOptions(options)
	if id == "" {
		panic(NewInnerError("GonerId is empty, must have gonerId option", MustHaveGonerId))
	}

	if nil == c.GetTomById(id) {
		c.Bury(goner, options...)
	}
	return c
}

func (c *cemetery) ReplaceBury(goner Goner, options ...GonerOption) (err error) {
	newTomb := c.buildTomb(goner, options...)
	if newTomb.GetId() == "" {
		err = ReplaceBuryIdParamEmptyError()
		return
	}

	oldTomb, buried := c.tombMap[newTomb.GetId()]
	c.tombMap[newTomb.GetId()] = newTomb

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

	c.tombs = append(c.tombs, newTomb)
	_, err = c.reviveOneFromTomb(newTomb)
	if err != nil {
		return err
	}
	return c.replaceTombsGonerField(newTomb.GetId(), goner, oldGoner, buried)
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
				v = BlackMagic(v)
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
	if tag == "" {
		return
	}
	list := strings.SplitN(tag, ",", 2)
	switch len(list) {
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
		aTomb := c.GetTomById(id)
		deps = append(deps, aTomb)
		if aTomb == nil {
			err = CannotFoundGonerByIdError(id)
			return
		}

		goner := aTomb.GetGoner()
		if IsCompatible(field.Type, goner) {
			setFieldValue(v, goner)
			suc = true
			return
		}

		err = c.checkRevive(aTomb)
		if err != nil {
			return
		}

		if suc, err = c.reviveByVampire(goner, extConfig, v); err != nil || suc {
			return
		}

		if suc, err = c.reviveByVampire2(goner, extConfig, v, field); err != nil || suc {
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

func (c *cemetery) reviveByVampire(goner Goner, extConfig string, v reflect.Value) (suc bool, err error) {
	if builder, ok := goner.(Vampire); ok {
		err = builder.Suck(extConfig, v)
		return err == nil, err
	}
	return false, nil
}

func (c *cemetery) reviveByVampire2(goner Goner, extConfig string, v reflect.Value, field reflect.StructField) (suc bool, err error) {
	if builder, ok := goner.(Vampire2); ok {
		err = builder.Suck(extConfig, v, field)
		return err == nil, err
	}
	return false, nil
}

func (c *cemetery) reviveFieldByType(gonerType reflect.Type, field reflect.StructField, v reflect.Value) (deps []Tomb, suc bool) {
	goneTypeName := getNameByType(gonerType)
	filedName := fmt.Sprintf("%s.%s", goneTypeName, field.Name)

	container := c.getGonerContainerByType(field.Type, filedName)
	if container != nil {
		setFieldValue(v, container.GetGoner())
		suc = true
		deps = append(deps, container)
	}
	return
}

func (c *cemetery) reviveFieldByProvider(tag string, field reflect.StructField, v reflect.Value) (deps []Tomb, suc bool, err error) {
	if aTomb, ok := c.providerMap[field.Type]; ok {
		err = c.checkRevive(aTomb)
		if err != nil {
			return
		}
		deps = append(deps, aTomb)
		vampire2 := aTomb.GetGoner().(Vampire2)
		_, extConfig := parseGoneTagId(tag)

		err = vampire2.Suck(extConfig, v, field)

		suc = err == nil
	}
	return
}

func (c *cemetery) reviveSpecialTypeFields(field reflect.StructField, v reflect.Value) (deps []Tomb, suc bool) {
	t := field.Type
	switch t.Kind() {

	case reflect.Slice: //support inject slice
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

	case reflect.Map: //support inject map
		if t.Key().Kind() == reflect.String { //key of map must be string
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
		suc = false
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

func getNameByType(gonerType reflect.Type) string {
	goneTypeName := gonerType.PkgPath()
	if goneTypeName == "" {
		goneTypeName = gonerType.Name()
	} else {
		goneTypeName = goneTypeName + "/" + gonerType.Name()
	}

	if goneTypeName == "" {
		goneTypeName = "[Anonymous Goner]"
	}
	return goneTypeName
}

func (c *cemetery) ReviveOne(goner any) (deps []Tomb, err error) {
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
			v = BlackMagic(v)
		}

		//do not inject multiple times
		if !v.IsZero() {
			continue
		}

		var suc bool
		var tmpDeps []Tomb

		// inject by id
		if tmpDeps, suc, err = c.reviveFieldById(tag, field, v); err != nil {
			return
		} else if suc {
			deps = append(deps, tmpDeps...)
			continue
		}

		// inject by type
		if tmpDeps, suc = c.reviveFieldByType(gonerType, field, v); suc {
			deps = append(deps, tmpDeps...)
			continue
		}

		if tmpDeps, suc, err = c.reviveFieldByProvider(tag, field, v); err != nil {
			return
		} else if suc {
			deps = append(deps, tmpDeps...)
			continue
		}

		// inject special types
		if tmpDeps, suc = c.reviveSpecialTypeFields(field, v); suc {
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
	if err != nil {
		return nil, err
	}
	tomb.GonerIsRevive(true)
	return
}

func (c *cemetery) ReviveAllFromTombs() error {
	sort.Sort(c.tombs)
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

func (c *cemetery) prophesy() (err error) {
	var tombs Tombs = c.GetTomByType(obsessionType)
	tombs = append(tombs, c.GetTomByType(obsessionType2)...)
	sort.Sort(tombs)

	for _, tomb := range tombs {
		goner := tomb.GetGoner()
		switch goner.(type) {
		case Prophet:
			err = goner.(Prophet).AfterRevive()
		case Prophet2:
			err = goner.(Prophet2).AfterRevive()
		}

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
	return c.tombs.GetTomByType(t)
}

func (c *cemetery) getGonerContainerByType(t reflect.Type, name string) Tomb {
	tombs := c.GetTomByType(t)
	if len(tombs) > 0 {
		var container Tomb

		for _, tmp := range tombs {
			if tmp.IsDefault(t) {
				container = tmp
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

func (c *cemetery) InjectFuncParameters(fn any, injectBefore func(pt reflect.Type, i int) any, injectAfter func(pt reflect.Type, i int)) (args []reflect.Value, err error) {
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
				args = append(args, reflect.ValueOf(x))
				continue
			}
		}

		x := getOnlyOne(pt, i+1)
		if x != nil {
			args = append(args, reflect.ValueOf(x))
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

		args = append(args, parameter.Elem())
		if injectAfter != nil {
			injectAfter(pt, i)
		}
	}
	return
}

func BlackMagic(v reflect.Value) reflect.Value {
	return reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()
}
