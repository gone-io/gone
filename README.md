<p>
   English&nbsp ｜&nbsp <a href="README_CN.md">中文</a>
</p>

[![license](https://img.shields.io/badge/license-MIT-blue)](LICENSE)
[![GoDoc](https://pkg.go.dev/badge/github.com/gone-io/gone.jsonvalue?utm_source=godoc)](https://pkg.go.dev/github.com/gone-io/gone/v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/gone-io/gone)](https://goreportcard.com/report/github.com/gone-io/gone)
[![codecov](https://codecov.io/gh/gone-io/gone/graph/badge.svg?token=H3CROTTDZ1)](https://codecov.io/gh/gone-io/gone)
[![Build and Test](https://github.com/gone-io/gone/actions/workflows/go.yml/badge.svg)](https://github.com/gone-io/gone/actions/workflows/go.yml)
[![Release](https://img.shields.io/github/release/gone-io/gone.svg?style=flat-square)](https://github.com/gone-io/gone/releases)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)

<img src="docs/assert/logo.png" width = "100" alt="logo" />


# 🚀 Gone - Lightweight Dependency Injection Framework for Go

## 💡 Framework Introduction

Gone is a lightweight dependency injection framework based on Golang tags, implementing component dependency management through concise annotations. Here is a typical usage example (a struct embedded with gone.Flag, which we call Goner):

```go
type Dep struct {
    gone.Flag
    Name string
}

type Component struct {
    gone.Flag
    dep *Dep        `gone:"*"` //Dependency injection
    log gone.Logger `gone:"*"` //Inject gone.Logger

    // Inject configuration, get value from environment variable GONE_NAME; 
    // if using components like goner/viper, values can be obtained from config 
    // files or config centers.
    // Reference documentation: https://github.com/gone-io/goner
    name string     `gone:"config:name"`
}

func (c *Component) Init() {
    c.log.Infof(c.dep.Name) //Use dependency
    c.log.Infof(c.name) //Use configuration
}
```

## ✨ Core Features

- **Comprehensive Dependency Injection Support**
  - Struct field injection (supports private fields)
  - Function parameter injection (auto-matching by type)
  - Configuration parameter injection (supports environment variables, config centers and config files)
  - Third-party component injection (via Provider mechanism)
  👉 [Detailed Documentation](docs/inject.md)
- Supports defining initialization methods, service start/stop methods and related lifecycle hook functions for Goners, enabling automated service management and custom operations.
- Provides [ecosystem goner components](https://github.com/gone-io/goner) supporting configuration, logging, database, LLM, observability and more.
- Provides [scaffolding tool gonectl](https://github.com/gone-io/gonectl) supporting project creation, component management, code generation, test mocking, compilation and running.

### Architecture
<img src="docs/assert/architecture.png" width = "800" alt="architecture"/>


## 🏁 Quick Start

### Environment Preparation
1. Install required tools
```bash
go install github.com/gone-io/gonectl@latest
go install go.uber.org/mock/mockgen@latest
```

### Create Project
```bash
gonectl create myproject
cd myproject
```

### Run Project
```bash
go mod tidy
gonectl run ./cmd/server
```

## More Documents

- 👉🏻 [docs](./docs)
- 👉🏻 [goner](https://github.com/gone-io/goner)
- 👉🏻 [gonectl](https://github.com/gone-io/gonectl)

## Release Notes

👉🏻 https://github.com/gone-io/gone/releases


## Contribution

If you find any bugs or have feature requests, feel free to [submit an issue](https://github.com/gone-io/gone/issues/new)
or [submit a pull request](https://github.com/gone-io/gone/pulls).

## Contact

If you have any questions, welcome to contact us through:

- [Github Discussions](https://github.com/gone-io/gone/discussions)
- Scan QR code to add WeChat, passcode: gone

  <img src="docs/assert/qr_dapeng.png" width = "250" alt="dapeng wx qr code"/>

## License

`gone` is released under the MIT License, see [LICENSE](./LICENSE) for details.
