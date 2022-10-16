package gone

import (
	"os"
	"os/signal"
	"reflect"
	"syscall"
)

// Run
// ```go
//
//	    // 加载服务
//		func LoadServer(c Cemetery) error {
//			c.Bury(goneXorm.New())
//			c.Bury(goneGin.New())
//			return nil
//		}
//
//	    // 加载组件
//		func LoadComponent(c Cemetery) error {
//			c.Bury(componentA.New())
//			c.Bury(componentB.New())
//		}
//
// gone.Run(LoadServer, LoadComponent)
//
// ```
func Run(digGraves ...Digger) {
	New(digGraves...).Start()
}

// New 新建Heaven
func New(digGraves ...Digger) Heaven {
	return &heaven{
		cemetery:  NewCemetery(),
		digGraves: digGraves,
		signal:    make(chan os.Signal),
	}
}

type heaven struct {
	cemetery Cemetery

	digGraves []Digger

	beforeStartHandlers []Process
	afterStartHandlers  []Process
	beforeStopHandlers  []Process
	afterStopHandlers   []Process

	signal chan os.Signal
}

func getAngelType() reflect.Type {
	var angelPtr *Angel = nil
	return reflect.TypeOf(angelPtr).Elem()
}

func (h *heaven) Start() {
	for _, digGrave := range h.digGraves {
		err := digGrave(h.cemetery)
		if err != nil {
			panic(err)
		}
	}

	err := h.cemetery.revive()
	if err != nil {
		panic(err)
	}

	angleTombs := h.cemetery.GetTomByType(getAngelType())
	for _, tomb := range angleTombs {
		angel := tomb.GetGoner().(Angel)
		h.BeforeStart(angel.Start)
		h.BeforeStop(angel.Stop)
	}

	// start Handlers 顺序调用：先注册的先调用
	for _, before := range h.beforeStartHandlers {
		err := before(h.cemetery)
		if err != nil {
			panic(err)
		}
	}

	for _, after := range h.afterStartHandlers {
		err := after(h.cemetery)
		if err != nil {
			panic(err)
		}
	}

	signal.Notify(h.signal, syscall.SIGINT, syscall.SIGTERM)
	<-h.signal

	// stop Handlers 逆序调用：先注册的后调用
	for i := len(h.beforeStopHandlers) - 1; i >= 0; i-- {
		before := h.beforeStopHandlers[i]
		err := before(h.cemetery)
		if err != nil {
			panic(err)
		}
	}

	for i := len(h.afterStopHandlers) - 1; i >= 0; i-- {
		before := h.afterStopHandlers[i]
		err := before(h.cemetery)
		if err != nil {
			panic(err)
		}
	}
}

func (h *heaven) Stop() {}

func (h *heaven) BeforeStart(p Process) Heaven {
	h.beforeStartHandlers = append(h.beforeStopHandlers, p)
	return h
}
func (h *heaven) AfterStart(p Process) Heaven {
	h.afterStopHandlers = append(h.afterStartHandlers, p)
	return h
}

func (h *heaven) BeforeStop(p Process) Heaven {
	h.beforeStopHandlers = append(h.beforeStopHandlers, p)
	return h
}
func (h *heaven) AfterStop(p Process) Heaven {
	h.afterStopHandlers = append(h.afterStopHandlers, p)
	return h
}
