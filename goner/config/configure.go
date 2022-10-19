package config

import (
	"errors"
	"fmt"
	"github.com/gone-io/gone"
	"github.com/magiconair/properties"
	"reflect"
	"strconv"
	"time"
)

type propertiesConfigure struct {
	gone.Flag
	gone.Logger `gone:"gone-logger"`
	props       *properties.Properties
	cemetery    gone.Cemetery `gone:"gone-cemetery"`
}

func (c *propertiesConfigure) Get(key string, v interface{}, defaultVal string) error {
	if c.props == nil {
		env := GetEnv("")
		c.Infof("Use Env: %s\n", env)
		c.props = c.mustGetProperties()
	}
	return c.parseKeyFromProperties(key, v, defaultVal, c.props)
}

func (c *propertiesConfigure) parseKeyFromProperties(key string, value interface{}, defaultVale string, props *properties.Properties) error {
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return errors.New("type of value must be ptr")
	}

	el := rv.Elem()

	switch rv.Elem().Kind() {
	default:
		return errors.New(fmt.Sprintf("<%s>(%v) is not support", el.Type().Name(), el.Type().Kind()))
	case reflect.Struct:
		k := key + "."
		conf := props.FilterStripPrefix(k)
		err := conf.Decode(value)
		if err != nil {
			c.Errorf("err:", err)
		}
	case reflect.Bool:
		def, _ := strconv.ParseBool(defaultVale)
		el.SetBool(props.GetBool(key, def))
	case reflect.Int:
		def, _ := strconv.ParseInt(defaultVale, 10, 32)
		confVal := props.GetInt(key, int(def))
		el.SetInt(int64(confVal))
	case reflect.Int64:
		if isDuration(el.Type()) {
			var duration time.Duration
			var err error
			if defaultVale == "" {
				duration = 0
			} else {
				duration, err = time.ParseDuration(defaultVale)
				if err != nil {
					c.Errorf("err:", err)
				}
			}
			confVal := props.GetParsedDuration(key, duration)
			el.Set(reflect.ValueOf(confVal))
		} else {
			def, _ := strconv.ParseInt(defaultVale, 10, 64)
			confVal := props.GetInt64(key, def)
			el.Set(reflect.ValueOf(confVal))
		}
	case reflect.Uint:
		def, _ := strconv.ParseUint(defaultVale, 10, 32)
		confVal := props.GetUint(key, uint(def))
		el.Set(reflect.ValueOf(confVal))
	case reflect.Uint64:
		def, _ := strconv.ParseUint(defaultVale, 10, 64)
		confVal := props.GetUint64(key, def)
		el.Set(reflect.ValueOf(confVal))
	case reflect.Float64:
		def, _ := strconv.ParseFloat(defaultVale, 64)
		confVal := props.GetFloat64(key, def)
		el.Set(reflect.ValueOf(confVal))
	case reflect.String:
		confVal := props.GetString(key, defaultVale)
		rv.Elem().SetString(confVal)
	}

	return nil
}

func (c *propertiesConfigure) isInTestKit() bool {
	return c.cemetery.GetTomById(gone.IdGoneTestKit) != nil
}

func (c *propertiesConfigure) mustGetProperties() *properties.Properties {
	var props *properties.Properties
	var err error

	properties.LogPrintf = c.Warnf

	if c.isInTestKit() {
		props, err = GetTestProperties()
	} else {
		props, err = GetProperties("")
	}
	if err != nil {
		panic(err)
	}
	return props
}

func isDuration(t reflect.Type) bool { return t == reflect.TypeOf(time.Second) }
