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
	- [Dependency Injection and Startup](#dependency-injection-and-startup)
	- [🌐Web Service](#web-service)
	- [💡Concepts](#concepts)
	- [🌰 More Examples:](#-more-examples)
	- [🪜🧰🛠️ Component Library (👉🏻 More components are under development... 💪🏻 ヾ(◍°∇°◍)ﾉﾞ 🖖🏻)](#️-component-library--more-components-are-under-development--ヾﾉﾞ-)
	- [📚Full Documentation](#full-documentation)
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

## Dependency Injection and Startup
Here's an example:
```go
package main

import (
	"fmt"
	"github.com/gone-io/gone"
)

type Worker struct {
	gone.Flag // Anonymously embedding gone.Flag structure makes it a Goner, it can be injected as a dependency into other Goners or receive injections from other Goners.
	Name      string
}

func (w *Worker) Work() {
	fmt.Printf("I am %s, and I am working\n", w.Name)
}

type Manager struct {
	gone.Flag                         // Anonymously embedding gone.Flag structure makes it a Goner, it can be injected as a dependency into other Goners or receive injections from other Goners.
	*Worker   `gone:"manager-worker"` // Named injection GonerId="manager-worker" for Worker instance.
	workers   []*Worker               `gone:"*"` // Inject all Workers into an array.
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
				Bury(ordinary, ordinaryWorker2)
			return nil
		}).
		// Run method supports dependency injection of parameters in its function.
		Run(func(manager *Manager) {
			manager.Manage()
		})
}
```
Summary:
1. In the Gone framework, dependencies are abstracted as Goners, which can be injected into each other.
2. By anonymously embedding the gone.Flag, the structure implements the Goner interface.
3. Before starting, load all Goners into the framework using the Bury function.
4. Use the Run method to start, where the function supports dependency injection of parameters.

[Full Documentation](https://goner.fun/)

Let's use Gone to write a web service below!

## 🌐Web Service
```go
package main

import (
	"fmt"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner"
)

// Implement a Goner. What is a Goner? => https://goner.fun/guide/core-concept.html#goner-%E9%80%9D%E8%80%85
type controller struct {
	gone.Flag // Goner tag, when anonymously embedded, a structure implements Goner
	gone.RouteGroup `gone:"gone-gin-router"` // Inject root routes
}

// Implement the Mount method to mount routes; the framework will automatically execute this method
func (ctr *controller) Mount() gone.GinMountError {

	// Define request structure
	type Req struct {
		Msg string `json:"msg"`
	}

	// Register the handler for `POST /hello`
	ctr.POST("/hello", func(in struct {
		to  string `gone:"http,query"` // Inject http request Query parameter To
		req *Req   `gone:"http,body"`  // Inject http request Body
	}) any {
		return fmt.Sprintf("to %s msg is: %s", in.to, in.req.Msg)
	})

	return nil
}

func main() {
	// Start the service
	gone.Serve(func(cemetery gone.Cemetery) error {
		// Call the framework's built-in component, load the gin framework
		_ = goner.GinPriest(cemetery)

		// Bury a controller-type Goner in the cemetery
		// What does bury mean? => https://goner.fun/guide/core-concept.html#burying
		// What is a cemetery? => https://goner.fun/guide/core-concept.html#cemetery
		cemetery.Bury(&controller{})
		return nil
	})
}
```

Run the above code: go run main.go, the program will listen on port 8080. Test it using curl:
```bash
curl -X POST 'http://localhost:8080/hello' \
    -H 'Content-Type: application/json' \
	--data-raw '{"msg": "你好呀？"}'
```

The result is as follows:
```
{"code":0,"data":"to  msg is: 你好呀？"}
```
[Quick Start](https://goner.fun/quick-start/)

## 💡Concepts
> The code we write is ultimately lifeless unless it is run.
In Gone, components are abstracted as Goners, whose properties can inject other Goners. Before Gone starts, all Goners need to be buried in the cemetery; after Gone starts, all Goners will be resurrected to establish a Heaven, "everyone in Heaven is no longer incomplete, and what they want will be satisfied."

[Core Concepts](https://goner.fun/guide/core-concept.html)

## 🌰 More Examples:

> Detailed examples can be found in the [example](example) directory, and more will be completed in the future.

## 🪜🧰🛠️ Component Library (👉🏻 More components are under development... 💪🏻 ヾ(◍°∇°◍)ﾉﾞ 🖖🏻)
- [goner/cumx](goner/cmux),
  a wrapper for `github.com/soheilhy/cmux`, used to reuse the same port to implement multiple protocols;
- [goner/config](goner/config), used to implement configuration for **Gone-App**
- [goner/gin](goner/gin),
  a wrapper for `github.com/gin-gonic/gin`, providing web services
- [goner/logrus](goner/logrus),
  a wrapper for `github.com/sirupsen/logrus`, providing logging services
- [goner/tracer](goner/tracer),
  providing log tracing, providing a unified `tracer```markdown
  Id` for the same request chain
- [goner/xorm](goner/xorm),
  a wrapper for `xorm.io/xorm`, used for database access; when using it, import the database driver as needed;
- [goner/redis](goner/redis),
  a wrapper for `github.com/gomodule/redigo`, used for interacting with redis
- [goner/schedule](goner/schedule),
  a wrapper for `github.com/robfig/cron/v3`, used for setting timers
- [emitter](https://github.com/gone-io/emitter), encapsulates event handling, which can be used for **DDD**'s **Event Storm**
- [goner/urllib](goner/urllib),
  encapsulates `github.com/imroc/req/v3`, used for sending HTTP requests, and connects the traceId between server and client

## 📚[Full Documentation](https://goner.fun/)

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