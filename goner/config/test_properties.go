package config

import (
	"fmt"
	"github.com/magiconair/properties"
	"os"
	"path"
)

//TestKit 模式会运行下面代码

const test = "_test"
const testExt = test + ext
const defaultTestConfigFile = defaultConf + testExt

func GetTestProperties() (props *properties.Properties, err error) {
	var configDir string
	configDir, err = lookupConfigDir("")
	props, err = buildTestProps(configDir)

	if err != nil {
		return
	}
	return props, fixVariableConfig(props)
}

func lookupConfigDir(begin string) (configDir string, err error) {
	if begin == "" {
		begin, err = os.Getwd()
		if err != nil {
			return
		}
	}

	_, err = os.Stat(path.Join(begin, "go.mod"))
	if err == nil { // 文件存在
		configDir = path.Join(begin, "config")
		_, err = os.Stat(configDir)
		return
	}

	if os.IsNotExist(err) { // 文件不存在
		return lookupConfigDir(path.Dir(begin))
	}
	//出错
	return
}

func buildTestProps(configDir string) (*properties.Properties, error) {
	env := GetEnv("")
	files := []string{
		path.Join(configDir, defaultConfigFile),
		path.Join(configDir, defaultTestConfigFile),
		path.Join(configDir, fmt.Sprintf("%s%s", env, ext)),
		path.Join(configDir, fmt.Sprintf("%s%s", env, testExt)),
	}

	configDir, _ = os.Getwd()
	configDir = path.Join(configDir, "testdata", configPath)
	files = append(files,
		path.Join(configDir, defaultConfigFile),
		path.Join(configDir, defaultTestConfigFile),
		path.Join(configDir, fmt.Sprintf("%s%s", env, ext)),
		path.Join(configDir, fmt.Sprintf("%s%s", env, testExt)),
	)

	return properties.LoadFiles(files, properties.UTF8, true)
}
