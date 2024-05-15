<p align="left">
   English&nbsp ｜&nbsp <a href="README_CN.md">中文</a>
</p>
<img src="docs/assert/logo.png" width = "250" alt="logo" align=center />

- [💡Concepts](#concepts)
- [🌰 More Examples:](#-more-examples)
- [🪜🧰🛠️ Component Library (👉🏻 More components are under development... 💪🏻 ヾ(◍°∇°◍)ﾉﾞ 🖖🏻)](#️-component-library--more-components-are-under-development--ヾﾉﾞ-)
- [📚Full Documentation](#full-documentation)


# Gone

[![license](https://img.shields.io/badge/license-GPL%20V3-blue)](LICENSE) [![GoDoc](https://pkg.go.dev/badge/github.com/gone-io/gone.jsonvalue?utm_source=godoc)](http://godoc.org/github.com/gone-io/gone)
[![Go Report Card](https://goreportcard.com/badge/github.com/gone-io/gone)](https://goreportcard.com/report/github.com/gone-io/gone)
[![codecov](https://codecov.io/gh/gone-io/gone/graph/badge.svg?token=H3CROTTDZ1)](https://codecov.io/gh/gone-io/gone)

First and foremost, Gone is a lightweight, Golang-based dependency injection framework inspired by the Spring Framework in Java. Secondly, the Gone framework includes a series of built-in components that provide a complete set of web development solutions through these components, offering services configuration, logging, tracing, service calls, database access, message middleware, and other commonly used microservice capabilities.

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
