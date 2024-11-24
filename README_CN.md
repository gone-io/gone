<p align="left">
    <a href="README.md">English</a>&nbsp ｜&nbsp 中文
</p>

[![license](https://img.shields.io/badge/license-GPL%20V3-blue)](LICENSE)
[![GoDoc](https://pkg.go.dev/badge/github.com/gone-io/gone.jsonvalue?utm_source=godoc)](http://godoc.org/github.com/gone-io/gone)
[![Go Report Card](https://goreportcard.com/badge/github.com/gone-io/gone)](https://goreportcard.com/report/github.com/gone-io/gone)
[![codecov](https://codecov.io/gh/gone-io/gone/graph/badge.svg?token=H3CROTTDZ1)](https://codecov.io/gh/gone-io/gone)
[![Build and Test](https://github.com/gone-io/gone/actions/workflows/go.yml/badge.svg)](https://github.com/gone-io/gone/actions/workflows/go.yml)
[![Release](https://img.shields.io/github/release/gone-io/gone.svg?style=flat-square)](https://github.com/gone-io/gone/releases)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)  

<img src="docs/assert/logo.png" width = "100" alt="logo" />

- [Gone](#gone)
	- [Gone 是什么？](#gone-是什么)
	- [特性](#特性)
	- [依赖注入与启动](#依赖注入与启动)
	- [🌐Web服务](#web服务)
	- [💡概念](#概念)
	- [🌰 更多例子：](#-更多例子)
	- [🪜🧰🛠️ 组件库（👉🏻 更多组件正在开发中...，💪🏻 ヾ(◍°∇°◍)ﾉﾞ，🖖🏻）](#️-组件库-更多组件正在开发中-ヾﾉﾞ)
	- [📚完整文档](#完整文档)
	- [贡献](#贡献)
	- [联系方式](#联系方式)
	- [许可证](#许可证)

# Gone
## Gone 是什么？

Gone 是一个轻量级的golang依赖注入框架，并且适配了一些列第三方组件用于快速开始编写一个云原生的微服务。

## 特性
- 依赖注入，支持对结构体属性和函数参数自动注入
- **[Gonectr](https://github.com/gone-io/gonectr)**，生成项目、生成辅助代码、编译和启动项目
- 单元测试方案，基于接口的mock测试
- 多种组件，可插拔，支持云原生、微服务

<img src="docs/assert/architecture.png" width = "600" alt="architecture"/>

## 快速开始
1. 安装 [gonectr](https://github.com/gone-io/gonectr) 和 [mockgen](https://github.com/uber-go/mock/tree/main)
    ```bash
    go install github.com/gone-io/gonectr@latest
    go install go.uber.org/mock/mockgen@latest
    ```
2. 创建一个项目
    ```bash
    gonectr create myproject
    ```
3. 运行项目
    ```bash
    cd myproject
    gonectr run ./cmd/server
    ```
   或者，使用make命令运行，如果你已经安装[make](https://www.gnu.org/software/make/):
    ```bash
    cd myproject
    make run
    ```
   或者使用docker compose来运行:
    ```bash
    cd myproject
    docker compose build
    docker compose up
    ```

## 依赖注入与启动
看一个例子：
```go
package main

import (
	"fmt"
	"github.com/gone-io/gone"
)

type Worker struct {
	gone.Flag //匿名嵌入了 gone.Flag的结构体就是一个 Goner，可以被作为依赖注入到其他Goner，或者接收其他 Goner 的注入
	Name      string
}

func (w *Worker) Work() {
	fmt.Printf("I am %s, and I am working\n", w.Name)
}

type Manager struct {
	gone.Flag                         //匿名嵌入了 gone.Flag的结构体就是一个 Goner，可以被作为依赖注入到其他Goner，或者接收其他 Goner 的注入
	*Worker   `gone:"manager-worker"` //具名注入 GonerId="manager-worker" 的 Worker 实例
	workers   []*Worker               `gone:"*"` //将所有Worker注入到一个数组
}

func (m *Manager) Manage() {
	fmt.Printf("I am %s, and I am managing\n", m.Name)
	for _, worker := range m.workers {
		worker.Work()
	}
}

func main() {
	managerRole := &Manager{}

	managerWorker := &Worker{Name: "Scott"}
	ordinaryWorker1 := &Worker{Name: "Alice"}
	ordinaryWorker2 := &Worker{Name: "Bob"}

	gone.
		Prepare(func(cemetery gone.Cemetery) error {
			cemetery.
				Bury(managerRole).
				Bury(managerWorker, gone.GonerId("manager-worker")).
				Bury(ordinaryWorker1).
				Bury(ordinaryWorker2)
			return nil
		}).
		//Run方法中的函数支持参数的依赖注入
		Run(func(manager *Manager) {
			manager.Manage()
		})
}
```
总结：
1. 在Gone框架中，依赖被抽象为 Goner，Goner 之间可以互相注入
2. 在结构体中匿名嵌入 gone.Flag，结构体就实现了 Goner接口
3. 在启动前，将所有 Goners 通过 Bury函数加载到框架中
4. 使用Run方法启动，Run方法中的函数支持参数的依赖注入

[完整文档](https://goner.fun/zh/)


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

## 更新记录
### v1.2.1
- 定义 **gone.Provider**，一个工厂函数用于将 不是 **Goner** 的外部组件（结构体、结构体指针、函数、接口）注入到 属性需要注入的Goner；
- 修复 `gone.NewProviderPriest` 无法为 生成接口类型的**gone.Provider**生成Priest; 
- 为`goner/gorm`编写测试代码，补齐其他测试代码；文档更新。

### v1.2.0
- 提供一种新的 `gone.GonerOption`，可以将按类型注入，将构造注入类型实例的任务代理给一个实现了`Suck(conf string, v reflect.Value, field reflect.StructField) error`的**Goner**；
- 提供了一个用于实现**Goner Provider**的辅助函数：`func NewProviderPriest[T any, P any](fn func(tagConf string, param P) (T, error)) Priest` ；
- 给`goner/xorm` 集群模式提供策略配置的方案；
- 完善`goner/gorm`代码 和 做功能测试，支持多种数据库的接入。

### v1.1.1
- goner/xorm 支持集群 和 多数据库，最新文档：https://goner.fun/zh/references/xorm.html
- 新增 goner/gorm，封装`gorm.io/gorm`，用于数据库的访问，暂时只支持mysql，完善中...

## 贡献
如果您发现了错误或有功能请求，可以随时[提交问题](https://github.com/gone-io/gone/issues/new)，同时欢迎[提交拉取请求](https://github.com/gone-io/gone/pulls)。

## 联系方式
如果您有任何问题，欢迎通过以下方式联系我们：
- [Github 讨论](https://github.com/gone-io/gone/discussions)
- 扫码加微信，暗号：gone

  <img src="docs/assert/qr_dapeng.png" width = "250" alt="dapeng wx qr code"/>

## 许可证
`gone` 在 MIT 许可证下发布，详情请参阅 [LICENSE](./LICENSE) 文件。