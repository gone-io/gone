<p>
   English&nbsp ｜&nbsp <a href="goner-create-example_CN.md">中文</a>
</p>

# How to Create a Goner Component for Gone Framework [Part 1] - Integrating with Apollo Configuration Center

- [How to Create a Goner Component for Gone Framework \[Part 1\] - Integrating with Apollo Configuration Center](#how-to-create-a-goner-component-for-gone-framework-part-1---integrating-with-apollo-configuration-center)
	- [Introduction](#introduction)
	- [Gone Framework and Goner Component Overview](#gone-framework-and-goner-component-overview)
	- [Apollo Configuration Center Overview](#apollo-configuration-center-overview)
	- [Core Approach to Developing Apollo Goner Component](#core-approach-to-developing-apollo-goner-component)
	- [Core Implementation and Explanation](#core-implementation-and-explanation)
		- [1. Apollo Client Component Implementation](#1-apollo-client-component-implementation)
		- [2. Initializing Apollo Client](#2-initializing-apollo-client)
		- [3. Configuration Retrieval Implementation](#3-configuration-retrieval-implementation)
		- [4. Configuration Value Setting Utility Function](#4-configuration-value-setting-utility-function)
		- [5. Configuration Monitoring and Automatic Value Updates](#5-configuration-monitoring-and-automatic-value-updates)
		- [Providing `gone.LoadFunc` for Easy Use](#providing-goneloadfunc-for-easy-use)
	- [Example Usage of Apollo Goner Component](#example-usage-of-apollo-goner-component)
		- [1. Writing Local Configuration Files (Supporting Multiple Formats: JSON, YAML, TOML, Properties, etc.)](#1-writing-local-configuration-files-supporting-multiple-formats-json-yaml-toml-properties-etc)
		- [2. Using Apollo Configuration in Services](#2-using-apollo-configuration-in-services)
		- [3. Importing Apollo Component](#3-importing-apollo-component)
	- [Advanced Usage](#advanced-usage)
		- [1. Monitoring Configuration Changes](#1-monitoring-configuration-changes)
		- [2. Supporting Multiple Namespaces](#2-supporting-multiple-namespaces)
	- [Best Practices](#best-practices)
	- [Conclusion](#conclusion)
	- [References](#references)

## Introduction

In microservice architecture, a configuration center is a crucial infrastructure component that centrally manages configuration information for various services and enables dynamic configuration updates. Apollo is an excellent distributed configuration center open-sourced by Ctrip. This article will explain in detail how to create a Goner component based on the Gone framework to integrate with Apollo configuration center, achieving unified configuration management.

## Gone Framework and Goner Component Overview

Gone is a dependency injection framework based on Go language, while Goner is a reusable component developed based on the Gone framework. By creating Goner components, we can modularize specific functionalities for reuse across different projects.

## Apollo Configuration Center Overview

Apollo configuration center consists of the following main parts:
- Configuration Management Interface (Portal): For user configuration management
- Configuration Service: Provides configuration retrieval interfaces
- Client SDK: Interacts with the server to retrieve/monitor configuration changes

## Core Approach to Developing Apollo Goner Component

1. First use **goner viper** to get local Apollo connection configuration information
2. Encapsulate Apollo client to provide **configuration retrieval** and **monitoring capabilities**
3. Implement `gone.Configure` interface to directly inject Apollo configuration center values into required components
4. Implement automatic configuration update mechanism to monitor configuration changes and update corresponding variable values
5. Support parsing and conversion of different configuration types

## Core Implementation and Explanation

### 1. Apollo Client Component Implementation

Source code: [apollo/client.go](https://github.com/gone-io/goner/blob/goner-example/apollo/client.go)

```go
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

The `apolloClient` struct defines various fields for the Apollo client component:
- Dependency-injected components: `changeListener`, `testFlag`, `logger`
- Apollo configuration items: `appId`, `cluster`, `ip`, etc.
- Control options: `watch` (whether to monitor configuration changes), `useLocalConfIfKeyNotExist` (whether to use local configuration when configuration doesn't exist)

### 2. Initializing Apollo Client

```go
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

The `Init` method completes Apollo client initialization:
1. Creates local configuration source (using viper component)
2. Reads Apollo-related configuration items from local configuration
3. Creates Apollo client configuration
4. Starts Apollo client
5. Adds change listener if configuration monitoring is enabled

### 3. Configuration Retrieval Implementation

```go
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

The `Get` method is the core implementation for configuration retrieval:
1. If configuration monitoring is enabled, associates the configuration key with the variable reference
2. If Apollo client is not initialized, retrieves from local configuration
3. Iterates through all namespaces, attempting to retrieve configuration from Apollo
4. If retrieval is successful, sets the configuration value to the variable
5. If retrieval from Apollo fails and local configuration use is allowed, retrieves from local configuration

### 4. Configuration Value Setting Utility Function

```go
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

The `setValue` function is used to set configuration values to variables:
1. If the configuration value is a string, sets it directly
2. If the configuration value is another type, first converts it to a JSON string, then sets it

### 5. Configuration Monitoring and Automatic Value Updates

```go
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

`changeListener` implements Apollo client's configuration change listening interface:
- `Init` method initializes a map to store configuration keys and corresponding variable references
- `Put` method associates configuration keys with variable references
- `OnChange` method is called when configuration changes, it iterates through all changed configurations, finds corresponding variable references, and updates their values
- `OnNewestChange` method is required by the interface but has no specific logic in this example

### Providing `gone.LoadFunc` for Easy Use
```go
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

This code uses `gone.OnceLoad` to ensure components are loaded only once, and registers two components through the `loader.Load` method:
- `apolloClient`: implements the `gone.Configure` interface for configuration retrieval
- `changeListener`: for monitoring configuration changes

## Example Usage of Apollo Goner Component

### 1. Writing Local Configuration Files (Supporting Multiple Formats: JSON, YAML, TOML, Properties, etc.)
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

### 2. Using Apollo Configuration in Services

```go
type MyService struct {
	gone.Flag
	
	// Service configuration
	serverPort int `gone:"config,server.port"`
	timeout    int `gone:"config,service.timeout"`
}

func (s *MyService) Init() {
	// Using configuration
	fmt.Printf("Service started, port: %d, timeout: %d milliseconds\n", s.serverPort, s.timeout)
}
```

### 3. Importing Apollo Component

```go
import (
	"github.com/gone-io/gone/v2"
	"github.com/gone-io/goner/apollo"
)

func main() {
	gone.
		Loads(
			apollo.Load,
			// Other components...
		).
		Load(&MyService{}).
		Run(func() {
			// ...
        })
}
```

## Advanced Usage

### 1. Monitoring Configuration Changes

By enabling `apollo.watch` configuration, automatic configuration updates can be achieved:

**Note**: Fields that need dynamic updates **must use pointer types** to be effective.

```go
type MyService struct {
    gone.Flag
    
    // Service configuration
    serverPort *int `gone:"config,server.port"`
    timeout    *int `gone:"config,service.timeout"`
}

func (s *MyService) Init() {
	
	go func() {
		for {
            fmt.Printf("Service running, port: %d, timeout: %d milliseconds\n", *s.serverPort, *s.timeout)
			time.Sleep(2 * time.Second)
        }
    }
}
```

### 2. Supporting Multiple Namespaces

Apollo supports multiple namespaces, which can be specified in `apollo.namespace` using comma separation:
```yml
# config/default.yml

# In configuration file
apollo.namespace: application,test.yml,database
```

**Note**: Non-properties type namespaces need to include the file extension to properly retrieve values from Apollo.

This way, configuration will be searched sequentially in these namespaces.

## Best Practices

1. **Layered Configuration Management**: Manage configurations by dimensions such as application, environment, and cluster

2. **Default Value Handling**: Always provide reasonable default values when retrieving configurations to avoid application crashes due to missing configurations

3. **Graceful Degradation**: Applications should be able to continue running using locally cached configurations when Apollo service is unavailable

4. **Configuration Change Validation**: Validate the rationality of critical configuration changes to avoid system issues caused by incorrect configurations

5. **Monitoring and Alerting**: Monitor and alert on key events such as configuration retrieval failures and configuration changes

## Conclusion

Through this article, we've learned how to create a Goner component based on the Gone framework to integrate with Apollo configuration center, achieving unified configuration management. This approach not only simplifies configuration management complexity but also provides dynamic configuration update capabilities, making applications more flexible and maintainable.

In actual projects, we can extend the Apollo Goner component based on requirements, such as adding support for more configuration types, enhancing configuration change handling logic, etc., to meet the needs of different scenarios.

## References
1. Apollo Official Documentation: https://www.apolloconfig.com/
2. Gone Framework Documentation: https://github.com/gone-io/gone
3. Apollo Go Client: https://github.com/apolloconfig/agollo