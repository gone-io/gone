package gone

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"
)

// GonerIds for Gone framework inner Goners
const (
	// IdGoneHeaven , The GonerId of Heaven Goner, which represents the program itself, and which is injected by default when it starts.
	IdGoneHeaven GonerId = "gone-heaven"

	// IdGoneCemetery , The GonerId of Cemetery Goner, which is Dependence Injection Key Goner, and which is injected by default.
	IdGoneCemetery GonerId = "gone-cemetery"

	// IdGoneTestKit , The GonerId of TestKit Goner, which is injected by default when using gone.Test or gone.TestAt to run test code.
	IdGoneTestKit GonerId = "gone-test-kit"

	//IdConfig , The GonerId of Config Goner, which can be used for Injecting Configs from files or envs.
	IdConfig GonerId = "config"

	//IdGoneConfigure , The GonerId of Configure Goner, which is used to read configs from devices.
	IdGoneConfigure GonerId = "gone-configure"

	// IdGoneTracer ,The GonerId of Tracer
	IdGoneTracer GonerId = "gone-tracer"

	// IdGoneLogger , The GonerId of Logger
	IdGoneLogger GonerId = "gone-logger"

	// IdGoneCMux , The GonerId of CMuxServer
	IdGoneCMux GonerId = "gone-cmux"

	// IdGoneGin , IdGoneGinRouter , IdGoneGinProcessor, IdGoneGinProxy, IdGoneGinResponser, IdHttpInjector;
	// The GonerIds of Goners in goner/gin, which integrates gin framework for web request.
	IdGoneGin              GonerId = "gone-gin"
	IdGoneGinRouter        GonerId = "gone-gin-router"
	IdGoneGinSysMiddleware GonerId = "gone-gin-sys-middleware"
	IdGoneGinProxy         GonerId = "gone-gin-proxy"
	IdGoneGinResponser     GonerId = "gone-gin-responser"
	IdHttpInjector         GonerId = "http"

	// IdGoneXorm , The GonerId of XormEngine Goner, which is for xorm engine.
	IdGoneXorm GonerId = "gone-xorm"

	// IdGoneRedisPool ,IdGoneRedisCache, IdGoneRedisKey, IdGoneRedisLocker, IdGoneRedisProvider
	// The GonerIds of Goners in goner/redis, which integrates redis framework for cache and locker.
	IdGoneRedisPool     GonerId = "gone-redis-pool"
	IdGoneRedisCache    GonerId = "gone-redis-cache"
	IdGoneRedisKey      GonerId = "gone-redis-key"
	IdGoneRedisLocker   GonerId = "gone-redis-locker"
	IdGoneRedisProvider GonerId = "gone-redis-provider"

	// IdGoneSchedule , The GonerId of Schedule Goner, which is for schedule in goner/schedule.
	IdGoneSchedule GonerId = "gone-schedule"

	// IdGoneReq , The GonerId of urllib.Client Goner, which is for request in goner/urllib.
	IdGoneReq GonerId = "gone-urllib"
)

const (
	RequestIdHeaderKey = "X-Request-Id"
	TraceIdHeaderKey   = "X-Trace-Id"
)

// PanicTrace used for getting panic stack
func PanicTrace(kb int, skip int) []byte {
	stack := make([]byte, kb<<10) //4KB
	length := runtime.Stack(stack, true)

	_, filename, fileLine, ok := runtime.Caller(skip)
	start := 0
	if ok {
		start = bytes.Index(stack, []byte(fmt.Sprintf("%s:%d", filename, fileLine)))
		stack = stack[start:length]
	}

	line := []byte("\n")
	start = bytes.Index(stack, line) + 1
	stack = stack[start:]
	end := bytes.LastIndex(stack, line)
	if end != -1 {
		stack = stack[:end]
	}

	e := []byte("\ngoroutine ")
	end = bytes.Index(stack, e)
	if end != -1 {
		stack = stack[:end]
	}
	stack = bytes.TrimRight(stack, "\n")
	return stack
}

