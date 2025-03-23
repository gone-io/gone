# Gone框架依赖注入核心路径详解

- [Gone框架依赖注入核心路径详解](#gone框架依赖注入核心路径详解)
	- [1. 组件定义与Goner接口](#1-组件定义与goner接口)
	- [2. 组件加载过程](#2-组件加载过程)
		- [2.1 Core结构体](#21-core结构体)
		- [2.2 组件加载流程](#22-组件加载流程)
	- [3. 依赖检查与循环依赖检测](#3-依赖检查与循环依赖检测)
		- [3.1 依赖收集](#31-依赖收集)
		- [3.2 组件安装](#32-组件安装)
	- [4. 字段注入实现与Provider机制](#4-字段注入实现与provider机制)
		- [4.1 字段注入](#41-字段注入)
		- [4.2 Provider机制](#42-provider机制)
	- [5. 生命周期管理](#5-生命周期管理)
		- [5.1 组件初始化](#51-组件初始化)
		- [5.2 应用程序生命周期](#52-应用程序生命周期)
		- [5.3 生命周期钩子](#53-生命周期钩子)
	- [6. 函数参数注入](#6-函数参数注入)
	- [7. 依赖注入流程总结](#7-依赖注入流程总结)


Gone是一个轻量级的Go语言依赖注入框架，它通过简洁的API和灵活的组件管理机制，帮助开发者构建模块化、可测试的应用程序。本文将详细介绍Gone框架的依赖注入核心路径，从组件定义到依赖注入的完整流程。

## 1. 组件定义与Goner接口

Gone框架中的所有组件都必须实现`Goner`接口，这是一个标记接口，用于标识可以被Gone框架管理的组件。

```go
// Goner是所有由Gone管理的组件必须实现的基础接口
// 它作为一个标记接口，用于标识可以被加载到Gone容器中的类型
type Goner interface {
	goneFlag()
}
```

为了简化组件的定义，Gone提供了一个`Flag`结构体，任何嵌入了这个结构体的类型都自动实现了`Goner`接口：

```go
// Flag是一个标记结构体，用于标识可以被gone框架管理的组件
// 在其他结构体中嵌入这个结构体表示它可以用于gone的依赖注入
type Flag struct{}

func (g *Flag) goneFlag() {}
```

组件定义示例：

```go
// 定义一个简单的组件
type MyComponent struct {
    gone.Flag  // 嵌入Flag以实现Goner接口
    // 组件的字段
    Dependency *AnotherComponent `gone:"*"` // 使用gone标签声明依赖
}
```

## 2. 组件加载过程

Gone框架通过`Core.Load`方法加载组件到容器中。这个过程包括组件注册、Provider检测和依赖关系建立。

### 2.1 Core结构体

`Core`是Gone框架的核心，负责组件的加载、依赖注入和生命周期管理：

```go
type Core struct {
	Flag
	coffins []*coffin

	nameMap            map[string]*coffin
	typeProviderMap    map[reflect.Type]*wrapProvider
	typeProviderDepMap map[reflect.Type]*coffin
	loaderMap          map[LoaderKey]bool
	log                Logger `gone:"*"`
}
```

### 2.2 组件加载流程

`Core.Load`方法是组件加载的入口，它完成以下工作：

1. 创建组件的coffin包装器
2. 处理命名组件的注册
3. 检测并注册Provider
4. 应用加载选项（如默认实现、加载顺序等）

```go
func (s *Core) Load(goner Goner, options ...Option) error {
	if goner == nil {
		return NewInnerError("goner cannot be nil - must provide a valid Goner instance", LoadedError)
	}
	co := newCoffin(goner)

	if namedGoner, ok := goner.(NamedGoner); ok {
		co.name = namedGoner.GonerName()
	}

	// 应用加载选项
	for _, option := range options {
		if err := option.Apply(co); err != nil {
			return ToError(err)
		}
	}

	// 处理命名组件的注册
	if co.name != "" {
		// 检查名称冲突并处理
		// ...
	}

	// 添加到组件列表
	s.coffins = append(s.coffins, co)

	// 检测并注册Provider
	provider := tryWrapGonerToProvider(goner)
	if provider != nil {
		co.needInitBeforeUse = true
		co.provider = provider

		// 注册Provider
		// ...
	}
	return nil
}
```

## 3. 依赖检查与循环依赖检测

Gone框架在初始化组件前，会先检查依赖关系，确保没有循环依赖，并确定最佳的初始化顺序。

### 3.1 依赖收集

`Core.Check`方法负责收集所有组件之间的依赖关系，并检测循环依赖：

```go
func (s *Core) Check() ([]dependency, error) {
	// 收集依赖关系
	depsMap, err := s.collectDeps()
	if err != nil {
		return nil, ToError(err)
	}

	// 检查循环依赖并获取最佳初始化顺序
	deps, orders := checkCircularDepsAndGetBestInitOrder(depsMap)
	if len(deps) > 0 {
		return nil, circularDepsError(deps)
	}

	// 添加字段填充操作
	for _, co := range s.coffins {
		orders = append(orders, dependency{co, fillAction})
	}
	return RemoveRepeat(orders), nil
}
```

### 3.2 组件安装

`Core.Install`方法按照确定的顺序执行组件的初始化：

```go
func (s *Core) Install() error {
	// 获取初始化顺序
	orders, err := s.Check()
	if err != nil {
		return ToError(err)
	}

	// 按顺序执行字段填充和初始化
	for i, dep := range orders {
		if dep.action == fillAction {
			if err := s.safeFillOne(dep.coffin); err != nil {
				s.log.Debugf("failed to %s at order[%d]: %s", dep, i, err)
				return ToError(err)
			}
		}
		if dep.action == initAction {
			if err = s.safeInitOne(dep.coffin); err != nil {
				s.log.Debugf("failed to %s at order[%d]: %s", dep, i, err)
				return ToError(err)
			}
		}
	}
	return nil
}
```

## 4. 字段注入实现与Provider机制

Gone框架通过反射机制实现字段注入，并支持多种Provider机制来创建和提供依赖实例。

### 4.1 字段注入

`Core.fillOne`方法负责为组件注入依赖：

```go
func (s *Core) fillOne(coffin *coffin) error {
	goner := coffin.goner

	// 调用BeforeInit钩子
	if initiator, ok := goner.(BeforeInitiatorNoError); ok {
		initiator.BeforeInit()
	}

	// 使用反射获取结构体字段
	elem := reflect.TypeOf(goner).Elem()
	elemV := reflect.ValueOf(goner).Elem()

	// 遍历所有字段
	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)
		v := elemV.Field(i)

		// 查找gone标签
		if tag, ok := field.Tag.Lookup(goneTag); ok {
			goneName, extend := ParseGoneTag(tag)
			if goneName == "" {
				goneName = defaultProviderName
			}

			// 获取依赖
			co, err := s.getDepByName(goneName)
			if err != nil {
				return ToErrorWithMsg(err, fmt.Sprintf("failed to find dependency %q for field %q in type %q", goneName, field.Name, GetTypeName(elem)))
			}

			// 尝试直接注入兼容类型
			if IsCompatible(field.Type, co.goner) {
				v.Set(reflect.ValueOf(co.goner))
				continue
			}

			// 尝试使用Provider提供依赖
			if co.provider != nil && field.Type == co.provider.Type() {
				provide, err := co.provider.Provide(extend)
				if err != nil {
					return ToErrorWithMsg(err, fmt.Sprintf("provider %T failed to provide value for field %q in type %q", co.goner, field.Name, GetTypeName(elem)))
				} else if provide != nil {
					v.Set(reflect.ValueOf(provide))
					continue
				}
			}

			// 尝试使用NamedProvider
			if provider, ok := co.goner.(NamedProvider); ok {
				// ...
			}

			// 尝试使用StructFieldInjector
			if injector, ok := co.goner.(StructFieldInjector); ok {
				// ...
			}
		}
	}

	coffin.isFill = true
	return nil
}
```

### 4.2 Provider机制

Gone框架支持多种Provider接口，用于创建和提供依赖实例：

```go
// Provider是一个泛型接口，用于提供T类型的依赖
type Provider[T any] interface {
	Goner
	Provide(tagConf string) (T, error)
}

// NamedProvider接口用于基于名称和类型创建依赖
type NamedProvider interface {
	NamedGoner
	Provide(tagConf string, t reflect.Type) (any, error)
}

// NoneParamProvider是一个简化的Provider接口
type NoneParamProvider[T any] interface {
	Goner
	Provide() (T, error)
}
```

`Core.Provide`方法实现了依赖查找和创建的逻辑：

```go
func (s *Core) Provide(tagConf string, t reflect.Type) (any, error) {
	// 尝试使用类型Provider
	if provider, ok := s.typeProviderMap[t]; ok && provider != nil {
		provide, err := provider.Provide(tagConf)
		if err != nil {
			s.log.Warnf("provider %T failed to provide value for type %s: %v",
				provider, GetTypeName(t), err)
		} else if provide != nil {
			return provide, nil
		}
	}

	// 尝试查找默认实现
	c := s.getDefaultCoffinByType(t)
	if c != nil {
		return c.goner, nil
	}

	// 处理切片类型的特殊情况
	if t.Kind() == reflect.Slice {
		// ...
	}

	return nil, NewInnerError(
		fmt.Sprintf("no provider or compatible type found for %s", GetTypeName(t)),
		NotSupport)
}
```

## 5. 生命周期管理

Gone框架提供了完整的组件生命周期管理，包括初始化、启动和停止阶段。

### 5.1 组件初始化

`Core.initOne`方法负责初始化单个组件：

```go
func (s *Core) initOne(c *coffin) error {
	goner := c.goner
	if initiator, ok := goner.(InitiatorNoError); ok {
		initiator.Init()
	}
	if initiator, ok := goner.(Initiator); ok {
		if err := initiator.Init(); err != nil {
			return ToError(err)
		}
	}
	c.isInit = true
	return nil
}
```

### 5.2 应用程序生命周期

`Application`结构体负责管理整个应用的生命周期，包括启动和停止：

```go
type Application struct {
	Flag

	loader  *Core    `gone:"*"`
	daemons []Daemon `gone:"*"`

	beforeStartHooks []Process
	afterStartHooks  []Process
	beforeStopHooks  []Process
	afterStopHooks   []Process

	signal chan os.Signal
}
```

`Application`提供了一系列方法来管理应用程序的生命周期：

```go
// 启动应用程序
func (s *Application) start() {
	// 执行启动前钩子
	for _, fn := range s.beforeStartHooks {
		fn()
	}

	// 启动所有守护进程
	for _, daemon := range s.daemons {
		err := daemon.Start()
		if err != nil {
			panic(err)
		}
	}

	// 执行启动后钩子
	for _, fn := range s.afterStartHooks {
		fn()
	}
}

// 停止应用程序
func (s *Application) stop() {
	// 执行停止前钩子
	for _, fn := range s.beforeStopHooks {
		fn()
	}

	// 按照相反的顺序停止所有守护进程
	for i := len(s.daemons) - 1; i >= 0; i-- {
		err := s.daemons[i].Stop()
		if err != nil {
			panic(err)
		}
	}

	// 执行停止后钩子
	for _, fn := range s.afterStopHooks {
		fn()
	}
}
```

### 5.3 生命周期钩子

Gone框架提供了多种生命周期钩子，允许组件在应用程序的不同阶段执行自定义逻辑：

```go
// 定义钩子函数类型
type Process func()

// 启动前钩子
type BeforeStart func(Process)

// 启动后钩子
type AfterStart func(Process)

// 停止前钩子
type BeforeStop func(Process)

// 停止后钩子
type AfterStop func(Process)
```

钩子的注册和执行顺序：

```go
// 注册启动前钩子
func (s *Application) beforeStart(fn Process) {
	// 注意：启动前钩子是按照后进先出的顺序执行的
	s.beforeStartHooks = append([]Process{fn}, s.beforeStartHooks...)
}

// 注册启动后钩子
func (s *Application) afterStart(fn Process) {
	// 注意：启动后钩子是按照先进先出的顺序执行的
	s.afterStartHooks = append(s.afterStartHooks, fn)
}
```

## 6. 函数参数注入

Gone框架不仅支持结构体字段注入，还支持函数参数注入，这使得可以直接在函数中使用依赖：

```go
// 函数参数注入
func (s *Core) InjectFuncParameters(fn any, injectBefore FuncInjectHook, injectAfter FuncInjectHook) (args []reflect.Value, err error) {
	ft := reflect.TypeOf(fn)

	if ft.Kind() != reflect.Func {
		return nil, NewInnerError(fmt.Sprintf("cannot inject parameters: expected a function, got %v", ft.Kind()), NotSupport)
	}

	in := ft.NumIn()

	for i := 0; i < in; i++ {
		pt := ft.In(i)
		paramName := fmt.Sprintf("parameter #%d (%s)", i+1, GetTypeName(pt))

		injected := false

		// 尝试使用自定义注入钩子
		if injectBefore != nil {
			v := injectBefore(pt, i, false)
			if v != nil {
				args = append(args, reflect.ValueOf(v))
				injected = true
			}
		}

		// 尝试使用标准依赖注入
		if !injected {
			if v, err := s.Provide("", pt); err != nil && !IsError(err, NotSupport) {
				return nil, ToErrorWithMsg(err, fmt.Sprintf("failed to inject %s in %s", paramName, GetFuncName(fn)))
			} else if v != nil {
				args = append(args, reflect.ValueOf(v))
				injected = true
			}
		}

		// 尝试创建并填充结构体参数
		if !injected {
			// ...
		}

		// 尝试使用后置注入钩子
		if injectAfter != nil {
			// ...
		}

		if !injected {
			return nil, NewInnerError(fmt.Sprintf("no suitable injector found for %s in %s", paramName, GetFuncName(fn)), NotSupport)
		}
	}
	return
}
```

函数包装器，用于执行带有注入参数的函数：

```go
func (s *Core) InjectWrapFunc(fn any, injectBefore FuncInjectHook, injectAfter FuncInjectHook) (func() []any, error) {
	args, err := s.InjectFuncParameters(fn, injectBefore, injectAfter)
	if err != nil {
		return nil, err
	}

	return func() (results []any) {
		values := reflect.ValueOf(fn).Call(args)
		for _, arg := range values {
			// 处理返回值
			// ...
			results = append(results, arg.Interface())
		}
		return results
	}, nil
}
```

## 7. 依赖注入流程总结

Gone框架的依赖注入核心路径可以总结为以下几个步骤：

1. **组件定义**：通过嵌入`gone.Flag`结构体实现`Goner`接口，使组件可被Gone框架管理。

2. **组件加载**：使用`Core.Load`或`Application.Load`方法将组件加载到容器中，可以指定名称、默认实现等选项。

3. **依赖检查**：框架检查组件之间的依赖关系，确保没有循环依赖，并确定最佳的初始化顺序。

4. **依赖注入**：框架通过反射机制，为每个组件注入其依赖项，支持多种注入方式：
   - 直接注入兼容类型
   - 使用Provider创建依赖实例
   - 使用NamedProvider基于名称和类型创建依赖
   - 使用StructFieldInjector自定义注入逻辑

5. **组件初始化**：按照确定的顺序初始化组件，调用`BeforeInit`和`Init`方法。

6. **应用程序启动**：执行启动前钩子，启动所有守护进程，执行启动后钩子。

7. **应用程序运行**：应用程序正常运行，处理业务逻辑。

8. **应用程序停止**：执行停止前钩子，按照相反的顺序停止所有守护进程，执行停止后钩子。

通过这一系列步骤，Gone框架实现了灵活、强大的依赖注入机制，帮助开发者构建模块化、可测试的应用程序。