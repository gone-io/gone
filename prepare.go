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

func (p *Preparer) Run(fns ...any) {
	AfterStopSignalWaitSecond = 0
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
