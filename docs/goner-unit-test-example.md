# 如何给Gone框架编写Goner组件（下）——给对接Apollo的Goner组件编写测试代码

- [如何给Gone框架编写Goner组件（下）——给对接Apollo的Goner组件编写测试代码](#如何给gone框架编写goner组件下给对接apollo的goner组件编写测试代码)
  - [引言](#引言)
  - [编写“可测试”的代码](#编写可测试的代码)
  - [对外部模块进行Mock](#对外部模块进行mock)
    - [对`gone.Configure`的Mock](#对goneconfigure的mock)
    - [对`startWithConfig`的Mock](#对startwithconfig的mock)
  - [编写测试代码](#编写测试代码)
    - [测试初始化逻辑](#测试初始化逻辑)
    - [测试配置获取功能](#测试配置获取功能)
    - [测试配置变更监听功能](#测试配置变更监听功能)
  - [总结](#总结)


> 本文源代码：[https://github.com/gone-io/goner/apollo](https://github.com/gone-io/goner/tree/v0.0.7/apollo)

## 引言
在上一篇文章[如何给Gone框架编写Goner组件[上]——编写一个Goner对接Apollo配置中心》](./goner-create-example.md)中，我们详细介绍了如何在Gone框架中实现一个Apollo配置中心组件。然而，仅仅实现功能是不够的，为了确保组件的可靠性和稳定性，我们必须为其编写充分的单元测试。本文以Apollo组件为例，深入探讨如何在Gone框架中构建高质量的单元测试，帮助开发者打造更健壮的组件。

## 编写“可测试”的代码
正如我在另一篇文章[《如何对Golang代码进行单元测试？》](https://blog.csdn.net/waitdeng/article/details/146349708)中提到的，编写单元测试的前提是编写“可测试”的代码，并采用设计可测试代码的实践方法。以以下代码为例，我们需要思考：

- 需要测试哪些部分？
- 如何对这些部分进行测试？

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

针对上述代码的测试较为困难，主要原因在于它依赖了两个外部系统：`viper` 和 `agollo`。其中，对于`viper`我们可以通过本地配置文件或环境变量来解决，而对于`agollo`则需要搭建一套Apollo服务，这在自动化测试环境中成本较高。  
因此，我们应关注的是`apolloClient`的初始化逻辑，而不必测试viper的配置读取或agollo的启动。为此，可以将对外部模块的依赖进行外部化，改写后的代码如下：

```go
func (s *apolloClient) init(localConfigure gone.Configure, startWithConfig func(loadAppConfig func() (*config.AppConfig, error)) (agollo.Client, error)) {
	type tuple struct {
		v          any
		defaultVal string
	}

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
		err := localConfigure.Get(k, t.v, t.defaultVal)
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
	client, err := startWithConfig(func() (*config.AppConfig, error) {
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

func (s *apolloClient) Init() {
	s.localConfigure = viper.New(s.testFlag)
	s.init(s.localConfigure, agollo.StartWithConfig)
}
```

通过这种改造，我们可以在测试时只关注`init()`函数的逻辑，而不必依赖实际的外部模块，从而大大降低了测试成本。

## 对外部模块进行Mock
针对改造后的`init()`函数，其依赖主要集中在两个方面：
- `localConfigure`（类型为`gone.Configure`）
- `startWithConfig`函数（签名为`func(loadAppConfig func() (*config.AppConfig, error)) (agollo.Client, error)`）

### 对`gone.Configure`的Mock
我们可以利用**mockgen**工具直接生成接口的模拟实现，命令如下：
```bash
go install go.uber.org/mock/mockgen@latest
mockgen -package=apollo github.com/gone-io/gone/v2 Configure > gone_mock_test.go
```

### 对`startWithConfig`的Mock
首先，利用**mockgen**生成`agollo.Client`接口的模拟实现：
```bash
mockgen -package=apollo github.com/apolloconfig/agollo/v4 Client > agollo_mock_test.go
```
然后，为测试`startWithConfig`构建一个模拟函数：
```go
mockClient := NewMockClient(ctrl)
mockedStartWithConfig = func(loadAppConfig func() (*config.AppConfig, error)) (agollo.Client, error) {
	return mockClient, nil
}
```

## 编写测试代码

### 测试初始化逻辑
该测试用例主要验证以下几点：
1. 配置项是否正确读取
2. 默认值是否生效
3. Apollo客户端是否被正确创建

```go
func TestApolloClient_Init(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 创建模拟对象
	localConfigure := NewMockConfigure(ctrl)

	// 设置模拟对象的行为
	localConfigure.EXPECT().Get("apollo.appId", gomock.Any(), "").Return(nil).Do(
		func(key string, v any, defaultVal string) {
			*(v.(*string)) = "testApp"
		},
	)
	// ... 对其他配置项进行相应的Mock设置 ...

	mockClient := NewMockClient(ctrl)

	// 创建apolloClient实例
	client := &apolloClient{
		changeListener: &changeListener{},
	}
	client.localConfigure = localConfigure

	// 执行初始化
	client.init(localConfigure, func(loadAppConfig func() (*config.AppConfig, error)) (agollo.Client, error) {
		return mockClient, nil
	})

	// 验证配置是否正确读取
	assert.Equal(t, "testApp", client.appId)
	assert.Equal(t, "default", client.cluster)
	// ... 对其他配置项进行验证 ...
}
```

### 测试配置获取功能
此测试用例涵盖了以下场景：
1. 成功从Apollo获取配置
2. 当Apollo获取失败时，能够回退到本地配置
3. 禁用本地配置时的行为

```go
func TestApolloClient_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 创建模拟对象
	localConfigure := NewMockConfigure(ctrl)
	mockClient := NewMockClient(ctrl)
	mockCache := NewMockCacheInterface(ctrl)

	// 设置模拟对象的行为
	mockClient.EXPECT().GetConfigCache("application").Return(mockCache).AnyTimes()
	mockCache.EXPECT().Get("test.key").Return("test-value", nil).AnyTimes()

	// 创建apolloClient实例
	client := &apolloClient{
		localConfigure:            localConfigure,
		apolloClient:              mockClient,
		namespace:                 "application",
		changeListener:            &changeListener{},
		watch:                     false,
		useLocalConfIfKeyNotExist: true,
	}

	// 测试从Apollo获取配置
	var value string
	err := client.Get("test.key", &value, "default-value")
	assert.Nil(t, err)
	assert.Equal(t, "test-value", value)

	// 测试在Apollo获取失败时使用本地配置
	mockCache.EXPECT().Get("test.not-exist").Return(nil, errors.New("key not found")).AnyTimes()
	localConfigure.EXPECT().Get("test.not-exist", gomock.Any(), "default-value").Return(nil).Do(
		func(key string, v any, defaultVal string) {
			*(v.(*string)) = "local-value"
		},
	)

	var localValue string
	err = client.Get("test.not-exist", &localValue, "default-value")
	assert.Nil(t, err)
	assert.Equal(t, "local-value", localValue)
}
```

### 测试配置变更监听功能
此测试用例主要验证：
1. 配置监听是否正确注册
2. 当配置发生变化时，值是否能被正确更新

```go
func TestApolloClient_Get_WithWatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 创建并设置必要的模拟对象
	// ...

	// 创建changeListener并初始化
	listener := &changeListener{}
	listener.Init()

	// 创建apolloClient实例，设置watch为true
	client := &apolloClient{
		// ...
		watch: true,
	}

	// 测试获取配置时，带有监听功能
	var value string
	err := client.Get("test.key", &value, "default-value")
	assert.Nil(t, err)
	assert.Equal(t, "test-value", value)

	// 验证监听器是否正确注册了该key
	_, exists := listener.keyMap["test.key"]
	assert.True(t, exists)

	// 模拟配置变更通知
	changes := make(map[string]*storage.ConfigChange)
	changes["test.key"] = &storage.ConfigChange{
		OldValue:   "test-value",
		NewValue:   "new-value",
		ChangeType: storage.MODIFIED,
	}

	changeEvent := &storage.ChangeEvent{
		Changes: changes,
	}

	// 触发配置变更通知
	listener.OnChange(changeEvent)

	// 验证配置值是否已被更新
	assert.Equal(t, "new-value", value)
}
```

## 总结
通过上述测试用例，我们实现了对Apollo组件核心功能的全面覆盖，主要体现在以下几点：

1. **依赖注入与接口抽象**  
   将外部依赖（如viper和agollo）外部化，使代码具备更好的可测试性。

2. **Mock外部模块**  
   使用mockgen生成模拟对象，避免了在测试环境中对实际Apollo服务的依赖，大大降低了测试成本。

3. **完善的测试场景设计**  
   覆盖了配置读取、获取和变更监听等关键功能，确保组件在各种场景下均能稳定运行。

4. **提升代码可维护性**  
   通过单元测试为后续代码维护和重构提供了可靠保障，同时也为其他Gone组件的开发提供了可借鉴的测试方法。

这种测试方法不仅能够确保组件功能的正确性，还能显著提高代码质量和开发效率，是构建健壮系统的重要实践。