// GetFuncName get function name
func GetFuncName(f any) string {
	return strings.TrimSuffix(runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name(), "-fm")
}

// GetInterfaceType get interface type
func GetInterfaceType[T any](t *T) reflect.Type {
	return reflect.TypeOf(t).Elem()
}

type defaultType struct {
	t reflect.Type
}

func (d defaultType) option() {}

func IsDefault[T any](t *T) GonerOption {
	return defaultType{t: GetInterfaceType(t)}
}

type provideType struct {
	t []reflect.Type
}

func (d provideType) option() {}

// Provide a kind of GonerOption, which can be used in burying Vampire2(which is implemented `Suck(conf string, v reflect.Value, field reflect.StructField) error`) to framework.
// Provide will get the type of the object and put it in the GonerOption. When A goner need to be injected,
// the framework will find the Vampire2 which Tag by the GonerOption and call the Suck method to create the goner and inject it.
func Provide(objs ...any) GonerOption {
	m := make(map[reflect.Type]any)

	for _, o := range objs {
		m[reflect.TypeOf(o)] = o
	}
	p := provideType{
		t: make([]reflect.Type, 0),
	}
	for o, _ := range m {
		p.t = append(p.t, o)
	}
	return p
}

// WrapNormalFnToProcess warp a func to Process
func WrapNormalFnToProcess(fn any) Process {
	return func(cemetery Cemetery) error {
		args, err := cemetery.InjectFuncParameters(fn, nil, nil)
		if err != nil {
			return err
		}

		results := reflect.ValueOf(fn).Call(args)
		for _, result := range results {
			if err, ok := result.Interface().(error); ok {
				return err
			}
		}
		return nil
	}
}

// IsCompatible t Type can put in goner
func IsCompatible(t reflect.Type, goner any) bool {
	gonerType := reflect.TypeOf(goner)

	switch t.Kind() {
	case reflect.Interface:
		return gonerType.Implements(t)
	case reflect.Struct:
		return gonerType.Elem() == t
	default:
		return gonerType == t
	}
}

func setFieldValue(v reflect.Value, ref any) {
	t := v.Type()

	switch t.Kind() {
	case reflect.Interface, reflect.Pointer, reflect.Slice, reflect.Map:
		v.Set(reflect.ValueOf(ref))
	default:
		v.Set(reflect.ValueOf(ref).Elem())
	}
	return
}

type timeUseRecord struct {
	UseTime time.Duration
	Count   int64
}

var mapRecord = make(map[string]*timeUseRecord)

// TimeStat record the time of function and avg time
func TimeStat(name string, start time.Time, logs ...func(format string, args ...any)) {
	since := time.Since(start)
	if mapRecord[name] == nil {
		mapRecord[name] = &timeUseRecord{}
	}
	mapRecord[name].UseTime += since
	mapRecord[name].Count++

	var log func(format string, args ...any)
	if len(logs) == 0 {
		log = func(format string, args ...any) {
			fmt.Printf(format, args...)
		}
	} else {
		log = logs[0]
	}

	log("%s executed %v times, took %v, avg: %v\n",
		name,
		mapRecord[name].Count,
		mapRecord[name].UseTime,
		mapRecord[name].UseTime/time.Duration(mapRecord[name].Count),
	)
}

func RunTest(fn any, priests ...Priest) {
	Prepare(priests...).testKit().Run(fn)
}

// Test Use for writing test cases, refer to [example](https://github.com/gone-io/gone/blob/main/example/test/goner_test.go)
func Test[T Goner](fn func(goner T), priests ...Priest) {
	RunTest(func(in struct {
		cemetery Cemetery `gone:"*"`
	}) {
		ft := reflect.TypeOf(fn)
		t := ft.In(0).Elem()
		theTombs := in.cemetery.GetTomByType(t)
		if len(theTombs) == 0 {
			panic(CannotFoundGonerByTypeError(t))
		}
		fn(theTombs[0].GetGoner().(T))
	}, priests...)
}

