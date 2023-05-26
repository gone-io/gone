# gone

[![license](https://img.shields.io/badge/license-GPL%20V3-blue)](LICENSE)
[![GoDoc](https://pkg.go.dev/badge/github.com/gone-io/gone.jsonvalue?utm_source=godoc)](http://godoc.org/github.com/gone-io/gone)

## 0. 框架定位
**做一个对Spring程序员最友好的Golang框架**


**广告**：长期寻觅一起完善和维护的框架的朋友："看你骨骼惊奇，就是你了🫵"  
  
**有意者请加微信👇，邀请你入群：**  

<img src=docs/assert/qr_dapeng.png width=200px />


## 1. 这是什么？

- 一个类似 **Java Spring** 的 **Golang** **依赖注入** 框架
- 一个不断完善的 **微服务解决方案**
- 更多信息，参考 [Gone Story](docs/gone-story.md)

## 2. 怎么使用？

> 所有的代码都应该封装到一个个叫 **Goner** 容器中，**Goner** 的概念可以类比 **Spring** 中的 **Spring Bean**

- **Goner** 是依赖注入的最小单位
- **Goner** 可以封装框架组件
- **Goner** 也可以是业务组件，比如一个 Service、一个 Controller、一个 Client、一个 Dao 等

> 下面是一个简单的例子，完整代码在[这里](https://github.com/gone-io/examples/tree/main/simple)

### 2.1. 编写一个 **Goner**

- 定一个 **`struct`**
- 组合 `gone.Flag`，将其标记为一个 Goner
- 定一个"构造函数"

- 如下：

  ```go
  package user

  import "github.com/gone-io/gone"

  // 1. 定义 Goner：userService
  type userService struct {
      gone.Flag //2. 聚合 gone.Flag，使其实现gone.Goner接口成为一个Goner
  }

  //NewUserService 3. 定义构造函数
  func NewUserService() gone.Goner {
      return &userService{}
  }
  ```

### 2.2. 给 **Goner** 依赖的属性注入值

- 假设 `user.userService` 的一个方法依赖`redis.Cache`

  ```go
  package user

  import (
      "fmt"
      "github.com/gone-io/examples/simple/interface/service"
      "github.com/gone-io/gone"
      "github.com/gone-io/gone/goner/redis"
  )

  // 1. 定义 Goner：userService
  type userService struct {
      gone.Flag             //2. 聚合 gone.Flag，使其实现gone.Goner接口成为一个Goner
      cache     redis.Cache `gone:"gone-redis-cache"` //4. 标记需要注入的依赖，这里表示在`cache`属性上注入一个ID=`gone-redis-cache`的 Goner 组件
  }

  func (s *userService) GetUserInfo(id int64) (*service.UserInfo, error) {
      info := new(service.UserInfo)
      key := fmt.Sprintf("user-%d", id)
      return info, s.cache.Get(key, info) //5.使用注入的依赖完成业务
  }

  // NewUserService 3. 定义 `userService` 构造函数
  func NewUserService() gone.Goner {
      return &userService{}
  }
  ```

- 假设 `student.studentService` 依赖 `redis.userService`
- 给 `student.studentService` 增加一个 `AfterRevive() gone.AfterReviveError`（Goner 上的`AfterRevive`在属性注入完后自动运行）

  ```go
  package student

  import (
      "github.com/gone-io/examples/simple/interface/service"
      "github.com/gone-io/gone"
      "github.com/gone-io/gone/goner/logrus"
  )

  // 1. 定义 Goner：studentService
  type studentService struct {
      gone.Flag                  // 2.  聚合 gone.Flag，使其实现gone.Goner接口成为一个Goner
      service.User `gone:"*"`    //4. 聚合 service.User，这里的 `gone:"*"` 表示 `按类型注入` 一个Goner
      log          logrus.Logger `gone:"gone-logger"` //6. 注入一个用于日志打印的Goner
  }

  func (s *studentService) GetStudentInfo(id int64) (*service.UserInfo, error) {
      return s.GetUserInfo(id) //5. 调用 User 的 `GetUserInfo` 来实现 `GetStudentInfo`方法
  }

  // AfterRevive 6.该方法会在 studentService 属性被注入完成后自动运行
  func (s *studentService) AfterRevive() gone.AfterReviveError {
      info, err := s.GetUserInfo(100)
      if err != nil {
          s.log.Errorf("get info err:%v", err) //调用日志Goner，打印错误日志
      } else {
          s.log.Infof("student info:%v", info) //调用日志Goner，打印student info
      }
      return nil
  }

  // NewStudentService 3. 定义 `studentService` 构造函数
  func NewStudentService() gone.Goner {
      return &studentService{}
  }

  ```

### 2.3. 启动程序

- 增加 main 函数，调用 gone.Run
- 给 gone.Run 方法提供一个 `Priest` 函数（在 **Gone** 中，**加载** **Goner** 的函数 叫 **Priest———牧师**；**Goner**
  其实是**逝者**的意思）
- 在 `Priest` 函数 中 “安葬” **Goner** （**Priest———牧师**，对 **Goner** 的加载过程叫 **Bury———安葬**
  ，[更多概念](docs/gone-story.md)）

  ```go
  package main

  import (
      "github.com/gone-io/examples/simple/student"
      "github.com/gone-io/examples/simple/user"
      "github.com/gone-io/gone"
  )

  // 1. 增加 main 函数，调用 gone.Run
  func main() {
      //2. 给 gone.Run 方法提供一个 `Priest` 函数
      gone.Run(Priest)
  }

  func Priest(cemetery gone.Cemetery) error {
      // 3. "安葬" Goner
      cemetery.Bury(user.NewUserService()) // 3.1 在 `Priest` 函数中 "安葬" `user.NewUserService()`构造出来的 Goner
      cemetery.Bury(student.NewStudentService()) // 3.2 在 `Priest` 函数中 "安葬" `user.NewStudentService()`构造出来的 Goner
      return nil
  }
  ```

## 3. 🌰 更多例子：

> 在[example](example)目录可以找到详细的例子，后续会补充完成的帮忙手册。

## 4. 🔣 组件库（👉🏻 更多组件正在开发中...，💪🏻 ヾ(◍°∇°◍)ﾉﾞ，🖖🏻）

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

## 5. ⚙️ TODO LIST

- grpc，封装 github.com/grpc/grpc
- 完善文档
- 完善英文文档
- 完善测试用例

## 6. ⚠️ 注意

- 尽量不用使用 struct（结构体）作为 `gone` 标记的字段，由于 struct 在 golang 中是值拷贝，可能导致相关依赖注入失败的情况
- 下面这些 Goner 上的方法都不应该是阻塞的
    - `AfterRevive(Cemetery, Tomb) ReviveAfterError`
    - `Start(Cemetery) error`
    - `Stop(Cemetery) error`
    - `Suck(conf string, v reflect.Value) SuckError`



---
**入群交流吧？添加微信👇︎，邀你入群！🤟**  

<img src=docs/assert/qr_nuoyi.png width=200px />
