package gone

type Preparer struct {
	heaven Heaven
}

func (p *Preparer) BeforeStart(fn any) *Preparer {
	p.heaven.BeforeStart(WrapNormalFnToProcess(fn))
	return p
}

func (p *Preparer) AfterStart(fn any) *Preparer {
	p.heaven.AfterStart(WrapNormalFnToProcess(fn))
	return p
}

func (p *Preparer) BeforeStop(fn any) *Preparer {
	p.heaven.BeforeStop(WrapNormalFnToProcess(fn))
	return p
}

func (p *Preparer) AfterStop(fn any) *Preparer {
	p.heaven.AfterStop(WrapNormalFnToProcess(fn))
	return p
}

func (p *Preparer) SetAfterStopSignalWaitSecond(sec int) {
	p.heaven.SetAfterStopSignalWaitSecond(sec)
}

func (p *Preparer) Run(fns ...any) {
	p.SetAfterStopSignalWaitSecond(0)
	for _, fn := range fns {
		p.AfterStart(fn)
	}
	p.heaven.
		Install().
		Start().
		Stop()
}

func (p *Preparer) Serve(fns ...any) {
	for _, fn := range fns {
		p.AfterStart(fn)
	}
	p.heaven.
		Install().
		Start().
		WaitEnd().
		Stop()
}

func Prepare(priests ...Priest) *Preparer {
	h := New(priests...)

	return &Preparer{
		heaven: h,
	}
}

// Run 开始运行一个Gone程序；`gone.Run` 和 `gone.Serve` 的区别是：
// 1. gone.Serve启动的程序，主协程会调用 Heaven.WaitEnd 挂起等待停机信号，可以用于服务程序的开发
// 2. gone.Run启动的程序，主协程则不会挂起，运行完就结束，适合开发一致性运行的代码
func Run(priests ...Priest) {
	Prepare(priests...).Run()
}

// Serve 开始服务，参考[Run](#Run)
func Serve(priests ...Priest) {
	Prepare(priests...).Serve()
}
