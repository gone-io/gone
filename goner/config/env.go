package config

import (
	"flag"
	"os"
)

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
