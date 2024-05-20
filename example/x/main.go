package main

import (
	"fmt"
	"github.com/gone-io/gone"
)

type Worker struct {
	gone.Flag // Goner标志，匿名内嵌`gone.Flag`表示该结构体为一个Goner
}

func (w *Worker) Do() {
	fmt.Println("worker do")
}

type Boss struct {
	gone.Flag // Goner标志，匿名内嵌`gone.Flag`表示该结构体为一个Goner

	seller *Worker `gone:"*"` //注入Worker
}

func (b *Boss) Do() {
	fmt.Println("boss do")
	b.seller.Do()
}

func main() {
	gone.
		Prepare(func(cemetery gone.Cemetery) error {
			cemetery.
				Bury(&Boss{}).
				Bury(&Worker{})
			return nil
		}).
		//AfterStart 是一个hook函数，关于hook函数请参考文档：https://goner.fun/zh/guide/hooks.html
		AfterStart(func(in struct {
			boss *Boss `gone:"*"` //注入Boss
		}) {
			in.boss.Do()
		}).
		Run()
}
