package main

import "github.com/gone-io/gone"

type Worker struct {
	gone.Flag
	Name string
}

type Boss struct {
	gone.Flag

	Name string
}

func main() {
	gone.
		Prepare(func(cemetery gone.Cemetery) error {
			cemetery.Bury(&Boss{Name: "Jim"}, gone.GonerId("boss-jim"))
			cemetery.Bury(&Worker{Name: "Bob"}, gone.GonerId("worker-bob"))
			return nil
		}).
		BeforeStart(func() {
			println("第1个 BeforeStart 函数")

		}).
		BeforeStart(func(in struct {
			worker Worker `gone:"worker-bob"`
			boss   Boss   `gone:"*"`
		}) {
			println("第2个 BeforeStart 函数")
			println("boss:", in.boss.Name)
			println("worker:", in.worker.Name)
		}).
		BeforeStart(func() error {
			println("第3个 BeforeStart 函数")

			return nil
		}).
		Run()
}
