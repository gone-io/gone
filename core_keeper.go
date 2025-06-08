package gone

import "reflect"

func newKeeper() *keeper {
	return &keeper{
		coffins:        []*coffin{},
		nameMap:        make(map[string]*coffin),
		defaultTypeMap: make(map[reflect.Type]*coffin),
	}
}

type keeper struct {
	Flag
	coffins        []*coffin
	nameMap        map[string]*coffin
	defaultTypeMap map[reflect.Type]*coffin
}

func (s *keeper) getAllCoffins() []*coffin {
	return s.coffins
}

func (s *keeper) getByName(name string) *coffin {
	return s.nameMap[name]
}

func (s *keeper) getByTypeAndPattern(t reflect.Type, pattern string) (coffins []*coffin) {
	for _, co := range s.coffins {
		if co.onlyForName {
			continue
		}

		if err := co.CoundProvide(t, false); err == nil && isMatch(co.name, pattern) {
			coffins = append(coffins, co)
		}
	}

	SortCoffins(coffins)
	return coffins
}

func (s *keeper) selectOneCoffin(t reflect.Type, pattern string, warn func()) (depCo *coffin) {
	if depCos := s.getByTypeAndPattern(t, pattern); depCos != nil && len(depCos) > 0 {
		l := len(depCos)
		if l == 1 {
			depCo = depCos[0]
		} else if l > 1 {
			for _, c := range depCos {
				if c.isDefault(t) {
					depCo = c
					break
				}
			}
			if depCo == nil {
				if warn != nil {
					warn()
				}
				depCo = depCos[0]
			}
		}
	}
	return
}

func (s *keeper) load(goner Goner, options ...Option) error {
	if goner == nil {
		return NewInnerError("goner cannot be nil - must provide a valid Goner instance", LoadedError)
	}
	co := newCoffin(goner)

	for _, o := range options {
		if err := o.Apply(co); err != nil {
			return ToError(err)
		}
	}

	if co.name != "" {
		if _, ok := s.nameMap[co.name]; ok && !co.forceReplace {
			return NewInnerErrorWithParams(LoadedError, "goner with name %q is already loaded - use ForceReplace() option to override", co.name)
		} else {
			s.nameMap[co.name] = co
		}
	}

	var forceReplaceFind = false
	if co.forceReplace && co.name != "" {
		var replacedCo *coffin
		for i := range s.coffins {
			if s.coffins[i].name == co.name {
				replacedCo = s.coffins[i]
				s.coffins[i] = co
				forceReplaceFind = true
				break
			}
		}
		for t, typeCo := range s.defaultTypeMap {
			if typeCo == replacedCo {
				delete(s.defaultTypeMap, t)
			}
		}
	}

	if !forceReplaceFind {
		s.coffins = append(s.coffins, co)
	}

	for t := range co.defaultTypeMap {
		if _, ok := s.defaultTypeMap[t]; ok {
			return NewInnerErrorWithParams(
				LoadedError,
				"type %q is already registered as default - cannot use IsDefault option when Loading named provider: %q",
				GetTypeName(t),
				co.Name(),
			)
		} else {
			s.defaultTypeMap[t] = co
		}
	}
	return nil
}
