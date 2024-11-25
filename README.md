<p align="left">
   English&nbsp ｜&nbsp <a href="README_CN.md">中文</a>
</p>

[![license](https://img.shields.io/badge/license-GPL%20V3-blue)](LICENSE) 
[![GoDoc](https://pkg.go.dev/badge/github.com/gone-io/gone.jsonvalue?utm_source=godoc)](http://godoc.org/github.com/gone-io/gone)
[![Go Report Card](https://goreportcard.com/badge/github.com/gone-io/gone)](https://goreportcard.com/report/github.com/gone-io/gone)
[![codecov](https://codecov.io/gh/gone-io/gone/graph/badge.svg?token=H3CROTTDZ1)](https://codecov.io/gh/gone-io/gone)
[![Build and Test](https://github.com/gone-io/gone/actions/workflows/go.yml/badge.svg)](https://github.com/gone-io/gone/actions/workflows/go.yml)
[![Release](https://img.shields.io/github/release/gone-io/gone.svg?style=flat-square)](https://github.com/gone-io/gone/releases)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)  

<img src="docs/assert/logo.png" width = "100" alt="logo" align=center />


- [Gone](#gone)
	- [What Is Gone?](#what-is-gone)
	- [Features](#features)
    - [Quick Start](#quick-start)
	- [Full Documentation](https://goner.fun/)
	- [Contributing](#contributing)
	- [Contact](#contact)
	- [License](#license)


# Gone
## What Is Gone?

Gone is a lightweight dependency injection framework for Golang, designed to integrate with a variety of third-party components, enabling rapid development of cloud-native microservices.

## Features
- Dependency injection: Supports automatic injection of struct fields and function parameters.
- **[Gonectr](https://github.com/gone-io/gonectr)**: Generates projects, auxiliary code, compiles, and starts the project.
- Unit testing solution: Mock testing based on interfaces.
- Multiple pluggable components: Supports cloud-native and microservices architectures.
  
<img src="docs/assert/architecture.png" width = "600" alt="architecture"/>

## Quick Start
1. Install [gonectr](https://github.com/gone-io/gonectr) and [mockgen](https://github.com/uber-go/mock/tree/main)
    ```bash
    go install github.com/gone-io/gonectr@latest
    go install go.uber.org/mock/mockgen@latest
    ```
2. Create a new project
    ```bash
    gonectr create myproject
    ```
3. Run the project
    ```bash
    cd myproject
    gonectr run ./cmd/server
    ```
    Or use run Make command if you have installed [make](https://www.gnu.org/software/make/):
    ```bash
    cd myproject
    make run
    ```
    Or with docker compose:
    ```bash
    cd myproject
    docker compose build
    docker compose up
    ```

## [Full Documentation](https://goner.fun/)

## Contributing
If you have a bug report or feature request, you can [open an issue](https://github.com/gone-io/gone/issues/new), and [pull requests](https://github.com/gone-io/gone/pulls) are also welcome.

## Changelog
### v1.2.1
- Introduced **gone.Provider**, a factory function for injecting external components (such as structs, struct pointers, functions, and interfaces) that are not **Goner** into Goners filed which tag by `gone`.
- Fixed an issue where `gone.NewProviderPriest` failed to create a Priest for **gone.Provider** instances that generate interface types.
- Added test cases for `goner/gorm` and completed other missing test cases; updated documentation accordingly.

### v1.2.0
- Introduced a new `gone.GonerOption`, enabling type-based injection by delegating the task of constructing injected type instances to a **Goner** that implements `Suck(conf string, v reflect.Value, field reflect.StructField) error`.
- Added a helper function for implementing **Goner Provider**: `func NewProviderPriest[T any, P any](fn func(tagConf string, param P) (T, error)) Priest`.
- Provided a strategy configuration solution for the cluster mode in `goner/xorm`.
- Improved the `goner/gorm` code and conducted functional tests to support integration with multiple databases.

### v1.1.1
- `goner/xorm` now supports clustering and multiple databases. Latest documentation: https://goner.fun/references/xorm.html
- Added `goner/gorm`, a wrapper for `gorm.io/gorm` for database access. Currently, only MySQL is supported, and improvements are ongoing.


## Contact
If you have questions, feel free to reach out to us in the following ways:
- [Github Discussion](https://github.com/gone-io/gone/discussions)

## License
`gone` released under MIT license, refer [LICENSE](./LICENSE) file.