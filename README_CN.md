<p align="left">
    <a href="README.md">English</a>&nbsp ｜&nbsp 中文
</p>

[![license](https://img.shields.io/badge/license-GPL%20V3-blue)](LICENSE)
[![GoDoc](https://pkg.go.dev/badge/github.com/gone-io/gone.jsonvalue?utm_source=godoc)](http://godoc.org/github.com/gone-io/gone)
[![Go Report Card](https://goreportcard.com/badge/github.com/gone-io/gone)](https://goreportcard.com/report/github.com/gone-io/gone)
[![codecov](https://codecov.io/gh/gone-io/gone/graph/badge.svg?token=H3CROTTDZ1)](https://codecov.io/gh/gone-io/gone)
[![Build and Test](https://github.com/go-kod/kod/actions/workflows/go.yml/badge.svg)](https://github.com/go-kod/kod/actions/workflows/go.yml)
[![Release](https://img.shields.io/github/release/gone-io/gone.svg?style=flat-square)](https://github.com/gone-io/gone/releases)

<img src="docs/assert/logo.png" width = "200" alt="logo" align=center />

- [Gone](#gone)
	- [🌐Web服务](#web服务)
	- [💡概念](#概念)
	- [🌰 更多例子：](#-更多例子)
	- [🪜🧰🛠️ 组件库（👉🏻 更多组件正在开发中...，💪🏻 ヾ(◍°∇°◍)ﾉﾞ，🖖🏻）](#️-组件库-更多组件正在开发中-ヾﾉﾞ)
	- [📚完整文档](#完整文档)


# Gone


Gone首先是一个轻量的，基于Golang的，依赖注入框架，灵感来源于Java中的Spring Framework；其次，Gone框架中包含了一系列内置组件，通过这些组件提供一整套Web开发方案，提供服务配置、日志追踪、服务调用、数据库访问、消息中间件等微服务常用能力。

[完整文档](https://goner.fun/zh/)

下面使用Gone来编写一个Web服务吧！

## 🌐Web服务
```go
package main

import (
	"fmt"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner"
)

// 实现一个Goner，什么是Goner？ => https://goner.fun/zh/guide/core-concept.html#goner-%E9%80%9D%E8%80%85
type controller struct {
	gone.Flag //goner 标记，匿名嵌入后，一个结构体就实现了Goner
	gone.RouteGroup `gone:"gone-gin-router"` //注入根路由
}

// 实现 Mount 方法，挂载路由；框架会自动执行该方法
func (ctr *controller) Mount() gone.GinMountError {

	// 定义请求结构体
	type Req struct {
		Msg string `json:"msg"`
	}

	//注册 `POST /hello` 的 处理函数
	ctr.POST("/hello", func(in struct {
		to  string `gone:"http,query"` //注入http请求Query参数To
		req *Req   `gone:"http,body"`  //注入http请求Body
	}) any {
		return fmt.Sprintf("to %s msg is: %s", in.to, in.req.Msg)
	})

	return nil
}

func main() {
	//启动服务
	gone.Serve(func(cemetery gone.Cemetery) error {
		// 调用框架内置组件，加载gin框架
		_ = goner.GinPriest(cemetery)

		//将 一个controller类型的Goner埋葬到墓园
		//埋葬是什么意思？ => https://goner.fun/zh/guide/core-concept.html#bury-%E5%9F%8B%E8%91%AC
		//墓园是什么意思？ => https://goner.fun/zh/guide/core-concept.html#cemetery-%E5%A2%93%E5%9B%AD
		cemetery.Bury(&controller{})
		return nil
	})
}
```

运行上面代码：go run main.go，程序将监听8080端口，使用curl测试：
```bash
curl -X POST 'http://localhost:8080/hello' \
    -H 'Content-Type: application/json' \
	--data-raw '{"msg": "你好呀？"}'
```

结果如下：
```
{"code":0,"data":"to  msg is: 你好呀？"}
```
[快速开始](https://goner.fun/zh/quick-start/)


## 💡概念
> 我们编写的代码终究只是死物，除非他们被运行起来。
在Gone中，组件被抽象为Goner（逝者），Goner属性可以注入其他的Goner。Gone启动前，需要将所有 Goners 埋葬（Bury）到墓园（cemetery）；Gone启动后，会将所有 Goners 复活，建立一个 天国（Heaven），“天国的所有人都不再残缺，他们想要的必定得到满足”。

[核心概念](https://goner.fun/zh/guide/core-concept.html)

## 🌰 更多例子：

> 在[example](example)目录可以找到详细的例子，后续会补充完成的帮忙手册。

## 🪜🧰🛠️ 组件库（👉🏻 更多组件正在开发中...，💪🏻 ヾ(◍°∇°◍)ﾉﾞ，🖖🏻）
- [goner/cumx](goner/cmux)，
  对 `github.com/soheilhy/cmux` 的封装，用于复用同一个端口实现多种协议；
- [goner/config](goner/config)，用于实现对 **Gone-App** 配置
- [goner/gin](goner/gin)，对 `github.com/gin-gonic/gin`封装，提供 web 服务
- [goner/logrus](goner/logrus)，
  对 `github.com/sirupsen/logrus`封装，提供日志服务
- [goner/tracer](goner/tracer)，
  提供日志追踪，可以给同一条请求链路提供统一的 `tracerId`
- [goner/xorm](goner/xorm)，
  封装 `xorm.io/xorm`，用于数据库的访问；使用时，按需引用数据库驱动；
- [goner/redis](goner/redis)，
  封装 `github.com/gomodule/redigo`，用于操作 redis
- [goner/schedule](goner/schedule)，
  封装 `github.com/robfig/cron/v3`，用于设置定时器
- [emitter](https://github.com/gone-io/emitter)，封装事件处理，可以用于 **DDD** 的 **事件风暴**
- [goner/urllib](goner/urllib),
  封装了 `github.com/imroc/req/v3`，用于发送http请求，打通了server和client的traceId

## 📚[完整文档](https://goner.fun/zh/)
