<p>
    <a href="goner-create-example.md">English</a>&nbsp ｜&nbsp 中文
</p>

# 如何给Gone框架编写Goner组件（上）——编写一个Goner对接Apollo配置中心

- [如何给Gone框架编写Goner组件（上）——编写一个Goner对接Apollo配置中心](#如何给gone框架编写goner组件上编写一个goner对接apollo配置中心)
	- [引言](#引言)
	- [Gone框架与Goner组件简介](#gone框架与goner组件简介)
	- [Apollo配置中心简介](#apollo配置中心简介)
	- [编写Apollo Goner组件的核心思路](#编写apollo-goner组件的核心思路)
	- [核心代码实现与讲解](#核心代码实现与讲解)
		- [1. Apollo客户端组件实现](#1-apollo客户端组件实现)
		- [2. 初始化Apollo客户端](#2-初始化apollo客户端)
		- [3. 配置获取实现](#3-配置获取实现)
		- [4. 配置值设置工具函数](#4-配置值设置工具函数)
		- [5. 配置监听和自动更新依赖注入的值](#5-配置监听和自动更新依赖注入的值)
		- [提供`gone.LoadFunc`函数，方便使用](#提供goneloadfunc函数方便使用)
	- [使用Apollo Goner组件的示例](#使用apollo-goner组件的示例)
		- [1. 编写本地配置文件，支持多种配置格式：JSON、YAML、TOML、Properties 等](#1-编写本地配置文件支持多种配置格式jsonyamltomlproperties-等)
		- [2. 在服务中使用Apollo配置](#2-在服务中使用apollo配置)
		- [3. 引入Apollo组件](#3-引入apollo组件)
	- [高级用法](#高级用法)
		- [1. 监听配置变更](#1-监听配置变更)
		- [2. 支持多命名空间](#2-支持多命名空间)
	- [最佳实践](#最佳实践)
	- [结论](#结论)
	- [参考资源](#参考资源)


## 引言

在微服务架构中，配置中心是一个非常重要的基础设施，它能够集中管理各个服务的配置信息，实现配置的动态更新。Apollo是携程开源的一款优秀的分布式配置中心，本文将详细讲解如何基于Gone框架编写一个Goner组件对接Apollo配置中心，实现配置的统一管理。

## Gone框架与Goner组件简介

Gone是一个基于Go语言的依赖注入框架，而Goner则是基于Gone框架开发的可复用组件。通过编写Goner组件，我们可以将特定功能模块化，便于在不同项目中复用。

## Apollo配置中心简介

Apollo配置中心主要由以下部分组成：
- 配置管理界面（Portal）：供用户管理配置
- 配置服务（ConfigService）：提供配置获取接口
- 客户端SDK：与服务端交互，获取/监听配置变化

## 编写Apollo Goner组件的核心思路

1. 先通过**goner viper**获取本地关于Apollo连接的配置信息
2. 封装Apollo客户端，提供**配置获取**和**监听能力**
3. 实现`gone.Configure`接口，将Apollo配置中心的值直接注入到需要的组件中
4. 实现配置自动更新机制，监控配置变更，并更新对应的变量值
5. 支持不同类型配置的解析和转换

## 核心代码实现与讲解

### 1. Apollo客户端组件实现

源代码：[apollo/client.go](https://github.com/gone-io/goner/blob/goner-example/apollo/client.go)

```go:https://github.com/gone-io/goner/blob/goner-example/apollo/client.go
type apolloClient struct {
	gone.Flag
	localConfigure gone.Configure
	apolloClient   agollo.Client

	changeListener *changeListener `gone:"*"`
	testFlag       gone.TestFlag   `gone:"*" option:"allowNil"`
	logger         gone.Logger     `gone:"*" option:"lazy"`

	appId                     string //`gone:"config,apollo.appId"`
	cluster                   string //`gone:"config,apollo.cluster"`
	ip                        string //`gone:"config,apollo.ip"`
	namespace                 string //`gone:"config,apollo.namespace"`
	secret                    string //`gone:"config,apollo.secret"`
	isBackupConfig            bool   //`gone:"config,apollo.isBackupConfig"`
	watch                     bool   //`gone:"config,apollo.watch"`
	useLocalConfIfKeyNotExist bool   //`gone:"config,apollo.useLocalConfIfKeyNotExist"`
}
```

`apolloClient`结构体定义了Apollo客户端组件的各个字段：
- 依赖注入的组件：`changeListener`、`testFlag`、`logger`
- Apollo配置项：`appId`、`cluster`、`ip`等
- 控制选项：`watch`（是否监听配置变更）、`useLocalConfIfKeyNotExist`（当配置不存在时是否使用本地配置）

### 2. 初始化Apollo客户端

```go:https://github.com/gone-io/goner/blob/goner-example/apollo/client.go
func (s *apolloClient) Init() {
	s.localConfigure = viper.New(s.testFlag)

	m := map[string]*tuple{
		"apollo.appId":                     {v: &s.appId, defaultVal: ""},
		"apollo.cluster":                   {v: &s.cluster, defaultVal: "default"},
		"apollo.ip":                        {v: &s.ip, defaultVal: ""},
		"apollo.namespace":                 {v: &s.namespace, defaultVal: "application"},
		"apollo.secret":                    {v: &s.secret, defaultVal: ""},
		"apollo.isBackupConfig":            {v: &s.isBackupConfig, defaultVal: "true"},
		"apollo.watch":                     {v: &s.watch, defaultVal: "false"},
		"apollo.useLocalConfIfKeyNotExist": {v: &s.useLocalConfIfKeyNotExist, defaultVal: "true"},
	}
	for k, t := range m {
		err := s.localConfigure.Get(k, t.v, t.defaultVal)
		if err != nil {
			panic(err)
		}
	}

	c := &config.AppConfig{
		AppID:          s.appId,
		Cluster:        s.cluster,
		IP:             s.ip,
		NamespaceName:  s.namespace,
		IsBackupConfig: s.isBackupConfig,
		Secret:         s.secret,
	}
	client, err := agollo.StartWithConfig(func() (*config.AppConfig, error) {
		return c, nil
	})
	if err != nil {
		panic(err)
	}
	s.apolloClient = client
	if s.watch {
		client.AddChangeListener(s.changeListener)
	}
}
```

`Init`方法完成了Apollo客户端的初始化工作：
1. 创建本地配置源（使用viper组件）
2. 从本地配置中读取Apollo相关配置项
3. 创建Apollo客户端配置
4. 启动Apollo客户端
5. 如果启用了配置监听，则添加变更监听器

### 3. 配置获取实现

```go:https://github.com/gone-io/goner/blob/goner-example/apollo/client.go
func (s *apolloClient) Get(key string, v any, defaultVal string) error {
	if s.watch {
		s.changeListener.Put(key, v)
	}

	if s.apolloClient == nil {
		return s.localConfigure.Get(key, v, defaultVal)
	}

	namespaces := strings.Split(s.namespace, ",")
	for _, ns := range namespaces {
		cache := s.apolloClient.GetConfigCache(ns)
		if cache != nil {
			if value, err := cache.Get(key); err == nil {
				err = setValue(v, value)
				if err != nil {
					s.warnf("try to set `%s` value err:%v\n", key, err)
				} else {
					return nil
				}
			} else {
				s.warnf("get `%s` value from apollo ns(%s) err:%v\n", key, ns, err)
			}
		}
	}
	if s.useLocalConfIfKeyNotExist {
		return s.localConfigure.Get(key, v, defaultVal)
	}
	return nil
}
```

`Get`方法是获取配置的核心实现：
1. 如果启用了配置监听，则将配置键与变量引用关联起来
2. 如果Apollo客户端未初始化，则从本地配置获取
3. 遍历所有命名空间，尝试从Apollo获取配置
4. 如果获取成功，则将配置值设置到变量中
5. 如果从Apollo获取失败且允许使用本地配置，则从本地配置获取

### 4. 配置值设置工具函数

```go:https://github.com/gone-io/goner/blob/goner-example/apollo/client.go
func setValue(v any, value any) error {
	if str, ok := value.(string); ok {
		return gone.ToError(gone.SetValue(reflect.ValueOf(v), v, str))
	} else {
		marshal, err := json.Marshal(value)
		if err != nil {
			return gone.ToError(err)
		}
		return gone.ToError(gone.SetValue(reflect.ValueOf(v), v, string(marshal)))
	}
}
```

`setValue`函数用于将配置值设置到变量中：
1. 如果配置值是字符串，则直接设置
2. 如果配置值是其他类型，则先转换为JSON字符串，再设置

### 5. 配置监听和自动更新依赖注入的值

```go:https://github.com/gone-io/goner/blob/goner-example/apollo/client.go
type changeListener struct {
    gone.Flag
    keyMap map[string]any
    logger gone.Logger `gone:"*" option:"lazy"`
}

func (c *changeListener) Init() {
    c.keyMap = make(map[string]any)
}

func (c *changeListener) Put(key string, v any) {
    c.keyMap[key] = v
}

func (c *changeListener) OnChange(event *storage.ChangeEvent) {
    for k, change := range event.Changes {
        if v, ok := c.keyMap[k]; ok && change.ChangeType == storage.MODIFIED {
            err := setValue(v, change.NewValue)
            if err != nil && c.logger != nil {
                c.logger.Warnf("try to change `%s` value  err: %v\n", k, err)
            }
        }
    }
}

func (c *changeListener) OnNewestChange(*storage.FullChangeEvent) {}
```

`changeListener`实现了Apollo客户端的配置变更监听接口：
- `Init`方法初始化一个map用于存储配置键与对应的变量引用
- `Put`方法将配置键与变量引用关联起来
- `OnChange`方法在配置变更时被调用，它会遍历所有变更的配置，找到对应的变量引用，并更新其值
- `OnNewestChange`方法是接口要求实现的，但在本例中没有具体逻辑

### 提供`gone.LoadFunc`函数，方便使用
```go:https://github.com/gone-io/goner/blob/goner-example/apollo/client.go
var load = gone.OnceLoad(func(loader gone.Loader) error {
    err := loader.
        Load(
            &apolloClient{},
            gone.Name(gone.ConfigureName),
            gone.IsDefault(new(gone.Configure)),
            gone.ForceReplace(),
        )
    if err != nil {
        return err
    }
    return loader.Load(&changeListener{})
})

func Load(loader gone.Loader) error {
    return load(loader)
}
```

这段代码使用`gone.OnceLoad`确保组件只被加载一次，并通过`loader.Load`方法注册了两个组件：
- `apolloClient`：实现了`gone.Configure`接口，用于获取配置
- `changeListener`：用于监听配置变更


## 使用Apollo Goner组件的示例

### 1. 编写本地配置文件，支持多种配置格式：JSON、YAML、TOML、Properties 等
```yml
# config/default.yml
apollo:
  appId: SampleApp
  cluster: default
  ip: http://127.0.0.1:8080
  namespace: application,test.yml
  secret: your-secret
  isBackupConfig: false
  watch: false
  useLocalConfIfKeyNotExist: true
```



### 2. 在服务中使用Apollo配置

```go
type MyService struct {
	gone.Flag
	
	// 服务配置
	serverPort int `gone:"config,server.port"`
	timeout    int `gone:"config,service.timeout"`
}

func (s *MyService) Init() {
	// 使用配置
	fmt.Printf("服务启动，端口：%d，超时：%d毫秒\n", s.serverPort, s.timeout)
}
```

### 3. 引入Apollo组件

```go
import (
	"github.com/gone-io/gone/v2"
	"github.com/gone-io/goner/apollo"
)

func main() {
	gone.
		Loads(
			apollo.Load,
			// 其他组件...
		).
		Load(&MyService{}).
		Run(func() {
			// ...
        })
}
```

## 高级用法

### 1. 监听配置变更

通过启用`apollo.watch`配置，可以实现配置的自动更新：

**注意**： 需要动态更新的字段，**必须使用指针类型**才有效。

```go
type MyService struct {
    gone.Flag
    
    // 服务配置
    serverPort *int `gone:"config,server.port"`
    timeout    *int `gone:"config,service.timeout"`
}

func (s *MyService) Init() {
	
	go func() {
		for {
            fmt.Printf("服务启动，端口：%d，超时：%d毫秒\n", *s.serverPort, *s.timeout)
			time.Sleep(2 * time.Second)
        }
    }
}
```

### 2. 支持多命名空间

Apollo支持多个命名空间，可以通过逗号分隔的方式在`apollo.namespace`中指定：
```yml
# config/default.yml

# 在配置文件中设置
apollo.namespace: application,test.yml,database
```

**注意**：不是properties类型的namespace需要带后缀名才能正常的从Apollo上获取到值。

这样，在获取配置时会依次从这些命名空间中查找。

## 最佳实践

1. **配置分层管理**：将配置按照应用、环境、集群等维度进行分层管理

2. **默认值处理**：获取配置时始终提供合理的默认值，避免因配置缺失导致应用崩溃

3. **优雅降级**：当Apollo服务不可用时，应用应能够使用本地缓存的配置继续运行

4. **配置变更验证**：对于关键配置的变更，应进行合理性验证，避免错误配置导致系统问题

5. **监控与告警**：对配置获取失败、配置变更等关键事件进行监控和告警

## 结论

通过本文的讲解，我们了解了如何基于Gone框架编写一个Goner组件对接Apollo配置中心，实现配置的统一管理。这种方式不仅简化了配置管理的复杂度，还提供了配置动态更新的能力，使应用更加灵活和可维护。

在实际项目中，我们可以根据需求对Apollo Goner组件进行扩展，例如添加更多的配置类型支持、增强配置变更的处理逻辑等，以满足不同场景的需求。

## 参考资源
1. Apollo官方文档：https://www.apolloconfig.com/
2. Gone框架文档：https://github.com/gone-io/gone
3. Apollo Go客户端：https://github.com/apolloconfig/agollo
