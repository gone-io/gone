package main

import (
	"fmt"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/config"
	"time"
)

type SubConf struct {
	X string `properties:"x"`
	Y string `properties:"y"`
}

type UseConfig struct {
	gone.Flag

	int      int           `gone:"config,my.conf.int"`
	int8     int8          `gone:"config,my.conf.int8"`
	printInt *int          `gone:"config,my.conf.int8"` //指针 指向int
	float64  float64       `gone:"config,my.conf.float64"`
	string   string        `gone:"config,my.conf.string"`
	bool     bool          `gone:"config,my.conf.bool"`
	duration time.Duration `gone:"config,my.conf.duration"`
	defaultV string        `gone:"config,my.conf.default,default=ok"`

	sub *SubConf `gone:"config,my.conf.sub"` //指针，指向结构体

	subs []SubConf `gone:"config,my.conf.subs"` //数组

	ints []int `gone:"config,my.conf.ints"` //基础类型数组，暂时不支持
}

func (g *UseConfig) AfterRevive() gone.AfterReviveError {
	fmt.Printf("int=%d\n", g.int)
	fmt.Printf("int8=%d\n", g.int8)
	fmt.Printf("printInt=%d\n", *g.printInt)
	fmt.Printf("float64=%f\n", g.float64)
	fmt.Printf("string=%s\n", g.string)
	fmt.Printf("bool=%t\n", g.bool)
	fmt.Printf("duration=%v\n", g.duration)
	fmt.Printf("defaultV=%s\n", g.defaultV)
	fmt.Printf("sub.x=%v\n", g.sub)
	fmt.Printf("subs=%v\n", g.subs)
	fmt.Printf("ints=%v\n", g.ints)

	return nil
}

func main() {
	gone.Run(func(cemetery gone.Cemetery) error {
		_ = config.Priest(cemetery)
		cemetery.Bury(&UseConfig{})
		return nil
	})
}
