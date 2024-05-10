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

func (p *Preparer) Run() {
	AfterStopSignalWaitSecond = 0
	p.heaven.
		Install().
		Start().
		Stop()
}

func (p *Preparer) Serve() {
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
