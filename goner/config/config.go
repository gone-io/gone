package config

import (
	"github.com/gone-io/gone"
	"reflect"
	"strings"
)

type config struct {
	gone.Flag
	configure Configure `gone:"gone-configure"`
}

func parseConfAnnotation(tag string) (key string, defaultVal string) {
	splitArray := strings.Split(tag, ",")
	key = strings.TrimSpace(splitArray[0])
	if len(splitArray) > 1 {
		for i := 1; i < len(splitArray); i++ {
			s := splitArray[i]
			arr := strings.Split(s, "=")
			if strings.TrimSpace(arr[0]) == "default" {
				if len(arr) > 1 {
					defaultVal = arr[1]
				}
				break
			}
		}
	}
	return key, defaultVal
}

func (f *config) Suck(conf string, v reflect.Value) gone.SuckError {
	key, defaultVal := parseConfAnnotation(conf)

	if reflect.Ptr == v.Kind() {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return f.configure.Get(key, v.Addr().Interface(), defaultVal)
}

// Configure 配置接口
type Configure interface {
	//GetProperties 将获取`key`所对应的值，值将写入到参数`v`中；参数`v`，只接受指针类型；如果`key`对应的值不存在，将使用defaultVal
	Get(key string, v interface{}, defaultVal string) error
}
