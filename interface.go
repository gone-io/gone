package gone

import "reflect"

// Goner 逝者
type Goner interface {
	goneFlag()
}

type identity interface {
	GetId() GonerId
}

// GonerId 逝者ID
type GonerId string

// Tomb 坟墓，逝者的容器
type Tomb interface {
	SetId(GonerId) Tomb
	GetId() GonerId
	GetGoner() Goner
	GonerIsRevive(flags ...bool) bool
}

// Cemetery 墓园
type Cemetery interface {
	Goner

	bury(goner Goner, ids ...GonerId) Tomb
	Bury(Goner, ...GonerId) Cemetery  // 埋葬，将逝者埋葬到墓园
	ReplaceBury(Goner, GonerId) error // 替换性埋葬

	revive() error // 复活，对逝者进行复活，让他们升入天堂
	reviveOne(tomb Tomb) (deps []Tomb, err error)
	reviveOneDep(tomb Tomb) (deps []Tomb, err error)

	GetTomById(GonerId) Tomb
	GetTomByType(reflect.Type) []Tomb
}

type SuckError error
type Vampire interface {
	Suck(conf string, v reflect.Value) SuckError
}

type ReviveAfterError error
type ReviveAfter interface {
	After(Cemetery, Tomb) ReviveAfterError
}

// Priest 神父，负责给Goner下葬
type Priest func(cemetery Cemetery) error

type Process func(cemetery Cemetery) error
type Heaven interface {
	Start()
	Stop()

	BeforeStart(Process) Heaven
	AfterStart(Process) Heaven

	BeforeStop(Process) Heaven
	AfterStop(Process) Heaven
}

type Angel interface {
	Goner
	Start(Cemetery) error
	Stop(Cemetery) error
}
