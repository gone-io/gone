package gone

import (
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"
)

// Run 开始运行一个Gone程序；`gone.Run` 和 `gone.Serve` 的区别是：
// 1. gone.Serve启动的程序，主协程会调用 Heaven.WaitEnd 挂起等待停机信号，可以用于服务程序的开发
// 2. gone.Run启动的程序，主协程则不会挂起，运行完就结束，适合开发一致性运行的代码
//
//	    // 定义加载服务的Priest函数
//		func LoadServer(c Cemetery) error {
//			c.Bury(goneXorm.New())
//			c.Bury(goneGin.New())
//			return nil
//		}
//
//	    // 加载组件的Priest函数
//		func LoadComponent(c Cemetery) error {
//			c.Bury(componentA.New())
//			c.Bury(componentB.New())
//		}
//
//
//		gone.Run(LoadServer, LoadComponent)//开始运行
func Run(priests ...Priest) {
	AfterStopSignalWaitSecond = 0
	New(priests...).
		Install().
		Start().
		Stop()
}

// Serve 开始服务，参考[Run](#Run)
func Serve(priests ...Priest) {
	New(priests...).
		Install().
		Start().
		WaitEnd().
		Stop()
}

// New 新建Heaven; Heaven 代表了一个应用程序；
func New(priests ...Priest) Heaven {
	cemetery := newCemetery()
	h := heaven{
		Logger:     &defaultLogger{},
		cemetery:   cemetery,
		priests:    priests,
		signal:     make(chan os.Signal),
		stopSignal: make(chan struct{}),
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

	signal     chan os.Signal
	stopSignal chan struct{}
}

func getAngelType() reflect.Type {
	var angelPtr *Angel = nil
	return reflect.TypeOf(angelPtr).Elem()
}

func (h *heaven) SetLogger(logger Logger) SetLoggerError {
	h.Logger = logger
	return nil
}

func (h *heaven) GetHeavenStopSignal() <-chan struct{} {
	return h.stopSignal
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

func (h *heaven) Install() Heaven {
	h.install()
	h.installAngelHook()
	return h
}

func (h *heaven) Start() Heaven {
	h.startFlow()
	return h
}

func (h *heaven) WaitEnd() Heaven {
	signal.Notify(h.signal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	s := <-h.signal
	h.Infof("gone system will quit for receive signal(%s)\n", s.String())
	return h
}

func (h *heaven) End() Heaven {
	h.signal <- syscall.SIGINT
	return h
}

// AfterStopSignalWaitSecond 收到停机信号后，退出程序等待的时间
var AfterStopSignalWaitSecond = 10

func (h *heaven) Stop() Heaven {
	h.stopFlow()
	close(h.stopSignal)

	if AfterStopSignalWaitSecond > 0 {
		h.Infof("WAIT %d SECOND TO STOP!!", AfterStopSignalWaitSecond)
	}
	for i := 0; i < AfterStopSignalWaitSecond; i++ {
		h.Infof("Stop in %d seconds.", AfterStopSignalWaitSecond-i)
		<-time.After(time.Second)
	}
	return h
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
