package gone_viper

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

var envFlag = flag.String("env", "", "environment")
var confFlag = flag.String("conf", "", "config directory")

// GetEnv get environment, fetch value from command line flag(-env) first, then from environment variable(ENV), then use default value
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

// GetConfDir get config directory, fetch value from command line flag(-conf) first, then from environment variable(CONF)
func GetConfDir() string {
	flag.Parse()
	if *confFlag != "" {
		return *confFlag
	}
	return os.Getenv(EConf)
}

func MustGetExecutableConfDir() string {
	dir, err := os.Executable()
	if err != nil {
		panic(gone.ToError(err))
	}
	return path.Join(filepath.Dir(dir), ConPath)
}

func MustGetWorkDir() string {
	workDir, err := os.Getwd()
	if err != nil {
		panic(gone.ToError(err))
	}
	return workDir
}

// ConfSetting config settings, include config file dir path and config file name(do not include file extension)
type ConfSetting struct {
	ConfigPath string
	ConfigName string
}

func lookForModDir(workDir string) string {
	if workDir == "/" {
		return ""
	}
	modFile := path.Join(workDir, "go.mod")
	if _, err := os.Stat(modFile); os.IsNotExist(err) {
		return lookForModDir(path.Dir(workDir))
	}
	return workDir
}

// GetConfSettings get config settings
func GetConfSettings(isInTestKit bool) (configs []ConfSetting) {
	var configPaths []string

	executableConfDir := MustGetExecutableConfDir()

	configPaths = append(configPaths, executableConfDir)
	workDir := MustGetWorkDir()

	configPaths = append(configPaths, path.Join(MustGetWorkDir(), ConPath))

	if isInTestKit {
		dir := lookForModDir(workDir)
		if dir != "" {
			dir = path.Join(dir, ConPath)
			configPaths = append(configPaths, dir)
		}

		testDataConfig := path.Join(workDir, "testdata", ConPath)
		if testDataConfig != dir {
			configPaths = append(configPaths, testDataConfig)
		}
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
