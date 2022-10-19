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
func Run(priests ...Priest) {
	New(priests...).Start()
}

// New 新建Heaven
func New(priests ...Priest) Heaven {
	cemetery := newCemetery()
	h := heaven{
		Logger:   &defaultLogger{},
		cemetery: cemetery,
		priests:  priests,
		signal:   make(chan os.Signal),
	}

	h.
		cemetery.
		Bury(&h, IdGoneHeaven).
		Bury(cemetery, IdGoneCemetery)
	return &h
}

type heaven struct {
	Flag

	Logger   `gone:"gone-logger"`
	cemetery Cemetery

	priests []Priest

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

func (h *heaven) burial() {
	for _, priest := range h.priests {
		err := priest(h.cemetery)
		if err != nil {
			panic(err)
		}
	}
}

func (h *heaven) install() {
	h.burial()

	err := h.cemetery.revive()
	if err != nil {
		panic(err)
	}
}

func (h *heaven) installAngelHook() {
	angleTombs := h.cemetery.GetTomByType(getAngelType())
	for _, tomb := range angleTombs {
		angel := tomb.GetGoner().(Angel)
		h.BeforeStart(angel.Start)
		h.BeforeStop(angel.Stop)
	}
}

func (h *heaven) startFlow() {
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
}

func (h *heaven) stopFlow() {
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

func (h *heaven) Start() {
	h.install()
	h.installAngelHook()

	signal.Notify(h.signal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	h.startFlow()
	s := <-h.signal
	h.Infof("gone system will quit for receive signal(%s)", s.String())
	h.stopFlow()
}

func (h *heaven) Stop() {
	h.signal <- syscall.SIGINT
}

func (h *heaven) BeforeStart(p Process) Heaven {
	h.beforeStartHandlers = append(h.beforeStartHandlers, p)
	return h
}
func (h *heaven) AfterStart(p Process) Heaven {
	h.afterStartHandlers = append(h.afterStartHandlers, p)
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
