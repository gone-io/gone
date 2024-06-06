package gone

import (
	"github.com/soheilhy/cmux"
	"net"
	"reflect"
	"xorm.io/xorm"
)

//go:generate sh -c "mockgen -package=gone -self_package=github.com/gone-io/gone -source=interface.go -destination=mock_test.go"

type Flag struct{}

func (g *Flag) goneFlag() {}

// Goner which is an abstraction of injectable objects: can inject other Goner, can be injected by other Goner.
type Goner interface {
	goneFlag()
}

type identity interface {
	GetId() GonerId
}

type GonerOption interface {
	option()
}

// GonerId Goner's id
type GonerId string

func (GonerId) option() {}

type IsDefault bool

func (IsDefault) option() {}

// Tomb container of Goner
type Tomb interface {
	SetId(GonerId) Tomb
	GetId() GonerId
	GetGoner() Goner
	GonerIsRevive(flags ...bool) bool

	IsDefault() bool
	SetDefault(isDefault bool) Tomb
}

type SetLoggerError error
type DefaultLogger interface {
	SetLogger(logger SimpleLogger) SetLoggerError
}

// Cemetery which is for burying and reviving Goner
type Cemetery interface {
	Goner

	//Bury a Goner to the Cemetery
	Bury(Goner, ...GonerOption) Cemetery

	//BuryOnce a Goner to the Cemetery, if the Goner is already in the Cemetery, it will be ignored
	BuryOnce(goner Goner, options ...GonerOption) Cemetery

	//ReplaceBury replace the Goner in the Cemetery with a new Goner
	ReplaceBury(Goner, GonerId) error

	//ReviveOne Revive a Goner from the Cemetery
	ReviveOne(goner any) (deps []Tomb, err error)

	//ReviveAllFromTombs Revive all Goner from the Cemetery
	ReviveAllFromTombs() error

	//GetTomById return the Tomb by the GonerId
	GetTomById(GonerId) Tomb

	//GetTomByType return the Tombs by the GonerType
	GetTomByType(reflect.Type) []Tomb

	InjectFuncParameters(fn any, injectBefore func(pt reflect.Type, i int) any, injectAfter func(pt reflect.Type, i int, obj *any)) (args []any, err error)
}

// Priest A function which has A Cemetery parameter, and return an error. use for Burying Goner
type Priest func(cemetery Cemetery) error

// Process a function which has a Cemetery parameter, and return an error. use for hooks
type Process func(cemetery Cemetery) error

type Heaven interface {

	//Install do some prepare before start
	Install() Heaven

	//WaitEnd make program block until heaven stop
	WaitEnd() Heaven

	//End send a signal to heaven to stop
	End() Heaven

	//Start make heaven start
	Start() Heaven
	Stop() Heaven

	//GetHeavenStopSignal return a channel to listen the signal of heaven stop
	GetHeavenStopSignal() <-chan struct{}

	//BeforeStart add a hook function which will execute before start;
	BeforeStart(Process) Heaven

	//AfterStart add a hook function which will execute after start
	AfterStart(Process) Heaven

	//BeforeStop add a hook function which will execute before stop
	BeforeStop(Process) Heaven

	//AfterStop add a hook function which will execute after stop
	AfterStop(Process) Heaven

	SetAfterStopSignalWaitSecond(sec int)
}

type AfterReviveError error

// Prophet A interface which has a AfterRevive method
type Prophet interface {
	Goner

	//AfterRevive A method which will execute after revive
	// Deprecate: use `AfterRevive() error` instead
	AfterRevive() AfterReviveError
}

type Prophet2 interface {
	Goner

	//AfterRevive A method which will execute after revive
	AfterRevive() error
}

type Angel interface {
	Goner
	Start(Cemetery) error
	Stop(Cemetery) error
}

type SuckError error
type Vampire interface {
	Goner
	Suck(conf string, v reflect.Value) SuckError
}

type Vampire2 interface {
	Goner
	Suck(conf string, v reflect.Value, field reflect.StructField) error
}

// Error normal error
type Error interface {
	error
	Msg() string
	Code() int
}

// InnerError which has stack
type InnerError interface {
	Error
	Stack() []byte
}

// Logger log interface
type Logger interface {
	Tracef(format string, args ...any)
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Printf(format string, args ...any)
	Warnf(format string, args ...any)
	Warningf(format string, args ...any)
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
	Panicf(format string, args ...any)

	Trace(args ...any)
	Debug(args ...any)
	Info(args ...any)
	Print(args ...any)
	Warn(args ...any)
	Warning(args ...any)
	Error(args ...any)
	Fatal(args ...any)
	Panic(args ...any)

	Traceln(args ...any)
	Debugln(args ...any)
	Infoln(args ...any)
	Println(args ...any)
	Warnln(args ...any)
	Warningln(args ...any)
	Errorln(args ...any)
	Fatalln(args ...any)
	Panicln(args ...any)
}

// Tracer Log tracking, which is used to assign a unified traceId to the same call link to facilitate log tracking.
type Tracer interface {

	//SetTraceId to set `traceId` to the calling function. If traceId is an empty string, an automatic one will
	//be generated. TraceId can be obtained by using the GetTraceId () method in the calling function.
	SetTraceId(traceId string, fn func())

	//GetTraceId Get the traceId of the current goroutine
	GetTraceId() string

	//Go Start a new goroutine instead of `go func`, which can pass the traceid to the new goroutine.
	Go(fn func())

	//Recover use for catch panic in goroutine
	Recover()

	//RecoverSetTraceId SetTraceId and Recover
	RecoverSetTraceId(traceId string, fn func())
}

type XormEngine interface {
	xorm.EngineInterface
	Transaction(fn func(session xorm.Interface) error) error
	Sqlx(sql string, args ...any) *xorm.Session
	GetOriginEngine() xorm.EngineInterface
}

//-----------

// CMuxServer cMux 服务，用于复用同一端口监听多种协议，参考文档：https://pkg.go.dev/github.com/soheilhy/cmux
type CMuxServer interface {
	Match(matcher ...cmux.Matcher) net.Listener
	MatchWithWriters(matcher ...cmux.MatchWriter) net.Listener
	GetAddress() string
}
