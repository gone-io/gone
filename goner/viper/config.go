package gone_viper

import (
	"encoding/json"
	"errors"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/config"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"reflect"
	"strconv"
	"time"
)

func NewConfigure() (gone.Goner, gone.GonerId, gone.GonerOption) {
	return &configure{}, gone.IdGoneConfigure, gone.IsDefault(true)
}

type configure struct {
	gone.Flag
	cemetery gone.Cemetery `gone:"*"`

	conf *viper.Viper
}

func (c *configure) Get(key string, v any, defaultVal string) error {
	if c.conf == nil {
		err := c.readConfig()
		if err != nil {
			return err
		}
	}

	return c.get(key, v, defaultVal)
}

func (c *configure) isInTestKit() bool {
	return c.cemetery != nil && c.cemetery.GetTomById(gone.IdGoneTestKit) != nil
}

func (c *configure) readConfig() (err error) {
	configs := config.GetConfSettings(c.isInTestKit())

	conf := viper.New()
	for _, setting := range configs {
		vConf := viper.New()
		vConf.SetConfigName(setting.ConfigName)
		vConf.AddConfigPath(setting.ConfigPath)
		err := vConf.ReadInConfig()
		if err != nil {
			var configFileNotFoundError viper.ConfigFileNotFoundError
			ok := errors.As(err, &configFileNotFoundError)
			if !ok {
				return gone.ToError(err)
			}
			continue
		}

		err = conf.MergeConfigMap(vConf.AllSettings())
		if err != nil {
			return gone.ToError(err)
		}
	}

	c.conf = conf
	return
}

func (c *configure) get(key string, value any, defaultVale string) error {
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return gone.NewInnerError("type of value must be ptr", gone.NotCompatible)
	}

	el := rv.Elem()
	v := c.conf.Get(key)
	if v == nil {
		return getDefault(el, defaultVale)
	}
	return getConf(key, el, c.conf)
}

func getConf(key string, v reflect.Value, vConf *viper.Viper) error {
	switch v.Kind() {
	default:
		err := vConf.UnmarshalKey(key, v.Addr().Interface())
		if err != nil {
			return gone.ToError(err)
		}
	case reflect.String:
		v.SetString(vConf.GetString(key))

	case reflect.Bool:
		v.SetBool(vConf.GetBool(key))

	case reflect.Int, reflect.Int32, reflect.Int16, reflect.Int8:
		v.SetInt(vConf.GetInt64(key))

	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		v.SetUint(vConf.GetUint64(key))

	case reflect.Float64, reflect.Float32:
		v.SetFloat(vConf.GetFloat64(key))

	case reflect.Int64:
		if v.Type() == reflect.TypeOf(time.Second) {
			v.Set(reflect.ValueOf(vConf.GetDuration(key)))
		} else {
			v.SetInt(vConf.GetInt64(key))
		}
	}
	return nil
}

func getDefault(v reflect.Value, defaultVale string) error {
	switch v.Kind() {
	default:
		var defaultValeMap any
		err := json.Unmarshal([]byte(defaultVale), &defaultValeMap)
		if err != nil {
			return gone.ToError(err)
		}
		err = mapstructure.Decode(defaultValeMap, v.Addr().Interface())
		if err != nil {
			return gone.ToError(err)
		}
	case reflect.String:
		v.SetString(defaultVale)
	case reflect.Bool:
		def, _ := strconv.ParseBool(defaultVale)
		v.SetBool(def)

	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		def, _ := strconv.ParseUint(defaultVale, 10, 64)
		v.SetUint(def)

	case reflect.Int, reflect.Int32, reflect.Int16, reflect.Int8:
		def, _ := strconv.ParseInt(defaultVale, 10, 32)
		v.SetInt(def)

	case reflect.Float64, reflect.Float32:
		def, _ := strconv.ParseFloat(defaultVale, 64)
		v.SetFloat(def)

	case reflect.Int64:
		if v.Type() == reflect.TypeOf(time.Second) {
			duration, err := time.ParseDuration(defaultVale)
			if err != nil {
				return gone.ToError(err)
			}
			v.Set(reflect.ValueOf(duration))
		} else {
			def, _ := strconv.ParseInt(defaultVale, 10, 64)
			v.SetInt(def)
		}
	}
	return nil
}
