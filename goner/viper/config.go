package gone_viper

import (
	"errors"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/internal/json"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var load = gone.OnceLoad(func(loader gone.Loader) error {
	return loader.Load(
		&configure{},
		gone.Name(gone.ConfigureName),
		gone.IsDefault(new(gone.Configure)),
		gone.ForceReplace(),
	)
})

func Load(loader gone.Loader) error {
	return load(loader)
}

// Priest Deprecated, use Load instead
func Priest(loader gone.Loader) error {
	return Load(loader)
}

type configure struct {
	gone.Flag
	test []gone.TestFlag `gone:"*"`
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
	return len(c.test) > 0
}

func (c *configure) readConfig() (err error) {
	configs := GetConfSettings(c.isInTestKit())
	conf := viper.New()
	conf.SetEnvPrefix("gone")
	conf.AutomaticEnv()
	conf.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

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
		return gone.NewInnerError("type of value must be ptr", gone.NotSupport)
	}

	el := rv.Elem()
	v := c.conf.Get(key)
	if v == nil {
		return getDefault(el, defaultVale)
	}
	return getConf(key, el, c.conf)
}

func getNameAndDefault(field reflect.StructField, tagName string) (name string, defaultValue string) {
	tag := field.Tag.Get(tagName)
	if tag == "" {
		tag = field.Name
	}
	specs := strings.Split(tag, ",")
	name = specs[0]

	if len(specs) > 1 {
		for _, s := range specs {
			split := strings.SplitAfterN(s, "=", 2)
			if len(split) == 2 && split[0] == "default=" {
				defaultValue = split[1]
				break
			}
		}
	}
	return
}

func DefaultValueDecoderConfig(decoderConfig *mapstructure.DecoderConfig) {
	decoderConfig.DecodeHook = mapstructure.ComposeDecodeHookFunc(
		func(from reflect.Type, to reflect.Type, data any) (any, error) {
			if from.Kind() == reflect.Map && to.Kind() == reflect.Struct {
				dataMap, ok := data.(map[string]any)
				if !ok {
					return data, nil
				}
				var setDefaultValue func(to reflect.Type, v map[string]any)

				setDefaultValue = func(to reflect.Type, v map[string]any) {
					fieldNum := to.NumField()
					for i := 0; i < fieldNum; i++ {
						field := to.Field(i)
						if !field.IsExported() {
							continue
						}
						name, defaultValue := getNameAndDefault(field, decoderConfig.TagName)
						_, has := v[name]

						if !has {
							if field.Type.Kind() == reflect.Struct {
								m := map[string]any{}
								setDefaultValue(field.Type, m)
								v[name] = m
							} else if defaultValue != "" {
								v[name] = defaultValue
							}
						}
					}
				}
				setDefaultValue(to, dataMap)
			}
			return data, nil
		},
		decoderConfig.DecodeHook,
	)
}

func getConf(key string, v reflect.Value, vConf *viper.Viper) error {
	switch v.Kind() {
	default:
		err := vConf.UnmarshalKey(key, v.Addr().Interface(),
			DefaultValueDecoderConfig,
		)
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
