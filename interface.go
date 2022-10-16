package gone

import "reflect"

// Goner 逝者
type Goner interface {
	gone()
}

// GonerId 逝者ID
type GonerId string

// Tomb 坟墓，逝者的容器
type Tomb interface {
	SetId(GonerId) Tomb
	GetId() GonerId
	GetGoner() Goner
}

// Cemetery 墓园
type Cemetery interface {
	Bury(Goner, GonerId) Tomb        // 埋葬，将逝者埋葬到墓园
	ReplaceBury(Goner, GonerId) Tomb // 替换

	revive() error // 复活，对逝者进行复活，让他们升入天堂

	GetTomById(GonerId) Tomb
	GetTomByType(reflect.Type) []Tomb
}

type BuildError error
type Builder interface {
	Build(conf string, pointer interface{}) BuildError
}

type ReviveAfterError error
type ReviveAfter interface {
	After(Cemetery, Tomb) ReviveAfterError
}

//  Goner Example
//	type jim struct {
//		DeadFlag
//
//		XMan XMan `revive:"x-man"`
//	}
//
//	type XMan struct {
//		DeadFlag
//	}

// Digger 掘墓
type Digger func(Cemetery Cemetery) error

type Process func(Cemetery) error
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
