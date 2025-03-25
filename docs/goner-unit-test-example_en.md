# How to Write Unit Tests for Gone Framework's Goner Components [Part 2] - Writing Tests for Apollo Configuration Center Component
- [How to Write Unit Tests for Gone Framework's Goner Components \[Part 2\] - Writing Tests for Apollo Configuration Center Component](#how-to-write-unit-tests-for-gone-frameworks-goner-components-part-2---writing-tests-for-apollo-configuration-center-component)
  - [Introduction](#introduction)
  - [Writing "Testable" Code](#writing-testable-code)
  - [Mocking External Modules](#mocking-external-modules)
    - [Mocking `gone.Configure`](#mocking-goneconfigure)
    - [Mocking `startWithConfig`](#mocking-startwithconfig)
  - [Writing Test Cases](#writing-test-cases)
    - [Testing Initialization Logic](#testing-initialization-logic)
    - [Testing Configuration Retrieval](#testing-configuration-retrieval)
    - [Testing Configuration Change Listening](#testing-configuration-change-listening)
  - [Summary](#summary)


> Source code: [https://github.com/gone-io/goner/apollo](https://github.com/gone-io/goner/tree/v0.0.7/apollo)

## Introduction
In our previous article [How to Write Goner Components for Gone Framework [Part 1] - Creating a Goner Component for Apollo Configuration Center](./goner-create-example_en.md), we detailed how to implement an Apollo configuration center component in the Gone framework. However, implementing functionality alone is not enough. To ensure the component's reliability and stability, we must write comprehensive unit tests. Using the Apollo component as an example, this article delves into how to build high-quality unit tests in the Gone framework, helping developers create more robust components.

## Writing "Testable" Code
As mentioned in my other article [How to Write Unit Tests for Golang Code?](https://blog.csdn.net/waitdeng/article/details/146349708), the prerequisite for writing unit tests is creating "testable" code and adopting practices designed for testability. Looking at the following code example, we need to consider:

- Which parts need to be tested?
- How can we test these parts?

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

Testing the above code is challenging primarily because it depends on two external systems: `viper` and `agollo`. While we can handle `viper` through local configuration files or environment variables, setting up an Apollo service for `agollo` would be costly in an automated testing environment.  
Therefore, we should focus on testing the initialization logic of `apolloClient` rather than testing viper's configuration reading or agollo's startup. To achieve this, we can externalize the dependencies on external modules. Here's the refactored code:

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

With this refactoring, we can focus on testing the `init()` function's logic without depending on actual external modules, significantly reducing testing costs.

## Mocking External Modules
For the refactored `init()` function, the dependencies are mainly concentrated in two aspects:
- `localConfigure` (type `gone.Configure`)
- `startWithConfig` function (signature `func(loadAppConfig func() (*config.AppConfig, error)) (agollo.Client, error)`)

### Mocking `gone.Configure`
We can use the **mockgen** tool to directly generate mock implementations of the interface:
```bash
go install go.uber.org/mock/mockgen@latest
mockgen -package=apollo github.com/gone-io/gone/v2 Configure > gone_mock_test.go
```

### Mocking `startWithConfig`
First, use **mockgen** to generate mock implementations of the `agollo.Client` interface:
```bash
mockgen -package=apollo github.com/apolloconfig/agollo/v4 Client > agollo_mock_test.go
```
Then, build a mock function for testing `startWithConfig`:
```go
mockClient := NewMockClient(ctrl)
mockedStartWithConfig = func(loadAppConfig func() (*config.AppConfig, error)) (agollo.Client, error) {
	return mockClient, nil
}
```

## Writing Test Cases

### Testing Initialization Logic
This test case primarily verifies:
1. Whether configuration items are correctly read
2. Whether default values are effective
3. Whether the Apollo client is correctly created

```go
func TestApolloClient_Init(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock objects
	localConfigure := NewMockConfigure(ctrl)

	// Set mock object behavior
	localConfigure.EXPECT().Get("apollo.appId", gomock.Any(), "").Return(nil).Do(
		func(key string, v any, defaultVal string) {
			*(v.(*string)) = "testApp"
		},
	)
	// ... Set up mocks for other configuration items ...

	mockClient := NewMockClient(ctrl)

	// Create apolloClient instance
	client := &apolloClient{
		changeListener: &changeListener{},
	}
	client.localConfigure = localConfigure

	// Execute initialization
	client.init(localConfigure, func(loadAppConfig func() (*config.AppConfig, error)) (agollo.Client, error) {
		return mockClient, nil
	})

	// Verify if configurations are correctly read
	assert.Equal(t, "testApp", client.appId)
	assert.Equal(t, "default", client.cluster)
	// ... Verify other configuration items ...
}
```

### Testing Configuration Retrieval
This test case covers the following scenarios:
1. Successfully retrieving configuration from Apollo
2. Falling back to local configuration when Apollo retrieval fails
3. Behavior when local configuration is disabled

```go
func TestApolloClient_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock objects
	localConfigure := NewMockConfigure(ctrl)
	mockClient := NewMockClient(ctrl)
	mockCache := NewMockCacheInterface(ctrl)

	// Set mock object behavior
	mockClient.EXPECT().GetConfigCache("application").Return(mockCache).AnyTimes()
	mockCache.EXPECT().Get("test.key").Return("test-value", nil).AnyTimes()

	// Create apolloClient instance
	client := &apolloClient{
		localConfigure:            localConfigure,
		apolloClient:              mockClient,
		namespace:                 "application",
		changeListener:            &changeListener{},
		watch:                     false,
		useLocalConfIfKeyNotExist: true,
	}

	// Test retrieving configuration from Apollo
	var value string
	err := client.Get("test.key", &value, "default-value")
	assert.Nil(t, err)
	assert.Equal(t, "test-value", value)

	// Test using local configuration when Apollo retrieval fails
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

### Testing Configuration Change Listening
This test case primarily verifies:
1. Whether configuration listening is correctly registered
2. Whether values are correctly updated when configuration changes

```go
func TestApolloClient_Get_WithWatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create and set up necessary mock objects
	// ...

	// Create and initialize changeListener
	listener := &changeListener{}
	listener.Init()

	// Create apolloClient instance with watch set to true
	client := &apolloClient{
		// ...
		watch: true,
	}

	// Test configuration retrieval with listening functionality
	var value string
	err := client.Get("test.key", &value, "default-value")
	assert.Nil(t, err)
	assert.Equal(t, "test-value", value)

	// Verify if the listener correctly registered the key
	_, exists := listener.keyMap["test.key"]
	assert.True(t, exists)

	// Simulate configuration change notification
	changes := make(map[string]*storage.ConfigChange)
	changes["test.key"] = &storage.ConfigChange{
		OldValue:   "test-value",
		NewValue:   "new-value",
		ChangeType: storage.MODIFIED,
	}

	changeEvent := &storage.ChangeEvent{
		Changes: changes,
	}

	// Trigger configuration change notification
	listener.OnChange(changeEvent)

	// Verify if the configuration value has been updated
	assert.Equal(t, "new-value", value)
}
```

## Summary
Through the above test cases, we have achieved comprehensive coverage of the Apollo component's core functionality, primarily reflected in the following aspects:

1. **Dependency Injection and Interface Abstraction**  
   Externalizing external dependencies (such as viper and agollo) makes the code more testable.

2. **Mocking External Modules**  
   Using mockgen to generate mock objects avoids dependency on actual Apollo services in the testing environment, significantly reducing testing costs.

3. **Comprehensive Test Scenario Design**  
   Covering key functionalities such as configuration reading, retrieval, and change listening ensures the component operates stably in various scenarios.

4. **Improved Code Maintainability**  
   Unit tests provide reliable guarantees for subsequent code maintenance and refactoring while also offering referenceable testing methods for other Gone components.

This testing methodology not only ensures component functionality correctness but also significantly improves code quality and development efficiency, making it an essential practice in building robust systems.