<p align="left">
   English&nbsp ｜&nbsp <a href="README_CN.md">中文</a>
</p>

[![license](https://img.shields.io/badge/license-GPL%20V3-blue)](LICENSE) 
[![GoDoc](https://pkg.go.dev/badge/github.com/gone-io/gone.jsonvalue?utm_source=godoc)](https://pkg.go.dev/github.com/gone-io/gone/v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/gone-io/gone)](https://goreportcard.com/report/github.com/gone-io/gone)
[![codecov](https://codecov.io/gh/gone-io/gone/graph/badge.svg?token=H3CROTTDZ1)](https://codecov.io/gh/gone-io/gone)
[![Build and Test](https://github.com/gone-io/gone/actions/workflows/go.yml/badge.svg)](https://github.com/gone-io/gone/actions/workflows/go.yml)
[![Release](https://img.shields.io/github/release/gone-io/gone.svg?style=flat-square)](https://github.com/gone-io/gone/releases)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)  

<img src="docs/assert/logo.png" width = "100" alt="logo" align=center />


- [Gone](#gone)
  - [Gone is a Golang tag-based dependency injection framework](#gone-is-a-golang-tag-based-dependency-injection-framework)
  - [Features](#features)
    - [Architecture](#architecture)
    - [Lifecycle](#lifecycle)
  - [Quick Start](#quick-start)
  - [Update History](#update-history)
    - [v2.0.5](#v205)
    - [v2.0.4](#v204)
    - [v2.0.3](#v203)
    - [v2.0.0](#v200)
    - [v1.2.1](#v121)
    - [v1.2.0](#v120)
    - [v1.1.1](#v111)
  - [Contribution](#contribution)
  - [Contact](#contact)
  - [License](#license)


# Gone
## Gone is a Golang tag-based dependency injection framework

Gone is a lightweight golang dependency injection framework. Here's a simple example (a structure that embeds gone.Flag, which we call a "Goner"):

```go
package main
import "github.com/gone-io/gone/v2"

type Dep struct {
    gone.Flag
    Name string
}
type Component struct {
    gone.Flag
    dep *Dep        `gone:"*"` //dependency injection
    log gone.Logger `gone:"*"`
}
func (c *Component) Init() {
    c.log.Infof(c.dep.Name) //using dependency
}
func main() {
    gone.
       NewApp().
       // Register and load components
       Load(&Dep{Name: "Component Dep"}).
       Load(&Component{}).
       //run
       Run()
}
```

## Features
- Supports struct property injection, including private field injection [👉🏻 Dependency Injection Introduction](docs/Inject_en.md)
- Supports function parameter injection, injecting by function parameter type [👉🏻 Dependency Injection Introduction](docs/Inject_en.md)
- Provider mechanism, supports injecting external components into Goners: [👉🏻 Gone V2 Provider Mechanism Introduction](docs/provider_en.md)
- Supports code generation, automatically completing component registration and loading (via [Gonectr](https://github.com/gone-io/gonectr))
- Supports interface-based mock unit testing
- Supports [Goner components](https://github.com/gone-io/goner) for microservice development
- Supports defining initialization methods `Init()` and `BeforeInit()` for Goners
- Supports writing service-type Goners: implementing `Start() error` and `Stop() error`, the framework will automatically call Start() and Stop() methods
- Supports hooks like `BeforeStart`, `AfterStart`, `BeforeStop`, `AfterStop` for executing custom operations when services start and stop

### Architecture
<img src="docs/assert/architecture.png" width = "600" alt="architecture"/>

### Lifecycle
<img src="docs/assert/flow.png" width = "600" alt="flow"/>

## Quick Start
1. Install [gonectr](https://github.com/gone-io/gonectr) and [mockgen](https://github.com/uber-go/mock/tree/main)
    ```bash
    go install github.com/gone-io/gonectr@latest
    go install go.uber.org/mock/mockgen@latest
    ```
2. Create a project
    ```bash
    gonectr create myproject
    ```
3. Run the project
    ```bash
    cd myproject
    go mod tidy
    gonectr run ./cmd/server
    ```

## Update History
### v2.0.5
- Added `option:"lazy"` tag to support lazy field injection, see [documentation](docs/lazy_fill_en.md)
- Note: Fields marked with `option:"lazy"` cannot be used in the `Init`, `Provide`, and `Inject` methods

### v2.0.4
- Added SetValue function for unified handling of various configuration value types
- Refactored existing type handling logic, using reflection to improve generality

### v2.0.3
- Added `option:"allowNil"` tag to support [graceful handling of optional dependencies](docs/allow_nil_en.md)
- Improved tests and documentation

### v2.0.0
Version 2 has been extensively updated, removing unnecessary concepts. Please refer to: [Gone@v2 Instructions](./docs/v2-update_en.md) before use.

### v1.2.1
- Defined **gone.Provider**, a factory function for injecting external components (structs, struct pointers, functions, interfaces) that are not **Goners** into Goner properties requiring injection
- Fixed `gone.NewProviderPriest` which couldn't generate Priests for interface-type **gone.Provider**
- Wrote test code for `goner/gorm`, completed other test codes; documentation updated.

### v1.2.0
- Provided a new `gone.GonerOption` that can inject by type, delegating the task of constructing type instances to a **Goner** that implements `Suck(conf string, v reflect.Value, field reflect.StructField) error`
- Provided a helper function for implementing **Goner Provider**: `func NewProviderPriest[T any, P any](fn func(tagConf string, param P) (T, error)) Priest`
- Provided strategy configuration for cluster mode in `goner/xorm`
- Improved `goner/gorm` code and functional testing, supporting connection to various databases.

### v1.1.1
- goner/xorm supports clusters and multiple databases, latest documentation: https://goner.fun/zh/references/xorm.html
- Added goner/gorm, encapsulating `gorm.io/gorm` for database access, currently only supports MySQL, in progress...

## Contribution
If you find errors or have feature requests, feel free to [submit an issue](https://github.com/gone-io/gone/issues/new), and [pull requests](https://github.com/gone-io/gone/pulls) are welcome.

## Contact
If you have any questions, please contact us through:
- [Github Discussions](https://github.com/gone-io/gone/discussions)
- Scan the QR code to add WeChat, with the message: gone

  <img src="docs/assert/qr_dapeng.png" width = "250" alt="dapeng wx qr code"/>

## License
`gone` is released under the MIT License, please see the [LICENSE](./LICENSE) file for details.