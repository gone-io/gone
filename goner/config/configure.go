package config

import (
	"fmt"
	"github.com/gone-io/gone"
	"github.com/magiconair/properties"
	"path"
	"reflect"
	"strconv"
	"time"
)

func NewConfigure() (gone.Goner, gone.GonerId, gone.GonerOption, gone.GonerOption) {
	return &configure{}, gone.IdGoneConfigure, gone.IsDefault(new(gone.Configure)), gone.Order0
}

type configure struct {
	gone.Flag
	gone.Logger `gone:"gone-logger"`
	props       *properties.Properties
	cemetery    gone.Cemetery `gone:"gone-cemetery"`
}

func (c *configure) Get(key string, v any, defaultVal string) error {
	if c.props == nil {
		props, err := c.getProperties()
		if err != nil {
			return gone.ToError(err)
		}
		c.props = fixExpand(props)
	}
	return getFromProperties(key, v, defaultVal, c.props)
}

const SliceMaxSize = 100

func fixExpand(props *properties.Properties) *properties.Properties {
	newProperties := properties.NewProperties()

	keys := props.Keys()
	for _, key := range keys {
		value, ok := props.Get(key)
		if ok {
			_, _, _ = newProperties.Set(key, value)
		}
	}
	return newProperties
}

func getFromProperties(
	key string,
	value any,
	defaultVale string,
	props *properties.Properties,
) error {
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return gone.NewInnerError("type of value must be ptr", gone.NotCompatible)
	}

	el := rv.Elem()

	switch el.Kind() {
	default:
		return gone.NewInnerError(fmt.Sprintf("<%s>(%v) is not support", el.Type().Name(), el.Type().Kind()), gone.NotCompatible)

	case reflect.String:
		confVal := props.GetString(key, defaultVale)
		el.SetString(confVal)

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

	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		def, _ := strconv.ParseUint(defaultVale, 10, 64)
		confVal := props.GetUint64(key, def)
		el.SetUint(confVal)

	case reflect.Float64, reflect.Float32:
		def, _ := strconv.ParseFloat(defaultVale, 64)
		confVal := props.GetFloat64(key, def)
		el.SetFloat(confVal)

	case reflect.Struct:
		conf := props.FilterStripPrefix(key + ".") //filtered expand not get corrected value, so need fixExpand
		err := conf.Decode(value)
		return gone.ToError(err)

	case reflect.Slice:
		sliceElementType := el.Type().Elem()

		for i := 0; i < SliceMaxSize; i++ {
			k := fmt.Sprintf("%s[%d].", key, i)
			conf := props.FilterStripPrefix(k) //filtered expand not get corrected value, so need fixExpand
			if conf.Len() == 0 {
				break
			}

			err := decodeSliceElement(sliceElementType, k, conf, el)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func decodeSliceElement(sliceElementType reflect.Type, k string, conf *properties.Properties, el reflect.Value) error {
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

func (c *configure) isInTestKit() bool {
	return c.cemetery != nil && c.cemetery.GetTomById(gone.IdGoneTestKit) != nil
}

const ext = ".properties"

func (c *configure) getProperties() (*properties.Properties, error) {
	configs := GetConfSettings(c.isInTestKit())
	var filePaths []string
	for _, conf := range configs {
		filePaths = append(filePaths, path.Join(conf.ConfigPath, conf.ConfigName+ext))
	}

	properties.LogPrintf = c.Debugf
	props, err := properties.LoadFiles(filePaths, properties.UTF8, true)
	return props, gone.ToError(err)
}

func isDuration(t reflect.Type) bool { return t == reflect.TypeOf(time.Second) }