// TestAt Use for writing test cases, test a specific ID of Goner
func TestAt[T Goner](id GonerId, fn func(goner T), priests ...Priest) {
	RunTest(func(in struct {
		cemetery Cemetery `gone:"*"`
	}) {
		theTomb := in.cemetery.GetTomById(id)
		if theTomb == nil {
			panic(CannotFoundGonerByIdError(id))
		}
		g, ok := theTomb.GetGoner().(T)
		if !ok {
			panic(NotCompatibleError(reflect.TypeOf(g).Elem(), reflect.TypeOf(theTomb.GetGoner()).Elem()))
		}
		fn(g)
	}, priests...)
}

// NewBuryMockCemeteryForTest make a new Cemetery for test
func NewBuryMockCemeteryForTest() Cemetery {
	return newCemetery()
}

func (p *Preparer) testKit() *Preparer {
	type Kit struct {
		Flag
	}
	p.heaven.(*heaven).cemetery.Bury(&Kit{}, IdGoneTestKit)
	return p
}

/*
Test Use for writing test cases
example:
```go

	gone.Prepare(priests...).Test(func(in struct{
	    cemetery Cemetery `gone:"*"`
	}) {

	  // test code
	})

```
*/
func (p *Preparer) Test(fn any) {
	p.testKit().AfterStart(fn).Run()
}

// TagStringParse parse tag string to map
// example: "a=1,b=2" -> map[string]string{"a":"1","b":"2"}
func TagStringParse(conf string) map[string]string {
	conf = strings.TrimSpace(conf)
	specs := strings.Split(conf, ",")
	m := make(map[string]string)
	for _, spec := range specs {
		spec = strings.TrimSpace(spec)
		pairs := strings.Split(spec, "=")
		if len(pairs) == 1 && pairs[0] != "" {
			m[pairs[0]] = ""
		} else if len(pairs) > 1 && pairs[0] != "" {
			m[pairs[0]] = pairs[1]
		}
	}
	return m
}

type provider[T any] struct {
	Flag
	cemetery Cemetery `gone:"*"`
	create   func(tagConf string) (T, error)
}

func (p *provider[T]) Suck(conf string, v reflect.Value, field reflect.StructField) error {
	obj, err := p.create(conf)
	if err != nil {
		return ToError(err)
	}
	v.Set(reflect.ValueOf(obj))
	return nil
}

// NewProviderPriest create a provider priest function for goner from a function like: `func(tagConf string, injectableStructParam struct{}) (provideType T, err error)`
// example:
// ```go
// type MyGoner struct {}
//
//	func NewMyGoner(tagConf string, param struct{
//		depGoner1 MyGoner1 `gone:"*"` // inject dep
//		depGoner2 MyGoner2 `gone:"*"` // inject dep
//		configStr string `gone:"config,my.config.str"` // inject config from config file
//	}) (MyGoner, error) {
//
//		// do something
//		return MyGoner{}, nil
//	}
//
// var priest = NewProviderPriest(NewMyGoner)
// ```
func NewProviderPriest[T any, P any](fn func(tagConf string, param P) (T, error)) Priest {
	p := provider[T]{}
	p.create = func(tagConf string) (T, error) {
		args, err := p.cemetery.InjectFuncParameters(fn, func(pt reflect.Type, i int) any {
			if i == 0 {
				return tagConf
			}
			return nil
		}, nil)

		if err != nil {
			return *new(T), err
		}

		results := reflect.ValueOf(fn).Call(args)
		if results[1].IsNil() {
			return results[0].Interface().(T), nil
		}
		return *new(T), ToError(results[1].Interface())
	}

	return func(cemetery Cemetery) error {
		cemetery.Bury(&p, Provide(*new(T)))
		return nil
	}
}
