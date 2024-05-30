package gone

import (
	"errors"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"
)

// New 新建Heaven; Heaven 代表了一个应用程序；
func New(priests ...Priest) Heaven {
	cemetery := newCemetery()
	h := heaven{
		SimpleLogger:              &defaultLogger{},
		cemetery:                  cemetery,
		priests:                   priests,
		signal:                    make(chan os.Signal),
		stopSignal:                make(chan struct{}),
		afterStopSignalWaitSecond: AfterStopSignalWaitSecond,
	}

	h.
		cemetery.
		Bury(&h, IdGoneHeaven, IsDefault(true)).
		Bury(cemetery, IdGoneCemetery, IsDefault(true))
	return &h
}

type heaven struct {
	Flag

	SimpleLogger `gone:"gone-logger"`
	cemetery     Cemetery

	priests []Priest

	beforeStartHandlers []Process
	afterStartHandlers  []Process
	beforeStopHandlers  []Process
	afterStopHandlers   []Process

	signal     chan os.Signal
	stopSignal chan struct{}

	afterStopSignalWaitSecond int
}

func (h *heaven) SetAfterStopSignalWaitSecond(sec int) {
	h.afterStopSignalWaitSecond = sec
}

func getAngelType() reflect.Type {
	var angelPtr *Angel = nil
	return reflect.TypeOf(angelPtr).Elem()
}

func (h *heaven) SetLogger(logger SimpleLogger) SetLoggerError {
	h.SimpleLogger = logger
	return nil
}

func (h *heaven) GetHeavenStopSignal() <-chan struct{} {
	return h.stopSignal
}

func (h *heaven) burial() {
	for _, priest := range h.priests {
		err := priest(h.cemetery)
		h.panicOnError(err)
	}
}

func (h *heaven) install() {
	h.burial()

	err := h.cemetery.ReviveAllFromTombs()
	h.panicOnError(err)
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
		h.panicOnError(err)
	}

	for _, after := range h.afterStartHandlers {
		err := after(h.cemetery)
		h.panicOnError(err)
	}
}

func (h *heaven) panicOnError(err error) {
	if err == nil {
		return
	}
	var iErr InnerError
	if errors.As(err, &iErr) {
		h.Errorf("%s\n", iErr.Error())
	}
	panic(err)
}

func (h *heaven) stopFlow() {
	// stop Handlers 逆序调用：先注册的后调用
	for i := len(h.beforeStopHandlers) - 1; i >= 0; i-- {
		before := h.beforeStopHandlers[i]
		err := before(h.cemetery)
		h.panicOnError(err)
	}

	for i := len(h.afterStopHandlers) - 1; i >= 0; i-- {
		before := h.afterStopHandlers[i]
		err := before(h.cemetery)
		h.panicOnError(err)
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
var AfterStopSignalWaitSecond = 5

func (h *heaven) Stop() Heaven {
	h.stopFlow()
	close(h.stopSignal)

	if h.afterStopSignalWaitSecond > 0 {
		h.Infof("WAIT %d SECOND TO STOP!!", h.afterStopSignalWaitSecond)
	}
	for i := 0; i < h.afterStopSignalWaitSecond; i++ {
		h.Infof("Stop in %d seconds.", h.afterStopSignalWaitSecond-i)
		<-time.After(time.Second)
	}
	return h
}

func (h *heaven) BeforeStart(p Process) Heaven {
	h.beforeStartHandlers = append([]Process{p}, h.beforeStartHandlers...)
	return h
}
func (h *heaven) AfterStart(p Process) Heaven {
	h.afterStartHandlers = append(h.afterStartHandlers, p)
	return h
}

func (h *heaven) BeforeStop(p Process) Heaven {
	h.beforeStopHandlers = append([]Process{p}, h.beforeStopHandlers...)
	return h
}
func (h *heaven) AfterStop(p Process) Heaven {
	h.afterStopHandlers = append(h.afterStopHandlers, p)
	return h
}
