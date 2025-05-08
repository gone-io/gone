<p>
    <a href="README.md">English</a>&nbsp ｜&nbsp 中文
</p>

[![license](https://img.shields.io/badge/license-MIT-blue)](LICENSE)
[![GoDoc](https://pkg.go.dev/badge/github.com/gone-io/gone.jsonvalue?utm_source=godoc)](https://pkg.go.dev/github.com/gone-io/gone/v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/gone-io/gone)](https://goreportcard.com/report/github.com/gone-io/gone)
[![codecov](https://codecov.io/gh/gone-io/gone/graph/badge.svg?token=H3CROTTDZ1)](https://codecov.io/gh/gone-io/gone)
[![Build and Test](https://github.com/gone-io/gone/actions/workflows/go.yml/badge.svg)](https://github.com/gone-io/gone/actions/workflows/go.yml)
[![Release](https://img.shields.io/github/release/gone-io/gone.svg?style=flat-square)](https://github.com/gone-io/gone/releases)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)

<img src="docs/assert/logo.png" width = "100" alt="logo" />


# 🚀 Gone - Go语言轻量级依赖注入框架

## 💡 框架简介

Gone 是一个基于Golang标签的轻量级依赖注入框架，通过简洁的注解实现组件依赖管理。下面是一个典型的使用示例（嵌入了gone.Flag的结构体，我们称之为Goner）：

```go
type Dep struct {
	gone.Flag
	Name string
}

type Component struct {
	gone.Flag
	dep *Dep        `gone:"*"` //依赖注入
	log gone.Logger `gone:"*"` //注入 gone.Logger

  // 注入配置, 从环境变量 GONE_NAME 中获取值；如果使用goner/viper 等组件可以可以从配置文件或者配置中心获取值。
  // 参考文档：https://github.com/gone-io/goner
  name string     `gone:"config:name"`
}

func (c *Component) Init() {
	c.log.Infof(c.dep.Name) //使用依赖
  c.log.Infof(c.name) //使用配置
}
```

## ✨ 核心特性

- **全面的依赖注入支持**
  - 结构体属性注入（支持私有字段）
  - 函数参数注入（按类型自动匹配）
  - 配置参数注入（支持环境变量、配置中心和配置文件）
  - 第三方组件注入（通过Provider机制）
  👉 [详细文档](docs/inject_CN.md)
- 支持为 Goner 定义初始化方法、服务启动停止方法及相关生命周期钩子函数，实现自动化的服务管理和自定义操作。
- 提供[生态goner组件库](https://github.com/gone-io/goner)，支持配置、日志、数据库、大模型、可观察等功能；
- 提供[脚手架工具gonectl](https://github.com/gone-io/gonectl)，支持项目创建、组件管理、代码生成、测试mock、编译和运行。

### 架构
<img src="docs/assert/architecture.png" width = "600" alt="architecture"/>

### 生命周期
<img src="docs/assert/flow.png" width = "600" alt="flow"/>

## 🏁 快速开始

### 环境准备
1. 安装必要工具
```bash
go install github.com/gone-io/gonectl@latest
go install go.uber.org/mock/mockgen@latest
```

### 创建项目
```bash
gonectl create myproject
cd myproject
```

### 运行项目
```bash
go mod tidy
gonectl run ./cmd/server
```

## 更新记录

👉🏻 https://github.com/gone-io/gone/releases


## 贡献

如果您发现了错误或有功能请求，可以随时[提交问题](https://github.com/gone-io/gone/issues/new)
，同时欢迎[提交拉取请求](https://github.com/gone-io/gone/pulls)。

## 联系方式

如果您有任何问题，欢迎通过以下方式联系我们：

- [Github 讨论](https://github.com/gone-io/gone/discussions)
- 扫码加微信，暗号：gone

  <img src="docs/assert/qr_dapeng.png" width = "250" alt="dapeng wx qr code"/>

## 许可证

`gone` 在 MIT 许可证下发布，详情请参阅 [LICENSE](./LICENSE) 文件。