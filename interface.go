package gone

import "reflect"

// Goner which is an abstraction of injectable objects: can inject other Goner, can be injected by other Goner.
type Goner interface {
	goneFlag()
}

type identity interface {
	GetId() GonerId
}

// GonerId Goner's id
type GonerId string

// Tomb container of Goner
type Tomb interface {
	SetId(GonerId) Tomb
	GetId() GonerId
	GetGoner() Goner
	GonerIsRevive(flags ...bool) bool
}

type SetLoggerError error
type DefaultLogger interface {
	SetLogger(logger Logger) SetLoggerError
}

// Cemetery which is for burying and reviving Goner
type Cemetery interface {
	//DefaultLogger

	Goner

	//bury(goner Goner, ids ...GonerId) Tomb

	//Bury a Goner to the Cemetery
	Bury(Goner, ...GonerId) Cemetery

	//ReplaceBury replace the Goner in the Cemetery with a new Goner
	ReplaceBury(Goner, GonerId) error

	//ReviveOne Revive a Goner from the Cemetery
	ReviveOne(goner any) (deps []Tomb, err error)

	//ReviveAllFromTombs Revive all Goner from the Cemetery
	ReviveAllFromTombs() error

	//reviveOneFromTomb(tomb Tomb) (deps []Tomb, err error)
	reviveDependence(tomb Tomb) (deps []Tomb, err error)

	//GetTomById return the Tomb by the GonerId
	GetTomById(GonerId) Tomb

	//GetTomByType return the Tombs by the GonerType
	GetTomByType(reflect.Type) []Tomb
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

	//DefaultLogger
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
