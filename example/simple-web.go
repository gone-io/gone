package main

import (
	"fmt"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner"
)

// 实现一个Goner，什么是Goner？ => https://goner.fun/zh/guide/core-concept.html#goner-%E9%80%9D%E8%80%85
type controller struct {
	gone.Flag                  //goner 标记，匿名嵌入后，一个结构体就实现了Goner
	gone.RouteGroup `gone:"*"` //注入根路由
}

// Mount 用于挂载路由；框架会自动执行该方法
func (ctr *controller) Mount() gone.GinMountError {
	// 定义请求结构体
	type Req struct {
		Msg string `json:"msg"`
	}

	//注册 `POST /hello` 的 处理函数
	ctr.
		POST("/hello", func(in struct {
			to  string `gone:"http,query"` //注入http请求Query参数To
			req *Req   `gone:"http,body"`  //注入http请求Body
		}) any {
			return fmt.Sprintf("to %s msg is: %s", in.to, in.req.Msg)
		}).
		GET("/hello", func() string {
			return "hello"
		})

	return nil
}

func main() {
	gone.
		Default.
		LoadPriest(goner.GinPriest). //加载 `goner/gin` 用于提供web服务
		Load(&controller{}).         //加载我们前面定义的controller
		Serve()                      // 启动服务
}
