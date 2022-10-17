package config

import (
	"errors"
	"flag"
	"fmt"
	"github.com/gone-io/gone"
	"github.com/magiconair/properties"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"time"
)

func NewConfigure() (gone.Goner, gone.GonerId) {
	return &propertiesConfigure{}, gone.IdGoneConfigure
}

type propertiesConfigure struct {
	gone.GonerFlag
	gone.Logger `gone:"gone-logger"`
	props       *properties.Properties
}

func (c *propertiesConfigure) Get(key string, v interface{}, defaultVal string) error {
	if c.props == nil {
		env := GetEnv("")
		c.Infof("Use Env: %s", env)
		c.props = c.MustGet()
	}
	return c.ParseKeyFromProperties(key, v, defaultVal, c.props)
}

func (c *propertiesConfigure) ParseKeyFromProperties(key string, value interface{}, defaultVale string, props *properties.Properties) error {
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

func isDuration(t reflect.Type) bool { return t == reflect.TypeOf(time.Second) }

func (c *propertiesConfigure) MustGet(envParams ...string) *properties.Properties {
	props, err := Get(envParams...)
	if err != nil {
		panic(err)
	}
	return props
}

// Get 读取环境变量ENV，读取参数 --env
// 读取配置的目录：程序所在目录，程序运行目录
// 配置文件读取顺序：config/default.properties，config/${env}.properties，后面的覆盖前面的
func Get(envParams ...string) (*properties.Properties, error) {
	var env = ""
	if len(envParams) > 0 {
		env = envParams[0]
	}

	env = GetEnv(env)

	var filenames = make([]string, 0)

	executableDir, err := getExecutableDir()
	if err == nil {
		filenames = append(filenames,
			path.Join(executableDir, configPath, defaultConfigFile),
			path.Join(executableDir, configPath, fmt.Sprintf("%s%s", env, fileType)),
		)
	}

	wordDir, err := os.Getwd()
	if err == nil {
		filenames = append(filenames,
			path.Join(wordDir, configPath, defaultConfigFile),
			path.Join(wordDir, configPath, fmt.Sprintf("%s%s", env, fileType)),
		)
	}

	confDir := getConfDir()
	if confDir != "" {
		filenames = append(filenames,
			path.Join(confDir, defaultConfigFile),
			path.Join(confDir, fmt.Sprintf("%s%s", env, fileType)),
		)
	}

	if len(filenames) == 0 {
		return nil, errors.New("cannot read config path")
	}

	props, err := properties.LoadFiles(filenames, properties.UTF8, true)
	if err != nil {
		return nil, err
	}

	err = fixVariableConfig(props)
	if err != nil {
		return nil, err
	}
	return props, nil
}

func fixVariableConfig(props *properties.Properties) error {
	keys := props.Keys()
	for _, k := range keys {
		v, ok := props.Get(k)
		if ok {
			_, _, err := props.Set(k, v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

const configPath = "config"
const fileType = ".properties"
const defaultConfigFile = "default.properties"

const defaultEnv = "local"

var envFlag = flag.String("env", "", "环境变量，默认为local")
var confFlag = flag.String("conf", "", "配置目录")

// GetEnv 获取环境变量
func GetEnv(env string) string {
	if env != "" {
		return env
	}

	flag.Parse()
	if *envFlag != "" {
		return *envFlag
	}

	env = os.Getenv("ENV")
	if env != "" {
		return env
	}
	return defaultEnv
}

func getConfDir() string {
	flag.Parse()
	if *confFlag != "" {
		return *confFlag
	}
	return os.Getenv("CONF")
}

func getExecutableDir() (string, error) {
	dir, err := os.Executable()
	if err != nil {
		return "", err
	}

	return filepath.Dir(dir), nil
}
