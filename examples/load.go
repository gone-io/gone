package main

import "github.com/gone-io/gone"

func main() {
	type Dep struct {
		gone.Flag
		Name string
	}

	// 加载一个Dep
	gone.Load(&Dep{})

	//加载 一个名为dep1的Dep
	gone.Load(&Dep{}, gone.Name("dep1"))

	//支持链式调用
	gone.
		Load(&Dep{}, gone.Name("dep3")).
		Load(&Dep{}, gone.Name("dep2"))

	//通过加载函数批量加载
	gone.Loads(func(loader gone.Loader) error {
		err := loader.Load(&Dep{}, gone.Name("dep4"))
		if err != nil {
			return gone.ToError(err)
		}
		err = loader.Load(&Dep{}, gone.Name("dep5"))

		return err
	})
}
