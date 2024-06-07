package config

import (
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

func (c *propertiesConfigure) Get(key string, v any, defaultVal string) error {
	if c.props == nil {
		env := GetEnv("")
		c.Infof("==>Use Env: %s", env)
		var err error
		c.props, err = c.mustGetProperties()
		if err != nil {
			return gone.ToError(err)
		}
	}
	return parseKeyFromProperties(key, v, defaultVal, c.props)
}

const SliceMaxSize = 100

type PropertiesConfigure interface {
	FilterStripPrefix(prefix string) *properties.Properties
	Decode(x interface{}) error
	GetParsedDuration(key string, def time.Duration) time.Duration
	GetBool(key string, def bool) bool
	GetInt(key string, def int) int
	GetInt64(string, int64) int64
	GetUint(string, uint) uint
	GetUint64(string, uint64) uint64
	GetFloat64(string, float64) float64
	GetString(string, string) string
	Len() int
}

func parseKeyFromProperties(
	key string,
	value any,
	defaultVale string,
	props PropertiesConfigure,
) error {
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return gone.NewInnerError("type of value must be ptr", gone.NotCompatible)
	}

	el := rv.Elem()

	switch rv.Elem().Kind() {
	default:
		return gone.NewInnerError(fmt.Sprintf("<%s>(%v) is not support", el.Type().Name(), el.Type().Kind()), gone.NotCompatible)
	case reflect.Struct:
		k := key + "."
		conf := props.FilterStripPrefix(k)
		err := conf.Decode(value)
		return gone.ToError(err)
	case reflect.Slice:
		sliceElementType := el.Type().Elem()

		for i := 0; i < SliceMaxSize; i++ {
			k := fmt.Sprintf("%s[%d].", key, i)
			conf := props.FilterStripPrefix(k)
			if conf.Len() == 0 {
				break
			}
			err := decodeSlice(sliceElementType, k, conf, el)
			if err != nil {
				return err
			}
		}

	case reflect.Bool:
		def, _ := strconv.ParseBool(defaultVale)
		el.SetBool(props.GetBool(key, def))
	case reflect.Int, reflect.Int32, reflect.Int16, reflect.Int8:
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
					return gone.ToError(err)
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

	case reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		def, _ := strconv.ParseUint(defaultVale, 10, 64)
		confVal := props.GetUint64(key, def)
		el.SetUint(confVal)

	case reflect.Float64, reflect.Float32:
		def, _ := strconv.ParseFloat(defaultVale, 64)
		confVal := props.GetFloat64(key, def)
		el.SetFloat(confVal)

	case reflect.String:
		confVal := props.GetString(key, defaultVale)
		rv.Elem().SetString(confVal)
	}

	return nil
}

func decodeSlice(sliceElementType reflect.Type, k string, conf PropertiesConfigure, el reflect.Value) error {
	switch sliceElementType.Kind() {
	case reflect.Struct:
		v := reflect.New(sliceElementType)
		err := conf.Decode(v.Interface())
		if nil != err {
			return gone.NewInnerError(fmt.Sprintf("config %s err:%s", k, err.Error()), gone.NotCompatible)
		}
		el.Set(reflect.Append(el, v.Elem()))
	case reflect.Pointer:
		if sliceElementType.Elem().Kind() == reflect.Struct {
			v := reflect.New(sliceElementType.Elem())
			err := conf.Decode(v.Interface())
			if nil != err {
				return gone.NewInnerError(fmt.Sprintf("config %s err:%s", k, err.Error()), gone.NotCompatible)
			}
			el.Set(reflect.Append(el, v))
		} else {
			return gone.NewInnerError(fmt.Sprintf("config %s err: bad type", k), gone.NotCompatible)
		}
	default:
		return gone.NewInnerError(fmt.Sprintf("config %s err: bad type", k), gone.NotCompatible)
	}
	return nil
}

func (c *propertiesConfigure) isInTestKit() bool {
	return c.cemetery != nil && c.cemetery.GetTomById(gone.IdGoneTestKit) != nil
}

func (c *propertiesConfigure) mustGetProperties() (*properties.Properties, error) {
	properties.LogPrintf = c.Warnf

	if c.isInTestKit() {
		return GetTestProperties()
	} else {
		return GetProperties("")
	}
}

func isDuration(t reflect.Type) bool { return t == reflect.TypeOf(time.Second) }
