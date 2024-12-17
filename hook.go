package gone

// BeforeStartProvider provides the BeforeStart hook registration function.
// This provider allows other components to register functions to be called before application start.
type BeforeStartProvider struct {
	Flag
	preparer *Preparer `gone:"*"`
}

func (s *BeforeStartProvider) Provide() (BeforeStart, error) {
	return s.preparer.beforeStart, nil
}

// AfterStartProvider provides the AfterStart hook registration function.
// This provider allows other components to register functions to be called after application start.
type AfterStartProvider struct {
	Flag
	preparer *Preparer `gone:"*"`
}

func (s *AfterStartProvider) Provide() (AfterStart, error) {
	return s.preparer.afterStart, nil
}

// BeforeStopProvider provides the BeforeStop hook registration function.
// This provider allows other components to register functions to be called before application stop.
type BeforeStopProvider struct {
	Flag
	preparer *Preparer `gone:"*"`
}

func (s *BeforeStopProvider) Provide() (BeforeStop, error) {
	return s.preparer.beforeStop, nil
}

// AfterStopProvider provides the AfterStop hook registration function.
// This provider allows other components to register functions to be called after application stop.
type AfterStopProvider struct {
	Flag
	preparer *Preparer `gone:"*"`
}

func (s *AfterStopProvider) Provide() (AfterStop, error) {
	return s.preparer.afterStop, nil
}
