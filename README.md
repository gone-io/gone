# gone

这是gone框架的第二版，第一版在[这里](https://gitlab.openviewtech.com/gone/gone#gone)

## 这是个啥？

这是一个依赖注入框架，应该是"最类似spring"的一个golang的依赖注入框架。可以将`Goner`理解为`Spring Bean`
，代码中只需要编写各种功能的`Goner`即可完成业务开发。在[example](example)目录可以找到详细的例子，后续会补充完成的帮忙手册。

## 概念

> gone的意思是 `走了，去了，没了，死了`，那么Gone管理都是Goner(逝者)  
> 存在一片神秘墓园，安葬在这里的逝者，灵魂会升入天国。天国指定的牧师可以将Goner葬入这片墓园...

- Heaven: 天国 🕊☁️
- Heaven.Start: 天国开始运行；Goner永生，直到天崩地裂
- Heaven.Stop:  天国崩塌，停止运行
- Cemetery: 墓园 🪦
- Cemetery.Bury:  安葬
- Cemetery.revive: 复活Goner，将其升入天国；对于Goner则是完成了属性的的注入（或者装配）
- Tomb: 坟墓 ⚰️
- Priest: 神父✝️，负责给Goner下葬
- Goner: 逝者 💀；是对可注入对象的抽象：可以注入其他Goner，可以被注入其他Goner；
- Prophet: 先知；如果一个Goner是先知，他被复活后会去执行`AfterRevive() AfterReviveError`方法，去窥视神的旨意
- Prophet.AfterRevive: 复活后执行的方法
- Angel: 天使 𓆩♡𓆪 ，实现了`Start(gone.Cemetery) error` 和 `Stop(gone.Cemetery) error`方法的Goner，升入天国后被变成天使
- Angel.Start: 天使左翼，开始工作；能力越大责任越大，天使是要工作的
- Angel.Stop: 天使右翼，停止工作；
- Vampire: 吸血鬼 🧛🏻‍，实现了`Suck(conf string, v reflect.Value) gone.SuckError`
  方法的是吸血鬼；吸血鬼是一个邪恶的存在，他可能毁掉整个天国。理论上吸血行为可以制造Goner，但是这可能会导致循环依赖，从而破坏系统。
- Vampire.Suck: 吸血鬼"吸血行为"

### 四种Goner

- 普通Goner
  > 普通Goner，可以用于抽象App中的Service、Controller、Client等常见的组件。
  > 如果Goner提供了方法 **`AfterRevive(Cemetery, Tomb) ReviveAfterError`**，在升入天国后会被调用。
- 先知Prophet
  > 先知，复活后会去执行`AfterRevive() AfterReviveError`方法
- 天使Angel
  > 天使会在天国承担一定的职责：启动阶段，天使的`Start`方法会被调用；停止阶段，天使的`Stop`方法会被调用；所以天使适合抽象"
  需要启停控制"的组件。
- 吸血鬼Vampire
  > 吸血鬼，具有吸血的能力，可以通过`Suck`方法去读取/写入被标记的字段；可以抽象需要控制其他组件某个属性的行为。

## 注入配置

## 普通Goner下葬

```go
package goner_demo

import "github.com/gone-io/gone"

type XGoner struct {
	gone.GonerFlag
}

type Demo struct {
	gone.GonerFlag
	a  *XGoner     `gone:"x-goner"` // x-goner 是 GonerId; 支持使用非导出属性
	A  XGoner      `gone:"x-goner"` // x-goner 是 GonerId; 支持结构体；⚠️尽量不要这样使用，由于结构体是值拷贝，会导致不能深度复制的问题
	A1 *XGoner     `gone:"x-goner"` // x-goner 是 GonerId; 支持结构体的指针
	A2 interface{} `gone:"x-goner"` // x-goner 是 GonerId; 支持接口

	B  *XGoner       `gone:"*"` //  支持匿名注入
	B1 []interface{} `gone:"*"` // 支持匿名注入数组
}
```

## 对吸血鬼下葬，被下葬的是Goner是一个Vampire

> 吸血鬼是一种邪恶的生物，他可以读取/吸入被注入的Goner的属性

```go
package goner_demo

import (
	"github.com/gone-io/gone"
	"github.com/magiconair/properties/assert"
	"reflect"
)

type ConfigVampire struct {
	gone.GonerFlag
}

func (*ConfigVampire) Suck(conf string, v reflect.Value) gone.SuckError {
	// conf = abc.dex,xxx|xxx
	// v = Demo.a 的 reflect.Value

	return nil
}

const ConfigVampireId = "x-config"

type Demo struct {
	// 吸血鬼不会被注入到属性中，而是会在属性上调用`Vampire.Suck`函数完成吸血，吸血鬼可以读取、写入属性的值
	a int `gone:"x-config,abc.dex,xxx|xxx"` //普通Goner会忽略GonerId(x-config)后面的字符串`abc.dex,xxx|xxx`; 而吸血鬼会用来进行"吸血"
}

func Priest(cemetery gone.Cemetery) error {
	cemetery.Bury(&ConfigVampire{}, ConfigVampireId)
	cemetery.Bury(&Demo{})
	return nil
}

func run() {
	gone.Run(Priest)
}
```

## 使用

- 启动

```go
package main

import "github.com/gone-io/gone"

func main() {
	gone.Run(func(cemetery gone.Cemetery) error {
		//下葬Goner
		return nil
	})
}

```

## 代码生成(生成`Priest`函数)

> 在gone框架中提供了一个同名的代码生成工具，他的作用是 扫描文件目录中标记了 `//go:gone`
> 的函数，为这些函数生成一个 `Priest`函数；

- 安装gone
    ```shell 
    go install github.com/gone-io/gone/tools/gone@v0.0.3
    ```
- 使用，参考 [example/app/Makefile](example/app/Makefile)
    ```shell
    gone -s ${scan_package_dir} -p ${pkgName} -f ${funcName} -o ${output_dir} [-w] --stat
    ```
- Demo
    ```shell
    # 进入本仓库的例子目录 
    cd example/app
    
    # 安装gone
    go install github.com/gone-io/gone/tools/gone@v0.0.4
    
    # 生成 priest.go 文件
    gone -s internal -p internal -f Priest -o internal/priest.go
    ```
  将生成文件`internal/priest.go`，内容如下：
    ```go
    // Code generated by gone; DO NOT EDIT.
    package internal
    import (
        "github.com/gone-io/gone/example/app/internal/worker"
        "github.com/gone-io/gone"
    )
    
    func Priest(cemetery gone.Cemetery) error {
        cemetery.Bury(worker.NewPrintWorker())
        worker.Priest(cemetery)
        return nil
    }
    ```

## 组件库

- `github.com/gone-io/gone/goner/cumx`  
  对 `github.com/soheilhy/cmux` 进行封装，用于复用同一个端口实现多种协议；
- `github.com/gone-io/gone/goner/config`  
  完成 gone-app 的配置
- `github.com/gone-io/gone/goner/gin`  
  对`github.com/gin-gonic/gin`封装，提供web服务
- `github.com/gone-io/gone/goner/logrus`  
  对`github.com/sirupsen/logrus`封装，提供日志服务
- `github.com/gone-io/gone/goner/tracer`  
  提供日志追踪，可以用于给同一条请求链路提供统一的tracerId
- `github.com/gone-io/gone/goner/xorm`  
  封装`xorm.io/xorm`，用于数据库的访问；使用时，按需引用数据库驱动；
- `github.com/gone-io/gone/goner/redis`
  封装`github.com/gomodule/redigo`，用于操作redis
- `github.com/gone-io/gone/goner/schedule`
  封装 `github.com/robfig/cron/v3`，用于设置定时器

## TODO LIST

- emitter，封装事件处理
- grpc，封装 github.com/grpc/grpc

## 📢注意

- 尽量不用使用 struct（结构体）作为 `gone` 标记的字段，由于struct在golang中是值拷贝，可能导致相关依赖注入失败的情况
- 下面这些Goner上的方法都不应该是阻塞的
    - `AfterRevive(Cemetery, Tomb) ReviveAfterError`
    - `Start(Cemetery) error`
    - `Stop(Cemetery) error`
    - `Suck(conf string, v reflect.Value) SuckError`