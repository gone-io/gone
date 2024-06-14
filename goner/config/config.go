package config

import (
	"github.com/gone-io/gone"
	"reflect"
	"strings"
)

type Configure = gone.Configure

func NewConfig() (gone.Vampire, gone.GonerId, gone.GonerOption, gone.GonerOption) {
	return &config{}, gone.IdConfig, gone.IsDefault(true), gone.Order0
}

type config struct {
	gone.Flag
	configure Configure `gone:"gone-configure"`
}

func ParseConfAnnotation(tag string) (key string, defaultVal string) {
	splitArray := strings.Split(tag, ",")
	key = strings.TrimSpace(splitArray[0])
	if strings.Contains(key, "=") {
		split := strings.Split(key, "=")
		key = split[0]
		if len(split) > 1 {
			defaultVal = split[1]
		}
		return
	}
	if len(splitArray) > 1 {
		for i := 1; i < len(splitArray); i++ {
			s := splitArray[i]
			arr := strings.Split(s, "=")
			if strings.TrimSpace(arr[0]) == "default" {
				if len(arr) > 1 {
					defaultVal = arr[1]
				}
				break
			}
		}
	}
	return key, defaultVal
}

func (f *config) Suck(conf string, v reflect.Value) gone.SuckError {
	key, defaultVal := ParseConfAnnotation(conf)

	if reflect.Ptr == v.Kind() {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return f.configure.Get(key, v.Addr().Interface(), defaultVal)
}
