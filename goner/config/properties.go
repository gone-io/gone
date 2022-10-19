package config

import (
	"errors"
	"flag"
	"fmt"
	"github.com/magiconair/properties"
	"os"
	"path"
	"path/filepath"
)

const configPath = "config"
const ext = ".properties"
const defaultConf = "default"
const defaultEnv = "local"
const defaultConfigFile = defaultConf + ext

// GetProperties 读取环境变量ENV，读取参数 --env
// 读取配置的目录：程序所在目录，程序运行目录
// 配置文件读取顺序：config/default.properties，config/${env}.properties，后面的覆盖前面的
func GetProperties(envParams ...string) (*properties.Properties, error) {
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
			path.Join(executableDir, configPath, fmt.Sprintf("%s%s", env, ext)),
		)
	}

	wordDir, err := os.Getwd()
	if err == nil {
		filenames = append(filenames,
			path.Join(wordDir, configPath, defaultConfigFile),
			path.Join(wordDir, configPath, fmt.Sprintf("%s%s", env, ext)),
		)
	}

	confDir := getConfDir()
	if confDir != "" {
		filenames = append(filenames,
			path.Join(confDir, defaultConfigFile),
			path.Join(confDir, fmt.Sprintf("%s%s", env, ext)),
		)
	}

	if len(filenames) == 0 {
		return nil, errors.New("cannot read config path")
	}

	props, err := properties.LoadFiles(filenames, properties.UTF8, false)
	if err != nil {
		return nil, err
	}

	return props, fixVariableConfig(props)
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
