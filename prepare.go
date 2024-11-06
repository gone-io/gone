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

func (p *Preparer) Load(goner Goner) *Preparer {
	p.heaven.GetCemetery().Bury(goner)
	return p
}
func (p *Preparer) Bury(goner Goner) *Preparer {
	return p.Load(goner)
}

func (p *Preparer) LoadPriest(priests ...Priest) *Preparer {
	for _, priest := range priests {
		err := priest(p.heaven.GetCemetery())
		if err != nil {
			panic(err)
		}
	}
	return p
}

func Prepare(priests ...Priest) *Preparer {
	h := New(priests...)

	return &Preparer{
		heaven: h,
	}
}

var Default = Prepare()

/*
Run A Gone Program；

gone.Run vs gone.Serve:

- gone.Run, The main goroutine never hangs, and the program is terminated when the main goroutine exits.

- gone.Serve, The main goroutine calls Heaven.WaitEnd and hangs, and the program waiting for the stop signal for exiting.
*/
func Run(priests ...Priest) {
	Prepare(priests...).Run()
}

// Serve Start for A Gone Server Program.
func Serve(priests ...Priest) {
	Prepare(priests...).Serve()
}
