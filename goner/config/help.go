package config

import (
	"flag"
	"github.com/gone-io/gone"
	"os"
	"path"
	"path/filepath"
)

const defaultEnv = "local"
const EEnv = "ENV"
const EConf = "CONF"
const ConPath = "config"

const TestSuffix = "_test"

const DefaultConf = "default"

var envFlag = flag.String("env", "", "environment，默认为local")
var confFlag = flag.String("conf", "", "config directory")

// GetEnv get environment
func GetEnv() string {
	flag.Parse()
	if *envFlag != "" {
		return *envFlag
	}

	env := os.Getenv(EEnv)
	if env != "" {
		return env
	}
	return defaultEnv
}

func GetConfDir() string {
	flag.Parse()
	if *confFlag != "" {
		return *confFlag
	}
	return os.Getenv(EConf)
}

func GetExecutableConfDir() (string, error) {
	dir, err := os.Executable()
	if err != nil {
		return "", gone.ToError(err)
	}

	return path.Join(filepath.Dir(dir), ConPath), nil
}

type ConfSetting struct {
	ConfigPath string
	ConfigName string
}

func GetConfSettings(isInTestKit bool) (configs []ConfSetting, err error) {
	var configPaths []string

	executableConfDir, err := GetExecutableConfDir()
	if err != nil {
		return
	}

	configPaths = append(configPaths, executableConfDir)
	wordDir, err := os.Getwd()
	if err != nil {
		return nil, gone.ToError(err)
	}
	configPaths = append(configPaths, path.Join(wordDir, ConPath))

	if isInTestKit {
		configPaths = append(configPaths, path.Join(wordDir, "testdata", ConPath))
	}

	settingConfPath := GetConfDir()
	if settingConfPath != "" {
		configPaths = append(configPaths, settingConfPath)
	}

	envConf := GetEnv()

	for _, configPath := range configPaths {
		configs = append(configs,
			ConfSetting{ConfigPath: configPath, ConfigName: DefaultConf},
		)

		if isInTestKit {
			configs = append(configs,
				ConfSetting{ConfigPath: configPath, ConfigName: DefaultConf + TestSuffix},
			)
		}

		configs = append(configs,
			ConfSetting{ConfigPath: configPath, ConfigName: envConf},
		)

		if isInTestKit {
			configs = append(configs,
				ConfSetting{ConfigPath: configPath, ConfigName: envConf + TestSuffix},
			)
		}
	}
	return
}
