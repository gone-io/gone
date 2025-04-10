package main

import "github.com/gone-io/gone/v2"

type Dependence struct {
	gone.Flag //嵌入gone.Flag，实现gone.Goner接口
	Name      string
}

type UseDependence struct {
	gone.Flag              //嵌入gone.Flag，实现gone.Goner接口
	Dependence *Dependence `gone:""` //使用gone标签标记，表示该属性需要注入
}

func main() {
	gone.
		NewApp().
		Load(&UseDependence{}).
		Load(&Dependence{Name: "i am a dependence."}).

		// run 函数支持参数按类型自动注入
		Run(func(
			u *UseDependence, //按类型注入 *UseDependence
			//d *Dependence,
		) {
			println(u.Dependence.Name)
		})
}
