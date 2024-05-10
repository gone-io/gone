package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner"
	"github.com/gone-io/gone/goner/xorm"
)

type Demo struct {
	Id   int64
	Data string
}

type SqlExecutor struct {
	gone.Flag
	db xorm.Engine `gone:"gone-xorm"`
}

func (e *SqlExecutor) Execute() {
	err := e.db.Sync(new(Demo))
	if err != nil {
		println(err.Error())
		return
	}

	demo := Demo{Data: "hello gone"}

	_, err = e.db.Insert(&demo, Demo{Data: "The most Spring programmer-friendly Golang framework, dependency injection, integrates Web. "})

	if err != nil {
		println(err.Error())
		return
	}

	fmt.Printf("%v", demo)

	var list []Demo
	err = e.db.Find(&list)
	if err != nil {
		println(err.Error())
		return
	}

	fmt.Printf("demo records:%v", list)

}

func main() {
	gone.
		Prepare(func(cemetery gone.Cemetery) error {
			_ = goner.XormPriest(cemetery)
			cemetery.Bury(&SqlExecutor{})
			return nil
		}).
		AfterStart(func(in struct {
			e SqlExecutor `gone:"*"`
		}) {
			in.e.Execute()
		}).
		Run()
}
