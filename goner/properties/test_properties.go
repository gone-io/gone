package properties

import (
	"fmt"
	"github.com/gone-io/gone"
	"github.com/magiconair/properties"
	"os"
	"path"
	"path/filepath"
)

//TestKit 模式会运行下面代码

const test = "_test"
const testExt = test + ext
const defaultTestConfigFile = defaultConf + testExt

func GetTestProperties() (props *properties.Properties, err error) {
	var configDir string
	configDir, err = lookupConfigDir("")
	if err != nil {
		return nil, gone.ToError(err)
	}
	props, err = buildTestProps(configDir)
	return props, gone.ToError(err)
}

func lookupConfigDir(begin string) (configDir string, err error) {
	if begin == "" {
		begin, err = os.Getwd()
		if err != nil {
			return
		}
	}

	_, err = os.Stat(filepath.Join(begin, "go.mod"))
	if err == nil { // 文件存在
		configDir = filepath.Join(begin, "config")
		_, err = os.Stat(configDir)
		return
	}

	if os.IsNotExist(err) { // 文件不存在
		return lookupConfigDir(filepath.Dir(begin))
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
	var err error
	files, err = filterNotExistedFiles(files)
	if err != nil {
		return nil, gone.ToError(err)
	}

	return properties.LoadFiles(files, properties.UTF8, true)
}
