package main

import "github.com/gone-io/gone"

type Worker struct {
	gone.Flag // Goner标志，匿名内嵌`gone.Flag`表示该结构体为一个Goner
	Name      string
}

type Boss struct {
	gone.Flag // Goner标志，匿名内嵌`gone.Flag`表示该结构体为一个Goner

	seller *Worker `gone:"*"` //匿名注入，如果有存在多个Worker，则注入其中一个，通常是第一个
}

func main() {
	gone.Run(func(cemetery gone.Cemetery) error {
		cemetery.Bury(&Boss{})
		cemetery.Bury(&Worker{Name: "小王"})
		cemetery.Bury(&Worker{Name: "小张"})
		return nil
	})
}
